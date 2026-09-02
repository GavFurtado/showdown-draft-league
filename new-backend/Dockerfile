FROM golang:1.23-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS dev
RUN go install github.com/githubnemo/CompileDaemon@latest
ENV PATH=$PATH:/go/bin
ENV TZ=Asia/Kolkata
EXPOSE 8080
CMD CompileDaemon \
    --build="go build -o /app/server /app/cmd/main.go" \
    --command="/app/server" \
    --recursive=true \
    --polling

FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server /app/cmd/main.go

FROM alpine:3.19 AS prod
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
