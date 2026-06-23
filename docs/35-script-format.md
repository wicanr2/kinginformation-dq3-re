# DQ3 腳本 / 事件 / 轉場格式(重新整理,靜態 RE 全解)

> 把散在 docs/31/34 的事件系統重新逐 handler 反組譯、整理成一份可直接 implement 的格式規格。
> 重點修正:**門/階梯/出城的 dispatch 是 section header `+0xc` 的「轉場表」純靜態資料**
> (不是 type-2 事件)——先前以為「68 個多 section 大城無 type-2、門追不到、要 debugger」是
> 誤判,門就在 `+0xc` 表裡。以 `assets_raw/DQ3.EXE` 逐段反組譯 + CTY00 實檔驗證。

## 位址對照(避免再踩 file/logical 陷阱)

`tools/re_disasm.py` / `tools/dis.sh` 印 **file offset**;`exe_funcs.json`/`sub_XXXX` 用 logical。
`file = logical + 0x1370`。本檔位址一律 **file offset**。DGROUP 變數位址(`[0x….]`)為段內 offset。

## 一、CTY section header(每個 section 自帶)

`section_base = CTY[ [0x256a]*2 ]`(`word[i]` = section i 的 base;`0xffff`=空;無 count 前綴)。
header 欄位(相對 `section_base`,皆為 2-byte 指標,相對 section_base):

| off | 內容 | 反組譯出處 |
|---|---|---|
| +0 / +2 | **NPC/物件清單**指標(`[0x526c]` 選:Y<0x12c 用 +0,否則 +2)| 0x4549 |
| +8 | **examine 事件表**指標(調べる/寶箱/給道具/傳送;`0xffff`=無)| 0x9d51 |
| **+0xc** | **轉場表**指標(門/階梯/出城;`0xffff`=無)★ | 0x48ae / 0x4a03 |
| +0xe | **layout** 指標 → `{w(2), h(2), tiles…}`,tiles 在 +4 | 0x443f |
| +0x10/+0x11 | 地圖旗標 / `[0xd77]` 遭遇安全旗標(0=可遇敵)| docs/13 |
| +0x13/+0x14 | 玩家 spawn X / Y | 0x9d51 |
| +0x15.. | 地圖參數;`[0xd71]`=地圖 id | |

> ★ `+8`(examine 事件)與 `+0xc`(轉場)是**兩張獨立的表**,別混。大城常 `+8=0xffff`(無 examine
> 事件)但 `+0xc` 有一整排門。

## 二、tile 格式與屬性

tile = u16:**低 byte = BLK tile index**;**高 byte** = 屬性/事件位元(見下)。
每個 tile index 另有一個 **屬性 word**:`attr = [ tile_low*2 + 0x308e ]`(來自 `BLKBMn.DAT`)。

走動/互動時對 tile 高 byte + attr word 的位元判讀(玩家移動 handler `0x2e0e`、examine `0x9cf6`):

| 來源 | 位元 | 語意 |
|---|---|---|
| tile 高 byte | `0x20` | **NPC 佔位**(NPC 載入時蓋上 `0x20\|slot_idx`,見四)|
| tile 高 byte | 低 5 bit `0x1f` | **subid**(索引 examine 事件表 / 轉場表)|
| attr word | bit0 `0x0001` | **阻擋(牆)** → 撞牆音效(0x2eb0)|
| attr word | bit3 `0x0008` | **有 examine 事件**(`test [bx+0x308e],8`,0x9d2c)|
| attr word | 高 3 bit `0xe000` | **轉場 tile**(門/階梯)→ 種類 = `rol+and 3` 存 `[0x258a]`(0x2ea1)|
| attr word | bit5 `0x0020` | 繪圖 overlay 屬性(視窗繪製器 0x334d)|

## 三、examine 事件系統(`+8` 表)— 調べる / 寶箱 / 給道具 / 傳送

踩到/面向 tile 按 Enter →(地表 `0x9cf6` / 城鎮 `0x9f7f`)讀 tile,`attr&8` 才有事件。
事件項由 tile 高 byte 低 5 bit 當 subid 索引:

