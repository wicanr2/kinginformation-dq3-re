#!/usr/bin/env python3
"""mz_roundtrip.py — DQ3.EXE 結構化拆解 → 重組,證實 byte-identical。

RE 正確性驗證的最確定訊號:把原版拆成「具名結構表示」(MZ header 各欄位 +
reloc table + header padding + load image segments + trailing pad),再從那份
表示重組,sha256 與原版完全相同。攻不下成乾淨指令的段以 raw bytes 保留——
目標是「能重建出完全相同的檔」,不是好看。

子命令:
  split  <EXE> <OUTDIR>   把 EXE 拆成 manifest.json + 各區段 .bin
  build  <OUTDIR> <OUT>   從 manifest + .bin 重組
  verify <EXE> <OUTDIR>   split→build→sha256 比對(一鍵 pass/fail)

設計:每個 byte 都歸屬到某個 region(無洞、無重疊);build 依 region 順序串接。
這保證「拆得出就一定組得回」,真正的考驗是 region 切分是否覆蓋全檔且不漏。
"""
import sys, os, json, struct, hashlib

MZ_FIELDS = ['e_magic','e_cblp','e_cp','e_crlc','e_cparhdr','e_minalloc',
             'e_maxalloc','e_ss','e_sp','e_csum','e_ip','e_cs','e_lfarlc','e_ovno']


def sha(b):
    return hashlib.sha256(b).hexdigest()


def parse_header(d):
    vals = struct.unpack_from('<14H', d, 0)
    return dict(zip(MZ_FIELDS, vals))


def split(exe_path, outdir):
    d = open(exe_path, 'rb').read()
    os.makedirs(outdir, exist_ok=True)
    hdr = parse_header(d)
    hdr_bytes = hdr['e_cparhdr'] * 16          # 0x1370
    reloc_off = hdr['e_lfarlc']                # 0x22
    reloc_cnt = hdr['e_crlc']                  # 1232
    reloc_end = reloc_off + reloc_cnt * 4      # 0x1362

    regions = []

    def emit(name, start, end, kind):
        """登記一個 region,把 bytes 寫成 .bin。"""
        blob = d[start:end]
        fn = f'{name}.bin'
        open(os.path.join(outdir, fn), 'wb').write(blob)
        regions.append({'name': name, 'file': fn, 'start': start,
                        'end': end, 'len': end - start, 'kind': kind,
                        'sha256': sha(blob)})

    # 1) MZ header 前 28 bytes(14 個 u16 欄位)——以欄位值重建,不存 bin
    #    其餘 header(reloc 前的 gap、reloc table、reloc 後到 hdr_bytes 的 padding)各自登記。
    # region: header fields(0..0x1c)用結構欄位重建
    regions.append({'name': 'mz_fields', 'kind': 'mz_header_fields',
                    'start': 0, 'end': 0x1c, 'len': 0x1c,
                    'fields': hdr,
                    'sha256': sha(d[0:0x1c])})
    # region: gap between header fields end (0x1c) and reloc table start
    if reloc_off > 0x1c:
        emit('hdr_gap', 0x1c, reloc_off, 'raw')
    elif reloc_off < 0x1c:
        raise SystemExit('reloc table overlaps header fields — unexpected layout')
    # region: reloc table
    emit('reloc_table', reloc_off, reloc_end, 'reloc')
    # region: header padding (reloc_end .. hdr_bytes)
    if hdr_bytes > reloc_end:
        emit('hdr_pad', reloc_end, hdr_bytes, 'raw')

    # 2) load image (hdr_bytes .. EOF) — 先整塊存(後續可再細分成 segment)
    emit('load_image', hdr_bytes, len(d), 'image')

    manifest = {
        'source': os.path.basename(exe_path),
        'file_size': len(d),
        'file_sha256': sha(d),
        'header_bytes': hdr_bytes,
        'reloc': {'offset': reloc_off, 'count': reloc_cnt, 'end': reloc_end},
        'regions': regions,
    }
    json.dump(manifest, open(os.path.join(outdir, 'manifest.json'), 'w'),
              indent=2, ensure_ascii=False)
    # 覆蓋率自檢:regions 必須無洞、無重疊、覆蓋全檔
    _check_coverage(regions, len(d))
    print(f'split OK: {len(regions)} regions, file_sha256={manifest["file_sha256"]}')
    return manifest


def _check_coverage(regions, total):
    rs = sorted(regions, key=lambda r: r['start'])
    pos = 0
    for r in rs:
        if r['start'] != pos:
            raise SystemExit(f'COVERAGE GAP/OVERLAP at {pos:#x}: next region {r["name"]} starts {r["start"]:#x}')
        pos = r['end']
    if pos != total:
        raise SystemExit(f'COVERAGE END MISMATCH: covered {pos:#x}, file {total:#x}')


def build(outdir, out_path):
    manifest = json.load(open(os.path.join(outdir, 'manifest.json')))
    regions = sorted(manifest['regions'], key=lambda r: r['start'])
    buf = bytearray()
    for r in regions:
        if r['kind'] == 'mz_header_fields':
            f = r['fields']
            blob = struct.pack('<14H', *[f[k] for k in MZ_FIELDS])
            assert len(blob) == r['len'], 'header field pack size mismatch'
        else:
            blob = open(os.path.join(outdir, r['file']), 'rb').read()
        if len(blob) != r['len']:
            raise SystemExit(f'region {r["name"]} len mismatch {len(blob)} != {r["len"]}')
        buf.extend(blob)
    open(out_path, 'wb').write(buf)
    print(f'build OK: {len(buf)} bytes -> {out_path}  sha256={sha(bytes(buf))}')
    return bytes(buf)


def verify(exe_path, outdir):
    orig = open(exe_path, 'rb').read()
    split(exe_path, outdir)
    rebuilt = build(outdir, os.path.join(outdir, '_rebuilt.bin'))
    o, r = sha(orig), sha(rebuilt)
    print('--- VERIFY ---')
    print('orig   sha256:', o)
    print('rebuilt sha256:', r)
    if o == r:
        print('RESULT: PASS — byte-identical round-trip (100%)')
        return 0
    # 找第一個不同 byte
    n = min(len(orig), len(rebuilt))
    for i in range(n):
        if orig[i] != rebuilt[i]:
            print(f'RESULT: FAIL — first diff at {i:#x}: orig={orig[i]:#04x} rebuilt={rebuilt[i]:#04x}')
            break
    else:
        print(f'RESULT: FAIL — length differs orig={len(orig)} rebuilt={len(rebuilt)}')
    return 1


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print(__doc__); sys.exit(1)
    cmd = sys.argv[1]
    if cmd == 'split':
        split(sys.argv[2], sys.argv[3])
    elif cmd == 'build':
        build(sys.argv[2], sys.argv[3])
    elif cmd == 'verify':
        sys.exit(verify(sys.argv[2], sys.argv[3]))
    else:
        print(__doc__); sys.exit(1)
