/* byte-match PoC — Turbo C 2.01 codegen 對原版 DQ3.EXE leaf 函式。
 *
 * 原版觀察:函式多半「無 BP frame、無參數、只存取 [DS:abs] 全域、near ret(c3)」。
 * 此檔用「無參數 + 全域」風格逼近,比對 codegen 形態。
 *
 * 原版 sub_e6b9(亂數狀態前進,16 bytes):
 *   a15a0b 051890 d1c0 d1c0 d1c0 a35a0b c3
 *   = ax=[g]; ax+=0x9018; rol ax,1; rol ax,1; rol ax,1; [g]=ax; ret
 */

unsigned g_rng;          /* 對應 [0xb5a] */
volatile unsigned g_tick;/* 對應 [0x05] BIOS tick(形態比對) */

/* A:純全域 LCG advance,無參數。期望出 rol ax,1 ×3。 */
unsigned rng_adv(void)
{
    /* (x<<3)|(x>>13) = rol 16-bit by 3。TC 2.01 對 unsigned 常量旋轉的 codegen 觀察點。 */
    g_rng += 0x9018;
    g_rng = (g_rng << 3) | (g_rng >> 13);
    return g_rng;
}

/* B:忙等 tick==7,對應 sub_683a(無 cli/sti,純 C 版做形態比對)。 */
void wait_tick7(void)
{
    g_tick = 0;
    while (g_tick != 7)
        ;
}

/* C:對應 sub_32a3 — 把 [bx+0x265d] 連續 13 個 word 設 0xff(步進 2)。
 *   原版:cx=0xd; bx=8; loop{ [bx+0x265d]=0xff(byte); bx+=2 }。
 *   注意原版寫的是 byte(c687),陣列基底 0x265d、步進 2。 */
unsigned char g_arr[64];
void fill_ff(void)
{
    int i;
    for (i = 0; i < 13; i++)
        g_arr[i * 2] = 0xff;
}
