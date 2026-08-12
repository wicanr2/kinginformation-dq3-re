package game

import (
	"fmt"
	"sort"

	"github.com/wicanr2/dq3_remake_ebitan/internal/battle"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	dosrng "github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
)

// 戰鬥指令的內部語意。畫面順序不是固定五項；原版 D3TXT00 rec441..444 依
// 「首位可行動者可逃跑」與「角色是否會咒文」組成 3 或 4 項選單。
const (
	bcWar   = 0 // 戰う
	bcFlee  = 1 // 逃げる
	bcDef   = 2 // 防御
	bcItem  = 3 // 道具(藥草)
	bcSpell = 4 // 咒文
)

const herbCode = 0x41 // 藥草 item id
const herbHeal = 30   // DQ3_HERB_HEAL

// MaxEnemies:一場戰鬥敵群上限(對齊 C MAXE=8、docs/13 encounter_build_group 群量上限)。
const MaxEnemies = 8

// enemyGroup 是原版 formation 的一筆 {怪物 id, 數量}。同一筆共用一次 HP 擲值；
// 不同筆可放不同怪物；版本專屬內容由 game pack 提供。
type enemyGroup struct {
	monID int
	count int
}

// Battle text roles are engine semantics; the referenced text ID and glyph
// stream come from the versioned game pack. Keeping the role names here avoids
// putting a DQ3 record number or player-visible sentence in the engine.
const (
	battleTextStatusSleeping   = "status_sleeping"
	battleTextStatusWoke       = "status_woke"
	battleTextNoSpell          = "no_spell"
	battleTextMPInsufficient   = "mp_insufficient"
	battleTextActorFled        = "actor_fled"
	battleTextActorHealed      = "actor_healed"
	battleTextActorMissed      = "actor_missed"
	battleTextActorDamage      = "actor_damage"
	battleTextActorAttack      = "actor_attack"
	battleTextVictory          = "victory"
	battleTextExpReward        = "exp_reward"
	battleTextGoldReward       = "gold_reward"
	battleTextSpellSealed      = "spell_sealed"
	battleTextEnemySleep       = "enemy_sleep"
	battleTextBuffAttack       = "buff_attack"
	battleTextBuffDefense      = "buff_defense"
	battleTextEnemySealed      = "enemy_sealed"
	battleTextEnemyConfused    = "enemy_confused"
	battleTextCurePoison       = "cure_poison"
	battleTextCureStatus       = "cure_status"
	battleTextSpellNoEffect    = "spell_no_effect"
	battleTextItemEmpty        = "item_empty"
	battleTextPalpunteBlowAway = "palpunte_blow_away"
	battleTextPalpunteDrainMP  = "palpunte_drain_mp"
	battleTextPartyDefeated    = "party_defeated"
)

type battleTextDefinition struct {
	value   string
	glyph   []uint16
	columns int
	lines   int
}

type battleMessage struct {
	role string
	text battleTextDefinition
	vars [][]int
}

// enemyUnit 是群戰中一隻敵的即時狀態。怪物資料、AI 與 sprite 都逐隻保存，讓原版混合
// formation 不會退化成「代表怪物 × N」。
type enemyUnit struct {
	monID       int
	hp, max     int
	mp          int // D3MNS +0x04；帕魯朋特的吸 MP 效果會消耗
	atk, def    int // 有效攻/防(索瑪+光之珠弱化後之值)
	agi         int // D3MNS +0x0b；敵我混排行動順序
	spr         *dq3data.MonsterSprite
	positionRaw int  // 原版 active enemy +0x03；pack-owned EGA raw horizontal offset
	fled        bool // 本場已逃走(不計入擊殺數→經驗/金錢排除,對齊 C g_fled)
	status      int  // W3:異常狀態位元(statusParalysis/statusSealed/statusBlind,見下方 const)
}

func (e *enemyUnit) alive() bool { return e.hp > 0 && !e.fled }

// 異常狀態位元(docs/data/spell-effects-research.md;W3)。共用於 enemyUnit.status、
// battleActor.status、Battle.heroStatus。睡眠／混亂在 remake 共用 statusParalysis，
// 但不是永久麻痺：原版角色狀態 +0x38 bit0x20 在每次行動前以 roll<=0x64 清除，
// 怪物狀態 bit0x80 則以 roll<0x64 清除；醒來當回合仍不行動。statusSealed/statusBlind
// 為 W3 新增(C 版對應的是「敵方全域」g_party_sealed/g_party_blind，W3 額外做
// 「玩家對單一敵單位」的鏡像，故落在 per-unit 欄位)。
const (
	statusPoison    = 1 << iota // 中毒(既有位元,占位保留;W3 未接毒傷結算)
	statusParalysis             // 睡眠/混亂(拉里荷144、美達巴尼152;敵方鏡像同)
	statusSealed                // 瑪荷頓156:不能施咒
	statusBlind                 // 瑪努莎158:物攻 ~50% 失手
)

type battlePhase int

// curGlyph 是共用 UI renderer 的資料槽，不是版本常數；production bootstrap
// 由 pack 的 battle_scene.cursor_glyph 寫入。零值只會出現在未安裝 pack 的
// lightweight fixture，不能作為正式 fallback。
var curGlyph int

const (
	phCommand     battlePhase = iota // 等玩家下指令
	phSpell                          // 選咒文
	phTargetEnemy                    // 選單體敵方目標
	phTargetAlly                     // 選單體我方目標
	phMessage                        // 顯示訊息(A 推進)
	phEnd                            // 勝/敗/逃(A 關閉戰鬥)
)

// Battle 是一場戰鬥的狀態。
type Battle struct {
	mons            *dq3data.Monsters
	shp             []byte
	mpal            []dq3data.Color // MNSBK.PAL(怪物色盤 + packbg 色盤)
	tx              *dq3data.Text
	nameText        *dq3data.Text                 // D3TXT00 名表(咒文名 rec)
	scr             []byte                        // PACKBG.SCR(戰鬥背景)
	bg              *dq3data.PackBG               // pack 選定頁的可見戰鬥場景帶
	background      gamepack.BattleBackgroundData // archive/layout/default selector，由 pack 提供
	backgroundReady bool
	bgPaletteBank   int

	active                 bool
	monID                  int                    // formation 第一筆怪物 id；單群相容與索瑪判斷用
	spr                    *dq3data.MonsterSprite // formation 第一筆 sprite；舊測試/單群相容用
	enemies                []enemyUnit            // 敵群(N 隻,docs/72 A3;單敵 = 1 元素陣列)
	lightOrb               bool                   // 開戰時持光之珠(索瑪 0x7c 戰用;弱化後清)
	heroHP                 int
	heroMax                int
	heroAtk                int
	heroDef                int
	heroAgi                int
	cursor                 int
	msg                    string // 目前訊息的 pack value；供 log／測試，畫面走 msgData.glyph
	msgData                battleMessage
	messageQueue           []battleMessage // 已結算回合的後續逐筆訊息；msg 是目前顯示項
	phase                  battlePhase
	result                 int // 0 進行中、1 勝、2 敗、3 逃
	gotExp                 int
	gotGold                int
	gotDrop                int  // 原版 D3MNS +0x26；-1 表示無掉落
	suppressDrop           bool // 原版 battle state bit1；劇情第二戰等明確禁止掉落
	rng                    *dosrng.RNG
	flashCol               int // >0:受擊閃光殘餘幀
	defending              bool
	heroHerbs              int   // 開戰時持有藥草數
	usedHerbs              int   // 本戰用掉藥草數(戰後從背包扣)
	heroLevel              int   // 我方等級(敵 AI 逃跑門檻用)
	heroMP                 int   // 目前 MP(施咒消耗)
	heroMaxMP              int   //
	spells                 []int // 已學可施放咒文 rec
	spellCursor            int
	companions             []*battleActor // 同伴(狀態列顯示 + 可各自下令及被鎖定)
	heroStatus             int            // 主角異常狀態位元(statusParalysis 等;166-168 解咒清此欄)
	commandActor           int            // 目前下令者：0=隊長，1..=同伴
	commands               []battleCommand
	resolving              bool // 正在執行已收集命令；避免舊單步入口重複推進整回合
	pending                battleCommand
	targetCursor           int
	targetFrom             battlePhase // 取消目標選擇時回 phCommand 或 phSpell
	messageResume          battlePhase // 非回合訊息關閉後回到的選單；phEnd 表示正常回合訊息
	actionActor            int
	actionTarget           int
	statusSleepingText     string // pack value；僅供 log／舊測試相容
	statusWokeText         string // pack value；僅供 log／舊測試相容
	statusSleepingID       string // battleTextStatusSleeping 的 pack role ID
	statusWokeID           string // battleTextStatusWoke 的 pack role ID
	texts                  map[string]battleTextDefinition
	messageLayout          gamepack.WindowLayout      // pack-owned shared battle message rect
	commandLayout          gamepack.BattlePanelLayout // pack-owned command/selection panel
	enemyLayout            gamepack.BattlePanelLayout // pack-owned enemy name/count panel
	sceneLayout            gamepack.BattleSceneLayout // pack-owned field band and cursor glyph
	sceneLayoutReady       bool
	formationPosition      gamepack.BattleFormationPosition // pack-owned active +0x03 projection
	formationPositionReady bool
	commandLabels          map[int][2]int // pack-owned command glyph pairs

	// per-battle 修正狀態(W3,docs/data/spell-effects-research.md;每場 startGroup 歸零,
	// 對齊 C reset_battle_mods() 類型的每戰暫態修正。151/154 是單體、155 是我方全體，
	// 故隊長與同伴各自持有 atkPct/defPct；partyBlind/partySealed 仍是全隊狀態，
	// 由敵施 158/156 設下,C 版全檔無中途清除 → 持續整場戰鬥(非單回合)。
	heroAtkPct  int  // 隊長物攻修正；拜基魯多為單體
	heroDefPct  int  // 隊長守備修正；史卡拉單體、史克魯多全體
	partyBlind  bool // 我方陷入幻惑(敵施158)→ 我方物攻 ~50% 失手
	partySealed bool // 我方咒文被封(敵施156)→ 我方無法施咒
	enemyAtkPct int  // 敵方物攻%(敵自施151,上限400)
	enemyDefPct int  // 敵方守備%(敵自施154/155,上限300)

	// 設定選單驅動(config.CombatInfo / CombatHurtFx;由 Game 在 start() 前設好,
	// start() 不重置 —— 與既有 lightOrb 同一套「呼叫端先設欄位」慣例)。
	showInfo     bool // false:不畫敵 HP 數字(remake 增強項,C 版無此顯示)
	hurtFxFrames int  // 受擊閃光殘餘幀數(取代硬寫 4;0=不閃,對齊 config.CombatHurtFx 換算)

	hits hitList // 指令窗/咒文子選單可點區塊(draw() 重建,phCommand/phSpell 互斥共用一份)
}

