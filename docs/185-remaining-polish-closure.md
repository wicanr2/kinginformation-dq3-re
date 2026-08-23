# 185 — 四項殘餘 polish 靜態閉合（2026-08-23）

## 範圍與輸入

本切片只閉合 record345 姓名插值、蘭西爾完成旗標、幻惑／封咒存續，以及 D3MNS
`+0x27`。不擴張為全域文字控制碼、全旗標、全部咒文或整張怪物結構重命名。

- 輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4，image `ida-pro-9.4-ver3:py312-x11-v4`。
- 位址：IDA linear；`logical=linear-0x10000`、`file=logical+0x1370`。
- 可重建 exporter：`tools/ida_dump_remaining_polish.py`。原始 EXE 唯讀；暫存 `.i64` 與
  sidecar 只在 `/tmp`，原始名稱、bytes、operand 與 xref 均保留。

## 1. record345 的 `0xFFFB` 是隊長姓名（confirmed）

`sub_16856` 從 `DS:259C=1` 開始按隊伍順序掃每人八格；全滿時它會暫時留下
`party_count+1`。但真正顯示端 `sub_1C425` 在 `0x1C554` 明確執行
`mov word ptr ds:259Ch,1`，再以 `DI=0x159` 呼叫文字 renderer。`sub_21651` 對 selector
非 0／7 的路徑執行 `(selector-1)*2 → DS:4F15`，再取角色 `+3` 姓名。因此 record345
的 `0xFFFB` owner 是 active slot 1，也就是隊長。

remake 已新增 `common:text.battle.drop.inventory_full`，全隊滿格時不寫物品、仍結算
EXP／gold，並以 pack 文字顯示隊長姓名。glyph words 與 D3TXT00 record345 完全一致。

## 2. 蘭西爾 flag `0x13` 沒有 gameplay writer（confirmed negative）

原版讀端仍成立：handler37 `0x159E4` 與 world loader `0x12691` 都以 `BX=0x13` 呼叫
`sub_16F09`。本輪重新列舉 `sub_16EDF`（SET）與 `sub_16EF4`（CLEAR）的所有 direct
callers，並保留每個 caller 前十筆原始指令；沒有任何 setter caller 的 BX provenance
為 `0x13`。既有 handler37、handler62、洞窟 handlers85／86 與 CTY23 寶箱也都沒有 writer。

因此本未發售 binary 的 completed 對話／入口分支是可讀但不可由正常流程寫入的殘留分支。
remake 正確行為是保留 reader、絕不在接受試煉／取藍寶珠／復隊時合成 writer。

## 3. 幻惑／封咒沒有我方全隊中途解除狀態（confirmed removal）

現行 D3MNS 39/39 action ledger 已閉合原版 selector `sub_199DC` 與全部 production
definition；其中沒有「敵施 rec156／158 → 我方全隊 blind／sealed」action。舊 Go 的
`partyBlind`／`partySealed` 在修正前只有 start reset 與測試直接寫入，production 沒有
writer，來源只是早期 C prototype 的近似。

因此問題不是「原版何時中途解除」，而是這兩個我方全隊狀態不應存在。已移除欄位、玩家
物攻 miss consumer、玩家咒文封鎖 consumer及對應 synthetic tests。玩家正式施放瑪荷頓
（156）／瑪努莎（158）仍寫目標 `enemyUnit.status`；pack 沒有回合 timer 或 cure consumer，
故這兩個敵方個體狀態存續到本場結束，再隨 battle instance 銷毀。

## 4. D3MNS `+0x27` 是本 EXE 未使用／保留欄（strong）

IDA 對全部已分析 code 掃描 `DS:0xD78..0xDA0` direct displacement：HP、MP、AI、抗性、
EXP、gold、`+0x25` 掉落 threshold、`+0x26` item 與 `+0x28` encounter weight 均有命中；
唯一沒有 direct operand 的尾端欄位是 `DS:0xD9F`（record `+0x27`）。既有 record loader／
consumer 也沒有 address-taken 後再加 `0x27` 的資料流。

「本 EXE 沒有已分析 direct consumer」為 confirmed；「設計用途是 padding／保留欄」仍只到
strong。remake 保留 raw `Unknown27` 供 round-trip，不命名、不消費、不放進 game-pack 規則；
這已足以關閉 remake polish，不再把它列成玩法缺口。

## 驗收與停止線

- pack parity：record345 glyph words、來源 record 與 evidence 必須通過 validation。
- component：全隊滿格不 mutation、仍結算 gold，且開啟 record345；156／158 只影響敵方個體。
- source audit：不得再出現 production `partyBlind`／`partySealed` 或「flag0x13 writer 待實作」。
- 沒有新的 binary 或玩家可見反證時，不重新開啟這四項；`+0x27` 不得改名成掉落率或
  boss repeat count。
