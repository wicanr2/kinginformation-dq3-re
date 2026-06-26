# RE 攻關日誌:`[0x722]` state machine(runner 觸發系統)

> 目標:逐行反組譯 `[0x722]` 的 state machine,搞清它到底是什麼、誰設它 → 決定 warp/portal/
> sub2 對話分支能不能接。**逐步記錄,成功失敗都留。** 位址:file = logical + 0x1370。

---

## Step 1:全掃 `[0x722]` 存取 → ★ 推翻「無 setter」

掃程式區所有指令(capstone),運算元含 `0x722`:
- **寫(setter):57 處**(大量 `mov word [0x722], 1`、少數 `=5/=ax/=bx/=si/=dx`、`inc/dec/add/sub`)
- **讀:165 處**(大量 `cmp [0x722], 1` / `cmp [0x722], 2`,也有 `cmp ...,3/4/9`)
- 其他:12 處(`push/pop [0x722]` 存堆疊、2 個 imul 假陽性)

**★ 重大更正**:docs/31/44/47 說「event id `[0x722]` 資料驅動、無靜態 setter」**是錯的**——
`[0x722]` 有 **57 個明確 setter**。且值域小(1/2/3/4/5/9),讀多為 `cmp ==1/2`。
⇒ `[0x722]` 比較像**狀態機的 state/mode 變數**(不是純資料驅動的 event id)。
之前「卡 runner」的前提需要重新檢視。

**最密集 cluster**:file 0x10a46–0x10e8f(logical 0xf6d6–0xfb1f)有極多 inc/dec/cmp/mov [0x722]
與 dx/ax/0 互動 → 疑為 **state machine 主體**。下一步逐行反組譯它。

## Step 2:反組譯密集 cluster(0xf6d6)→ 是「座標→region 命中測試」非事件分派

```
cmp cx,0xc0 / cmp dx,0x1a       ; 座標界檢
dx=(cx>>5)+1                    ; 座標 → cell 索引
cmp [0x722], dx                 ; 比對當前 [0x722] 與算出的 region
[0x71c]==3 分支:讀 [si+0x18/0x1a]=region{x,y}、[0x72c]=寬、[si+0x14]=高
迴圈掃 region 清單(6 byte/項 {x,y,..}),算命中 index=bp → cmp [0x722],bp; je 命中
```
⇒ **`[0x722]` = 當前 region/page 索引**(座標除算 + region 清單命中比對),**不是 runner event id**。
docs 把 `[0x722]` 當「event id」八成是**誤判**。下一步:找真正 runner 事件分派 `call [bx+0x3baa]`,
看 bx(event id)從哪來 —— 確認 [0x722] 與 event id 的關係。

## Step 3:找事件派發 → logical 0x9871 `call [bx+0x3baa]`

`call [bx+0x3baa]` 僅 1 處:L09871(file 0xabe1);`lea si,[0x3bb4]`(sub2 表)2 處:0x4ff2/0x96fe。
反組譯 runner 看 bx(event id)來源 + 與 [0x722] 關係。

## Step 4:✓ runner 派發確認 — bx = [0x722]

runner(函式 0xabb2 / L09842):
```
L09863: cmp byte [0x726],1 ; je → 跳過派發    ; [0x726]==1 gate
L0986a: mov bx, [0x722]                       ; ★ bx = [0x722]
L0986e: dec bx ; shl bx,1                      ; bx=([0x722]-1)*2
L09871: call [bx + 0x3baa]                     ; 派發 [0x722] 號 handler
```
⇒ **`[0x722]` 就是 runner 的 event/region 索引**。

## ★★ 核心結論(推翻舊 blocker)

`[0x722]` 同時被兩條路設定,**兩條都是靜態可決定的**:
1. **座標→region 命中測試**(Step 2,L0f6d6 cluster):玩家座標 → 除算 + region 清單比對 → 算出
   region index → 寫 [0x722]。region 清單是資料(可抽)。**玩家走到某區 → [0x722] 自動 = 該區號**。
2. **57 個明確 setter**(Step 1):故事推進直接 `mov [0x722], N`(進城/事件後等)。

⇒ **舊結論「[0x722] 無靜態 setter、純動態、追不到」是錯的**。它有 57 setter + 座標 region 規則,
**全部靜態**。所以 **warp/portal/sub2 的「何時觸發」是可以靜態 RE 出來並 wiring 的**——
不需要 DOSBox 動態!之前卡住是因為把 [0x722] 誤判成「不可知的動態 event id」。

