/* sub_e91e — VGA Graphics Controller 暫存器初始化 (一串 word OUT 到 0x3ce)
 *
 * 原版 (seg0 0xe91e, 20 bytes, near ret, 無 BP frame):
 *   bace03   mov dx,0x3ce
 *   b80300   mov ax,3      ; ef out dx,ax
 *   b80508   mov ax,0x805  ; ef out dx,ax
 *   b80700   mov ax,7      ; ef out dx,ax
 *   b808ff   mov ax,0xff08 ; ef out dx,ax
 *   c3       ret
 *
 * = outpw(0x3ce, 0x0003); outpw(0x3ce,0x0805); outpw(0x3ce,0x0007); outpw(0x3ce,0xff08);
 * outpw(port,val) 在 MSC 為 conio intrinsic -> mov dx,port; mov ax,val; out dx,ax。
 * 無 local、無迴圈 -> 應為 frameless;dx 為常數,MSC 可能每次重載 dx (待比對)。
 */
#include <conio.h>
#pragma intrinsic(outpw)

void sub_e91e(void)
{
    outpw(0x3ce, 0x0003);
    outpw(0x3ce, 0x0805);
    outpw(0x3ce, 0x0007);
    outpw(0x3ce, 0xff08);
}
