# Указываем базовый образ для этапа сборки
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/api/main.go

# Указываем базовый образ для финального этапа
FROM alpine:latest

# Порт будет передан из docker-compose.yaml через переменную окружения
ENV PORT=${PORT:-3000}

# Устанавливаем glibc
RUN apk --no-cache add libc6-compat

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE ${PORT}

CMD ["./main"]