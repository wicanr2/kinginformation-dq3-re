#!/usr/bin/env python3
"""產生 dq3_remake/src/dq3_restored_sprites.c:把 bug #3 的兩隻未完成 boss sprite
(128 歐里狄加 / 129 五頭龍大王)復原資料內嵌成 C,供 remake 在原版 DQ3MNS.SHP 空 sprite
時回退顯示(不修改版權素材、不依賴 work/DQ3MNS_fixed.SHP)。

來源:tools/make_sprites.py 由參考圖(dq3_org_pic/,gitignore)轉成 MNSBK.PAL plane-major
sprite——128 取 FC 風 128-sprite.jpeg;129 取現代 render 後紫→綠重映,對齊實機索瑪戰的綠
King Hydra。本檔為 committed artifact(比照 dq3_zhuyin_table.c);參考圖本身不入版控。

docker: tools/dockrun.sh tools/gen_restored_sprites.py(需先有 dq3_org_pic/ 參考圖)
"""
import importlib.util, os, struct, sys

# 重用 make_sprites 的 build_grid/encode/JOBS/GREEN_REMAP
spec = importlib.util.spec_from_file_location("mks", "tools/make_sprites.py")
mks = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mks)   # 執行時即生成 sprites + work/DQ3MNS_fixed.SHP

OUT = "dq3_remake/src/dq3_restored_sprites.c"
lines = [
    '/* dq3_restored_sprites.c — bug #3 復原 boss sprite(128 歐里狄加 / 129 五頭龍大王)。',
    ' * 生成檔(tools/gen_restored_sprites.py);格式 = MNSBK.PAL plane-major,同 DQ3MNS.SHP',
    ' * 單隻 sprite:u16 w_bytes、u16 h、4 plane(plane3→0)。原版這兩格為空,remake 回退本資料。',
    ' * 來源說明見生成器 docstring。請勿手改,改參考圖後重跑生成器。 */',
    '#include "dq3_restored_sprites.h"',
    '#include <string.h>',
    '',
]
for mid in (128, 129):
    blob = mks.sprites[mid]
    arr = ', '.join(str(b) for b in blob)
    lines.append('static const unsigned char SPR_%d[%d] = { %s };' % (mid, len(blob), arr))
    lines.append('')

lines += [
    'int dq3_restored_sprite(int id, dq3_monster_sprite *out)',
    '{',
    '    const unsigned char *d; int n;',
    '    if (id == 128) { d = SPR_128; n = (int)sizeof SPR_128; }',
    '    else if (id == 129) { d = SPR_129; n = (int)sizeof SPR_129; }',
    '    else return -1;',
    '    {',
    '        int wb = d[0] | (d[1] << 8); int h = d[2] | (d[3] << 8);',
    '        int W = (wb & 0x7fff) * 8, s, r, b, bit; int plane_sz = (wb & 0x7fff) * h;',
    '        wb &= 0x7fff;',
    '        if (W <= 0 || h <= 0 || W > DQ3_SHP_MAXW || h > DQ3_SHP_MAXH) return -1;',
    '        if (4 + plane_sz * 4 > n) return -1;',
    '        memset(out, 0, sizeof *out); out->w = W; out->h = h;',
    '        for (s = 0; s < 4; s++) {',
    '            int pl = 3 - s; int b0 = 4 + s * plane_sz;',
    '            for (r = 0; r < h; r++)',
    '                for (b = 0; b < wb; b++) {',
    '                    unsigned char v = d[b0 + r * wb + b];',
    '                    for (bit = 0; bit < 8; bit++)',
    '                        if (v & (0x80 >> bit)) out->px[r][b*8+bit] |= (unsigned char)(1 << pl);',
    '                }',
    '        }',
    '        for (r = 0; r < h; r++)',
    '            for (b = 0; b < W; b++)',
    '                out->opaque[r][b] = out->px[r][b] ? 1 : 0;',
    '    }',
    '    return 0;',
    '}',
    '',
]
open(OUT, "w").write('\n'.join(lines))
print("wrote", OUT, "128:%dB 129:%dB" % (len(mks.sprites[128]), len(mks.sprites[129])))
