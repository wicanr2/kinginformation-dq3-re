# DQ3 Go/Ebiten remake 接手記憶

> 更新：2026-08-10。接手先讀 `CLAUDE.md`、`CONTEXT.md`，再讀
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
`0.1.27`、`dq3_cht` content `0.1.32`；不要把本段早期的版本號當成現行值。IDA Pro 9.4 證明商人聚落 `(210,64)`
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
下方最新版本段落與 `docs/74` 為準，不沿用本段早期的剩餘工作清單。
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
schema `0.1.9` 新增 `interface.json`；原版共用對話框可見尺寸為 `(152,238,352,96)`、文字 inset、
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
加入／離隊及單人試煉復隊均會重置 trail，避免跨場景瞬移。其餘 battle 長尾與母親逐格護送仍是
未閉合工作；同日也把標題 idle attract 接回 versioned pack：`interface.json.attract` 宣告
`TITH.P`→`TITO.P` 八張卡、1200 frame delay／hold，manifest 保存每張 PCX 的尺寸／SHA-256，
輸入會中斷輪播並回主選單；`title_attract_warrior.png` 是 runtime 對拍。PCX／順序達 D3/E2，
能力條逐幀填充與淡入淡出仍待 V3。商店／教會／旅社逐窗、日夜 palette、BGM／SFX 及
Android／桌面 release 仍是未閉合工作。

2026-08-09 另一個高可見度切片已完成：原版 `FIRST.SCR`（112000 bytes）經
`DecodeRowInterleaved4BPP` 以每列 plane0..3 的 row-interleaved 格式解碼，專用 16 色
palette 與 `FIRST.SCR` asset key 已移入 `interface.json.new_game_confirmation`，manifest
保存尺寸／SHA-256。正式 `ngConfirm` renderer 先畫 pack 背景再疊能力 stat panel；runtime
圖 `dq3_remake_ebitan/docs/ng_confirm.png` 已更新，原始檔尺寸／解碼與配色維持 D2，runtime 背景達 V2，panel 幾何仍待
同狀態 V3。game-pack schema 現為 `0.1.27`；不得把此前 `0.1.22`／`0.1.23`／`0.1.24`／`0.1.25` 歷史段落當成目前版本。

2026-08-09 戰鬥文字串接切片已完成：`interface.json.battle_texts` 將 25 個跨版本 battle role
指向 `texts.json` 的 D3TXT00 record；`battle.go` 直接消費原始 glyph/control stream，角色／怪物
名稱與數字以 runtime 變數插入，缺少對映或 glyph 時 `NewGameWithPack` fail closed。勝利、經驗、
金錢與全滅也已改回原始 record 順序；相關 role、record、推論等級與 consumer 見 `docs/108`，
契約同步於 `docs/84`。Docker＋Xvfb 的 `internal/...` 與 `game` 全套測試通過，並重產十張受影響
戰鬥 PNG；逃跑失敗未找到獨立原版句子，故不以「沒打中」猜代。這只完成戰鬥文字資料與 renderer，
逐動作 timing、動畫／音效、完整抗性／formation／掉落及跨平台 release 仍未閉合。

後續修正也已把戰鬥訊息外框資料化：`interface.json.battle_message` 以 DQ3.EXE DGROUP
`0x3e6e` 的 shared `win_rect` 提供可見 `(152,238,352,96)`、inset `(16,16)`、20 欄／4 行；
`sub_1fb36` 的 `(360,112)` 只是備份範圍，已由 `sub_1fd30` 反證。訊息階段不再沿用左下
150×108 指令框，也不畫空的右側敵名框。`docs/108`、`docs/109` 與
`dq3_remake_ebitan/docs/battle_message_queue.png` 是本輪證據；缺 geometry 時 production
`NewGameWithPack` fail closed。其餘戰鬥 timing／動畫／音效與 V3 長尾仍未完成。

同日以 IDA Pro 9.4 追閉戰鬥下方 HUD：`DS:0x4112`（file `0x1a252`）的指令框與
`byte_28f00`（file `0x1a270`）的敵名／數量框已進 `interface.json.battle_command`／
`battle_enemy`。可見幾何分別為 `(144,248,128,(menu_rows+2)*16)` 與
`(288,248,256,(active_groups+2)*16)`；`sub_1c1d8`、`sub_1f908`、`sub_1b053`、
`sub_1b101` 的 writer→consumer、raw bytes、IDA hash 與推論等級集中在 `docs/109`。
原始 style pattern、spell／target 其他 rect 及數字 writer 的 raw 內部轉換尚是 V3 長尾，不可從這批
資料宣稱整個戰鬥畫面已完成。受影響的 `battle_party_command.png`、`battle_ally_target.png`、
`battle_palpunte_menu.png`、`battle_enemy_target.png`、`battle_message_queue.png` 已重產。

## 固定工程方法

1. 每輪閉合一個垂直切片：正式入口 → 狀態交易 → 畫面/聲音 → 下一節點 → save/load。
2. 未知值 fail-closed；禁止用 C remake、經典 DQ3 慣例或「合理值」填 production。
3. 每個參數追原版「誰在何時寫入」，並記錄明確位址口徑。
4. component test 與 production-input trace 分級；只有正常玩家可達才可關閉主線 GAP。
5. 進度只維護 `docs/74`；歷史計畫保留短索引，細節從 Git 歷史取回。

## 工作樹注意

接手先執行 `git status --short`。未追蹤的 Android libs、scratchpad 與 `tools/dosbox_verify_*`、
`tools/tmp_*` 可能是使用者檔案；不得刪除、覆寫或誤納入 commit。

2026-08-09 戰鬥指令標籤資料化：`interface.json.battle_command_labels` 新增 `war`、
`flee`、`defend`、`item`、`spell` 五組原版 D3TXT00.FON 雙 glyph（D2，對映與欄距
證據見 `docs/109-battle-hud-rects-re.md`）。`battle.go` 已移除對應的版本專屬 table；
`NewGameWithPack` 會要求 labels 契約並在缺少時 fail closed，battle draw fixture 則明確
安裝內建 pack。這只完成資料邊界，不代表逐 label IDA writer、戰鬥框線樣式、逐動作
timing／動畫／音效、抗性／formation／掉落或 release 已完成。

