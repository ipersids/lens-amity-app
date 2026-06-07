FROM golang:1.25-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod ./
COPY go.sum ./

RUN go mod download
