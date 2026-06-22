/* sub_e96d — 以等差數列 (公差 0x54) 填一個 word 陣列 (0x15e 個元素)
 *
 * 原版 (seg0 0xe96d, 21 bytes, near ret):
 *   8d3e9a28   lea di,[0x289a]
 *   b95e01     mov cx,0x15e
 *   b80000     mov ax,0
 *  .l:
 *   8905       mov [di],ax
 *   055400     add ax,0x54
 *   83c702     add di,2
 *   e2f6       loop .l
 *   c3         ret
 *
 * = word arr[0x15e]; v=0; for(i;cx) { *p++=v; v+=0x54; }
 * 重點:lea di + loop (cx 計數) + di 步進 2;ax 為等差值,起始 0 (mov ax,0)。
 * 陣列絕對位址 (0x289a) 由 linker 指派,屬資料 reloc 差異 (非 codegen)。
 */
unsigned g_tbl[0x15e];

void sub_e96d(void)
{
    unsigned *p = g_tbl;
    unsigned v = 0;
    int n = 0x15e;
    do {
        *p++ = v;
        v += 0x54;
    } while (--n);
}
