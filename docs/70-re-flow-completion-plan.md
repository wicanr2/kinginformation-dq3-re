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

### R-2 沙曼歐莎怪力魔 boss〔2026-07-28 已實作並驗證事件切片〕

**EXE 精確證據（取代下方 2026-07-04 的推測）**：

- handler：`DQ3.EXE` file `0x5682..0x5732`。
- 入口逐項判定：story flag `0x42` 已設、原版日夜 byte 為夜、CTY `0x2c`（44）、
  section 1、**玩家座標恰為 `(14,7)`**。沒有 facing 判定；舊「面向 `(14,6)`」是錯誤推測。
- 揭露：rec97 → clear42/set10 → 重載 NPC → rec98 → battle config 含 monster `0x59`（89）。
- 戰敗／逃跑：clear10、restore42，可再次使用鏡。
- 勝利：rec99、強制白天、clear21/set22。EXE 沒有 clear10，因此 remake 也保留 flag10。
- `D3MNS.DAT` monster89 `+0x26=0x62`、`+0x27=200`，交叉證實變身杖獎勵；
  拉之鏡 `0x61` 不消耗。

**remake 驗證**：

- `TestMirrorProductionInputTrace` 從合法 CTY44 checkpoint 起，只經 `Game.step`：
  命令窗 → 道具 → 拉之鏡 → rec97/98 → 怪力魔戰鬥 → 正式攻擊輸入 → 勝利交易。
- 另有戰敗回滾、勝利旗標／獎勵、CTY44 sec1／日夜／story bits 存讀檔回歸。
- runtime 圖：`dq3_remake_ebitan/docs/samanosa_mirror_reveal.png`、
  `samanosa_boss_troll.png`。
- 證據等級：事件內操作與跨存檔為 E3；**從前一個主線節點自然走到沙曼歐莎仍屬 campaign
  reachability 缺口**，不可把這個 checkpoint slice 說成全中盤 E3。

<details><summary>2026-07-04 舊 spec（保留作錯誤推測紀錄，不再作實作依據）</summary>

**沙曼歐莎城 = CTY44**(rec71 自報 + `docs/maps/cty/CTY44.png`);**只有 5 個 section sec0-4**
(先前誤寫 13 = ad-hoc parser 掃過界的垃圾;正確口徑=OpenTown/dq3_scene section 邊界檢查)。

**已定案(可實作)**:
- **假國王**:CTY44 **sec1 (14,6)** sprite=15 sub=0 **夜晚限定**(day 表 word0=0xffff 白天無人、night 表 word2 才有;
  同格三筆記錄由 byte5 故事旗標 fl 決定載哪筆:fl=0x42 rec96「憤怒」/ fl=0x22 rec116「打呼」)。對上攻略「夜晚廚房後門→頂層寢室」。
- **真國王**:CTY44 **sec4 (17,31)** sprite=15 sub=0 dlg=92(rec92「怪物奪走變身杖變成我的形體」)。sec3=牢獄(rec90 提示「拉之鏡在南方洞窟」)。
- **文字序列**(txt05 逐字核對):rec96 未持鏡預設 / rec97「鏡中床上躺的是怪物」reveal / rec98「受死吧」→戰鬥 / rec99 戰後「假王消失、救真王、白天來臨」/ rec116 解決後打呼。
- **Boss**:怪89 怪力魔 HP400 攻80 防180 速60 EXP2500 G105 不施咒。**獎勵 0x62 變身杖** + 假王消失 + 真王還原 + **強制切白天**(對齊 [0x526c],rec99 敘事線索)。

