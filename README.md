# 精訊版《勇者鬥惡龍 III》反組譯與 remake

> 桌面版 v0.1.34 已由 checkpoint `9d639d0` 正式發布；公開 patch 不含原版素材，
> checksum 與驗證界線見 [`docs/131`](docs/131-release-v0.1.34.md)。

本專案研究精訊資訊在 1990 年代製作的中文版 DQ3
（程式內題名 *Dragon Fighter III／傳說的終章*），並以原版 DOS 程式與資料為證據，
製作可在現代平台執行的 Go／Ebiten remake。

## 專案主軸：保存台灣遊戲文化資產

這不只是把一款舊遊戲換成新的執行檔；我們要保存的是台灣 1990 年代「移植、中文化、
玩家社群接力維修」的整套文化與技術脈絡。現存的 1994–1995 BBS 史料與精訊音樂研究
記錄，留下了未正式發售版本的身世、玩家攻略、當年自行摸索的修正方法，以及 DOS 時代
中文字庫、注音輸入、地圖引擎與 Sound Blaster FM 音訊整合的工藝。這些內容和原始
`DQ3.EXE` 一樣，都是本專案要保存、註明來源、讓後人可以重新檢查的對象。

我們以三層方式保存：

- **玩家記憶與社群史料**：保留杜勝利、孔方兄、青衫詩客等攻略作者的來源與致敬，並連結
  [台灣本土攻略原文收錄](references/walkthroughs/README.md) 及 [1994–1995 BBS 歷史紀錄](docs/history/dq3-bbs-1994.md)。
- **技術與聲音工藝**：記錄 16×16 繁中文字模、注音／英數選字、精訊地名與劇本，以及
  [精訊音樂保存紀錄](docs/58-taiwan-jingxun-music-preservation.md) 和 [MT-32 再詮釋的建置經驗](docs/59-munt-mt32-build.md)。
- **可回查的工程證據**：原始檔、位址、資料格式、推論等級與 remake 的 JSON／runtime
  分開保存；原版素材、影片與 ROM 仍由使用者合法持有，絕不放入公開 Git。

歷史敘述與現行產品狀態刻意分層：攻略與 BBS 是文化史料及路線交叉佐證，`DQ3.EXE`、
原始資料與同狀態實機結果才是行為 oracle；進度以 [`docs/74`](docs/74-ebiten-remake-completion-plan.md)
為準。早期 C/SDL prototype 的完成聲明保留在 Git 歷史與研究文件中，不直接當成現行
Go／Ebitengine 產品的 parity 證據。

截至 2026-08-12，本專案採用使用者確認的 E2／D2 快速驗收：戰鬥 pack raw、loader、
formation／抗性 decoder、正式事件輸入與 save/load 的 Docker headless targeted tests 已通過。
本輪已修正 CTY07→CTY08 的正式 section 路由：依原始 transition table，並用正式道具／戰鬥
輸入處理塔區遭遇，現在可抵達 CTY08 sec3；本輪再修正 CTY13 金字塔開關路徑在遭遇 modal
時未送正式戰鬥輸入的 trace blocker，已可由正式路徑取得魔法鑰匙並通過 save/load。完整
campaign 隨後以正式 `InputState` clean replay 至 `THE END`（63.92 秒）並通過；完整 `game`
回歸以一次性分片程序跑完 288 個頂層測試。因此主線可宣稱 campaign E3 通過，但逐畫面／音效 V3 的長尾仍待
後續對拍，本輪指定的 AppImage、Windows ZIP 與 macOS ZIP 已產出；Android 已另開保存優先
UX 並完成 Docker debug APK（目前僅 `build-only`，尚無 emulator／真機驗收），WASM 仍不在本輪 release 目標。現行進度與工作順序以
[`docs/74-ebiten-remake-completion-plan.md`](docs/74-ebiten-remake-completion-plan.md) 為準。
本次可公開的 patch checksum、Docker smoke 界線與不公開的本機完整版清單見
[`docs/122`](docs/122-release-20260810.md)。
本輪另以 IDA／原始資料閉合 CTY59 `handler41` 的 Y=4 設施分支；它已達 D3／E2／V1，
不代表 CTY60／61 階段或全程畫面 parity 已完成。
日邦格八頭大蛇兩場固定編隊的背景 selector 也已由 IDA 9.4 閉合並遷入 game pack：正式四人
trace 現分別為洞窟與沙漠，達 near-state V2；這是必經 Boss 的可見修正，不把通用地形背景或
逐畫面 V3 一併宣稱完成。該 schema 更新後的正式 `InputState` replay 仍在 Docker＋Xvfb 於
67.24 秒抵達 `THE END`，詳見 [`docs/128`](docs/128-battle-background-selector-re.md)。
其餘必經單頭目（怪力魔、巴拉摩斯與索瑪三連戰）同樣已改由原版固定 record 的 page／palette
selector 啟動，不再落回 generic 草地；目前只有 D2／runtime V1，沒有同狀態原版畫格時不升格
為 V2／V3，詳見 [`docs/129`](docs/129-required-boss-backgrounds-re.md)。
地表四人 HUD 也已由 IDA 的原始 window／內容 writer 訂正為中央 `(152,238,352,80)` 四欄，
姓名、H／M、職業與等級不再使用目測寬框；正式輸入 runtime 達 near-state V2，詳見
[`docs/130`](docs/130-party-field-hud-re.md)。

