package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type StandardMessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

type User struct {
	ID           string    `json:"id"`
	NIK          string    `json:"nik"`
	Nama         string    `json:"nama"`
	Email        *string   `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID    string  `json:"id"`
	NIK   string  `json:"nik"`
	Nama  string  `json:"nama"`
	Email *string `json:"email,omitempty"`
	Role  string  `json:"role"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:    u.ID,
		NIK:   u.NIK,
		Nama:  u.Nama,
		Email: u.Email,
		Role:  u.Role,
	}
}

type LoginRequest struct {
	NIK      string `json:"nik"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type AuthClaims struct {
	UserID string `json:"user_id"`
	NIK    string `json:"nik"`
	Nama   string `json:"nama"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
