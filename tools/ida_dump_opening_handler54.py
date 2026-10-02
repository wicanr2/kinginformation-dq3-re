"""IDA 9.4 開場函式與直接 consumer 的非破壞批次匯出。

Usage:
  idat -A -c -o/work/opening.i64 \
    "-S/tools/ida_dump_opening_handler54.py /work/opening-handler54-ida.json 0x147b" \
    /input/DQ3.EXE

The output deliberately retains file offsets, IDA linear addresses, raw bytes,
original IDA names, xrefs, and an explicit inference level.  It does not rename
or otherwise annotate the disposable database.
"""

import hashlib
import json
import os
import sys

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

INPUT_BLOB = b""
INPUT_PATH = ""
REVIEW_LEDGER = [
    (0x10077, 0x100ab, "strong", "創角返回後清畫面、開共用視窗、消費生日文字；只閉合首頁"),
    (0x100ab, 0x100ae, "confirmed", "第18次正常Enter後生日consumer已返回；未追加EOF按鍵，限定出生交易順序"),
    (0x100b5, 0x100c4, "confirmed", "生日返回後才寫DGROUP4F33/4F35=5,5並呼叫11900；房間渲染未對拍"),
    (0x100cd, 0x100d5, "confirmed", "同一次第18次Enter後消費record83並抵達內嵌等待；限自然固定seed創角"),
    (0x15002, 0x15010, "strong", "共用DGROUP3E6E視窗caller；生日首頁record404字模框已逐點閉合"),
    (0x20a07, 0x20a3a, "strong", "四平面清零writer；生日首頁黑底由自然輸入動態閉合，不外推其他caller"),
    (0x214f8, 0x21501, "strong", "一般字模SI+=2、BP+=3，即24px步距；生日首頁已閉合"),
    (0x215b7, 0x215ce, "strong", "0xfff5姓名插值後只越過控制碼本身；下一word不是參數"),
    (0x21514, 0x21530, "strong", "row3的EOF分支捲動後返回，沒有按鍵等待；生日自然caller已閉合"),
    (0x21558, 0x2157e, "strong", "0xfffc推進行並等待，再續寫同一文字流；保留／捲動完整畫面未閉合"),
    (0x21651, 0x216bf, "strong", "姓名consumer保存並還原SI；BP增加名字長度乘3；生日姓名0已閉合"),
    (0x28c3e, 0x28c4a, "confirmed", "raw共用視窗前12bytes及record404字模畫布；限定自然生日首頁，其他欄位／場景未外推"),
    (0x272ed, 0x272ef, "confirmed", "DGROUP251D原始初值1e00即30；自然生日／房間等待觀測亦為30，只限新遊戲初值"),
    (0x11971, 0x11991, "strong", "房間重繪原點為玩家減9、7，範圍20×15；隔離陰影字組試作剩302像素，正式房間尚未閉合"),
    (0x1311a, 0x1311e, "confirmed", "section+0x12經DH寫DGROUP0B2D；自然家中外界圖塊為71，限定原始欄位與寫入"),
    (0x11dd8, 0x11ddc, "strong", "房間視野Y界外且layer為0時讀DGROUP0B2D作圖塊；不外推其他layer"),
    (0x11e47, 0x11e4b, "strong", "房間視野X界外且layer為0時讀DGROUP0B2D作圖塊；不外推其他layer"),
    (0x11ed0, 0x11eee, "strong", "NPC朝向與動畫位元選圖庫指標；第18次自然房間觀測原始索引43／33，初始姿態writer仍未知"),
    (0x1fc57, 0x1fcc6, "strong", "共用視窗偏移8px陰影；ROR AAAA逐行遮罩與16位元AND讀寫，不是清文字內容；限dosgolem字組鎖存契約"),
    (0x21b98, 0x21bac, "strong", "僅設定VGA繪圖暫存器；沒有額外24px清底，不能把此helper稱作文字清底"),
]
CONTINUATION_RANGES = {0x100ab, 0x100b5, 0x100cd, 0x21514, 0x21558}


def review_evidence(start):
    if start in (0x1311a, 0x11dd8, 0x11e47, 0x11ed0, 0x1fc57, 0x21b98):
        return ("docs/188：2026-10-02房間渲染DRAFT；IDA原始bytes／writer-consumer與自然19次IRQ1唯讀觀測；"
                "原版收據SHA-256 b077e99c3b38f58b3326df8f8d25ed72b74d58f351a71d41374bc5063db4358c；"
                "字組陰影隔離試作302像素，非正式完整parity，NPC初始姿態與箭頭仍待閉合")
    if start in (0x272ed, 0x11971):
        return ("docs/188：2026-10-02有限返回／初始clock READY及房間DRAFT；原版冷啟動收據SHA-256 "
                "a23584e6b08ece0ad065f3e26efe217c1d3ad5f96d4eaada9fb4664d1eff3165；原始bytes與唯讀flow；不代表房間CONFORMED")
    if start in CONTINUATION_RANGES:
        return ("docs/188：2026-10-02續頁DRAFT的已審查RE；原版19次IRQ1／38次送達與純讀取flow，"
                "收據SHA-256 f8752bdf2c4e8d333c53b05bd591ccd211d34d6f16c516fde8643db0f06790df；"
                "原始bytes／database xref；不代表production或房間渲染已符合")
    return "docs/188：2026-10-01有限首頁READY；固定seed冷啟動17次IRQ1與完整畫布核對；原始bytes／database xref"


