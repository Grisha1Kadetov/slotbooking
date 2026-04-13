package auth

import (
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/utctime"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     user.Role `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DummyLoginRequest struct {
	Role user.Role `json:"role"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      user.Role  `json:"role"`
	CreatedAt *time.Time `json:"createdAt"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

func ToUserResponse(u user.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: utctime.TimePointerToUTC(u.CreatedAt),
	}
}

func ToRegisterResponse(u user.User) RegisterResponse {
	return RegisterResponse{
		User: ToUserResponse(u),
	}
}

func ToTokenResponse(token string) TokenResponse {
	return TokenResponse{
		Token: token,
	}
}
