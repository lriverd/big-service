package infrastructure

import (
	"context"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserFirestoreRepository struct {
	client     *firestore.Client
	collection string
}

func NewUserRepository(client *firestore.Client) *UserFirestoreRepository {
	return &UserFirestoreRepository{client: client, collection: "users"}
}

func (r *UserFirestoreRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, err
		}
		return nil, err
	}
	var user domain.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

func (r *UserFirestoreRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	iter := r.client.Collection(r.collection).Where("email", "==", email).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		return nil, err
	}
	var user domain.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

func (r *UserFirestoreRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = "user"
	}

	data := map[string]interface{}{
		"email":     user.Email,
		"name":      user.Name,
		"photoUrl":  user.PhotoURL,
		"role":      user.Role,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	}
	if user.PasswordHash != "" {
		data["passwordHash"] = user.PasswordHash
	}

	ref, _, err := r.client.Collection(r.collection).Add(ctx, data)
	if err != nil {
		return nil, err
	}
	user.ID = ref.ID
	return user, nil
}

func (r *UserFirestoreRepository) UpdateLastLoginAt(ctx context.Context, id string, t time.Time) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "lastLoginAt", Value: t},
	})
	return err
}

func (r *UserFirestoreRepository) Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	updates := []firestore.Update{
		{Path: "updatedAt", Value: time.Now().UTC()},
	}
	if req.Name != nil {
		updates = append(updates, firestore.Update{Path: "name", Value: *req.Name})
	}
	if req.PhotoURL != nil {
		updates = append(updates, firestore.Update{Path: "photoUrl", Value: *req.PhotoURL})
	}

	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *UserFirestoreRepository) IncrementReputationScore(ctx context.Context, id string, delta int) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "reputationScore", Value: firestore.Increment(delta)},
	})
	return err
}

func (r *UserFirestoreRepository) SetDailySpotLimitOverride(ctx context.Context, id string, limit int, expiresAt time.Time) error {
	_, err := r.client.Collection(r.collection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "dailySpotLimitOverride", Value: limit},
		{Path: "dailySpotLimitOverrideExpiresAt", Value: expiresAt},
	})
	return err
}

func (r *UserFirestoreRepository) List(ctx context.Context, limit, offset int, search string) ([]*domain.UserPublic, int, error) {
	query := r.client.Collection(r.collection).OrderBy("createdAt", firestore.Desc)

	// Firestore doesn't support LIKE queries, so we do prefix matching
	if search != "" {
		searchLower := strings.ToLower(search)
		query = r.client.Collection(r.collection).
			Where("name", ">=", searchLower).
			Where("name", "<=", searchLower+"\uf8ff")
	}

	// Get total count
	total, _ := r.countQuery(ctx, search)

	iter := query.Offset(offset).Limit(limit).Documents(ctx)
	var users []*domain.UserPublic
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		var u domain.User
		if err := doc.DataTo(&u); err != nil {
			continue
		}
		users = append(users, &domain.UserPublic{
			ID:       doc.Ref.ID,
			Name:     u.Name,
			PhotoURL: u.PhotoURL,
		})
	}
	return users, total, nil
}

func (r *UserFirestoreRepository) Count(ctx context.Context) (int, error) {
	return r.countQuery(ctx, "")
}

func (r *UserFirestoreRepository) countQuery(ctx context.Context, search string) (int, error) {
	var results []*firestore.DocumentSnapshot
	var err error
	if search != "" {
		searchLower := strings.ToLower(search)
		results, err = r.client.Collection(r.collection).
			Where("name", ">=", searchLower).
			Where("name", "<=", searchLower+"\uf8ff").
			Documents(ctx).GetAll()
	} else {
		results, err = r.client.Collection(r.collection).Documents(ctx).GetAll()
	}
	if err != nil {
		return 0, err
	}
	return len(results), nil
}