### V3 近似驗收政策（2026-08-11）

使用者允許在尚缺同狀態原版逐畫面、逐幀時鐘或逐 PCM 停頓證據的項目採近似方式補正。
這類結果統一標為 `V3-approx`：保留原始資料、事件順序、正式輸入、JSON／D3 證據與
存讀檔契約，但不宣稱逐像素、逐色盤或逐波形 parity。這項政策只處理可見呈現與時序的
近似，不放寬玩家流程 gate，也不把 `unknown` 或未閉合的原版語意升格為 `confirmed`。
標題閒置後的八張職業 attract 也已由 game-pack 接線，順序／PCX 達 D3、輪播 timing 達 E2；
能力條逐幀填充與淡入淡出仍待 V3。
正常桌面／mobile 啟動現在先由 game-pack 播放 `TITA.P`、`TITB.P`、`TITD.P`、`TITE.P`、
`TITP.P` 五張開機年代／巨龍過場；任一正式輸入可跳過並交回標題。素材 identity 達 D2、
runtime 達 E2／V1，但每張停留 120 幀是可重播設定，不是原版逐幀 timing；排序、淡入淡出、
TITP 位置與開場音效仍待 V3，詳見 [`docs/120`](docs/120-opening-cutscene-re.md)。
新遊戲能力確認畫面的 `FIRST.SCR` raw 背景、record 407 的十三個具名欄位、三層 EGA backdrop
與 `beveled_2px` frame 已全由 game-pack 載入。正式輸入的固定確認 checkpoint 已完成
640×350 同狀態 V3 靜態對拍（AE=1,474／0.658%）；原始 hash、殘差與反證見
[`docs/126`](docs/126-newgame-confirmation-v3-static-comparison.md)。palette register、游標閃爍、
能力條與整段創角的動態 timing 尚未由這張固定畫面宣稱完成。
地表詳細狀況窗則已依 `DGROUP 0x3DA8`／D3TXT00 record 407 資料化，runtime 達 E2／V2；
狀況子選單、隊員詳情與道具／裝備逐窗仍待同狀態 V3，詳見 [`docs/116`](docs/116-field-status-panel-re.md)。

## 先讀這些

- [`PROJECT_MEMORY.md`](PROJECT_MEMORY.md)：接手時應保留的少量穩定決策。
- [`CONTEXT.md`](CONTEXT.md)：術語、位址口徑與文件索引。
- [`docs/74`](docs/74-ebiten-remake-completion-plan.md)：Go／Ebiten remake 的現行完成計畫。
- [`docs/69`](docs/69-why-remake-looked-done-but-wasnt.md)：過往實作經驗與後續改善方法。
- [`docs/00`](docs/00-re-methodology.md)：反組譯方法、證據分級與常見陷阱。
- [`references/walkthroughs/README.md`](references/walkthroughs/README.md)：台灣本土攻略原文、作者與出處。
- [`docs/history/dq3-bbs-1994.md`](docs/history/dq3-bbs-1994.md)：1994–1995 BBS 一手史料整理。
- [`docs/58-taiwan-jingxun-music-preservation.md`](docs/58-taiwan-jingxun-music-preservation.md)：精訊音樂與台灣電玩史保存紀錄。
- [`docs/124-android-ux-spec.md`](docs/124-android-ux-spec.md)：保存優先的 Android 版 UX 與本輪 build-only 界線。
- [`docs/125-acceptance-20260811.md`](docs/125-acceptance-20260811.md)：Go／Android／原版取樣驗收與 HUD 框線勘誤。
- [`docs/128`](docs/128-battle-background-selector-re.md)：固定編隊背景 archive／palette selector 的原始資料、勘誤與 V2 界線。
- [`docs/129`](docs/129-required-boss-backgrounds-re.md)：必經單頭目固定 record 到 renderer 的 IDA 接線與 V1 界線。
- [`docs/130`](docs/130-party-field-hud-re.md)：地表四人 HUD 原始 window／內容 writer、pack 契約與 near-state V2 勘誤。

