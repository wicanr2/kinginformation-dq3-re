# 82 — 羅馬利亞金皇冠、臨時王位與辭位流程

> 2026-07-29；EXE 位址皆為 file offset。現行 Go／Ebitengine 實作使用通用
> `temporary_role` primitive，版本專屬值由 `dq3_cht` game pack 提供。

## 原版入口與任務

CTY02 section 1 白天 NPC raw record `{7,2,4,16,9,43,0}` 是 handler 9。DQ3.EXE
file `0x65ee..0x6646` 的首次任務分支：

1. 測試 story flag `0x2c`。
2. flag 為 1 時令 selected item=`0x33`，再搜尋全隊背包。
3. 尚未持有金皇冠時依序播放 D3TXT02 rec45、rec15。
4. 這條分支不修改旗標與道具，再談一次會重播任務。

持皇冠後由 file `0x6649` 起：

1. `mov word [si],0x00ff` 從搜尋命中的角色 item slot 移除一件金皇冠，再清 flag `0x2c`。
2. 播放 rec49、rec50；第一次選 No 播 rec51，再回到 rec50，不能直接退出。
3. 選 Yes 播 rec48，清一般勇者 flag `0x2b`、設臨時王位 flag `0x27`，重建場景。
4. 日後重談走 rec46→rec47；No 播 rec52 並離開，Yes 可再次接受王位。

角色圖 writer 位於 file `0x668b..0x66aa`：女性以 `AX=0x1d`、男性以 `AX=0x10`，
`BP=1` 呼叫角色圖 loader。換成 `DQ3MAN.BLS` entry base 後分別是 116、64。runtime
角色圖只由 canonical flag `0x27` 派生，存檔不另存第二份角色狀態。

## 地下競技場辭位

CTY02 section 3 的 raw record `{4,11,31,16,13,39,0}` 是玩家處於臨時王位時顯示的前國王。
它的 **CTY raw handler 是 13**；先前把 jump-table slot 編號誤寫成 handler 11，已更正。
真正 consumer 為 file `0x674b..0x67e6`：

1. rec68 後第一層 Yes 播 rec69，繼續當王。
2. 第一層 No 播 rec70，進第二層確認。
3. 第二層 No 播 rec71→rec69，仍繼續當王。
4. 第二層 Yes 播 rec72，之後清 flag `0x27`、設 flag `0x2b`，恢復一般隊伍角色圖。

本機完整實況 12:42–13:14 可見玩家在地下競技場左上與前國王交談並完成兩層選擇，
與 CTY、EXE、D3TXT 三份證據一致。

## 夜間王宮 gate

正式長流程取得金皇冠後回到羅馬利亞時可能已是夜晚。這不是碰撞錯誤：

- CTY02 section 0 白天表把守衛放在 `(13,6)`、`(16,6)`，中央兩格可通。
- 夜間表把守衛放在 `(14,6)`、`(15,6)`，兩名 NPC 的 `ctrl=0`，王宮中央通道封閉。
- 旅店成功交易依 DQ3.EXE file `0x876a..0x8778` 將日夜 selector 寫回白天並重設時刻；
  詳見 [`docs/83`](83-inn-and-romaly-route-audit.md)。

因此正式 trace 夜間抵達時先由正常設施入口住宿，再走白天王宮路線；不可讓 NPC 穿透、
忽略碰撞或捏造地下側門。

## Game-pack 與驗收

`events.json` 的 `dq3:event.romaly_crown_kingship` 保存：

- offer／restore NPC selector；
- item `0x33`、flags `0x2c/0x2b/0x27`；
- 男女角色圖 entry base；
- rec15、45–52、68–72 對應的穩定 text ID。

Go 只實作跨版本的具名狀態機、交易與角色圖派生。驗收證據：

- `TestDQ3RomalyTemporaryRoleMatchesOriginalEXECTYAndText` 逐欄對 EXE、CTY 日夜表及
  D3TXT glyph words。
- `TestTemporaryRoleQuestAndForcedItemReturn` 與
  `TestTemporaryRoleRestoreTwoStageChoiceAndSaveDerivation` 鎖定所有 Yes／No 分支、
  交易時序與角色圖派生。
- `TestOpeningProductionInputTrace` 從新遊戲正式取得皇冠；夜間住宿、交冠、先拒絕再接受、
  從標題讀檔、走到競技場辭位、再存讀檔並正常離城，全程只用 `InputState`。
- Runtime 圖：
  [`romaly_king_crown_quest.png`](../dq3_remake_ebitan/docs/romaly_king_crown_quest.png)、
  [`romaly_crown_return.png`](../dq3_remake_ebitan/docs/romaly_crown_return.png)、
  [`romaly_kingship_choice.png`](../dq3_remake_ebitan/docs/romaly_kingship_choice.png)、
  [`romaly_temporary_king.png`](../dq3_remake_ebitan/docs/romaly_temporary_king.png)、
  [`romaly_restore_intro.png`](../dq3_remake_ebitan/docs/romaly_restore_intro.png)、
  [`romaly_restore_confirm.png`](../dq3_remake_ebitan/docs/romaly_restore_confirm.png)。

此切片已達 E3；runtime 畫面已有 V2 證據，但逐幀動畫、音訊 cue 與同幀像素級 V3 仍屬
全案視聽長尾，不能由事件流程通過推成整個 remake 完成。
