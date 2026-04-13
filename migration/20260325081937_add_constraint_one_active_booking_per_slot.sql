-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS bookings_one_active_per_slot_idx 
ON bookings (slot_id)
WHERE status = 'active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS bookings_one_active_per_slot_idx;
-- +goose StatementEnd
