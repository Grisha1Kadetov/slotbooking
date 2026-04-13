package room

import (
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/room"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/utctime"
	"github.com/google/uuid"
)

type RoomResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Capacity    *int       `json:"capacity"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type RoomCreateResponse struct {
	Room RoomResponse `json:"room"`
}

type RoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

type RoomCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

func (r RoomCreateRequest) ToModel() room.Room {
	return room.Room{
		ID:          uuid.Nil,
		Name:        r.Name,
		Description: &r.Description,
		Capacity:    &r.Capacity,
	}
}

func ToRoomResponse(r room.Room) RoomResponse {
	return RoomResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Capacity:    r.Capacity,
		CreatedAt:   utctime.TimePointerToUTC(r.CreatedAt),
	}
}

func ToRoomsResponse(rooms []room.Room) RoomsResponse {
	res := make([]RoomResponse, len(rooms))
	for i, r := range rooms {
		res[i] = ToRoomResponse(r)
	}

	return RoomsResponse{Rooms: res}
}
