# 84 — 精訊版 DQ 共用 game pack：JSON 欄位契約

> 狀態：架構契約草案 v0.1（尚未代表 loader 已完成）
>
> 適用範圍：Go／Ebitengine 共用核心，以及未來的 `dq1_cht`、`dq2_cht`、`dq3_cht`
> 資料包。第一個實作者與相容性基準是 `dq3_cht`。

## 1. 目的與邊界

資料包要讓已由原版證實的設定能在不修改 Go 程式、不重新編譯桌面執行檔的情況下修正，
同時避免把猜測值包裝成「可調參數」後失去來源。

```text
Ebitengine 共用核心
  ├─ input / renderer / audio / save / battle state machine
  ├─ game-pack loader + validator + script primitives
  └─ versioned save migration
             │
             ├─ dq1_cht/
             ├─ dq2_cht/
             └─ dq3_cht/
                  ├─ manifest.json
                  ├─ data/*.json
                  └─ 原版合法素材（不進 Git）
```

### 1.1 放進 JSON

- 角色職業、初始能力、成長表、經驗門檻。
- 道具、裝備、咒文、怪物、遭遇群組及戰鬥數值。
- 商店、旅館、教會及其他設施的貨架、費用與服務參數。
- 城鎮／場景索引、入口、warp、寶箱、Boss trigger、事件條件與資料驅動腳本。
- UI record ID、音樂／音效 cue、資源檔名與各代差異。

### 1.2 保留在 Go

- Ebitengine lifecycle、輸入抽象、繪圖、音訊播放、碰撞及存檔 I/O。
- 戰鬥／事件的狀態機與有明確語意的 script primitive。
- schema、交叉引用、範圍、資源存在性及版本相容性驗證。
- DOS 原始格式 decoder。它同時是產生 JSON 與 parity test 的 oracle，不因 JSON 化刪除。

### 1.3 不適合 JSON 化

- 大型 tilemap、sprite、字型、圖片、音樂及原版 binary blob；它們留在原格式或轉成適合的
  binary／媒體格式，由 manifest 引用。
- 任意 Go function 名稱、任意運算式、可執行程式碼或沒有固定語意的 script opcode。
- 尚未由原版證實、只因「玩起來合理」而新增的 production 預設值。

## 2. 目錄與載入順序

建議的資料包：

```text
gamepacks/dq3_cht/
  manifest.json
  data/
    rules.json
    classes.json
    spells.json
    items.json
    monsters.json
    encounters.json
    facilities.json
    worlds.json
    scenes.json
    events.json
    ui.json
    audio.json
  assets/                 # 本機合法原版素材；預設 gitignore
```

載入順序固定為：

1. loader 內建的 schema version 支援表；
2. `manifest.json`；
3. manifest 明列的 JSON；
4. 可選的外部 override 檔；
5. 完整驗證與 canonical content hash；
6. 成功後一次替換 active pack；任一步失敗都不啟動半套資料。

Desktop 應允許 `--game-pack /path/to/dq3_cht` 或等價環境設定讀外部資料，因此改 JSON 不必
重編譯。熱重載只允許在標題畫面或明確安全 checkpoint；戰鬥、事件交易或存檔中途不得替換
資料。Android／Web 若把資料嵌進 APK／WASM，更新仍需重包；若要免重包，須另做已簽章、
具版本及 hash 驗證的外部資料包下載機制。

## 3. 全域格式規則

### 3.1 JSON 語法

- UTF-8、無 BOM、標準 JSON；不依賴註解、尾逗號或鍵順序。
- 整數一律用十進位。原版位址、bit mask 或 raw ID 可用 `"0x001e"` 字串表示，避免
  JavaScript number、正負號及顯示格式歧義。
- loader 必須拒絕未知欄位。格式演進必須提高 `schema_version`，不可悄悄忽略拼錯的鍵。
- 陣列中 ID 不得重複；所有引用必須存在；同一 namespace 內大小寫敏感。

### 3.2 ID

跨遊戲的穩定 ID 使用：

```text
namespace:local_id
```

