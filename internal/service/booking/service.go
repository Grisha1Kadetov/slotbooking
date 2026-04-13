package booking

import (
	"context"
	"errors"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
	slotModel "github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	"github.com/google/uuid"
)

type BookingService struct {
	repo         repository
	slotProvider slotProvider
}

func New(repo repository, slotProvider slotProvider) *BookingService {
	return &BookingService{
		repo:         repo,
		slotProvider: slotProvider,
	}
}

func (s *BookingService) Create(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	slot, err := s.slotProvider.GetById(ctx, b.SlotId)
	if err != nil {
		if errors.Is(err, slotModel.ErrNotFound) {
			return booking.Booking{}, booking.ErrSlotNotFound
		}
		return booking.Booking{}, err
	}
	if slot.StartAt.UTC().Before(time.Now().UTC()) {
		return booking.Booking{}, booking.ErrOldBooking
	}
	return s.repo.Create(ctx, b)
}

func (s *BookingService) GetListAll(ctx context.Context, page, pageSize int) ([]booking.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return s.repo.GetListAll(ctx, page, pageSize)
}

func (s *BookingService) GetByUserId(ctx context.Context, userID uuid.UUID) ([]booking.Booking, error) {
	return s.repo.GetByUserId(ctx, userID)
}

func (s *BookingService) CancelById(ctx context.Context, bookingID uuid.UUID, userId uuid.UUID) (booking.Booking, error) {
	book, err := s.repo.GetById(ctx, bookingID)
	if err != nil {
		return booking.Booking{}, err
	}
	if book.UserId != userId {
		return booking.Booking{}, booking.ErrNotOwner
	}
	return s.repo.CancelById(ctx, bookingID)
}
