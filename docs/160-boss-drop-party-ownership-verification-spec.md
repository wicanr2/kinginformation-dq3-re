# 160 — 必經 Boss 掉落的隊伍 ownership 驗收規格（2026-08-22）

## 問題

`docs/157` 接回原版戰後掉落 writer 後，物品會寫入隊伍中第一個有空位的角色。八頭大蛇
第一戰前診斷為：勇者含裝備已使用 11 格，但全隊仍有 7 格空位。怪75必掉草薙大劍
`0x14` 後，舊 production assertion 仍用 hero-only `hasItem(0x14)`，把寫入同伴的成功掉落
誤報為遺失。

## 證據

- `docs/95`：怪75 `D3MNS +0x25=0xff`、`+0x26=0x14`；第一戰使用正常掉落，第二戰由
  `DGROUP 0x2518 bit1` 抑制。
- `docs/157`：IDA Pro 9.4 confirmed `sub_1C425 → sub_1684E/sub_16856`，依主角至同伴
  順序掃每人八格，寫第一個 `0x00ff`；全隊滿格才不寫。
- 原始輸入仍為 `DQ3.EXE` 115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；位址、bytes、工具與
  sidecar 契約見 `docs/157`。本切片不新增或改寫原版語意。

推論等級：怪75必掉、隊伍第一空位 writer 與第二戰抑制均為 **confirmed**；最新 trace
在戰前有 7 格隊伍空位是 remake runtime 直接觀測。hero-only assertion 不符合 writer，
屬已證實的驗收錯誤。

## 規格與驗收

1. 第一戰後以 `hasPartyItem(0x14)` 驗證取得，不要求道具必在勇者。
2. 紫寶珠在第二戰後寶箱前仍須以 `hasPartyItem(0x69)==false` 驗證，避免同樣 ownership
   假設造成假陽性／假陰性。
3. 不搬移掉落、不直接補道具、不改掉落 writer、怪物記錄或事件旗標。
4. 正式 trace 必須越過第一戰後自動轉場；後續第一個 blocker 另立 spec。
