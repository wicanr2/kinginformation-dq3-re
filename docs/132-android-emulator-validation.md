# Android emulator 動態驗證

## 驗證範圍

- 儲存庫 HEAD：`1f6ce696b1162e7f9c2b25738daa7e9ca3e0f947`
- APK SHA-256：`38c222238a9dab068e37c5bf86eb3bad7c561f7104e32797c1a8e7c153a5e8ad`
- emulator：37.1.11
- system image：API 34 x86_64 revision 14（Android 14）
- 加速：Docker 內掛入 `/dev/kvm`，`environment.txt` 為 `emulator_accel=on`
- 網路：`--network none`

驗收入口使用 [`android/tools`](../dq3_remake_ebitan/android/tools/README.md) 的有界腳本；APK
安裝、Activity、觸控、HOME 暫停／恢復、force-stop／重啟與 crash 掃描是相互獨立的 gate。

## 結果

| Gate | 結果 | 證據 |
|---|---|---|
| Android framework 開機 | 通過 | `sys.boot_completed=1`，package／activity／window service 可回應 |
| APK 安裝 | 通過 | `adb install -r`：`Performing Streamed Install`、`Success` |
| Activity 啟動 | **失敗** | `MainActivity` 的 main thread 發生 `FATAL EXCEPTION` |
| 前景 Activity | 失敗 | `topResumedActivity`／`mCurrentFocus` 均回到 Nexus Launcher |
| 觸控、截圖 | 未執行 | Activity gate 失敗即關閉 |
| pause／resume、force-stop／restart | 未執行 | Activity gate 失敗即關閉 |

崩潰鏈已由 Android runtime 堆疊直接閉合：

```text
MainActivity.onCreate(MainActivity.java:21)
→ applyGameChrome(MainActivity.java:41)
→ PhoneWindow.getInsetsController(PhoneWindow.java:3968)
→ DecorView.getWindowInsetsController() 的 receiver 為 null
→ NullPointerException／Activity 被 force-finish
```

原始碼在 `setContentView(ebitenView)` 前呼叫 `applyGameChrome()`；Android 14 此時尚沒有可供
`PhoneWindow.getInsetsController()` 使用的 decor。程式雖檢查回傳的 controller 是否為 null，
但例外在取得 controller 的過程中已先發生，因此該檢查無法保護此路徑。

## 判定與下一步

這是 **Android runtime 產品缺陷**，不是 emulator、APK 安裝或素材載入失敗。Android 狀態仍
為 `build-only`；不得宣稱動態啟動、觸控或 lifecycle 通過。

最小修正應只調整 Android 外殼的沉浸式 UI 套用時機：先建立／掛上 content view，再於 decor
可用時取得 `WindowInsetsController`，並保留 `onWindowFocusChanged(true)` 的重套行為。修正後
用同一 emulator、同一腳本重跑全部 gate；不要以單元測試或 `am start` 的 `Status: ok` 取代
前景 Activity 與 logcat 證據。

完整未追蹤證據保留於 `/tmp/dq3-android-kvm-smoke-20260812/`，包含
`activity-first.txt`、`logcat-first.txt`、`install.log`、`start.log`、`environment.txt` 與
`emulator.log`。APK、原版素材及動態輸出不加入 Git。
