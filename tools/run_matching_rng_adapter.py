"""Docker-only C/ASM RNG adapter experiment; scope and receipts in docs/25.

compile runs in dq3-msc; cpu runs in the existing Go image. Originals, compiler
and dosgolem are read-only inputs. The adapter is a diagnostic, never production.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import time

from omf_matching_probe import read_object, resolve_ds_offsets
from run_matching_probe import COMPILER_HASHES, ORIGINAL_HASH, compare


ROOT = Path(__file__).resolve().parents[1]
C_SOURCE = """extern unsigned g_rng;
unsigned long rngpair(unsigned limit) {
    unsigned value;
    value = g_rng + 0x9018u;
    value = (value << 3) | (value >> 13);
    g_rng = value;
    return ((unsigned long)(value % limit) << 16) | (unsigned long)(value / limit);
}
"""


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_json(path, data):
    with path.open("x") as stream:
        stream.write(json.dumps(data, ensure_ascii=False, indent=2) + "\n")


def original():
    path = ROOT / "assets_raw/DQ3.EXE"
    raw = path.read_bytes()
    if len(raw) != 115282 or sha(path) != ORIGINAL_HASH:
        raise ValueError("Original input identity differs")
    return path, raw


def compile_probe(args):
    path, raw = original()
    evidence = json.loads(args.ida_evidence.read_text())
    if evidence["input"]["sha256"] != ORIGINAL_HASH or evidence["tool"]["version"] != "9.4":
        raise ValueError("IDA input/tool identity differs")
    args.output.mkdir(exist_ok=False)
    compiler = {}
    for name, expected in COMPILER_HASHES.items():
        component = args.msc_root / "BIN" / name
        if sha(component) != expected:
            raise ValueError("Mounted compiler component differs: " + name)
        compiler[name] = {"sha256": expected, "size": component.stat().st_size}
    # The MZ header size comes from the actual input, not a compiler hypothesis.
    header = int.from_bytes(raw[8:10], "little") * 16
    assemblies = []
    for start, end, count in ((0xe6b9, 0xe6c9, 7), (0xe6c9, 0xe6e7, 13)):
        record = next(f for f in evidence["functions"] if int(f["ida_linear_start"], 16) == 0x10000 + start)
        instructions = record["instructions"]
        if len(instructions) != count:
            raise ValueError("IDA instruction boundary differs")
        recorded = b"".join(bytes.fromhex(i["file_bytes"]) for i in instructions)
        wanted = raw[header + start:header + end]
        if recorded != wanted:
            raise ValueError("IDA original bytes differ")
        source = ROOT / "re/match" / ("sub_" + format(start, "x") + ".asm")
        binary = args.output / (source.stem + ".bin")
        subprocess.run(["nasm", "-f", "bin", str(source), "-o", str(binary)], check=True, timeout=20)
        if binary.read_bytes() != wanted:
            raise ValueError("Semantic ASM differs")
        assemblies.append({"original_name": record["original_name"], "logical_start": hex(start),
                           "ida_linear_start": hex(0x10000 + start), "file_start": hex(header + start),
                           "source": str(source.relative_to(ROOT)), "source_sha256": sha(source),
                           "code_sha256": sha(binary), "code_size": len(wanted), "exact": True})
    source = args.output / "rngpair.c"
    source.write_text(C_SOURCE, encoding="ascii")
    commands = ["@echo off", "set PATH=C:\\BIN", "set INCLUDE=C:\\INCLUDE\\INCLUDE",
                "set LIB=C:\\LIB", "set TMP=D:\\", "d:",
                "C:\\BIN\\CL.EXE /c /AS /Ox rngpair.c > compile.log",
                "if errorlevel 1 goto failed", "echo SUCCESS > DONE.TXT", "goto end",
                ":failed", "echo FAIL > DONE.TXT", ":end"]
    (args.output / "go.bat").write_text("\r\n".join(commands) + "\r\n", encoding="ascii")
    config = ("[sdl]\noutput=surface\n[cpu]\ncycles=fixed 100000\n[mixer]\nnosound=true\n"
              "[autoexec]\nmount c " + str(args.msc_root) + "\nmount d " + str(args.output)
              + "\nd:\ncall go.bat\nexit\n")
    (args.output / "dosbox.conf").write_text(config, encoding="ascii")
    started = time.monotonic()
    with (args.output / "dosbox.log").open("wb") as stream:
        process = subprocess.run(["dosbox", "-conf", str(args.output / "dosbox.conf"), "-exit"],
                                 stdout=stream, stderr=subprocess.STDOUT, timeout=90, check=False)
    seconds = time.monotonic() - started
    marker = args.output / "DONE.TXT"
    if process.returncode or not marker.is_file() or marker.read_text().strip() != "SUCCESS":
        raise ValueError("C compilation failed; inspect preserved output")
    obj = read_object(args.output / "RNGPAIR.OBJ")
    code, fixups = resolve_ds_offsets(obj, "_rngpair", {"_g_rng": 0x0b5a})
    if not fixups or any(f["symbol"] != "_g_rng" for f in fixups):
        raise ValueError("C candidate data relocation differs")
    (args.output / "c-code.bin").write_bytes(code)
    # A deliberately wrong data placement is a separate negative input.
    wrong, _ = resolve_ds_offsets(obj, "_rngpair", {"_g_rng": 0x0b5b})
    (args.output / "wrong-data-code.bin").write_bytes(wrong)
    adapter = ROOT / "re/match/rng_adapter.asm"
    subprocess.run(["nasm", "-f", "bin", str(adapter), "-o", str(args.output / "adapter.bin")],
                   check=True, timeout=20)
    negative_adapters = {
        "swapped-results": adapter.read_text().replace("    add sp, 2", "    xchg ax, dx\n    add sp, 2"),
        "wrong-argument": adapter.read_text().replace("    push bx\n    call", "    push cx\n    call"),
        "missing-zero-guard": adapter.read_text().replace("    jne .nonzero", "    jmp .nonzero"),
    }
    for name, content in negative_adapters.items():
        assembly = args.output / (name + ".asm")
        assembly.write_text(content, encoding="ascii")
        subprocess.run(["nasm", "-f", "bin", str(assembly), "-o", str(args.output / (name + ".bin"))],
                       check=True, timeout=20)
    fingerprint = []
    # Bounded printable version markers are leads for these compiler inputs only.
    for name in COMPILER_HASHES:
        binary = (args.msc_root / "BIN" / name).read_bytes()
        for match in re.finditer(rb"[\x20-\x7e]{8,160}", binary):
            text = match.group().decode("ascii")
            if "Microsoft" in text or "Version" in text or "5.10" in text or "5.1" in text:
                fingerprint.append({"input_component": name, "file_offset": hex(match.start()),
                                    "text": text, "inference_level": "confirmed bytes in mounted compiler; original compiler unknown"})
    files = {p.name: sha(p) for p in args.output.iterdir() if p.suffix == ".bin"}
    write_json(args.output / "compile-receipt.json", {
        "schema_version": 1, "scope": "authored diagnostic C and ASM adapter; no production changes",
        "input": {"path": str(path.relative_to(ROOT)), "size": len(raw), "sha256": ORIGINAL_HASH},
        "ida_evidence": {"path": str(args.ida_evidence), "sha256": sha(args.ida_evidence),
                         "address_space": evidence["address_space"]},
        "compiler_inputs": compiler, "candidate_component_markers": fingerprint,
        "producer_sha256": sha(Path(__file__)), "omf_parser_sha256": sha(ROOT / "tools/omf_matching_probe.py"),
        "c_source_sha256": sha(source), "adapter_source": str(adapter.relative_to(ROOT)),
        "adapter_source_sha256": sha(adapter), "object_sha256": obj["sha256"],
        "c_flags": "/c /AS /Ox", "c_fixups": fixups,
        "c_bound_raw_compare": compare(raw[header + 0xe6c9:header + 0xe6e7], code),
        "semantic_asm": assemblies, "artifact_hashes": files,
        "adapter_code_size": len((args.output / "adapter.bin").read_bytes()),
        "compile_wall_seconds": seconds, "tool": subprocess.check_output(["nasm", "-v"], text=True).strip(),
        "original_compiler": "unknown", "C_exact_candidates": int(code == raw[header + 0xe6c9:header + 0xe6e7]),
    })
    print(json.dumps({"C_bytes": len(code), "adapter_bytes": len((args.output / "adapter.bin").read_bytes()),
                      "ASM_exact_bytes": 46, "compile_seconds": seconds}))


def cpu_probe(args):
    original()
    compiled = json.loads((args.compiled / "compile-receipt.json").read_text())
    for name, expected in compiled["artifact_hashes"].items():
        if sha(args.compiled / name) != expected:
            raise ValueError("Compiled artifact differs: " + name)
    args.output.mkdir(exist_ok=False)
    sources = {}
    authored = ROOT / "tools/dosgolem_matching_rng_adapter.go"
    with tempfile.TemporaryDirectory(prefix="dq3-rng-adapter-") as temporary:
        root = Path(temporary)
        for path in sorted((args.dosgolem_root / "internal").rglob("*.go")):
            if path.name.endswith("_test.go"):
                continue
            relative = path.relative_to(args.dosgolem_root)
            target = root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, target)
            sources[str(relative)] = sha(path)
        shutil.copyfile(args.dosgolem_root / "go.mod", root / "go.mod")
        sources["go.mod"] = sha(args.dosgolem_root / "go.mod")
        target = root / "cmd/matching-rng-adapter/main.go"
        target.parent.mkdir(parents=True)
        shutil.copyfile(authored, target)
        subprocess.run(["gofmt", "-w", str(target)], check=True, timeout=20)
        shutil.copyfile(target, args.output / "probe-source.go")
        env = os.environ.copy()
        env.update(GOCACHE=str(root / "cache"), GOMODCACHE=str(root / "modules"),
                   GOPROXY="off", GOTOOLCHAIN="local", CGO_ENABLED="0", GOMAXPROCS="2")
        binary = args.output / "rng-adapter-probe"
        started = time.monotonic()
        with (args.output / "build.log").open("wb") as stream:
            subprocess.run(["go", "build", "-p", "2", "-trimpath", "-o", str(binary),
                            "./cmd/matching-rng-adapter"], cwd=root, env=env,
                           stdout=stream, stderr=subprocess.STDOUT, timeout=180, check=True)
        seconds = time.monotonic() - started
    with (args.output / "run.log").open("wb") as stream:
        subprocess.run([str(binary), "-exe", str(ROOT / "assets_raw/DQ3.EXE"),
                        "-compiled", str(args.compiled), "-output", str(args.output / "receipt.json")],
                       stdout=stream, stderr=subprocess.STDOUT, timeout=120, check=True)
    receipt = json.loads((args.output / "receipt.json").read_text())
    if (receipt["case_count"] != 131152 or len(receipt["negative_cases"]) != 4
            or receipt["counts"] != {"bound10_all_u16_seeds": 65536, "bound0_all_u16_seeds": 65536,
                                     "boundary_inputs": 80}):
        raise ValueError("Controlled input coverage differs")
    if not all(case["rejected"] for case in receipt["negative_cases"]):
        raise ValueError("Negative adapter was accepted")
    write_json(args.output / "producer-meta.json", {
        "producer_sha256": sha(Path(__file__)), "authored_go_source_sha256": sha(authored),
        "formatted_go_source_sha256": sha(args.output / "probe-source.go"),
        "compiled_receipt_sha256": sha(args.compiled / "compile-receipt.json"),
        "engine_sources": sources,
        "engine_sources_canonical_sha256": hashlib.sha256(json.dumps(sources, sort_keys=True, separators=(",", ":")).encode()).hexdigest(),
        "engine_scope": "non-test internal Go sources plus go.mod; no cmd/probe scratch",
        "go_version": subprocess.check_output(["go", "version"], text=True).strip(),
        "build_seconds": seconds, "binary_sha256": sha(binary),
        "receipt_sha256": sha(args.output / "receipt.json"),
    })
    print(json.dumps({"cases": receipt["case_count"], "CPU_seconds": receipt["wall_seconds"],
                      "build_seconds": seconds, "negative_cases": receipt["negative_cases"]}))


def audit_probe(args):
    original()
    compiled = json.loads((args.compiled / "compile-receipt.json").read_text())
    cpu = json.loads((args.cpu / "receipt.json").read_text())
    meta = json.loads((args.cpu / "producer-meta.json").read_text())
    # Both phases must have consumed the current authored producers.
    if compiled["producer_sha256"] != sha(Path(__file__)) or meta["producer_sha256"] != sha(Path(__file__)):
        raise ValueError("Producer freshness differs")
    if compiled["omf_parser_sha256"] != sha(ROOT / "tools/omf_matching_probe.py"):
        raise ValueError("OMF parser freshness differs")
    if compiled["adapter_source_sha256"] != sha(ROOT / compiled["adapter_source"]):
        raise ValueError("Adapter source freshness differs")
    if compiled["c_source_sha256"] != hashlib.sha256(C_SOURCE.encode("ascii")).hexdigest():
        raise ValueError("Authored C candidate differs")
    if meta["authored_go_source_sha256"] != sha(ROOT / "tools/dosgolem_matching_rng_adapter.go"):
        raise ValueError("Go probe freshness differs")
    if meta["formatted_go_source_sha256"] != sha(args.cpu / "probe-source.go"):
        raise ValueError("Formatted Go source differs")
    for key, path in (("compiled_receipt_sha256", args.compiled / "compile-receipt.json"),
                      ("receipt_sha256", args.cpu / "receipt.json"), ("binary_sha256", args.cpu / "rng-adapter-probe")):
        if meta[key] != sha(path):
            raise ValueError("CPU provenance differs: " + key)
    for relative, expected in meta["engine_sources"].items():
        if sha(args.dosgolem_root / relative) != expected:
            raise ValueError("Selected engine source differs: " + relative)
    for name, expected in compiled["artifact_hashes"].items():
        if sha(args.compiled / name) != expected:
            raise ValueError("Compiled artifact differs: " + name)
    for assembly in compiled["semantic_asm"]:
        if sha(ROOT / assembly["source"]) != assembly["source_sha256"]:
            raise ValueError("Exact ASM source differs")
    evidence = Path(compiled["ida_evidence"]["path"])
    if sha(evidence) != compiled["ida_evidence"]["sha256"]:
        raise ValueError("Consumed IDA evidence differs")
    # Repeat code/data inputs must reproduce; wall time is deliberately excluded.
    repeat = json.loads((args.repeat_compiled / "compile-receipt.json").read_text())
    for key in ("artifact_hashes", "object_sha256", "c_source_sha256", "c_fixups", "compiler_inputs"):
        if compiled[key] != repeat[key]:
            raise ValueError("Independent rebuild differs: " + key)
    if not (cpu["register_and_persistent_state_equal"] and cpu["source_ASM_full_outcome_equal"]
            and not cpu["full_memory_equivalent"] and not cpu["byte_exact_C"]
            and compiled["C_exact_candidates"] == 0 and len(cpu["negative_cases"]) == 4
            and cpu["calls_with_final_scratch_stack_differences"] == 65616
            and cpu["counts"] == {"bound10_all_u16_seeds": 65536, "bound0_all_u16_seeds": 65536,
                                  "boundary_inputs": 80}
            and all(case["rejected"] for case in cpu["negative_cases"])):
        raise ValueError("Adapter result scope differs")
    ledger = json.loads((ROOT / "tools/ida_rng_abi_ledger.json").read_text())
    for annotation in ledger["annotations"]:
        raw = (ROOT / ledger["input_path"]).read_bytes()
        offset = int(annotation["file_offset"], 16)
        expected = bytes.fromhex(annotation["bytes"])
        if raw[offset:offset + len(expected)] != expected:
            raise ValueError("Original ABI ledger bytes differ")
    dynamic = ledger["dynamic_evidence"]
    if sha(ROOT / dynamic["path"]) != dynamic["sha256"]:
        raise ValueError("Prior local ABI receipt differs")
    # Positive control proves the existing documentation entry is searchable.
    index = (ROOT / "docs/25-match-progress.md").read_text()
    if "ida_matching_probe.py" not in index:
        raise ValueError("Known indexed positive control missing")
    for name in ("run_matching_rng_adapter.py", "dosgolem_matching_rng_adapter.go", "rng_adapter.asm"):
        if name not in index:
            raise ValueError("New tool lacks documentation entry: " + name)
    for folder in (args.compiled, args.cpu):
        for path in folder.rglob("*"):
            if path.stat().st_uid != os.getuid() or path.stat().st_gid != os.getgid():
                raise ValueError("Output ownership differs: " + str(path))
            if path.is_dir() and path.name.endswith(".md"):
                raise ValueError("Markdown directory in output")
    result = {"input_unchanged": True, "producer_and_engine_fresh": True, "independent_artifact_rebuild_equal": True,
              "receipt_sha256": sha(args.cpu / "receipt.json"), "compiled_receipt_sha256": sha(args.compiled / "compile-receipt.json"),
              "register_and_persistent_state_equal": True, "full_memory_equivalent": False,
              "C_exact_candidates": 0, "semantic_ASM_exact_bytes": 46, "controlled_inputs": 131152,
              "negative_cases_rejected": 4, "prior_ABI_ledger_bytes_and_receipt_verified": True,
              "original_compiler": "unknown", "normal_player_path": False, "UID": os.getuid(), "GID": os.getgid()}
    write_json(args.cpu / "final-audit.json", result)
    print(json.dumps(result))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="phase", required=True)
    compiler = sub.add_parser("compile")
    compiler.add_argument("--output", type=Path, required=True)
    compiler.add_argument("--ida-evidence", type=Path, required=True)
    compiler.add_argument("--msc-root", type=Path, default=Path("/msc"))
    cpu = sub.add_parser("cpu")
    cpu.add_argument("--output", type=Path, required=True)
    cpu.add_argument("--compiled", type=Path, required=True)
    cpu.add_argument("--dosgolem-root", type=Path, default=Path("/dosgolem"))
    audit = sub.add_parser("audit")
    audit.add_argument("--compiled", type=Path, required=True)
    audit.add_argument("--repeat-compiled", type=Path, required=True)
    audit.add_argument("--cpu", type=Path, required=True)
    audit.add_argument("--dosgolem-root", type=Path, default=Path("/dosgolem"))
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Run with host UID/GID")
    {"compile": compile_probe, "cpu": cpu_probe, "audit": audit_probe}[args.phase](args)


if __name__ == "__main__":
    main()
