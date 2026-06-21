# 字型格式:D3TXT00.FON / SETTXT.FON / CHINA.FON

精訊版 DQ3 的中文點陣字型已破解。三個 `.FON` 共用同一種**字模(glyph)位元排列**,差別在容器結構。

## 字模位元排列(三檔共用,已驗證)

| 屬性 | 值 |
|---|---|
| 字模尺寸 | **16 × 16 像素** |
| 排列 | **row-major**(逐列),每列由上而下 |
| 每列 bytes | 2(16 bit) |
| bit 順序 | **MSB-first**(bit 15 = 最左像素) |
| 每字 bytes | 32 |

解碼一個字模(`blob` 為位元組陣列,`base` 為字模起始):

```python
for y in range(16):
    row = (blob[base + y*2] << 8) | blob[base + y*2 + 1]
    for x in range(16):
        pixel_on = (row >> (15 - x)) & 1
```

驗證方式:以 `D3TXT00.FON` 為羅塞塔石碑,按此排列 render 出完全正確的 `0-9 A-Z`、注音符號與遊戲中文字,故排列確認無誤。

## D3TXT00.FON — 遊戲內字型(完全解出)

- 47,232 bytes = **1,476 個 16×16 字模**,無 header,整檔即連續字模。
- 內容 = 遊戲實際使用的全部字元,依序大致為:
  - ASCII:`0-9`、`A-Z`、標點符號
  - **注音符號**(ㄅㄆㄇㄈ…)
  - **遊戲中文詞彙**:職業(勇/戰/武/僧/魔/賢/商/遊)、武器、防具、道具、咒文、地名、人物、系統訊息用字
  - 羅馬數字 `I II III IV`、★、箭頭、`OK` 等符號
- `SETTXT.FON`(6,368 bytes = 199 字)= SETUP 程式用的小字集,同排列。
- 對應的 `D3TXT00.CF1`(39 bytes)/ `D3TXT00.CF2`(5 bytes)疑為字碼對照 / 編碼設定,待解。

## CHINA.FON — 母字庫(變寬字型,逐字定位待解)

192,200 bytes,容器結構:

```
+0x00  22 個 32-bit LE offset (offset table,0x00..0x57)
       offs[0]=0x58 同時標示 table 結束
0x58.. 22 個 section,offs[i] .. offs[i+1]
```

每段開頭是一張遞減的 16-bit 值索引表(sec14:558, 589, 588, 587, …, 551 連續遞減),
推測為字碼索引;其後是字模資料。

### 關鍵發現:這是變寬(proportional)字型,非固定 16×16 格

以**已驗證正確的 `D3TXT00.FON` 字模當範本字典**,在 CHINA.FON 逐 byte 搜尋完全相符的
32-byte 片段,得 **1,372 個精確命中** → 證實:

1. **`D3TXT00.FON` 是 `CHINA.FON` 的子集**(同一套字模設計),CHINA 為母字庫。
2. **字模不在固定 32-byte 格點上**:以每段最早命中為起點,`(offset − start) % 32 == 0`
   的比例僅 2–17%,且每段命中的 `offset mod 32` 有 13–31 種不同值。
3. 真值位置呈「數段 32-間距的連續全形字模 + 不規則間隙」(sec14 相鄰差:32, 32, 32, 64,
   737, 64, 32, 32, 64, 225, …)。間隙含非全形(半形 8×16=16 byte?)或變長資料,
   故整段沒有單一相位能對齊——這就是早期「整字被攔腰切半 / 字跨行」的根因。

### 字模實為 16×14(row14/row15 恆為行距留白)

以 D3TXT00 的 1,460 個有效字模統計各列全空比例:**row14 = 97.9% 全空、row15 = 99.2% 全空**,
其餘列平均有筆畫。即字身佔 row0..13(16×14),row14/15 為行距留白。此特徵是對齊的關鍵約束:
半格錯位的視窗在這兩列必有筆畫,可據以否決。

### 抽取解法(已解):run-based 相位對齊 + 16×14 留白約束

CHINA.FON 為「連續 32-byte 字模 + 段首索引表 + 字模間零星插入位元組」的變位佈局。
**多數字模以固定 32-byte 間距、同一相位連續排列(一條 run)**;字模之間偶有**1-byte 插入**
(sec14 某處 gap=737=23×32+**1**,該 1 byte 使後續字模起始 offset 的奇偶相位翻轉)。
因此整段沒有單一相位能對齊,固定步長必失敗。改用 run-based 對齊:

