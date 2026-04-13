package room

import (
	"encoding/json"
	"net/http"

	errrend "github.com/Grisha1Kadetov/slotbooking/internal/pkg/errorrenderer"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/render"
)

type RoomHandler struct {
	service roomService
	l       log.Logger
}

func New(s roomService, l log.Logger) *RoomHandler {
	return &RoomHandler{service: s, l: l}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var request RoomCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errrend.RenderBadRequestError(w, r, "invalid JSON")
		return
	}
	if request.Name == "" {
		errrend.RenderBadRequestError(w, r, "name is required")
		return
	}
	requstModel := request.ToModel()
	room, err := h.service.CreateRoom(r.Context(), requstModel)
	if err != nil {
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to create room", log.Err(err))
		return
	}
	response := RoomCreateResponse{ToRoomResponse(room)}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func (h *RoomHandler) GetAllRooms(w http.ResponseWriter, r *http.Request) {
	room, err := h.service.GetAllRooms(r.Context())
	if err != nil {
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to get list of rooms", log.Err(err))
		return
	}
	response := ToRoomsResponse(room)
	render.JSON(w, r, response)
}
