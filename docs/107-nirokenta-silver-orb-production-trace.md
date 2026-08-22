# 尼羅肯特銀寶珠 handler49：逆向證據與正式交易

> 本文件曾被其他現行文件引用但實體遺失。2026-08-22 由相同原始檔重新以
> IDA Pro 9.4 匯出，不以舊摘要冒充證據；重建原因與容量訂正見 `docs/172`。

## 原始輸入與工具

- `assets_raw/DQ3.EXE`：115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/CTY64.DAT`：2,035 bytes，SHA-256
  `6ab43f6173c550aeb0187d259b5bc7e040aa406bd2fefacd8b1890ba1eaf3d9f`。
- IDA Pro 9.4；位址基準為 IDA linear，`logical=linear-0x10000`、
  `file=linear-0xEC90`。
- 匯出器：`tools/ida_dump_nirokenta_silver_orb.py`；gitignored sidecar：
  `work/nirokenta-silver-orb-ida-20260822.json`。
- 原始函式名、位址、bytes、operand 與 xref 均保留；下列語意是附加註解，
  沒有修改 EXE、CTY 或把 rename 當成證據。

## handler49 交易（confirmed）

`sub_15D70`（IDA linear `0x15D70..0x15DAF`）是 DGROUP handler table 的
handler49 consumer：

1. `0x15D73..0x15D79` 令 `BX=0x4B` 並呼叫 `sub_16F09` 測 story flag `0x4B`。
2. flag 未 set 時，`0x15D7D` 以 `DI=0x0C0D` 顯示 repeat text，然後返回。
3. flag set 時，`0x15D88` 以 `DI=0x0C0C` 顯示成功 text。
4. `0x15D90` 把 `DS:2593` 設為 item `0x6B`，`0x15D96` 呼叫 `sub_1684E`。
5. `0x15D99..0x15DA9` 檢查 `DS:0726`：只有 storage 成功才以
   `sub_16EF4` 清 flag `0x4B`；容量失敗走 `sub_15037`，不清 flag。

因此「顯示成功對話」不是獎勵已提交的證據；item writer 成功與 flag clear 才是
不可分割的 transaction。

## 共用全隊八格 writer（confirmed）

`sub_1684E` 先播放／等待取得效果後落入 `sub_16856`。`sub_16856`
（IDA linear `0x16856..0x16898`）：

- 從 `DS:5077` 取得 active party 人數，`DS:259C=1` 起依 party pointer table
  `DS:4F15` 順序掃描。
- 每名角色從 `+0x3A` 起最多比較八個 word，空槽 sentinel 為 `0x00FF`。
- 第一個空槽把 `DS:0726` 清為 0，再把 `DS:2593` 的 `0x6B` 寫入該槽。
- 全隊皆滿時把 `DS:0726` 設為 1 並返回，沒有越界追加、沒有覆蓋既有物品。

`sub_1684E` 的直接 caller xref 包含 `0x15D96`，所以 handler49 與這個 writer 是
同一條可回查鏈，不是由其他獎勵事件類推。共用 writer 的其他 caller 見 `docs/161`。

## CTY selector 與 remake 對應

- 正式 CTY64 loader 解析出的 section0 NPC 位於 `(16,10)`、subtype2、
  `handler_raw=49`；production `talkNPCItemReward` 同時比對 CTY、section、tile 與
  handler，任一不符均不接管。
- game pack event `dq3:event.nirokenta_silver_orb` 保存 flag `75`、item `107`
  與成功／事後 text ID；玩家文字仍由 `texts.json` 提供，Go 不含顯示句子。
- runtime 必須使用 `grantPartyItem`。成功才 clear present flag；全隊滿格時保持 flag、
  不給 item，讓玩家可騰出一格後再交談。

上述 selector 的 CTY raw identity 與正常玩家交談屬 D3；視窗逐 pixel、取得動畫 frame
與 PCM wall-clock 尚未做同狀態原版對拍，故不宣稱 V3。

## 驗收契約

- 主角滿格但同伴有空格：依 party order 給第一個可用同伴。
- 全隊每人含裝備皆八格：不給 item、不清 flag；騰出一格後重談可成功。
- 成功後重談：只顯示 after text，不再給第二顆。
- 成功狀態經 save/load：item ownership 與 flag clear 均保留。
- boot 起正式 trace 必須由尼羅肯特洞窟抵達 CTY64，再以正常 NPC 交談取得；
  direct handler call、直接給 item 或直接清 flag 不算 E3。
