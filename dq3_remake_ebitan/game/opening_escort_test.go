package game

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"testing"
)

func TestOpeningMotherEscortUsesVisibleFrames(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.startOpening()
	// 局部帶路 fixture 明確建立生日返回後的場景；正式玩家輸入另行驗證。
	if !g.enterOpeningScene() {
		t.Fatal("出生場景 fixture 載入失敗")
	}
	g.dlg.prelude, g.dlg.preludeFrame = nil, nil
	g.dlg.layout = g.pack.DialogueWindowLayout()
	if !g.startMotherEscort() {
		t.Fatal("CTY00 sec4 應找到 pack 指定母親並開始逐格帶路")
	}
	steps := 0
	for g.openingEscortAnimating() && steps < 1000 {
		g.advanceOpeningEscort()
		steps++
	}
	if steps <= len(g.openingEscort.Frames) {
		t.Fatalf("帶路應有可見 hold frames，僅 %d 次更新", steps)
	}
	if g.cur.sec != g.openingEscort.Section || g.px != 5 || g.py != 5 || g.cur.npcAt(10, 10) < 0 {
		t.Fatal("家中序列未保持主角位置")
	}
	// 局部城鎮fixture；完整家中選圖及手動接近由正常InputState測試驗證。
	if !g.finishMotherEscort() {
		t.Fatal("城鎮fixture失敗")
	}
	g.openingEscortPhase = 1
	g.openingEscortIndex = 0
	g.applyOpeningArrivalFrame(g.openingEscort.ArrivalFrames[0])
	for g.openingEscortAnimating() && steps < 1500 {
		if err := g.advanceOpeningEscort(); err != nil {
			t.Fatal(err)
		}
		steps++
	}
	dialogueFrame := g.openingEscort.ArrivalFrames[g.openingEscort.DialogueFrameIndex].Player
	if g.curCty != 0 || g.cur == nil || g.cur.sec != 0 || g.px != dialogueFrame.X || g.py != dialogueFrame.Y {
		t.Fatalf("opening 對話序列觸發位置錯：cty=%d sec=%v pos=(%d,%d)，want=(%d,%d)",
			g.curCty, currentSceneSection(g.cur), g.px, g.py, dialogueFrame.X, dialogueFrame.Y)
	}
	if !g.dlg.open || g.openingEscortDialogue != 0 {
		t.Fatal("抵達 pack 指定格後應開第一段對話")
	}
	if !reflect.DeepEqual(g.dlg.buf, g.cur.dlgText.Record(80)) {
		t.Fatal("第一段不是 D3TXT01 record 80")
	}
	if g.dlg.layout.GlyphHoldFrames != 1 || g.dlg.revealCells != 0 || g.dlg.prelude == nil || g.dlg.shadow == nil {
		t.Fatalf("逐字初值錯：hold=%d visible=%d", g.dlg.layout.GlyphHoldFrames, g.dlg.revealCells)
	}
	g.dlg.Tick()
	if g.dlg.revealCells != 1 {
		t.Fatal("滿 pack hold frames 應顯示一個 glyph cell")
	}
	if g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatalf("rec80 關閉前不得交易旗標：17=%v 50=%v", g.storyFlag(0x17), g.storyFlag(0x50))
	}
	g.dlg.open = false
	if !g.resumeOpeningEscortAfterDialogue() {
		t.Fatal("record80 關閉後應恢復最後三步")
	}
	if g.dlg.open || g.openingEscortDialogue != 1 || g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatalf("record80 返回後應保持旗標，不能開 record79：open=%v index=%d flag17=%v flag50=%v",
			g.dlg.open, g.openingEscortDialogue, g.storyFlag(0x17), g.storyFlag(0x50))
	}
	lastIndex := g.openingEscort.DialogueFrameIndex
	for g.openingEscortAnimating() && steps < 2000 {
		g.advanceOpeningEscort()
		steps++
		if g.openingEscortIndex != lastIndex && g.openingEscortIndex >= 0 {
			lastIndex = g.openingEscortIndex
			if g.storyFlag(0x17) || !g.storyFlag(0x50) {
				t.Fatal("最後一步之前提早交易旗標")
			}
		}
	}
	if g.px != 21 || g.py != 17 || g.storyFlag(0x17) == false || g.storyFlag(0x50) {
		t.Fatalf("剩餘步行完成狀態錯：pos=(%d,%d) flag17=%v flag50=%v",
			g.px, g.py, g.storyFlag(0x17), g.storyFlag(0x50))
	}
	if n := g.cur.npcs[0]; n.recordIndex != 0 || n.x != 22 || n.y != 18 || n.facing != 0 {
		t.Fatalf("原始母親 record0 未停在 (22,18) 朝下：%+v", n)
	}
}

