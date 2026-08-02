# 奧莉薇亞海岬至蓋亞之劍 production trace

## 結論

schema `0.1.21`／content `0.1.25` 已把愛的回憶後續閉合成正式玩家切片：

```text
CTY36 取得愛的回憶 → 原始轉場離開幽靈船 → 同座標恢復停泊船 mode1
→ 航行至地表 (76,54) → records597/598 → clear flag0x35
→ 沿海岬航道進 CTY55 → 調查寶箱取得蓋亞之劍0x0f → save/load
```

全程由 boot 起的同一條 `InputState` trace 抵達，沒有直接設定座標、授予道具、呼叫事件
handler 或跳過戰鬥。畫面目前只有 runtime V1；尚未定位原版影片的完全同狀態畫格，故不標
V2／V3。

## 原始輸入與位址口徑

- 輸入：`assets_raw/DQ3.EXE`
- 大小：115282 bytes
- SHA-256：`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- 主要工具：IDA Pro 9.4；沿用同一輸入建立的 `/tmp/dq3-ghostship-ida/DQ3.i64` 與受控
  sidecar。原始名稱、位址與運算元不改名；本文語意是附加註記。
- 位址換算：`file = IDA linear - 0xec90`。本文的 `linear` 均指 IDA image linear address。
- 第二證據：直接比對 `DQ3.EXE`、`CTY55.DAT`、`D3TXT00.TXT` 原始 bytes／records；
  `TestDQ3OliviaCapeAndGaiaSwordMatchOriginalEXECTYAndText` 鎖定同值。

## 證據 ledger

| 原始定位 | 原始結果 | remake 對應 | 等級 |
|---|---|---|---|
| linear `0x126bd..0x126d3` | 玩家必須在 `(76,54)`，且 flag0x35 set | `coordinate_item_gate_events.coordinate/active_story_flag_raw` | confirmed |
| linear `0x126d5..0x126e3` | 寫 item0x64，呼叫 inventory search，依結果分支 | `required_item_raw_id=100` | confirmed |
| linear `0x126e5..0x12701` | 有道具顯示 records597／598，再 clear flag0x35 | 兩個 text ID、成功後提交 clear；不消耗道具 | confirmed |
| linear `0x12704..0x12723` | 缺道具顯示 record597，日夜 byte加20，direction3 呼叫移動五次 | failure direction3／steps5／day-night steps20 | confirmed static；runtime 可見移動 E2 |
| `CTY55.DAT` file `0x56` | `01 0f 00 48`；section0 event0、tile `(14,22)` | JSON `treasure_events`，給0x0f並 clear present flag0x48 | confirmed |
| `D3TXT00.TXT` records597／598 | 悲歌、愛立克／歐里比雅重逢與解除詛咒 | 完整 glyph/control words，沒有 Go runtime 文字 | confirmed |

手動使用愛的回憶仍只顯示 record596；解除詛咒是地表座標 consumer，兩者不可合併。

## 實作修正

- 新增有限 `coordinate_item_gate_events` 契約；JSON 不能放任意 callback 或條件式。
- CTY55 蓋亞之劍從舊 Go `treasures` table 移至 game-pack JSON。
- 幽靈船出口回到停泊船同一世界座標時恢復 mode1。先前未恢復會把玩家困在不可步行水格，
  正式 trace 已實際撞出並修正這個連通性問題。
- 海岬成功 branch 等 record598 完整關閉後才 clear flag0x35；失敗 branch 逐 frame 執行五步，
  不直接瞬移。

## 驗收

- 原始 EXE／CTY／D3TXT parity test：通過。
- 成功、失敗、重播抑制、道具不消耗與五步水流 component tests：通過。
- boot 起 production trace 延伸至 CTY55 寶箱及 save/load：通過。
- runtime 圖：
  - `docs/img/olivia_cape_memory_reunion.png`（V1）
  - `docs/img/gaia_sword_obtained.png`（V1）

## 下一個 blocker

現行舊碼仍把蓋亞之劍 `0x0f` 寫成「地表任意位置使用即設 remake flag0x32」，這不是原版
規格。後續 IDA 已確認特殊 item table 只列 `0x0f`，派發至 `(logical) 0x3d65`；唯一成功
座標是 `(65,109)`，成功後進入動畫、套用 `DGROUP 0x3b96` 地圖 patch 並 set 原版
world-state bit `0x20`。patch 尺寸／完整 bytes、重建 consumer 與 DOSBox 可見結果仍待閉合，
因此尚未修改 production。詳細證據與下一輪順序見 [`docs/106`](106-gaia-sword-volcano-re-worklist.md)。
