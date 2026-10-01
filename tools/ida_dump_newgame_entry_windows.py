"""IDA 9.4 非破壞匯出：新遊戲視窗與輸入 writer；未審查語意醒目標為 unknown。"""
import hashlib
import json
import os

import ida_auto
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
with open(path, "rb") as stream:
    blob = stream.read()
sha = hashlib.sha256(blob).hexdigest()
if len(blob) != 115282 or sha != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
    raise RuntimeError("原版輸入版本不符，停止匯出")

# 新語意須審查後逐筆加入；索引不從歷史散文自動推導。
ledger = {
    0x19E8E: {
        "semantic": "主選單 raw window：flags3、x_byte28/y150/width_byte22/height64、record475、cursor_byte30/y166；保留 bytes",
        "inference_level": "confirmed", "evidence": "IDA 10006→1F4E3／1F908 + docs/113；原版冷啟動與主選單試作全畫布 RGB=0",
    },
    0x10006 - 0xEC90: {
        "semantic": "主選單 raw window 的 DS:3D4E 取址入口；保留原始 lea si 運算元",
        "inference_level": "confirmed", "evidence": "IDA entry_range + docs/113；冷啟動 IRQ1 Enter 到主選單",
    },
    0x1FC57 - 0xEC90: {
        "semantic": "raw x+1 byte/y+8；逐列旋轉 0xAAAA，四平面 word AND 視窗陰影",
        "inference_level": "confirmed", "evidence": "IDA writer_range 1FC57..1FCC5 + docs/113；主選單／初始命名試作全畫布 RGB=0",
    },
    0x1FDB1 - 0xEC90: {
        "semantic": "四邊帶 XOR；1FDE4 每列旋轉，平面回跳 1FDE6 不再旋轉",
        "inference_level": "confirmed", "evidence": "IDA 1FD30..1FE10 + docs/113；主選單試作全畫布 RGB=0",
    },
    0x1FD42 - 0xEC90: {
        "semantic": "DS:727 為框線 XOR 平面選擇 consumer；本次主選單執行期值5，初始化 writer 未確認",
        "inference_level": "confirmed", "evidence": "IDA 原始 mov bl,ds:727h + docs/113；原版顏色及全畫布試作，不外推其他階段初值",
    },
    0x211B6 - 0xEC90: {
        "semantic": "16x16 不透明字模 writer；前景 DS:258F，未選平面寫零（palette writer 仍未知）",
        "inference_level": "confirmed", "evidence": "IDA 211B6..2121B 與 21B98 GC 契約 + docs/113；record475/451/452/456 完整畫布驗證",
    },
    0x211DA - 0xEC90: {
        "semantic": "DS:258F 為字模前景色 consumer；本次主選單／初始命名執行期值8，palette 初始化 writer 未確認",
        "inference_level": "confirmed", "evidence": "IDA 原始 mov bl,ds:258Fh + docs/113；原版 glyph/palette 觀察及全畫布試作",
    },
    0x1126F - 0xEC90: {
        "semantic": "命名游標 XOR 平面8/4，共兩 byte、15列；保留原始 AND es:[di],al",
        "inference_level": "confirmed", "evidence": "IDA 10E55/11087→1123C→1126F + docs/113；初始命名的姓名／字盤游標全畫布驗證",
    },
}

# 導航與模式交易逐項審查；只對 docs/113 限定的 45 格命名盤成立。
for ea, semantic in (
    (0x11087, "英數／注音共用導航入口；保留原始 sub_11087 與 caller xref"),
    (0x110C9, "上移減欄數 DGROUP2702，負值加全格數 DGROUP2704"),
    (0x110EC, "下移加欄數 DGROUP2702，達全格數後減 DGROUP2704"),
    (0x11110, "左移 dec cx；全格數循環，raw0 左移為 raw44，不限制同列"),
    (0x11134, "右移 inc cx；全格數循環，raw44 右移為 raw0，不限制同列"),
    (0x10F86, "注音盤欄數 writer：DGROUP2702=9；原始 word_274D2"),
    (0x10F8C, "注音盤全格數 writer：DGROUP2704=45；原始 word_274D4"),
    (0x10DE8, "功能第一列由注音切英數：移除模式 bit1，未寫 raw 游標"),
    (0x10DED, "功能第一列加入英數 bit2；原版實際 raw35 保留"),
    (0x10E4F, "功能交易移除焦點 bit4；原版 mode5 轉 mode2"),
    (0x10EB0, "英數 writer 畫 record453 後呼叫共用導航；既有 raw 游標保留"),
    (0x10F80, "注音 writer 初始化 raw 游標為0；反向切換沿用此既有行為"),
):
    ledger[ea - 0xEC90] = {
        "semantic": semantic, "inference_level": "confirmed",
        "evidence": "IDA database caller/writer/consumer + docs/113 READY；issue4-name/mode 冷啟動 IRQ1 狀態／PNG 收據；注音重設只作原始 writer 斷言",
    }

