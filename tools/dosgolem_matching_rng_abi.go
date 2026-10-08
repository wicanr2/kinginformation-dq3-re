// DQ3 Issue #5: controlled local RNG ABI probe; entry and scope in docs/25.
// Build inside an isolated copy of the dosgolem module; never run on the host.
// This deliberately injects registers, state and a near-return stack. It is not
// a normal-player-path receipt and does not control the production game's RNG.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/bits"
	"os"
	"reflect"
	"time"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

const originalHash = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
const stateOffset = 0x0b5a
const returnIP = 1
const stackSegment = 0x8000
const stackPointer = 0xff00

type result struct {
	Registers    [8]uint16 `json:"registers"`
	High         [8]uint16 `json:"high_registers"`
	Segments     [4]uint16 `json:"segments"`
	IP           uint16    `json:"ip"`
	Flags        uint16    `json:"flags"`
	State        uint16    `json:"state_word"`
	Writes       []uint32  `json:"modified_addresses"`
	Instructions int       `json:"instructions"`
}

type probe struct {
	m            *machine.Machine
	cs, ds       uint16
	stateAddress uint32
	active       bool
	writes       []uint32
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func newProbe(data []byte) *probe {
	m := machine.New()
	if err := m.LoadEXE(data); err != nil {
		panic(err)
	}
	cs := uint16(m.ImageBase >> 4)
	// The original startup MOV AX,seg / MOV DS,AX is at logical 92AB/92AE.
	if m.Read8(cpu.Addr(cs, 0x92ab)) != 0xb8 {
		panic("startup DS carrier differs")
	}
	ds := m.Read16(cpu.Addr(cs, 0x92ac))
	p := &probe{m: m, cs: cs, ds: ds, stateAddress: cpu.Addr(ds, stateOffset)}
	m.WatchWrites(0, machine.MemSize-1, func(address uint32, old, value byte) {
		if p.active {
			p.writes = append(p.writes, address)
		}
	})
	return p
}

func (p *probe) run(entry, seed, bound, initialFlags uint16) (result, [8]uint16, [4]uint16, [8]uint16) {
	c := p.m.CPU
	c.Model = cpu.Model80386
	c.Reset()
	c.Halted = false
	c.DivErrors = nil
	c.Cycles = 0
	c.R = [8]uint16{0xa55a ^ seed, 0x1357, 0x2468, bound, stackPointer, 0x5678, 0x6789, 0x789a}
	c.Hi = [8]uint16{0x1101, 0x2202, 0x3303, 0x4404, 0x5505, 0x6606, 0x7707, 0x8808}
	c.Seg = [4]uint16{0x7000, p.cs, stackSegment, p.ds}
	c.FSeg, c.GSeg = 0x6000, 0x6100
	c.IP = entry
	c.SetFlags(initialFlags)
	p.m.Write16(p.stateAddress, seed)
	p.m.Write16(cpu.Addr(stackSegment, stackPointer), returnIP)
	before, segments, high := c.R, c.Seg, c.Hi
	p.writes = p.writes[:0]
	p.active = true
	count := 0
	for c.IP != returnIP || c.Seg[cpu.CS] != p.cs {
		if count >= 32 {
			panic("local routine exceeded 32-instruction budget")
		}
		if err := c.Step(); err != nil {
			panic(err)
		}
		count++
	}
	p.active = false
	if len(c.DivErrors) != 0 || c.Halted || c.FSeg != 0x6000 || c.GSeg != 0x6100 {
		panic("unexpected exception, halt or segment modification")
	}
	return result{c.R, c.Hi, c.Seg, c.IP, c.Flags, p.m.Read16(p.stateAddress), append([]uint32(nil), p.writes...), count}, before, segments, high
}

func verify(r result, before [8]uint16, segments [4]uint16, high [8]uint16, entry, seed, bound uint16, stateAddress uint32) {
	advanced := bits.RotateLeft16(seed+0x9018, 3)
	expected := before
	expected[cpu.SP] += 2
	state := advanced
	count := 7
	if entry == 0xe6b9 {
		expected[cpu.AX] = advanced
	} else if entry == 0xe6c9 {
		if bound == 0 {
			expected[cpu.DX] = 0
			state = seed
			count = 4
		} else {
			expected[cpu.AX] = advanced / bound
			expected[cpu.DX] = advanced % bound
			count = 11
		}
	} else {
		panic("unknown local entry")
	}
	if r.Registers != expected || r.Segments != segments || r.High != high || r.State != state || r.IP != returnIP || r.Instructions != count {
		panic(fmt.Sprintf("ABI differs entry=%04x seed=%04x BX=%04x got=%+v expected_R=%x state=%04x", entry, seed, bound, r, expected, state))
	}
	if entry == 0xe6c9 && bound == 0 {
		if len(r.Writes) != 0 {
			panic("zero bound wrote state")
		}
	} else {
		// WatchWrites reports value changes, not stores of equal bytes.
		var expectedChanges []uint32
		if uint8(seed) != uint8(state) {
			expectedChanges = append(expectedChanges, stateAddress)
		}
		if uint8(seed>>8) != uint8(state>>8) {
			expectedChanges = append(expectedChanges, stateAddress+1)
		}
		if !reflect.DeepEqual(r.Writes, expectedChanges) {
			panic("unexpected modified memory footprint")
		}
	}
	// DIV flags are architecturally undefined, so no scalar model guesses them.
	if entry == 0xe6b9 {
		sum := seed + 0x9018
		var wanted uint16
		if advanced&1 != 0 {
			wanted |= cpu.CF
		}
		if bits.OnesCount8(uint8(sum))%2 == 0 {
			wanted |= cpu.PF
		}
		if (seed^0x9018^sum)&0x10 != 0 {
			wanted |= cpu.AF
		}
		if sum == 0 {
			wanted |= cpu.ZF
		}
		if sum&0x8000 != 0 {
			wanted |= cpu.SF
		}
		if ((advanced >> 15) ^ (advanced & 1)) != 0 {
			wanted |= cpu.OF
		}
		mask := uint16(cpu.CF | cpu.PF | cpu.AF | cpu.ZF | cpu.SF | cpu.OF)
		if r.Flags&mask != wanted {
			panic("defined core flags differ")
		}
	}
}

func main() {
	exe := flag.String("exe", "", "pristine original executable")
	rebuilt := flag.String("rebuilt", "", "private original-byte scaffold with source-assembled RNG slices")
	output := flag.String("output", "", "new JSON receipt")
	wordControl := flag.String("word-control", "", "optional source-compiled MSC word echo code")
	pairControl := flag.String("pair-control", "", "optional source-compiled MSC DX:AX pair echo code")
	flag.Parse()
	if os.Getuid() == 0 {
		panic("probe must not run as root")
	}
	raw, err := os.ReadFile(*exe)
	if err != nil {
		panic(err)
	}
	other, err := os.ReadFile(*rebuilt)
	if err != nil {
		panic(err)
	}
	if len(raw) != 115282 || digest(raw) != originalHash || digest(other) != originalHash {
		panic("input or assembled scaffold identity differs")
	}
	a, b := newProbe(raw), newProbe(other)
	started := time.Now()
	counts := map[string]int{"core_all_u16_seeds": 0, "bound10_all_u16_seeds": 0, "bound0_all_u16_seeds": 0, "boundary_inputs": 0}
	var totalInstructions uint64
	samples := []map[string]any{}
	check := func(label string, entry, seed, bound, flags uint16, sample bool) {
		x, pre, segs, high := a.run(entry, seed, bound, flags)
		y, _, _, _ := b.run(entry, seed, bound, flags)
		verify(x, pre, segs, high, entry, seed, bound, a.stateAddress)
		if !reflect.DeepEqual(x, y) {
			panic("original and assembled source ABI differ")
		}
		counts[label]++
		totalInstructions += uint64(x.Instructions + y.Instructions)
		if sample {
			samples = append(samples, map[string]any{"entry_logical": fmt.Sprintf("0x%x", entry), "seed_fixed_before_call": fmt.Sprintf("0x%04x", seed), "bound_BX": bound, "initial_flags": flags, "original": x, "source_assembled": y})
		}
	}
	for seed := 0; seed < 65536; seed++ {
		check("core_all_u16_seeds", 0xe6b9, uint16(seed), 10, 2, seed == 0x1357)
		check("bound10_all_u16_seeds", 0xe6c9, uint16(seed), 10, 2, seed == 0x1357)
		check("bound0_all_u16_seeds", 0xe6c9, uint16(seed), 0, 2, seed == 0x1357)
	}
	seeds := []uint16{0, 1, 0x6fe7, 0x6fe8, 0x7fff, 0x8000, 0xfffe, 0xffff}
	bounds := []uint16{1, 2, 3, 4, 6, 10, 100, 256, 0x8000, 0xffff}
	for _, seed := range seeds {
		for _, bound := range bounds {
			check("boundary_inputs", 0xe6c9, seed, bound, 0x0e47, false)
		}
	}
	for _, entry := range []uint16{0xe6b9, 0xe6c9} {
		for _, flags := range []uint16{0x0202, 0x0e47} {
			check("boundary_inputs", entry, 0x1357, 10, flags, true)
		}
	}
	known, _, _, _ := a.run(0xe6c9, 0x1357, 10, 2)
	if known.Registers[cpu.AX] == known.Registers[cpu.DX] {
		panic("predefined wrong-return-register negative is inconclusive")
	}
	var compilerControls []map[string]any
	for _, control := range []struct{ name, path string }{{"word", *wordControl}, {"pair", *pairControl}} {
		if control.path == "" {
			continue
		}
		code, err := os.ReadFile(control.path)
		if err != nil {
			panic(err)
		}
		p := newProbe(raw)
		const entry = 0x400
		p.m.WriteBytes(cpu.Addr(p.cs, entry), code)
		c := p.m.CPU
		c.Model = cpu.Model80386
		c.Reset()
		c.Halted = false
		c.DivErrors = nil
		c.R = [8]uint16{0xa55a, 0x1357, 0x2468, 10, stackPointer, 0x5678, 0x6789, 0x789a}
		c.Seg = [4]uint16{0x7000, p.cs, stackSegment, p.ds}
		c.IP = entry
		c.SetFlags(2)
		// BX=10 conflicts deliberately with the predeclared stack argument BEEF.
		p.m.Write16(cpu.Addr(stackSegment, stackPointer), returnIP)
		p.m.Write16(cpu.Addr(stackSegment, stackPointer+2), 0xbeef)
		steps := 0
		for c.IP != returnIP || c.Seg[cpu.CS] != p.cs {
			if steps >= 32 {
				panic("compiler control exceeded local budget")
			}
			if err := c.Step(); err != nil {
				panic(err)
			}
			steps++
		}
		if c.R[cpu.SP] != stackPointer+2 || c.R[cpu.BP] != 0x5678 || c.R[cpu.BX] != 10 {
			panic("compiler C control stack/register contract differs")
		}
		if control.name == "word" && (c.R[cpu.AX] != 0xbeef || c.R[cpu.DX] != 0x2468) {
			panic("MSC word return does not consume stack argument / return AX")
		}
		if control.name == "pair" && (c.R[cpu.DX] != 0xbeef || c.R[cpu.AX] != 0xbeef>>1) {
			panic("MSC long return does not use DX:AX")
		}
		compilerControls = append(compilerControls, map[string]any{"case": control.name, "code_sha256": digest(code), "BX_input": 10, "stack_argument": 0xbeef, "AX_output": c.R[cpu.AX], "DX_output": c.R[cpu.DX], "SP_output": c.R[cpu.SP], "instructions": steps, "scope": "mounted MSC candidate ABI only"})
	}
	receipt := map[string]any{
		"input":                     map[string]any{"path": *exe, "size": len(raw), "sha256": digest(raw)},
		"source_assembled_scaffold": map[string]any{"path": *rebuilt, "sha256": digest(other), "source_assembled_bytes": 46, "retained_original_bytes": len(raw) - 46},
		"address_space":             map[string]any{"kind": "dosgolem real-mode linear", "load_base": a.m.ImageBase, "CS": a.cs, "DS": a.ds, "state_address": a.stateAddress, "IDA_load_base": "0x10000"},
		"counts":                    counts, "total_original_and_candidate_instructions": totalInstructions, "wall_seconds": time.Since(started).Seconds(),
		"samples": samples, "negative_wrong_AX_remainder_assumption_rejected": true,
		"compiler_controls": compilerControls,
		"scope":             map[string]any{"state_injection": true, "direct_entry": true, "CPU_reentry": true, "normal_player_path": false, "production_RNG_locked": false, "seed_method": "predeclared exhaustive enumeration plus fixed boundary vectors; no rerolls"},
		"inference_level":   "confirmed local instruction ABI; no campaign or original compiler claim",
		"flags_limit":       "DIV arithmetic flags not assigned guessed scalar expectations; both byte-identical routines compared on the same dosgolem engine",
		"observer_limit":    "WatchWrites reports modified byte values; equal-value stores are identified statically, not counted dynamically",
	}
	file, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(receipt); err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("RNG_ABI_PASS core=%d bound10=%d bound0=%d boundaries=%d instructions=%d seconds=%.6f\n", counts["core_all_u16_seeds"], counts["bound10_all_u16_seeds"], counts["bound0_all_u16_seeds"], counts["boundary_inputs"], totalInstructions, time.Since(started).Seconds())
}
