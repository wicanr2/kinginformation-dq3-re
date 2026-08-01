package game

import (
	"fmt"
	"sort"

	"github.com/wicanr2/dq3_remake_ebitan/internal/battle"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
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

// 戰鬥指令 glyph 標籤。
var battleCmdLabels = map[int][2]int{
	bcWar: {107, 207}, bcFlee: {629, 630}, bcDef: {203, 204},
	bcItem: {402, 1354}, bcSpell: {429, 430},
}

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

// enemyUnit 是群戰中一隻敵的即時狀態。怪物資料、AI 與 sprite 都逐隻保存，讓原版混合
// formation 不會退化成「代表怪物 × N」。
type enemyUnit struct {
	monID    int
	hp, max  int
	mp       int // D3MNS +0x04；帕魯朋特的吸 MP 效果會消耗
	atk, def int // 有效攻/防(索瑪+光之珠弱化後之值)
	agi      int // D3MNS +0x0b；敵我混排行動順序
	spr      *dq3data.MonsterSprite
	fled     bool // 本場已逃走(不計入擊殺數→經驗/金錢排除,對齊 C g_fled)
	status   int  // W3:異常狀態位元(statusParalysis/statusSealed/statusBlind,見下方 const)
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
	mons     *dq3data.Monsters
	shp      []byte
	mpal     []dq3data.Color // MNSBK.PAL(怪物色盤 + packbg 色盤)
	tx       *dq3data.Text
	nameText *dq3data.Text                            // D3TXT00 名表(咒文名 rec)
	scr      []byte                                   // PACKBG.SCR(戰鬥背景)
	bg       *[dq3data.PackBGH][dq3data.PackBGW]uint8 // 解碼後背景(草原 page22)

	active             bool
	monID              int                    // formation 第一筆怪物 id；單群相容與索瑪判斷用
	spr                *dq3data.MonsterSprite // formation 第一筆 sprite；舊測試/單群相容用
	enemies            []enemyUnit            // 敵群(N 隻,docs/72 A3;單敵 = 1 元素陣列)
	lightOrb           bool                   // 開戰時持光之珠(索瑪 0x7c 戰用;弱化後清)
	heroHP             int
	heroMax            int
	heroAtk            int
	heroDef            int
	heroAgi            int
	cursor             int
	msg                string
	messageQueue       []string // 已結算回合的後續逐筆訊息；msg 是目前顯示項
	phase              battlePhase
	result             int // 0 進行中、1 勝、2 敗、3 逃
	gotExp             int
	gotGold            int
	gotDrop            int  // 原版 D3MNS +0x26；-1 表示無掉落
	suppressDrop       bool // 原版 battle state bit1；劇情第二戰等明確禁止掉落
	rng                *dosrng.RNG
	flashCol           int // >0:受擊閃光殘餘幀
	defending          bool
	heroHerbs          int   // 開戰時持有藥草數
	usedHerbs          int   // 本戰用掉藥草數(戰後從背包扣)
	heroLevel          int   // 我方等級(敵 AI 逃跑門檻用)
	heroMP             int   // 目前 MP(施咒消耗)
	heroMaxMP          int   //
	spells             []int // 已學可施放咒文 rec
	spellCursor        int
	companions         []*battleActor // 同伴(狀態列顯示 + 可各自下令及被鎖定)
	heroStatus         int            // 主角異常狀態位元(statusParalysis 等;166-168 解咒清此欄)
	commandActor       int            // 目前下令者：0=隊長，1..=同伴
	commands           []battleCommand
	resolving          bool // 正在執行已收集命令；避免舊單步入口重複推進整回合
	pending            battleCommand
	targetCursor       int
	targetFrom         battlePhase // 取消目標選擇時回 phCommand 或 phSpell
	messageResume      battlePhase // 非回合訊息關閉後回到的選單；phEnd 表示正常回合訊息
	actionActor        int
	actionTarget       int
	statusSleepingText string // game pack common:text.battle.status.sleeping
	statusWokeText     string // game pack common:text.battle.status.woke

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
		b.emit(b.statusWokeText)
	} else {
		b.emit(b.statusSleepingText)
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
// HP 上限對齊 C dq3_battlescene_run:「單一擲值套全組」(同種同 HP 上限,非逐隻各擲),
// 故 N=1 時的亂數序列與舊版單敵 start() 完全等價(既有測試不受影響)。
func (b *Battle) startGroup(monID, count int, seed int64, hp heroParams, comps []*battleActor) bool {
	if count < 1 {
		count = 1
	}
	if count > MaxEnemies {
		count = MaxEnemies
	}
	return b.startFormation([]enemyGroup{{monID: monID, count: count}}, seed, hp, comps)
}

// startFormation 開一場原版混合編隊戰。每個 group 依序解碼，且每筆只擲一次 HP，
// 所以 startGroup 的 RNG 序列與既有同種群戰完全相容。
func (b *Battle) startFormation(groups []enemyGroup, seed int64, hp heroParams, comps []*battleActor) bool {
	if len(groups) == 0 {
		return false
	}
	b.rng = dosrng.New(uint16(seed))
	b.enemies = nil
	b.monID, b.spr = 0, nil
	for _, group := range groups {
		if len(b.enemies) >= MaxEnemies {
			break
		}
		count := group.count
		if count < 1 {
			continue
		}
		if count > MaxEnemies-len(b.enemies) {
			count = MaxEnemies - len(b.enemies)
		}
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
		hpRoll := int(st.HPBase) + b.rng.Next(int(st.HPRand)+1)
		eatk, edef := int(st.Atk), int(st.Def)
		if group.monID == 0x7c && b.lightOrb {
			eatk /= 3
			edef /= 4
			b.lightOrb = false
		}
		for i := 0; i < count; i++ {
			b.enemies = append(b.enemies, enemyUnit{
				monID: group.monID, hp: hpRoll, max: hpRoll, mp: int(st.MP),
				atk: eatk, def: edef, agi: int(st.Agi), spr: spr,
			})
		}
	}
	if len(b.enemies) == 0 {
		return false
	}
	if b.bg == nil && b.scr != nil { // 首戰解碼草原背景(page22,對 game3.png)
		b.bg, _ = dq3data.DecodePackBG(b.scr, 22)
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
	b.msg, b.gotExp, b.gotGold, b.gotDrop, b.suppressDrop = "", 0, 0, -1, false
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
					b.msg, b.phase = "還沒學會咒文", phMessage
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
				b.msg, b.phase, b.messageResume = "MP 不足", phMessage, phSpell
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
				b.msg = b.messageQueue[0]
				b.messageQueue = b.messageQueue[1:]
			} else if b.result != 0 {
				b.msg = ""
				b.phase = phEnd
			} else if b.messageResume != phEnd {
				b.msg = ""
				b.phase = b.messageResume
				b.messageResume = phEnd
			} else {
				b.msg = ""
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

func (b *Battle) emit(msg string) {
	if msg == "" {
		return
	}
	if b.resolving {
		b.messageQueue = append(b.messageQueue, msg)
		return
	}
	b.msg = msg
}

func (b *Battle) beginMessagePlayback() {
	b.phase = phMessage
	if len(b.messageQueue) == 0 {
		return
	}
	b.msg = b.messageQueue[0]
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
		b.emit("陷入異常,無法行動!")
		return
	}
	switch b.cursor {
	case bcFlee: // 逃げる
		if battle.FleeOK(b.heroAgi, b.fleeResist(), b.roll()) {
			b.result, b.phase = 3, phMessage
			b.emit("逃走成功")
			return
		}
		b.emit("逃走失敗")
	case bcDef: // 防御(本回合受傷減半)
		b.defending = true
		b.emit("防御")
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
			b.emit(fmt.Sprintf("回復 %d", heal))
		} else {
			b.emit("沒有藥草")
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
			b.emit("幻惑,揮空了")
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
		b.emit(fmt.Sprintf("給予 %d 傷害", dmg))
		b.flashCol = b.hurtFxFrames
	}
	if b.allEnemiesDead() {
		b.settleVictoryRewards()
		b.result, b.phase = 1, phMessage
		b.emit(fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold))
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
	if b.partySealed { // 瑪荷頓(敵施156)→ 我方無法施咒,但仍耗費本回合(對齊 C)
		b.emit("咒文被封印,無法施放!")
		return
	}
	caster := b.actionActor
	if b.actorMP(caster) < def.MP {
		b.phase = phMessage
		b.emit("MP 不足")
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
		total := 0
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
			total += val
		}
		b.emit(fmt.Sprintf("回復 %d", total))
	case spell.Sleep:
		n := b.applyEnemyStatus(def.Target, b.actionTarget, statusParalysis)
		b.emit(fmt.Sprintf("%d 隻敵人陷入睡眠!", n))
	case spell.BuffAtk:
		for _, target := range b.allySpellTargets(def.Target, b.actionTarget) {
			b.addActorAtkPct(target, 100)
		}
		b.emit("同伴的攻擊力提升了!")
	case spell.BuffDef:
		for _, target := range b.allySpellTargets(def.Target, b.actionTarget) {
			b.addActorDefPct(target, 50)
		}
		b.emit("同伴的守備力提升了!")
	case spell.Seal:
		n := b.applyEnemyStatus(def.Target, b.actionTarget, statusSealed)
		b.emit(fmt.Sprintf("%d 隻敵人的咒文被封印!", n))
	case spell.Blind:
		n := b.applyEnemyStatus(def.Target, b.actionTarget, statusBlind)
		b.emit(fmt.Sprintf("%d 隻敵人陷入幻惑!", n))
	case spell.CurePoison: // 基阿里166
		for _, target := range b.allySpellTargets(def.Target, b.actionTarget) {
			b.clearActorStatus(target, statusPoison)
		}
		b.emit("解毒了!")
	case spell.CureStatus: // 基阿里克167/薩梅哈168(remake 簡化:麻痺/混亂/睡眠共用一位元)
		for _, target := range b.allySpellTargets(def.Target, b.actionTarget) {
			b.clearActorStatus(target, statusParalysis)
		}
		b.emit("異常解除了!")
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
			}
			b.flashCol = b.hurtFxFrames
			b.emit(fmt.Sprintf("咒文波及 %d 隻敵人", hit))
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
				b.emit(fmt.Sprintf("咒文 %d 傷害", val))
			}
			b.flashCol = b.hurtFxFrames
		}
	}
	if b.allEnemiesDead() {
		b.settleVictoryRewards()
		b.result, b.phase = 1, phMessage
		b.emit(fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold))
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
			}
		}
		if b.heroHP > 0 && b.heroStatus&statusParalysis == 0 && b.rng.Next(256) < 0x96 {
			b.heroStatus |= statusParalysis
			affected++
		}
		for _, c := range b.companions {
			if c.hp > 0 && c.status&statusParalysis == 0 && b.rng.Next(256) < 0x96 {
				c.status |= statusParalysis
				affected++
			}
		}
		if affected == 0 {
			b.emit("但是這個咒文沒有效。")
		} else {
			b.emit("敵我之間有睡意‥‥")
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
		b.heroHP = heal(b.heroHP, b.heroMax)
		for _, c := range b.companions {
			c.hp = heal(c.hp, c.maxHP)
		}
		b.emit("全體同伴的傷口復原了。")
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
			}
		}
		if affected == 0 {
			b.emit("但是這個咒文沒有效。")
		} else {
			b.emit("怪物被吹走了。")
		}
	case 12: // effect type1：隨機一敵，吸取 1..min(10,MP)-1（原版 0 會改成 1）
		alive := make([]int, 0, len(b.enemies))
		for i := range b.enemies {
			if b.enemies[i].alive() {
				alive = append(alive, i)
			}
		}
		if len(alive) == 0 {
			b.emit("但是這個咒文沒有效。")
			return
		}
		tgt := alive[b.rng.Next(len(alive))]
		chance, _ := b.mons.SpellChance(b.enemies[tgt].monID, 59)
		// effect type1 caller file 0xe69a uses JA failure，與其他效果的 JB
		// success 不同；因此 roll==chance 仍成功。
		if b.rng.Next(256) > chance || b.enemies[tgt].mp <= 0 {
			b.emit("但是這個咒文沒有效。")
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
		b.emit(fmt.Sprintf("奪取了 %d 點的MP", drain))
	default:
		b.emit("但是這個咒文沒有效。")
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
	b.msg = ""
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
	b.emit(fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold))
}