def clean_line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def item_record(ea):
    size = max(1, ida_bytes.get_item_size(ea))
    raw = ida_bytes.get_bytes(ea, size) or b""
    file_offset = ida_loader.get_fileregion_offset(ea)
    refs_from = []
    for dst in idautils.CodeRefsFrom(ea, False):
        refs_from.append({
            "ida_linear": hex(dst),
            "original_name": ida_funcs.get_func_name(dst) or idc.get_name(dst) or "",
        })
    result = {
        "ida_linear": hex(ea),
        "file_offset": hex(file_offset) if file_offset >= 0 else None,
        "bytes": raw.hex(),
        "bytes_basis": "IDA loaded linear；可能包含MZ relocation；原始檔案另列file_bytes",
        "file_bytes": INPUT_BLOB[file_offset:file_offset + size].hex() if file_offset >= 0 else None,
        "disassembly": clean_line(ea),
        "code_refs_from": refs_from,
        "data_refs_from": [{"ida_linear": hex(dst), "original_name": idc.get_name(dst) or ""}
                           for dst in idautils.DataRefsFrom(ea)],
        "source_path": INPUT_PATH,
        "source_size": len(INPUT_BLOB),
        "source_sha256": hashlib.sha256(INPUT_BLOB).hexdigest(),
        "semantic": "UNKNOWN：開場定位線索，尚未審查writer／consumer閉環",
        "inference_level": "unknown",
        "warning": "⚠ unknown：不可將自動名稱或定位線索當成已證實語意",
        "evidence": "本次非破壞IDA database匯出；docs/188、docs/192只供定位",
    }
    for start, end, level, semantic in REVIEW_LEDGER:
        if start <= ea < end:
            result.update(inference_level=level, semantic=semantic,
                          reviewed_range={"ida_linear_start": hex(start), "ida_linear_end_exclusive": hex(end)},
                          evidence=review_evidence(start))
            result["warning"] = "" if level == "confirmed" else "⚠ " + level + "：未達已證實，不能外推未驗收玩家路徑"
            break
    return result


