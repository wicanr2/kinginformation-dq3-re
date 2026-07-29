# 86 — 諾亞尼爾：夢幻紅寶石與覺醒粉 production trace

> 狀態：2026-07-29，D3／E3；玩家可見畫面目前 V2，仍待 DOSBox 同狀態逐像素 V3。

## 結論

從羅馬利亞辭位後的合法 checkpoint 起，正式 `InputState` 路線已完成：

```text
羅馬利亞離城 → CTY04 沉睡村 → CTY05 精靈女王線索
→ CTY11 sec3 調查取得 0x59 → CTY05 handler16 原地換成 0x5a
→ CTY04 正式道具選單使用 → 村民甦醒 → save/load → 正式離村
```

本切片發現並移除一個會讓驗收假通過的錯誤：舊 Go 特例寫
`g.flags[0x31] = true`，但 NPC visibility 使用的是另一套原版 `storyBits`；而且
`0x31` 新遊戲本來就為 set。舊 trace 只檢查 `storyFlag(0x31)`，因此沒有證明覺醒粉
造成任何狀態改變。

## 原版閉合證據

### CTY11 夢幻紅寶石

`CTY11.DAT` sec3 的 examine table subid0 原始四 bytes：

```text
03 59 00 2f
```

即 type3、item `0x59`、present flag `0x2f`。原版語意為 flag set 時可取，取得後 clear。

### CTY05 精靈女王

`CTY05.DAT` sec0 `(17,7)` 是 subtype2、raw handler16。`DQ3.EXE`：

- file `0x68f7`：selected item = `0x59`；
- file `0x690a`：`mov word ptr [si], 0x5a`，直接取代原背包格；
- 缺道具播放 `D3TXT02` rec90，成功播放 rec96。

因此不是「刪除後 append」，也沒有第二份 milestone 或合理預設。

### CTY04 覺醒粉

`DQ3.EXE` file `0x5320..0x5365`：

1. `[0x256c] == 4`：只檢查目前 CTY，沒有 section、座標或 facing gate；
2. file `0x5330` 呼叫通用 selected-item consumer；
3. `SET 0x26`；
4. `CLEAR 0x31`；
5. `DI=0x257`，即 `D3TXT00` rec599「啊!村莊的人們醒來了!」；
6. file `0x5347..0x5353` 呼叫場景／地圖重載鏈。

`CTY04.DAT` sec0 同時含 11 筆 flag `0x26` 的清醒居民與 11 筆 flag `0x31` 的沉睡替身。
新遊戲 story bits 為 0–39 clear、40 起 set，故事件前只顯示 `0x31`；成功交易後反轉為
`0x26`。這也證明舊「設 0x31 表示甦醒」完全顛倒。

## game-pack 與引擎邊界

版本專屬設定位於 `data/events.json` 的 `quest_item_chain_events`：

- treasure CTY／section／subid／type／item／present flag；
- exchange NPC selector、required/granted item、before/success text ID；
- location-use CTY、set/clear flags、success text ID、scene reload。

Go 只保留有限的共用 primitive：

```text
一次性 treasure → 原背包格換物 → town-gated item use → flag mutation → scene reload
```

`itemuse.KindOf(0x5a)`、`scriptedTable` handler16 row 與 baked treasure row 均已移除；
缺 selector／文字／引用時失敗即關閉，不設 Go fallback。

## checkpoint 修正

由羅馬利亞存檔繼續時另發現 Go save 只保存城內座標，沒有保存進城前的
`overPx/overPy`。讀檔離城因此錯回阿里阿罕附近，而不是羅馬利亞地表入口。存檔格式新增
`over_px`、`over_py` 與明確的 `overworld_position_v2` marker；新版精確 round-trip，
舊城內存檔則由原版 CTY 入口表遷移，沒有用零值猜測。正式 trace 已由此 checkpoint
正常北行至 CTY04，並在甦醒後再次由標題讀檔離村。

## 驗收

- `TestDQ3NoanielQuestItemChainMatchesOriginalEXECTYAndText`
  逐 byte／word 對 `DQ3.EXE`、CTY04／05／11 與 D3TXT00／02。
- 四個 component tests 鎖定 present flag、原地換物、位置失敗不消耗、成功旗標與重載。
- `TestOpeningProductionInputTrace` 從 boot 延伸至整條事件，事件後 save/load，讀檔仍可
  正式由 CTY04 邊界離村。
- runtime 圖：
  `dq3_remake_ebitan/docs/noaniel_sleeping.png`、
  `dq3_remake_ebitan/docs/fairy_queen_ruby_exchange.png`、
  `dq3_remake_ebitan/docs/noaniel_awakened.png`。

下一個 campaign audit 從甦醒後離村 checkpoint 前往金字塔，定位按鈕／魔法鑰匙流程
第一個玩家 blocker。
