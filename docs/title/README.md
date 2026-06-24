# 標題 / 開場畫面

`FIRST.SCR` = 遊戲開場畫面,經反組譯解出格式後 render(`tools/title_render.py`):

- **640 × 350 像素、4bpp(16 色)**
- **row-interleaved planar**:每一掃描列 = plane0..3 各 80 byte(= 320 byte/列),350 列 × 320 = 112,000 byte
- 調色盤:**開場插圖不用 `DQ3.PAL`**。原版顯示開場前另設一組獨立的 16 色 VGA DAC,膚色在
  **index 14**(`DQ3.PAL[14]` 是灰色,直接套會把臉畫成灰的——舊版 bug)。正確 16 色由實機截圖
  (`dq3_org_pic/727339426…`,姓名輸入畫面同一張開場圖)反推:女主角區(右側無視窗遮擋)逐
  index 取所有像素的中位色;主色(膚/紅/藍/金/白)樣本量大、穩定,膚色 idx14 釘到 DQ 標準肉色
  `fcd89c`。硬編在 `tools/title_render.py` 的 `OPENING_PAL`。

`first_opening.png` = 還原結果:DQ3 雙主角開場原畫(鳥山明風格),膚色已修正為正確肉色。

## TIT*.P = 標準 ZSoft PCX

`TIT*.P`(TITA~TITP、TIT0~3 等)為**標準 ZSoft PCX 檔**(`tools/title_pcx.py` 解碼)。
原本以為的「自訂壓縮表頭」其實就是 PCX 的 128 byte 標準表頭——`0x0a` 是 PCX magic
(廠商碼 ZSoft),不是自訂壓縮旗標:

| offset | 欄位 | 值 | 說明 |
|---|---|---|---|
| 0 | manufacturer | `0x0a` | PCX magic(ZSoft) |
| 1 | version | `0x05` | v3.0+,header 內含 16 色 palette |
| 2 | encoding | `0x01` | RLE |
| 3 | bitsPerPixel | `0x01` | 每 plane 1 bit |
| 4–11 | Xmin,Ymin,Xmax,Ymax | 0,0,639,349 | 邊界框 → 640×350 |
| 12–15 | HDpi,VDpi | 640,350 | |
| 16–63 | EGA palette | 16 色×3 byte | **各檔自帶**,不用 DQ3.PAL |
| 65 | nplanes | `4` | 4 個 bit-plane → 4bpp/16 色 |
| 66–67 | bytesPerLine | `80` | 每 plane 每列 80 byte |

**RLE 解碼**:讀一 byte `b`;若 `(b & 0xC0) == 0xC0` → `count = b & 0x3F`、下一 byte 為值,輸出 `count` 份;
否則 `b` 為單一 literal。解到 `H × nplanes × bytesPerLine = 350 × 4 × 80 = 112,000` byte 為止。

**重組像素**:每列 320 byte = plane0..3 各 80 byte 依序排(plane-sequential per row,與 FIRST.SCR 同),
`pixel = bit(p0) | bit(p1)<<1 | bit(p2)<<2 | bit(p3)<<3`,套用 header 內 16 色 palette。

## 各檔內容(對 references/game1~3.png 驗證)

- **`title_logo.png`(= TITG.P)** = **標題畫面正解**,與 `references/game1.png` 像素吻合:
  DRAGON FIGHTER 3D logo + 寶珠權杖 + 藍色弧線 +「傳說的終章」+ III +「© 1993 精訊資訊有限公司」,
  橘色漸層天空、綠色山丘與城堡剪影。
- `title_bg.png`(= TITF.P)= 標題的**底圖層**(同樣天空/山丘,但無 logo 與文字;logo 與中文字為另一層,
  推測 runtime 疊上 TITF 上方)。
- `ending.png`(= TIT3.P)= 片尾「THE END / TO BE CONTINUE DRAGONFIGHTER I」+ DRAGON FIGHTER III logo。
- `scene_1989.png`(= TITA.P)、`scene_1993.png`(= TITD.P)= 年代過場(1989 / 1993 黑白人物群像)。
- 其餘 TITB/C/E/H~P、TIT0~2 = 角色群像、雪山湖泊、城堡等過場/章節插圖,palette 分兩組
  (TITA~E 灰階組、TITF~O 彩色組)。

`tools/title_pcx.py <IN.P> <OUT.png>` 可解任一張(docker uv venv 內跑)。
