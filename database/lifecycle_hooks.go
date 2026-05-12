package database

import "context"

type PostCreate interface {
	PostCreate(ctx context.Context, db Provider) error // function to call post-create
}

type PreUpdate interface {
	PreUpdate(ctx context.Context, db Provider, existingValues CrudRecord) error // function to call pre-update
}

type CanOnlyDelete interface {
	CrudRecord
	CanOnlyDelete(ctx context.Context, db Provider, userId UserId) bool
}

type PostUpdate interface {
	PostUpdate(ctx context.Context, db Provider) error // function to call post-update
}

type PreDelete interface {
	PreDelete(ctx context.Context, db Provider) error // function to call pre-delete
}

type PostDelete interface {
	PostDelete(ctx context.Context, db Provider) error // function to call post-delete
}
