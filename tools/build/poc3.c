/* PoC round 3 — 逼近 sub_e6b9 的 codegen 細節。
 * 原版 (16 bytes, near ret):
 *   a15a0b   mov ax,[g]
 *   051890   add ax,0x9018
 *   d1c0 d1c0 d1c0   rol ax,1 x3
 *   a35a0b   mov [g],ax
 *   c3       ret
 * 重點:先載入暫存器 -> add -> rol(三次 rol ax,1)-> 回存。
 *
 * 以下逐一試不同 C 寫法,看哪個逼出 a1/05/d1c0³/a3 形態。
 */
unsigned g_rng;

/* v1:用區域變數承接,rol 用 ((x<<3)|(x>>13)) */
unsigned f1(void)
{
    unsigned x = g_rng;
    x += 0x9018;
    x = (x << 3) | (x >> 13);
    g_rng = x;
    return x;
}

/* v2:rol 用內建?TC 2.01 無 _rotl 宣告於 stdlib;試 char 化 */
unsigned f2(void)
{
    unsigned x = g_rng + 0x9018;
    x = (x << 3) | (x >> 13);
    return (g_rng = x);
}

/* v3:純加法 + 單次 shift,純粹比對 mov ax,[g];add ax,imm;mov [g],ax 形態 */
unsigned f3(void)
{
    g_rng += 0x9018;
    return g_rng;
}
