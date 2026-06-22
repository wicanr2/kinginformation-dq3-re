# DQ3.EXE 戰鬥系統:進入點 / 主迴圈 / 指令 / 怪物資料 / sprite(RE → C)

把精訊版 DQ3 的戰鬥子系統從 DQ3.EXE 反組譯成可讀 C。戰鬥畫面對照
`references/game3.png`(藍天草原背景、6 隻史萊姆橫排、上方隊伍狀態框
勇者/武鬥家/僧侶/魔法師 HP/MP、下方指令 戰/逃/防/道具)。反組譯後各畫面元素
都對應到具體函式,與 game3.png 逐項吻合(見文末對照表)。

工具:`tools/re_battle_dis.py`(docker capstone 內跑,給 seg0 off / dump D3MNS.DAT、
DQ3MNS.SHP)。位址空間 = seg0 邏輯 offset(file off = 0x1370 + off)。主資料段
DS=0x14dd(file base 0x16140)。反編譯 C 在 `re/battle.c`,宣告在 `re/battle.h`。

```
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone \
           && python tools/re_battle_dis.py"          # 戰鬥函式群
           # python tools/re_battle_dis.py d3mns       # D3MNS.DAT 結構
           # python tools/re_battle_dis.py shp         # DQ3MNS.SHP 表頭
           # python tools/re_battle_dis.py 0xNNNN n    # 任意 off 反組譯
```

## 戰鬥進入點:野外行走 → 步數遭遇 → 戰鬥

戰鬥不是獨立狀態,而是掛在**主迴圈每幀的 HUD/事件處理尾端**。呼叫鏈:

```
main_loop (sub_93e3)
  └─ update_hud (sub_9530)              ; 每幀末
       └─ battle_check_encounter (sub_bd97)   ; 步數倒數,歸 0 才起戰
            └─ battle_main (sub_bddf)          ; 戰鬥主流程
```

`battle_check_encounter`(sub_bd97)= 經典 **RPG 步數遭遇計數器**:

- 地表移動(或城鎮中 `[0xd77]==0`)時,每走一步把遭遇計數 `[0x52f4]` 減 1。
- 計數歸 0 → 擲新的遭遇步數,並進入 `battle_main`:
  - **事件強制遭遇**(`[0x4f46] & 0x1000`):步數 = `rng(5)+1`(短距離,連戰用)。
  - **一般遭遇**:重擲 `rng(0x12)` 直到 ≥10,再視 `[0x526c]`(近城 / 特定地形)決定
    是否 `-2`(地形影響遇敵頻率)。

→ 與 docs/10 對齊:狀態 handler 中 `[0x4f3b] >= 2`(模式旗標)時 Space/Enter 改走
`sub_434f`——`[0x4f3b]` 即「是否在戰鬥 / 過場模式」;戰鬥期間玩家輸入被導向別處。

## 戰鬥主流程:`battle_main`(sub_bddf)

```
battle_main:
  [0x2517]=0                              ; 遭遇結果旗標(0=戰鬥續行)
  encounter_build_group (sub_a7d5)        ; 由玩家 tile 查遭遇區 + 生成敵群
  if [0x2517]!=0: return                  ; 遭遇被取消(空群 / 無敵地形)
  alive = party_alive_count (sub_c5f6)    ; 數存活我方 → bp
  if alive==0: [0x2518]=0; return         ; 無人可戰 → 退出
  [0x24b4]=alive; [0x24b3]=enemy_total
  battle_setup_party (sub_c8c6)           ; 攤平隊伍 HP/MP/狀態 → [0x063a]
  battle_enter_screen (sub_bfd1)          ; 載入 packbg 戰鬥背景(頁由地形 sub_d9f8 決定)
  battle_draw_status (sub_c572)           ; 畫我方狀態框(隊伍人數定高度)
  battle_draw_names (sub_c59b)            ; 逐隻畫敵名 + 怪物群 sprite
  battle_command_loop (sub_c08b)          ; 逐角色下指令 + 回合執行
  alive = party_alive_count
  if alive!=0:                            ; ── 勝利 / 繼續
     battle_redraw_field (sub_bcf2); 重載地圖; battle_leave_screen (sub_c03f)
  else:                                   ; ── 逃跑
     battle_leave_screen; battle_run_away (sub_c7d9)
```

