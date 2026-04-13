package booking

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/auth"
	bookingModel "github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
	errrend "github.com/Grisha1Kadetov/slotbooking/internal/pkg/errorrenderer"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type BookingHandler struct {
	service bookingService
	l       log.Logger
}

func New(s bookingService, l log.Logger) *BookingHandler {
	return &BookingHandler{
		service: s,
		l:       l,
	}
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.ActorFromContext(r.Context())
	if !ok {
		errrend.RenderError(w, r, http.StatusUnauthorized, string(errrend.StatusUnauthorized), "unauthorized")
		return
	}

	var request CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errrend.RenderBadRequestError(w, r, "invalid JSON")
		return
	}

	b := bookingModel.Booking{
		SlotId: request.SlotID,
		UserId: actor.UserId,
		Status: bookingModel.StatusActive,
	}

	createdBooking, err := h.service.Create(r.Context(), b)
	if err != nil {
		if errors.Is(err, bookingModel.ErrSlotNotFound) {
			errrend.RenderError(w, r, http.StatusNotFound, string(errrend.StatusSlotNotFound), "slot not found")
			return
		}
		if errors.Is(err, bookingModel.ErrSlotAlreadyBooked) {
			errrend.RenderError(w, r, http.StatusConflict, string(errrend.StatusSlotAlreadyBooked), "slot is already booked")
			return
		}
		if errors.Is(err, bookingModel.ErrOldBooking) {
			errrend.RenderBadRequestError(w, r, "slot is in the past")
			return
		}

		errrend.RenderInternalError(w, r)
		h.l.Error("failed to create booking", log.Err(err))
		return
	}

	response := ToCreateResponse(createdBooking)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func (h *BookingHandler) GetListAll(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20

	pageRaw := r.URL.Query().Get("page")
	if pageRaw != "" {
		parsedPage, err := strconv.Atoi(pageRaw)
		if err != nil || parsedPage < 1 {
			errrend.RenderBadRequestError(w, r, "page must be a positive integer")
			return
		}
		page = parsedPage
	}

	pageSizeRaw := r.URL.Query().Get("pageSize")
	if pageSizeRaw != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil || parsedPageSize < 1 {
			errrend.RenderBadRequestError(w, r, "pageSize must be a positive integer")
			return
		}
		if parsedPageSize > 100 {
			errrend.RenderBadRequestError(w, r, "pageSize must be less or equal 100")
			return
		}
		pageSize = parsedPageSize
	}

	bookings, total, err := h.service.GetListAll(r.Context(), page, pageSize)
	if err != nil {
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to get bookings list", log.Err(err))
		return
	}

	response := ToListResponse(bookings, page, pageSize, total)
	render.JSON(w, r, response)
}

func (h *BookingHandler) GetMy(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.ActorFromContext(r.Context())
	if !ok {
		errrend.RenderError(w, r, http.StatusUnauthorized, string(errrend.StatusUnauthorized), "unauthorized")
		return
	}

	bookings, err := h.service.GetByUserId(r.Context(), actor.UserId)
	if err != nil {
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to get user bookings", log.Err(err))
		return
	}

	response := ToMyListResponse(bookings)
	render.JSON(w, r, response)
}

func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.ActorFromContext(r.Context())
	if !ok {
		errrend.RenderError(w, r, http.StatusUnauthorized, string(errrend.StatusUnauthorized), "unauthorized")
		return
	}

	bookingIdRaw := chi.URLParam(r, "bookingId")

	bookingId, err := uuid.Parse(bookingIdRaw)
	if err != nil {
		errrend.RenderBadRequestError(w, r, "invalid bookingId")
		return
	}

	cancelledBooking, err := h.service.CancelById(r.Context(), bookingId, actor.UserId)
	if err != nil {
		if errors.Is(err, bookingModel.ErrNotFound) {
			errrend.RenderError(w, r, http.StatusNotFound, string(errrend.StatusBookingNotFound), "booking not found")
			return
		}
		if errors.Is(err, bookingModel.ErrNotOwner) {
			errrend.RenderError(w, r, http.StatusForbidden, string(errrend.StatusForbidden), "cannot cancel another user's booking")
			return
		}

		errrend.RenderInternalError(w, r)
		h.l.Error("failed to cancel booking", log.Err(err))
		return
	}

	response := ToCancelResponse(cancelledBooking)
	render.JSON(w, r, response)
}
