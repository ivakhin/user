package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

type RepoRedis struct {
	Client *redis.Client
}

func (r *RepoRedis) Read(ctx context.Context, id int) (User, error) {
	var user User

	cmd := r.Client.Get(ctx, strconv.Itoa(id))

	if err := cmd.Err(); err != nil {
		return user, fmt.Errorf("redis GET command: %w", err)
	}

	val, err := cmd.Result()
	if err != nil {
		return user, fmt.Errorf("redis GET result: %w", err)
	}

	if err := bson.Unmarshal([]byte(val), &user); err != nil {
		return user, fmt.Errorf("redis decode user: %w", err)
	}

	return user, nil
}

func (r *RepoRedis) Write(ctx context.Context, users Users) error {
	for _, user := range users {
		value, err := bson.Marshal(user)
		if err != nil {
			return fmt.Errorf("redis encode user: %w", err)
		}

		cmd := r.Client.Set(ctx, strconv.Itoa(user.ID), value, redis.KeepTTL)
		if err := cmd.Err(); err != nil {
			return fmt.Errorf("redis SET command: %w", err)
		}
	}

	return nil
}