2026-08-09 地表命令窗也已資料化：`interface.json.field_command_labels` 保存標題與六項
正式命令的雙 glyph，`cmdmenu.go` 不再含 DQ3 專屬字模 table；`NewGameWithPack` 缺欄位
即拒絕啟動。這是 D2，證據與限制見 `docs/110`；不代表 battle timing／音效／V3 或 release
已完成。當時 pack schema `0.1.27`、content `0.1.32`（後續欄位切片沿用此 schema 版本）。

2026-08-09 又完成戰鬥場景帶資料化：`interface.json.battle_scene` 保存 `field_y0=80`、
`field_y1=246`、`ground_y=232` 與 `cursor_glyph=0x77`；`battle.go` 不再持有這組
DQ3 專屬 renderer 常數，production 缺欄位即 fail closed。這是 D2 engine/data 邊界切片，
證據與限制見 [`docs/111`](docs/111-battle-scene-layout-re.md)；框線樣式、逐動作 timing／
動畫／音效、完整抗性／formation／掉落與跨平台 release 仍未閉合。

2026-08-09 開場／創角 glyph 也已移入 `interface.json.new_game_labels`：主選單、性別、能力
確認及英數姓名盤的特殊格均由 game-pack 提供，`NewGameWithPack` 缺欄位即 fail closed，
`NewGameFlow`、`NameInput`、`GenderSelect` 與酒館不再保存這批 DQ3 raw glyph table。這批
依 `D3TXT00.FON`（47232 bytes、SHA-256
`c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`）及開場實機畫面交叉
核對，為 D2；證據與完整 role 對映見 [`docs/112`](docs/112-newgame-labels-re.md)。面板／姓名盤
幾何已另由 `interface.json.new_game_geometry` 保存：`0x29088`／`0x290a6`／`0x290c4`／
`0x290e0` 與能力確認 `0x28b78`／`0x28b92`／`0x28bc6` 的 raw window 欄位、9×5／16px
姓名盤步距及 640×350 像素 rect 均由 IDA 9.4 writer／consumer 與 DOSBox 截圖交叉閉合；
整體仍標 D2，證據見 [`docs/113`](docs/113-newgame-geometry-re.md)。本輪已修正 production
創角階段切換到 `FIRST.SCR` 黑底立繪，並把 NewGameFlow／NameInput／GenderSelect 與酒館
改讀 pack geometry；功能列文字已接入 pack，框線 pattern、逐幀游標／閃爍與逐幀演出仍未
升為創角 V3。

2026-08-09 開場姓名流程再校正：IDA Pro 9.4 的 `sub_10e8e/sub_10f5b/sub_10dc8` 與
D3TXT00.TXT records 451–456 證實畫面是 9×5、45 個 raw cell；語意 cell 採
`cell=col*5+row`，注音／英數的 raw35（cell43、⊙）才進右側五列功能列，功能列第五列
才設完成旗標。`interface.json.new_game_labels` 現保存兩組 45 格、標題、雙箭頭、五列
功能文字與 rec456 組字提示；`NameInput` 已移除舊 38 格／直接 OK 近似，正式 production
trace 與 touch hit box 都走同一份 pack 資料。新增 parity／function-focus tests；完整
開場 trace 已在單一 Docker＋Xvfb 容器重跑通過，四張 runtime PNG 也已重新產出。geometry 的
`name_title=(233,46)`、`name_text=(249,62)` 來自 raw window／`sub_10e55`，若後續 DOSBox
同輸入畫面推翻，必須追加勘誤而不覆寫 `docs/113` 原始證據。

2026-08-09 README 戰鬥畫面勘誤：原設定 `CombatInfo=true` 會在怪物上方繪製 `H/HP`，但
這是 remake 診斷增強、不是精訊版原版 UI；`internal/config.Default` 已改為 `false`，設定
仍可由 `combat_info` 明確開啟。README 引用的八頭大蛇、巴哈拉達與怪力魔戰鬥 PNG 已在
Docker／Xvfb 以關閉診斷資訊重產，避免玩家可見文字落在框線外；battle frame／文字框的
完整 V3 對拍仍待原版同狀態核對。

同日驗證：`TestOpeningProductionInputTrace` 在 Docker＋Xvfb 以正式 `InputState` 重播通過
（311.44 秒，從新遊戲至 `THE END`）；NameInput 空名回傳仍由主角呼叫端 fail closed、酒館
則回退職業名。README 戰鬥圖的 `combat_info=0` 重產也已完成。

2026-08-10 完成指定桌面發佈收尾：Docker 內建立 Linux AppImage、Windows x86_64 ZIP、
macOS Intel／Apple Silicon ZIP，並以正式 renderer PNG 剪出 1280×700、30 fps 的
`dist/dq3-promo-20260810.mp4`。完整檔案、大小、SHA-256、Docker 工具鏈與驗證界線集中於
[`docs/114`](docs/114-release-artifacts.md)；Android／WASM 不屬本輪三平台目標。這些
`dist/` 產物不納入 Git，且沒有包含原版資產。

2026-08-10 以 Docker 內 IDA Pro 9.4 重查遭遇 gate，勘誤 `docs/13`、`docs/34`、`docs/35`：
`sub_130cf`（file `0x443f` 起）將 CTY section `+0x11` 寫入 `DS:[0xd77]`，而
`sub_1BD97`（`dec [0x52f4]` 位於 file `0xd118`）在 CTY 模式比較該值，**0=安全、非0=可遇敵**。
Go 現在由 `Town.EncounterFlag`／`Scene.encounterFlag` 保存 raw，正式地表與遭遇 CTY 移動使用
`DS:[0x52f4]` 對映的步數計數器；安全 CTY、飛行與聖水 gate 均不誤觸發，save JSON 保存
`encounter_step`。完整輸入雜湊、IDA linear／logical／file 位址與限制見 [`docs/115`](docs/115-encounter-cty-gate-re.md)。
本輪只做受影響的針對性測試，沒有重跑長達 311 秒的完整 production trace；強制遭遇、完整
encounter pack JSON、戰鬥 V3 與音效仍依 `docs/74` 保持未完成。

