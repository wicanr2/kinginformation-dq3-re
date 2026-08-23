package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// talkSpecialShop 把 pack 選定的 scripted NPC 接入共用商店狀態機。引擎只執行通用的
// 兩旗標交易；城鎮、道具與 story flag 身分全部由 pack 提供。
func (g *Game) talkSpecialShop(n *npcInst) bool {
	if g.pack == nil {
		return false
	}
	for _, event := range g.pack.SpecialShopEvents() {
		if !g.scriptedNPCMatches(n, event.NPC) {
			continue
		}
		g.openSpecialShop(event)
		return true
	}
	return false
}

func (g *Game) openSpecialShop(event gamepack.SpecialShopEvent) {
	g.openFacility(event.FacilityK)
	if !g.shop.active {
		return
	}
	g.shop.kind = "special"
	g.shop.unlockSellItemRawID = event.UnlockSellItemRawID
	g.shop.unlockFlagRaw = event.UnlockFlagRaw
	g.shop.conditionalItemRawID = event.ConditionalItemRawID
	g.shop.conditionalItemFlagRaw = event.ConditionalItemFlagRaw
	if !g.storyFlag(event.UnlockFlagRaw) && g.storyFlag(event.ConditionalItemFlagRaw) {
		g.shop.codes = append(g.shop.codes, event.ConditionalItemRawID)
	}
}

func (g *Game) completeSpecialShopSale(itemRawID int) {
	if g.shop.kind == "special" && itemRawID == g.shop.unlockSellItemRawID &&
		g.shop.unlockFlagRaw >= 0 {
		g.setStoryFlag(g.shop.unlockFlagRaw, false)
	}
}

func (g *Game) completeSpecialShopPurchase(itemRawID int) {
	if g.shop.kind != "special" || itemRawID != g.shop.conditionalItemRawID ||
		g.shop.conditionalItemFlagRaw < 0 {
		return
	}
	g.setStoryFlag(g.shop.conditionalItemFlagRaw, false)
	for i, code := range g.shop.codes {
		if code == itemRawID {
			g.shop.codes = append(g.shop.codes[:i], g.shop.codes[i+1:]...)
			if g.shop.cursor >= len(g.shop.codes) && g.shop.cursor > 0 {
				g.shop.cursor--
			}
			break
		}
	}
}
