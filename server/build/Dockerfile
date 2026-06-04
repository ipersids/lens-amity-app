# build stage
FROM go:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrator

# migration stage
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrate .

CMD ["./api"]
