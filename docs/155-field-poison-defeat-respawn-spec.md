# 155 — 原野中毒致死、全隊倒下與記錄點復歸規格（2026-08-22）

## 問題與停止線

`docs/153` 訂正怪物行動 RNG、`docs/154` 修正正式玩家補給策略後，
`TestOpeningProductionInputTrace` 已自然推進至 CTY55。當時三名現役角色皆為
`HP=0`、`conditionPoison`，但 `Game` 仍停在一般地圖狀態；之後無法開啟魯拉選單。
這不是事件路由或咒文資料缺失，而是 remake 只有戰鬥結束後的全滅近似，沒有承接
`applyHazardStep` 的原版死亡 consumer。

本切片只閉合：

```text
正式成功移動 → 每員地形／毒傷 → HP 歸零與死亡位 → 存活人數 →
部分死亡重排／全隊倒下訊息 → 最後記錄點復歸
```

不擴張研究其他 field status、教會價格、戰鬥動畫或逐幀淡出。

## 輸入、工具與位址契約

- 輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：Docker image `ida-pro-9.4-ver3:py312-x11-v2`，IDA Pro 9.4＋IDAPython。
- 可重生 exporter：`tools/ida_dump_field_poison_defeat.py`。
- sidecar：`work/field-poison-defeat-20260822.json`（不加入 Git；以 exporter、輸入 hash
  與本文件 raw anchors 重建）。
- 下列位址均為 **IDA linear**；seg0 project logical = linear − `0x10000`，
  seg0 file = logical + `0x1370`。原始 `sub_*` 名稱完整保留。

IDA 9.4 的 16-bit decompiler 沒有產生可靠 pseudocode；結論以 database function boundary、
xref、原始 bytes 與逐指令資料流為主，不把反編譯器缺輸出誤當作缺少 consumer。

## 已證實：每步死亡 writer 與存活人數判定

既有 `docs/137` 已閉合 `sub_193E3 → sub_19530 → sub_1964E`。本次 sidecar 再確認：

```text
0x196AF xor ah,ah                 ; 本員本步 damage byte
0x196B1 mov bx,[si+16h]           ; current HP
0x196B4 cmp bx,ax
0x196B6 jg  0x196CC               ; HP > damage：只扣血
0x196B8 mov word ptr [si+16h],0
0x196BD or  word ptr [si+38h],80h ; death bit
0x196C2 inc byte_27284            ; 本步新增死亡人數
0x196C6 call sub_1EBD8            ; 更新本員狀態圖
```

`sub_19530` 在逐員 consumer 後只要 `byte_27284 != 0` 就呼叫 `sub_1C5F6`。
後者逐一掃描現役隊伍：死亡位已 set 者不計；死亡位未 set 且 HP>0 時 `BP++`；
死亡位未 set 但 HP=0 時補寫 `0x80`。因此回傳的 `BP` 是**仍存活人數**，不是死亡人數：

```text
0x1C600 mov bp,0
0x1C60A test word ptr [di+38h],80h
0x1C611 cmp  word ptr [di+16h],0
0x1C617 inc  bp
0x1C61B or   word ptr [di+38h],80h
```

推論等級：**confirmed**（正式 movement caller、writer、同一 caller 內的 aggregate
consumer 與完整分支均閉合）。

## 已證實：部分死亡與全隊倒下分支

`sub_19530` 在 `sub_1C5F6` 後直接測試 `BP`：

```text
0x19621 cmp bp,0
0x19624 jz  0x1962C
0x19626 call sub_1BCF2             ; 尚有存活者
...
0x1962C call sub_15002             ; 全隊倒下
0x1962F mov  di,169h               ; D3TXT00 record 361
0x19632 call sub_21414
0x19637 call sub_15010
0x1963A call sub_1ED39
0x19643 call sub_1C03F
0x19646 call sub_1C7D9
```

`sub_1BCF2` 在隊長死亡時把第一位未死亡角色交換到 active slot 0，再重建每員狀態；
它不復活死亡同伴。`D3TXT00` record 361 的 raw glyph stream 是
`[65525,529,437,487,627,423,57]`，解碼為「{隊伍名}們被打倒了!」。

推論等級：**confirmed**。這也推翻現行 Go 註解「全滅後全隊滿血復活」；原版部分死亡
只重排存活隊長，全隊倒下才進下一節的復歸 consumer。

