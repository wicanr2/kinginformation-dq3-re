# 84 — 精訊版 DQ 共用 game pack：JSON 欄位契約

> 狀態：v0.1 已有嚴格 loader、canonical hash、存檔 pack identity 與逐批擴充的有限事件 primitive；
> 其餘資料表依垂直切片逐批遷移。
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
| `schema_version` | string | 是 | 現行資料契約版本為 `"0.1.27"`。 |
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

`data` 可選擇提供 `audio` 路徑。存在時，檔案必須是同一 schema version 的
`audio.json`，由 loader 嚴格驗證；未提供音效 cue 不得在 Go 端補上版本專屬軌號。
`assets` 的 `size`／`sha256` 是外部原版素材的完整性 gate；引擎讀取 pack 指定的路徑並在
載入時核對，不得把 `TITG.P`、`TIT3.P` 等檔名重新寫成共用引擎常數。

最小示意：

```json
{
  "schema_version": "0.1.27",
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
    "facilities": "data/facilities.json",
    "characters": "data/characters.json",
    "texts": "data/texts.json",
    "events": "data/events.json",
    "interface": "data/interface.json",
    "audio": "data/audio.json"
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

### 6.8 `characters.json` 與 `texts.json`

`characters.json` 保存版本專屬的初始人物資料。引擎只認識語意角色，不認識 DQ3 ID：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `default_refs.new_game_player` | string | 是 | 指向本 pack 的新遊戲玩家 default。 |
| `default_refs.registered_party_member` | string | `party` capability 時必填 | 指向酒場／登錄隊員 default；無隊伍版本可省略。 |
| `defaults[].id` | string | 是 | pack-owned 穩定 ID。 |
| `defaults[].equipment` | object | 是 | `weapon/armor/shield/head`；空槽明寫 `null`，item 0 仍是合法道具。 |
| `defaults[].evidence` | object | 是 | 初值 writer 與 consumer 的 D3 證據。 |
| `party_sprite.asset` | string | `party` capability 時必填 | 隊伍角色圖的 pack asset；引擎不假設檔名。 |
| `party_sprite.entries[]` | object[] | `party` capability 時必填 | `{class_raw,gender_raw,entry_base,evidence}`；每個 class/gender 只能有一筆。 |
| `party_sprite.entries[].entry_base` | int | 是 | 角色圖格式的原始 entry base；保留原始 record 對映，不由共用引擎套公式。 |
| `party_sprite.entries[].evidence` | object | 是 | asset bytes、對映與玩家可見角色的 D3 證據。 |

runtime 將 `null` 解為 `-1` 空槽，不用 `0` 當哨兵。reference validator 只檢查
`default_refs` 是否指向同 pack 的 default，不硬寫 `dq3:` ID。

`texts.json` 是玩家可見文字的 canonical 資料；事件與畫面只引用 text ID：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `definitions[].id` | string | 是 | 穩定 text ID。 |
| `value` | string | 是 | 可讀、可修改的原版繁體中文轉寫，保留換行。 |
| `glyph_codes` | int[] | 是 | 原版字模及控制碼 words；runtime 直接消費，不設 Go 字串 fallback。 |
| `layout` | object | 是 | `dialogue`、`menu_label`、`menu_record` 或 `battle_message`；對話另列 columns／lines_per_page。`menu_record` 是本身含框線／換行的原版 legacy record，不套用共用對話框容量。 |
| `source` | object | 是 | `legacy_record` 的檔名／record，或 `glyph_map` 字模來源。 |
| `evidence` | object | 是 | 玩家可見文字來源與 consumer。 |

DQ3 精訊版的 `party_sprite` 對映由 `DQ3MST.BLS` 與 [`docs/27`](27-bls-character-sprites.md) 定錨；
男／女各職業 entry base 是 pack 資料，不是 Go fallback。原版影片
`dq3_real_video/frames/f000295.jpg`、`f000300.jpg`、`f000900.jpg` 的四人縱列與四欄
H/M/HUD 可作玩家可見 oracle；runtime 對拍圖見
[`party_field_hud.png`](../dq3_remake_ebitan/docs/img/party_field_hud.png)。

目前甘達特 rec84–87、羅馬利亞 rec15／45–52／68–72、精靈女王 rec90／96、
諾亞尼爾全域 rec599、金字塔 D3TXT03 rec86–89、波魯多加 D3TXT04 rec24–28、
諾魯德 D3TXT04 rec87–90、巴哈拉達 D3TXT03 rec86–87／106–120／124、
戰鬥睡眠持續／醒來 D3TXT00 rec349–350／354–355，以及戰鬥行動／結算
rec262、270、317–318、319、326、330、332、335、338、340、346、348、351、353、
361、364、366、372–373、396、399，均已遷入
`data/texts.json`；parity test 逐 word
對原始 D3TXT，修改 pack 會改 canonical content hash，舊存檔不得靜默套用。

### 6.9 `events.json`

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

第一個已實作 primitive 是 `boss_surrender_events`，供「開場對話→指定編隊→勝利道歉→
Yes/No（No 可重問）→接受後清旗標」的原版事件使用。它不是甘達特專用 opcode；
DQ1／DQ2 若有相同交易可直接使用，沒有則不必宣告。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `schema_version` | string | 是 | 必須等於 manifest schema。 |
| `boss_surrender_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id` | string | 是 | 穩定 namespaced event ID。 |
| `[].kind` | enum | 是 | 此表固定為 `boss_surrender`。 |
| `[].trigger.kind` | enum | 是 | 現行固定 `scene_tile_subid`。 |
| `[].trigger.cty_raw` | int | 是 | 原版 CTY number；版本專屬，不得硬寫在 Go。 |
| `[].trigger.section` | int | 是 | CTY section。 |
| `[].trigger.tile_subid` | int | 是 | tile high byte 低五位，合法 0–31。 |
| `[].trigger.tiles` | `{x,y}[]` | 是 | 所有可觸發 tile；至少一格，tile 座標、左上原點。 |
| `[].presence_flag_raw` | int | 是 | 事件/NPC 仍存在時為 set 的原版 story flag。 |
| `[].dialogue_text_ids` | object | 是 | `intro/apology/accept/reject/choice_yes/choice_no` 六個穩定 text ID；實際字串與原版 record 在 `texts.json`。 |
| `[].formation.background_raw` | int | 是 | 原版 formation background byte。 |
| `[].formation.page_raw` | int | 是 | 原版 sprite page byte。 |
| `[].formation.groups` | object[] | 是 | 依原順序列 `{monster_raw_id,count}`；同 monster 的分離 group 不得擅自合併。 |
| `[].formation.raw_bytes_hex` | string | 是 | 原始 formation bytes，供 parity test；引擎不把它當可執行腳本。 |
| `[].clear_flag_raw` | int | 是 | 接受求饒後清除的 flag；v0.1 必須等於 `presence_flag_raw`。 |
| `[].treasure_gates` | object[] | 是 | 可為空；每筆含 `cty_raw/section/tile/item_raw_id/while_flag_set`。 |
| `[].evidence` | object | 是 | 事件入口、交易與可見 consumer 的 D3 證據。 |

DQ3 精訊版甘達特事件示例的 canonical 值位於
`internal/gamepack/packs/dq3_cht/data/events.json`，不在本文複製第二份易過期 JSON。
loader 會拒絕未知欄位、非法 subid／flag／monster/count、不一致 clear flag、空 trigger/
formation 與缺 D3 consumer；production 不提供 Go fallback。

第二個已實作 primitive 是 `temporary_role_events`，供「任務道具交還→強制／可選身分
切換→另一 NPC 以兩層選擇恢復原身分」的流程使用。名稱描述跨遊戲行為，不含羅馬利亞、
金皇冠、國王或 DQ3 固有 ID。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `temporary_role_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id` | string | 是 | 穩定 namespaced event ID。 |
| `[].kind` | enum | 是 | 此表固定為 `temporary_role`。 |
| `[].offer_npc` | object | 是 | `{cty_raw,section,tile:{x,y},handler_raw}`；任務／交物／再次接受的 scripted NPC。 |
| `[].restore_npc` | object | 是 | 同上；active role 時提供恢復流程的 scripted NPC。 |
| `[].pending_flag_raw` | int | 是 | 任務尚待必要道具時為 set 的原版 story flag。 |
| `[].required_item_raw_id` | int | 是 | 交還時只消耗一件的原版 item ID。 |
| `[].normal_role_flag_raw` | int | 是 | 一般身分為 set 的原版 story flag。 |
| `[].active_role_flag_raw` | int | 是 | 臨時身分為 set 的原版 story flag；不得與其他三個 flag 重複。 |
| `[].role_sprite_entry_bases` | object | 是 | `{male,female}`；原版角色圖 entry base，皆須為非負整數。 |
| `[].dialogue_text_ids` | object | 是 | 必須含 `quest_greeting/quest_request/return_praise/forced_offer/forced_reject/accept/later_intro/later_offer/later_reject/restore_intro/continue_role/restore_reconsider/restore_reject/restore_accept/choice_yes/choice_no`，所有引用須存在。 |
| `[].evidence` | object | 是 | 入口、交易、選擇分支、旗標與玩家可見 consumer 的 D3 證據。 |

引擎交易時序固定：必要道具存在才在 `return_praise` 前消耗一件並清 pending flag；接受
對白關閉後才交換 normal／active flags；恢復接受對白關閉後才反向交換。角色圖只由 active
flag 與 pack entry 派生，不在存檔複製 transient sprite state。缺文字、selector 或引用時
整個事件 fail closed，不部分扣道具。DQ3 canonical 範例見 `events.json`，原版證據見
[`docs/82`](82-romaly-king-production-trace.md)。

`temporary_solo_challenges` 描述「指定 NPC 詢問→Yes/No→暫時只保留隊長→世界入口→另一
NPC 還原完全相同隊伍」的有限交易。它不允許 JSON 提供任意隊伍程式，也不把完成旗標的
讀分支誤當 writer：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `temporary_solo_challenges` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定為 `temporary_solo_challenge`。 |
| `[].entry_npc`／`[].return_npc` | object | 是 | `{cty_raw,section,tile:{x,y},handler_raw}`；不得相同。 |
| `[].completed_flag_raw` | int | 是 | 已完成分支的原版 read flag；本 primitive 不負責寫入。 |
| `[].mode_mask_raw` | int | 是 | 原版暫時模式 raw bit anchor；必須為 `1..255`，引擎只實作具名狀態。 |
| `[].solo_world_position` | object | 是 | `{x,y,layer}`；接受對白關閉後的原版世界落點。 |
| `[].return_town_destination` | object | 是 | `{cty_raw,section,x,y}`；返回對白關閉後的正常場景落點。 |
| `[].cave_messages` | object[] | 是 | 至少一筆 `{cty_raw,section,handler_raw,text_id}`；只允許顯示文字。 |
| `[].dialogue_text_ids` | object | 是 | `prompt/accept/reject/completed/return/cave_back/cave_empty/choice_yes/choice_no` 均須存在。 |
| `[].evidence` | object | 是 | 隊伍 count writer／restore consumer、入口、文字與可見隊伍變化的 D3 證據。 |

存檔必須在試煉途中保存 active event ID 與完整離隊角色 records，不能只保存現役單人隊伍；
返回時按原順序恢復。缺 selector、文字或 pack event 一律 fail closed。DQ3 canonical 範例、
完成旗標 writer 的未知狀態及 production trace 見 [`docs/100`](100-lancel-courage-blue-orb-production-trace.md)。

第三個已實作 primitive 是 `quest_item_chain_events`，供「一次性寶物→指定 NPC 原地換物→
指定地點使用結果道具→切換場景旗標與重載」的有限流程使用。它不包含諾亞尼爾、精靈女王
或 DQ3 固有名稱；其他版本有相同交易才能使用。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `quest_item_chain_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id` | string | 是 | 穩定 namespaced event ID。 |
| `[].kind` | enum | 是 | 此表固定為 `quest_item_chain`。 |
| `[].treasure` | object | 是 | `cty_raw/section/tile_subid/event_type_raw/item_raw_id/present_flag_raw`；present flag 採原版 set=可取、clear=已取語意。 |
| `[].exchange.npc` | object | 是 | `{cty_raw,section,tile:{x,y},handler_raw}` 的 scripted NPC selector。 |
| `[].exchange.required_item_raw_id` | int | 是 | 必須等於 treasure item；存在時才可交易。 |
| `[].exchange.granted_item_raw_id` | int | 是 | 直接取代原背包格的結果道具，不刪除後 append。 |
| `[].exchange.before_text_id` | string | 是 | 缺前置道具時的穩定 text ID。 |
| `[].exchange.success_text_id` | string | 是 | 成功換物時的穩定 text ID。 |
| `[].use.item_raw_id` | int | 是 | 必須等於交換所得道具。 |
| `[].use.location_kind` | enum | 是 | v0.1 只支援 `town`，不接受任意條件式。 |
| `[].use.cty_raw` | int | 是 | 成功使用的原版 CTY number。 |
| `[].use.set_flags_raw` | int[] | 是 | 成功後依序 SET；至少一筆，與 clear 陣列不得重複。 |
| `[].use.clear_flags_raw` | int[] | 是 | 成功後依序 CLEAR；至少一筆。 |
| `[].use.success_text_id` | string | 是 | 使用成功的穩定 text ID。 |
| `[].use.reload_scene` | boolean | 是 | v0.1 必須為 true，旗標交易後重載目前 section。 |
| `[].evidence` | object | 是 | treasure、交換、use writer／consumer 與可見結果的 D3 證據。 |

引擎只實作上述固定交易；位置錯誤時不消耗，缺 text/reference 時在任何 inventory／flag
mutation 前失敗即關閉。DQ3 canonical 範例與原版 parity test 見 `events.json`、
[`docs/86`](86-noaniel-awakening-production-trace.md)。

`choice_item_exchange_events` 描述「無道具資訊詢問／持有道具交換」兩組 Yes／No，成功後
原格換物並提交固定世界狀態的有限交易。它不允許任意 callback：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `choice_item_exchange_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定為 `choice_item_exchange`。 |
| `[].npc` | object | 是 | `{cty_raw,section,tile:{x,y},handler_raw}`。 |
| `[].available_flag_raw` | int | 是 | 原版事件可用旗標；clear 時只顯示 after。 |
| `[].required_item_raw_id`／`granted_item_raw_id` | int | 是 | 不同的 `0..255` ID；成功時原地取代第一個命中格。 |
| `[].activate_world_object` | object | 是 | `{object_id,position:{x,y,layer}}`；成功交易啟用指定追蹤世界物件並寫入其座標，不得改寫玩家座標。 |
| `[].set_world_state_mask_raw` | int | 是 | 成功時 OR 的非零 16-bit 原版 mask。 |
| `[].clear_story_flags_raw` | int[] | 是 | 成功時清除；至少一筆。 |
| `[].dialogue_text_ids` | object | 是 | `introduction/interested/hint/offer/success/after/reject/choice_yes/choice_no` 全部必填。 |
| `[].evidence` | object | 是 | choice、inventory writer、世界副作用與文字 record 的 D3 證據。 |

