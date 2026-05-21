package domain_test

import (
	"testing"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/comments/domain"
)

func TestComment(t *testing.T) {
	now := time.Now()
	c := domain.Comment{
		ID: "c1", SpotID: "s1", UserID: "u1", Text: "Great spot",
		Likes: 3, Liked: true, CreatedAt: now,
	}
	if c.Likes != 3 || !c.Liked {
		t.Error("unexpected comment fields")
	}
}

func TestCommentWithUser(t *testing.T) {
	photo := "http://photo.url"
	c := domain.Comment{
		ID: "c1", User: &domain.UserInfo{ID: "u1", Name: "Test", PhotoURL: &photo},
	}
	if c.User.Name != "Test" || *c.User.PhotoURL != photo {
		t.Error("unexpected user info")
	}
}

func TestCommentLike(t *testing.T) {
	cl := domain.CommentLike{CommentID: "c1", UserID: "u1", CreatedAt: time.Now()}
	if cl.CommentID != "c1" {
		t.Error("unexpected comment like fields")
	}
}

func TestCreateCommentRequest(t *testing.T) {
	req := domain.CreateCommentRequest{Text: "This is a test comment"}
	if req.Text != "This is a test comment" {
		t.Error("unexpected text")
	}
}

func TestUpdateCommentRequest(t *testing.T) {
	req := domain.UpdateCommentRequest{Text: "Updated comment text"}
	if req.Text != "Updated comment text" {
		t.Error("unexpected text")
	}
}

