#!/bin/bash
set -euo pipefail

LOG_PREFIX="[Frontend Deploy]"
# CloudflareのHTMLキャッシュ7日間より長く、旧版への移行後8日間保持する。
STATIC_ASSET_RETENTION_MINUTES=11520

if [ "$#" -ne 2 ]; then
    echo "$LOG_PREFIX ERROR: Usage: $0 <build_output_dir> <deploy_dir>" >&2
    exit 1
fi

BUILD_OUTPUT_DIR=${1%/}
DEPLOY_DIR=${2%/}
if [ ! -f "$BUILD_OUTPUT_DIR/index.html" ] || [ ! -d "$BUILD_OUTPUT_DIR/_next/static" ]; then
    echo "$LOG_PREFIX ERROR: Incomplete frontend build: $BUILD_OUTPUT_DIR" >&2
    exit 1
fi

# 空のマーカーのmtimeを退役時刻として使う。asset本体のmtimeは変更しない。
RETIREMENT_DIR="$DEPLOY_DIR/.static-asset-retirement"
mkdir -p "$DEPLOY_DIR/_next/static" "$RETIREMENT_DIR"
asset_list=$(mktemp)
trap 'rm -f -- "$asset_list"' EXIT
trap 'echo "$LOG_PREFIX ERROR: Deployment failed at line $LINENO" >&2' ERR

# HTMLを公開する前にassetを配置し、再登場したassetの古い期限を解除する。
rsync -av "$BUILD_OUTPUT_DIR/_next/static/" "$DEPLOY_DIR/_next/static/"
find "$BUILD_OUTPUT_DIR/_next/static" -type f -print0 > "$asset_list"
while IFS= read -r -d '' source_asset; do
    relative_path=${source_asset#"$BUILD_OUTPUT_DIR/_next/static/"}
    rm -f -- "$RETIREMENT_DIR/$relative_path"
done < "$asset_list"

rsync -av --delete --exclude '/_next/static/' --exclude '/.static-asset-retirement/' "$BUILD_OUTPUT_DIR/" "$DEPLOY_DIR/"

# HTMLの切替成功後に初めて退役したassetを記録する。再デプロイで期限を延長しない。
# 移行前から残っていたassetも、初回検出から8日間は削除しない。
find "$DEPLOY_DIR/_next/static" -type f -print0 > "$asset_list"
while IFS= read -r -d '' deployed_asset; do
    relative_path=${deployed_asset#"$DEPLOY_DIR/_next/static/"}
    marker="$RETIREMENT_DIR/$relative_path"
    if [ ! -e "$BUILD_OUTPUT_DIR/_next/static/$relative_path" ] && [ ! -e "$marker" ]; then
        mkdir -p -- "$(dirname -- "$marker")"
        touch -- "$marker"
    fi
done < "$asset_list"

find "$RETIREMENT_DIR" -type f -mmin "+$STATIC_ASSET_RETENTION_MINUTES" -print0 > "$asset_list"
while IFS= read -r -d '' marker; do
    relative_path=${marker#"$RETIREMENT_DIR/"}
    if [ ! -e "$BUILD_OUTPUT_DIR/_next/static/$relative_path" ]; then
        rm -f -- "$DEPLOY_DIR/_next/static/$relative_path"
        rm -f -- "$marker"
    fi
done < "$asset_list"

find "$DEPLOY_DIR/_next/static" "$RETIREMENT_DIR" -depth -mindepth 1 -type d -empty -delete

echo "$LOG_PREFIX Deployment successful"