函式體含多個近乎重複的「回合段」(本幀重入 / 多輪),`re/battle.c` 取第一輪結構作
代表,語意對齊;後段為回合重播與勝 / 逃 / 全滅三分支結算。

## 怪物群生成:`encounter_build_group`(sub_a7d5 / 主體 0xa86d)

1. **遭遇區查表**:由玩家 tile `[0x4f2f]`(X)/`[0x4f31]`(Y)算遭遇區編號——地表時
   `region = [(Y&0xf0)+(X>>4) + 0x4966]`,與門檻 `[0x52f5]` 取大;海面 / 邊界
   (`Y>=0x12c`…)走固定 region。`region-1 << 5` 索引遭遇表(每項 0x20 bytes)。
2. 取遭遇表項 → 候選怪物 id 串 `[0x231b]`(0xff 終止)。
3. **點數預算分配**:`[0x2513]=0x26`(38)為「戰鬥點數預算」;對每個候選 id,讀
   D3MNS 出現權重 `[id*0x29 + 0x0da0]`,以 `floor(budget/weight)` 算最多隻數,擲亂數
   定實際隻數,扣預算,寫敵群表 `[0x2321]`(sprite id)/ `[0x231f]`(群隻數,上限 8)。

→ game3.png 的「史萊姆 ×6」即此處生成的一群同種怪;`[0x231f]=6`、`[0x2321]` 填 6 個
同一 sprite id。

## 怪物資料結構 D3MNS.DAT

loader `sub_17ee`:整檔讀進 **DS:0x0d78**(`d3mns_buf`)。

| 屬性 | 值 |
|---|---|
| 檔案大小 | 5330 (0x14d2) bytes |
| 記錄大小 | **0x29 = 41 bytes / 筆** |
| 記錄數 | **130**(5330 / 41 = 130,整除無餘)|
| 索引 | `monster_id * 0x29`(sub_a86d 的 `mul bl, bl=0x29`)|
| 載入位址 | DS:0x0d78 |
| 出現權重欄位 | 記錄相對 **+0x28**(每筆最後一 byte);表基底 DS:0x0da0 = 0x0d78+0x28 |

確定欄位:
- **+0x28 = 出現權重 / 群量基準**(`encounter_build_group` 讀此欄,`>0` 才作遭遇候選;
  以此與點數預算 0x26 相除決定生成隻數)。

待逐筆對照(目前推定,需對 D3TXT 怪名與 DOSBox 戰鬥數值驗證):記錄 +0x00..0x27 的
HP / MP / 攻擊 / 防禦 / 速度 / 經驗值 / 金錢 / sprite 編號等欄位。`re/battle.h` 以
`d3mns_record { u8 fields[0x28]; u8 spawn_weight; }` 表示(定長 41,已驗證對齊)。

> 怪名來源:`battle_draw_names`(sub_c59b)以 `msg_id = sprite_id + 0x258` 經文字繪製器
> `lcall 111b:264` 畫出 → 怪名在 D3TXT 訊息表的 ID 偏移 0x258 起(可與 docs/03 文字表
> 交叉定位每隻怪的中文名,如 `史萊姆`)。

## 怪物 sprite:DQ3MNS.SHP

loader `sub_b19e`:handle 常駐 **DS:0x249b**(0 才開檔)。

| 屬性 | 值 |
|---|---|
| 檔案大小 | 1,377,210 (0x1503ba) bytes |
| 檔首 | **32-bit offset table**(每筆 u32,LE)|
| 表長 | 首個 offset = 0x20c → **131 筆**(130 怪 + 1 sentinel,與 D3MNS 130 筆對齊)|
| sprite 表頭 | 8 bytes(讀進 DS:0x248f)|
| sprite0 表頭 | `10 00 7b 00 00 00 00 00` = {w=16, h=123, data_pos…}|