例如 `dq3:class.hero`、`dq3:item.herb`、`common:service.revive`。`raw_id` 只是精訊版原始
資料的數字索引，不可當跨代 API。DQ1、DQ2 不必建立假的 DQ3 class 或 spell，只需宣告
自己實際支援的 feature。

### 3.3 數值與單位

欄位名稱必須帶出單位或文件明定單位：

- `*_hp`、`*_mp`：整數點數；
- `*_gold`、`*_exp`：非負整數；
- `*_frames`：以固定 simulation tick 計；
- `*_tiles`：地圖格；
- `*_percent` 不使用浮點，改用 `chance_numerator/chance_denominator` 或原版 `roll_max`；
- 原版 byte table 保留原始 roll 語意，不先轉成四捨五入百分比。

### 3.4 缺值、null 與預設值

- 必填欄位缺少即失敗。
- 非必填欄位的預設值必須在本文件逐欄明載；未記載即沒有預設值。
- `null` 只表示 schema 明定的「無此值」，不可同時代表未知、0、沿用或未實作。
- 未知原版值不得填 0。研究資料用 `status: "unknown"` 留在 evidence ledger，production
  pack 在必需值仍未知時應 fail closed。

## 4. 共用欄位

所有具原版對應的 record 可包含：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `id` | string | 是 | 穩定 namespaced ID。 |
| `raw_id` | string | 否 | 原版 table index／ID，以 hex 字串保存；無原始 ID 時省略。 |
| `name_text_id` | string | 視 record | 指向 `ui.json`／文字資源的 ID，不把顯示字串複製到規則表。 |
| `enabled` | boolean | 否 | 預設 `true`；只用於原版確有但該代停用的內容。 |
| `features` | string[] | 否 | 此 record 要求的引擎 capability；預設空陣列。 |
| `evidence` | object | 是 | production 數值的來源。欄位見下表。 |

`evidence`：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `level` | enum | 是 | `D1`、`D2` 或 `D3`；production 初值至少 D2，流程值必須 D3。 |
| `source_kind` | enum | 是 | `exe`、`data_file`、`dosbox`、`video`、`manual`。 |
| `source` | string | 是 | 例如 `DQ3.EXE`、`ITEM.DAT`。 |
| `address_space` | enum | 條件必填 | `file`、`logical`、`dgroup`、`record`、`none`。 |
| `address` | string | 條件必填 | `0x...` 或 `record:...`; `address_space=none` 時省略。 |
| `consumer` | string | D3 必填 | consumer 位址／函式及玩家可見效果摘要。 |
| `doc` | string | 是 | repo 內證據文件路徑與章節。 |
| `note` | string | 否 | 限制或尚未閉合部分，不放推測性 production 值。 |

例：

```json
{
  "id": "common:service.revive",
  "raw_id": "0x0003",
  "evidence": {
    "level": "D3",
    "source_kind": "exe",
    "source": "DQ3.EXE",
    "address_space": "file",
    "address": "0x19dac",
    "consumer": "file 0x85ff..0x8696：扣款、清死亡旗標、HP/MP 回滿",
    "doc": "docs/40-facility-shops.md#教會"
  }
}
```

## 5. `manifest.json`

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `schema_version` | string | 是 | 資料契約版本，例如 `"0.1.0"`。 |
| `pack_id` | string | 是 | 例如 `"dq3_cht"`；只允許小寫 ASCII、數字及底線。 |
| `game` | enum | 是 | `dq1`、`dq2`、`dq3`。 |
| `edition` | string | 是 | 本專案使用 `"cht_jingxun"`。 |
| `content_version` | string | 是 | pack 內容版本；不得代替 schema version。 |
| `engine_api` | string | 是 | 相容範圍，例如 `">=0.1.0 <0.2.0"`。 |
| `title_text_id` | string | 是 | 遊戲標題文字 ID。 |
| `entry_event_id` | string | 是 | 新遊戲正式入口事件。 |
| `save_namespace` | string | 是 | 各代存檔隔離鍵。 |
| `capabilities` | string[] | 是 | 例如 `party`、`classes`、`day_night`、`vehicles.ship`。 |
| `data` | object | 是 | logical table 名稱到相對 JSON 路徑；路徑不得離開 pack root。 |
| `assets` | object | 是 | logical asset 名稱、路徑、大小及 hash。 |

