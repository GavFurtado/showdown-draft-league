# Use a Go base image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Install CompileDaemon for hot reloading
RUN go install github.com/githubnemo/CompileDaemon@latest

# Copy the go.mod and go.sum files to download dependencies

COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code, including the .env file
COPY . .

# Expose the port your Go application listens on (replace 8080 with your app's port)
EXPOSE 8080

# Command to run CompileDaemon
# It watches for changes, rebuilds the app, and runs the new executable

CMD CompileDaemon --build="go build -o /app/server /app/cmd/main.go" --command="/app/server" --recursive=true
