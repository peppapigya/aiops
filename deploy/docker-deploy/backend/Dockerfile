FROM docker.m.daocloud.io/library/golang:1.25-alpine AS builder

WORKDIR /build
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

RUN apk add --no-cache git curl ca-certificates \
    &&git clone -b master --depth 1  https://github.com/peppapigya/aiops.git \
    && cd  aiops/backend/ \
    && go mod tidy \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o aiops ./cmd/server

FROM docker.m.daocloud.io/library/alpine:3.20.2
ENV TZ=Asia/Shanghai
WORKDIR /app

COPY --from=builder /build/aiops/backend/aiops .
COPY --from=builder /build/aiops/backend/config/ ./config/

RUN chmod +x ./aiops &&\
    apk add --no-cache curl ca-certificates
EXPOSE 8081
HEALTHCHECK  --interval=30s --timeout=5s --start-period=5s --retries=3 CMD curl -f http://localhost:8081/health || exit 1

CMD ["./aiops"]