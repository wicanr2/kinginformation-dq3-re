# 72 — RE 做了、兩版 remake 都沒實作的缺口盤點(2026-07-03)

> **修正進度(2026-07-03 當天,見 `docs/73` worklist)**:本盤點的高信心缺口多數當天已修,狀態如下標。
> 讀本檔前先看此表,別把已修項當「還缺」:
>
> | 缺口 | 修正批 | 狀態 |
> |---|---|---|
> | A1 玩家狀態/buff 咒文 | W3 | **進行中**(ebitan)|
> | A2 文字插值 | W1 | ✅ ebitan(commit 52ce817)|
> | A3 怪物群隻數 | W2 | ✅ ebitan(commit 0269012)|
> | A4 酒場兩段式招募 | W4 | ✅ ebitan(commit 5507a0a)|
> | D. C oracle NPC 三 bug | W5 | ✅ C(commit f27b51d)|
> | B1 沙曼歐莎怪力魔 / C1 EBG / C2 0xd4f | — | 延後(需 RE / 低信心)|
>
> 註:A1-A4 的修正目前只在 **ebitan**(可玩主目標);C 版同缺項除 D 段(W5)外未回補,若要拿 C 對拍需另補。

> 3 個 subagent 分區(戰鬥/怪物、地圖/事件/NPC、文字/UI/音訊)交叉盤點 RE docs vs 兩版 code
> (C `dq3_remake/`、ebitan `dq3_remake_ebitan/`),協調者逐項獨立 grep 核實。只收「**兩版都沒做**」。
> 校準原型 = NPC story-flag 過濾(docs/71,原版 RE 了、兩版都漏)。方法:truth-in-code,不信 docs marker,
> 兩版都實際 grep(negative 證據)才算。docs 描述的「remake 簡化版」多已過期,以 code 為準。

## A. 兩版都缺 · 高信心 · 有玩感影響(建議優先)

### A1 ★ 玩家狀態/buff/debuff 咒文完全不能施放(最高影響)
- **機制**:拉里荷(睡眠)、拜基魯多(攻↑)、史卡拉/史克魯多(守↑)、瑪努莎(幻惑)、瑪荷頓(封咒)、
  美達巴尼(混亂)等 rec143-160。原版 base==0 = 輔助/狀態咒(docs/re-log-spell-effect-dispatch)。
- **依據**:`docs/re-log-spell-effect-dispatch.md`、`docs/data/spell-effects-research.md`、`docs/16`(習得表:勇者 lv16 學拉里荷等)。
- **C**:`dq3_spelldef.c` 全表只有 `DQ3_SK_DMG/HEAL/REVIVE`;rec143-160 不在表 → `pick_caster_spell_kind` 取不到 def → 玩家選單直接沒有。狀態效果只在敵方施咒分支接了,玩家側從未接。
- **ebitan**:`internal/spell/spell.go` Kind 只有 `Dmg/Heal/Revive`,同樣過濾;且 `game/battle.go` **全無** status/PARALYSIS/blind/sealed 欄位——連敵方對玩家下狀態都沒有。
- **影響**:DQ3 標誌性戰術(睡王怪、拜基魯多爆發、史卡拉硬撐)兩版都打不出來;ebitan 額外連中狀態都不會。

### A2 ★ 文字插值:道具名 + 數值渲染成空白(普遍可讀性)
- **機制**:對話控制碼 `VAR_ITEM`(0xfffa 道具名)、`VAR_NUM`(0xfff9 金額/數值)插值。
- **依據**:`docs/03-text-format.md` §控制碼、`docs/12-exe-commands.md`(file 0x1290a.. 各插值分支、`[0x2593]` 存數值)。
- **C**:`dq3_text_set_var_num` 有定義**零呼叫端**(dead);`dq3_text.c:134` 明確把 `VAR_ITEM` 排除在名字替換外。只有 VAR_NAME(主角/受話者名)真接了(main.c:399)。
- **ebitan**:`dialogue.go:45/108` 把所有 `v>=0xffed`(含全部插值碼)一律畫空白;`text.go:14` 註解自陳。**零插值邏輯**。
- **影響**:「你得到了{道具}!」「花費了{金額}G」等系統訊息(D3TXT00 rec181-280 一批)兩版全留白。
- 對齊記憶 `re-retro-text-inline-var-codes`(插值碼要消耗參數,否則亂碼/留白)。

### A3 怪物群隻數缺權重預算(遭遇節奏)
- **機制**:原版用戰鬥點數預算(0x26=38)對候選怪依 `spawn_weight`(D3MNS +0x28)`floor(budget/weight)` 定隻數上限 → 同種多隻(史萊姆×6)。
- **依據**:`docs/13` §怪物群生成、`docs/16`。
- **C**:`spawn_weight` 只解析(`dq3_monster.c:47`)不用;實際隻數 = `1 + grnd()%3`(main.c:1313,均勻亂數)。
- **ebitan**:`SpawnWeight` 同樣只解析;`startEncounter`(game.go:1146)**固定挑 1 隻**。
- **影響**:無同種多隻遭遇,弱怪失去數量壓力/刷經驗特性,遭遇節奏偏離原版。

