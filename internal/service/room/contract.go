package room

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/room"
)

type roomRepository interface {
	Create(ctx context.Context, room room.Room) (room.Room, error)
	GetAll(ctx context.Context) ([]room.Room, error)
}
