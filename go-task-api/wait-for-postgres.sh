#!/bin/sh
until nc -z postgres 5432; do
  echo "Waiting for postgres..."
  sleep 1
done
echo "Postgres is up!"
exec "$@"
