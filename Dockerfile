FROM ubuntu:24.04

# 必要なパッケージをインストール
RUN apt update && \
    apt install -y golang git gcc && \
    apt clean && rm -rf /var/lib/apt/lists/*

# 作業ディレクトリ
WORKDIR /app

# Goファイルをコピー
COPY . .

# Goの依存を解決
RUN go get -u github.com/gin-gonic/gin

RUN go mod download

# Goバイナリをビルド
RUN go build -o main main.go

# ポート開放
EXPOSE 8080

# デーモンとして実行
CMD ["./backend"]