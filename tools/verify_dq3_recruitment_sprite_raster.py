"""核對正常招募畫面的人物差異來源；入口與證據限制見 docs/188。

只讀原始素材與既有完整 PNG，不產生替代畫面、不修改相位。
此診斷不能證明動畫時序或正式 remake 的畫面 parity。
"""

import argparse
import hashlib
import json
from pathlib import Path
import re
import struct
import sys

from verify_dosgolem_mother_return import png


ASSETS = {
    "DQ3MST.BLS": (115206, "a1a48eaf6c13ae73472d5ff77769fa538c19e048c24496d20218a89f3230f244"),
    "DQ3MAN.BLS": (222726, "823f57e0724e36ac8ed1aa472f59d2e8fb059f171e05e3aafc457e386f77158d"),
    "CTY00.DAT": (7546, "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836"),
    "DQ31.BLK": (65286, "5996d95d743fb8e1fe8d3ad29c513f5241af2c3574cf9f652dce57ebd6ba3298"),
}
SOURCE_HASH = "cf0730f10e48673e2da6702c77a6e458269cfe0153216b8770b7d3889a08e829"
PACK_HASH = "sha256:19f6124c5f03c9f14c2909b94be5ac48438df956ea8b59e82af58c8168c558f6"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def plane_pixel(raw, base, x, y, row_bytes=4, height=24):
    offset, bit = y * row_bytes + x // 8, 0x80 >> (x % 8)
    return sum(1 << (3 - p) for p in range(4)
               if raw[base + p * row_bytes * height + offset] & bit)


