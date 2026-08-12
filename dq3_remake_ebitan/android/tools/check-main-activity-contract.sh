#!/usr/bin/env bash
set -euo pipefail

SOURCE_PATH="${1:-$(dirname "$0")/../app/src/main/java/com/wicanr2/dq3/MainActivity.java}"

if [[ ! -f "$SOURCE_PATH" ]]; then
    printf '找不到 MainActivity 原始碼：%s\n' "$SOURCE_PATH" >&2
    exit 1
fi

# 此檢查刻意只驗證原始碼結構並採 fail-closed；生命週期修改不得再次把
# chrome 設定移到 content view 之前。
python3 - "$SOURCE_PATH" <<'PY'
from pathlib import Path
import sys

source_path = Path(sys.argv[1])
source = source_path.read_text(encoding="utf-8")

on_create_start = source.find("protected void onCreate(")
if on_create_start < 0:
    raise SystemExit("缺少 onCreate 宣告")
# MainActivity 的方法均以 @Override 分隔；使用下一個標記作為窄邊界，
# 避免用第一個大括號結束處截斷含有內層區塊的方法。
on_create_end = source.find("\n    @Override", on_create_start)
if on_create_end < 0:
    raise SystemExit("找不到 onCreate 後的生命週期方法邊界")
on_create = source[on_create_start:on_create_end]

content_index = on_create.find("setContentView(ebitenView);")
chrome_index = on_create.find("applyGameChrome();")
if content_index < 0:
    raise SystemExit("onCreate 缺少 setContentView(ebitenView)")
if chrome_index < 0:
    raise SystemExit("onCreate 缺少 applyGameChrome()")
if content_index >= chrome_index:
    raise SystemExit("fail-closed：onCreate 必須先 setContentView，再 applyGameChrome")

if "if (ebitenView != null)" not in source:
    raise SystemExit("fail-closed：生命週期回呼必須防護尚未初始化的 ebitenView")

focus_start = source.find("public void onWindowFocusChanged(boolean hasFocus)")
if focus_start < 0:
    raise SystemExit("缺少 onWindowFocusChanged 宣告")
focus = source[focus_start:]
if "if (hasFocus)" not in focus or "applyGameChrome();" not in focus:
    raise SystemExit("fail-closed：視窗取得焦點時必須重新套用 chrome")

print(f"MainActivity 生命週期契約通過：{source_path}")
PY