```
ev_tbl = section_base + word[section_base+8]
entry  = ev_tbl + subid*4 + 1          ; ★ 表首 1 byte 跳過(count/header)
type   = entry[0]                       ; 1 byte
param  = entry[1..2]                    ; u16
p2     = entry[3]                       ; 1 byte = 旗標 id(見「旗標閘」)
```

派發(`0x9d86`,以 `type` 分):

| type | handler(file)| 動作 |
|---|---|---|
| 0 | 0x9ec1 | **寶箱/調べる**:`param==0` → rec 0x108「在腳邊調查‥」+ 0xf3「可惜是空的」;`param!=0` → 給道具 param + 清旗標 |
| 1 / 3 | 0x9de9 | **給道具進隊伍背包**:掃隊伍 8 格找 `0xff` 空格存入 param;滿了印 rec 0x13a「行李滿」。type3 另設 `[0x726]=1`(續事件鏈)+ 印 rec 0xc11 |
| 2 | 0x9f32 | **傳送/門**(舊式):目的 = `[param*3 + 0x4ea0]` = `{dest, X, Y}`(warp 表)→ `call 0xd1f9` 轉場 |
| else | 0x9f71 | 給道具 `[0x24fc]` + rec 0x1e4 |

> ⚠ **修正 docs/31 誤標**:type 1/3(`0x9de9`)是**給道具**(找 `0xff` 空格存入),**不是**
> 「找鑰匙的條件檢查」。`0x9de9` 全段已反組譯確認(0x9de9–0x9ec0)。

### 旗標閘(一次性事件的核心)— bit 陣列 `[0x4f70]`

每個事件項的 `p2` 是**旗標 id**。事件觸發前先測旗標,觸發後清旗標:

```
分派(0x9d7c):  [0x24fc]=p2;  if test_flag(p2)!=1 → 跳過(事件未 armed/已消耗)
handler 尾段:   clear_flag(p2)              ; 消耗 → 不再觸發
```

旗標函式(`id` in bx):

| 函式(file)| 動作 |
|---|---|
| 0x825e | **set_flag**:`[0x4f70 + id/8] \|= (0x80 ror (id&7))` |
| 0x8264 | **clear_flag**:`[0x4f70 + id/8] &= ~(0x80 ror (id&7))` |
| 0x8279 | **test_flag**:`al = ([0x4f70+id/8] & (0x80 ror (id&7))) ? 1 : 0` |

**語意**:旗標 **set = 事件 armed(可觸發)**;清 = 已消耗。NPC 可見性也用同一 bit 陣列(見四)。

### tile 外觀隨旗標改變(寶箱開過變空)— `0x4703`

section 載入時全圖預掃(`0x46ac`),對每個事件 tile 測旗標:
- 旗標 **set(armed)** → tile 不動(顯示未開的寶箱/物件)。
- 旗標 **清(消耗)** → 改寫 tile:`tile_low += 1`(下一個 BLK index = 開過外觀)+ 清高 byte subid 位
  → 不再觸發。**寶箱關 = index N、開 = N+1**。

## 四、NPC / 物件(`+0`/`+2` 表)

清單格式:`{ count(1 byte), records… }`,每筆 **7 byte**(載入器 `0x4540`–`0x45e2`):

| byte | 欄位 | 說明 |
|---|---|---|
| 0 | **X** | 地圖格 X(位置 = `Y*width + X`)|
| 1 | **Y** | 地圖格 Y |
| 2 | **sprite 群 id** | 載入時去重映射到 sprite 快取槽(`0x265d` 表;`0x4634`)|
| 3 | **方向 / 移動控制** | bits0-1=**朝向**(0-3)、**bit2 `0x04`=移動開關**(set=隨機走動 / clear=靜止)、bit7 `0x80`=暫時凍結。★ 舊註「bits0x38→handler 0x62e9/0x6355/0x839f」是**誤標**(那些位址是劇情/旗標碼,非 mover);真 mover = `L02025`(file 0x3395),見 §九 |
| 4 | **對話 / 參數 id** | → `[0x2597]`(NPC 互動時用,接對話)|
| 5 | **flags** | 低 3 bit = **可見性旗標 id**(查 `[0x4f70]`;旗標清 → NPC 隱形不載入,0x457c)|
| 6 | b6 | 載入時被覆寫:存「NPC 站的那格原始 tile byte」供 NPC 移動後還原(`[di+6]`)|

