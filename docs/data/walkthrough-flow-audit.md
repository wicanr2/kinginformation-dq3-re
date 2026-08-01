# 攻略全流程 × remake 接線盤點(code 核實)

> 來源:**杜勝利** DOS 版 DQ3 攻略(56 步詳細流程,`references/walkthroughs/dusheng_dq3.txt`(已收錄,註明出處))+
> **青衫詩客(小邱)/ 孔方兄** 1994–95 BBS 條列攻略(`docs/history/dq3-bbs-1994.md`,聚焦當機點 + 改檔)。
> 著作權歸原作者,此處僅摘要對照。**現況一律以 code 為準**(rulebook 63:marker 不可信,grep code 核實;
> 本檔 2026-06-27 全面對 `dq3_remake/src/` 重驗)。

## 接線機制總覽(先懂這個再看缺口)

remake 的劇情接線分兩種,缺口只可能出在「需顯式接線」那種:

1. **資料驅動(已涵蓋大部分)**:
   - `dq3_treasure.c`(**121 筆**,自 CTY*.DAT examine 事件表萃取):寶箱 + 「調查地面得物」隱藏物。
   - `dq3_sub2.c` / `dq3_scripted.c`:NPC 給物 / 檢查 / transform(byte4 跳表)。
   - 商店(`dq3_shopdata.c`,真實 ITEM.DAT)、門 tier(盜賊/魔法/最終鑰匙)。
   → 只要玩家走到該地圖、該筆資料存在,**自動生效**,不需逐項手接。
2. **顯式接線(main.c 條件事件)**:里程碑、boss 劇情、puzzle、跨道具 gate —— **缺口集中在此**。

## ✅ 已接且 code 核實(game_tester 80/80)

| 範疇 | 內容 | code 位置 |
|---|---|---|
| 主線里程碑 | START→盜賊鑰匙→魔法玉→羅馬利亞→達瑪→船→彩虹→下降→索瑪 | `dq3_progress.c` |
| 地點事件 | 覺醒粉解諾阿尼魯催眠(CTY4)/ 提頓牢房綠寶珠夜 gated(CTY20)/ 精靈祠堂雲雨之杖(CTY92)/ 八頭大蛇(CTY19)/ 甘達特巢穴(CTY14)/ 黑胡椒救人 gate(CTY15)/ 金皇冠 gate(CTY10)/ 不死鳥六珠(**CTY70**；舊 CTY82 誤判已由 EXE+影片更正) | `main.c`/Ebiten `phoenix.go` |
| **賢者轉職 gate** | 達瑪轉賢者需《領悟之書》(遊人免);Ch21 加爾那之塔 ⚠**現用碼 0x40 錯,應 0x4a,見 A′** | `main.c:2321` |
| A boss | 巴拉摩斯本體 + 索瑪前序列(怨靈122→殭屍123→索瑪124)+ 甘達特 + 八頭大蛇 + 五頭龍/歐里狄加 | `main.c` / boss token |
| 道具鏈(寶箱表) | 領悟之書 0x4a・勇者之盾 0x40・祈禱戒指 0x48・流星手環 0x49・雷神之杖 0x0a・世界樹之葉 0x4c・回音之笛 0x60・拉之鏡 0x61・愛的回憶 0x64・銀豎琴 0x71・妖精之笛 0x77・太陽之石 0x72・藍寶珠 0x67・紅寶珠 0x68・生命戒指 0x70… | `dq3_treasure.c` |
| 道具鏈(事件) | 夢幻紅寶石→覺醒粉・蓋亞之劍開火山・乾渴壺吸海(USE)・變身杖(怪力魔)→船員之骨・草薙之劍(八頭大蛇)・彩虹水滴合成+架橋・六珠→拉米亞 | `main.c`/`dq3_sub2.c` |

## ❌ 真正缺口(code 0 命中核實,2026-06-27)

