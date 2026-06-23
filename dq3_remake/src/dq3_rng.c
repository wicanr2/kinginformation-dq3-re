/* dq3_rng.c — DQ3.EXE 亂數(鏡射 file 0xfa39)。 */
#include "dq3_rng.h"

void dq3_rng_seed(dq3_rng *r, uint16_t seed) { r->state = seed; }

int dq3_rng_next(dq3_rng *r, int n)
{
    uint16_t s;
    if (n == 0) return 0;                       /* L0e6c9: cmp bx,0; je → dx=0 */
    s = (uint16_t)(r->state + 0x9018);          /* add ax,0x9018 */
    s = (uint16_t)((s << 3) | (s >> 13));       /* rol ax,1 ×3 */
    r->state = s;                               /* mov [0xb5a],ax */
    return (int)(s % (unsigned)n);              /* div bx → 餘數 */
}
