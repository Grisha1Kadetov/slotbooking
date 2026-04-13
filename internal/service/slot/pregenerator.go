package slot

import (
	"context"
	"time"
)

type Pregenerator struct {
	roomRepo roomLister
	slots    pregenerateService
	now      func() time.Time
	tick     time.Duration
}

func NewPregenerator(roomRepo roomLister, slots pregenerateService, now func() time.Time) *Pregenerator {
	return &Pregenerator{
		roomRepo: roomRepo,
		slots:    slots,
		now:      now,
		tick:     time.Minute * 10,
	}
}

func (p *Pregenerator) Run(ctx context.Context) error {
	lastDay := p.now().UTC().Day()
	ticker := time.NewTicker(p.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			now := p.now().UTC()
			if now.Day() == lastDay {
				continue
			}
			lastDay = now.Day()

			rooms, err := p.roomRepo.GetAll(ctx)
			if err != nil {
				return err
			}

			for _, r := range rooms {
				if err := p.slots.PreGenerateSlotsByRoomIdWithDuration(ctx, now, 0, r.ID); err != nil {
					return err
				}
			}
		}
	}
}
