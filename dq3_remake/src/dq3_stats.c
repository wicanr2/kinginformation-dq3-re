/* dq3_stats.c — 角色數值/升級系統實作(成長表 + 門檻表 + #4/#5/#6 修正)。
 * 資料改用 dq3_exedata(自 DQ3.EXE DGROUP 抽出編入,生成檔),不再執行期讀 DQ3.EXE。 */
#include "dq3_stats.h"
#include "dq3_exedata.h"
#include <stdio.h>
#include <string.h>

int dq3_stats_load(dq3_stats *st, const char *assets_dir, int apply_bug4_fix,
                   char *err, int errcap)
{
    int c, lv;
    (void)assets_dir; (void)err; (void)errcap;   /* 不再需要 DQ3.EXE */
    memset(st, 0, sizeof *st);

    /* 成長表 + 門檻表:自內建 dq3_exedata 複製 */
    for (c = 0; c < DQ3_NUM_CLASS; c++) {
        memcpy(st->growth[c], dq3x_growth[c], 14);
        for (lv = 0; lv <= DQ3_MAX_LEVEL; lv++)
            st->thresh[c][lv] = dq3x_thresh[c][lv];
    }

    /* #4 勇者(class0)MP 成長修正:base 列+2、slope 列+3 */
    st->apply_bug4_fix = apply_bug4_fix;
    if (apply_bug4_fix) {
        st->growth[0][2] = 0x08;   /* MP base 3→8 */
        st->growth[0][3] = 0x0a;   /* MP slope 5→10 */
    }

    st->loaded = 1;
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

/* #6 各屬性上限:主屬性(力量/速度/耐力/聰明/運氣)為 byte → 上限 255(原版超 255 wrap 歸 0 = bug);
 * HP/MP 為 3 位數值 → 上限 999。clamp(不 wrap)即 #6 修正。 */
int dq3_stat_cap_for(dq3_stat_kind k)
{
    return (k == DQ3_STAT_HP || k == DQ3_STAT_MP) ? DQ3_STAT_HPMP_CAP : DQ3_STAT_PRIMARY_CAP;
}
static uint16_t add_clamped_kind(uint16_t cur, int delta, dq3_stat_kind k)
{
    int v = (int)cur + delta, cap = dq3_stat_cap_for(k);
    if (v < 0) v = 0;
    if (v > cap) v = cap;          /* #6:依屬性別 clamp,不 wrap */
    return (uint16_t)v;
}

int dq3_member_change_class(dq3_member *m, const dq3_stats *st, int new_cls)
{
    int k;
    if (!m || new_cls <= 0 || new_cls >= DQ3_NUM_CLASS) return -1;  /* 0=勇者,不可轉成勇者 */
    if (m->cls == 0) return -1;                                     /* 勇者不可轉職 */
    m->cls = new_cls;
    m->level = 1;
    m->exp = st->thresh[new_cls][1];                /* level1 門檻 */
    for (k = 0; k < DQ3_STAT_COUNT; k++)
        m->stat[k] = (uint16_t)(m->stat[k] / 2);    /* DQ3 換職:保留一半屬性 */
    m->cur_hp = m->stat[DQ3_STAT_HP];               /* 重置滿 */
    m->cur_mp = m->stat[DQ3_STAT_MP];
    return 0;
}

void dq3_member_init(dq3_member *m, const dq3_stats *st, int cls, int level)
{
    int k;
    if (level < 1) level = 1;
    if (level > DQ3_MAX_LEVEL) level = DQ3_MAX_LEVEL;
    m->cls = cls; m->level = level;
    m->exp = st->thresh[cls][level];                  /* 該等級的門檻經驗 */
    for (k = 0; k < DQ3_STAT_COUNT; k++) {
        int t = dq3_stats_growth_target(st, cls, (dq3_stat_kind)k, level);
        m->stat[k] = add_clamped_kind(0, t, (dq3_stat_kind)k);   /* #6:初始值也 clamp(高階主屬性 > 255)*/
    }
    m->cur_hp = m->stat[DQ3_STAT_HP];   /* 起始滿血/滿魔 */
    m->cur_mp = m->stat[DQ3_STAT_MP];
}

int dq3_member_gain_exp(dq3_member *m, const dq3_stats *st, uint32_t add)
{
    int newlv, gained = 0, k;
    m->exp += add;                                              /* uint32,不溢位 */
    newlv = dq3_stats_level_for_exp(st, m->cls, m->exp, 1);     /* #5:夾在 [1,43] */
    while (m->level < newlv) {
        int lv0 = m->level, lv1 = m->level + 1;
        for (k = 0; k < DQ3_STAT_COUNT; k++) {                              /* 逐屬性套成長 delta */
            int t0 = dq3_stats_growth_target(st, m->cls, (dq3_stat_kind)k, lv0);
            int t1 = dq3_stats_growth_target(st, m->cls, (dq3_stat_kind)k, lv1);
            int delta = t1 - t0; if (delta < 0) delta = 0;
            m->stat[k] = add_clamped_kind(m->stat[k], delta, (dq3_stat_kind)k);  /* #6:依屬性別 clamp 255/999 */
            if (k == DQ3_STAT_HP) m->cur_hp = add_clamped_kind(m->cur_hp, delta, DQ3_STAT_HP);  /* 升級加血到 cur */
            if (k == DQ3_STAT_MP) m->cur_mp = add_clamped_kind(m->cur_mp, delta, DQ3_STAT_MP);
        }
        m->level = lv1; gained++;
    }
    return gained;
}
