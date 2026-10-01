"""IDA 9.4 非破壞匯出：開場檔名 xref、caller 與原始指令；語意預設 unknown。"""
import hashlib
import json
import os

import ida_auto
import ida_ida
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_loader
import ida_nalt
import idautils
import idc


ida_auto.auto_wait()
path = ida_nalt.get_input_file_path()
with open(path, 'rb') as stream:
    blob = stream.read()
sha = hashlib.sha256(blob).hexdigest()
if len(blob) != 115282 or sha != '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c':
    raise RuntimeError('輸入版本與受版控語意索引不符，停止匯出')
# 原始 file offset 作 key，附加語意不覆寫原始名稱／指令。
# 等級只涵蓋這裡寫明的語意，不外推整個函式或尚未對拍的按鍵／音訊。
ledger = {
    0x13460: ('F 背景→DFP loader→標誌動畫→G 標題的 caller', 'confirmed'),
    0x13512: ('標誌 y=348..92，每步 -2，共129位置，每位置等待2 ticks', 'confirmed'),
    0x135b6: ('第七圖塊逐列四平面、色號零仍覆寫、底部裁切', 'confirmed'),
    0x13634: ('六圖塊按 byte 交錯四平面、色號零透明、底部裁切', 'confirmed'),
    0x13845: ('顯示頁切換完成的 RET 觀測點；原始 return stack 可區分動畫 caller', 'confirmed'),
    0x13521: ('寫入起始基準 y=348', 'confirmed'),
    0x13533: ('讀取 DGROUP 5BDE 的 y_offset', 'confirmed'),
    0x1353e: ('讀取 DGROUP 5BE0 的 x_byte', 'confirmed'),
    0x13559: ('第七圖塊 y_offset=33', 'confirmed'),
    0x1355f: ('第七圖塊 x_byte=9', 'confirmed'),
    0x1356b: ('本位置翻頁後等待2 ticks', 'confirmed'),
    0x1358d: ('下個位置基準 y 減2', 'confirmed'),
    0x13592: ('減到90停止；最後顯示基準 y=92', 'confirmed'),
}

def annotation(offset):
    meaning, level = ledger.get(offset, ('待審查原始指令', 'unknown'))
    return {'semantic': meaning, 'inference_level': level,
            'evidence': ('docs/196-dosgolem-opening-sequence-parity.md；IDA bytes/caller/consumer＋dosgolem129次自然翻頁'
                         if level == 'confirmed' else 'UNKNOWN：未審查；不可視為既定語意'),
            'source': sha}
functions = set()
anchors = []
for pattern in [b'TITA', b'TITB', b'TITC', b'TITD', b'TITE', b'TITF', b'DFP', b'TITG']:
    offset = 0
    while True:
        offset = blob.find(pattern, offset)
        if offset < 0:
            break
        ea = ida_loader.get_fileregion_ea(offset)
        refs = []
        if ea != idc.BADADDR:
            for delta in range(-16, len(pattern)):
                for ref in idautils.XrefsTo(ea + delta):
                    fn = ida_funcs.get_func(ref.frm)
                    if fn:
                        functions.add(fn.start_ea)
                    refs.append({'ida_linear': hex(ref.frm),
                                 'file_offset': hex(ida_loader.get_fileregion_offset(ref.frm)),
                                 'xref_type': ref.type, 'original_function': ida_funcs.get_func_name(ref.frm),
                                 'disassembly': ida_lines.generate_disasm_line(ref.frm, ida_lines.GENDSM_REMOVE_TAGS)})
        anchors.append({'pattern': pattern.decode(), 'file_offset': hex(offset),
                        'ida_linear': hex(ea), 'bytes': blob[offset:offset+24].hex(),
                        'inference_level': 'unknown', 'source': sha, 'xrefs': refs})
        offset += len(pattern)
# Turbo Pascal 檔名有長度前綴，也可能經 DS 相對位移取址，字串本身沒有 direct xref。
# 此掃描只提供數值線索，仍保留 unknown；由下面的資料庫函式／caller 關係匯出。
operand_anchors = []
for ea in idautils.Heads(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
    if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
        continue
    for op in range(2):
        if idc.get_operand_type(ea, op) not in (idc.o_imm, idc.o_mem, idc.o_displ):
            continue
        if idc.get_operand_value(ea, op) not in (0x5af4, 0x5af6):
            continue
        fn = ida_funcs.get_func(ea)
        if fn:
            functions.add(fn.start_ea)
        operand_anchors.append({'ida_linear': hex(ea), 'file_offset': hex(ida_loader.get_fileregion_offset(ea)),
                                'original_operand': idc.print_operand(ea, op), 'inference_level': 'unknown',
                                'source': sha})
# 已匯出 caller sub_220F0 的直接呼叫：只追標誌初始化、動畫與共用等待。
for ea in (0x221A2, 0x2222A, 0x22631, 0x224C4, 0x227A7, 0x227BC):
    fn = ida_funcs.get_func(ea)
    if fn:
        functions.add(fn.start_ea)
# 呼叫者與一層被呼叫者由資料庫圖取得；限制為命中檔名的子系統。
roots = set(functions)
for start in roots:
    for ref in idautils.CodeRefsTo(start, False):
        fn = ida_funcs.get_func(ref)
        if fn:
            functions.add(fn.start_ea)
    for ea in idautils.FuncItems(start):
        for dst in idautils.CodeRefsFrom(ea, False):
            fn = ida_funcs.get_func(dst)
            if fn:
                functions.add(fn.start_ea)
items = []
for start in sorted(functions):
    fn = ida_funcs.get_func(start)
    instructions = []
    for ea in idautils.FuncItems(start):
        size = ida_bytes.get_item_size(ea)
        original_offset = ida_loader.get_fileregion_offset(ea)
        instructions.append({'ida_linear': hex(ea),
                             'file_offset': hex(ida_loader.get_fileregion_offset(ea)),
                             'bytes': (ida_bytes.get_bytes(ea, size) or b'').hex(),
                             'disassembly': ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS),
                             **annotation(original_offset)})
    items.append({'original_function': ida_funcs.get_func_name(start), 'ida_linear': hex(start),
                  'file_offset': hex(ida_loader.get_fileregion_offset(start)),
                  'end_ida_linear': hex(fn.end_ea),
                  **annotation(ida_loader.get_fileregion_offset(start)), 'instructions': instructions})
result = {'input': {'path': path, 'size': len(blob), 'sha256': sha},
          'tool': {'name': 'IDA Pro', 'version': ida_kernwin.get_kernel_version(),
                   'address_space': 'IDA linear 與原始 MZ file offset；各欄明示，未轉為 logical'},
          'warning': '原始名稱／位址／bytes 保留；附加語意自動合併受版控 file-offset ledger。UNKNOWN 項未審查，不可當事實；未改寫 database。',
          'function_count': len(list(idautils.Functions())), 'anchors': anchors,
          'operand_anchors': operand_anchors, 'functions': items}
with open(idc.ARGV[1] + '.tmp', 'w', encoding='utf-8') as stream:
    json.dump(result, stream, ensure_ascii=False, indent=2)
    stream.write('\n')
os.replace(idc.ARGV[1] + '.tmp', idc.ARGV[1])
idc.qexit(0)
