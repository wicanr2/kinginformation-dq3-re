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
	for i, item := range g.inventory {
		if item == required {
			g.inventory[i] = granted
			return true
		}
	}
	for i := range g.companions {
		for j, item := range g.companions[i].Inventory {
			if item == required {
				g.companions[i].Inventory[j] = granted
				return true
			}
		}
	}
	return false
}

// grantPartyItem follows the original common reward consumer: search the
// leader, then companions, and store in the first personal inventory with a
// free slot. Equipped items count toward each eight-slot record capacity.
func (g *Game) grantPartyItem(code int) bool {
	if g.pack == nil {
		return false
	}
	slots := g.pack.ItemActions().PersonalInventorySlots
	heroCount := len(g.inventory)
	for _, equipped := range g.equip {
		// 現行裝備空值是 -1；raw item 0 是合法武器，不可當空槽。
		if equipped >= 0 {
			heroCount++
		}
	}
	if heroCount < slots {
		g.inventory = append(g.inventory, code)
		return true
	}
	for _, member := range g.companions {
		if member.itemCount() < slots {
			member.Inventory = append(member.Inventory, code)
			return true
		}
	}
	return false
}
