#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables from .env file if exists
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
  echo "run.sh: Loaded environment variables from .env"
else
  echo "run.sh: No .env file found. Using default or system environment variables."
fi

# recompile on dev builds every time
if [ "$ENV" = "dev" ]; then
  if ! command -v swag &> /dev/null; then
    echo "run.sh: WARNING: swag is not installed. API documentation may be outdated."
    echo "run.sh:   Install with: go install github.com/swaggo/swag/cmd/swag@v1.8.12"
  else
    echo "run.sh: Regenerating swagger docs..."
    swag init -g cmd/main.go -o api/
    echo
  fi
  go run ./cmd/main.go
else
  go build -o ./build/server .
  ./server
fi
