package slot

import "errors"

var (
	ErrNotFound     = errors.New("slot not found")
	ErrRoomNotFound = errors.New("room not found")
	ErrTimeRange    = errors.New("invalid time range")
	ErrTimeOverlap  = errors.New("slot time overlaps with another slot")
)
