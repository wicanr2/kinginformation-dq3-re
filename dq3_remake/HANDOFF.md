# DQ3 RE/Remake — Session 交接文件

> 寫於 2026-06-26。新 session 啟動於 `/home/anr2/dq3`(git repo)。
> **先讀本檔 → 再讀 `WORKLIST.md` → 動手前比對 `~/.claude/rules/00-rules-index.md` 觸發表**。

---

## 0. 一句話目標

把未發售的精訊 DOS DQ3(16-bit real-mode、large model、MSC 編譯的 `DQ3.EXE`)反組譯還原成
可重編譯的 **C99 + SDL2 remake**(`dq3_remake/`),行為對齊原版(DOSBox 當 oracle),
目標:完整還原 + 可破關。GitHub: `github.com/wicanr2/kinginformation-dq3-re`。

---

## 1. 當前狀態(git ground truth)

- HEAD = `0e859a0` 草薙之劍 0x14 接線(剛提交,工作樹乾淨)。
- **build 乾淨**(docker `cmake --build`,BUILD_EXIT=0)。
- **game_tester 79/79 全綠**(容器內跑,見 §5 正確跑法)。
- 未追蹤檔皆已 gitignore(`new_map_dq3/`、`docs/img/` 等),`git status` 乾淨。

近期提交序(新→舊):
```
0e859a0 草薙之劍 0x14 接線:擊敗八頭大蛇勝利分支給予
7d186af repo 整理:RE 工具腳本入版控 + 暫存截圖 gitignore
4c64620 彩虹水滴 bug#2 修正確認 + 架彩虹橋用途接線 + 文王→女王全改
36085e1 review 第一性原理調查 + glyph 234「文」→「女」更正
5427667 標題主選單(新遊戲 / 繼續冒險)+ 串接 slot 讀檔
1780ba5 隨時存讀檔(F5/F9)+ 讀檔回原位置
5a8c311 多存檔 slot(6 格):記錄點/F10 選格存 + DQ3_LOAD=N 讀
3ec4958 文件對齊:README/WORKLIST/audit(73/73→現 79/79)
834a7d5 special 事件全盤點 + 分類持久化
```

---

## 2. 硬約束(使用者定,不可違反)

- **不用 subagent**:全部 main thread 自己做(使用者明確要求)。
- **docker first**:RE 工具、編譯、跑遊戲一律 docker,不污染 host;Python 用 docker uv.venv。
- **版權**:杜勝利/青衫攻略、原版素材 gitignore,只摘要 + 標出處,不入庫。
- **git push 已預先授權**;commit 結尾加
  `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`。
- **繁體中文**回應。
- **用 ground truth 不靠記憶**:記憶不一定同步更新;斷言前回原始資料/git 驗證(見 §6)。
- **第一性原理**:撞牆別反射性下「封死/要動態/DOSBox」,錨定實例→找 sink→運算元反向追靜態源。

---

## 3. RE 速查(位址換算 + 關鍵事實)

### 位址換算
- file = logical + `0x1370`
- DGROUP:file = `0x16140` + DS_offset
- rec = di − `0xbb8`
- sub2 jumptable @ `0x3bb4`

### 已確立的機制事實
- **`[0x722]` = region/page index**(座標 hit-test 算出的頁/區索引),**不是** runner event id。57 個 setter。
- **dlg = byte4**(表索引)經 `find_npc_b4` 確認。但 handler 顯示的 rec ≠ byte4 — handler 內部自決顯哪個 rec。
- **sub2 handlers 是過場腳本**(對話/NPC 動畫 via `0x67c5`/`0x21ef`/flag/warp via `0xbe89`=`0xd1f9`),
  **不含** `[0x2321]` 敵群表 / `sub_bfd1` battle_enter / `0x7bbe` GIVE。戰鬥是劇情強制或另一機制。
- **開戰** = `battle_enter_screen (sub_bfd1)`,敵群寫 `[0x2321]`(sprite id)/`[0x231f]`(數量)。
- **位置型道具用法**:DQ3_USE_GAIA/DRAIN/AWAKEN/FAIRYFLUTE/RAINBOW — kind 回主迴圈
  (`g_item_world_eff`),依地點條件決定是否消耗。
