-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    room_id UUID NOT NULL REFERENCES rooms (id),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT slots_time_range_check CHECK (start_at < end_at),
    CONSTRAINT slots_no_overlap_per_room EXCLUDE USING GIST (
        room_id WITH =,
        tstzrange (start_at, end_at, '[)') WITH &&
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS slots;
-- +goose StatementEnd