def row(ea):
    offset = ida_loader.get_fileregion_offset(ea)
    annotation = ledger.get(offset, {
        "semantic": "UNKNOWN：待審查，不可視為既定語意",
        "inference_level": "unknown",
        "evidence": "docs/113-newgame-geometry-re.md；Issue #4 尚在證據審查",
    })
    return {
        "ida_linear": hex(ea), "file_offset": hex(offset),
        "original_name": idc.get_name(ea),
        "bytes": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
        "disassembly": ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS),
        "source_path": path, "source_size": len(blob), "source_sha256": sha,
        "original_size": ida_bytes.get_item_size(ea), **annotation,
    }

functions = set()
anchors = []
for ea in (0x28B1E, 0x2735F, 0x254F7):
    refs = []
    for ref in idautils.XrefsTo(ea):
        refs.append({**row(ref.frm), "xref_type": ref.type})
        fn = ida_funcs.get_func(ref.frm)
        if fn is not None:
            functions.add(fn.start_ea)
    anchors.append({**row(ea), "xrefs": refs})
for ea in (0x1000A, 0x10030, 0x10624, 0x10854, 0x10D17, 0x10DC8, 0x10E55, 0x10E8E, 0x10F5B,
           0x11087, 0x11172, 0x1123C, 0x1126F, 0x1F4E3, 0x1F590, 0x1F63C,
           0x1F779, 0x1FB36, 0x1FC57, 0x1FD30, 0x1FDB1,
           0x1FCE1, 0x1F908, 0x211B6, 0x21286, 0x213C4, 0x21B98):
    fn = ida_funcs.get_func(ea)
    if fn is not None:
        functions.add(fn.start_ea)
items = []
for start in sorted(functions):
    fn = ida_funcs.get_func(start)
    refs = [{**row(ref.frm), "xref_type": ref.type,
             "original_function": ida_funcs.get_func_name(ref.frm)}
            for ref in idautils.XrefsTo(start)]
    items.append({**row(start), "original_function": ida_funcs.get_func_name(start),
                  "end_ida_linear": hex(fn.end_ea), "xrefs": refs,
                  "instructions": [row(ea) for ea in idautils.FuncItems(start)]})
data = []
for ea in range(0x28A00, 0x29100, 16):
    offset = ida_loader.get_fileregion_offset(ea)
    data.append({"ida_linear": hex(ea), "file_offset": hex(offset),
                 "bytes": (ida_bytes.get_bytes(ea, 16) or b"").hex(),
                 "source_path": path, "source_size": len(blob), "source_sha256": sha,
                 "original_size": 16, "semantic": "UNKNOWN：原始視窗候選資料，待 writer 閉合",
                 "inference_level": "unknown", "evidence": "本次 IDA 原始 bytes 匯出"})
result = {
    "input": {"path": path, "size": len(blob), "sha256": sha},
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
             "address_space": "IDA linear／原始 MZ file offset，各欄分開標示"},
    "warning": "不改名、不改指令、不覆寫原版；UNKNOWN 不是確定語意。間接寫入不由直接 xref 清單排除。",
    "functions": items, "data": data, "anchors": anchors,
    "entry_range": [row(ea) for ea in idautils.Heads(0x10000, 0x10030)],
    # IDA 自動函式邊界在舊 16-bit 共用 frame writer 有分裂；保留資料庫
    # 分類與原始歸屬，不擅自合併函式，只另匯出原始區間。
    "writer_range": [row(ea) for ea in idautils.Heads(0x1FB36, 0x1FCE1)
                     if ida_bytes.is_code(ida_bytes.get_flags(ea))],
}
with open(idc.ARGV[1] + ".tmp", "w", encoding="utf-8") as stream:
    json.dump(result, stream, ensure_ascii=False, indent=2)
    stream.write("\n")
os.replace(idc.ARGV[1] + ".tmp", idc.ARGV[1])
idc.qexit(0)
