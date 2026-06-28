# Boss / special 事件正式觸發點盤點(第一性原理)

> 使用者洞見(2026-06-26):boss 不是會走動的 NPC,是地圖上**固定一格(或數格)放著 sprite 的物件**,
> 玩家走過去按**空白鍵(examine)**就觸發開戰。所以「boss 正式觸發點」== 某 CTY/section/座標上
> 的 `kind=special` 事件格。純靜態可定位,不需實機探索。

## 權威來源

`docs/data/npc_dialogue.json` —— 已 RE 解碼的全 CTY NPC dump,每筆欄位:
`sec, bank, x, y, sprite(entry), sub, kind(talk/facility/special), dlg(事件 record 號), text`。
這是反組譯產物,比手刻 section parser 可靠(手刻易踩 section 偏移錯位陷阱,見記憶
`re-section-table-count-prefix-offby1`)。工具:`tools/list_special_events.py`。

- `kind=special` = examine 觸發事件(boss / 機關 / 劇情觸發點),全遊戲共 **64 個**。
- `dlg` = 事件 record 號;boss 事件如八頭大蛇 **CTY19 sec1 (35,12) dlg=45** 對上既接的 byte4=45。
- ⚠ `dlg` 是事件 id,**不保證等同** dq3_scripted 的 byte4(多數一致,個案需以對話/攻略確認)。
- ⚠ special **同時涵蓋** boss、機關(石門/按鈕)、give 道具劇情 NPC,不是全部都是 boss。

## 辨識 boss 的結構特徵(第一性原理)

1. **2×2(或多格)同 sprite 群 + 同 dlg**:boss 大 sprite 跨 4 格,4 筆 special 同 sprite/同 dlg。
2. **對話為「‥‥‥‥」(沉默)**:boss 開戰前不說話,examine 即進戰鬥。
3. sprite entry **不一定大**(八頭大蛇 b2=14);關鍵是 special + 結構,非 sprite 數值大小。

## 已確認的 boss 觸發點

| CTY | sec | 座標 | 結構 | 狀態 |
|---|---|---|---|---|
| 19 | 1 | (35,12) | 八頭大蛇 dlg=45 sprite=14 | ✅ 已接(main.c cur_cty==19 b4==45)|
| 14 | 1 | (14,12)(14,13)(15,12)(15,13) | dlg=58 examine 事件 | ✅ 已接 = **甘達特盜賊巢穴**(2026-06-27 確認:main.c CTY14 examine → 甘達特手下怪27 → 甘達特怪26 → flag 0x211,與古布達黑胡椒鏈閉環)。**非巴拉摩斯城**(舊「身分未定」推測已更正)|

### CTY14 調查結論(2026-06-26,誠實修正)

**已確定(靜態 ground truth)**:
- CTY14 = overworld (106,114),緊鄰 CTY15 巴哈拉達鎮 (101,119);只有 sec1、12 NPC、**0 facility(無店)**。
- sec1 中央 (14,12)(14,13)(15,12)(15,13) 規整 2×2:talk 對話全「‥‥‥‥」沉默 + sub=2 special dlg=58。
- 無店 + 全沉默 + 2×2 examine 事件 = **boss 房擺位**(boss 開戰前不說話,examine 進戰鬥)。這是個 boss 觸發點。

**未能確定(需追 EXE 事件 dispatch)**:
- **boss 身分未定**。原推論「sprite=46 = 特定 boss」**已推翻**:sprite=46 在 CTY84(達姆杜拉)是普通 talk 居民,
  是通用人物 sprite,非怪物專屬。
- `special` 的 `dlg` 是**事件 id,非 D3TXT record 號**(special 無 text 欄);把 dlg=58 對到 D3TXT03 rec58
  「流星手環提示」是**錯誤對照**,兩者無關。
- dlg(事件 id)→ 怪 id 的映射在 EXE 事件處理碼裡(間接跳表 dispatch,見記憶 `re-indirect-jumptable-dispatch`),
  不在文字/NPC 表。要定身分得靜態追該 dispatch;當前實機/工具環境受限,留待後續。

