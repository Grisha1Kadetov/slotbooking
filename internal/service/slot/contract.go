package slot

import (
	"context"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/room"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	"github.com/google/uuid"
)

type slotRepository interface {
	CreateBulk(ctx context.Context, slots []slot.Slot) ([]slot.Slot, error)
	GetById(ctx context.Context, slotId uuid.UUID) (slot.Slot, error)
	GetAvailableByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slot.Slot, error)
	HasAnyByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error)
}

type scheduleProvider interface {
	GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedule.Schedule, error)
}

type roomProvider interface {
	ExistsById(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type pregenerateService interface {
	PreGenerateSlotsByRoomIdWithDuration(ctx context.Context, dateFromInclude time.Time, nextDays int, roomId uuid.UUID) error
}

type roomLister interface {
	GetAll(ctx context.Context) ([]room.Room, error)
}
