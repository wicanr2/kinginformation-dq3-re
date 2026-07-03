# RE 流程補完計畫:定位 → 分批實作(拔後門可破關)

> 目標:把 ebitan remake 剩餘的 A 級流程斷點補齊,達成驗收 =「無 debug 鍵,從新遊戲走到破關,對齊原版」。
> 前置盤點見 `docs/65` §剩餘工作真相盤點。剩餘斷點幾乎全卡 **RE 定位**(城/座標),非實作。
> 本計畫每批 = **先 RE 定位(產 spec)→ 驗 spec(截圖/攻略 oracle)→ 實作 → 獨立核實 → commit**。

## oracle 基建(已備,三證交叉)

1. **cty_loc 地理**(docs/32):地表座標 → CTY 號;上/下層 world PNG 已標。
2. **D3TXT 自報地名**(docs/32 §劇本反推):每城 NPC「歡迎到XX」句 → CTY↔地名。關鍵對照:
   - txt03 = 諾阿尼魯/阿莎拉慕/伊席斯/巴哈拉達/劇揚(**巴拉摩斯城在伊席斯南、不死鳥祠堂在諾阿尼魯北**)
   - txt05 = 姆歐魯/**沙曼歐莎**(怪力魔)
3. **DOSBox warp 截圖**(`tools/dosbox_warp_cty.sh <idx>`):patch 起始 CTY → 開機進內裝 → 截圖比對。
4. **已解碼 NPC dump**(`docs/data/npc_dialogue.json` + `tools/list_special_events.py`):每 NPC 的 sec/座標/sprite/dlg。

> 方法紀律(rulebook 62/64):有解碼器 + 已知輸出(攻略/截圖)→ 交叉比對定位座標,不追資料流;
> 搜尋落空≠不存在,換 query。

## 相依關係(決定批次順序)

```
R-1 露依達酒場 ──(獨立,單一已知 CTY00)
R-2 沙曼歐莎怪力魔 ──(獨立中盤;isNight() 已就緒 ✅)
R-3 不死鳥祠堂+六珠+飛行坐騎 ──(基礎;gate 終盤)
        └─> R-4 巴拉摩斯城+descend ──(需坐騎飛抵)
                    └─> R-5 龍王城+終盤入口 wire(拔 Z/U/Enter)──(需 descend)
```

## 批次

### R-1 露依達酒場正式入口 + 移除示範隊〔最易,無跨城 RE〕
- **RE 定位**:CTY00(阿里阿罕)哪個 section = 冒險者登錄所(2F)/酒場;櫃台 NPC 座標。
  oracle:`warp CTY00` 各 section 截圖 + txt00 rec527「露依達的店」/rec550「登錄所」+ npc_dialogue CTY0。
- **實作**:酒場入口 NPC(examine → 開 tavern 招募 UI,tavern.go 已存在)+ 移除 `member.go` 預建示範隊
  (改為酒場逐一創角招募後才有同伴)。
- **驗收**:新遊戲→進酒場門→招募→隊伍成形;無示範隊殘留。
- 模型:RE 定位=Opus+haiku(warp 截圖);實作=sonnet。

### R-2 沙曼歐莎怪力魔 boss〔獨立中盤〕
- **RE 定位**:沙曼歐莎城 CTY 號(txt05 自報 + cty_loc + warp 截圖);假國王 NPC 在頂層寢室 section/座標;
  觸發機制=**夜晚 + 持拉之鏡0x61 + examine/USE 假國王**(釐清 byte4=44 是否即此，或為拉之鏡 USE 事件；
  注意 CTY54 (8,2) byte4=44 已證=雪地草原 transform 老人，**非**此)。
- **實作**:isNight() gate(✅ 已備)+ 持0x61 gate → 怪89(HP400)boss → 勝得變身杖0x62 + 真國王復位。
- **驗收**:夜晚持鏡 examine 假國王 → 開戰 → 勝取變身杖;非夜/無鏡不觸發。
- 模型:RE=Opus;實作=sonnet。

### R-3 不死鳥祠堂 + 六珠 gate + 飛行坐騎〔基礎大批,gate 終盤〕
- **RE 定位**:不死鳥祠堂 CTY(諾阿尼魯北雪白陸地,txt03 + cty_loc 上層 + warp);祭壇 NPC 座標;
  六珠(綠0x66/藍/紅/紫/黃/銀0x6b)現有收集點盤點(treasure.go/scripted 對帳,哪些已可得、哪些缺)。
- **實作**:① 六珠齊全 gate;② 祭壇事件 → 拉米亞復活;③ **飛行坐騎系統**(新 traversal 子系統:
  騎乘態飛越地表、可降落於指定點/城鎮)。這是本計畫最大一塊。
- **驗收**:集齊六珠→祭壇→得坐騎→可飛行降落。
- 模型:RE=Opus;六珠盤點=sonnet;飛行系統移植=sonnet(對 C `dq3_*` 坐騎碼,若有)。
- ⚠ 若 C remake 亦未做飛行坐騎,則以原版 EXE RE + 攻略為 spec 自製(登記 spec 缺口)。

### R-4 巴拉摩斯城 + boss + descend 自然觸發〔需 R-3 坐騎〕
- **RE 定位**:巴拉摩斯城 CTY(伊席斯南,txt03 + cty_loc + warp);boss trigger 座標;城內寶箱
  (魔神之斧0x19、祈禱戒指0x48第二枚);蓋亞那洞穴跳坑位置(巴拉摩斯城右側)。
- **實作**:飛抵巴拉摩斯城 → boss 121(**須打兩次**)→ 勝設 flag 0x213 → 蓋亞那跳坑 examine →
  `descend()` 自然觸發(**拔 game.go:592 KeyU**)。
- **驗收**:騎坐騎飛至→打倒巴拉摩斯兩次→跳坑→自動下降アレフガルド,全程無 KeyU。
- 模型:RE=Opus;實作=sonnet。

### R-5 龍王城 + 終盤自然入口 wire + 拔所有後門〔需 R-4,收尾驗收〕
- **RE 定位**:龍王城 CTY(乘不死鳥飛至,光之珠0x65);索瑪神殿(アレフガルド)入口座標;
  彩虹橋走位→索瑪城 path。
- **實作**:龍王城→光之珠0x65;彩虹橋 walk → 索瑪神殿入口 → `startZomaSeq()` 自然觸發
  (**拔 game.go:597 KeyZ**);拔 `game.go:714 Enter→進城`、`game.go:718 Cancel→出城` 後門。
- **驗收(D 段最終)**:**無任何 debug 鍵,新遊戲一路到破索瑪** + 使用者實測確認(docs/66 全 56 步)。
- 模型:RE=Opus;實作=sonnet;最終驗收=Opus 打包 build + 使用者實測。

## 每批共通紀律

- **先定位再實作**:RE 未定案(城/座標)前不派實作,避免邊做邊猜燒 token(rulebook 65)。
- **獨立核實才收**:sonnet 交回 → Opus 重跑完整套件(game+internal 全綠)+ 視覺 dump 肉眼比對 + 抽查關鍵邏輯(rulebook 45/65)。
- **debug 鍵當待移除清單**:R-4/R-5 逐一拔(KeyU/KeyZ/Enter/Cancel);驗收時全拔重跑。
- **完整性 > 投報(HARD)**:不以低投報跳步;卡關換 oracle(warp 截圖 / 原版錄影 / 攻略)。
- **版權**:warp 截圖 / 影格不入公開版控。

## 進度

| 批 | RE 定位 | 實作 | 狀態 |
|---|---|---|---|
| R-1 露依達酒場 | ⬜ | ⬜ | 待 |
| R-2 沙曼歐莎怪力魔 | ⬜ | ⬜ | 待(isNight ✅)|
| R-3 不死鳥+六珠+飛行 | ⬜ | ⬜ | 待 |
| R-4 巴拉摩斯城+descend | ⬜ | ⬜ | 待(需 R-3)|
| R-5 龍王城+終盤 wire+拔後門 | ⬜ | ⬜ | 待(需 R-4)|
