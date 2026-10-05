package gamepack

import (
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
)

// ItemWordEncoding retains required zero/nonzero fields through strict decoding.
// No raw masks or version-specific metadata are supplied by the engine.
type ItemWordEncoding struct {
	Empty               *int `json:"empty"`
	CodeMask            *int `json:"code_mask"`
	WornMask            *int `json:"worn_mask"`
	CurseMask           *int `json:"curse_mask"`
	TransferBlockedMask *int `json:"transfer_blocked_mask"`
}

type ItemWordMetadata struct {
	EquipmentPart  *int  `json:"equipment_part"`
	CursedWhenWorn *bool `json:"cursed_when_worn"`
	DropForbidden  *bool `json:"drop_forbidden"`
}

type ItemStorageDefinition struct {
	DropBlockedMask *int               `json:"drop_blocked_mask"`
	Encoding        ItemWordEncoding   `json:"encoding"`
	PartCount       *int               `json:"part_count"`
	Items           []ItemWordMetadata `json:"items"`
	Evidence        Evidence           `json:"evidence"`
}

func (p *Pack) itemWordContract() (itemstore.Encoding, []itemstore.Metadata, error) {
	if p == nil || p.Characters.ItemStorage == nil {
		return itemstore.Encoding{}, nil, fmt.Errorf("characters.item_storage is required")
	}
	s := p.Characters.ItemStorage
	if s.DropBlockedMask == nil || *s.DropBlockedMask <= 0 || *s.DropBlockedMask > 65535 {
		return itemstore.Encoding{}, nil, fmt.Errorf("item_storage requires drop word mask")
	}
	fields := []*int{s.Encoding.Empty, s.Encoding.CodeMask, s.Encoding.WornMask,
		s.Encoding.CurseMask, s.Encoding.TransferBlockedMask}
	for _, value := range fields {
		if value == nil || *value < 0 || *value > 65535 {
			return itemstore.Encoding{}, nil, fmt.Errorf("item_storage requires explicit word encoding")
		}
	}
	if *s.DropBlockedMask&*s.Encoding.CodeMask != 0 {
		return itemstore.Encoding{}, nil, fmt.Errorf("drop word mask overlaps item identity")
	}
	if s.PartCount == nil || *s.PartCount < 1 || *s.PartCount > len([4]int{}) || len(s.Items) == 0 {
		return itemstore.Encoding{}, nil, fmt.Errorf("item_storage requires equipment parts and archive metadata")
	}
	e := itemstore.Encoding{
		Empty: uint16(*s.Encoding.Empty), CodeMask: uint16(*s.Encoding.CodeMask),
		WornMask: uint16(*s.Encoding.WornMask), CurseMask: uint16(*s.Encoding.CurseMask),
		TransferBlockedMask: uint16(*s.Encoding.TransferBlockedMask), PartCount: *s.PartCount,
	}
	metadata := make([]itemstore.Metadata, len(s.Items))
	for code, item := range s.Items {
		if item.EquipmentPart == nil || item.CursedWhenWorn == nil || item.DropForbidden == nil {
			return itemstore.Encoding{}, nil, fmt.Errorf("item_storage.items[%d] requires all metadata fields", code)
		}
		metadata[code] = itemstore.Metadata{Part: *item.EquipmentPart, CursedWhenWorn: *item.CursedWhenWorn}
	}
	return e, metadata, nil
}

func (p *Pack) validateItemStorage() error {
	e, metadata, err := p.itemWordContract()
	if err != nil {
		return err
	}
	if err := validateEvidence(p.Characters.ItemStorage.Evidence); err != nil {
		return fmt.Errorf("item_storage evidence: %w", err)
	}
	if level := p.Characters.ItemStorage.Evidence.Level; level != "D2" && level != "D3" {
		return fmt.Errorf("item_storage requires reviewed D2/D3 evidence")
	}
	// A synthetic empty buffer validates the explicit codec only; it is never
	// a character initializer or a fallback for missing character words.
	words := make([]uint16, p.Events.ItemActions.PersonalInventorySlots)
	for position := range words {
		words[position] = e.Empty
	}
	if _, err := itemstore.FromWords(len(words), e, metadata, words); err != nil {
		return fmt.Errorf("item_storage: %w", err)
	}
	return nil
}

// ValidateItemStorageAgainstItems must run against the archive actually loaded
// at boot. Schema/content versions do not prove count, shape or item metadata.
func (p *Pack) ValidateItemStorageAgainstItems(items *dq3data.Items) error {
	e, metadata, err := p.itemWordContract()
	if err != nil {
		return err
	}
	if items == nil || items.Count() != len(metadata) {
		return fmt.Errorf("item_storage metadata does not match actual item archive shape/count")
	}
	for code, item := range metadata {
		if item.Part != items.EquipSlot(code) || item.CursedWhenWorn != items.CursedWhenEquipped(code) || *p.Characters.ItemStorage.Items[code].DropForbidden != items.DropForbidden(code) {
			return fmt.Errorf("item_storage metadata differs from original item record %d", code)
		}
	}
	// Validate the complete codec as well; callers cannot bypass it by invoking
	// only the archive comparison on a modified Pack.
	if _, err := itemstore.FromWords(1, e, metadata, []uint16{e.Empty}); err != nil {
		return err
	}
	return nil
}

func (p *Pack) ItemStoreFromWords(words []uint16) (itemstore.Store, error) {
	e, metadata, err := p.itemWordContract()
	if err != nil {
		return itemstore.Store{}, err
	}
	return itemstore.FromWords(p.Events.ItemActions.PersonalInventorySlots, e, metadata, words)
}

func (p *Pack) CharacterItems(id string) (itemstore.Store, bool) {
	if p == nil {
		return itemstore.Store{}, false
	}
	d, ok := p.charDefaults[id]
	if !ok {
		return itemstore.Store{}, false
	}
	words, err := characterWords(d)
	if err != nil {
		return itemstore.Store{}, false
	}
	store, err := p.ItemStoreFromWords(words)
	return store, err == nil
}

func characterWords(d *CharacterDefault) ([]uint16, error) {
	if d.ItemWords == nil {
		return nil, fmt.Errorf("%s item_words is required", d.ID)
	}
	words := make([]uint16, len(d.ItemWords))
	for position, value := range d.ItemWords {
		if value == nil || *value < 0 || *value > 65535 {
			return nil, fmt.Errorf("%s item_words[%d] must be an explicit word", d.ID, position)
		}
		words[position] = uint16(*value)
	}
	return words, nil
}

func (p *Pack) NewGamePlayerItems() (itemstore.Store, bool) {
	return p.CharacterItems(p.Characters.DefaultRefs.NewGamePlayer)
}

func (p *Pack) RegisteredPartyMemberItems() (itemstore.Store, bool) {
	return p.CharacterItems(p.Characters.DefaultRefs.RegisteredPartyMember)
}

// DecodeItemStore requires the already reviewed pack/archive contract.
func (p *Pack) DecodeItemStore(raw []byte) (itemstore.Store, error) {
	e, metadata, err := p.itemWordContract()
	if err != nil {
		return itemstore.Store{}, err
	}
	return itemstore.Decode(p.Events.ItemActions.PersonalInventorySlots, e, metadata, raw)
}
