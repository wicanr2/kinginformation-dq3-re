// Issue #5 diagnostic C/ASM adapter; Docker-only entry and scope in docs/25.
// Explicit register, memory and stack injection. No normal-player-path claim.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

const inputHash = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
const stateWord = 0x0b5a
const returnOffset = 1
const stackSeg = 0x8000
const stackTop = 0xff00
const originalEntry = 0xe6c9
const adapterEntry = 0x0400
const cEntry = 0x0500

type outcome struct {
	Registers    [8]uint16 `json:"registers"`
	High         [8]uint16 `json:"high_registers"`
	Segments     [4]uint16 `json:"segments"`
	ExtraSeg     [2]uint16 `json:"extra_segments"`
	IP           uint16    `json:"ip"`
	Flags        uint16    `json:"flags"`
	State        uint16    `json:"state_word"`
	Changes      []uint32  `json:"modified_addresses"`
	Instructions int       `json:"instructions"`
	ScratchStack string    `json:"final_scratch_stack_hex"`
}

type localProbe struct {
	m      *machine.Machine
	cs, ds uint16
	state  uint32
	active bool
	writes []uint32
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func loadProbe(raw []byte) *localProbe {
	m := machine.New()
	if err := m.LoadEXE(raw); err != nil {
		panic(err)
	}
	cs := uint16(m.ImageBase >> 4)
	if m.Read8(cpu.Addr(cs, 0x92ab)) != 0xb8 {
		panic("startup DS carrier differs")
	}
	ds := m.Read16(cpu.Addr(cs, 0x92ac))
	p := &localProbe{m: m, cs: cs, ds: ds, state: cpu.Addr(ds, stateWord)}
	m.WatchWrites(0, machine.MemSize-1, func(address uint32, old, value byte) {
		if p.active {
			p.writes = append(p.writes, address)
		}
	})
	return p
}

func (p *localProbe) run(entry, seed, bound, initialFlags uint16) (outcome, error) {
	c := p.m.CPU
	c.Model = cpu.Model80386
	c.Reset()
	c.Halted, c.DivErrors, c.Cycles = false, nil, 0
	c.R = [8]uint16{0xa55a ^ seed, 0x1357, 0x2468, bound, stackTop, 0x5678, 0x6789, 0x789a}
	c.Hi = [8]uint16{0x1101, 0x2202, 0x3303, 0x4404, 0x5505, 0x6606, 0x7707, 0x8808}
	c.Seg = [4]uint16{0x7000, p.cs, stackSeg, p.ds}
	c.FSeg, c.GSeg = 0x6000, 0x6100
	c.IP = entry
	c.SetFlags(initialFlags)
	// Fresh zeroed scratch stack for each predetermined call, on both sides.
	p.m.WriteBytes(cpu.Addr(stackSeg, stackTop-128), make([]byte, 130))
	p.m.WriteBytes(p.state-1, make([]byte, 6))
	p.m.Write16(p.state, seed)
	p.m.Write16(cpu.Addr(stackSeg, stackTop), returnOffset)
	p.writes = p.writes[:0]
	p.active = true
	defer func() { p.active = false }()
	steps := 0
	for c.IP != returnOffset || c.Seg[cpu.CS] != p.cs {
		if steps >= 128 {
			return outcome{}, fmt.Errorf("local call exceeded 128 instructions")
		}
		if err := c.Step(); err != nil {
			return outcome{}, err
		}
		steps++
		if c.Halted || len(c.DivErrors) > 0 {
			return outcome{}, fmt.Errorf("halt/division exception: %v", c.DivErrors)
		}
	}
	stack := make([]byte, 128)
	for i := range stack {
		stack[i] = p.m.Read8(cpu.Addr(stackSeg, stackTop-128) + uint32(i))
	}
	return outcome{c.R, c.Hi, c.Seg, [2]uint16{c.FSeg, c.GSeg}, c.IP, c.Flags,
		p.m.Read16(p.state), append([]uint32(nil), p.writes...), steps, hex.EncodeToString(stack)}, nil
}

func modeledResult(r outcome, seed, bound uint16) bool {
	expected := [8]uint16{0xa55a ^ seed, 0x1357, 0x2468, bound, stackTop + 2, 0x5678, 0x6789, 0x789a}
	state := seed
	if bound == 0 {
		expected[cpu.DX] = 0
	} else {
		state = bits.RotateLeft16(seed+0x9018, 3)
		expected[cpu.AX], expected[cpu.DX] = state/bound, state%bound
	}
	return r.Registers == expected && r.State == state
}

func registerStateEqual(a, b outcome) bool {
	return a.Registers == b.Registers && a.High == b.High && a.Segments == b.Segments &&
		a.ExtraSeg == b.ExtraSeg && a.IP == b.IP && a.State == b.State
}

func readCode(root, name string) []byte {
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		panic(err)
	}
	return data
}