// hurtFxFrames:受傷特效時長 ms → 動畫幀數(~16ms/幀,對齊 60fps 迴圈)。0 或負值 = 不閃。
// 移植自 config.CombatHurtFx 語意(0=關;上限 2000ms)。
func hurtFxFrames(ms int) int {
	if ms <= 0 {
		return 0
	}
	f := ms / 16
	if f < 1 {
		f = 1
	}
	return f
}

// setTextDefinitions installs the validated pack-owned battle text streams.
// Production startup supplies every role; unit tests may construct Battle
// directly, in which case missing roles remain non-rendering technical IDs.
func (b *Battle) setTextDefinitions(defs map[string]battleTextDefinition) {
	b.texts = make(map[string]battleTextDefinition, len(defs))
	for role, def := range defs {
		def.glyph = append([]uint16(nil), def.glyph...)
		b.texts[role] = def
	}
	if def, ok := b.texts[battleTextStatusSleeping]; ok {
		b.statusSleepingID, b.statusSleepingText = battleTextStatusSleeping, def.value
	}
	if def, ok := b.texts[battleTextStatusWoke]; ok {
		b.statusWokeID, b.statusWokeText = battleTextStatusWoke, def.value
	}
}

// setMessageLayout installs the validated pack-owned battle message window.
// Direct Battle fixtures may leave it empty; production startup rejects that
// omission before a player can enter a battle.
func (b *Battle) setMessageLayout(layout gamepack.WindowLayout) {
	b.messageLayout = layout
}

func (b *Battle) setCommandLayout(layout gamepack.BattlePanelLayout) {
	b.commandLayout = layout
}

func (b *Battle) setEnemyLayout(layout gamepack.BattlePanelLayout) {
	b.enemyLayout = layout
}

// setSceneLayout installs the validated pack-owned battle field geometry.
// Direct Battle fixtures may leave it empty; production bootstrap rejects that
// omission before a player can enter a battle.
func (b *Battle) setSceneLayout(layout gamepack.BattleSceneLayout) {
	b.sceneLayout = layout
	b.sceneLayoutReady = layout.ID != ""
	if b.sceneLayoutReady {
		curGlyph = layout.CursorGlyph
	}
}

// setFormationPosition installs the confirmed raw formation projection.  A
// production pack must provide it; direct unit fixtures may omit it and retain
// the historical renderer-only spacing for isolated tests.
func (b *Battle) setFormationPosition(layout gamepack.BattleFormationPosition) {
	b.formationPosition = layout
	b.formationPositionReady = layout.Evidence.Level != ""
}

// setBackgroundSource installs a fully validated pack-owned background
// archive.  Direct Battle fixtures may omit it; production bootstrap must not.
func (b *Battle) setBackgroundSource(background gamepack.BattleBackgroundData, scr []byte, palette []dq3data.Color) bool {
	if background.PageStrideBytes <= 0 || background.PageCount <= 0 ||
		background.PaletteBankCount <= 0 || background.PaletteEntriesPerBank <= 0 ||
		len(scr) != background.PageCount*background.PageStrideBytes ||
		len(palette) != background.PaletteBankCount*background.PaletteEntriesPerBank {
		return false
	}
	b.scr, b.mpal = scr, palette
	b.background, b.backgroundReady = background, true
	b.bg, b.bgPaletteBank = nil, 0
	return true
}

func (b *Battle) selectBackground(selector gamepack.BattleBackgroundSelector) bool {
	if !b.backgroundReady || selector.PageRaw < 0 || selector.PageRaw >= b.background.PageCount ||
		selector.PaletteBankRaw < 0 || selector.PaletteBankRaw >= b.background.PaletteBankCount {
		return false
	}
	bg, ok := dq3data.DecodePackBG(b.scr, selector.PageRaw, dq3data.PackBGFormat{
		PageStrideBytes:  b.background.PageStrideBytes,
		FieldOffsetBytes: b.background.FieldOffsetBytes,
		FieldChunkBytes:  b.background.FieldChunkBytes,
		WidthPixels:      b.background.WidthPixels,
		FieldRows:        b.background.FieldRows,
		PlaneCount:       b.background.PlaneCount,
	})
	if !ok {
		return false
	}
	b.bg, b.bgPaletteBank = bg, selector.PaletteBankRaw
	return true
}

// setCommandLabels installs the version-owned glyph pair for each stable
// command role. Missing pairs are intentionally non-rendering for direct unit
// fixtures; production NewGameWithPack rejects an absent contract first.
func (b *Battle) setCommandLabels(labels gamepack.BattleCommandLabels) {
	b.commandLabels = map[int][2]int{
		bcWar:   {labels.War.PrimaryGlyph, labels.War.SecondaryGlyph},
		bcFlee:  {labels.Flee.PrimaryGlyph, labels.Flee.SecondaryGlyph},
		bcDef:   {labels.Defend.PrimaryGlyph, labels.Defend.SecondaryGlyph},
		bcItem:  {labels.Item.PrimaryGlyph, labels.Item.SecondaryGlyph},
		bcSpell: {labels.Spell.PrimaryGlyph, labels.Spell.SecondaryGlyph},
	}
}

func (b *Battle) showMessage(m battleMessage) {
	b.msgData = m
	b.msg = m.text.value
}

// emitText resolves a stable role to pack-owned glyph/control words. A
// missing role never falls back to a Go sentence; the technical role remains
// in msg so tests/logs can expose the broken contract while the renderer draws
// nothing. Built-in production packs are rejected before this path is reached.
func (b *Battle) emitText(role string, vars ...[]int) {
	m := battleMessage{role: role, text: b.texts[role]}
	m.vars = make([][]int, len(vars))
	for i, v := range vars {
		m.vars[i] = append([]int(nil), v...)
	}
	if m.text.value == "" {
		m.text.value = role
	}
	if b.resolving {
		b.messageQueue = append(b.messageQueue, m)
		return
	}
	b.showMessage(m)
}

// emitStatus keeps direct Battle unit tests that provide only the historical
// status value working, while the production path always uses a pack role.
func (b *Battle) emitStatus(role, legacyValue string, vars ...[]int) {
	if role == "" {
		if legacyValue == "" {
			return
		}
		m := battleMessage{text: battleTextDefinition{value: legacyValue}, vars: vars}
		if b.resolving {
			b.messageQueue = append(b.messageQueue, m)
		} else {
			b.showMessage(m)
		}
		return
	}
	b.emitText(role, vars...)
}

func (b *Battle) actorNameGlyphs(actor int) []int {
	if actor == 0 {
		return append([]int(nil), classNames[0]...)
	}
	if actor-1 >= 0 && actor-1 < len(b.companions) {
		c := b.companions[actor-1]
		if len(c.name) > 0 {
			return append([]int(nil), c.name...)
		}
		if c.class >= 0 && c.class < len(classNames) {
			return append([]int(nil), classNames[c.class]...)
		}
	}
	return nil
}

func (b *Battle) enemyNameGlyphs(index int) []int {
	if index < 0 || index >= len(b.enemies) || b.nameText == nil {
		return nil
	}
	var out []int
	for _, code := range b.nameText.Record(0x258 + b.enemies[index].monID) {
		if code < dq3data.GlyphMax {
			out = append(out, int(code))
		}
	}
	return out
}

// battleActor 是戰鬥中一名同伴的即時數值(隊長為 hero* 欄位,同伴用此)。
type battleActor struct {
	name           []int
	class, level   int
	hp, maxHP      int
	mp, maxMP      int
	atk, def, agi  int
	spells         []int
	defending      bool
	atkPct, defPct int
	status         int // W3:異常狀態位元(statusParalysis 等,見 enemyUnit 旁 const)
}

type battleCommand struct {
	kind   int
	spell  int
	target int // 敵方 index，或我方 actor index；不需目標時為 -1
}

type turnEntry struct {
	enemy bool
	index int
	speed int
}

// pct:val 套用百分比修正(基準 100)。共用於物攻/守備 buff/debuff(拜基魯多/史卡拉等)。
func pct(val, p int) int { return val * p / 100 }

func (b *Battle) roll() int { return b.rng.Next(256) }

// 原版玩家側 IDA 0x1b518..0x1b521（file 0xc888..0xc891）：
// cmp al,0x64; ja still-asleep，因此 100 也會醒來。
func playerStatusWakes(roll int) bool { return roll <= 0x64 }

// 原版怪物側 IDA 0x1a9c9..0x1a9db（file 0xbd39..0xbd4b）：
// cmp al,0x64; jb wake，因此 100 仍維持睡眠。
func enemyStatusWakes(roll int) bool { return roll < 0x64 }

// consumeActorSleep 對齊原版角色行動前的睡眠／混亂 consumer。回傳 true 表示本回合
// 已由異常狀態消耗；即使本次醒來也必須跳過原本指令。
func (b *Battle) consumeActorSleep(actor int) bool {
	if b.actorStatus(actor)&statusParalysis == 0 {
		return false
	}
	if playerStatusWakes(b.roll()) {
		b.clearActorStatus(actor, statusParalysis)
		b.emitStatus(b.statusWokeID, b.statusWokeText, b.actorNameGlyphs(actor))
	} else {
		b.emitStatus(b.statusSleepingID, b.statusSleepingText, b.actorNameGlyphs(actor))
	}
	return true
}