> 下列在攻略有、但 `dq3_remake/src/` grep **零命中**或明確標「待」。多屬「需顯式接線的 scripted 事件」,
> **非可玩性阻塞**(主線里程碑流仍可破關),屬忠實度長尾。

> ⚠ **道具碼勘誤(2026-06-27,從 D3TXT00 名表 offset+1 解碼,對 ~9 個 gameplay/render 實證碼驗證)**:
> 舊 `re-log-722` 在 0x40–0x4a 範圍**標籤系統性位移**。正確碼:**雷神之杖 0x0a、勇者之盾 0x40、
> 祈禱戒指 0x48、流星手環 0x49、領悟之書 0x4a、魔神之斧 0x19、蜘蛛絲 0x46**。下方已用正確碼重盤。

1. ~~雷神之杖未接~~ → **修正:雷神之杖 0x0a 已在 treasure 表(CTY38)**。原誤判因搜中文字串「雷神」
   (不在 .c)0 命中 + 用錯碼;data-driven 已放置,**非缺口**。
2. ~~勇者之盾 0x46 不在表~~ → **修正:勇者之盾真碼 0x40,已在 treasure 表(CTY87)**。**非缺口**。
3. 乾渴壺 `0x5e` 已由 CTY76 正常推石流程取得；現行 Go trace 以 119 步正式方向輸入
   完成 handler30、passage、event0 調查與 save/load，見 `docs/97`。下一個 blocker 是
   海中淺灘使用乾渴壺與最終鑰匙取得。
4. ~~王者之劍商店解鎖未接~~ → **修正:已接**(main.c:1247 瑪依拉 CTY81 2F 道具店主)。實作為 transform NPC:
   歐里空金屬 0x6d + 淨 12500G(賣22500−買35000)→ 王者之劍 0x1c。實機驗證通過、game_tester 已測。**非缺口**。

### A′. ★賢者轉職 gate 比對錯碼 — ✅ 已修(2026-06-27)
- bug:`main.c:2321/2329` 賢者 gate 檢查/消耗 **0x40(實為勇者之盾)**,領悟之書真碼 **0x4a**。
- **攻略反證(無需 disasm,gate 為 remake 新增之 Ch21 要求,非重現原版碼)**:treasure 表 CTY18 = 0x4a 領悟之書
  + 0x50 力量種子 + 0x36 鐵頭盔 = 杜 Ch21 加爾那之塔三寶;CTY87 = 0x40 勇者之盾 + 0x4f 生命果實 = 杜 Ch48 洛特洞窟。
  → **CTY18=加爾那之塔(領悟之書 0x4a)、CTY87=洛特洞窟(勇者之盾 0x40)**,碼修後攻略全自洽。
- **修法**:gate 0x40→0x4a(檢查 + 消耗);game_tester 改測 CTY18 0x4a 領悟之書 + CTY87 0x40 勇者之盾(81/81)。
  直接驗證:持 0x4a→轉賢者成功+消耗書;無 0x4a→拒;0x40 不再被誤接受。

### B. 過場 / 解謎事件未接
5. **諾魯特密道** — ✅ **已接(2026-06-27)**(杜 Ch16-17):★又一碼勘誤 **0x5b = 國王的信(非船;
   item 表無「船」=vehicle ship.owned)**。諾魯特 gate(CTY62/63 byte4=50,handler 0x5daf,檢查 0x5b)早已在
   dq3_scripted。舊 C remake 曾以自造 flag `0x215` 補授信；2026-07-29 重查原版 handler26
   已證實真正 gate 是 story flag `0x37`：首次 rec24 關閉後授《國王的信》0x5b、clear
   `0x37`、顯示 rec25。Go/Ebitengine 已以 game-pack primitive 與正式 production trace
   閉合授信。2026-07-29 再以 IDA、CTY62、原版影片校正：諾魯德不是自動護送隊伍；
   handler50 只讓 NPC 引路，玩家自行踩 `(50,15)` 才進 handler57。此段已達 E3，
   信件不消耗（`docs/88`）。
