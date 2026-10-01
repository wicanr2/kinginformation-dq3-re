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

## 2026-10-01 同伴持有皇冠的 gate 勘誤

工作依據：[Issue #3](https://github.com/wicanr2/kinginformation-dq3-re/issues/3)。
**READY：限全隊持有權查找與一件還冠交易。** 原有主線讓主角持冠，未涵蓋同伴持有；
本輪正式路線在同伴取得皇冠並存讀檔後，國王仍播放找皇冠任務。Production gate 使用
只查主角的 `countItem`，與本文件原先記錄的全隊搜尋、實際全隊移除 consumer 不一致。
這是測試覆蓋缺口所掩蓋的引擎缺陷，不推翻原有 EXE 證據，也不改物品資料。

輸入：`assets_raw/DQ3.EXE`，115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
工具：IDA Pro 9.4、`ida-pro-9.4-idapython:locked-v1`；一次性 database，原版唯讀。
位址：`file = IDA linear − 0x10000 + 0x1370`；DGROUP 偏移保持原始運算元。

| 原始定位／bytes | 附加語意 | 推論等級與證據 |
|---|---|---|
| IDA `0x152b3`／file `0x6623`，`c70693253300`；IDA `0x152b9`／file `0x6629`，`e8e015` | handler9 把 selected item 寫成皇冠 `0x33`，呼叫原名 `sub_1689C` | confirmed：固定 EXE bytes／直接 call xref |
| `sub_1689C`，IDA `0x1689c..0x168df`／file `0x7c0c..0x7c4f` | `[DS:5077h]` 為迴圈次數；由 `[bx+4F15h]` 取角色指標、加 `0x3a`，逐角色掃八個 u16 槽；先以 `&0xff` 比對 `[DS:2593h]` | confirmed：callee 原始迴圈、取址與讀取；沒有只限首名或存活角色的分支 |
| IDA `0x168d9`／file `0x7c49`，`c606260700`；IDA `0x152bc`／file `0x662c`，`803e260700` | 命中留下 SI 指向槽位，寫 `[DS:726h]=0`；caller 以此選還冠分支 | confirmed：callee writer → caller consumer |
| IDA `0x152d9`／file `0x6649`，`c704ff00` | `[si]=0xff` 只清搜尋命中的一格，再清 pending flag、顯示 rec49 | confirmed：原始間接 writer／後續旗標與文字 call |

實作契約：pending flag 為真時查目前全隊持有權；無皇冠仍播任務且不消耗，命中則
依主角→同伴順序移除一件，再清 pending flag、播放 return_praise。其他選項與角色圖
沿用原規格。新 component 必須涵蓋同伴持有及兩件只消耗一件；正式 trace 必須保留
取得皇冠時的持有者與隊伍存讀檔，再由正常對話還冠與辭位。

可重現 IDA 匯出入口：`tools/ida_dump_temporary_role_item_gate.py`，輸出 sidecar 到
`/tmp/dq3-crown-gate-ida.json`；每筆指令自動附加原始定位、語意、等級與本節證據。
匯出未註記部分保持醒目的 UNKNOWN，不從名稱推導規則。原始 EXE 與 database 不入 Git。
本輪首份 sidecar 為空，沒有採用；明定 UTF-8 並增加工具日誌後於同一 image 重跑，
核對非空、輸入 hash、版本及原始 caller／callee，才採用證據。首次空檔的根因未獨立驗證。

原版正常玩家路線／畫面仍待 dosgolem 收據；本節靜態資料流不能宣稱完整主線同狀態 parity。

本輪 remake 驗證：新增同伴持有兩件的 component 在修正前失敗、修正後通過；
既有任務、強制選項、辭位與存檔角色圖派生亦通過。正式新遊戲 trace 已驗證同伴持冠、
隊伍物品／持有者存讀檔、正常還冠、王位／辭位，再繼續至日邦格；未重新設定亂數種子。
其餘 game、全部 internal 與 desktop build 通過；整段主線仍由 Issue #2 重驗。
局部引擎修正已通過以上驗證，原版 dosgolem 動態驗收保持 pending，不升格 V3。

本輪限定實作狀態為 **CONFORMED：全隊持有權與單件消耗**，不包含原版動態同狀態。
`TestTemporaryRoleCompanionCrownReturnConsumesOneItem` 可指定
`DQ3_DUMP_COMPANION_CROWN=<PNG 路徑>` 擷取已顯字的元件畫面；父目錄須先存在。
本輪 `work/dosgolem-opening/issue3-companion-crown-return.png` 為 640×350，18491 bytes，
SHA-256 `e00ca60fbecb99a80ce9738834ea3a9d4cbc67a221040ca332c41c0918ca3a2e`；
已目視確認還冠對白及視窗，這是元件畫面，沒有冒稱 dosgolem 同狀態收據。
舊 `TestDumpNewGameScreens` 在更早的魔法球測試設定停止，未取得還冠畫面；故改在已通過的
限定元件重生 PNG。首張只含逐字顯示的空白初始幀，依資料包等待長度推進文字後重取。

重生匯出（原版唯讀，database 僅存在容器 `/tmp`，sidecar 寫入主機 `/tmp`）：

```bash
test -d /home/anr2/dq3 && test -d /tmp && timeout 120s docker run --rm \
  --network none --memory 3g --cpus 2 --pids-limit 192 -u "$(id -u):$(id -g)" \
  --mount type=bind,src=/home/anr2/dq3,dst=/repo,readonly \
  --mount type=bind,src=/tmp,dst=/out ida-pro-9.4-idapython:locked-v1 \
  bash -c 'idat -A -c -o/tmp/dq3-crown-gate.i64 "-S/repo/tools/ida_dump_temporary_role_item_gate.py /out/dq3-crown-gate-reviewed.json" /repo/assets_raw/DQ3.EXE'
```

本輪 reviewed sidecar：248027 bytes，SHA-256
`564f7c375960d9c308f73bff9d914f0f05e8e6f91e40153d1099bd87e8fe3163`；
含 288 筆原始指令，其中 12 筆附上述限定語意，其餘保持 unknown。
