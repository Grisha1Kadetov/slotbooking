package auth

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
)

type authService interface {
	Register(ctx context.Context, email, password string, role user.Role) (user.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GenerateDummyToken(role user.Role) (string, error)
}