- **魔王是地圖固定格 sprite**(非亂跑 NPC):走過去按空白觸發。CTY14 = 甘達特巢穴(已接線)。

### 重要 id(已驗)
- 道具:`0x14` 草薙之劍 / `0x1c` 王者之劍(atk120/35000/mask 0x80 勇者專用)/ `0x5f` 黑暗之燈 /
  `0x60` 回音之笛 / `0x61` 拉之鏡 / `0x65` 光之珠 / `0x69` 紫寶珠 / `0x6b` 銀寶珠 /
  `0x70` 生命戒指 / `0x71` 銀豎琴 / `0x72` 太陽之石 / `0x73` 雲雨之杖 / `0x74` 精靈的守護 /
  `0x75` 彩虹水滴 / `0x77` 妖精之笛
- 怪物:26 甘達特(HP551)/ 27 甘達特手下(HP81)/ 75 八頭大蛇(HP801)/ 121 巴拉摩斯 /
  124(0x7c) 索瑪
- glyph:**234 = 「女」**(原誤標「文」,已修)/ 430 才是真「文」。要正確 glyph 序列**從原版
  D3TXT record 解**,不靠 unicode_map 反查(見 memory `re-glyph-mislabel-verify-via-record`)。

---

## 4. 已完成 + 待辦

### 本階段已完成
- **B 道具鏈**:草薙之劍 0x14(八頭大蛇勝利給)、王者之劍 0x1c 鏈、彩虹水滴架橋。
- **A boss**:CTY14 甘達特巢穴觸發點接線(走到格→怪27→怪26→flag 0x211)。
- **存讀檔**:6 slot + F5/F9 隨時存讀 + 讀檔回原位置(`restore_position`)+ 標題主選單
  (新遊戲 / 繼續冒險)。存檔格式 magic `DQ3SAVE5`,path-based 多 slot。
- **glyph 234 文→女**修正(龍之女王正名);全文件/UI 標籤同步。
- **彩虹水滴 bug#2**:合成已產 0x75(非銀寶珠 0x6b);補上「架彩虹橋」用途接線。
- **special 事件全盤點**:53 個 kind=special 全分類(`docs/data/special-events-classified.json`)。

### 立即待辦(下一步,定義明確)
1. **更新 `docs/data/quest-items.md` 的 7 個「待」**(第一性原理已查證):
   - `0x14` 草薙之劍 → ✅(已接線,commit 0e859a0,八頭大蛇勝利分支)。
   - `0x5f` 黑暗之燈、`0x60` 回音之笛、`0x61` 拉之鏡、`0x70` 生命戒指、`0x71` 銀豎琴
     → 6 個其實**已可由寶箱系統取得**(`dq3_treasures` 表已含;ORBSCAN 實機驗證有掉:
     太陽之石「獲得道具 0x72」、生命戒指 (44,26)、銀豎琴 (2,4)、黑暗之燈 (4,9)、拉之鏡 (16,16))
     → 把「待」改 ✅ 並註明寶箱座標(stale label)。
   - `0x72` 太陽之石 第 38 行「待」是**重複舊條**(第 42 行已 ✅ 下層 CTY80 寶箱)→ 合併/刪重複。
2. **WORKLIST.md 數字對齊**:仍寫 73/73,實際 79/79;順手更新。
3. **繼續「修正其他缺口」**:quest 全 ✅ 後,回 `special-events-classified.json` 找剩餘
   2 個 review 長尾(CTY6 折扣 / CTY65 換蛋)是否真缺口(先前第一性原理結論:兩者皆**非**真實
   接線需求,別過度解讀;若無證據就維持 review 標記)。

---

## 5. 工具 + 正確跑法(踩過的坑)

### build / 測試(必須容器內,用 /repo /build 絕對路徑)
```bash
cd /home/anr2/dq3
# build
docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc \
  'cd /build && cmake --build .'
# game_tester(★ 一定要在容器內,且傳 /repo /build —— 在 host 直接跑會 7/78 全假失敗)
docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc \
  'bash /repo/tools/game_tester.sh /repo/assets_raw /build/dq3_remake'
```
> 坑:`game_tester.sh` 內部硬編 `/repo`、`/build` 路徑(只存在於容器),host 直跑→所有 binary 缺失→
> PASS=7 FAIL=71 假象。**不是回歸,是環境**。