載入動作(每筆):
1. `flags` 低 3 bit 查旗標 → 清則跳過(NPC 不存在,如劇情前/後的 NPC)。
2. 7 byte 複製到 `[0xb66 + slot*8]`(8-byte 工作槽;第 8 byte `[di+6]` 存原 tile)。
3. **把 NPC 蓋進地圖**:`map[Y*width+X]` 高 byte = `0x20 | slot_idx`(供碰撞:玩家撞 `0x20` 位 → NPC 互動 `0x2ec5`;供繪圖)。
4. `byte2` 經 `0x265d` sprite 快取去重(多個同型 NPC 共用一份 sprite)。

**NPC 移動**:見 §九(per-frame 隨機走動 mover,**非** `0x6240` —— 該位址是場景 init,舊標誤)。
→ remake 要做動的 NPC 時實作這層;靜態擺放可先只用 X/Y/sprite/flags。

## 五、轉場系統(`+0xc` 表)★ 門 / 階梯 / 出城 — 本次主解

走到 attr `0xe000` 的轉場 tile → 移動 handler 設 `[0x4f46]` bit0(`0x2e8c`)→ 移動後 `call 0x488f`。
`0x488f` 讀 **section_base + 0xc 的轉場表**,以 tile 高 byte subid(`[0xb55]&0x1f`)索引,
每項 **4 byte**:

```
trans_tbl = section_base + word[section_base+0xc]
entry     = trans_tbl + subid*4
destCTY   = entry[0]      ; = [0x256c] 比對
destSec   = entry[1]      ; = [0x256a];  0xff=出城回地表  0xfe=特殊([0x5051]=1)
destX     = entry[2]
destY     = entry[3]
```

分派(`0x48d7`–`0x4972`):

| 條件 | 動作 |
|---|---|
| `destSec==0xff` | **出城回地表**(`0x49f9`):X/Y 用 entry[2]/[3],皆 0 則用存的地表位置 `[0x5053/5055]`;`[0x5051]`=1/3 → **Y += 0x12c(300)= 地下層**(對上 cty_loc 地下 Y−300)|
| `destSec==0xfe` | 特殊(`[0x5051]=1`)|
| `destCTY != [0x256c]` | **跨 CTY 載入**(`0x49d1`):設新 CTY,從 `[CTY*4 + 0x748]`(cty_loc 表)取地表座標放人 → `call 0x43d1` load_cty |
| `destSec == [0x256a]` | 同 section,只設玩家 X/Y |
| 否則 | **同 CTY 換 section**:`[0x256a]=destSec` → `0x481b`/`0x4386` 重載 section → 設 X/Y |

`[0x258a]` = 轉場種類(attr 頂位元 `rol+and 3`):1=門、2/4/5=階梯類(`0x488f` 內分別播音效/動畫)。

### CTY00(阿里阿罕)實檔驗證

```
section0 (42×43 全鎮): ev@+8=0xffff(無 examine 事件)  TRANS@+0xc 有 4 門:
  [0]{cty0,sec1,(14,6)} [1]{cty0,sec2,(8,2)} [2]{cty0,sec3,(7,1)} [3]{cty0,sec4,(10,10)}
section1 (16×12 室內):
  [0]{cty0,sec0,(19,36)} 出回鎮上   [1]{cty16,sec0,…} 跨 CTY   …
```

⇒ **所有城鎮(含先前卡住的多 section 大城)的建築門/室內外/上下樓/出城,全在 `+0xc` 表**,
4-byte `{destCTY,destSec,X,Y}`,純靜態。remake 直接讀此表即可 dispatch,**不需動態分析**。

## 六、寫死座標的劇情事件(城鎮 examine `0x9f7f` + scripted-event runner)

少數劇情點不走表,而是寫死座標判斷:
- `0x9f7f`:`if [0x4f2f]==X && [0x4f31]==Y` → 設 `[0x2593]` + 跑對話序列(如某鎮特定格)。
- **scripted-event runner `0xabb2`**:`id=[0x722]` → `call [id*2 + 0x3baa]` 跳表 → handler。
  **#2 彩虹水滴合成 = id 83**(handler `0x776c` 尾段 `0x77ce`,docs/31)。
