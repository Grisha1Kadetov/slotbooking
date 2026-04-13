package utctime

import "time"

func TimePointerToUTC(t *time.Time) *time.Time {
	v := *t
	v = v.UTC()
	return &v
}
