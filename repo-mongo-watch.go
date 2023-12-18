package user

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("user not found")

type RepoMongoDBCached struct {
	Collection       *mongo.Collection
	mu               sync.RWMutex
	store            map[int]User
	decodeErrHandler func(context.Context, error)
	err              error
	closeFn          func()
}

func (r *RepoMongoDBCached) Read(_ context.Context, id int) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.err != nil {
		return User{}, r.err
	}

	user, ok := r.store[id]
	if !ok {
		return user, ErrNotFound
	}

	return user, nil
}

type message struct {
	Doc   User          `bson:"fullDocument"`
	Type  operationType `bson:"operationType"`
	DocID struct {
		ID int `bson:"_id"`
	} `bson:"documentKey"`
}

type operationType string

const (
	operationTypeInsert  operationType = "insert"
	operationTypeReplace operationType = "replace"
	operationTypeUpdate  operationType = "update"
	operationTypeDelete  operationType = "delete"
)

func NewRepoMongoDBCached(
	ctx context.Context,
	collection *mongo.Collection,
	errHandler func(context.Context, error),
) (*RepoMongoDBCached, error) {
	count, err := collection.CountDocuments(ctx, bson.M{}, options.Count())
	if err != nil {
		return nil, fmt.Errorf("get users count: %w", err)
	}

	r := &RepoMongoDBCached{
		Collection:       collection,
		mu:               sync.RWMutex{},
		store:            make(map[int]User, count),
		decodeErrHandler: errHandler,
		err:              nil,
	}

	go r.watch(ctx)

	if err := r.setAll(ctx); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *RepoMongoDBCached) Healthcheck() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.err
}

func (r *RepoMongoDBCached) Close() {
	r.closeFn()
}

//nolint:nlreturn
func (r *RepoMongoDBCached) watch(ctx context.Context) {
	var msg message

	ctx, r.closeFn = context.WithCancel(ctx)

	for {
		opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

		stream, err := r.Collection.Watch(ctx, mongo.Pipeline{}, opts)
		if err != nil {
			r.setErr(err)
			break
		}

		for stream.Next(ctx) {
			if err := stream.Err(); err != nil {
				r.setErr(err)
				break
			}

			err := stream.Decode(&msg)
			if err != nil {
				r.handleDecodeErr(ctx, err)
				continue
			}

			switch msg.Type {
			case operationTypeInsert, operationTypeReplace, operationTypeUpdate:
				r.set(msg.Doc)
			case operationTypeDelete:
				r.delete(msg.Doc.ID)
			}

			msg = message{} //nolint:exhaustruct
		}
	}
}

func (r *RepoMongoDBCached) setAll(ctx context.Context) error {
	cur, err := r.Collection.Find(ctx, bson.M{}, options.Find())
	if err != nil {
		return fmt.Errorf("mongodb get user cursor: %w", err)
	}

	var user User

	for cur.Next(ctx) {
		if err := cur.Err(); err != nil {
			return fmt.Errorf("mongodb user stream: %w", err)
		}

		if err := cur.Decode(&user); err != nil {
			r.decodeErrHandler(ctx, err)

			continue
		}

		r.set(user)
		user = User{} //nolint:exhaustruct
	}

	return nil
}

func (r *RepoMongoDBCached) set(user User) {
	r.mu.Lock()
	r.store[user.ID] = user
	r.mu.Unlock()
}

func (r *RepoMongoDBCached) delete(id int) {
	r.mu.Lock()
	delete(r.store, id)
	r.mu.Unlock()
}

func (r *RepoMongoDBCached) setErr(err error) {
	r.mu.Lock()
	r.err = err
	r.mu.Unlock()
}

func (r *RepoMongoDBCached) handleDecodeErr(ctx context.Context, err error) {
	if r.decodeErrHandler != nil {
		r.decodeErrHandler(ctx, err)
	}
}
