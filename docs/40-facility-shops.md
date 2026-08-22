# 設施系統（商店／旅社／教會／記錄點）— facility block 與 dispatcher 靜態 RE

> 把 facility block 格式與共用 dispatcher 從 EXE 追到 `CTY*.DAT`，並萃取商店品項。
> 結論:**商店貨架不靠猜、不靠攻略佐證,而是逐城寫死在 CTY 地圖資料裡**,由一個共用的
> 設施 dispatcher 依 NPC 的 byte4 索引出來。raw block 仍須有同 section NPC consumer 才是
> 玩家可用設施；不能把 block inventory 擴張成「所有設施均已閉合」。

## 一、設施 dispatcher(file 0x839f)

玩家對店員 NPC「話す/調べる」→ 走到設施 dispatcher。`di` 進來時 = 該 NPC 的記錄指標。

```
0x839f:  bl = [di+4]              ; NPC byte4 = 設施索引(docs/35 NPC 記錄第 5 欄「對話/參數 id」)
         bh = 0;  bx <<= 1        ; bx = byte4 * 2
         di = 6 + [0xb24]         ; [0xb24] = 當前 section_base;di → section header +6
         es = [0x2536]            ; seg_cty(載入的 CTYxx.DAT)
         si = es:[di] + [0xb24]   ; section header +6 = 設施指標表(相對 section_base)
         si += bx                 ; + byte4*2 → 表內第 byte4 項
         di = es:[si] + [0xb24]   ; 該項 = 設施 block 偏移(相對 section_base)
         al = es:[di]             ; block[0] = 設施類型
         ah = 0;  ax <<= 1
         si = 0x623 + ax          ; 跳表 base DGROUP 0x623
         call [si]                ; 依類型派發
```

關鍵:**section header `+6` 是一張先前未命名的欄位 = 設施指標表**(docs/35 表只列了
`+0/+2` NPC、`+8` examine、`+0xc` 轉場、`+0xe` layout;`+6` 補上)。設施表項與 block
都是**相對 section_base 的偏移**(故每次讀完都 `+[0xb24]`)。

## 二、設施跳表(DGROUP 0x623,5 種)

| type | handler(file)| 設施 | 歡迎詞 rec(D3TXT00)|
|---|---|---|---|
| 0 | 0x86f5 | **旅社**(inn)| 0x11a `歡迎光臨,這裡是旅社,` |
| 1 | 0x8814 | **武器/防具店** | 0x118 `歡迎光臨,` → 0x11b/0x11c 買/賣 |
| 2 | 0x8adf | **道具店** | 0x119 `歡迎光臨,這裡是道具店` |
| 3 | 0x83d8 | **教會**(神官)| 0x129 `我是主持正義之人。` |
| 4 | 0x2719 | **記錄點**(冒險之書)| 0xfd `你們要把旅行的經驗` / 0xfa `選擇一組記憶吧` |

## 三、設施 block 格式(在 CTY 資料內)

```
facility_block:
  +0  type           ; 0..4(跳表索引)
  ── type 1(武防店)/ 2(道具店)──
  +1  count          ; 品項數
  +2.. item_id[count]; 道具碼(= ITEM.DAT index;名稱 = D3TXT00 rec (id+1))
  ── type 0(旅社)──
  +1  cost(u16)      ; 住宿費 raw；handler 0x86f5 乘「目前仍存活的隊員數」
```

賣單 builder(file 0x8848)的對應反組譯:`es=[0x2536]; cl=es:[di]`(count)→ 逐 byte
讀 item id 複製到 buffer `[0x662]`,每項價格查 `[item*7 + 0x1f9 + 2]`(**ITEM.DAT 載到
DGROUP 0x1f9**,7 byte/筆,+2/+3 = 價格,對上 docs/22)。買單渲染 file 0x8fa1 →
0x9bcf(選單幾何,品項數存 `[0x5077]`、清單指標存 `[0x3ee8]`)。

## 四、萃取(資料驅動,可重生)

`tools/gen_shop_data.py` 直接讀 `assets_raw/CTY*.DAT`:逐 section 讀 header `+6` 設施表
(表長 = `(首 block 偏移 − 表頭)/2`,blocks 緊接表後)→ 解出每個 block 的 type 與品項,
產出 `dq3_remake/src/dq3_shopdata.c`(+`.h`)。**全 89 個 CTY 共 91 設施、46 商店**。

驗證(三條獨立佐證,皆自洽):
- **CTY00 阿里阿罕**最弱(檜木棒/木棒/銅劍/布衣…)、`0x1e 布的衣服`對上 docs/22 主角開場初始裝備。
- **CTY38 道具店**含 `5d 隱身草`；杜勝利攻略另記「朗錫爾道具店隱身草」，但攻略地名
  不能反推 CTY 編號。production 貨架以 CTY raw 資料為準，朗錫爾 CTY47／75 是另一組入口。
- **王者之劍(35000)不在任何商店** → 對上攻略「瑪依拉村二樓寶物」,非商品。CTY86 出現破壞之劍(終盤)。

## 五、全鎮商店品項表

