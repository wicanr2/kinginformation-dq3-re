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

// enemyUnit 是群戰中一隻敵的即時狀態。W2(docs/72 A3):同種群戰,故 monID/atk/def 目前組內
// 皆同值(對齊 C ehp[MAXE] 陣列 + 共用 eatk/edef/spr);逐隻獨立只留 hp/fled,monID/atk/def
// 保留逐隻欄位是為將來混群(W3+)鋪路,非本批要做。
type enemyUnit struct {
	monID    int
	hp, max  int
	mp       int  // D3MNS +0x04；帕魯朋特的吸 MP 效果會消耗
	atk, def int  // 有效攻/防(索瑪+光之珠弱化後之值)
	agi      int  // D3MNS +0x0b；敵我混排行動順序
	fled     bool // 本場已逃走(不計入擊殺數→經驗/金錢排除,對齊 C g_fled)
	status   int  // W3:異常狀態位元(statusParalysis/statusSealed/statusBlind,見下方 const)
}

func (e *enemyUnit) alive() bool { return e.hp > 0 && !e.fled }

// 異常狀態位元(docs/data/spell-effects-research.md;W3)。共用於 enemyUnit.status、
// battleActor.status、Battle.heroStatus。對齊 C DQ3_STATUS_*:睡眠(拉里荷144)與混亂(美達巴尼152)
// remake 映射同一位元(統稱 statusParalysis,即 C DQ3_STATUS_PARALYSIS)——本回合起不能行動,
// 且對齊 C 行為(dq3_battlescene.c 全檔搜尋無中途自動清除):持續整場戰鬥,需靠解咒/道具清除
// (166/167/168)或戰鬥結束(reset)。statusSealed/statusBlind 為 W3 新增(C 版對應的是「敵方
// 全域」g_party_sealed/g_party_blind,W3 額外做「玩家對單一敵單位」的鏡像,故落在 per-unit 欄位)。
const (
	statusPoison    = 1 << iota // 中毒(既有位元,占位保留;W3 未接毒傷結算)
	statusParalysis             // 睡眠/混亂(拉里荷144、美達巴尼152;敵方鏡像同)
	statusSealed                // 瑪荷頓156:不能施咒
	statusBlind                 // 瑪努莎158:物攻 ~50% 失手
)

type battlePhase int

const (
	phCommand battlePhase = iota // 等玩家下指令
	phSpell                      // 選咒文
	phMessage                    // 顯示訊息(A 推進)
	phEnd                        // 勝/敗/逃(A 關閉戰鬥)
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

	active       bool
	monID        int                    // 群組代表種 id(同種群戰,全組共用;name 查詢/0x7c 判斷用此)
	spr          *dq3data.MonsterSprite // 群組共用 sprite(同種群戰,W2 範圍)
	enemies      []enemyUnit            // 敵群(N 隻,docs/72 A3;單敵 = 1 元素陣列)
	lightOrb     bool                   // 開戰時持光之珠(索瑪 0x7c 戰用;弱化後清)
	heroHP       int
	heroMax      int
	heroAtk      int
	heroDef      int
	heroAgi      int
	cursor       int
	msg          string
	phase        battlePhase
	result       int // 0 進行中、1 勝、2 敗、3 逃
	gotExp       int
	gotGold      int
	rng          *dosrng.RNG
	flashCol     int // >0:受擊閃光殘餘幀
	defending    bool
	heroHerbs    int   // 開戰時持有藥草數
	usedHerbs    int   // 本戰用掉藥草數(戰後從背包扣)
	heroLevel    int   // 我方等級(敵 AI 逃跑門檻用)
	heroMP       int   // 目前 MP(施咒消耗)
	heroMaxMP    int   //
	spells       []int // 已學可施放咒文 rec
	spellCursor  int
	companions   []*battleActor // 同伴(狀態列顯示 + 可各自下令及被鎖定)
	heroStatus   int            // 主角異常狀態位元(statusParalysis 等;166-168 解咒清此欄)
	commandActor int            // 目前下令者：0=隊長，1..=同伴
	commands     []battleCommand
	resolving    bool // 正在執行已收集命令；避免舊單步入口重複推進整回合

	// per-battle 修正狀態(W3,docs/data/spell-effects-research.md;每場 startGroup 歸零,
	// 對齊 C reset_battle_mods()/g_eatk_pct 等——百分比整數,基準 100)。玩家自施 151/154/155
	// 影響 party*、敵自施影響 enemy*;partyBlind/partySealed 對齊 C g_party_blind/g_party_sealed,
	// 由敵施 158/156 設下,C 版全檔無中途清除 → 持續整場戰鬥(非單回合)。
	partyAtkPct int  // 我方物攻%(拜基魯多151 我方自施,上限400)
	partyDefPct int  // 我方守備%(154/155 我方自施升 / 敵鏡像146/147降,25..300;W3 只做前者,後者留 TODO)
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
	name          []int
	class, level  int
	hp, maxHP     int
	mp, maxMP     int
	atk, def, agi int
	spells        []int
	defending     bool
	status        int // W3:異常狀態位元(statusParalysis 等,見 enemyUnit 旁 const)
}

