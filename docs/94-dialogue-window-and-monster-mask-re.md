# 對話視窗幾何與怪物 AND-mask 逆向閉合

> 2026-08-01；IDA Pro 9.4，原始 `DQ3.EXE`／`DQ3MNS.SHP`。本文只記錄可重現結論。

## 1. 共用對話視窗

原始結構位於 DGROUP `0x3e6e`（file `0x19fae`）：

```text
0b 01 13 00 ee 00 2c 00 60 00 ...
```

- `sub_15002`（logical `0x5002`，file `0x6372`）清 DGROUP `0x259b`，把 `si` 指向
  DGROUP `0x3e6e`，再呼叫 `sub_1f590`；多個 scripted-event caller 共用此入口。
- `sub_1f590`（logical `0xf590`，file `0x10900`）把 `[si+2]` 寫入文字 X、`[si+4]`
  寫入文字 Y，再消費視窗文字記錄。
- `sub_1fb36` 以 `[si+2]`／`[si+4]` 作左上角，將 `[si+6]` 加 1 VGA byte、
  `[si+8]` 加 `0x10` pixel，備份完整外框區。
- `sub_1fc57` 以 `[si+2]+1` byte、`[si+4]+8` 清內容，row count 取 `[si+8]`、
  width 取 `[si+6]`，閉合 `+6=寬、+8=高`。

換算：X=`0x13*8=152`、Y=`0xee=238`、內容寬=`0x2c*8=352`、內容高=`0x60=96`；
外框寬=`(0x2c+1)*8=360`、高=`0x60+0x10=112`，正好到 640×350 畫布底部。
文字 renderer 再加 `(16,16)`，形成 20 個 16px 欄與每頁 4 行。原版影片同狀態畫面亦顯示
中央偏下的窄框；舊 remake `(24,244,592,96)` 是未經原版資料證實的全寬近似值。

production canonical 值現存於 `data/interface.json`；Go renderer 不保留 DQ3 fallback。

## 2. 怪物透明遮罩

`sub_1b1fe` 先呼叫 `sub_1b31a` 套 AND-mask，再呼叫 `sub_1b2af` 寫四個色彩 plane：

1. sprite header 是 `{u16 width_bytes, u16 height}`。
2. 隨後是四段 `width_bytes*height` 的 plane-major 色彩資料。
3. `sub_1b31a` 跳到 `4 + 4*width_bytes*height`，逐列呼叫 `sub_1b37c` 解遮罩。
4. RLE top bits：`0x40` 重複 `0xff`、`0x80` 重複 `0x00`、`0xc0` 讀下一 byte
   重複、`0x00` 將目前 byte 當單一 literal。
5. 遮罩會 AND 到背景：bit 1 保留背景（透明），bit 0 清背景並接受後續色彩（不透明）。

所以色彩索引與透明度互相獨立；`MNSBK.PAL` index 0 是黑色，怪物可有 index 0 且不透明的
像素。舊 decoder 的 `Opaque = Px != 0` 會讓巴哈拉塔守衛／甘達特出現背景穿洞。

## 3. 驗收鎖

- game-pack parity test 鎖定 EXE file `0x19fae` 原始 bytes 與 JSON 外框換算。
- monster component test 鎖定守衛 id57 同時存在「黑色且不透明」與透明背景像素。
- runtime 驗收須重產 README 戰鬥圖及所有受對話框幾何影響的舊圖，再與原版同狀態目視比較。