所有成功文字先解析，才提交換物、物件啟用、world-state 與旗標交易；拒絕與缺道具分支不得修改
持久狀態。DQ3 canonical 範例與 production trace 見 [`docs/104`](104-transform-staff-ghost-ship-production-trace.md)。

`tracked_world_objects` 描述可移動、可追蹤且可進入的有限地表物件。其座標屬於存檔狀態，
玩家位置與物件位置必須分離；引擎只提供固定的視窗重建、定位文字與入口 primitive：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `tracked_world_objects` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定物件 ID；kind 固定為 `tracked_world_entrance`。 |
| `[].active_world_state_mask_raw` | int | 是 | 非零原版 world-state mask；未啟用時不繪製、不追蹤、不接受入口。 |
| `[].layer` | int | 是 | 原版地表層；現行有限契約為 `0`。 |
| `[].entrance_cty_raw`／`entrance_section` | int | 是 | 玩家與物件同座標時載入的目的場景。 |
| `[].entry_vehicle_mode` | enum | 是 | 現行只接受 `disembark_ship_if_aboard`，入口時停泊船並切回徒步。 |
| `[].tile_raw_by_rng_low2` | int[4] | 是 | 依原版亂數狀態低二位選取的四個 raw tile；不可改成動畫計時器。 |
| `[].tracker_item_raw_id` | int | 是 | 開啟方向／距離定位流程的原版道具 ID。 |
| `[].tracker_dialogue_text_ids` | object | 是 | `preamble/west/east/north/south` 五個穩定 text ID；距離由引擎綁定數字變數。 |
| `[].relocation` | object | 是 | 固定視窗尺寸、原點偏移、水平／垂直保留邊界及候選座標。 |
| `[].relocation.candidates` | object[] | 是 | 非空 `{x,y,layer}` 陣列；重複項保留原版權重。 |
| `[].evidence` | object | 是 | 啟用 writer、座標表、視窗 consumer、入口與定位文字的 D3 證據。 |

