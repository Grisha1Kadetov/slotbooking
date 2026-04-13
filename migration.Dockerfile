FROM golang:1.25

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY migration /app/migration
COPY migration-entrypoint.sh /app/migration-entrypoint.sh

RUN chmod +x /app/migration-entrypoint.sh

ENTRYPOINT ["/app/migration-entrypoint.sh"]