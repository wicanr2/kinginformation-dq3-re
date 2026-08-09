# DQ3 地表命令窗字模資料化 ledger

> 2026-08-09。這是玩家可見文字的資料邊界切片；不改寫原始資產，也不把語意名稱
> 當成原版符號。字模索引只附加在 game pack JSON，Go 引擎保留輸入與繪製流程。

## 輸入與證據

| 欄位 | 值 |
|---|---|
| 原始字模來源 | `assets_raw/D3TXT00.FON` |
| 來源大小／SHA-256 | `47232` bytes／`c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02` |
| 位址基準 | `none`；這批是 FON 字模索引，不是 EXE 位址 |
| 主要證據 | `docs/data/glyph_unicode_map.json`、原版 DOSBox 命令窗畫面 |
| 推論等級 | `D2`：索引與玩家可見字形／既有命令流程相符；尚未逐一匯出原版 label writer sidecar |

## 對映

| role | primary／secondary glyph | 字形 | inference |
|---|---:|---|---|
| `title` | `274／664` | 命／令 | D2 |
| `talk` | `443／449` | 對／話 | D2 |
| `spell` | `429／430` | 咒／文 | D2 |
| `status` | `665／666` | 狀／況 | D2 |
| `item` | `402／237` | 道／具 | D2 |
| `equip` | `197／409` | 裝／備 | D2 |
| `examine` | `544／545` | 調／查 | D2 |

資料位於 `dq3_remake_ebitan/internal/gamepack/packs/dq3_cht/data/interface.json` 的
`field_command_labels`。`cmdmenu.go` 不再持有這組版本專屬字模；`NewGameWithPack`
缺少契約會 fail closed，直接 unit fixture 可省略但不能代表 production fallback。

## 限制與後續

這批只證實字模資料與現行命令窗可見結果，未把 `sub_1C03F` 等原版 writer 的每個
位址／raw bytes 匯出，因此不升級為 D3。若日後變更字形，必須保留舊斷言與反證，補上
原始 writer→consumer sidecar 及同狀態 DOSBox 對拍。
