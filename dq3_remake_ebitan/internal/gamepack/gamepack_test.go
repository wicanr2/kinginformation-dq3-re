package gamepack

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuiltinDQ3ReviveCosts(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	for level, want := range map[int]int{0: 10, 1: 10, 3: 10, 4: 20, 40: 1610, 41: 1610} {
		if got := p.ReviveCost(level); got != want {
			t.Errorf("ReviveCost(%d)=%d, want %d", level, got, want)
		}
	}
}

func TestDQ3ReviveCostsMatchOriginalEXE(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	const off = 0x19dac
	if len(exe) < off+40*2 {
		t.Fatalf("DQ3.EXE too small for revive table: %d", len(exe))
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	for level := 1; level <= 40; level++ {
		want := int(binary.LittleEndian.Uint16(exe[off+(level-1)*2:]))
		if got := p.ReviveCost(level); got != want {
			t.Errorf("level %d JSON=%d, DQ3.EXE file %#x=%d", level, got, off+(level-1)*2, want)
		}
	}
}

func TestLoadRejectsUnknownAndInvalidData(t *testing.T) {
	validManifest := `{
	  "schema_version":"0.1.0","pack_id":"test","game":"dq3","edition":"cht_jingxun",
	  "content_version":"0.1.0","engine_api":">=0.1.0 <0.2.0",
	  "title_text_id":"x:title","entry_event_id":"x:new","save_namespace":"test",
	  "capabilities":[],"data":{"facilities":"facilities.json"},"assets":{}
	}`
	validFacilities := `{
	  "schema_version":"0.1.0","service_definitions":[{
	    "id":"common:service.revive",
	    "pricing":{"formula_id":"common:formula.level_table","level_cap":1,"costs_gold":[10]},
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"reader","doc":"docs/x.md"}
	  }]
	}`
	tests := []struct {
		name, manifest, facilities, want string
	}{
		{"unknown manifest field", strings.Replace(validManifest, `"assets":{}`, `"assets":{},"typo":1`, 1), validFacilities, "unknown field"},
		{"path escape", strings.Replace(validManifest, `"facilities.json"`, `"../facilities.json"`, 1), validFacilities, "pack-relative"},
		{"cost length", validManifest, strings.Replace(validFacilities, `"level_cap":1`, `"level_cap":2`, 1), "must equal"},
		{"unknown facilities field", validManifest, strings.Replace(validFacilities, `"schema_version":"0.1.0"`, `"schema_version":"0.1.0","typo":1`, 1), "unknown field"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"manifest.json":   {Data: []byte(tc.manifest)},
				"facilities.json": {Data: []byte(tc.facilities)},
			}
			_, err := Load(fsys)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load error=%v, want substring %q", err, tc.want)
			}
		})
	}
}
