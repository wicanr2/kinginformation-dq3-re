package game

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// This optional test consumes the separately validated cold-boot receipt.
// Full canvas differences are recorded; NPC timing and original saves remain open.
func TestDosgolemKingReturnNormalInput(t *testing.T) {
	path := os.Getenv("DQ3_KING_RETURN_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_KING_RETURN_ORIGINAL to the verified cold-boot return receipt")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Scenario string            `json:"scenario"`
		Injected bool              `json:"game_state_injection"`
		Inputs   []json.RawMessage `json:"player_input"`
		Queued   []string          `json:"queued_events"`
		Ready    []string          `json:"return_events"`
		Cameras  []string          `json:"camera_events"`
	}
	if err = json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	if source.Scenario != "king_return_cold_ready" || source.Injected || len(source.Ready) != 85 || len(source.Inputs) != 100+len(source.Queued) {
		t.Fatal("unverified return source shape")
	}
	fields := func(line string) map[string]string {
		result := map[string]string{}
		for _, word := range strings.Fields(line) {
			if pair := strings.SplitN(word, "=", 2); len(pair) == 2 {
				result[pair[0]] = pair[1]
			}
		}
		return result
	}
	number := func(row map[string]string, key string, base int) int {
		t.Helper()
		n, e := strconv.ParseInt(row[key], base, 64)
		if e != nil {
			t.Fatalf("invalid native %s: %v", key, e)
		}
		return int(n)
	}
	type event struct {
		step int
		kind string
		row  map[string]string
	}
	var events []event
	for _, line := range source.Queued {
		row := fields(line)
		events = append(events, event{number(row, "step", 10), row["kind"], row})
	}
	for _, line := range source.Ready {
		row := fields(line)
		events = append(events, event{number(row, "step", 10), "observe", row})
	}
	sort.Slice(events, func(i, j int) bool { return events[i].step < events[j].step })
	cameras := map[int]map[string]string{}
	for _, line := range source.Cameras {
		row := fields(line)
		cameras[number(row, "ordinal", 10)] = row
	}
	g, _ := runDosgolemThroneLanding(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	up := InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}
	enter := InputState{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}
	step := func(in InputState) {
		t.Helper()
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
	}
	press := func(in InputState) {
		t.Helper()
		for i := 0; g.cd > 0 && !g.dlg.open && i < 120; i++ {
			step(idle)
		}
		if g.cd > 0 && !g.dlg.open {
			t.Fatal("cooldown did not finish")
		}
		step(in)
		step(idle)
		for i := 0; g.cd > 0 && !g.dlg.open && i < 120; i++ {
			step(idle)
		}
	}
	naturalIdle := func() {
		t.Helper()
		status := g.pack.Interface.FieldIdleStatus
		if !status.Enabled(g.curCty, g.cur.sec) {
			t.Fatalf("original idle scene missing: CTY=%d section=%d", g.curCty, g.cur.sec)
		}
		for i := 0; i < 300 && !g.fieldIdle.open; i++ {
			prior := g.fieldIdle.elapsed
			step(idle)
			if g.fieldIdle.open && prior < status.HoldFrames()-1 {
				t.Fatal("idle window opened before declared threshold")
			}
		}
		if !g.fieldIdle.open {
			t.Fatalf("original idle window missing: CTY=%d section=%d position=%d,%d", g.curCty, g.cur.sec, g.px, g.py)
		}
	}
	waitText := func() {
		t.Helper()
		for i := 0; i < 3000 && g.dlg.open && !g.dlg.waitingForConfirm(); i++ {
			step(idle)
		}
	}
	naturalIdle()
	press(up)
	for i := 0; i < 14; i++ {
		press(up)
	}
	naturalIdle()
	for i := 0; i < 5; i++ {
		press(enter)
	}
	press(up)
	waitText()
	for i := 0; i < 9; i++ {
		press(enter)
		waitText()
	}
	if g.px != 9 || g.py != 7 || g.heroGold != 50 || g.dlg.open {
		t.Fatal("normal audience checkpoint missing")
	}
	prefix := strings.TrimSuffix(filepath.Base(path), "-receipt.json")
	var samples []map[string]any
	writePNG := func(label string) {
		t.Helper()
		if out := os.Getenv("DQ3_KING_RETURN_OUT"); out != "" {
			f, e := os.Create(out + "-" + label + ".png")
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
		}
	}
	completed, idleInputs, observations := 0, 0, 0
	for _, event := range events {
		row := event.row
		ordinal := number(row, "ordinal", 10)
		switch event.kind {
		case "motion":
			if row["ida_linear"] != "1997c" || ordinal != completed+1 {
				t.Fatal("native motion boundary/order differs")
			}
			direction, ok := map[int]int{0x50: 0, 0x48: 1, 0x4b: 2, 0x4d: 3}[number(row, "scan", 16)]
			if !ok {
				t.Fatal("unsupported native motion")
			}
			press(InputState{DirHeld: direction, DirEdge: direction, AnyKeyEdge: true})
			completed++
		case "idle_dismiss":
			if number(row, "scan", 16) != 0x1c || ordinal != completed {
				t.Fatal("native idle input/order differs")
			}
			before, _ := json.Marshal([]any{g.px, g.py, g.curCty, g.cur.sec, g.heroGold, testInventory(g.items), g.items.Equipment(), g.storyBits})
			naturalIdle()
			g.renderFrame()
			name := fmt.Sprintf("%s-idle-%02d.png", prefix, number(row, "packet", 10))
			full := sourceCanvasDifference(t, g, path, name)
			f, e := os.Open(filepath.Join(filepath.Dir(path), name))
			if e != nil {
				t.Fatal(e)
			}
			img, e := png.Decode(f)
			f.Close()
			if e != nil {
				t.Fatal(e)
			}
			s := g.pack.Interface.FieldIdleStatus
			w := g.pack.Interface.PartyHUD.WindowLayout
			left, _ := g.pack.TextDefinition(s.LeftTextID)
			right, _ := g.pack.TextDefinition(s.RightTextID)
			body, _ := g.pack.TextDefinition(s.ColumnTextID)
			w.Width = (left.Layout.Columns + right.Layout.Columns + len(g.partyHUDActors())*body.Layout.Columns) * dq3data.GlyphPx
			windowDiff, shadowDiff := 0, 0
			for y := w.Y; y < w.Y+w.Height+s.Shadow.OffsetY; y++ {
				for x := w.X; x < w.X+w.Width+s.Shadow.OffsetX; x++ {
					r, green, b, _ := img.At(x, y).RGBA()
					i := (y*ScreenW + x) * 4
					if g.rgba[i] != uint8(r>>8) || g.rgba[i+1] != uint8(green>>8) || g.rgba[i+2] != uint8(b>>8) {
						shadowDiff++
						if x < w.X+w.Width && y < w.Y+w.Height {
							windowDiff++
						}
					}
				}
			}
			if windowDiff != 0 {
				t.Fatalf("idle content differs at motion %d: window=%d full=%d", ordinal, windowDiff, full)
			}
			label := fmt.Sprintf("idle-%02d", number(row, "packet", 10))
			writePNG(label)
			samples = append(samples, map[string]any{"label": label, "ordinal": ordinal, "cty": g.curCty, "section": g.cur.sec, "full_rgb_difference": full, "window_rgb_difference": windowDiff, "window_shadow_rgb_difference": shadowDiff})
			press(enter)
			idleInputs++
			after, _ := json.Marshal([]any{g.px, g.py, g.curCty, g.cur.sec, g.heroGold, testInventory(g.items), g.items.Equipment(), g.storyBits})
			if string(before) != string(after) {
				t.Fatal("idle wait/dismiss changed gameplay state")
			}
		case "observe":
			if ordinal != completed || row["ida_linear"] != "1991d" {
				t.Fatal("native normal observation/order differs")
			}
			if g.px != number(row, "player_x", 10) || g.py != number(row, "player_y", 10) {
				t.Fatalf("native position differs at %d: remake=%d,%d original=%s,%s", ordinal, g.px, g.py, row["player_x"], row["player_y"])
			}
			data, e := os.ReadFile(filepath.Join(spineAssetsDir(t), fmt.Sprintf("CTY%02d.DAT", g.curCty)))
			if e != nil {
				t.Fatal(e)
			}
			if int(binary.LittleEndian.Uint16(data[g.cur.sec*2:])) != number(row, "raw0b24", 16) {
				t.Fatal("native section base differs")
			}
			camera := g.activeSceneCamera()
			native := cameras[ordinal]
			if camera == nil {
				t.Fatalf("native camera differs at %d CTY=%d section=%d", ordinal, g.curCty, g.cur.sec)
			}
			if native != nil && (native["ida_linear"] != "11991" || g.px-camera.AnchorX != number(native, "origin_x", 10) || g.py-camera.AnchorY != number(native, "origin_y", 10)) {
				t.Fatalf("measured native camera differs at %d", ordinal)
			}
			actor, e := hex.DecodeString(row["actor"])
			if e != nil || len(actor) != 128 {
				t.Fatal("native actor shape differs")
			}
			var inventory []int
			for off := 0x3a; off < 0x4a; off += 2 {
				item := binary.LittleEndian.Uint16(actor[off:])
				if item != 0xff && item&0x8000 == 0 {
					inventory = append(inventory, int(item))
				}
			}
			if !reflect.DeepEqual(testInventory(g.items), inventory) || g.heroGold != number(row, "gold_lo", 16)+(number(row, "gold_hi", 16)<<16) || fmt.Sprintf("%x", g.storyBits) != row["flags"] {
				t.Fatal("native inventory/gold/story differs")
			}
			g.renderFrame()
			label := fmt.Sprintf("step-%02d", ordinal)
			full := sourceCanvasDifference(t, g, path, prefix+"-return-ready-"+fmt.Sprintf("%02d", ordinal)+".png")
			writePNG(label)
			samples = append(samples, map[string]any{"label": label, "ordinal": ordinal, "cty": g.curCty, "section": g.cur.sec, "x": g.px, "y": g.py, "origin_x": g.px - camera.AnchorX, "origin_y": g.py - camera.AnchorY, "gold": g.heroGold, "full_rgb_difference": full})
			observations++
		default:
			t.Fatal("unknown native event")
		}
	}
	if completed != 85 || observations != 85 || 100+completed+idleInputs != len(source.Inputs) {
		t.Fatal("native return input count differs")
	}
	if err = g.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, e := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if e != nil {
		t.Fatal(e)
	}
	for _, in := range []InputState{{DirHeld: -1, DirEdge: -1, Enter: true}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: -1, Enter: true}} {
		if e = loaded.step(in); e != nil {
			t.Fatal(e)
		}
	}
	if loaded.showTitle || loaded.curCty != g.curCty || loaded.cur.sec != g.cur.sec || loaded.px != g.px || loaded.py != g.py || loaded.heroGold != g.heroGold || !reflect.DeepEqual(testInventory(loaded.items), testInventory(g.items)) || !reflect.DeepEqual(loaded.items.Equipment(), g.items.Equipment()) || !reflect.DeepEqual(loaded.storyBits, g.storyBits) {
		t.Fatal("title load lost return state")
	}
	camera := g.activeSceneCamera()
	if !reflect.DeepEqual(loaded.activeSceneCamera(), camera) {
		t.Fatal("title load lost camera binding")
	}
	g = loaded
	g.renderFrame()
	writePNG("title-load")
	oldX, oldY := g.px, g.py
	press(InputState{DirHeld: 0, DirEdge: 0, AnyKeyEdge: true})
	if g.px != oldX || g.py != oldY+1 {
		t.Fatal("normal loaded route could not continue")
	}
	g.renderFrame()
	writePNG("resume-one-step")
	report := map[string]any{"normal_inputs": len(source.Inputs), "motions": completed, "extra_idle_inputs": idleInputs, "observations": observations, "state_injection": false, "schema": g.pack.Schema(), "content_hash": g.pack.ContentHash(), "samples": samples, "title_save_load_state": true, "normal_loaded_continuation": true, "original_save_load_parity": false, "audio_parity": false, "full_rgb_parity": false}
	if out := os.Getenv("DQ3_KING_RETURN_OUT"); out != "" {
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(out+".json", append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("normal inputs=%d motions=%d idle=%d content=%s", len(source.Inputs), completed, idleInputs, g.pack.ContentHash())
}