DQ3 幽靈船的 18 筆候選座標、四個 tile、船員之骨 records740–744、動態入口與存檔
round-trip 均有 EXE／CTY／D3TXT parity test；完整證據見 [`docs/104`](104-transform-staff-ghost-ship-production-trace.md)。

`coordinate_item_gate_events` 描述踩入固定世界座標後自動發生的有限道具 gate；不得以 JSON
表達任意條件式或 callback：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `coordinate_item_gate_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定為 `coordinate_item_gate`。 |
| `[].coordinate` | object | 是 | 唯一 `{x,y,layer}`；重複座標失敗即關閉。 |
| `[].active_story_flag_raw` | int | 是 | set 時事件才啟用的原版 story flag。 |
| `[].required_item_raw_id` | int | 是 | 成功分支搜尋的原版道具 ID。 |
| `[].clear_story_flag_raw` | int | 是 | 成功文字完整關閉後清除；現行有限契約必須等於 active flag。 |
| `[].consume_required_item` | boolean | 是 | 現行有限契約必須為 false。 |
| `[].failure_direction`／`failure_steps` | int | 是 | 缺道具時逐 frame 執行的原版方向與正整數步數，不可瞬移。 |
| `[].failure_day_night_steps` | int | 是 | 失敗分支一次加入的非負原版日夜步數。 |
| `[].dialogue_text_ids` | object | 是 | `approach/success` 兩個穩定 text ID；失敗只顯示 approach。 |
| `[].evidence` | object | 是 | 座標入口、flag reader/writer、inventory search、文字與強制移動 consumer 的 D3 證據。 |

DQ3 奧莉薇亞海岬是 canonical 範例；原始 EXE／D3TXT 同值、成功／失敗 branch、正式船路與
save/load 見 [`docs/105`](105-olivia-cape-gaia-sword-production-trace.md)。

不屬於多階段任務的原版一次性寶箱放在 `treasure_events`，不得再新增 Go
`treasures` table 或以「合理值」補欄位：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `treasure_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id` | string | 是 | 穩定 namespaced event ID。 |
| `[].kind` | enum | 是 | 固定為 `treasure`。 |
| `[].treasure` | object | 是 | `cty_raw/section/tile_subid/event_type_raw/item_raw_id/present_flag_raw`；present flag 採原版 set=可取、clear=已取。 |
| `[].evidence` | object | 是 | CTY raw entry、EXE dispatcher／inventory consumer 與玩家可見結果的 D3 證據。 |

`item_use_effects` 將版本專屬 raw item ID 接到引擎的有限具名 primitive；不得用 JSON
放任意條件式、callback 或程式碼：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `item_use_effects` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 namespaced ID；kind 固定為 `item_use`。 |
| `[].item_raw_id` | int | 是 | game pack 原始道具 ID；同一 pack 不可重複。 |
| `[].effect_id` | enum | 是 | 支援 `force_day_night_phase`、`temporary_invisibility`、`reveal_world_map_patch`；未知值 fail closed。 |
| `[].location_kind` | enum | 是 | 日夜與地圖 patch 使用 `overworld`；暫時隱形使用 `any`。不接受任意條件運算式。 |
| `[].day_night_phase` | int | 是 | pack canonical 晝夜相位 `0..3`；DQ3 黑夜為 2。 |
| `[].reset_day_night_steps` | boolean | 是 | 是否依原版 transaction 重設現行 phase step；`force_day_night_phase` 必須為 true，暫時隱形必須為 false。 |
| `[].step_count` | int | 是 | 暫態效果剩餘步數；`temporary_invisibility` 必須大於 0，非步數效果必須為 0。 |
| `[].consume` | boolean | 是 | 成功後是否消耗；黑暗燈為 false，隱形草原版會清除選中物品槽，故為 true。 |
| `[].required_layer`／`[].required_vehicle` | int／enum | patch 必填 | 原版地表層與載具 gate；現行 vehicle 只接受 `ship`。 |
| `[].use_tile` | `{x,y}` | patch 必填 | 唯一成功的原始 world coordinate。 |
| `[].required_story_flag_raw`／`[].required_story_flag_set` | int／boolean | patch 必填 | 有限 story-bit gate；不接受任意布林式。 |
| `[].set_world_state_mask` | int | patch 必填 | 原版持久 world-state bit；必須非零。 |
| `[].clear_story_flags_raw` | int[] | patch 必填 | 成功後依原始 handler clear；必須包含 gate flag。 |
| `[].map_patch` | object | patch 必填 | **持久** world-state consumer 讀檔時重建的 `layer/origin/width/height/tiles_raw`；row-major 長度必須等於寬×高，tile 皆為 0–255。 |
| `[].direct_map_patch` | object | 否 | 道具 handler 當幀直接寫入現行 viewport 的表；欄位同 `map_patch`。若原版 direct writer 與 reload consumer 使用不同 table，兩者必須分開保存；只在成功使用當幀套用，讀檔不得把它當持久資料。 |
| `[].animation_palette_mode_raw`／`[].animation_cycles` | int／int | patch 必填 | 保存原版 transition 參數；不代表引擎已達逐幀 V3。 |
| `[].evidence` | object | 是 | item dispatcher、pointer table、writer、clock/palette consumer 與可見結果的 D3 證據。 |

引擎必須先以 pack selector claim 道具，再進舊相容 dispatcher；否則同一 raw ID 可能被 Go
fallback 重新解釋。location／目前 phase gate 失敗不消耗、不重設 clock，也不得用 Go 預設值
補缺欄位。`temporary_invisibility` 只設定具名暫態效果與步數；它不得直接內嵌守衛座標、
NPC 移動或任意碰撞規則。`reveal_world_map_patch` 只能執行固定 gate、state bit 與有限
row-major tile replacement；讀檔以 `map_patch` 持久表重建，不能用 remake flag 取代。
若 `direct_map_patch` 存在，必須另有 IDA／原版 evidence 證明當幀 writer；不得把 direct 表
誤當成 reload 表。DQ3 canonical 範例見 `events.json`、
[`docs/93`](93-teidon-dark-lamp-production-trace.md) 與
[`docs/96`](96-invisibility-grass-eginbear-production-trace.md)、
[`docs/98`](98-thirsty-pitcher-final-key-production-trace.md)。

`tracking_guard_events` 描述「玩家踏入有限 trigger 後，指定 NPC 沿單一軸追蹤；具名暫態
效果可略過追蹤」的有限 primitive。它不能執行任意腳本，也不能把 DQ3 地圖寫進引擎：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `tracking_guard_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 namespaced ID；kind 固定為 `tracking_guard`。 |
| `[].cty_raw`／`[].section` | int | 是 | 原始場景 selector。 |
| `[].trigger_tiles` | object[] | 是 | 至少一筆 `{tile,tile_subid,handler_raw}`；座標不得重複。 |
| `[].guard` | object | 是 | 有限 `ScriptedNPCSelector`；以原始 CTY／section／起始 tile／handler 身分定位。 |
| `[].track_axis` | enum | 是 | 現行只接受 `horizontal`；未知軸 fail closed。 |
| `[].bypass_effect_id` | enum | 是 | 現行只接受 `temporary_invisibility`；必須對應已驗證的具名引擎效果。 |
| `[].evidence` | object | 是 | trigger、NPC writer、效果 bit consumer、玩家可見阻擋／通過結果的 D3 證據。 |

