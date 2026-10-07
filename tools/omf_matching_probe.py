"""Bounded OMF reader for the single-function matching experiment in docs/25.

Format source: TIS OMF 1.1, https://openwatcom.org/ftp/devel/docs/omf.pdf.
Select code by PUBDEF and SEGDEF. Resolve only explicit 16-bit, segment-relative,
external-symbol FIXUPP references with frame method F5 and a supplied DS offset.
Other relocations are refused; this module is not a general linker.
"""

import hashlib
from pathlib import Path
import struct


class UnsupportedOMF(ValueError):
    """The input exceeds the experiment's explicitly supported OMF subset."""


class Reader:
    def __init__(self, data):
        self.data = data
        self.position = 0

    def take(self, count):
        end = self.position + count
        if end > len(self.data):
            raise ValueError("Truncated OMF field")
        result = self.data[self.position:end]
        self.position = end
        return result

    def byte(self):
        return self.take(1)[0]

    def word(self):
        return struct.unpack("<H", self.take(2))[0]

    def index(self):
        first = self.byte()
        return ((first & 0x7F) << 8) | self.byte() if first & 0x80 else first

    def name(self):
        return self.take(self.byte()).decode("ascii")

    def communal_size(self):
        first = self.byte()
        if first < 0x80:
            return first
        widths = {0x81: 2, 0x84: 3, 0x88: 4}
        if first not in widths:
            raise UnsupportedOMF("Unsupported COMDEF numeric leaf")
        return int.from_bytes(self.take(widths[first]), "little")

    def more(self):
        return self.position < len(self.data)


