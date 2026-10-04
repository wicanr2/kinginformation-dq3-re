# DQ3 Go／Ebitengine 現行工作清單

更新：2026-10-05。遠端 Issue 是本輪工作的權威；唯一詳細計畫仍是 [`docs/74`](docs/74-ebiten-remake-completion-plan.md)；
本檔只保存可快速接手的目前順序，不收錄歷史 C/SDL 工作。

| 本輪工作 | 狀態與驗證界線 | 權威入口 |
|---|---|---|
| 原版新遊戲／創角／母親開場 | 進行中：正常指令窗、導覽及HUD有限CONFORMED，十二張完整RGB0，正常存讀檔與下一步通過。下一步道具206七列／父窗／action，RGB差44816、仍DRAFT；現況只查CONTEXT | [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)、[docs/188](docs/188-opening-escort-to-castle-spec.md) |
| 原版六幕開場 | 已完成本輪限定色號／RGB 對拍，其他相位與音訊未知 | [Issue #1（已關閉）](https://github.com/wicanr2/kinginformation-dq3-re/issues/1)、[docs/196](docs/196-dosgolem-opening-sequence-parity.md) |
| 正式玩家路線回歸 | 本輪完整game484覆蓋、433不同頂層／119子PASS、51選用SKIP；internal171／375及11套件PASS、4選用SKIP。正常THE END190.25秒、Linux desktop PASS，必驗指令窗零SKIP；原版完整動態對拍未知 | [Issue #2](https://github.com/wicanr2/kinginformation-dq3-re/issues/2)、[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4) |
| 同伴持冠還冠持有權檢查 | 已由 fc78bb5 推送，原版局部靜態閉環、元件、正式還冠／辭位及存讀檔通過；原版動態路線待 dosgolem | [Issue #3（已關閉）](https://github.com/wicanr2/kinginformation-dq3-re/issues/3)、[docs/82](docs/82-romaly-king-production-trace.md) |

下方為 2026-08-26 功能與交付歷史，不能覆蓋上表的現行驗收，也不限制本輪使用者授權的對拍目標。

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
| remake 玩家流程 | 六項功能皆已接入；開場轉場後自動行走已依 [`docs/188`](docs/188-opening-escort-to-castle-spec.md) 改為兩階段 pack 序列，rec80→rec79、逐字顯示與王座勇者像素檢查均通過 | 否；完整 trace 後段仍會在既有 Lv20 練級上限失敗，與本切片無關 |
| 原版證據限制 | 藥草、祈禱之戒與聖水參數已由 [`docs/191`](docs/191-field-item-parameter-re.md) 訂正；目前限制是王座同頁原版畫格、開場淡入淡出、職業能力條逐格及部分音效仍只到 V1／V2／unknown | 否；文件與程式必須保留近似／未知標記，不得冒稱 exact |
| 交付 | `v0.1.36` 已由 commit `d6184f4` 建立 Linux x86_64 AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP 的公開 patch 與本機完整版；公開 patch 私有素材為 0 | 已完成；checksum、smoke 與限制見 [`docs/194`](docs/194-release-v0.1.36.md) |

2026-08-25 音訊抽樣：正式場景使用的七軌 OGG 已由 [`docs/193`](docs/193-audio-sampling-polish-20260825.md)
完成格式、時長、音量、長靜音、完整解碼與 SHA-256 驗收；沒有 clipping 或近似靜音。
field 軌的 3.112 秒尾段安靜區保留為編曲／loop-point 未知，不在沒有原版循環邊界時裁切。
跨平台人耳與尚未閉合完成／恢復時序的 EBG 事件 cue 是可選 V3，不是功能 blocker。

2026-08-26 音訊 runtime polish：戰鬥結束不再固定播放 field，而依結算後場景恢復
castle／town／dungeon／field；音樂 OFF→ON 亦立即恢復當前 title／battle／scene／ending cue。
具名路由測試與七軌重驗通過，見 [`docs/195`](docs/195-audio-transition-runtime-polish.md)。
音訊維度自評由 78% 提升至 84%，總分為 90.25%（對外仍取整 90%）。commit `1dc5546`
已重包為 `dist-all/v0.1.38-local/full/` 的四個本機完整版；尚未建立新的公開 release。

2026-08-25 勘誤：[`docs/192`](docs/192-opening-dialogue-v3-closure.md) 已證實先前 rec80
「小型城堡／錯誤背景」是 BLK 位平面解碼錯誤，修正後同狀態背景與 rec80→rec79 已閉合；
record 78 的勇者及角色關係亦已由像素差異測試確認。王座同頁原版畫格、開場淡入淡出與
職業能力條逐格仍缺足夠 oracle，維持 V2／unknown，不反向推翻 campaign E3，也不使用
「全畫面 V3」措辭。

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
公開 patch 使用 v0.1.36 tag。現行 schema `0.1.53`／content `0.1.59` 已於
`dist-all/v0.1.36/` 完成四平台架構檔（Linux、Windows、兩種 macOS 架構）的 patch／full
同 checkpoint 封裝，公開 release 只含四個 patch。Android host audio／真機、macOS 真機與全遊戲 V3 仍是
可選驗收，不得誤寫成已通過。

本輪依使用者指示不以完整回歸作完成條件。若沒有新的玩家可見差異、正式發行需求或更強
原版證據，不重新開啟已完成切片，也不把可選 V3 長尾改列為功能 blocker。
