package domain

import "time"

type User struct {
	ID                              string     `json:"id" firestore:"-"`
	Email                           string     `json:"email" firestore:"email"`
	Name                            string     `json:"name" firestore:"name"`
	PhotoURL                        *string    `json:"photoUrl" firestore:"photoUrl"`
	Role                            string     `json:"role,omitempty" firestore:"role"`
	PasswordHash                    string     `json:"-" firestore:"passwordHash"`
	CreatedAt                       time.Time  `json:"createdAt" firestore:"createdAt"`
	UpdatedAt                       time.Time  `json:"updatedAt" firestore:"updatedAt"`
	LastLoginAt                     *time.Time `json:"lastLoginAt,omitempty" firestore:"lastLoginAt"`
	ReputationScore                 int        `json:"reputationScore" firestore:"reputationScore"`
	DailySpotLimitOverride          *int       `json:"-" firestore:"dailySpotLimitOverride,omitempty"`
	DailySpotLimitOverrideExpiresAt *time.Time `json:"-" firestore:"dailySpotLimitOverrideExpiresAt,omitempty"`
}

// ActiveDailySpotLimitOverride returns the user's temporary daily spot-creation
// limit override, if one is currently in effect (set and not yet expired).
// "Active" is computed here rather than stored as a boolean to avoid needing a
// background job to flip it off when it expires.
func (u *User) ActiveDailySpotLimitOverride(now time.Time) *int {
	if u.DailySpotLimitOverride == nil {
		return nil
	}
	if u.DailySpotLimitOverrideExpiresAt != nil && now.After(*u.DailySpotLimitOverrideExpiresAt) {
		return nil
	}
	return u.DailySpotLimitOverride
}

type UserPublic struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	PhotoURL *string    `json:"photoUrl"`
	Stats    *UserStats `json:"stats,omitempty"`
}

type UserStats struct {
	SpotsCreated  int     `json:"spotsCreated"`
	CommentsCount int     `json:"commentsCount"`
	RatingsCount  int     `json:"ratingsCount"`
	AverageRating float64 `json:"averageRating"`
}

type UserWithStats struct {
	User
	Stats UserStats `json:"stats"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	PhotoURL *string `json:"photoUrl"`
}

type Favorite struct {
	ID        string    `json:"id" firestore:"-"`
	UserID    string    `json:"userId" firestore:"userId"`
	SpotID    string    `json:"spotId" firestore:"spotId"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}

type FavoriteSpot struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Region  string    `json:"region"`
	Rating  float64   `json:"rating"`
	AddedAt time.Time `json:"addedAt"`
}
