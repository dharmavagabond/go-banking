#!/bin/sh

DB_DSN="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@postgres-db:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable" task migrate:up

exec "$@"
