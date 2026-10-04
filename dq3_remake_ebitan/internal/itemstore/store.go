// Package itemstore owns ordered physical item words. Contract and evidence:
// docs/188, "A 共用格位核心 READY"; pack integration remains a separate gate.
package itemstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Encoding is supplied by a validated game pack, without version defaults.
type Encoding struct {
	Empty, CodeMask, WornMask, CurseMask, TransferBlockedMask uint16
	PartCount                                                 int
}

// Metadata is indexed by the actual item archive. Part -1 means no gear part.
// Eligibility (class, gender, conditions) remains the caller's responsibility.
type Metadata struct {
	Part           int
	CursedWhenWorn bool
}

// Entry retains the physical identity even when duplicate codes are displayed.
type Entry struct {
	Position, Code, Part int
	Word                 uint16
}

// Store has one writable source of ownership. Mutations copy their backing
// words, so copying a Store value cannot mutate a different owner by alias.
type Store struct {
	words    []uint16
	encoding Encoding
	metadata []Metadata
}

func FromWords(capacity int, e Encoding, metadata []Metadata, words []uint16) (Store, error) {
	if capacity < 1 || len(words) != capacity || e.CodeMask == 0 ||
		e.CodeMask&(e.CodeMask+1) != 0 || e.WornMask == 0 || e.CurseMask == 0 ||
		e.CodeMask&(e.WornMask|e.CurseMask) != 0 || e.WornMask&e.CurseMask != 0 ||
		e.Empty&(e.WornMask|e.CurseMask) != 0 || e.TransferBlockedMask == 0 ||
		e.TransferBlockedMask&e.CodeMask != 0 || e.PartCount < 1 ||
		len(metadata) < 1 || len(metadata) > int(e.CodeMask)+1 ||
		int(e.Empty&e.CodeMask) < len(metadata) {
		return Store{}, fmt.Errorf("missing or invalid item word contract")
	}
	for code, item := range metadata {
		if item.Part < -1 || item.Part >= e.PartCount {
			return Store{}, fmt.Errorf("invalid equipment part for item %d", code)
		}
	}
	for position, word := range words {
		if word != e.Empty && int(word&e.CodeMask) >= len(metadata) {
			return Store{}, fmt.Errorf("invalid item identity at physical slot %d", position)
		}
	}
	return Store{append([]uint16(nil), words...), e, append([]Metadata(nil), metadata...)}, nil
}

func (s Store) valid() bool { return len(s.words) > 0 && len(s.metadata) > 0 }

func (s Store) Words() []uint16 { return append([]uint16(nil), s.words...) }

func (s Store) Clone() Store {
	return Store{s.Words(), s.encoding, append([]Metadata(nil), s.metadata...)}
}

func (s Store) At(position int) (Entry, bool) {
	if position < 0 || position >= len(s.words) || s.words[position] == s.encoding.Empty {
		return Entry{}, false
	}
	word := s.words[position]
	code := int(word & s.encoding.CodeMask)
	return Entry{position, code, s.metadata[code].Part, word}, true
}

