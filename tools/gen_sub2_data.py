#!/usr/bin/env python3
"""子型2 scripted-event NPC 跳表(DGROUP 0x3bb4)→ remake 資料。
RE(docs/42):NPC 互動子型 2(handler 0x6355)以 byte4*2 索引 DS 0x3bb4 跳表 → 呼叫 handler。
handler = 旗標條件對話 NPC:依故事旗標(test_flag 0x8279)選 section-bank 對話 rec(di=0xbb8+offset);
部分另呼叫特殊 sub(0x2719 記錄點 / 0xd1f9 傳送 / 0x16dd 話す列表 等)。
remake 無完整旗標狀態 → 取 handler 的「主對話 rec」(第一個 rec offset)顯示;純特殊無對話者標 0xff。
產出 dq3_remake/src/dq3_sub2.{c,h}。"""
import struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
import os
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
d=open(os.path.join(ROOT,'assets_raw/DQ3.EXE'),'rb').read()
md=Cs(CS_ARCH_X86, CS_MODE_16); DG=0x16140; base=DG+0x3bb4
SPECIAL={0x2719:'record',0xd1f9:'warp',0x16dd:'talk_list'}
def scan(fl):
    rec=None; special=None
    for ins in md.disasm(d[fl:fl+0x120],fl):
        m,o=ins.mnemonic,ins.op_str
        if m=='mov' and o.startswith('di, 0x'):
            v=int(o.split('0x')[1],16)
            if 0xbb8<=v<0xd00 and rec is None: rec=v-0xbb8
        if m=='call':
            for k,nm in SPECIAL.items():
                if ('0x%x'%k) in o and special is None: special=nm
        if m in('ret','retf'): break
    return rec,special
N=40
tab=[]
for i in range(N):
    w=struct.unpack_from('<H',d,base+i*2)[0]
    if w==0: tab.append((0xff,None)); continue
    tab.append(scan(w+0x1370))
L=['/* dq3_sub2.c — 子型2 scripted-event NPC 主對話 rec(自 DGROUP 0x3bb4 跳表萃取,勿手改)。',
   ' * RE docs/42:byte4 → handler → 旗標條件對話。此表取 handler 的主對話 rec offset',
   ' * (section bank 內 rec = offset;0xff=該 handler 無單純對話/純特殊事件)。*/',
   '#include "dq3_sub2.h"','',
   'const unsigned char dq3_sub2_rec[%d] = {'%N]
for i in range(0,N,10):
    L.append('  '+', '.join('0x%02x'%(tab[j][0] if tab[j][0] is not None else 0xff) for j in range(i,min(i+10,N)))+',')
L+=['};','const int dq3_sub2_rec_len = %d;'%N,'',
    'int dq3_sub2_dialogue(int byte4)',
    '{',
    '    if (byte4 < 0 || byte4 >= dq3_sub2_rec_len) return -1;',
    '    return dq3_sub2_rec[byte4] == 0xff ? -1 : dq3_sub2_rec[byte4];',
    '}']
open(os.path.join(ROOT,'dq3_remake/src/dq3_sub2.c'),'w').write('\n'.join(L)+'\n')
open(os.path.join(ROOT,'dq3_remake/include/dq3_sub2.h'),'w').write(
'''/* dq3_sub2.h — 子型2 scripted-event NPC 主對話 rec 表(docs/42)。 */
#ifndef DQ3_SUB2_H
#define DQ3_SUB2_H
extern const unsigned char dq3_sub2_rec[];
extern const int dq3_sub2_rec_len;
/* 子型2 NPC(byte4)的主對話 rec(section bank 內;-1=純特殊事件無單純對話)。 */
int dq3_sub2_dialogue(int byte4);
#endif
''')
nd=sum(1 for r,s in tab if r!=0xff and r is not None)
nsp=sum(1 for r,s in tab if s)
print('子型2:%d handler,%d 有主對話 rec,%d 帶特殊事件(record/warp/talk_list)'%(N,nd,nsp))
for i,(r,s) in enumerate(tab):
    if r!=0xff or s: print('  byte4=%2d rec=%s special=%s'%(i,('0x%02x'%r if r is not None and r!=0xff else '-'),s or '-'))
