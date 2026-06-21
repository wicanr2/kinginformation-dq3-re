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

結論:全形漢字字模為 16×16(32 byte),但**逐字位置須由每段偏移/索引表決定**,
不能用固定步長硬切。精確逐字抽取留待:(a) 解開索引表語意,或 (b) 反組譯 `DQ3.EXE`
的字型載入/繪製程式取得定位邏輯。

### 已產出的地面真值

`tools/font_match.py` 產生 1,372 筆 `(china_byte_offset → d3txt00_glyph_index)` 對照,
為下一輪逐字抽取的錨點(僅為位置數據,非字模本身)。

## 工具

| 工具 | 用途 |
|---|---|
| `tools/dockrun.sh` | 在 docker uv venv 跑 Python(Pillow),不污染 host |
| `tools/font_extract.py` | 抽 D3TXT00 / SETTXT / CHINA 三檔,產生 atlas PNG 到 `assets_out/font/` |
| `tools/font_*.py` | 偵察過程工具(stride 掃描、col-major 測試、起點偵測等) |

atlas PNG 為版權衍生物,不入版控;自備 `assets_raw/` 後執行 `tools/dockrun.sh tools/font_extract.py` 重新產生。

## 待辦

- [ ] CHINA.FON 每段索引表精確長度與語意(字碼對照?)→ 解開後可建「內部字碼 ↔ 字模 ↔ Big5/Unicode」對照表
- [ ] `D3TXT00.CF1/.CF2`、`SETTXT.CF1/.CF2` 解析
- [ ] 與 `D3TXT*.TXT` 文字腳本結合,把遊戲文字解成可讀中文
