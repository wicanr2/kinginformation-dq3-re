# NPC 對話系統 + 逐句 mapping

> 把「城鎮/迷宮裡每個 NPC 站哪、是什麼、說什麼」從地圖資料 + 對話檔解出來,並處理對話變數 {V}。
> 與設施(docs/40)、寶箱(docs/41)同一套 section 資料;這份補上 NPC 與對話。

## 一、NPC 記錄(section header +0/+2 清單)

清單格式 `{count(1B), records...}`,每筆 **7 byte**(docs/34/35 §四):

| byte | 欄位 | 說明 |
|---|---|---|
| 0,1 | X, Y | 地圖格座標 |
| 2 | sprite 群 id | BLS sprite(載入去重)|
| 3 | **互動子型 / 移動控制** | bits0-1=朝向、bit2=移動開關、**bits3-5 `0x38`=互動子型**、bit7=凍結 |
| 4 | **dlg / 設施索引** | 對話 NPC=對話 id;設施 NPC=設施表索引(docs/40)|
| 5 | flags | 低 3 bit=**可見性旗標**(劇情前/後 NPC 變體)|
| 6 | b6 | 執行期覆寫(站立底 tile)|

## 二、互動分流(dispatcher 0x6291 → 子型 = `(b3>>3)&7`)

```
al=[di+3];  test al,0x38
== 0  → 子型 0:對話    [0x2597]=[di+4];  call talk(0x62e9)
否則  → al=(al&0x38)>>3:
   1 → 對話(同上)
   2 → 特殊 handler 0x62d9
   3..7 → 設施 dispatcher 0x839f(b4=設施索引,docs/40)
```

## 三、對話文字定位

對話 handler 0x62e9:`di=[0x2597](=b4); di+=0xbb8; 印記錄 di`。

- **對話記錄 = b4**(該城所屬 D3TXT bank 的本地 rec;printer 全域定址加 0xbb8)。
- **可見性旗標**(byte5 低 3 bit)決定同一位置的 NPC 變體:阿里阿罕劇情前(flag 0x01,「魔王要消滅世界
  是胡說的吧」)vs 打倒巴拉摩斯後(flag 0x03,「歡迎回到阿里阿罕!你們打倒巴拉摩斯…」),同格不同人。
- **多城共用一個 D3TXTnn**,各佔不同 b4 區段不撞號:D3TXT01 內 阿里阿罕=b4 0-9/79/86-92、雷貝=b4 55-64。

### CTY → 對話檔(D3TXTnn)★ 已靜態全解 — bank 寫在 section header +0x17

對話檔 = **D3TXT0\<bank\>**,`bank` = **每個 section 的 header 偏移 `+0x17` 那個 byte**(就在 CTY 檔裡)。
**純靜態、資料驅動,不需 overlay/DOSBox。** 反組譯鏈(全 file offset):

```
sub_4428(section header 解析,0x44d2):
   di = section_base + 0x15;  bx = [di+2]            ; 讀 section_base + 0x17 起的 word
   [0xb58] = bl                                      ; ★ bank = section_base + 0x17 的 byte
   [0xd71] = bh                                      ; = 地圖 id(section_base + 0x18)
load_cty 尾段(0x4526):
   ax = [0xb58]                                      ; bank
   call 0x2bd5                                        ; 對話文字載入器
0x2bd5:  if [0x2528]==bank: return                    ; 同 bank 不重載
         [0x2528]=bank;  div 10 → 數字;  poke 進 DGROUP "d3txt00.txt" 模板的數字位 [0x79]/[0x7a]
         int 21 開 d3txtNN.txt → 載入對話文字段
```

> 檔名模板是**小寫**存在 DGROUP(`cty00.dat`、`d3txt00.fon`、`d3txt00.txt`@0x74);先前搜大寫 "D3TXT"
> 沒命中,誤判「在 overlay 封死」。實際 bank 就在 CTY section header,直接讀。

**每 section 各自有 bank**(多 section 大城可跨檔:CTY12 sec=2/3/3/2、CTY43=5/5/2)。
全 89 城讀出 bank 皆落 **1-9**。驗證(逐句內容對上):

| CTY | bank → 檔 | 內容佐證 |
|---|---|---|
| 0 阿里阿罕 | 1 → D3TXT01 | 「被海包圍的大陸」「歐里狄加的後代」「北去雷貝鎮」 |
| 1 雷貝 | 1 → D3TXT01 | 「歡迎到雷貝鎮」「盜賊的鑰匙…拿吉米之塔」 |
| 2 羅馬利亞 | 2 → D3TXT02 | 「拯救遙遠北方的諾阿尼魯」「新的國王誕生了!{名}王萬歲」 |
| 17 達瑪神殿 | 6 → D3TXT06 | 「我想轉成武鬥家」「職業經驗不夠是無法轉職的」(轉職神殿)|