## 已證實：最後記錄點 writer 與敗北復歸

### 記錄點 writer

記錄設施 handler `sub_1527E` 在玩家確認、完成冒險之書交易後，把六個目前位置欄位
複製到專用保存欄位，再呼叫 persistence consumer `sub_115E2`：

```text
0x15349 [4F2F] → [4F48]   ; world X
0x1534F [4F31] → [4F4A]   ; world Y
0x15355 [4F33] → [4F4C]
0x1535B [4F35] → [4F4E]
0x15361 [256A] → [4F50]
0x15367 [256C] → [4F52]
0x15373 call sub_115E2
```

`sub_10030` 的新遊戲 initializer 同時寫目前／保存 world `(0x99,0xAE)`，並把
`[256A]/[4F50]=4`、`[256C]/[4F52]=0`；`(153,174)` 與原始 `cty_loc[0]` 的
阿里阿罕入口一致。這證明保存欄不是「最近進過的 town pointer」。

### 全隊倒下 consumer

`sub_1C7D9` 顯示 D3TXT00 record 362：

> 從某處傳來一個聲音：「你們就這樣放棄了嗎？」「再給你們一次機會吧。」

raw glyph stream 為
`[511,416,631,513,461,401,546,632,313,61,65532,59,60,528,529,633,398,537,634,635,423,534,58,65532,59,60,573,559,528,529,401,572,636,415,499,56]`。
之後它只對 `DS:507F` 所指的第一名角色：

```text
0x1C819 and word ptr [si+38h],0FF7Fh ; clear death bit
0x1C81E mov ax,[si+2Ah]               ; max HP
0x1C821 mov [si+16h],ax               ; full HP
```

其餘死亡成員沒有在此被補血或清死亡位。接著保存 world X/Y 與另外四個位置欄位被
寫回 current 欄，`sub_13008` 重新載入位置，再由 `sub_1BCF2` 重建 active party。

推論等級：**confirmed**（記錄點 writer、persistence caller、全滅 reader、唯一復活
writer、location reload consumer 與兩筆玩家可見 record 已閉合）。保存欄中的四個非 world
座標欄位在原程式內仍以 raw address 描述；remake 可用具型別的 CTY／section／local position
保存同一玩家可見交易，但不可把原始欄位逐一臆命名。

## Remake 規格

1. `applyHazardStep` 對每名本步由正 HP 降至 0 的角色標記死亡；至少一人存活時留在原地，
   死亡角色不得被旅店或一般移動暗中復活。
2. 全隊倒下時，正式 modal 依序顯示 pack text record 361、362；輸入只推進對話，不得穿透
   到移動或指令窗。
3. record 362 階段只讓第一名角色清死亡並回滿 HP；其他現役同伴維持 0 HP／死亡。
   MP、EXP、道具、旗標與目前世界交易不得藉由載入舊 snapshot 回捲。
4. 位置回到最後一次成功 `Save`／記錄點保存的位置。`Game.Save` 應先更新具型別 respawn
   location，再序列化；`Load` 應還原該位置。舊 Go 存檔沒有欄位時，以該存檔自身的當前位置
   作唯一可證明的 migration，不猜最近城鎮。
5. 新遊戲在第一次記錄前使用當前正式開場／阿里阿罕位置建立初始 respawn；不可硬寫 DQ3
   CTY、座標到共用 defeat primitive。
6. 戰鬥全滅與 field 全滅共用同一復歸 transaction；戰鬥既有 record 361 message 不得重複
   排入，仍須接 record 362 與相同單人復活／位置恢復。
7. record 362 的實字與 glyph 只放 game-pack JSON；Go 僅持有穩定 text ID 與通用 modal stage。

## 驗收

- EXE parity test 鎖定上述 branch bytes、record 361／362 glyph stream 與記錄點 writer。
- component：一人毒死時不傳送、存活者成為可操作隊長語意；全隊毒死依兩筆 modal，僅第一名
  滿 HP、其餘保持死亡，位置回最後 `Save`，進度不回捲。
- component：戰鬥全滅走同一 respawn transaction，且不重複 record 361。
- save/load round-trip：respawn location 持久；舊欄位缺失 migration 使用 loaded position。
- 完整 `TestOpeningProductionInputTrace` 由標題以正式 `InputState` 抵達 `THE END`；若出現下一個
  blocker，另開新 spec，不回退本文件已證實的原版 transaction。