## 決策點(待定)
要 wiring 的話,路徑:
- (A) 抽**座標→region 命中測試的 region 清單**([0x71c]/[0x72c]/region structs)→ 重建「玩家在哪區」。
- (B) 對每個 warp/portal handler,讀它 `cmp [0x722],N` 的 N → 對應到 region/事件。
- (C) sub2 對話分支:同樣 handler 內 `cmp [0x722]/旗標` → 選 rec(用戶推測 sub2 也吃 state machine,待驗)。

## Step 5:✓ 確認 user 推測 — sub2 = runner 同一張跳表

`0x3bb4 - 0x3baa = 10 byte = 5 entries`。⇒ **sub2 表(0x3bb4)= runner 表(0x3baa)偏移 5 格**,
**同一張跳表**。sub2 NPC byte4=N → 派發 runner event(N+5)的同一 handler。
⇒ **sub2 對話分支 = runner 事件的一部分**(user 推測正確)。warp/portal/sub2 三者**同源**,
都由 [0x722] / event 索引驅動,而 [0x722] 全靜態可決定(Step 4)。

## Step 6:✓ 樣本完整解 — sub2 條件對話 handler(拿吉米老人 byte4=12)

handler L0x537f 逐行:
```
test_flag(0x2d):                      ; 0x8279=test_flag, 0x8264=set_flag
  未設 → di=0xbee → rec54「我的夢想快實現」→ 結束
  已設 → di=0xbeb(rec51 intro); cmp [0x722],1(region):
           ==1 → [0x2593]=0x55(給盜賊鑰匙)+ 0x7bbe(give)+ di=0xbec(rec52 給予話)+ set_flag(0x2d)
           !=1 → di=0xbed(rec53)
```
**di→rec 公式**:`rec = di − 0xbb8`(gen_sub2 已有;驗:波魯多加 di=0xbd4→rec28 ✓;
波魯多加 handler test_flag(0x37/0x38) + [0x2593]=0x5c 胡椒,同模式)。

⇒ **樣本驗證成功**:sub2 條件對話 = `test_flag(X) [+ cmp [0x722],region] → 選 rec(di−0xbb8)[+ 給物 [0x2593]]`。
**全可靜態抽**。每個 handler 都這結構 → Step 2(全 wiring)可系統化:逐 handler 抽
{flag, before_rec, after_rec, give_item, region_cond}。

## Step 7:✓ 樣本 wiring 完成(拿吉米老人條件對話)

main.c:對話拿吉米老人 → 未有鑰匙=給+rec52「給予話」、已有=rec53「後話(我的夢想和靈魂…)」。
實測兩態 + in-game 渲染(najimi_after.png)都對。**穿幫(拿到鑰匙仍講 before 話)修復。**

★ **Step 1 結論:條件對話 wiring 路徑完全打通**(handler 反組譯 → flag/region/give/rec 全可抽 →
di−0xbb8 取 rec → remake 接)。**不需 DOSBox。** → 可進 Step 2(系統化全 wiring)。

## Step 8:Leve 魔法玉老人(byte4=7)條件對話 wiring

handler L0x521f:test_flag(0x2a≈持鑰匙開門)未設→rec65 before「海對岸有勇者」;
設+給 0x58 魔法玉+rec60 給予話。**更正**:之前 Leve wiring 用「任一 sec1 NPC」不精準,
真正給予者是 **byte4=7**(3,3)。已改精準 + 條件對話。實測兩態 ✓。

## Step 9:✓ data-driven 通用 handler(Step 2 架構完成)

`dq3_scripted.{h,c}`:scripted NPC 表 {byte4,cty,give_item,prereq_item,milestone,
before_rec,give_rec,after_rec};main.c 一段通用 handler 跑全部(前置未滿→before、
持前置未給→給+milestone+give_rec、已給→after)。拿吉米(12)+Leve(7)改走通用表,
取代逐個 if-else。實測兩 NPC 三態正確,31/31。**加新 scripted NPC = 表加一行**
(從 docs/data/sub2-handlers.md 抽 byte4/give_item/recs,查 CTY+前置)。