> 補充:**file→地名 區域指紋**(掃各檔出現的地名 rec 495-514)可交叉佐證 bank 分組,但 bank 本身
> 已由 section header +0x17 直接定出,不需推定。

## 四、對話變數 {V}(控制碼)

記錄是 2-byte word 串;`>=0xffed` 為控制碼,部分後接 **+1 word 參數**(須一併消耗):

| 碼 | 語意 | 參數 | remake 替換 |
|---|---|---|---|
| 0xfffe/0xfffd | 換行 | — | 換行 |
| 0xfffc | 捲動 + 等鍵 | — | 分頁 |
| 0xfff5 VAR0 | 變數子字串(城鎮對話最常見)| +1 | **主角/受話者名** |
| 0xfffb VAR_ENT | 實體/隊員名 | +1 | 主角/受話者名 |
| 0xfff9 VAR_NUM | 數值(金額)| +1 | 十進位數字 |
| 0xfffa VAR_ITEM | 道具名 | +1 | (略;城鎮對話罕見)|
| 0xffed/0xfff6 | 索引字串 / variant | +1 | 主角名(後備)|

> ★ 原 remake renderer 把控制碼當空白但**未消耗 +1 參數** → 參數被誤畫成字模(亂碼)。已修正:
> `dq3_text_draw_record` 消耗參數並替換實字;`dq3_text_set_var_name`(主角名)/`dq3_text_set_var_num`
> 設定 context,`main.c` 在創角/讀檔/酒場後以隊長名更新。例:阿里阿罕 b4=0x4f 由「{名},怎麼了?」
> 渲染為「<主角名>,怎麼了?」。

## 五、逐句 mapping(已驗證:D3TXT01)

完整全城 NPC 記錄(位置/sprite/子型/dlg/可見旗標)在 **`docs/data/npc_dialogue.json`**
(`tools/gen_npc_dialogue.py` 產;71 城、768 NPC、對話型 643)。早期區逐句樣本:

### CTY00 阿里阿罕(節選)

| 位置 | dlg | 可見 | 對話 |
|---|---|---|---|
| (37,16) | 0x00 | 前 | 阿里阿罕是個被海所包圍的大陸,據說在海的對岸有更大的陸地存在著。 |
| (28,35) | 0x07 | 前 | 你是那個勇者歐里狄加的後代嗎?你父親是一個多麼棒的人啊! |
| (2,28) | 0x09 | 前 | 這裡是阿里阿罕,北去的話就是雷貝鎮。 |
| (8,37) | 0x4f | 前 | {名},怎麼了?趕快去晉見國王吧。 |
| (2,28) | 0x56 | **後** | 歡迎回到阿里阿罕!你們打倒巴拉摩斯的傳聞已經流傳到這裡了喔。 |
| (6,1) | 0x5d | **後** | 我從海的彼岸,活著‥‥回‥來了‥。 |

### CTY01 雷貝(節選)

| 位置 | dlg | 對話 |
|---|---|---|
| (3,9) | 0x39 | 歡迎到雷貝鎮。 |
| (17,4) | 0x3a | 有盜賊的鑰匙了嗎?太好了。這個村莊的南方,也有通往拿吉米之塔的通道呢。 |
| (4,19) | 0x37 | 往東方去,越過山丘的話,會有一處小泉水。 |

## 待 RE / 待補

- ~~CTY→D3TXTnn 表~~ **已解(section header +0x17,見上)**。
- ~~remake 城鎮「話す」接 byte4 → 對話記錄~~ **已接(見下)**。
- 特殊子型 2(handler 0x62d9)語意待解。

## remake 落地(話す NPC 互動)

- `dq3_scene.dlg_bank` = section header +0x17(`dq3_town_load` 解析);主迴圈每幀依當前 section bank
  以 `dq3_dialogue_set_bank`(→ `dq3_text_reload`,只換文字檔保留字型)切換對話檔。
- 面向 NPC 按 Enter:`dq3_scene_npc_at(faceX,faceY)` 找 NPC,子型 `(ctrl>>3)&7 < 3`(對話型)→
  `dq3_dialogue_open(byte4)` + `set_dialogue_hero`({V}=主角名);設施型(≥3)略過(另由商店/旅社入口處理)。
- 驗證:`dq3_dialogue_test` —— CTY00 +0x17==1、talk NPC b4=0x4f → D3TXT01 非空(「{名},怎麼了?…」)、
  `set_bank` reload D3TXT02、越界拒絕,全通過。
- 待補:系統訊息(「空的」rec 0xf3、「行李滿」0x13a)與道具名(rec=code+1)語意上屬 D3TXT00,
  目前共用 section bank 檔(latent;原版 D3TXT00 常駐),不影響 NPC 對話正確性。
