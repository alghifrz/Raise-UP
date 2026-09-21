# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api \
	&& CGO_ENABLED=0 GOOS=linux go build -o /bin/seed ./cmd/seed

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 1000 appuser

COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/seed /app/seed

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/api"]
