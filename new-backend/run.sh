#!/bin/bash

# Load environment variables from .env file if exists
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo "[INFO] No .env file found. Using default or system environment variables."
fi

export APP_BASE_URL="http://localhost:${PORT}"

# recompile on dev builds every time
if [ "$ENV" = "dev" ]; then
  go run ./cmd/main.go
else
  go build -o ./build/server .
  ./server
fi
