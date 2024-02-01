FROM golang:1.21-alpine AS builder

WORKDIR /app
ENV DATABASE_URL=""

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./weather-api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/weather-api ./weather-api
COPY views ./views

EXPOSE 8080

CMD ["./weather-api"]
