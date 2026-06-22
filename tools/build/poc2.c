/* PoC round 2 — 精準逼近 sub_683a 忙等。
 * 原版 (19 bytes, near ret):
 *   fa            cli
 *   c70605000000  mov word [0x0005], 0
 *   fb            sti
 * .l:
 *   fa            cli
 *   a10500        mov ax, [0x0005]
 *   fb            sti
 *   3d0700        cmp ax, 7
 *   75f6          jne .l
 *   c3            ret
 *
 * 推測 C:對絕對位址 0x0005 的 16-bit volatile,以 disable()/enable() 包夾。
 *   disable() = inline cli (TC intrinsic),enable() = inline sti。
 *   *(volatile int *)5 在 small/compact model = DS:0005 → a10500 / c70605。
 */
#include <dos.h>   /* disable(), enable() — 產生 cli/sti */

#define TICK (*(volatile int *)5)   /* small model: DS:0005 */

void wait_tick7(void)
{
    disable();
    TICK = 0;
    enable();
    do {
        disable();
        /* 強制讀進暫存器再比較 */
    } while ((enable(), TICK) != 7);
}

/* 變體:用區域變數承接讀值,逼出 mov ax,[5]; cmp ax,7 */
void wait_tick7b(void)
{
    int x;
    disable();
    TICK = 0;
    enable();
    for (;;) {
        disable();
        x = TICK;
        enable();
        if (x == 7) break;
    }
}
