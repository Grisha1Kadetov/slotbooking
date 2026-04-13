package auth

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/google/uuid"
)

type service interface {
	ParseToken(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error)
}