## Step 10:★ 驗證糾錯 — 區分 GIVE(0x7bbe)vs TAKE(0x7c0c)

逐個驗證時發現 byte4=16 用 `call 0x7c0c`(取/檢查)非 `0x7bbe`(給)→ 它是**需求道具** NPC 非給予者。
更新 gen_sub2_full 區分 give/take。重列:**純給予 NPC** = byte4 7,12,31,52(其餘 8/9=羅馬利亞收皇冠、
16/44/50=需求道具、25/26=Portoga barter、33-36=warp、48/49=barter)。
⇒ 待加表:**byte4=31(CTY22 給0x4b)、byte4=52(CTY67 給0x65)**。其餘 TAKE/barter 另特例或已接。

## Step 11:★★ section 偏移表「count-prefix」差一格 bug → NPC 座標全錯位(再糾錯)

逐一驗證 byte4=31/52 時撞牆:warp 到 section 0 怎麼觸發都沒反應。
反查 `dq3_town.c::dq3_town_load` 註解才發現:**CTY section 偏移表無 count 前綴**,
`so = word[section]`(word[0]=section0)。我臨時寫的 `find_npc_b4.py` 卻把 word[0] 當 count、
從 word[1] 起算(`offs=[u16(2+2i)]`)→ **整個差一格**:回報的「sect0」其實是 section1,
且漏掉真正的 section0。導致:

- byte4=31 被誤標 CTY22 **sect0** (3,4),實際在 **sect1** (3,4)。warp section 0 自然觸發不到。
- byte4=52 用同樣壞 parser 的 `_sweep52.py` 掃 → 誤判「無任何 sub2 NPC 帶 b4=52」,
  我一度把它當 runner/region 事件**錯誤排除**。

修正 parser(`so=word[i]`,無 count)後重掃,兩者都還原為乾淨 sub2 GIVE NPC:
- byte4=31 → CTY22 **sect1** (3,4) GIVE 0x4b,tflag/sflag 0x3c,rec45/46。
- byte4=52 → CTY67 **sect0** (14,24) GIVE 0x65 光之珠,sflag 0x4e,rec71(龍女王,對抗索瑪關鍵)。

**教訓**:NPC「從哪來」撞牆時別退回「動態/runner」結論(memory: dont-retreat-to-overlay-dosbox)——
先確認自己的解析器與遊戲實際 loader(C 端權威)逐欄位對齊。一個 off-by-one 的 header 解讀
就讓整批座標錯位、進而誤判「資料不存在」。

## Step 12:✓ 權威全掃 `tools/sweep_sub2_npcs.py` → `docs/data/sub2-npcs-live.md`

用修正後 parser 掃全部 CTY 的**實存** sub2 NPC,逐個交叉比對跳表 handler 的
GIVE/TAKE/tflag/sflag/region/warp/recs。這份才是 wiring 權威清單(取代憑 handler 表反推 CTY)。

實測落表(35/35 playthrough 綠):
| byte4 | CTY/sect/(x,y) | 行為 |
|---|---|---|
| 7  | CTY1 s1 (3,3)   | GIVE 0x58 魔法玉(tflag 0x2a 持鑰匙)|
| 12 | CTY8 s3 (9,9)   | GIVE 0x55 盜賊鑰匙(tflag 0x2d)|
| 31 | CTY22 s1 (3,4)  | GIVE 0x4b 水槍(flag 0x3c)|
| 52 | CTY67 s0 (14,24)| GIVE 0x65 光之珠(sflag 0x4e,無前置)|

**純 GIVE sub2 NPC 至此全數接齊。** 其餘待接(另特例/需擴充):
- 條件 flag-gated give:byte4=84(CTY16 GIVE 0x10,flag 0xe5)、33/34(CTY44 GIVE 0x66 + warp)。
- barter(TAKE+GIVE,需通用 handler 加 TAKE):25(CTY15)、48(CTY83)、49(CTY64)。
- 純 TAKE/收物(非給予,另特例):8/9(皇冠)、16(CTY5)、44(CTY54)、50(CTY62/63)、26(Portoga 已特例)。

## Step 13:★ 區塊邊界修正 → barter「批」其實是 flag 鏈子任務,非簡單 barter

加 block 邊界(jmp 0x6380/ret 截斷)後重掃,推翻先前「25/48/49 是 barter」的判讀 ——
那是 sweep 掃過 handler 邊界把**下一個** handler 的 give/take 算進來的假象。真實單區塊下:

