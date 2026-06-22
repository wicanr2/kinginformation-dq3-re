/* sub_32a3 — 對某 byte 陣列從 offset 8 起、步進 2、連續 13 個元素設 0xff
 *
 * 原版 (seg0 0x32a3, 17 bytes, near ret, 無 BP frame):
 *   b90d00     mov cx,0xd          ; 13 次
 *   bb0800     mov bx,8            ; 起始 index 8
 *  .l:
 *   c6875d26ff mov byte [bx+0x265d],0xff
 *   83c302     add bx,2
 *   e2f6       loop .l
 *   c3         ret
 *
 * = register int bx=8,cx=13; do { arr[bx]=0xff; bx+=2; } while(--cx);
 * 要點:用 register 變數 (MSC 5.1 /Ox 才會把 bx/cx 留在暫存器、不開 BP frame),
 * do-while(--cx) -> loop;陣列 base 0x265d 為 linker 指派絕對位址 (data reloc 差異)。
 */
unsigned char g_flags[64];

void sub_32a3(void)
{
    register int bx = 8;
    register int cx = 13;
    do {
        g_flags[bx] = 0xff;
        bx += 2;
    } while (--cx);
}
