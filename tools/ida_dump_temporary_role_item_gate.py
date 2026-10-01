"""IDA 9.4：非破壞匯出還冠 gate／callee，語意入口為 docs/82。"""

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

ida_auto.auto_wait()
source = Path(ida_nalt.get_input_file_path())
blob = source.read_bytes()
sha = hashlib.sha256(blob).hexdigest()
if len(blob) != 115282 or sha != '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c':
    raise RuntimeError('輸入版本與受審查證據不符')
if len(idc.ARGV) != 2:
    raise RuntimeError('需要唯一 sidecar 輸出路徑')

# 以原始 file offset 為 key；等級只適用 docs/82 審查的局部資料流。
ledger = {
    0x6623: 'handler9 selected item=0x33',
    0x6629: '呼叫原名 sub_1689C 搜尋全隊物品',
    0x662c: '消費搜尋結果 DS:726h',
    0x6649: '間接清除 SI 所指命中槽位，只移除一件',
    0x7c0c: '以 DS:5077h 決定角色迴圈次數',
    0x7c1d: '由原始角色指標表 bx+4F15h 取址',
    0x7c21: '個人道具起點為角色指標+0x3a',
    0x7c26: '遮去裝備高 byte 後比對原始道具 ID',
    0x7c34: '每名角色掃八個 u16 物品欄',
    0x7c40: 'loop 遍歷後續角色，不只首名',
    0x7c42: '未命中 writer：DS:726h=1',
    0x7c49: '命中 writer：DS:726h=0；SI 保留命中槽位',
}

def instruction(ea):
    offset = ea - 0x10000 + 0x1370
    meaning = ledger.get(offset)
    fn = ida_funcs.get_func(ea)
    return {
        'ida_linear': hex(ea), 'file_offset': hex(offset),
        'function_original': ida_funcs.get_func_name(fn.start_ea) if fn else None,
        'bytes_hex': (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b'').hex(),
        'raw_instruction': ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or '',
        'semantic': meaning or 'UNKNOWN：未審查指令，不能視為既定語意',
        'inference_level': 'confirmed' if meaning else 'unknown',
        'evidence': 'docs/82-romaly-king-production-trace.md#2026-10-01-同伴持有皇冠的-gate-勘誤',
        'source_sha256': sha,
    }

start, end = 0x1527e, 0x1531b
targets = set()
for ea in idautils.Heads(start, end):
    if idc.print_insn_mnem(ea).lower() == 'call':
        targets.update(idautils.CodeRefsFrom(ea, False))
callees = []
for target in sorted(targets):
    fn = ida_funcs.get_func(target)
    callees.append({
        'target_ida_linear': hex(target),
        'function_original': ida_funcs.get_func_name(fn.start_ea) if fn else None,
        'start_ida_linear': hex(fn.start_ea) if fn else None,
        'end_ida_linear_exclusive': hex(fn.end_ea) if fn else None,
        'xrefs_to': [{'from_ida_linear': hex(x.frm), 'xref_type': int(x.type)}
                     for x in idautils.XrefsTo(target)],
        'instructions': [instruction(ea) for ea in idautils.FuncItems(fn.start_ea)] if fn else [],
    })
result = {
    'schema': 'dq3.temporary_role_item_gate_evidence.v1',
    'input': {'path': str(source), 'size': len(blob), 'sha256': sha},
    'tool': {'name': 'IDA Pro', 'version': ida_kernwin.get_kernel_version(),
             'address_space': 'IDA linear; logical=linear-0x10000; file=logical+0x1370'},
    'gate_instructions': [instruction(ea) for ea in idautils.Heads(start, end)],
    'callees': callees,
}
output = Path(idc.ARGV[1])
output.write_text(json.dumps(result, ensure_ascii=True, indent=2) + '\n', encoding='utf-8')
if not output.stat().st_size:
    raise RuntimeError('sidecar 輸出為空')
idc.qexit(0)