6. **帶商人建城 questline** — ✅ **核心已接(2026-06-27)**(杜 Ch33-36):
   - **RE 全解**:老人 handler **0x5aba**(非 NPC 跳表 → runner/region 事件,無放置 NPC):`cl=[0x5077]隊伍數`,
     迴圈掃隊員 `cmp [si+1],6`(class==6=**商人**);無商人→rec1「只要有商人」;有商人→rec2-5(如何/是嗎/謝謝)
     + clear_flag 0x23(建城)。對話 D3TXT07 rec0-5。
   - **接線**:CTY83 (16,2) 老人 NPC 位:掃隊伍 class6 商人 → `dq3_roster_remove`(寄存,裝備回阿里阿罕預存所,杜)
     + flag 0x216 建城完成 + rec5「謝謝你」;無商人→rec1。黃寶珠 0x6a(Ch41 革命後)已在 treasure 表 CTY83。
   - ✅ **視覺增生已做(2026-06-27)**:RE 發現 CTY83 的 17 個商店/居民 NPC **vis_flag=0x05**(原版「建城後才顯示」旗標)。
     remake 原本載入全部 NPC(忽略 vis_flag)→ 城鎮永遠建好態。修:`dq3_scene_hide_unbuilt(cur,0x05,16,2)` per-frame——
     未建城(flag 0x216 未設)隱藏 vis_flag 0x05 NPC(留老人 (16,2)),呈空草原;帶商人建城後重入 → 商店/居民全顯
     (可買更多東西=成就感,使用者要求)。圖驗證:建城前空草原、建城後居民現身。game_tester 93/93。
7. **耶進貝亞倉庫番推三石** — ✅ **Go/Ebitengine 正式流程已閉合（2026-08-02）**。
   - IDA 證明原版 pushability 是 runtime NPC `ctrl bit0x40`，不是舊 C prototype 的
     `B2==40`；隱形 `0xc000` 尚在時不得推動。
   - handler30 固定檢查 NPC slot0–2 的 Y 是否全為 5；成功 clear story flag `0x3b`
     並刷新 passage。沒有 remake 自造的 `g_sokoban_solved` 或 debug token。
   - CTY76 event0 `{item=0x5e,flag=0x3a}` 已遷入 game-pack `treasure_events`；event1／
     flag `0x3b` 是 passage 狀態，不是第二個可取寶物。正式 trace、raw bytes 與圖見 `docs/97`。
8. **勇氣神殿「單獨戰鬥」gate** — ✅ **Go/Ebitengine 正式流程閉合（2026-08-02）**。
   - 非破壞性 IDA 9.4 重查推翻舊結論：handler37 **不寫** flag `0x13`；接受時保存
     active party count、強制 count=1、寫 world `(82,165)` 並設 mode bit `0x80`。handler62
     播 rec12 後還原 count 並清 mode bit。flag `0x13` 只有已完成分支 reader，writer 仍 unknown。
   - 舊 C remake 的 `{82,165}: flag0x13→CTY75, else→CTY47` 方向也相反；原版 reader 是
     clear→CTY75、set→CTY47。錯誤 `owPortal` 已從 Ebiten 移除，由 game-pack 有限事件接管。
   - CTY23 sec2 藍寶珠 raw event 為 `01 67 00 ad`；D3TXT06 rec9–12／84 與 D3TXT07
     rec66–67 已逐 word parity。單人機制不是 moot：正式 trace 與原版影片都驗證四人→單人→四人。
   - schema `0.1.15`、途中／復隊 save-load、兩張 runtime PNG 與完整證據見 [`docs/100`](../100-lancel-courage-blue-orb-production-trace.md)。

> **B-6/B-8 範圍校正(2026-06-27)**:原列為「玩家感受明顯的快速接線」**低估了**——兩者 NPC 尚未 RE 識別、
> 且需新機制(商人寄存+建城狀態 / 神父 gate+單獨戰鬥),屬需各自一輪 RE 的中型工項,非 quick win。
> 獎勵道具(0x6a/0x67)皆已 data-driven 可取,**不阻塞破關**。