| CTY | 設施(計) | 武防店品項 | 道具店品項 |
|---|---|---|---|
| 0 | 旅社/武防店/道具店/教會 | 檜木棒、木棒、銅劍、布的衣服、旅人的衣服、皮甲胄、皮盾 | 藥草、驅毒草、蓋美拉翅膀 |
| 1 | 旅社/武防店/道具店/教會 | 聖刀、有刺的鞭子、鎖鍊刀、功夫裝、皮甲胄、皮帽子、皮盾 | 纏頭巾、藥草、驅毒草、聖水、蓋美拉翅膀 |
| 2 | 旅社/武防店/道具店/教會 | 鎖鍊刀、鐵槍、鋼劍、皮甲胄、龜殼甲胄、鎖子甲、青銅盾 | 鐵的圍裙、皮帽子、藥草、驅毒草、聖水、蓋美拉翅膀、滿月草 |
| 3 | 旅社/武防店/道具店/教會 | 鐵槍、鋼劍、鐵爪、鐵甲、武鬥裝、青銅盾、鐵盾 | 鐵的圍裙、藥草、驅毒草、聖水、蓋美拉翅膀、滿月草 |
| 4 | 旅社/道具店 | — | 魔導士之杖、防禦衣、聖水、蓋美拉翅膀、滿月草、蜘蛛絲 |
| 6 | 旅社/武防店×3/道具店×2/教會 | 鎖子甲、鐵甲、青銅盾、鐵盾、鐵的圍裙、危險的泳衣(另兩攤見資料)| 驅毒草、滿月草(另一攤全藥品)|
| 12 | 旅社/武防店/道具店/教會 | 鋼劍、鐵斧、大剪刀、鐵甲、防禦衣、鋼甲、鐵盾 | 皮帽子、藥草、驅毒草、聖水、蓋美拉翅膀、滿月草 |
| 15 | 旅社/道具店/教會 | — | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草、蜘蛛絲 |
| 16 | 旅社/道具店 | — | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草 |
| 20 | 旅社/武防店/教會 | 理力之杖、大鐵鎚、屠僵屍劍、魔法的鎧甲、魔法法衣、青銅盾、鐵面具 | — |
| 22 | 旅社×2/武防店×2/道具店×2/教會×2 | 審判之杖、大鐵鎚、鋼甲、防禦衣、武鬥裝、鐵面具 | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草、毒蛾粉 |
| 32 | 旅社/武防店/道具店/教會 | 檜木棒、木棒、銅劍、布的衣服、旅人的衣服、皮甲胄、皮盾 | 藥草、驅毒草、蓋美拉翅膀 |
| 38 | 旅社/武防店/道具店 | 檜木棒、木棒、毒針、有刺的鞭子、布的衣服、華麗的衣服、功夫裝 | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草、毒蛾粉、隱身草 |
| 43 | 旅社/武防店/道具店；另有無 NPC consumer 的教會 block | 鐵爪、審判之杖、屠僵屍劍、屠龍劍、魔法的鎧甲、鏡盾、魔導士之杖 | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草、毒蛾粉 |
| 47 | 旅社/武防店/道具店 | 聖刀、鐵槍、鐵斧、大鐵鎚、防禦衣、魔法法衣、鐵面具 | 聖水、蓋美拉翅膀、隱身草 |
| 59 | 道具店 | — | 藥草、驅毒草 |
| 60 | 旅社/道具店 | — | 藥草、聖水、蓋美拉翅膀、皮帽子 |
| 61 | 旅社/武防店/道具店 | 屠僵屍劍、屠龍劍、審判之杖、魔法的鎧甲、魔法法衣、龍鱗甲、鏡盾 | 藥草、聖水、蓋美拉翅膀、毒蛾粉、蜘蛛絲、隱身草 |
| 75 | 旅社/武防店/道具店 | 聖刀、鐵槍、鐵斧、大鐵鎚、防禦衣、魔法法衣、鐵面具 | 聖水、蓋美拉翅膀、隱身草 |
| 79 | 旅社/武防店/道具店/教會 | 魔導士之杖、審判之杖、屠龍劍、飛鷹劍、魔法法衣、龍鱗甲、天使的衣袍 | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草、隱身草 |
| 81 | 旅社/武防店/道具店 | 鋼劍、大剪刀、龍鱗甲、水之羽衣、鏡盾、伊歐那順 | 藥草、聖水、蓋美拉翅膀、滿月草 |
| 83 | 旅社/武防店/道具店 | 屠僵屍劍、屠龍劍、審判之杖、魔法的鎧甲、魔法法衣、龍鱗甲、鏡盾 | 藥草、聖水、蓋美拉翅膀、毒蛾粉、蜘蛛絲、隱身草 |
| 84 | 旅社/武防店/道具店 | 屠僵屍劍、審判之杖、魔法法衣、龍鱗甲、力盾、華麗的衣服 | 藥草、驅毒草、聖水、蓋美拉翅膀、滿月草、毒蛾粉 |
| 86 | 旅社/武防店/教會 | 飛鷹劍、破壞之劍、防禦衣、魔法的鎧甲、鏡盾、鐵面具 | — |

