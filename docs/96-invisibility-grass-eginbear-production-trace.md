# 96 — 隱形草與愛丁貝亞守衛：IDA 證據與正式玩家流程

> 更新：2026-08-02（Asia/Taipei）
> 範圍：道具 `0x5d`、CTY38 原始貨架、CTY39 section 0 守衛 gate、CTY76 入口
> 結論：機制與設定資料已達 D3／E3；畫面達 V2，尚無同狀態原版截圖，不能標 V3。

## 1. 本輪修正的錯誤前提

舊 trace 把「朗錫爾村 CTY47／勇氣神殿態 CTY75」直接當成隱形草貨架入口，但正式船路
只能抵達 CTY47，且該入口的 NPC 連通區沒有可達的道具店。原始商店表盤點則明確指出
**CTY38 道具店貨架含 `0x5d`**（見 [`docs/40`](40-facility-shops.md)）。因此正式 trace
改由 CTY38 買入，不再讓地名線索取代 SHP.DAT／runtime 貨架 oracle。

這項修正不表示 CTY38 的地名已重新命名；它只證明目前 production 流程使用的貨架與原始
商店資料一致。CTY47／75 的朗錫爾雙入口與勇氣試煉已由後續切片閉合，見
[`docs/100`](100-lancel-courage-blue-orb-production-trace.md)；本文件不再以早期 gate 假說
描述該流程。

## 2. 非破壞性 IDA 證據

| 項目 | 內容 |
|---|---|
| 工具 | IDA Pro 9.4，Docker image `ida-pro-9.4-ver2:latest` |
| 原始輸入 | `assets_raw/DQ3.EXE`，唯讀掛載 |
| SHA-256 | `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| database | 一次性 `/tmp/dq3-invis-ida`；未加入 Git，未改原始檔 |
| 位址口徑 | IDA linear address、`(logical)`、`(file)` 與 `DGROUP` 明確分列 |

### 2.1 道具派發與消耗

`0x5d` 的 item pointer table entry 位於 `(file) 0x197e2`，word 為 `0x3fff`。handler
範圍為 `(logical) 0x3fff..0x4011`、`(file) 0x536f..0x5381`，原始 19 bytes：

```text
e8 f7 0c e8 3f a6 c6 06 f7 52 19 90 81 0e 46 4f 00 c0 c3
```

閉合資料流：

```text
道具選單 item 0x5d
  → pointer entry 0x3fff
  → call (logical) 0x4cf9 清除所選角色物品槽
  → write DGROUP 0x52f7 = 0x19
  → OR DGROUP 0x4f46 |= 0xc000
  → 移動步數／可見性 consumer
