# DQ3 Go/Ebiten remake 接手記憶

> 更新：2026-07-29。接手先讀 `CLAUDE.md`、`CONTEXT.md`，再讀
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
- 最新 production audit：[`docs/83`](docs/83-inn-and-romaly-route-audit.md)
- 香巴尼塔甘達特原版事件與正常路徑：[`docs/85`](docs/85-shanpane-kandar-production-trace.md)
- 近期 IDA/影片證據：[`docs/75`](docs/75-phoenix-orbs-re.md)、
  [`docs/76a`](docs/76-baramos-gaia-re.md)、[`docs/76b`](docs/76-r5-endgame-realignment.md)、
  [`docs/77`](docs/77-r5b-castle-aftermath.md)

目前核心終盤切片已接通，boot 起的正式 trace 已從 CTY01 延伸到 CTY02 羅馬利亞：玩家
以正式道具選單使用魔法球，原版 flag `0x51`、四格牆重建、道具消耗、CTY30 tier-1 門、
CTY31 出洞、羅馬利亞國王 rec45→rec15 任務與王座存讀檔均已驗證。
**尚未完成從新遊戲開始的無 debug 全流程驗收**。正式 trace 已由標題完成前置主線，
建立四人隊並以原版初始布衣及國王裝備配置；合法遭遇練至 Lv10、住宿、購物、登上
CTY10，擊敗甘達特×1＋手下×3，完成求饒、取得金皇冠、冒險之書 round-trip，並從
標題「載入進度」正式離塔返回羅馬利亞。下一個 blocker 是 handler9 持皇冠後的
還冠／強制讓位狀態機（EXE file `0x6649` 起）；見 `docs/82`。

遊戲設定將逐批移至 versioned JSON game pack，長期讓同一 Go／Ebitengine core 支援
精訊版 DQ1／DQ2／DQ3。原始 DAT／EXE decoder 必須保留為 parity oracle；JSON 值仍需
D2/D3 證據，未知值不得以可調參數名義猜填。Desktop 可讀外部 pack 免重編譯，
Android／Web 嵌入 pack 則仍需重包。

2026-07-29 起的硬規則：版本專屬 CTY/section/座標、subid、story flag、dialogue record、
formation、寶箱 gate、價格與 cue 不得新增成 Go 常數/table。甘達特事件已是第一個完整
`events.json` + `texts.json` + 通用 `boss_surrender` primitive + EXE/CTY/D3TXT parity
範例；人物初始裝備也由 `characters.json` 提供。細則見 `AGENTS.md` 與 `docs/84`。

## 固定工程方法

1. 每輪閉合一個垂直切片：正式入口 → 狀態交易 → 畫面/聲音 → 下一節點 → save/load。
2. 未知值 fail-closed；禁止用 C remake、經典 DQ3 慣例或「合理值」填 production。
3. 每個參數追原版「誰在何時寫入」，並記錄明確位址口徑。
4. component test 與 production-input trace 分級；只有正常玩家可達才可關閉主線 GAP。
5. 進度只維護 `docs/74`；歷史計畫保留短索引，細節從 Git 歷史取回。

## 工作樹注意

接手先執行 `git status --short`。未追蹤的 Android libs、scratchpad 與 `tools/dosbox_verify_*`、
`tools/tmp_*` 可能是使用者檔案；不得刪除、覆寫或誤納入 commit。
