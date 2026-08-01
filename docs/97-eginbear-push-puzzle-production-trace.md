# 愛丁貝亞地下室推石與乾渴壺 production trace

> 更新：2026-08-02（Asia/Taipei）
> 範圍：CTY76 section 0、推動 NPC、handler30、passage flag `0x3b`、乾渴壺 `0x5e`

## 1. 結論

CTY76 的三個物件不是由 sprite `B2==40` 這個值決定「可推」。原版玩家碰撞 consumer
檢查被撞 NPC runtime record 的 `ctrl bit 0x40`；前方一格須在地圖內、地形可走且沒有
另一個 NPC。`DGROUP 0x4f46 & 0xc000 != 0` 時 consumer 會提前返回，因此隱形狀態反而
不能推物。

三個 NPC 推至 Y=5 後，CTY76 raw handler30 檢查 runtime slot 0、1、2 的 Y byte；三者
全部等於 5 才清 story flag `0x3b`，刷新 passage event tiles。passage 後的 event0 才是
可取得的乾渴壺；event1／flag `0x3b` 是 passage 狀態，不是第二個可收集寶物。

## 2. 原始資料

輸入：

- `DQ3.EXE` SHA-256：
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- `CTY76.DAT`：section 0，版面 `21×17`。

CTY76 parser anchors：

| 類別 | 原始值 |
|---|---|
| special handler table | raw handler30（`0x1e`） |
| NPC slot 0–2 | `(3,10)`、`(5,10)`、`(7,10)`；各 `B2=40`、`ctrl=0xc0` |
| 完成列 | low tile 91：`(4,5)`、`(5,5)`、`(6,5)` |
| handler30 triggers | `(7,5)`、`(4..7,6)`、`(4..6,7)`；subid1 |
| passage event tiles | `(7,5)` low 137、`(7,6)` low 139；完成後各加一並清 subid |
| event0 | `{type=1,item=0x5e,present_flag=0x3a}`，tile `(19,2)` |
| event1 | `{type=1,item=0x5e,present_flag=0x3b}`，屬 passage／handler 狀態 |

## 3. IDA Pro 9.4 證據

分析在一次性 Docker 容器內進行；原始 EXE 唯讀，臨時 `.i64`、匯出 sidecar 與人工語意
均位於 `/tmp`，不修改原始檔也不加入 Git。

### 3.1 玩家碰撞 writer

玩家移動函式呼叫 `(logical) 0xde2a..0xe056`（IDA linear `0x1de2a..0x1e056`）。入口 bytes 位於
`(file) 0xf19a`：

```text
f7 06 46 4f 00 c0 74 03 e9 89 00
8a dd 80 e3 1f 32 ff 88 2e 7c 06
d1 e3 d1 e3 d1 e3 8d 3e 66 0b 03 fb
b0 40 84 45 03
```

閉合資料流：

```text
玩家移動 caller
  → 由 tile/NPC stamp 取得被撞 NPC slot
  → test DGROUP 0x4f46,0xc000；非零直接失敗
  → runtime base DGROUP 0x0b66 + slot*8
  → test [npc+3],0x40
  → 以面向增量算 NPC 前方格
  → bounds／tile attr bit0／NPC stamp consumer
  → 寫回 [npc+0]=X、[npc+1]=Y，更新舊／新 tile stamp
  → 玩家才進入 NPC 原位置
```

因此 `B2=40` 只是這三個物件的 sprite key；把它當 pushability 是舊 C prototype 的錯誤
外推。正式引擎不再用 sprite 判定。

### 3.2 handler30 完成 consumer

CTY76 special handler raw 30 經 DGROUP sub2 table 指向 `(logical) 0x586b`，
`(file) 0x6bdb` bytes：

```text
bb 3b 00 e8 98 16 3c 01 74 01 c3
b9 03 00 8d 3e 66 0b 80 7d 01 05 75 f2
83 c7 08 e2 f5 bb 3b 00 e8 66 16 e8 47 db
bd 09 00 9a 40 02 53 10 9a b2 03 53 10 c3
```

