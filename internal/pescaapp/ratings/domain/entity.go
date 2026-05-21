package domain

import "time"

type Rating struct {
	ID        string    `json:"id" firestore:"-"`
	SpotID    string    `json:"spotId" firestore:"spotId"`
	UserID    string    `json:"userId" firestore:"userId"`
	User      *UserInfo `json:"user,omitempty" firestore:"-"`
	Stars     int       `json:"stars" firestore:"stars"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type UserInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	PhotoURL *string `json:"photoUrl"`
}

type RatingStats struct {
	AverageRating float64            `json:"averageRating"`
	TotalRatings  int                `json:"totalRatings"`
	Distribution  map[string]int     `json:"distribution"`
}

type CreateRatingRequest struct {
	Stars int `json:"stars" binding:"required,min=1,max=5"`
}