較早的 Markdown 可作為研究索引；進度以 `docs/74` 為主，規格則回到 `DQ3.EXE`、原始資料、
DOSBox 實機與本機完整影片核對。

## 現行產品

### `dq3_remake_ebitan/`：主要 remake

Go 1.24 + Ebitengine 的跨平台實作，是目前主要開發的產品線。已建立的基礎包括：

- 原版 palette、BLK、地表、CTY、角色、文字、怪物、道具與音訊資料解析器。
- 新遊戲姓名／性別、開場家中、母親事件、王城獎勵、酒場登錄與四人隊伍。
- 地表／城鎮移動、NPC、日夜、門與轉場、商店、宿屋、教會、道具、裝備與存讀檔。
- 多敵戰鬥、角色成長、部分戰鬥咒文與原版怪物資料。
- 船、不死鳥、巴拉摩斯後下降、六珠事件、終盤連戰與結局的正式玩家流程。
- 部分場景咒文及其 MP、使用條件與玩家可見效果。
- 桌面共用輸入、觸控層與 Android 綁定骨架。

這是能力概覽，不作逐項完成度聲明；各功能的玩家入口、設定值與對拍狀態請查現行計畫及相關
RE 文件。

### 怪物圖鑑與 128／129 歷史回補

Git 歷史中的 `01f023c`（2026-06-22）記錄了未完成怪物槽位的回補：`id 128` 歐里狄加
（Ortega）與 `id 129` 五頭龍大王（King Hydra）原本沒有可供解碼的 sprite，父子決鬥因此
可能在 blit 階段讀到空資料。後續 `8a538a5` 將圖鑑補到 130 格，並把 `id 129` 的配色改成
對齊實機索瑪最終戰的綠色；這裡是歷史研究紀錄，不把提交訊息當成新的原版規格。

回補工具 [`tools/make_sprites.py`](tools/make_sprites.py) 以參考圖產生符合 `MNSBK.PAL` 的
16 色、4-plane planar SHP 資料；[`tools/style_sprites.py`](tools/style_sprites.py) 再附上
原版 consumer 所需的逐列 RLE AND-mask。產生的 `work/DQ3MNS_fixed.SHP`，以及存在時供
Ebitengine mobile embed 使用的 `dq3_remake_ebitan/mobile/assets/DQ3MNS.SHP`，都是從唯讀
`assets_raw/DQ3MNS.SHP` 衍生的執行版副本；原始基線不會被覆寫，也不納入 Git。重新匯出的透明
PNG 保留原始槽位與像素比例，並以與現有圖鑑相同的深色索引背景展示：

![130 格怪物圖鑑](docs/monsters/monster_sheet.png)

![128／129 回補（同一 MNS palette 與像素縮放）](docs/monsters/restored_128_129.png)

個別匯出可直接檢查 sprite 邊界與透明遮罩：[`spr_128.png`](docs/monsters/spr_128.png)、
[`spr_129.png`](docs/monsters/spr_129.png)。格式、雜湊與風格檢查見
[`docs/121`](docs/121-monster-128-129-style-audit.md)。

目前可見成果（以下為現行 Ebitengine runtime 證據；未另標示者是 V2，不代表已完成原版
同狀態逐像素 V3）。本輪新增由正式新遊戲輸入 trace 產生的終盤固定片尾畫面與配樂 cue；
其餘場景仍依各圖的 V1/V2 標註，不能把單一終盤畫格擴大解讀成全程 V3。終盤證據／剩餘
畫面工作見 [`docs/106`](docs/106-gaia-sword-volcano-re-worklist.md)。