守衛移動只能在成功走入宣告 trigger 後發生；目標格若被地形或另一 NPC 占用則失敗即關閉。
引擎不把暫時隱形改寫成全域 NPC 穿透：原版玩家移動 consumer 仍會對 NPC stamp 回滾。
DQ3 canonical 範例與正式 trace 見 `events.json` 與 [`docs/96`](96-invisibility-grass-eginbear-production-trace.md)。

`npc_push_rule` 描述原版城鎮碰撞 consumer 對所有 NPC 共用的推動 gate；
`push_puzzle_events` 只描述「玩家踏入有限 trigger 後，原始 handler 以固定 slot／軸值判斷
完成並清 story flag」的有限 primitive。兩者均為 typed 欄位，不接受任意 JSON 表達式：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `npc_push_rule` | object | 是 | 全 pack 共用的 NPC 推動規則；不得省略。 |
| `.ctrl_mask`／`.ctrl_value` | int／int | 是 | 對 runtime NPC control byte 套用的 bit gate；value 不得含 mask 外位元。 |
| `.blocked_by_effect_id` | enum | 是 | 現行只接受 `temporary_invisibility`；對齊原版碰撞 consumer 的提前 return。 |
| `.evidence` | object | 是 | 碰撞 caller、control-bit consumer、座標 writer 與副作用的 D3 證據。 |
| `push_puzzle_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 namespaced ID；kind 固定為 `push_puzzle`。 |
| `[].cty_raw`／`[].section` | int | 是 | 原始場景 selector。 |
| `[].trigger_tiles` | object[] | 是 | 至少一筆 `{tile,tile_subid,handler_raw}`；座標不得重複。 |
| `[].npcs` | object[] | 是 | 至少一筆；每筆保存 `slot`、`initial_tile`、`ctrl_raw`。slot 在 NPC 移動後仍是穩定身分，座標與完整 control byte 只作 parity anchor。 |
| `[].completion_axis` | enum | 是 | 現行只接受 `y`；必須忠實描述原始 handler 的比較方式，不可改成較「合理」的目標配對。 |
| `[].completion_value` | int | 是 | 所有宣告 slot 必須達到的原始軸值。 |
| `[].clear_story_flag_raw` | int | 是 | 完成時清除的原始 story flag；範圍 1–511。 |
| `[].evidence` | object | 是 | NPC runtime record、碰撞 writer、handler consumer、CTY trigger 與正式輸入結果的 D3 證據。 |

引擎可推任何符合全域 `npc_push_rule` 的 runtime NPC；前方出界、地形阻擋、另一 NPC 占格、
ctrl mask 不符或具名暫態效果仍生效時都必須失敗即關閉。`push_puzzle_events` 宣告的 slot
只供完成 handler 判定，不限制共用碰撞規則。完成 handler 清 flag 後，沿用通用 event-tile rebuild 更新
passage；一次性寶物另由 `treasure_events` 宣告。DQ3 canonical 範例與正式 trace 見
`events.json` 與 [`docs/97`](97-eginbear-push-puzzle-production-trace.md)。

取得成功必須在同一幀清 present flag、把 low tile 加一並清 event subid；重新載入場景時由
同一 present flag 重建已開啟外觀。加爾那之塔《領悟之書》canonical 範例與完整正式 trace
見 `events.json` 與 [`docs/91`](91-garuna-satori-book-production-trace.md)。

第四個已實作 primitive 是 `two_step_floor_switch_gates`，供「踩地板開關→Yes/No→第一鍵
武裝→第二鍵成功或陷阱→解鎖事件格與一次性寶物」的有限流程使用。它不包含金字塔、
魔法鑰匙或 DQ3 固有名稱；原版沒有相同行為的版本宣告空陣列。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `two_step_floor_switch_gates` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id` | string | 是 | 穩定 namespaced event ID。 |
| `[].kind` | enum | 是 | 固定為 `two_step_floor_switch_gate`。 |
| `[].cty_raw`／`[].section` | int | 是 | 原版事件場景 selector。 |
| `[].switches` | object[] | 是 | 至少兩筆；每筆為 `tile:{x,y}/tile_subid/handler_raw`。座標不得重複；raw handler 只作 parity anchor，引擎不執行它。 |
| `[].completion_tile_subid` | int | 是 | 已武裝後可完成事件的 subid，必須存在於 `switches`。 |
| `[].clear_flag_raw` | int | 是 | 成功後清除並交給通用 event-tile rebuild 的原版 story flag。 |
| `[].trap_transition_subid_raw` | int | 是 | 原始 CTY transition table subid；供 parity test，不作任意跳址。 |
| `[].trap_destination` | object | 是 | 經原始 transition 驗證的 `{cty_raw,section,x,y}`；引擎只載入此有限目的地。 |
| `[].unlocked_treasure` | object | 是 | `cty_raw/section/tile_subid/event_type_raw/item_raw_id/present_flag_raw`；場景必須與 gate 相同。 |
| `[].dialogue_text_ids` | object | 是 | `prompt/pressed/trap/success/choice_yes/choice_no`，所有引用必須存在。 |
| `[].success_sfx_raw` | int | 是 | 原始成功音效 cue，0–255。 |
| `[].evidence` | object | 是 | 開關入口、暫存狀態、成功／陷阱 consumer 與可見結果的 D3 證據。 |

transient armed state 不進存檔；成功交易持久化的是原版 story flags 與 treasure present
flag。選否不武裝，錯誤第二鍵先顯示 trap 再載宣告目的地，正確第二鍵先清旗標並以通用
event-tile rebuild 更新目前場景。DQ3 canonical 範例與完整證據見 `events.json`、
[`docs/87`](87-pyramid-switch-magic-key-production-trace.md)。

