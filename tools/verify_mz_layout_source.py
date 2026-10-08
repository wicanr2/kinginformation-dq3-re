"""Rebuild typed MZ header/file-end data; entry: docs/25-match-progress.md.

render_layout consumes only JSON values, never original executable bytes.
The CLI uses the original EXE strictly downstream for byte comparison.
This does not materialize any executable code or loaded-image data.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import struct


FIELDS = ["signature", "last_page_bytes", "pages", "relocation_count", "header_paragraphs",
          "minimum_extra_paragraphs", "maximum_extra_paragraphs", "initial_ss", "initial_sp",
          "checksum", "initial_ip", "initial_cs", "relocation_table_offset", "overlay_number"]


def uint(value, maximum):
    if type(value) is not int or not 0 <= value <= maximum:
        raise ValueError("Unsigned field is absent or out of range")
    return value


def render_layout(model):
    if model["schema_version"] != 1 or model["relocation_location_encoding"] != "64KiB-windows":
        raise ValueError("Unreviewed layout source contract")
    fields = model["fields"]
    if set(fields) != set(FIELDS):
        raise ValueError("MZ fields are incomplete or unknown")
    values = [uint(fields[name], 65535) for name in FIELDS]
    if fields["signature"] != 0x5A4D or fields["pages"] < 1 or fields["last_page_bytes"] > 511:
        raise ValueError("Invalid original MZ signature/page fields")
    if fields["minimum_extra_paragraphs"] > fields["maximum_extra_paragraphs"]:
        raise ValueError("MZ extra-memory interval is invalid")
    header_size = fields["header_paragraphs"] * 16
    declared = (fields["pages"] - 1) * 512 + (fields["last_page_bytes"] or 512)
    if declared != model["declared_file_length"] or declared < header_size:
        raise ValueError("Declared file length differs from the MZ fields")
    extra = model["unknown_extension_words"]
    if not isinstance(extra, list):
        raise ValueError("Unknown header extension requires typed words")
    header = bytearray(struct.pack("<14H", *values))
    for word in extra:
        header.extend(struct.pack("<H", uint(word, 65535)))
    if len(header) != fields["relocation_table_offset"]:
        raise ValueError("Relocation table offset differs from typed prefix")
    locations = model["ordered_relocation_word_logical_offsets"]
    if not isinstance(locations, list) or len(locations) != fields["relocation_count"]:
        raise ValueError("Relocation count differs from typed source")
    decoded = []
    for text in locations:
        if not isinstance(text, str) or not re.fullmatch(r"0x[0-9a-f]+", text):
            raise ValueError("Relocation position is not a canonical logical offset")
        logical = int(text, 16)
        if not 0 <= logical < declared - header_size - 1:
            raise ValueError("Relocation word crosses the declared image")
        decoded.append(logical)
        # Original location encoding, distinct from target/caller CS frames.
        header.extend(struct.pack("<HH", logical & 65535, (logical >> 16) * 4096))
    if decoded != sorted(set(decoded)):
        raise ValueError("Original relocation ordering/uniqueness differs")
    padding = model["header_padding"]
    count = uint(padding["count"], 65535)
    header.extend(bytes([uint(padding["u8_value"], 255)]) * count)
    if len(header) != header_size:
        raise ValueError("Typed MZ header does not fill its exact extent")
    if fields["initial_cs"] * 16 + fields["initial_ip"] >= declared - header_size:
        raise ValueError("MZ entry is outside its declared image")
    suffix = bytearray()
    for word in model["unknown_after_declared_length_u16"]:
        suffix.extend(struct.pack("<H", uint(word, 65535)))
    if declared + len(suffix) != model["input"]["size"]:
        raise ValueError("Opaque file-end extent differs from original identity")
    return bytes(header), bytes(suffix)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    root = Path(__file__).resolve().parents[1]
    source = root / "re/match/mz_layout.json"
    model = json.loads(source.read_text())
    header, suffix = render_layout(model)
    original = root / model["input"]["path"]
    raw = original.read_bytes()
    if sha(original) != model["input"]["sha256"] or len(raw) != model["input"]["size"]:
        raise ValueError("Original comparison identity differs")
    if header != raw[:len(header)] or suffix != raw[model["declared_file_length"]:]:
        raise ValueError("Source-built header/file-end differs from original")
    args.output.mkdir(exist_ok=False)
    (args.output / "header.bin").write_bytes(header)
    (args.output / "file-end.bin").write_bytes(suffix)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)),
               "source_sha256": sha(source), "input": model["input"],
               "header_bytes": len(header), "header_sha256": hashlib.sha256(header).hexdigest(),
               "declared_file_length": model["declared_file_length"],
               "file_end_bytes": len(suffix), "file_end_sha256": hashlib.sha256(suffix).hexdigest(),
               "relocation_count": model["fields"]["relocation_count"],
               "header_and_file_end_byte_exact": True, "whole_goal_complete": False,
               "scope": "Typed metadata layout only; original code/body is not a build input"}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"header_bytes": len(header), "file_end_bytes": len(suffix),
                      "relocations": receipt["relocation_count"], "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
