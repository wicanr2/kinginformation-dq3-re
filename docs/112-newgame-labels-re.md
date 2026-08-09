# 112 — 開場／創角字模資料化證據

## 結論

`interface.json.new_game_labels` 現在保存精訊版開場主選單、性別、能力確認與姓名輸入
畫面（rec451–456）的原始 glyph index。`NewGameFlow`、`NameInput`、`GenderSelect` 與酒館共用
同一份受驗證 game-pack 資料；production `NewGameWithPack` 缺欄位即 fail closed，Go
renderer 不再保存這批版本專屬字模常數。

本批的 glyph stream 已與 `new_game_geometry` 一起接入正式 renderer；主選單、姓名盤與能力
確認面板的座標／尺寸證據另見 [`docs/113`](113-newgame-geometry-re.md)。目前尚未閉合的是
EGA 框線 pattern、逐幀游標／閃爍與同狀態演出，不把單張 PNG 誤稱為完整 V3 parity。

## 資料契約

| role | glyph index | 畫面用途 |
|---|---:|---|
| `title` | `106,187,207,710,179` | 開場標題 |
| `start` | `113,689,488,711` | 遊戲開始 |
| `load` | `712,519,696,283` | 載入進度 |
| `male` / `female` | `775,674` / `234,674` | 性別選項／能力確認性別 |
| `level` / `hp` / `mp` | `419,420` / `22,30` / `27,30` | 能力確認左欄 |
| `agility` | `282,283` | 詳細狀況窗的「速度」欄；不是開場確認右欄第一列 |
| `luck` / `max_hp` / `max_mp` | `276,426,425,427` / `291,163,22,30` / `291,163,27,30` | 開場能力確認右欄的「運氣點數／最大HP／最大MP」 |
| `attack` / `defense` / `experience` | `623,624` / `340,409` / `525,526` | 開場能力確認右欄的「攻擊力／守備力／經驗」 |
| `sex` / `hero` / `cloth` | `674,417` / `106,187` / `190,149,191,192` | 能力確認固定標籤 |
| `prompt` | `398,546,194,229,456,534` | 「這個人可以嗎」 |
| `yes` / `no` | `399` / `678` | 能力確認選項 |
| `name_title` | `690,519,691,692` | `輸入姓名` 標題（rec451） |
| `name_left_arrow` / `name_right_arrow` | `693,693` / `694,694` | 姓名列左右雙箭頭（rec451） |
| `name_zhuyin_grid` | 45 格 | 注音盤逐列 glyph（rec452，含特殊格） |
| `name_alnum_grid` | 45 格 | 英數／符號盤逐列 glyph（rec453） |
| `function_zhuyin` / `function_alnum` | 5×glyph rows | 右側功能列（rec452／453） |
| `input_mode_zhuyin` | `690,519,702,313` | 注音組字提示（rec456） |
| `backspace` / `ok` | `11` / `29,25` | 舊契約保留；正式 rec453 使用完整 45 格，不把符號格誤當退格 |

## 證據與限制

- 輸入字型：`D3TXT00.FON`，47232 bytes，SHA-256
  `c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`。
- `assets_raw/D3TXT00.TXT` SHA-256 為
  `38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b`；
  `docs/data/d3txt_codes.json` 的 D3TXT00 records 451–456 與原版開場／創角實機畫面交叉核對上述 index；component test
  `TestDQ3NewGameLabelsMatchOriginalGlyphs` 會檢查字型大小、雜湊與完整 role 對映。
- 目前推論等級為 **D2**：字模資料與玩家可見畫面已對齊，但本批尚未另行匯出每個
  glyph 的 IDA writer→consumer sidecar，因此不宣稱 D3。
- `docs/dosbox/v3_01_afterstats.png` 的右欄逐 glyph 解碼修正了先前把「運氣點數／最大HP／最大MP」
  誤標成「速度／HP／MP」的欄位。`agility` 保留為詳細狀況窗的獨立 role；開場確認 renderer
  現在只使用 `luck`、`max_hp`、`max_mp`、`attack`、`defense`、`experience` 六列。
- 本次只閉合欄位語意與 JSON→renderer 對映；目前外框仍是簡化 1px 白框，原版 EGA
  pattern／藍色選項框與同輸入逐像素對拍仍是 V3 缺口。
- Docker＋Xvfb 重新產出的 `dq3_remake_ebitan/docs/ng_confirm.png` SHA-256 為
  `da3507180fbc598ce005edbbd386c54a74f70c7dcf00dac6b2fe4a3e302ba7c2`；它只證明目前
  runtime 可重現，不取代原版同輸入 V3 對拍。
- 此欄位只描述字模，不描述文字控制碼或任意 JSON 程式碼。姓名輸入的 raw 45 格→欄優先
  cell 換算、功能列焦點與完成 gate 仍由共用狀態機處理；畫面幾何由
  `new_game_geometry` 契約提供。

原始 glyph index、pack 版本與 evidence 由
[`docs/84-game-pack-json-contract.md`](84-game-pack-json-contract.md) 的 interface 契約
統一管理；不可把本表複製回 Go 常數或以攻略／C remake 改寫。
