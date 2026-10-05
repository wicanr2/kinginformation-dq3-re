package game

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestFieldItemDropPhysicalPositionAndFailures(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	words := []uint16{0x00ff, 1, 3, 0x00ff, 0x00ff, 0x00ff, 0x00ff, 0x801e}
	store, err := g.pack.ItemStoreFromWords(words)
	if err != nil {
		t.Fatal(err)
	}
	g.items, g.panelActor, g.panelCursor = store, 0, 0
	if !g.selectPanelItem() || g.itemSelected != 1 {
		t.Fatal("normal selection did not translate visible row to physical position")
	}
	if !g.dropSelectedItem() {
		t.Fatal("hole before first visible item blocked drop")
	}
	words[1] = 0x00ff
	if !reflect.DeepEqual(g.items.Words(), words) {
		t.Fatal("drop cleared visible ordinal instead of physical position")
	}
	g.itemSelected = 7
	before, rng := g.snapshot(), g.prng
	if g.dropSelectedItem() || !equalFieldSave(before, g.snapshot()) || rng != g.prng {
		t.Fatal("worn failure consumed state")
	}
	g.itemSelected = 99
	if g.dropSelectedItem() || !equalFieldSave(before, g.snapshot()) {
		t.Fatal("invalid selection consumed state")
	}
}

func TestFieldItemDropPromptRequiresFreshConfirmation(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		g := fieldSaveLoadComponentGame(t)
		s, p := g.pack.Interface.FieldItems.Drop, g.pack.Interface.FieldItems.GivePresentation
		id := s.SuccessTextID
		if blocked {
			id = s.BlockedTextID
		}
		codes, _ := g.pack.TextGlyphCodes(id)
		d := Dialogue{layout: p.Window}
		d.openRecord(codes)
		d.prelude = &gamepack.OpeningPrelude{VariableCodeWords: p.VariableCodeWords, ReturnMode: "confirm"}
		d.varGlyph = map[uint16][]int{uint16(*s.ActorVariableCode): {0}, uint16(*s.ItemVariableCode): {210, 210, 210}}
		g.itemGivePrompt = &fieldItemPromptState{dialogue: d}
		before, rng := g.snapshot(), g.prng
		count := d.pageCellCount() * p.Window.GlyphHoldFrames
		for i := 0; i < count; i++ {
			g.stepItemGivePrompt(InputState{Enter: true, Confirm: true, Cancel: true, SaveMenu: true, LoadMenu: true})
			if g.itemGivePrompt == nil || !equalFieldSave(before, g.snapshot()) || rng != g.prng {
				t.Fatal("drop reveal consumed fresh key or changed persistence")
			}
		}
		if !g.itemGivePrompt.dialogue.waitingForConfirm() {
			t.Fatal("drop reveal incomplete")
		}
		g.stepItemGivePrompt(InputState{Cancel: true, SaveMenu: true, LoadMenu: true})
		if g.itemGivePrompt == nil {
			t.Fatal("unrelated input dismissed drop result")
		}
		g.stepItemGivePrompt(InputState{Enter: true})
		if g.itemGivePrompt != nil || !equalFieldSave(before, g.snapshot()) || rng != g.prng {
			t.Fatal("drop fresh-key return changed state")
		}
	}
}

func TestFieldItemDropCapacityFailurePreservesWords(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	store, err := g.pack.ItemStoreFromWords([]uint16{0, 0x00ff, 0x00ff, 0x00ff, 0x00ff, 0x00ff, 0x00ff, 0x801e})
	if err != nil {
		t.Fatal(err)
	}
	g.items, g.panelActor, g.itemSelected = store, 0, 0
	g.heroName = make([]int, 20)
	g.rgba = make([]byte, ScreenW*ScreenH*4)
	before, rng := g.snapshot(), g.prng
	if g.beginSingleHeroDrop() || g.itemGivePrompt != nil || !equalFieldSave(before, g.snapshot()) || rng != g.prng {
		t.Fatal("oversized binding consumed item before validating presentation")
	}
}

