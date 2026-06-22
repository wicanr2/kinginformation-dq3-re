/* dq3_packbg.c — 戰鬥背景 PACKBG.SCR 解碼實作。 */
#include "dq3_packbg.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int dq3_packbg_decode(const char *dir, int page,
                      uint8_t out[DQ3_PACKBG_H][DQ3_PACKBG_W], char *err, int errcap)
{
    char path[2048]; FILE *f; long pos; size_t need = 0x6e00; uint8_t *buf;
    int r, pl, b, bit, pb = DQ3_PACKBG_W / 8;   /* 80 */
    size_t o = 0;
    #define FAIL(m) do{ if(err)snprintf(err,errcap,"%s",m); if(f)fclose(f); free(buf); return -1; }while(0)

    buf = (uint8_t *)malloc(need);
    f = NULL;
    if (!buf) { if(err)snprintf(err,errcap,"OOM"); return -1; }
    snprintf(path, sizeof path, "%s/PACKBG.SCR", dir);
    f = fopen(path, "rb");
    if (!f) FAIL("open PACKBG.SCR");
    pos = (long)page * 0x3d80 + page;
    if (fseek(f, pos, SEEK_SET) != 0) FAIL("seek");
    if (fread(buf, 1, need, f) != need) FAIL("read");
    fclose(f); f = NULL;

    memset(out, 0, (size_t)DQ3_PACKBG_H * DQ3_PACKBG_W);
    for (r = 0; r < DQ3_PACKBG_H; r++)
        for (pl = 0; pl < 4; pl++)
            for (b = 0; b < pb; b++) {
                uint8_t v = (o < need) ? buf[o] : 0; o++;
                for (bit = 0; bit < 8; bit++)
                    if (v & (0x80 >> bit)) out[r][b*8 + bit] |= (uint8_t)(1 << pl);
            }
    free(buf);
    return 0;
    #undef FAIL
}
