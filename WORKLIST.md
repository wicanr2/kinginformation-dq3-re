# DQ3 Go／Ebitengine 現行工作清單

更新：2026-08-23。唯一詳細計畫仍是 [`docs/74`](docs/74-ebiten-remake-completion-plan.md)；
本檔只保存可快速接手的目前順序，不收錄歷史 C/SDL 工作。

| 順序 | 工作 | 狀態 | 直接證據 |
|---:|---|---|---|
| 1 | 戰鬥道具 | 完成（D3／E2） | [`docs/179`](docs/179-battle-item-selector-runtime-spec.md) |
| 2 | 原始怪物表實際使用的 action | 完成（39/39 definitions，D2／E2） | [`docs/180`](docs/180-monster-action-runtime-spec.md) |
| 3 | 剩餘合法野外咒文 | 完成（D3／E2） | [`docs/181`](docs/181-field-support-spells-runtime-spec.md) |
| 4 | 商店賣出／售價／裝備交易 | **進行中：下一項** | `docs/156` 與商店 caller |
| 5 | 船直接進城／離船／出城復船 | 待辦 | world transition caller |
| 6 | 必要選單與玩家可見支線 | 待盤點後逐項閉合 | 只接受正常玩家路徑 discrepancy |

不屬於功能完成 gate：全遊戲逐畫面 V3、DOS PCM/DAC/PIT wall-clock 深挖、未被原始資料
使用的 helper/action。這些只能由新的玩家可見差異或正式發行需求重新開啟。