- **★ 祠堂觸發座標已解(靜態,改走 CTY 資料)**:`CTY93` section0(17×17)NPC 清單只有**一隻祭司**,
  在 **(8,8)**(`ctrl=0x03`=靜止、`dlg=0x4b`/`0x4d` 兩變體=合成前/後依旗標、`flags=0x28` 可見性閘);
  玩家 spawn=(15,8)。調べる (8,8) 祭司 → 合成。三證據齊:CTY 資料 + 劇本 `txt08[32]`「這裡是神聖的
  祠堂」+ 地理 docs/32(CTY93 下層 SE)。剩微鏈:NPC `dlg` id → scripted-event 83 的精確派發(資料驅動)。

## 七、對 remake 的落地對映

| 機制 | 資料來源 | remake 動作 |
|---|---|---|
| examine 事件 | section `+8` 表,4-byte `{type,param,p2}` | 已部分(`dq3_scene.events`);補旗標閘 + tile+1 |
| 旗標 | `[0x4f70]` bit 陣列 | 加 `dq3_flags`(set/clear/test,id→bit)|
| NPC | section `+0`/`+2`,7-byte | 解析 X/Y/sprite/flags;可見性查旗標;(移動之後做)|
| **門/階梯/出城** | section **`+0xc`** 表,4-byte `{destCTY,destSec,X,Y}` | **`dq3_town_load` 解析 `+0xc` → `scene.transitions[subid]`;主迴圈踩轉場 tile → 依表 dispatch** |
| 跨 CTY 落點 | `0x748` cty_loc(已 extract)| 出城/跨 CTY 用 cty_loc 地表座標 + 地下 Y+300 |

## 八、鑰匙門(locked door)★ 已靜態全解 — 內嵌在轉場 handler,非獨立機制

> 用「鑰匙道具 id 反查」破解(方法論 docs/00 §1/§4):先解出鑰匙 id,再全檔找對它的比較,
> 一步命中,不必沿舊位址盲掃。**糾正舊標**:`re/player.c` 行註把 attr **低 byte** bits6-7
> (`0x00c0`)籠統標為「角落/出口 warp 類別」;door handler 證實**低 byte bits6-7 = 門所需鑰匙等級**。
> (player.c 移動碰撞實際讀的是 attr **高 byte** bits6-7 = word `0xc000` 做角落/warp_pending 判定,
> 是**不同欄位**,兩者勿混——這也是位址/欄位混標導致誤判的一例,見 docs/00 §3。)

**鑰匙道具**(D3TXT00 道具名解出):`0x55` 盜賊的鑰匙、`0x56` 魔法鑰匙、`0x57` 最終鑰匙。
**等級** = `id − 0x54` → 1 / 2 / 3(盜賊 < 魔法 < 最終)。

開門在城鎮轉場 handler `0x488f` 內,兩段(全靜態,純資料判定):

1. **算全隊最高鑰匙等級 → `[0x2593]`**(`0x48c3`,file 0x5c33;由 `0x4816` 呼叫):
   ```
   外圈 [0x5077] 名隊員:si = [memberidx*2 + 0x4f15];si += 0x3a   ; 隊員道具陣列(8 格 u16)
   內圈 cx=8:  dx=[si];  if 0x55<=dl<=0x57:  tier = dl-0x54
              if tier > [0x2593]: [0x2593] = tier              ; 取最大(=最高階鑰匙)
   ```
2. **讀面向 tile,夠級就開門**(`0x4906`,file 0x5c76;由 `0x3f7d`/`0x4862`/`0x4890` 呼叫):
   ```
   if [0x4f2d]!=1: skip                       ; 僅城鎮
   面向格 = 玩家(X,Y) + dir 位移([dir*2+0xb35] / +0xb37);di = layout offset
   attr = [ tile_low*2 + 0x308e ]
   if (attr & 0x00c0)==0: skip                 ; ★ bits6-7=0 → 非鎖門
   need = (attr_low >> 6) & 3                   ; ★ 門所需鑰匙等級(rol bl,1 ×2; and 3)→ [0x2515]
   if [0x2593] < need: skip(維持鎖住)          ; 隊伍最高鑰匙 < 門需求 → 開不了
   ── 開門:[di] = tile_low & 0x1f(清事件高位)   ; 改 tile 為通行
            [di+1] &= 0xe0(清屬性低位)
            [0x2591] = 0x3744 + [0x4f09] + [dir*2+0xb4d](開門訊息);lcall 0x10b6,0x19c 播動畫/音效
   ```

