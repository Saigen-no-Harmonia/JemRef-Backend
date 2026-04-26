# ===== Build Stage =====
FROM golang:1.25-alpine AS builder
WORKDIR /app

# キャッシュを効かせるため依存関係だけ先に処理
COPY go.mod go.sum ./
RUN go mod download

# CGOを無効化にし、バイナリ作成
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/server


# ===== Runtime Stage =====
FROM alpine:3.19

# appのみコピー（軽量化のため）
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app/ .

EXPOSE 8080

CMD ["./app"]