// heroParams 是開戰時由 Game 傳入的主角當前數值(等級推導,見 game.heroStats)。
type heroParams struct {
	level, curHP, maxHP, atk, def, agi int
	herbs                              int   // 持有藥草數(戰鬥中可用)
	mp, maxMP                          int   // 目前/最大 MP
	spells                             []int // 已學可施放咒文 rec
}

// start 開一場單敵戰鬥(monID + 主角數值 + 同伴)。等同 startGroup(monID, 1, …)。
func (b *Battle) start(monID int, seed int64, hp heroParams, comps []*battleActor) bool {
	return b.startGroup(monID, 1, seed, hp, comps)
}

// startGroup 開一場 N 隻同種怪的群戰(docs/72 A3)。count 夾在 [1, MaxEnemies]。跳過空 sprite 的怪。
func (b *Battle) startGroup(monID, count int, seed int64, hp heroParams, comps []*battleActor) bool {
	if count < 1 {
		count = 1
	}
	if count > MaxEnemies {
		count = MaxEnemies
	}
	return b.startFormation([]enemyGroup{{monID: monID, count: count}}, seed, hp, comps)
}

// startFormation 開一場原版混合編隊戰。sub_1AAA1 先以全部群組的
// D3MNS +0x28×count 從 0x26 反向扣除，再由 sub_1AAD5 加 2；每隻
// sub_1AB2C 都各自擲一次 HP，並將 active +0x03 以 weight×2 前進。
func (b *Battle) startFormation(groups []enemyGroup, seed int64, hp heroParams, comps []*battleActor) bool {
	return b.startFormationWithBackground(groups, seed, hp, comps, nil)
}

// startFormationWithBackground runs the same combat construction with an
// optional pack selector for scripted formations.  nil means the pack's
// validated default selector; direct fixtures without a pack remain isolated.
func (b *Battle) startFormationWithBackground(groups []enemyGroup, seed int64, hp heroParams, comps []*battleActor, selector *gamepack.BattleBackgroundSelector) bool {
	if len(groups) == 0 {
		return false
	}
	b.rng = dosrng.New(uint16(seed))
	b.enemies = nil
	b.monID, b.spr = 0, nil
	// Normalize the same group/count sequence that will be materialized below.
	// The original builder budgets every accepted member before it writes the
	// first active record; doing this as a read-only pass keeps the RNG stream
	// unchanged while retaining the raw +0x28 weight contract.
	normalized := make([]enemyGroup, 0, len(groups))
	remaining := MaxEnemies
	formationWeight := 0
	for _, group := range groups {
		if remaining <= 0 {
			break
		}
		count := group.count
		if count < 1 {
			continue
		}
		if count > remaining {
			count = remaining
		}
		st, ok := b.mons.Stat(group.monID)
		if !ok {
			return false
		}
		normalized = append(normalized, enemyGroup{monID: group.monID, count: count})
		formationWeight += count * int(st.SpawnWeight)
		remaining -= count
	}
	if len(normalized) == 0 {
		return false
	}
	positionRaw := 0
	if b.formationPositionReady {
		// sub_1AAA1 starts at raw 0x26, subtracts the complete formation
		// budget, and sub_1AAD5 adds two before the first member.
		positionRaw = b.formationPosition.OriginRaw - formationWeight +
			b.formationPosition.FirstMemberOffsetRaw
	}
	for _, group := range normalized {
		count := group.count
		spr, err := dq3data.DecodeMonsterSprite(b.shp, group.monID)
		if err != nil {
			return false
		}
		st, ok := b.mons.Stat(group.monID)
		if !ok {
			return false
		}
		if len(b.enemies) == 0 {
			b.monID, b.spr = group.monID, spr
		}
		eatk, edef := int(st.Atk), int(st.Def)
		if group.monID == 0x7c && b.lightOrb {
			eatk /= 3
			edef /= 4
			b.lightOrb = false
		}
		for i := 0; i < count; i++ {
			hpRoll := int(st.HPBase) + b.rng.Next(int(st.HPRand)+1)
			b.enemies = append(b.enemies, enemyUnit{
				monID: group.monID, hp: hpRoll, max: hpRoll, mp: int(st.MP),
				atk: eatk, def: edef, agi: int(st.Agi), spr: spr,
				positionRaw: positionRaw,
			})
			if b.formationPositionReady {
				positionRaw += b.formationPosition.WeightStepMultiplier * int(st.SpawnWeight)
			}
		}
	}
	if len(b.enemies) == 0 {
		return false
	}
	if b.backgroundReady {
		selected := b.background.DefaultSelector
		if selector != nil {
			selected = *selector
		}
		if !b.selectBackground(selected) {
			return false
		}
	}
	b.heroMax, b.heroHP = hp.maxHP, hp.curHP
	b.heroAtk, b.heroDef, b.heroAgi, b.heroLevel = hp.atk, hp.def, hp.agi, hp.level
	b.heroHerbs, b.usedHerbs, b.defending = hp.herbs, 0, false
	b.heroMP, b.heroMaxMP, b.spells = hp.mp, hp.maxMP, hp.spells
	b.companions = comps
	b.spellCursor = 0
	b.heroStatus = 0
	b.commandActor = 0
	b.commands = emptyCommands(1 + len(comps))
	b.resolving = false
	b.messageResume = phEnd
	// W3 per-battle 修正狀態歸零(對齊 C reset_battle_mods();百分比基準 100)。
	b.heroAtkPct, b.heroDefPct, b.enemyAtkPct, b.enemyDefPct = 100, 100, 100, 100
	for _, c := range b.companions {
		c.atkPct, c.defPct = 100, 100
	}
	b.partyBlind, b.partySealed = false, false
	b.cursor, b.phase, b.result = 0, phCommand, 0
	b.msg, b.msgData, b.gotExp, b.gotGold, b.gotDrop, b.suppressDrop = "", battleMessage{}, 0, 0, -1, false
	b.messageQueue = nil
	b.active = true
	b.seekCommandActor(0)
	return true
}

func (b *Battle) actorAlive(i int) bool {
	if i == 0 {
		return b.heroHP > 0
	}
	return i-1 < len(b.companions) && b.companions[i-1].hp > 0
}

func (b *Battle) actorStatus(i int) int {
	if i == 0 {
		return b.heroStatus
	}
	return b.companions[i-1].status
}

func (b *Battle) actorSpells(i int) []int {
	if i == 0 {
		return b.spells
	}
	return b.companions[i-1].spells
}

func (b *Battle) actorMP(i int) int {
	if i == 0 {
		return b.heroMP
	}
	return b.companions[i-1].mp
}

func (b *Battle) setActorMP(i, v int) {
	if i == 0 {
		b.heroMP = v
		return
	}
	b.companions[i-1].mp = v
}

func (b *Battle) actorHP(i int) (cur, max int) {
	if i == 0 {
		return b.heroHP, b.heroMax
	}
	c := b.companions[i-1]
	return c.hp, c.maxHP
}

func (b *Battle) setActorHP(i, v int) {
	if i == 0 {
		b.heroHP = v
		return
	}
	b.companions[i-1].hp = v
}

func (b *Battle) clearActorStatus(i, bits int) {
	if i == 0 {
		b.heroStatus &^= bits
		return
	}
	b.companions[i-1].status &^= bits
}

func (b *Battle) actorAtkPct(i int) int {
	if i == 0 {
		return b.heroAtkPct
	}
	return b.companions[i-1].atkPct
}

func (b *Battle) actorDefPct(i int) int {
	if i == 0 {
		return b.heroDefPct
	}
	return b.companions[i-1].defPct
}

func (b *Battle) addActorAtkPct(i, add int) {
	if i == 0 {
		b.heroAtkPct += add
		if b.heroAtkPct > 400 {
			b.heroAtkPct = 400
		}
		return
	}
	c := b.companions[i-1]
	c.atkPct += add
	if c.atkPct > 400 {
		c.atkPct = 400
	}
}

func (b *Battle) addActorDefPct(i, add int) {
	if i == 0 {
		b.heroDefPct += add
		if b.heroDefPct > 300 {
			b.heroDefPct = 300
		}
		return
	}
	c := b.companions[i-1]
	c.defPct += add
	if c.defPct > 300 {
		c.defPct = 300
	}
}

// commandMenu 對齊 D3TXT00 rec441..444：
// 有咒/不可逃、無咒/可逃、有咒/可逃、無咒/不可逃。
func (b *Battle) commandMenu(actor int) []int {
	hasSpell := len(b.actorSpells(actor)) > 0
	canFlee := actor == b.firstCommandActor()
	switch {
	case hasSpell && canFlee:
		return []int{bcWar, bcSpell, bcFlee, bcItem} // rec443
	case hasSpell:
		return []int{bcWar, bcSpell, bcDef, bcItem} // rec441
	case canFlee:
		return []int{bcWar, bcFlee, bcDef, bcItem} // rec442
	default:
		return []int{bcWar, bcDef, bcItem} // rec444
	}
}

func (b *Battle) firstCommandActor() int {
	for i := 0; i < 1+len(b.companions); i++ {
		if b.actorAlive(i) && b.actorStatus(i)&statusParalysis == 0 {
			return i
		}
	}
	return 0
}

func (b *Battle) seekCommandActor(from int) bool {
	for i := from; i < 1+len(b.companions); i++ {
		if b.actorAlive(i) && b.actorStatus(i)&statusParalysis == 0 {
			b.commandActor, b.cursor = i, 0
			return true
		}
	}
	return false
}

func (b *Battle) previousCommandActor() bool {
	for i := b.commandActor - 1; i >= 0; i-- {
		if b.actorAlive(i) && b.actorStatus(i)&statusParalysis == 0 {
			b.commandActor, b.cursor = i, 0
			return true
		}
	}
	return false
}

func (b *Battle) queueCommand(cmd battleCommand) {
	b.commands[b.commandActor] = cmd
	if b.seekCommandActor(b.commandActor + 1) {
		b.phase = phCommand
		return
	}
	b.resolveRound()
}