最小示意：

```json
{
  "schema_version": "0.1.0",
  "pack_id": "dq3_cht",
  "game": "dq3",
  "edition": "cht_jingxun",
  "content_version": "0.1.0",
  "engine_api": ">=0.1.0 <0.2.0",
  "title_text_id": "dq3:text.title",
  "entry_event_id": "dq3:event.new_game",
  "save_namespace": "dq3_cht",
  "capabilities": ["party", "classes", "day_night", "vehicles.ship", "vehicles.phoenix"],
  "data": {
    "rules": "data/rules.json",
    "classes": "data/classes.json",
    "facilities": "data/facilities.json"
  },
  "assets": {}
}
```

## 6. 規則與內容表

以下是第一版需要穩定的資料表。欄位新增或語意修改必須同步更新本文件、schema version、
loader 測試及至少一個 pack fixture。

### 6.1 `rules.json`

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `max_party_size` | int | 是 | 合法範圍 1–8；DQ3 精訊版應由證據填值。 |
| `max_level` | int | 是 | 等級上限。 |
| `inventory_slots_per_member` | int | 是 | 每角色持有欄位數。 |
| `money_limit_gold` | int | 是 | 金錢上限。 |
| `simulation_hz` | int | 是 | 固定邏輯 tick；改動會影響 replay，相容性需升版。 |
| `battle` | object | 是 | 逃跑、先攻、傷害、狀態與 reward 的具名規則 ID。 |
| `world` | object | 是 | 日夜步數、tile size、遇敵 accumulator 等規則。 |

公式不可直接放任意字串。應引用引擎已註冊且有測試的 `formula_id`，參數另列。例如：

```json
{
  "formula_id": "common:formula.revive_level_table",
  "parameters": {
    "costs_gold": [10, 10, 10, 20]
  }
}
```

### 6.2 `classes.json`

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `classes` | object[] | 是 | 職業 records；DQ1 沒有轉職系統也可只列單一角色 archetype。 |
| `classes[].id` | string | 是 | 職業 ID。 |
| `classes[].raw_id` | string | 否 | 原始 class byte。 |
| `classes[].growth_table` | int[][] | 是 | 每列的能力順序由 `stat_order` 明定。 |
| `classes[].stat_order` | string[] | 是 | 例如 `strength/vitality/agility/hp/mp/intelligence/luck`。 |
| `classes[].exp_thresholds` | int[] | 是 | index 與 `first_level` 對應，必須單調不減。 |
| `classes[].spell_school_ids` | string[] | 否 | 預設空陣列。 |
| `classes[].equip_tags` | string[] | 否 | 裝備允許標籤；不得混用 raw bit mask 而不說明 bit order。 |

### 6.3 `spells.json`

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `spells[].id` | string | 是 | 咒文 ID。 |
| `spells[].mp_cost` | int | 是 | 非負。 |
| `spells[].contexts` | enum[] | 是 | `field`、`battle`。 |
| `spells[].target` | enum | 是 | `self`、`ally_one`、`ally_all`、`enemy_one`、`enemy_group`、`enemy_all`。 |
| `spells[].effect_id` | string | 是 | 共用引擎 effect primitive。 |
| `spells[].parameters` | object | 是 | effect schema 決定的 typed 參數。 |
| `schools[].learn[].level` | int | 是 | 習得等級。 |
| `schools[].learn[].spell_id` | string | 是 | 必須引用存在的 spell。 |

### 6.4 `items.json`

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `items[].id` | string | 是 | 道具 ID。 |
| `items[].category` | enum | 是 | `weapon/armor/shield/helmet/tool/key/quest` 等。 |
| `items[].price_gold` | int | 是 | 原版買價；不可由賣價反推。 |
| `items[].sell_price_gold` | int/null | 是 | 不可販售時為 `null`。 |
| `items[].equip` | object/null | 是 | 攻防值、slot 及允許職業；非裝備為 `null`。 |
| `items[].use` | object/null | 是 | `effect_id`、contexts、消耗時機及 typed parameters。 |
| `items[].flags` | string[] | 是 | quest、cursed、consumable 等具名旗標。 |

