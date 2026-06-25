# 精訊 DQ3 二進位結構規格(canonical 資料字典)

> 把 RE 出來的**所有二進位格式**集中成一份正式規格,供 remake/工具引用,避免散落各 docs。
> 規則:① 每個結構列「來源(檔/EXE 位址)+ 欄位 + 型別 + 語意」② 偏移用十六進位 ③ 改格式先改這裡。
> 詳細推導見對應 docs;本檔只放**定版結構**。位址:EXE 為 file offset(logical+0x1370);DGROUP 為段內 offset。

## 1. CTY 檔(`CTYxx.DAT`)— 城鎮/迷宮

```
檔首: word[i] = section i 的 base offset(相對檔首);word[0]=section0;0xffff=空。無 count 前綴。
      section 數 = word[0] / 2(section0 緊接 offset 表後)。
```

### 1.1 section header(相對 section_base,各 section 自帶)— docs/34/35/40/42

| off | 型別 | 欄位 | 語意 |
|---|---|---|---|
| +0x00 / +0x02 | word ptr | NPC/物件清單指標 | `[0x526c]` 選(Y<0x12c 用 +0,否則 +2)|
| **+0x06** | word ptr | **設施指標表** | 商店/旅社/教會/記錄(docs/40)|
| +0x08 | word ptr | examine 事件表指標 | 調べる/寶箱/給道具/傳送(0xffff=無)|
| +0x0c | word ptr | 轉場表指標 | 門/階梯/出城(0xffff=無)|
| +0x0e | word ptr | layout 指標 | → `{w(2),h(2),tiles…}`,tiles 在 +4 |
| +0x10/+0x11 | byte | 地圖旗標 / `[0xd77]` 遭遇安全旗標 | 0=可遇敵 |
| +0x13/+0x14 | byte | 玩家 spawn X / Y | |
| **+0x17** | byte | **對話 bank** | → `D3TXT0<bank>`(docs/42)|
| +0x18 | byte | `[0xd71]` 地圖 id | |

所有指標相對 section_base;表項偏移亦相對 section_base。

### 1.2 NPC 記錄(`+0`/`+2` 清單:`{count(1B), records}`,每筆 7 byte)— docs/34/42

| byte | 欄位 | 語意 |
|---|---|---|
| 0,1 | X, Y | 地圖格座標(位置 = Y*寬+X)|
| 2 | sprite 群 id | BLS sprite key(載入去重)|
| 3 | ctrl | bits0-1=朝向、bit2(0x04)=移動開關、**bits3-5(0x38)=互動子型**、bit7=凍結 |
| 4 | byte4 | **互動子型 0/1**=對話 rec;**2**=scripted-event handler idx;**3-7**=設施 idx(docs/40/42)|
| 5 | flags | 低 3 bit=可見性旗標 id(`[0x4f70]`;清=NPC 不載入)|
| 6 | b6 | 執行期覆寫(站立底 tile)|

**互動子型** = `(ctrl>>3)&7`。

### 1.3 設施 block(`+6` 設施表[byte4] 指到的 block)— docs/40

```
block[0] = type   ; 0=旅社 1=武器/防具店 2=道具店 3=教會 4=記錄點(跳表 DGROUP 0x623)
── type 1/2(商店)──
block[1] = count          ; 品項數
block[2..] = item_id[count] ; = ITEM.DAT index;名 = D3TXT00 rec (id+1)
── type 0(旅社)── block[1..2] = 住宿費 raw(× 人數)
```
設施表長度 = (首 block 偏移 − 表頭)/2。dispatcher: file 0x839f(NPC byte4 索引)。

### 1.4 examine 事件表(`+8`:`{count(1B), entries}`,每筆 4 byte)— docs/31/35

```
entry = { type(1B), param(u16), flag(1B) }
type 0 = 寶箱/調査:param==0 空;param!=0 給道具 param(handler 0x9ec1)
type 1/3 = 給道具進背包 param(0x9de9)
type 2 = ★傳送:目的 = warp 表 [param*3 + 0x4ea0](§4)→ 切場景(handler 0x9f32→0xd1f9)
flag = 一次性旗標 id([0x4f70] bit;set=armed/未取)
```

### 1.5 轉場表(`+0xc`:無 count,每筆 4 byte)— docs/35