第五個已實作 primitive 是 `staged_vehicle_exchange_events`，供「首次交談分兩段授予任務
道具→後續檢查交換道具→消耗後授予停泊交通工具→完成後固定對話」的有限流程使用。
它不包含波魯多加、國王、黑胡椒或 DQ3 固有名稱；其他版本沒有相同行為時宣告空陣列。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `staged_vehicle_exchange_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id` | string | 是 | 穩定 namespaced event ID。 |
| `[].kind` | enum | 是 | 固定為 `staged_vehicle_exchange`。 |
| `[].npc` | object | 是 | `{cty_raw,section,tile:{x,y},handler_raw}`；正式 scripted NPC selector。 |
| `[].quest_present_flag_raw` | int | 是 | 首次任務仍存在時為 set 的原版 story flag。 |
| `[].granted_item_raw_id` | int | 是 | 首次任務介紹關閉後授予一件的原版 item ID。 |
| `[].exchange_available_flag_raw` | int | 是 | 後續交換仍可進行時為 set 的原版 story flag。 |
| `[].required_item_raw_id` | int | 是 | 成功交換時只消耗一件的原版 item ID。 |
| `[].vehicle.kind` | enum | 是 | 現行只支援 `ship`；未知種類載入失敗。 |
| `[].vehicle.world_position` | object | 是 | `{x,y,layer}`；交通工具授予後的原版地表停泊格。 |
| `[].vehicle.clear_flags_raw` | int[] | 是 | 成功交易後依序清除；至少一筆，不得重複。 |
| `[].dialogue_text_ids` | object | 是 | `quest_intro/item_grant/need_item/success/after` 五個穩定 text ID。 |
| `[].evidence` | object | 是 | 入口、writer、table/state、consumer 與玩家可見副作用的 D3 證據。 |

交易時序固定：`quest_intro` 關閉後才給任務道具、清首次旗標並開啟 `item_grant`；缺交換
道具時只顯示 `need_item`；成功時原地消耗一件、清宣告旗標、建立停泊 vehicle，再顯示
`success`；完成後只顯示 `after`。任何 reference、vehicle kind、layer 或 D3 evidence
不合法時，loader 在 runtime mutation 前失敗即關閉。DQ3 canonical 值與原版證據見
`events.json`、[`docs/50`](50-ship-acquisition.md)。

第六個已實作 primitive 是 `guided_passage_events`，供「指定 NPC 檢查必要道具→NPC
沿固定路徑引路→交回正常玩家輸入→玩家踩 scene tile trigger→對話後 NPC 沿固定路徑
返回→原子更新旗標」的有限流程使用。它不包含諾魯德、國王信或 DQ3 固有 ID。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `guided_passage_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定為 `guided_passage`。 |
| `[].guide_npc`／`[].completed_npc` | object | 是 | 同 CTY／section／handler 的起始與完成 NPC selector。 |
| `[].interaction_tile` | `{x,y}` | 是 | 玩家合法交談格，必須與起始 NPC 正交相鄰。 |
| `[].completion_trigger` | object | 是 | `scene_tile_subid` selector；含 `cty_raw/section/tile_subid/tiles`。引擎只在真實玩家移動後查驗，並以 `special_handlers[tile_subid-1]` 對照 handler。 |
| `[].scene_handler_raw` | int | 是 | 原始 special-handler table 值，僅作有限 dispatcher 與 parity anchor。 |
| `[].required_item_raw_id` | int | 是 | 必要道具；是否消耗由下一欄明載。 |
| `[].consume_required_item` | boolean | 是 | 只在完成交易時依原版決定是否消耗。 |
| `[].guide_present_flag_raw` | int | 是 | 起始 NPC 可見旗標。 |
| `[].animation_pending_flag_raw` | int | 是 | NPC 引路完成、等待玩家踩 trigger 的暫態原版旗標。 |
| `[].completed_flag_raw` | int | 是 | 完成 NPC 可見旗標；三個旗標不得重複。 |
| `[].guide_path`／`[].guide_return_path` | `{x,y}[]` | 是 | 起點相接的正交 waypoint；只描述路徑，不含任意程式。 |
| `[].guide_end_facing`／`[].guide_return_end_facing` | enum | 是 | 各段原始 mode=2 轉身結果：`down/up/left/right`。 |
| `[].step_frames` | int | 是 | 每格固定 simulation frames，合法 1–60。 |
| `[].guide_movement_raw_hex`／`[].guide_return_raw_hex` | string／string[] | 是 | 原始 movement triplets，供 EXE parity test；引擎不執行 raw bytes。 |
| `[].dialogue_text_ids` | object | 是 | `introduction/acceptance/guide_ready/after` 四個穩定 text ID。 |
| `[].evidence` | object | 是 | caller→writer→table/state→consumer 與玩家可見結果的 D3 證據。 |

引擎在第一段 NPC 動畫後必須交回正常輸入，不得自動搬移隊伍；只有玩家實際踩到已驗證
的 tile/subid/handler 三重 selector 才能開始完成對話。缺道具、文字、selector、raw anchor
或 final facing 時在任何 mutation 前失敗即關閉。DQ3 canonical 值、IDA 位址與正式 trace
見 `events.json`、[`docs/88`](88-norud-guided-passage-production-trace.md)。

第七個已實作 primitive 是 `hostage_rescue_events`，供「守衛問答與戰鬥→踩格觸發
NPC movement→牆上開關與俘虜 movement→返回 boss 與求饒→跨 CTY 獎勵 NPC 一次性交物」
的有限流程使用。它不包含巴哈拉達、甘達特、黑胡椒或 DQ3 固有 ID。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `hostage_rescue_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定為 `hostage_rescue`。 |
| `[].guard_npcs`／`[].boss_npcs` | selector[] | 是 | `{cty_raw,section,tile,handler_raw}`；各至少一筆。 |
| `[].guard_present_flag_raw`／`[].boss_present_flag_raw` | int | 是 | 兩階段 NPC 可見性的原版 story flag。 |
| `[].guard_formation`／`[].boss_formation` | formation | 是 | `background_raw/page_raw/groups/raw_bytes_hex`；groups 依原始順序保留。 |
| `[].approach_trigger`／`[].switch_trigger` | trigger | 是 | `scene_tile_subid`、CTY、section、subid 與所有觸發格。 |
| `[].approach_pending_flag_raw`／`[].captive_present_flag_raw` | int | 是 | NPC 動畫階段的原版旗標。 |
| `[].approach_movement` | movement | 是 | `npc/path/end_facing/raw_bytes_hex`；path 為正交 waypoint。 |
| `[].captive_movements` | movement[] | 是 | 至少一段；同上。起點可承接上一段 movement 的終點。 |
| `[].completion_set_flags_raw`／`[].completion_clear_flags_raw` | int[] | 是 | 接受求饒後的原子旗標交易；兩陣列皆不得空或重複。 |
| `[].reward_npc` | selector | 是 | 跨 CTY 的一次性獎勵 NPC；欄位名稱不得綁定特定遊戲道具。 |
| `[].reward_available_flag_raw`／`[].reward_item_raw_id` | int | 是 | set 時可領；成功後清 flag 並給一件 item。實際道具與文字由各 game pack 提供。 |
| `[].step_frames` | int | 是 | 每格 movement simulation frames，合法 1–60。 |
| `[].dialogue_text_ids` | object | 是 | 守衛、救援、開關、重逢、boss、獎勵與 Yes／No 的全部穩定 text ID。 |
| `[].evidence` | object | 是 | caller→writer→table/state→consumer 與玩家可見結果的 D3 證據。 |

引擎只接受上述固定階段；選否不消耗、不改旗標，守衛／boss 戰敗不完成交易，接受 boss
求饒後才原子切換完成旗標。獎勵背包新增成功後才清 availability flag。scene cache 在
story flag 實際改變時失效，避免跨 CTY NPC 沿用事件前清單。缺 selector、文字、formation、
movement anchor 或 D3 evidence 時在任何 mutation 前失敗即關閉。DQ3 canonical 值、
IDA／CTY／D3TXT 證據與正式 trace 見 `events.json`、[`docs/89`](89-baharata-rescue-production-trace.md)。

`item_actions` 定義共用道具動作選單與個人物品容量；玩家可見標籤不得寫在 Go：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `item_actions.personal_inventory_slots` | int | 是 | 每名角色總物品格數；DQ3 精訊版為 8，裝備亦占格。 |
| `item_actions.text_ids` | object | 是 | `use/give/drop` 三個穩定 text ID；DQ3 來源為 D3TXT00 rec421。 |
| `item_actions.evidence` | object | 是 | 選單 consumer 與個人物品 writer／reader 的 D3 證據。 |