### 6.5 `monsters.json` 與 `encounters.json`

怪物欄位至少包含 `hp_base`、`hp_roll`、`exp`、`gold`、能力、抗性、行動表、掉落、sprite
引用及 evidence。`hp_roll` 必須保存原版隨機語意，不只保存觀察到的 max HP。

遭遇分成：

- `regions`：世界格／region byte 到 encounter table；
- `tables`：候選 monster group、原始權重及隊形；
- `groups`：怪物 ID、最小／最大數量、最多群數；
- `rules`：步數 accumulator、地形／晝夜修正及禁止遇敵條件。

validator 必須檢查怪物引用、數量上下限、權重非負、region 完整性及不可達 table。

### 6.6 `facilities.json`

頂層 `service_definitions` 定義跨場景共用的服務與公式；`facilities` 才是各場景的 NPC
設施實例。第一批 loader 已實作並驗證 `common:service.revive`，場景 instance 遷移時再
加入 `facilities`。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `service_definitions[].id` | string | 是 | 共用服務 ID。 |
| `service_definitions[].pricing` | object | 是 | 具名公式、level cap 與完整費用表。 |
| `service_definitions[].evidence` | object | 是 | 公式／表來源；流程費用需 D3。 |
| `facilities[].id` | string | 是 | 場所內的設施實例 ID。 |
| `service_id` | string | 是 | 例如 `common:service.inn`、`common:service.revive`。 |
| `scene_id` | string | 是 | 所在場景。 |
| `npc_selector` | object | 是 | 原始 NPC subtype／record 等可驗證 selector，不只寫座標。 |
| `inventory` | string[] | 條件必填 | 商店貨架 item IDs。 |
| `pricing` | object | 是 | fixed、per_alive_member、level_table 等具名公式與參數。 |
| `dialogue` | object | 是 | 開場、確認、成功、失敗、取消的 text IDs。 |

教會復活表示例：

```json
{
  "id": "dq3:facility.aliahan_church_revive",
  "service_id": "common:service.revive",
  "scene_id": "dq3:scene.aliahan",
  "npc_selector": {"facility_type_raw": "0x0002"},
  "pricing": {
    "formula_id": "common:formula.level_table",
    "costs_gold": [10, 10, 10, 20, 30, 40, 50, 70, 90, 110]
  },
  "dialogue": {
    "confirm_text_id": "dq3:text.church.revive_confirm",
    "success_text_id": "dq3:text.church.revive_success"
  },
  "evidence": {
    "level": "D3",
    "source_kind": "exe",
    "source": "DQ3.EXE",
    "address_space": "file",
    "address": "0x19dac",
    "consumer": "file 0x85ff..0x8696",
    "doc": "docs/40-facility-shops.md#教會"
  }
}
```

實際 DQ3 pack 必須列完整 40 筆，不可因文件示例截短而讓 loader 自動補值。

### 6.7 `worlds.json`、`scenes.json`

`worlds` 保存 world/layer、地圖資源、城鎮入口、交通工具規則及 encounter region 引用。
`scenes` 保存 CTY／section 的資源組、spawn、出口、warp、NPC table、event table、
palette、文字 bank 及 audio cue。

座標一律是 tile 座標，明定原點與軸方向；原始檔若是 pixel 或 byte offset，轉換公式放
evidence note。`spawn` 必須有 `reason`（new_game、town_entry、warp_return 等），不可再把
section header 的某組座標當所有入口的 fallback。

### 6.8 `events.json`

事件使用有限、具型別的 primitive：

```json
{
  "id": "dq3:event.king_first_audience",
  "trigger": {"kind": "region", "scene_id": "dq3:scene.aliahan_throne", "region_id": "audience"},
  "when": [{"op": "flag_is_clear", "flag_id": "dq3:flag.king_reward"}],
  "steps": [
    {"op": "dialogue", "text_id": "dq3:text.king.first_audience"},
    {"op": "give_items", "item_ids": ["dq3:item.club", "dq3:item.herb"]},
    {"op": "add_gold", "amount_gold": 50},
    {"op": "set_flag", "flag_id": "dq3:flag.king_reward"}
  ],
  "transaction": "atomic",
  "evidence": {
    "level": "D3",
    "source_kind": "exe",
    "source": "DQ3.EXE",
    "address_space": "file",
    "address": "0x140a",
    "consumer": "opening runner 至 handler56 的玩家可見交易",
    "doc": "docs/74-ebiten-remake-completion-plan.md#32-開場第一輪-exe-trace-已確認的-mismatch"
  }
}
```

