# DGROUP 靜態資料表目錄(DQ3.EXE)

我們的反組譯先前只 RE 了**碼**(`re/*.c`),資料表只記位址。本目錄把 DGROUP 的
**靜態資料表一次抽出編目**(offset/size/語意/值),由 `tools/extract_dgroup_tables.py`
生成內建 C(`dq3_remake/src/dq3_exedata.c`),**remake 不再執行期讀 DQ3.EXE**。

> 位址慣例:`file = 0x16140 + DS_off`(HDR 0x1370 + DGROUP seg 0x14dd×16)。

| 表 | DS off | file off | size | 語意 | 用途 |
|---|---|---|---|---|---|
| 成長表 | 0x4366 | 0x1a4a6 | 8×14 | 每職業 HP/MP/… base+slope | dq3_stats(#4 MP / 成長)|
| 升級門檻指標 | 0x43d6 | 0x1a516 | 8×u16 | → 各職業 44×u32 累積經驗 | dq3_stats(#5 clamp)|
| 地形表 | 0x4df6 | 0x1af36 | 256 | tile→地形類別(0/1/3…) | 戰鬥背景選頁(packbg)|
| 物件事件表 | 0x37c4 | 0x19904 | 256×3 | tile→事件碼(0xff=無)| 確認/對話/觸發(sub_9cd6)|

## 成長表(0x4366,8 職業 × 14 byte)
列格式:`HPb HPs MPb MPs b4b b4s b6b b6s b8b b8s bAb bAs c d`(docs/23)
- 0 勇者: `06 06 03 05 03 05 08 10 08 02 06 06 06 06`
- 1 戰士: `08 06 04 06 03 03 08 14 00 00 04 04 04 04`
- 2 武鬥家: `0a 08 04 08 03 09 08 16 00 00 06 04 06 06`
- 3 僧侶: `04 06 02 05 03 04 07 0e 05 0e 05 08 08 08`
- 4 魔法使者: `04 04 01 05 03 06 06 0c 05 10 06 0a 08 08`
- 5 賢者: `04 06 02 05 03 04 07 0e 05 0c 08 08 08 08`
- 6 商人: `04 06 02 05 03 05 08 10 00 00 06 04 06 06`
- 7 遊玩者: `04 06 02 05 03 05 07 0e 00 00 06 06 06 0a`

## 升級門檻(0x43d6 指標 → 各 44 entry)
指標(DS off):0x43e6, 0x4496, 0x4546, 0x45f6, 0x46a6, 0x4756, 0x4806, 0x48b6;每職業 44 entry(lv0..43)。
- 勇者 lv1/10/43 門檻 = 29 / 4364 / 960002(範例)

## 地形表(0x4df6,tile→地形)
前 48 tile:`3 3 3 3 0 0 0 0 0 0 0 0 0 0 0 0 0 0 3 3 3 3 0 0 0 0 3 3 3 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0`(地形類別,戰鬥背景 page=[0xd73]+terrain*8,docs/13)

## 物件事件表(0x37c4,tile→事件 3 byte)
每 entry 3 byte;`sub_9cd6` 以前方 tile 查 `[tile*3+0x37c4]`,事件碼決定對話/觸發。
非空範例:tile0=0a0d06, tile1=500d0c, tile2=b40d04, tile3=141d06, tile4=231d0c, tile5=641d05

## 再生方式
```
tools/dockrun.sh tools/extract_dgroup_tables.py
```
→ 重生 `dq3_remake/{include,src}/dq3_exedata.*` 與本目錄。版權:數值平衡表依 decomp
慣例納入可重編譯來源;大型創作素材(sprite/劇本/音樂)仍為外部 asset。
