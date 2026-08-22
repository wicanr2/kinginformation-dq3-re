"""非破壞性匯出 monster46 action gate／選 bit／目標亂數順序證據。

IDA Pro 9.4 batch（原始 EXE 必須先複製到容器內可建立 database 的暫存位置）：

  idat -A '-Stools/ida_dump_monster46_action_rng.py OUT D3MNS.DAT' DQ3.EXE

本工具只讀 IDA database 的原始名稱、函式邊界、bytes、CFG/xref 與指定 DGROUP
table；不 rename、不寫 comment、不 patch、不修改原始 EXE。所有語意仍須逐 claim
審查；輸出中的 raw instruction／xref 是證據，並非自動命名後的事實。

本輪特別保存：

* sub_1A973 的 cast gate 兩個 caller 分支；
* sub_199DC → sub_19AD6 的 mask selection 與 sub_19AD6 的 bounded RNG；
* monster46 bit9／bit41 的 remap／handler pointer；
* bit41 handler 前後的所有已知 RNG call-site；
* target-selection 候選函式及其 direct/indirect caller 限制。

輸出 sidecar 應放在 gitignored work/，不可加入 Git。
"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_gdl
import ida_kernwin
import ida_lines
import ida_nalt
import idautils
import idc


MONSTER_ID = 46
MONSTER_STRIDE = 0x29
REMAP_IDA = 0x28700       # DGROUP 0x3930
HANDLERS_IDA = 0x2871E    # DGROUP 0x394e
RNG_256_IDA = 0x1E6B9
RNG_BOUNDED_IDA = 0x1E6C9

FUNCTIONS = (
    0x19207,
    0x199DC,
    0x19AD6,
    0x19B4C,
    0x19C6B,
    0x19CB5,
    0x1A3EE,
    0x1A6B5,
    0x1A6D1,
    0x1A973,
    0x1AB83,
    0x1AC05,
    0x1ACCE,
    0x1C08B,
    0x1C642,
    0x1C673,
    0x1D49A,
    0x1D881,
    0x1E6B9,
    0x1E6C9,
)

RAW_RANGES = (
    (0x191F0, 0x19299, "alternate_context_selection"),
    (0x199DC, 0x19BC3, "action_gate_remap_dispatch"),
    (0x19C40, 0x19D90, "target_scope_dispatch_candidates"),
    (0x1A973, 0x1AAA1, "enemy_action_caller"),
    (0x1A3EE, 0x1A4A0, "monster46_bit41_handler_window"),
    (0x1AC05, 0x1AFDE, "physical_and_common_damage_window"),
    (0x1AB83, 0x1ABAE, "normal_living_party_target_selection"),
    (0x1C630, 0x1C6D0, "enemy_target_setup_candidates"),
    (0x1D850, 0x1D8D0, "spell_effect_setup_candidate"),
)

KEY_SITES = (
    0x19207, 0x1922C, 0x19230, 0x19244,
    0x19AD6, 0x19AE3, 0x19B15, 0x19B1E, 0x19B2C, 0x19B45,
    0x199F3, 0x199FB, 0x19A0C, 0x19A0F, 0x19A17, 0x19A37,
    0x19A53, 0x19A58, 0x19A5E, 0x19A66, 0x19A6E, 0x19AC1,
    0x19ACA, 0x19AD1,
    0x1A9F9, 0x1AA01, 0x1AA0F, 0x1AA16, 0x1AA4F, 0x1AA57,
    0x1AA5A, 0x1AA61, 0x1AA65, 0x1AA69, 0x1AA6C, 0x1AA73,
    0x1AA79, 0x1AA7C,
    0x1A3EE, 0x1A3F6, 0x1A402, 0x1A406, 0x1A40D, 0x1A410,
    0x1A414, 0x1A418, 0x1A42B, 0x1A430,
    0x1AB83, 0x1AB91, 0x1AB95, 0x1ABA2, 0x1ABA8,
)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def fn_name(ea):
    fn = ida_funcs.get_func(ea)
    return ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"


def instruction(ea):
    size = ida_bytes.get_item_size(ea)
    if size <= 0:
        size = 1
    fn = ida_funcs.get_func(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disasm(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "inference_level": "confirmed",
        "evidence_source": "IDA Pro 9.4 raw item heads and bytes",
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


def code_refs_from(ea):
    refs = []
    for target in idautils.CodeRefsFrom(ea, False):
        refs.append({
            "target_ida_linear": hex(target),
            "target_function_original": fn_name(target),
            "inference_level": "confirmed_direct_code_ref",
            "evidence_source": "IDA Pro 9.4 CodeRefsFrom",
        })
    return refs


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(ea),
            "inference_level": "unknown",
            "error": "no_function_boundary",
        }
    instructions = []
    for item in idautils.FuncItems(fn.start_ea):
        rec = instruction(item)
        rec["direct_code_refs_from"] = code_refs_from(item)
        instructions.append(rec)
    blocks = []
    try:
        for block in ida_gdl.FlowChart(fn):
            blocks.append({
                "start_ida_linear": hex(block.start_ea),
                "end_ida_linear": hex(block.end_ea),
                "successors": [hex(x.start_ea) for x in block.succs()],
                "predecessors": [hex(x.start_ea) for x in block.preds()],
                "inference_level": "confirmed",
                "evidence_source": "IDA Pro 9.4 FlowChart",
            })
    except Exception as exc:
        blocks.append({
            "inference_level": "unknown",
            "error": type(exc).__name__,
        })
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to_function_start": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "flow_blocks": blocks,
        "instructions": instructions,
        "inference_level": "confirmed_raw_boundary_and_xrefs",
        "semantic_scope": "raw instructions/CFG/xrefs only; claim semantics require review",
        "evidence_source": "IDA Pro 9.4 function database",
    }


def raw_range(start, end, label):
    return {
        "label": label,
        "start_ida_linear": hex(start),
        "end_ida_linear": hex(end),
        "instructions": [instruction(ea) for ea in idautils.Heads(start, end)],
        "inference_level": "confirmed_raw_range",
        "evidence_source": "IDA Pro 9.4 item heads and raw bytes; gaps preserved",
    }


def set_bits(mask):
    return [bit for bit in range(48) if mask[bit // 8] & (0x80 >> (bit % 8))]


def action_value(bit, remap):
    if bit < 16:
        return bit, "direct_bit_0_15"
    if bit < 18:
        return bit + 2, "bit_plus_2"
    return remap[bit - 18], "DGROUP_0x3930_remap"


def table_word(ea):
    return {
        "entry_ida_linear": hex(ea),
        "raw_word_le": hex(ida_bytes.get_word(ea)),
        "bytes_hex": (ida_bytes.get_bytes(ea, 2) or b"").hex(),
        "inference_level": "confirmed_raw_table_entry",
        "evidence_source": "IDA Pro 9.4 data bytes",
    }


def rng_xrefs(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {
            "target_ida_linear": hex(ea),
            "target_function_original": "<no-func>",
            "inference_level": "unknown",
        }
    return {
        "target_ida_linear": hex(fn.start_ea),
        "target_function_original": ida_funcs.get_func_name(fn.start_ea),
        "xrefs_to": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "inference_level": "confirmed_direct_xrefs_only",
        "evidence_source": "IDA Pro 9.4 xref database",
    }


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) != 3:
        raise RuntimeError("需要 OUT_JSON、D3MNS.DAT")
    output_path = Path(idc.ARGV[1])
    monsters_path = Path(idc.ARGV[2])
    exe_path = Path(ida_nalt.get_input_file_path())
    exe = exe_path.read_bytes()
    monsters = monsters_path.read_bytes()
    start = MONSTER_ID * MONSTER_STRIDE
    record = monsters[start:start + MONSTER_STRIDE]
    if len(record) != MONSTER_STRIDE:
        raise RuntimeError("monster46 record truncated")
    mask = record[14:20]
    remap = bytes(ida_bytes.get_bytes(REMAP_IDA, 30) or b"")
    if len(remap) != 30:
        raise RuntimeError("action remap table unavailable")

    actions = []
    for bit in set_bits(mask):
        action, source = action_value(bit, remap)
        entry = None
        handler = None
        if action >= 0x14:
            entry = HANDLERS_IDA + (action - 0x14) * 2
            handler_word = ida_bytes.get_word(entry)
            handler = {
                **table_word(entry),
                "handler_logical": hex(handler_word),
                "handler_ida_linear": hex(handler_word + 0x10000),
                "handler_file": hex(handler_word + 0x1370),
            }
        actions.append({
            "mask_bit": bit,
            "action_raw": hex(action),
            "action_raw_int": action,
            "mapping_source": source,
            "handler_pointer": handler,
            "inference_level": "confirmed_raw_mapping; handler semantics require claim review",
        })

    key_site_records = []
    for ea in KEY_SITES:
        rec = instruction(ea)
        rec["direct_code_refs_from"] = code_refs_from(ea)
        key_site_records.append(rec)

    payload = {
        "schema": "dq3.ida_monster46_action_rng.v1",
        "input": {
            "exe": {"path": str(exe_path), "size": len(exe), "sha256": sha(exe)},
            "monsters": {"path": str(monsters_path), "size": len(monsters), "sha256": sha(monsters)},
        },
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "script": "tools/ida_dump_monster46_action_rng.py",
        },
        "address_contract": {
            "ida_linear": "logical + 0x10000",
            "file_main_load": "logical + 0x1370",
            "dgroup_tables": "IDA linear addresses are recorded explicitly; do not substitute file offsets",
        },
        "monster_record": {
            "id": MONSTER_ID,
            "stride": MONSTER_STRIDE,
            "file_offset": hex(start),
            "raw_hex": record.hex(),
            "action_gate_u8": record[13],
            "action_mask_hex": mask.hex(),
            "action_mask_set_bits_msb_first": set_bits(mask),
            "actions": actions,
            "inference_level": "confirmed_raw_data; action semantics require caller/handler review",
        },
        "tables": {
            "action_remap": {
                "dgroup": "0x3930",
                "ida_linear": hex(REMAP_IDA),
                "raw_hex": remap.hex(),
                "inference_level": "confirmed_raw_table",
            },
            "handler_pointer_table": {
                "dgroup": "0x394e",
                "ida_linear": hex(HANDLERS_IDA),
                "entries_for_monster46_actions": [x["handler_pointer"] for x in actions if x["handler_pointer"]],
                "inference_level": "confirmed_raw_table; indirect call target semantics require caller review",
            },
        },
        "rng_targets": {
            "rng_256_candidate": rng_xrefs(RNG_256_IDA),
            "rng_bounded_candidate": rng_xrefs(RNG_BOUNDED_IDA),
            "inference_level": "confirmed_direct_xrefs_only; indirect calls remain unknown",
        },
        "key_sites": key_site_records,
        "functions": {hex(ea): function_record(ea) for ea in FUNCTIONS},
        "raw_ranges": [raw_range(start, end, label) for start, end, label in RAW_RANGES],
        "claim_boundary": {
            "confirmed": [
                "monster46 mask raw bytes and MSB-first set-bit indices",
                "sub_1A973 direct cast-gate call sites and sub_199DC direct xrefs",
                "sub_19AD6 bounded-selection call site and sub_199DC remap/handler call path",
                "raw handler pointer table entry for action 0x35",
            ],
            "strong": [
                "RNG helper roles when inferred from calling convention and bounded argument",
            ],
            "unknown": [
                "unlisted indirect callers or target selection performed behind unresolved function pointers",
                "whether any single-target RNG occurs in a handler not captured by the listed direct calls",
            ],
        },
    }
    output_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if output_path.stat().st_size == 0:
        raise RuntimeError("empty sidecar")
    idc.qexit(0)


main()