func TestFieldItemDropDosgolemNormalInputComparison(t *testing.T) {
	runFieldItemOrderedNormalAt230(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const sourceHash = "5ca9fce6caff5bd1c1f27b489fba1ef9f687a808374d7fd544e94728d45f1971"
		source := filepath.Join(dir, "issue4-item-drop-normal-r3-source-r1-receipt.json")
		raw, err := os.ReadFile(source)
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
			t.Fatal("normal drop source identity differs", err)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if err := json.Unmarshal(raw, &src); err != nil || len(src.States) != 261 || len(src.Queued) != 261 || len(src.Artifacts) != 591 {
			t.Fatal("normal drop source shape differs", err)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(filepath.Join(dir, a.Path))
			if filepath.Base(a.Path) != a.Path || e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("normal drop source artifact differs", a.Path, e)
			}
		}
		before, rng := g.snapshot(), g.prng
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var samples []map[string]any
		for n := 231; n <= 261; n++ {
			in := idle
			in.AnyKeyEdge = true
			scan, e := strconv.ParseInt(src.Queued[n-1]["scan"], 16, 64)
			if e != nil {
				t.Fatal(e)
			}
			switch scan {
			case 0x39:
				in.Confirm = true
			case 0x1c:
				in.Enter = true
			case 0x50:
				in.DirEdge = 0
			case 0x48:
				in.DirEdge = 1
			case 0x4d:
				in.DirEdge = 3
			default:
				t.Fatal("unsupported normal drop input", n)
			}
			if e := g.step(in); e != nil {
				t.Fatal(e)
			}
			if e := g.step(idle); e != nil {
				t.Fatal(e)
			}
			if src.States[n-1]["phase"] == "waiting" {
				if g.itemGivePrompt == nil {
					t.Fatal("normal action did not open native message", n)
				}
				for updates := 0; !g.itemGivePrompt.dialogue.waitingForConfirm(); updates++ {
					if updates >= 100 {
						t.Fatal("native action did not reach fresh-key wait", n)
					}
					if e := g.step(idle); e != nil {
						t.Fatal(e)
					}
				}
			}
			actor, e := hex.DecodeString(src.States[n-1]["actor"])
			if e != nil || len(actor) != 128 {
				t.Fatal("original drop actor shape", e)
			}
			words := make([]uint16, 8)
			for i := range words {
				words[i] = binary.LittleEndian.Uint16(actor[0x3a+2*i:])
			}
			if !reflect.DeepEqual(g.items.Words(), words) {
				t.Fatalf("normal physical drop words packet%d got%04x want%04x", n, g.items.Words(), words)
			}
			actual := g.snapshot()
			want := before
			want.Items, want.itemStore = actual.Items, actual.itemStore
			if !equalFieldSave(want, actual) || rng != g.prng {
				t.Fatal("drop changed state beyond reviewed physical words", n)
			}
			if n == 247 {
				entry, valid := g.selectedItemEntry()
				if !valid || entry.Position != 1 || entry.Code != 1 {
					t.Fatal("visible first row after hole must select physical1", entry, valid)
				}
			}
			if n == 257 {
				entry, valid := g.selectedItemEntry()
				if !valid || entry.Position != 7 || entry.Word != 0x801e {
					t.Fatal("worn final physical slot selection differs", entry, valid)
				}
			}
			if n == 241 || n == 250 || n == 260 {
				d := &g.itemGivePrompt.dialogue
				s := g.pack.Interface.FieldItems.Drop
				id := s.SuccessTextID
				if n == 260 {
					id = s.BlockedTextID
				}
				codes, _ := g.pack.TextGlyphCodes(id)
				if !reflect.DeepEqual(d.buf, codes) {
					t.Fatal("native drop message differs", n)
				}
				if n != 260 {
					code := 0
					if n == 250 {
						code = 1
					}
					if !reflect.DeepEqual(d.varGlyphs(uint16(*s.ActorVariableCode)), g.equipActorName(0)) || !reflect.DeepEqual(d.varGlyphs(uint16(*s.ItemVariableCode)), itemNameGlyphs(g.shop.nameText, code)) {
						t.Fatal("drop actor/item binding differs", n)
					}
				}
			}
			if n == 233 || n == 242 || n == 251 || n == 261 {
				if g.itemGivePrompt != nil || g.dlg.open || g.panel != panelNone || g.cmd.open {
					t.Fatal("fresh Enter did not return to field", n)
				}
			}
			g.renderFrame()
			path := filepath.Join(dest, fmt.Sprintf("drop-packet-%03d.png", n))
			f, e := os.Create(path)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf("issue4-item-drop-normal-r3-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			encoded, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			var npcs []map[string]int
			for _, npc := range g.cur.npcs {
				npcs = append(npcs, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "words": words, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(encoded)), "png_size": len(encoded), "npc_visuals": npcs})
			t.Logf("normal drop packet%d complete640x350 RGB difference=%d", n, diff)
		}
		b, e := json.MarshalIndent(map[string]any{"scope": "normal new-game to261 sole healthy hero: two drops and worn rejection", "source_sha256": sourceHash, "samples": samples, "game_state_injection": false, "only_reviewed_slot_transaction": true, "rng_unchanged": true, "fresh_enter_returns_to_field": true, "save_version": saveFormatVersion, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "animation_timing_parity": false}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(dest, "drop-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}
