#!/usr/bin/env bash
# 已停用：此檔原本只會封裝歷史 dq3_remake C/SDL 產品。
set -euo pipefail

cat >&2 <<'NOTICE'
此腳本屬於已停用的 dq3_remake C/SDL 歷史產品，不得用來建立現行發佈包。
現行產品是 dq3_remake_ebitan；請使用 dq3_remake_ebitan/build.sh，在 Docker 內完成
Go／Ebitengine 建置與驗收，再依現行 release 流程封裝 Windows ZIP。
NOTICE
exit 2