## 正式 trace 重播訂正（2026-08-22）

接線後第一次完整重播在 CTY55 section0 走向 `(14,22)` 時正常觸發本規格的全滅
consumer：勇者與兩名同伴皆為 `HP=0`、poison，對話停在 rec361 階段。這證明舊 trace
之所以能繼續，是舊 Go 的「全隊補滿」近似暗中掩蓋了長途航行後未治療的毒；它不是
CTY55 地圖不連通，也不是新 respawn 實作退化。

第一個修正方案曾在奧莉薇亞事件後魯拉回 CTY2，準備進教會逐員解毒；乾淨重播卻在
入口 `(14,23)` 走向教會 `(2,13)` 途中再次開啟 rec361。原因是毒傷由每個成功移動步消費，
低 HP 隊伍連抵達 facility 前的二十餘步都無法承受；先在城外等待白天也會遇到同一問題。
因此該方案已推翻，不應保留成 current trace。

現行合法玩家策略改為在幽靈船前的 CTY38 道具店，與聖水一起透過正式商店 transaction
預備至少三份驅毒草；奧莉薇亞事件完成後、再次移動前，依 rec421「使用」與正式選人 modal
替仍存活者逐員解毒。死亡角色不能使用驅毒草；若當時已有死者，才在仍存活且已解毒的勇者
帶領下魯拉到 CTY2，經教會先復活、再逐員解毒並住宿，然後由 CTY38 港口登船。所有道具、
MP、金錢、船位、教會與旅店交易都由 production
handler 消費；不得改寫 HP、condition、位置、repel 或 RNG。這只是 deterministic 驗收路線，
不宣稱是原版唯一攻略，也不改動本文件已證實的敗北 transaction。

逐員解毒／必要時教會復活後的下一次重播已能安全離開 CTY2，但在 CTY38→CTY55 航程的
正式遭遇中全滅，並正確回到 CTY38 記錄點；舊 `traceSailToTown` 因未消費 defeat modal，
最後只報成「未進 CTY55」。隊伍在 CTY38 已正式購買聖水，故現行路線在啟航前透過道具
選單使用一瓶，使既有原版驅敵 consumer 覆蓋此航段。這不改 encounter table、repel 初值或
RNG；若未持有聖水就 fail closed，不由 trace 注入效果。

只以「是否有人死亡」決定要不要進 CTY2 的版本仍在返航遭遇中全滅：三人雖已由驅毒草
解除 poison，卻保留長航路消耗後的低 HP，沒有觸發教會分支便直接啟航。現行 gate 改成
任何現役角色死亡或 `current HP < max HP` 都走同一個正式 CTY2 教會／旅店補給；這是玩家
可見能力值導出的策略，不是提高角色數值。住宿後才回 CTY38 登船並使用聖水。

單次聖水仍短於完整 CTY38→CTY55 航線；`repel` 降到安全餘量後發生遭遇並再次全滅。
因此只有這條已知長航線要求 trace 在 `repel<=8` 時，若背包仍有聖水就再次經正式道具
選單使用；沒有道具時接受遭遇。不得延長單瓶效果或改全域航行 RNG。

後續診斷顯示失敗航次沒有任何 battle log，且終點 `shipAboard=false`；先前將它解讀為
遭遇全滅是錯的。奧莉薇亞強制移動／補給分支結束後不保證仍在船上，所以進長航路 helper
前必須按正式地表路徑重新登上停泊船，並在使用聖水後斷言仍 aboard。這是 trace 前置條件
訂正，不是船或聖水 production consumer 已被證明有錯。

再加 aboard 斷言後證明使用聖水不會離船；真正問題是先魯拉補給會把船重定位到 CTY38，
該停泊點與 CTY55 不在同一個可航連通區，航路 fallback 隨即上岸。現行路線不再為殘血／
死亡同伴離開海岬水域：CTY38 預先正式購買藥草，海岬後以驅毒草及藥草治療仍存活角色，
死亡同伴維持原版死亡狀態，待取得蓋亞之劍 checkpoint 後再走既有教會路線。這保留原船位，
也不讓 trace 以不可達港口冒稱有效補給策略。
