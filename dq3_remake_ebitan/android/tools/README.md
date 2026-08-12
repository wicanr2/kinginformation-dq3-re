# Android 模擬器 smoke 工具

本工具固定使用 Android emulator 37.1.11 與 API 34 x86_64 system image revision 14。
它只驗證 APK 安裝、Activity 啟動、觸控輸入通道、暫停／恢復與重新啟動；不代表真機
效能、音訊體感或原版 V3 parity。

建置 image：

```bash
docker build \
  -t dq3-android-emulator:20260812-r1 \
  -f dq3_remake_ebitan/android/tools/Dockerfile.emulator \
  dq3_remake_ebitan/android/tools
```

執行時必須限制資源、掛入 APK 與輸出目錄，並提供可寫的 `/.android` tmpfs；emulator 37
會在該處建立 JWKS 與 ADB 金鑰：

```bash
timeout 540s docker run --rm \
  --network none \
  --memory 6g \
  --cpus 2 \
  --pids-limit 1024 \
  --device /dev/kvm \
  --group-add 993 \
  --tmpfs /.android:rw,uid=1000,gid=1000,mode=0700 \
  -u 1000:1000 \
  -v /absolute/path/to/app-debug.apk:/input/app-debug.apk:ro \
  -v /absolute/path/to/output:/out \
  dq3-android-emulator:20260812-r1 \
  /input/app-debug.apk /out
```

`--group-add` 應使用主機 `/dev/kvm` 的實際群組 ID；上例的 `993` 是 2026-08-12 驗證主機值。
腳本會把實際選用的加速模式寫入 `environment.txt`。若主機沒有把 `/dev/kvm` 提供給容器，
工具會退回純 TCG。2026-08-12 的 KVM 驗證已完成 Android 14 framework 開機與 APK 安裝，
`MainActivity.applyGameChrome()` 的啟動空指標也已修正；production 音訊版仍受 headless host
audio backend 阻塞。腳本會要求 `firstWindowDrawn=true` 且不得殘留 DQ3 splash，避免只憑
resumed Activity 誤報通過。隔離靜音探針通過不代表正式音訊版或真機通過；完整證據見
[`docs/132`](../../../docs/132-android-emulator-validation.md)。
