package slot

import (
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	"github.com/google/uuid"
)

type SlotResponse struct {
	ID     uuid.UUID `json:"id"`
	RoomID uuid.UUID `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type ListAvailableResponse struct {
	Slots []SlotResponse `json:"slots"`
}

func ToSlotResponse(s slot.Slot) SlotResponse {
	return SlotResponse{
		ID:     s.ID,
		RoomID: s.RoomID,
		Start:  s.StartAt,
		End:    s.EndAt,
	}
}

func ToListAvailableResponse(slots []slot.Slot) ListAvailableResponse {
	responseSlots := make([]SlotResponse, 0, len(slots))
	for _, s := range slots {
		responseSlots = append(responseSlots, ToSlotResponse(s))
	}

	return ListAvailableResponse{
		Slots: responseSlots,
	}
}
