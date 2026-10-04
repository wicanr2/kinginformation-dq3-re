package itemstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

const originalEXEHash = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
const originalItemHash = "7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d"

func verifiedFile(t *testing.T, path, wantHash string, wantSize int) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != wantHash || (wantSize >= 0 && len(raw) != wantSize) {
		t.Fatalf("oracle identity differs: %s, size %d, sha256 %s", path, len(raw), got)
	}
	return raw
}

func originalContract(t *testing.T) (Encoding, []Metadata) {
	t.Helper()
	assets := os.Getenv("DQ3_ASSETS")
	if assets == "" {
		assets = "../../../assets_raw"
	}
	exe := verifiedFile(t, filepath.Join(assets, "DQ3.EXE"), originalEXEHash, 115282)
	item := verifiedFile(t, filepath.Join(assets, "ITEM.DAT"), originalItemHash, 896)
	// IDA Pro 9.4 exports in docs/147 and docs/188. These are ORIGINAL file
	// offsets, with original bytes, not rebased IDA memory bytes.
	for _, check := range []struct {
		file int
		word string
	}{
		{0x4c99, "3dff00"},               // IDA linear13929: empty sentinel.
		{0x4cab, "81269125ff00"},         // linear1393B: low-byte identity reader.
		{0x4d78, "a900e0"},               // linear13A08: cross-owner flag gate.
		{0x4d82, "c705ff00"},             // linear13A12: clear source in place.
		{0x85c4, "83c63ab90800"},         // linear17254: eight physical words.
		{0x85ca, "f70400407404c704ff00"}, // linear1725A: church curse removal.
		{0x9408, "b80080"},               // linear18098: worn bit.
		{0x940d, "ba000e"},               // linear1809D: metadata curse mask.
		{0x9416, "b80040"},               // linear180A6: curse bit.
		{0x941e, "b8ff7f2104"},           // linear180AE: clear only worn bit.
	} {
		want, err := hex.DecodeString(check.word)
		if err != nil {
			t.Fatal(err)
		}
		if got := exe[check.file : check.file+len(want)]; !bytes.Equal(got, want) {
			t.Fatalf("original file %x: got %x, want %x", check.file, got, want)
		}
	}
	items, err := dq3data.OpenItems(item)
	if err != nil {
		t.Fatal(err)
	}
	metadata := make([]Metadata, len(item)/7)
	for code := range metadata {
		part := -1
		switch item[code*7+4] >> 5 {
		case 1:
			part = 0
		case 2:
			part = 1
		case 3:
			part = 3
		case 4:
			part = 2
		}
		cursed := binary.LittleEndian.Uint16(item[code*7+4:code*7+6])&0x0e00 != 0
		if items.EquipSlot(code) != part || items.CursedWhenEquipped(code) != cursed {
			t.Fatalf("original decoder disagrees with ITEM record %d", code)
		}
		metadata[code] = Metadata{Part: items.EquipSlot(code), CursedWhenWorn: items.CursedWhenEquipped(code)}
	}
	e, _ := fixtureContract()
	return e, metadata
}

func TestOriginalWordContractAndActualItemMetadata(t *testing.T) {
	e, metadata := originalContract(t)
	s, err := FromWords(8, e, metadata, []uint16{0x801e, 0, 3, 27, 0xff, 0xff, 0xff, 0xff})
	if err != nil {
		t.Fatal(err)
	}
	if !s.Wear(1) || !s.Wear(2) {
		t.Fatal("actual original metadata must permit weapon replacement")
	}
	requireWords(t, s, []uint16{0x801e, 0, 0x8003, 27, 0xff, 0xff, 0xff, 0xff})
	if !s.Wear(3) || s.Wear(2) {
		t.Fatal("actual original curse writer and replacement gate")
	}
	requireWords(t, s, []uint16{0x801e, 0, 3, 0xc01b, 0xff, 0xff, 0xff, 0xff})
}

func TestAcceptedDosgolemPhysicalWordsBeforeAndAfterSelfGift(t *testing.T) {
	directory := os.Getenv("DQ3_ITEM_ORACLE_DIR")
	if directory == "" {
		t.Skip("optional accepted original source; enable DQ3_ITEM_ORACLE_DIR for parity validation")
	}
	const sourceHash = "aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4"
	raw := verifiedFile(t, filepath.Join(directory, "issue4-field-item-reorder-r2-source-r3-receipt.json"), sourceHash, -1)
	var source struct {
		OriginalHash string `json:"original_sha256"`
		Revision     string `json:"upstream_revision"`
		Seed         string `json:"seed"`
		Artifacts    []struct {
			Path   string `json:"path"`
			Size   int    `json:"size"`
			SHA256 string `json:"sha256"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	if source.OriginalHash != originalEXEHash || source.Revision != "2f44a68ebfc54b28fb15dd4a34510b0b04a5415d" || source.Seed != "1357" || len(source.Artifacts) != 498 {
		t.Fatal("accepted source identity or seed differs")
	}
	readPersistent := func(packet int) []byte {
		t.Helper()
		name := fmt.Sprintf("issue4-field-item-reorder-r2-persistent-%03d.bin", packet)
		for _, artifact := range source.Artifacts {
			if artifact.Path == name {
				if artifact.Size != 2172 {
					t.Fatal("original persistent region shape differs")
				}
				return verifiedFile(t, filepath.Join(directory, name), artifact.SHA256, artifact.Size)
			}
		}
		t.Fatalf("accepted source lacks packet %d", packet)
		return nil
	}
	before, after, reopened := readPersistent(224), readPersistent(225), readPersistent(230)
	// DGROUP50B9..50C8; persist begins DGROUP4F29. Address spaces stay explicit.
	const offset = 0x50b9 - 0x4f29
	words := make([]uint16, 8)
	for position := range words {
		words[position] = binary.LittleEndian.Uint16(before[offset+position*2:])
	}
	e, metadata := originalContract(t)
	s, err := FromWords(len(words), e, metadata, words)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Give(&s, 0) {
		t.Fatal("accepted original self-gift")
	}
	want := append([]byte(nil), before...)
	for position, word := range s.Words() {
		binary.LittleEndian.PutUint16(want[offset+position*2:], word)
	}
	if !bytes.Equal(want, after) || !bytes.Equal(after, reopened) {
		t.Fatal("core differs from accepted complete original 2172-byte persistent transaction")
	}
	encoded, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Decode(8, e, metadata, encoded)
	if err != nil {
		t.Fatal(err)
	}
	requireWords(t, loaded, s.Words())
	if len(loaded.Worn()) != 1 || loaded.Worn()[0].Position != 7 || loaded.Entries()[6].Position != 7 {
		t.Fatal("snapshot cannot rebuild a compact or equipment-first sequence")
	}
}
