package domain

import "time"

type Comment struct {
	ID        string     `json:"id" firestore:"-"`
	SpotID    string     `json:"spotId" firestore:"spotId"`
	UserID    string     `json:"userId" firestore:"userId"`
	User      *UserInfo  `json:"user,omitempty" firestore:"-"`
	Text      string     `json:"text" firestore:"text"`
	Likes     int        `json:"likes" firestore:"likes"`
	Liked     bool       `json:"liked" firestore:"-"`
	CreatedAt time.Time  `json:"createdAt" firestore:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt" firestore:"updatedAt"`
	DeletedAt *time.Time `json:"-" firestore:"deletedAt"`
}

type UserInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	PhotoURL *string `json:"photoUrl"`
}

type CommentLike struct {
	ID        string    `json:"id" firestore:"-"`
	CommentID string    `json:"commentId" firestore:"commentId"`
	UserID    string    `json:"userId" firestore:"userId"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
}

type CreateCommentRequest struct {
	Text string `json:"text" binding:"required,min=1,max=500"`
}

type UpdateCommentRequest struct {
	Text string `json:"text" binding:"required,min=1,max=500"`
}