它先測 flag `0x3b`；若仍為 set，從 `DGROUP 0x0b66` 起掃三個 8-byte NPC record，逐筆
比較 `[slot+1] == 5`。全部成立才 clear `0x3b`、刷新場景並播放原始 cue。它不比較 X，
也不按 sprite key 搜尋 NPC。

## 4. 推論等級 ledger

| 結論 | 推論等級 | 依據 |
|---|---|---|
| pushability 來自 runtime NPC `ctrl bit0x40` | `proven` | 玩家移動 caller、slot 計算、bit consumer、座標 writer |
| 隱形狀態會禁止推動 | `proven` | consumer 入口 `test [0x4f46],0xc000` 與失敗分支；component 邊界測試 |
| 完成條件是 slot0–2 的 Y 全為 5 | `proven` | handler30 原始 bytes、固定三次迴圈與 `[di+1]` consumer |
| 完成後 clear flag `0x3b` 並刷新 passage | `proven` | handler30 flag caller、CTY event tiles、runtime tile rebuild |
| event0／flag `0x3a` 才是可取得乾渴壺 | `proven` | CTY event table、tile `(19,2)`、type1 inventory consumer、正式命令輸入 |
| `B2==40` 表示所有可推物件 | `hypothesis`（已推翻） | 只有 CTY76 sprite 資料符合；原版 consumer 不讀 B2 |
| ctrl bit `0x40` 是全域 NPC 推動 gate，不限 CTY76 | `strong inference` | 碰撞 consumer 未讀 CTY；全 CTY raw 掃描另見 CTY01 sec0 slot6、CTY27 sec0 slot1 具有該 bit，component test 鎖定跨事件行為 |

## 5. game-pack 與正式玩家輸入

schema `0.1.12` 新增 pack 層級 `npc_push_rule`，保存 `ctrl mask/value` 與阻擋效果；有限
`push_puzzle_events` 則保存 CTY／section、八個 trigger、三個完成判定 NPC slot、初始座標、
完整 `ctrl_raw` parity anchor、完成軸／值與 clear flag。這個分層避免把 CTY76 的完成名單
誤當成全域碰撞 consumer 的可推名單。
乾渴壺 event0 另遷入 `treasure_events`；legacy Go `treasures` table 的 CTY76 subid0／1
已移除。

`TestOpeningProductionInputTrace` 從全新遊戲既有 CTY76 合法 checkpoint 繼續：

1. 只送正式方向輸入走完 119 步可解路，沒有座標、NPC 或 story flag 注入。
2. 最後一推後驗證三個 runtime slot 均在 Y=5、flag `0x3b` 已清、passage 已刷新。
3. 正常走到 `(19,2)`，開命令窗選「調查」，取得 item `0x5e` 並清 present flag `0x3a`。
4. 保存、回標題讀檔，驗證仍在 CTY76、乾渴壺存在、`0x3a/0x3b` 均保持 clear。

2026-08-02 Docker／Xvfb 結果：延長後的 `TestOpeningProductionInputTrace` 通過（約 27 秒）。

## 6. 畫面證據

| 狀態 | runtime state | 圖片 | 等級 |
|---|---|---|---|
| 推石前 | 玩家 `(5,12)`；三個 NPC slot 在 `(3,10)/(5,10)/(7,10)`；passage flag `0x3b` set | [`eginbear_push_puzzle_before.png`](img/eginbear_push_puzzle_before.png) | V2 |
| 完成後 | 玩家 `(4,6)`；三個 slot 均在 Y=5；flag `0x3b` clear，passage event tile 已刷新 | [`eginbear_push_puzzle_solved.png`](img/eginbear_push_puzzle_solved.png) | V2 |

兩張圖都由 production renderer 與正式方向輸入產生；尚未取得同狀態 DOSBox 原版截圖，
所以維持 V2，不宣稱 V3。

## 7. 尚未完成

- 下一個 production blocker 是乾渴壺在海中淺灘的正式使用、地圖變化與最終鑰匙取得。
- 推石前／完成後 runtime PNG 已保存；仍需原版同狀態 DOSBox 畫面升至 V3。
- 本切片完成不代表全新遊戲至 THE END 全流程完成。
