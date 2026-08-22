"""非破壞性匯出 DQ3 商店賣出與 DGROUP 0x0b62 selector 證據。"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_kernwin
import ida_lines
import ida_nalt
import ida_segment
import idautils
import idc


def line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

hits = []
for seg_ea in idautils.Segments():
    seg = ida_segment.getseg(seg_ea)
    for ea in idautils.Heads(seg.start_ea, seg.end_ea):
        if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
            continue
        dis = line(ea)
        if "0B62h" not in dis and not (0x1776F <= ea < 0x1793B):
            continue
        hits.append({
            "ida_linear": hex(ea),
            "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
            "disassembly": dis,
            "level": "raw_instruction",
        })

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_shop_sell_evidence.v1",
    "input": {"path": str(input_path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(), "address_space": "linear"},
    "annotation_contract": {"mode": "non_destructive_export", "original_names_preserved": True},
    "instructions": hits,
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