README 內的戰鬥圖使用 Docker 正式 renderer，並使用預設 `combat_info=0`。其中日邦格兩張
已於 2026-08-12 以正式四人 `InputState` trace 重產：固定編隊 selector 分別取原版 page 35／
bank 5 與 page 26／bank 5，作為 near-state V2 證據；它們不是逐像素 V3。敵人上方的 `H/HP`
是可選的 remake 診斷資訊，不是精訊版原版畫面，因此不應出現在這些對拍圖；玩家可見的戰鬥
文字也必須由框線／面板 renderer 佈局，若文字落在框外即視為尚未通過 V3 對拍。
怪力魔、巴拉摩斯與索瑪三連戰也已使用各自固定 record 的背景 page／bank；由於尚無可重播的
同狀態原版畫格，README 不把 Docker runtime 圖包裝成 V2 對拍。

![Ebiten：原版新遊戲開場](dq3_remake_ebitan/docs/opening_home_rec82.png)

![Ebiten：場景咒文選單](dq3_remake_ebitan/docs/field_spell_menu.png)

![Ebiten：諾魯德密道正式流程](dq3_remake_ebitan/docs/norud_passage_trigger.png)

![Ebiten：巴哈拉達救援的甘達特混合編隊（使用原版 SHP AND-mask）](dq3_remake_ebitan/docs/baharata_boss_formation.png)

![Ebiten：救援後取得黑胡椒](dq3_remake_ebitan/docs/baharata_pepper_received.png)

![Ebiten：波魯多加取船後首次航行](docs/img/ship_first_sailing.png)

![Ebiten：達瑪神殿賢者轉職選單](docs/img/dhama_menu.png)

![Ebiten：達瑪神殿轉職成功](docs/img/dhama_reclass_success.png)

![Ebiten：提頓取得黑暗之燈](docs/img/teidon_dark_lamp_obtained.png)

![Ebiten：地表使用黑暗之燈後進入夜晚](docs/img/teidon_dark_lamp_night.png)

![Ebiten：日邦格八頭大蛇第一戰（正式四人 trace、洞窟背景，near-state V2）](docs/img/jipang_orochi_first_battle.png)

![Ebiten：第一戰後依原版移動至宮殿並顯示 rec70](docs/img/jipang_orochi_first_post.png)

![Ebiten：拒絕無姬後的第二戰（正式四人 trace、沙漠背景，near-state V2）](docs/img/jipang_orochi_second_battle.png)

![Ebiten：第二戰後正式調查取得紫寶珠](docs/img/jipang_purple_orb_obtained.png)

![Ebiten：愛丁貝亞可見玩家被守衛追蹤阻擋](docs/img/eginbear_guard_visible_block.png)

![Ebiten：使用隱形草後守衛留在左欄，玩家由右欄通過](docs/img/eginbear_invisibility_pass.png)

![Ebiten：愛丁貝亞地下室推石前](docs/img/eginbear_push_puzzle_before.png)

![Ebiten：三顆石頭推至完成列後 passage 開啟](docs/img/eginbear_push_puzzle_solved.png)

![Ebiten：船上使用乾渴壺後顯現最終鑰匙祠堂](docs/img/thirsty_pitcher_revealed.png)

![Ebiten：祠堂內正式調查取得最終鑰匙](docs/img/final_key_obtained.png)

![Ebiten：提頓夜間 tier3 牢門，使用最終鑰匙前](docs/img/teidon_final_key_door_closed.png)

![Ebiten：由正式道具選單使用最終鑰匙後，牢門開啟且鑰匙保留](docs/img/teidon_final_key_door_open.png)

![Ebiten：提頓夜間 handler35 綠色寶珠對話](docs/img/teidon_green_orb_dialogue.png)

![Ebiten：朗錫爾勇氣試煉神父詢問是否單獨挑戰](docs/img/lancel_courage_trial_prompt.png)

![Ebiten：勇氣洞窟正式調查寶箱取得藍寶珠](docs/img/courage_cave_blue_orb_obtained.png)

![Ebiten：推開海盜村入口物件後露出地下密道](docs/img/pirates_red_orb_hidden_entrance.png)

![Ebiten：地下藏寶室正式調查取得紅寶珠](docs/img/pirates_red_orb_obtained.png)

![Ebiten：商人城正式交付商人的首次確認](docs/img/merchant_settlement_first_offer.png)

