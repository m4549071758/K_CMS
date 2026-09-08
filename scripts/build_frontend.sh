#!/bin/bash
set -euo pipefail

# ビルドスクリプト - フロントエンドのビルドとデプロイを実行
# 使用方法: ./build_frontend.sh <action> <article_id>
# 例: ./build_frontend.sh create abc-123

LOG_PREFIX="[Frontend Build]"
FRONTEND_DIR="/root/blog"
BUILD_OUTPUT_DIR="/root/blog/out"
DEPLOY_DIR="/var/www/html/blog"
# CloudflareがHTMLを7日間キャッシュするため、旧HTMLが参照する静的assetを8日間保持する。
STATIC_ASSET_RETENTION_MINUTES=11520

# 引数のチェック
if [ $# -lt 2 ]; then
    echo "$LOG_PREFIX ERROR: Missing arguments"
    echo "Usage: $0 <action> <article_id>"
    exit 1
fi

ACTION=$1
ARTICLE_ID=$2

echo "$LOG_PREFIX ============================================"
echo "$LOG_PREFIX Starting build process..."
echo "$LOG_PREFIX Action: $ACTION"
echo "$LOG_PREFIX Article ID: $ARTICLE_ID"
echo "$LOG_PREFIX Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "$LOG_PREFIX ============================================"

# フロントエンドディレクトリの存在確認
if [ ! -d "$FRONTEND_DIR" ]; then
    echo "$LOG_PREFIX ERROR: Frontend directory not found: $FRONTEND_DIR"
    exit 1
fi

# フロントエンドディレクトリに移動
cd "$FRONTEND_DIR" || {
    echo "$LOG_PREFIX ERROR: Failed to change directory to $FRONTEND_DIR"
    exit 1
}

# 依存関係のインストール（package.jsonが変更されている可能性を考慮）
echo "$LOG_PREFIX Checking dependencies..."
if npm install --silent; then
    echo "$LOG_PREFIX Dependencies up to date"
else
    echo "$LOG_PREFIX WARNING: npm install failed, continuing with existing dependencies..."
fi

# ビルド実行
echo "$LOG_PREFIX Running npm run build..."
BUILD_START=$(date +%s)

if npm run build; then
    BUILD_END=$(date +%s)
    BUILD_DURATION=$((BUILD_END - BUILD_START))
    echo "$LOG_PREFIX Build completed successfully in ${BUILD_DURATION}s"
else
    echo "$LOG_PREFIX ERROR: Build failed"
    exit 1
fi

# ビルド出力ディレクトリの存在確認
if [ ! -d "$BUILD_OUTPUT_DIR" ]; then
    echo "$LOG_PREFIX ERROR: Build output directory not found: $BUILD_OUTPUT_DIR"
    exit 1
fi

# デプロイディレクトリが存在しない場合は作成
if [ ! -d "$DEPLOY_DIR" ]; then
    echo "$LOG_PREFIX Creating deploy directory: $DEPLOY_DIR"
    mkdir -p "$DEPLOY_DIR" || {
        echo "$LOG_PREFIX ERROR: Failed to create deploy directory"
        exit 1
    }
fi

# HTMLなどは完全同期するが、Cloudflareに残る旧HTMLが参照する静的assetは保持する。
echo "$LOG_PREFIX Deploying build output to $DEPLOY_DIR..."
mkdir -p "$DEPLOY_DIR/_next/static"

if ! rsync -av "$BUILD_OUTPUT_DIR/_next/static/" "$DEPLOY_DIR/_next/static/"; then
    echo "$LOG_PREFIX ERROR: Failed to deploy static assets"
    exit 1
fi

if ! rsync -av --delete --exclude '/_next/static/' "$BUILD_OUTPUT_DIR/" "$DEPLOY_DIR/"; then
    echo "$LOG_PREFIX ERROR: Failed to deploy build output"
    exit 1
fi

echo "$LOG_PREFIX Deployment successful"

# 現行ビルドに存在せず、保持期間を過ぎた静的assetだけを削除する。
while IFS= read -r -d '' deployed_asset; do
    relative_path=${deployed_asset#"$DEPLOY_DIR/"}
    if [ ! -e "$BUILD_OUTPUT_DIR/$relative_path" ]; then
        rm -f -- "$deployed_asset"
    fi
done < <(find "$DEPLOY_DIR/_next/static" -type f -mmin "+$STATIC_ASSET_RETENTION_MINUTES" -print0)

find "$DEPLOY_DIR/_next/static" -depth -type d -empty -delete

# 所有者をnginx:nginxに変更
echo "$LOG_PREFIX Changing ownership to nginx:nginx..."
if chown -R nginx:nginx "$DEPLOY_DIR"; then
    echo "$LOG_PREFIX Ownership changed successfully"
else
    echo "$LOG_PREFIX ERROR: Failed to change ownership"
    exit 1
fi

# 完了メッセージ
echo "$LOG_PREFIX ============================================"
echo "$LOG_PREFIX Build and deploy completed successfully!"
echo "$LOG_PREFIX Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "$LOG_PREFIX ============================================"

exit 0
