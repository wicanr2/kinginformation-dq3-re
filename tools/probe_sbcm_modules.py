"""Docker-only original SBCM OMF module navigation; indexed in docs/25.

Relocation-excluded searches are candidates only. They cannot establish
matching source coverage, whole relocations, low-level eligibility or the
compiler of the primary program. Retain original module/public/IDA names.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import struct

from omf_matching_probe import Reader, UnsupportedOMF, read_object


ROOT = Path(__file__).resolve().parents[1]
EXE_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
LIB_HASH = "01b242cb99193d006e23b73babe122887e59713410f88798283a9120df1d0683"


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def modules(raw):
    if raw[0] != 0xf0:
        raise ValueError("Not an OMF library header")
    page = struct.unpack_from("<H", raw, 1)[0] + 3
    dictionary, blocks = struct.unpack_from("<IH", raw, 3)
    if page < 16 or page > 32768 or page & (page - 1) or dictionary + blocks * 512 != len(raw):
        raise ValueError("Library page/dictionary contract differs")
    offset = page
    result = []
    while offset < dictionary:
        kind, length = struct.unpack_from("<BH", raw, offset)
        if kind == 0xf1:
            if offset + 3 + length != dictionary:
                raise ValueError("LIBEND does not close the module area")
            return result, {"page_bytes": page, "dictionary_file_offset": hex(dictionary), "dictionary_blocks": blocks,
                            "LIBEND_file_offset": hex(offset), "dictionary_sha256": sha(raw[dictionary:])}
        if kind != 0x80:
            raise ValueError("Expected aligned THEADR at " + hex(offset))
        start = offset
        records = []
        while offset < dictionary:
            kind, length = struct.unpack_from("<BH", raw, offset)
            if length < 1 or offset + 3 + length > dictionary:
                raise ValueError("Invalid library module record range")
            record = raw[offset:offset + 3 + length]
            if record[-1] and sum(record) & 255:
                raise ValueError("Library module record checksum differs")
            records.append({"library_file_offset": hex(offset), "kind": hex(kind), "length": length,
                            "checksum": "omitted" if record[-1] == 0 else "verified"})
            offset += len(record)
            if kind == 0x8a:
                break
        else:
            raise ValueError("Module lacks supported 16-bit MODEND")
        name_length = raw[start + 3]
        name = raw[start + 4:start + 4 + name_length].decode("ascii")
        if len(name) != name_length or 1 + name_length >= struct.unpack_from("<H", raw, start + 1)[0]:
            raise ValueError("Invalid THEADR module name")
        result.append({"original_module_name": name, "library_file_start": hex(start), "size": offset - start,
                       "sha256": sha(raw[start:offset]), "records": records, "raw": raw[start:offset]})
        aligned = (offset + page - 1) // page * page
        if any(raw[offset:aligned]):
            raise ValueError("Nonzero library alignment padding")
        offset = aligned
    raise ValueError("Library lacks LIBEND")


def candidates(raw, code, fixups):
    widths = {0: 1, 1: 2, 2: 2, 3: 4, 4: 1}
    excluded = set()
    for fix in fixups:
        if fix["location_type"] not in widths:
            raise UnsupportedOMF("Unknown relocation field width")
        affected = set(range(fix["offset"], fix["offset"] + widths[fix["location_type"]]))
        if max(affected) >= len(code) or excluded & affected:
            raise ValueError("Relocation fields cross/overlap the code segment")
        excluded |= affected
    runs = []
    start = 0
    for offset in range(len(code) + 1):
        if offset == len(code) or offset in excluded:
            if start < offset:
                runs.append((start, offset))
            start = offset + 1
    if not runs:
        return [], len(excluded)
    begin, end = max(runs, key=lambda run: run[1] - run[0])
    if end - begin < 12:
        return [], len(excluded)
    anchor = code[begin:end]
    cursor = 0
    hits = []
    while True:
        found = raw.find(anchor, cursor)
        if found < 0:
            break
        cursor = found + 1
        location = found - begin
        if location < 0x1370 or location + len(code) > len(raw):
            continue
        if all(raw[location + lo:location + hi] == code[lo:hi] for lo, hi in runs):
            hits.append(location)
            if len(hits) > 128:
                raise ValueError("Ambiguous candidate count exceeds bound")
    return hits, len(excluded)


def decode_lidata(raw):
    """Decode only USE16 LIDATA without subsequent relocation on repeated data.

    The original object stays intact. Re-encoded LEDATA is a parser view,
    never source recovery. Retain an original-record map for all FIXUPP.
    """
    offset = 0
    view = bytearray()
    origins = {}
    decoded = []
    last_was_lidata = False

    def block(reader, depth=0):
        if depth > 16:
            raise ValueError("LIDATA nesting exceeds bound")
        repeat, count = reader.word(), reader.word()
        content = b"".join(block(reader, depth + 1) for _ in range(count)) if count else reader.take(reader.byte())
        if len(content) * repeat > 65536:
            raise ValueError("LIDATA expansion exceeds USE16 segment")
        return content * repeat

    while offset < len(raw):
        kind, length = struct.unpack_from("<BH", raw, offset)
        record = raw[offset:offset + 3 + length]
        if len(record) != length + 3 or length < 1 or record[-1] and sum(record) & 255:
            raise ValueError("Invalid original LIDATA object record")
        if kind == 0x9c and last_was_lidata:
            raise UnsupportedOMF("FIXUPP on repeated LIDATA remains unsupported")
        origins[len(view)] = offset
        if kind == 0xa2:
            reader = Reader(record[3:-1])
            segment_index, segment_offset = reader.index(), reader.word()
            encoded_header_size = reader.position
            header = record[3:3 + encoded_header_size]
            content = b""
            while reader.more():
                content += block(reader)
                if len(content) > 65536:
                    raise ValueError("Combined LIDATA expansion exceeds USE16 segment")
            payload = header + content
            if len(payload) + 1 > 65535:
                raise UnsupportedOMF("Decoded LEDATA record exceeds 16-bit record size")
            converted = bytes([0xa0]) + struct.pack("<H", len(payload) + 1) + payload
            converted += bytes([-sum(converted) & 255])
            view.extend(converted)
            decoded.append({"original_record_file_offset": hex(offset), "segment_index": segment_index,
                            "segment_offset": hex(segment_offset), "encoded_bytes": length - 1 - encoded_header_size,
                            "decoded_bytes": len(content), "decoded_sha256": sha(content)})
            last_was_lidata = True
        else:
            view.extend(record)
            if kind == 0xa0:
                last_was_lidata = False
        offset += len(record)
    return bytes(view), origins, decoded


def resolve_self_segment(code, fixes, segment_index, file_offset):
    """Resolve actual same-private-segment OFFSET16 references only.

    Original SDK fixups with other frames/targets or nonzero encoded addends
    stay unresolved. No relocation-excluded comparison establishes equality.
    """
    linked = bytearray(code)
    results = []
    for fix in fixes:
        if (fix["location_type"] != 1 or fix["target"] != {"method": 0, "datum": segment_index}
                or fix["frame"] != {"method": 0, "datum": segment_index}):
            raise UnsupportedOMF("Non-self OFFSET16 fixup requires independent symbol/frame evidence")
        location = fix["offset"]
        if not 0 <= location or location + 2 > len(linked):
            raise ValueError("Self fixup crosses the original code segment")
        addend = struct.unpack_from("<H", linked, location)[0]
        if addend:
            raise UnsupportedOMF("Nonzero SDK encoded addend requires a separate producer contract")
        value = fix["displacement"]
        if fix["segment_relative"]:
            if not 0 <= value <= 65535:
                raise ValueError("Self-segment OFFSET16 overflows")
        else:
            value -= location + 2
            if not -32768 <= value <= 32767:
                raise ValueError("Self-relative displacement exceeds supported signed16")
        struct.pack_into("<H", linked, location, value & 65535)
        results.append({**fix, "original_addend": addend, "actual_value": value,
                        "encoded_operand": hex(value & 65535), "EXE_operand_file_offset": hex(file_offset + location)})
    return bytes(linked), results


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--ida-inventory", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    exe_path, lib_path = ROOT / "assets_raw/DQ3.EXE", ROOT / "assets_raw/SBCM.LIB"
    raw, library = exe_path.read_bytes(), lib_path.read_bytes()
    if len(raw) != 115282 or sha(raw) != EXE_HASH or len(library) != 35840 or sha(library) != LIB_HASH:
        raise ValueError("Original EXE/library input differs")
    inventory = json.loads(args.ida_inventory.read_text())
    if inventory["input"]["sha256"] != EXE_HASH or inventory["tool"]["version"] != "9.4":
        raise ValueError("IDA identity differs")
    args.output.mkdir(exist_ok=False)
    parsed, header = modules(library)
    results = []
    for index, module in enumerate(parsed):
        obj_path = args.output / ("m" + format(index, "03d") + ".obj")
        obj_path.write_bytes(module.pop("raw"))
        module["artifact_name"] = obj_path.name
        try:
            view, record_origins, decoded = decode_lidata(obj_path.read_bytes())
            view_path = args.output / (obj_path.stem + "-parser-view.obj")
            view_path.write_bytes(view)
            obj = read_object(view_path)
            for fix in obj["fixups"]:
                fix["record_file_offset"] = record_origins[fix["record_file_offset"]]
        except (ValueError, UnsupportedOMF) as error:
            module.update(status="REFUSED", reason=str(error))
        else:
            module.update(status="PARSED", original_publics=obj["publics"], original_externals=obj["externals"],
                          decoded_LIDATA=decoded, parser_view_sha256=sha(view), segments=[])
            for n, segment in enumerate(obj["segments"][1:], 1):
                row = {"segment_index": n, "original_segment_name": segment["name"], "class": segment["class"],
                       "size": segment["length"], "written_bytes": sum(segment["written"]), "sha256": sha(segment["bytes"])}
                if segment["class"] == "CODE" and segment["length"] and all(segment["written"]):
                    fixes = [fix for fix in obj["fixups"] if fix["segment_index"] == n]
                    hits, excluded = candidates(raw, bytes(segment["bytes"]), fixes)
                    row.update(fixups=fixes, relocation_bytes_excluded=excluded, candidates=[])
                    for offset in hits:
                        linear = offset - 0x1370 + 0x10000
                        publics = []
                        for pub in obj["publics"]:
                            if pub["segment_index"] != n:
                                continue
                            ea = linear + pub["offset"]
                            owner = next((f for f in inventory["functions"] if f["ida_linear_start"] == hex(ea)), None)
                            publics.append({"library_public": pub, "original_IDA_name": owner["original_name"] if owner else None,
                                            "ida_linear": hex(ea), "logical": hex(ea - 0x10000), "file_offset": hex(offset + pub["offset"])})
                        candidate = {"file_start": hex(offset), "ida_linear_start": hex(linear), "logical_start": hex(linear - 0x10000),
                                     "publics": publics, "inference_level": "hypothesis relocation-excluded module candidate",
                                     "warning": "Actual fixup resolution, ownership and low-level eligibility not verified"}
                        try:
                            linked, applied = resolve_self_segment(bytes(segment["bytes"]), fixes, n, offset)
                        except UnsupportedOMF as error:
                            candidate.update(whole_segment_equal=False, fixup_status="UNRESOLVED", reason=str(error))
                        else:
                            equal = linked == raw[offset:offset + len(linked)]
                            candidate.update(whole_segment_equal=equal, fixup_status="RESOLVED", actual_fixups=applied,
                                             resolved_segment_sha256=sha(linked))
                            if equal:
                                candidate.update(inference_level="confirmed complete linked-segment byte identity",
                                                 warning="Metadata ownership only; low-level eligibility and readable source still require review")
                        row["candidates"].append(candidate)
                module["segments"].append(row)
        results.append(module)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__).read_bytes()),
               "OMF_parser_sha256": sha((ROOT / "tools/omf_matching_probe.py").read_bytes()),
               "input": {"path": "assets_raw/SBCM.LIB", "size": len(library), "sha256": LIB_HASH},
               "original_EXE_sha256": EXE_HASH, "IDA_inventory_sha256": sha(args.ida_inventory.read_bytes()),
               "library_header": header, "modules": results, "source_coverage_increment": 0,
               "scope": "Original OMF metadata and relocation-excluded navigation only; no C/ASM source completion"}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"modules": len(results), "REFUSED": [r["original_module_name"] for r in results if r["status"] == "REFUSED"],
                      "candidates": [(r["original_module_name"], s["size"], [c["ida_linear_start"] for c in s.get("candidates", [])])
                                     for r in results for s in r.get("segments", []) if s.get("candidates")]}))


if __name__ == "__main__":
    main()
