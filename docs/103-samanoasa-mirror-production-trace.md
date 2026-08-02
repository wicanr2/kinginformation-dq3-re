# 拉之鏡與沙曼歐莎假王正式流程

狀態：拉之鏡原始資料與 common consumer 已達 D3；同一條 boot production trace 已從
CTY60 自然走過 CTY41→42、CTY43 關卡、CTY44、CTY24 拉之鏡、save/load 與假王戰，
此切片現為 E3／V1。

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

`TestOpeningProductionInputTrace` 不使用 debug key、內部事件函式或狀態注入，目前證實：

1. 從商人交付並 save/load 的 CTY58 checkpoint 正式離城；
2. 由同一世界座標重進，ordered gate 自然選到 CTY60；
3. 航行至雪島 CTY41，以正式最終鑰匙開門，經右側旅人之門到 CTY42；
4. CTY42 bottom row 的同一 transition subid 使用 raw destination `(213,123)` 離場；
5. 地表西行先進 CTY43，再走原始跨 CTY portal 到 CTY44，沒有穿越其他入口或直接載入城鎮；
6. 由 CTY44 關卡正常離場，徒步到 CTY24，取得拉之鏡並通過 save/load；
7. 使用黑暗燈、再次經 CTY43 入 CTY44，夜間走到 sec1 `(14,7)`，由正式道具選單使用
   拉之鏡，完成 record97／98、monster89、變身杖 `0x62` 與旗標交易。

IDA 9.4 證實 `DQ3.EXE` file `0x49f9..0x4a40` 重新讀取 transition X/Y；只有兩者皆零時
才回退 remembered coordinates，否則直接寫 world X/Y，完全不讀 facing。舊 remake
依 facing 把玩家額外推出兩格的行為與測試已刪除。原版入口同列兩格判定則由 file
`0x43a6..0x43d0` 證實；跨 CTY43 關卡是正式路線，不可把入口當成普通地表格穿過。

`worldEntranceGrace` 目前推論等級為 `strong`：精確出口座標、同列兩格入口 matcher、
諾魯德東口的地表連通性與完整 production trace 共同證明離場後首步不能立刻重進；但原版
保存／清除此一次性狀態的 writer-consumer 尚未在 IDA 閉合。因此不得把欄位名稱或
scene-mode 猜想升格成 confirmed；後續仍需用原版同操作畫面或直接狀態 writer 補到 D3。

## 玩家可見證據

- [使用拉之鏡揭露假王](../dq3_remake_ebitan/docs/samanosa_mirror_reveal.png)
- [怪力魔戰](../dq3_remake_ebitan/docs/samanosa_boss_troll.png)

兩張圖已由本批現行 Ebitengine renderer 重生（內容與舊檔相同），屬 V1 runtime 證據；尚未宣稱與精訊 DOS
同狀態逐像素一致。
