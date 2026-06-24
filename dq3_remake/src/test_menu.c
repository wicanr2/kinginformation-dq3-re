/* test_menu.c — 通用選單導航驗證(不繪製 → 對 draw_glyph 給 stub)。 */
#include "dq3_menu.h"
#include <stdio.h>

/* stub:nav 測試不繪製 */
void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{ (void)t;(void)fb;(void)fb_w;(void)fb_h;(void)x;(void)y;(void)idx;(void)fg; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(void)
{
    dq3_menu m;
    uint16_t a[2] = {106, 187};   /* 勇者 */
    int i;

    printf("== 選單導航 ==\n");
    dq3_menu_init(&m, 40, 40);
    for (i = 0; i < 4; i++) dq3_menu_add(&m, a, 2);
    CHECK(m.n_items == 4, "加 4 項");
    CHECK(m.cursor == 0, "游標起始 0");

    dq3_menu_input(&m, 0x50);                 /* 下 */
    CHECK(m.cursor == 1, "下 → 1");
    dq3_menu_input(&m, 0x48);                 /* 上 */
    CHECK(m.cursor == 0, "上 → 0");
    dq3_menu_input(&m, 0x48);                 /* 上(從 0 繞回尾)*/
    CHECK(m.cursor == 3, "上繞回尾 → 3");
    dq3_menu_input(&m, 0x50);                 /* 下(從尾繞回頭)*/
    CHECK(m.cursor == 0, "下繞回頭 → 0");

    dq3_menu_input(&m, 0x50); dq3_menu_input(&m, 0x50);  /* → 2 */
    CHECK(dq3_menu_input(&m, 0x1c) == 2, "Enter → 回選中 index 2");
    CHECK(dq3_menu_input(&m, 0x39) == -1, "非導航鍵 → -1");

    printf("== 滿項防呆 ==\n");
    { dq3_menu f; int k, last = 0;
      dq3_menu_init(&f, 0, 0);
      for (k = 0; k < DQ3_MENU_MAX_ITEMS + 3; k++) last = dq3_menu_add(&f, a, 2);
      CHECK(f.n_items == DQ3_MENU_MAX_ITEMS, "上限後不再加");
      CHECK(last == -1, "超量加 → -1");
    }

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
