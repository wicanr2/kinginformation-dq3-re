# DQ3 Go/Ebiten remake 接手記憶

> 更新：2026-08-02。接手先讀 `CLAUDE.md`、`CONTEXT.md`，再讀
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
`docs/99`。**尚未完成從新遊戲開始到 THE END 的無 debug 全流程驗收**；下一個 audit 從
提頓綠色寶珠合法 checkpoint 繼續。地表水面顏色及四輪 palette transition 仍未達 V3，
不得因流程閉合而略過。

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
人物初始裝備由 `characters.json` 提供。細則見 `AGENTS.md` 與 `docs/84`。

## 固定工程方法

1. 每輪閉合一個垂直切片：正式入口 → 狀態交易 → 畫面/聲音 → 下一節點 → save/load。
2. 未知值 fail-closed；禁止用 C remake、經典 DQ3 慣例或「合理值」填 production。
3. 每個參數追原版「誰在何時寫入」，並記錄明確位址口徑。
4. component test 與 production-input trace 分級；只有正常玩家可達才可關閉主線 GAP。
5. 進度只維護 `docs/74`；歷史計畫保留短索引，細節從 Git 歷史取回。

## 工作樹注意

接手先執行 `git status --short`。未追蹤的 Android libs、scratchpad 與 `tools/dosbox_verify_*`、
`tools/tmp_*` 可能是使用者檔案；不得刪除、覆寫或誤納入 commit。