func emptyCommands(n int) []battleCommand {
	out := make([]battleCommand, n)
	for i := range out {
		out[i].target = -1
	}
	return out
}

func (b *Battle) resetCommandRound() {
	b.commands = emptyCommands(1 + len(b.companions))
	b.commandActor, b.cursor, b.spellCursor = 0, 0, 0
	// 隊長死亡且所有仍存活同伴都處於睡眠／混亂時，沒有人可以下令。
	// 不可退回已死亡 actor 0 的命令窗；直接以空命令結算敵方回合，待回合
	// 訊息播放完後再重新判斷是否已有可動隊員。
	if !b.seekCommandActor(0) {
		b.resolveRound()
	}
}

func (b *Battle) aliveEnemyIndices() []int {
	out := make([]int, 0, len(b.enemies))
	for i := range b.enemies {
		if b.enemies[i].alive() {
			out = append(out, i)
		}
	}
	return out
}

func (b *Battle) aliveActorIndices() []int {
	out := make([]int, 0, 1+len(b.companions))
	for i := 0; i < 1+len(b.companions); i++ {
		if b.actorAlive(i) {
			out = append(out, i)
		}
	}
	return out
}

func (b *Battle) beginTarget(cmd battleCommand, from battlePhase) {
	b.pending, b.targetFrom, b.targetCursor = cmd, from, 0
	switch cmd.kind {
	case bcWar:
		b.phase = phTargetEnemy
	case bcItem:
		b.phase = phTargetAlly
	case bcSpell:
		def, ok := spell.GetDef(cmd.spell)
		if !ok {
			b.phase = from
			return
		}
		switch def.Target {
		case spell.TargetEnemyOne:
			b.phase = phTargetEnemy
		case spell.TargetAllyOne:
			b.phase = phTargetAlly
		default:
			b.queueCommand(cmd)
		}
	default:
		b.queueCommand(cmd)
	}
}

func moveTargetCursor(cur, n, dir int) int {
	if n <= 0 {
		return 0
	}
	if dir == 0 || dir == 3 {
		return (cur + 1) % n
	}
	if dir == 1 || dir == 2 {
		return (cur + n - 1) % n
	}
	return cur
}

// firstAliveEnemy 是行動前原目標已死亡時的 fallback；正常玩家輸入會先進 phTargetEnemy。
func (b *Battle) firstAliveEnemy() int {
	for i := range b.enemies {
		if b.enemies[i].alive() {
			return i
		}
	}
	return -1
}

// allEnemiesDead:組內全部 HP=0(含逃走)→ true(勝利判定)。
func (b *Battle) allEnemiesDead() bool {
	for i := range b.enemies {
		if b.enemies[i].hp > 0 {
			return false
		}
	}
	return true
}

// killedCount:組內「被擊殺」隻數(排除逃走者)。
// 只在 allEnemiesDead()==true 時呼叫才有意義。
func (b *Battle) killedCount() int {
	n := 0
	for i := range b.enemies {
		if !b.enemies[i].fled {
			n++
		}
	}
	return n
}

func (b *Battle) victoryRewards() (exp, gold int) {
	for i := range b.enemies {
		if b.enemies[i].fled {
			continue
		}
		st, ok := b.mons.Stat(b.enemies[i].monID)
		if !ok {
			continue
		}
		exp += int(st.Exp)
		gold += int(st.Gold)
	}
	return exp, gold
}

// settleVictoryRewards 對齊 DQ3.EXE logical 0xc509..0xc527：勝利後以 formation 第一隻
// 怪的 D3MNS +0x25 作 rng(256) <= rate，成功取 +0x26 item。劇情可由原版 battle
// state 明確 suppress；不能把 +0x27 誤當 rate。
func (b *Battle) settleVictoryRewards() {
	b.gotExp, b.gotGold = b.victoryRewards()
	b.gotDrop = -1
	if b.suppressDrop || b.mons == nil {
		return
	}
	st, ok := b.mons.Stat(b.monID)
	if !ok || b.roll() > int(st.DropRate) {
		return
	}
	b.gotDrop = int(st.DropItem)
}

func (b *Battle) fleeResist() int {
	resist := 0
	for i := range b.enemies {
		if !b.enemies[i].alive() {
			continue
		}
		if st, ok := b.mons.Stat(b.enemies[i].monID); ok && int(st.FleeResist) > resist {
			resist = int(st.FleeResist)
		}
	}
	return resist
}

// input 處理一幀輸入,回傳戰鬥是否結束(關閉)。點格(P2)= 游標移過去 + 等同 A 選定;
// tapIdx 對照 b.hits(上一幀 draw() 依當前 phase 建的那組:指令 5 格 或 咒文清單)。
func (b *Battle) input(in InputState) (closed bool) {
	tapIdx := -1
	if in.Tapped {
		tapIdx = b.hits.at(in.TapX, in.TapY)
	}
	switch b.phase {
	case phCommand:
		menu := b.commandMenu(b.commandActor)
		confirm := in.Confirm
		if tapIdx >= 0 && tapIdx < len(menu) {
			b.cursor, confirm = tapIdx, true
		}
		switch {
		case confirm:
			kind := menu[b.cursor]
			if kind == bcSpell { // 咒文 → 目前下令者的咒文子選單
				if len(b.actorSpells(b.commandActor)) > 0 {
					b.spellCursor, b.phase = 0, phSpell
				} else {
					b.emitText(battleTextNoSpell, b.actorNameGlyphs(b.commandActor))
					b.phase = phMessage
				}
			} else {
				b.beginTarget(battleCommand{kind: kind, target: -1}, phCommand)
			}
		case in.Cancel:
			b.previousCommandActor()
		case in.DirEdge == 0: // 下
			b.cursor = (b.cursor + 1) % len(menu)
		case in.DirEdge == 1: // 上
			b.cursor = (b.cursor + len(menu) - 1) % len(menu)
		}
	case phSpell:
		spells := b.actorSpells(b.commandActor)
		confirm := in.Confirm
		if tapIdx >= 0 && tapIdx < len(spells) {
			b.spellCursor, confirm = tapIdx, true
		}
		switch {
		case in.Cancel:
			b.phase = phCommand
		case confirm:
			rec := spells[b.spellCursor]
			if def, ok := spell.GetDef(rec); ok && b.actorMP(b.commandActor) < def.MP {
				b.emitText(battleTextMPInsufficient, b.actorNameGlyphs(b.commandActor))
				b.phase, b.messageResume = phMessage, phSpell
			} else {
				b.beginTarget(battleCommand{kind: bcSpell, spell: rec, target: -1}, phSpell)
			}
		case in.DirEdge == 0:
			b.spellCursor = (b.spellCursor + 1) % len(spells)
		case in.DirEdge == 1:
			b.spellCursor = (b.spellCursor + len(spells) - 1) % len(spells)
		}
	case phTargetEnemy:
		targets := b.aliveEnemyIndices()
		if len(targets) == 0 {
			b.resolveRound()
			break
		}
		if tapIdx >= 0 && tapIdx < len(targets) {
			b.targetCursor = tapIdx
			in.Confirm = true
		}
		switch {
		case in.Cancel:
			b.phase = b.targetFrom
		case in.Confirm:
			cmd := b.pending
			cmd.target = targets[b.targetCursor]
			b.queueCommand(cmd)
		default:
			b.targetCursor = moveTargetCursor(b.targetCursor, len(targets), in.DirEdge)
		}
	case phTargetAlly:
		targets := b.aliveActorIndices()
		if len(targets) == 0 {
			b.resolveRound()
			break
		}
		if tapIdx >= 0 && tapIdx < len(targets) {
			b.targetCursor = tapIdx
			in.Confirm = true
		}
		switch {
		case in.Cancel:
			b.phase = b.targetFrom
		case in.Confirm:
			cmd := b.pending
			cmd.target = targets[b.targetCursor]
			b.queueCommand(cmd)
		default:
			b.targetCursor = moveTargetCursor(b.targetCursor, len(targets), in.DirEdge)
		}
	case phMessage:
		if in.Confirm {
			if len(b.messageQueue) > 0 {
				b.showMessage(b.messageQueue[0])
				b.messageQueue = b.messageQueue[1:]
			} else if b.result != 0 {
				b.msg, b.msgData = "", battleMessage{}
				b.phase = phEnd
			} else if b.messageResume != phEnd {
				b.msg, b.msgData = "", battleMessage{}
				b.phase = b.messageResume
				b.messageResume = phEnd
			} else {
				b.msg, b.msgData = "", battleMessage{}
				b.phase = phCommand
				b.resetCommandRound()
			}
		}
	case phEnd:
		if in.Confirm {
			b.active = false
			return true
		}
	}
	return false
}

func (b *Battle) beginMessagePlayback() {
	b.phase = phMessage
	if len(b.messageQueue) == 0 {
		return
	}
	b.showMessage(b.messageQueue[0])
	b.messageQueue = b.messageQueue[1:]
}

