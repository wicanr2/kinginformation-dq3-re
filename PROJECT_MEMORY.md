# DQ3 Go/Ebiten remake 接手記憶

> 更新：2026-08-09。接手先讀 `CLAUDE.md`、`CONTEXT.md`，再讀
> [`docs/74-ebiten-remake-completion-plan.md`](docs/74-ebiten-remake-completion-plan.md)。
> 本檔只保存不易過期的決策；逐項狀態不要在此重複維護。

## 目標與完成定義

現行產品是 `dq3_remake_ebitan/`。玩家須由全新遊戲開始，只用正式遊戲輸入，依精訊版流程抵達
THE END；關鍵事件的入口、設定資料、畫面、聲音、副作用與 save/load 均有原版證據。

以下都不算完成：編譯成功、單元測試全綠、直接呼叫內部事件函式、用 debug 鍵/env 串場景，
或只證明 parser/handler 存在。

## 證據與產品線

- 本機原版影片／DOSBox：玩家可見流程與操作 oracle。
- IDA、原始 EXE/CTY/文字/資產：caller、設定值與副作用 oracle。
- 攻略：路線與定位線索，不代替原始資料。
- `dq3_remake/` C/SDL 版：機制與 parser 參考，不是原版 oracle。
- 舊 RE/進度 Markdown：線索；若與近期證據或 code 衝突，以證據重查。

失敗原因、事件契約與驗收紀律集中在
[`docs/69-why-remake-looked-done-but-wasnt.md`](docs/69-why-remake-looked-done-but-wasnt.md)。

## 現況入口

- 唯一 current plan：[`docs/74-ebiten-remake-completion-plan.md`](docs/74-ebiten-remake-completion-plan.md)
- 共用 game-pack JSON 契約：
  [`docs/84`](docs/84-game-pack-json-contract.md)
- 原版流程 oracle：[`docs/66-original-flow-oracle.md`](docs/66-original-flow-oracle.md)
- 最新 production audit：[`docs/89`](docs/89-baharata-rescue-production-trace.md)
- 最新取船／航行 audit：[`docs/90`](docs/90-portoga-ship-production-trace.md)
- 最新加爾那之塔／睡眠恢復 audit：
  [`docs/91`](docs/91-garuna-satori-book-production-trace.md)
- 最新達瑪轉職 audit：
  [`docs/92`](docs/92-dhama-reclass-production-trace.md)
- 最新提頓《黑暗之燈》audit：
  [`docs/93`](docs/93-teidon-dark-lamp-production-trace.md)
- 最新日邦格八頭大蛇兩階段 audit：
  [`docs/95`](docs/95-jipang-orochi-production-trace.md)
- 最新隱形草／愛丁貝亞守衛 audit：
  [`docs/96`](docs/96-invisibility-grass-eginbear-production-trace.md)
- 最新愛丁貝亞推石／乾渴壺 audit：
  [`docs/97`](docs/97-eginbear-push-puzzle-production-trace.md)
- 最新乾渴壺使用／最終鑰匙 audit：
  [`docs/98`](docs/98-thirsty-pitcher-final-key-production-trace.md)
- 最新蘭西爾勇氣試煉／藍寶珠 audit：
  [`docs/100`](docs/100-lancel-courage-blue-orb-production-trace.md)
- 最新海盜村密道／紅寶珠 audit：
  [`docs/101`](docs/101-pirates-red-orb-production-trace.md)
- 香巴尼塔甘達特原版事件與正常路徑：[`docs/85`](docs/85-shanpane-kandar-production-trace.md)
- 近期 IDA/影片證據：[`docs/75`](docs/75-phoenix-orbs-re.md)、
  [`docs/76a`](docs/76-baramos-gaia-re.md)、[`docs/76b`](docs/76-r5-endgame-realignment.md)、
  [`docs/77`](docs/77-r5b-castle-aftermath.md)