def frame_pixel(raw, frame, x, y, background):
    base = 6 + frame * 480
    require(base >= 6 and base + 480 <= len(raw), "BLS frame out of bounds")
    transparent = raw[base + 384 + y * 4 + x // 8] & (0x80 >> (x % 8))
    return background if transparent else plane_pixel(raw, base, x, y)


def verify(args):
    source_raw = args.source_receipt.read_bytes()
    require(sha(source_raw) == SOURCE_HASH, "unaccepted original receipt identity")
    source = json.loads(source_raw)
    require(source["seed"] == "1357" and source["normal_inputs"] == 242
            and all(len(source[k]) == 204 for k in ("queued", "consumed", "states")),
            "original input conditions differ")
    manifest = {a["path"]: a for a in source["artifacts"]}
    require(len(manifest) == len(source["artifacts"]), "duplicate original artifact")
    for name, artifact in manifest.items():
        require(Path(name).name == name, "original artifact path differs")
        raw = (args.original / name).read_bytes()
        require(len(raw) == artifact["size"] and sha(raw) == artifact["sha256"],
                f"original artifact differs: {name}")
    log_raw = (args.original / "issue4-recruit-yes-r1.log").read_bytes()
    require(sha(log_raw) == source["log_sha256"], "original observer log differs")
    observations = [dict(re.findall(r"(\w+)=(\S+)", s)) for s in log_raw.decode().splitlines()
                    if s.startswith("DQ3_RECRUIT_SPRITE ")]
    require(len(observations) == 260 and all(o["phase0004"] == "1" for o in observations),
            "original observed phase differs")
    runtime_raw = (args.runtime / "recruitment-continue.json").read_bytes()
    runtime = json.loads(runtime_raw)
    require(runtime["pack_schema"] == "0.13.0" and runtime["pack_hash"] == PACK_HASH,
            "runtime pack differs")
    require(runtime["seed"] == 0x1357, "runtime seed differs")
    raw_assets = {}
    for name, (size, digest) in ASSETS.items():
        raw = (args.assets / name).read_bytes()
        require(len(raw) == size and sha(raw) == digest, f"asset identity differs: {name}")
        raw_assets[name] = raw
    cty, blk = raw_assets["CTY00.DAT"], raw_assets["DQ31.BLK"]
    section = struct.unpack_from("<H", cty, 0)[0]
    layout = section + struct.unpack_from("<H", cty, section + 14)[0]
    width, height = struct.unpack_from("<HH", cty, layout)
    row_bytes, tile_height, tile_count = struct.unpack_from("<HHH", blk)
    require((section, width, height, row_bytes, tile_height, tile_count)
            == (12, 42, 43, 4, 24, 170), "map or tile shape differs")
    # 角色名與 raw frame 僅用於此已接受路線的診斷，未進入引擎。
    actors = [
        ("hero", "DQ3MST.BLS", 4, 2, 18, 288, 168, None),
        ("counter", "DQ3MAN.BLS", 200, 2, 16, 288, 120, 0x88),
        ("lower_right", "DQ3MAN.BLS", 26, 8, 21, 480, 240, 0x8f),
    ]
    expected, actor_reports = {}, []
    # 色盤直接取自原版色號／PNG配對，不補未出現的色號。
    indexed_path = args.original / "issue4-recruit-yes-r1-packet-194-ready.bin"
    original194 = args.original / "issue4-recruit-yes-r1-packet-194-ready.png"
    indices = indexed_path.read_bytes()
    w, h, _, original_rgb = png(original194)
    require((w, h, len(indices)) == (640, 350, 224000), "original canvas shape differs")
    palette = {}
    for index, color in zip(indices, original_rgb):
        require(index not in palette or palette[index] == color, "inconsistent original palette")
        palette[index] = color
    for name, asset, frame, wx, wy, sx, sy, record in actors:
        if record is not None:
            raw_record = cty[record:record + 7]
            require(tuple(raw_record[:2]) == (wx, wy), "NPC raw position differs")
            raw_direction = raw_record[3] & 3
            require(((raw_record[2] - 4) * 4 + raw_direction) * 2 == frame,
                    "NPC raw asset frame differs")
        tile = cty[layout + 4 + 2 * (wy * width + wx)]
        require(tile < tile_count, "background tile out of bounds")
        tile_base = 6 + tile * row_bytes * tile_height * 4
        local_difference = 0
        for y in range(24):
            for x in range(32):
                background = plane_pixel(blk, tile_base, x, y)
                colors = [frame_pixel(raw_assets[asset], frame + phase, x, y, background)
                          for phase in (0, 1)]
                require(all(c in palette for c in colors), "required palette entry absent")
                offset = (sy + y) * 640 + sx + x
                require(offset not in expected, "overlapping diagnostic actors")
                expected[offset] = (palette[colors[0]], palette[colors[1]], name)
                local_difference += colors[0] != colors[1]
        actor_reports.append({"actor": name, "asset": asset, "raw_frames": [frame, frame + 1],
                              "file_ranges": [[6 + f * 480, 6 + (f + 1) * 480]
                                              for f in (frame, frame + 1)],
                              "cty_record_file": hex(record) if record is not None else None,
                              "background_tile": tile, "full_actor_difference": local_difference})
    samples = []
    for sample in runtime["samples"]:
        number = sample["packet"]
        if not 194 <= number <= 204:
            continue
        require(sample["hero_walk"] == 0, "observed runtime hero phase differs")
        for record in (14, 15):
            visible = [n for n in sample["npc_visuals"] if n["record"] == record]
            require(len(visible) == 1 and visible[0]["walk"] == 0,
                    "observed runtime NPC phase differs")
        original_path = args.original / f'issue4-recruit-yes-r1-packet-{number:03d}-{sample["original_phase"]}.png'
        runtime_path = args.runtime / f"recruitment-continue-packet-{number:03d}.png"
        ow, oh, _, a = png(original_path)
        rw, rh, _, b = png(runtime_path)
        require((ow, oh, rw, rh) == (640, 350, 640, 350), "sample canvas shape differs")
        differences = [i for i, (old, new) in enumerate(zip(b, a)) if old != new]
        counts = {name: 0 for name, *_ in actors}
        for offset in differences:
            require(offset in expected, "difference outside known actor raster")
            old, new, name = expected[offset]
            require((b[offset], a[offset]) == (old, new), "difference color not explained by raw frames")
            counts[name] += 1
        require(len(differences) == sample["full_rgb_difference"], "runtime difference metadata differs")
        if number in (194, 204):
            # 兩張無modal畫面核對完整三個人物與底圖，包含所有透明像素。
            require(all((b[i], a[i]) == (old, new) for i, (old, new, _) in expected.items()),
                    "complete actor/background raster differs")
        samples.append({"packet": number, "full_rgb_difference": len(differences),
                        "explained_difference": counts, "unexplained_difference": 0,
                        "original_png": str(original_path), "runtime_png": str(runtime_path),
                        "original_png_sha256": sha(original_path.read_bytes()),
                        "runtime_png_sha256": sha(runtime_path.read_bytes())})
    require([s["packet"] for s in samples] == list(range(194, 205)), "required samples missing")
    return {"scope": "read-only whole-canvas difference diagnosis; never replacement/cropped parity",
            "source_receipt_sha256": sha(source_raw), "runtime_receipt_sha256": sha(runtime_raw),
            "source_receipt": str(args.source_receipt), "runtime_directory": str(args.runtime),
            "asset_directory": str(args.assets), "python_version": sys.version,
            "tool_sha256": sha(Path(__file__).read_bytes()), "assets": ASSETS,
            "original_sprite_observations": len(observations),
            "actors": actor_reports, "samples": samples,
            "raster_explanation": "confirmed for these captures only",
            "animation_timing": "unknown; separate game counter and per-consumer reads required",
            "production_changed": False, "remake_parity": False}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--assets", type=Path, required=True)
    parser.add_argument("--original", type=Path, required=True)
    parser.add_argument("--source-receipt", type=Path, required=True)
    parser.add_argument("--runtime", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), "refuse to overwrite existing receipt")
    result = verify(args)
    with args.output.open("x") as stream:
        json.dump(result, stream, indent=2, ensure_ascii=False)
        stream.write("\n")
    print(json.dumps({"samples": len(result["samples"]), "unexplained": 0,
                      "parity": False, "output": str(args.output)}, ensure_ascii=False))


if __name__ == "__main__":
    main()