// execTurn 執行玩家指令 + 敵方回合,結算。
func (b *Battle) execTurn() {
	if !b.resolving {
		b.commands = emptyCommands(1 + len(b.companions))
		target := b.firstAliveEnemy()
		if b.cursor == bcItem {
			target = 0
		}
		b.commands[0] = battleCommand{kind: b.cursor, target: target}
		for i := range b.companions {
			b.commands[i+1] = battleCommand{kind: bcWar, target: b.firstAliveEnemy()}
		}
		b.resolveRound()
		return
	}
	b.defending = false
	if b.heroStatus&statusParalysis != 0 { // 睡眠/混亂(敵施144/152):本回合任何指令皆無法行動
		b.emitStatus(b.statusSleepingID, b.statusSleepingText, b.actorNameGlyphs(b.actionActor))
		return
	}
	switch b.cursor {
	case bcFlee: // 逃げる
		if battle.FleeOK(b.heroAgi, b.fleeResist(), b.roll()) {
			b.result, b.phase = 3, phMessage
			b.emitText(battleTextActorFled, b.actorNameGlyphs(b.actionActor))
			return
		}
		// D3TXT00 has a confirmed success record (348), but no separately
		// identified original failure sentence. Do not substitute the attack
		// miss record here; an unverified message is worse than silence.
	case bcDef: // 防御(本回合受傷減半)
		b.defending = true
	case bcItem: // 道具:用藥草回 HP
		target := b.actionTarget
		if target < 0 || target >= 1+len(b.companions) || !b.actorAlive(target) {
			target = 0
		}
		cur, max := b.actorHP(target)
		if b.usedHerbs < b.heroHerbs && cur < max {
			heal := herbHeal
			if cur+heal > max {
				heal = max - cur
			}
			b.setActorHP(target, cur+heal)
			b.usedHerbs++
			b.emitText(battleTextActorHealed, b.actorNameGlyphs(target), digitGlyphs(heal))
		} else {
			b.emitText(battleTextItemEmpty)
		}
	default: // 戰う：使用已選敵方目標；若目標先被擊倒則退到下一個存活敵。
		tgt := b.actionTarget
		if tgt < 0 || tgt >= len(b.enemies) || !b.enemies[tgt].alive() {
			tgt = b.firstAliveEnemy()
		}
		if tgt < 0 {
			break // 組內已無存活敵(全數逃走中)→ 揮空,直接進勝負判定
		}
		if b.partyBlind && b.roll() < 128 { // 我方幻惑(敵施158瑪努莎)→ ~50% 揮空(對齊 C g_party_blind)
			b.emitText(battleTextActorMissed)
			break
		}
		crit := 0
		if b.roll() < 8 { // ~1/32 會心
			crit = 1
		}
		atk := pct(b.heroAtk, b.actorAtkPct(0))       // 拜基魯多151
		def := pct(b.enemies[tgt].def, b.enemyDefPct) // 敵154/155自施升守備
		dmg := battle.PhysDamage(atk, def, b.roll(), crit)
		b.enemies[tgt].hp -= dmg
		if b.enemies[tgt].hp < 0 {
			b.enemies[tgt].hp = 0
		}
		b.emitText(battleTextActorDamage, b.enemyNameGlyphs(tgt), digitGlyphs(dmg))
		b.flashCol = b.hurtFxFrames
	}
	if b.allEnemiesDead() {
		b.finishVictory()
		return
	}
}

// execSpell 施放咒文 rec:DMG 傷敵 / HEAL 補己 / W3 新增輔助狀態咒(睡眠/buff/封咒/幻惑/解咒)。
// 瑪荷頓 sealed(敵施封我方咒)在此擋下；被封仍算用掉這次行動。
func (b *Battle) execSpell(rec int) {
	if !b.resolving {
		b.commands = emptyCommands(1 + len(b.companions))
		target := 0
		if def, ok := spell.GetDef(rec); ok && def.Target == spell.TargetEnemyOne {
			target = b.firstAliveEnemy()
		}
		b.commands[0] = battleCommand{kind: bcSpell, spell: rec, target: target}
		for i := range b.companions {
			b.commands[i+1] = battleCommand{kind: bcWar, target: b.firstAliveEnemy()}
		}
		b.resolveRound()
		return
	}
	def, ok := spell.GetDef(rec)
	if !ok {
		b.phase = phCommand
		return
	}
	caster := b.actionActor
	if b.partySealed { // 瑪荷頓(敵施156)→ 我方無法施咒,但仍耗費本回合(對齊 C)
		b.emitText(battleTextSpellSealed, b.actorNameGlyphs(caster))
		return
	}
	if b.actorMP(caster) < def.MP {
		b.phase = phMessage
		b.emitText(battleTextMPInsufficient, b.actorNameGlyphs(caster))
		return
	}
	b.setActorMP(caster, b.actorMP(caster)-def.MP)
	b.defending = false
	switch def.Kind {
	case spell.Heal:
		targets := []int{b.actionTarget}
		if def.Target == spell.TargetAllyGroup {
			targets = b.aliveActorIndices()
		}
		for _, target := range targets {
			if !b.actorAlive(target) {
				continue
			}
			cur, max := b.actorHP(target)
			val := spell.CastValue(def.Base, b.roll())
			if cur+val > max {
				val = max - cur
			}
			b.setActorHP(target, cur+val)
			if val > 0 {
				b.emitText(battleTextActorHealed, b.actorNameGlyphs(target), digitGlyphs(val))
			}
		}
	case spell.Sleep:
		n := b.applyEnemyStatus(def.Target, b.actionTarget, statusParalysis)
		for _, target := range b.aliveEnemyIndices() {
			if b.enemies[target].status&statusParalysis != 0 {
				b.emitText(battleTextEnemySleep, b.enemyNameGlyphs(target))
			}
		}
		if n == 0 {
			b.emitText(battleTextSpellNoEffect)
		}
	case spell.BuffAtk:
		targets := b.allySpellTargets(def.Target, b.actionTarget)
		for _, target := range targets {
			b.addActorAtkPct(target, 100)
			b.emitText(battleTextBuffAttack, b.actorNameGlyphs(target))
		}
	case spell.BuffDef:
		targets := b.allySpellTargets(def.Target, b.actionTarget)
		for _, target := range targets {
			b.addActorDefPct(target, 50)
			b.emitText(battleTextBuffDefense, b.actorNameGlyphs(target), digitGlyphs(50))
		}
	case spell.Seal:
		n := b.applyEnemyStatus(def.Target, b.actionTarget, statusSealed)
		for _, target := range b.aliveEnemyIndices() {
			if b.enemies[target].status&statusSealed != 0 {
				b.emitText(battleTextEnemySealed, b.enemyNameGlyphs(target))
			}
		}
		if n == 0 {
			b.emitText(battleTextSpellNoEffect)
		}
	case spell.Blind:
		n := b.applyEnemyStatus(def.Target, b.actionTarget, statusBlind)
		for _, target := range b.aliveEnemyIndices() {
			if b.enemies[target].status&statusBlind != 0 {
				b.emitText(battleTextEnemyConfused, b.enemyNameGlyphs(target))
			}
		}
		if n == 0 {
			b.emitText(battleTextSpellNoEffect)
		}
	case spell.CurePoison: // 基阿里166
		targets := b.allySpellTargets(def.Target, b.actionTarget)
		for _, target := range targets {
			b.clearActorStatus(target, statusPoison)
			b.emitText(battleTextCurePoison, b.actorNameGlyphs(target))
		}
	case spell.CureStatus: // 基阿里克167/薩梅哈168(remake 簡化:麻痺/混亂/睡眠共用一位元)
		targets := b.allySpellTargets(def.Target, b.actionTarget)
		for _, target := range targets {
			b.clearActorStatus(target, statusParalysis)
			b.emitText(battleTextCureStatus, b.actorNameGlyphs(target))
		}
	case spell.Palpunte:
		b.execPalpunte(b.rng.Next(16))
	default: // Dmg
		if def.Target == spell.TargetEnemyGroup { // 群體咒:命中每隻存活敵,各自獨立擲值
			hit := 0
			for i := range b.enemies {
				if !b.enemies[i].alive() {
					continue
				}
				v := spell.CastValue(def.Base, b.roll())
				b.enemies[i].hp -= v
				if b.enemies[i].hp < 0 {
					b.enemies[i].hp = 0
				}
				hit++
				b.emitText(battleTextActorDamage, b.enemyNameGlyphs(i), digitGlyphs(v))
			}
			b.flashCol = b.hurtFxFrames
			if hit == 0 {
				b.emitText(battleTextSpellNoEffect)
			}
		} else {
			tgt := b.actionTarget
			if tgt < 0 || tgt >= len(b.enemies) || !b.enemies[tgt].alive() {
				tgt = b.firstAliveEnemy()
			}
			if tgt >= 0 {
				val := spell.CastValue(def.Base, b.roll())
				b.enemies[tgt].hp -= val
				if b.enemies[tgt].hp < 0 {
					b.enemies[tgt].hp = 0
				}
				b.emitText(battleTextActorDamage, b.enemyNameGlyphs(tgt), digitGlyphs(val))
			}
			b.flashCol = b.hurtFxFrames
		}
	}
	if b.allEnemiesDead() {
		b.finishVictory()
		return
	}
}

func (b *Battle) allySpellTargets(target spell.Target, selected int) []int {
	if target == spell.TargetAllyGroup {
		return b.aliveActorIndices()
	}
	if selected >= 0 && selected < 1+len(b.companions) && b.actorAlive(selected) {
		return []int{selected}
	}
	if b.actorAlive(b.actionActor) {
		return []int{b.actionActor}
	}
	return nil
}

func (b *Battle) applyEnemyStatus(target spell.Target, selected, bit int) int {
	targets := []int{selected}
	if target == spell.TargetEnemyGroup {
		targets = b.aliveEnemyIndices()
	}
	n := 0
	for _, i := range targets {
		if i < 0 || i >= len(b.enemies) || !b.enemies[i].alive() {
			continue
		}
		b.enemies[i].status |= bit
		n++
	}
	return n
}

