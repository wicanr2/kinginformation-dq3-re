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

### [0x4f70] = 遊戲通用 256-flag story-flag 系統(flag API,file 0x824f/0x8264/0x8279)

| 位址 | 作用 |
|---|---|
| 0x824f | **SET**:`or [id>>3 +0x4f70], (0x80 ror id&7)` |
| 0x8264 | **CLEAR**:`and [id>>3 +0x4f70], ~mask` |
| 0x8279 | **GET**:`test …; al=1/0` |

寶箱(`dq3_treasure.h`「一次性旗標 id [0x4f70] bit」)、里程碑都用同一陣列。

## Ground truth:新遊戲初始 [0x4f70] 狀態(EXE 靜態資料映像)

DGROUP 檔基準用資料錨點定出:`lea dx,[0xd0]; int21 open` 開 `dq3man.bls`(字串在 file 0x16210)
→ **DGROUP file base = 0x16210 − 0xd0 = 0x16140**(`dq3mst.bls` @0xc5→0x16205 交叉驗證)。
故 [0x4f70] 陣列在 **file 0x16140 + 0x4f70 = 0x1b0b0**,32 byte:

```
00 00 00 00 00 ff ff ff … ff   → flag 0-39【清】、flag 40-255【設】
```

**新遊戲時:byte5 < 40 的 NPC 隱藏(旗標清,劇情事件才設起→出現)、byte5 ≥ 40 顯示(基礎人口/初始在場)。**

> ⚠ 走過的彎路(誠實記):先前試「掃 EXE 對 SET(0x824f)的呼叫 → 有 setter=條件旗標」的 heuristic,
> 結果把 flag 35(掛 371 NPC 的基礎旗標)誤判成條件 → CTY02 從 31 掉到 1(空城)。改讀 EXE 靜態
> 映像的**實際初值**才對:flag 40-255 初始就設,heuristic 誤判的 43/66/128… 其實是基礎人口。
> 教訓:別用「有沒有 setter」推初始狀態,直接讀資料映像(對齊 docs 一貫「資料是真相」紀律)。

## 修正(ebitan,已實作)

- `internal/dq3data/townmap.go`:`NPCStoryInitFlags [32]byte` = 上述初值(flag 0-39 清、40-255 設)。
- `game/game.go`:`Game.storyBits [32]byte` + `storyFlag(id)/setStoryFlag(id,v)/initStoryBits()`;
  **獨立於 `g.flags`**(remake 里程碑 0x211/0x213 等),避免誤設 boss doneFlag 造成 boss 被跳過。
  新遊戲 init 與重置點都呼叫 `initStoryBits()`。
- `game/worldmap.go` `loadTownSceneSec`:加 `flagSet func(int) bool` 參數,NPC loop 內
  `if flagSet != nil && !flagSet(n.Flags) { continue }`(對齊原版 je skip);玩家路徑傳 `g.storyFlag`,
  測試傳 `nil`(不過濾)。
- C remake 待同步(`dq3_scene.c` 同缺此層;`dq3_flags_init` 目前全清零,需改為此初值才能接 NPC 過濾)。

### 驗證(數量,對 ground truth)

| CTY sec0 白天 | 未過濾 | 起始顯示 |
|---|---|---|
| CTY00 阿里阿罕 | 24 | **15** |
| CTY02 羅馬利亞 | 31 | 15 |
| CTY21 | 37 | 19 |
| CTY25 | 25 | 13 |

全 89 城掃描:**0 個城起始顯示掉到 ≤2**(無空城);阿里阿罕視覺 dump 確認仍有人但不再擠滿。
完整套件 game+internal 全綠、vet clean。

## 待辦(誠實揭露)

- **條件 NPC 的事件設旗標(byte5<40)尚未逐一 wire**:這些 NPC 起始隱藏是**對的**(對齊原版),
  但它們在原版某劇情事件後會出現;ebitan 需在對應事件呼叫 `setStoryFlag(id, true)` 才會現出。
  屬「隨事件批次補」的後續工作(與 R 系列事件實作一起做),非本次範圍。當前狀態:**起始/基礎人口已對齊**。
- 部分 byte5≥40 的旗標原版會被事件 **CLEAR**(NPC 離開,如 file 0x1568 clear flag 80);同屬事件 wire 範圍。
