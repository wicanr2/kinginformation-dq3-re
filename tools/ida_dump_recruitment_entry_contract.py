"""IDA 9.4 首次招募選單契約的非破壞有界匯出；證據入口為 docs/188。

在唯讀 DQ3.EXE 上建立一次性 database，再以 -S 傳入輸出 JSON 路徑。
保留原始名稱、bytes、xref type 與位址；候選不推定為讀／寫或已證實語意。
"""

import hashlib
import json
from pathlib import Path
import struct

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_loader
import ida_nalt
import idautils
import idc


def main():
    ida_auto.auto_wait()
    path = ida_nalt.get_input_file_path()
    with open(path, "rb") as stream:
        raw = stream.read()
    digest = hashlib.sha256(raw).hexdigest()
    if len(raw) != 115282 or digest != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
        raise RuntimeError("DQ3.EXE identity differs")
    header_size = struct.unpack_from("<H", raw, 8)[0] * 16
    relocation_count = struct.unpack_from("<H", raw, 6)[0]
    relocation_table = struct.unpack_from("<H", raw, 0x18)[0]
    loaded_base = ida_loader.get_fileregion_ea(header_size)
    if loaded_base < 0 or loaded_base % 16:
        raise RuntimeError("MZ loaded base differs")
    loaded_segment = loaded_base // 16
    relocation_offsets = []
    for i in range(relocation_count):
        offset, segment = struct.unpack_from("<HH", raw, relocation_table + i * 4)
        relocation_offsets.append(header_size + segment * 16 + offset)
    ledger_path = Path("/repo/tools/ida_recruitment_entry_ledger.json")
    ledger_raw = ledger_path.read_bytes()
    ledger = json.loads(ledger_raw)
    if ledger["input_size"] != len(raw) or ledger["input_sha256"] != digest:
        raise RuntimeError("reviewed ledger input differs")
    annotations = {}
    for annotation in ledger["annotations"]:
        ea, off = int(annotation["ida_linear"], 16), int(annotation["file_offset"], 16)
        value = bytes.fromhex(annotation["bytes"])
        if ea in annotations or off != ida_loader.get_fileregion_offset(ea) or raw[off:off + len(value)] != value:
            raise RuntimeError("reviewed ledger original bytes differ")
        if ida_bytes.get_bytes(ea, len(value)) != value or annotation["inference_level"] not in ("confirmed", "strong"):
            raise RuntimeError("reviewed ledger database or inference differs")
        annotations[ea] = annotation

    def row(ea):
        off = ida_loader.get_fileregion_offset(ea)
        size = ida_bytes.get_item_size(ea)
        owner = ida_funcs.get_func(ea)
        expected = bytearray(raw[off:off + size]) if off >= 0 else None
        relocations = []
        if expected is not None:
            for relocation in relocation_offsets:
                if off <= relocation < off + size:
                    if relocation + 2 > off + size:
                        raise RuntimeError("MZ relocation crosses IDA item")
                    original_word = struct.unpack_from("<H", raw, relocation)[0]
                    loaded_word = (original_word + loaded_segment) & 0xffff
                    struct.pack_into("<H", expected, relocation - off, loaded_word)
                    relocations.append({"file_offset": hex(relocation), "original_word": hex(original_word),
                                        "loaded_word": hex(loaded_word), "load_segment": hex(loaded_segment)})
            if ida_bytes.get_bytes(ea, size) != expected:
                raise RuntimeError("IDA loaded bytes do not match original MZ relocation")
        result = {
            "ida_linear": hex(ea), "file_offset": hex(off),
            "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "file_bytes": raw[off:off + size].hex() if off >= 0 else None,
            "expected_loaded_bytes": expected.hex() if expected is not None else None,
            "mz_relocations": relocations,
            "disassembly": ida_lines.tag_remove(ida_lines.generate_disasm_line(ea, 0) or ""),
            "original_name": idc.get_name(ea),
            "original_function": ida_funcs.get_func_name(owner.start_ea) if owner else None,
            "function_start": hex(owner.start_ea) if owner else None,
            "function_end": hex(owner.end_ea) if owner else None,
            "ds_register": hex(idc.get_sreg(ea, "ds")),
            "code_refs_from": [hex(v) for v in idautils.CodeRefsFrom(ea, False)],
            "data_refs_from": [hex(v) for v in idautils.DataRefsFrom(ea)],
            "source_path": path, "source_size": len(raw), "source_sha256": digest,
            "inference_level": "unknown",
            "warning": "⚠ unknown：候選定位，尚未閉合動態writer／consumer",
            "evidence": "docs/188：2026-10-04 首次招募consumer；未審查項目保持unknown",
        }
        if ea in annotations:
            annotation = annotations[ea]
            result.update({"inference_level": annotation["inference_level"],
                           "added_semantic": annotation["semantic"], "consumer": annotation["consumer"],
                           "evidence": annotation.get("review_reference", ledger["review_reference"]),
                           "dynamic_evidence": annotation.get("dynamic_evidence", ledger["dynamic_evidence"]),
                           "warning": "" if annotation["inference_level"] == "confirmed"
                           else "⚠ strong：尚未動態閉合，不當作完整parity"})
        return result

    def refs(ea):
        return [{**row(ref.frm), "xref_type": ref.type} for ref in idautils.XrefsTo(ea)]

    ranges = [(0x1036d,0x103d7),(0x15002,0x15038),(0x1f4e3,0x1fe11),(0x21414,0x21651)]
    sections = []
    for start, end in ranges:
        rows = [row(ea) for ea in idautils.Heads(start, end)]
        if not rows:
            raise RuntimeError("empty IDA range")
        sections.append({"ida_linear_start": hex(start), "ida_linear_end_exclusive": hex(end),
                         "instructions": rows, "entry_xrefs": refs(start)})
    candidates = []  # 本切片不掃描未接到玩家入口的操作數候選。
    window_fields = []
    for offset in (0x3e32,0x3e6e):
        ea=0x24dd0+offset
        window_fields.append({"dgroup_offset":hex(offset),"ida_linear":hex(ea),"file_offset":hex(ea-0xec90),"file_bytes":raw[ea-0xec90:ea-0xec90+32].hex(),"field_xrefs":[{"offset":hex(i),"refs":refs(ea+i)} for i in range(0,24,2)]})
    data = {
        "input": {"path": path, "size": len(raw), "sha256": digest},
        "mz_load": {"header_size": header_size, "load_segment": hex(loaded_segment),
                    "relocation_count": relocation_count},
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script_sha256": hashlib.sha256(open(__file__, "rb").read()).hexdigest(),
                 "address_space": "IDA linear；MZ file=linear−0xEC90；DGROUP基底linear0x24DD0"},
        "ranges": sections, "recruitment_entry_window_fields":window_fields,
        "raw_window_3e9c": {"ida_linear":hex(0x28C6C), "file_offset":hex(0x28C6C-0xEC90), "file_bytes":raw[0x28C6C-0xEC90:0x28C6C-0xEC90+20].hex(), "xref_limit":"indirect SI consumer traced separately"},
        "operand_candidates": candidates,
        "reviewed_ledger": {"path": str(ledger_path), "sha256": hashlib.sha256(ledger_raw).hexdigest(),
                            "annotations": ledger["annotations"]},
        "xref_limit": "直接xref不涵蓋暫存器／指標的間接讀寫；operand候選未依位置或助憶碼分類。",
    }
    with open(idc.ARGV[1], "x", encoding="utf-8") as stream:
        json.dump(data, stream, ensure_ascii=False, indent=2)
        stream.write("\n")


if Path(idc.ARGV[1]).exists():
    idc.qexit(2)  # 已存在的輸出保持不變。

try:
    main()
except Exception as exc:
    with open(idc.ARGV[1], "x", encoding="utf-8") as stream:
        json.dump({"error": repr(exc)}, stream)
    idc.qexit(1)
else:
    idc.qexit(0)
