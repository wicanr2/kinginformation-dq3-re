/* sub_e6b9 — 亂數狀態前進 (LCG, byte-match target)
 *
 * 原版 (seg0 0xe6b9, 16 bytes, near ret):
 *   a1 5a0b  mov ax,[0xb5a] ; 05 1890 add ax,0x9018
 *   d1c0 d1c0 d1c0  rol ax,1 x3 ; a3 5a0b mov [0xb5a],ax ; c3 ret
 *
 * 無 BP frame、無 local。讀全域 -> +0x9018 -> rol 3 -> 回存 -> 回傳 (ax)。
 * _rotl 以 intrinsic 內聯 (常數 count=3 -> 3 x rol ax,1)。
 * 全域絕對位址 (0xb5a) 由 linker 指派,屬資料 reloc 差異 (非 codegen)。
 */
#include <stdlib.h>
#pragma intrinsic(_rotl)

unsigned g_rng;

unsigned sub_e6b9(void)
{
    return g_rng = _rotl(g_rng + 0x9018, 3);
}