type battleCommand struct {
	kind  int
	spell int
}

type turnEntry struct {
	enemy bool
	index int
	speed int
}

// pct:val 套用百分比修正(基準 100)。共用於物攻/守備 buff/debuff(拜基魯多/史卡拉等)。
func pct(val, p int) int { return val * p / 100 }

func (b *Battle) roll() int { return b.rng.Next(256) }

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
	spr, err := dq3data.DecodeMonsterSprite(b.shp, monID)
	if err != nil {
		return false
	}
	st, ok := b.mons.Stat(monID)
	if !ok {
		return false
	}
	if count < 1 {
		count = 1
	}
	if count > MaxEnemies {
		count = MaxEnemies
	}
	b.rng = dosrng.New(uint16(seed))
	b.monID, b.spr = monID, spr
	hpRoll := int(st.HPBase) + b.rng.Next(int(st.HPRand)+1)
	// 敵有效攻防;索瑪(0x7c)+光之珠 → 驅散黑暗結界:攻/3、防/4(移植 battlescene 二階段)。
	eatk, edef := int(st.Atk), int(st.Def)
	if monID == 0x7c && b.lightOrb {
		eatk /= 3
		edef /= 4
		b.lightOrb = false // 用畢清旗標
	}
	b.enemies = make([]enemyUnit, count)
	for i := range b.enemies {
		b.enemies[i] = enemyUnit{
			monID: monID, hp: hpRoll, max: hpRoll, mp: int(st.MP),
			atk: eatk, def: edef, agi: int(st.Agi),
		}
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
	b.commands = make([]battleCommand, 1+len(comps))
	b.resolving = false
	// W3 per-battle 修正狀態歸零(對齊 C reset_battle_mods();百分比基準 100)。
	b.partyAtkPct, b.partyDefPct, b.enemyAtkPct, b.enemyDefPct = 100, 100, 100, 100
	b.partyBlind, b.partySealed = false, false
	b.cursor, b.phase, b.result = 0, phCommand, 0
	b.msg, b.gotExp, b.gotGold = "", 0, 0
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

func (b *Battle) resetCommandRound() {
	b.commands = make([]battleCommand, 1+len(b.companions))
	b.commandActor, b.cursor, b.spellCursor = 0, 0, 0
	b.seekCommandActor(0)
}

// firstAliveEnemy:第一個存活敵的 index(remake 簡化目標選擇:玩家物攻/單體咒固定指向組內
// 第一個存活敵,非原版 C pick_alive_enemy 的隨機選;敵方 AI 選我方目標仍維持隨機,見 enemyTurn)。
// 無存活敵回 -1。
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

// killedCount:組內「被擊殺」隻數(排除逃走者;對齊 C (en-g_fled) 算經驗/金錢)。
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
				b.queueCommand(battleCommand{kind: kind})
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
				b.msg, b.phase = "MP 不足", phMessage
			} else {
				b.queueCommand(battleCommand{kind: bcSpell, spell: rec})
			}
		case in.DirEdge == 0:
			b.spellCursor = (b.spellCursor + 1) % len(spells)
		case in.DirEdge == 1:
			b.spellCursor = (b.spellCursor + len(spells) - 1) % len(spells)
		}
	case phMessage:
		if in.Confirm {
			b.phase = phCommand
			b.msg = ""
			if b.result != 0 {
				b.phase = phEnd
			} else {
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

// execTurn 執行玩家指令 + 敵方回合,結算。
func (b *Battle) execTurn() {
	if !b.resolving {
		b.commands = make([]battleCommand, 1+len(b.companions))
		b.commands[0] = battleCommand{kind: b.cursor}
		for i := range b.companions {
			b.commands[i+1] = battleCommand{kind: bcWar}
		}
		b.resolveRound()
		return
	}
	st, _ := b.mons.Stat(b.monID)
	b.defending = false
	if b.heroStatus&statusParalysis != 0 { // 睡眠/混亂(敵施144/152):本回合任何指令皆無法行動
		b.msg = "陷入異常狀態,無法行動!"
		return
	}
	switch b.cursor {
	case bcFlee: // 逃げる
		if battle.FleeOK(b.heroAgi, int(st.FleeResist), b.roll()) {
			b.msg, b.result, b.phase = "逃走成功", 3, phMessage
			return
		}
		b.msg = "逃走失敗"
	case bcDef: // 防御(本回合受傷減半)
		b.defending = true
		b.msg = "防御"
	case bcItem: // 道具:用藥草回 HP
		if b.usedHerbs < b.heroHerbs && b.heroHP < b.heroMax {
			heal := herbHeal
			if b.heroHP+heal > b.heroMax {
				heal = b.heroMax - b.heroHP
			}
			b.heroHP += heal
			b.usedHerbs++
			b.msg = fmt.Sprintf("回復 %d", heal)
		} else {
			b.msg = "沒有藥草"
		}
	default: // 戰う:主角攻擊(目標=第一個存活敵,見 firstAliveEnemy 簡化說明)
		tgt := b.firstAliveEnemy()
		if tgt < 0 {
			break // 組內已無存活敵(全數逃走中)→ 揮空,直接進勝負判定
		}
		if b.partyBlind && b.roll() < 128 { // 我方幻惑(敵施158瑪努莎)→ ~50% 揮空(對齊 C g_party_blind)
			b.msg = "揮空了(幻惑)"
			break
		}
		crit := 0
		if b.roll() < 8 { // ~1/32 會心
			crit = 1
		}
		atk := pct(b.heroAtk, b.partyAtkPct)          // 拜基魯多151(我方自施)
		def := pct(b.enemies[tgt].def, b.enemyDefPct) // 敵154/155自施升守備
		dmg := battle.PhysDamage(atk, def, b.roll(), crit)
		b.enemies[tgt].hp -= dmg
		if b.enemies[tgt].hp < 0 {
			b.enemies[tgt].hp = 0
		}
		b.msg = fmt.Sprintf("給予 %d 傷害", dmg)
		b.flashCol = b.hurtFxFrames
	}
	if b.allEnemiesDead() { // 勝:擊殺數(排除逃走)× 單隻經驗/金錢(對齊 C apply_victory_exp)
		killed := b.killedCount()
		b.gotExp, b.gotGold = killed*int(st.Exp), killed*int(st.Gold)
		b.msg, b.result, b.phase = fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold), 1, phMessage
		return
	}
}

// execSpell 施放咒文 rec:DMG 傷敵 / HEAL 補己 / W3 新增輔助狀態咒(睡眠/buff/封咒/幻惑/解咒)。
// 瑪荷頓 sealed(敵施封我方咒)在此擋下；被封仍算用掉這次行動。
func (b *Battle) execSpell(rec int) {
	if !b.resolving {
		b.commands = make([]battleCommand, 1+len(b.companions))
		b.commands[0] = battleCommand{kind: bcSpell, spell: rec}
		for i := range b.companions {
			b.commands[i+1] = battleCommand{kind: bcWar}
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
		b.msg = "咒文被封印,無法施放!"
		return
	}
	if b.heroMP < def.MP {
		b.msg, b.phase = "MP 不足", phMessage
		return
	}
	b.heroMP -= def.MP
	b.defending = false
	switch def.Kind {
	case spell.Heal:
		val := spell.CastValue(def.Base, b.roll())
		if b.heroHP+val > b.heroMax {
			val = b.heroMax - b.heroHP
		}
		b.heroHP += val
		b.msg = fmt.Sprintf("回復 %d", val)
	case spell.Sleep: // 拉里荷144/美達巴尼152:敵單體(第一個存活敵)陷入睡眠/混亂
		if tgt := b.firstAliveEnemy(); tgt >= 0 {
			b.enemies[tgt].status |= statusParalysis
			b.msg = "敵人陷入睡眠!"
		}
	case spell.BuffAtk: // 拜基魯多151:我方物攻 ×2(上限400%,對齊 C g_eatk_pct 鏡像)
		b.partyAtkPct += 100
		if b.partyAtkPct > 400 {
			b.partyAtkPct = 400
		}
		b.msg = "我方攻擊力提升了!"
	case spell.BuffDef: // 史卡拉154/史克魯多155:我方守備 +50%(上限300%,對齊 C g_edef_pct 鏡像)
		b.partyDefPct += 50
		if b.partyDefPct > 300 {
			b.partyDefPct = 300
		}
		b.msg = "我方守備力提升了!"
	case spell.Seal: // 瑪荷頓156:敵單體被封咒
		if tgt := b.firstAliveEnemy(); tgt >= 0 {
			b.enemies[tgt].status |= statusSealed
			b.msg = "敵人的咒文被封印!"
		}
	case spell.Blind: // 瑪努莎158:敵單體陷入幻惑(物攻易失手)
		if tgt := b.firstAliveEnemy(); tgt >= 0 {
			b.enemies[tgt].status |= statusBlind
			b.msg = "敵人陷入幻惑!"
		}
	case spell.CurePoison: // 基阿里166
		b.heroStatus &^= statusPoison
		b.msg = "解毒了!"
	case spell.CureStatus: // 基阿里克167/薩梅哈168(remake 簡化:麻痺/混亂/睡眠共用一位元)
		b.heroStatus &^= statusParalysis
		b.msg = "狀態解除了!"
	case spell.Palpunte:
		b.execPalpunte(b.rng.Next(16))
	default: // Dmg
		if def.Target == spell.TargetGroup { // 群體咒:命中每隻存活敵,各自獨立擲值(對齊 C cast_spell_effect)
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
			b.msg = fmt.Sprintf("咒文波及 %d 隻敵人", hit)
		} else { // 單體咒:目標=第一個存活敵(簡化,見 firstAliveEnemy 說明)
			tgt := b.firstAliveEnemy()
			if tgt >= 0 {
				val := spell.CastValue(def.Base, b.roll())
				b.enemies[tgt].hp -= val
				if b.enemies[tgt].hp < 0 {
					b.enemies[tgt].hp = 0
				}
				b.msg = fmt.Sprintf("咒文 %d 傷害", val)
			}
			b.flashCol = b.hurtFxFrames
		}
	}
	if b.allEnemiesDead() {
		st, _ := b.mons.Stat(b.monID)
		killed := b.killedCount()
		b.gotExp, b.gotGold = killed*int(st.Exp), killed*int(st.Gold)
		b.msg, b.result, b.phase = fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold), 1, phMessage
		return
	}
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
			b.msg = "但是這個咒文沒有效。"
		} else {
			b.msg = "敵我之間瀰漫著睡意‥‥"
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
		b.msg = "全體同伴的傷口復原了。"
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
			b.msg = "但是這個咒文沒有效。"
		} else {
			b.msg = "怪物被吹走了。"
		}
	case 12: // effect type1：隨機一敵，吸取 1..min(10,MP)-1（原版 0 會改成 1）
		alive := make([]int, 0, len(b.enemies))
		for i := range b.enemies {
			if b.enemies[i].alive() {
				alive = append(alive, i)
			}
		}
		if len(alive) == 0 {
			b.msg = "但是這個咒文沒有效。"
			return
		}
		tgt := alive[b.rng.Next(len(alive))]
		chance, _ := b.mons.SpellChance(b.enemies[tgt].monID, 59)
		// effect type1 caller file 0xe69a uses JA failure，與其他效果的 JB
		// success 不同；因此 roll==chance 仍成功。
		if b.rng.Next(256) > chance || b.enemies[tgt].mp <= 0 {
			b.msg = "但是這個咒文沒有效。"
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
		b.heroMP += drain // 原版直接 add [caster+0x18]，不以 max MP 封頂
		b.msg = fmt.Sprintf("奪取了 %d 點的MP", drain)
	default:
		b.msg = "但是這個咒文沒有效。"
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
	ai, aiOK := b.mons.AI(b.monID)
	for _, e := range order {
		if b.result != 0 {
			return
		}
		if e.enemy {
			b.enemyAction(e.index, ai, aiOK)
			continue
		}
		if !b.actorAlive(e.index) || b.actorStatus(e.index)&statusParalysis != 0 {
			continue
		}
		cmd := b.commands[e.index]
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
		} else if !b.wipedOut() {
			b.phase = phMessage
		}
	}
}

