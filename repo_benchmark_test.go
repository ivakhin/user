package user

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoReplicaSetURI  = "mongodb://localhost:27017/?replicaSet=rs0"
	mongoStandaloneURI  = "mongodb://localhost:27018"
	mongoDBName         = "bench"
	mongoCollectionName = "bench"
	redisAddr           = "localhost:6379"
	usersCount          = 10000
)

func BenchmarkMongoDB(b *testing.B) {
	b.ReportAllocs()

	ctx := context.TODO()

	repo := &RepoMongoDB{
		Collection: mongoCollection(ctx, mongoStandaloneURI),
	}

	for i := 0; i < b.N; i++ {
		_, _ = repo.Read(ctx, rand.Intn(usersCount))
	}
}

func BenchmarkRedis(b *testing.B) {
	b.ReportAllocs()

	ctx := context.TODO()

	repo := &RepoRedis{
		Client: redis.NewClient(&redis.Options{Addr: redisAddr}), //nolint:exhaustruct
	}

	for i := 0; i < b.N; i++ {
		_, _ = repo.Read(ctx, rand.Intn(usersCount))
	}
}

func BenchmarkMongoDBWatch(b *testing.B) {
	b.ReportAllocs()

	ctx := context.TODO()

	repo, err := NewRepoMongoDBCached(
		ctx,
		mongoCollection(ctx, mongoReplicaSetURI),
		func(_ context.Context, err error) { panic(err) })
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		_, _ = repo.Read(ctx, rand.Intn(usersCount))
	}
}

func TestWriteFixtures(t *testing.T) {
	t.Skip() // comment line and run test to write fixtures

	writeFixtures(context.TODO(), usersCount)
}

func TestDropFixtures(t *testing.T) {
	t.Skip() // comment line and run test to drop fixtures

	ctx := context.TODO()

	assert.NoError(t, mongoCollection(ctx, mongoStandaloneURI).Database().Drop(ctx))
	assert.NoError(t, mongoCollection(ctx, mongoReplicaSetURI).Database().Drop(ctx))
	assert.NoError(t, redis.NewClient(&redis.Options{Addr: redisAddr}).FlushAll(ctx).Err()) //nolint:exhaustruct
}

func mongoCollection(ctx context.Context, uri string) *mongo.Collection {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	database := client.Database(mongoDBName, options.Database())

	return database.Collection(mongoCollectionName, options.Collection())
}

func writeFixtures(ctx context.Context, count int) {
	const batchSize = 10000

	writers := []RepoWriter{
		&RepoMongoDB{Collection: mongoCollection(ctx, mongoStandaloneURI)},
		&RepoMongoDB{Collection: mongoCollection(ctx, mongoReplicaSetURI)},
		&RepoRedis{Client: redis.NewClient(&redis.Options{Addr: redisAddr})}, //nolint:exhaustruct
	}

	users := make(Users, 0, batchSize)

	for i := 0; i < count; i++ {
		users = append(users, randUser(i))
		if len(users)%batchSize == 0 {
			writeFixtureBatch(ctx, writers, users)
			users = make(Users, 0, batchSize)
		}
	}
}

func writeFixtureBatch(ctx context.Context, writers []RepoWriter, users Users) {
	wg := sync.WaitGroup{}
	wg.Add(len(writers))

	for _, writer := range writers {
		writer := writer
		go func() {
			if err := writer.Write(ctx, users); err != nil {
				panic(err)
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
