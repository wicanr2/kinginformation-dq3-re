# 158 — 原野道具持有者選擇與個人物品丟棄規格（2026-08-22）

## 問題與停止線

戰後掉落改回每人八格後，正式 trace 在羅馬利亞武防店遇到戰士個人物品欄已滿；現行
道具面板只列勇者 inventory，玩家無法以正式操作丟棄同伴藥草。原版則在打開道具命令時
先選持有者，再操作該角色的八格。本切片閉合持有者選擇、個人清單、rec421 動作選單、
同一持有者的丟棄 writer，以及使用／給予的 ownership 傳遞；不研究 item list 的逐像素框線。

## 輸入、工具與位址契約

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4＋IDAPython；image `ida-pro-9.4-ver3:py312-x11-v3`。
- exporter：`tools/ida_dump_field_item_owner_actions.py`。
- sidecar：`work/field-item-owner-actions-20260822.json`（不加入 Git）。
- IDA linear 位址；`logical=linear-0x10000`、`file=logical+0x1370`。原始函式名與
  raw function-boundary unknown 均保留；未以 rename 取代位址證據。

## 已證實的原版垂直鏈

### 1. 道具命令先選持有者

`0x1372F` 起的 raw runner（此 fresh IDA database 未建立函式邊界，因此只按 raw window
引用，不冒稱函式名稱）：

```text
0x1373A mov al,[5077h]          ; 現行隊伍人數
0x1373F cmp al,1
0x13743 mov word [0722h],1      ; 單人自動選第一人
0x1374C..5B                     ; 多人開 party selector sub_1885F
0x13768 mov bx,[0722h]          ; 1-based 選定角色
0x1376C mov [259Ch],bx          ; 文字插值 owner
0x13770 mov [062Dh],bx          ; item owner
0x13774 mov [0741h],bx          ; item-use target 初值
0x13778 call sub_138F8          ; 計算該角色八格非空數
```

若該角色沒有物品，`0x137F1` 顯示 record263「身上沒有攜帶物品」。有物品時
`0x137B8` 將 menu record 設成 `0x1A5`／421，`0x137BE` 設三列；D3TXT00 record421
即「使用／給予／丟掉」。特定職業另有第四列，不在本切片實作範圍。

### 2. 個人八格清單與選取

- `sub_138F8`：`BX` 減一、乘二，取 `DS:4F15[BX] + 0x3A`，掃八個 word，略過
  `0x00FF`，回傳非空數。
- `sub_13829/sub_13853`：再次把 `DS:0722` 保存成 owner `DS:062D`，把 item list 的
  1-based 游標保存成 `DS:062F`。
- `sub_13919`：以 `DS:062D` 定位角色 `+0x3A`，依 `DS:062F` 選第 N 個非空 word，
  將低八位 raw item 寫入 `DS:2591`。

因此 owner 與 item index 都是角色局部狀態；不能先把所有物品攤平成勇者清單再操作。

### 3. 給予與丟棄

給予分支 `0x139AB..0x13A9F` 以 `DS:062D` 為來源、一般 party selector 的
`DS:0722` 為目標。不同角色時掃目標八格 `0x00FF`，成功才把來源 word 寫成
`0x00FF` 並把 raw word 寫入目標；同一角色時 `0x13A62..0x13A9D` 只重排該角色清單。

丟棄分支 `0x13ABC..0x13B0F`：

```text
0x13ABC mov bx,[062Dh]          ; 選定持有者
0x13AC0 mov [259Ch],bx
0x13AC4 call sub_19834          ; DS:4F15[owner-1] 角色 pointer
0x13AD2 call sub_13919          ; 選定該角色的 item word
0x13AD5 test ax,0E000h          ; 特殊／裝備 bit gate
0x13AE7 mov al,[item+1FEh]
0x13AEB test al,2               ; 不可丟棄 metadata gate
0x13AF8 mov word ptr [si],00FFh ; 合法時只清來源 slot
0x13AFC mov di,0115h            ; record277「角色把物品丟掉」
0x13B02 call sub_18197          ; 清單／衍生狀態更新
```

死亡角色走 record278；不可丟棄物品走另一錯誤 record。現行 pack 尚未資料化完整
不可丟棄 metadata，因此本切片只要求既有允許丟棄的藥草等物品依 owner 正確移除；未知
metadata 不得由 Go 猜新規則。

推論等級：owner selector、`DS:062D/062F`、角色八格 reader、不同 owner 給予 writer、
同 owner reorder、合法丟棄 `0x00FF` writer 與 records263／277／278 均為 **confirmed**。
第四動作列及全部不可丟棄 item 身分為 **unknown／切片外**。

## Remake 規格

1. 多人隊伍開啟道具命令先進通用 party-member owner selector；單人可直接選主角。
2. 清單、游標、使用、給予、丟棄都操作 owner 的 personal inventory；不得默認勇者。
3. 取消階層為 action／item list → owner selector → 關閉，不能跳過 owner 層。
4. 給予目標包含全隊；不同 owner 受八格容量限制，同 owner 不複製道具。
5. 使用型道具必須從選定 owner 的確切 slot 消耗；效果目標仍依各 item effect 另行選擇。
6. 所有玩家可見標籤與訊息仍由 game-pack text ID；本切片不在 Go 加 DQ3 record 或中文。

## 驗收

- EXE parity：鎖定 `0x1373A..0x13778`、`0x13829..0x13941`、
  `0x139AB..0x13A9F`、`0x13ABC..0x13B0F` 的代表 bytes。
- component：選同伴→列其物品→丟藥草；選同伴物品使用時只消耗該 owner；給予的來源／
  目標容量與同 owner 不複製。
- production：羅馬利亞正式打開道具命令、選戰士、丟一株藥草，再由商店把青銅盾買給
  戰士並用正式裝備面板穿上。

## 2026-08-22 實作結果

- `panelActor` 現在同時保存原野道具持有者；多人由正式 owner selector 進入，單人自動
  選勇者。清單、使用、給予與丟棄均透過該角色的 personal inventory。
- 不同角色給予遵守含裝備在內的八格容量；同角色給予只把選定欄位移到尾端，不複製。
  消耗與丟棄都依選定 slot 移除，不再以 raw id 掃描勇者背包。
- `field_item_owner_test.go` 已鎖定同伴丟棄、同伴使用、跨角色給予、同角色重排與上述
  EXE 代表 bytes。Docker＋Xvfb targeted tests 通過。
- 正式 `TestOpeningProductionInputTrace` 已從標題越過羅馬利亞：戰士經 owner selector
  丟棄藥草、購得並裝上青銅盾，勇者亦購得並裝上龜殼甲胄；重播繼續到取船後航線才在
  monster51 遭遇全滅。因此本切片達 E3；整條 campaign 的現行 E3 仍為 pending，不能引用
  較早的 `THE END` checkpoint 冒稱目前工作樹已通過。

## 後續 trace 訂正

共用 grant 接回後，《領悟之書》可能由任一有空位角色持有。舊
`traceGiveInventoryItem` 寫死由勇者出發，並把「勇者不再持有」的斷言機械改成
`hasPartyItem` 後形成永遠失敗的條件。現行 helper 依隊伍順序尋找實際 owner，必要時以
正式丟棄藥草替目標角色騰格，再走 owner selector／rec421／全隊 target selector；物品
原本就在轉職目標時不做無意義重排。驗收改為「全隊仍持有，且指定同伴持有」。
