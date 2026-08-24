# DQ3 Go／Ebitengine 現行工作清單

更新：2026-08-24。唯一詳細計畫仍是 [`docs/74`](docs/74-ebiten-remake-completion-plan.md)；
本檔只保存可快速接手的目前順序，不收錄歷史 C/SDL 工作。

| 順序 | 工作 | 狀態 | 直接證據 |
|---:|---|---|---|
| 1 | 戰鬥道具 | 完成（D3／E2） | [`docs/179`](docs/179-battle-item-selector-runtime-spec.md) |
| 2 | 原始怪物表實際使用的 action | 完成（39/39 definitions，D2／E2） | [`docs/180`](docs/180-monster-action-runtime-spec.md) |
| 3 | 剩餘合法野外咒文 | 完成（D3／E2） | [`docs/181`](docs/181-field-support-spells-runtime-spec.md) |
| 4 | 商店賣出／售價／裝備交易 | 完成（一般店 D3／E2） | [`docs/182`](docs/182-shop-sell-runtime-spec.md) |
| 5 | 船直接進城／離船／出城復船 | 完成（D2／E2） | [`docs/183`](docs/183-ship-town-entry-transaction.md) |
| 6 | 必要選單與玩家可見支線 | 完成（功能 E2；主線 E3） | 瑪依拉特殊店／王者之劍見 [`docs/184`](docs/184-maira-kings-sword-special-shop.md)；狀況／道具／裝備正常入口既有，逐窗 V3 非功能 gate |

不屬於功能完成 gate：全遊戲逐畫面 V3、DOS PCM/DAC/PIT wall-clock 深挖、未被原始資料
使用的 helper/action。這些只能由新的玩家可見差異或正式發行需求重新開啟。

至此六項功能 worklist 已清空。現行狀態分成三類，後續不得混寫：

| 分類 | 現況 | 是否阻塞 remake 功能完成 |
|---|---|---|
| remake 玩家流程 | 六項功能皆已接入；2026-08-24 game test 找到的開場轉場後自動行走缺段已依 [`docs/188`](docs/188-opening-escort-to-castle-spec.md) 改為兩階段 pack 序列；component、正式創角 trace 開場段與王座勇者像素差異均通過 | 否；完整 trace 後段仍會在既有 Lv20 練級上限失敗，與本切片無關 |
| 原版證據限制 | 藥草治療量、祈禱之戒 MP 回復量、聖水步數仍是 classic／舊 C 近似；部分逐窗、逐幀與音效只到 V1／V2／unknown | 否；文件與程式必須保留近似／未知標記，不得冒稱 exact |
| 交付 | `v0.1.35` 已由 commit `b26bb48` 建立 Linux x86_64 AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP 的公開 patch 與本機完整版；公開 patch 私有素材為 0 | 已完成；checksum、smoke 與限制見 [`docs/190`](docs/190-release-v0.1.35.md) |

2026-08-24 開場抽樣對拍確認：rec80 的事件順序與文字語意相符，但原版城鎮道路／水道背景
與 remake 小型城堡圖塊顯著不符；rec78 勇者可見且角色關係相符，但 viewport、palette、
對話框及分頁仍不是 V3。詳見 [`docs/189`](docs/189-opening-sampled-parity-20260824.md)。這些
不反向推翻 campaign E3，但阻止 README／release 使用「視覺忠實完成」或完整 V3 的措辭。

2026-08-24 玩家路徑 polish：正常產品不再自動退回 FM；注音組字回到獨立下方面板，
`ㄨㄤˇ` 以明示相容別名補「王」；母親帶路改為 pack-owned 可見逐格序列；NPC 待機切換
兩幀，移動時由 CTY ctrl 方向同步 renderer 面向。具名測試與證據見 [`docs/187`](docs/187-opening-player-path-polish.md)。

2026-08-23 polish：record345 隊長姓名、蘭西爾 `0x13` 無 writer 的否定性證據、舊 C
`partyBlind/partySealed` 死狀態，以及 D3MNS `+0x27` 未使用 raw 欄均已由 [`docs/185`](docs/185-remaining-polish-closure.md)
閉合，不再列為 remake 待辦。

## 本機交付（已完成）

2026-08-23 已完成：

1. `dist-all/v0.1.36-local/full/`：Linux x86_64 AppImage、Windows x86_64 ZIP、macOS
   x86_64／arm64 ZIP，均為本機完整版；包含 18 軌 MT-32 OGG 與 CONTROL／PCM ROM。
2. `dist-all/v0.1.35-local/promo/`：72 秒有音樂推廣片、實機開場來源、contact sheet、
   FFprobe、音量／非靜音驗收及音樂來源雜湊。
3. 四包皆核對架構、CRC／解包、checkpoint、修正版 `DQ3MNS.SHP` 與 IDA sidecar 排除；Linux
   AppImage 另通過 Docker＋Xvfb 8 秒啟動。macOS 沒有真機驗收，維持靜態驗證標記。
4. `SHA256SUMS.txt` 與 `LOCAL-DELIVERY.txt` 已放在同一版本根目錄。

同日依玩家實際操作反證修正 BLS 上／左 frame、NPC 交談轉向、`F1` HELP、`F2`／`S`
系統設定及一般完整版無音樂；具名測試、pack validator、桌面純編譯、ZIP CRC／架構與 Linux
ALSA null sink 啟動 smoke 已通過，見 [`docs/186`](docs/186-player-controls-audio-polish.md)。

2026-08-24 曾以當時的 schema `0.1.50`／content `0.1.56` source 建立
`dist-all/v0.1.37-local/full/linux-amd64/dq3-remake-v0.1.37-local-full-linux-amd64.AppImage`。
其 `AppRun` 明確把 `DQ3_ASSETS` 指向包內 `usr/share/dq3/assets_raw`，修正前包把 OGG 放在該處、
卻指向另一目錄而退回 FM 的封裝錯誤；包內含 18 軌 OGG 與兩個 MT-32 ROM，Docker＋Xvfb／
ALSA null smoke 通過，SHA-256 為
`719affa467a17daca51b621cd85faaee5136f5fe9e4e53ac70419f76f24d5321`。這是 Linux 單平台
本機更新；它早於 `docs/188` 的開場修正，也不代表 Windows／macOS 已重包到同一 source
checkpoint。

依使用者決定，完整版與含 MT-32 音樂的推廣片都只在本機保留，不加入 Git 或公開 release；
公開 patch 使用 v0.1.35 tag。現行 `0.1.51/0.1.57` 已於 `dist-all/v0.1.35/` 重包；Android host audio／真機、macOS 真機與全遊戲 V3 仍是
可選驗收，不得誤寫成已通過。

本輪依使用者指示不以完整回歸作完成條件。若沒有新的玩家可見差異、正式發行需求或更強
原版證據，不重新開啟已完成切片，也不把可選 V3 長尾改列為功能 blocker。
