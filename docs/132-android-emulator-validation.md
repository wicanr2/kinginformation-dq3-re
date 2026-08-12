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

## 2026-08-12 Luna Max 修正與第二層驗證

`MainActivity` 已採最小 lifecycle 修正：`Seq`／`EbitenView`／`setContentView` 完成後才首次
呼叫 `applyGameChrome()`，`onPause()`／`onResume()` 也防護尚未初始化的 view。新增
`check-main-activity-contract.sh`，以失敗即關閉方式鎖定首次 chrome 套用順序、焦點重套與
lifecycle 防護。

修正後的 build-only 產物：

- AAR：122,650,265 bytes，SHA-256
  `32db89d8af50b492292b2fe7f4a5d31d965da775269caa199410e7decdaa2c9d`。
- debug APK：125,713,466 bytes，SHA-256
  `69b678b237f1efb9902ea0bdc6e825be284e0217d6fdd84b92c94125d33b7b26`。
- `aapt`：`com.wicanr2.dq3`、label「傳說的終章」、compile SDK 35、target SDK 34、
  `MainActivity`；均維持原契約。

KVM emulator 已證明原空指標消失，APK 可安裝，Activity 可保持 resumed，HOME 回復與
force-stop 後冷啟動都有存活 PID，logcat 沒有 DQ3 crash／ANR。不過人工檢查推翻腳本最初的
「通過」：production APK 的畫面仍停在 Android splash／全黑，`firstWindowDrawn=false`。
兩個 DQ3 process 的最後進度都停在 OboeAudio `openStreamInternal/configure`，沒有 Go panic。

隔離、未進 production 的診斷探針只把 mobile 傳入 `game.NewGame` 的 music FS 改為 `nil`；
同一 KVM emulator 隨即在約 2.7 秒完成冷啟動，顯示 DQ3 標題與開場年份，且 install、前景、
觸控截圖、HOME 回復、force-stop／restart、PID、無 crash／ANR 全部通過。這把「headless
emulator host audio backend 令同步 `audio.NewContext` 阻塞」提升為強推論；AAR 內資產缺失
已被排除。診斷探針不加入 Git，也不以永久停用 Android 音樂作正式修正。

smoke 腳本因此新增 `firstWindowDrawn=true` 且不得殘留 DQ3 splash 的玩家可見 gate。現行
Android 判定是：**Activity 啟動崩潰已修；production 音訊版的 headless emulator 畫面 gate
仍受 host audio backend 阻塞；靜音診斷探針完整通過。** 真機或具可用 PulseAudio／ALSA
backend 的 emulator 尚需重跑 production APK，不能把探針結果升格為正式 Android 驗收完成。