```

handler 沒有城鎮／地表 gate，因此 game pack 使用 `location_kind:"any"`；成功會消耗，
`step_count` 為 25。它與咒文雷米歐爾共用 timer/state，但道具寫 `0xc000`，咒文路徑寫
`0x8000` 加同一個 25 步 timer。

### 2.2 CTY39 守衛 writer 與 consumer

CTY39 section 0 base 是 `(file) 0x0004`；`(file) 0x0c2f` 是 section 1，不能混用。
section 0 special handler table 在 `(file) 0x00b2`，raw byte `0x1d`（29）。DGROUP
sub2 table `0x3bb4` 的 entry 29 指向 `(logical) 0x582c`（IDA linear `0x1582c`）。

`(logical) 0x582c..0x586a` 先測 `DGROUP 0x4f46 bit 0x8000`：

- 未隱形：比較 guard slot 0 X（`DGROUP 0x0b66`）與玩家 X（`DGROUP 0x4f33`），再以
  `(logical) 0x67c5` movement interpreter 及 `DGROUP 0x3cc9/0x3cd0` 左右腳本追到同欄。
- 隱形：bit 已 set 時直接 return，守衛不橫移。

CTY39 原始資料再提供：

- trigger：section 0 的 `(13,37)`、`(14,37)`，subid 1，handler raw 29；
- guard slot 0：起點 `(13,36)`、sprite 22、movement ctrl 0、handler/dialogue raw 15；
- spawn：`(13,42)`。

玩家移動 `(logical) 0x1a6a..0x1b91`（IDA linear `0x11a6a..0x11b91`）仍會讀
tile/NPC stamp，呼叫 consumer `(logical) 0xde2a`（IDA linear `0x1de2a`）；NPC 占格時
移動會回滾。故**隱形不等於全域穿越 NPC**：它只阻止
handler29 追蹤，玩家必須從守衛未占據的右欄通過。

## 3. 推論等級 ledger

| 結論 | 推論等級 | 依據 |
|---|---|---|
| `0x5d` 成功後清物品槽、timer=25、set `0xc000` | `proven` | pointer、handler bytes、writer 與 inventory consumer |
| handler29 在可見時橫移 guard slot0、隱形時 return | `proven` | CTY raw handler、IDA xref、writer、movement consumer |
| 隱形不提供一般 NPC 穿透 | `proven` | 玩家移動函式與 NPC stamp consumer，加 component 邊界測試 |
| CTY38 runtime 貨架販售 `0x5d` | `proven` | 原始商店表解析、正式 facility/shop transaction |
| 原版玩家慣用地名稱 CTY38 的最終 canonical 名稱 | `unknown` | 本輪不以攻略地名反推 CTY 編號 |
| 25 步暫態是否應跨原版存檔保存 | `unknown` | 本輪只證明道具已消耗、checkpoint 場景可續；不宣稱 timer 持久化 |

## 4. Remake 對應

schema `0.1.11` 新增／擴充：

- `item_use_effects[].step_count` 與 `temporary_invisibility`；
- `tracking_guard_events`：有限 CTY／section、trigger、guard selector、追蹤軸及 bypass effect；
- CTY parser 修正：即使 event table pointer 為 `0xffff`，special handler table 仍以最近的
  transition/layout pointer 決定結尾，CTY39 raw handler29 不再漏讀。

Go 引擎只實作具名暫態效果與通用水平追蹤 primitive；`0x5d`、CTY39、座標、subid、
handler、timer 與 selector 均留在 versioned game-pack JSON。未知 effect、axis、引用或空
trigger 一律失敗即關閉。

## 5. 正式玩家輸入與存讀檔

`TestOpeningProductionInputTrace` 從全新遊戲開始，沒有 debug key、env、grant item、座標
注入或直接事件呼叫，完成：

1. 正常練級、商店裝備與補給，完成日邦格兩階段事件及紫寶珠 save/load。
2. 正式離村、登船、航行至 CTY38，由 facility NPC 開原始貨架並買入 `0x5d`。
3. 正式抵達 CTY39，走到 `(14,38)`，由 rec421 道具選單選「使用」。
4. 驗證物品消耗、timer `25→24→23`，守衛保持 `(13,36)`，玩家由右欄到 `(14,36)`。
5. 由原始 transition 進 CTY76 section 0，保存、回標題讀檔後仍在 CTY76，隱形草維持已消耗。

2026-08-02 Docker／Xvfb 執行結果：`TestOpeningProductionInputTrace` 通過（約 22 秒）。

## 6. 畫面證據

| 狀態 | runtime state | 圖片 | 等級 |
|---|---|---|---|
| 可見玩家被追蹤 | 玩家 `(14,37)`；守衛由 `(13,36)` 移到 `(14,36)`，下一步阻擋 | [`eginbear_guard_visible_block.png`](img/eginbear_guard_visible_block.png) | V2 |
| 隱形玩家通過 | 玩家 `(14,36)`、timer 23；玩家 sprite 隱藏，守衛仍在 `(13,36)` | [`eginbear_invisibility_pass.png`](img/eginbear_invisibility_pass.png) | V2 |

第二張圖看不到玩家是效果本身，不可只靠畫面推斷座標；座標、timer 與 NPC slot 由同一
runtime component test 鎖定。尚未取得同狀態 DOSBox 原版截圖，故維持 V2。

## 7. 尚未完成

- CTY76 地下室推石／乾渴之壺已由下一切片閉合，見 [`docs/97`](97-eginbear-push-puzzle-production-trace.md)。
- CTY38 的 canonical 地名仍應另以原始文字／場景證據校正，不由貨架內容反推。
- 本切片完成不代表新遊戲至 THE END 全流程完成。
