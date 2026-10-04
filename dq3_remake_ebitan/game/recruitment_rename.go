package game

// docs/188 的單人隊伍 READY。原版重新依隊伍指標選人，
// 改名目標是主角；正在觀看的名冊角色與能力頁副本保持。
func (g *Game) startRecruitmentRename() {
	rc := &g.recruit
	if rc.selection == nil || rc.selection.ViewRename == nil || rc.viewFlow == nil ||
		rc.selection.ViewRename.TargetScope != "singleton_party_leader" || len(g.companions) != 0 ||
		len(g.heroName) == 0 || g.tavern.geometry == nil || g.tavern.labels == nil {
		return
	}
	nf := &NewGameFlow{stage: ngName, labels: g.tavern.labels, geometry: g.tavern.geometry}
	nf.ni.setLabels(nf.labels)
	nf.ni.Init()
	rc.renameFlow, rc.stage = nf, rcViewRename
}

func (g *Game) recruitmentRenameInput(in InputState) {
	rc := &g.recruit
	if rc.renameFlow == nil || rc.selection == nil || rc.selection.ViewRename == nil {
		return
	}
	ni := &rc.renameFlow.ni
	tap := -1
	if in.Tapped {
		tap = ni.hits.at(in.TapX, in.TapY)
	}
	confirmed, canceled := ni.input(in, tap)
	if !canceled {
		if !confirmed || rc.selection.ViewRename.RequireNonemptyName == nil ||
			*rc.selection.ViewRename.RequireNonemptyName && len(ni.nameBuf) == 0 ||
			len(ni.nameBuf) > rc.selection.NameCapacity {
			return
		}
		g.heroName = append([]int(nil), ni.nameBuf...)
		g.dlg.heroName = g.heroName
	}
	rc.renameFlow, rc.viewFlow = nil, nil
	rc.startSelectionText(rc.selection.AgainTextID, rcAgain)
}
