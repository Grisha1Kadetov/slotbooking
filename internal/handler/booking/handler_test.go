package booking_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handlerbooking "github.com/Grisha1Kadetov/slotbooking/internal/handler/booking"
	authmodel "github.com/Grisha1Kadetov/slotbooking/internal/model/auth"
	bookingmodel "github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type bookingServiceMock struct {
	createFunc      func(ctx context.Context, b bookingmodel.Booking) (bookingmodel.Booking, error)
	getListAllFunc  func(ctx context.Context, page, pageSize int) ([]bookingmodel.Booking, int, error)
	getByUserIdFunc func(ctx context.Context, userID uuid.UUID) ([]bookingmodel.Booking, error)
	cancelByIdFunc  func(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (bookingmodel.Booking, error)
}

func (m *bookingServiceMock) Create(ctx context.Context, b bookingmodel.Booking) (bookingmodel.Booking, error) {
	if m.createFunc == nil {
		return bookingmodel.Booking{}, errors.New("unexpected Create call")
	}
	return m.createFunc(ctx, b)
}

func (m *bookingServiceMock) GetListAll(ctx context.Context, page, pageSize int) ([]bookingmodel.Booking, int, error) {
	if m.getListAllFunc == nil {
		return nil, 0, errors.New("unexpected GetListAll call")
	}
	return m.getListAllFunc(ctx, page, pageSize)
}

func (m *bookingServiceMock) GetByUserId(ctx context.Context, userID uuid.UUID) ([]bookingmodel.Booking, error) {
	if m.getByUserIdFunc == nil {
		return nil, errors.New("unexpected GetByUserId call")
	}
	return m.getByUserIdFunc(ctx, userID)
}

func (m *bookingServiceMock) CancelById(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (bookingmodel.Booking, error) {
	if m.cancelByIdFunc == nil {
		return bookingmodel.Booking{}, errors.New("unexpected CancelById call")
	}
	return m.cancelByIdFunc(ctx, bookingID, userID)
}

func requestWithActor(method, path string, body []byte, actor authmodel.Actor) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	ctx := authmodel.WithActor(req.Context(), actor)
	return req.WithContext(ctx)
}

func requestWithURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestBookingHandler_Create(t *testing.T) {
	actor := authmodel.Actor{UserId: uuid.New()}

	t.Run("unauthorized", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := httptest.NewRequest(http.MethodPost, "/bookings/create", nil)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := requestWithActor(http.MethodPost, "/bookings/create", []byte("{"), actor)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("slot not found", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{
			createFunc: func(ctx context.Context, b bookingmodel.Booking) (bookingmodel.Booking, error) {
				return bookingmodel.Booking{}, bookingmodel.ErrSlotNotFound
			},
		}, nil)
		body := []byte(`{"slotId":"` + uuid.New().String() + `"}`)
		req := requestWithActor(http.MethodPost, "/bookings/create", body, actor)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("slot already booked", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{
			createFunc: func(ctx context.Context, b bookingmodel.Booking) (bookingmodel.Booking, error) {
				return bookingmodel.Booking{}, bookingmodel.ErrSlotAlreadyBooked
			},
		}, nil)
		body := []byte(`{"slotId":"` + uuid.New().String() + `"}`)
		req := requestWithActor(http.MethodPost, "/bookings/create", body, actor)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
		}
	})

	t.Run("old booking", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{
			createFunc: func(ctx context.Context, b bookingmodel.Booking) (bookingmodel.Booking, error) {
				return bookingmodel.Booking{}, bookingmodel.ErrOldBooking
			},
		}, nil)
		body := []byte(`{"slotId":"` + uuid.New().String() + `"}`)
		req := requestWithActor(http.MethodPost, "/bookings/create", body, actor)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		slotID := uuid.New()
		bookingID := uuid.New()
		createdAt := time.Now().UTC()
		h := handlerbooking.New(&bookingServiceMock{
			createFunc: func(ctx context.Context, b bookingmodel.Booking) (bookingmodel.Booking, error) {
				if b.SlotId != slotID {
					t.Fatalf("slotId = %v, want %v", b.SlotId, slotID)
				}
				if b.UserId != actor.UserId {
					t.Fatalf("userId = %v, want %v", b.UserId, actor.UserId)
				}
				if b.Status != bookingmodel.StatusActive {
					t.Fatalf("status = %q, want %q", b.Status, bookingmodel.StatusActive)
				}
				return bookingmodel.Booking{
					Id:        bookingID,
					SlotId:    b.SlotId,
					UserId:    b.UserId,
					Status:    b.Status,
					CreatedAt: &createdAt,
				}, nil
			},
		}, nil)
		body := []byte(`{"slotId":"` + slotID.String() + `"}`)
		req := requestWithActor(http.MethodPost, "/bookings/create", body, actor)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
		}

		var resp handlerbooking.CreateResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Booking.ID != bookingID {
			t.Fatalf("booking id = %v, want %v", resp.Booking.ID, bookingID)
		}
		if resp.Booking.SlotID != slotID {
			t.Fatalf("slot id = %v, want %v", resp.Booking.SlotID, slotID)
		}
	})
}

