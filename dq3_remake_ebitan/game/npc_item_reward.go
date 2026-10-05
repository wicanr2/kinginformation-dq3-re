package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

func (g *Game) talkNPCItemReward(n *npcInst) bool {
	if g.pack == nil || g.cur == nil || n == nil {
		return false
	}
	for _, event := range g.pack.NPCItemRewardEvents() {
		selector := event.NPC
		if selector.CTYRaw != g.curCty || selector.Section != g.cur.sec ||
			selector.Tile != (gamepack.TileCoordinate{X: n.x, Y: n.y}) ||
			selector.HandlerRaw != n.b4 {
			continue
		}
		if event.RequiredItemRawID != nil {
			if !g.hasPartyItem(*event.RequiredItemRawID) {
				// The original handler has a distinct no-item branch, but its
				// dialogue record is not yet closed in the evidence ledger. Fail
				// closed without granting or consuming anything when that optional
				// text is absent.
				if event.DialogueTextIDs.Before != "" {
					g.openPackText(event.DialogueTextIDs.Before)
				}
				return true
			}
			if !g.openPackText(event.DialogueTextIDs.Success) {
				return true
			}
			if !event.ReplaceRequired || !g.replacePartyItem(*event.RequiredItemRawID, event.GrantedItemRaw) {
				return true
			}
			g.setStoryFlag(event.PresentFlagRaw, false)
			for _, flag := range event.ClearStoryFlagsRaw {
				g.setStoryFlag(flag, false)
			}
			g.noticeCode, g.noticeTimer = event.GrantedItemRaw, 120
			return true
		}
		if !g.storyFlag(event.PresentFlagRaw) {
			g.openPackText(event.DialogueTextIDs.After)
			return true
		}
		if !g.openPackText(event.DialogueTextIDs.Success) {
			return true
		}
		if g.grantPartyItem(event.GrantedItemRaw) {
			g.setStoryFlag(event.PresentFlagRaw, false)
			g.noticeCode, g.noticeTimer = event.GrantedItemRaw, 120
		}
		return true
	}
	return false
}

// replacePartyItem preserves the original party-order inventory slot while
// transforming one quest item into another. It never removes an item before
// confirming the replacement owner, so a full/invalid party fails closed.
func (g *Game) replacePartyItem(required, granted int) bool {
	for actor := 0; actor <= len(g.companions); actor++ {
		s := g.actorItemStore(actor)
		for _, entry := range s.Entries() {
			if entry.Code == required {
				return s.Replace(entry.Position, granted)
			}
		}
	}
	return false
}

// grantPartyItem 對齊原版共用取得物品 writer：依隊長至同伴尋找第一個
// 有空位的個人物品欄；每名角色的裝備亦占八格容量。
func (g *Game) grantPartyItem(code int) bool {
	if g.pack == nil {
		return false
	}
	for actor := 0; actor <= len(g.companions); actor++ {
		if _, ok := g.actorItemStore(actor).Add(code); ok {
			return true
		}
	}
	return false
}
