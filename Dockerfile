FROM golang:1.19 as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o update-app-cache cmd/update_app_cache/main.go

FROM redis:7.0.4
COPY --from=builder /build/update-app-cache /
