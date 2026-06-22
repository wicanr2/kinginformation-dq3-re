/* sub_edab — 把 src (指標存於全域 [0x25d1]) 複製 0x30 bytes 到固定陣列 [0x35c2]
 *
 * 原版 (seg0 0xedab, 18 bytes, near ret, 無 BP frame):
 *   06         push es
 *   1e         push ds
 *   07         pop es          ; ES=DS
 *   8b36d125   mov si,[0x25d1] ; src offset (near 指標全域)
 *   8d3ec235   lea di,[0x35c2] ; dest 固定陣列
 *   b93000     mov cx,0x30     ; 48 bytes
 *   f3a4       rep movsb
 *   07         pop es
 *   c3         ret
 *
 * = memcpy(g_dest, g_src, 0x30);  (MSC 5.1 memcpy intrinsic 對 near 指標 inline rep movsb)
 * g_src 是 near 指標全域;ES 由 push ds/pop es 設成 DS (small model 同段)。
 * 全域絕對位址 (0x25d1/0x35c2) 由 linker 指派,屬資料 reloc 差異。
 */
#include <memory.h>
#pragma intrinsic(memcpy)

char *g_src;           /* [0x25d1] near 指標 */
char  g_dest[0x30];    /* [0x35c2] */

void sub_edab(void)
{
    memcpy(g_dest, g_src, 0x30);
}