// execPalpunte 執行 rec180 的原版 16 格亂數表。
//
// DQ3.EXE file 0xe5d8、DGROUP 0x3908：只有 index 0/2/8/11/12 非 NULL，
// 其中 2 與 8 共用 handler；其餘 11 格顯示「沒有有效」。此函式接受已擲出的
// index，讓資料表契約能以不依賴 RNG seed 的測試鎖定。
func (b *Battle) execPalpunte(effect int) {
	switch effect {
	case 0: // effect type3 → 全敵睡眠；再逐一以 150/256 令存活隊員睡眠
		affected := 0
		for i := range b.enemies {
			if !b.enemies[i].alive() {
				continue
			}
			chance, _ := b.mons.SpellChance(b.enemies[i].monID, 59)
			if b.rng.Next(256) < chance {
				b.enemies[i].status |= statusParalysis
				affected++
				b.emitText(battleTextEnemySleep, b.enemyNameGlyphs(i))
			}
		}
		if b.heroHP > 0 && b.heroStatus&statusParalysis == 0 && b.rng.Next(256) < 0x96 {
			b.heroStatus |= statusParalysis
			affected++
			b.emitText(battleTextStatusSleeping, b.actorNameGlyphs(0))
		}
		for i, c := range b.companions {
			if c.hp > 0 && c.status&statusParalysis == 0 && b.rng.Next(256) < 0x96 {
				c.status |= statusParalysis
				affected++
				b.emitText(battleTextStatusSleeping, b.actorNameGlyphs(i+1))
			}
		}
		if affected == 0 {
			b.emitText(battleTextSpellNoEffect)
		}
	case 2, 8: // BP=0x55 → 每名存活隊員回復 85..94，個別擲值並封頂
		heal := func(cur, max int) int {
			if cur <= 0 {
				return cur
			}
			n := b.rng.Next(95)
			for n < 85 {
				n = b.rng.Next(95)
			}
			cur += n
			if cur > max {
				cur = max
			}
			return cur
		}
		old := b.heroHP
		b.heroHP = heal(b.heroHP, b.heroMax)
		if b.heroHP > old {
			b.emitText(battleTextActorHealed, b.actorNameGlyphs(0), digitGlyphs(b.heroHP-old))
		}
		for i, c := range b.companions {
			old = c.hp
			c.hp = heal(c.hp, c.maxHP)
			if c.hp > old {
				b.emitText(battleTextActorHealed, b.actorNameGlyphs(i+1), digitGlyphs(c.hp-old))
			}
		}
	case 11: // sub_1D6C7：各敵依 spellID59 抗性判定，被吹走者不給 EXP/Gold
		affected := 0
		for i := range b.enemies {
			if !b.enemies[i].alive() {
				continue
			}
			chance, _ := b.mons.SpellChance(b.enemies[i].monID, 59)
			if b.rng.Next(256) < chance {
				b.enemies[i].hp = 0
				b.enemies[i].fled = true
				affected++
				b.emitText(battleTextPalpunteBlowAway, b.enemyNameGlyphs(i))
			}
		}
		if affected == 0 {
			b.emitText(battleTextSpellNoEffect)
		}
	case 12: // effect type1：隨機一敵，吸取 1..min(10,MP)-1（原版 0 會改成 1）
		alive := make([]int, 0, len(b.enemies))
		for i := range b.enemies {
			if b.enemies[i].alive() {
				alive = append(alive, i)
			}
		}
		if len(alive) == 0 {
			b.emitText(battleTextSpellNoEffect)
			return
		}
		tgt := alive[b.rng.Next(len(alive))]
		chance, _ := b.mons.SpellChance(b.enemies[tgt].monID, 59)
		// effect type1 caller file 0xe69a uses JA failure，與其他效果的 JB
		// success 不同；因此 roll==chance 仍成功。
		if b.rng.Next(256) > chance || b.enemies[tgt].mp <= 0 {
			b.emitText(battleTextSpellNoEffect)
			return
		}
		limit := b.enemies[tgt].mp
		if limit > 10 {
			limit = 10
		}
		drain := b.rng.Next(limit)
		if drain == 0 {
			drain = 1
		}
		b.enemies[tgt].mp -= drain
		b.setActorMP(b.actionActor, b.actorMP(b.actionActor)+drain) // 原版不以 max MP 封頂
		b.emitText(battleTextPalpunteDrainMP, b.enemyNameGlyphs(tgt), digitGlyphs(drain))
	default:
		b.emitText(battleTextSpellNoEffect)
	}
}

func (b *Battle) turnSpeed(agi int) int {
	base := agi / 2
	if base < 1 {
		base = 1
	}
	return base + b.rng.Next(base)
}

// resolveRound 對齊原版 sub_d6bf：收完全隊命令後，才把存活隊員與敵人依敏捷擲值混排。
func (b *Battle) resolveRound() {
	b.msg, b.msgData = "", battleMessage{}
	b.messageQueue = nil
	for _, c := range b.companions {
		c.defending = false
	}
	b.defending = false
	order := make([]turnEntry, 0, 1+len(b.companions)+len(b.enemies))
	if b.heroHP > 0 {
		order = append(order, turnEntry{index: 0, speed: b.turnSpeed(b.heroAgi)})
	}
	for i, c := range b.companions {
		if c.hp > 0 {
			order = append(order, turnEntry{index: i + 1, speed: b.turnSpeed(c.agi)})
		}
	}
	for i := range b.enemies {
		if b.enemies[i].alive() {
			order = append(order, turnEntry{enemy: true, index: i, speed: b.turnSpeed(b.enemies[i].agi)})
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return order[i].speed > order[j].speed })

	b.resolving = true
	defer func() { b.resolving = false }()
	for _, e := range order {
		if b.result != 0 {
			break
		}
		if e.enemy {
			ai, aiOK := b.mons.AI(b.enemies[e.index].monID)
			b.enemyAction(e.index, ai, aiOK)
			continue
		}
		if !b.actorAlive(e.index) || b.consumeActorSleep(e.index) {
			continue
		}
		cmd := b.commands[e.index]
		b.actionActor, b.actionTarget = e.index, cmd.target
		if e.index == 0 {
			b.cursor = cmd.kind
			if cmd.kind == bcSpell {
				b.execSpell(cmd.spell)
			} else {
				b.execTurn()
			}
		} else {
			b.execCompanionCommand(e.index-1, cmd)
		}
	}
	if b.result == 0 {
		if b.allEnemiesDead() {
			b.finishVictory()
		} else {
			b.wipedOut()
		}
	}
	b.beginMessagePlayback()
}

func (b *Battle) finishVictory() {
	b.settleVictoryRewards()
	b.result, b.phase = 1, phMessage
	b.emitText(battleTextVictory)
	b.emitText(battleTextExpReward, digitGlyphs(b.gotExp))
	b.emitText(battleTextGoldReward, digitGlyphs(b.gotGold))
}

func (b *Battle) execCompanionCommand(i int, cmd battleCommand) {
	c := b.companions[i]
	switch cmd.kind {
	case bcFlee:
		if battle.FleeOK(c.agi, b.fleeResist(), b.roll()) {
			b.result, b.phase = 3, phMessage
			b.emitText(battleTextActorFled, b.actorNameGlyphs(i+1))
		}
	case bcDef:
		c.defending = true
	case bcItem:
		target := cmd.target
		if target < 0 || target >= 1+len(b.companions) || !b.actorAlive(target) {
			target = i + 1
		}
		cur, max := b.actorHP(target)
		if b.usedHerbs < b.heroHerbs && cur < max {
			heal := herbHeal
			if cur+heal > max {
				heal = max - cur
			}
			b.setActorHP(target, cur+heal)
			b.usedHerbs++
			b.emitText(battleTextActorHealed, b.actorNameGlyphs(target), digitGlyphs(heal))
		} else {
			b.emitText(battleTextItemEmpty)
		}
	case bcSpell:
		b.execSpell(cmd.spell)
	default:
		tgt := cmd.target
		if tgt < 0 || tgt >= len(b.enemies) || !b.enemies[tgt].alive() {
			tgt = b.firstAliveEnemy()
		}
		if tgt < 0 {
			return
		}
		if b.partyBlind && b.roll() < 128 {
			b.emitText(battleTextActorMissed)
			return
		}
		atk := pct(c.atk, b.actorAtkPct(i+1))
		def := pct(b.enemies[tgt].def, b.enemyDefPct)
		dmg := battle.PhysDamage(atk, def, b.roll(), 0)
		b.enemies[tgt].hp -= dmg
		if b.enemies[tgt].hp < 0 {
			b.enemies[tgt].hp = 0
		}
		b.emitText(battleTextActorDamage, b.enemyNameGlyphs(tgt), digitGlyphs(dmg))
		b.flashCol = b.hurtFxFrames
	}
	if b.allEnemiesDead() {
		b.finishVictory()
	}
}

// aliveTargets:存活隊員索引(-1=隊長,0..n=同伴)。
func (b *Battle) aliveTargets() []int {
	var t []int
	if b.heroHP > 0 {
		t = append(t, -1)
	}
	for i, c := range b.companions {
		if c.hp > 0 {
			t = append(t, i)
		}
	}
	return t
}

// damageMember:對目標(-1=隊長,i=同伴)扣血;回是否為隊長被打(供防御減半)。
func (b *Battle) damageMember(target, dmg int) {
	if target < 0 {
		b.heroHP -= dmg
		if b.heroHP < 0 {
			b.heroHP = 0
		}
		return
	}
	b.companions[target].hp -= dmg
	if b.companions[target].hp < 0 {
		b.companions[target].hp = 0
	}
}

func (b *Battle) targetDef(target int) int {
	if target < 0 {
		return b.heroDef
	}
	return b.companions[target].def
}

// actorIndex converts an enemy target representation (-1 for the leader,
// zero-based companion index otherwise) to the shared actor index used by
// name/status helpers (0 leader, 1.. companions).
func actorIndex(target int) int {
	if target < 0 {
		return 0
	}
	return target + 1
}

// setTargetStatus:對目標(-1=隊長,i=同伴)加異常狀態位元(敵鏡像144/152 睡眠/混亂用)。
func (b *Battle) setTargetStatus(target, bit int) {
	if target < 0 {
		b.heroStatus |= bit
		return
	}
	b.companions[target].status |= bit
}

// enemyTurn:敵方回合(移植 C do_turn 敵方段:逐隻存活敵各行動一次——逃跑 → 施咒 → 物攻;
// 逃跑/施咒/物攻皆鎖定隨機存活隊員;任一隻行動後我方全滅即中斷,結算敗;全體行動完畢才收尾訊息)。
// 混合 formation 必須逐隻依 monID 取 AI；同種群自然得到相同資料。
func (b *Battle) enemyTurn() {
	for i := range b.enemies {
		ai, aiOK := b.mons.AI(b.enemies[i].monID)
		b.enemyAction(i, ai, aiOK)
		if b.result != 0 {
			return
		}
	}
	b.phase = phMessage
}