boot 起的正式 trace 已由諾魯德 checkpoint 繼續閉合巴哈拉達救援：正常補給與練級後
步行至 CTY14，開原版魔法門，依序完成四守衛、handler59／60 俘虜移動、甘達特
`monster56×1 + monster57×3`、求饒選項、返回 CTY15 與古布達交談取得黑胡椒 `0x5c`，
並通過 save/load。事件旗標、編隊、NPC／trigger selector、movement、文字與黑胡椒交易
均由 game pack 提供；詳見 `docs/89`。其後已由正式標題讀檔、旅店補給、
魯拉、王城交談完成黑胡椒換船，驗證原版旗標交易、停泊 `(25,73,L0)`、取船後
save/load、正式登船與首次航行；詳見 `docs/90`。schema `0.1.6` 已把 CTY18 sec1
`{type=1,item=0x4a,flag=0xb7}` 遷入 `treasure_events`；正式 trace 已處理航海遭遇、
繞完加爾那之塔不同連通區、取得《領悟之書》、立即切換開箱 tile 並通過 save/load。
同批以 IDA 證明玩家睡眠／混亂每回合 `roll<=100`、怪物側 `roll<100` 可醒，修除怪物
43 令全隊永久不能行動的死循環；詳見 `docs/91`。其後已正常離塔、航行至 CTY17、
正式練所選同伴至 Lv20、用魔法鑰匙入殿、經 rec421 把領悟之書交給該同伴，並由
handler39 完成賢者轉職與 save/load；詳見 `docs/92`。其後已從該合法 checkpoint
航行至 CTY20，於白天由命令窗調查取得《黑暗之燈》`0x5f`，在地表由 rec421 正式使用，
依原版 gate 進入黑夜且不消耗道具；save/load 後再航行抵達 CTY19，詳見 `docs/93`。
同一條正式 trace 又由 CTY19 正式入口完成 handler45 第一戰、怪75原始掉落草薙大劍、
NPC slot0 與玩家各自的右移兩步、CTY21 rec70、handler36 No 分支第二戰、掉落抑制及
同格寶箱取得紫寶珠，並通過 save/load；詳見 `docs/95`。同一條 trace 已再由 CTY21
正常補給、從 CTY38 原始貨架買入隱形草 `0x5d`，在 CTY39 由正式道具選單使用，使
handler29 不再把守衛橫移到玩家同欄，從右欄進入 CTY76 並通過 save/load；詳見
`docs/96`。其後已由 119 步正式方向輸入完成 CTY76 三石解謎；原版以 NPC runtime
`ctrl bit0x40` 判定可推，handler30 固定檢查 slot0–2 的 Y=5，成功 clear flag `0x3b`
並開 passage。再由命令窗調查 event0 取得乾渴壺 `0x5e` 並通過 save/load；詳見
`docs/97`。同一條 trace 已再由正式船路抵達 `(146,53)`，使用乾渴壺套用原版
world-state bit `0x08`／`5×4` 地表 patch，進 CTY40 調查取得最終鑰匙並完成 save/load；
詳見 `docs/98`。同一條 boot trace 已再由正式魯拉回 CTY15；原版 handler 同步把船移到
`(104,126)`，玩家登船航行至提頓入口外使用黑暗之燈，再從 rec421 明確使用最終鑰匙開
tier3 牢門。夜間 handler35 把綠色寶珠交給第一個有空格的隊員並通過 save/load；詳見
`docs/99`。同一條 boot trace 又已航行到蘭西爾、住宿切白天、用最終鑰匙進神殿，經
handler37 正式接受試煉使四人暫時變單人，走 CTY23 原始轉場取得藍寶珠，在試煉途中
save/load，再由 CTY75 handler62 原樣復隊並完成第二次 save/load；詳見 `docs/100`。
同一條 trace 已再由正式魯拉至 CTY15 觸發船隻重定位，登船航行至 CTY27，推開具
`ctrl bit0x40` 的入口物件，走密道取得紅寶珠 `0x68`，完成存讀檔並驗證 flag `0x3f`
同時控制寶箱與入口物件 visibility；詳見 `docs/101`。現行 pack manifest 為 schema
`0.1.22`、`dq3_cht` content `0.1.27`；不要把本段早期的版本號當成現行值。IDA Pro 9.4 證明商人聚落 `(210,64)`
的原版入口順序為 `flag0x23 clear→CTY58`，否則依序測 `0x47 set→59`、
`0x42 set→60`、`0x48 set→61`、最後 CTY83；舊 C／Go「所有 set 即取」表已推翻。
入口已由 game-pack JSON 載入；三個 stage gate 已閉合到最終鑰匙 `0x47`、假王事件
`0x42`、蓋亞之劍 `0x48`。商人交付已從紅寶珠 checkpoint 以正式酒場與船陸輸入
閉合，含兩次確認、滿載逐件訊息、共用預存所、建城者身份及 save/load。
IDA 同時訂正地表 seam（X `242/2`、Y `204/2`）與船 mode1 的 attr bit1/bit0 分支。
CTY83 黃寶珠原始 event `01 6a 00 4a` 已遷入 JSON；舊 Go table 把 present flag set
誤作「已取得」而會卡死來源，已移除。handler48 的 record42／43、flag0x4a 分支、保存的
建城者姓名、男性 flag0x15／女性 flag0x23 與 save/load 已達 component E2／V1。同一條
boot trace 已由商人交付自然重進 CTY60，並正式航行到 CTY41、走 CTY42 原始出口
`(213,123)`、穿過 CTY43 關卡進 CTY44，再由 CTY24 取得拉之鏡並完成 save/load。
同一條 trace 已使用黑暗燈，在 CTY44 sec1 `(14,7)` 由道具選單使用拉之鏡，閉合
record97／98、monster89、變身杖 `0x62` 與勝利旗標交易，故此段升為 campaign E3／V1。
IDA 9.4 證實離城 consumer 直接使用 transition X/Y、兩者皆零才回退 remembered
coordinates，不讀 facing；舊 remake 額外推出兩格的近似已移除。後續狀態以本檔下方的
`0.1.22` 段落與 `docs/74` 為準，不沿用本段早期的剩餘工作清單。
2026-08-09 已用一次性 Docker+Xvfb 通過 `TestOpeningProductionInputTrace`：從新遊戲、正式
InputState、不中途注入座標／道具／事件函式，依序完成 CTY83 黃寶珠、六珠／拉米亞、巴拉摩斯、
自然下降、愛列夫加特、蓋亞之劍火山、尼羅肯特銀寶珠、索瑪城奧爾特加事件、封咒三連戰、
光之珠弱化、王城冊封與 `THE END`。中途的雲雨之杖、彩虹水滴、蓋亞之劍、銀寶珠及關鍵戰前
均有 save/load round-trip；主線流程目前是 campaign E3／runtime V1。地表水面顏色、四輪
palette transition、逐畫面／音效 V3、Android／桌面 release 驗收仍是後續工作，不能與流程
完成混稱。

