/* test_zhuyin.c — 注音輸入(查表 + 組字 + 挑字)驗證。 */
#include "dq3_zhuyin.h"
#include <stdio.h>

void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{ (void)t;(void)fb;(void)fb_w;(void)fb_h;(void)x;(void)y;(void)idx;(void)fg; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

static int has(const uint16_t *a, int n, int v) { int i; for(i=0;i<n;i++) if(a[i]==v) return 1; return 0; }

int main(void)
{
    uint16_t cand[32]; int n, i, idx;
    dq3_zhuyin z;

    printf("== 注音查表(ㄕˋ → 是 glyph 399)==\n");
    n = dq3_zhuyin_lookup(17, 0, 0, 3, cand, 32);   /* ㄕ(聲母17) + ˋ(4聲=tone3) */
    printf("  ㄕˋ 候選數 = %d\n", n);
    CHECK(n > 0, "ㄕˋ 有候選");
    CHECK(has(cand, n, 399), "候選含 是(glyph 399)");
    CHECK(dq3_zhuyin_lookup(99, 0, 0, 0, cand, 32) == 0, "不存在音節 → 0 候選");

    printf("== 組字流程:選ㄕ → 選ˋ → 挑「是」==\n");
    dq3_zhuyin_init(&z);
    z.cursor = 16; dq3_zhuyin_input(&z, 0x1c);       /* 聲母格 16 = ㄕ */
    CHECK(z.sh == 17, "選ㄕ → sh=17");
    CHECK(z.phase == DQ3_ZH_COMPOSE, "尚在組字");
    z.cursor = 39; dq3_zhuyin_input(&z, 0x1c);       /* 聲調格 39 = ˋ(tone3)→ 查表 */
    CHECK(z.phase == DQ3_ZH_PICK && z.ncand > 0, "選ˋ → 進候選窗");

    idx = -1; for (i = 0; i < z.ncand; i++) if (z.cand[i] == 399) idx = i;
    CHECK(idx >= 0, "候選窗含 是(399)");
    z.cand_cur = idx;
    {
        int g = dq3_zhuyin_input(&z, 0x1c);          /* Enter 選定 */
        CHECK(g == 399, "挑定 → 回 glyph 399(是)");
    }
    CHECK(z.phase == DQ3_ZH_COMPOSE && z.sh == 0 && z.tone == -1, "選定後清組字緩衝");

    printf("== ESC 放棄組字 ==\n");
    dq3_zhuyin_init(&z); z.cursor = 0; dq3_zhuyin_input(&z, 0x1c);  /* 選ㄅ */
    CHECK(z.sh == 1, "選ㄅ → sh=1");
    dq3_zhuyin_input(&z, 0x01);                       /* ESC */
    CHECK(z.sh == 0, "ESC → 清組字");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
