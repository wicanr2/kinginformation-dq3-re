/* test_nameinput.c — 英數姓名輸入 widget 驗證(不繪製 → draw_glyph stub)。 */
#include "dq3_nameinput.h"
#include <stdio.h>

void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{ (void)t;(void)fb;(void)fb_w;(void)fb_h;(void)x;(void)y;(void)idx;(void)fg; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

/* 把游標移到指定 cell(直接設,測 commit 用)*/
static void at(dq3_nameinput *ni, int cell) { ni->cursor = cell; }

int main(void)
{
    dq3_nameinput ni;

    printf("== cell → glyph 對映(英數)==\n");
    CHECK(dq3_nameinput_cell_glyph(0) == 0,  "cell0 → glyph0(數字0)");
    CHECK(dq3_nameinput_cell_glyph(9) == 9,  "cell9 → glyph9(數字9)");
    CHECK(dq3_nameinput_cell_glyph(10) == 15, "cell10 → glyph15(A)");
    CHECK(dq3_nameinput_cell_glyph(35) == 40, "cell35 → glyph40(Z)");
    CHECK(dq3_nameinput_cell_glyph(DQ3_NI_CELL_OK) == 0xffff, "OK 格 → 非字元");

    printf("== 選字 / 退格 / 完成 ==\n");
    dq3_nameinput_init(&ni);
    at(&ni, 10); dq3_nameinput_input(&ni, 0x1c);   /* A */
    at(&ni, 11); dq3_nameinput_input(&ni, 0x1c);   /* B */
    at(&ni, 2);  dq3_nameinput_input(&ni, 0x1c);   /* 2 */
    CHECK(ni.len == 3 && ni.buf[0]==15 && ni.buf[1]==16 && ni.buf[2]==2, "輸入 A B 2 → 緩衝 {15,16,2}");
    at(&ni, DQ3_NI_CELL_BS); dq3_nameinput_input(&ni, 0x1c);  /* 退格 */
    CHECK(ni.len == 2, "退格 → len 2");
    at(&ni, DQ3_NI_CELL_OK); dq3_nameinput_input(&ni, 0x1c);  /* 完成 */
    CHECK(ni.done == 1, "OK → done");

    printf("== 上限 / 導航 ==\n");
    dq3_nameinput_init(&ni);
    { int k; for (k = 0; k < 8; k++) { at(&ni, 0); dq3_nameinput_input(&ni, 0x1c); } }
    CHECK(ni.len == DQ3_NI_NAME_MAX, "超過上限不再加");
    dq3_nameinput_init(&ni);
    dq3_nameinput_input(&ni, 0x4d);   /* 右 */
    CHECK(ni.cursor == 1, "右 → cell1");
    dq3_nameinput_input(&ni, 0x50);   /* 下 */
    CHECK(ni.cursor == 1 + DQ3_NI_COLS, "下 → +9");
    dq3_nameinput_input(&ni, 0x4b);   /* 左 */
    CHECK(ni.cursor == DQ3_NI_COLS, "左 → cell9");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
