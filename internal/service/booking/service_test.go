package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	bookingservice "github.com/Grisha1Kadetov/slotbooking/internal/service/booking"
	"github.com/google/uuid"
)

type repositoryMock struct {
	createFunc      func(ctx context.Context, b booking.Booking) (booking.Booking, error)
	getListAllFunc  func(ctx context.Context, page, pageSize int) ([]booking.Booking, int, error)
	getByUserIdFunc func(ctx context.Context, userID uuid.UUID) ([]booking.Booking, error)
	getByIdFunc     func(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error)
	cancelByIdFunc  func(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error)
}

func (m *repositoryMock) Create(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	if m.createFunc == nil {
		return booking.Booking{}, errors.New("unexpected Create call")
	}
	return m.createFunc(ctx, b)
}

func (m *repositoryMock) GetListAll(ctx context.Context, page, pageSize int) ([]booking.Booking, int, error) {
	if m.getListAllFunc == nil {
		return nil, 0, errors.New("unexpected GetListAll call")
	}
	return m.getListAllFunc(ctx, page, pageSize)
}

func (m *repositoryMock) GetByUserId(ctx context.Context, userID uuid.UUID) ([]booking.Booking, error) {
	if m.getByUserIdFunc == nil {
		return nil, errors.New("unexpected GetByUserId call")
	}
	return m.getByUserIdFunc(ctx, userID)
}

func (m *repositoryMock) GetById(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error) {
	if m.getByIdFunc == nil {
		return booking.Booking{}, errors.New("unexpected GetById call")
	}
	return m.getByIdFunc(ctx, bookingID)
}

func (m *repositoryMock) CancelById(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error) {
	if m.cancelByIdFunc == nil {
		return booking.Booking{}, errors.New("unexpected CancelById call")
	}
	return m.cancelByIdFunc(ctx, bookingID)
}

type slotProviderMock struct {
	getByIdFunc func(ctx context.Context, slotID uuid.UUID) (slot.Slot, error)
}

func (m *slotProviderMock) GetById(ctx context.Context, slotID uuid.UUID) (slot.Slot, error) {
	if m.getByIdFunc == nil {
		return slot.Slot{}, errors.New("unexpected GetById call")
	}
	return m.getByIdFunc(ctx, slotID)
}

func TestBookingService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := uuid.New()
	slotID := uuid.New()

	tests := []struct {
		name    string
		repo    *repositoryMock
		slots   *slotProviderMock
		wantErr error
	}{
		{
			name: "success",
			repo: &repositoryMock{
				createFunc: func(ctx context.Context, b booking.Booking) (booking.Booking, error) {
					b.Id = uuid.New()
					return b, nil
				},
			},
			slots: &slotProviderMock{
				getByIdFunc: func(ctx context.Context, gotSlotID uuid.UUID) (slot.Slot, error) {
					if gotSlotID != slotID {
						t.Fatalf("GetById() slotID = %v, want %v", gotSlotID, slotID)
					}
					return slot.Slot{
						ID:      slotID,
						StartAt: time.Now().Add(time.Hour),
					}, nil
				},
			},
		},
		{
			name: "slot not found",
			repo: &repositoryMock{},
			slots: &slotProviderMock{
				getByIdFunc: func(ctx context.Context, slotID uuid.UUID) (slot.Slot, error) {
					return slot.Slot{}, slot.ErrNotFound
				},
			},
			wantErr: booking.ErrSlotNotFound,
		},
		{
			name: "old booking",
			repo: &repositoryMock{},
			slots: &slotProviderMock{
				getByIdFunc: func(ctx context.Context, slotID uuid.UUID) (slot.Slot, error) {
					return slot.Slot{
						ID:      slotID,
						StartAt: time.Now().Add(-time.Hour),
					}, nil
				},
			},
			wantErr: booking.ErrOldBooking,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := bookingservice.New(tt.repo, tt.slots)
			got, err := service.Create(ctx, booking.Booking{
				SlotId: slotID,
				UserId: userID,
				Status: booking.StatusActive,
			})

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if got.Id == uuid.Nil {
				t.Fatal("Create() returned empty booking ID")
			}
		})
	}
}

func TestBookingService_GetListAll_NormalizesPagination(t *testing.T) {
	t.Parallel()

	repo := &repositoryMock{
		getListAllFunc: func(ctx context.Context, page, pageSize int) ([]booking.Booking, int, error) {
			if page != 1 {
				t.Fatalf("page = %d, want 1", page)
			}
			if pageSize != 20 {
				t.Fatalf("pageSize = %d, want 20", pageSize)
			}
			return []booking.Booking{}, 0, nil
		},
	}

	service := bookingservice.New(repo, &slotProviderMock{})

	_, _, err := service.GetListAll(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("GetListAll() error = %v", err)
	}
}

func TestBookingService_GetByUserId(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	expected := []booking.Booking{{Id: uuid.New(), UserId: userID}}

	repo := &repositoryMock{
		getByUserIdFunc: func(ctx context.Context, gotUserID uuid.UUID) ([]booking.Booking, error) {
			if gotUserID != userID {
				t.Fatalf("GetByUserId() userID = %v, want %v", gotUserID, userID)
			}
			return expected, nil
		},
	}

	service := bookingservice.New(repo, &slotProviderMock{})

	got, err := service.GetByUserId(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetByUserId() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetByUserId() len = %d, want 1", len(got))
	}
	if got[0].UserId != userID {
		t.Fatalf("GetByUserId() userID = %v, want %v", got[0].UserId, userID)
	}
}

func TestBookingService_CancelById(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bookingID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	cancelledBooking := booking.Booking{
		Id:     bookingID,
		UserId: ownerID,
		Status: booking.StatusCancelled,
	}

	tests := []struct {
		name    string
		repo    *repositoryMock
		userID  uuid.UUID
		wantErr error
	}{
		{
			name: "success",
			repo: &repositoryMock{
				getByIdFunc: func(ctx context.Context, gotBookingID uuid.UUID) (booking.Booking, error) {
					if gotBookingID != bookingID {
						t.Fatalf("GetById() bookingID = %v, want %v", gotBookingID, bookingID)
					}
					return booking.Booking{
						Id:     bookingID,
						UserId: ownerID,
						Status: booking.StatusActive,
					}, nil
				},
				cancelByIdFunc: func(ctx context.Context, gotBookingID uuid.UUID) (booking.Booking, error) {
					if gotBookingID != bookingID {
						t.Fatalf("CancelById() bookingID = %v, want %v", gotBookingID, bookingID)
					}
					return cancelledBooking, nil
				},
			},
			userID: ownerID,
		},
		{
			name: "not owner",
			repo: &repositoryMock{
				getByIdFunc: func(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error) {
					return booking.Booking{
						Id:     bookingID,
						UserId: ownerID,
					}, nil
				},
			},
			userID:  otherUserID,
			wantErr: booking.ErrNotOwner,
		},
		{
			name: "booking not found",
			repo: &repositoryMock{
				getByIdFunc: func(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error) {
					return booking.Booking{}, booking.ErrNotFound
				},
			},
			userID:  ownerID,
			wantErr: booking.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := bookingservice.New(tt.repo, &slotProviderMock{})

			got, err := service.CancelById(ctx, bookingID, tt.userID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CancelById() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("CancelById() error = %v", err)
			}
			if got.Status != booking.StatusCancelled {
				t.Fatalf("CancelById() status = %q, want %q", got.Status, booking.StatusCancelled)
			}
		})
	}
}