func (b *Battle) enemyAction(i int, ai dq3data.MonsterAI, aiOK bool) {
	if i < 0 || i >= len(b.enemies) || !b.enemies[i].alive() {
		return
	}
	if b.enemies[i].status&statusParalysis != 0 {
		if enemyStatusWakes(b.roll()) {
			b.enemies[i].status &^= statusParalysis
			b.emitStatus(b.statusWokeID, b.statusWokeText, b.enemyNameGlyphs(i))
		} else {
			b.emitText(battleTextEnemySleep, b.enemyNameGlyphs(i))
		}
		return
	}
	if aiOK && ai.FleeRate > 0 && b.heroLevel >= int(ai.FleeThresh) && b.roll() <= int(ai.FleeRate) {
		b.enemies[i].hp, b.enemies[i].fled = 0, true
		b.emitText(battleTextActorFled, b.enemyNameGlyphs(i))
		return
	}
	targets := b.aliveTargets()
	if len(targets) == 0 {
		b.result, b.phase = 2, phMessage
		b.emitText(battleTextPartyDefeated, b.actorNameGlyphs(0))
		return
	}
	tgt := targets[b.rng.Next(len(targets))]
	if aiOK && ai.CastProb > 0 && b.enemies[i].status&statusSealed == 0 && b.roll() < int(ai.CastProb) {
		if bits := spell.MonsterSpellBits(ai.SpellMask); len(bits) > 0 {
			rec := spell.MonsterSpellRec(bits[b.rng.Next(len(bits))])
			if def, ok := spell.GetDef(rec); ok {
				switch def.Kind {
				case spell.Heal:
					val := spell.CastValue(def.Base, b.roll())
					b.enemies[i].hp += val
					if b.enemies[i].hp > b.enemies[i].max {
						b.enemies[i].hp = b.enemies[i].max
					}
					b.emitText(battleTextActorHealed, b.enemyNameGlyphs(i), digitGlyphs(val))
					return
				case spell.Sleep:
					b.setTargetStatus(tgt, statusParalysis)
					b.emitText(battleTextEnemySleep, b.actorNameGlyphs(actorIndex(tgt)))
					return
				case spell.BuffAtk:
					b.enemyAtkPct += 100
					if b.enemyAtkPct > 400 {
						b.enemyAtkPct = 400
					}
					b.emitText(battleTextBuffAttack, b.enemyNameGlyphs(i))
					return
				case spell.BuffDef:
					b.enemyDefPct += 50
					if b.enemyDefPct > 300 {
						b.enemyDefPct = 300
					}
					b.emitText(battleTextBuffDefense, b.enemyNameGlyphs(i), digitGlyphs(50))
					return
				case spell.Seal:
					b.partySealed = true
					b.emitText(battleTextEnemySealed, b.actorNameGlyphs(actorIndex(tgt)))
					return
				case spell.Blind:
					b.partyBlind = true
					b.emitText(battleTextEnemyConfused, b.actorNameGlyphs(actorIndex(tgt)))
					return
				default:
					val := spell.CastValue(def.Base, b.roll())
					b.damageMember(tgt, val)
					b.emitText(battleTextActorDamage, b.actorNameGlyphs(actorIndex(tgt)), digitGlyphs(val))
					b.wipedOut()
					return
				}
			}
		}
	}
	if b.enemies[i].status&statusBlind != 0 && b.roll() < 128 {
		b.emitText(battleTextActorMissed)
		return
	}
	eatk := pct(b.enemies[i].atk, b.enemyAtkPct)
	actor := 0
	if tgt >= 0 {
		actor = tgt + 1
	}
	edef := pct(b.targetDef(tgt), b.actorDefPct(actor))
	edmg := battle.PhysDamage(eatk, edef, b.roll(), 0)
	if tgt < 0 && b.defending || tgt >= 0 && b.companions[tgt].defending {
		edmg /= 2
	}
	b.damageMember(tgt, edmg)
	b.emitText(battleTextActorAttack, b.enemyNameGlyphs(i))
	b.emitText(battleTextActorDamage, b.actorNameGlyphs(actor), digitGlyphs(edmg))
	b.wipedOut()
}

// wipedOut:我方全滅 → 設敗、回 true(呼叫端應立即結束回合,別讓其餘敵繼續行動)。
func (b *Battle) wipedOut() bool {
	if len(b.aliveTargets()) == 0 {
		b.result, b.phase = 2, phMessage
		b.emitText(battleTextPartyDefeated, b.actorNameGlyphs(0))
		return true
	}
	return false
}

// drawName:畫 D3TXT00 記錄 rec 的 glyph 序列(咒名/敵名;非 code+1)。
func (b *Battle) drawName(rgba []byte, x, y, rec int, fg dq3data.Color) {
	if b.nameText == nil {
		return
	}
	col := 0
	for _, v := range b.nameText.Record(rec) {
		if v < dq3data.GlyphMax {
			drawGlyph(rgba, b.nameText, x+col*dq3data.GlyphPx, y, int(v), fg)
			col++
		}
	}
}