`rura_navigation` 保存傳送狀態機使用的版本專屬目的地與載具落點；引擎不得硬寫 CTY
順序、文字 record、特殊 arrival offset 或船座標：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `rura_navigation.destinations` | object[] | 是 | 可為空陣列；順序就是原版目的地 index，不得依 CTY 排序重建。 |
| `[].cty_raw` | int | 是 | 原始目的 CTY；同一 pack 不得重複。 |
| `[].name_record_raw` | int | 是 | 目的地選單的原始文字 record；實際字串仍由文字資源提供。 |
| `[].arrival_offset` | `{x,y}` | 是 | 套在 CTY world location 的玩家落點差值；保留 CTY47 等原版特例。 |
| `[].ship_world_position` | `{x,y,layer}` | 是 | 持有船時由同一目的 index 寫入的正規化船落點；原始第二層 Y 值須由 parity test 驗證轉層，不能由 renderer 猜。 |
| `rura_navigation.evidence` | object | 是 | 目的 index、CTY table、玩家 writer、船 table／writer 與玩家可登船結果的 D3 證據。 |

現行 DQ3 canonical table 來自 EXE file `0x16c00`（DGROUP `0x0ac0`）與 file
`0x16c14`（DGROUP `0x0ad4`）；handler logical `0x4c2b..0x4cf8`。完整證據與
production trace 見 [`docs/99`](99-teidon-final-key-green-orb-production-trace.md)。

`world_entrance_variants` 保存「同一 world coordinate 依 ordered story-flag chain 載入
不同 scene record」的版本專屬資料。這不是一般傳送表，也不能簡化成無順序的
`flag → CTY` map：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `world_entrance_variants` | object[] | 是 | 可為空陣列；同一 `{world_tile,layer}` 不得重複。 |
| `[].id` | string | 是 | 穩定入口 ID；同一 pack 不得重複。 |
| `[].world_tile` | `{x,y}` | 是 | 原版 world coordinate；不得放在 Go 常數表。 |
| `[].layer` | int | 是 | `0` 地表、`1` 下層；不同 layer 不可互相命中。 |
| `[].branches` | object[] | 是 | 非空 ordered branches；引擎由 index 0 起取第一個命中項。 |
| `[].branches[].flag_raw` | int | 是 | 原版 story flag；同一入口不得重複。 |
| `[].branches[].when_flag_set` | bool | 是 | set／clear 命中極性；不得假設所有 branch 相同。 |
| `[].branches[].cty_raw` | int | 是 | 命中時載入的原始 scene record；含 default 在內不得重複。 |
| `[].default_cty_raw` | int | 是 | 所有 branch 都未命中時的 scene record。 |
| `[].evidence` | object | 是 | 座標 reader、flag tests、目的 writer 與 scene loader consumer 的 D3 證據。 |

DQ3 商人聚落 `(210,64)` 的第一個 branch 是 `flag0x23 clear → CTY58`，後三個則是
set gate；若寫成統一「set 即取」會在新遊戲直接載入革命後 CTY83。完整 IDA 證據見
[`docs/102`](102-merchant-settlement-world-entrance-re.md)。

`settlement_founder_events` 與 `settlement_founder_followups` 分別描述建城者交付交易及後續
場景的唯讀對話。引擎只保存離隊角色身分、共用預存所與有限對話序列，不得知道商人職業、
城鎮階段、性別 sprite flag 或黃寶珠：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `settlement_founder_events` | object[] | 是 | 可為空陣列；交付事件的穩定集合。 |
| `[].npc` | selector | 是 | 原版交付 NPC 的 `{cty_raw,section,tile,handler_raw}`。 |
| `[].required_class_raw` | int | 是 | 由 active party 依原版順序搜尋的職業。 |
| `[].require_alive` | bool | 是 | 第一個職業命中者是否必須存活；不可跳過死亡者改找後一人。 |
| `[].completion_flag_raw` | int | 是 | 交易成功後設定的原版完成旗標。 |
| `[].founder_visibility_flags_by_gender_raw` | int[2] | 是 | index `0/1` 對應男／女建城者在後續 CTY 使用的原版可見旗標；交易完成才設定。 |
| `[].shared_storage_capacity` | int | 是 | 原版共用預存所容量；滿載不撤銷交付，而是逐件顯示滿載文字。 |
| `[].dialogue_text_ids` | object | 是 | 兩次確認、拒絕、成功、事後與滿載文字的穩定 text ID。 |
| `settlement_founder_followups` | object[] | 是 | 可為空陣列；只讀取既有 founder／flags，不得修改劇情或物品。 |
| `[].npc` | selector | 是 | 後續 subtype2 NPC 的原版 selector。 |
| `[].required_flags_set_raw`／`required_flags_clear_raw` | int[] | 是 | 兩陣列不得重複或矛盾；依 JSON 順序選第一個完整命中項。 |
| `[].dialogue_text_ids` | string[] | 是 | 非空的阻塞對話序列；所有姓名控制碼綁定已保存建城者，不得回退成勇者姓名。 |
| `[].evidence` | object | 是 | handler、旗標 reader、文字 record 與玩家可見結果的 D3 證據。 |

DQ3 handler48 以 `flag0x4a`（黃寶珠 present flag）選擇 `record42→43` 或只顯示
`record42`。它不直接給道具；黃寶珠仍由 CTY83 event `01 6a 00 4a` 與共同寶箱 consumer
交易。兩條路徑不得合併成「與建城者交談即得到寶珠」。

`conditional_facility_events` 描述原版 subtype2 NPC 的座標條件設施分支。這類 handler
不是一般 `facility` subtype：只有玩家位於指定 Y 座標時才透過原始 section facility
指標表開啟設施，其他位置播放指定的重複對話。引擎只解讀有限的座標比較、facility
索引與文字引用，不得把 CTY、handler 或對話句子寫回 Go：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `conditional_facility_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；`kind` 固定為 `conditional_facility`。 |
| `[].npc` | selector | 是 | `{cty_raw,section,tile,handler_raw}`；需與原始 NPC 完整相符。 |
| `[].required_player_y` | int | 是 | 原版 handler 的玩家 Y gate；不是 NPC Y，也不可用鄰近格猜測。 |
| `[].facility_k` | int | 是 | 命中時送入原始 section facility table 的 byte4 索引。 |
| `[].fallback_text_id` | string | 是 | 未命中座標 gate 時的可重複對話 text ID。 |
| `[].evidence` | object | 是 | compare、分支 caller、facility consumer、文字 record 與玩家可見結果的 D3 證據。 |

DQ3 CTY59 handler41 在 IDA linear `0x15bba`（file `0x6f2a`）比較 `DS:4f35==4`；命中
時以 `BX=0` 呼叫共用設施 dispatcher（CTY59 的兩項道具店），否則以 `di=0xbc0`
播放 D3TXT07 record8。這個欄位契約保留兩條路徑，避免把「走到同一 NPC」誤簡化成永遠
開店或永遠只顯示對話。

`npc_item_reward_events` 描述「指定 NPC 依 present flag 顯示成功／事後文字，並把一件
道具交給第一個有空格的隊員」的有限 primitive：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `npc_item_reward_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定 `npc_item_reward`。 |
| `[].npc` | selector | 是 | `{cty_raw,section,tile,handler_raw}` 的正式 subtype2 NPC。 |
| `[].present_flag_raw` | int | 是 | set 時可領取；成功存入背包後才清除。 |
| `[].granted_item_raw` | int | 是 | 原版 item ID；不等同跨遊戲穩定 ID。 |
| `[].dialogue_text_ids` | object | 是 | `grant/after` 兩個穩定 text ID；實際字串不得進 Go。 |
| `[].evidence` | object | 是 | NPC selector、flag gate、item writer、全隊背包 consumer 與玩家可見結果的 D3 證據。 |

