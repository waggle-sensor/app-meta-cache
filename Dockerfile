FROM golang:1.19 as builder
WORKDIR $GOPATH/src/github.com/waggle-sensor/app-meta-cache
ARG TARGETARCH
COPY . .
RUN CGO_ENABLED=0 go build -o update-app-cache cmd/update_app_cache/main.go \
  && mkdir -p /app \
  && cp update-app-cache /app

FROM redis:7.0.4
COPY --from=builder /app/update-app-cache /