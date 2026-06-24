# 精訊版 DQ3 道具特效修正(bugs.md #7)

`bugs.md` item 4 列了三個「道具該有的特殊效果在精訊版失效」的問題:隼劍(飛鷹劍)只打
一次、魔法鎧甲無抗魔、祈禱之戒永不壞。docs/18 已把這三項定位為「功能未實作 / 與前提
衝突」,並指出真正要補的是程式而非資料。本文件接手,聚焦一個務實問題:**能不能用
binary patch 直接在 DOS EXE 補回這些效果,而不是只能等 SDL remake**。

結論先講:三項裡 **#7a(隼劍雙擊)做出來了**,而且不需要獨立的 code cave —— 靠改寫
EXE 內一段既有的判定碼、復用引擎原本就藏著的「再攻擊一次」路徑即可,同長度、不位移、
開機驗證不崩。**#7b(魔甲抗魔)需要在咒文扣血路徑掃描受方裝備,那是一個迴圈,EXE 內
沒有安全可放新碼的空間,誠實標註留給 C 層 / SDL2**。**#7c(祈禱戒指)反組譯確認損壞
邏輯本來就在、約 25% 會觸發,與「永不壞」的前提衝突,不需 patch**。

| # | 道具 | 結論 | 做法 | 信心 |
|---|---|---|---|---|
| 7a | 飛鷹劍(隼劍)item 0x6e | **已修** | 同長度區段改寫 + 復用既有 re-attack 路徑 | 高 |
| 7b | 魔法鎧甲 item 0x2b | 留 C 層 | 需裝備掃描迴圈,無安全 code cave | 高(根因) |
| 7c | 祈禱之戒 item 0x48 | 不需修 | 損壞邏輯本就存在(~25%),與前提衝突 | 高(衝突) |

機讀版:`docs/data/codecave_patches.json`;產生器:`tools/re_codecave_patch.py`。

---

## 先確認:官方有沒有修?

青衫攻略「修改:DQ3.EXE」一節列了 8 段當年社群用的 EXE patch:聲霸卡防當、魔王打不死、
敵人消失、五頭龍當機、某些機器當機、穿牆、致命一擊、生命不減。**沒有任何一段是隼劍 /
魔甲 / 祈禱戒指**。這反過來印證 docs/18 的判斷:這三項在精訊版是「功能根本沒接上」,
連當年的玩家社群也沒有對應修正碼可抄。所以 #7 沒有「最穩的官方碼」可用,要自己補。

---

## code cave 探勘:沒有安全的 in-EXE cave

任務原本設想用 code cave(在 EXE 空白區寫新碼、原碼插 jmp 導過去)。實際探勘 DQ3.EXE
後,結論是 **resident 主碼裡沒有可用的 code cave**:

- **連續 NOP 最長只有 6 byte**(file 0xa40f)。其餘 padding 幾乎都是函式之間 3 byte 的
  對齊填充,放不下任何含迴圈的新碼。
- 檔內確實有大型 0 填充區(file 0x16ca1 長 6578、file 0x189c9 長 3473),但它們落在
  **seg-0x1000 group 的 overlay 資料段**(緊鄰 `nvoc.vcx`/`fvoc.vcx` 檔名字串與資料
  表),不是固定 `CS` 可執行的程式區。
- 更關鍵:EXE 標頭 `e_minalloc = 0`,代表程式不向 DOS 多要 image 以外的 BSS;它在
  執行期用到的大緩衝(sprite buffer 段 0x1e8b、文字段 0x0f4f 等,皆為非 relocated 的
  runtime 值)是向 DOS 配置在 **image 之上**的記憶體,並不落在那些檔內 0 區。也就是說
  那些 0 區是資料段的一部分,不能保證戰鬥全程不被引擎當資料覆寫。

沒有「可執行、段固定、戰鬥全程不被覆寫」的 cave,硬塞會冒崩遊戲的風險。依專案紀律
(寧可少修一個、保證修正版不崩),改採另一條路:**不放新碼,改寫既有判定 + 復用既有
路徑**。這條路只有 #7a 走得通。

---

## #7a 隼劍雙擊 —— 復用引擎藏著的「再攻擊一次」機制

### 關鍵發現:re-attack 路徑本來就存在

反組譯玩家攻擊執行段(傷害計算 `sub` @ seg0 0xacce / file 0xc03e,唯一呼叫者在
file 0xc912)時,在傷害套用完之後(file 0xc1fa)有一段被忽略的判定:

```
f00c1fa: cmp byte [0x24fe], 0      ; [0x24fe] = 第二擊旗標
f00c1ff: je   0xc204               ; 旗標==0 → 檢查要不要再打一次
f00c201: jmp  0xc2f5               ; 旗標!=0 → 結束
f00c204: mov  si, [0x24a9]         ; 防方結構
f00c208: mov  cl, [si+0x16]        ; 防方 HP 低位元組
f00c20b: test cl, 0x10             ; ★ 觸發條件:HP bit4(語意可疑,形同未接線)
f00c20e: jne  0xc213               ; 成立 → 再攻擊
f00c210: jmp  0xc2f5               ; 不成立 → 結束
f00c213: mov  cl, 3 ... mov [0x24fe],1 ... jmp 0xbf97  ; ★ 重入攻擊 handler 打第二擊
```

`0xc213` 起的區塊會 `mov [0x24fe],1`(設第二擊旗標,防無限迴圈)再 `jmp 0xbf97`
**重入攻擊 handler** —— 也就是引擎本來就有「再打一次」的完整機制(會重印「XX 的攻擊!」
訊息、重新計傷),只是它的觸發條件被接到一個語意可疑的位元(防方 HP 低位元 bit4)上,
形同沒接線。這非常符合 docs/18 的旁證:D3TXT 戰鬥訊息「攻擊!」「受到 N 點傷害」成對
重複出現,暗示原設計本有連兩擊意圖,但攻擊碼從未正確觸發第二份。

### 修法:把觸發條件換成「武器是飛鷹劍」

只要把 `0xc1fa..0xc212` 這 25 byte 的判定區改寫成「第一擊(`[0x24fe]==0`)且攻擊者
武器 == 飛鷹劍(item 0x6e)就觸發第二擊」,既有的 re-attack 機制就會正確地讓飛鷹劍打
兩次。攻擊者由 `[0x259c]`(1-based index)取得,武器在隊員實體 `+0x3a` 首格。

改寫後 25 byte(file 0xc1fa,seg0 0xae8a):

```
f00c1fa: 80 3e fe 24 00   cmp byte [0x24fe], 0    ; 第二擊旗標
f00c1ff: 75 11            jne 0xc212              ; 已打過第二擊 → ret
f00c201: 8b 1e 9c 25      mov bx, [0x259c]        ; 攻擊者 index
f00c205: 4b               dec bx
f00c206: d1 e3            shl bx, 1               ; (idx-1)*2
f00c208: 8b b7 15 4f      mov si, [bx+0x4f15]     ; 攻擊者實體指標
f00c20c: 80 7c 3a 6e      cmp byte [si+0x3a], 0x6e ; 首格裝備(武器)== 飛鷹劍?
f00c210: 74 01            je  0xc213              ; 是 → 落入既有 re-attack 區塊
f00c212: c3               ret                     ; 否 → 結束本次攻擊
```

byte 變動(同長度,不位移):

| file offset | 原 | 新 | file offset | 原 | 新 |
|---|---|---|---|---|---|
| 0xc1ff | 74 | 75 | 0xc209 | 4c | b7 |
| 0xc200 | 03 | 11 | 0xc20a | 16 | 15 |
| 0xc201 | e9 | 8b | 0xc20b | f6 | 4f |
| 0xc202 | f1 | 1e | 0xc20c | c1 | 80 |
| 0xc203 | 00 | 9c | 0xc20d | 10 | 7c |
| 0xc204 | 8b | 25 | 0xc20e | 75 | 3a |
| 0xc205 | 36 | 4b | 0xc20f | 03 | 6e |
| 0xc206 | a9 | d1 | 0xc210 | e9 | 74 |
| 0xc207 | 24 | e3 | 0xc211 | e2 | 01 |
| 0xc208 | 8a | 8b | 0xc212 | 00 | c3 |

(`0xc1fa..0xc1fe` 的 `cmp byte [0x24fe],0` 與原碼相同,不變;共 20 byte 實際變動。)

### 兩個讓它剛好 fit 的細節

1. **失敗路徑用 1-byte `ret` 取代原 far jmp**:原碼「結束」是 `jmp 0xc2f5`,而 0xc2f5
   其實就是一個 `c3`(ret)—— 那個 jmp 只是繞路 ret。此處棧已平衡(前面 0xc1f4 已
   `pop bp/bx/di`),直接 `ret` 即可,省下原 far jmp 的 2 byte,讓新碼剛好塞進 25 byte。
