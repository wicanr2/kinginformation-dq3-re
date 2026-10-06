package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// Recorded native states isolate the automatic mover. These are component
// seeds and fixtures, never a replacement for the normal-input player trace.
func TestNPCMotionRecordedTurnLayers(t *testing.T) {
	dir := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional native four-layer component oracle")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "issue4-npc-turn-layers-component-r1-fixture-r1.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "e7ba85ee6c4645229d38c46a859808c36d55dc38a57751cca2b24f84866d5498" {
		t.Fatal("native layer fixture identity", err)
	}
	var source struct {
		NormalPlayerPath        bool   `json:"normal_player_path"`
		StateInjection          bool   `json:"state_injection"`
		CPURegisterReentry      bool   `json:"cpu_register_reentry"`
		EmulatorSnapshotRestore bool   `json:"emulator_snapshot_restore"`
		ComponentSeed           string `json:"component_seed"`
		NPCSlots                string `json:"npc_slots"`
		Slot                    int    `json:"slot"`
		Cases                   []struct {
			Entry  map[string]string `json:"entry"`
			Return map[string]string `json:"return_state"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(raw, &source); err != nil || source.NormalPlayerPath || !source.StateInjection || !source.CPURegisterReentry || source.EmulatorSnapshotRestore || source.ComponentSeed != "1e2c" || len(source.Cases) != 4 {
		t.Fatal("native controlled layer scope", err)
	}
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	before, err := hex.DecodeString(source.NPCSlots)
	if err != nil || len(before) != 112 {
		t.Fatal("native table", err)
	}
	value := func(row map[string]string, key string, base int) int {
		t.Helper()
		n, err := strconv.ParseInt(row[key], base, 32)
		if err != nil {
			t.Fatal(key, err)
		}
		return int(n)
	}
	for _, c := range source.Cases {
		sc := &Scene{w: 42, h: 43, hiMap: make([]byte, 42*43), attr: &dq3data.BlockAttr{A: make([]uint16, 256)}}
		sc.tileAt = func(x, y int) int { return 0 }
		for j := 0; j < len(before)/8; j++ {
			b := before[j*8:]
			sc.npcs = append(sc.npcs, npcInst{x: int(b[0]), y: int(b[1]), ctrl: int(b[3])})
		}
		n := sc.npcs[source.Slot]
		sc.hiMap[n.y*sc.w+n.x] = byte(value(c.Entry, "word", 16) >> 8)
		sc.npcRng.Seed(uint16(value(c.Entry, "seed", 16)))
		g := &Game{cur: sc, inTown: true, pack: p, px: value(c.Entry, "player_x", 10), py: value(c.Entry, "player_y", 10)}
		g.npcStep(source.Slot)
		n = sc.npcs[source.Slot]
		if n.ctrl != value(c.Return, "ctrl", 16) || n.x != value(c.Return, "x", 10) || n.y != value(c.Return, "y", 10) || sc.npcRng.State() != uint16(value(c.Return, "seed", 16)) {
			t.Fatalf("native layer%s got ctrl%02x (%d,%d) seed%04x", c.Entry["layer"], n.ctrl, n.x, n.y, sc.npcRng.State())
		}
	}
	t.Log("four controlled native layers matched; explicit state injection and CPU reentry; not normal player parity")
}

func TestNPCMotionRecordedNativeEntryReturn(t *testing.T) {
	dir := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional native component oracle")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "issue4-npc-move-state-normal-r1-component-r1.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "6586abb198648dc27859901813f2c40cad156579b511c101db820eb7023b69b4" {
		t.Fatal("native component identity", err)
	}
	var source struct {
		SourceSHA256 string `json:"source_sha256"`
		Cases        []struct {
			Entry, Return map[string]string
			Events        []map[string]string
		}
	}
	if err = json.Unmarshal(raw, &source); err != nil || source.SourceSHA256 != "b2fdf7818b4d8ca07d97d5796db6ac241490a6899f8352ab2396cabb3b96c9c6" || len(source.Cases) != 1672 {
		t.Fatal("native source shape", err)
	}
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	value := func(row map[string]string, name string, base int) int {
		t.Helper()
		n, err := strconv.ParseInt(row[name], base, 32)
		if err != nil {
			t.Fatal(name, err)
		}
		return int(n)
	}
	turns, rolls, moves := 0, 0, 0
	for i, c := range source.Cases {
		before, err := hex.DecodeString(c.Entry["npc_slots"])
		if err != nil {
			t.Fatal(err)
		}
		after, err := hex.DecodeString(c.Return["npc_slots"])
		if err != nil {
			t.Fatal(err)
		}
		w, h := value(c.Entry, "width", 10), value(c.Entry, "height", 10)
		tiles := make([]byte, w*h)
		sc := &Scene{w: w, h: h, hiMap: make([]byte, w*h), attr: &dq3data.BlockAttr{A: make([]uint16, 256)}}
		sc.tileAt = func(x, y int) int { return int(tiles[y*w+x]) }
		for j := 0; j < len(before)/8; j++ {
			b := before[j*8:]
			sc.npcs = append(sc.npcs, npcInst{x: int(b[0]), y: int(b[1]), ctrl: int(b[3])})
		}
		idx := value(c.Entry, "slot", 10)
		n := sc.npcs[idx]
		word := value(c.Entry, "cell_word", 16)
		sc.hiMap[n.y*w+n.x] = byte(word >> 8)
		tiles[n.y*w+n.x] = byte(word)
		if c.Entry["target_valid"] == "true" {
			x, y := value(c.Entry, "target_x", 10), value(c.Entry, "target_y", 10)
			word := value(c.Entry, "target_word", 16)
			sc.hiMap[y*w+x] = byte(word >> 8)
			tiles[y*w+x] = byte(word)
			sc.attr.A[word&255] = uint16(value(c.Entry, "target_attr", 16))
		}
		sc.npcRng.Seed(uint16(value(c.Entry, "seed", 16)))
		g := &Game{cur: sc, inTown: true, pack: p, px: value(c.Entry, "player_x", 10), py: value(c.Entry, "player_y", 10)}
		g.npcStep(idx)
		if sc.npcRng.State() != uint16(value(c.Return, "seed", 16)) {
			t.Fatalf("case %d packet%s native RNG mismatch: got%04x want%s", i, c.Entry["packet"], sc.npcRng.State(), c.Return["seed"])
		}
		for j, n := range sc.npcs {
			b := after[j*8:]
			if n.x != int(b[0]) || n.y != int(b[1]) || n.ctrl != int(b[3]) {
				t.Fatalf("case %d packet%s actor%d got(%d,%d,%02x) want(%d,%d,%02x)", i, c.Entry["packet"], j, n.x, n.y, n.ctrl, b[0], b[1], b[3])
			}
		}
		for _, e := range c.Events {
			switch e["ida_linear"] {
			case "1207f":
				turns++
			case "120b0":
				rolls++
			case "121b2":
				moves++
			}
		}
	}
	if turns != 19 || rolls != 110 || moves != 5 {
		t.Fatal("native branch coverage", turns, rolls, moves)
	}
	t.Logf("1672 controlled native components; turns%d step rolls%d moves%d; full NPC/global RNG parity remains unknown", turns, rolls, moves)

	// Reconstruct complete scans from the same observed sequence. Only the first
	// and last observation groups are excluded because their scan boundaries are
	// outside this observer's bounded window. No seed search or reroll is used.
	assets := os.Getenv("DQ3_ASSETS")
	townRaw, err := os.ReadFile(filepath.Join(assets, "CTY00.DAT"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(townRaw)) != "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836" {
		t.Fatal("native town identity", err)
	}
	town, err := dq3data.OpenTown(townRaw, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	attrRaw, err := os.ReadFile(filepath.Join(assets, blkbmFile(mapBlkNum[0])))
	if err != nil {
		t.Fatal(err)
	}
	boundaries := []int{0}
	for i := 1; i < len(source.Cases); i++ {
		a, b := source.Cases[i-1].Entry, source.Cases[i].Entry
		ba, _ := hex.DecodeString(a["npc_slots"])
		bb, _ := hex.DecodeString(b["npc_slots"])
		ia, ib := value(a, "slot", 10)*8, value(b, "slot", 10)*8
		if a["player_x"] != b["player_x"] || a["player_y"] != b["player_y"] || int(bb[ib+1])*town.W+int(bb[ib]) <= int(ba[ia+1])*town.W+int(ba[ia]) {
			boundaries = append(boundaries, i)
		}
	}
	boundaries = append(boundaries, len(source.Cases))
	complete, repeated := 0, 0
	for k := 1; k < len(boundaries)-2; k++ {
		start, end := boundaries[k], boundaries[k+1]
		first, last := source.Cases[start].Entry, source.Cases[end-1].Return
		before, _ := hex.DecodeString(first["npc_slots"])
		after, _ := hex.DecodeString(last["npc_slots"])
		sc := &Scene{w: town.W, h: town.H, hiMap: append([]byte(nil), town.HiMap...), tileAt: town.Tile, attr: dq3data.OpenBlockAttr(attrRaw)}
		for j := 0; j < len(before)/8; j++ {
			b := before[j*8:]
			sc.npcs = append(sc.npcs, npcInst{x: int(b[0]), y: int(b[1]), ctrl: int(b[3])})
		}
		seen := map[string]bool{}
		for j := start; j < end; j++ {
			c := source.Cases[j]
			if j > start && c.Entry["seed"] != source.Cases[j-1].Return["seed"] {
				t.Fatal("uncontrolled RNG inside native scan", j)
			}
			if seen[c.Entry["slot"]] {
				repeated++
			}
			seen[c.Entry["slot"]] = true
			if c.Entry["target_valid"] == "true" {
				x, y := value(c.Entry, "target_x", 10), value(c.Entry, "target_y", 10)
				tile := town.Tile(x, y)
				if tile != value(c.Entry, "target_word", 16)&255 || int(sc.attr.A[tile]) != value(c.Entry, "target_attr", 16) {
					t.Fatal("scan terrain differs from raw town", j)
				}
			}
		}
		sc.npcRng.Seed(uint16(value(first, "seed", 16)))
		g := &Game{cur: sc, inTown: true, pack: p, px: value(first, "player_x", 10), py: value(first, "player_y", 10)}
		g.npcTick()
		if sc.npcRng.State() != uint16(value(last, "seed", 16)) {
			t.Fatalf("native scan%d calls%d RNG got%04x want%s", k, end-start, sc.npcRng.State(), last["seed"])
		}
		for j, n := range sc.npcs {
			b := after[j*8:]
			if n.x != int(b[0]) || n.y != int(b[1]) || n.ctrl != int(b[3]) {
				t.Fatalf("native scan%d actor%d mismatch", k, j)
			}
		}
		complete++
	}
	if complete == 0 || repeated == 0 {
		t.Fatal("missing native live scan coverage", complete, repeated)
	}
	t.Logf("native full cell scans%d; later-cell repeated actor evaluations%d; source clock origin unchanged", complete, repeated)
}
