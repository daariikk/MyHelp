FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY services/polyclinic-service ./services/polyclinic-service
WORKDIR /workspace/services/polyclinic-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /workspace/bin/polyclinic-service ./cmd

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /workspace/bin/polyclinic-service .
COPY --from=builder /workspace/services/polyclinic-service/config ./config
# Обновляем конфиг для контейнерной среды
RUN sed -i 's/localhost:8084/0.0.0.0:8084/' config/local.yaml
RUN sed -i 's/localhost:5432/db:5432/' config/local.yaml
EXPOSE 8084
ENV POLYCLINIC_SERVICE_CONFIG_PATH=/app/config/local.yaml
ENTRYPOINT ["./polyclinic-service"]