> 教訓:`kind=special` 能可靠定位「**哪一格是 examine 觸發事件**」(回答了使用者「boss 在哪一格」),
> 但「**那格觸發哪隻怪**」需另一層 RE(事件 id → 怪 id dispatch)。前者已解,後者待追。
> 不靠單一線索(sprite 值)臆斷身分——sprite 是通用的,會誤判。

### CTY14 boss 身分:追 EXE handler + 對話 record 後的結論(2026-06-26)

用 `tools/dis_sub2_handler.py` 反組譯 byte4=58 handler(file 0x07250,跳表 0x3bb4):
```
L0x5ee0: call 0x6372; [0xb34]=1; di=0xc22(rec106); lcall 渲染;
         cmp [0x722],1; jne …; di=0xc23(rec107); jmp 0x6380(dispatcher)
```
- **byte4=58 本身不含戰鬥呼叫**,是 boss 房**前置對話 NPC**(沉默 2×2 + examine 觸發對話)。
- 對話 = D3TXT03 rec106/107/108:**「你們是要來加入的嗎?」「大王現在不在,你們下次再來吧」
  「那麼就不能讓你們通過,看招吧」** —— 典型**盜賊巢穴守衛 boss 房對話**。
- D3TXT03 涵蓋區(docs/32:112)= 諾阿尼魯/阿莎拉慕/伊席斯/**巴哈拉達**/劇揚;CTY14 (106,114) 緊鄰
  巴哈拉達 CTY15 (101,119)。
- 對照 D3TXT03 救人段尾「**甘達特「呼呼呼,我回來的正是時候!**」+ 攻略古布達/達妮亞救人(巴哈拉達洞窟
  打甘達特手下 → 遇返回的甘達特)。

**結論(高信度)**:CTY14 sec1 = **甘達特的盜賊巢穴 / 巴哈拉達洞窟 boss 房**;守衛 boss = **甘達特手下(怪27)**,
深處 boss = **甘達特(怪26)**。與既接的古布達救人鏈(flag 0x211,boss:27;boss:26)**完全閉環**——
這正是該救人事件的「城內正式觸發點」所在(byte4=58 對話 → region [0x722]=1 觸發戰鬥)。

**戰鬥已接**(2026-06-27,與上表行 30 一致):`main.c` CTY14 sec1 examine `byte4=58` → 未救出(flag 0x211)則
打甘達特手下(怪27)、救出後給後話。原版精確 region 觸發(`[0x722]=1` region handler)雖未逐指令重現,
但 remake 以 byte4=58 examine 達成等效劇情強制戰,玩家走進 CTY14 巢穴即觸發。

## 巴拉摩斯城 CTY 定位(未決)

- BBS 史料(docs/history/dq3-bbs-1994.md):「打贏巴拉摩斯後回阿里阿罕,索瑪現身震垮蓋亞那洞穴牆→跳下進異界」。
- npc_dialogue.json 提「巴拉摩斯」的 CTY:CTY00(阿里阿罕,破關後傳聞 ×4)、CTY16、CTY25(×4)。
  這些是**提到**巴拉摩斯的台詞,非巴拉摩斯**所在城**。
- 巴拉摩斯本體已接(`baramos` boss token + flag 0x213 + 索瑪前序列)。**CTY14 已確認 = 甘達特盜賊巢穴
  (非巴拉摩斯城)** —— 舊「CTY14 為巴拉摩斯城候選」推測已更正。巴拉摩斯「所在城」的精確 overworld
  定位非破關必須(boss 已可觸發),屬 RE 長尾。

## 全 64 special 清單

見 `tools/list_special_events.py`(無參數列全部;`CTYnn` 篩單城;`--text` 附對話)。
已接的 byte4/dlg:scripted 表 {7,12,16,25,31,44,49,50,52,84} + main.c 特例 {(14,58)甘達特巢穴,(15,25),(19,45),(82,64)}。
其餘為提示/居民/設施對話(special-events-audit 已逐一判讀,0 review 真缺口),非 boss。
