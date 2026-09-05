# マルチステージビルド
# Stage 1: Go アプリケーションのビルド
FROM golang:1.21-bookworm AS builder

# 作業ディレクトリを設定
WORKDIR /app

# SQLiteドライバーのCGOビルドに必要です。
RUN apt-get update \
  && apt-get install --no-install-recommends -y build-essential \
  && rm -rf /var/lib/apt/lists/*

# go.mod, go.sum をコピー
COPY go.mod go.sum ./

# 依存ライブラリを取得（キャッシュレイヤー）
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
# CGO_ENABLED=1 は SQLite ドライバーが C 拡張を使用するため必須です。
# ビルド実行環境のアーキテクチャをそのまま使い、CGOクロスコンパイルを避けます。
RUN CGO_ENABLED=1 go build -o commute2invoice ./src/cmd

# Stage 2: ランタイムイメージ
# alpine:latest ベースで軽量化
FROM debian:bookworm-slim

# 必要なシステムパッケージをインストール
# curl: ヘルスチェック用
# ca-certificates: HTTPS通信用
RUN apt-get update \
  && apt-get install --no-install-recommends -y ca-certificates curl fontconfig fonts-noto-cjk wkhtmltopdf \
  && rm -rf /var/lib/apt/lists/*

# 日本語フォント（Noto Sans JP）をインストール
# PDF生成時に日本語テキストが正しく表示されるために必須
# 作業ディレクトリを設定
WORKDIR /app

# SQLiteの既定保存先を作成します。名前付きボリュームをマウントしない起動でも利用できます。
RUN mkdir -p /data

# ビルドステージからバイナリをコピー
COPY --from=builder /app/commute2invoice .
COPY --from=builder /app/src/frontend ./src/frontend

# データボリュームマウントポイント（DB ファイルの保存先）
VOLUME ["/data"]

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# ポート 8080 を公開
EXPOSE 8080

# アプリケーション起動
CMD ["./commute2invoice"]