func (b *Battle) finishVictory() {
	st, _ := b.mons.Stat(b.monID)
	killed := b.killedCount()
	b.gotExp, b.gotGold = killed*int(st.Exp), killed*int(st.Gold)
	b.msg, b.result, b.phase = fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold), 1, phMessage
}

func (b *Battle) execCompanionCommand(i int, cmd battleCommand) {
	c := b.companions[i]
	switch cmd.kind {
	case bcFlee:
		st, _ := b.mons.Stat(b.monID)
		if battle.FleeOK(c.agi, int(st.FleeResist), b.roll()) {
			b.msg, b.result, b.phase = "逃走成功", 3, phMessage
		} else {
			b.msg = "逃走失敗"
		}
	case bcDef:
		c.defending = true
		b.msg = "防御"
	case bcItem:
		if b.usedHerbs < b.heroHerbs && c.hp < c.maxHP {
			heal := herbHeal
			if c.hp+heal > c.maxHP {
				heal = c.maxHP - c.hp
			}
			c.hp += heal
			b.usedHerbs++
			b.msg = fmt.Sprintf("回復 %d", heal)
		} else {
			b.msg = "沒有藥草"
		}
	case bcSpell:
		b.execCompanionSpell(i, cmd.spell)
	default:
		tgt := b.firstAliveEnemy()
		if tgt < 0 {
			return
		}
		if b.partyBlind && b.roll() < 128 {
			b.msg = "揮空了(幻惑)"
			return
		}
		atk := pct(c.atk, b.partyAtkPct)
		def := pct(b.enemies[tgt].def, b.enemyDefPct)
		dmg := battle.PhysDamage(atk, def, b.roll(), 0)
		b.enemies[tgt].hp -= dmg
		if b.enemies[tgt].hp < 0 {
			b.enemies[tgt].hp = 0
		}
		b.msg = fmt.Sprintf("給予 %d 傷害", dmg)
		b.flashCol = b.hurtFxFrames
	}
	if b.allEnemiesDead() {
		b.finishVictory()
	}
}