- 真正單區塊 give/take 只有 **byte4=26**(Portoga 胡椒→船,已特例)與 **52**(龍女王給光之珠,已接)。
- 其餘給予型(7/12/31/49/...)都是 `test_flag(prereq); je GIVE` 結構:**給予分支在 je 目標之後**
  (另一塊),給予**閘在 flag**。`tools/decode_sub2_struct.py` 解出完整雙區塊 →`docs/data/sub2-struct.md`:
  每個 {always_rec, prereq_flag(je極性), before/give_rec, give/take item, set_flag}。
- 道具經 inventory.h/main.c 定錨:0x5b 船、0x5c 黑胡椒、0x6b 銀寶珠、0x65 光之珠、0x73 雲雨之杖…。
  48-50/63-68 是「船/銀寶珠/0x66」的多 NPC flag 鏈(0x4a→0x4b→0x4c)子任務。

**決策**:完整 flag 機器(誰設各 prereq flag = [0x722] region machine + examine 事件圖)是另一條
大型 RE 線;4 個主要給予 NPC 觀感行為已驗證。依目標優先序(可破關>polish>驗證>打包),
忠實結構**存資料**(sub2-struct.md)供後續精修,接線維持已驗證觀感給予,時間投入完成度+polish+打包。

## Step 14:剩餘 sub2 handler 接線完成(2026-06-26)

`tools/sub2_worklist.py`(行為 × 真實 NPC × 已接 交叉比對)產出權威清單,把剩餘有真實 NPC 的
give/take handler 全部接齊:

### 給予型(give_item)— 觀感給予,逐一遊戲內驗證
| byte4 | CTY/位置 | 給物 | 對白(D3TXT bank)|
|---|---|---|---|
| 25 | CTY15 巴哈拉達 (5,24) | 0x5c 黑胡椒 | 胡椒商古布達獲救後給(rec118/119/120,bank3)|
| 49 | CTY64 (16,10) | 0x6b 銀寶珠 | 「粉碎大魔王陰謀」→ 給(rec84/85,bank1)|
| 84 | CTY16 (33,24) | 0x10 誘惑之劍 | 與卡爾洛斯重逢答謝(rec98/99,bank4)|

### 檢查型(require_item)— 新增 handler 分支(0x7c0c 檢查、**不消耗**)
`dq3_scripted` 加 `require_item` 欄;持物→success rec、缺物→need rec。判定 0x7c0c 為「檢查」非
「消耗」的關鍵證據:**byte4=50 require 船 0x5b**(不可能消耗玩家的船 → 必為 gate 檢查)。

| byte4 | CTY/位置 | 需求物 | 缺物 rec / 持物 rec |
|---|---|---|---|
| 16 | CTY5 精靈之村 (17,7) | 0x59 夢幻紅寶石 | rec47 / rec90 |
| 44 | CTY54 (8,2) | 0x62 變身杖 | rec61 / rec56 |
| 50 | CTY62/63 (cty=0xff) | 0x5b 船 | rec89 / rec87(渡海 gate)|

### 不接(查證後排除)
- **byte4=35**(give 0x66):handler 存在但**無任何 NPC 引用** → 開發殘留。
- **byte4=64**(require 0x66 綠寶珠):handler 無對白 rec → 非乾淨對話 NPC(byte4 63-68 同檢查 0x66,
  疑為非對話 gate 或 decoder 未涵蓋),暫不接。
- 道具語意定錨:0x59 夢幻紅寶石 / 0x62 變身杖 / 0x66 綠寶珠 / 0x5c 黑胡椒 / 0x6b 銀寶珠
  (font 渲染 D3TXT00 rec=code+1)。

⇒ **所有「有真實 NPC 的 sub2 give/take handler」接齊**:給予 7/12/25/31/49/52/84 +
特例 9(Romaly 收皇冠)/26(Portoga 胡椒換船)+ 檢查型 16/44/50。game_tester **32/32** 全綠。
道具經濟與條件對話的 sub2 子系統至此完成靜態還原 + remake 落地。

## Step 15:綠寶珠 questline — 提頓村牢房犯人(runner 給予)+ 祭壇 stub(2026-06-26)