![Ebiten：CTY59 handler41 Y≠4 的原始對話分支](dq3_remake_ebitan/docs/img/merchant_settlement_handler41_dialogue.png)

![Ebiten：CTY59 handler41 Y=4 的兩項道具店分支](dq3_remake_ebitan/docs/img/merchant_settlement_handler41_shop.png)

![Ebiten：四人縱列與四欄 H/M/等級隊伍 HUD（pack 驅動 E2）](dq3_remake_ebitan/docs/img/party_field_hud.png)

![Ebiten：標題閒置後的戰士職業 attract 卡（pack 驅動 E2）](dq3_remake_ebitan/docs/img/title_attract_warrior.png)

![Ebiten：新遊戲能力確認畫面的 runtime 參考圖（固定 checkpoint 的 V3 靜態證據另見 docs/126）](dq3_remake_ebitan/docs/ng_confirm.png)

![Ebiten：革命後調查建城者座位後方取得黃寶珠](docs/img/merchant_revolution_yellow_orb_obtained.png)

![Ebiten：沙曼歐莎夜間使用拉之鏡揭露假王](dq3_remake_ebitan/docs/samanosa_mirror_reveal.png)

![Ebiten：拉之鏡事件進入怪力魔戰](dq3_remake_ebitan/docs/samanosa_boss_troll.png)

![Ebiten：船員之骨定位後航行至動態幽靈船（V1，待原版同狀態對拍）](docs/img/ghost_ship_overworld.png)

![Ebiten：幽靈船內正式調查取得愛的回憶（V1，待原版同狀態對拍）](docs/img/ghost_ship_loves_memory_chest.png)

![Ebiten：愛的回憶解除奧莉薇亞海岬詛咒（V1，待原版同狀態對拍）](docs/img/olivia_cape_memory_reunion.png)

![Ebiten：牢獄祠堂正式調查取得蓋亞之劍（V1，待原版同狀態對拍）](docs/img/gaia_sword_obtained.png)

![Ebiten：正式新遊戲流程抵達 THE END（TIT3.P 終盤片尾，V3）](dq3_remake_ebitan/docs/img/ending_the_end_runtime.png)

### `dq3_remake/`：C99 + SDL2 參考實作

此目錄保存較早的現代化 C prototype、parser、測試與實驗性流程，可協助理解資料格式與既有
機制。由於它也包含早期近似值，移植設定資料時會再以原版證據核對。

### `re/`：反組譯筆記與部分 C 重建

此目錄保存對 `DQ3.EXE` 的函式分析、資料結構與部分 C 表達，也包含可重組原位指令的工具鏈
及 byte-match PoC。已驗證的範圍請查 [`docs/17`](docs/17-build-toolchain.md)、
[`docs/19`](docs/19-re-correctness.md) 與 [`docs/25`](docs/25-match-progress.md)。

## 目前的完成里程碑

全程版本以以下項目作為驗收方向：

1. 從全新遊戲開始，以正式玩家輸入抵達 THE END。
2. 主要事件入口、順序、gate、勝敗分支、道具與旗標交易對齊精訊版。
3. 關鍵 UI、角色／NPC、地圖、日夜、轉場、音樂及音效有原版畫面或資料佐證。
4. 關鍵事件前後可正常存檔、讀檔並繼續。
5. Linux AppImage、Windows ZIP 與 macOS ZIP 共用同一套 Go game core。

截至 2026-08-12，E2 targeted 驗收已涵蓋 pack raw→loader→battle／event→save/load；IDA
閉合的四條 story-flag transaction（handler74／69／70／33）已遷入
`events.json.story_flag_runtime_events` 並由正式 command／battle path 接線，
並以既有 PNG 作畫面證據；CTY07→CTY08 與 CTY13 金字塔 route 已依原始 transition table、
正式道具／戰鬥輸入修正，CTY13 checkpoint 已可取得魔法鑰匙並保存／讀回；完整
`TestOpeningProductionInputTrace` 已於本修正後由新遊戲重播至 `THE END`，因此主線達目前 E3 gate。第 3 項仍有 V1／V2 與 V3 長尾，
boss repeat-N、逐動作 frame、PCM wall-clock 與可見 palette transition 沒有可安全填入的
靜態 production 值，仍是 V3／unknown 限制；formation 的 raw EGA position／stride 已納入
E2 pack／renderer。日邦格兩場必經 Boss 的固定背景則已按原版 archive page／palette bank
達 near-state V2；怪力魔、巴拉摩斯與索瑪三連戰已由相同 raw fixed-record selector 接入 runtime
V1。通用地形背景沒有藉此宣稱 parity。完整靜態收斂與停止條件見
[`docs/123`](docs/123-static-battle-daynight-re.md) 與 [`docs/129`](docs/129-required-boss-backgrounds-re.md)。三種桌面發佈包與推廣片的雜湊／界線見
[`docs/114`](docs/114-release-artifacts.md)；後續工作以 [`docs/74`](docs/74-ebiten-remake-completion-plan.md)
的 E2 handoff 與未決清單為準。