讀取流程(sub_b19e):

```
if h_d3mns==0: open "DQ3MNS.SHP" → [0x249b]
seek(id*4)              ; offset table 第 id 筆
read 8 → [0x248f]       ; sprite 表頭 { u16 w, u16 h, u16 data_lo, u16 data_hi }
seek(data_hi:data_lo)   ; [0x2493]/[0x2495] 指像素資料
read 像素分頁 → VGA planar 驅動 seg 0x111b blit(4-bit planar,AND-mask 透明)
```

顯示(sub_b16f `mob_draw_group`):逐隻取 `[0x2321]` 的 sprite id,x 偏移累進到
`[0x2477]`,呼叫 sub_b19e 載入並貼出 → game3.png 的 6 隻史萊姆橫排即此 x 累進迴圈。

繪圖底層:VGA planar 驅動 seg 0x111b(docs/06,`out 0x3c4` Map Mask 4 plane 輪寫、
row pitch 0x54、`and es:[di]` 透明遮罩),與 BLK/SHP/packbg 同 4-bit planar 家族。

## 戰鬥背景:packbg(頁由地形決定)

> **更正(原誤記「page 0x13」)**:`battle_enter_screen` 的 `bp=0x13` 是 `lcall 1053:240`
> 這個 file op 的參數,**不是** packbg 頁。真正的背景頁由選頁 wrapper `sub_d9f8` 算出。

`battle_enter_screen`(sub_bfd1)→ `sub_c688` → **`sub_d9f8`(選頁)** → `load_packbg_page`:
- `terrain = [tile + 0x4df6]`(tile→地形類別表,DGROUP);`page = [0xd73] + terrain*8`
  (`terrain==3` → page 0x19);`[0xd73]` 初值 0。
- 每地形專屬 16 色子調色盤:`[0x25d1] = MNSBK.PAL + [0xd73]*0x30`。
- `PACKBG.SCR` 為每 `0x6e00`(640×88,4-plane row-interleaved)一張背景的**密集陣列**
  (共 132 張,全部解出乾淨);**草原戰鬥背景 = 陣列索引 22**(藍天 + 綠地,對 game3.png)。
  (`page*0x3d80` 公式僅 page 0 與陣列對齊;remake 直接用 `index*0x6e00`。)