def main():
    global INPUT_BLOB, INPUT_PATH
    ida_auto.auto_wait()
    if len(idc.ARGV) < 3:
        raise RuntimeError("output path and target file offset are required")
    output = idc.ARGV[1]
    target_file = int(idc.ARGV[2], 0)
    input_path = ida_nalt.get_input_file_path()
    with open(input_path, "rb") as fh:
        blob = fh.read()
    INPUT_BLOB, INPUT_PATH = blob, input_path
    if len(blob) != 115282 or hashlib.sha256(blob).hexdigest() != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
        raise RuntimeError("輸入DQ3.EXE大小或雜湊不符，停止匯出")

    target_ea = ida_loader.get_fileregion_ea(target_file)
    if target_ea == idc.BADADDR:
        for ea in idautils.Heads(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
            off = ida_loader.get_fileregion_offset(ea)
            size = max(1, ida_bytes.get_item_size(ea))
            if off >= 0 and off <= target_file < off + size:
                target_ea = ea
                break
    if target_ea == idc.BADADDR:
        raise RuntimeError(f"file offset {target_file:#x} is not mapped")

    fn = ida_funcs.get_func(target_ea)
    if fn is None:
        raise RuntimeError(f"no function contains IDA address {target_ea:#x}")

    callers = []
    for ref in idautils.CodeRefsTo(fn.start_ea, False):
        caller = ida_funcs.get_func(ref)
        callers.append({
            "callsite_ida_linear": hex(ref),
            "callsite_file_offset": hex(ida_loader.get_fileregion_offset(ref)),
            "original_function": ida_funcs.get_func_name(caller.start_ea) if caller else "",
            "disassembly": clean_line(ref),
        })

    selected_heads = list(idautils.FuncItems(fn.start_ea))
    selected_range = None
    if len(idc.ARGV) > 3:
        end_file = int(idc.ARGV[3], 0)
        if not target_file < end_file <= len(blob) or end_file - target_file > 4096:
            raise RuntimeError("額外範圍須為有效且不超過4096bytes的原始檔案區間")
        last_ea = ida_loader.get_fileregion_ea(end_file - 1)
        if last_ea == idc.BADADDR:
            raise RuntimeError("額外範圍末端未映射到IDA位址")
        last_head = ida_bytes.get_item_head(last_ea)
        end_ea = last_head + max(1, ida_bytes.get_item_size(last_head))
        selected_heads = list(idautils.Heads(target_ea, end_ea))
        if not selected_heads:
            raise RuntimeError("額外範圍沒有可匯出的資料庫項目")
        actual_end_file = ida_loader.get_fileregion_offset(last_head) + ida_bytes.get_item_size(last_head)
        selected_range = {"file_start": hex(target_file), "requested_file_end_exclusive": hex(end_file),
                          "file_end_exclusive": hex(actual_end_file),
                          "ida_linear_start": hex(target_ea), "ida_linear_end_exclusive": hex(end_ea),
                          "note": "有界資料庫區間；末端保留完整原始指令，不合併或修改函式邊界"}
    instructions = [item_record(ea) for ea in selected_heads]
    related_starts = set()
    for ea in selected_heads:
        for target in idautils.CodeRefsFrom(ea, False):
            related = ida_funcs.get_func(target)
            if related is not None and related.start_ea != fn.start_ea:
                related_starts.add(related.start_ea)
    related_functions = []
    for start in sorted(related_starts):
        related = ida_funcs.get_func(start)
        related_functions.append({
            **item_record(start),
            "original_function": ida_funcs.get_func_name(start),
            "function_end_ida_linear": hex(related.end_ea),
            "instructions": [item_record(ea) for ea in idautils.FuncItems(start)],
            "xrefs": [{**item_record(ref.frm), "xref_type": ref.type}
                      for ref in idautils.XrefsTo(start)],
        })
    result = {
        "evidence_contract": {
            "semantic_annotation": f"file {target_file:#x}／IDA linear {target_ea:#x} 所在原始函式；僅作導覽",
            "inference_level": "unknown",
            "note": "原始函式導覽標籤；語意須逐項審查caller／writer／consumer，不能把標籤當證據。",
        },
        "input": {
            "path": input_path,
            "size": len(blob),
            "sha256": hashlib.sha256(blob).hexdigest(),
        },
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "export_script_sha256": hashlib.sha256(open(__file__, "rb").read()).hexdigest(),
            "address_space": "IDA linear；MZ file=linear−0xEC90；DGROUP基底linear0x24DD0",
        },
        "target": {
            "requested_file_offset": hex(target_file),
            "ida_linear": hex(target_ea),
            "original_function": ida_funcs.get_func_name(fn.start_ea),
            "function_start_ida_linear": hex(fn.start_ea),
            "function_start_file_offset": hex(ida_loader.get_fileregion_offset(fn.start_ea)),
            "function_end_ida_linear": hex(fn.end_ea),
        },
        "callers": callers,
        "instructions": instructions,
        "selected_range": selected_range,
        "data_xrefs_note": "DS相對運算元與間接讀寫未必產生直接xref；空清單不能證明沒有writer或consumer。",
        "review_ledger": [{"ida_linear_start": hex(a), "ida_linear_end_exclusive": hex(b),
                           "inference_level": level, "semantic": semantic,
                           "evidence": review_evidence(a)}
                          for a, b, level, semantic in REVIEW_LEDGER],
        "related_functions": related_functions,
        "raw_windows": [{**item_record(ea), "length_bytes": 30,
                         "raw_hex": (ida_bytes.get_bytes(ea, 30) or b"").hex(),
                         "xrefs": [{**item_record(ref.frm), "xref_type": ref.type}
                                   for ref in idautils.XrefsTo(ea)]}
                        for ea in (0x28c3e,)],
        "raw_data_refs": [{**item_record(ea),
                           "length_bytes": 2,
                           "raw_hex": (ida_bytes.get_bytes(ea, 2) or b"").hex(),
                           "xrefs": [{**item_record(ref.frm), "xref_type": ref.type,
                                      "original_function": ida_funcs.get_func_name(ref.frm)}
                                     for ref in idautils.XrefsTo(ea)]}
                          for ea in (0x24dd0 + 0x0b2d, 0x272ed)],
    }
    repo_root = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
    result["resolution_backlinks"] = []
    for relative, marker in (
        ("docs/94-dialogue-window-and-monster-mask-re.md", "2026-10-01 勘誤：raw"),
        ("docs/84-game-pack-json-contract.md", "創角後生日旁白（schema 0.1.58"),
    ):
        with open(os.path.join(repo_root, relative), encoding="utf-8") as stream:
            text = stream.read()
        if marker not in text or "188-opening-escort-to-castle-spec.md" not in text:
            raise RuntimeError(relative + " 缺生日旁白勘誤或契約入口")
        result["resolution_backlinks"].append({"older_spec": relative, "required_marker": marker,
                                                "evidence_spec": "docs/188-opening-escort-to-castle-spec.md",
                                                "source_sha256": hashlib.sha256(blob).hexdigest(), "checked": True})
    os.makedirs(os.path.dirname(output), exist_ok=True)
    with open(output, "w", encoding="utf-8") as fh:
        json.dump(result, fh, ensure_ascii=False, indent=2)
        fh.write("\n")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    output = idc.ARGV[1] if len(idc.ARGV) > 1 else "/tmp/opening-handler54-ida-error.json"
    with open(output, "w", encoding="utf-8") as fh:
        json.dump({"error": repr(exc)}, fh, ensure_ascii=False, indent=2)
        fh.write("\n")
    idc.qexit(1)
