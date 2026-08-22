"""非破壞性匯出 word_272E1==3／cue9／record334 的完整資料流。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_special_physical.py OUT FVOC.VCX NVOC.VCX' DQ3.EXE

只讀 IDA database 的原始名稱、bytes、函式邊界、CFG 與 xref；不 rename、
不寫 comment、不改原始 EXE。語意須在匯出後逐 claim 審查，名稱本身不是證據。
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
	0x13F98,
	0x1469F,
	0x14CF9,
    0x19530,
    0x19810,
    0x199DC,
    0x19B4C,
    0x19BF7,
    0x19C6B,
    0x19CB5,
    0x1A38C,
    0x1AC05,
    0x1ACCE,
    0x1AFE2,
    0x1B4F6,
    0x1BFD1,
    0x1BF7F,
    0x1C1D8,
    0x1C954,
    0x1CB3C,
    0x20770,
    0x208E2,
    0x21414,
)

RAW_RANGES = (
	(0x13D80, 0x13EB0, "field_item_clear_condition_handlers"),
    (0x13F70, 0x13FC0, "other_bit_0x10_writer_context"),
    (0x197F0, 0x19835, "bit_0x10_battle_state_consumer"),
    (0x199DC, 0x19B4C, "action_remap_and_dispatch"),
    (0x1A180, 0x1A4F0, "action_handlers_near_word_272E1_writers"),
    (0x1AC05, 0x1ACC8, "enemy_physical_or_special_entry"),
    (0x1ACCE, 0x1AFDE, "common_result_cue9_record334_and_damage"),
    (0x1AFE2, 0x1B061, "result_state_initializer"),
    (0x1B4F6, 0x1B5A8, "player_physical_entry"),
	(0x1CB30, 0x1CB90, "bit_0x10_spell_entry_consumer"),
	(0x1CC60, 0x1CCE0, "battle_clear_condition_callers"),
)

XREF_TARGETS = (
	0x1469F,
	0x14CF9,
	0x1CCB2,
    0x199DC,
    0x1A38C,
    0x1AC05,
    0x1ACCE,
    0x1AFE2,
    0x1B4F6,
    0x20770,
    0x208E2,
    0x21414,
    0x25B45,
    0x27287,
    0x272E1,
    0x27363,
    0x29D16,
)

CUES = (1, 3, 4, 6, 9, 11)


def relative_status_operand_candidates():
    """列出 +0x38 狀態 word 與 0x10／反遮罩 0xef 的直接指令候選。

    這不是語意判定：16-bit 程式大量透過不同 base register 間接存取角色
    record，IDA 的 data xref 不會自動連到每筆 record。匯出原始位址、bytes
    與所屬函式，供後續逐筆 caller／consumer 審查。
    """
    out = []
    for ea in idautils.Heads(0x10000, 0x2AEE0):
        if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
            continue
        line = disasm(ea).lower()
        if "38h" not in line:
            continue
        if not any(token in line for token in ("10h", "0efh", "0ffefh")):
            continue
        rec = instruction(ea)
        rec["inference_level"] = "candidate_only"
        rec["evidence_source"] = (
            "IDA Pro 9.4 raw operand scan; base-register provenance and semantics require review"
        )
        out.append(rec)
    return out


def immediate_bit_0x10_candidates():
    """列出程式段內直接測試／寫入 0x10 的指令，供人工追 base provenance。"""
    out = []
    for ea in idautils.Heads(0x10000, 0x2AEE0):
        if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
            continue
        line = disasm(ea).lower().strip()
        if not line.startswith(("test", "or ", "and ")) or "10h" not in line:
            continue
        rec = instruction(ea)
        rec["inference_level"] = "candidate_only"
        rec["evidence_source"] = (
            "IDA Pro 9.4 immediate-mask scan; register/data provenance and semantics require review"
        )
        out.append(rec)
    return out


def all_relative_status_operand_candidates():
    """完整列出程式段內直接存取角色 record +0x38 的指令。"""
    out = []
    for ea in idautils.Heads(0x10000, 0x2AEE0):
        if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
            continue
        if "38h" not in disasm(ea).lower():
            continue
        rec = instruction(ea)
        rec["inference_level"] = "candidate_only"
        rec["evidence_source"] = (
            "IDA Pro 9.4 +0x38 operand scan; base-register provenance and semantics require review"
        )
        out.append(rec)
    return out


def sha(data):
    return hashlib.sha256(data).hexdigest()


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def instruction(ea):
    size = ida_bytes.get_item_size(ea)
    fn = ida_funcs.get_func(ea)
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
    blocks = []
    for block in ida_gdl.FlowChart(fn):
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


def raw_range(start, end, label):
    return {
        "label": label,
        "start_ida_linear": hex(start),
        "end_ida_linear": hex(end),
        "instructions": [instruction(ea) for ea in idautils.Heads(start, end)],
        "inference_level": "confirmed",
        "evidence_source": "IDA Pro 9.4 item heads and raw bytes; function gaps preserved",
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
    "schema": "dq3.ida_special_physical.v1",
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
    "raw_ranges": [raw_range(*args) for args in RAW_RANGES],
    "exact_xrefs": {hex(ea): [xref_record(x) for x in idautils.XrefsTo(ea)] for ea in XREF_TARGETS},
    "relative_status_operand_candidates": relative_status_operand_candidates(),
    "immediate_bit_0x10_candidates": immediate_bit_0x10_candidates(),
    "all_relative_status_operand_candidates": all_relative_status_operand_candidates(),
    "voc_descriptors": {
        "FVOC.VCX": [voc_descriptor(fvoc, cue) for cue in CUES],
        "NVOC.VCX": [voc_descriptor(nvoc, cue) for cue in CUES],
    },
}

output_path.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
