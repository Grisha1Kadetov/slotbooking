package slot

import (
	"errors"
	"net/http"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	errrend "github.com/Grisha1Kadetov/slotbooking/internal/pkg/errorrenderer"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

type SlotHandler struct {
	service slotService
	l       log.Logger
}

func New(s slotService, l log.Logger) *SlotHandler {
	return &SlotHandler{
		service: s,
		l:       l,
	}
}

func (h *SlotHandler) GetAvailableList(w http.ResponseWriter, r *http.Request) {
	roomIDRaw := chi.URLParam(r, "roomId")
	if roomIDRaw == "" {
		errrend.RenderBadRequestError(w, r, "roomId is required")
		return
	}

	roomID, err := uuid.Parse(roomIDRaw)
	if err != nil {
		errrend.RenderBadRequestError(w, r, "invalid roomId")
		return
	}

	dateRaw := r.URL.Query().Get("date")
	if dateRaw == "" {
		errrend.RenderBadRequestError(w, r, "date is required")
		return
	}

	date, err := time.Parse(dateLayout, dateRaw)
	if err != nil {
		errrend.RenderBadRequestError(w, r, "date must be in YYYY-MM-DD format")
		return
	}

	slots, err := h.service.GetAvailableByRoomIdAndDate(r.Context(), roomID, date)
	if err != nil {
		if errors.Is(err, slot.ErrRoomNotFound) {
			errrend.RenderError(w, r, http.StatusNotFound, string(errrend.StatusRoomNotFound), "room not found")
			return
		}
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to get available slots", log.Err(err))
		return
	}

	response := ToListAvailableResponse(slots)
	render.JSON(w, r, response)
}
