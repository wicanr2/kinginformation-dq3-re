#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

for script in package_appimage.sh package_win.sh; do
  output=""
  status=0
  output="$(bash "$ROOT/tools/$script" 2>&1)" || status=$?

  if [ "$status" -ne 2 ]; then
    printf '%s 預期 exit 2，實際 exit %s\n%s\n' "$script" "$status" "$output" >&2
    exit 1
  fi
  case "$output" in
    *"已停用的 dq3_remake C/SDL 歷史產品"*"現行產品是 dq3_remake_ebitan"*"dq3_remake_ebitan/build.sh"*) ;;
    *)
      printf '%s 缺少 fail-closed 或現行入口說明：\n%s\n' "$script" "$output" >&2
      exit 1
      ;;
  esac
done

printf '舊 C/SDL 發包入口皆已失敗即關閉。\n'
