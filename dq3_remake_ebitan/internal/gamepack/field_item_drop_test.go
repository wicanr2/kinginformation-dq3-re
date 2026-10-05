package gamepack

import (
	"bytes"
	"encoding/binary"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldItemDropOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	exe, raw, items := originalItemStorageInputs(t)
	for _, check := range []struct {
		linear int
		bytes  []byte
	}{
		{0x13af3, []byte{0x3d, 0, 0}}, {0x13af8, []byte{0xc7, 4, 0xff, 0}},
		{0x13afc, []byte{0xbf, 0x15, 1}}, {0x13b09, []byte{0xbf, 0x10, 1}},
	} {
		offset := check.linear - 0xec90
		if !bytes.Equal(exe[offset:offset+len(check.bytes)], check.bytes) {
			t.Fatal("original drop instruction differs", check.linear)
		}
	}
	for code, m := range p.Characters.ItemStorage.Items {
		hasValue := binary.LittleEndian.Uint16(raw[code*7+2:]) != 0
		if m.DropHasValue == nil || *m.DropHasValue != hasValue {
			t.Fatal("actual original price gate differs", code)
		}
		entry := itemstore.Entry{Code: code, Word: uint16(code)}
		if p.ItemDropAllowed(entry) != (hasValue && !*m.DropForbidden) {
			t.Fatal("unworn drop eligibility differs", code)
		}
		entry.Word |= uint16(*p.Characters.ItemStorage.DropBlockedMask)
		if p.ItemDropAllowed(entry) {
			t.Fatal("blocked word accepted", code)
		}
	}
	if p.ItemDropAllowed(itemstore.Entry{Code: -1}) || p.ItemDropAllowed(itemstore.Entry{Code: items.Count()}) {
		t.Fatal("invalid identity accepted")
	}
	if err := p.ValidateItemStorageAgainstItems(items); err != nil {
		t.Fatal(err)
	}
	*p.Characters.ItemStorage.Items[0].DropHasValue = !*p.Characters.ItemStorage.Items[0].DropHasValue
	if p.ValidateItemStorageAgainstItems(items) == nil {
		t.Fatal("incorrect price metadata accepted")
	}
	s := p.Interface.FieldItems.Drop
	for _, check := range []struct {
		id     string
		record int
	}{{s.SuccessTextID, 277}, {s.BlockedTextID, 272}} {
		d, ok := p.TextDefinition(check.id)
		if !ok || d.Source.Record == nil || *d.Source.Record != check.record {
			t.Fatal("native drop record reference differs")
		}
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", d.Source.File))
		if err != nil {
			t.Fatal(err)
		}
		codes := dq3data.LoadText(nil, b).Record(check.record)
		if len(codes) != len(d.GlyphCodes) {
			t.Fatal("drop record shape differs")
		}
		for i, c := range codes {
			if int(c) != d.GlyphCodes[i] {
				t.Fatal("drop glyph source differs")
			}
		}
	}
	if *s.ActorVariableCode != 0xfffb || *s.ItemVariableCode != 0xfff9 {
		t.Fatal("original drop binding differs")
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.ContentVersion(), p.ContentHash())
}

func TestFieldItemDropRejectsBrokenContract(t *testing.T) {
	for _, edit := range []func(*Pack){
		func(p *Pack) { p.Interface.FieldItems.Drop = nil },
		func(p *Pack) { p.Interface.FieldItems.Drop.SuccessTextID = "unknown" },
		func(p *Pack) { p.Interface.FieldItems.Drop.BlockedTextID = "unknown" },
		func(p *Pack) { p.Interface.FieldItems.Drop.BlockedTextID = p.Interface.FieldItems.Drop.SuccessTextID },
		func(p *Pack) { p.Interface.FieldItems.Drop.ActorVariableCode = nil },
		func(p *Pack) { p.Interface.FieldItems.Drop.ItemVariableCode = nil },
		func(p *Pack) {
			*p.Interface.FieldItems.Drop.ItemVariableCode = *p.Interface.FieldItems.Drop.ActorVariableCode
		},
		func(p *Pack) { p.Interface.FieldItems.Drop.Evidence.Level = "D2" },
		func(p *Pack) {
			d, _ := p.TextDefinition(p.Interface.FieldItems.Drop.SuccessTextID)
			d.GlyphCodes = append(d.GlyphCodes, *p.Interface.FieldItems.Drop.ActorVariableCode)
		},
		func(p *Pack) {
			d, _ := p.TextDefinition(p.Interface.FieldItems.Drop.BlockedTextID)
			d.GlyphCodes[0] = *p.Interface.FieldItems.Drop.ItemVariableCode
		},
		func(p *Pack) {
			d, _ := p.TextDefinition(p.Interface.FieldItems.Drop.SuccessTextID)
			d.GlyphCodes = append(d.GlyphCodes, 65532)
		},
	} {
		p, err := BuiltinDQ3()
		if err != nil {
			t.Fatal(err)
		}
		edit(p)
		if p.validateFieldItems() == nil {
			t.Fatal("broken drop contract accepted")
		}
	}
}
