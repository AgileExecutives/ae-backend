#!/bin/bash

# Wait for database to be ready
set -e

host="$1"
port="$2"
user="$3"
shift 3
cmd="$@"

echo "Waiting for PostgreSQL at $host:$port..."

until pg_isready -h "$host" -p "$port" -U "$user"; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - executing command"
exec $cmd