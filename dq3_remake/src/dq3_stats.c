/* dq3_stats.c — 角色數值/升級系統實作(成長表 + 門檻表 + #4/#5/#6 修正)。 */
#include "dq3_stats.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* DQ3.EXE DGROUP 定位(docs/05 / docs/23) */
#define EXE_HDR      0x1370
#define EXE_DGROUP   0x14dd
#define GROWTH_DS    0x4366
#define THR_PTR_DS   0x43d6
#define GROWTH_FILE  (EXE_HDR + EXE_DGROUP * 16 + GROWTH_DS)   /* 0x1a4a6 */
#define THR_PTR_FILE (EXE_HDR + EXE_DGROUP * 16 + THR_PTR_DS)  /* 0x1a516 */
#define DS_TO_FILE(off) (EXE_HDR + EXE_DGROUP * 16 + (off))

static uint16_t rd16(const uint8_t *d, size_t o) { return (uint16_t)(d[o] | (d[o+1] << 8)); }
static uint32_t rd32(const uint8_t *d, size_t o) {
    return (uint32_t)d[o] | ((uint32_t)d[o+1] << 8) | ((uint32_t)d[o+2] << 16) | ((uint32_t)d[o+3] << 24);
}

int dq3_stats_load(dq3_stats *st, const char *assets_dir, int apply_bug4_fix,
                   char *err, int errcap)
{
    char path[2048]; FILE *f; long sz; uint8_t *d = NULL; size_t n;
    int c, lv;
    #define FAIL(m) do { if (err) snprintf(err, errcap, "%s", m); free(d); return -1; } while (0)

    memset(st, 0, sizeof *st);
    snprintf(path, sizeof path, "%s/DQ3.EXE", assets_dir);
    f = fopen(path, "rb");
    if (!f) FAIL("open DQ3.EXE failed");
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    if (sz <= 0) { fclose(f); FAIL("DQ3.EXE empty"); }
    d = (uint8_t *)malloc((size_t)sz);
    if (!d || fread(d, 1, (size_t)sz, f) != (size_t)sz) { fclose(f); FAIL("read DQ3.EXE"); }
    fclose(f); n = (size_t)sz;

    /* 成長表:8 職業 × 14 byte */
    if (GROWTH_FILE + DQ3_NUM_CLASS * 14 > n) FAIL("growth table OOB");
    for (c = 0; c < DQ3_NUM_CLASS; c++)
        memcpy(st->growth[c], d + GROWTH_FILE + c * 14, 14);

    /* 門檻表:指標表 8 × u16(DS off)→ 各 44 entry × u32 */
    if (THR_PTR_FILE + DQ3_NUM_CLASS * 2 > n) FAIL("threshold ptr table OOB");
    for (c = 0; c < DQ3_NUM_CLASS; c++) {
        uint16_t ds_off = rd16(d, THR_PTR_FILE + c * 2);
        size_t base = DS_TO_FILE(ds_off);
        if (base + (size_t)(DQ3_MAX_LEVEL + 1) * 4 > n) FAIL("threshold table OOB");
        for (lv = 0; lv <= DQ3_MAX_LEVEL; lv++)
            st->thresh[c][lv] = rd32(d, base + (size_t)lv * 4);
    }

    /* #4 勇者(class0)MP 成長修正:base 列+2、slope 列+3 */
    st->apply_bug4_fix = apply_bug4_fix;
    if (apply_bug4_fix) {
        st->growth[0][2] = 0x08;   /* MP base 3→8 */
        st->growth[0][3] = 0x0a;   /* MP slope 5→10 */
    }

    st->loaded = 1;
    free(d);
    return 0;
    #undef FAIL
}

uint16_t dq3_stats_growth_target(const dq3_stats *st, int cls, dq3_stat_kind kind, int level)
{
    int base_off = (int)kind * 2;     /* 每屬性 2 byte:base, slope */
    uint16_t base, slope, t;
    if (cls < 0 || cls >= DQ3_NUM_CLASS) return 0;
    base  = st->growth[cls][base_off];
    slope = st->growth[cls][base_off + 1];
    /* target = base + (slope×level)/2,16-bit(對齊 mul→shr→add 路徑) */
    t = (uint16_t)(base + (uint16_t)((slope * (uint16_t)level) >> 1));
    return t;
}

int dq3_stats_level_for_exp(const dq3_stats *st, int cls, uint32_t exp, int fixed)
{
    int level = 1;
    if (cls < 0 || cls >= DQ3_NUM_CLASS) return 1;

    if (fixed) {
        /* #5 修正:level 夾在 [1,43];達上限即停,絕不越界讀表 */
        while (level < DQ3_MAX_LEVEL && exp >= st->thresh[cls][level + 1])
            level++;
        return level;
    }

    /* 原版 bug 重現:不 clamp;level 超過 43 後越界讀「下一職業 entry[0]=0」
     * → exp>=0 永遠成立 → 連升。以 cap(99)防無限迴圈;回傳 >43 代表 runaway。 */
    {
        const uint32_t *flat = &st->thresh[0][0];   /* 8×44 連續 */
        int max_idx = DQ3_NUM_CLASS * (DQ3_MAX_LEVEL + 1) - 1;
        int base = cls * (DQ3_MAX_LEVEL + 1);
        while (level < 99) {
            int idx = base + level + 1;             /* 查 threshold[level+1] */
            uint32_t thr;
            if (idx > max_idx) break;               /* 真的掃出全表才停 */
            thr = flat[idx];                        /* 越界時讀到相鄰職業 entry */
            if (exp >= thr) level++; else break;
        }
        return level;
    }
}

uint16_t dq3_stats_add_clamped(uint16_t cur, int delta)
{
    int v = (int)cur + delta;     /* #6:寬型別運算 */
    if (v < 0) v = 0;
    if (v > DQ3_STAT_CAP) v = DQ3_STAT_CAP;   /* 顯式上限,不 wrap */
    return (uint16_t)v;
}