func (s Store) Entries() []Entry {
	var entries []Entry
	for position := range s.words {
		if entry, ok := s.At(position); ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (s Store) Inventory() []Entry {
	var entries []Entry
	for _, entry := range s.Entries() {
		if entry.Word&s.encoding.WornMask == 0 {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (s Store) Worn() []Entry {
	var entries []Entry
	for _, entry := range s.Entries() {
		if entry.Word&s.encoding.WornMask != 0 {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (s Store) firstEmpty() int {
	for position, word := range s.words {
		if word == s.encoding.Empty {
			return position
		}
	}
	return -1
}

// Add writes a plain item into the first physical empty slot. Price/reward
// transactions must commit only after this operation succeeds.
func (s *Store) Add(code int) (int, bool) {
	if s == nil || !s.valid() || code < 0 || code >= len(s.metadata) {
		return -1, false
	}
	position := s.firstEmpty()
	if position < 0 {
		return -1, false
	}
	words := s.Words()
	words[position] = uint16(code)
	s.words = words
	return position, true
}

// Remove clears only the selected physical slot. The caller must first apply
// the selected action's eligibility/confirmation/consumption gate.
func (s *Store) Remove(position int) bool {
	if s == nil {
		return false
	}
	if _, ok := s.At(position); !ok {
		return false
	}
	words := s.Words()
	words[position] = s.encoding.Empty
	s.words = words
	return true
}

func (s *Store) MoveToEnd(position int) bool {
	if s == nil {
		return false
	}
	entry, ok := s.At(position)
	if !ok {
		return false
	}
	words := s.Words()
	copy(words[position:], words[position+1:])
	words[len(words)-1] = entry.Word
	s.words = words
	return true
}

func (s Store) compatible(other Store) bool {
	if s.encoding != other.encoding || len(s.metadata) != len(other.metadata) {
		return false
	}
	for code := range s.metadata {
		if s.metadata[code] != other.metadata[code] {
			return false
		}
	}
	return true
}

// Give preserves the complete selected word. A self-gift rotates every later
// physical slot, including empties. A cross-owner gift validates both owners
// and all gates before either owner changes.
func (s *Store) Give(destination *Store, position int) bool {
	if s == nil || destination == nil || !s.valid() || !destination.valid() {
		return false
	}
	if s == destination {
		return s.MoveToEnd(position)
	}
	entry, ok := s.At(position)
	if !ok || !s.compatible(*destination) {
		return false
	}
	empty := destination.firstEmpty()
	if empty < 0 || entry.Word&s.encoding.TransferBlockedMask != 0 {
		return false
	}
	sourceWords, destinationWords := s.Words(), destination.Words()
	sourceWords[position] = s.encoding.Empty
	destinationWords[empty] = entry.Word
	s.words, destination.words = sourceWords, destinationWords
	return true
}

// Wear changes flags in place. It does not apply class/gender eligibility.
// A cursed previously worn item blocks replacing that equipment part.
func (s *Store) Wear(position int) bool {
	if s == nil {
		return false
	}
	selected, ok := s.At(position)
	if !ok || selected.Part < 0 {
		return false
	}
	for _, entry := range s.Worn() {
		if entry.Position != position && entry.Part == selected.Part && entry.Word&s.encoding.CurseMask != 0 {
			return false
		}
	}
	words := s.Words()
	for _, entry := range s.Entries() {
		if entry.Position != position && entry.Part == selected.Part {
			words[entry.Position] &^= s.encoding.WornMask
		}
	}
	words[position] |= s.encoding.WornMask
	if s.metadata[selected.Code].CursedWhenWorn {
		words[position] |= s.encoding.CurseMask
	}
	s.words = words
	return true
}

// RemoveCursed removes all matching words, including unworn words. Church
// confirmation, fee and one-time payment remain an external transaction.
func (s *Store) RemoveCursed() int {
	if s == nil || !s.valid() {
		return 0
	}
	words := s.Words()
	removed := 0
	for position, word := range words {
		if word != s.encoding.Empty && word&s.encoding.CurseMask != 0 {
			words[position] = s.encoding.Empty
			removed++
		}
	}
	if removed > 0 {
		s.words = words
	}
	return removed
}

const storageVersion = 1 // Generic serialization version, not game data.

// MarshalJSON prevents accidental serialization as an empty object merely
// because the owning fields are private.
func (s Store) MarshalJSON() ([]byte, error) { return s.Encode() }

// A JSON decoder cannot supply the required external archive/pack contract.
// Require Decode explicitly rather than constructing a default owner.
func (s *Store) UnmarshalJSON(_ []byte) error {
	return fmt.Errorf("item word decoding requires an explicit archive and encoding contract")
}

func (s Store) Encode() ([]byte, error) {
	if !s.valid() {
		return nil, fmt.Errorf("uninitialized item word store")
	}
	return json.Marshal(struct {
		Version int      `json:"storage_version"`
		Words   []uint16 `json:"words"`
	}{storageVersion, s.Words()})
}

// Decode refuses legacy/ambiguous records. It never migrates bag/equipment or
// supplies absent slots. The caller still validates the enclosing game save.
func Decode(capacity int, e Encoding, metadata []Metadata, raw []byte) (Store, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return Store{}, fmt.Errorf("item word snapshot must be an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return Store{}, err
		}
		key, ok := token.(string)
		if !ok || (key != "storage_version" && key != "words") {
			return Store{}, fmt.Errorf("unknown item word snapshot field")
		}
		if _, exists := fields[key]; exists {
			return Store{}, fmt.Errorf("duplicate item word snapshot field %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return Store{}, err
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Store{}, fmt.Errorf("null item word snapshot field %q", key)
		}
		fields[key] = value
	}
	if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
		return Store{}, fmt.Errorf("unfinished item word snapshot object")
	}
	var version int
	var encodedWords []json.RawMessage
	if len(fields) != 2 {
		return Store{}, fmt.Errorf("missing item word snapshot fields")
	}
	if err := json.Unmarshal(fields["storage_version"], &version); err != nil {
		return Store{}, err
	}
	if err := json.Unmarshal(fields["words"], &encodedWords); err != nil {
		return Store{}, err
	}
	if version != storageVersion || len(encodedWords) != capacity || capacity < 1 {
		return Store{}, fmt.Errorf("missing or unsupported item word snapshot")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Store{}, fmt.Errorf("trailing item word snapshot data")
	}
	words := make([]uint16, capacity)
	for position, word := range encodedWords {
		if bytes.Equal(bytes.TrimSpace(word), []byte("null")) {
			return Store{}, fmt.Errorf("null item word at physical slot %d", position)
		}
		if err := json.Unmarshal(word, &words[position]); err != nil {
			return Store{}, fmt.Errorf("invalid item word at physical slot %d: %w", position, err)
		}
	}
	return FromWords(capacity, e, metadata, words)
}
