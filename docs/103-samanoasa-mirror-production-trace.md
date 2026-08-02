# 拉之鏡與沙曼歐莎假王正式流程

狀態：拉之鏡原始資料與 common consumer 已達 D3，既有假王事件 component 為 E2／V1；
boot trace 已自然重進 CTY60，但尚未穿過 CTY41→42 旅人之門抵達沙曼歐莎，故不是 E3。

## 原始輸入與定位

- `CTY24.DAT`：14,874 bytes；SHA-256
  `4c29d303d308b05dd5ee29a577e9d8878e9dc1d90a25f8b6c507a5841ba48a5d`。
- CTY24 section 2 base=file `0x2c77`，event table=file `0x2c90`；event0 位於 file
  `0x2c91`，raw `01 61 00 9f`，即 type1、item `0x61`、present flag `0x9f`。
- 通用調查事件 consumer 為 `DQ3.EXE` file `0x9d86..0x9e41`。type1 寫入道具成功後才清除
  present flag；因此新遊戲初始 set 的 `0x9f` 代表寶物仍存在，不是已取得。
- 假王事件入口與交易已由 `DQ3.EXE` file `0x5682..0x5732` 證實：夜晚、CTY44 sec1、
  玩家 `(14,7)` 使用 `0x61`，依序顯示 record97／98、進 monster89；勝利給 `0x62`、
  clear `0x21`、set `0x22` 並切回白天，敗／逃則復原假王旗標。

以上結論為 `confirmed`：CTY raw record、EXE writer／consumer 與正式玩家可見流程已閉合。
位址採 file offset；沒有把 IDA linear address 與檔案位址混用。

## Remake 修正

舊 `game/treasure.go` 的 `{24,2,0,1,97,159}` 只是一列 baked table，且 legacy reader 把
present flag set 解釋成「已取得」，會讓拉之鏡在正常初始旗標下消失。本批已：

1. 刪除該 Go 列，遷入 `dq3_cht` game-pack 的
   `dq3:event.shamanoasa_mod_change_mirror`；
2. 以原始 CTY／EXE parity test 鎖定 selector 與 consumer bytes；
3. 沿用共用 `collectQuestTreasure` transaction，缺 JSON 時失敗即關閉；
4. 將 content version 升為 `0.1.22`，schema 不變。

## 正式玩家輸入現況

`TestOpeningProductionInputTrace` 不使用 debug key、內部事件函式或狀態注入，目前只證實：

1. 從商人交付並 save/load 的 CTY58 checkpoint 正式離城；
2. 由同一世界座標重進，ordered gate 自然選到 CTY60；
3. 依原版攻略應再航行至雪島祠堂 CTY41，經三旅人之門到 CTY42，才可徒步至沙曼歐莎。

目前第一個 blocker 是 CTY42：正式最終鑰匙可開出口前的門，但現行 graph 所列世界出口
與 runtime 可達格／落點無法閉合，離場後也無法在不重入 CTY42 的情況下走向 CTY44。
曾測試以邊界外法線取代 facing 的推測性修正，仍未閉合，已撤回且未保留在 production。
下一輪必須由 CTY41／42 raw tile、transition subid 與 DQ3.EXE transition consumer 重新
追查；在此之前不得宣稱自然取得拉之鏡或清除商人城 `flag0x42` gate。

## 玩家可見證據

- [使用拉之鏡揭露假王](../dq3_remake_ebitan/docs/samanosa_mirror_reveal.png)
- [怪力魔戰](../dq3_remake_ebitan/docs/samanosa_boss_troll.png)

兩張圖由現行 Ebitengine renderer 重生，屬 V1 runtime 證據；尚未宣稱與精訊 DOS 同狀態
逐像素一致。
