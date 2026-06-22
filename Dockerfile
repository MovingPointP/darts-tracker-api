# ビルドステージ
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 依存関係のダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# OpenAPIドキュメントの自動生成
RUN go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/server/main.go

# ビルド
# 実行ファイル: server を生成
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# 実行ステージ
FROM alpine:latest

# 証明書検証パッケージのインストール
RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./server"]