primitive 必須由 engine registry 定義 input/output、失敗語意及存檔效果。禁止 `eval`、
Go symbol、任意跳址或「事件失敗仍部分扣款」。不支援的 opcode 應在載入時失敗。

### 6.9 `ui.json`、`audio.json`

- `ui.json` 將穩定 text ID 映射至原版 bank/record/glyph sequence；保留 raw record，
  避免在規則表散落中文字串。
- `audio.json` 將 cue ID 映射至原版 track/VOC 或合法替代資源，並記錄 loop、優先權、
  scene transition 行為。
- UI layout 若跨三代共用，可用具名 component 參數；逐像素座標必須記錄 canvas、anchor
  與單位，不可依目前視窗縮放值猜測。

## 7. Override 與不重新編譯

外部 override 只覆蓋 schema 明列為 `tunable` 的欄位。原版 fidelity pack 預設不接受
無 evidence 的 override；開發／實驗模式必須在畫面與 log 明示為 modified pack，且產生不同
content hash，避免其 screenshot 或 save 被誤當原版對拍。

第一版不做任意 JSON merge。override 使用明確 patch record：

```json
{
  "schema_version": "0.1.0",
  "base_pack_id": "dq3_cht",
  "base_content_hash": "sha256:...",
  "changes": [
    {
      "record_id": "dq3:class.hero",
      "field": "exp_thresholds",
      "value": [0, 87, 174]
    }
  ]
}
```

loader 必須拒絕不存在欄位、非 tunable 欄位及 base hash 不符。正式 parity test 永遠以
未套 override 的 canonical pack 執行。

## 8. 驗證與測試契約

每一批 hardcode 遷移須同時具備：

1. schema unit test：合法 fixture 可載入，未知欄位與非法值會失敗；
2. reference test：所有 ID、資源、事件 opcode、formula/effect 均可解析；
3. original parity test：JSON 與 EXE／DAT decoder 的同一筆原始資料逐欄相等；
4. transaction test：成功、取消、資源不足、重複觸發不產生錯誤副作用；
5. production trace：從正常 checkpoint 以 `InputState` 抵達；
6. save test：存檔記錄 `pack_id/schema_version/content_hash`，不相容時明確拒絕或 migration；
7. visual test：受影響畫面更新 runtime PNG，不能以 JSON snapshot 代替畫面對拍。

## 9. 遷移順序

1. 建立 `internal/gamepack` 型別、嚴格 JSON decoder、validator 及最小 fixture。
2. 遷移教會／旅館費用與商店貨架，因目前 progression trace 正被這些設定阻塞。
3. 遷移職業成長、經驗門檻、咒文及戰鬥常數。
4. 遷移怪物、遭遇及掉落；原始 decoder 繼續作 oracle。
5. 遷移 world location、map block、warp、寶箱、Boss trigger。
6. 最後遷移 event/script；每次只搬一個已由玩家流程閉合的垂直切片。
7. DQ3 從新遊戲到 THE END 通過後，用 DQ1 建第二個最小 pack 驗證共用介面；若 DQ1
   需要改 core，先判斷是缺 capability，不能把 DQ1 資料硬套成 DQ3 概念。

## 10. 變更規則

- 本文件是 JSON 欄位的 canonical Markdown；其他文件連結至此，不另抄易過期欄位表。
- schema 的 breaking change 必須提升 major/minor、提供 migration 說明並更新所有 pack。
- 每個原版數值只應有一個 canonical JSON owner；Go、測試與文件不可再各留一份手抄表。
- converter 可由原始 binary 產生 JSON，但生成結果仍需 review、evidence 與 deterministic diff。
- 「能改 JSON」不是 fidelity 證據；數值是否正確仍由原版 writer／table／consumer 與實機對拍決定。
