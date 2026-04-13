package auth

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/google/uuid"
)

type Actor struct {
	UserId uuid.UUID
	Role   user.Role
}

type actorKey struct{}

func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, a)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(actorKey{}).(Actor)
	return a, ok
}