// This compares the town leg at its natural production checkpoint. The known
// home/modal mismatch remains outside this state-only assertion and is recorded.
func TestDosgolemMotherArrivalStateComparison(t *testing.T) {
	receiptPath := os.Getenv("DQ3_MOTHER_FINISH_ORIGINAL")
	if receiptPath == "" {
		t.Skip("set DQ3_MOTHER_FINISH_ORIGINAL to the freshly generated original receipt")
	}
	b, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Scenario string   `json:"scenario"`
		Injected bool     `json:"game_state_injection"`
		Events   []string `json:"mother_entry_events"`
	}
	if err = json.Unmarshal(b, &original); err != nil {
		t.Fatal(err)
	}
	if original.Scenario != "mother_finish" || original.Injected {
		t.Fatal("不是原版自然續跑收據")
	}
	type state struct{ PlayerX, PlayerY, LeaderX, LeaderY, LeaderFacing int }
	var expected []state
	fieldsRE := regexp.MustCompile(`(\w+)=(\S+)`)
	addresses := map[string]bool{"10130": true, "10148": true, "10163": true, "1017e": true, "10199": true, "101b4": true, "101c5": true, "101e6": true, "101ef": true, "101f8": true}
	for _, event := range original.Events {
		fields := map[string]string{}
		for _, pair := range fieldsRE.FindAllStringSubmatch(event, -1) {
			fields[pair[1]] = pair[2]
		}
		if !addresses[fields["ida_linear"]] {
			continue
		}
		raw, err := hex.DecodeString(fields["npc0"])
		if err != nil || len(raw) != 8 {
			t.Fatal("原始NPC觀察欄位不完整")
		}
		x, err := strconv.Atoi(fields["player_x"])
		if err != nil {
			t.Fatal(err)
		}
		y, err := strconv.Atoi(fields["player_y"])
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, state{x, y, int(raw[0]), int(raw[1]), []int{0, 2, 1, 3}[raw[3]&3]})
	}
	if len(expected) != 42 {
		t.Fatalf("原版狀態數=%d want42", len(expected))
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "arrival-save.json"))
	g, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	g.prng.Seed(0x1357)
	pictureSeed := uint16(0x151b)
	g.homeSelection.seed = &pictureSeed
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	for _, scan := range []int{0x1c, 0x1c, 0x48, 0x4b, 0x1c, 0x1c, 0x4d, 0x50, 0x1c, 0x48, 0x4b, 0x1c, 0x48, 0x1c, 0x1c, 0x1c, 0x1c} {
		in := idle
		switch scan {
		case 0x1c:
			in.Enter = true
		case 0x48:
			in.DirEdge = 1
		case 0x50:
			in.DirEdge = 0
		case 0x4b:
			in.DirEdge = 2
		case 0x4d:
			in.DirEdge = 3
		}
		step(in)
	}
	confirmations := 0
	for i := 0; i < 4000 && !g.homeAwait; i++ {
		in := idle
		if g.dlg.waitingForConfirm() || g.homeSelection.active {
			in.Enter = true
			confirmations++
		}
		step(in)
	}
	if !g.homeAwait {
		t.Fatal("正常選圖沒有交還控制")
	}
	for _, dir := range []int{0, 0, 2, 2, 0, 0, 0, 3, 3, 3, 3, 3, 3} {
		for g.cd > 0 {
			step(idle)
		}
		in := idle
		in.DirHeld = dir
		step(in)
	}
	for i := 0; i < 100 && g.openingEscortPhase == 5; i++ {
		step(idle)
	}
	if confirmations != 7 || g.openingEscortPhase != 1 || g.openingEscortIndex != 0 {
		t.Fatal("正常創角無法抵達城鎮帶路checkpoint")
	}
	observed := make([]state, 0, len(expected))
	lastIndex := -1
	capture := func(index int) {
		t.Helper()
		if index < 0 || index >= len(expected) {
			t.Fatal("狀態索引越界")
		}
		var npc *npcInst
		for i := range g.cur.npcs {
			if g.cur.npcs[i].recordIndex == 0 {
				npc = &g.cur.npcs[i]
				break
			}
		}
		if npc == nil {
			t.Fatal("遺失原始母親record0")
		}
		got := state{g.px, g.py, npc.x, npc.y, npc.facing}
		if got != expected[index] {
			t.Fatalf("原版checkpoint frame%d got%+v want%+v", index, got, expected[index])
		}
		observed = append(observed, got)
		lastIndex = index
	}
	capture(0)
	for i := 0; i < 4000 && g.openingEscortPhase != 2; i++ {
		step(idle)
		if g.openingEscortIndex != lastIndex {
			capture(g.openingEscortIndex)
		}
	}
	if g.openingEscortPhase != 2 || !g.dlg.open || lastIndex != 38 || g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatal("城門對話與旗標gate不符")
	}
	for i := 0; i < 2000 && !g.dlg.waitingForConfirm(); i++ {
		step(idle)
	}
	if !g.dlg.waitingForConfirm() {
		t.Fatal("record80未達內嵌確認")
	}
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	for i := 0; i < 2000 && g.openingEscortPhase == 2; i++ {
		step(idle)
	}
	if g.dlg.open || g.openingEscortPhase != 3 || g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatal("單次確認後應自動EOF，旗標仍未交易")
	}
	for i := 0; i < 1000 && g.openingEscortAnimating(); i++ {
		before := g.openingEscortIndex
		step(idle)
		if g.openingEscortIndex < 0 {
			capture(len(expected) - 1)
		} else if g.openingEscortIndex != before {
			capture(g.openingEscortIndex)
			if g.storyFlag(0x17) || !g.storyFlag(0x50) {
				t.Fatal("最後一步之前交易旗標")
			}
		}
	}
	if len(observed) != 42 || g.px != 21 || g.py != 17 || !g.storyFlag(0x17) || g.storyFlag(0x50) {
		t.Fatal("原版最後位置與旗標交易不符")
	}
	assertArrivalCanvas(t, g, receiptPath)
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []InputState{
		{DirHeld: -1, DirEdge: -1, Enter: true},
		{DirHeld: -1, DirEdge: 0},
		{DirHeld: -1, DirEdge: -1, Enter: true},
	} {
		if err = loaded.step(in); err != nil {
			t.Fatal(err)
		}
	}
	if loaded.showTitle || loaded.px != g.px || loaded.py != g.py || loaded.curCty != g.curCty || loaded.cur.sec != g.cur.sec || !loaded.storyFlag(0x17) || loaded.storyFlag(0x50) {
		t.Fatal("帶路完成後存讀檔遺失位置或旗標")
	}
	camera, restoredCamera := g.activeSceneCamera(), loaded.activeSceneCamera()
	if camera == nil || restoredCamera == nil || *camera != *restoredCamera {
		t.Fatal("正常標題讀檔遺失目的場景camera")
	}
	if prefix := os.Getenv("DQ3_MOTHER_STATE_OUT"); prefix != "" {
		f, err := os.Create(prefix + "-final.png")
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
		closeErr := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		report := map[string]any{"scope": "正常創角、家中選圖及手動接近後的城鎮42狀態與最後1020A完整RGB；完整開場及音訊仍未V3", "original_receipt": receiptPath, "original_seed": "1357一次自然Lv1入口", "remake_seed": "1357首次正式InputState前", "remake_confirmations_before_town": confirmations, "castle_confirmations": 1, "state_count": len(observed), "states": observed, "save_load_player_flags": true, "save_load_camera": true, "final_canvas_rgb_parity": true, "visual_parity": false, "audio_parity": false, "pack_schema": g.pack.Schema(), "pack_content_hash": g.pack.ContentHash()}
		out, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(prefix+"-receipt.json", append(out, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Log(fmt.Sprintf("正常家中選圖、手動接近、城鎮42狀態及最後完整RGB通過；標題讀檔camera維持，完整開場與音訊未完成"))
}

func TestOpeningKingAudienceRendersHero(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = throne, throne, ctyAliahanCastle, true
	g.px, g.py, g.facing = aliahanKingX, aliahanKingY+1, 1
	g.showTitle = false
	g.renderFrame()
	visible := append([]byte(nil), g.rgba...)
	g.remoaru = 1
	g.renderFrame()

	camX := clampi(g.px-ViewCols/2, 0, max0(throne.w-ViewCols))
	camY := clampi(g.py-ViewRows/2, 0, max0(throne.h-ViewRows))
	sx, sy := (g.px-camX)*TileW, (g.py-camY)*TileH
	different := false
	for y := sy; y < sy+TileH && !different; y++ {
		for x := sx; x < sx+TileW; x++ {
			o := (y*ScreenW + x) * 4
			if visible[o] != g.rgba[o] || visible[o+1] != g.rgba[o+1] || visible[o+2] != g.rgba[o+2] {
				different = true
				break
			}
		}
	}
	if !different {
		t.Fatal("王座正前方的可見畫面與透明狀態相同：勇者 sprite 未繪出")
	}
}

func TestOpeningArrivalMissingActorFailsClosed(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.startOpening()
	if !g.enterOpeningScene() {
		t.Fatal("出生場景 fixture 載入失敗")
	}
	broken := *g.openingEscort
	unknownRecord := 999
	broken.ArrivalLeaderRecord = &unknownRecord
	g.openingEscort = &broken
	if !g.startMotherEscort() {
		t.Fatal("家中 fixture 應可開始")
	}
	home := g.cur
	for i := 0; i < 1000 && g.openingEscortAnimating(); i++ {
		if err := g.advanceOpeningEscort(); err != nil {
			t.Fatal(err)
		}
	}
	if g.finishMotherEscort() || g.cur != home || g.cur.sec != broken.Section || g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatal("缺少原始領路NPC時跳過轉場或消耗旗標")
	}
}

// TestDumpOpeningCastleArrivalPath is an opt-in evidence helper: it derives a
// legal shortest tile path from the original CTY00 data between handler54's
// transition landing and the castle portal.  The printed coordinates are only
// a candidate until checked against the original video; production JSON must
// not be generated from this test alone.
func TestDumpOpeningCastleArrivalPath(t *testing.T) {
	if os.Getenv("DQ3_DUMP_OPENING_CASTLE_PATH") == "" {
		t.Skip("set DQ3_DUMP_OPENING_CASTLE_PATH=1 to derive the CTY00 candidate route")
	}
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !g.finishMotherEscort() {
		t.Fatal("無法載入 pack 指定的城鎮抵達段")
	}
	path := tracePath(g.cur, g.px, g.py, 21, 9)
	if len(path) == 0 {
		t.Fatal("CTY00 landing cannot reach the tile before the castle portal")
	}
	x, y := g.px, g.py
	t.Logf("route start=(%d,%d)", x, y)
	for i, dir := range path {
		dx, dy := dirDelta(dir)
		x, y = x+dx, y+dy
		t.Logf("route[%d]=(%d,%d) dir=%d", i+1, x, y, dir)
	}
}
