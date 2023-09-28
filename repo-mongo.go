package user

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RepoMongoDB struct {
	Collection *mongo.Collection
}

func (r *RepoMongoDB) Read(ctx context.Context, id int) (User, error) {
	var user User

	res := r.Collection.FindOne(ctx,
		bson.M{
			"_id": id,
		},
		options.FindOne())

	if err := res.Err(); err != nil {
		return user, fmt.Errorf("find user: %w", err)
	}

	if err := res.Decode(&user); err != nil {
		return User{}, fmt.Errorf("decode user: %w", err)
	}

	return user, nil
}

func (r *RepoMongoDB) Write(ctx context.Context, users Users) error {
	opts := options.InsertMany()

	_, err := r.Collection.InsertMany(ctx, users.toInterface(), opts)
	if err != nil {
		return fmt.Errorf("insert users: %w", err)
	}

	return nil
}

func (u Users) toInterface() []interface{} {
	res := make([]interface{}, 0, len(u))
	for _, user := range u {
		res = append(res, interface{}(user))
	}

	return res
}
