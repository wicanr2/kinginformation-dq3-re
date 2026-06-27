# 經驗教訓(Lessons Learned)— 反組譯 / remake 接線 / 人機協作

> 本檔整理精訊 DQ3 反組譯 + SDL2 重製專案在「多輪協作」中**反覆踩到的雷、根因、與更正後的紀律**。
> 與 [`docs/00-re-methodology.md`](00-re-methodology.md)(可重用 RE *技術*)互補:00 講「怎麼 RE」,
> 本檔講「**怎麼工作才不繞遠路、不下錯結論**」。每條 = 踩雷 → 第一性原理根因 → 鐵則 → 本專案實例。
> 跨 session 版見 `~/.claude` memory(`dq3-worklist-stale-loop`、`dq3-no-dosbox-debugger` 等)。

---

## 1. 別鬼打牆:已完成的工作別當「未完成」重查

**踩雷**:反覆把前幾輪「早就做完」的工作當成待辦又去重新調查——手動選咒選單、注音↔英數切換、
packbg 戰鬥背景、設定選單 UI、scripted warp、怪物施狀態…每次都「發現」其實已完成。純浪費。

**根因(第一性原理)**:WORKLIST/markdown 累積了「做完卻沒標 `[x]`」的項目;我**信任 marker 而非
ground truth(code)**。dated 盤點文件(`docs/47`、`walkthrough-flow-audit`)更會隨工作完成而整批變錯。

**鐵則**:
- 動手調查任何「剩餘/未完成」項**前,先 `grep` 實際 code 確認它是否已實作**。`[ ]`/「未接」不可信。
- 驗證完是已完成 → **當下立刻**標 `[x]` + 收進「已完成」摘要,別留給下一輪重踩。
- **WORKLIST / code 是唯一真相**;dated 盤點(audit)只是歷史快照,核實一律回 code。

**實例**:本專案一輪內就清掉 6+ 個 stale `[ ]`;walkthrough-flow-audit 的「❌ 沒串起來」清單對 code
核實後幾乎全錯(boss/道具鏈早已 ✅),整段重寫。

---

## 2. Ground truth = code + 使用者(親自玩過破關)

**踩雷**:① 依不完整的 `dq3_treasures` 抽取表判「盜賊鑰匙斷鏈」(實為攻略裡老人給);
② 下「原版無自動晝夜」結論(使用者:**走路就會變夜**);③ 標「scripted warp 未接線」(使用者:**已接**)。

**根因**:把「我靜態掃不到」當「不存在」;把自己的推論凌駕於玩過破關的使用者之上。

**鐵則**:
- 使用者玩過破關 = **第一手 ground truth**;使用者指正時**先採信並回頭重查**,別替錯結論辯護。
- 內容缺口預設是「**remake 還沒 RE/接**」,**不是**「原版沒有 / build 沒做」(精訊中文化是做完的)。
- 查道具/事件來源:**先翻攻略**(references/ + BBS),再靜態掃。

---

## 3. 第一性原理:靜態反追溯源優先,不退回動態 / DOSBox

**踩雷**:撞牆就想下「封死 / 要動態 / 需 DOSBox debugger 看 crash 位址」。**此環境根本沒有 DOSBox**。

**根因**:「在 overlay / 是 runner 資料驅動 / 看不出來」是 thought-terminating cliché;多數「值/機制
從哪來」純靜態就能答,且不深。

**鐵則**:
- 「這值/設定/行為從哪來」→ 一律**先靜態反追**:錨定已知實例 → 找 sink(open/print/查表/扣血)→
  運算元 register-by-register 反走到一張表/一個 header 欄位/一個常數。搜尋落空 = 換 query,非不存在。
- **不提議 DOSBox debugger / 逐畫面 oracle 當解法或下一步**;真 runtime-only(crash 在哪條指令 fault)
  且靜態窮盡仍無解 → 標「此環境不驗 / 用使用者指定值」,不掛在 DOSBox 上等。

**實例**(本專案**純靜態解出**):物理/咒文傷害公式、教會復活費 level 表(DGROUP 0x3c6c)、旅社費
(設施 block +1)、祈禱之戒 ~25% 損壞、overworld portal 0x396e 全分支、怪物施狀態 base 分類。

---

## 4. 柵欄原則(Chesterton's Fence)要用對:反證前先排除「另有存在理由」

**踩雷**:用「黑暗之燈(強制變夜的道具)存在 → 所以夜晚不會自動來」當柵欄反證,**論證錯了**——
黑暗之燈仍有獨立用途(立即/室內強制變夜),它存在**不能**反證自動循環。

