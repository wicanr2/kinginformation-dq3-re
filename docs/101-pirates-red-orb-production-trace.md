# 海盜村紅寶珠正式流程追蹤（2026-08-02）

## 結論

由 `docs/100` 的藍寶珠復隊合法 checkpoint 出發，正式流程已能經魯拉、登船、航行至
CTY27，推開入口物件進入地下藏寶室，取得紅寶珠 `0x68`，再完成存檔、回標題讀檔與入口
物件消失驗證。全程使用正式 `InputState`，沒有座標注入、debug key 或直接呼叫事件函式。

- 流程證據：**E3**。
- 畫面證據：**V1**。已產生現行 runtime PNG，但本輪在本機完整影片中沒有定位到原版同狀態
  畫格，不能宣稱 V2/V3。
- game-pack schema：`0.1.15`（未變更）。
- `dq3_cht` content：`0.1.17`。

## 原版資料與 IDA 證據

本輪以唯讀方式掛載原始檔，在一次性 Docker 容器內使用 IDA Pro 9.4；IDA database、sidecar
與授權資料均留在 `/tmp` 或原始安裝位置，沒有修改原始檔，也沒有加入 Git。

- `DQ3.EXE` SHA-256：
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- CTY27 sec0 base：`0x0004 (file)`。
- 日／夜入口物件：`0x0025 (file)`／`0x0034 (file)`，raw
  `1a 09 28 c1 0e 3f 00`。可直接證實座標 `(26,9)`、runtime control `0xc1` 含
  `bit 0x40`、dialogue 14、visibility flag `0x3f`；`B2=0x28` 的精確美術語意尚未證實，
  因此只稱「可推物件」。
- CTY27 sec0 transition0：`0x0081 (file)`，raw `1b 01 06 03`，普通入口到
  CTY27 sec1 `(6,3)`。
- transition1：`0x0085 (file)`，raw `1b 01 05 09`，密道入口到 sec1 `(5,9)`。
- `(26,9)` tile：`0x02ef (file)`，raw `21 01`，高位 subid 1。
- CTY27 sec1 event table：`0x0a5e (file)`，raw
  `03 01 68 00 3f 01 50 00 40 01 43 00 41`；event0 是 item `0x68`、flag `0x3f`。

IDA 主要 consumer chain：

1. `sub_1DE2A` `(logical) 0x1de2a..0x1e057` 先檢查 world effect `0xc000`；非零時
   禁止推動，再於 `(logical) 0x1de4c` 檢查 NPC runtime control `bit 0x40`，驗證前方 tile
   與 NPC stamp 後移動物件一格。
2. `sub_13534` `(logical) 0x13534..0x1372f` 讀 section transition table，使用玩家所在
   tile subid（`DGROUP 0x0b55 & 0x1f`）乘以 4，取得目的 `{cty,section,x,y}`；跨 CTY
   remembered-coordinate 分支在 `(logical) 0x13661`／`0x49d1 (file)`。
3. 可重現的指令錨點另鎖定 `0x48aa (file)` 的 transition 讀取、`0xf19a (file)` 的
   invisibility/world-effect gate、`0xf1bc (file)` 的 `ctrl bit0x40` 判定。

推論等級：上述 raw 欄位、reader 與玩家可見結果為 D2/D3；「入口物件在故事語意上是大石」
只由攻略與外觀支持，尚未由 sprite table 閉合，因此不作 production identifier。

## 正式玩家路徑

直接從蘭西爾離開後，船仍停在先前座標，玩家無法由陸路抵達。正式 trace 因此先由道具選單
魯拉至 CTY15；已證實的原版魯拉副作用會把船重定位到 `(104,126)`。玩家再正式登船、航行到
CTY27 外圍，依序完成：

1. 走到入口物件下方 `(26,10)`。
2. 按上推動具 `ctrl bit0x40` 的物件，露出 transition subid 1，進入 sec1 `(5,9)`。
3. 走到 event0 寶箱並由調查命令取得紅寶珠 `0x68`；原版 present flag `0x3f` 在取得後
   **clear**，不是 set。
4. 存檔、回標題、正式讀檔，驗證紅寶珠仍由全隊持有且 flag 維持 clear。
5. 回到 sec0；同一 present flag `0x3f` clear 後，入口物件於場景重載時不再出現。

> 2026-08-22 勘誤：舊版步驟把 `0x3f` 寫成「取得時 set」，與 CTY present-flag
> consumer、現行 `collectPackTreasure` 及正式 trace 相反；本段已訂正並保留錯誤原因。
> 共用取得 writer 接回後，紅寶珠也可能合法落入同伴，取得與 save/load 驗收改為全隊持有，
> 不再假設勇者 owner。

普通入口 `(6,3)` 即使門已開也無法走到密室寶箱；密道 transition 不是可省略的裝飾流程。

## 驗證與畫面

- `TestDQ3PiratesRedOrbMatchesOriginalEXEAndCTY`：鎖定 EXE consumer 指令與 CTY27 raw data。
- `TestPiratesRedOrbPushEntranceAndTreasure`：鎖定 push gate、密道 transition、寶珠交易與
  scene reload visibility。
- `TestOpeningProductionInputTrace`：從 boot 起的既有正式流程接續至紅寶珠、存讀檔與返回。
- Runtime 圖：
  - [`pirates_red_orb_hidden_entrance.png`](img/pirates_red_orb_hidden_entrance.png)
  - [`pirates_red_orb_obtained.png`](img/pirates_red_orb_obtained.png)

本機影片已以有界 contact sheet 搜尋，但沒有找到原版紅寶珠同狀態畫格；影片中的修改數值也
不得作參數 oracle。後續若取得同狀態 DOSBox 畫面，才可把本切片由 V1 升級為 V2/V3。

## 下一個 blocker

紅寶珠完成後，P3 第一候選是商人建城／黃寶珠；仍須從本輪合法 checkpoint 重播，以玩家
實際遇到的第一個 blocker 為準，不能因後段已有孤立 handler 測試而跳過連接路徑。
