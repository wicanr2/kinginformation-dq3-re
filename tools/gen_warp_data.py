#!/usr/bin/env python3
"""warp 表 0x4ea0(scripted 傳送目的)→ remake 資料。
RE(docs/31/34/35):section +8 examine 事件 type-2 → warp[param]={dest,X,Y}@0x4ea0(3-byte)→ 切場景
(原版餵轉場執行器 0xd1f9)。dest=目的 CTY 號、(X,Y)=落點。產出 dq3_remake/src/dq3_warp.{c,h}。"""
import os
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
exe=open(os.path.join(ROOT,'assets_raw/DQ3.EXE'),'rb').read(); DG=0x16140; base=DG+0x4ea0
N=120
ent=[(exe[base+i*3],exe[base+i*3+1],exe[base+i*3+2]) for i in range(N)]
L=['/* dq3_warp.c — scripted 傳送表 0x4ea0(自 DQ3.EXE 萃取,勿手改)。',
   ' * RE docs/31/35:section +8 examine type-2 → warp[param]={dest_cty,X,Y} → 切場景。*/',
   '#include "dq3_warp.h"','',
   'const dq3_warp dq3_warps[%d] = {'%N]
for d_,x,y in ent:
    L.append('  { %3d, %3d, %3d },'%(d_,x,y))
L+=['};','const int dq3_warps_len = %d;'%N,'',
    'int dq3_warp_get(int param, int *dest_cty, int *x, int *y)',
    '{',
    '    if (param < 0 || param >= dq3_warps_len) return -1;',
    '    if (dq3_warps[param].dest == 0 && dq3_warps[param].x == 0 && dq3_warps[param].y == 0) return -1;',
    '    if (dest_cty) *dest_cty = dq3_warps[param].dest;',
    '    if (x) *x = dq3_warps[param].x;',
    '    if (y) *y = dq3_warps[param].y;',
    '    return 0;',
    '}']
open(os.path.join(ROOT,'dq3_remake/src/dq3_warp.c'),'w').write('\n'.join(L)+'\n')
open(os.path.join(ROOT,'dq3_remake/include/dq3_warp.h'),'w').write(
'''/* dq3_warp.h — scripted 傳送表 0x4ea0(type-2 examine 事件目的地,docs/31/35)。 */
#ifndef DQ3_WARP_H
#define DQ3_WARP_H
typedef struct { unsigned char dest, x, y; } dq3_warp;   /* dest=目的 CTY、(x,y)=落點 */
extern const dq3_warp dq3_warps[];
extern const int dq3_warps_len;
/* 取 warp[param];回 0 並填 dest_cty/x/y,空項回 -1。 */
int dq3_warp_get(int param, int *dest_cty, int *x, int *y);
#endif
''')
nz=sum(1 for d_,x,y in ent if not(d_==0 and x==0 and y==0))
print('warp 表:%d 項(%d 非空)'%(N,nz))