func main() {
	exe := flag.String("exe", "", "pristine original")
	compiled := flag.String("compiled", "", "new compiler/assembler artifacts")
	output := flag.String("output", "", "new local receipt")
	flag.Parse()
	if os.Getuid() == 0 {
		panic("run as host UID/GID")
	}
	raw, err := os.ReadFile(*exe)
	if err != nil {
		panic(err)
	}
	if len(raw) != 115282 || digestBytes(raw) != inputHash {
		panic("original identity differs")
	}
	a, exact, candidate := loadProbe(raw), loadProbe(raw), loadProbe(raw)
	coreCode, boundCode := readCode(*compiled, "sub_e6b9.bin"), readCode(*compiled, "sub_e6c9.bin")
	if len(coreCode) != 16 || len(boundCode) != 30 {
		panic("exact ASM scope differs")
	}
	exact.m.WriteBytes(cpu.Addr(exact.cs, 0xe6b9), coreCode)
	exact.m.WriteBytes(cpu.Addr(exact.cs, originalEntry), boundCode)
	adapter, cCode := readCode(*compiled, "adapter.bin"), readCode(*compiled, "c-code.bin")
	if len(adapter) > cEntry-adapterEntry || len(cCode) > 0x100 {
		panic("diagnostic code slots overlap")
	}
	candidate.m.WriteBytes(cpu.Addr(candidate.cs, adapterEntry), adapter)
	candidate.m.WriteBytes(cpu.Addr(candidate.cs, cEntry), cCode)
	started := time.Now()
	counts := map[string]int{"bound10_all_u16_seeds": 0, "bound0_all_u16_seeds": 0, "boundary_inputs": 0}
	var samples []map[string]any
	extraStackCalls, finalStackDifferences, rawFlagDifferences, zeroFlagDifferences := 0, 0, 0, 0
	var originalInstructions, exactInstructions, candidateInstructions uint64
	check := func(label string, seed, bound, flags uint16, sample bool) {
		x, err := a.run(originalEntry, seed, bound, flags)
		if err != nil || !modeledResult(x, seed, bound) {
			panic(fmt.Sprintf("original model differs: %v seed=%04x BX=%04x", err, seed, bound))
		}
		y, err := exact.run(originalEntry, seed, bound, flags)
		if err != nil || !reflect.DeepEqual(x, y) {
			panic("source ASM control differs")
		}
		z, err := candidate.run(adapterEntry, seed, bound, flags)
		if err != nil || !registerStateEqual(x, z) {
			panic(fmt.Sprintf("C/ASM register-state mismatch seed=%04x BX=%04x: %v got=%+v original=%+v", seed, bound, err, z, x))
		}
		// Arithmetic DIV flags are undefined. Control flags must still agree.
		if x.Flags&(cpu.IF|cpu.DF|cpu.TF) != z.Flags&(cpu.IF|cpu.DF|cpu.TF) {
			panic("control flags differ")
		}
		if x.Flags != z.Flags {
			rawFlagDifferences++
			if bound == 0 {
				zeroFlagDifferences++
				panic("zero-path flags differ")
			}
		}
		stackWrites := 0
		for _, address := range z.Changes {
			if address == candidate.state || address == candidate.state+1 {
				continue
			}
			start := cpu.Addr(stackSeg, stackTop-128)
			if address < start || address >= cpu.Addr(stackSeg, stackTop) {
				panic(fmt.Sprintf("C candidate modified memory outside state/scratch stack: %x", address))
			}
			stackWrites++
		}
		if bound != 0 && stackWrites == 0 {
			panic("C call produced no observed scratch stack changes")
		}
		if bound == 0 && len(z.Changes) != 0 {
			panic("zero bound wrote memory")
		}
		if stackWrites > 0 {
			extraStackCalls++
		}
		if x.ScratchStack != z.ScratchStack {
			finalStackDifferences++
		}
		originalInstructions += uint64(x.Instructions)
		exactInstructions += uint64(y.Instructions)
		candidateInstructions += uint64(z.Instructions)
		counts[label]++
		if sample {
			samples = append(samples, map[string]any{"seed_fixed_before_call": fmt.Sprintf("0x%04x", seed),
				"bound_BX": bound, "initial_flags": flags, "original": x, "source_ASM": y,
				"C_ASM_adapter": z, "register_and_persistent_state_equal": true,
				"raw_flags_equal": x.Flags == z.Flags, "additional_scratch_stack_changes": stackWrites,
				"final_scratch_stack_equal": x.ScratchStack == z.ScratchStack,
				"full_memory_changes_equal": reflect.DeepEqual(x.Changes, z.Changes)})
		}
	}
	for seed := 0; seed < 65536; seed++ {
		check("bound10_all_u16_seeds", uint16(seed), 10, 2, seed == 0x1357)
		check("bound0_all_u16_seeds", uint16(seed), 0, 2, seed == 0x1357)
	}
	seeds := []uint16{0, 1, 0x6fe7, 0x6fe8, 0x7fff, 0x8000, 0xfffe, 0xffff}
	bounds := []uint16{1, 2, 3, 4, 6, 10, 100, 256, 0x8000, 0xffff}
	for _, seed := range seeds {
		for _, bound := range bounds {
			check("boundary_inputs", seed, bound, 0x0e47, seed == 0xffff && bound == 0xffff)
		}
	}
	positiveSeconds := time.Since(started).Seconds()
	var negativeCases []map[string]any
	for _, test := range []struct {
		name, adapter, code string
		seed, bound         uint16
	}{
		{"swapped-results", "swapped-results.bin", "c-code.bin", 0x1357, 10},
		{"wrong-argument", "wrong-argument.bin", "c-code.bin", 0x1357, 10},
		{"missing-zero-guard", "missing-zero-guard.bin", "c-code.bin", 0x1357, 0},
		{"wrong-data-placement", "adapter.bin", "wrong-data-code.bin", 0x1357, 10},
	} {
		p := loadProbe(raw)
		p.m.WriteBytes(cpu.Addr(p.cs, adapterEntry), readCode(*compiled, test.adapter))
		p.m.WriteBytes(cpu.Addr(p.cs, cEntry), readCode(*compiled, test.code))
		got, failure := p.run(adapterEntry, test.seed, test.bound, 2)
		wanted, err := a.run(originalEntry, test.seed, test.bound, 2)
		if err != nil {
			panic(err)
		}
		rejected := failure != nil || !registerStateEqual(wanted, got)
		if !rejected {
			panic("predefined negative adapter unexpectedly accepted: " + test.name)
		}
		reason := "register/persistent-state mismatch"
		if failure != nil {
			reason = failure.Error()
		}
		negativeCases = append(negativeCases, map[string]any{"case": test.name, "rejected": rejected,
			"seed_fixed_before_call": fmt.Sprintf("0x%04x", test.seed), "BX": test.bound, "reason": reason})
	}
	if extraStackCalls != 65616 || finalStackDifferences != 65616 {
		panic("nonzero-bound scratch footprint count differs")
	}
	if counts["bound10_all_u16_seeds"] != 65536 || counts["bound0_all_u16_seeds"] != 65536 || counts["boundary_inputs"] != 80 {
		panic("predeclared input coverage differs")
	}
	receipt := map[string]any{
		"schema_version": 1, "input": map[string]any{"path": *exe, "size": len(raw), "sha256": inputHash},
		"address_space": map[string]any{"kind": "dosgolem real-mode linear", "load_base": a.m.ImageBase,
			"CS": a.cs, "DS": a.ds, "state_address": a.state, "original_entry_logical": "0xe6c9",
			"original_entry_IDA_linear": "0x1e6c9", "adapter_entry_logical": "0x0400", "C_entry_logical": "0x0500"},
		"case_count": 131152, "counts": counts, "wall_seconds": positiveSeconds,
		"instructions":                        map[string]any{"original": originalInstructions, "exact_ASM": exactInstructions, "C_ASM_adapter": candidateInstructions},
		"register_and_persistent_state_equal": true, "source_ASM_full_outcome_equal": true,
		"calls_with_additional_stack_changes": extraStackCalls, "calls_with_raw_flag_differences": rawFlagDifferences,
		"calls_with_final_scratch_stack_differences": finalStackDifferences,
		"zero_bound_flag_differences":                zeroFlagDifferences,
		"full_memory_equivalent":                     false, "byte_exact_C": false,
		"samples": samples, "negative_cases": negativeCases,
		"code_hashes": map[string]any{"adapter": digestBytes(adapter), "C": digestBytes(cCode),
			"exact_ASM_core": digestBytes(coreCode), "exact_ASM_bound": digestBytes(boundCode)},
		"scope": map[string]any{"direct_entry": true, "state_injection": true, "CPU_reentry": true,
			"normal_player_path": false, "production_changed": false,
			"seed_method": "predeclared complete u16 enumeration plus fixed boundaries; no rerolls"},
		"inference_level": "confirmed local register/state results on selected dosgolem engine; original compiler and campaign unknown",
		"flags_limit":     "raw differences retained; DIV arithmetic flags have no guessed architectural expected values",
		"observer_limit":  "modified byte values observed; stores of equal values not counted; scratch stack reset equally before each call",
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
	fmt.Printf("RNG_ADAPTER_PASS cases=131152 register_state=true extra_stack=%d raw_flags=%d seconds=%.6f\n", extraStackCalls, rawFlagDifferences, positiveSeconds)
}