> 語意:門 tile 的 attr 低 byte bits6-7 編入**所需鑰匙等級**(1/2/3);隊伍持有 ≥ 該級的鑰匙即可開
> (最終鑰匙開所有門)。開門 = 就地把 tile 改成通行並清事件/屬性位元(`seg_cty` 寫回),非走 +0xc 表。
> remake:`dq3_town_load` 已解 attr;另加 `dq3_member` 鑰匙等級掃描 + 面向 tile 開鎖(改 tile + 清 attr)。

## 九、NPC 移動(per-frame 隨機走動 mover)★ 已靜態全解

> 由「會動 vs 不會動 NPC」現象反推:差別在 **slot[3] bit2(0x04)移動開關**。
> mover 確認用 RNG(file 0xfa39);**糾正舊標**:doc 標的 `0x6240`(分派器)/`0x62e9`/`0x6355`/`0x839f`
> (handler)全是場景 init / 劇情旗標碼,非 mover。真 mover = **`L02025`(file 0x3395)**,逐 NPC 呼叫。

**slot[3] 移動控制**:bits0-1=朝向(0-3)、**bit2 `0x04`=移動開關**、bit7 `0x80`=暫時凍結。

**步進規則**(`L02025` per NPC,每幀;`si`=該 NPC 8-byte 槽 `[0xb66]+slot*8`):
```
RNG(4);  if rng!=1: return                 ; ★ 節流:每幀 1/4 機率才評估這隻
if [si+3] & 0x80: return                    ; 凍結中
if !([si+3] & 4): return                    ; ★ bit2 清 = 靜止 NPC(不動)
dir = RNG(4)                                ; 隨機方向
if dir == ([si+3] & 3):  try_step(dir)      ; 與當前朝向同 → 直接走一步
else:
    if RNG(20)==1:                          ; 1/20 機率才轉向
        new = (cur ± 1) & 3                  ; ±1 視 npc_index 與 count/2
        [si+3] = ([si+3]&0xfc) | new;  try_step(new)
    ; 否則本幀不動

try_step(dir):                              ; 落步前 5 道閘(任一不過 → 取消 / 反向)
  target = (slot.X,slot.Y) + ([dir*4+0xb35], [dir*4+0xb37])   ; ★ 與玩家同一張方向位移表
  ① target 界內(0..[0xb28]寬 / 0..[0xb2a]高)
  ② |玩家X − targetX| ≥ 3 且 |玩家Y − targetY| ≥ 3            ; ★ 不走進玩家 3 格內
  ③ target ≠ 玩家所在格;若 == → 朝向反轉(dir+2)、不走
  ④ target tile 高 byte bit `0x20` = 已有別的 NPC → 擋
  ⑤ target tile attr bit0(`[idx*2+0x308e]` low!=0)= 牆 → 擋
  全過 → 更新 slot X/Y + 還原舊格(slot[6])+ 重蓋新格(0x20|slot)
```

> 語意:NPC 每幀僅 ¼ 機率被評估(走得慢、不規則);移動者多沿當前朝向走、偶爾(1/20)轉向;
> 不會貼近玩家(≥3 格)、不會疊在另一 NPC 或牆上。**靜態 NPC = bit2 清**(村裡站著不動的人)。
> remake:NPC 槽加 byte3;town tick 每幀對 bit2 set 的 NPC 跑此步進(RNG + 5 閘 + 改 tile)。

## 待 RE(縮小後)

- scripted-event 祠堂觸發**座標已解**(CTY93 (8,8) 祭司,見 §六);剩 NPC dlg id → event 83 派發微鏈。
- ~~NPC 移動~~ **已解(見上 §九)**;~~鑰匙門~~ **已解(§八)**。
- ~~鑰匙門~~ **已解(見上 §八)**。
