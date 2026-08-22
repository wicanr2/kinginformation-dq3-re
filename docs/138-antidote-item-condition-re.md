# 驅毒草 item 0x42：IDA dispatcher、消耗與解毒 spec（2026-08-22）

> 歷史 blocker：本切片已完成，後續正式 trace 已由標題抵達 `THE END`；現況以 `docs/74`
> 最新 checkpoint 為準，不得把下述修正前缺口重新列為 current work。

## 問題與先前歧義

poison 已接成 battle→角色／save→移動→教會 E2，但正常道具入口仍讓驅毒草落入未知
fallback。歷史 C 文件又宣稱「沒有中毒不消耗」，沒有原版 handler 證據。這使勇氣試煉
單人回程若在洞窟中毒，只能賭遭遇 RNG，且文件會把 C prototype 當成原版規格。

## 輸入、工具與位址口徑

- 輸入：`assets_raw/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4＋IDAPython，Docker image `ida-pro-9.4-ver3:py312-x11-v2`。
- 原始檔唯讀；IDA database 在容器 `/tmp`，可重生 sidecar 為
  `work/antidote-item-ida.json`；腳本為 `tools/ida_dump_antidote_item.py`。
- IDA linear = logical + `0x10000`；file = logical + `0x1370`。所有 rename／名稱只供導覽，
  本輪沒有修改 database annotation 或原始 EXE。

## confirmed 原版鏈

1. item-use dispatcher logical `0x3ccf` 對 raw item `>=0x41` 做
   `(item-0x41)*2`，讀 DGROUP `0x366a` pointer table。
2. item `0x42` 讀 DGROUP `0x366c`／file `0x197ac`，raw word `0x3dc3`，進
   logical `0x3dc3`：

   ```text
   call sub_14CF9
   mov  dx, 0x40
   mov  bp, 0x16c
   call sub_1469F
   retn
   ```

3. `sub_14CF9` 由目前物品持有者取角色 pointer，掃 `+0x3a` 起八個 item words，命中
   selected raw item 後立即寫 `0x00ff`。因此驅毒草在效果判定前就消耗；「無中毒不消耗」
   是已推翻的 C prototype 斷言。
4. `sub_1469F` 以已選目標 `DGROUP 0x0741` 取角色：
   - 死亡 `+0x38 bit0x80`：顯示 rec341，已消耗 item，不改狀態；
   - 未死亡但 `+0x38 bit0x40` 未設：顯示 rec341，已消耗 item；
   - 中毒：`AND +0x38, ~0x40`，更新狀態 UI，顯示 `bp=0x16c` 即 rec364，回 AL=1。
5. D3TXT00 原始 glyph：rec341
   `[508,399,435,436,494,410,147,431,273,56]`；rec364
   `[65531,486,638,149,137,464,423,56]`。此處採 decoder／game-pack 契約，已去除
   原始 record 終止碼 `0xffff`；原始 bytes 仍由 parity test 回讀。

推論等級：**confirmed**（dispatcher→raw table→handler→inventory writer→status
reader/writer→玩家文字 consumer 已閉合）。

## Remake spec

- game-pack `item_use_effects` 新增有限 `clear_condition` primitive：
  `condition_id=common:condition.poison`、`target_scope=party_member`、`consume=true`，並引用
  success／no-effect text ID；raw item `0x42` 只存在 DQ3 pack。
- 引擎在玩家選「使用」後進通用 party-member target modal，包含主角與同伴；取消選人前不
  消耗。確認目標後依原版順序先移除 item，再判斷死亡／condition；沒有作用也不退還。
- 成功只清指定成員 poison，不碰其他 condition。畫面文字由 pack text ID 開啟，不在 Go
  寫句子或 DQ3 record。
- schema/reference validation 必須拒絕未知 condition、非 party-member scope、未消耗與缺
  success/no-effect text。EXE table word、handler bytes、文字 glyph 需有 parity tests。
- 正式 trace 在 CTY23 取得藍寶珠後，若主角中毒且仍持驅毒草，必須經正常道具清單→使用
  →選主角解除，再由正式方向輸入返回 CTY75；不得直接清 condition。

## 停止線與未決

- **後續勘誤（2026-08-22）：**本文件當時未解滿月草 `0x45`；`docs/152` 已獨立閉合
  持久角色模型、怪物 action3 writer、戰鬥命令 gate、場景 40 步 consumer 與 field item
  handler。此歷史停止線不得再當成現行 pending；仍未閉合的是戰鬥道具清單／選擇器。
- 現行全域 inventory 與原版 per-character 八格持有者仍不完全相同；本切片只要求使用目標
  精確，物品持有者 ownership 是既有架構差距，不能由本 handler 自動宣稱已閉合。
- 成功／失敗訊息的同狀態 V3 畫面尚未取得；本切片最高為 E2／V1。
