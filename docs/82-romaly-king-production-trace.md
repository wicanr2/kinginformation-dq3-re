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

持皇冠後由 file `0x6649` 起進入還冠／王位分支。2026-07-29 重新反組譯閉合如下：

1. file `0x6623` 令 selected item=`0x33`，`0x7c0c` 搜尋隊伍背包；找到時
   `[0x726]=0`，跳入 `0x6649`。
2. `mov word [si],0x00ff` 從找到的角色 item slot 移除金皇冠，再呼叫 `0x8264`
   清 story flag `0x2c`。
3. 依序播放 D3TXT02 rec49「帶回金皇冠」與 rec50「代替我統治國家」。
4. Yes/No 選 No 時播 rec51「不要推辭」並回到 rec50；首次還冠不能直接離開此迴圈。
5. 選 Yes 播 rec48，清 flag `0x2b`、設 flag `0x27`，重建國王場景並進入玩家為王狀態。
6. 日後 `0x2c` 已清時重談會走 rec46→rec47；此時 No 播 rec52 並離開，Yes 再進
   rec48 的讓位結果。

flag primitive 已由 file `0x824f`（set）、`0x8264`（clear）、`0x8279`（test）逐指令
確認。此分支尚未接線，現行 handler 明確 fail closed：不自動收皇冠，也不替玩家接受王位。
它是金皇冠 E3 後的第一個 production blocker。

## Ebiten 驗收

- 新增 CTY02 handler9 正式 NPC consumer。
- `TestRomalyKingFirstQuestRecords` 鎖定 rec45→rec15 與無狀態交易。
- `TestOpeningProductionInputTrace` 從新遊戲正常抵達王座，以正式命令窗談話，完成兩段對白，
  再驗證 CTY02 section 1 save/load。
- Runtime 圖：
  [`romaly_king_crown_quest.png`](../dq3_remake_ebitan/docs/romaly_king_crown_quest.png)。

後續 campaign audit 已合法擊敗甘達特、取得金皇冠、從標題讀檔並返回羅馬利亞；
見 `docs/85`。因此本文件的持皇冠分支現在就是 current blocker。
