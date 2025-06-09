FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY services/appointment-service ./services/appointment-service
WORKDIR /workspace/services/appointment-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /workspace/bin/appointment-service ./cmd

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /workspace/bin/appointment-service .
COPY --from=builder /workspace/services/appointment-service/config ./config
# Обновляем конфиг для контейнерной среды
RUN sed -i 's/localhost:8085/0.0.0.0:8085/' config/local.yaml
RUN sed -i 's/localhost:5432/db:5432/' config/local.yaml
EXPOSE 8085
ENV APPOINTMENT_SERVICE_CONFIG_PATH=/app/config/local.yaml
ENTRYPOINT ["./appointment-service"]