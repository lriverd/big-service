package domain

import "time"

type LoginRequest struct {
	IDToken string `json:"idToken" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
}

type AuthResponse struct {
	User         AuthUser `json:"user"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
}

type AuthUser struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	PhotoURL  *string   `json:"photoUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type GoogleClaims struct {
	Email   string
	Name    string
	Picture string
}

