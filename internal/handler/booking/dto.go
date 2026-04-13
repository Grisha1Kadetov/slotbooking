package booking

import (
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/utctime"
	"github.com/google/uuid"
)

type CreateRequest struct {
	SlotID               uuid.UUID `json:"slotId"`
	CreateConferenceLink bool      `json:"createConferenceLink"`
}

type BookingResponse struct {
	ID             uuid.UUID      `json:"id"`
	SlotID         uuid.UUID      `json:"slotId"`
	UserID         uuid.UUID      `json:"userId"`
	Status         booking.Status `json:"status"`
	ConferenceLink *string        `json:"conferenceLink"`
	CreatedAt      *time.Time     `json:"createdAt"`
}

type CreateResponse struct {
	Booking BookingResponse `json:"booking"`
}

type CancelResponse struct {
	Booking BookingResponse `json:"booking"`
}

type ListResponse struct {
	Bookings   []BookingResponse `json:"bookings"`
	Pagination Pagination        `json:"pagination"`
}

type MyListResponse struct {
	Bookings []BookingResponse `json:"bookings"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

func ToBookingResponse(b booking.Booking) BookingResponse {
	return BookingResponse{
		ID:             b.Id,
		SlotID:         b.SlotId,
		UserID:         b.UserId,
		Status:         b.Status,
		ConferenceLink: b.ConferenceLink,
		CreatedAt:      utctime.TimePointerToUTC(b.CreatedAt),
	}
}

func ToCreateResponse(b booking.Booking) CreateResponse {
	return CreateResponse{
		Booking: ToBookingResponse(b),
	}
}

func ToCancelResponse(b booking.Booking) CancelResponse {
	return CancelResponse{
		Booking: ToBookingResponse(b),
	}
}

func ToListResponse(bookings []booking.Booking, page, pageSize, total int) ListResponse {
	responseBookings := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		responseBookings = append(responseBookings, ToBookingResponse(b))
	}

	return ListResponse{
		Bookings: responseBookings,
		Pagination: Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}
}

func ToMyListResponse(bookings []booking.Booking) MyListResponse {
	responseBookings := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		responseBookings = append(responseBookings, ToBookingResponse(b))
	}

	return MyListResponse{
		Bookings: responseBookings,
	}
}
