package domain

import "context"

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
	List(ctx context.Context, limit, offset int, search string) ([]*UserPublic, int, error)
	Count(ctx context.Context) (int, error)
}

type FavoriteRepository interface {
	List(ctx context.Context, userID string, limit, offset int) ([]*Favorite, int, error)
	Add(ctx context.Context, fav *Favorite) error
	Remove(ctx context.Context, userID, spotID string) error
	Exists(ctx context.Context, userID, spotID string) (bool, error)
}

