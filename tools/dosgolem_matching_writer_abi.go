// Controlled local writer ABI oracle; docs/25-match-progress.md.
// This injects registers, data and a near-return stack, then enters CFE8.
// It is not a normal-player-path or whole-source-EXE receipt. No RNG executes.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

const inputSHA = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
const entryIP, calleeIP, returnIP = 0xcfe8, 0xd2ab, 1
const stackSeg, stackTop = 0x8000, 0xff00

type change struct {
	Address  uint32
	Old, New byte
}
type outcome struct {
	Registers                   [8]uint16
	High                        [8]uint16
	Segments                    [4]uint16
	ExtraSegments               [2]uint16
	IP, Flags, Word, AXAtCallee uint16
	Changes                     []change
	Scratch                     []byte
	Instructions                int
}
type probe struct {
	m       *machine.Machine
	cs, ds  uint16
	active  bool
	changes []change
}

func digest(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }
func load(raw, code []byte) *probe {
	m := machine.New()
	if err := m.LoadEXE(raw); err != nil {
		panic(err)
	}
	cs := uint16(m.ImageBase >> 4)
	if m.Read8(cpu.Addr(cs, 0x92ab)) != 0xb8 {
		panic("startup DS carrier differs")
	}
	p := &probe{m: m, cs: cs, ds: m.Read16(cpu.Addr(cs, 0x92ac))}
	if code != nil {
		m.WriteBytes(cpu.Addr(cs, entryIP), code)
	}
	m.WatchWrites(0, machine.MemSize-1, func(a uint32, old, new byte) {
		if p.active {
			p.changes = append(p.changes, change{a, old, new})
		}
	})
	return p
}
func (p *probe) run(ax, index uint16, table byte, flags uint16) outcome {
	if table&0x18 != 0 {
		panic("only the declared no-action branch is allowed")
	}
	c := p.m.CPU
	c.Model = cpu.Model80386
	c.Reset()
	c.Halted, c.DivErrors, c.Cycles = false, nil, 0
	c.R = [8]uint16{ax, 0x1357, 0x2468, 0x3456, stackTop, 0x5678, 0x6789, 0x789a}
	c.Hi = [8]uint16{0x1101, 0x2202, 0x3303, 0x4404, 0x5505, 0x6606, 0x7707, 0x8808}
	c.Seg = [4]uint16{0x7000, p.cs, stackSeg, p.ds}
	c.FSeg, c.GSeg, c.IP = 0x6000, 0x6100, entryIP
	c.SetFlags(flags)
	p.m.Write16(cpu.Addr(p.ds, 0x24b5), index)
	p.m.Write8(cpu.Addr(p.ds, uint16(0x37c5+uint32(index)*3)), table)
	p.m.Write16(cpu.Addr(p.ds, 0x0743), 0xbeef)
	p.m.WriteBytes(cpu.Addr(stackSeg, stackTop-32), make([]byte, 36))
	p.m.Write16(cpu.Addr(stackSeg, stackTop), returnIP)
	p.changes = p.changes[:0]
	p.active = true
	defer func() { p.active = false }()
	steps, calls := 0, 0
	var atCall uint16
	for c.IP != returnIP || c.Seg[cpu.CS] != p.cs {
		if steps >= 64 {
			panic("local routine exceeded its instruction budget")
		}
		if c.Seg[cpu.CS] != p.cs {
			panic("unexpected code frame")
		}
		if c.IP == calleeIP {
			atCall = c.R[cpu.AX]
			calls++
		}
		if c.IP == 0xd2bf || c.IP == 0xd2c7 || c.IP == 0xd2cf {
			panic("action branch reached")
		}
		if err := c.Step(); err != nil {
			panic(err)
		}
		steps++
	}
	if calls != 1 || c.Halted || len(c.DivErrors) != 0 {
		panic("invalid local return")
	}
	return outcome{c.R, c.Hi, c.Seg, [2]uint16{c.FSeg, c.GSeg}, c.IP, c.Flags,
		p.m.Read16(cpu.Addr(p.ds, 0x0743)), atCall, append([]change(nil), p.changes...),
		append([]byte(nil), p.m.Mem[cpu.Addr(stackSeg, stackTop-32):cpu.Addr(stackSeg, stackTop+4)]...), steps}
}
func main() {
	exe := flag.String("exe", "", "pristine original EXE")
	goodPath := flag.String("good", "", "source-compiled AX-live code")
	badPath := flag.String("bad", "", "same-profile C code which discards incoming AX")
	output := flag.String("output", "", "new JSON receipt")
	flag.Parse()
	if os.Getuid() == 0 {
		panic("use host UID/GID")
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			panic(e)
		}
		return b
	}
	raw, good, bad := read(*exe), read(*goodPath), read(*badPath)
	if len(raw) != 115282 || digest(raw) != inputSHA || !bytes.Equal(good, raw[0xe358:0xe362]) {
		panic("input or matched C differs")
	}
	if len(bad) != 9 || bytes.Equal(good, bad) {
		panic("discarded-AX negative differs from its declared scope")
	}
	a, b, n := load(raw, nil), load(raw, good), load(raw, bad)
	if !bytes.Equal(a.m.Mem, b.m.Mem) {
		panic("initial original/candidate memory differs")
	}
	counts := map[string]int{"all_u16_AX": 0, "boundary_vectors": 0}
	mismatches, memorySamples := 0, 0
	samples := []map[string]any{}
	started := time.Now()
	check := func(label string, ax, index uint16, table byte, flags uint16, fullMemory bool) {
		x, y, z := a.run(ax, index, table, flags), b.run(ax, index, table, flags), n.run(ax, index, table, flags)
		want := [8]uint16{ax & 0xff00, 0x1357, 0x2468, uint16(uint32(index) * 3), stackTop + 2, 0x5678, 0x6789, 0x789a}
		if x.Registers != want || x.Word != 0 || x.AXAtCallee != ax || x.Instructions != 15 {
			panic(fmt.Sprintf("original ABI differs: %+v", x))
		}
		if !reflect.DeepEqual(x, y) {
			panic("source C and original outputs or writes differ")
		}
		if fullMemory {
			if !bytes.Equal(a.m.Mem, b.m.Mem) {
				panic("sampled complete VM memory differs")
			}
			memorySamples++
		}
		if z.AXAtCallee != 0 || z.Registers[cpu.AX] != 0 || z.Word != 0 {
			panic("negative did not discard AX as declared")
		}
		if z.Registers[cpu.AX] != x.Registers[cpu.AX] {
			mismatches++
		}
		counts[label]++
		if label == "all_u16_AX" && (ax == 0xa55a || ax == 0x005a) {
			samples = append(samples, map[string]any{"AX": ax, "index": index, "table_raw_byte": table, "flags": flags, "original": x, "source_C": y, "discarded_AX_C": z})
		}
	}
	for ax := 0; ax < 65536; ax++ {
		check("all_u16_AX", uint16(ax), 0, 0, 2, ax&0xff == 0)
	}
	for _, index := range []uint16{0, 1, 2, 7, 15, 0x1234, 0x8000, 0xffff} {
		for _, ax := range []uint16{0x0055, 0x0155, 0x8055, 0xff55} {
			for _, flags := range []uint16{2, 0x0202, 0x0e47} {
				for _, table := range []byte{0, 1, 0x20, 0xe7} {
					check("boundary_vectors", ax, index, table, flags, false)
				}
			}
		}
	}
	if counts["all_u16_AX"] != 65536 || counts["boundary_vectors"] != 384 || mismatches != 65568 || memorySamples != 256 {
		panic("declared vector counts differ")
	}
	receipt := map[string]any{
		"input":                map[string]any{"path": *exe, "size": len(raw), "sha256": digest(raw)},
		"source_C":             map[string]any{"code_sha256": digest(good), "compiled_bytes": len(good)},
		"discarded_AX_control": map[string]any{"code_sha256": digest(bad), "compiled_bytes": len(bad), "AX_return_mismatches": mismatches},
		"address_space":        map[string]any{"kind": "dosgolem real-mode linear", "load_base": a.m.ImageBase, "CS": a.cs, "DS_injected_from_startup_carrier": a.ds, "entry_logical": "0xcfe8", "original_callee_logical": "0xd2ab"},
		"counts":               counts, "full_VM_memory_samples": memorySamples, "samples": samples, "wall_seconds": time.Since(started).Seconds(),
		"scope": map[string]any{"state_injection": true, "CPU_reentry": true, "code_patch_in_VM": true, "normal_player_path": false, "whole_source_EXE": false, "original_callee_retained": true, "RNG_executed": false, "seed_method": "not applicable; predeclared no-action path executes no RNG", "valid_gameplay_indices": "unknown"},
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(append(data, '\n')); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("65536 AX + 384 boundary cases; %d discarded-AX mismatches\n", mismatches)
}