全隊背包皆滿時必須不給物、不清 flag；不得把成功文字當成 transaction 成功。DQ3
提頓綠色寶珠範例與完整證據見 `events.json`、`texts.json` 及 [`docs/99`](99-teidon-final-key-green-orb-production-trace.md)。

第八個已實作 primitive 是 `reclass_events`，供「NPC 入口→選隊員→來源／等級 gate→
選目標職業→兩次確認→Lv1 能力與物品交易」使用，不包含達瑪、領悟之書或 DQ3 固有 ID。

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `reclass_events` | object[] | 是 | 可為空陣列；不得省略。 |
| `[].id`／`[].kind` | string／enum | 是 | 穩定 ID；kind 固定為 `reclass`。 |
| `[].npc` | selector | 是 | `{cty_raw,section,tile,handler_raw}` 的正式 scripted NPC。 |
| `[].minimum_level` | int | 是 | 所選角色的最低持久等級。 |
| `[].forbidden_source_classes_raw` | int[] | 是 | 禁止轉職的來源職業；不得與其他 class 清單重複。 |
| `[].basic_target_classes_raw` | int[] | 是 | 原版順序的基本目標職業。 |
| `[].advanced_target_class_raw` | int | 是 | 有額外 gate 的目標職業。 |
| `[].advanced_free_source_class_raw` | int | 是 | 不需 catalyst 即可選 advanced target 的來源職業。 |
| `[].advanced_required_item_raw_id` | int | 是 | 必須由所選角色親自持有的 catalyst item。 |
| `[].effect_id` | enum | 是 | 現行固定 `common:effect.reclass_original`。 |
| `[].personal_inventory_slots` | int | 是 | 交易掃描與卸裝後仍須遵守的個人物品格數。 |
| `[].class_name_text_ids_raw` | object | 是 | raw class 十進位字串鍵到穩定 text ID；必須涵蓋本事件可能讀寫的來源／目標職業，DQ3 完整提供 0..7。 |
| `[].dialogue_text_ids` | object | 是 | rec45–55、基本／advanced class menu 與 Yes／No 的穩定引用。 |
| `[].evidence` | object | 是 | 入口→gate→writer→consumer→玩家可見結果的 D3 證據。 |

共用 effect 固定保留既有咒文、清 EXP／回 Lv1、解除全部裝備，並只對 pack 宣告的欄位
套用原版減半交易；advanced catalyst 只可從所選角色移除。缺 selector、文字、class map、
個人物品或 catalyst 時在 mutation 前失敗即關閉。DQ3 canonical 值、IDA 證據與正式 trace
見 `events.json`、[`docs/92`](92-dhama-reclass-production-trace.md)。

### 6.10 `interface.json`

`interface.json` 保存版本專屬固定視窗幾何；引擎只實作通用繪圖契約。目前必填
`dialogue`，欄位為 `id`、`x`、`y`、`width`、`height`、`text_inset_x`、
`text_inset_y`、`columns`、`lines_per_page` 與 D3 `evidence`。座標與尺寸皆為 640×350
邏輯畫布的 pixel；不得以目前縮放後視窗推算。DQ3 canonical 可見框值為
`(152,238,352,96)`、inset `(16,16)`、20 欄、每頁 4 行；EXE 結構與 consumer 見
[`docs/94`](94-dialogue-window-and-monster-mask-re.md)。缺欄位或無效幾何一律 fail closed。

有 `party` capability 的 pack 另須提供 `party_hud`；除通用 window geometry 外，
`hp_label_glyph`、`mp_label_glyph`、`level_label_glyph` 也必須由 pack 指定。DQ3
精訊版 canonical 為 `(48,244,448,80)`、四欄、四列，對應原版影片的姓名／H／M／等級
狀態窗；缺資料時引擎不畫猜測版面。

能力確認畫面若使用獨立 raw screen，可在同一檔提供 `new_game_confirmation`；引擎只實作
row-interleaved 4bpp decoder 與 modal composition，尺寸、asset key、palette 與證據由 pack
提供：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `new_game_confirmation.id` | string | 是 | 穩定的版本專屬畫面 ID。 |
| `new_game_confirmation.asset_key` | string | 是 | `manifest.assets` 的 raw screen asset key。 |
| `width`／`height` | int | 是 | raw 畫布尺寸；寬度須為 8 的倍數。 |
| `format` | string | 是 | 現行共用 primitive 為 `row_interleaved_4bpp`。 |
| `palette` | RGB[] | 是 | 16 組 8-bit RGB；不可由引擎猜測或套用另一份 palette。 |
| `evidence` | object | 是 | raw bytes、palette、consumer 與原版畫面證據。 |

DQ3 canonical 使用 `FIRST.SCR`（112000 bytes、640×350）與 `docs/title/first_opening.png` 對拍；
能力確認 modal 會在此背景上疊加 stat panels，缺失資料則 fail closed 為黑底，不推導其他版本畫面。

`interface.json.battle_message` 保存戰鬥行動／結算訊息使用的版本專屬外框與文字網格，欄位
同 `dialogue`（`id`、`x`、`y`、`width`、`height`、`text_inset_x`、`text_inset_y`、
`columns`、`lines_per_page`、`evidence`）。DQ3 精訊版由 DQ3.EXE 的 DGROUP `0x3e6e`
共用 `win_rect` 證實為可見 `(152,238,352,96)`、inset `(16,16)`、20 欄／4 行；
`sub_1fb36` 的 `(360,112)` 只是關閉時備份背景範圍，不是玩家看見的框。battle
renderer 只能讀這個 pack layout，缺少資料時 production 建立失敗，不把訊息塞回指令框或
保留另一套 Go 座標。原始 bytes／consumer 見 [`docs/13`](13-exe-battle.md)，runtime 對拍
由 [`docs/108`](108-battle-text-pack.md) 的戰鬥訊息圖更新。

`interface.json.battle_command` 與 `interface.json.battle_enemy` 是戰鬥下方兩個
由原版 `win_rect` 直接消費的 panel。兩者除共用 window geometry 外，還必須指定
`height_mode="rows_plus_base"`、`base_rows`、`row_height` 以及各自的
`cursor_inset_x`、`row_inset_y`、`label_inset_x`、`secondary_label_inset_x`、
`value_inset_x`、`name_inset_x`、`count_inset_x`；這些是資料欄位，不得在 renderer 以
DQ3 座標補值。

DQ3 canonical（皆為 640×350 邏輯 pixel）為：

| panel | 原始定位／可見幾何 | 動態列規則 | 主要 offset |
|---|---|---|---|
| `battle_command` | file `0x1a252`／`DS:0x4112`；`(144,248,128,96)`（四列） | `(menu_rows+2)*16`；三列為 80 高 | cursor `16`、row y `16`、兩欄 x `32`／`80` |
| `battle_enemy` | file `0x1a270`／`byte_28f00`；`(288,248,256,48)`（一列） | `(active_groups+2)*16` | name x/y `32`／`16`、count x `144` |

writer／consumer、raw bytes、IDA linear／file 位址與尚未閉合的 style／baseline 長尾見
[`docs/109`](109-battle-hud-rects-re.md)。pack 缺任一 production panel 時
`NewGameWithPack` fail closed；直接 unit fixture 可以不載 pack，但不代表可執行遊戲有
版本 fallback。

