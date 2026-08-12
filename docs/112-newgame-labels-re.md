# 112 — 開場／創角字模資料化證據

## 結論

`interface.json.new_game_labels` 現在保存精訊版開場主選單、性別、能力確認（record 407）與
姓名輸入畫面（rec451–456）的原始 glyph index。`NewGameFlow`、`NameInput`、`GenderSelect` 與酒館共用
同一份受驗證 game-pack 資料；production `NewGameWithPack` 缺欄位即 fail closed，Go
renderer 不再保存這批版本專屬字模常數。

本批的 glyph stream 已與 `new_game_geometry` 一起接入正式 renderer；主選單、姓名盤與能力
確認面板的座標／尺寸證據另見 [`docs/113`](113-newgame-geometry-re.md)。能力確認的固定同狀態
畫面已另達 V3 靜態對拍；逐幀游標／閃爍與整段演出仍未閉合，詳見
[`docs/126`](126-newgame-confirmation-v3-static-comparison.md)。

## 資料契約

| role | glyph index | 畫面用途 |
|---|---:|---|
| `title` | `106,187,207,710,179` | 開場標題 |
| `start` | `113,689,488,711` | 遊戲開始 |
| `load` | `712,519,696,283` | 載入進度 |
| `male` / `female` | `775,674` / `234,674` | 性別選項／能力確認性別 |
| `level` / `hp` / `mp` | `12,419,12,420,667` / `12,12,22,30,667` / `12,12,27,30,667` | record 407 的左欄；前導空格與冒號也保留 |
| `strength` / `agility` / `vitality` / `intelligence` | `170,280,667` / `282,283,667` / `284,170,667` / `285,286,283,667` | 確認頁右欄的力量／速度／耐力／聰明度 |
| `luck` / `max_hp` / `max_mp` | `276,426,425,427,667` / `291,163,22,30,667` / `291,163,27,30,667` | 確認頁右欄的運氣點數／最大HP／最大MP |
| `attack` / `defense` / `experience` | `623,624,170,667` / `340,409,170,667` / `525,526,667` | 確認頁右欄的攻擊力／守備力／經驗 |
| `sex` / `hero` / `cloth` | `12,674,12,417,667` / `106,187` / `190,149,191,192` | 能力確認固定標籤 |
| `prompt` | `398,546,194,229,456,534,58` | 「這個人可以嗎？」；末尾 `58` 是問號 |
| `yes` / `no` / `choice_cursor` | `399` / `678` / `11` | 能力確認選項與原版游標 |
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
  `docs/data/d3txt_codes.json` 的 D3TXT00 record 407 與 records 451–456 和原版開場／創角實機畫面交叉核對上述 index；component test
  `TestDQ3NewGameLabelsMatchOriginalGlyphs` 會檢查字型大小、雜湊與完整 role 對映。
- 目前推論等級為 **D2**：字模資料與玩家可見畫面已對齊，但本批尚未另行匯出每個
  glyph 的 IDA writer→consumer sidecar，因此不宣稱 D3。
- `sub_10854 → sub_1834E` 與 record 407 的逐 glyph 解碼確認開場確認頁包含力量、速度、耐力、
  聰明度、運氣點數、最大HP、最大MP、攻擊力、守備力與經驗等十列；`agility` 在此頁是速度，
  也可在詳細狀況窗重用。舊有「只用六列」描述已推翻。
- 固定確認畫面已採 `beveled_2px` 大面板、專用 choice backdrop／cursor、原始 palette 與
  同輸入 V3 靜態對拍；它仍不包含逐幀游標、背景切換或音效時序。
- Docker＋Xvfb 重新產出的 `dq3_remake_ebitan/docs/ng_confirm.png` SHA-256 為
  `da3507180fbc598ce005edbbd386c54a74f70c7dcf00dac6b2fe4a3e302ba7c2`；它只證明目前
  runtime 可重現；最終同輸入 V3 靜態對拍的 hash／AE 以 `docs/126` 為準。
- 此欄位只描述字模，不描述文字控制碼或任意 JSON 程式碼。姓名輸入的 raw 45 格→欄優先
  cell 換算、功能列焦點與完成 gate 仍由共用狀態機處理；畫面幾何由
  `new_game_geometry` 契約提供。

原始 glyph index、pack 版本與 evidence 由
[`docs/84-game-pack-json-contract.md`](84-game-pack-json-contract.md) 的 interface 契約
統一管理；不可把本表複製回 Go 常數或以攻略／C remake 改寫。
