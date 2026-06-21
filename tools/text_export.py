#!/usr/bin/env python3
"""把 D3TXT 記錄轉成標註版文字 dump(控制碼以 {TAG} 表示,glyph 以 #索引 表示)。

由於精訊字型是自訂內部碼、無 Unicode 對照表,本工具不輸出可直接讀的中文,
而是輸出「結構化、可追溯」的記錄:每筆一行,控制碼語意化,glyph 用 16 進位索引。
真正肉眼可讀的中文請看 assets_out/txtNN_full.png(以 FON 字模 render)。

用法: text_export.py
"""
import json, os

DATA = 'docs/data/d3txt_codes.json'
OUTDIR = 'docs/data'

# 已從 atlas 與上下文確認的控制 / 標點對照
CTRL = {
    0xfffe: '\n      ',   # 換行 + 兩格縮排(後面常接 0x0c 0x0c)
    0xfffc: '\n--page--\n',  # 換頁 / 訊息段落分隔
    0xfffb: '{VAR}',     # 動態插入(名稱/道具/數值占位)
    0xfffa: '{VAR}',
    0xfff9: '{VAR}',
    0xfff6: '{VAR}',
    0xfff5: '{VAR}',
    0xffed: '{VAR}',
    0xfffd: '\n      ',  # 換行(變體)
}
# 低索引標點(由 atlas 確認)
PUNC = {
    0x0c: ' ', 0x37: ',', 0x38: '。', 0x39: '!', 0x3a: '?',
    0x3b: '*', 0x3c: '「', 0x3d: ':',
}


def fmt(codes):
    out = []
    for c in codes:
        if c in CTRL:
            out.append(CTRL[c])
        elif c in PUNC:
            out.append(PUNC[c])
        else:
            out.append(f'#{c:x}')  # glyph 索引(無 Unicode 對照)
    return ''.join(out)


def main():
    d = json.load(open(DATA))
    for k, v in d.items():
        lines = [f'# {k}  records={v["num_records"]}  table={v["table_bytes"]}B', '']
        for r in v['records']:
            lines.append(f'[{r["i"]}] off={r["off"]} {"<term>" if r["term"] else ""}')
            lines.append('    ' + fmt(r['codes']).replace('\n', '\n    '))
        open(f'{OUTDIR}/{k}_codes.txt', 'w').write('\n'.join(lines))
        print(f'{k} -> {OUTDIR}/{k}_codes.txt')


if __name__ == '__main__':
    main()