// execCompanionSpell 暫時把施法同伴放進既有 caster 欄位，並把真正隊長放到該同伴槽位。
// 如此 Palpunte 的全隊逐人效果仍只處理每人一次，結束後再原位寫回。
func (b *Battle) execCompanionSpell(i, rec int) {
	c := b.companions[i]
	savedHP, savedMax, savedMP, savedMaxMP, savedStatus := b.heroHP, b.heroMax, b.heroMP, b.heroMaxMP, b.heroStatus
	b.heroHP, b.heroMax, b.heroMP, b.heroMaxMP, b.heroStatus = c.hp, c.maxHP, c.mp, c.maxMP, c.status
	c.hp, c.maxHP, c.mp, c.maxMP, c.status = savedHP, savedMax, savedMP, savedMaxMP, savedStatus
	b.execSpell(rec)
	casterHP, casterMax, casterMP, casterMaxMP, casterStatus := b.heroHP, b.heroMax, b.heroMP, b.heroMaxMP, b.heroStatus
	b.heroHP, b.heroMax, b.heroMP, b.heroMaxMP, b.heroStatus = c.hp, c.maxHP, c.mp, c.maxMP, c.status
	c.hp, c.maxHP, c.mp, c.maxMP, c.status = casterHP, casterMax, casterMP, casterMaxMP, casterStatus
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
// AI(咒文機率/逃跑閾值)為同種共用欄位(對齊 C g_eai 單載),故迴圈外只查一次。
func (b *Battle) enemyTurn() {
	ai, aiOK := b.mons.AI(b.monID)
	for i := range b.enemies {
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
		b.msg = "敵人昏睡中,無法行動"
		return
	}
	if aiOK && ai.FleeRate > 0 && b.heroLevel >= int(ai.FleeThresh) && b.roll() <= int(ai.FleeRate) {
		b.enemies[i].hp, b.enemies[i].fled = 0, true
		b.msg = "敵人逃走了"
		return
	}
	targets := b.aliveTargets()
	if len(targets) == 0 {
		b.msg, b.result, b.phase = "全滅…", 2, phMessage
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
					b.msg = fmt.Sprintf("敵詠唱咒文回復 %d", val)
					return
				case spell.Sleep:
					b.setTargetStatus(tgt, statusParalysis)
					b.msg = "敵詠唱咒文,我方陷入異常狀態!"
					return
				case spell.BuffAtk:
					b.enemyAtkPct += 100
					if b.enemyAtkPct > 400 {
						b.enemyAtkPct = 400
					}
					b.msg = "敵方攻擊力上升了!"
					return
				case spell.BuffDef:
					b.enemyDefPct += 50
					if b.enemyDefPct > 300 {
						b.enemyDefPct = 300
					}
					b.msg = "敵方守備力上升了!"
					return
				case spell.Seal:
					b.partySealed = true
					b.msg = "我方咒文被封印了!"
					return
				case spell.Blind:
					b.partyBlind = true
					b.msg = "我方陷入幻惑!"
					return
				default:
					val := spell.CastValue(def.Base, b.roll())
					b.damageMember(tgt, val)
					b.msg = fmt.Sprintf("敵咒文 %d 傷害", val)
					b.wipedOut()
					return
				}
			}
		}
	}
	if b.enemies[i].status&statusBlind != 0 && b.roll() < 128 {
		b.msg = "敵人陷入幻惑,攻擊落空"
		return
	}
	eatk := pct(b.enemies[i].atk, b.enemyAtkPct)
	edef := pct(b.targetDef(tgt), b.partyDefPct)
	edmg := battle.PhysDamage(eatk, edef, b.roll(), 0)
	if tgt < 0 && b.defending || tgt >= 0 && b.companions[tgt].defending {
		edmg /= 2
	}
	b.damageMember(tgt, edmg)
	b.msg = fmt.Sprintf("敵攻擊 %d 傷害", edmg)
	b.wipedOut()
}

