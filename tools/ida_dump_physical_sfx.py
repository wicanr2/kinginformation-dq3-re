"""非破壞性匯出戰鬥物理攻擊／受擊 VOC 與訊息控制流。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_physical_sfx.py OUT FVOC.VCX NVOC.VCX' DQ3.EXE

輸出保留原始函式名、IDA linear 位址、bytes、xref、CFG、可用時的 decompiler
文字與原始 VOC descriptor；不 rename、不寫 database comment、不修改輸入。
"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_gdl
import ida_hexrays
import ida_kernwin
import ida_lines
import ida_nalt
import idautils
import idc


FUNCTIONS = (
    0x1A973,
    0x1ABAE,
    0x1AB83,
    0x1AC05,
    0x1ACCE,
    0x1AFC6,
    0x1B3C3,
    0x1B23F,
    0x1B3F3,
    0x1B4F6,
    0x1B7B0,
    0x18222,
    0x19207,
    0x19834,
    0x1F470,
    0x20770,
    0x208E2,
    0x21414,
)

KEY_ADDRESSES = (
    # enemy physical entry
    0x1AC0B, 0x1AC0E, 0x1AC13, 0x1AC16,
    # common damage branches
    0x1ADAE, 0x1ADB1, 0x1ADC2, 0x1ADCB, 0x1ADD0, 0x1ADD3,
    0x1ADDB, 0x1ADE4, 0x1ADE9,
    0x1AEAE, 0x1AEB1,
    0x1AF13, 0x1AF16, 0x1AF1B, 0x1AF20, 0x1AF23, 0x1AF30,
    0x1AF41, 0x1AF44, 0x1AF58, 0x1AF61, 0x1AF66,
    # shared miss and death consumers
    0x1AF86, 0x1AF94, 0x1AF9C, 0x1AFA4, 0x1AFA7,
    0x1AFAD, 0x1AFB3, 0x1AFB6, 0x1AFBB, 0x1AFC1,
    0x1AFC6, 0x1AFC9, 0x1AFCE, 0x1AFD7, 0x1AFDC,
    # player physical entry
    0x1B568, 0x1B56B, 0x1B570, 0x1B573, 0x1B578,
    0x1B57D, 0x1B580, 0x1B585, 0x1B5A2,
    # effect tail
    0x1B7B6, 0x1B7B9, 0x1B7BE,
)

CUES = (1, 3, 4, 6, 9, 11)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def instruction(ea):
    fn = ida_funcs.get_func(ea)
    size = ida_bytes.get_item_size(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disasm(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "inference_level": "confirmed",
        "evidence_source": "IDA Pro 9.4 raw bytes and disassembly",
    }


def xref_record(x):
    fn = ida_funcs.get_func(x.frm)
    return {
        "from_ida_linear": hex(x.frm),
        "xref_type": int(x.type),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "instruction": instruction(x.frm),
        "inference_level": "confirmed",
        "evidence_source": "IDA Pro 9.4 xref database",
    }


def decompile_text(fn):
    try:
        cfunc = ida_hexrays.decompile(fn.start_ea)
        if cfunc is None:
            return {"available": False, "inference_level": "unknown"}
        return {
            "available": True,
            "text": str(cfunc),
            "inference_level": "navigation_only",
            "evidence_source": "Hex-Rays output; raw instructions/CFG remain authoritative",
        }
    except Exception as exc:
        return {
            "available": False,
            "error": type(exc).__name__,
            "inference_level": "unknown",
        }


def flow_record(fn):
    graph = ida_gdl.FlowChart(fn)
    blocks = []
    for block in graph:
        blocks.append({
            "start_ida_linear": hex(block.start_ea),
            "end_ida_linear": hex(block.end_ea),
            "successors": [hex(x.start_ea) for x in block.succs()],
            "predecessors": [hex(x.start_ea) for x in block.preds()],
            "instructions": [instruction(ea) for ea in idautils.Heads(block.start_ea, block.end_ea)],
            "inference_level": "confirmed",
            "evidence_source": "IDA Pro 9.4 FlowChart and raw instructions",
        })
    return blocks


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "inference_level": "unknown", "error": "no_function"}
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to_function_start": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "flow_blocks": flow_record(fn),
        "decompiler": decompile_text(fn),
        "inference_level": "confirmed",
        "semantic_scope": "raw boundary/xrefs/CFG only; semantic interpretation requires claim-level review",
        "evidence_source": "IDA Pro 9.4 function and xref databases",
    }


def le24(data, off):
    return data[off] | data[off + 1] << 8 | data[off + 2] << 16


def voc_descriptor(data, cue):
    table_off = cue * 4
    if table_off + 4 > len(data):
        return {"cue_raw": cue, "inference_level": "unknown", "error": "table_oob"}
    off = int.from_bytes(data[table_off:table_off + 4], "little")
    if off + 6 > len(data):
        return {"cue_raw": cue, "file_offset": hex(off), "inference_level": "unknown", "error": "block_oob"}
    block_type = data[off]
    block_length = le24(data, off + 1)
    rate_div = data[off + 4]
    codec = data[off + 5]
    samples = max(0, block_length - 2)
    source_rate = 1_000_000 // (256 - rate_div) if block_type == 1 and codec == 0 else 0
    duration_ns = samples * 1_000_000_000 // source_rate if source_rate else 0
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
        "semantic_scope": "raw descriptor/source duration only; audible meaning and host wall-clock unknown",
        "evidence_source": "VCX offset table and type-1 block bytes",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 4:
    raise RuntimeError("需要 OUT_JSON、FVOC.VCX、NVOC.VCX")

output_path = Path(idc.ARGV[1])
exe_path = Path(ida_nalt.get_input_file_path())
fvoc_path = Path(idc.ARGV[2])
nvoc_path = Path(idc.ARGV[3])
exe, fvoc, nvoc = exe_path.read_bytes(), fvoc_path.read_bytes(), nvoc_path.read_bytes()

result = {
    "schema": "dq3.ida_physical_sfx.v1",
    "input": {
        "exe": {"path": str(exe_path), "size": len(exe), "sha256": sha(exe)},
        "fvoc": {"path": str(fvoc_path), "size": len(fvoc), "sha256": sha(fvoc)},
        "nvoc": {"path": str(nvoc_path), "size": len(nvoc), "sha256": sha(nvoc)},
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; seg0 logical = linear - 0x10000; seg0 file = logical + 0x1370",
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_levels": "raw instructions/xrefs/CFG/VOC descriptors confirmed; decompiler navigation-only; action meaning pending review",
    },
    "functions": [function_record(ea) for ea in FUNCTIONS],
    "key_addresses": [instruction(ea) for ea in KEY_ADDRESSES],
    "exact_xrefs": {
        hex(ea): [xref_record(x) for x in idautils.XrefsTo(ea)]
        for ea in (0x1AC05, 0x1ACCE, 0x1AFC6, 0x1B4F6, 0x1B7B0,
                   0x20770, 0x208E2, 0x25B45, 0x272E1)
    },
    "voc_descriptors": {
        "FVOC.VCX": [voc_descriptor(fvoc, cue) for cue in CUES],
        "NVOC.VCX": [voc_descriptor(nvoc, cue) for cue in CUES],
    },
}

output_path.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
