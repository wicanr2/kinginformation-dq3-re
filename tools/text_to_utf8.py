#!/usr/bin/env python3
"""用 glyph_unicode_map.json 把 D3TXT*.TXT 解成純文字 UTF-8。

控制碼語意化:
  0xfffe / 0xfffd -> 換行(\n + 行縮排去除)
  0xfffc          -> 換頁: \n---\n
  0xfffb/fffa/fff9/fff6/fff5/ffed -> 動態插值占位 {var}
  0xffff          -> 記錄終止(忽略)

用法:
  text_to_utf8.py            # dump 全部 10 檔 -> docs/script/txtNN.txt
  text_to_utf8.py <TXT>      # 印單檔到 stdout
"""
import struct, sys, json, os

FON_GLYPHS = 1476
MAP_PATH = "docs/data/glyph_unicode_map.json"

CTRL = {
    0xfffe: "\n", 0xfffd: "\n",
    0xfffc: "\n---\n",
    0xfffb: "{變數}", 0xfffa: "{變數}", 0xfff9: "{變數}",
    0xfff6: "{變數}", 0xfff5: "{變數}", 0xffed: "{變數}",
}


def load_map():
    m = json.load(open(MAP_PATH, encoding="utf-8"))
    return {int(k): v for k, v in m.items()}


def parse_records(txt):
    d = open(txt, "rb").read()
    first = struct.unpack("<H", d[0:2])[0]
    ptrs = [struct.unpack("<H", d[i:i+2])[0] for i in range(0, first, 2)]
    recs = []
    for i in range(len(ptrs) - 1):
        s, e = ptrs[i], ptrs[i+1]
        raw = d[s:e]
        codes = [struct.unpack("<H", raw[j:j+2])[0]
                 for j in range(0, len(raw) - (len(raw) % 2), 2)]
        recs.append(codes)
    return recs


def decode_record(codes, gmap):
    out = []
    miss = 0
    for c in codes:
        if c == 0xffff:
            continue
        if c in CTRL:
            out.append(CTRL[c])
        elif c < FON_GLYPHS:
            ch = gmap.get(c)
            if ch is None:
                out.append(f"[?{c}]")
                miss += 1
            else:
                # 0x0c(12)/0x0d(13) 縮排空白 -> 去掉行首視覺縮排但保留一般空白
                out.append(ch)
        else:
            out.append(f"[ctrl{c:04x}]")
            miss += 1
    return "".join(out), miss


def dump_file(txt, gmap):
    recs = parse_records(txt)
    lines = []
    total_miss = 0
    for i, codes in enumerate(recs):
        text, miss = decode_record(codes, gmap)
        total_miss += miss
        # 清掉因 0x0c 縮排造成的多餘行首空白(每行開頭的連續空格)
        text = "\n".join(ln.lstrip(" ") for ln in text.split("\n"))
        lines.append(f"[{i}] {text}")
    return lines, total_miss, len(recs)


def main():
    gmap = load_map()
    if len(sys.argv) > 1:
        lines, miss, n = dump_file(sys.argv[1], gmap)
        print("\n".join(lines))
        print(f"\n# records={n} unmapped_codes={miss}", file=sys.stderr)
        return
    os.makedirs("docs/script", exist_ok=True)
    grand = 0
    summary = []
    for i in range(10):
        txt = f"assets_raw/D3TXT0{i}.TXT"
        lines, miss, n = dump_file(txt, gmap)
        grand += miss
        summary.append((i, n, miss))
        header = (f"# D3TXT0{i}.TXT 純文字 (glyph index -> Unicode via glyph_unicode_map.json)\n"
                  f"# 記錄數={n}  未對應碼={miss}\n"
                  f"# 控制碼: 換行=實際換行  換頁=---  動態插值={{變數}}\n")
        with open(f"docs/script/txt0{i}.txt", "w", encoding="utf-8") as f:
            f.write(header + "\n" + "\n".join(lines) + "\n")
    for i, n, miss in summary:
        print(f"txt0{i}: records={n} unmapped={miss}")
    print(f"TOTAL unmapped codes = {grand}")


if __name__ == "__main__":
    main()