// wipedOut:我方全滅 → 設敗、回 true(呼叫端應立即結束回合,別讓其餘敵繼續行動)。
func (b *Battle) wipedOut() bool {
	if len(b.aliveTargets()) == 0 {
		b.msg, b.result, b.phase = "全滅…", 2, phMessage
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
	if b.spr != nil && len(b.enemies) > 0 {
		en := len(b.enemies)
		flash := b.flashCol > 0
		for i := range b.enemies {
			if !b.enemies[i].alive() {
				continue
			}
			gx := ScreenW*(i+1)/(en+1) - b.spr.W/2
			gy := groundY + 8 - b.spr.H
			for r := 0; r < b.spr.H; r++ {
				for x := 0; x < b.spr.W; x++ {
					if !b.spr.Opaque[r][x] {
						continue
					}
					c := dq3data.Color{R: 200, G: 200, B: 200}
					if idx := int(b.spr.Px[r][x]); idx < len(b.mpal) {
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
				cx := gx + b.spr.W/2
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
	}
	// 下方左:指令窗(120,236,150×108)—— 指令 或 咒文子選單
	{
		mx, my, mw, mh := 120, 236, 150, 108
		fillBox(rgba, mx, my, mw, mh, white)
		b.hits.reset()
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
		} else {
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
	// 下方右:敵名框(290,262,250×36)—— 敵名 = D3TXT00 record 0x258+monID + 數量
	{
		ex, ey, ew, eh := 290, 262, 250, 36
		fillBox(rgba, ex, ey, ew, eh, white)
		b.drawName(rgba, ex+12, ey+8, 0x258+b.monID, white)
		alive := 0
		for i := range b.enemies {
			if b.enemies[i].alive() {
				alive++
			}
		}
		drawNumber(rgba, b.tx, ex+ew-40, ey+8, alive, white)
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

// drawASCII 畫數字部分(非數字字元跳過;中文訊息 glyph 對照另補)。
func drawASCII(rgba []byte, tx *dq3data.Text, x, y int, s string, fg dq3data.Color) {
	col := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			drawGlyph(rgba, tx, x+col*dq3data.GlyphPx, y, int(c-'0'), fg)
			col++
		} else if c == ' ' {
			col++
		}
	}
}
