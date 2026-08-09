#!/usr/bin/env bash
set -eu

# 在已完成的 macOS 交叉編譯輸出旁建立可搬移的 .app 與 ZIP。
# 預期從 Docker 容器執行，且 OUT_ROOT 指向可寫的 dist 目錄。
OUT_ROOT=${OUT_ROOT:?請指定 OUT_ROOT}

for arch in x86_64 arm64; do
  base="${OUT_ROOT}/dq3-20260809-macos-${arch}"
  binary="${base}/dq3-remake"
  app="${base}/Dragon Quest III.app"
  test -x "${binary}"
  rm -rf "${app}" "${base}.zip"
  mkdir -p "${app}/Contents/MacOS"
  cp "${binary}" "${app}/Contents/MacOS/dq3-remake"
  chmod 0755 "${app}/Contents/MacOS/dq3-remake"
  printf '%s\n' \
    '<?xml version="1.0" encoding="UTF-8"?>' \
    '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
    '<plist version="1.0">' '<dict>' \
    '<key>CFBundleExecutable</key><string>dq3-remake</string>' \
    '<key>CFBundleIdentifier</key><string>org.dq3cht.remake</string>' \
    '<key>CFBundleName</key><string>Dragon Quest III 精訊版重製</string>' \
    '<key>CFBundleDisplayName</key><string>Dragon Quest III 精訊版重製</string>' \
    '<key>CFBundlePackageType</key><string>APPL</string>' \
    '<key>CFBundleVersion</key><string>2026.08.09</string>' \
    '<key>LSMinimumSystemVersion</key><string>10.13</string>' \
    '</dict>' '</plist>' > "${app}/Contents/Info.plist"
  BASE="${base}" python3 -c 'import os, zipfile; base=os.environ["BASE"]; app=os.path.join(base, "Dragon Quest III.app"); out=base+".zip"; z=zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED); [z.write(os.path.join(root, name), os.path.relpath(os.path.join(root, name), base)) for root, _, names in os.walk(app) for name in names]; z.close()'
done
