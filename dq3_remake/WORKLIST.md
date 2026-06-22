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
- **remake 階段④ C 層 bug 修正 ✅(可單測者)**:`dq3_stats`(#4/#5/#6)+ `dq3_combat`(#7a/#7b),各附決定性單元測試(`dq3_stats_test`/`dq3_combat_test`,build.sh 整合全通過)。#1/#2/#3 屬戰鬥/事件邏輯,根因+修正值已記錄,移植該系統時依正確值寫入即修復。

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

### 階段③ 對話/野外指令/戰鬥
- [ ] 注音姓名輸入(re/nameinput.c,docs/15:5×9 grid=0..44 1-D ring,Up=−9/Down=+9/Left=−1/Right=+1 mod45;組字 lcall 11c4:0x27;完成在功能列第5列)。
- [ ] 對話流程(re/commands.c,Enter sub_7c43→事件表 `[ft*3+0x37c4]`;文字繪製器 4 行/頁、控制碼換行/換頁/變數)。
- [ ] 野外指令選單(DS:0x3baa 12 指令:話す/移動/調べる…)。
- [~] 戰鬥(優先 2,進行中):
  - [x] **怪物資料 + sprite 基礎**:`dq3_monster` 讀 D3MNS.DAT(130×41 數值)+ 解 DQ3MNS.SHP(offset 表 + plane-major,MNSBK.PAL,色0 透明)。單測:史萊姆 HP6/exp4、金屬 exp4140、sprite 48×39。
  - [ ] 遭遇生成(sub_a7d5:遭遇區→候選怪→點數預算生群)、戰鬥主迴圈(指令戰/逃/防/道具)、戰鬥場景繪製(packbg 背景 + 怪群 sprite blit)。
  - [ ] 用復原的 Ortega(128)/Hydra(129)sprite(tools/make_sprites.py)填空槽。

### 階段④ 7 bug 全修進 C(根因見 docs/18,20,22,23)
- [x] **#1 巴拉摩斯打不死**:`dq3_battle_resolve` 正確結算(先判我方全滅含被吹飛→敗,再判敵全滅→勝)。單測:全隊被巴西魯拉吹飛+巴拉摩斯HP500 → 修正判敗、原版誤判勝(重現 #1)。
- [~] #2 彩虹水滴卡關(合成成品 item code 0x6b→0x75):根因/修正值已記錄(docs/18),屬合成事件 handler,移植事件系統時產出 0x75 即修復。
- [x] **#3 九頭龍/歐里狄加當機(缺 sprite)**:`dq3_monster_sprite_decode` 對空 sprite(id128/129)回 <0 = blit guard(不當機);單測坐實 id128/129 為空。復原 sprite(make_sprites.py)填槽待整合進戰鬥繪製。
- [x] **#4 勇者 MaxMP 成長偏低**:`dq3_stats` 從 DQ3.EXE 讀成長表,勇者 MP base 3→8/slope 5→10。單測:Lv43 MaxMP 110→223。
- [x] **#5 高等級升級錯亂**:`dq3_stats_level_for_exp(fixed)` clamp level≤43。單測:原版越界連升到 99、修正版夾 43。(binary patch 因 code cave≤8B 做不到,C 層單行 if)
- [x] **#6 數值 255 溢位**:`dq3_stats_add_clamped` 用 uint16 + 顯式 clamp(9999)。單測:200+100 原版 byte wrap=44、修正=300。
- [x] 三者以 `dq3_stats_test` 決定性單元測試驗證(build.sh 整合,全通過);資料從使用者 DQ3.EXE 讀,不入 git。
- [x] **#7a 隼劍雙擊**:`dq3_combat_num_attacks` 依武器 +5 bit7 決定 2 次(通用,非寫死 0x6e)。單測:飛鷹劍 2 次、普通劍 1 次。
- [x] **#7b 魔法鎧甲抗魔**:`dq3_combat_spell_damage` 掃 8 格裝備 +6 bit2,命中減半。單測:魔甲 80→40、無魔甲 80。(binary patch 因需掃描迴圈+無 cave 做不到,C 層完成)
- [x] #7a/#7b 以 `dq3_combat_test` 決定性單測驗證;ITEM.DAT(128×7)從使用者 DQ3.EXE 旁讀,不入 git。
- [ ] #7c 祈禱之戒:原版本就會壞(~25%),保持即可(不需修)
- [x] **#8 戰鬥後畫面變黃/綠不恢復**(玩家實機回報;docs/28):根因=戰鬥切 MNSBK.PAL 進 DAC、離場只重載 tile 不還原 DQ3.PAL。remake 從設計免疫:`dq3_scene_apply_palette` 契約,場景切換/戰鬥返回必重套目的 palette(已在 run_scene 落地)。原版 byte-level 修點留 DOSBox oracle 回合。

### 階段⑤ oracle 驗證 + 打包
- [ ] DOSBox 原版 vs remake **逐畫面比對一模一樣**(bug 場景=修正後正確行為);用 `tools/dosbox*` 截原版、remake headless 截圖比對。
- [ ] retro-cjk-hires-canvas:內部畫布原生解析、nearest 放大;CJK 16×16 正常。
- [ ] 海面 palette cycling(DQ3.PAL idx2/10 DAC 動畫)。
- [ ] 跨平台打包(Linux/Windows;AppImage/exe)。

## 關鍵參考 / 工具 / 紀律

- **格式/位址**:docs/02–16(素材)、docs/05–13/15(EXE 邏輯)、`docs/data/exe_funcs.json`(函式圖)、`re/*.c`(邏輯草稿,移植來源)。
- **編譯/驗證**:`dq3_remake/scripts/build.sh`(docker SDL2 編譯+headless 標題驗證)、`tools/dockrun.sh`(docker uv python)、`tools/dosbox_run.sh`+`dq3-dosbox` image(原版 oracle)。
- **原始素材**:`assets_raw/`(版權,gitignore;remake 執行期指向它)。
- **不公開**:整套可玩遊戲(`work/dq3_fixed_v1.zip`)、原始檔、第三方攻略/截圖、render 的版權畫面像素(`*.ppm`/`titg*.png` 等)——見各 `.gitignore`。
- **紀律**:不用 subagent;docker first;每階段 commit+push;commit message 繁中 + `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`。
