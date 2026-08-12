# Android 版專屬 UX 規格（保存優先）

本文件把既有 `docs/64-android-touch-ui-plan-v2.md` 的已實作部分整理成 Android 版
交付規格。Android 沒有原版實機 oracle，因此這裡的 UX 是 **port 專屬設計**，不冒充
精訊版原始畫面；遊戲畫布、中文字、事件順序與存檔契約仍以桌面／原版證據為準。

## 設計目標

1. 保留精訊版 640×350 畫布、繁中劇本、注音／英數選字與原版框線；手機只增加輸入與
   系統適配，不重新設計成現代卡牌或手遊 HUD。
2. 橫向單手可操作，拇指不必遮住對話框、戰鬥文字與隊伍狀態。
3. 觸控仍只產生跨平台 `InputState`；`game/` 不知道輸入來自 Android、鍵盤或測試 trace。
4. 不把任何原始遊戲檔、錄影、ROM 或生成的 `.aar`／APK 加入公開 Git。

## 畫面與控制

| 區域 | 行為 | 規格 |
|---|---|---|
| 中央 | 精訊版遊戲畫布 | 固定邏輯尺寸 640×350，等比 letterbox；不裁切、不拉伸 |
| 左下 | 浮動十字鍵 | 四方向、死區、按住續走；首次觸碰才顯示，避免遮住桌面版畫面 |
| 右下 | A／B | A 確定／對話／推進，B 取消／返回；edge 語意，多點觸控 |
| 右上 | 情境圖示鍵 | 標題顯示齒輪並開設定，命名顯示雙向切換圖示並切換注音／英數；其他場景隱藏 |
| 非控制區 | 直接點選 | 以現有 hit-list 對選單、商店、注音盤提供輔助；十字鍵＋A 永遠保留 |

控制疊在畫布四角，不把 pillarbox 黑邊當成輸入區。邏輯座標維持既有 `TouchUI` 常數，
避免每台手機的 DPI 或 PPI 猜值改變遊戲座標；安全區以 Android 沉浸式視窗及畫布內縮
共同處理。

## Android 外殼行為

- `sensorLandscape` 橫向啟動，隱藏 action bar 與系統列；使用者以系統手勢暫時叫出
  系統列時，回到前景再套回沉浸式畫面。
- `FLAG_KEEP_SCREEN_ON` 避免長對話、戰鬥或音樂保存展示時自動熄屏。
- `onPause`／`onResume` 呼叫 Ebitengine 的 suspend／resume；不在背景偷偷推進步數、
  遭遇或日夜時鐘。
- Android 應用程式名稱使用精訊版中文標題「傳說的終章」；不把英文暫名當成玩家可見
  的遊戲名稱。遊戲內文字仍全部來自 game-pack JSON。

## 文化保存與發佈界線

Android build 可以在本機合法持有的 `mobile/assets/` 上產生 `.aar`／APK，但這些產物
不等於可公開散布的遊戲包。公開 patch 只含引擎、資料契約、工具與文件；原版資產、
MT-32 ROM、原版影片及含資產的 APK 只留在本機或私有保存媒介。

README 的文化保存入口：

- [台灣本土攻略原文](../references/walkthroughs/README.md)
- [1994–1995 BBS 史料](history/dq3-bbs-1994.md)
- [精訊音樂保存紀錄](58-taiwan-jingxun-music-preservation.md)

## 驗收分級

- **Android build**：Docker 內 `ebitenmobile bind`／Gradle `assembleDebug` 成功，APK
  具備中文 label、橫向與沉浸式設定；記錄 SHA-256。
- **Android smoke**：APK 可安裝、`EbitenView` 建立、背景暫停／恢復不崩潰；需要 emulator
  或真機，不能以桌面包代替。
- **Android touch path**：以真機抽測命名（注音／英數）、標題設定、對話、地表移動、
  戰鬥 A/B、多點觸控與存讀檔。這些是 port UX gate，不升格為原版 V3 parity。

本輪若只有 Docker build 而沒有 emulator／真機，狀態標為 `build-only`，不可宣稱
Android 實機驗收完成。

## 本次 Docker build 紀錄

- Go AAR：`dist/dq3-android-current-20260811.aar`（SHA-256：`cc0c12730c39190362e9699d07232e553d86d4a76fbdb209004f28965a4c8566`）
- 產物：`dist/dq3-android-debug-20260811.apk`（ignored，本機含合法素材的 debug 包）
- SHA-256：`ea830ea06dfde28ce3e912669092663c4cde03069575a8fd695c1e1915efe62c`
- SDK 35 `/opt/android-sdk/build-tools/35.0.0/aapt dump badging`：`com.wicanr2.dq3`、version `1.0`、label「傳說的終章」、landscape feature
- 狀態：`build-only`；目前沒有 emulator／真機輸入與生命週期 smoke 證據。

### 2026-08-12 最新 source 重建

- Go AAR（`/tmp`、不納入 Git）：122,651,070 bytes；SHA-256
  `8d7a1051ce9c66383a5a548114961126a4cc897b92e1f62b90ea94b451e95ef7`。
- debug APK（`/tmp`、含本機合法素材、不納入 Git）：125,714,249 bytes；SHA-256
  `38c222238a9dab068e37c5bf86eb3bad7c561f7104e32797c1a8e7c153a5e8ad`。
- SDK 35 `aapt`：`com.wicanr2.dq3`、label「傳說的終章」、`MainActivity`、
  `sensorLandscape`、四 ABI；狀態仍為 `build-only`。
- 動態 smoke 工具已固定為 `dq3-android-emulator:20260812-r1`：Android emulator 37.1.11、
  API 34 x86_64 system image revision 14，建置來源與有界腳本位於
  [`android/tools`](../dq3_remake_ebitan/android/tools/README.md)。無 KVM 的純 TCG＋SwiftShader
  測試中，ADB 於 181 秒上線、Android 14 可辨識、boot animation 於 247 秒停止，但 480 秒內
  `sys.boot_completed` 始終為空，package manager／framework 尚未 ready，故沒有執行 APK
  安裝。完成 smoke 前不得宣稱安裝、觸控、lifecycle 或存讀檔通過。
- 上述重建刻意使用 `LC_ALL=C`、離線 cache；`build-android.sh` 以
  `JAVA_TOOL_OPTIONS=-Dfile.encoding=UTF-8` 固定 `ebitenmobile` 內部 `javac` 編碼，因此 Go
  匯出符號的繁中註解不再造成 US-ASCII 編譯失敗。
