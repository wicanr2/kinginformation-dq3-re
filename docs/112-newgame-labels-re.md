# 112 — 開場／創角字模資料化證據

## 結論

`interface.json.new_game_labels` 現在保存精訊版開場主選單、性別、能力確認與姓名輸入
特殊格所需的原始 glyph index。`NewGameFlow`、`NameInput`、`GenderSelect` 與酒館共用
同一份受驗證 game-pack 資料；production `NewGameWithPack` 缺欄位即 fail closed，Go
renderer 不再保存這批版本專屬字模常數。

本批只資料化 glyph stream。主選單、姓名盤與能力確認面板的座標／尺寸仍是另一個待閉合
的 V3 geometry slice；移入 JSON 不代表目前畫面已與原版逐像素一致。

## 資料契約

| role | glyph index | 畫面用途 |
|---|---:|---|
| `title` | `106,187,207,710,179` | 開場標題 |
| `start` | `113,689,488,711` | 遊戲開始 |
| `load` | `712,519,696,283` | 載入進度 |
| `male` / `female` | `775,674` / `234,674` | 性別選項／能力確認性別 |
| `level` / `hp` / `mp` | `419,420` / `22,30` / `27,30` | 能力確認左欄 |
| `agility` / `attack` / `defense` / `experience` | `282,283` / `623,624` / `340,409` / `525,526` | 能力確認右欄 |
| `sex` / `hero` / `cloth` | `674,417` / `106,187` / `190,149,191,192` | 能力確認固定標籤 |
| `prompt` | `398,546,194,229,456,534` | 「這個人可以嗎」 |
| `yes` / `no` | `399` / `678` | 能力確認選項 |
| `backspace` / `ok` | `11` / `29,25` | 英數姓名盤特殊格 |

## 證據與限制

- 輸入字型：`D3TXT00.FON`，47232 bytes，SHA-256
  `c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`。
- `glyph_unicode_map.json` 與原版開場／創角實機畫面交叉核對上述 index；component test
  `TestDQ3NewGameLabelsMatchOriginalGlyphs` 會檢查字型大小、雜湊與完整 role 對映。
- 目前推論等級為 **D2**：字模資料與玩家可見畫面已對齊，但本批尚未另行匯出每個
  glyph 的 IDA writer→consumer sidecar，因此不宣稱 D3。
- 此欄位只描述字模，不描述文字控制碼、輸入規則或任意 JSON 程式碼。姓名輸入的注音
  組字資料仍由共用 `zhuyin` 元件處理；畫面幾何待後續 `new_game_geometry` 契約閉合。

原始 glyph index、pack 版本與 evidence 由
[`docs/84-game-pack-json-contract.md`](84-game-pack-json-contract.md) 的 interface 契約
統一管理；不可把本表複製回 Go 常數或以攻略／C remake 改寫。
