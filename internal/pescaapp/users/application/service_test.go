package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/users/application"
	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
)

type mockUserRepo struct {
	findByIDFn  func(ctx context.Context, id string) (*domain.User, error)
	updateFn    func(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error)
	listFn      func(ctx context.Context, limit, offset int, search string) ([]*domain.UserPublic, int, error)
	findByEmail func(ctx context.Context, email string) (*domain.User, error)
	createFn    func(ctx context.Context, user *domain.User) (*domain.User, error)
	countFn     func(ctx context.Context) (int, error)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmail != nil {
		return m.findByEmail(ctx, email)
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return user, nil
}
func (m *mockUserRepo) Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int, search string) ([]*domain.UserPublic, int, error) {
	return m.listFn(ctx, limit, offset, search)
}
func (m *mockUserRepo) Count(ctx context.Context) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}
	return 0, nil
}

func TestUserService_GetUser(t *testing.T) {
	repo := &mockUserRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, Name: "Test", Email: "t@t.com"}, nil
		},
	}
	c := cache.New(time.Minute, time.Minute)
	svc := application.NewUserService(repo, c)

	user, err := svc.GetUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("expected u1, got %s", user.ID)
	}

	// Second call should hit cache
	user2, err := svc.GetUser(context.Background(), "u1")
	if err != nil || user2.ID != "u1" {
		t.Error("cache miss or error")
	}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	c := cache.New(time.Minute, time.Minute)
	svc := application.NewUserService(repo, c)

	_, err := svc.GetUser(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	name := "Updated"
	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
			return &domain.User{ID: id, Name: *req.Name}, nil
		},
	}
	c := cache.New(time.Minute, time.Minute)
	svc := application.NewUserService(repo, c)

	user, err := svc.UpdateUser(context.Background(), "u1", domain.UpdateUserRequest{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "Updated" {
		t.Errorf("expected Updated, got %s", user.Name)
	}
}

func TestUserService_ListUsers(t *testing.T) {
	repo := &mockUserRepo{
		listFn: func(ctx context.Context, limit, offset int, search string) ([]*domain.UserPublic, int, error) {
			return []*domain.UserPublic{{ID: "u1", Name: "Test"}}, 1, nil
		},
	}
	c := cache.New(time.Minute, time.Minute)
	svc := application.NewUserService(repo, c)

	users, total, err := svc.ListUsers(context.Background(), 20, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