### C. 終盤戰鬥
9. **索瑪城大魔人／最終戰（2026-07-28 更正）**：本機完整原版影片顯示大魔人為城內一般隨機遭遇，
   可一次出現兩隻；CTY90 sec5 六個 events 是 item0x4c、flags0xef..0xf4 的世界樹之葉寶箱，不是守衛。
   IDA formations `4EFA/4EFF/4EF0` 分別為怨靈122、殭屍123、索瑪124。舊 remake 自造的
   `6×monster106 + flag0x214` 已撤回；正式序列改為 `122→123→D3TXT07 rec72→124`。

### D. 場景轉換
10. **蓋亞那洞穴跳坑 in-game 觸發** — ✅ **R-4 重新依原版接通(2026-07-28)**:
    舊 runner ev86/`0x783d` 結論是位址高位遺漏造成的索瑪終戰誤認。真正流程是巴拉摩斯
    CLR 0x29→阿里阿罕 rec98/99 CLR 0x4d→CTY72 第二筆 transition→CTY77→`0xfe`
    出口到 DQ3UND `(85,67)`；Ebiten KeyU 已移除。

### E. Post-game(青衫專屬)
11. **全破後裝備變洛特系** — ✅ **已接(2026-06-27,時代結尾)**(青衫:王者之劍→洛特之劍、光之鎧甲→洛特之甲、勇者之盾→洛特之盾)。
    上網查證:精訊版 ITEM.DAT **本無洛特道具**(decode 確認)→ 依使用者指示「原版沒有就自己加」補上。
    破索瑪(run_finale)→ `dq3_set_loto_blessed(1)` + flag 0x217:王者之劍 0x1c 攻120→150(洛特之劍)、
    光之鎧甲 0x3f 守70→95(洛特之鎧)、勇者之盾 0x40 守35→55(洛特之盾)+ 「そして伝説へ…」冊封結尾
    (接通 DQ1/DQ2 洛特血脈,圓「傳說的終章」副標)。game_tester 91/91。

## ✅ 「正確地不做」——非缺口(忠實度上應排除)

> 這些是 review 時最容易誤判成「缺口」的,特別釘住(Chesterton's fence):

- **青衫/孔方兄列的當機 bug**:龍王城父親-五頭龍當機、巴拉摩斯/索瑪「打不死」、彩虹水滴→銀寶珠、
  寶箱開不開 —— 正是 remake **已在 C 層修掉的 7 個 bug**。忠實還原 = **正確地不重現這些 crash**,不是缺口。
- **精訊版原本就漏的**(青衫/孔方兄明指):大地盔甲(勇氣洞窟)、某墓前種子等 FC 有、精訊無 ——
  remake 對齊精訊版,**正確地不補**。
- **沙曼歐莎南洞「全寶箱變米米克」**(青衫):若 treasure 表忠實萃取自精訊 CTY 資料,本就如此,非缺口。

## 現況解讀（2026-08-02）

本檔是舊 C/SDL prototype 的線索盤點，不是 Go/Ebitengine 完成證明。單元測試、raw treasure
table 或可直接呼叫的 handler 都不能代替正常玩家入口。唯一 current plan 是
[`docs/74`](../74-ebiten-remake-completion-plan.md)。

目前 boot 起的 Go production trace 已閉合隱形草、CTY39 守衛、CTY76 推石、乾渴壺取得、
船上唯一座標使用、原版 `5×4` 地圖 patch、CTY40 最終鑰匙、魯拉船重定位、提頓夜間牢門、
綠色寶珠及各 checkpoint save/load。詳見 `docs/98`、`docs/99`；下一個 blocker 必須從
綠色寶珠合法 checkpoint 繼續重播後才判定。其餘
舊項目也必須以正式 `InputState`、save/load 與原版證據逐項重新驗收。
