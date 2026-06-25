/* test_ship.c — 船系統:海 attr 指紋辨識 + 登船/航行/上岸/不上山 狀態機(docs/48)。
 * 自建 scene(免連 SDL)。attr:0=陸(0x02 可走)、1=海(0x25 BLK+0x20)、2=山(0x07 BLK 無 0x20)。 */
#include "dq3_ship.h"
#include "dq3_scene.h"
#include <stdio.h>
#include <string.h>

/* ---- stub:船測試不走素材路徑,僅供連結 dq3_scene.c ---- */
void dq3_set_palette(const dq3_color *pal, int count) { (void)pal; (void)count; }
int  dq3_charsprite_load(dq3_charsprite *cs, const char *a, const char *b, int e, char *err, int cap)
{ (void)cs; (void)a; (void)b; (void)e; (void)err; (void)cap; return -1; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

#define W 9
#define H 9
static uint8_t  idx[W*H];
static uint8_t  himap[W*H];
static uint16_t attr3[3] = { 0x0002, 0x0025, 0x0007 };  /* 0=陸 1=海 2=山 */

/* 左 3 欄陸地,其餘海;(6,4) 放一座山(海中礁/山,船不可越)。 */
static void build(dq3_scene *s)
{
    int x, y;
    memset(s, 0, sizeof *s);
    s->map_w = W; s->map_h = H;
    s->index_map = idx; s->hi_map = himap;
    s->attr = attr3; s->attr_count = 3;
    memset(himap, 0, sizeof himap);
    for (y = 0; y < H; y++)
        for (x = 0; x < W; x++)
            idx[y*W + x] = (x < 3) ? 0 : 1;     /* x<3 陸,否則海 */
    idx[4*W + 6] = 2;                            /* 海中山 */
    s->facing = 0;
}

int main(void)
{
    dq3_scene s; dq3_ship sh;

    printf("== 海 attr 指紋辨識 ==\n");
    build(&s);
    CHECK(dq3_ship_navigable(&s, 5, 4) == 1, "海格(0x25)可航行");
    CHECK(dq3_ship_navigable(&s, 1, 4) == 0, "陸地(0x02)不可航行");
    CHECK(dq3_ship_navigable(&s, 6, 4) == 0, "海中山(0x07)不可航行(無 0x20)");
    CHECK(dq3_ship_navigable(&s, -1, 4) == 0, "邊界外不可航行");
    CHECK(dq3_scene_walkable(&s, 5, 4) == 0, "海格步行被擋");

    printf("== 登船(踏上停泊船格)==\n");
    build(&s); dq3_ship_init(&sh);
    sh.owned = 1; sh.aboard = 0; sh.px = 3; sh.py = 4; sh.layer = 0;  /* 船停岸邊海格 */
    s.px = 2; s.py = 4;                                               /* 玩家在陸地 (2,4),右邊即船 */
    CHECK(dq3_ship_input(&s, &sh, 0x4d, 0) == DQ3_SHIP_BOARD, "向船格(右)→ 登船");
    CHECK(sh.aboard == 1 && s.px == 3 && s.py == 4, "登船後 aboard,玩家在船格");

    printf("== 航行 / 不上山 ==\n");
    CHECK(dq3_ship_input(&s, &sh, 0x4d, 0) == DQ3_SHIP_SAILED, "海上向右 → 航行");
    CHECK(s.px == 4 && s.py == 4, "航行後移動到 (4,4)");
    dq3_ship_input(&s, &sh, 0x4d, 0);                 /* (4,4)→(5,4) */
    CHECK(s.px == 5 && s.py == 4, "續航行到 (5,4),山在 (6,4) 右側");
    CHECK(dq3_ship_input(&s, &sh, 0x4d, 0) == DQ3_SHIP_BLOCKED, "向海中山(右)→ 擋,不上山");
    CHECK(s.px == 5 && s.py == 4, "撞山不移動");

    printf("== 上岸(船停原水格)==\n");
    build(&s); dq3_ship_init(&sh);
    sh.owned = 1; sh.aboard = 1; sh.layer = 0;
    s.px = 3; s.py = 4;                               /* 在岸邊海格,左即陸地 (2,4) */
    CHECK(dq3_ship_input(&s, &sh, 0x4b, 0) == DQ3_SHIP_DISEMBARK, "向陸地(左)→ 上岸");
    CHECK(sh.aboard == 0 && sh.px == 3 && sh.py == 4, "上岸後船停原水格 (3,4)");
    CHECK(s.px == 2 && s.py == 4, "玩家踏上陸地 (2,4)");

    printf("== 未取得船不能登船 ==\n");
    build(&s); dq3_ship_init(&sh);
    sh.owned = 0; sh.px = 3; sh.py = 4; sh.layer = 0;
    s.px = 2; s.py = 4;
    CHECK(dq3_ship_input(&s, &sh, 0x4d, 0) == DQ3_SHIP_BLOCKED, "未 owned:向船格 → 擋(海不可步行)");

    printf("\n%s (fail=%d)\n", g_fail ? "== FAIL ==" : "== ALL PASS ==", g_fail);
    return g_fail ? 1 : 0;
}
