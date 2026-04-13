#!/bin/sh
set -e

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

/go/bin/goose -dir /app/migration up