package room

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/room"
)

type RoomService struct {
	repo roomRepository
}

func New(repo roomRepository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) CreateRoom(ctx context.Context, r room.Room) (room.Room, error) {
	return s.repo.Create(ctx, r)
}

func (s *RoomService) GetAllRooms(ctx context.Context) ([]room.Room, error) {
	return s.repo.GetAll(ctx)
}
