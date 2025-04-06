﻿FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o otech-server ./cmd/server

FROM alpine:3.15

WORKDIR /app

COPY --from=builder /app/otech-server .

COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./otech-server"]