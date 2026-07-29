# 82 — 羅馬利亞國王：金皇冠任務入口

> 2026-07-29；位址為 `DQ3.EXE` file offset。

## 先前錯誤記憶

舊研究日誌寫「皇冠 → 還羅馬利亞國王（已接）」，但現行 Ebiten `scriptedTable` 沒有
handler 8/9，玩家對 CTY02 section 1 國王談話不會出現任何內容。舊 C remake 的完成宣稱
不能當 Go/Ebiten production 證據。

## 原版證據

CTY02 section 1 白天 NPC `(7,2)` 是 scripted handler 9。DQ3.EXE
file `0x65ee..0x6646` 的首次任務分支：

1. GET story flag `0x2c`。
2. flag 為 1 時，令 selected item=`0x33` 並查隊伍背包。
3. 尚未持有金皇冠時，依序播放 D3TXT02 rec45 與 rec15。
4. 這條分支不修改 flag、不給／收道具；再次談話會重播任務。

rec45 歡迎繼承歐里狄加傳說的勇者；rec15 明說盜賊甘達特偷走金皇冠，請玩家找回。

持皇冠後由 file `0x6649` 起進入王位 Yes/No 分支，會清 `0x2c`、改 `0x2b/0x27`、
重建場景並涉及玩家選擇。該 optional 分支尚未接線，目前明確 fail-closed：不自動收皇冠，
也不替玩家接受王位。

## Ebiten 驗收

- 新增 CTY02 handler9 正式 NPC consumer。
- `TestRomalyKingFirstQuestRecords` 鎖定 rec45→rec15 與無狀態交易。
- `TestOpeningProductionInputTrace` 從新遊戲正常抵達王座，以正式命令窗談話，完成兩段對白，
  再驗證 CTY02 section 1 save/load。
- Runtime 圖：
  [`romaly_king_crown_quest.png`](../dq3_remake_ebitan/docs/romaly_king_crown_quest.png)。

下一個 campaign audit 從國王任務後正常出城，前往香巴尼塔 CTY10；不得用既有孤立 boss/
寶箱測試取代沿途可達性。
