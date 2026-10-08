"""Docker-only original SDK data-layout review, indexed in docs/25.

Preserve initial values and inference levels. Hardware-derived seed conversion
is not original runtime parity. Unreferenced payload meaning remains unknown.
"""

import argparse
import hashlib
import json
from pathlib import Path
import re
import struct


ROOT = Path(__file__).resolve().parents[1]
TABLES = {
    "CTVMEM.ASM": [(0x85, 14, "api_dispatch", 0x236dc), (0xa1, 8, "stream_dispatch", 0x23538)],
    "CMFDRV.ASM": [(0x229, 8, "status_dispatch", 0x24620), (0x239, 4, "controller_dispatch", 0x2482b),
                   (0x241, 16, "channel_dispatch", 0x24932), (0x261, 15, "api_dispatch", 0x24a61),
                   (0x27f, 3, "negative_api_dispatch", 0x24a70)],
}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--layout", type=Path, required=True)
    parser.add_argument("--platform", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    raw_path = ROOT / "assets_raw/DQ3.EXE"
    raw = raw_path.read_bytes()
    layout = json.loads(args.layout.read_text())
    if layout["input"]["sha256"] != sha(raw_path) or layout["tool"]["version"] != "9.4":
        raise ValueError("Original IDA identity differs")
    platform = json.loads(args.platform.read_text())
    if sha(args.platform.parent / "soundblaster.cpp") != platform["sha256"]:
        raise ValueError("Pinned platform source differs")
    text = (args.platform.parent / "soundblaster.cpp").read_text()
    table = re.search(r"e2_incr_table\[4\]\[9\]\s*=\s*\{(.*?)\n\};", text, re.S)
    if not table:
        raise ValueError("Pinned E2 contract table missing")
    rows = [[int(value, 0) for value in re.findall(r"-?0x[0-9a-fA-F]+|-?\d+", row)]
            for row in re.findall(r"\{([^{}]+)\}", table[1])]
    if len(rows) != 4 or any(len(row) != 9 for row in rows):
        raise ValueError("Pinned E2 contract shape differs")
    if not re.search(r"sb\.e2\.value\s*=\s*0xaa", text) or not re.search(r"sb\.e2\.count\s*=\s*0", text):
        raise ValueError("Pinned E2 reset contract differs")
    state = 0xaa
    outputs = []
    for count, value in enumerate((0x06, 0x6b)):
        state += sum(rows[count][bit] for bit in range(8) if value >> bit & 1) + rows[count][8]
        outputs.append(state & 255)
    if outputs != [0x3a, 0x08] or platform["derived_DMA_bytes"] != outputs:
        raise ValueError("Source-derived DMA result differs")
    if platform["derived_word"] != "0x083a":
        raise ValueError("E2 platform-derived result differs")
    reviews = []
    for module in layout["modules"]:
        base = int(module["ida_linear_start"], 16)
        start = int(module["file_start"], 16)
        heads = {int(row["ida_linear"], 16): row for row in module["instructions"]}
        tables = []
        for offset, count, name, consumer in TABLES[module["SDK_metadata_name"]]:
            if consumer not in heads or heads[consumer]["mnemonic"] != "call":
                raise ValueError("Declared table consumer is not original CALL")
            words = []
            for index in range(count):
                value = struct.unpack_from("<H", raw, start + offset + 2 * index)[0]
                target = base + value
                if module["SDK_metadata_name"] == "CTVMEM.ASM" and name == "api_dispatch" and index == 6:
                    if value != 0x6b06 or base + 0x83a not in heads:
                        raise ValueError("Original E2 mutable dispatch slot differs")
                    words.append({"index": index, "original_initial_word": hex(value), "type": "mutable_u16",
                                  "hardware_contract_derived_target": hex(base + 0x83a),
                                  "inference_level": "confirmed initial bytes; strong DMA-address/consumer closure; emulator-contract-derived mutation",
                                  "runtime_hardware_parity": False})
                else:
                    if target not in heads:
                        raise ValueError("Table word does not resolve to an original instruction head")
                    words.append({"index": index, "original_initial_word": hex(value), "type": "near_code_offset16",
                                  "original_IDA_target": hex(target), "original_IDA_name": heads[target]["original_name"],
                                  "inference_level": "confirmed bytes/head; legal index closure scoped by documented caller"})
            tables.append({"name": name, "module_offset": hex(offset), "word_count": count,
                           "original_consumer_IDA_linear": hex(consumer), "original_consumer": heads[consumer], "words": words})
        gaps = []
        for gap in module["uncovered_regions"]:
            lo, hi = int(gap["ida_linear_start"], 16), int(gap["ida_linear_end_exclusive"], 16)
            contents = bytes.fromhex(gap["raw_bytes"])
            if raw[int(gap["file_start"], 16):int(gap["file_start"], 16) + hi - lo] != contents:
                raise ValueError("Original gap bytes differ")
            if lo == 0x24bf3:
                stack_writer = heads[0x249b9]
                if stack_writer["file_bytes"] != "bc3713" or hi - base != 0x1337 or any(contents):
                    raise ValueError("CMF IRQ stack extent/initial bytes differ")
                role, level = "50-word interrupt stack reserve; SS=CS and SP at exclusive end", "confirmed scoped stack geometry and initial zero bytes"
            elif lo == 0x236a9:
                role, level = "unreferenced 11-byte initialized payload; meaning unknown", "strong non-code layout only; no semantic name or source-completion approval"
            else:
                role, level = "driver header, mutable state, near-handler tables and parameter arrays", "confirmed original data references and table words; unidentified fields retain unknown names"
            gaps.append({"IDA_linear_start": hex(lo), "IDA_linear_end_exclusive": hex(hi), "module_offset": hex(lo - base),
                         "size": len(contents), "sha256": hashlib.sha256(contents).hexdigest(), "role": role,
                         "inference_level": level, "literal_initial_values_known": True,
                         "no_original_code_array_approved": True})
        reviews.append({"SDK_metadata_name": module["SDK_metadata_name"], "tables": tables, "regions": gaps})
    result = {"schema_version": 1, "producer_sha256": sha(Path(__file__)), "input_EXE_sha256": sha(raw_path),
              "IDA_layout_sha256": sha(args.layout), "platform_receipt_sha256": sha(args.platform), "modules": reviews,
              "unreferenced_payload_meaning_known": False, "whole_modules_ready": False,
              "scope": "Data values/structures review; no source completion or original hardware wall-clock claim"}
    with args.output.open("x") as stream:
        stream.write(json.dumps(result, indent=2) + "\n")
    print(json.dumps({"tables": sum(len(m["tables"]) for m in reviews), "data_regions": 4, "E2_initial_word_preserved": "0x6b06", "IRQ_stack_bytes": 100, "whole_modules_ready": False}))


if __name__ == "__main__":
    main()