同日收尾切片已把終盤資產與配樂接回 versioned pack：`manifest.json` 以 SHA-256／尺寸宣告
`title_image=TITG.P`、`ending_image=TIT3.P`，`data/audio.json` 以穩定 scene cue 提供
TITLE/FIELD/CASTLE/TOWN/DUNGEON/BATTLE/ENDING（ENDING 為原版 track 17）。loader 會在 Docker runtime 讀取並 fail closed；正式 trace 在冊封後捲完
`ENDTXT.TXT` 產生 `dq3_remake_ebitan/docs/img/ending_the_end_runtime.png`，與
`docs/title/ending.png` 逐像素一致。這只把固定終盤畫面與 cue 升為 V3；終盤文字 timing、
其餘場景／戰鬥畫面與跨平台 release 仍未完成，不得將單一終盤圖誤宣稱全程 parity。

同日已補閉合 CTY59 `handler41`：IDA Pro 9.4 的原始分支在玩家 Y=4 呼叫 section
facility index 0，否則選 D3TXT07 record8；兩條路徑均由 `conditional_facility_events`
JSON 提供。`TestConditionalFacilityOriginalCoordinateGate` 以原版建城後 NPC 可見旗標，
從正式 `cmdTalk` 驗證對話／兩項道具店，`internal/gamepack` parity test 鎖定 EXE bytes
與 record8 glyph，並產出 [對話狀態圖](dq3_remake_ebitan/docs/img/merchant_settlement_handler41_dialogue.png)
與 [店鋪狀態圖](dq3_remake_ebitan/docs/img/merchant_settlement_handler41_shop.png)（V1）。
這只關閉 handler41 的事件分支；CTY60／61 階段與全程 V2/V3 仍以 `docs/74`、
`docs/102` 的未閉合項為準；CTY83 自然路線已由 boot production trace 閉合。

遊戲設定將逐批移至 versioned JSON game pack，長期讓同一 Go／Ebitengine core 支援
精訊版 DQ1／DQ2／DQ3。原始 DAT／EXE decoder 必須保留為 parity oracle；JSON 值仍需
D2/D3 證據，未知值不得以可調參數名義猜填。Desktop 可讀外部 pack 免重編譯，
Android／Web 嵌入 pack 則仍需重包。