**根因**:「某設計存在 ⇒ 某事必不成立」的反證,只有在**該設計沒有其他存在理由**時才成立。

**鐵則**:拆/用一道柵欄(既有設計)前,先重建它「為何存在、擋住什麼」;若它有**不只一個**存在理由,
就不能拿它反證另一件事。判它多餘/錯誤前,先確認自己比建造者更懂它為何存在。

---

## 5. 斷言「做了 X / X 成立」前先驗證(grep / dump / diff / 單測)

**踩雷**:工具回顯(bash stdout、Read)可能幻覺假 PASS / 假 commit;憑印象標「某機制做 X」常錯。

**鐵則**:
- 宣告完成**只在** `git diff/log/status` + 檔案中介數字(`PASS=N`)驗證後;bash 回顯最不可靠。
- 描述一段碼做什麼前,**grep 它的呼叫端**(真被執行嗎?方向對嗎?死碼?)。
- 視覺/UI feature:**dump PPM + pixel-diff + 肉眼 read**(本專案:晝夜 phase0 vs phase2 diff 494437、
  設定/裝備/寶箱 overlay 畫面 dump 確認),不靠「應該會動」。

---

## 6. 改動「已測試系統」:用「可選參數 + NULL=現行」做零回歸

**踩雷顧慮**:把確定性升級系統改成忠實 rng 成長,會動到 #4/#5/#6 既有單測 → 回歸風險。

**鐵則**:新增 `_rng` 變體(rng 參數),舊函式 = 呼叫 `_rng(…, NULL)`;設計成 **`rng==NULL` 完全重現
現行確定性行為** → 既有測試零修改全綠,遊戲路徑傳真 rng 即得忠實隨機。新行為另加單測(rng≤上限/
可重現/有落差/NULL=確定性)。

**實例**:`dq3_member_init_rng`/`gain_exp_rng`(RE sub_d9cc:`+= rng(0..(target−當前))`),game_tester 79/79 不變。

---

## 7. dated 盤點 / audit 文件會過期:完成工作後回頭更正,別讓它誤導未來

**踩雷**:`docs/47`、`walkthrough-flow-audit`、`boss-trigger-points` 等盤點記錄「當時狀態」,工作完成卻
沒回頭改 → 後來把已完成項當缺口、把已定案的身分當「未定」(CTY14 被標「巴拉摩斯城候選」實為甘達特巢穴)。

**鐵則**:盤點/audit 文件加「★ 歷史盤點,核實以 code+WORKLIST 為準」警示;**錯誤斷言(說錯、非只 stale)
直接刪除/更正**,不只加註。完成一段工作時,順手把對應 audit 的該項標 ✅。

---

## 8. 記憶衛生:存「教訓」不存「狀態」

**原則**:跨 session memory 放**耐用的方法論 / 流程教訓 / 已驗證事實**(羅塞塔字型對齊、位址基準陷阱、
glyph 234=女、別鬼打牆、別依賴 DOSBox)——這些不會過期。**專案狀態(誰做了沒)放 WORKLIST/code**,
那才是唯一真相、會變動。如此 memory 永遠正確,狀態永遠在一處。

---

## 9. 使用者要求時:研究驅動實作(先查真實資料再決定怎麼做)

**做法**:接「咒文效果」「野外咒文」這類需忠實的 feature,**先查一手史料 + 攻略 + web + RE** 確認真實
機制(精訊 1994 BBS「DQ3 MAGIC LIST」+ Dragon Quest Wiki/Realm of Darkness + `re-log-spell-effect-dispatch`),
**再決定 remake 怎麼對應簡化模型**,並把「研究來源 + 實作對照 + 忠實度標註」寫成文件(`spell-effects-research.md`)。
不憑記憶臆造數值;確切值未 RE 到就用 classic 值 + 註明出處。

---

## 10. 自建工具補原版缺口,而非一次性 hack

**做法**:原版字庫缺 UI 詞的字(「設」「版」)→ 建**可重用 rasterize pipeline**
(`tools/gen_custom_glyphs.py`:CJK 字型 → 原版 16×16 點陣格式 + 自包含 draw 函式;
`dq3_text_draw_glyph` 對 idx≥0xf000 fallback),而非硬塞單一 bitmap。下次缺字一行搞定。

---

> **一句話總結**:**code 是唯一真相**,使用者(破關)是 ground truth,撞牆先靜態反追(別等 DOSBox),
> 斷言前先驗證,改測試過的系統用「NULL=現行」零回歸,盤點文件會過期要回頭改,記憶只存教訓。
