package domain

import "time"

type LoginRequest struct {
	IDToken string `json:"idToken" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
}

type PasswordLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
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