2026-08-10 戰鬥 frame 契約勘誤與接線：IDA Pro 9.4 追完
`sub_1f590 → sub_1fd30 → sub_1fdb1`，確認 `0x031a`、`0x0b11`、`0x0912` 是
`win_rect` 的 raw flags／背景備份槽，不是三種 frame style。`sub_1fd30` 依共用
EGA writer 畫四邊；與原版 `dosbox/orochi_boss.png` 像素核對後，可見結果保存為
`interface.json` 各 battle window 的 `frame`（白色 1px border、黑色 interior，D2
pack evidence；推論說明見 [`docs/109`](docs/109-battle-hud-rects-re.md)）。
`fillPackBox` 已讓戰鬥訊息、指令、敵名及隊伍狀態框消費 JSON，
`NewGameWithPack` 缺戰鬥 frame 即 fail closed；直接 fixture 才保留導覽用 fallback。
`internal/gamepack` 的 EXE/HUD parity、`game` 編譯、`TestBattleCommand`、
`TestMultiEnemy`／`TestHit` 均在 Docker 通過；本輪未重播 311 秒完整 trace，戰鬥逐動作
timing／動畫／音效、抗性／formation／掉落仍是 V3 長尾。

2026-08-10 敵人數量列基線資料化：`sub_1b101` 的名稱／數量 writer 與
`dosbox/orochi_boss.png`（SHA-256 `746d7a95c2456fa6338438e6b61ff4fe120670304ebb0f98fdb98c36ef0cad55`）
均顯示數量字與名稱同一個可見 16px row；`interface.json.battle_enemy` 新增
`count_inset_y=0`（相對 `text_inset_y`），`battle.go` 只消費 pack 欄位，validator
拒絕負值／缺欄位，`TestDQ3BattleHUDRectsMatchOriginalEXE` 鎖定此值。這是 strong+D2
的局部畫面契約；數字 writer 內部 raw DX 轉換、spell／target rect、逐動作 timing／
動畫／音效仍不等於戰鬥 V3。Docker＋Xvfb 的 `TestDumpMultiEnemyBattle` 另產出
[`battle_count_row_runtime.png`](dq3_remake_ebitan/docs/battle_count_row_runtime.png)
（SHA-256 `a9c49e091d25bc28d018cf5363c07f75dac4a8029a91cf78f9da60b179518eb8`），僅作
debug 視覺核對，不取代正式入口。

2026-08-10 詳細狀況窗已完成一個可回查的資料化切片。IDA Pro 9.4 由
`lea si, ds:3DA8h`→`sub_1F4E3` 閉合原版視窗結構：`DGROUP 0x3DA8`、file `0x19EE8`、
flags `0x0301`、位置 `(152,46)`、尺寸 `352×192`、標題 D3TXT00 record 407；原始
`D3TXT00` glyph/control stream 仍完整保存在 `docs/data/d3txt_codes.json` 與 game pack
文字資料。`interface.json.field_status` 現在保存 window/frame、22×12 網格、職業／性別
glyph 與具名欄位 anchor；`panel.go` 只依 pack 填入主角姓名、職業、性別、等級、HP／MP、
五項能力、最大值、攻擊、防禦與經驗，沒有版本專屬文字或座標 fallback。缺資料時
`NewGameWithPack` fail closed。證據與推論等級見 `docs/116-field-status-panel-re.md`，
欄位契約見 `docs/84-game-pack-json-contract.md`；Docker＋Xvfb runtime 對拍為
`dq3_remake_ebitan/docs/status_detail_runtime.png`，目前 E2／V2，尚非完整狀況子選單、
隊員詳情、道具／裝備逐窗或同狀態 V3。

本輪發行範圍仍只包含三種桌面包：Linux AppImage、Windows x86_64 ZIP、macOS ZIP（Intel
與 Apple Silicon 各一包）；Android、WASM 不列入本輪驗收或發佈宣稱。所有建置、測試、
截圖與影片工作必須在一次性 Docker 容器內執行；提交前清點並清除不再使用的專案容器，
保留使用者未追蹤的 Android libs、scratchpad 與反組譯工具檔。

狀況窗切片後已在 Docker 重新編譯並覆蓋同一批三平台桌面包；最新檔案尺寸／SHA-256、
AppImage `squashfs-root/usr/bin/dq3-remake` 內嵌執行檔對照，以及 Windows MinGW 工具鏈
來源均記錄於 `docs/114-release-artifacts.md`。推廣片仍為同一組版控 runtime PNG 剪輯的
`dist/dq3-promo-20260810.mp4`；不要把未納入本輪的 Android／WASM 產物加入發佈清單。

2026-08-10 開場能力確認欄位勘誤：重新放大 `docs/dosbox/v3_01_afterstats.png`，並以 D3TXT00
record 407／`docs/data/glyph_unicode_map.json` 逐 glyph 解碼，確認原版右欄六列為「運氣點數、
最大HP、最大MP、攻擊力、守備力、經驗」。舊的 `agility`／`hp`／`mp` 對映把詳細狀況窗欄位
錯套到開場 panel，已由 `interface.json.new_game_labels` 的 `luck`、`max_hp`、`max_mp` 修正；
`NewGameFlow` 正式消費 `stats.LUCK`、最大 HP／MP，缺資料仍由 pack validator fail closed。
這是 D2 設定資料切片與 runtime V2 修正，並非完整 V3：目前框線仍是簡化白框，原版 EGA
pattern／藍色選項框／同狀態逐像素對拍未閉合。證據與 role 對映見 [`docs/112`](docs/112-newgame-labels-re.md)。

2026-08-10 開場外框色彩切片：IDA Pro 9.4 以 `sub_10854` 的
`byte_28b78/byte_28b92/byte_28bc6` caller、`sub_1f590 → sub_1fd30 → sub_1fdb1` 四邊
writer 及 `DS:0x727`（file `0x16867`）閉合共用 frame primitive。DOSBox
`v3_01_afterstats.png` 的能力／提示外框實際 RGB 是 `(255,223,255)`、內部黑；原版
`confirm_choice` 仍有藍黑交錯 pattern，不能由 solid RGB 欄位假稱完成。`NewGameGeometry`
新增 pack-owned `frame`，NewGameFlow／酒館改用該 frame，`NewGameWithPack` 缺欄位即
fail closed。這是 D2／runtime V2；完整 pattern 與同狀態 V3 仍依 [`docs/117`](docs/117-newgame-frame-re.md)
保留為後續工作。

