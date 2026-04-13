package slot

import (
	"context"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	"github.com/google/uuid"
)

type slotService interface {
	GetAvailableByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slot.Slot, error)
}