`interface.json.battle_scene` 保存戰鬥場景帶與目標／命令游標的版本資料：
`field_y0`、`field_y1`、`ground_y` 與 `cursor_glyph`。這些欄位只接受具名的
renderer primitive；缺少 `battle_scene` 時 production `NewGameWithPack` fail closed，
不可在 `battle.go` 以另一版的座標或 glyph 補值。DQ3 精訊版目前為
`80..246`、草地基線 `232`、游標 glyph `0x77`（★），證據等級 D2，完整限制見
[`docs/111`](111-battle-scene-layout-re.md)。框線 style pattern、逐動作動畫／停頓／音效
仍不得由這組欄位推成 V3。

標題閒置巡禮若有原版證據，可在同一檔提供 `attract`；引擎只實作輪播狀態機，不能在
Go 內列出職業或檔名：

| 欄位 | 型別 | 必填 | 說明 |
|---|---:|---:|---|
| `attract.id` | string | 是 | 穩定的版本專屬巡禮 ID。 |
| `attract.start_delay_frames` | int | 是 | 標題無輸入後開始巡禮的固定邏輯幀數；非負。 |
| `attract.frames` | object[] | 是 | 有序全螢幕卡片；不得為空或重複 asset key。 |
| `frames[].asset_key` | string | 是 | `manifest.assets` 的 PCX asset key；未知引用 fail closed。 |
| `frames[].hold_frames` | int | 是 | 該卡片停留的固定邏輯幀數；正整數。 |
| `attract.evidence`／`frames[].evidence` | object | 是 | 影片輪播順序／timing 與 PCX 檔案、consumer 的證據。 |

DQ3 canonical `TITH.P`→`TITO.P` 八張職業卡、每卡 1200 個 60Hz frame 的證據與 runtime
對拍見 [`docs/67`](67-attract-assets-re.md) 及 `dq3_remake_ebitan/docs/img/title_attract_warrior.png`。
若其他版本只有標題而沒有經證實巡禮，省略 `attract` 即保持普通標題流程；不得以預設
檔名或合理秒數補上。

`interface.json.battle_texts` 是 production 戰鬥訊息的角色→文字 ID 對映；欄位名稱是
跨版本的 engine role，實際句子、D3TXT record、glyph/control words 與證據仍屬
`texts.json`。目前 DQ3 pack 提供睡眠／醒來、咒文／MP gate、逃跑、攻擊／傷害／沒打中、
勝利／經驗／金錢、狀態／增益／解毒、帕魯朋特及全滅等 25 個角色。每個目標文字的
`layout.kind` 必須為 `battle_message`，renderer 直接消費 `glyph_codes`，不把 Unicode
字串重新猜成字模，也不在 Go 保留玩家可見 fallback；缺少對映、文字或 glyph stream
時，production `NewGameWithPack` fail closed。欄位與原始 record 對照及尚未閉合的動作
停頓／動畫／音效長尾見 [`docs/108`](108-battle-text-pack.md)。

`interface.json.battle_command_labels` 則保存戰鬥下方五種穩定指令角色的雙 glyph
對映：`war`、`flee`、`defend`、`item`、`spell`。每項必須有非負的
`primary_glyph`／`secondary_glyph` 與 evidence；這些數值是 D3TXT00.FON 的原版
字模索引，不是 Unicode 字串，也不是可執行規則。DQ3 的目前對映與 command rect 的
writer／consumer 交叉參照見 [`docs/109`](109-battle-hud-rects-re.md)。正式 pack
缺少這組契約時啟動即拒絕，直接 unit fixture 可省略但不得被當成 production fallback。

`interface.json.field_command_labels` 以同一個雙 glyph 契約保存地表命令窗的標題與六項
指令：`title`、`talk`、`spell`、`status`、`item`、`equip`、`examine`。字模索引與
evidence 仍由 DQ3 pack 擁有；`cmdmenu.go` 只依穩定欄位繪製與處理輸入，不得再放
版本專屬中文或 raw glyph table。正式 pack 缺少這組契約時 `NewGameWithPack` 必須
fail closed；目前 DQ3 對映維持 D2，後續若要升為 D3 仍需補原版 writer／consumer sidecar。

`interface.json.new_game_labels` 保存開場主選單、創角性別／能力確認，以及英數姓名盤
特殊格的 glyph stream。欄位固定為 `title`、`start`、`load`、`male`、`female`、`level`、
`hp`、`mp`、`agility`、`attack`、`defense`、`experience`、`sex`、`hero`、`cloth`、
`prompt`、`yes`、`no`、`backspace`、`ok`，每欄都是非空的 `int[]` 原始 glyph index；
不接受任意 JSON code 或直接玩家可見句子。整個物件必須有 `evidence`，目前 DQ3 為
D2、來源 `D3TXT00.FON + glyph_unicode_map.json`，完整對映與限制見
[`docs/112`](112-newgame-labels-re.md)。

production `NewGameWithPack` 缺少 `new_game_labels` 即拒絕啟動；`NewGameFlow`、
`NameInput`、`GenderSelect` 與酒館只依此契約繪製。這只封存字模資料，主選單／能力確認
面板座標仍須另以原版 writer／consumer 閉合後放入 geometry 契約；不可把目前 renderer
的座標視為已證實的跨版本預設。

### 6.10 `staged_boss_events`

`staged_boss_events` 描述具有兩個正式入口、戰間移動／轉場、選項分支與事後寶箱 gate 的
有限 boss 狀態機。兩戰分別提供 `first_npc`／`second_npc`、原始編隊、掉落政策與勝利旗標
交易；`first_npc_post_battle_directions` 與 `first_post_battle_directions` 分別保存旗標交易前
NPC 及交易後玩家的具名方向，`treasure_gates` 明列第二戰前不可取得的
道具。文字一律使用 `dialogue_text_ids`，不得在 Go 寫版本專屬句子。

掉落政策目前只接受 `original` 與 `suppress`：前者必須走原始怪物資料的掉落欄位，後者對應
原版戰鬥狀態的掉落抑制，不可改成「戰後直接給道具」。每個 selector、編隊、旗標、路徑與
寶箱 gate 都需 D3 證據；缺值或未知方向 fail closed。第一個實例是
`dq3:event.jipang_orochi`，證據見 [`docs/95`](95-jipang-orochi-production-trace.md)。

### 6.11 `ui.json`、`audio.json`

- `ui.json` 將穩定 text ID 映射至原版 bank/record/glyph sequence；保留 raw record，
  避免在規則表散落中文字串。
- `audio.json` 將 cue ID 映射至原版 track/VOC 或合法替代資源，並記錄 loop、優先權、
  scene transition 行為。現行最小結構為：

```json
{
  "schema_version": "0.1.27",
  "cues": {
    "ending": {
      "kind": "music",
      "track": 17,
      "loop": true,
      "priority": 100,
      "evidence": {"level": "D3", "source_kind": "video", "address_space": "none",
        "source": "…", "consumer": "ending sequence", "doc": "docs/61-music-scene-mapping.md"}
    }
  }
}
```

cue 的 `track` 只能存在於 versioned pack；共用引擎以穩定 cue ID 查詢。`ending` 的
DQ3 精訊版值與 `TIT3.P` 終盤畫面同一條正式玩家路徑驗證，不能由畫面名稱或 C prototype
猜填。
- UI layout 若跨三代共用，可用具名 component 參數；逐像素座標必須記錄 canvas、anchor
  與單位，不可依目前視窗縮放值猜測。

## 7. Override 與不重新編譯

外部 override 只覆蓋 schema 明列為 `tunable` 的欄位。原版 fidelity pack 預設不接受
無 evidence 的 override；開發／實驗模式必須在畫面與 log 明示為 modified pack，且產生不同
content hash，避免其 screenshot 或 save 被誤當原版對拍。

第一版不做任意 JSON merge。override 使用明確 patch record：

```json
{
  "schema_version": "0.1.27",
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