2026-08-10 開場 EGA 邊線遮罩切片：重新依 IDA `sub_1fd30`／`sub_1fdb1` 的 `bh=0xaa`
與 raw `DS:0x727` plane mask 閉合 logical 邊線的局部奇偶遮罩；`FrameStyle` 新增必填的
`border_pattern`，只接受 engine primitive `solid` 或 `checkerboard_1px`，禁止 pack 放任意
JSON code。新遊戲 frame 使用 `checkerboard_1px`，其未命中的邊線像素保留 FIRST.SCR 底圖；
戰鬥／對話／HUD／狀況窗明載 `solid`，validator 對缺值與未知 pattern fail closed。
新增 `frame_pattern_test.go` 鎖定底圖保留語意，Docker＋Xvfb 重產並覆蓋
`dq3_remake_ebitan/docs/ng_menu.png`、`ng_name.png`、`ng_gender.png`、`ng_confirm.png`。
這是 strong／D2 的 engine-data slice；`confirm_choice` 藍色選擇圖樣、FIRST.SCR palette、
逐幀游標／動畫與同狀態 V3 仍未閉合，不能將本切片宣稱為完整創角 parity。

2026-08-10 勘誤：`confirm_choice` 的「藍色選項框未接」已不再是靜態設定缺口。原版
DOSBox 同輸入證據（raw linear `0x28bc6`）已閉合 `112×64` 藍黑 checkerboard backdrop、
`80×32` 中央黑色 content、`98×50` lavender／膚色交錯 2px frame；`stats_right_rows`
也由錯誤 `y=96` 校正為 `y=126` 起六列。`interface.json` 現保存三個具名欄位，engine
只接受 `checkerboard_frame_2px` 註冊 primitive，缺欄位／accent／pattern fail closed；
Docker frame unit test、gamepack test、Xvfb dump 通過，`dq3_remake_ebitan/docs/ng_confirm.png`
已更新。這是 D2／V2 靜態畫面切片；palette register、游標閃爍／能力條逐幀時序與完整
stat panel V3 仍未完成。詳見 [`docs/118`](docs/118-newgame-choice-backdrop-re.md)。

同日第一性原理日夜結論：四階段 clock、`DarkenPalette`、住宿／黑暗之燈 reset、日／夜
NPC table 及 story-flag filter 已在 engine 與 pack 正式存在，故不重寫日夜機制。剩餘工作是
用原版 raw clock／palette 對拍，並按 [`docs/71`](docs/71-npc-story-flag-visibility.md)
逐一閉合 21 個條件 flag 的入口→writer→pack state→consumer→玩家可見副作用；已有動態
pack flag 不重複造表，無完整閉環者維持 `unknown`／fail closed。

2026-08-10 開機過場切片：先前 `main.go`／mobile 直接進標題，造成玩家第一幀與精訊版不符。
`interface.json.opening` 現保存五張已定位 PCX（`TITA.P`、`TITB.P`、`TITD.P`、`TITE.P`、
`TITP.P`）及 manifest 大小／SHA-256；共用 engine 只消費 asset key、停留幀與跳過規則。
桌面／mobile 的正式入口會啟動 sequence，無輸入推進，任一正式輸入停止並讓同一按鍵進入標題；
Docker 的 gamepack、開機過場與正式 opening trace 測試已通過。素材 identity／靜態接線是 D2／E2／V1，
每張 120 幀、排序、淡入淡出、TITP 是否為標題前最後一張及開場音效仍是限制，不得宣稱完整
cutscene V3；證據與下一步見 [`docs/120`](docs/120-opening-cutscene-re.md)。

2026-08-10 發佈包重建覆蓋前一批產物：目前 `dist/dq3-20260810-*` 是本輪 Go/Ebitengine
原始碼（含 `opening` sequence）在 Docker 內重新編譯的 Linux AppImage、Windows x86_64 ZIP、
macOS Intel ZIP 與 Apple Silicon ZIP；`dist/dq3-promo-20260810.mp4` 也重新剪入五張開機卡，
固定為 1280×700、30 fps、1350 幀／45 秒。最新大小與 SHA-256 只以 [`docs/114`](docs/114-release-artifacts.md)
為準；本段之前的同日期雜湊是歷史產物，不可複製到 release note。AppImage 已在 Docker
`--appimage-extract-and-run`＋唯讀 `assets_raw` 啟動 smoke；Windows／macOS 只做 PE／Mach-O
架構與 ZIP 內容檢查，未在 Linux 主機假裝執行。完整 311 秒主線 trace 沿用既有通過紀錄，
本輪依使用者指示只跑受影響的快速測試，不能把這次重包解讀為戰鬥／音效 V3 完成。

2026-08-10 release 分流：公開 release 只放不含原版素材的 Linux AppImage、Windows x86_64
ZIP、macOS Intel／Apple Silicon ZIP；最新檔名、尺寸、SHA-256 與 Docker smoke 界線見
[`docs/122`](docs/122-release-20260810.md)。先前 `dist/release-20260810/full/` 的四個本機
完整版仍是歷史產物，使用的舊 SHP hash 不可再沿用；下一次完整版必須以目前
`work/DQ3MNS_fixed.SHP`（SHA-256 `eb846763ee5d5b1582d0e54678f238539b02e997f359f618d833ce8bac2b8adb`）
替換合法 `assets_raw/` 的同名副本。不得上傳或把原版素材加入 Git；Linux headless smoke
只在沒有 `/dev/snd` 時省略 MT-32 初始化，不能把此 smoke 當成音效 parity。

2026-08-10 128／129 sprite 風格稽核與局部整理：`01f023c`／`8a538a5` 的歷史提交確認
外部回補背景與綠色 Hydra 的來源；`tools/make_sprites.py` 與 `tools/style_sprites.py`
現在均保留 opaque mask、尺寸、姿態與 4-plane 封裝，只把 palette alias、明暗群組與固定
相位抖點整理得更接近 1--127 的原生像素節奏，並同步更新現行 Go fallback
`internal/dq3data/restored_sprites.go`。後續又把與原版 consumer 相容的逐列 RLE AND-mask
附回 128／129，產生可直接載入的 `work/DQ3MNS_fixed.SHP`，並同步到 ignored 的
`dq3_remake_ebitan/mobile/assets/DQ3MNS.SHP`；`assets_raw/DQ3MNS.SHP` 仍是未修改的唯讀基線。
`spr_128.png`、`spr_129.png`、完整 `monster_sheet.png` 與 128／129 聯絡圖均已在 Docker
內重新產生；這仍是外部回補的 `strong` 風格推論，不是原始美術 bytes 的 `confirmed` 證據。
最新 SHP／embedded copy 與每張圖雜湊見 [`docs/121`](docs/121-monster-128-129-style-audit.md)。

