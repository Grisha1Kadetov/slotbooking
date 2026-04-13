-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, password_hash, role)
VALUES (
        '11111111-1111-1111-1111-111111111111',
        'admin@test.com',
        'dummy',
        'admin'
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'user@test.com',
        'dummy',
        'user'
    )
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users
WHERE
    id IN (
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-2222-2222-222222222222'
    );
-- +goose StatementEnd
