"""IDA 9.4 NPC圖層consumer的非破壞有界匯出；證據入口為 docs/188。

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
import ida_ua
import ida_xref
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
    ledger_path = Path("/repo/tools/ida_npc_animation_ledger.json")
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
        loaded_value = bytearray(value)
        for relocation in relocation_offsets:
            if off <= relocation < off + len(value):
                if relocation + 2 > off + len(value):
                    raise RuntimeError("reviewed ledger relocation crosses byte range")
                original_word = struct.unpack_from("<H", raw, relocation)[0]
                struct.pack_into("<H", loaded_value, relocation - off, (original_word + loaded_segment) & 0xffff)
        if ida_bytes.get_bytes(ea, len(value)) != loaded_value or annotation["inference_level"] not in ("confirmed", "strong"):
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
            "evidence": "docs/188：2026-10-06 正常422戶外NPC mover候選；未審查項目保持unknown",
        }
        if ea in annotations:
            annotation = annotations[ea]
            result.update({"inference_level": annotation["inference_level"],
                           "added_semantic": annotation["semantic"], "consumer": annotation["consumer"],
                           "evidence": annotation.get("evidence", annotation.get("review_reference", ledger["review_reference"])),
                           "dynamic_evidence": annotation.get("dynamic_evidence", ledger["dynamic_evidence"]),
                           "warning": "" if annotation["inference_level"] == "confirmed"
                           else "⚠ strong：尚未動態閉合，不當作完整parity"})
        return result

    def refs(ea):
        return [{**row(ref.frm), "xref_type": ref.type} for ref in idautils.XrefsTo(ea)]

    ranges=[]
    for requested_ea in (0x12025,0x120aa):
        owner=ida_funcs.get_func(requested_ea)
        if owner is None:raise RuntimeError("IDA function missing: "+hex(requested_ea))
        pair=(owner.start_ea,owner.end_ea)
        if pair not in ranges:ranges.append(pair)
    mover_owner=ida_funcs.get_func(0x12025)
    for ref in idautils.XrefsTo(mover_owner.start_ea):
        if not ref.iscode:continue
        owner=ida_funcs.get_func(ref.frm)
        if owner is not None and (owner.start_ea,owner.end_ea) not in ranges:ranges.append((owner.start_ea,owner.end_ea))
    for ref in idautils.XrefsTo(0x24dd0+0x0b32):
        if ref.type != ida_xref.dr_W:continue
        owner=ida_funcs.get_func(ref.frm)
        if owner is not None and (owner.start_ea,owner.end_ea) not in ranges:ranges.append((owner.start_ea,owner.end_ea))
    literal_candidates=[]
    for start_ea in idautils.Functions():
        func=ida_funcs.get_func(start_ea)
        for ea in idautils.FuncItems(start_ea):
            insn=ida_ua.insn_t()
            if not ida_ua.decode_insn(insn,ea):continue
            if any(op.type==ida_ua.o_mem and op.addr==0x0b32 for op in insn.ops):
                literal_candidates.append(row(ea))
                pair=(func.start_ea,func.end_ea)
                if pair not in ranges:ranges.append(pair)
    sections=[]
    for start,end in ranges:
        rows=[row(ea) for ea in idautils.Heads(start,end)]
        if not rows:raise RuntimeError("empty IDA range")
        sections.append({"ida_linear_start":hex(start),"ida_linear_end_exclusive":hex(end),"instructions":rows,"entry_xrefs":refs(start)})
    decoded_candidates=[]
    operand_candidates=[]
    raw_windows=[]
    for offset,size in ((0x0b32,2),(0x0b64,2),(0x0b66,8),(0x0b28,4),(0x0b35,16),(0x0b5a,2)):
        ea=0x24dd0+offset;off=ida_loader.get_fileregion_offset(ea)
        raw_windows.append({"dgroup_offset":hex(offset),"ida_linear":hex(ea),"file_offset":hex(off),"bytes":raw[off:off+size].hex(),"inference_level":"unknown","xref_to":refs(ea)})
    data={"unresolved_b32_literal_candidates":literal_candidates,"input":{"path":path,"size":len(raw),"sha256":digest},"tool":{"name":"IDA Pro","version":ida_kernwin.get_kernel_version(),"address_space":"IDA linear; MZ file=linear-EC90; DGROUP base linear24DD0","script_sha256":hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},"ranges":sections,"raw_windows":raw_windows,"decoded_candidates":decoded_candidates,"operand_candidates":operand_candidates,"reviewed_ledger":{"sha256":hashlib.sha256(ledger_raw).hexdigest(),"annotations":ledger["annotations"]},"render_candidates":{hex(offset): refs(0x24dd0+offset) for offset in [0x0b32,0x0b64,0x0b66,0x0b28,0x0b35]},"xref_limit":"Direct xrefs exclude indirect accesses; no instruction-mnemonic read/write guessing"}
    with open(idc.ARGV[1],"x",encoding="utf-8") as stream:json.dump(data,stream,ensure_ascii=False,indent=2);stream.write("\n")

if Path(idc.ARGV[1]).exists():idc.qexit(2)
try:main()
except Exception as exc:
    with open(idc.ARGV[1],"x",encoding="utf-8") as stream:json.dump({"error":repr(exc)},stream)
    idc.qexit(1)
else:idc.qexit(0)
