# DQ3MAN.BLS 角色 sprite 格式(RE 偵察,空間解碼已解、色彩待校)

remake 地表/城鎮目前以佔位方框代表玩家;真主角 sprite 來源為 `DQ3MAN.BLS`。
本文記錄對該檔的反組譯偵察結果:**空間佈局已解、色彩對應待 DOSBox oracle 校正**。

## 檔案概況

| 檔 | 大小 | 備註 |
|---|---|---|
| `DQ3MAN.BLS` | 222726 | 角色(人物)圖庫;= 6B 表頭 + 58 × 0xf00 page |
| `DQ3MST.BLS` | 115206 | master(待查) |
| `DQ3LIN.BLS` | 46086 | line/陣列(待查);另有 `DQ3LIN.BRP` / `DQ3MAN.BRP` |

`222726 = 6 + 58 × 0xf00`,印證 `re/loaders.c::load_blk_page`(sub 0xff0e)的
分頁存取:`base = 0x3c0×bp + 6`、`pos = ax×0xf00 + base`,每次讀 `0x3c0`(960)B,
一個 page(0xf00)含 4 個 0x3c0 子塊。handle = `h_bls_man`(DS:0x2609),
目的段 `seg_blspage`(DS:0x2534)。

> 註:docs/08 舊將 `sub_ec9e`(開 dq3mst.bls/dq3man.bls)標為「BLS 音樂」,
> 為早期推定。實際 `DQ3MAN.BLS` 解出的是**角色圖形**(見下),該標籤應更正為角色圖庫。

## 表頭(6 bytes)

`04 00 18 00 e8 00` → `u16 w_bytes=4`(像素寬 32)、`u16 h=24`、`u16 count=0xe8=232`。
與 BLK tile 表頭同形(docs/04),但 body 不是 162×384 的平面排列(232×384≠body),
故 count 與實際排列關係待釐清(可能另有偏移表 / 變動尺寸 / 分頁邊界)。

## Sprite 像素格式(空間解碼已解)

body(檔首 +6 起)以 **384 byte / 32×24、4-bit plane-major** 切片即可 render 出
**可辨識的角色圖形**(頭盔、臉、髮、軀幹輪廓):

```
每 sprite = 4 plane 連續,plane 順序 plane3 → plane0(同 SHP,Map Mask ah 0x08→0x01);
每 plane = 24 row × 4 byte,每 byte = 8 像素的該 plane bit(MSB-first,bit0x80=最左);
色號 = Σ(plane_bit << plane),0..15。
```

工具:`tools/bls_probe.py`(多假設 sheet)、`tools/bls_probe2.py`(前 32 隻放大、色0=灰格)。
驗證:`work/bls_A_384_p3210_*.png`、`work/bls_chars_dq3.png` 呈現一排排角色頭像/人物
(形狀連貫、邊緣乾淨),確認 **w=4/h=24/plane-major plane3-first 的空間解碼正確**。

## 待解:色彩對應(palette)

空間正確,但用 `DQ3.PAL` / `MNSBK.PAL` 前 16 色直接上色,角色「背景/填色」呈彩虹雜色,
非乾淨透明背景。研判角色 sprite 用**專屬子調色盤**(palette 某段位移 / 每 sprite base),
由 EXE 的角色 blit 常式設定 —— 尚未對齊。

下一步(留後續回合):
1. 反組譯角色 blit 常式(找 `seg_blspage` 讀取點 + Map Mask + palette base 設定)。
2. 以 **DOSBox 原版 oracle** 截地表主角實際畫面,反推正確 palette 與「哪隻 = 地表主角下/左/上/右」。
3. 取代 remake scene 的佔位方框。

在此之前,remake 以佔位方框代表玩家(誠實標示未定稿),不貼上色不正確的 sprite。

## 後續發現(對照 DOSBox oracle,docs/29)

oracle 截到地表/城鎮主角為**小 sprite(~16px、透明底)**,金盔/髮 + 藍衣 + 膚色。據此再驗:

- **16 寬(w_bytes=2)重解 16×24/16×16**(`tools/bls_probe3.py`):糊掉,非乾淨單隻 → 不是單純 16 寬平面排列。
- **32×24 掃 plane 順序 × palette 視窗**(`tools/bls_probe4.py`,MNSBK 10 個 16 色窗 + DQ3 窗):
  角色頭像始終清楚,但**背景恆為填色(非透明),不隨視窗變透明**。
- 研判:**DQ3MAN.BLS 的 32×24 sprite 較可能是戰鬥/選單用的「大角色圖(帶底)」,
  不是地表那種 16px 透明走路 sprite**。地表小走路 sprite 來源另在他處,候選:
  `DQ3MST.BLS`(115KB)/ `DQ3LIN.BLS`(46KB)/ BLK tile 某段 / 分頁小 sprite 段。
- 下一步:反組譯地表場景的「玩家 sprite blit」呼叫(找讀取的段/檔與 sprite 編號),
  定位小走路 sprite 真正來源 + 子調色盤;以 oracle 畫面(`tools/oracle_crop_hero.py` 裁出)為基準校正。

## 格式定案(反組譯角色 blit sub_1ed0 + masked 解碼,已驗)

反組譯角色 sprite blit `sub_0x1ed0`(file 0x3240,wrapper `sub_0x1ea7`)解出**確切格式**:

```
out 0x3c4, 0x0802         ; VGA Sequencer Map Mask = plane(0x08→0x01 逐 plane)
cx = 0x18 (24 rows)       ; 高 = 24
es=[0x2532](VRAM 0xa000) ds=[0x2534](seg_blspage)
每 row 讀 2 word(4 byte=32px):dest = (dest AND mask[bx]) OR data[si]
bx = si + 0x180           ; ★ 遮罩在資料 +0x180(384)
```

→ **每角色 frame = 32×24,4-bit plane-major(plane3→0,384B 資料)+ AND 遮罩(+0x180,
透明 blit)**。先前「背景填色」就是**沒套遮罩**。masked 解碼 + **stride 960**
(= 222720 body / 232 entry)+ 遮罩 bit=1 透明 + **DQ3.PAL 低 16 色**,render 出
**完全乾淨的角色 sprite**(透明底):國王(紅披風冠)、老人、勇者(深髮藍衣 / 金髮藍衣,
各 **4 frame = 4 方向**)、白頭巾、士兵、騎士等。`tools/bls_probe5.py`(`bls_s960_dq3.png`)。

**結論**:DQ3MAN.BLS **就是**地表/城鎮角色(含主角)的 sprite 來源;32×24 masked、每角色
4 方向 frame、stride 960、DQ3.PAL。先前誤判「帶底大角色圖」是漏了 AND 遮罩 + stride 對不齊。
→ 可直接在 remake 解碼並以透明 blit 取代佔位框(見 `dq3_sprite` 模組)。