(只列有商店的城;CTY8/45/46/49/51/85 等只有旅社或教會,無商品。完整逐攤資料見
`dq3_remake/src/dq3_shopdata.c`,含 `sec`/`k`/旅社住宿費。)

> 2026-08-22 訂正：上表的「設施」原先只枚舉 section facility blocks，會把孤立 block
> 誤寫成玩家可用設施。IDA `sub_1702F` 證明還必須有 facility subtype NPC 以 `byte4`
> 索引該 block。CTY43 section0 雖有 index2 type3 教會 block，但日／夜 NPC 表都沒有
> facility consumer 指向 index2；`(12,4)` 的 `byte4=2` NPC 是 subtype0 對話者。因此
> CTY43 教會不可用，完整 raw records 與 hash 見 `docs/140`。

## 六、remake 落地

- `dq3_shopdata.{c,h}`:`dq3_facility` 陣列(cty/sec/k/type/count/item_off/inn_cost)+
  品項池 `dq3_shop_itempool`;`dq3_shop_items(cty, type, &items)` 取某城某類商店品項。
- `dq3_facility_at(cty, sec, k)`:依 (CTY, section, NPC byte4=設施索引) 精確查設施。
- **走到店員 NPC → 開該攤(已接,docs/42)**:Enter 面向設施型 NPC(子型 `(ctrl>>3)&7 >= 3`)→
  `dq3_facility_at(cur_cty, cur->section, byte4)` → 依 type:武器/道具店開 `shop_modal`(該攤精確品項
  `&dq3_shop_itempool[item_off]`,非合併)、旅社/教會/記錄點開對應 D3TXT00 歡迎詞、記錄點另 autosave。
  `dq3_scene.section` 新增(town_load 設)供查表。B 鍵商店捷徑保留為合併版 fallback。
- `dq3_dialogue_test` 驗:CTY00 sec0 設施 NPC byte4 → `dq3_facility_at` 全命中、k1=武防店(7 品項)。

## 教會服務(handler 0x83d8,3 選單)

- option1 `0x849b` = **解毒**(rec 0x136「誰要解毒」)
- option2 `0x853c` = **解詛咒**(rec 0x137「誰要解詛咒」)
- option3 `0x85ff` = **復活**(rec 0x138「誰需要復活」;測實體 `[+0x38]&0x80` 死亡旗標)
- 復活 handler 讀角色 level `+0x15`，夾至 40，再查 DGROUP `0x3c6c`／file
  `0x19dac` 的 40-word 表；不是舊 C remake 的 `level×10`。
- Lv1..40 費用：
  `10,10,10,20,30,40,50,70,90,110,130,150,170,200,230,260,290,330,370,410,`
  `450,490,530,580,630,680,730,790,850,910,970,1030,1090,1160,1230,1300,`
  `1370,1450,1530,1610`。
- file `0x8689..0x8696` 在付款後清死亡 bit，並把 current HP/MP 寫回 max HP/MP。

## 現行 remake 狀態與待補

- Ebiten 旅店已依原版閉合：file `0x86f5` 呼叫 `0xd966` 計算存活人數，`0x8705`
  乘 raw 單價，付款後 `0x87cb` 逐員把 `+0x16/+0x18` 寫成 `+0x2a/+0x2c`
  （HP/MP 上限），並跳過死亡旗標 `+0x38 & 0x80`。金幣不足不扣款也不恢復。
- Ebiten `facChurch` 已開正式三選單 modal；復活會選隊員、顯示 rec301 確認、按完整
  40-level 表扣款，成功後 HP/MP 回滿。存活者、取消與金錢不足均不扣款。
- 復活費已移至 versioned `dq3_cht` JSON game pack；測試直接逐 word 比對
  `DQ3.EXE` file `0x19dac`。欄位契約見 `docs/84-game-pack-json-contract.md`。
- 新遊戲 production trace 已在阿里阿罕低危區實際發生陣亡後，經設施 NPC、教會 modal、
  旅店及正式戰鬥輸入練到 Lv4，再繼續抵達羅馬利亞。
- 解毒已由 pack condition、battle／角色／save、移動 consumer 與教會固定 5G transaction
  接成 E2。解詛咒不是 `+0x38` flag，而是掃描角色 `+0x3a` 起八個物品 word 的
  `bit0x4000`，費用為 `level×100`，付款後把命中 word 寫成 `0x00ff`。後續已由
  `sub_17ED9` 證實 ITEM `+4/+5 & 0x0e00` 是該 bit 的 writer gate；五筆 raw ID、同部位
  換裝拒絕與教會交易已接成 E2。完整鏈見 `docs/147`。
- 同城多攤(CTY6/22 多家武防店):`dq3_facility_at` 已能逐攤(by k),走到不同店員開不同攤。

> 現行 Go 限制：`facilityForCty(cty,k)` 尚未接收 section，雖 raw table 保存 `sec`，lookup
> 目前會忽略它。CTY43 section1 與 CTY22 section1 這類同 CTY 多 section 設施因此尚未做
> 原版 runtime parity；不得把舊 C 的 `dq3_facility_at(cty,sec,k)` 完成度套到 Ebitengine。