func (b *Battle) execCompanionCommand(i int, cmd battleCommand) {
	c := b.companions[i]
	switch cmd.kind {
	case bcFlee:
		if battle.FleeOK(c.agi, b.fleeResist(), b.roll()) {
			b.result, b.phase = 3, phMessage
			b.emit("逃走成功")
		} else {
			b.emit("逃走失敗")
		}
	case bcDef:
		c.defending = true
		b.emit("防御")
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
			b.emit(fmt.Sprintf("回復 %d", heal))
		} else {
			b.emit("沒有藥草")
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
			b.emit("幻惑,揮空了")
			return
		}
		atk := pct(c.atk, b.actorAtkPct(i+1))
		def := pct(b.enemies[tgt].def, b.enemyDefPct)
		dmg := battle.PhysDamage(atk, def, b.roll(), 0)
		b.enemies[tgt].hp -= dmg
		if b.enemies[tgt].hp < 0 {
			b.enemies[tgt].hp = 0
		}
		b.emit(fmt.Sprintf("給予 %d 傷害", dmg))
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
			b.emit(b.statusWokeText)
		} else {
			b.emit(b.statusSleepingText)
		}
		return
	}
	if aiOK && ai.FleeRate > 0 && b.heroLevel >= int(ai.FleeThresh) && b.roll() <= int(ai.FleeRate) {
		b.enemies[i].hp, b.enemies[i].fled = 0, true
		b.emit("敵人逃走了")
		return
	}
	targets := b.aliveTargets()
	if len(targets) == 0 {
		b.result, b.phase = 2, phMessage
		b.emit("全滅‥")
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
					b.emit(fmt.Sprintf("敵使用咒文回復 %d", val))
					return
				case spell.Sleep:
					b.setTargetStatus(tgt, statusParalysis)
					b.emit("敵使用咒文,我方陷入異常!")
					return
				case spell.BuffAtk:
					b.enemyAtkPct += 100
					if b.enemyAtkPct > 400 {
						b.enemyAtkPct = 400
					}
					b.emit("敵方攻擊力上升了!")
					return
				case spell.BuffDef:
					b.enemyDefPct += 50
					if b.enemyDefPct > 300 {
						b.enemyDefPct = 300
					}
					b.emit("敵方守備力上升了!")
					return
				case spell.Seal:
					b.partySealed = true
					b.emit("我方咒文被封印了!")
					return
				case spell.Blind:
					b.partyBlind = true
					b.emit("我方陷入幻惑!")
					return
				default:
					val := spell.CastValue(def.Base, b.roll())
					b.damageMember(tgt, val)
					b.emit(fmt.Sprintf("敵咒文 %d 傷害", val))
					b.wipedOut()
					return
				}
			}
		}
	}
	if b.enemies[i].status&statusBlind != 0 && b.roll() < 128 {
		b.emit("敵人陷入幻惑,攻擊落空")
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
	b.emit(fmt.Sprintf("敵攻擊 %d 傷害", edmg))
	b.wipedOut()
}

