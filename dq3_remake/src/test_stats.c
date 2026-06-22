/* test_stats.c — 數值/升級系統單元測試:示範 #4/#5/#6 的原版 bug vs C 修正。
 *
 * 這是 rule 60 要求的「快速、決定性 pass/fail 訊號」:證明 C 重寫層確實修好了
 * binary patch 修不了的 #5/#6,以及 #4 的成長提升。
 * 用法:dq3_stats_test [dir]。資料用內建 dq3_exedata,不需 DQ3.EXE。回傳 0=全通過。
 */
#include "dq3_stats.h"
#include <stdio.h>
#include <string.h>

static int g_fail = 0;
#define CHECK(cond, msg) do { \
    if (cond) { printf("  [PASS] %s\n", msg); } \
    else { printf("  [FAIL] %s\n", msg); g_fail++; } } while (0)

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    char err[256] = {0};
    dq3_stats orig, fix;
    uint16_t mp_orig, mp_fix;
    int lv_fixed, lv_buggy;
    uint32_t big_exp;

    if (dq3_stats_load(&orig, assets, 0, err, sizeof err) != 0) {
        fprintf(stderr, "load(orig) failed: %s\n", err); return 2;
    }
    if (dq3_stats_load(&fix, assets, 1, err, sizeof err) != 0) {
        fprintf(stderr, "load(fix) failed: %s\n", err); return 2;
    }

    printf("== Bug #4:勇者 MaxMP 成長 ==\n");
    mp_orig = dq3_stats_growth_target(&orig, 0, DQ3_STAT_MP, 43);
    mp_fix  = dq3_stats_growth_target(&fix,  0, DQ3_STAT_MP, 43);
    printf("  勇者 Lv43 MaxMP target: 原版=%u  修正=%u\n", mp_orig, mp_fix);
    CHECK(mp_orig < 130, "原版勇者 Lv43 MaxMP 偏低(<130,放不出比荷瑪拉)");
    CHECK(mp_fix >= 200, "修正後勇者 Lv43 MaxMP >=200(可放全體補血咒)");
    CHECK(mp_fix > mp_orig, "修正後 MaxMP 高於原版");

    printf("== Bug #5:高等級升級錯亂(越界連升)==\n");
    /* 給超過 Lv43 門檻的經驗,觀察是否連升越界 */
    big_exp = orig.thresh[0][DQ3_MAX_LEVEL] + 1000000u;
    lv_fixed = dq3_stats_level_for_exp(&fix,  0, big_exp, 1);
    lv_buggy = dq3_stats_level_for_exp(&orig, 0, big_exp, 0);
    printf("  exp=%u  →  修正版 level=%d  原版(bug)level=%d\n", big_exp, lv_fixed, lv_buggy);
    CHECK(lv_fixed == DQ3_MAX_LEVEL, "修正版 level 夾在 43(不越界)");
    CHECK(lv_buggy > DQ3_MAX_LEVEL, "原版越界連升 → level 衝破 43(重現 bug)");

    printf("== Bug #6:屬性 255 溢位 ==\n");
    /* 原版 8-bit:200+100=300 會 wrap 成 44;C 修正版用 u16+clamp 維持 300 */
    {
        uint16_t fixed_val = dq3_stats_add_clamped(200, 100);
        uint8_t  wrap_val  = (uint8_t)(200 + 100);     /* 模擬原版 byte 截斷 */
        printf("  200+100:  原版(byte wrap)=%u  修正(u16)=%u\n", wrap_val, fixed_val);
        CHECK(wrap_val == 44, "原版 byte 截斷 200+100 → 44(重現 wrap)");
        CHECK(fixed_val == 300, "修正版 u16 維持 300(不 wrap)");
        CHECK(dq3_stats_add_clamped(9000, 5000) == DQ3_STAT_CAP, "超上限 clamp 到 9999");
    }

    printf("\n%s (%d failures)\n", g_fail ? "== 有測試未通過 ==" : "== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