2026-07-29 起的硬規則：版本專屬 CTY/section/座標、subid、story flag、dialogue record、
formation、寶箱 gate、價格與 cue 不得新增成 Go 常數/table。甘達特事件已是第一個完整
`events.json` + `texts.json` + 通用 `boss_surrender` primitive + EXE/CTY/D3TXT parity
範例；羅馬利亞 `temporary_role` 是第二個完整範例；諾亞尼爾 `quest_item_chain` 是
第三個；金字塔 `two_step_floor_switch_gate` 是第四個；波魯多加
`staged_vehicle_exchange` 是第五個，NPC selector、原版旗標、任務／交換道具、船停泊
座標與五段文字均由 JSON 提供；諾魯德 `guided_passage` 是第六個，互動格、scene
trigger、兩段 NPC 路徑與最終朝向、旗標、道具及四段文字均由 JSON 提供；巴哈拉達
`hostage_rescue` 是第七個，四守衛、兩個 scene trigger、三段 NPC movement、兩組
formation、完成旗標、古布達交易與全部玩家文字均由 JSON 提供。
獨立一次性寶箱另使用 `treasure_events`；加爾那之塔《領悟之書》是第一個從舊 Go
treasure table 遷出的完整 D3／E3 範例。
達瑪 `reclass` 是第八個有限 primitive；NPC selector、Lv20、職業清單、賢者 gate、
個人物品八格、文字及 effect ID 均由 JSON 提供。Go 只保留通用狀態機；轉職保留既有
咒文、VIT 不減半、全卸裝，且領悟之書只從所選角色移除。
schema `0.1.8` 的 `item_use_effects` 是道具效果資料化的第一個有限契約；提頓寶箱與
黑暗燈 raw ID、旗標、適用場景、目標日夜 phase、步數重設及不消耗語意均由 JSON 提供，
Go 不知道《黑暗之燈》或 DQ3 專屬 raw ID。
schema `0.1.9` 新增 `interface.json`；原版共用對話框 `(152,238,360,112)`、文字 inset、
20 欄／4 行已由 EXE rect 的 caller、writer 與 consumer 證實並移出 Go。怪物透明度也已
改由 `DQ3MNS.SHP` 四色 plane 後的獨立 RLE AND-mask 解碼，禁止再把黑色色號 0 當透明。
證據見 `docs/94`；受影響的 production 截圖已重產並目視核對。
schema `0.1.10` 新增 `staged_boss_events`；日邦格兩個 NPC selector、兩組編隊、NPC／玩家
戰間路徑、掉落政策、旗標交易、寶箱 gate 與 rec66–70 均由 JSON 提供。怪物 parser 已校正
`+0x25` 掉落閾值、`+0x26` 掉落道具，`+0x27` 保持未知；詳見 `docs/95`。
schema `0.1.11` 新增 `item_use_effects.step_count`、`temporary_invisibility` 與
`tracking_guard_events`；隱形草的 raw ID、消耗、25 步 timer、CTY39 trigger、guard
selector 與 bypass effect 均由 JSON 提供。CTY parser 也已修正無 event table 時漏讀
special handler 的錯誤；詳見 `docs/96`。
schema `0.1.12` 新增全域 `npc_push_rule` 與 `push_puzzle_events`，詳見 `docs/97`。
schema `0.1.13` 新增 `reveal_world_map_patch`；乾渴壺的船／座標／旗標 gate、world-state、
完整 `5×4` tile patch 與原始動畫參數均由 JSON 提供，CTY40 最終鑰匙也遷入
`treasure_events`；詳見 `docs/98`。
schema `0.1.14` 新增 `rura_navigation`、鑰匙 `door_key_tier` 與
`npc_item_reward_events`；魯拉 CTY／文字 record／arrival offset／船落點、三把鑰匙 tier、
提頓 NPC selector／綠色寶珠 flag／文字均由 JSON 提供，並有 EXE／CTY／D3TXT parity
與 boot production trace；詳見 `docs/99`。
schema `0.1.15` 新增 `temporary_solo_challenges`；入口／返回 NPC、完成旗標 reader、mode
mask、世界／場景落點、洞窟訊息與所有文字均由 JSON 提供。原版 handler37 不寫 flag0x13，
其 writer 維持 unknown；試煉途中完整離隊角色 records 會進存檔，handler62 後按原序復隊。
CTY23 藍寶珠亦已遷入 `treasure_events`；詳見 `docs/100`。
schema `0.1.16` 新增 `world_entrance_variants`；同一 world coordinate 的 ordered
flag branches、每支極性、layer、目的 CTY 與 default 全由 JSON 提供。商人聚落舊表把
第一支 clear gate 讀成 set gate，已由 IDA linear `0x125fe..0x12681` 推翻；詳見 `docs/102`。
人物初始裝備由 `characters.json` 提供。細則見 `AGENTS.md` 與 `docs/84`。