### A4 酒場兩段式招募(登錄所建名冊 ≠ 酒場入隊)
- **機制**:原版 2F 登錄所建角色「僅登錄入名冊」→ 1F 酒場另選「找同伴參加」才入隊(D3TXT00 rec527-550)。
- **依據**:`docs/36-original-tavern-playthrough.md`(92-104)。
- **C**:`dq3_tavern.c:54` 創角當下 `dq3_party_add` 自動入隊(有 roster 畫面可 toggle,但非「創角=僅登錄」語意)。
- **ebitan**:`game.go:578/580` 即建即入隊,滿 3 人 `append(companions[1:],m)` **頂替最舊**(無 roster 狀態)。
- **影響**:無法先建多角色再挑選出隊;ebitan 滿隊自動頂替**可能意外丟失隊友**。

## B. 兩版都缺 · 已在計畫 / 需再 RE

### B1 沙曼歐莎假國王 → 拉之鏡 reveal → 怪力魔 → 變身杖鏈
- **機制**:沙曼歐莎(夜)持拉之鏡(0x61)照假國王現形 → 怪力魔(怪89 HP400)→ 掉變身杖(0x62)。
- **依據**:`docs/re-log-722` Step20(標「待」)、`docs/data/quest-items.md`、`docs/66` 步35b。
- **C**:拉之鏡可撿,但 reveal/怪力魔觸發零命中;變身杖只能靠 debug `boss:89:0x62` token。
- **ebitan**:同樣零命中;`bosstrigger.go` 自陳「怪力魔座標未定案」。
- **影響**:正常遊玩摸不到變身杖 → B 段道具鏈(船員之骨→幽靈船→蓋亞之劍)兩版都卡死。
- **狀態（2026-07-28）**：EXE handler `0x5682..0x5732` 已定位並完成 Ebiten 實作。
  精確 gate 是玩家站 `(14,7)`，不是舊推測的 facing；正式道具 UI、rec97/98、怪89、
  勝敗 flags、0x62、白天與 save/load 均有回歸。剩餘缺口是 campaign 前段自然抵達 CTY44。

## C. 低信心 / 已知 TODO

- **C1 EBG 事件音效 cue**(旅館睡眠/教堂復活/神宮):`docs/61` 已自標「先載著日後接」。C 載入無觸發、ebitan **連載入都沒有**(`grep EBG` 零命中)。中低。
- **C2 攻方狀態傷害減半**(`[bx+0xd4f]==3 → dmg>>=1`,docs/13):兩版 `0xd4f` 零命中;但**原版觸發條件 RE 本身未定**,implement 前需再 RE。低。

## D. 反向發現:C oracle 曾對 NPC 不忠實 → **已由 W5 修正(commit f27b51d)**

盤點時發現「已驗證的 C port」在 NPC 上有三處不忠實(**與校準範例同源但方向相反**:通常 C 是 oracle,這裡 C 落後 ebitan)。**W5 已全部回補**,C 恢復為可靠 NPC oracle:

| 項 | C 發現時 | C 現況(W5 後)| ebitan |
|---|---|---|---|
| b2<4 NPC 丟棄 | `dq3_scene.c:173 if(b2<4)continue` 丟(約 1661 NPC) | ✅ b2<4→entry0 fallback | 已修(fallback entry0)|
| sprite cache 上限 | `n_npc_spr<8`(應 13) | ✅ `npc_spr[13]` | Go map 無上限 |
| story-flag 可見性過濾 | 零實作 | ✅ 獨立 `g_npc_vis[32]` + 過濾(test_npcvis 24→15)| 已修(commit 8b76ac1)|

→ **C 不是可靠的 NPC oracle**;若要拿 C 對拍 NPC,需先回補這三項(否則 C「少 sprite + 過多/過少 NPC」的錯會被當基準)。

## 排除清單(核實兩版都有做 / 誤判,不算缺口)

物理傷害公式(弱攻miss/會心×2)、咒傷 base/2+rng、魔甲抗魔#7b、飛鷹劍雙擊#7a、AI逃跑/施咒機率/bitmask選咒、升級clamp#5、屬性9999clamp#6、NPC日夜雙清單、晝夜步數/palette、戰鬥受傷震動+閃光、道具使用(藥草/驅毒/聖水/滿月/翼)、戰後palette#8(架構免疫)、NPC隨機走動mover、鎖門鑰匙等級、owportal城變體、type-2 examine warp、寶箱一次性旗標、對話4行分頁、注音命名組字、魯拉傳送(C有ebitan缺→非兩版)、`dq3_locwarp`(冗餘資料非dead-code缺口)。

> 遭遇區座標→region→候選怪(docs/39)、VAR_NAME 主角名插值 = **僅 ebitan 缺、C 有** → 不列本表(本表只收兩版都缺)。
