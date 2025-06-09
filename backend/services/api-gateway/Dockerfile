FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY services/api-gateway ./services/api-gateway
WORKDIR /workspace/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /workspace/bin/api-gateway ./cmd

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /workspace/bin/api-gateway .
COPY --from=builder /workspace/services/api-gateway/config ./config
# Обновляем конфиг для контейнерной среды
RUN sed -i 's/localhost:8082/0.0.0.0:8082/' config/local.yaml
RUN sed -i 's/localhost:5432/db:5432/' config/local.yaml
RUN sed -i 's|http://localhost:8085|http://appointment-service:8085|g' config/local.yaml
RUN sed -i 's|http://localhost:8083|http://account-service:8083|g' config/local.yaml
RUN sed -i 's|http://localhost:8084|http://polyclinic-service:8084|g' config/local.yaml
EXPOSE 8082
ENV API_GATEWAY_CONFIG_PATH=/app/config/local.yaml
ENTRYPOINT ["./api-gateway"]