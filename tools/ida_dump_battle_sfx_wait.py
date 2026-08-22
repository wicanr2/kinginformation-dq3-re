"""非破壞性匯出戰鬥 VOC cue、driver dispatch 與 wait call-site 證據。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_battle_sfx_wait.py OUT FVOC.VCX NVOC.VCX' DQ3.EXE

輸出保留原始函式名、IDA linear 位址、bytes、xref 與原始 VOC descriptor；
不 rename、不寫 comment、不修改 database 或輸入檔。
"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_nalt
import idautils
import idc


FUNCTIONS = (
    0x1A973,  # enemy actor consumer; flee cue 0x15
    0x1AC05,  # physical/special entrance
    0x1ACCE,  # damage branches
    0x1B47E,  # player command/action branch
    0x1B4F6,  # player physical/flee-related branch
    0x1B7B0,  # effect tail
    0x1D86D,
    0x1D881,
    0x1D8D6,
    0x20770,  # VOC cue dispatch
    0x208A7,  # bank selection/load
    0x208E2,  # driver completion wait
    0x22CF5,  # Sound Blaster command dispatch
    0x236B4,  # driver dispatch
)
KEY_CALLS = (
    0x1AA37,  # enemy flee: raw cue writer BP=0x15
    0x1AA3A,  # enemy flee: BP=0x15 -> sub_20770
    0x1AA47,  # enemy flee wait after message
    0x1AC21,  # physical/special cue call (review exact BP writer nearby)
    0x1B4D3,  # player flee: raw cue writer BP=0x0d
    0x1B4D6,  # player flee: BP=0x0d -> sub_20770
    0x1B4E3,  # player flee wait after message
    0x20770,
    0x208E2,
)
CUES = (1, 4, 6, 8, 9, 13, 16, 21)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def line(ea):
    return ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""


def instruction(ea):
    fn = ida_funcs.get_func(ea)
    size = ida_bytes.get_item_size(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": line(ea),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "inference_level": "confirmed",
        "evidence_source": "IDA Pro 9.4 raw bytes and disassembly",
    }


def xref_record(x):
    fn = ida_funcs.get_func(x.frm)
    return {
        "from_ida_linear": hex(x.frm),
        "xref_type": int(x.type),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "instruction": instruction(x.frm),
        "inference_level": "confirmed",
        "evidence_source": "IDA Pro 9.4 xref database",
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(ea),
            "inference_level": "unknown",
            "error": "no_function",
        }
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to_function_start": [
            xref_record(x) for x in idautils.XrefsTo(fn.start_ea)
        ],
        "instructions": [instruction(i) for i in idautils.FuncItems(fn.start_ea)],
        "inference_level": "confirmed",
        "semantic_scope": "raw function boundary/instructions/xrefs only; function meaning requires claim-level review",
        "evidence_source": "IDA Pro 9.4 function database",
    }


def le24(data, off):
    return data[off] | data[off + 1] << 8 | data[off + 2] << 16


def le32(data, off):
    return int.from_bytes(data[off:off + 4], "little")


def voc_descriptor(data, cue):
    table_off = cue * 4
    if table_off + 4 > len(data):
        return {"cue_raw": cue, "inference_level": "unknown", "error": "table_oob"}
    off = le32(data, table_off)
    if off + 6 > len(data):
        return {"cue_raw": cue, "file_offset": hex(off), "inference_level": "unknown", "error": "block_oob"}
    block_type = data[off]
    block_length = le24(data, off + 1)
    rate_div = data[off + 4]
    codec = data[off + 5]
    samples = max(0, block_length - 2)
    source_rate = 0
    duration_ns = 0
    if block_type == 1 and codec == 0 and rate_div < 256:
        source_rate = 1_000_000 // (256 - rate_div)
        if source_rate > 0:
            duration_ns = samples * 1_000_000_000 // source_rate
    return {
        "cue_raw": cue,
        "file_offset": hex(off),
        "block_type": block_type,
        "block_length": block_length,
        "rate_div": rate_div,
        "codec": codec,
        "source_rate": source_rate,
        "source_samples": samples,
        "source_duration_nanos": duration_ns,
        "inference_level": "confirmed",
        "semantic_scope": "raw VOC descriptor and source-duration arithmetic; host wall-clock unknown",
        "evidence_source": "VCX table and type-1 VOC block bytes",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 4:
    raise RuntimeError("需要 OUT_JSON、FVOC.VCX、NVOC.VCX")

output_path = Path(idc.ARGV[1])
exe_path = Path(ida_nalt.get_input_file_path())
fvoc_path = Path(idc.ARGV[2])
nvoc_path = Path(idc.ARGV[3])
exe, fvoc, nvoc = (
    exe_path.read_bytes(), fvoc_path.read_bytes(), nvoc_path.read_bytes()
)

result = {
    "schema": "dq3.ida_battle_sfx_wait.v1",
    "input": {
        "exe": {"path": str(exe_path), "size": len(exe), "sha256": sha(exe)},
        "fvoc": {"path": str(fvoc_path), "size": len(fvoc), "sha256": sha(fvoc)},
        "nvoc": {"path": str(nvoc_path), "size": len(nvoc), "sha256": sha(nvoc)},
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": (
            "IDA linear; seg0 logical = linear - 0x10000; "
            "seg0 file = logical + 0x1370; segmented driver functions retain IDA linear"
        ),
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_levels": (
            "instructions/xrefs/VOC descriptors confirmed; player-flee cue 13 and "
            "enemy-flee cue 21 confirmed; remaining cue semantics and host wall-clock unknown"
        ),
    },
    "semantic_claims": [
        {
            "semantic": "player_fled uses FVOC cue 13 then waits for driver completion",
            "original_locations": ["IDA 0x1b4d3", "IDA 0x1b4d6", "IDA 0x1b4e3"],
            "raw_operands": ["mov bp,0x0d", "call sub_20770", "call sub_208E2"],
            "inference_level": "confirmed",
            "evidence_source": "sub_1B47E success branch writer/call/message/wait/consumer chain",
        },
        {
            "semantic": "enemy_fled uses FVOC cue 21 then waits for driver completion",
            "original_locations": ["IDA 0x1aa37", "IDA 0x1aa3a", "IDA 0x1aa47"],
            "raw_operands": ["mov bp,0x15", "call sub_20770", "call sub_208E2"],
            "inference_level": "confirmed",
            "evidence_source": "sub_1A973 flee branch writer/call/message/wait/consumer chain",
        },
        {
            "semantic": "host Sound Blaster wall-clock and resampled waveform identity",
            "original_locations": ["sub_208E2", "sub_236B4"],
            "inference_level": "unknown",
            "evidence_source": "static driver completion path lacks a host-visible timing oracle",
        },
    ],
    "functions": [function_record(ea) for ea in FUNCTIONS],
    "key_calls": [instruction(ea) for ea in KEY_CALLS],
    "exact_xrefs": {
        hex(ea): [xref_record(x) for x in idautils.XrefsTo(ea)]
        for ea in (0x20770, 0x208E2, 0x22CF5, 0x236B4)
    },
    "voc_descriptors": {
        "FVOC.VCX": [voc_descriptor(fvoc, cue) for cue in CUES],
        "NVOC.VCX": [voc_descriptor(nvoc, cue) for cue in CUES],
    },
}

output_path.write_text(
    json.dumps(result, ensure_ascii=False, indent=2) + "\n",
    encoding="utf-8",
)
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