// draw 畫整個戰鬥畫面到 rgba,1:1 對齊 C render():上狀態列 / 中怪群站綠地 / 下左指令 + 下右敵名。
func (b *Battle) draw(rgba []byte, scenePal []dq3data.Color) {
	white := dq3data.Color{R: 248, G: 248, B: 248}
	if len(scenePal) > 15 {
		white = scenePal[15]
	}
	red := dq3data.Color{R: 224, G: 48, B: 48}
	cyan := dq3data.Color{R: 64, G: 208, B: 208}
	black := dq3data.Color{}
	sky := dq3data.Color{R: 56, G: 120, B: 216}
	ground := dq3data.Color{R: 40, G: 128, B: 40}
	b.hits.reset()
	if !b.sceneLayoutReady {
		// 直接 Battle fixture 可省略場景資料；正式入口在 bootstrap 時已
		// fail closed，不得在 renderer 內以 DQ3 座標猜測補值。
		return
	}
	fieldY0 := b.sceneLayout.FieldY0
	fieldY1 := b.sceneLayout.FieldY1
	groundY := b.sceneLayout.GroundY
	curGlyph := b.sceneLayout.CursorGlyph

	// 背景:全黑 → pack 選定的場景帶；未安裝 pack 的 direct fixture 才保留純色底。
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			putPx(rgba, x, y, black)
		}
	}
	if b.bg != nil {
		for y := fieldY0; y < fieldY1; y++ {
			sr := (y - fieldY0) * b.bg.Height / (fieldY1 - fieldY0)
			if sr >= b.bg.Height {
				sr = b.bg.Height - 1
			}
			for x := 0; x < ScreenW && x < b.bg.Width; x++ {
				c := black
				idx := int(b.bg.Pixels[sr*b.bg.Width+x])
				if b.backgroundReady {
					idx += b.bgPaletteBank * b.background.PaletteEntriesPerBank
				}
				if idx < len(b.mpal) {
					c = b.mpal[idx]
				}
				putPx(rgba, x, y, c)
			}
		}
	} else { // 純色 fallback(天空 80..232 / 綠地 232..246)
		for y := fieldY0; y < fieldY1; y++ {
			c := sky
			if y >= groundY {
				c = ground
			}
			for x := 0; x < ScreenW; x++ {
				putPx(rgba, x, y, c)
			}
		}
	}
	// 怪群：正式 pack 使用 IDA sub_1B31A 的 EGA destination；直接 Battle
	// fixture 才保留舊的等距 fallback。已死／逃走者不畫，但位置不重排。
	if len(b.enemies) > 0 {
		en := len(b.enemies)
		flash := b.flashCol > 0
		targets := b.aliveEnemyIndices()
		for i := range b.enemies {
			spr := b.enemies[i].spr
			if spr == nil {
				// 舊單群／存檔相容狀態只有 Battle.spr；混合編隊則每隻優先使用自己的 sprite。
				spr = b.spr
			}
			if !b.enemies[i].alive() || spr == nil {
				continue
			}
			gx, gy := 0, 0
			if b.formationPositionReady {
				// IDA sub_1B31A computes
				// stride*(bottom_raw-sprite_height)+active+0x03.  Convert that
				// EGA byte address back to the pack's pixel surface without
				// replacing the raw stride with a guessed screen width.
				stride := b.formationPosition.EGAStrideBytes
				dst := (b.formationPosition.EGABottomRaw-spr.H)*stride + b.enemies[i].positionRaw
				gy = dst / stride
				gx = (dst % stride) * b.formationPosition.PixelsPerRawByte
			} else {
				gx = ScreenW*(i+1)/(en+1) - spr.W/2
				gy = groundY + 8 - spr.H
			}
			if b.phase == phTargetEnemy {
				for pos, target := range targets {
					if target != i {
						continue
					}
					if pos == b.targetCursor {
						drawGlyph(rgba, b.tx, gx+spr.W/2-8, gy-18, curGlyph, white)
					}
					b.hits.add(gx, gy, spr.W, spr.H, pos)
					break
				}
			}
			for r := 0; r < spr.H; r++ {
				for x := 0; x < spr.W; x++ {
					if !spr.Opaque[r][x] {
						continue
					}
					c := dq3data.Color{R: 200, G: 200, B: 200}
					if idx := int(spr.Px[r][x]); idx < len(b.mpal) {
						c = b.mpal[idx]
					}
					if flash {
						c = red
					}
					putPx(rgba, gx+x, gy+r, c)
				}
			}
			// 敵 HP(remake 增強):H + 數字,置中於怪上方。受設定選單「戰鬥資訊」開關(showInfo)控制。
			if b.showInfo {
				cx := gx + spr.W/2
				hy := gy - 32
				if hy < fieldY0 {
					hy = fieldY0
				}
				hc := white
				if b.enemies[i].max > 0 && b.enemies[i].hp*4 < b.enemies[i].max {
					hc = red
				}
				drawGlyph(rgba, b.tx, cx-26, hy, 22, white) // H
				drawNumber(rgba, b.tx, cx-8, hy, b.enemies[i].hp, hc)
			}
		}
		if b.flashCol > 0 { // 閃光倒數:每畫面一次,不隨隻數重複遞減
			b.flashCol--
		}
	}
	// 上方隊伍狀態列(X=0x13*8=152、Y=8;每欄 80px:名 / H+HP / M+MP / 等級;隊長 + 同伴)
	{
		tx0, ty0, colpx := 0x13*8, 8, 0xa*8
		cols := 1 + len(b.companions)
		fillPackBox(rgba, b.messageLayout, tx0-8, ty0-6, cols*colpx+8, 4*16+12)
		drawCol := func(cx int, name []int, hp, mp, level int, alive bool) {
			hc := white
			if !alive {
				hc = red
			}
			col := 0 // 名(class glyph)
			for _, g := range name {
				drawGlyph(rgba, b.tx, cx+col*16, ty0, g, white)
				col++
			}
			drawGlyph(rgba, b.tx, cx, ty0+16, 22, hc) // H
			drawNumber(rgba, b.tx, cx+18, ty0+16, hp, hc)
			drawGlyph(rgba, b.tx, cx, ty0+32, 27, white) // M
			drawNumber(rgba, b.tx, cx+18, ty0+32, mp, white)
			drawNumber(rgba, b.tx, cx+18, ty0+48, level, white)
		}
		var leaderName []int
		if b.heroLevel > 0 { // 隊長名用勇者職業 glyph
			leaderName = classNames[0]
		}
		drawCol(tx0, leaderName, b.heroHP, b.heroMP, b.heroLevel, b.heroHP > 0)
		for i, c := range b.companions {
			drawCol(tx0+(i+1)*colpx, classNames[c.class], c.hp, c.mp, c.level, c.hp > 0)
		}
		if b.phase == phTargetAlly {
			for pos, actor := range b.aliveActorIndices() {
				cx := tx0 + actor*colpx
				if pos == b.targetCursor {
					drawGlyph(rgba, b.tx, cx-16, ty0, curGlyph, white)
				}
				b.hits.add(cx-8, ty0-6, colpx, 4*16+12, pos)
			}
		}
	}
	if b.phase == phMessage {
		w := b.messageLayout
		if w.ID != "" {
			fillPackBox(rgba, w, w.X, w.Y, w.Width, w.Height)
			drawBattleMessage(rgba, b.tx, w.X+w.TextInsetX, w.Y+w.TextInsetY,
				w.Columns, b.msgData, white)
		}
	} else {
		// 下方左:指令窗(原版 sub_c169 的指令／咒文子選單)。production
		// 已在 NewGameWithPack 驗證並安裝 pack panel；缺資料的直接 fixture
		// 不畫此版本專屬面板，避免 renderer 自行捏造座標。
		if w := b.commandLayout; w.ID != "" {
			mx, my, mw, mh := w.X, w.Y, w.Width, w.Height
			cursorInsetX, rowInsetY := w.CursorInsetX, w.RowInsetY
			labelInsetX, secondaryLabelInsetX := w.LabelInsetX, w.SecondaryLabelInsetX
			valueInsetX, rowHeight := w.ValueInsetX, w.RowHeight
			if w.HeightMode == "rows_plus_base" {
				rows := 1
				switch b.phase {
				case phCommand:
					rows = len(b.commandMenu(b.commandActor))
				case phSpell:
					rows = len(b.actorSpells(b.commandActor))
					if rows > 5 {
						rows = 5
					}
				}
				if rows < 1 {
					rows = 1
				}
				mh = (rows + w.BaseRows) * rowHeight
			}
			fillPackBox(rgba, b.commandLayout.WindowLayout, mx, my, mw, mh)
			if b.phase == phSpell {
				spells := b.actorSpells(b.commandActor)
				// 原版視窗只有 5 行；已學咒文超過 5 筆時，讓目前游標保持在可見
				// 範圍內。舊版只畫前 5 筆，游標移到第 6 筆後會成為不可見選項。
				start := 0
				if b.spellCursor >= 5 {
					start = b.spellCursor - 4
				}
				for row, i := 0, start; row < 5 && i < len(spells); row, i = row+1, i+1 {
					rec := spells[i]
					y := my + rowInsetY + row*rowHeight
					if i == b.spellCursor {
						drawGlyph(rgba, b.tx, mx+cursorInsetX, y, curGlyph, white)
					}
					b.drawName(rgba, mx+labelInsetX, y, rec, white)
					if def, ok := spell.GetDef(rec); ok {
						drawNumber(rgba, b.tx, mx+valueInsetX, y, def.MP, cyan)
					}
					b.hits.add(mx, y-4, mw, rowHeight, i)
				}
			} else if b.phase == phTargetEnemy {
				targets := b.aliveEnemyIndices()
				if len(targets) > 0 {
					target := targets[b.targetCursor]
					y := my + rowInsetY
					drawGlyph(rgba, b.tx, mx+cursorInsetX, y, curGlyph, white)
					b.drawName(rgba, mx+labelInsetX, y, 0x258+b.enemies[target].monID, white)
					drawNumber(rgba, b.tx, mx+valueInsetX, y, target+1, cyan)
				}
			} else if b.phase == phTargetAlly {
				targets := b.aliveActorIndices()
				if len(targets) > 0 {
					actor := targets[b.targetCursor]
					y := my + rowInsetY
					drawGlyph(rgba, b.tx, mx+cursorInsetX, y, curGlyph, white)
					name := classNames[0]
					if actor > 0 {
						name = classNames[b.companions[actor-1].class]
					}
					for i, glyph := range name {
						drawGlyph(rgba, b.tx, mx+labelInsetX+i*16, y, glyph, white)
					}
				}
			} else if b.phase == phCommand {
				menu := b.commandMenu(b.commandActor)
				for i, kind := range menu {
					y := my + rowInsetY + i*rowHeight
					if i == b.cursor {
						drawGlyph(rgba, b.tx, mx+cursorInsetX, y, curGlyph, white)
					}
					if label, ok := b.commandLabels[kind]; ok {
						drawGlyph(rgba, b.tx, mx+labelInsetX, y, label[0], white)
						drawGlyph(rgba, b.tx, mx+secondaryLabelInsetX, y, label[1], white)
					}
					b.hits.add(mx, y-4, mw, rowHeight, i)
				}
			}
		}
		// 下方右:敵名框。混合 formation 逐種列名與存活數；同種群維持單列。
		{
			type namedCount struct{ monID, count int }
			var names []namedCount
			for i := range b.enemies {
				if !b.enemies[i].alive() {
					continue
				}
				found := false
				for j := range names {
					if names[j].monID == b.enemies[i].monID {
						names[j].count++
						found = true
						break
					}
				}
				if !found {
					names = append(names, namedCount{monID: b.enemies[i].monID, count: 1})
				}
			}
			if w := b.enemyLayout; w.ID != "" {
				ex, ey, ew, eh := w.X, w.Y, w.Width, w.Height
				if w.HeightMode == "rows_plus_base" {
					eh = (len(names) + w.BaseRows) * w.RowHeight
				}
				fillPackBox(rgba, b.enemyLayout.WindowLayout, ex, ey, ew, eh)
				for i, n := range names {
					nameY := ey + w.TextInsetY + i*w.RowHeight
					b.drawName(rgba, ex+w.NameInsetX, nameY, 0x258+n.monID, white)
					countY := nameY
					if w.CountInsetY != nil {
						countY += *w.CountInsetY
					}
					drawNumber(rgba, b.tx, ex+w.CountInsetX, countY, n.count, white)
				}
			}
		}
	}
}

// digitGlyphs 把整數轉十進位 glyph 序列(glyph 0..9 = 字元 '0'..'9',docs/03:64;val<0 視為 0)。
// drawNumber 與對話插值 VAR_NUM(dialogue.go varGlyphs)共用同一套轉換,別各自重寫。
func digitGlyphs(val int) []int {
	if val < 0 {
		val = 0
	}
	s := fmt.Sprintf("%d", val)
	out := make([]int, len(s))
	for i := 0; i < len(s); i++ {
		out[i] = int(s[i] - '0')
	}
	return out
}

// drawNumber 畫十進位數(digit glyph 0..9)。
func drawNumber(rgba []byte, tx *dq3data.Text, x, y, val int, fg dq3data.Color) {
	for i, gi := range digitGlyphs(val) {
		drawGlyph(rgba, tx, x+i*dq3data.GlyphPx, y, gi, fg)
	}
}

// drawBattleMessage consumes the original D3TXT glyph/control stream. Each
// VAR control consumes one runtime glyph slice from the message context; an
// unset variable reserves one blank cell just like the dialogue renderer.
func drawBattleMessage(rgba []byte, tx *dq3data.Text, x, y, cols int, m battleMessage, fg dq3data.Color) {
	lines := 4
	if m.text.columns > 0 {
		cols = m.text.columns
	}
	if m.text.lines > 0 {
		lines = m.text.lines
	}
	col, line := 0, 0
	varIdx := 0
	for i := 0; i < len(m.text.glyph); i++ {
		v := m.text.glyph[i]
		switch {
		case v == dq3data.TxtEnd:
			return
		case v == dq3data.TxtPage:
			return
		case v == dq3data.TxtNL || v == dq3data.TxtNL2:
			col, line = 0, line+1
			if line >= lines {
				return
			}
			continue
		case dq3data.IsVarInsert(v):
			var glyphs []int
			if varIdx < len(m.vars) {
				glyphs = m.vars[varIdx]
			}
			varIdx++
			if i+1 < len(m.text.glyph) {
				i++ // raw D3TXT control parameter word
			}
			n := len(glyphs)
			if n == 0 {
				n = 1
			}
			for j := 0; j < n; j++ {
				if j < len(glyphs) && tx != nil {
					drawGlyph(rgba, tx, x+col*16, y+line*16, glyphs[j], fg)
				}
				col++
				if col >= cols {
					col, line = 0, line+1
					if line >= lines {
						return
					}
				}
			}
			continue
		case v >= 0xffed:
			// Unknown control words remain a one-cell non-rendering token.
			col++
			if col >= cols {
				col, line = 0, line+1
				if line >= lines {
					return
				}
			}
			continue
		default:
			if v < dq3data.GlyphMax && tx != nil {
				drawGlyph(rgba, tx, x+col*16, y+line*16, int(v), fg)
			}
		}
		col++
		if col >= cols {
			col, line = 0, line+1
			if line >= lines {
				return
			}
		}
	}
}
