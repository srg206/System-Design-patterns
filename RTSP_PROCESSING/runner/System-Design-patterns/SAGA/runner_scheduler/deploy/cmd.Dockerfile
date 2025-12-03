FROM golang:1.24-alpine AS builder

ARG CMD_PATH=./cmd/scheduler/main.go

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main "${CMD_PATH}"

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]

