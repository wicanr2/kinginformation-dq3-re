/* dq3_owportal.c — overworld 旗標條件 portal 表(RE 0x396e,docs/44 §7 / docs/45 §3.2)。
 * 逐段反組譯讀出:玩家在某 overworld 位置 → test_flag 鏈選目的 CTY(同點依進度變城)。
 * 旗標 id 對 [0x4f70] bit 陣列(dq3_storyflags)。 */
#include "dq3_owportal.h"

typedef struct { unsigned char flag, cty; } portal_branch;
typedef struct { unsigned char x, y, n, def_cty; portal_branch br[4]; } owportal;

/* cty_loc[58]&[83] 同在 (210,64);(82,165)=cty_loc75/47;(54,129)=cty_loc71/72。皆地表。 */
static const owportal PORTALS[] = {
    { 210,  64, 4, 83, { {0x23,58}, {0x47,59}, {0x42,60}, {0x48,61} } },  /* 0x39cb 鏈 */
    {  82, 165, 1, 47, { {0x13,75} } },
    {  54, 129, 1, 72, { {0x4d,71} } },
};
#define NPORTAL ((int)(sizeof PORTALS / sizeof PORTALS[0]))

int dq3_owportal_resolve(int x, int y, const dq3_storyflags *flags)
{
    int i, k;
    for (i = 0; i < NPORTAL; i++) {
        if (PORTALS[i].x != x || PORTALS[i].y != y) continue;
        for (k = 0; k < PORTALS[i].n; k++)                 /* 旗標鏈:set 即取該城 */
            if (flags && dq3_flags_get(flags, PORTALS[i].br[k].flag))
                return PORTALS[i].br[k].cty;
        return PORTALS[i].def_cty;                          /* 無旗標 → 預設變體 */
    }
    return -1;
}
