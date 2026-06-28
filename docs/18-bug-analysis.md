# 精訊版 DQ3 bug 反組譯定位與修正

精訊版 DQ3 是 1993 年由精訊接手中文化、但因兩家公司條件談不攏而未正式發行的地下版。
青衫先生整理的攻略明白寫著:「由於修改的不完全,本遊戲的 BUG 非常之多,當機也相當頻繁,
對於沒有修改存檔與程式能力的玩家而言,會根本無法玩完這個遊戲。」

本文件把青衫記錄在 `bugs.md` 的 7 個 bug,逐一用反組譯定位到 `DQ3.EXE` 的確切位址、
說清楚為什麼會出現那個現象,並提出修正。

> ★ **更正(2026-06,以 code/JSON 為準)**:後續 RE 已把修正範圍擴大,本文部分舊斷言過期 ——
> 現有 **4 個 EXE binary patch(#1/#2/#4/#7a)+ #3 sprite 復原 = 5/7**(`tools/build_fixed_version.py`):
> #4(勇者MP)= EXE DGROUP 成長表 2-byte 資料 patch(`file 0x1a4a8/9`,`docs/data/stat_patches.json`),
> **非** dragon0.dat、成長表**在** EXE 映像內(DS:0x4366→`file 0x1a4a6`);
> #7a(隼劍雙擊)= 同長度區段改寫(`file 0xc1fa`,`docs/data/codecave_patches.json`),**已修非未實作**;
> #6 的 255-wrap **不是** 8-bit 成長公式(該公式全程 16-bit 正確,`docs/23`)。下方舊正文凡與此衝突者以本框為準。

兩個交叉佐證來源讓定位特別有把握:

- **青衫攻略內附的官方修正碼**:攻略「修改:DQ3.EXE」一節列了 8 段當年玩家社群用的
  EXE patch(含「魔王打不死修正法」「五頭龍大王當機修正」),以及存檔 HEX 修改法。
  這些 patch 的 byte pattern 在 EXE 內都能精準命中,等於替我們的反組譯定位做了驗證。
- **完整道具編號表**:攻略「道具修改列表」給了 128 個道具的 item code(彩虹水滴=0x75、
  黃寶珠=0x6a、祈禱戒指=0x48、魔法鎧甲=0x2b、飛鷹劍=0x6e…),直接對上 `ITEM.DAT` 結構。

位址慣例:`file offset` = EXE 檔內絕對位移;EXE 載入頭 `HDR = 0x1370`(e_cparhdr=0x137),
seg0 邏輯 off = file − 0x1370。seg0 主碼約 file 0x1370..0x11370;之後是 overlay segment
(boss 遭遇表等資料落在此區,只看 file offset)。反組譯工具 `tools/re_bug_dis.py`
(docker capstone 內跑);修正版產生器 `tools/re_bugpatch.py`。

---

## 摘要表

| # | bug | 性質 | 定位 | 修正 | 信心 |
|---|---|---|---|---|---|
| 1 | 巴拉摩斯打不死 | 卡關 | sub_aa69 file 0xbe02 | binary patch(2 byte) | 高 |
| 2 | 彩虹水滴誤拿黃/銀寶珠 | 卡關 | 合成事件 file 0x77e7 | binary patch(1 byte) | 高 |
| 3 | 九頭龍/五頭龍戰當機 | 卡關 | boss 遭遇表 file 0x1b02c + 缺 sprite | binary patch(3 byte) | 高 |
| 4 | 勇者 MaxMP 成長固定 +1 | 樂趣 | 成長表 EXE DGROUP file 0x1a4a6 | **EXE-data binary patch(file 0x1a4a8/9,已修)** | 高 |
| 5 | 高等級升級錯亂 | 樂趣 | 門檻查表 sub_ecdb file 0xecdb | C 層(越界,需 clamp) | 高(根因) |
| 6 | 數值 255 溢位 | 樂趣 | 成長公式 file 0xed3c | C 層(型別/clamp) | 中 |
| 7a | 隼劍只打一次 | 樂趣 | 攻擊碼/裝備特效 sub_c8c6 | **binary patch(file 0xc1fa 區段改寫,已修)** | 高(根因) |
| 7b | 魔法鎧甲無抗魔 | 樂趣 | 扣血器 0xc098/0xba71/0xc22e | C 層(功能未實作) | 高(根因) |
| 7c | 祈禱之戒永不壞 | 樂趣 | use handler file 0x5af4 | 與前提衝突,不需 patch | 高(衝突) |

「binary patch」= 已收進 `tools/re_bugpatch.py`、同長度 in-place、可直接套用驗證。
「C 層」= 功能未實作或需改型別/加 clamp,不是 1~3 byte 能翻轉,留給日後 SDL remake。

---

## Bug 1:巴拉摩斯打不死(卡關)— 信心:高

**現象**(青衫):打巴拉摩斯時,全隊被牠的「巴西魯拉(空間追放,把人吹出戰場)」吹走後,
卻莫名其妙判定玩家贏 —— 也就是勝負判定錯亂,正常打死巴拉摩斯反而不會結束。青衫原文:
「由於 BUG 的關係,巴拉摩斯是打不死的,除非剛好被毒針刺中要害,否則便必須靠修改。」
(「毒針刺中要害」= 道具 0x02 毒針的必殺秒殺 proc,是繞過此 bug 的唯一正規手段。)

**反組譯定位**:`sub_aa69`(file 0xbdd9,屬 `sub_a973`,由戰鬥指令迴圈 sub_c08b 呼叫),
file 0xbe02 起的「敵方累積值 clamp」:

```
saa7f: mov bx, [0x24a5]            ; bx = 選定敵人結構基底
saa83: cmp byte [bx+0x2334], 0      ; 累積值旗標==0 → 跳過
saa88: je   ret
saa8a: mov ax, [bx+0x2341]          ; ax = 增量欄
saa8e: add ax, [bx+0x2334]          ; ax += 累積值 [+0x2334]
saa92: cmp [bx+0x2336], ax          ; ★ 和 [+0x2336] 比 — 參照欄錯
saa96: jae  ok
saa98: mov ax, [bx+0x2336]          ; ★ clamp 取 [+0x2336]
saa9c: mov [bx+0x2334], ax
```

**根因**:這段做的是敵方某「累積值」的上限夾制,夾制基準引用了**錯誤的欄位 +0x2336**。
正確應參照 +0x2334(累積值本身)。引用錯欄使巴拉摩斯這隻 boss 的勝負/狀態累積判定走偏,
導致「全隊被吹飛(實際我方已無人)卻判我方贏」的錯誤結算。

**修正(已做成 binary patch)**:把兩處 `[bx+0x2336]` 改成 `[bx+0x2334]`,即兩個 disp16
的高位元組 byte `36`→`34`:

| file offset | 原 byte | 新 byte | 指令 |
|---|---|---|---|
| 0xbe04 | `36` | `34` | `cmp word[bx+0x2336]` → `[bx+0x2334]` |
| 0xbe0a | `36` | `34` | `mov ax,word[bx+0x2336]` → `[bx+0x2334]` |

這正是青衫攻略「修改:DQ3.EXE / 2.魔王打不死修正法」列出的官方碼
(`39 87 36 23 73 04 8B 87 36 23` → 把兩個 `36` 改 `34`),反組譯逐 byte 命中,互相印證。

**驗證**:DOSBox 載 `work/DQ3_fixed.EXE`,進巴拉摩斯戰正常把牠 HP 打到 0 應正常判勝、推進劇情。

---

## Bug 2:彩虹水滴誤拿黃寶珠(卡關)— 信心:高

**現象**(青衫):地下世界要先取得「彩虹水滴」才能架彩虹橋到魔王城,但理論上該拿到彩虹水滴的
時機卻會拿到「黃寶珠」,不修改就一定卡關。

**反組譯定位**:合成事件 handler,file 0x77ce–0x77f1(seg0 0x645e)。對應拉達多姆神聖祠堂
「雨和太陽合而為一就出現彩虹橋」的劇情:

```
s6461: mov [0x2593], 0x72          ; 太陽之石
s6467: call 0x7c0c                 ; 在道具欄找該物 → si 指向那一格
s646a: mov word [si], 0xff          ; 消耗太陽之石(0xff=空)
s646e: mov [0x2593], 0x73          ; 雲雨之杖
s6474: call 0x7c0c                 ; 找雲雨之杖 → si
s6477: mov word [si], 0x6b          ; ★ 成品就地寫入該格 — 寫成 0x6b
s647b: mov bx, 0x139               ; 設「已合成」旗標 0x139
s647e: call 0x8264
```

**根因**:太陽之石(0x72)+ 雲雨之杖(0x73)的合成成品 item code **被寫死成 0x6b(銀寶珠)**,
正確應是 **0x75(彩虹水滴)**。`mov word[si],imm` 把玩家道具欄裡雲雨之杖那一格就地換成成品。
因為 item code 是以 `mov word[mem],imm16` 形式寫死(不是 `mov ax,imm`),所以全 EXE 內
item code 0x75(彩虹水滴)從不被任何指令寫入 —— 合成事件是它唯一該產出的點,卻寫錯了碼。

> 註:青衫憑記憶寫「拿到黃寶珠(0x6a)」,程式真值是 0x6b(銀寶珠),差一個碼、同屬寶珠系列
> (口語誤差)。無論顯示成黃或銀寶珠,bug 本質相同:成品應是彩虹水滴 0x75。

**修正(已做成 binary patch)**:

| file offset | 原 byte | 新 byte | 說明 |
|---|---|---|---|
| 0x77e9 | `6b` | `75` | 合成成品 item code 0x6b(銀寶珠)→ 0x75(彩虹水滴) |

**驗證**:DOSBox 走到該合成事件,確認道具欄拿到「彩虹水滴」而非寶珠,可繼續架彩虹橋。

---

## Bug 3:九頭龍/五頭龍戰當機(卡關)— 信心:高

**現象**(青衫):進最終魔王城後有兩場必要戰鬥(`bugs.md` 稱九頭龍,攻略稱「五頭龍大王」)
會當機,不避掉就卡關。攻略給的避戰法是改存檔把那兩隻 boss 的 monster_id 改成 0x0A
(displacement 0x10/0x35 與 0x12/0x37 → 0x0A),另有 EXE patch「4.五頭龍大王當機修正」。

**根因 = 缺 sprite(未完成的怪物圖)**。這是「未發售/中文化未完成」最典型的事故:

- 兩隻 boss 的 monster_id:**0x80 = 歐里狄加**、**0x81 = 五頭龍大王**(怪名取自 D3TXT00
  記錄 `0x258 + id`,見 docs/16)。攻略明寫第一場是「和歐魯狄加戰鬥的那隻」,正是 0x80。
- 查 `DQ3MNS.SHP` offset table 這兩隻的 sprite 表頭:

  | id | 怪名 | sprite data off | w | h | 狀態 |
  |---|---|---|---|---|---|
  | 0x05 | 史萊姆(對照組) | 0x3e06 | 0x06 | 39 | 正常 |
  | 0x7c | 索瑪 | 0x136fa6 | 0x30 | 144 | 正常 |
  | 0x7d | 魔文 | 0x13df7e | 0x1a | 140 | 正常 |
  | 0x7e | 冰河魔人 | 0x141b58 | 0x12 | 126 | 正常 |
  | **0x80** | **歐里狄加** | 0x148516 | **0xfff8** | **65528** | **壞** |
  | **0x81** | **五頭龍大王** | 0x14b156 | **0x18ac** | **65528** | **壞** |

  0x80/0x81 的表頭前 4 byte 是未完成的填充值(`f8 ff f8 ff` / `ac 18 f8 ff`),不是真正的
  w/h。這正是 docs/16 怪物 agent 報告「130 隻中 128 隻可 render,最後兩隻 boss 大圖失敗」
  的那兩隻 —— 它們的 sprite 根本沒做完。

- 顯示鏈(re/battle.c):`mob_draw_group`(sub_b16f)→ `shp_load_sprite`(sub_b19e,
  file 0xc50e)把 sprite 整段讀進遠段 [0x2540] →色彩 blit `sub_b2af` 讀資料前 4 byte
  當 w/h,迴圈 `h rows × w bytes × 4 planes` 貼圖。當 w=0x18ac(6316)、h=0xfff8(65528)
  時,blit 迴圈量 = 6316 × 65528 × 4 ≈ 16 億 byte,遠超 21KB 的 sprite 緩衝 → 寫穿
  video / 讀穿緩衝 → **當機**(陣列越界 / 壞指標)。

**修正(已做成 binary patch)**:把 boss 遭遇腳本表(overlay 區 file 0x1b02c..)裡那兩隻
缺 sprite 的 boss 換成有效 sprite 的怪。原表是 3-byte 一項的腳本,boss 段 raw:

```
file 0x1b02c:  80 01 81 01 01 2d 08 7c 01 01 2a 08 81 01 ...
                ^^    ^^                          ^^
              0x80  0x81(第一場)               0x81(第二場)
```

| file offset | 原 byte | 新 byte | 說明 |
|---|---|---|---|
| 0x1b02c | `80` | `7e` | 歐里狄加 0x80(缺 sprite)→ 0x7e 冰河魔人(第一場) |
| 0x1b02e | `81` | `7d` | 五頭龍大王 0x81(缺 sprite)→ 0x7d 魔文(第一場) |
| 0x1b038 | `81` | `7c` | 五頭龍大王 0x81(缺 sprite)→ 0x7c 索瑪(第二場) |

替代怪 0x7c/0x7d/0x7e 的 SHP 表頭都正常(w<0x200,h<0x200),blit 不會越界。這與青衫攻略
「4.五頭龍大王當機修正」(`80→7E`、`81→7D`、`81→7C`)完全一致,等於官方驗證過的修法。

> 補充:更「忠於原作」的修法是替缺失的 0x80/0x81 sprite 補上有效圖、或在 blit 前加
> 「w/h 上限 guard(如 w>0x200 或 h>0x200 就跳過/用 placeholder)」。guard 版較通用但
> real-mode 原碼無空檔可塞,需 code cave / 改跳轉(非同長度),先標註待 SDL remake 實作;
> 本次採同長度的「換怪」patch,可立即驗證且與官方碼一致。

**驗證**:DOSBox 走到該兩場戰鬥,確認不再當機、能正常打完。

---

## Bug 4:勇者升級最大 MP 成長固定為 1(樂趣)— 信心:中

**現象**(青衫):勇者每次升級 MaxMP 只 +1,放不了幾個咒文 MP 就乾,連勇者專用全體補血咒文
「比荷瑪拉」都幾乎放不出來。

**反組譯定位**:升級數值成長公式 `sub_d9cc`(file 0xed3c),MP 區段 file 0xedb0 起:

```
mov al,[bx+0x4369]   ; MP 成長係數 slope(byte,來自 dragon0.dat 成長表)
mul dl               ; × 等級
shr ax,1             ; ÷2
add al,[bx+0x4368]   ; + MP base
adc ah,0
mov cx,[si+0x1e]     ; − 現值 MP
sub ax,cx            ; delta = 目標 − 現值
... delta>0 → 亂數微調(sub_fa57,回傳 ≥1)→ add [si+0x1e](MP)/[si+0x20](MaxMP) 同一 delta
```

**根因**:程式邏輯本身**正確** —— MaxMP(+0x20)與 MP(+0x1e)加的是同一個 delta,
公式是「目標值 = base + slope×level/2,寫回 = 目標 − 現值」,**沒有把 MaxMP 寫死成 1**。
所以「每級只 +1」不是程式碼硬編,而是**勇者那一列的 MP slope 在成長表(dragon0.dat)裡設得太小**:
slope 很小時,`slope×level/2` 每升一級的目標增量只比上一級多 0~1,再經 `sub_fa57` 亂數
(下限 1)夾出「每級 +1」。勇者是混合職業,設計上 MP 成長本就最低,精訊把這列值設得過低
(或 `shr 1` 把本就小的 slope 直接砍沒)。

**修正(C 層 / 資料層)**:真正的修法在 **dragon0.dat 成長表勇者列的 MP slope/base bytes**,
把勇者 MP slope 調高;或日後 SDL remake 重寫成長表時,拿掉對 MP 的 `/2` 過度衰減。

**受阻誠實標註**:成長表是 runtime 由 dragon0.dat(loader sub_155e)載入到 BSS,**不在
EXE 靜態映像內**(用 DGROUP 推算位址落在檔尾之外),所以勇者 MP slope 的確切值與在
dragon0.dat 內的 offset,本次靜態反組譯讀不到。要 100% 坐實「是表值偏低」,下一步需 dump
dragon0.dat、依 sub_155e 載入邏輯定位成長表 offset 再比對。傾向判定是表值問題,因為公式對
所有職業共用同一段碼,唯獨勇者表現異常 → 差異只能來自表。

---

## Bug 5:高等級(~Lv4x+)升級錯亂(樂趣)— 信心:高(根因)

**現象**(青衫):約 Lv4x 之後,每當經驗到升級標準,就會連升 30~40 級,還會學到一堆本來
那個職業學不到的咒文,造就 HP/MP 超高、會用所有咒文的怪物角色。

**反組譯定位**:升級外迴圈 `sub_d94c`(file 0xecbc)對每名角色反覆呼叫單級判定
`sub_ecdb`,**回傳 1 就再呼叫一次**(連升)。門檻查表 `sub_ecdb`(file 0xece9–0xecf5):

```
mov bl,[si+0x15]   ; bl = 等級(byte)
xor bh,bh
shl bx,1
shl bx,1           ; bx = 等級 × 4
mov dx,[bx+di+2]   ; 升級門檻 高 word
mov ax,[bx+di]     ; 升級門檻 低 word(每級 4 bytes = 32-bit 門檻)
```

**根因 = 等級當索引、越界讀表**。等級 `[si+0x15]` 是 byte,×4 當索引去查 32-bit 門檻表。
門檻表長度有限(一般到 Lv50/99)。等級一旦超過表長,`bx = level×4` **越界讀到鄰接資料**,
讀到的「門檻」變成垃圾小值 → 角色經驗(`[si+0x34]:[si+0x32]`)幾乎必然 ≥ 該垃圾門檻 →
回傳 1 → 外迴圈不停 `inc 等級` 連升,正是「一次升 30~40 級」。

「學到別職業咒文」同源:升級每跨一級就呼叫學咒文 `sub_db5f`(file 0xeecf),用等級
`[si+0x15]` 在咒文學習表 `[bx+0x36fb]` 線性掃描比對「第幾級學某咒」。等級值爆走(40+)後,
掃描指標越界滑進**相鄰職業的咒文學習表**,於是學到非本職咒文。`sub_db5f` 還用 `[si+1]`
(職業)分流不同職業表,索引一錯整個對應就亂。

> 這也解釋青衫的觀察:正因為這個越界 bug 讓勇者亂學咒文,才有機會放出他幾乎放不出來的
> 專用咒文「比荷瑪拉」(與 Bug 4 互動)。

**修正(C 層)**:取門檻/咒文前 `if (level >= MAX_LEVEL) level = MAX_LEVEL;`,或門檻表/
咒文表加哨兵終止值;升級迴圈加「等級已達上限就 break」。binary 同長度 patch 不易(real-mode
原碼無空檔塞 clamp,需 trampoline / code cave),先標 C 層修。

---

## Bug 6:數值 255 溢位(樂趣)— 信心:中

**現象**(青衫):除防禦力外,屬性只要超過 255 就從 0 重新算;印象中防禦力上限是 1023。

**反組譯定位 / 根因**:在升級成長公式 `sub_d9cc`(file 0xed3c)內,目標值的計算用
`mul dl`(slope × level → 限制在 byte×byte)+ `add al,base; adc ah,0` 這種「8-bit 加 +
進位」慣用法 —— 這是 C 編譯器把成長中間值算成 `unsigned char`(8-bit)的特徵。屬性升到 255
後再算下一級目標,`(u8)(base + slope×level/2)` **wrap 回 0**,`sub ax,現值` 得到大負數 →
`jle` 跳過 → 該級不再成長;若 wrap 後目標 < 現值,重算時就「從 0 重新算」,正是青衫描述的
「>255 歸 0」。攻擊/敏捷這類純由公式算的屬性最容易爆;HP/MP 因以 word 累加且通常先到別的
上限,較少觸發。(已搜尋確認屬性寫回是 `add word ptr [si+..]` word 加法、無 `mov [si+..],al`
的 byte 截斷寫回,故截斷發生在中間值型別而非寫回欄位。)

**防禦力 1023**:全檔唯一的 `and ax,0x3ff`(25 ff 03)在 file 0x9321,但上下文是道具/咒文
清單繪製(取 10-bit ID),**不是防禦力 clamp**;`cmp ax,0x3ff` 全檔無。所以「防禦力上限 1023」
比較可能是**防禦力欄位/顯示本身是 10-bit**(裝備防禦加總後存進 10-bit 欄位或顯示遮罩),
而非升級公式裡的明確 clamp 指令。

> 攻略旁證:女鬼面具「防禦力 255」是裝備加成而非成長,符合「防禦走不同(較寬)的數值路徑」。

**修正(C 層)**:成長公式中間值與屬性欄位一律用 `uint16_t`(或 `int`),
`target = base + slope*level/2` 全程 16-bit;寫回前做明確上限 `if(v>9999)v=9999;`,
不依賴自然 wrap。防禦力若確為 10-bit,改成 u16 + 明確 `min(v,1023)`(或解除上限)。
源於 C 型別,組語層無單點可同長度 patch。

**受阻誠實標註**:防禦力「上限 1023」是否真為 10-bit 欄位,需對 DOSBox 實機把防禦堆到
>1023 觀察行為佐證;目前判定為「欄位/顯示寬度問題,非公式 clamp」。

---

## Bug 7:物品功能 bug(樂趣)

精訊版把幾個道具的特殊效果做漏了。逐一分析(item 結構見下方附錄)。

### 7a:飛鷹劍(隼劍,item 0x6e)只打一次 — 信心:高(根因)

**現象**:FC DQ3 隼劍攻擊兩次,精訊版只打一次。

**根因 = double-hit 功能整段未實作**。裝備特效解碼 `battle_setup_party`(sub_c8c6,
file 0xdc36)掃 8 格裝備時,用的是**寫死兩個 item code 的白名單**(只比 0x4d、0x11),
完全不讀 item 記錄 +5 旗標。攻擊執行(玩家攻擊 handler file 0xbf75)只 `call` 一次傷害
計算,沒有攻擊次數變數 / `mov cx,2` / 外圈迴圈。全 EXE 搜不到任何 item 0x6e 的比較,
也搜不到攻擊路徑上對 0xc0 bit 的測試 —— 飛鷹劍 +5=0xc0(double-hit 旗標)躺在 ITEM.DAT
沒人讀。

> 旁證:D3TXT 怪/我方戰鬥訊息「{變數}的攻擊!」「受到{變數}點傷害」**成對重複出現**,
> 暗示原設計本有「連兩擊各印一次」意圖,但攻擊碼從未用到第二份。

**修正(C 層)**:攻擊執行依武器 item +5 bit7(0xc0 含 0x80)決定打兩次:

```c
uint8_t wpn = actor->equip[0].code;            // 角色結構 +0x3a 第一格 = 武器
int hits = (ITEM[wpn].flag5 & 0x80) ? 2 : 1;   // 飛鷹劍 +5=0xc0
for (int i = 0; i < hits; i++) { print_attack_msg(); apply_phys_damage(); }
```

原檔無 in-place 空間(功能不存在),需 code cave。0xc0 的精確 bit 語意建議 DOSBox 對 FC 校正。

### 7b:魔法鎧甲(item 0x2b)無抗魔法 — 信心:高(根因類型)

**現象**:FC DQ3 魔法鎧甲有抗魔法減傷,精訊版無效。

**根因 = 受咒文傷害減免功能整段未實作**。三條共用扣血器(物理 file 0xc098、通用
file 0xba71、file 0xc22e)套用傷害時**只看 [0xd75] 敵我旗標**,沒有任何「讀受方角色裝備
抗魔 → 打折傷害」的分支(唯一的屬性處理 [0x24b7] 是攻方咒文屬性,不是受方抗性)。
`battle_setup_party` 掃裝備只解碼 0x4d/0x11 兩碼,不讀 item +5/+6。全 EXE 搜 `cmp ...,0x2b`
(魔法鎧甲)/`0x2a`(天使袍)零命中。

魔法鎧甲 ITEM.DAT:+5=0x00、+6=0xd4;天使袍 +5=0xa0、+6=0x1c。兩者 +6 的 bit2(0x04)都
置位,若 FC 抗魔對應 +6 bit2,則資料其實正確,**純粹是程式從不讀取**(連天使袍抗魔在精訊版
同樣失效)。所以這是「碼根本沒檢查裝備抗性」,不只是資料缺旗標。

**修正(C 層)**:咒文傷害套用前掃裝備、命中抗魔 bit 就減傷:

```c
int magic_resist(Actor *t){
  for (int i=0;i<8;i++){ uint8_t c=t->equip[i].code;
    if (c!=0xff && (ITEM[c].flag6 & 0x04)) return 1; }   // +6 bit2 抗魔(語意待 FC 校正)
  return 0;
}
if (is_spell_damage && magic_resist(target)) dmg = dmg * 2 / 3;   // 或 >>1
```

只改資料(ITEM.DAT 0x2b 加旗標)無效,必須補程式。+6 哪個 bit = 抗魔,建議 DOSBox 對 FC 校正。

### 7c:祈禱之戒(item 0x48)永遠不會用壞 — 信心:高(與前提衝突)

**現象**(青衫):FC 原版祈禱戒指使用後有機率損壞消失,精訊版「永遠不會用壞」。

**反組譯定位**:祈禱戒指 use handler,file 0x5af4–0x5b0b(seg0 0x4784):

```
mov di, 0x187      ; 印 [391]「對著戒指祈禱。MP 恢復了。」
... (MP 恢復) ...
call 0xfa29        ; RNG:ax = rol16(seed+0x9018, 3)
cmp al, 0x40       ; 比 RNG 低位元組
ja  ret            ; al > 0x40 → 不壞
mov di, 0x188      ; 印 [392]「啊,戒指壞了。」
lcall 0x111b,0x264
mov ax, 0x48       ; 祈禱戒指 code
call 0x5b0c        ; remove_item:掃 +0x3a 8 格,找到該 code → [si]=0xff 移除
ret
```

**結論:破壞邏輯確實存在且會觸發**。把 RNG(`sub_e6b9`:`seed=rol16(seed+0x9018,3)`)做
10 萬次模擬,`al ≤ 0x40` 機率約 **25.5%** —— 每次用祈禱戒指約 1/4 機率會壞並從道具欄移除。
這條路徑單一、必經(MP 恢復後 fall-through 進入,無提前 ret),全 EXE 也只有這一處
`remove_item(0x48)` + 「戒指壞了」訊息。

就反組譯而言,祈禱之戒的耐久/損壞邏輯**正常運作(~25% 破壞率,與 FC「會壞」一致)**,
與青衫「永遠不會用壞」的描述**衝突**。`bugs.md` item 4 是把幾個道具籠統歸成「等等之類的」,
屬玩家憑記憶的概略陳述;7c 最可能是前提誤記,或破壞率被主觀低估。

**處置**:7c 不需 patch(現況碼正確)。唯一動態驗證點 = DOSBox 反覆使用祈禱戒指統計實際
破壞率;若實機確認真的不壞,再回頭查 use handler 是否另有未被此路徑覆蓋的選單入口。

---

## 修正版 patch 清單(已套入 tools/re_bugpatch.py)

執行 `python tools/re_bugpatch.py` → 讀 `assets_raw/DQ3.EXE`(唯讀)→ 套用下列 6 個同長度
in-place patch → 輸出 `work/DQ3_fixed.EXE`(work/ 已 gitignore)。套用前逐點驗證原 bytes
與預期相符,不符即中止(防錯位)。全部不更動 MZ header / segment 佈局 / relocation table。

| bug | file offset | 原 | 新 | 改法 | 信心 |
|---|---|---|---|---|---|
| 1 巴拉摩斯打不死 | 0xbe04 | 36 | 34 | cmp word[bx+0x2336]→[bx+0x2334] | 高 |
| 1 巴拉摩斯打不死 | 0xbe0a | 36 | 34 | mov ax,word[bx+0x2336]→[bx+0x2334] | 高 |
| 2 彩虹水滴 | 0x77e9 | 6b | 75 | 合成成品 0x6b(銀寶珠)→0x75(彩虹水滴) | 高 |
| 3 五頭龍當機 | 0x1b02c | 80 | 7e | 歐里狄加(缺sprite)→冰河魔人,第一場 | 高 |
| 3 五頭龍當機 | 0x1b02e | 81 | 7d | 五頭龍大王(缺sprite)→魔文,第一場 | 高 |
| 3 五頭龍當機 | 0x1b038 | 81 | 7c | 五頭龍大王(缺sprite)→索瑪,第二場 | 高 |

機讀版:`docs/data/bug_patches.json`(含未做成 patch 的 bug 4/5/6/7 的 C 層修法與理由)。

**驗證方式(DOSBox)**:
1. 開機:`work/DQ3_fixed.EXE` 應正常啟動進遊戲(6 byte 改動不影響開機路徑)。
2. Bug 1:巴拉摩斯戰把 HP 打到 0 應正常判勝。
3. Bug 2:走到彩虹橋合成事件,道具欄應拿到彩虹水滴(0x75)。
4. Bug 3:進那兩場 boss 戰不再當機(對手換成有效 sprite 的怪)。

**未做成 binary patch 的 bug**(需 C 層 / 資料層,留給 SDL remake):
- Bug 4 勇者 MP 成長:疑 dragon0.dat 成長表勇者列值偏低;需 dump 該檔比對。
- Bug 5 高等級升級錯亂:等級 byte 當索引越界;需 clamp(code cave)。
- Bug 6 數值 255 溢位:成長中間值 8-bit wrap;需改型別 + 明確 clamp。
- Bug 7a 隼劍 / 7b 魔法鎧甲:double-hit / 抗魔功能整段未實作;需補程式(code cave)。
- Bug 7c 祈禱之戒:破壞邏輯實際存在且 ~25% 會觸發,與前提衝突,不需 patch。

---

## 反組譯工具

- `tools/re_bug_dis.py` — bug 分析用反組譯/搜尋(docker capstone):`dis`/`seg`/`bytes`/
  `find <hexpat>`(?? 萬用)/`xref`。
- `tools/re_bugpatch.py` — 修正版 binary patch 產生器(讀原檔 → 套 patch → work/DQ3_fixed.EXE,
  套用前驗證原 bytes)。
- `docs/data/bug_patches.json` — patch 清單機讀版。

跑法(docker uv venv,不污染 host):
```
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone \
           && python tools/re_bug_dis.py dis 0xbe02 10"
python tools/re_bugpatch.py        # 產生 work/DQ3_fixed.EXE
```

## 殘留 / 待釐清

- **Bug 4 / 6 的資料面**:成長表(dragon0.dat)勇者 MP slope、防禦力欄位寬度,需 dump
  dragon0.dat + DOSBox 實機佐證才能 100% 坐實「表值 vs 公式」。
- **Bug 3 的「忠於原作」修法**:本次採同長度「換怪」patch(與官方碼一致);理想是補回
  0x80/0x81 的有效 sprite,或在 blit 加 w/h guard(需 code cave,留 SDL remake)。
- **Bug 7a/7b 的 bit 語意**:item +5 bit7(double-hit)、+6 bit2(抗魔)的精確語意,建議
  DOSBox 對 FC DQ3 行為實測校正後再定論。
- **Bug 7c**:與青衫前提衝突,需 DOSBox 實機統計祈禱戒指破壞率佐證。