`battle_leave_screen`(sub_c03f)重載地圖 BLK(`load_blk`)回到地表 / 城鎮
(註:離場未還原 DQ3.PAL → 原版 bug #8 戰後變黃綠,見 docs/28)。

## 視窗 / 文字座標系(從 RE 推導,非肉眼對圖)

戰鬥各框由畫框 routine **`sub_f590`(file 0x10900)** 依一個 win_rect 結構繪製;
文字由 **文字繪製器 `lcall 111b:214`/`:264`**(file 0x12734)畫。座標系:

- **`[0x716]` = 文字游標 X(VRAM byte,×8 = 像素)**;每字 X 進 **2 byte = 16px**。
- **`[0x718]` = 文字游標 Y(像素)**;換行碼 `0xfffe` 重設 X、Y += `0x10`(**行高 16px**)。
- 字模 16×16(docs/02)。文字記錄取自 `seg_txt`([0x252e],D3TXT)。

### win_rect 結構(`sub_f590` 解讀;si = rect 基底)

| 欄位 | 意義 |
|---|---|
| `+0` | 框樣式 id(交 sub_10ea6 畫邊框)|
| `+2` | 文字 X 原點(byte ×8 = px)→ [0x716] |
| `+4` | 文字 Y 原點(px)→ [0x718] |
| `+6` | 框高 |
| `+8` | 框寬(byte)|
| `+0xa` | 標題列文字記錄 id |
| `+0xc` | 內容行數(迴圈次數)|
| `+0xe` | 內容列文字記錄 id(逐行/逐欄)|
| `+0x10` | footer 文字記錄 id |
| `+0x14` | `0xfe` = **多欄模式**(每欄重設 Y、X 進 +0x16)|
| `+0x16` | 多欄 X 進(byte;狀態框 = 0xa = **80px/欄**)|

### 隊伍狀態框(`battle_draw_status` + win_rect [0x3e9c],RE 值)

- 文字原點 **X = 0x13 byte = 152px、Y = 8px**;**4 欄**(隊伍人數),**欄距 0xa byte = 80px**;
  框高 = `人數×0xa + 4`;標題/內容/footer 記錄 = `0x191/0x192/0x193`。
- 即 game3.png 上方「勇者|武鬥家|僧侶|魔法師」4 欄,每欄含名 / H+HP / M+MP / 職業+等級。

### 戰鬥訊息 / 敵名(`lcall 111b:264`,win_rect [0x3e6e])

- 以基準游標 `[0x3e70]`/`[0x3e72]`(`battle_enter_screen` 設 0x12/0xf8)起算:
  X = `[0x3e70]+2`、Y = `[0x3e72]+0x10 + 行*0x10`;敵名記錄 = `sprite_id + 0x258`。

> remake `dq3_battlescene` 依此座標系排版(字/行 16px、狀態 4 欄 80px),不再肉眼對圖。

### 框邊框 = 框線字模(rec 0x191/0x192/0x193)

`sub_f590` 先 `sub_10ea6`(**備份框下畫面供關框還原**,非畫邊框),再以 win_rect 的標題/
內容/footer 記錄畫出框。狀態框框體**藏在文字記錄裡的框線字模**(D3TXT00,渲染後即成框):

| 記錄 | 內容(glyph) | 角色 |
|---|---|---|
| `0x191` | `47 │54│ │54│ │54│ 50` | 左欄:`┌ │ │ │ └` |
| `0x192` | row0/4=`48×5`(─);row1=`22`(H);row2=`27`(M);row3=class 位 | 每員欄:`─` 頂底 + H/M/職列(數值由 fill `0x8222` 填)|
| `0x193` | `49 │54│ │54│ │54│ 51` | 右欄:`┐ │ │ │ ┘` |

框線 glyph(由記錄反查,docs/02 字模):`47=┌ 48=─ 49=┐ 50=└ 51=┘ 54=│`。
H/M/名/等級**值**不在記錄內,由 fill 函式 `0x8222` 讀隊伍緩衝 `[0x063a]` 填入。
remake 目前以單線白框 + 直接填值(視覺等價、位置已照 RE 座標);若要逐字格復刻框線字模,
需把框尺寸改成字格網格對齊(留待)。

## 回合 / 指令選單:戰 / 逃 / 防 / 道具

`battle_command_loop`(sub_c08b):本回合逐角色下指令。

- `sub_c169` 開「戰 / 逃 / 防 / 道具」指令選單(二級派發,對應 game3.png 左下
  `戰 鬥 / 逃 跑 / 防 禦 / 道 具`)。
- 結果旗標 `[0x2517]`:`0`=續行 / `1`=防禦或無事 / `2`=逃跑成功(跳結算)。
- 逐角色 loop(`[0x2515]` = 待下指令角色數):畫指令游標、依選擇派發攻擊 / 防禦 /
  道具 / 逃。
- 結算分支:
  - 敵全滅(`[0x24b3]==0`)→ `di=0x161` 勝利訊息 + 結算經驗 / 金錢(`sub_c954` /
    `sub_c425`)。
  - 我方全滅(`[0x24b4]==0`)→ `di=0x169` 全滅訊息 + 等鍵。

`party_alive_count`(sub_c5f6):掃 `[0x5077]` 名隊員(實體指標 `[bx+0x4f15]`),
HP(`[di+0x16]`)≠0 者計入,並設 0x80 已處理旗標;回 0 = 全滅。

`battle_setup_party`(sub_c8c6):4 名隊員各從結構 `[bx+0x4f0d]` 取 5 個 word
(+0x1a/+0x1c/+0x1e/+0x20/+0x24,推定 HP/MaxHP/MP/MaxMP/攻擊或經驗)+ 3 個 byte
(+0x01/+0x2e/+0x2f,推定 等級 / 狀態),攤平到戰鬥工作緩衝 `[0x063a]`;另掃 +0x3a
起 8 格裝備設睡眠(`0x4d`→bit0x800)/ 中毒(`0x11`&`0x80`→bit0x4000)狀態旗標。

→ game3.png 上方狀態框(H=HP、M=MP、勇/武/僧/魔 各一行)即 `battle_draw_status`
(sub_c572,框高 = 隊伍人數×0xa+4)+ `battle_setup_party` 攤平的 HP/MP word。

## 與 game3.png 逐項對照

| game3.png 畫面元素 | 對應函式 / 資料 |
|---|---|
| 藍天草原背景 | packbg.scr 地形選頁(草原=陣列索引22)(`battle_enter_screen` sub_bfd1)|
| 6 隻史萊姆橫排 | `encounter_build_group`(群量)+ `mob_draw_group`(x 累進 blit)|
| 史萊姆 sprite 圖 | DQ3MNS.SHP(offset table → 8B 表頭 → planar blit,sub_b19e)|
| 「史萊姆 6」敵名 + 隻數 | `battle_draw_names`(msg=sprite_id+0x258,sub_c59b);`[0x231f]`=6 |
| 上方 勇者/武鬥家/僧侶/魔法師 H/M 框 | `battle_draw_status`(sub_c572)+ `battle_setup_party`(sub_c8c6)|
| 左下 戰/逃/防/道具 指令 | `battle_command_loop`(sub_c08b)→ 指令選單 sub_c169 |

→ 反組譯結構與 game3.png **逐項吻合**:背景、敵群、敵名、隊伍狀態、指令選單皆對上。

## 反編譯產出

- `re/battle.h` — D3MNS 記錄結構、SHP 表頭結構、戰鬥全域(DS offset)、函式宣告。
- `re/battle.c` — 進入點 / 主流程 / 遭遇生成 / sprite 載入 / 背景切換 / 狀態列 /
  指令回合的可讀 C(real-mode large-model;runtime far-call 以 exe.h wrapper 表達;
  核心遊戲函式 forward-declare)。
- `tools/re_battle_dis.py` — 戰鬥函式反組譯 + D3MNS.DAT / DQ3MNS.SHP 結構 dump。

C 風格同專案慣例:求結構正確、可審閱,非位元精準;逐函式對 DOSBox 驗證為後續步驟。

## 殘留 / 待釐清

- **D3MNS.DAT 記錄 +0x00..0x27 逐欄語意**:HP / MP / 攻防 / 速度 / 經驗 / 金錢 / sprite
  編號的精確位移,需對 DOSBox 戰鬥數值 + D3TXT 怪名(msg 0x258 起)逐筆驗證。
- **DQ3MNS.SHP sprite 表頭 8 bytes 的 w/h/data_pos 確切欄序**:目前以 sprite0
  `{16,123,…}` 推定;需貼一隻怪物圖驗證(可仿 docs/11 的 CTY 貼圖驗證法,用 mnsbk.pal
  160 色 palette 著色)。
- **回合行動 / 傷害計算**:`sub_c08b` 二級指令派發的各動作子函式(攻擊傷害公式、
  呪文、道具效果)尚未逐一反編譯;目前到「指令選單 + 勝 / 逃 / 全滅結算」層。
- **連戰 / 多輪回合**:sub_bddf 後段重複回合段的迴圈條件(何時換下一回合)語意待補。
- 戰鬥背景音樂 mbg.mcx / ebg.mcx 的播放觸發點(Sound Blaster 段 0x129c)未在本任務範圍。
