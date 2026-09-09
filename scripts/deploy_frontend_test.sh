#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
DEPLOY_SCRIPT=${1:-"$SCRIPT_DIR/deploy_frontend.sh"}
root=$(mktemp -d)
trap 'rm -rf -- "$root"' EXIT
build="$root/build"
deploy="$root/published"
state="$deploy/.static-asset-retirement"
mkdir -p "$build/_next/static/chunks" "$deploy/_next/static/chunks"

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

publish() {
    if ! bash "$DEPLOY_SCRIPT" "$build" "$deploy" > "$root/deploy.log" 2>&1; then
        cat "$root/deploy.log" >&2
        fail 'deployment failed'
    fi
}

printf 'old-page' > "$deploy/index.html"
printf 'obsolete-page' > "$deploy/deleted.html"
printf 'old-chunk' > "$deploy/_next/static/chunks/old.js"
printf 'legacy-font' > "$deploy/_next/static/chunks/old font.woff2"
touch -d '30 days ago' "$deploy/_next/static/chunks/old.js" "$deploy/_next/static/chunks/old font.woff2"
printf 'new-page-with-current-chunk' > "$build/index.html"
printf 'current-chunk' > "$build/_next/static/chunks/current.js"
touch -d '30 days ago' "$build/_next/static/chunks/current.js"

publish
[ -f "$deploy/_next/static/chunks/old.js" ] || fail 'long-lived chunk deleted at retirement'
[ -f "$deploy/_next/static/chunks/old font.woff2" ] || fail 'retired font deleted at retirement'
[ "$(cat "$deploy/index.html")" = 'new-page-with-current-chunk' ] || fail 'HTML not updated'
[ ! -e "$deploy/deleted.html" ] || fail 'obsolete HTML retained'
[ "$(cat "$deploy/_next/static/chunks/current.js")" = 'current-chunk' ] || fail 'current chunk not published'
echo 'PASS: long-lived chunks and fonts survive retirement; HTML is synchronized'

# マーカーの時計を進め、待機なしで退役後の期限境界を検証する。
touch -d '11500 minutes ago' "$state/chunks/old.js"
publish
[ -f "$deploy/_next/static/chunks/old.js" ] || fail 'chunk deleted before eight days'
touch -d '11522 minutes ago' "$state/chunks/old.js"
publish
[ ! -e "$deploy/_next/static/chunks/old.js" ] || fail 'retirement deadline extended by repeated deployments'
[ -f "$deploy/_next/static/chunks/old font.woff2" ] || fail 'younger retirement deleted'
[ -f "$deploy/_next/static/chunks/current.js" ] || fail 'long-lived current chunk deleted'
echo 'PASS: retirement expires after eight days without sliding or deleting current assets'

# 期限切れ直前のassetを再公開し、次の退役から新たに8日間保持する。
printf 'returned-chunk' > "$build/_next/static/chunks/returned.js"
publish
rm -- "$build/_next/static/chunks/returned.js"
publish
touch -d '9 days ago' "$state/chunks/returned.js"
printf 'returned-chunk' > "$build/_next/static/chunks/returned.js"
touch -d '30 days ago' "$build/_next/static/chunks/returned.js"
publish
[ -f "$deploy/_next/static/chunks/returned.js" ] || fail 'reactivated chunk deleted'
rm -- "$build/_next/static/chunks/returned.js"
publish
[ -f "$deploy/_next/static/chunks/returned.js" ] || fail 'reactivated chunk inherited old retirement deadline'
echo 'PASS: reactivation cancels the old deadline; retirement starts again'

# 不完全なビルドや転送失敗で公開済みHTML・assetを削除しない。
mv -- "$build/index.html" "$root/index.html"
if bash "$DEPLOY_SCRIPT" "$build" "$deploy" > "$root/deploy.log" 2>&1; then
    fail 'incomplete build accepted'
fi
[ "$(cat "$deploy/index.html")" = 'new-page-with-current-chunk' ] || fail 'incomplete build changed HTML'
mv -- "$root/index.html" "$build/index.html"
mkdir "$root/bin"
printf '#!/bin/bash\nexit 23\n' > "$root/bin/rsync"
chmod +x "$root/bin/rsync"
if PATH="$root/bin:$PATH" bash "$DEPLOY_SCRIPT" "$build" "$deploy" > "$root/deploy.log" 2>&1; then
    fail 'asset transfer failure ignored'
fi
[ "$(cat "$deploy/index.html")" = 'new-page-with-current-chunk' ] || fail 'failed transfer changed HTML'
[ -f "$deploy/_next/static/chunks/returned.js" ] || fail 'failed transfer deleted retired chunk'
echo 'PASS: incomplete build and transfer failure leave published files intact'

# 初回デプロイで公開先ディレクトリがなくても動作する。
if ! bash "$DEPLOY_SCRIPT" "$build" "$root/fresh/published" > "$root/deploy.log" 2>&1; then
    cat "$root/deploy.log" >&2
    fail 'first deployment failed'
fi
[ "$(cat "$root/fresh/published/index.html")" = 'new-page-with-current-chunk' ] || fail 'first deployment missed HTML'
[ -f "$root/fresh/published/_next/static/chunks/current.js" ] || fail 'first deployment missed chunk'
echo 'PASS: first deployment creates the complete output'
