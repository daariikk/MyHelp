FROM golang:1.23-alpine AS builder
WORKDIR /workspace
# Копируем корневые go.mod и go.sum из контекста сборки (backend)
COPY go.mod go.sum ./
RUN go mod download
# Копируем исходники сервиса
COPY services/account-service ./services/account-service
WORKDIR /workspace/services/account-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /workspace/bin/account-service ./cmd

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /workspace/bin/account-service .
COPY --from=builder /workspace/services/account-service/config ./config
# Обновляем конфиг для контейнерной среды
RUN sed -i 's/localhost:8083/0.0.0.0:8083/' config/local.yaml
RUN sed -i 's/localhost:5432/db:5432/' config/local.yaml
EXPOSE 8083
ENV ACCOUNT_SERVICE_CONFIG_PATH=/app/config/local.yaml
ENTRYPOINT ["./account-service"]