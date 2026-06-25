#!/usr/bin/env bash
# 打包 dq3_remake 可交付 distributable(不含版權素材)。
# 產出 work/dist/ + work/dq3_remake_dist.tar.gz(work/ 為 gitignore)。
# 容器內建置 + 複製 binary;素材需使用者自備(見 DIST_README)。
# 用法(host):tools/package.sh   # 內部用 docker + dq3build volume
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

HOST_UID="$(id -u)"; HOST_GID="$(id -g)"
echo "== 1. 容器內 release 建置(全 target)=="
docker run --rm -e HOST_UID="$HOST_UID" -e HOST_GID="$HOST_GID" -v "$ROOT":/repo -v dq3build:/build dq3-remake bash -lc '
  cmake -S /repo/dq3_remake -B /build -DCMAKE_BUILD_TYPE=Release >/tmp/c.log 2>&1
  cmake --build /build -j >/tmp/m.log 2>&1 || { tail -30 /tmp/m.log; exit 1; }
  echo "   build OK: $(stat -c %s /build/dq3_remake) bytes"
  # 交付前最終驗證:全綠才打包
  bash /repo/tools/game_tester.sh /repo/assets_raw /build/dq3_remake >/tmp/gt.log 2>&1 || { echo "   game_tester FAIL"; tail -20 /tmp/gt.log; exit 2; }
  echo "   game_tester: $(grep -E "結果:PASS" /tmp/gt.log)"
  # 複製產物出來(掛載到 /repo/work)
  mkdir -p /repo/work/dist/bin /repo/work/dist/verify
  cp /build/dq3_remake /repo/work/dist/bin/
  for t in /build/dq3_*_test; do cp "$t" /repo/work/dist/bin/ 2>/dev/null || true; done
  cp /repo/tools/game_tester.sh /repo/tools/playthrough_check.sh /repo/tools/mainline_check.sh /repo/work/dist/verify/
  cp /repo/dq3_remake/DIST_README.md /repo/work/dist/
  chown -R "${HOST_UID}:${HOST_GID}" /repo/work/dist
'

echo "== 2. 啟動腳本 + 版本戳 =="
cat > work/dist/run.sh <<'RUN'
#!/usr/bin/env bash
# 啟動 dq3_remake。用法:./run.sh /path/to/assets_raw
set -eu
ASSETS="${1:?用法: ./run.sh /path/to/assets_raw}"
DIR="$(cd "$(dirname "$0")" && pwd)"
exec "$DIR/bin/dq3_remake" "$ASSETS" game
RUN
chmod +x work/dist/run.sh work/dist/verify/*.sh
git -C "$ROOT" rev-parse --short HEAD > work/dist/COMMIT.txt 2>/dev/null || echo "unknown" > work/dist/COMMIT.txt

echo "== 3. 打包 tarball =="
( cd work && tar czf dq3_remake_dist.tar.gz dist )
ls -la work/dq3_remake_dist.tar.gz
echo "== ✅ 打包完成:work/dq3_remake_dist.tar.gz(commit $(cat work/dist/COMMIT.txt))=="
echo "   內容:"; tar tzf work/dq3_remake_dist.tar.gz | head -30
