# R-5b：索瑪城事件與戰後拉達多姆（2026-07-28）

本輪以本機完整實機影片、CTY90/79/80 原始資料，以及 IDA Pro 9.4 產生的
`DQ3.EXE.i64/.asm` 三方交叉驗證。目標不是補一條可通關捷徑，而是撤回先前
「擊倒索瑪立刻播 ENDTXT」的錯誤流程。

## 1. 王座後的隱藏樓梯

- CTY90 sec0 `(23,3)`：event subid3 = `{type=4,param=0,flag=0xe5}`。
- IDA `sub_18986 → sub_18C01`：
  1. CLEAR event story flag；
  2. 當前 tile 低 byte `+1`；
  3. `hiMap &= 0xe0`，清除事件 subid；
  4. 顯示 D3TXT00 rec484「啊，發現一座隱藏樓梯！」。
- 顯露後的 tile 使用原始 transition[0]，進 CTY90 sec1 `(16,8)`。

Ebiten 已由正式命令窗「調查」走完整路徑，並測試 flag、tile mutation、訊息及
踩樓梯轉場；不再用 debug 鍵直接進最終層。

## 2. 歐魯迪卡橋上過場

原始 CTY90 sec4 NPC：

| 座標 | 身分 | 可見旗標 |
|---|---|---|
| `(12,28)` | 戰鬥演出 sprite | `e0` |
| `(13,28)` | 戰鬥演出 sprite | `e0` |
| `(13,28)` | handler90 垂死歐魯迪卡 | `0e` |

IDA：

- handler79 `sub_16494`：要求 flag `e0`，播放 D3TXT07 rec69，固定 formation
  `02 2a 08 80 01 81 01`，接著 CLEAR `e0`、SET `0e`。
- handler90 `sub_16724`：播放 D3TXT07 rec70，CLEAR `0e`。

影片中玩家隊長與 `(12/13,28)` sprite 的水平 tile 差，加上 CTY90 collision
交叉驗證後，玩家實際腳位為橋面 `(18,29)`；由右向左踏入便自動觸發。

Ebiten 已還原 rec69 → `e0/0e` NPC 換景 → rec70 → 清除 NPC 的自然流程，
且不誤開可由玩家下指令的通常戰鬥 UI。原版 formation 0x80 是固定演出；
其逐回合動畫仍列為後續視覺細化，不把一般 boss 戰冒充原版過場。

## 3. 「第二隻五頭龍」的判定

IDA handler81 `sub_16586` 的確保留 formation
`01 2a 08 81 01`，對應當年 BBS 所說的第二隻五頭龍。然而本專案的權威
`assets_raw/DQ3.EXE` 在原本 battle call 的三個 byte 已是 `NOP NOP NOP`；
本機完整影片進最終房後也直接走：

`巴拉摩斯怨靈 0x7a → 巴拉摩斯殭屍 0x7b → rec72 → 索瑪 0x7c`

因此這輪不憑歷史上另一 build 的當機點，額外捏造一場會改變本機精訊版流程的
戰鬥。formation 0x81 與修復 sprite 仍保留為版本差異證據。

## 4. 索瑪戰後不是立即結局

IDA：

- handler80 勝利分支 `sub_164CD` CLEAR flag `e1`，重建場景並回到愛列夫加特；
  它沒有直接授予洛特封號。
- CTY80 sec1 `(10,4)` 國王是 handler74 `sub_16346`。
- `e1` 已清時，handler74 進 `sub_1E713`，先顯示 D3TXT08 rec48，再跑固定 ending。

Ebiten 現在：

1. 打倒索瑪後只設 `cleared/msZoma`、CLEAR `e1`，回 CTY90 城外的下層地表；
2. 不先設 `lotoBlessed/0x217`，也不先開 ENDTXT；
3. 下層蓋美拉翅膀回 CTY79 拉達多姆外城；
4. 依原始 CTY79→CTY80 sec0→sec1 轉場進王座；
5. 自然與 handler74 國王交談，播 rec48；
6. 對白結束才設洛特封號與 `0x217`，接 ENDTXT。

本機影片使用傳送咒文回拉達多姆；Ebiten 的非戰鬥咒文選單仍未完成，所以目前先
讓同類 ReturnTown 道具在下層正確回 CTY79，而不是錯回地上阿里阿罕。傳送咒文
選單仍是全流程完成前必須補齊的項目。

## 5. 自動驗證與視覺證據

- `TestZomaHiddenStairNaturalExamine`
- `TestOrtegaBridgeEventNaturalMovement`
- `TestRunFinale`
- `TestZomaAftermathRadatomeRoute`
- `TestEndingScroll`
- `TestComponentSpine`

新增／重產畫面：

- `zoma_hidden_stair_before.png`
- `zoma_hidden_stair_revealed.png`
- `ortega_bridge_rec69.png`
- `ortega_dying_rec70.png`
- `zoma_aftermath_overworld.png`
- `radatome_return.png`
- `radatome_king_loto.png`
- `ending_scroll_start.png`
