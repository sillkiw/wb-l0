# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM golang:1.24.6  AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка бинаря consumer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/consumer ./cmd/consumer

############################
# Runtime stage
############################
FROM alpine:3.20



WORKDIR /app
COPY --from=build /out/consumer /app/consumer

ENV APP_ENV=docker

RUN adduser -D -u 10001 appuser
USER appuser

ENTRYPOINT ["/app/consumer"]
