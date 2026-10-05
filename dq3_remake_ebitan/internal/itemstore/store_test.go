package itemstore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// Version-specific values here are evidence fixtures, never engine defaults.
func fixtureContract() (Encoding, []Metadata) {
	e := Encoding{Empty: 0xff, CodeMask: 0xff, WornMask: 0x8000,
		CurseMask: 0x4000, TransferBlockedMask: 0xe000, PartCount: 4}
	metadata := make([]Metadata, 128)
	for code := range metadata {
		metadata[code].Part = -1
	}
	metadata[0].Part = 0
	metadata[3].Part = 0
	metadata[27] = Metadata{Part: 0, CursedWhenWorn: true}
	metadata[30].Part = 1
	return e, metadata
}

func fixture(t *testing.T, words []uint16) Store {
	t.Helper()
	e, metadata := fixtureContract()
	s, err := FromWords(len(words), e, metadata, words)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func requireWords(t *testing.T, s Store, want []uint16) {
	t.Helper()
	if got := s.Words(); !reflect.DeepEqual(got, want) {
		t.Fatalf("physical words: got %#v, want %#v", got, want)
	}
}

func TestAcceptedSelfGiftRetainsEmptyAndWornWord(t *testing.T) {
	// Normal original packets 225..230, receipt aad971bb (docs/188).
	s := fixture(t, []uint16{0x801e, 0, 1, 1, 3, 31, 31, 0xff})
	if !s.Give(&s, 0) {
		t.Fatal("normal self-gift rejected")
	}
	requireWords(t, s, []uint16{0, 1, 1, 3, 31, 31, 0xff, 0x801e})
	entries := s.Entries()
	if len(entries) != 7 || entries[6].Position != 7 || entries[6].Code != 30 ||
		len(s.Inventory()) != 6 || len(s.Worn()) != 1 || s.Worn()[0].Position != 7 {
		t.Fatalf("list must retain physical position, duplicates and equipment: %#v", entries)
	}
	if !s.MoveToEnd(7) {
		t.Fatal("last slot self-gift")
	}
	requireWords(t, s, []uint16{0, 1, 1, 3, 31, 31, 0xff, 0x801e})
}

func TestFirstEmptyItemZeroFullFailureAndRemovalHole(t *testing.T) {
	s := fixture(t, []uint16{0x801e, 0, 0xff, 1, 3, 31, 31, 0xff})
	if position, ok := s.Add(0); !ok || position != 2 {
		t.Fatal("code zero must fill first empty, not count as empty")
	}
	if position, ok := s.Add(31); !ok || position != 7 {
		t.Fatal("second acquisition must fill remaining physical empty")
	}
	before := s.Words()
	if _, ok := s.Add(1); ok {
		t.Fatal("full owner accepted another item")
	}
	requireWords(t, s, before)
	if !s.Remove(3) {
		t.Fatal("remove selected slot")
	}
	want := append([]uint16(nil), before...)
	want[3] = 0xff
	requireWords(t, s, want)
	if position, ok := s.Add(3); !ok || position != 3 {
		t.Fatal("acquisition must reuse the physical hole")
	}
}

func TestCrossOwnerTransactionAndBlockedWords(t *testing.T) {
	source := fixture(t, []uint16{0x801e, 0x1001, 0xff, 3})
	destination := fixture(t, []uint16{1, 0xff, 0x801e, 0xff})
	if !source.Give(&destination, 1) {
		t.Fatal("permitted full word transfer")
	}
	requireWords(t, source, []uint16{0x801e, 0xff, 0xff, 3})
	requireWords(t, destination, []uint16{1, 0x1001, 0x801e, 0xff})
	for _, word := range []uint16{0x8000, 0x4000, 0x2000, 0xe000} {
		t.Run(fmt.Sprintf("blocked-%04x", word), func(t *testing.T) {
			source := fixture(t, []uint16{word, 0xff})
			destination := fixture(t, []uint16{0xff, 3})
			sourceBefore, destinationBefore := source.Words(), destination.Words()
			if source.Give(&destination, 0) {
				t.Fatal("original 13A08 transfer gate must reject flagged word")
			}
			requireWords(t, source, sourceBefore)
			requireWords(t, destination, destinationBefore)
			if !source.Give(&source, 0) {
				t.Fatal("cross-owner mask must not prevent self-gift")
			}
			requireWords(t, source, []uint16{0xff, word})
		})
	}
	full := fixture(t, []uint16{0, 1})
	sourceBefore, fullBefore := source.Words(), full.Words()
	if source.Give(&full, 3) {
		t.Fatal("full recipient must reject transaction")
	}
	requireWords(t, source, sourceBefore)
	requireWords(t, full, fullBefore)
	otherContract := fixture(t, []uint16{0xff, 0xff})
	otherContract.encoding.TransferBlockedMask = 0x8000
	if source.Give(&otherContract, 3) {
		t.Fatal("cross-pack contract must reject transfer")
	}
	requireWords(t, source, sourceBefore)
	requireWords(t, otherContract, []uint16{0xff, 0xff})
}

func TestWearStaysInPlaceAndPreservesOpaqueFlags(t *testing.T) {
	s := fixture(t, []uint16{0x9000, 0xff, 3, 0x801e, 0x2001})
	if !s.Wear(2) {
		t.Fatal("replace permitted weapon")
	}
	requireWords(t, s, []uint16{0x1000, 0xff, 0x8003, 0x801e, 0x2001})
	if len(s.Worn()) != 2 || s.Worn()[0].Position != 2 || s.Worn()[1].Position != 3 {
		t.Fatal("equipment must derive from physical positions")
	}
	if s.Wear(4) {
		t.Fatal("non-gear item must not be assigned a guessed equipment part")
	}
	requireWords(t, s, []uint16{0x1000, 0xff, 0x8003, 0x801e, 0x2001})
	// An unworn curse flag is preserved, without interpreting unknown flags.
	s = fixture(t, []uint16{0x4000, 3, 0x801e})
	if !s.Wear(1) {
		t.Fatal("unworn cursed word is not the current locked weapon")
	}
	requireWords(t, s, []uint16{0x4000, 0x8003, 0x801e})
}

func TestCurseWriterReplacementGateAndChurchRemoval(t *testing.T) {
	s := fixture(t, []uint16{0x8000, 27, 3, 0x401f, 0x801e})
	if !s.Wear(1) {
		t.Fatal("equip cursed metadata candidate")
	}
	requireWords(t, s, []uint16{0, 0xc01b, 3, 0x401f, 0x801e})
	before := s.Words()
	if s.Wear(2) {
		t.Fatal("cannot replace currently cursed weapon")
	}
	requireWords(t, s, before)
	if removed := s.RemoveCursed(); removed != 2 {
		t.Fatalf("church must remove worn and unworn cursed words: %d", removed)
	}
	requireWords(t, s, []uint16{0, 0xff, 3, 0xff, 0x801e})
	if removed := s.RemoveCursed(); removed != 0 {
		t.Fatal("repeated church removal must not report another transaction")
	}
}

func TestNoMutableViewInputMetadataOrStoreCopyAliases(t *testing.T) {
	e, metadata := fixtureContract()
	words := []uint16{0x801e, 0, 0xff}
	s, err := FromWords(len(words), e, metadata, words)
	if err != nil {
		t.Fatal(err)
	}
	words[0] = 0xff
	metadata[0].Part = -1
	s.Words()[0] = 0xff
	s.Entries()[0].Word = 0xff
	s.Inventory()[0].Word = 0xff
	s.Worn()[0].Word = 0xff
	requireWords(t, s, []uint16{0x801e, 0, 0xff})
	copyValue, clone := s, s.Clone()
	if !copyValue.Wear(1) || !clone.Remove(0) {
		t.Fatal("independent Store copies")
	}
	requireWords(t, s, []uint16{0x801e, 0, 0xff})
	requireWords(t, copyValue, []uint16{0x801e, 0x8000, 0xff})
	requireWords(t, clone, []uint16{0xff, 0, 0xff})
	if !s.Remove(1) {
		t.Fatal("original owner mutation")
	}
	requireWords(t, copyValue, []uint16{0x801e, 0x8000, 0xff})
}

func TestSnapshotRoundTripPreservesAllPhysicalWords(t *testing.T) {
	s := fixture(t, []uint16{0x203f, 0xff, 0xc01b, 0, 0xff})
	raw, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	e, metadata := fixtureContract()
	loaded, err := Decode(5, e, metadata, raw)
	if err != nil {
		t.Fatal(err)
	}
	requireWords(t, loaded, s.Words())
	if !loaded.MoveToEnd(2) {
		t.Fatal("loaded owner must support next operation")
	}
	requireWords(t, loaded, []uint16{0x203f, 0xff, 0, 0xff, 0xc01b})
	requireWords(t, s, []uint16{0x203f, 0xff, 0xc01b, 0, 0xff})
}

func TestOrdinaryJSONCannotBypassExplicitArchiveContract(t *testing.T) {
	s := fixture(t, []uint16{0x801e, 0, 0xff})
	raw, err := json.Marshal(s)
	if err != nil || string(raw) != `{"storage_version":1,"words":[32798,0,255]}` {
		t.Fatalf("ordinary JSON marshal lost private owner: %s, %v", raw, err)
	}
	before := s.Words()
	if err := json.Unmarshal(raw, &s); err == nil {
		t.Fatal("ordinary unmarshal bypassed explicit archive/pack validation")
	}
	requireWords(t, s, before)
	var zero Store
	if _, err := json.Marshal(zero); err == nil {
		t.Fatal("uninitialized private fields serialized without an error")
	}
	if err := json.Unmarshal(raw, &zero); err == nil || zero.valid() {
		t.Fatal("ordinary JSON supplied a default store contract")
	}
}

func TestInvalidContractsIndicesAndUninitializedStoresReject(t *testing.T) {
	e, metadata := fixtureContract()
	mutations := []func(*Encoding){
		func(e *Encoding) { *e = Encoding{} },
		func(e *Encoding) { e.Empty = 0 },
		func(e *Encoding) { e.CodeMask = 0xf5 },
		func(e *Encoding) { e.WornMask = e.CodeMask },
		func(e *Encoding) { e.CurseMask = e.WornMask },
		func(e *Encoding) { e.TransferBlockedMask = 0 },
		func(e *Encoding) { e.TransferBlockedMask |= 1 },
		func(e *Encoding) { e.PartCount = 0 },
	}
	for index, mutation := range mutations {
		bad := e
		mutation(&bad)
		if _, err := FromWords(1, bad, metadata, []uint16{0xff}); err == nil {
			t.Fatalf("bad contract %d accepted", index)
		}
	}
	for _, words := range [][]uint16{nil, {0}, {0x80, 0xff}} {
		if _, err := FromWords(2, e, metadata, words); err == nil {
			t.Fatalf("invalid capacity or identity accepted: %#v", words)
		}
	}
	badMetadata := append([]Metadata(nil), metadata...)
	badMetadata[0].Part = e.PartCount
	if _, err := FromWords(1, e, badMetadata, []uint16{0}); err == nil {
		t.Fatal("out-of-range metadata part")
	}
	s := fixture(t, []uint16{0x801e, 0, 0xff})
	for _, position := range []int{-1, 2, 3} {
		before := s.Words()
		if s.Remove(position) || s.MoveToEnd(position) || s.Wear(position) || s.Give(&s, position) {
			t.Fatalf("invalid/empty physical position accepted: %d", position)
		}
		requireWords(t, s, before)
	}
	var zero Store
	if _, err := zero.Encode(); err == nil {
		t.Fatal("zero store encoded")
	}
	if _, ok := zero.Add(0); ok || zero.Remove(0) || zero.Wear(0) || zero.Give(&s, 0) || zero.RemoveCursed() != 0 {
		t.Fatal("zero store supplied a default contract")
	}
	var absent *Store
	if _, ok := absent.Add(0); ok || absent.Remove(0) || absent.Wear(0) || absent.Give(&s, 0) || absent.MoveToEnd(0) || absent.RemoveCursed() != 0 {
		t.Fatal("absent owner mutated")
	}
}

func TestMalformedLegacyAndAmbiguousSnapshotsReject(t *testing.T) {
	e, metadata := fixtureContract()
	for _, raw := range []string{
		`null`, `[]`, `{}`, `{"words":[0]}`, `{"storage_version":1}`,
		`{"storage_version":null,"words":[0]}`, `{"storage_version":1,"words":null}`,
		`{"storage_version":0,"words":[0]}`, `{"storage_version":2,"words":[0]}`,
		`{"storage_version":1.5,"words":[0]}`, `{"storage_version":"1","words":[0]}`,
		`{"storage_version":1,"words":[null]}`, `{"storage_version":1,"words":[-1]}`,
		`{"storage_version":1,"words":[65536]}`, `{"storage_version":1,"words":[1.5]}`,
		`{"storage_version":1,"words":["0"]}`, `{"storage_version":1,"words":[128]}`,
		`{"storage_version":1,"words":[0,255]}`, `{"storage_version":1,"words":[]}`,
		`{"inventory":[0],"equipment":[-1,30,-1,-1]}`,
		`{"storage_version":1,"words":[0],"inventory":[0]}`,
		`{"storage_version":1,"words":[0],"Words":[1]}`,
		`{"storage_version":1,"words":[0],"words":[1]}`,
		`{"storage_version":0,"storage_version":1,"words":[0]}`,
		`{"storage_version":1,"words":[0]} {}`, `{"storage_version":1,"words":[0]}x`,
		`{"storage_version":1,"words":[0]`,
	} {
		if _, err := Decode(1, e, metadata, []byte(raw)); err == nil {
			t.Errorf("malformed snapshot accepted: %s", raw)
		}
	}
}

func TestDerivedViewsAndExactPhysicalReplacement(t *testing.T) {
	s := fixture(t, []uint16{0x801e, 0xff, 0x2000, 0, 3, 0xff})
	if s.Equipment() != [4]int{-1, 30, -1, -1} || s.Count(0) != 2 {
		t.Fatal("derived equipment or code-zero count differs")
	}
	copyOwner := s
	if !s.Replace(2, 27) || s.Replace(1, 3) || s.Replace(3, 128) {
		t.Fatal("replacement must select an occupied slot and valid identity")
	}
	if s.RemoveCode(0, 2) != 1 || s.RemoveCode(30, 0) != 0 {
		t.Fatal("removal must count actual physical matches")
	}
	requireWords(t, s, []uint16{0x801e, 0xff, 27, 0xff, 3, 0xff})
	requireWords(t, copyOwner, []uint16{0x801e, 0xff, 0x2000, 0, 3, 0xff})
	if !reflect.DeepEqual(s.Codes(), []int{30, 27, 3}) {
		t.Fatal("code view omitted worn word or changed physical order")
	}
	var absent *Store
	if absent.RemoveCode(0, 1) != 0 || absent.Replace(0, 0) || absent.ClearFirstCursedState() || absent.ClearStatesAndRemove(0) {
		t.Fatal("absent owner mutated")
	}
}

func TestFirstCurseClearsAllStateAndAdvancedRemovesEveryBook(t *testing.T) {
	s := fixture(t, []uint16{0x601f, 0xc01b, 0xff, 0xa003, 0x404a, 0, 0x204a, 0xff})
	copyOwner := s
	if !s.HasCursed() || !s.ClearFirstCursedState() {
		t.Fatal("unworn curse was ignored")
	}
	requireWords(t, s, []uint16{31, 0xc01b, 0xff, 0xa003, 0x404a, 0, 0x204a, 0xff})
	if !s.ClearStatesAndRemove(74) {
		t.Fatal("reviewed advanced transition rejected")
	}
	requireWords(t, s, []uint16{31, 27, 0xff, 3, 0xff, 0, 0xff, 0xff})
	requireWords(t, copyOwner, []uint16{0x601f, 0xc01b, 0xff, 0xa003, 0x404a, 0, 0x204a, 0xff})
	before := s.Words()
	if s.HasCursed() || s.ClearFirstCursedState() || s.ClearStatesAndRemove(128) {
		t.Fatal("absent curse or invalid book accepted")
	}
	requireWords(t, s, before)
}