```
維持一條「目前相位」的 run:
  進入 run 起點:小窗內取 validity 局部極大者(避免落在真相位旁邊 ±1/±2)。
  run 內每個位置:在 {b-2..b+3} 重新對齊,取 validity 最高且合格者輸出, b 推進到該位置 +32。
    validity = 筆畫 4-連通性 − 孤立雜點懲罰 + 底部留白獎勵(row15==0、row14 近空)。
    底部留白是最強訊號:對齊乾淨的字 row15 全空、row14 近空;相位錯的字筆畫會溢進 row14/15。
  兩相位皆不合格(掉出 run)→ +1 滑動,直到重新找回 run 起點(吸收索引表/雜湊區)。
最後做一輪局部精修:每個輸出字在 ±2 內 snap 到 validity 局部極大(不越過鄰字),
修正 run 過程慢慢累積的相位漂移。
```

實測抽出約 **5,386 個完整中文字**;以 D3TXT00.FON 命中的 **1,372 個地面真值錨點**量測,
run-based 對齊的字落在錨點上的覆蓋率達 **≈93%**(歐、詠、愛、矮、藹、礙 等先前歪掉的字皆已正字)。
工具:`tools/font_align.py`(對齊核心)、`tools/font_greedy.py`(產 atlas)、
`tools/font_align_check.py`(對地面真值量測對齊正確率)、`tools/font_match.py`(產錨點)。

### 沒有「字模混淆」——是 1-byte 對齊漂移(更正先前判斷)

先前版本以為約 21% 的字被精訊「左右半對調 + 下移」**混淆**存放,並以「去混淆 transform」還原。
**這個判斷是錯的。** 程式驗證:所謂的「去混淆」運算

```python
hi = [blob[b+y*2]   for y in range(16)]
lo = [blob[b+y*2+1] for y in range(16)]
deobf_row[y] = ((lo[y-1] if y>=1 else 0) << 8) | hi[y]
```

在數學上**對所有 y≥1 完全等於從 offset b−1 讀取**(`rows16(b-1)[y]`),僅 row0 的高位元組被歸零有別。
也就是說:「去混淆」=**往前 1 byte 讀**,根本沒有對調或下移這回事。

真正的成因:字模之間的 **1-byte 插入**使部分字模落在**奇相位**的 offset。
舊版貪婪掃描固定 +32,一旦掉到錯相位就晚讀 1 byte,讀出來的字看起來就像「左右半對調 + 下移」
——那只是**對齊漂移**的視覺假象,不是字型層級的防拷混淆。**讀對位置(正確相位)就是正字。**

新的 run-based 對齊用 validity(尤其 row14/row15 底部留白乾淨度)鎖定正確相位,
連成 run、偵測 1-byte 插入翻相位、最後局部精修吸收累積漂移,從根本對齊,
不再需要任何「去混淆 / 還原」啟發式。

### 待清尾

- [ ] 殘留 ≈7% 對不齊的字:集中在段首索引表邊界、字模間夾雜的非全形(半形 8×16?)/變長資料,
      以及少數筆畫極密、validity 與鄰相位差距小的字;需更強的相位約束或解出段內結構才能補齊。
- [ ] 索引表語意(每段遞減 16-bit 值 = 內部字碼?)→ 建內部字碼 ↔ 字模 ↔ Big5/Unicode 對照
- [ ] 每段索引表精確長度與字模起點(目前靠 validity 滑動找回 run 起點,非由結構直接定位)

## 工具

| 工具 | 用途 |
|---|---|
| `tools/dockrun.sh` | 在 docker uv venv 跑 Python(Pillow),不污染 host |
| `tools/font_align.py` | CHINA.FON run-based 相位對齊核心(validity + run + 局部精修);被 greedy / check 共用 |
| `tools/font_greedy.py` | 用 font_align 抽 CHINA.FON 並產 atlas(全段 CHINA.greedy.atlas.png / 單段 greedy_secNN.png) |
| `tools/font_align_check.py` | 用 D3TXT00 錨點量測 run-based 對齊正確率(≈93% 覆蓋) |
| `tools/font_extract.py` | 抽 D3TXT00 / SETTXT / CHINA 三檔,產生 atlas PNG 到 `assets_out/font/` |
| `tools/font_*.py` | 偵察過程工具(stride 掃描、col-major 測試、起點偵測等) |

atlas PNG 為版權衍生物,不入版控;自備 `assets_raw/` 後執行 `tools/dockrun.sh tools/font_extract.py` 重新產生。

## 待辦

- [ ] CHINA.FON 每段索引表精確長度與語意(字碼對照?)→ 解開後可建「內部字碼 ↔ 字模 ↔ Big5/Unicode」對照表
- [ ] `D3TXT00.CF1/.CF2`、`SETTXT.CF1/.CF2` 解析
- [ ] 與 `D3TXT*.TXT` 文字腳本結合,把遊戲文字解成可讀中文