// wipedOut:我方全滅 → 設敗、回 true(呼叫端應立即結束回合,別讓其餘敵繼續行動)。
func (b *Battle) wipedOut() bool {
	if len(b.aliveTargets()) == 0 {
		b.result, b.phase = 2, phMessage
		b.emit("全滅‥")
		return true
	}
	return false
}

// 版面常數(1:1 對齊 C dq3_battlescene render,references/game3.png)。
const (
	fieldY0, fieldY1, groundY = 80, 246, 232 // 場景帶(天空+綠地)/ 綠地平線
	curGlyph                  = 0x77         // ►/★ 游標 glyph(對齊 C)
)

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

	// 背景:全黑 → 場景帶。有 packbg(草原 page22)→ 縮放進 80..246;否則純色 fallback。
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			putPx(rgba, x, y, black)
		}
	}
	if b.bg != nil { // packbg:88 row 縮放進 fieldY0..fieldY1(對齊 C g_sky 渲染)
		for y := fieldY0; y < fieldY1; y++ {
			sr := (y - fieldY0) * dq3data.PackBGH / (fieldY1 - fieldY0)
			if sr >= dq3data.PackBGH {
				sr = dq3data.PackBGH - 1
			}
			for x := 0; x < ScreenW && x < dq3data.PackBGW; x++ {
				c := black
				if idx := int(b.bg[sr][x]); idx < len(b.mpal) {
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
	// 怪群:橫排站綠地平線(對齊 C render:gx=(SCREEN_W*(i+1))/(en+1)-spr->w/2,en=組總隻數
	// 含已死/逃走者,位置不因中途減員而重排);受擊閃紅;死/逃不畫。
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
			gx := ScreenW*(i+1)/(en+1) - spr.W/2
			gy := groundY + 8 - spr.H
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
		fillBox(rgba, tx0-8, ty0-6, cols*colpx+8, 4*16+12, white)
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
	// 下方左:指令窗(120,236,150×108)—— 指令 或 咒文子選單
	{
		mx, my, mw, mh := 120, 236, 150, 108
		fillBox(rgba, mx, my, mw, mh, white)
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
				y := my + 24 + row*16
				if i == b.spellCursor {
					drawGlyph(rgba, b.tx, mx+8, y, curGlyph, white)
				}
				b.drawName(rgba, mx+28, y, rec, white)
				if def, ok := spell.GetDef(rec); ok {
					drawNumber(rgba, b.tx, mx+110, y, def.MP, cyan)
				}
				b.hits.add(mx, y-4, mw, 16, i)
			}
		} else if b.phase == phTargetEnemy {
			targets := b.aliveEnemyIndices()
			if len(targets) > 0 {
				target := targets[b.targetCursor]
				y := my + 24
				drawGlyph(rgba, b.tx, mx+8, y, curGlyph, white)
				b.drawName(rgba, mx+28, y, 0x258+b.enemies[target].monID, white)
				drawNumber(rgba, b.tx, mx+112, y, target+1, cyan)
			}
		} else if b.phase == phTargetAlly {
			targets := b.aliveActorIndices()
			if len(targets) > 0 {
				actor := targets[b.targetCursor]
				y := my + 24
				drawGlyph(rgba, b.tx, mx+8, y, curGlyph, white)
				name := classNames[0]
				if actor > 0 {
					name = classNames[b.companions[actor-1].class]
				}
				for i, glyph := range name {
					drawGlyph(rgba, b.tx, mx+28+i*16, y, glyph, white)
				}
			}
		} else if b.phase == phMessage {
			drawBattleMessage(rgba, b.tx, mx+12, my+16, 8, b.msg, white)
		} else if b.phase == phCommand {
			menu := b.commandMenu(b.commandActor)
			for i, kind := range menu {
				y := my + 24 + i*16
				if i == b.cursor {
					drawGlyph(rgba, b.tx, mx+8, y, curGlyph, white)
				}
				label := battleCmdLabels[kind]
				drawGlyph(rgba, b.tx, mx+28, y, label[0], white)
				drawGlyph(rgba, b.tx, mx+72, y, label[1], white)
				b.hits.add(mx, y-4, mw, 16, i)
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
		ex, ey, ew, eh := 290, 262, 250, 20+16*len(names)
		if eh < 36 {
			eh = 36
		}
		fillBox(rgba, ex, ey, ew, eh, white)
		for i, n := range names {
			y := ey + 8 + i*16
			b.drawName(rgba, ex+12, y, 0x258+n.monID, white)
			drawNumber(rgba, b.tx, ex+ew-40, y, n.count, white)
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

// battleMessageGlyph 是本地 battle fallback 訊息會用到的 D3TXT00.FON 字。
// 原版有對應的 rec330..376；待訊息變數 context 完整接入後可直接改以 record renderer。
var battleMessageGlyph = map[rune]int{
	' ': 13, ',': 55, '.': 43, '!': 57, ':': 61, '‥': 123, '。': 56,
	'M': 27, 'E': 19, 'X': 38, 'P': 30, 'G': 21,
	'上': 421, '不': 228, '中': 407, '之': 145, '了': 423, '予': 677,
	'人': 194, '伴': 602, '但': 508, '個': 546, '倒': 627, '備': 409,
	'傷': 553, '入': 519, '全': 594, '力': 170, '功': 195, '動': 619,
	'升': 422, '印': 501, '原': 555, '及': 1026, '取': 614, '口': 554,
	'同': 601, '吹': 180, '咒': 429, '回': 312, '失': 515, '奪': 613,
	'守': 340, '害': 471, '封': 500, '常': 581, '幻': 295, '御': 204,
	'復': 463, '怪': 469, '惑': 168, '意': 1076, '成': 610, '我': 493,
	'打': 487, '提': 1214, '揮': 652, '擊': 624, '攻': 623, '放': 634,
	'效': 418, '敗': 853, '敵': 478, '文': 430, '方': 547, '施': 1095,
	'是': 399, '有': 147, '毒': 137, '沒': 410, '法': 208, '波': 659,
	'滅': 912, '無': 495, '物': 470, '異': 643, '的': 149, '眠': 798,
	'睡': 628, '空': 328, '給': 559, '草': 174, '落': 1053, '著': 484,
	'藥': 244, '行': 531, '被': 437, '解': 464, '走': 612, '足': 1377,
	'逃': 629, '這': 398, '間': 1174, '防': 203, '除': 790, '陷': 1158,
	'隻': 680, '體': 486, '點': 425, '用': 240,
}

func drawBattleMessage(rgba []byte, tx *dq3data.Text, x, y, cols int, s string, fg dq3data.Color) {
	col, line := 0, 0
	for _, r := range s {
		glyph, ok := battleMessageGlyph[r]
		if r >= '0' && r <= '9' {
			glyph, ok = int(r-'0'), true
		}
		if ok && r != ' ' {
			drawGlyph(rgba, tx, x+col*16, y+line*16, glyph, fg)
		}
		col++
		if col >= cols {
			col, line = 0, line+1
			if line >= 4 {
				return
			}
		}
	}
}
