# dq3_remake WORKLIST(compact 後接續依據)

> 目標:把精訊版 DQ3(Dragon Fighter III, 1993)用現代 C + SDL2 **跨平台重製**,放在 `dq3_remake/`,
> 行為與原版**一模一樣**(DOSBox 原版當 oracle 逐畫面驗證),且 **7 個 bug 全在 C 層修對**。
> 紀律:**不用 subagent,全部自己在主執行緒做**;一切編譯/執行在 **docker**(不污染 host);每階段有產出就 commit+push。

## 現況快照(已完成)

- **素材全解碼**:字型(docs/02)、文字+UTF-8 劇本(03)、地圖 tile/世界圖/CTY 城鎮(04)、標題/PCX(docs/title)、怪物 sprite+數值(16)。
- **EXE 反組譯**:recon(05)、callmap+runtime API(06)、函式圖(08,`docs/data/exe_funcs.json`)、loaders(09)、戰鬥(13)、注音輸入(15)。`re/*.c` 為邏輯草稿。
- **RE 正確性鐵證**:整檔反組譯→nasm 重組 **sha256==原版(byte-identical 100%)**(docs/17)。編譯器=MSC 5.x(docs/19)。
- **bug 對照組(a)**:binary patch 修正版 5/7(#1/2/3/4/7a 修、#7c 不需、#5/6/7b 留 C),DOSBox 驗無回歸(docs/18,20);完整遊戲包 `work/dq3_fixed_v1.zip`(本機,不公開)。
- **remake 地基(階段①)✅**:`dq3_remake/` SDL2 骨架編譯成功,headless 解 `TITG.P` 顯示標題正確(`scripts/build.sh`)。
- **remake 階段② 進行中**:
  - `dq3_scene` 場景核心(攝影機/tile 貼圖/bit0 碰撞/走動)+ `dq3_field`(地表)/`dq3_town`(城鎮 CTY)薄載入器。headless 驗證走動+碰撞+捲動正確。
  - **真主角 sprite ✅**:DQ3MAN.BLS 完整破解(masked 32×24、stride960、4 方向,docs/27);`dq3_sprite` 解碼 + scene 透明 blit,entry16 金髮勇者顯示於地表/城鎮、朝向正確。DOSBox oracle 自動進遊戲打通(docs/29)。
- **remake 階段④ C 層 bug 修正(誠實狀態,見階段④ claim 校正)**:**已整合進可玩系統** = #1 結算、#2 合成(gate CTY93)、#3 blit guard、#8 palette、真傷害公式、**#4/#5/#6 升級系統**(`dq3_member` + 戰鬥勝利結算)、#7a/#7b(battlescene)。**邏輯實作+單測** 已大致收斂。#6 註:原版確切 wrap 點未定位、核心修法為 uint16 全程(非 clamp 數值)。
- **知識庫(repo 級)✅**(2026-06-23,接續 session 8db027c4 斷三次的請求):`CONTEXT.md`(canonical 術語表 + `docs/00–35` 主題索引 + 待釐清詞,即「避免重複採坑」入口)、`docs/00-re-methodology.md`(4 條可重用 RE 技巧:羅塞塔範本 / 跳表派發 / file⇄logical 位址基準陷阱 / 用資料內容反證標註)、`~/.claude` 跨 session 記憶 ×4。已 commit+push(`ecd4f5c`)。

## dq3_remake 剩餘 worklist

### 階段② 移植遊戲邏輯(主迴圈→可走動)
- [x] runtime shim:鍵盤 scancode→SDL keysym(0x48/0x50/0x4b/0x4d/Enter/Space/ESC)、計時、檔案 IO。(滑鼠 int33h→SDL 待戰鬥/選單時補)
- [x] 資產載入 C 化(地表):DQ3.PAL(`dq3_pal`)、DQ3.BLK(`dq3_blk`,32×24 4-plane planar)、DQ3CON.MAP(`<w><h>`+1byte/tile)、BLKBM.DAT(u16/tile 屬性,bit0=阻擋)。
- [x] **地表 field 引擎**(`dq3_field.c`):攝影機跟隨玩家(20×15 視窗,邊界夾住)+ tile 預解貼圖 + 方向移動 + bit0 碰撞。headless 驗證:走 16 步全成功、海/山正確擋住、捲動正確(`field0/1.png`)。
- [x] **scene 核心重構**:抽出 `dq3_scene`(攝影機/貼圖/bit0 碰撞/走動),地表/城鎮共用;`dq3_field`/`dq3_town` 降為薄載入器(deep module + adapters at edges)。
- [x] **城鎮 CTY 引擎**(`dq3_town.c`):解 section 偏移表→layout_ptr→`<w><h><spawn>`+w×h u16(低 byte=index),載 DQ3N.BLK+BLKBMN.DAT。headless 驗證 CTY00 sect0(16×12):磚牆/紅地毯/床/寶箱正確,走動有牆碰撞(`town0/1.png`)。
- [x] **真主角 sprite ✅**:破解 DQ3MAN.BLS 格式(反組譯 blit sub_1ed0:stride 960、32×24、4-plane plane-major + AND 遮罩@+0x180、DQ3.PAL,每角色 4 方向 frame,docs/27)。`dq3_sprite` C 解碼器 + scene 透明 blit;主角=entry16(金髮勇者,對 oracle 室內主角)、facing→frame 預設 {0下,1左,2上,3右}(frame0/2 已驗正面/背面)。實測:地表/城鎮顯示真主角、走動朝向正確、透明乾淨,取代佔位框。
  - DOSBox oracle 自動進遊戲打通(`tools/dosbox_to_overworld.sh`,docs/29):注音輸入完成序列 RE 出 → 截到起始房間 + 城鎮主角,作為 sprite/palette 校正與 phase⑤ 比對基準。
- [ ] 主迴圈整合(re/mainloop.c `sub_93e3` 結構):場景旗標 [0x4f2d] 0=地表/1=城鎮切換、過場、11-entry 狀態機跳表(指令/對話/戰鬥)。
- [ ] **里程碑**:能在城鎮/世界地圖走動,對 DOSBox 原版同畫面比對一致。(地表走動已通,城鎮 + 真主角 sprite + 換圖過場待補)

### 階段②.5 串接狀態機(game 模式)✅
- [x] **地表↔城鎮↔戰鬥串接**(`main.c` run_game):地表走動 + 步數隨機遭遇 → 互動戰鬥 → 回地表;SPACE 進/出城鎮。每次戰鬥/換場後 `dq3_scene_apply_palette` 重套目的場景色盤(**實際運用 bug #8 修正**)。headless 實測:走 8 步→遭遇 id1×2→腳本戰勝→回地表→進城鎮 CTY00(主角顯示)。
- [整合] **section/CTY 轉場(門/階梯/出城)— 改用 +0xc 轉場表(docs/35,主機制)**:
  逐 handler 重新 RE 後確認,**所有城鎮建築門/室內外/上下樓/出城都在 section header `+0xc` 的轉場表**
  (4-byte 項 `{destCTY,destSec,X,Y}`,by 轉場 tile 的 subid;轉場 tile = 屬性 attr&0xe000)。
  先前以為「68 個多 section 大城無 type-2、門追不到、需 debugger」是**誤判**——門就是 +0xc 純靜態資料。
  落地:`dq3_town_load` 解析 +0xc → `scene.transitions[]`;`dq3_scene_tile_transition()` 查詢;
  main.c 踩轉場 tile → 同 CTY 換 section / 跨 CTY 載入 / destSec==0xff 出城回地表。
  實測 CTY00 section0:7 門解析正確,(21,8)/(22,8)→CTY25、(8,14)→sec2、(19,36)→sec1… 端到端通。
  舊 type-2 warp(0x4ea0)路徑為較罕見的另一種門,保留 dq3x_warp 資料。

### 階段③ 對話/野外指令/戰鬥
- [x] **文字/對話系統**(`dq3_text`):載 D3TXT00.FON(16×16 字模)+ D3TXTNN.TXT(指標表+2-byte glyph index 記錄),渲染對話視窗,處理控制碼(0xfffe 換行/0xfffc 換頁/0xffed+ 變數占位)。實測渲染阿里阿罕 NPC 真繁中對話(「鎮上的人所說的魔王要消滅世界是胡說的吧」)。→ 解鎖對話/NPC/選單/敵名/道具名/系統訊息。
- [ ] 把文字接進:戰鬥(指令戰/逃/防/道具、敵名、HP 數字)、城鎮 NPC 對話觸發、道具/咒文選單。
- [ ] 注音姓名輸入(re/nameinput.c,docs/15:5×9 grid=0..44 1-D ring,Up=−9/Down=+9/Left=−1/Right=+1 mod45;組字 lcall 11c4:0x27;完成在功能列第5列)。
- [ ] 對話流程(re/commands.c,Enter sub_7c43→事件表 `[ft*3+0x37c4]`;文字繪製器 4 行/頁、控制碼換行/換頁/變數)。
- [ ] 野外指令選單(DS:0x3baa 12 指令:話す/移動/調べる…)。
- [~] 戰鬥(優先 2,進行中):
  - [x] **怪物資料 + sprite 基礎**:`dq3_monster` 讀 D3MNS.DAT(130×41 數值)+ 解 DQ3MNS.SHP(offset 表 + plane-major,MNSBK.PAL,色0 透明)。單測:史萊姆 HP6/exp4、金屬 exp4140、sprite 48×39。
  - [x] **戰鬥場景繪製**:`main.c` battle 模式 — MNSBK.PAL + 怪群 sprite blit(透明)+ 天空/地面背景。實測史萊姆×6 對 references/game3.png 一致(藍天/綠地/6 隻藍史萊姆)。
  - [x] **互動戰鬥迴圈**(`dq3_battlescene`):遭遇怪群 + 指令選單(戰/逃/防/道具,游標)+ 回合執行(物理傷害公式 `dq3_battle_phys_damage`,會心一擊、防禦減半、逃跑判定)+ #1 正確結算 + 隊伍 HP 條 HUD。可互動(方向+Enter)可 headless 腳本(DQ3_BATTLE_SCRIPT)。實測史萊姆×3 FFFF→勝。
  - [x] **戰鬥 CJK 文字**:敵名(D3TXT00 rec 0x258+id)、指令選單(戰鬥/逃跑/防禦/道具)、隊伍 HP 數字,皆以 dq3_text 字模渲染。戰鬥畫面與 game3.png 一致(經典重現)。
  - [ ] 潤飾:packbg 戰鬥背景解碼;遭遇生成 sub_a7d5;復原 Ortega/Hydra 填 #3 空槽;傷害公式對 DOSBox 校準。
  - [ ] 用復原的 Ortega(128)/Hydra(129)sprite(tools/make_sprites.py)填空槽。

### 階段④ 7 bug C 層修正(誠實狀態:邏輯實作+單測 vs 整合進可玩系統)

> **誠實校正(claim 整頓)**:多數 bug 的修法是「正確邏輯 + 決定性單測」,但**只有部分已整合
> 進實際可玩的遊戲系統**(因戰鬥/升級/能力值系統尚未完整)。標記:
> **[整合]** = 已接進可跑的系統;**[單測]** = 邏輯實作+單測過、尚未整合;**[~]** = 根因記錄、未實作。
> 不再籠統掛「已修」。

- [整合] **#1 巴拉摩斯打不死**:`dq3_battle_resolve` 正確結算(先判我方全滅含被吹飛→敗)。
  **已接進** `dq3_battlescene::do_turn`(同款結算)。單測:全隊被吹飛+巴拉摩斯HP500 → 修正判敗 vs 原版誤判勝。
- [整合] **#2 彩虹水滴卡關**:`dq3_inventory` + `dq3_synth_rainbow_drop`(對齊 RE file 0x77ce:消耗太陽之石0x72+雲雨之杖0x73→雲雨之杖格寫成品)。修正:產出彩虹水滴 0x75(原版誤產銀寶珠 0x6b)。**已接 in-game**:`dq3_synth_shrine_examine` + 劇情旗標 0x139(未合成過才觸發、成功設旗標),game 模式城鎮調べる持兩材料即合成(headless demo log:`result=0x75 flag0x139=1`)。`dq3_scripted_event_run` 鏡射 runner(0xabb2)的 id→handler 跳表(DS 0x3baa),合成 = scripted-event 83。單測:修正→0x75、原版→0x6b、缺料不合成、祠堂觸發+旗標、不重複、event 派發。**待(改動態)**:精確祠堂座標 — 靜態已窮舉(0x77ce=handler 0x776c 尾段;runner 0xabb2 **零靜態呼叫者**,經事件 VM 進入),需 DOSBox 中斷點 runner phys 0x9842 讀地圖/座標定位(docs/31)。
- [整合(部分)] **#3 九頭龍/歐里狄加當機(缺 sprite)**:`dq3_monster_sprite_decode` 對空 sprite 回 <0 = blit guard,
  **已整合**(battlescene 不畫空 sprite 不當機)。**但復原的 Ortega/Hydra sprite 尚未填入**(仍只是 guard,非顯示)。
- [整合] **#4 勇者 MaxMP 成長偏低**:`dq3_stats` 內建成長表,勇者 MP base 3→8/slope 5→10。
  **已整合**:`dq3_member`(隊員實體)init/gain_exp 用成長表;`dq3_battlescene` 勝利結算
  套用 → 升級時 MaxMP 依修正表成長。單測 Lv43 110→223、整合測 gain_exp→Lv43 MaxMP=223。
- [整合] **#5 高等級升級錯亂**:`dq3_stats_level_for_exp(fixed)` clamp level≤43。
  **已整合**:`dq3_member_gain_exp` 一律用 fixed 版求等級 → 隊員 level 夾 43,杜絕越界連升。
  單測:原版連升到 99 vs 夾 43;整合測:灌 9.96M exp → 隊員 level=43。
- [整合] **#6 數值 255 溢位**:`dq3_stats_add_clamped`(uint16+clamp 9999)。
  **已整合**:`dq3_member.stat[6]` 全程 uint16,`gain_exp` 升級套成長一律經 add_clamped 寫回
  → 設計上免疫 255 wrap(=#6 真正修法)。單測:200+100 wrap44 vs 300;整合測:屬性 ≤9999。
  註:原版確切 wrap 點仍未定位(docs/23);clamp 9999 為上限佔位,核心修正是 uint16 全程。
- [整合] **#7a 隼劍雙擊**:`dq3_combat_num_attacks`(武器 +5 bit7;機制對齊 RE file 0xc1fa 二擊重入)。
  **已接進** `dq3_battlescene::do_turn`(飛鷹劍員打 2 次)。實測:勇者(飛鷹劍)第1擊+第2擊、武鬥家普通武器 1 次。單測:飛鷹劍 2 次。
- [整合] **#7b 魔法鎧甲抗魔**:`dq3_combat_spell_damage`(裝備 +6 bit2 減半)。**已接進** battlescene 敵方咒文路徑。
  實測:僧侶(魔法鎧甲)受咒文減半(7 vs 別人 12-15)。單測:魔甲 80→40。
- [整合] **物理傷害公式**:`dq3_battle_phys_damage` = 反組譯真公式 `(atk−def)/2+rng((atk−def)/4)`(file 0xc03e,docs/13;
  非先前湊的 atk/2−def/4)。**已接進** battlescene。單測對齊。
- [ ] #7c 祈禱之戒:原版本就會壞(~25%),保持即可(不需修)
- [x] **#8 戰鬥後畫面變黃/綠不恢復**(玩家實機回報;docs/28):根因=戰鬥切 MNSBK.PAL 進 DAC、離場只重載 tile 不還原 DQ3.PAL。remake 從設計免疫:`dq3_scene_apply_palette` 契約,場景切換/戰鬥返回必重套目的 palette(已在 run_scene 落地)。原版 byte-level 修點留 DOSBox oracle 回合。

### 階段⑤ oracle 驗證 + 打包
- [ ] DOSBox 原版 vs remake **逐畫面比對一模一樣**(bug 場景=修正後正確行為);用 `tools/dosbox*` 截原版、remake headless 截圖比對。
- [ ] retro-cjk-hires-canvas:內部畫布原生解析、nearest 放大;CJK 16×16 正常。
- [ ] 海面 palette cycling(DQ3.PAL idx2/10 DAC 動畫)。
- [ ] 跨平台打包(Linux/Windows;AppImage/exe)。

### 接續中(session 8db027c4 收尾時點名、未動工的靜態 RE)
- [x] **鑰匙門機制 ✅ 靜態全解**(docs/35 §八):鑰匙 id=0x55/0x56/0x57(盜賊/魔法/最終,等級=id−0x54);
  內嵌城鎮轉場 handler `0x488f`:`0x48c3` 掃全隊 8 道具格取最高鑰匙級 `[0x2593]`;`0x4906` 讀面向 tile,
  attr **低 byte bits6-7**=門所需級,隊伍級≥需求 → 改 tile `[di]=low&0x1f` + 清 attr `[di+1]&=0xe0` + 開門動畫。
  糾正舊標(player.c 誤把低 byte 0xc0 標角落,實為鑰匙級;角落用的是高 byte 0xc000)。
  > 用「鑰匙道具 id 反查 + 回 EXE 重驗」破解,正是 session 走死胡同時回 dq3.exe/腳本重新檢視的成果。
  - [x] **remake 落地 ✅**:`dq3_inv_key_tier`(掃道具取最高鑰匙級,鏡射 0x48c3)、
    `dq3_scene_door_tier`/`_open_door`/`_try_open_facing_door`(鏡射 0x4906/0x4977)、
    main.c 城鎮被擋時試開門重試;單測 `dq3_door_test` + inventory 鑰匙級測試全綠,full build OK。
    剩:接真實隊伍道具來源(目前單一 `dq3_inventory`)、開門音效/動畫。
- [ ] **NPC 移動步進規則**:handler `0x62e9`/`0x6355`/`0x839f`(隨機/巡邏/靜止)的步進條件。
- [ ] **CTY→地名 對照收尾**(untracked WIP):`docs/maps/cty_name_fill.md` 只填到 CTY2;
  配套 `tools/_big.py`(cty_loc 疊圖)、`tools/dosbox_walk_test.sh`、`new_map_dq3/` 待決定納版控或 gitignore。

## 關鍵參考 / 工具 / 紀律

- **格式/位址**:docs/02–16(素材)、docs/05–13/15(EXE 邏輯)、`docs/data/exe_funcs.json`(函式圖)、`re/*.c`(邏輯草稿,移植來源)。
- **編譯/驗證**:`dq3_remake/scripts/build.sh`(docker SDL2 編譯+headless 標題驗證)、`tools/dockrun.sh`(docker uv python)、`tools/dosbox_run.sh`+`dq3-dosbox` image(原版 oracle)。
- **原始素材**:`assets_raw/`(版權,gitignore;remake 執行期指向它)。
- **不公開**:整套可玩遊戲(`work/dq3_fixed_v1.zip`)、原始檔、第三方攻略/截圖、render 的版權畫面像素(`*.ppm`/`titg*.png` 等)——見各 `.gitignore`。
- **紀律**:不用 subagent;docker first;每階段 commit+push;commit message 繁中 + `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`。