## 證據優先序

1. DOSBox 同狀態、同輸入的原版實機結果，以及本機完整通關影片的玩家可見結果。
2. 原版 `DQ3.EXE`、CTY、D3TXT、ITEM、怪物及其他原始資料。
3. 完整影片用於畫面、操作順序與路線；修改過的角色數值不作參數 oracle。
4. 當年 BBS 與攻略，用於找路線及歷史交叉佐證。
5. C/SDL remake 與較早文件，用於提供實作線索。

反組譯紀錄採用 `writer → table/state → consumer → 玩家可見副作用`，位址同時註明
`file offset`、`logical` 或 `DGROUP`，以便後續重現。

## 建置與測試

原版素材不納入 repo。請把合法持有的檔案放在 `assets_raw/`。

```bash
bash dq3_remake_ebitan/build.sh
```

Android 保存優先 UX 與建置界線見 [`docs/124`](docs/124-android-ux-spec.md)。本輪 Docker
已用最新 HUD／Boss 背景 source 重建 debug APK（含本機合法素材、ignored），SHA-256 為
`38c222238a9dab068e37c5bf86eb3bad7c561f7104e32797c1a8e7c153a5e8ad`；這只代表
`build-only`。專用 emulator image 已固定 Android emulator 37.1.11 與 API 34 revision 14；
KVM 下 Android 14 與 APK 安裝均成功，但 `MainActivity.applyGameChrome()` 在 content view
建立前取得 `WindowInsetsController` 而崩潰，故觸控、存讀檔與生命週期尚未驗收。證據與重現
方式見 [`docs/132`](docs/132-android-emulator-validation.md)。

桌面執行方式與 Android 建置需求見
[`dq3_remake_ebitan/README.md`](dq3_remake_ebitan/README.md)。

圖形測試可在 Xvfb 下編譯 `game.test` 後，從 `dq3_remake_ebitan/game/` 執行，讓
`../../assets_raw` 指向正確素材位置。若素材缺失而出現 `SKIP`，表示該項尚未實際執行。

## 主要資料與工具

| 路徑 | 用途 |
|---|---|
| `dq3_remake_ebitan/` | 現行 Go／Ebiten remake |
| `dq3_remake/` | C99／SDL2 歷史參考實作 |
| `re/` | 反組譯註記與部分 C 重建 |
| `docs/` | RE 證據、格式、流程與歷史文件 |
| `tools/` | 抽取、渲染、反組譯與 DOSBox 驗證工具 |
| `dq3_real_video/` | 本機原版實況（不納入 Git） |
| `assets_raw/` | 原版素材（不納入 Git） |

專案內可使用：

- `tools/dis.sh`：16-bit Capstone 線性反組譯。
- `tools/dosbox_*.sh`：原版實機重播與截圖。

## 已確認的研究成果

以下項目已有直接研究證據：

- 16×16 繁中文字模、D3TXT 記錄、CTY section、BLK tile、地表與怪物等格式已可解析。
- 原版為 16-bit real-mode DOS 程式；編譯器與函式指紋研究見 `docs/19`。
- 原版 bug、當年 patch 與 BBS 記錄已有多項交叉驗證。
- 場景、戰鬥、NPC、遭遇、咒文與事件 runner 的多個資料表及 consumer 已被定位。

素材範例：

![原版世界地圖資料還原](docs/maps/world_con.png)

![原版中文字庫 atlas](docs/fonts/D3TXT00.FON.atlas.png)

## 版權

原始遊戲程式、文字、音樂與美術版權屬其權利人，不納入本 repo。研究與 remake 執行需由
使用者自行提供合法持有的原版素材。歷史敘述會標示資料來源，尚待一手資料確認的內容則保留
為研究線索。