```
entry = { destCTY(1B), destSec(1B), X(1B), Y(1B) }   ; 由轉場 tile subid 索引
destSec 0xff = 出地表 overworld;0xfe = 出下層(設 [0x5051]=1 → Y+=300);
        其餘 = 同CTY換section / 跨CTY。
表長:讀到 layout 偏移前為止(無 count)。dispatch: file 0x488f / 0x49f9。
```

## 2. overworld 地圖(`DQ3CON.MAP` 地表 / `DQ3UND.MAP` 下層)— docs/04/43

```
檔首: w(u16), h(u16);之後 w*h 個 u8 = 每格 BLK tile index。
層: 玩家 overworld Y([0x4f31]) < 0x12c(300) → DQ3CON;≥300 → DQ3UND 讀 (Y−300)。
    (兩檔柄常開 [0x1f1]con/[0x1f3]und;判定 file 0x3fac。)
tile 可走性: attr = BLKBM.DAT[tile*2](u16);bit0(0x0001)=阻擋(步行);海 tile(88)bit0=1。
```

## 3. 全域表(EXE / DGROUP)

| 表 | DGROUP off | 結構 | 語意 |
|---|---|---|---|
| ITEM.DAT(載入)| 0x1f9 | 128×7B | 道具屬性(b0攻/b1防/b2-3價/b4類/b5旗/b6職)docs/22 |
| cty_loc(0x748)| 0x748 | 100×4B `{X, b1, Y(u16... 實 3B)}` | overworld 進城:index=CTY;map(0地表/1下層/0xff迷宮)|
| map_blknum | 0x0a04 | 89×1B | 每 CTY 的 BLK 號(1-5)|
| **warp 表** | **0x4ea0** | 113×3B `{dest, X, Y}` | type-2 傳送目的(dest=CTY、X/Y 落點)§4 |
| 設施跳表 | 0x623 | 5×2B | 設施 type → handler |
| scripted-event 跳表 | **0x3baa** | word/項 | runner 0xabb2 event-id → handler |
| 子型2 跳表 | **0x3bb4** | word/項 | = 0x3baa 偏移 5 項;NPC byte4 → handler |

> `0x3baa` 與 `0x3bb4` 是**同一張表偏移 5 項**:子型2 byte4 = runner event_id − 5(docs/44 §5c)。

### 3.1 scripted warp struct(0xd1f9 介面,5 byte)— docs/44 §4/§5b

```
struct @ <addr> = { b0(1B), cur_map(1B), dest(1B), X(1B), Y(1B) }
type-2: 動態建在 [0x2521];固定 scripted warp: 寫死 struct 在 0x4eb5/0x4ed5/…(0x4ea0 區內)
dest = 目的 CTY/map、X/Y = 落點。執行: call 0xd1f9。
```

### 3.2 overworld 旗標條件 portal(函式 0x396e)— docs/44 §7

```
玩家在某 overworld 位置(cty_loc[i] 或寫死 X,Y)→ 依 test_flag 鏈選目的 CTY:
  cty_loc[58](210,64): flag 0x23→58, 0x47→59, 0x42→60, 0x48→61, else→83
  (82,165): flag 0x13→75, else→47
  存位置: →36
  (54,129): flag 0x4d→71, else→72
dispatcher 0x39cb: [0x256c]=bp(目的CTY); call 0x4378(load_cty)。
語意 = 同一 overworld 點依進度載不同城(城鎮變體);**非下降**(目的多為地表城)。
```

## 4. 文字容器(`D3TXT00-09.TXT`)— docs/03

```
檔首: u16 指標表(ptr[0]=表大小);record i = data[ptr[i]..ptr[i+1]],2-byte LE 值串,0xffff 終止。
值 <1476 = D3TXT00.FON 第 N 個 16×16 字模;≥0xffed = 控制碼:
  0xfffe/0xfffd 換行、0xfffc 捲動、0xfff5/0xfffb/0xfff9/0xfffa/0xfff6/0xffed = 插值(+1 word 參數,須消耗)。
道具名 = D3TXT00 rec(item_id+1);NPC 對話 = section bank D3TXTnn rec(byte4)。
```

## 維護規則

- 新解出/修正一個結構 → **先改本檔**,再改 generator / remake。
- generator(`tools/gen_*.py`)是本規格的可執行落地;欄位語意以本檔為準。
- 改 EXE 全域表位址/結構 → 同步更新「來源位址」欄。