青衫攻略線索:**提頓村(CTY20)牢房犯人給綠寶珠 0x66**,送往遙遠南方雷亞姆蘭特祭壇。
這解開了 byte4=64「require 0x66 但無對白」之謎 —— 0x66 的**給予端不是 sub2 talk-NPC**:

- **給予端**:byte4=35 handler(L0x593a,give 0x66,rec38「把這個寶珠拿去吧」/rec39,flag 0x3e)
  **無 sub2 NPC 引用** → 是 runner/region 事件,原版閘在「**夜晚進村 + 開牢門**」(黑暗神燈變夜)。
  CTY20 的實體 NPC 是 **b4=29 sub=0** @ (16,2):平時是**屍體 + 牆上留書**(rec29「帶的寶珠留給有緣人」)。
  ⇒ remake 無日夜系統 → 簡化:talk 提頓村 (16,2) 犯人 → 給綠寶珠 0x66 + rec29 留書(main.c 特例)。
- **祭壇端**:byte4 **63-68** handler(L0x60b5-0x60d3,6 個近乎相同)只 `[0x2593]=0x66; call 0x7c0c`
  檢查綠寶珠,**無對白 rec、無 give、無 flag** → **stub handler**(未發售版祭壇事件未做完,
  同 ENDTXT/結局未定稿模式)。⇒ 不接(無觀感反應)。

**道具來源教訓再驗證**(memory `re-item-source-check-walkthrough-first`):0x66 一度被當「無給予 NPC」,
實際給予在 runner 事件 + 攻略才講得清(夜晚/牢房條件)。**sub2 清單掃不到 ≠ 道具無來源** —— runner/region
事件、examine、sub=0 特殊 NPC 都是合法給予管道。game_tester 33/33。

## Step 16:★ 不死鳥祠堂 = 六寶珠祭壇(攻略確認 + build 不完整)2026-06-26

青衫攻略:諾阿尼魯村(CTY4)往北雪白陸地的**不死鳥祠堂**,把**綠藍紅紫黃銀六珠**放祭壇 →
與守護巨蛋的守護者交談 → 復活不死鳥拉米亞(飛行坐騎)。**靜態資料逐項確認設計**:

- **六寶珠 = 連續道具 0x66-0x6b**(font 渲染 D3TXT00 逐一確認):
  0x66 綠寶珠 / 0x67 藍寶珠 / 0x68 紅寶珠 / 0x69 紫寶珠 / 0x6a 黃寶珠 / 0x6b 銀寶珠。
  **完全對上攻略「綠藍紅紫黃銀」**。(註:祭壇紅珠=0x68 紅寶珠,與支線 0x59 夢幻紅寶石不同物。)
- **祭壇檢查邏輯**(file 0x7450):`cx=6; [0x2593]=0x66; loop{ call 0x7c0c 檢查; [0x726]==0→失敗跳出;
  inc [0x2593] }` → 依序檢查 0x66..0x6b **全持有**才成功。
- **6 個祭壇位置 handler byte4=63-68**(L0x60b5-0x60d3,各 6 bytes):`mov bp, 0x8f..0x94; jmp 0x7446;
  mov bx,bp; test_flag; je`。即各位置綁一個**「該珠已放上」旗標 0x8f-0x94**。
  ⇒ 先前誤報「63-68 都 TAKE 0x66」是 decoder 看到 fall-through 到 0x7450 的 [0x2593]=0x66。
  byte4=64(唯一有真實 NPC,CTY82)= 0x90 藍珠位 → **CTY82 疑為不死鳥祠堂**。

- **★ build 不完整**:粗掃旗標 setter(`mov bx,0xNN; call set_flag`)——0x8f/0x90 各 1、**0x91-0x94 為 0**。
  即「放珠上祭壇」設旗標的觸發**大部分沒接** → 不死鳥復活序列在此未發售早期 build **未完成可觸發**
  (同下降 event 86、祭壇 stub、ENDTXT 未定稿模式)。

**結論(有確認嗎 → 有,且更精確)**:攻略描述的**設計**被資料完整佐證(六珠 id 連續、祭壇全檢、
6 位置旗標),但這個**早期 build 沒把放珠/復活觸發接完**。remake 已可取得其中 2 珠
(綠 0x66 提頓村、銀 0x6b CTY64);另 4 珠(藍/紅/紫/黃)給予端同屬 runner 事件,本 build 未必接齊。
