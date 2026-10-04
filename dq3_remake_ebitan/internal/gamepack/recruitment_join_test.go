package gamepack

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/opl2"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecruitmentJoinOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.RecruitmentJoin
	dir := filepath.Join("..", "..", "..", "assets_raw")
	read := func(file, hash string) []byte {
		t.Helper()
		d, e := os.ReadFile(filepath.Join(dir, file))
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(d)) != hash {
			t.Fatal("original identity differs", file)
		}
		return d
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	for i, id := range s.TextIDs() {
		d, _ := p.TextDefinition(id)
		c, _ := p.TextGlyphCodes(id)
		if *d.Source.Record != 536+i || !reflect.DeepEqual(c, tx.Record(536+i)) {
			t.Fatal("join record differs", id)
		}
	}
	for _, x := range []struct {
		offset int
		raw    string
	}{{0x1792, "bf1802"}, {0x179d, "bf1902"}, {0x17b4, "bd2000"}, {0x17bc, "bf1a02"}, {0x17ce, "eb09"}, {0x17d9, "c3"}, {0x171e, "bf1c02"}} {
		if fmt.Sprintf("%x", exe[x.offset:x.offset+len(x.raw)/2]) != x.raw {
			t.Fatalf("caller file%x differs", x.offset)
		}
	}
	a := s.Sound
	asset := p.Manifest.Assets[a.SourceAsset]
	raw := read(asset.Path, asset.SHA256)
	events, ticks, e := opl2.ParseEventStream(raw[a.Start:a.End])
	if e != nil || len(events) != a.EventCount || ticks != uint64(a.DeltaTicks) {
		t.Fatal("sound event stream differs", e)
	}
	if a.Start != 0x58b || a.End != 0x6b3 || len(events) != 81 || ticks != 461 || a.ClockDivisor != 12428 || a.ReferenceHz != 1193180 || a.HoldFrames() != 289 {
		t.Fatal("reviewed stream/timing differs")
	}
	if s.LeaderNameRole != "primary_actor" || s.NameControlCode != dq3data.TxtVarEnt || !s.RetainCallerBackdrop {
		t.Fatal("reviewed names/backdrop differ")
	}
	pcm, e := opl2.RenderEventStream(raw[a.Start:a.End], 44100, a.ClockDivisor, a.ReferenceHz)
	wantFrames := int((a.DeltaTicks*a.ClockDivisor*44100 + a.ReferenceHz - 1) / a.ReferenceHz)
	if e != nil || len(pcm) != wantFrames {
		t.Fatal("single FM duration", len(pcm), wantFrames, e)
	}
	nonzero := false
	for _, sample := range pcm {
		if sample != 0 {
			nonzero = true
			break
		}
	}
	if !nonzero {
		t.Fatal("single FM render is silent")
	}
}

func TestRecruitmentJoinRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "null_sound", "missing_zero", "unknown_sound", "unknown_text", "same_text", "wrong_name_role", "wrong_code", "backdrop", "unknown_asset", "bad_range", "bad_clock", "bad_hash", "bad_path", "unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			b, _ := json.Marshal(p.Interface.RecruitmentJoin)
			var m map[string]any
			json.Unmarshal(b, &m)
			sound := m["sound"].(map[string]any)
			switch name {
			case "missing":
				p.Interface.RecruitmentJoin = nil
			case "null_sound":
				m["sound"] = nil
			case "missing_zero":
				delete(sound, "start")
			case "unknown_sound":
				sound["guessed"] = true
			case "unknown_text":
				m["joined_text_id"] = "unknown"
			case "same_text":
				m["leader_text_id"] = m["joined_text_id"]
			case "wrong_name_role":
				m["leader_name_role"] = "joined_actor"
			case "wrong_code":
				m["name_control_code"] = 65530
			case "backdrop":
				m["retain_caller_backdrop"] = false
			case "unknown_asset":
				sound["source_asset"] = "unknown"
			case "bad_range":
				sound["end"] = 999999
			case "bad_clock":
				sound["reference_hz"] = 0
			case "bad_hash":
				sound["render_sha256"] = "x"
			case "bad_path":
				sound["render_file"] = "../wrong.ogg"
			case "unreviewed":
				sound["timing_evidence"].(map[string]any)["level"] = "D1"
			}
			if name != "missing" {
				b, _ = json.Marshal(m)
				var s RecruitmentJoin
				if e = json.Unmarshal(b, &s); e != nil {
					return
				}
				p.Interface.RecruitmentJoin = &s
			}
			if e = p.validateRecruitmentJoin(); e == nil {
				t.Fatal("broken join accepted")
			}
		})
	}
}