def read_object(path):
    """Preserve segment, public-symbol, external and relocation identities."""
    raw = Path(path).read_bytes()
    cursor = Reader(raw)
    names = [None]
    segments = [None]
    externals = [None]
    publics = []
    fixups = []
    frames = {}
    targets = {}
    last_data = None
    records = []
    ended = False
    while cursor.more():
        record_start = cursor.position
        kind = cursor.byte()
        size = cursor.word()
        if size < 1:
            raise ValueError("OMF record has no checksum byte")
        payload = cursor.take(size)
        checksum = payload[-1]
        if checksum and sum(raw[record_start:cursor.position]) & 0xFF:
            raise ValueError("OMF record checksum differs")
        records.append({"file_offset": record_start, "type": kind, "length": size,
                        "checksum_status": "omitted" if checksum == 0 else "verified"})
        r = Reader(payload[:-1])
        if ended:
            raise UnsupportedOMF("Multiple modules or bytes after MODEND")
        if kind == 0x96:
            while r.more():
                names.append(r.name())
        elif kind == 0x98:
            attributes = r.byte()
            if attributes & 1 or attributes >> 5 == 0:
                raise UnsupportedOMF("Only relocatable USE16 segments are supported")
            length = r.word()
            if attributes & 2:
                if length:
                    raise ValueError("Invalid BIG segment length")
                length = 65536
            name_index, class_index, overlay_index = r.index(), r.index(), r.index()
            segments.append({"name": names[name_index], "class": names[class_index],
                             "overlay": names[overlay_index], "length": length,
                             "bytes": bytearray(length), "written": bytearray(length)})
        elif kind in (0x8C, 0xB0):
            while r.more():
                name = r.name()
                type_index = r.index()
                definition = {"name": name, "type_index": type_index, "kind": kind}
                if kind == 0xB0:
                    data_type = r.byte()
                    if data_type == 0x62:
                        definition["size"] = r.communal_size()
                    elif data_type == 0x61:
                        definition["size"] = r.communal_size() * r.communal_size()
                    else:
                        raise UnsupportedOMF("Unsupported COMDEF data type")
                externals.append(definition)
        elif kind == 0x90:
            group_index, segment_index = r.index(), r.index()
            if not segment_index:
                raise UnsupportedOMF("Absolute PUBDEF is unsupported")
            while r.more():
                publics.append({"name": r.name(), "offset": r.word(), "type_index": r.index(),
                                "group_index": group_index, "segment_index": segment_index})
        elif kind == 0xA0:
            segment_index, offset = r.index(), r.word()
            content = r.take(len(r.data) - r.position)
            segment = segments[segment_index]
            end = offset + len(content)
            if end > segment["length"] or any(segment["written"][offset:end]):
                raise ValueError("LEDATA is out of bounds or overlapping")
            segment["bytes"][offset:end] = content
            segment["written"][offset:end] = b"\x01" * len(content)
            last_data = (segment_index, offset, len(content))
        elif kind == 0x9C:
            while r.more():
                first = r.byte()
                if not first & 0x80:
                    method = (first >> 2) & 7
                    is_frame = bool(first & 0x40)
                    if method not in ((0, 1, 2, 4, 5) if is_frame else (0, 1, 2)):
                        raise UnsupportedOMF("Unsupported THREAD method")
                    datum = r.index() if method <= 2 else None
                    thread = {"method": method, "datum": datum}
                    (frames if is_frame else targets)[first & 3] = thread
                    continue
                if last_data is None:
                    raise ValueError("FIXUP has no preceding LEDATA")
                location = ((first & 3) << 8) | r.byte()
                fix_data = r.byte()
                if fix_data & 0x80:
                    frame = frames[(fix_data >> 4) & 3]
                else:
                    method = (fix_data >> 4) & 7
                    if method not in (0, 1, 2, 4, 5):
                        raise UnsupportedOMF("Unsupported explicit FRAME method")
                    frame = {"method": method, "datum": r.index() if method <= 2 else None}
                if fix_data & 8:
                    target = targets[fix_data & 3]
                else:
                    target = {"method": fix_data & 3, "datum": r.index()}
                displacement = 0 if fix_data & 4 else r.word()
                seg, data_offset, data_length = last_data
                if location >= data_length:
                    raise ValueError("FIXUP is outside its LEDATA")
                fixups.append({"segment_index": seg, "offset": data_offset + location,
                               "location_type": (first >> 2) & 15,
                               "segment_relative": bool(first & 0x40),
                               "frame": dict(frame), "target": dict(target),
                               "displacement": displacement, "record_file_offset": record_start})
        elif kind == 0x8A:
            ended = True
        elif kind in (0x80, 0x88, 0x9A):
            pass  # Header, comments and groups retained in the raw object, not interpreted as code.
        else:
            raise UnsupportedOMF("Unsupported OMF record " + hex(kind))
    if not ended:
        raise ValueError("OMF module lacks MODEND")
    return {"path": str(path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest(),
            "segments": segments, "externals": externals, "publics": publics,
            "fixups": fixups, "records": records}


def select_function(obj, symbol):
    """Require one code public and complete bytes instead of choosing the largest segment."""
    matches = [s for s in obj["publics"] if s["name"] == symbol]
    if len(matches) != 1:
        raise ValueError("Expected exactly one matching PUBDEF")
    selected = matches[0]
    segment = obj["segments"][selected["segment_index"]]
    if segment["class"] != "CODE":
        raise ValueError("PUBDEF does not belong to a CODE segment")
    siblings = [s for s in obj["publics"] if s["segment_index"] == selected["segment_index"]]
    if len(siblings) != 1 or selected["offset"] != 0:
        raise UnsupportedOMF("Function extent requires a single code public at segment offset zero")
    if not segment["length"] or not all(segment["written"]):
        raise UnsupportedOMF("Code segment is empty or has unmaterialized gaps")
    return bytes(segment["bytes"]), selected


def resolve_ds_offsets(obj, symbol, placements):
    """Apply actual FIXUPP references using independently supplied DS-relative offsets."""
    code, selected = select_function(obj, symbol)
    linked = bytearray(code)
    resolved = []
    occupied = set()
    for fixup in obj["fixups"]:
        if fixup["segment_index"] != selected["segment_index"]:
            continue
        if (fixup["location_type"] != 1 or not fixup["segment_relative"]
                or fixup["frame"]["method"] != 5 or fixup["target"]["method"] != 2):
            raise UnsupportedOMF("FIXUP requires unsupported frame, target or location semantics")
        name = obj["externals"][fixup["target"]["datum"]]["name"]
        if name not in placements:
            raise UnsupportedOMF("Missing explicit symbol placement: " + name)
        offset = fixup["offset"]
        if offset + 2 > len(linked) or occupied.intersection((offset, offset + 1)):
            raise ValueError("FIXUP overlaps or crosses selected function")
        occupied.update((offset, offset + 1))
        addend = struct.unpack_from("<H", linked, offset)[0]
        value = placements[name] + fixup["displacement"] + addend
        if not 0 <= value <= 65535:
            raise ValueError("Explicit symbol placement overflows a 16-bit offset")
        struct.pack_into("<H", linked, offset, value)
        resolved.append({**fixup, "symbol": name, "original_addend": addend,
                         "supplied_ds_offset": placements[name], "resolved_operand": value})
    return bytes(linked), resolved