2. **不影響開機路徑**:patch 只動戰鬥攻擊段;DOSBox 載入 `work/DQ3_item_fixed.EXE`
   開機正常進到標題與主選單(「DRAGON FIGHTER III 傳說的終章 ©1993 精訊資訊有限公司」
   →「▶遊戲開始 / 載入進度」),畫面無損毀。

### 為何用 item code 0x6e 而非 +5 旗標

ITEM.DAT 確認飛鷹劍 0x6e 的 `+5 = 0xc0`(bit7 為 double-hit 旗標)。更通用的偵測是
測 `+5 & 0x80`,但那需要在 EXE 內取得 ITEM 資料段指標再查表,改動更大;此處直接比對
武器 item code 0x6e,最小改動、最低風險。若日後 SDL remake,改用 `+5 bit7` 較通用。

---

## #7b 魔甲抗魔 —— 留 C 層 / SDL2

咒文傷害套用在 file 0xc22e 起的路徑;受方結構在 `di`,側別由 `[0xd75]&2` 區分
(`==0` 表受方為我方)。要做抗魔,得在扣血前掃描受方 8 格裝備(`+0x3a`,stride 2)找
魔法鎧甲 0x2b(`+6 = 0xd4`),命中就把傷害打折。問題:

- 裝備不在固定槽位 → 必須一個**掃描迴圈**,不是單點比較。
- 咒文扣血區是密集的 live 碼,沒有像 #7a 那樣 25 byte 的等量可改寫判定區可借。
- resident 主碼又沒有安全 code cave 可放迴圈。

硬塞會破壞咒文傷害公式或越界,風險高。依紀律標註:**需 C 層 / SDL2 重寫**——在咒文
傷害套用前呼叫 `magic_resist(target)`(掃裝備命中抗魔 bit 就減傷)。資料面其實沒缺
(魔甲與天使袍 `+6 bit2` 都置位),純粹是程式從不讀取。

> 註:同一段咒文路徑裡已有一處既有的傷害折半(`[bx+0xd4f]==3` → `shr bx,1`),是
> 攻方相關的修正,不是受方裝備抗性;不能直接拿來當魔甲抗魔。

---

## #7c 祈禱之戒 —— 損壞邏輯本就存在,不需修

祈禱戒指 use handler(file 0x5af4)反組譯:

```
f005af4: mov al,1; call 0xfa29        ; RNG → al
f005af9: cmp al, 0x40
f005afb: ja  0x5b0b                   ; al>0x40(~74.6%)→ 不壞
f005afd: mov di,0x188; lcall 111b:264 ; 印「啊,戒指壞了。」
f005b05: mov ax,0x48; call 0x5b0c     ; remove_item:掃 8 格 +0x3a 找 0x48 → [si]=0xff 移除
```

`al ≤ 0x40` 即損壞,機率約 0x41/0x100 ≈ **25.4%**。損壞路徑單一、必經、有移除實作。
反組譯結論:**損壞邏輯正常運作(~25%),與青衫「永遠不會用壞」的前提衝突**。`bugs.md`
item 4 是把幾個道具籠統歸成「等等之類的」,屬玩家憑記憶的概略陳述;7c 最可能是前提
誤記。**不需 patch**;唯一動態驗證點 = DOSBox 反覆使用祈禱戒指統計實際破壞率。

---

## 套用與驗證

```bash
# 僅驗證原 bytes 相符(不輸出)
python tools/re_codecave_patch.py --verify
# 從原版產生只含 #7a 的修正版
python tools/re_codecave_patch.py            # → work/DQ3_item_fixed.EXE
# 疊在 A 的卡關修正版之上(#1/#2/#3 + #7a)
python tools/re_codecave_patch.py --base work/DQ3_fixed.EXE   # → work/DQ3_full_fixed.EXE
```

套用前逐 byte 比對 `codecave_patches.json` 的 `orig_bytes`,不符即中止(防底檔錯位)。

**已驗**:DOSBox(`dq3-dosbox` image)載入 `DQ3_item_fixed.EXE` 開機 → 標題 → 主選單
皆正常,畫面無損毀,證明 #7a patch 不影響開機與基本流程。