func TestBookingHandler_GetListAll(t *testing.T) {
	t.Run("invalid page", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=abc", nil)
		rr := httptest.NewRecorder()

		h.GetListAll(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid page size", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := httptest.NewRequest(http.MethodGet, "/bookings/list?pageSize=0", nil)
		rr := httptest.NewRecorder()

		h.GetListAll(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("too large page size", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := httptest.NewRequest(http.MethodGet, "/bookings/list?pageSize=101", nil)
		rr := httptest.NewRecorder()

		h.GetListAll(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		createdAt := time.Now().UTC()
		expected := []bookingmodel.Booking{{
			Id:        uuid.New(),
			SlotId:    uuid.New(),
			UserId:    uuid.New(),
			Status:    bookingmodel.StatusActive,
			CreatedAt: &createdAt,
		}}
		h := handlerbooking.New(&bookingServiceMock{
			getListAllFunc: func(ctx context.Context, page, pageSize int) ([]bookingmodel.Booking, int, error) {
				if page != 2 {
					t.Fatalf("page = %d, want %d", page, 2)
				}
				if pageSize != 10 {
					t.Fatalf("pageSize = %d, want %d", pageSize, 10)
				}
				return expected, 1, nil
			},
		}, nil)
		req := httptest.NewRequest(http.MethodGet, "/bookings/list?page=2&pageSize=10", nil)
		rr := httptest.NewRecorder()

		h.GetListAll(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		var resp handlerbooking.ListResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp.Bookings) != 1 {
			t.Fatalf("bookings len = %d, want %d", len(resp.Bookings), 1)
		}
		if resp.Pagination.Page != 2 || resp.Pagination.PageSize != 10 || resp.Pagination.Total != 1 {
			t.Fatalf("pagination = %+v, want page=2 pageSize=10 total=1", resp.Pagination)
		}
	})
}

func TestBookingHandler_GetMy(t *testing.T) {
	actor := authmodel.Actor{UserId: uuid.New()}

	t.Run("unauthorized", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
		rr := httptest.NewRecorder()

		h.GetMy(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("success", func(t *testing.T) {
		createdAt := time.Now().UTC()
		expected := []bookingmodel.Booking{{
			Id:        uuid.New(),
			SlotId:    uuid.New(),
			UserId:    actor.UserId,
			Status:    bookingmodel.StatusActive,
			CreatedAt: &createdAt,
		}}
		h := handlerbooking.New(&bookingServiceMock{
			getByUserIdFunc: func(ctx context.Context, userID uuid.UUID) ([]bookingmodel.Booking, error) {
				if userID != actor.UserId {
					t.Fatalf("userID = %v, want %v", userID, actor.UserId)
				}
				return expected, nil
			},
		}, nil)
		req := requestWithActor(http.MethodGet, "/bookings/my", nil, actor)
		rr := httptest.NewRecorder()

		h.GetMy(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		var resp handlerbooking.MyListResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp.Bookings) != 1 {
			t.Fatalf("bookings len = %d, want %d", len(resp.Bookings), 1)
		}
	})
}

func TestBookingHandler_Cancel(t *testing.T) {
	actor := authmodel.Actor{UserId: uuid.New()}

	t.Run("unauthorized", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := httptest.NewRequest(http.MethodPost, "/bookings/id/cancel", nil)
		rr := httptest.NewRecorder()

		h.Cancel(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid booking id", func(t *testing.T) {
		h := handlerbooking.New(&bookingServiceMock{}, nil)
		req := requestWithActor(http.MethodPost, "/bookings/bad/cancel", nil, actor)
		req = requestWithURLParam(req, "bookingId", "bad-uuid")
		rr := httptest.NewRecorder()

		h.Cancel(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("booking not found", func(t *testing.T) {
		bookingID := uuid.New()
		h := handlerbooking.New(&bookingServiceMock{
			cancelByIdFunc: func(ctx context.Context, gotBookingID uuid.UUID, gotUserID uuid.UUID) (bookingmodel.Booking, error) {
				if gotBookingID != bookingID {
					t.Fatalf("bookingID = %v, want %v", gotBookingID, bookingID)
				}
				if gotUserID != actor.UserId {
					t.Fatalf("userID = %v, want %v", gotUserID, actor.UserId)
				}
				return bookingmodel.Booking{}, bookingmodel.ErrNotFound
			},
		}, nil)
		req := requestWithActor(http.MethodPost, "/bookings/cancel", nil, actor)
		req = requestWithURLParam(req, "bookingId", bookingID.String())
		rr := httptest.NewRecorder()

		h.Cancel(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("not owner", func(t *testing.T) {
		bookingID := uuid.New()
		h := handlerbooking.New(&bookingServiceMock{
			cancelByIdFunc: func(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (bookingmodel.Booking, error) {
				return bookingmodel.Booking{}, bookingmodel.ErrNotOwner
			},
		}, nil)
		req := requestWithActor(http.MethodPost, "/bookings/cancel", nil, actor)
		req = requestWithURLParam(req, "bookingId", bookingID.String())
		rr := httptest.NewRecorder()

		h.Cancel(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
		}
	})

	t.Run("success", func(t *testing.T) {
		bookingID := uuid.New()
		createdAt := time.Now().UTC()
		h := handlerbooking.New(&bookingServiceMock{
			cancelByIdFunc: func(ctx context.Context, gotBookingID uuid.UUID, gotUserID uuid.UUID) (bookingmodel.Booking, error) {
				if gotBookingID != bookingID {
					t.Fatalf("bookingID = %v, want %v", gotBookingID, bookingID)
				}
				if gotUserID != actor.UserId {
					t.Fatalf("userID = %v, want %v", gotUserID, actor.UserId)
				}
				return bookingmodel.Booking{
					Id:        bookingID,
					SlotId:    uuid.New(),
					UserId:    actor.UserId,
					Status:    bookingmodel.StatusCancelled,
					CreatedAt: &createdAt,
				}, nil
			},
		}, nil)
		req := requestWithActor(http.MethodPost, "/bookings/cancel", nil, actor)
		req = requestWithURLParam(req, "bookingId", bookingID.String())
		rr := httptest.NewRecorder()

		h.Cancel(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		var resp handlerbooking.CancelResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Booking.ID != bookingID {
			t.Fatalf("booking id = %v, want %v", resp.Booking.ID, bookingID)
		}
		if resp.Booking.Status != bookingmodel.StatusCancelled {
			t.Fatalf("status = %q, want %q", resp.Booking.Status, bookingmodel.StatusCancelled)
		}
	})
}