### headless 跑遊戲(debug token)
```bash
docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc '
  export SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy
  DQ3_DEBUG="warp:14:13:12" DQ3_INPUT="..." /build/dq3_remake /repo/assets_raw game 2>&1'
```
- 環境變數:`DQ3_DEBUG`(warp/party/boss token)、`DQ3_INPUT`(鍵序)、`DQ3_DUMP`(PPM 截圖)、
  `DQ3_LOAD=N`(讀 slot N;N≥2 讀 `dq3_saveN.dat`,否則 slot0=`dq3_save.dat`)。
- **注意**:FFFFFFFF 攻擊腳本打不贏八頭大蛇 HP801,所以勝利分支(給草薙劍)headless 不易驗 —
  靠程式碼正確性(與既有紫寶珠 0x69 給予同模式)+ git diff 確認,不強求 runtime 證。

### RE 工具(容器內)
- `tools/list_special_events.py`(`--todo` 過濾已分類 / `--text` 附對話)
- `tools/find_boss_sprites.py`、`tools/dis_handler_full.py`(跟進分支)
- `tools/_big.py`(未追蹤,大檔反組譯輔助)

---

## 6. 紀律(memory,務必遵守)

- **`re-verify-via-git-not-tool-echo`**:工具 stdout/Read 可能幻覺(假 PASS/假 commit)。可靠序:
  `git log/diff/status` > python 寫檔再讀數字 > `grep -c` 數字 > bash stdout 回顯(最不可靠)。
  宣告完成**只在** git diff 驗證 + 檔案中介 PASS=N 之後。
- **`re-glyph-mislabel-verify-via-record`**:要某詞正確 glyph 序列,從原版 `assets_raw/D3TXTNN.TXT`
  解真實 record(格式 docs/03),不靠 `glyph_unicode_map.json` 反查(D3TXT00 系統檔含框線符,誤標多)。
- bash heredoc 內 python `-c` 別用反引號(bash 會先執行→刪內容);用 Write 寫檔再 python 讀。
- 其他 memory 見 `~/.claude/projects/-home-anr2-dq3/memory/MEMORY.md`(retro RE 系列教訓)。

---

## 7. 關鍵檔案地圖

| 檔案 | 用途 |
|---|---|
| `dq3_remake/src/main.c` | 主迴圈、場景、存讀檔、標題選單、boss/道具接線(大檔)|
| `dq3_remake/src/dq3_item_use.c` / `include/dq3_item_use.h` | 位置型道具效果(GAIA/DRAIN/AWAKEN/FAIRYFLUTE/RAINBOW)|
| `dq3_remake/src/dq3_save.c` / `include/dq3_save.h` | 存檔格式 DQ3SAVE5 + pos(含 in_town/layer/sec)|
| `dq3_remake/src/dq3_inventory.c` | `dq3_synth_rainbow_drop`(fixed=1→0x75,bug#2 已修)|
| `dq3_remake/include/dq3_runtime.h` | scancode(DQ3_SC_F5=0x3f / F9=0x43)|
| `docs/data/npc_dialogue.json` | NPC 權威 dump(root-owned,改要 docker)|
| `docs/data/special-events-classified.json` | 53 special 事件分類(避免重複盤點)|
| `docs/data/quest-items.md` | 道具取得鏈(§4 待辦:7「待」要更新)|
| `docs/data/glyph_unicode_map.json` | glyph→中文(234 已修文→女)|
| `docs/re-log-722-state-machine.md` | RE journal(已到 Step 42)|
| `docs/remake-wiring-lessons.md` | 接線教訓(含工具幻覺)|
| `dq3_remake/WORKLIST.md` / `README.md` | 進度(數字待對齊 79/79)|
| `tools/game_tester.sh` | 交付 gate(§5 正確跑法)|

---

## 8. 給新 session 的第一個動作建議

1. `git log --oneline -5` 確認 HEAD = `0e859a0`,`git status` 乾淨。
2. 跑 §5 的 docker game_tester 確認 79/79(別在 host 直跑)。
3. 執行 §4 立即待辦 1-2(更新 quest-items.md 7「待」+ WORKLIST 數字),git 驗證後提交。
4. 接 §4 待辦 3「繼續修正其他缺口」,用第一性原理盤 special 長尾。