schema `0.1.20` 延續 `choice_item_exchange_events`，並新增 `tracked_world_objects`。
IDA Pro 9.4 已推翻 CTY54 舊
`scriptedTable` 的立即交換近似：handler44 有無杖詢問與持杖交換兩組 Yes／No；成功會
原格把 `0x62` 換成 `0x63`、啟用座標為 `(150,90)` 的世界物件、OR world-state `0x04`、
clear flag `0x43`，再依序顯示 records60／61。七段 D3TXT06 文字與全部設定已遷入 JSON；
boot production trace 已從假王勝利正常住宿、魯拉、登船航行到 CTY54，驗證拒絕、接受與
save/load。幽靈船座標已與玩家座標分離；四個 tile 由 RNG state 低二位選擇，物件離開
`80×80` 活動視窗時依原版 18 筆候選表重定位。船員之骨 records740–744、同座標動態入口、
CTY36 轉場、愛的回憶 `0x64` 寶箱及其 save/load 已由正式 InputState trace 閉合為 E3／V1。
schema `0.1.22`／content `0.1.27` 延續 `coordinate_item_gate_events`，並新增
`direct_map_patch` 與 persistent `map_patch` 的分離契約。boot trace 從
愛的回憶 checkpoint 依 CTY36 原始轉場離船；同座標停泊船會恢復 mode1，避免玩家困在
水格。其後正式航行至 `(76,54)`，依 records597／598、item0x64 與 flag0x35 解除詛咒，
再進 CTY55 調查取得蓋亞之劍0x0f並完成 save/load。CTY55 event0 已從 Go treasure table
遷入 JSON；見 `docs/105`。2026-08-09 已完成蓋亞之劍火山切片：IDA 已確認特殊 item table
`DGROUP 0x3666/0x3668` 將 item0x0f 派發至 `(logical) 0x3d65`，只接受 `(65,109)`，成功
執行動畫、set 原版 world-state bit0x20；`DGROUP 0x3b96` 的 direct viewport table 與
`DGROUP 0x3a54` 的 persistent reload table 已分開存入 JSON，save/load 只重建後者。正式
boot trace 已走 CTY56 南洞口的 sec1→2→3→2→3→sec0 北出口，進 CTY64 由 handler49
取得銀寶珠，並通過 party-order inventory、save/load、重複對話。現行「任意地表 set
remake flag0x32」已移除；完整 bytes／consumer／位址證據與路徑見 `docs/106`。同一條 boot
trace 隨後已通過 CTY83、P4–P6 與 new game→THE END；後續只處理 V3 對拍、音效／音樂長尾、
跨平台 release 與受影響的畫面更新。

2026-08-09 另完成第一個高可見度玩家畫面切片：原版影片 `f000295`／`f000300`／`f000900`
證實四人縱列與四欄 H/M/等級 HUD；Go 現在以 `characters.json.party_sprite` 解析
`DQ3MST.BLS` 的 class/gender entry，透過 8 格移動前位置歷史繪製活人與隊尾棺材，並由
`interface.json.party_hud` 提供 `(48,244,448,80)` 幾何與 label glyph。正式 renderer 對拍圖為
`dq3_remake_ebitan/docs/img/party_field_hud.png`，component test 鎖定 pack 對映、trail 與
真實資產；目前達 E2／runtime 圖證，尚未宣稱同地圖同晝夜 V3。進入新場景、傳送、讀檔、
加入／離隊及單人試煉復隊均會重置 trail，避免跨場景瞬移。其餘 battle 長尾、母親／attract、
商店／教會／旅社逐窗、日夜 palette、BGM／SFX 及 Android／桌面 release 仍是未閉合工作。

## 固定工程方法

1. 每輪閉合一個垂直切片：正式入口 → 狀態交易 → 畫面/聲音 → 下一節點 → save/load。
2. 未知值 fail-closed；禁止用 C remake、經典 DQ3 慣例或「合理值」填 production。
3. 每個參數追原版「誰在何時寫入」，並記錄明確位址口徑。
4. component test 與 production-input trace 分級；只有正常玩家可達才可關閉主線 GAP。
5. 進度只維護 `docs/74`；歷史計畫保留短索引，細節從 Git 歷史取回。

## 工作樹注意

接手先執行 `git status --short`。未追蹤的 Android libs、scratchpad 與 `tools/dosbox_verify_*`、
`tools/tmp_*` 可能是使用者檔案；不得刪除、覆寫或誤納入 commit。
