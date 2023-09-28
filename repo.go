package user

import "context"

type Repo interface {
	Read(ctx context.Context, id int) (User, error)
}

type RepoWriter interface {
	Write(ctx context.Context, users Users) error
}