**待驗(需進戰鬥)**:裝備飛鷹劍實際打兩次的行為,需載入存檔進戰鬥才能確認,超出
有界自動化 harness 可靠覆蓋範圍;此處以靜態反組譯正確性 + 開機完整性背書。

---

## 工具

- `tools/re_codecave_scan.py` — code cave 探勘 + 攻擊 / 傷害 handler 反組譯(docker capstone)。
- `tools/re_codecave_patch.py` — #7 修正 patch 產生器(讀原檔→驗 orig bytes→套用→work/)。
- `docs/data/codecave_patches.json` — patch 與 cave 探勘結論機讀版。

## 殘留 / 待釐清

- #7a:飛鷹劍進戰鬥實打兩次,需存檔進戰鬥 DOSBox 實測;另可改用 `+5 bit7` 通用偵測。
- #7b:抗魔需裝備掃描迴圈,EXE 無安全 cave;留 SDL remake(`magic_resist` 打折)。
- #7c:與青衫前提衝突,需 DOSBox 統計祈禱戒指實際破壞率佐證。

## ITEM.DAT 7-byte 結構(裝備系統 RE)

`ITEM.DAT` = 128 筆 × 7 byte,index = 道具碼。反組譯 + 對照已知道具(BBS 數值表 docs/history)解出:

| byte | 語意 | 佐證 |
|---|---|---|
| `+0` | **攻擊力**(武器 b0;防具=0)| 木棒7、銅劍10、鋼劍28、理力之杖55、王者之劍120、破壞之劍110(對上 BBS)|
| `+1` | **防禦力**(防具 b1;武器=0)| 布衣2、皮甲胄8、鐵甲22、鋼甲32、魔法鎧甲40、光之盔甲70 |
| `+2/+3` | **價格**(u16 = b2 + b3×256)| 王者之劍 136×256+184 = **35000**(對上 BBS「三萬五」);鋼劍 1500、理力之杖 2500 |
| `+4` | **類別 / 裝備部位**(32=武器、64=防具身、24=道具)| — |
| `+5` | **武器特殊旗標**(bit7=雙擊 #7a)| 飛鷹劍 0xc0、理力之杖 0x80 |
| `+6` | **可裝備職業 bitmask**(`0x80>>cls`;0xff=全職業、0x80=勇者專用,#7b bit2=抗魔)| 王者之劍 / 光之盔甲 = 0x80(勇者專用,對上 BBS)、布衣 = 0xff |

remake:`dq3_item_attack/defense/price/category/can_equip`(dq3_combat);`dq3_recruit` 加
`weapon`/`armor` 槽;戰鬥 **攻擊力 = 力量 + 武器 b0、守備力 = 耐力/2 + 防具 b1**。實測:
勇者裝銅劍(攻+10)物理傷害提高。待:商店(買賣)+ 裝備指令 UI(可裝備職業用 b6 防呆)。

## 商店 / 開場 RE(部分)

- **道具店 handler**:file 0x8af3 印歡迎詞(D3TXT00 rec 0x119「歡迎光臨,這裡是道具店」),
  `[0x408a]=0x1b1` → 買單 overlay。**買到的道具放進玩家 entity 庫存**:0x8ab7 `bx=[0x722];
  si=[bx+0x4f15]+0x3a;cx=8` 掃 8 格找空(0xff)填入(entity +0x3a = 8 格道具欄)。
- **商店「貨架」來源**:店主 NPC 的庫存(entity +0x3a),載自 CTY NPC 資料;per-town 精確
  貨架表仍待深 RE(shop overlay + CTY NPC 庫存欄)。remake 暫用合理早期選品(真實 ITEM.DAT 價/限制)。
- **主角開場初始**(file 0x1c33):新遊戲主角 `[si+0x15]=1`(等級1)、`call 0xeecf`(學咒文 sub_db5f)、
  **`[si+0x3a]=0x1e` = 起始裝備「布的衣服」(0x1e)**。⇒ remake 主角創角後預設帶布衣。

### per-town 貨架表(界限)

CTY NPC 記錄僅 **7 byte**(`{X,Y,b2,b3,b4,flags,b6}`,docs/34),裝不下 8 格商店庫存 →
**商店貨架不在 CTY NPC**,而是 EXE 全域表(由 shop id 索引,經 overlay 買單 0x9bcf / `[0x408a]`
選單系統顯示)。該表追蹤需深入 overlay 選單資料流,**待續**。remake 暫用 curated 早期貨架。
