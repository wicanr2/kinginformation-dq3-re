# 精訊版《勇者鬥惡龍 III》反組譯與 remake

本專案研究精訊資訊在 1990 年代製作的中文版 DQ3
（程式內題名 *Dragon Fighter III／傳說的終章*），並以原版 DOS 程式與資料為證據，
製作可在現代平台執行的 Go／Ebiten remake。

截至 2026-08-10，`dq3_remake_ebitan/` 已由一次性 Docker＋Xvfb 的正式
`InputState` trace 從全新遊戲抵達 `THE END`，沿途核對設定資料、事件副作用、戰鬥與存讀檔。
主線 campaign 已達 E3；逐畫面／音效 V3 的長尾仍待後續對拍，本輪指定的 AppImage、Windows ZIP
與 macOS ZIP 已產出。Android／WASM 不屬本輪 release 目標。現行進度與工作順序以
[`docs/74-ebiten-remake-completion-plan.md`](docs/74-ebiten-remake-completion-plan.md) 為準。
本輪另以 IDA／原始資料閉合 CTY59 `handler41` 的 Y=4 設施分支；它已達 D3／E2／V1，
不代表 CTY60／61 階段或全程畫面 parity 已完成。
標題閒置後的八張職業 attract 也已由 game-pack 接線，順序／PCX 達 D3、輪播 timing 達 E2；
能力條逐幀填充與淡入淡出仍待 V3。
新遊戲能力確認畫面的 `FIRST.SCR` raw 背景也已由 game-pack 載入，原始檔尺寸／解碼與配色維持 D2，runtime 背景達 V2；
stat panel 幾何仍待同狀態 V3。
地表詳細狀況窗則已依 `DGROUP 0x3DA8`／D3TXT00 record 407 資料化，runtime 達 E2／V2；
狀況子選單、隊員詳情與道具／裝備逐窗仍待同狀態 V3，詳見 [`docs/116`](docs/116-field-status-panel-re.md)。

## 先讀這些

- [`PROJECT_MEMORY.md`](PROJECT_MEMORY.md)：接手時應保留的少量穩定決策。
- [`CONTEXT.md`](CONTEXT.md)：術語、位址口徑與文件索引。
- [`docs/74`](docs/74-ebiten-remake-completion-plan.md)：Go／Ebiten remake 的現行完成計畫。
- [`docs/69`](docs/69-why-remake-looked-done-but-wasnt.md)：過往實作經驗與後續改善方法。
- [`docs/00`](docs/00-re-methodology.md)：反組譯方法、證據分級與常見陷阱。

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

目前可見成果（以下為現行 Ebitengine runtime 證據；未另標示者是 V2，不代表已完成原版
同狀態逐像素 V3）。本輪新增由正式新遊戲輸入 trace 產生的終盤固定片尾畫面與配樂 cue；
其餘場景仍依各圖的 V1/V2 標註，不能把單一終盤畫格擴大解讀成全程 V3。終盤證據／剩餘
畫面工作見 [`docs/106`](docs/106-gaia-sword-volcano-re-worklist.md)。

README 內的戰鬥圖已在 2026-08-09 以 Docker 正式 renderer 重新產出，並使用預設
`combat_info=0`。敵人上方的 `H/HP` 是可選的 remake 診斷資訊，不是精訊版原版畫面，
因此不應出現在這些對拍圖；玩家可見的戰鬥文字也必須由框線／面板 renderer 佈局，若文字
落在框外即視為尚未通過 V3 對拍。

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

![Ebiten：日邦格八頭大蛇第一戰（原版怪物遮罩、背景與框內戰鬥文字）](docs/img/jipang_orochi_first_battle.png)

![Ebiten：第一戰後依原版移動至宮殿並顯示 rec70](docs/img/jipang_orochi_first_post.png)

![Ebiten：拒絕無姬後的第二戰（原版掉落抑制、無非原版怪物 HP 飄字）](docs/img/jipang_orochi_second_battle.png)

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

![Ebiten：新遊戲能力確認畫面的 FIRST.SCR 人物背景（pack 驅動 D2／V2）](dq3_remake_ebitan/docs/ng_confirm.png)

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

截至 2026-08-10，`TestOpeningProductionInputTrace` 已完成上述第 1、2、4 項的主線 E3
驗收；終盤固定片尾 `TIT3.P` 與 ENDING 配樂 cue 已由同一條正式 trace 產生 runtime V3
證據，第 3 項其餘場景仍以 runtime V1／部分 V2 為主；第 5 項指定桌面 release 已完成。
編譯、單元測試、事件切片與畫面 dump 分別提供不同層次的信心；三種桌面發佈包與
推廣片的雜湊／界線見 [`docs/114`](docs/114-release-artifacts.md)。主線已由正式玩家
流程串起，後續集中於其餘 V3 對拍、戰鬥／場景音效長尾與發佈驗收。

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