2026-08-10 靜態戰鬥／日夜 RE 收尾（IDA Pro 9.4、Docker、原始 `DQ3.EXE` SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`）：

- 正式戰鬥鏈 `sub_1BDDF → sub_1C08B → sub_1C34F → sub_1A973` 已證實玩家／敵人速度
  queue、排序、每個活躍 actor 每回合一次、死亡／睡眠／逃跑／施咒／物理分派。沒有找到
  boss 同回合 repeat-N 外層迴圈；另有 `loc_190D7`（file `0xa447`）的事件／AI-only
  runner 會重建 queue 最多 12 輪，但每輪每個 entry 仍只呼叫一次。boss 同回合多次行動
  維持 `unknown`，不可用連戰 queue 或 D3MNS 未命名欄位代替。
- `sub_1C425` 已閉合經驗／金錢分配與掉落：`DGROUP 0x2518 bit1` 抑制掉落，D3MNS
  `+0x25` 作 **`AL <= +0x25` 的 inclusive RNG gate**（`0xff` 保證通過）、`+0x26` 作
  item raw id；`+0x21/+0x23` 的經驗／金錢 consumer 已確認。剩餘只是正規化顯示名稱與
  各 scripted gate 組合的 parity，不得改寫原始比較或填入猜測值。
- `sub_1D7FB`（file `0xeb6b`）已閉合 D3MNS `+0x19..+0x1d` packed 2-bit resistance、
  `DS:0x36e5` effect thresholds 與 `DS:0x36f5` class 門檻 `0/68/180/255`；相等先遞增
  類別及連續 `0x00` sentinel 已回填 parser／parity test，不能用 first-`>=` bounds 近似。
  單體 `AL<=threshold` 與群體 `AL<threshold` comparator 也分開記錄；每個中文咒文／效果
  名稱的 JSON 展開仍待 sidecar。
- `sub_1A7D5`／`sub_1BF35` 已閉合候選八筆、38 點 encounter budget、權重、group／
  background／page 與 active position；`sub_1B16F`／`sub_1B31A`／`sub_1B2AF`／
  `sub_1B3C3` 證實 SHP 單 payload、RLE AND-mask、四 plane、位置及六輪 1/2 tick 重畫，
  目前沒有靜態證據支持每個動作另選 sprite frame。
- 音效不是 `sub_1683A`（那是 tick 等待）。戰鬥 `BP` cue 由 `sub_20770` 取
  `DS:0x253e` 四位元組指標，經 `sub_22CF5`／`sub_236B4` 進 Sound Blaster VOC driver；
  `sub_208A7` 選 `FVOC.VCX`／`NVOC.VCX`，`sub_208E2` 等待完成。cue 原始數值與 call-site
  已記錄，FVOC／NVOC raw block offset／length／rate byte 也已列入 sidecar；實際波形／
  玩家可見時長仍是 V3。
- 日夜 `sub_1EE23`（file `0x10193`）已證實有效 clock `0..0xef`、`0x78` 寫 night
  selector、`0xf0` 歸零寫 day selector；`sub_1EE76`（file `0x101e6`）查
  `DGROUP 0x25c5` 原始 12-byte index `01 00 00 00 01 02 03 04 04 04 03 02`，由
  `DQ3.PAL` buffer `DS:0x3232` 選 bank／上傳 DAC；`sub_131BD` 以 `[0x526c]` 選日／夜
  NPC 表並測 `[0x4f70]`。
- 21 個原版 story flag 的 SET literal→`sub_16EDF`→`[0x4f70]` 已建立 writer ledger；
  原版 writer 全部 `confirmed`。`0x0d`、`0x11`、`0x24`、`0x1a`、`0x1b` 的 remake
  production 事件／consumer 仍 `unknown`；不可把「runtime unknown」誤寫成「原版沒有
  writer」。`0x15` 及 `0x13` 不屬這 21 個清單。

完整位址、輸入 hash、IDA 位址基準、raw table 與限制集中於
[`docs/123`](docs/123-static-battle-daynight-re.md)；可重播的非 production sidecar 位於
[`docs/data/dq3-static-battle-daynight-evidence.json`](docs/data/dq3-static-battle-daynight-evidence.json)，
並已在 Docker 內以五份原始輸入的 hash 與 EXE palette index 通過 parity。日夜舊版未決描述已在
[`docs/data/day-night-system.md`](docs/data/day-night-system.md) 以新靜態補充覆蓋。
這一輪只完成靜態設定證據，不宣稱戰鬥逐動作畫面／聲音 V3；下一輪應將達 D2／D3 gate 的
欄位逐批遷入 versioned game-pack，再從合法 production checkpoint 補實機 timing、
formation、drop 與缺失 flag 的正式玩家入口。

2026-08-11 抗性 parser 勘誤：重新以 IDA `sub_1D7FB` 的 `cmp AL,[0x36e5+BX]`／相等
`inc BX` 重算後，確認舊 `SpellChance` 的 first-`>=` bounds 會在 spell 3、6、38、59
等邊界偏移；連續 `0x00` sentinel 也不可刪除。現已改用原始 threshold stream，限制
effect descriptor `0..59`，加入 spell `2/38/59`、怪物 121 與越界 fail-closed 測試；
Docker `go test ./internal/dq3data` 通過。這只修正 raw resistance decoder，不代表中文
效果名稱或玩家可見抗性 V3 已完成。

2026-08-11 靜態補證（IDA Pro 9.4、Docker、同一份 EXE SHA-256）：

- formation：`sub_1A7D5` 的 `dec bx; shl bx,5` 修正候選表索引口徑；`DGROUP 0x4966`
  區域表有效值為 `1..0x1b`，`DS:0x29826` 的區域 1..27 各 4×8-byte raw candidate rows
  已完整放入 sidecar（使用長度 `0x360`、SHA-256 `46edf59f...f3d69`）。固定事件編隊
  `DS:0x4eb5..0x4eff` 的 13 個原始 records 也已逐筆保留，未替 header／candidate byte
  猜欄位名稱。`sub_1AAA1→sub_1AAD5→sub_1AB2C` 證實 active `+0x03` 的 budget／weight
  writer 與 `sub_1B1FE` consumer；精確螢幕座標仍 unknown。
- 抗性：`DS:0x37c3`（file `0x19903`）60 筆 3-byte descriptor 的 raw hash 與
  `D3TXT00.TXT` record `0x79+id` 中文名稱已加入 sidecar。名稱只證實文字 record 對映，
  不可反推元素／狀態抗性；packed pair 仍只輸出 raw class／threshold。
- 五個 story flag：handler74 `sub_16346→sub_1E713` 閉合 `0x0d` 的 map reload／animation；
  handler69 `sub_16139` 閉合 `0x11` 的 map-cell write、SET→`sub_134AB→sub_131BD→
  sub_11971→sub_12C9C`→CLEAR；handler70 `sub_1622A` 閉合 `0x1a/0x1b` 的 post-battle
  reset、日 selector／palette、DQ3BLK／DQ3UND rebuild；handler33 `sub_15904` 閉合
  `0x24` 的 rec／clear/set／close。89 個 CTY、768 筆 NPC 的 full byte5 掃描命中
  `0x0d:26/0x11:1/0x1a:6/0x1b:25/0x24:4`，證明 generic `sub_131BD` consumer 有 raw
  records；Go production writer、direct GET caller 與玩家可見事件仍 unknown。
- boss：正式 queue 與 alternate `loc_190D7` 都沒有同 actor 同回合 repeat-N；negative static
  evidence 更完整，但 boss 多次行動的資料條件／數量仍 unknown。SHP 仍只找到 monster-id
  單 payload，VOC 只確認 `BP→DS:0x253e→sub_22CF5→sub_236B4→sub_208E2` 與 raw VCX
  blocks；動作 frame、PCM 波形與玩家停頓不由反組譯單獨升為 V3。

本輪 sidecar／[`docs/123`](docs/123-static-battle-daynight-re.md)／[`docs/119`](docs/119-daynight-story-flag-audit.md)
已追加 raw 位址、來源 hash、consumer 與推論等級；沒有把未知值寫入 game-pack 或 Go fallback。

2026-08-11 攻略×IDA 勘誤：本輪把 `docs/66` 的主線 step 43–45、`docs/75`、`docs/76`、
`docs/77` 與原始 CTY 逐格對讀後，確認 `0x1a/0x1b` 不是龍之女王／彩虹材料中繼。龍之女王
光之珠是 CTY67 handler52（step 43）；CTY65 sec0 `(8,3)` handler70 `sub_1622A` 是
巴拉摩斯戰鬥後交易（step 44），會重置 party、day selector／clock／palette，並同時 SET
`0x1a`／`0x1b`、重建 `DQ3BLK`／`DQ3UND`。CTY00／CTY25 的 full-byte NPC records 與
「打倒巴拉摩斯」等文字是其 generic visibility consumer。先前錯誤說法保留在
`docs/data/special-events-audit.md` 的歷史勘誤段並明示已推翻；現行 runtime 仍是 Go
缺少兩旗標 production writer，不得把未接線值冒充 E3。

五旗標目前唯一安全結論：原版 handler／writer→`[0x4f70]`→generic NPC consumer 已以
IDA 9.4 靜態證實；`0x0d`（索瑪後期）、`0x11`（六珠祭壇暫存）、`0x1a/0x1b`
（巴拉摩斯戰後）、`0x24`（沙曼歐莎條件居民）的 Go 正式玩家入口仍須另開 E3 垂直切片，
不能以攻略地名或自有 `cleared`／動畫狀態猜填 JSON。

2026-08-11 E2 milestone 決策（使用者確認）：本輪改採快速 E2／D2 完成標準，使用 IDA Pro
9.4、原始 EXE／DAT／SHP／VOC／CTY bytes、可重建 decoder、Docker deterministic trace、
schema/reference/parity tests 與既有 reference PNG 作驗收，不重新啟動 DOSBox。只把
`confirmed`／必要的 `strong` 證據遷入 versioned game-pack；boss repeat-N、精確座標、逐動作
frame、PCM 播放／停頓及五個旗標的完整玩家語意仍保留 `unknown`／fail-closed，不得在
README／release 宣稱 V3 parity。E2 工作先完成 JSON／engine integration、正式 InputState
trace、事件／戰鬥交易與 save/load，再做 headless contract 驗收。

E2 第一個垂直切片已完成：修正 `EncounterTables.Slot` 的原版 `dec region; shl 5` 索引，
region `1..27` 現正確對應 raw slot `0..26`；新增 pack `data/battle.json` 保存 encounter
候選／固定編隊與 resistance raw／中文 record 對映。`NewGameWithPack` 由嚴格驗證的 pack raw
建立 encounter decoder，不再直接從 `DQ3.EXE` 讀取版本專屬遭遇表；Docker gamepack／dq3data
測試及 Xvfb headless 編譯通過。這是 E2/D2 loader contract，不提升 boss repeat、動畫／音效
或玩家可見事件為 V3。

同一切片也把 `D3MNS.DAT`、`DQ3MNS.SHP`、`PACKBG.SCR`、`MNSBK.PAL`、`FVOC.VCX` 的
battle asset 路徑與完整性移入 manifest；SHP 同時接受原始 bytes 與已產生 128/129 版本的
受控 alternate hash，兩者都必須精確匹配，不能接受任意檔案。

2026-08-11 E2 headless 驗收收尾：沒有啟動 DOSBox；沿用專案 `Dockerfile.test`，以 UID/GID、
唯讀 `/home/anr2/go/pkg/mod` 與 `xvfb-run` 執行。`go test ./internal/...` 全綠；
`TestE2BattlePackEncounterTrace`、`TestBaramosProductionInputTrace`、
`TestMirrorProductionInputTrace`、`TestMirrorWinSaveRestore`、save round-trip、battle queue、
混合 formation、逐敵行動與全滅獎勵測試均通過。這證實 pack raw → loader → production battle／
event → save/load 的 E2/D2 閉環；既有 PNG 僅作畫面 V2 證據。boss repeat-N、逐動作 frame、PCM
波形／玩家停頓、精確 formation 座標及五個 story flag 的 remake 可見入口仍為 `unknown`／V3
gap，不能在 README 或 release 說成完整原版 parity。這批只更新文件與測試驗收紀錄，尚未
commit/push。

附帶回歸觀察：完整 `TestOpeningProductionInputTrace` 在本輪 Docker 重播於 CTY7 入口後的
CTY7→CTY8 section route fixture 停在 `(7,6)` 失敗；它不屬於使用者選定的 E2 gate，因此沒有
用 debug state 修補或把結果寫成 E3 通過。若後續要回到完整 campaign，必須先修正式 portal
路徑，並由合法 checkpoint 重播。

勘誤（2026-08-11）：上述段落是修正前快照。現行 route fixture 已依原始 CTY07→CTY08
transition chain，先由雷貝正式道具店購買聖水，再以道具／戰鬥 `InputState` 穿越遭遇區，
可抵達 CTY08 sec3 並接上拿吉米老人；完整 trace 的後段仍有 CTY13 路徑 fixture 未收尾，
所以仍不宣稱 E3。README 的 `V3-approx` 政策只允許呈現／時序近似，不改變這個 gate。

2026-08-11 CTY13 勘誤：前述「CTY13 路徑 fixture 未收尾」是修正前快照。CTY13 sec2 的
原始 section header `+0x11=10` 允許遭遇，`(25,24)`／`(25,25)` 原始 tile 均可走；真正
阻塞是 `traceWalkToNoPortal` 在遭遇 modal 時仍送方向鍵，沒有走正式 battle consumer。
現已補 `traceResolveBattle` 與正式聖水輸入，從合法前段 trace 可抵達 CTY13 sec2、完成雙
開關、取得 `0x56` 魔法鑰匙並通過 save/load（targeted E2）。完整 campaign 在此修正後尚未
重播完畢，故仍不重新宣稱 E3。

2026-08-11 formation／PCM 補證勘誤：IDA Pro 9.4 的 `sub_1AAA1→sub_1AAD5→sub_1AB2C→sub_1B31A`
已閉合 raw position 公式：第一筆為 `0x26 - Σ(D3MNS +0x28×count) + 2`，後續每隻以
`2×D3MNS +0x28` 前進；`sub_1AB2C` 逐個體擲 HP，不能再用同組代表 HP／起點 `0x26`。這些
常數已遷入 `battle.json.formation_position`，renderer 依 EGA `0x54` stride／`0x102` bottom
投影；raw formula 為 confirmed，像素面只標 strong static，未宣稱 V3 畫面。VOC decoder／
gaudio sidecar 另保存 cue 的 source rate、sample count、可重建 duration；`PlaySFX` host
wall-clock、boss repeat-N、逐動作 SHP frame 仍 unknown／V3 gap。`NVOC` cue21 raw offset
勘誤為 `0x1f973`；完整證據在 [`docs/123`](docs/123-static-battle-daynight-re.md) 與
`docs/data/dq3-static-battle-daynight-evidence.json`。

## 2026-08-11 Android UX／build checkpoint

使用者要求開啟 Android 版工作；本輪採保存優先的 port UX，不改寫精訊版 640×350 畫布、
繁中劇本、注音／英數選字或 game-pack 邊界。`docs/124-android-ux-spec.md` 定義浮動四向
十字鍵、右下 A/B、標題／命名情境鍵、非控制區選單點選、橫向沉浸式外殼、`FLAG_KEEP_SCREEN_ON`
與 `onPause`／`onResume` 規則。Android manifest 的玩家可見 App 名稱改為「傳說的終章」，
遊戲內文字仍只能來自 JSON／text ID。

Docker build 使用既有 `rich2-go-android:20260809`（SDK 35、AGP 8.2.2），並以 UID 1000
在隔離 HOME／Gradle cache 中執行；`compileSdk` 固定 35、`buildToolsVersion` 固定 35.0.0，
`targetSdk` 維持 34。最新 Go AAR 為 `dist/dq3-android-current-20260811.aar`，SHA-256
`cc0c12730c39190362e9699d07232e553d86d4a76fbdb209004f28965a4c8566`；debug APK 已產生於 ignored
`dist/dq3-android-debug-20260811.apk`，
SHA-256 `ea830ea06dfde28ce3e912669092663c4cde03069575a8fd695c1e1915efe62c`；`aapt` 確認
套件 `com.wicanr2.dq3`、中文 label 與 landscape feature。這只是 `build-only`，尚無 emulator
或真機 touch／lifecycle／音訊／存讀檔證據，不得把桌面包或 Docker build 宣稱為 Android 實機
完成。APK 含本機合法素材且不加入 Git；Docker 工作容器已清理。

## 2026-08-11 IDA evidence → story-flag runtime E2 checkpoint

本輪依 IDA Pro 9.4 已閉合的 handler／caller／consumer 直接接線，沒有以攻略或合理預設補
值。`dq3_cht` game pack schema 升至 `0.1.29`，`events.json` 新增嚴格驗證的
`story_flag_runtime_events`；引擎只執行有限 primitive，缺欄位或未知值 fail closed。

- handler74／CTY80：國王冊封完成時 clear `0xf5`、set `0x0d`、scene reload；Ending writer
  由 `advanceEndingKing` 接入。
- handler69／CTY70：六珠復活動畫完成後寫 map cell byte `0x80`，短暫 set `0x11` 並於
  場景重建前 clear；map patch 以 `phoenix_map_cell_written` 進 save/load，拉米亞停放座標與
  動畫 frame/tick 全由 pack 提供。
- handler70／CTY65 `(8,3)`：巴拉摩斯固定編隊戰後 clear `0x29`、set `0x1a/0x1b`，重設
  day phase／step、palette；掛在既有 `onBattleEnd` 正式 boss trigger。
- handler33／CTY44 `(13,27)`：正式「話す」入口先開原始 `DI=0x0bff`，依 `di-0xbb8` 換算
  D3TXT record 71，再 clear `0x3d`、set `0x24/0x21`、重載條件 NPC。

新增 `game/story_flags_test.go` 鎖定 pack evidence、正式 `InputState` command path、Phoenix
map/save replay、Baramos aftermath 與 ending writer；Docker＋Xvfb targeted tests、
`internal/gamepack` tests 通過。這是 E2 engine/data／save transaction closure；不宣稱原版
同狀態 V3、完整 campaign E3、逐幀 SHP／PCM wall-clock、精確 palette register 或 boss
repeat-N。前文的「runtime unknown／missing」保留作歷史勘誤，現行狀態以本節、`docs/119`、
`docs/123` 及 pack JSON 為準。

## 2026-08-11 驗收勘誤：current checkpoint 尚未重播閉合

本輪以現行 Go／Ebitengine 測試器、Docker＋Xvfb、Android 靜態工具鏈與 DOSBox oracle
取樣驗收。`internal/...`、觸控 UX 契約測試全綠；完整 `TestOpeningProductionInputTrace`
在 `traceBoardAndSailShip`／`renderFrame` 逾時，排除該長測後唯一觀察到的失敗是
`TestKandarTowerProductionRouteOpensThiefKeyDoor` 的 Lv1 單人遭遇全滅。Android APK 可
重現但沒有 emulator／裝置，屬 build-only。原版開場室內、新遊戲能力確認與 Orochi 戰鬥
只形成 near-state／V2 對照，不能提升為像素 V3。歷史曾抵達 `THE END` 的 checkpoint
不能取代本輪正式輸入重播；因此目前不宣稱 campaign E3。完整指令、雜湊、畫面差異與
下一個最小動作見 [`docs/125-acceptance-20260811.md`](docs/125-acceptance-20260811.md)。

## 2026-08-11 後續重播勘誤：campaign E3 已閉合

上一節是當時受限重播的歷史快照，已被同日更強的完整重播推翻，保留它只為追溯錯誤成因。
`boardPhoenix` 在下層前往拉米亞時把 battle 保留的 movement cooldown 誤當一般 cooldown，
因而無限送空白輸入；修正後會以正式戰鬥選單結束隨機遭遇再繼續移動，未注入座標、事件或
資源。`TestOpeningProductionInputTrace` 隨後以正式 `InputState` 從新遊戲到 `THE END` 通過
（63.92 秒），含 P4–P6、巴拉摩斯、下層、索瑪三連戰、冊封及中途 save/load。

同時修正甘達特塔 route fixture：它的 Lv1 單人狀態只意圖驗證盜賊鑰匙門與 transition，卻把
塔區遭遇誤混入；現以正式道具面板使用聖水管理遭遇，不改怪物資料。Docker＋Xvfb 將全部
288 個 `game` 頂層測試分 18 個一次性程序批次執行後全數通過；`internal/...` 亦全綠。
因此現行主線狀態是 **campaign E3／runtime V1**，不是原版同狀態畫面與音效 V3；Android
仍只有 build-only。完整命令、終局 PNG 雜湊及舊結論訂正見
[`docs/125-acceptance-20260811.md`](docs/125-acceptance-20260811.md)。

## 2026-08-11 能力確認 HUD 框線像素勘誤

以 `dosbox/v3_01_afterstats.png` 的 640×350 logical pixels 逐點取樣後，確認能力確認頁
三個大面板（StatsLeft／StatsEquipment／StatsRight）外框是連續 lavender
`RGB(255,223,255)`，上／下各兩列、左／右各兩欄；先前把 EGA `bh=0xaa` 直接投影為
`checkerboard_1px` 的結論已標為推翻。`interface.json.new_game_geometry.frame` 現改用
具名 `solid_2px` primitive；中央 `confirm_choice` 的藍黑 checkerboard 與 lavender／膚色
`checkerboard_frame_2px` 不變。新增 frame regression test，完整證據與勘誤見
[`docs/117`](docs/117-newgame-frame-re.md)。此修正是 D2／V2 靜態 HUD 對拍，不提升逐幀
palette、游標動畫或整段創角流程為 V3。

## 2026-08-12 current memory：能力確認 V3 靜態與 RE 停止條件

現行 game pack schema 是 `0.1.31`，DQ3 content version 是 `0.1.36`。能力確認頁的 DQ3
資料已完全留在 `interface.json`：record 407 glyph stream、13 個具名 stat field、raw EGA
window backdrop layer、`frame_edge_widths`、`beveled_2px` 與 `FIRST.SCR` palette；Go 只保留
跨版本 renderer／validator primitive。

原版正式創角輸入（名稱 `0`、男性、確認）與 remake fixture 在 640×350 得到 AE=1,474
（0.658%）的 V3 靜態 checkpoint。此結果只適用固定畫面；游標／palette／淡入與音效
時序未被單張 PNG 宣稱完成。原始 hash、殘差、實作與反證見 [`docs/126`](docs/126-newgame-confirmation-v3-static-comparison.md)。

IDA Pro 9.4 的新 audit 重查 queue builder／actor consumer xref、alternate runner、SHP redraw
及 VOC dispatch，確認可靜態閉合的 formation、resistance、五個 story flag、clock／palette
bank 已在資料／runtime 接線；boss repeat-N、逐動作 frame、PCM wall-clock 與可見 palette
transition 沒有安全 production 值。除非取得新原版 frame／音訊 trace 或具體新 caller，
不要再為這些項目猜 JSON、加 Go fallback 或廣泛重跑反組譯。完整界線見
[`docs/123`](docs/123-static-battle-daynight-re.md)。

## 2026-08-12 current memory：八頭大蛇固定編隊背景 V2 已閉合

不要再以舊的單勇者 PNG 或修正前的天空／草地 trace 評估日邦格戰鬥。固定編隊 record 的
`raw[1]`／`raw[2]` 已由 IDA Pro 9.4 關聯至 archive page／palette bank，並資料化為
`events.json` 的 `background.{page_raw,palette_bank_raw}`；`battle.json.background` 保存完整
`PACKBG.SCR`／`MNSBK.PAL` 契約，renderer 沒有 DQ3 fallback。

正式 `InputState` replay 的第一、第二戰目前分別顯示洞窟與沙漠，達 near-state V2；不是同隊伍、
同數值、同動態 phase 的 V3。這個窄修正是必要的 faithful-remake 工作，因兩場皆為必經且背景
錯配很醒目；但 campaign 已 E3，且尚無新 player-visible discrepancy，故停止擴張成 generic
terrain selector RE。通用戰鬥只保留已驗證草地 baseline `0/0`，不聲稱全地形 parity。可回查的
raw、hash、推論等級、訂正與停止條件見 [`docs/128`](docs/128-battle-background-selector-re.md)。
