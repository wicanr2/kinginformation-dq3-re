# 商人聚落世界入口分支逆向證據

狀態：入口選擇與商人交付已閉合（D3／E3）；CTY59 handler41 的 Y=4 設施分支已由
IDA／CTY／D3TXT 閉合並落成 `conditional_facility_events`（D3／E2）；CTY83 黃寶珠資料、
牢中對話與元件交易已達 D3／E2／V1。同一條 boot production trace 已從合法 checkpoint
以正式輸入自然進入 CTY83、完成黃寶珠與 save/load（E3）；剩餘是跨階段與原版同狀態
V2／V3 對拍，不把元件圖誤標成完整逐畫面 parity。

## 問題

同一個地表座標 `(210,64)` 對應 CTY58、59、60、61、83。歷史 C remake 與舊 Go
`owPortals` 把它解釋成「旗標設立即選相對應 CTY」，導致新遊戲在 `0x23` 未設時直接載入
革命後的 CTY83。這與 CTY58 的原始 NPC handler 及 D3TXT07 劇情順序矛盾。

## 輸入與工具

- `DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- CTY58／59／60／61／83 與 D3TXT07 均從原始資料唯讀分析。
- IDA Pro 9.4，database 位址基準為 IDA linear address；工作 database 與 sidecar 位於
  `/tmp/dq3-merchant-ida/`，不加入 Git。

## 入口 writer → consumer（confirmed）

`DQ3.EXE` file `0x396e..0x39f1`（IDA linear `0x125fe..0x12681`）先比較目前地表座標
與 `cty_loc[58]`，再依序測試 story flag：

| 順序 | 原始分支 | 寫入目的 CTY | 玩家可見階段 |
|---:|---|---:|---|
| 1 | `flag 0x23 == clear` | 58 | 初始聚落，可交付商人 |
| 2 | `flag 0x23 == set && flag 0x47 == set` | 59 | 第一階段 |
| 3 | 上述未命中且 `flag 0x42 == set` | 60 | 第二階段 |
| 4 | 上述未命中且 `flag 0x48 == set` | 61 | 第三階段 |
| 5 | 全部未命中 | 83 | 革命／牢獄階段 |

目的 CTY 寫入 `DGROUP 0x256c`，section `0` 寫入 `DGROUP 0x256a`，再呼叫共同 CTY
loader。這證明分支是 ordered decision，不是任意 flag-to-CTY map。

新遊戲原始 story-bit 映像為 flag `0..39` clear、`40..511` set，因此初值恰為：
`0x23` clear，`0x47/0x42/0x48` set。第一個分支必定選 CTY58；交付商人後 handler
會 set `0x23`，下一次進入才會選 CTY59。

## 商人交付入口（confirmed，已落地）

CTY58 section0 的 subtype2 NPC `(7,15)`、`byte4=40` 透過 `DGROUP 0x3bb4` handler table
到 `DQ3.EXE` file `0x6e12..0x6f29`（IDA linear `0x15aa2..0x15bb9`）。該 handler：

1. 顯示 D3TXT07 record 0，尋找 active party 中 class raw `6` 的成員；
2. 依 D3TXT07 record 1–4 完成缺商人與兩次確認；
3. 成功顯示 record 5 並 set flag `0x23`；
4. 將商人的八個物品逐一送進 `DGROUP 0x526f`、容量 `0x78` 的共用預存區；
5. 保存 18-byte 商人身份資料到 `DGROUP 0x5058`，清除該角色可用狀態並從 active party
   pointer table 移除；
6. 依性別 discriminator 設定 `0x23` 或 `0x15` 的 NPC 顯示旗標並重載場景 NPC。

這條行為不能用「從 roster 刪除」近似。Go 現已以 `settlement_founder`
有限狀態機落地：只掃 active party，保存建城者身份，將裝備與個人道具移入
共用預存所，再從 active party 移除，不回 roster。founder 與 storage 均有
save/load round-trip。

### 預存所滿載 consumer（confirmed）

IDA Pro 9.4 的 `sub_16CE2`（IDA linear `0x16ce2..0x16d04`；file
`0x8052..0x8074`）以 `CX=0x78`、`SI=DGROUP 0x526f` 掃描第一個 `0xff`；找到時
寫入 `AL`，滿載時呼叫全域 D3TXT00 record `0x135`（十進位 309）「你存入的道具
太多了」。handler40 對八個物品槽逐一呼叫且忽略回傳值，所以每一件溢位道具
都顯示一次 record309，然後仍完成商人交付。component test 已鎖定這個次數與
「最後一次訊息後才離隊」時序。

## Remake 對應

- `events.json.world_entrance_variants` 保存座標、layer、ordered branches、分支極性及 CTY。
- 共用 engine 的 `worldEntranceResolve` 只理解有限契約，不知道 DQ3 的座標、旗標或 CTY。
- schema/reference validation 拒絕重複座標、重複 flag、重複 CTY、越界值及缺少 evidence。
- parity test 鎖定五個商人聚落狀態及 `(54,129)` 的另一條原版入口分支。
- `settlement_founder_events` 保存 NPC selector、職業、存活條件、完成旗標、
  共用預存所容量與全部 text ID；Go 只保留通用二次確認與 transaction primitive。
- `TestOpeningProductionInputTrace` 已由紅寶珠合法 checkpoint 透過露易達酒場
  寄放隊員、登錄並招募商人，正式魯拉、登船、航行進 CTY58，完成兩次
  確認與 save/load，無 debug key、座標或旗標注入。

## 玩家可見證據

- [CTY58 交付前](img/merchant_settlement_cty58_before.png)
- [原始自我介紹](img/merchant_settlement_introduction.png)
- [第一次商人確認](img/merchant_settlement_first_offer.png)

這三張是真實 CTY58 素材與 production renderer 的 V1 runtime PNG。本機 59 分鐘實況
`dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4` 已抽樣前 37.5
分鐘，尚未找到同狀態 CTY58 片段；因此目前誠實標為 V1，不宣稱原版對拍 V2。

### CTY83 handler48 與黃寶珠（confirmed）

- `CTY83.DAT`：2,841 bytes；SHA-256
  `ae882639131e1be2f62b21f6643edfd0e5b6f497df692314d182a799e0a46934`。
- section0 的 event table 在 file `0xf4`；tile `(4,2)` 的 subid2 entry 位於 file
  `0xfd`，原始 bytes 為 `01 6a 00 4a`，即 type1、item `0x6a`、present flag `0x4a`。
- handler48 為 logical `0x5d50..0x5d70`（IDA linear `0x15d50..0x15d70`，
  file `0x70c0..0x70e0`）：固定顯示 D3TXT07 record42，讀 flag `0x4a`；flag set 時
  再顯示 record43「在我座位後面的東西，你們就拿去吧」，clear 時跳過 record43。
- common examine dispatcher logical `0x8986`（IDA linear `0x18986`）先讀 event 的 p2
  present flag；type1／3 分支呼叫 `sub_18A79`，成功後共同流程清 flag。handler48 本身
  不給黃寶珠，兩條路徑不可合併。
- IDA Pro 9.4 sidecar 位於 `/tmp/dq3-revolution-ida/revolution_ida.txt`；輸入 EXE 雜湊與
  本文件前述相同，位址基準為 IDA linear。原始 file raw 的 far-call segment relocation
  word 是 `0x111b`，IDA 載入後顯示 `0x211b`；parity test 鎖 file bytes，不混用兩種基準。

舊 `treasures` Go table 使用「flag set＝已取得」，與原版 present flag 語意相反；因
`flag0x4a` 初值為 set，會讓黃寶珠一開始就被判定為已取。該列已刪除，改由
`treasure_events` JSON 使用「set＝存在、成功後 clear」。handler48 的 record42／43、
旗標分支與建城者姓名插值也由 `settlement_founder_followups` JSON 提供。男性建城者的
後續 sprite flag `0x15`、女性 `0x23` 已加入交付 transaction；舊實作只設完成 flag
`0x23`，會漏掉男性建城者的後續可見狀態。

目前 runtime 圖片：

- [CTY83 革命場景](img/merchant_revolution_cty83.png)
- [牢中建城者 record42](img/merchant_revolution_imprisoned.png)
- [座位後方提示 record43](img/merchant_revolution_seat_hint.png)
- [調查取得黃寶珠](img/merchant_revolution_yellow_orb_obtained.png)

圖片使用真正 CTY83、D3TXT07 與 production renderer，已目視確認；但 fixture 由已證實
狀態載入，尚未從商人交付 checkpoint 自然完成沙曼歐莎與蓋亞之劍鏈，因此只標 V1／
component E2，不宣稱 E3 或原版同狀態 V2。

## 建城 stage gate writers（confirmed）

- `0x47` 是 CTY40 section0 最終鑰匙寶箱的 present flag；原始 event
  `01 57 00 47` 經共同寶箱 consumer 成功給物後清除。見 `docs/98`。
- `0x42` 是沙曼歐莎假王事件 gate；原版 file `0x5682..0x5732` 在拉之鏡成功入口
  clear `0x42`，戰敗／逃跑會 restore，勝利維持 clear。見 `docs/74` 的 R-2 證據。
- `0x48` 是 CTY55 section0 蓋亞之劍寶箱的 present flag；原始 event0 file `0x56`
  為 `01 0f 00 48`，共同寶箱 consumer 成功給物後清除。見 `docs/data/quest-items.md`。

因此建城階段由既有主線事件推進，不是時間、步數或 remake 專用「建城完成」旗標。晚交付
商人時，下一次進城依玩家已清除的 gate 直接選較後 CTY，是 ordered chain 的自然結果。

## CTY59 handler41 設施分支（confirmed）

IDA Pro 9.4 的 `sub_15BBA`（IDA linear `0x15bba..0x15bd4`；file `0x6f2a..0x6f46`）
先比較 `DS:4f35`（玩家 Y）與 `4`。未命中時設定 `di=0xbc0`，由 D3TXT07 loader
解析為 record8「啊啊,是{變數}……」；命中時設 `BX=0` 呼叫共用設施 dispatcher
`sub_17034`，其後從當前 CTY section 的 facility pointer table 取 type。CTY59 section0
的 index0 block 是 type2、兩項貨架，因此這不是一般 facility subtype，而是「同一個
handler 在 Y=4 才開道具店」的有限分支。

此鏈已由 `conditional_facility_events` 提供 selector、Y gate、facility index、record8
text ID 與 D3 evidence；production component test 以正式 `cmdTalk` 驗證 Y≠4 對話與
Y=4 貨架兩條結果，並產出兩張現行 runtime PNG。原始 EXE／D3TXT parity test 鎖定
file bytes 與 record8，沒有把 CTY59 或 handler41 常數放回共用 Go 流程。

- [Y≠4 對話分支 runtime PNG](../dq3_remake_ebitan/docs/img/merchant_settlement_handler41_dialogue.png)
- [Y=4 道具店分支 runtime PNG](../dq3_remake_ebitan/docs/img/merchant_settlement_handler41_shop.png)

## 尚未閉合

- CTY60／61 的階段差異，以及這些階段的原版同狀態 V2／V3。
- CTY83 黃寶珠與 CTY59 handler41 的 runtime 圖仍需原版同狀態 V2／V3 對拍；玩家路線
  與 save/load 已由 boot production trace 閉合。
- 從原版實況定位 CTY58 同狀態畫面，將上述 runtime PNG 由 V1 升為 V2。
