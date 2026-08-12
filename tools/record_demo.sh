#!/usr/bin/env bash
# 相容入口：舊版會在主機跑 ffmpeg，且錄製已停用的 C/SDL 產品，現已失敗即關閉。
set -euo pipefail

cat >&2 <<'EOF'
錯誤：舊 C/SDL 推廣片錄製流程已停用。
現行產品是 dq3_remake_ebitan；請在一次性 Docker 容器內執行：

  tools/build_ebiten_promo_video.sh

建置完成後再以 tools/verify_promo_video.sh 驗證，不能只檢查 MP4 是否含 audio stream。
EOF
exit 2