**待定/待實測(誠實)**:
- **reveal 觸發的確切 EXE 分支位址未反組譯到**:sub=0 NPC 的 talk dispatcher(0x6291→0x62e9)是純靜態印表機、無條件分支/戰鬥呼叫 → reveal 必是**另一條路徑**,最可能=「持拉之鏡從道具選單『使用』」的**位置相關 item-use**(比照既有 `DQ3_USE_AWAKEN/DRAIN/GAIA/FAIRYFLUTE` 家族)。⚠ 但那家族本身也是「經典機制+攻略驗證落地」非逐位址證實(docs/49 已反證 id 直接 cmp 分支不存在)。→ 實作比照此工程慣例(`DQ3_USE_MIRROR`:CTY44+sec1+面向(14,6)+夜晚+持0x61 → boss89 → 0x62+設後旗標),誠實度同既有家族;要逐位址證實需另撥一輪反組譯 item-use 選單分派。
- **「後」旗標精確 bit index**:fl=0x22 是 flag id,需按 docs/60 公式(byte=id>>3, mask=0x80 ror(id&7))換算 [0x4f70] 的 byte/bit。
- DOSBox 不可靠 → 以上靜態 RE + 資料交叉,未動態驗;實作後用 game_tester 或 host DOSBox 行為驗證。

**實作(下週派 sonnet)**:`DQ3_USE_MIRROR` 位置 item-use(CTY44 sec1 面向(14,6)+isNight✅+持0x61)→ 怪89 boss → 勝給0x62+設後旗標+切白天+真王對話切換。isNight() 已備。
- 模型:RE=Opus(已定 spec);實作=sonnet;行為驗證=game_tester/dump。

</details>

### R-3 不死鳥祠堂 + 六珠 gate + 飛行坐騎〔基礎大批,gate 終盤〕
- **RE 定位**:不死鳥祠堂 CTY(諾阿尼魯北雪白陸地,txt03 + cty_loc 上層 + warp);祭壇 NPC 座標;
  六珠(綠0x66/藍/紅/紫/黃/銀0x6b)現有收集點盤點(treasure.go/scripted 對帳,哪些已可得、哪些缺)。
- **實作**:① 六珠齊全 gate;② 祭壇事件 → 拉米亞復活;③ **飛行坐騎系統**(新 traversal 子系統:
  騎乘態飛越地表、可降落於指定點/城鎮)。這是本計畫最大一塊。
- **驗收**:集齊六珠→祭壇→得坐騎→可飛行降落。
- 模型:RE=Opus;六珠盤點=sonnet;飛行系統移植=sonnet(對 C `dq3_*` 坐騎碼,若有)。
- ⚠ 若 C remake 亦未做飛行坐騎,則以原版 EXE RE + 攻略為 spec 自製(登記 spec 缺口)。

### R-4 巴拉摩斯城 + boss + descend 自然觸發〔需 R-3 坐騎〕
- **RE 定位**:✅ CTY66 外城、CTY65 boss 房 `(8,3)` handler70；formation `4EDF`
  是怪121(`0x79`)一隻。舊「打兩次」是精訊版打不死 bug 的錄影重試，不是規則。
- **實作**:✅ 單戰勝利→remake flag0x213 + CLR 原版 flag0x29→回阿里阿罕 rec98/99→
  CLR flag0x4d→同點 CTY71 切 CTY72→跳坑 transition 到 CTY77→`0xfe` 自然下降。
- **驗收**:✅ production InputState trace 全程無 KeyU；完整實機長流程仍併入 R-5 驗收。
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
| R-1 露依達酒場 | ✅ | ✅ | 正式入口／登錄／招募／正常出城 trace 完成 |
| R-2 沙曼歐莎怪力魔 | ✅ EXE `0x5682..0x5732`：玩家 `(14,7)`、flags42/10/21/22、rec97-99、怪89 | ✅ 正式 item UI、勝敗交易、存讀檔、runtime 圖 | 事件切片完成；campaign 自然可達性待中盤串接 |
| R-3 不死鳥+六珠+飛行 | ⬜ | ⬜ | 待 |
| R-4 巴拉摩斯城+descend | ⬜ | ⬜ | 待(需 R-3)|
| R-5 龍王城+終盤 wire+拔後門 | ⬜ | ⬜ | 待(需 R-4)|
