# DQ3 Go／Ebitengine 現行工作清單

更新：2026-08-23。唯一詳細計畫仍是 [`docs/74`](docs/74-ebiten-remake-completion-plan.md)；
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
| remake 玩家流程 | campaign E3；上述六項皆已接入正式入口 | 否，已完成 |
| 原版證據限制 | 藥草治療量、祈禱之戒 MP 回復量、聖水步數仍是 classic／舊 C 近似；部分逐窗、逐幀與音效只到 V1／V2／unknown | 否；文件與程式必須保留近似／未知標記，不得冒稱 exact |
| 交付 | `dist-all/v0.1.35-local/` 已含四個完整版與有音樂推廣片；公開版仍是 v0.1.34 | 已完成本機交付；依使用者決定不建立 tag／GitHub Release |

2026-08-23 polish：record345 隊長姓名、蘭西爾 `0x13` 無 writer 的否定性證據、舊 C
`partyBlind/partySealed` 死狀態，以及 D3MNS `+0x27` 未使用 raw 欄均已由 [`docs/185`](docs/185-remaining-polish-closure.md)
閉合，不再列為 remake 待辦。

## 本機交付（已完成）

2026-08-23 已完成：

1. `dist-all/v0.1.35-local/full/`：Linux x86_64 AppImage、Windows x86_64 ZIP、macOS
   x86_64／arm64 ZIP，均為本機完整版。
2. `dist-all/v0.1.35-local/promo/`：72 秒有音樂推廣片、實機開場來源、contact sheet、
   FFprobe、音量／非靜音驗收及音樂來源雜湊。
3. 四包皆核對架構、CRC／解包、checkpoint、修正版 `DQ3MNS.SHP` 與 IDA sidecar 排除；Linux
   AppImage 另通過 Docker＋Xvfb 8 秒啟動。macOS 沒有真機驗收，維持靜態驗證標記。
4. `SHA256SUMS.txt` 與 `LOCAL-DELIVERY.txt` 已放在同一版本根目錄。

依使用者決定，完整版與含 MT-32 音樂的推廣片都只在本機保留，不加入 Git、不發布、不建立
新 tag。至此必要 worklist 為零；Android host audio／真機、macOS 真機與全遊戲 V3 仍是
可選驗收，不得誤寫成已通過。

本輪依使用者指示不以完整回歸作完成條件。若沒有新的玩家可見差異、正式發行需求或更強
原版證據，不重新開啟已完成切片，也不把可選 V3 長尾改列為功能 blocker。
