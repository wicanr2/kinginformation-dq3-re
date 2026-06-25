/* dq3_rng.c — DQ3 亂數,雙模式(DOS 忠實 / REAL xorshift32)。見 dq3_rng.h。 */
#include "dq3_rng.h"

static int g_rng_mode = DQ3_RNG_DOS;   /* 預設忠實 */

void dq3_rng_set_mode(int mode) { g_rng_mode = (mode == DQ3_RNG_REAL) ? DQ3_RNG_REAL : DQ3_RNG_DOS; }
int  dq3_rng_get_mode(void) { return g_rng_mode; }

void dq3_rng_seed(dq3_rng *r, uint16_t seed)
{
    r->state = seed;
    /* 32-bit state 由 seed 混出,確保非 0(xorshift 要求)*/
    r->state32 = (uint32_t)seed * 2654435761u + 0x9e3779b9u;
    if (r->state32 == 0) r->state32 = 1u;
}

int dq3_rng_next(dq3_rng *r, int n)
{
    if (n == 0) return 0;                       /* L0e6c9: cmp bx,0; je → 0 */
    if (g_rng_mode == DQ3_RNG_REAL) {           /* 真實:xorshift32(週期 2³²-1)*/
        uint32_t x = r->state32;
        x ^= x << 13; x ^= x >> 17; x ^= x << 5;
        r->state32 = x;
        return (int)(x % (unsigned)n);
    }
    {                                           /* DOS 忠實:16-bit add/rol3 */
        uint16_t s = (uint16_t)(r->state + 0x9018);   /* add ax,0x9018 */
        s = (uint16_t)((s << 3) | (s >> 13));         /* rol ax,1 ×3 */
        r->state = s;                                 /* mov [0xb5a],ax */
        return (int)(s % (unsigned)n);                /* div bx → 餘數 */
    }
}
