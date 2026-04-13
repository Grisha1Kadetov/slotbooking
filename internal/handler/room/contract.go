package room

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/room"
)

type roomService interface {
	CreateRoom(ctx context.Context, room room.Room) (room.Room, error)
	GetAllRooms(ctx context.Context) ([]room.Room, error)
}
