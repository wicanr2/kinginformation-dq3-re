# 71 — NPC 可見性 story-flag 過濾:第一張地圖 NPC「過多」的根因與修正(2026-07-03)

> 使用者實測:remake 第一張地圖(阿里阿罕 CTY00)NPC **過多**,非原版數量,質疑 sec 地圖機制沒搞懂。
> 查證確認:原版 NPC 載入有**三層**,兩版 remake 都漏了第三層(逐 NPC story-flag 可見性過濾)。
> 這是靜態反組譯 + EXE 靜態資料映像當 ground truth 定案(未用 DOSBox)。與 docs/60(日夜兩份清單)互補。

## 原版 NPC 載入三層(反組譯 DQ3.EXE)

```
① 日夜選表:section base+0=白天表 / base+2=黑夜表([0x526c] 選)      ← docs/60 §2.3
② 讀表:{count byte, 7-byte 記錄 × count}                            ← docs/60 §2.4
③ 逐 NPC 過濾:記錄 byte5 = story-flag id → test [0x4f70] bit → 清則跳過  ← 本文(兩版 remake 缺這層)
```

第三層在 docs/60 §2.2 就反組譯出來了,但當時的 remake 實作(§4)只做了 ①②。

### 過濾迴圈(file 0x4568,disassembly 確鑿)

```asm
mov  bl, es:[si+5]          ; byte5 = story-flag id(整個 byte,0-255)
xor  bh, bh
mov  cl, 7 ; and cl, bl     ; cl = id & 7(bit 位置)
mov  ah, 0x80 ; ror ah, cl  ; mask
shr  bx,1 ×3                ; bx = id >> 3(byte index)
test byte [bx+0x4f70], ah   ; 測旗標 bit
je   skip                   ; ★ bit 清 → 跳過(不載入這個 NPC);設 → 載入
```

### [0x4f70] = 遊戲通用 story-flag 系統(flag API,file 0x824f/0x8264/0x8279)

| 位址 | 作用 |
|---|---|
| 0x824f | **SET**:`or [id>>3 +0x4f70], (0x80 ror id&7)` |
| 0x8264 | **CLEAR**:`and [id>>3 +0x4f70], ~mask` |
| 0x8279 | **GET**:`test …; al=1/0` |

寶箱(`dq3_treasure.h`「一次性旗標 id [0x4f70] bit」)、里程碑都用同一陣列。

## Ground truth:新遊戲初始 [0x4f70] 狀態(EXE 靜態資料映像)

DGROUP 檔基準用資料錨點定出:`lea dx,[0xd0]; int21 open` 開 `dq3man.bls`(字串在 file 0x16210)
→ **DGROUP file base = 0x16210 − 0xd0 = 0x16140**(`dq3mst.bls` @0xc5→0x16205 交叉驗證)。
故 [0x4f70] 陣列在 **file 0x16140 + 0x4f70 = 0x1b0b0**。舊文只讀 32 byte，
誤稱上限為 256 flags；R-3 handler 實際用到 `0x12c..0x131`，API 的 BX 也沒有 8-bit 截斷。
現讀 64 byte：

```
00 00 00 00 00 ff ff ff … ff   → flag 0-39【清】、flag 40-511【設】
```

**新遊戲時:byte5 < 40 的 NPC 隱藏(旗標清,劇情事件才設起→出現)、byte5 ≥ 40 顯示(基礎人口/初始在場)。**

> ⚠ 走過的彎路(誠實記):先前試「掃 EXE 對 SET(0x824f)的呼叫 → 有 setter=條件旗標」的 heuristic,
> 結果把 flag 35(掛 371 NPC 的基礎旗標)誤判成條件 → CTY02 從 31 掉到 1(空城)。改讀 EXE 靜態
> 映像的**實際初值**才對:flag 40-255 初始就設,heuristic 誤判的 43/66/128… 其實是基礎人口。
> 教訓:別用「有沒有 setter」推初始狀態,直接讀資料映像(對齊 docs 一貫「資料是真相」紀律)。

## 修正(ebitan,已實作)

- `internal/dq3data/townmap.go`:`NPCStoryInitFlags [32]byte` = 上述初值(flag 0-39 清、40-255 設)。
- `game/game.go`:`Game.storyBits [32]byte` + `storyFlag(id)/setStoryFlag(id,v)/initStoryBits()`;
  **獨立於 `g.flags`**（舊 remake 的高位里程碑空間），避免誤設 boss doneFlag 造成 boss 被跳過。
  新遊戲 init 與重置點都呼叫 `initStoryBits()`。
