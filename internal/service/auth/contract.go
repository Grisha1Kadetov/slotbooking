package auth

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
)

type userRepository interface {
	Create(ctx context.Context, u user.User) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
}
