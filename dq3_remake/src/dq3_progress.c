/* dq3_progress.c — remake 主線進度旗標流(#1)。順序取自杜勝利攻略。 */
#include "dq3_progress.h"

/* 里程碑順序(= 攻略進度);RAINBOW/DESCEND 鏡射既有 EXE 旗標。 */
static const int MS_SEQ[DQ3_MS_COUNT] = {
    DQ3_MS_START, DQ3_MS_THIEF_KEY, DQ3_MS_MAGIC_BALL, DQ3_MS_ROMALY, DQ3_MS_DHAMA,
    DQ3_MS_SHIP, DQ3_MS_RAINBOW, DQ3_MS_DESCEND, DQ3_MS_ZOMA
};
static const char *MS_NAME[DQ3_MS_COUNT + 1] = {
    "未開始", "阿里阿罕開場", "盜賊鑰匙", "魔法球", "羅馬利亞", "達瑪轉職",
    "取船", "彩虹水滴", "下降アレフガルド", "索瑪(終盤)"
};

/* RAINBOW/DESCEND 里程碑同步既有 EXE 旗標(避免兩套狀態)。 */
static int eff_flag(int ms)
{
    if (ms == DQ3_MS_RAINBOW) return DQ3_FLAG_RAINBOW_SYNTHED;
    if (ms == DQ3_MS_DESCEND) return DQ3_FLAG_DESCENDED;
    return ms;
}

int dq3_progress_done(const dq3_storyflags *f, int milestone_flag)
{
    return f ? dq3_flags_get(f, eff_flag(milestone_flag)) : 0;
}

void dq3_progress_set(dq3_storyflags *f, int milestone_flag)
{
    if (f) dq3_flags_set(f, eff_flag(milestone_flag), 1);
}

int dq3_progress_stage(const dq3_storyflags *f)
{
    int i;
    for (i = 0; i < DQ3_MS_COUNT; i++)
        if (!dq3_progress_done(f, MS_SEQ[i])) return i;   /* 第一個未完成 = 目前階段 */
    return DQ3_MS_COUNT;
}

const char *dq3_progress_name(int stage)
{
    if (stage < 0) stage = 0;
    if (stage > DQ3_MS_COUNT) stage = DQ3_MS_COUNT;
    return MS_NAME[stage];
}
