package booking

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	"github.com/google/uuid"
)

type repository interface {
	Create(ctx context.Context, b booking.Booking) (booking.Booking, error)
	GetListAll(ctx context.Context, page, pageSize int) ([]booking.Booking, int, error)
	GetByUserId(ctx context.Context, userID uuid.UUID) ([]booking.Booking, error)
	CancelById(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error)
	GetById(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error)
}

type slotProvider interface {
	GetById(ctx context.Context, slotId uuid.UUID) (slot.Slot, error)
}