- `game/worldmap.go` `loadTownSceneSec`:加 `flagSet func(int) bool` 參數,NPC loop 內
  `if flagSet != nil && !flagSet(n.Flags) { continue }`(對齊原版 je skip);玩家路徑傳 `g.storyFlag`,
  測試傳 `nil`(不過濾)。
- **C remake ✅ 已同步(W5,commit f27b51d)**:`dq3_scene.c` 加獨立 `g_npc_vis[32]` + `dq3_npc_vis_init`
  (flag0-39清/40-255設,不共用 `dq3_storyflags` 避免誤設 boss doneFlag)+ `dq3_scene_load_npcs` 過濾;
  `test_npcvis` 驗 CTY00 24→15。**注意**:C 的 `dq3_flags_init` 仍全清零(那是里程碑陣列,與 NPC 可見性
  是**不同 id 空間**,不可混用——見 W5 commit 說明)。

### 驗證(數量,對 ground truth)

| CTY sec0 白天 | 未過濾 | 起始顯示 |
|---|---|---|
| CTY00 阿里阿罕 | 24 | **15** |
| CTY02 羅馬利亞 | 31 | 15 |
| CTY21 | 37 | 19 |
| CTY25 | 25 | 13 |

全 89 城掃描:**0 個城起始顯示掉到 ≤2**(無空城);阿里阿罕視覺 dump 確認仍有人但不再擠滿。
完整套件 game+internal 全綠、vet clean。

## 條件 NPC 事件 wiring spec(給 R 系列逐事件接)

條件旗標(byte5 0-39,起始清)由**遍布全遊戲的 21 個獨立劇情事件**各自 SET(反組譯 `call 0x824f`
的 `mov bx/bl,imm` 前綴掃出)。這些 NPC 起始隱藏是**對的**(對齊原版);ebitan 在實作對應事件時,
於該點呼叫 `g.setStoryFlag(id, true)`(離開/消失型呼 `false`)即現出/隱藏對應 NPC。**沒有捷徑:
旗標由事件設,事件不存在就無從接**(硬接=猜時機=編造機制,不做)。基礎設施 `setStoryFlag` 已就緒。

| flag id | 原版 SET 位址(file) | 掛的 NPC 數(全城) | 備註/區域線索 |
|---|---|---|---|
| 23 (0x17) | 0x156b | 少 | 開場區(near new-game init);同段 clear flag 80 |
| 24 (0x18) | 0x162a | — | 開場區 |
| 38 (0x26) | 0x5336 | — | UI di=0xc9 |
| 34 (0x22) | 0x5735 | — | — |
| 25 (0x19) | 0x5622 / 0x5673 / 0x719c | 多 | — |
| 16 (0x10) | 0x56c8 | — | di=0xc19 |
| 39 (0x27) | 0x6688 | — | di=0xbeb/0xbe8 |
| 36 (0x24) | 0x6c88 | — | di=0xbff |
| 33 (0x21) | 0x6c8e | — | di=0xbff |
| 31 (0x1f) | 0x6d4a | — | si=[0x4ed5] |
| 35 (0x23) | 0x6ea3 / 0x6f1d | **371** | di=0xbbd(大量基礎-ish 人口,需確認事件很早)|
| 32 (0x20) | 0x7053 | — | si=[0x4ed0/0x3cee] |
| 22 (0x16) | 0x716a | — | di=0xc10 si=[0x3d03] |
| 28 (0x1c) | 0x7232 | — | — |
| 20 (0x14) | 0x736a | — | di=0xc28 |
| 37 (0x25) | 0x73f0 | — | di=0xc2c |
| 17 (0x11) | 0x7563 | — | di=[0xb26] |
| 26 (0x1a) | 0x760a | — | — |
| 27 (0x1b) | 0x7610 | 8(阿里阿罕 sec0) | — |
| 14 (0x0e) | 0x7830 | — | di=0xbfd si=[0x4ee9] |
| 13 (0x0d) | 0xfab0 | — | 後期(高位址)|

> wiring 流程(每個 R 系列事件批):實作事件 X → 反組譯確認 X 對應的原版 [0x4f70] flag id
> (上表)→ 在 ebitan 事件點呼 `g.setStoryFlag(id, true)` → 對應條件 NPC 現出 → dump 核。
> ⚠ ebitan 的 `g.flags`（舊 remake 高位里程碑）與原版 `[0x4f70]` story flags 是
> **不同 id 空間**，不得混用；
> storyBits 專供 NPC 可見性,別把里程碑 id 塞進去。
> 部分 byte5≥40 旗標原版會被事件 **CLEAR**(NPC 離開,如 file 0x1568 clear flag 80);同理接 `false`。
