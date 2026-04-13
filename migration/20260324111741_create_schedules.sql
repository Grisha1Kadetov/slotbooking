-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    room_id UUID NOT NULL UNIQUE REFERENCES rooms (id),
    days_of_week INT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    CHECK ( days_of_week <@ ARRAY[1, 2, 3, 4, 5, 6, 7] ),
    CHECK (start_time < end_time)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schedules;
-- +goose StatementEnd
