/* test_door.c — 鎖門開鎖驗證(docs/35 §八;鏡射 DQ3.EXE 0x4906/0x4977)。
 *
 * 只測 dq3_scene 的純邏輯門函式;為免拖入 SDL/sprite,對 dq3_scene.c 參照但
 * 本測試未走到的兩個外部符號提供 stub。 */
#include "dq3_scene.h"
#include "dq3_sprite.h"
#include "dq3_runtime.h"
#include <stdio.h>
#include <string.h>

/* ---- stub:door 測試不會呼叫,僅供連結 ---- */
void dq3_set_palette(const dq3_color *pal, int count) { (void)pal; (void)count; }
int  dq3_charsprite_load(dq3_charsprite *cs, const char *a, int e, char *err, int cap)
{ (void)cs; (void)a; (void)e; (void)err; (void)cap; return -1; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

/* 3×3 場景:中心 (1,1) 放一扇 tier-N 鎖門;周圍為通行地板。 */
static uint8_t  idx[9];
static uint8_t  hi[9];
static uint16_t attr[4];   /* tile index 0..3 的屬性 */

static void build(dq3_scene *s, int door_tier, uint8_t door_subid)
{
    int i;
    memset(s, 0, sizeof *s);
    s->map_w = s->map_h = 3;
    s->index_map = idx; s->hi_map = hi;
    s->attr = attr; s->attr_count = 4;
    for (i = 0; i < 9; i++) { idx[i] = 0; hi[i] = 0; }   /* tile 0 = 地板 */
    attr[0] = 0x0000;                 /* 地板:可走、非門 */
    attr[1] = (uint16_t)(door_tier << 6);   /* tile 1 = 門:bits6-7 = 所需等級 */
    attr[2] = (uint16_t)((door_subid & 0x1f) ? 0 : 0); /* 開門後的目標 tile(可走)*/
    /* 門放中心 (1,1):index=1,高 byte subid = 開門後 tile index */
    idx[1*3+1] = 1;
    hi[1*3+1]  = door_subid;          /* 高 byte:低 5 bit = 開門後 index、高位保留 */
}

int main(void)
{
    dq3_scene s;

    printf("== 鎖門所需等級(0x4906 test attr,0xc0)==\n");
    build(&s, 2, 0x02);
    CHECK(dq3_scene_door_tier(&s, 1, 1) == 2, "門 attr bits6-7=2 → 所需等級 2");
    CHECK(dq3_scene_door_tier(&s, 0, 0) == 0, "地板 → 非門(0)");
    CHECK(dq3_scene_door_tier(&s, 9, 9) == 0, "界外 → 0");

    printf("== 試開面向門(鑰匙等級閘)==\n");
    build(&s, 2, 0x02);
    s.px = 1; s.py = 0; s.facing = 0;        /* 站門上方、面向下 → 面向格=(1,1) */
    CHECK(dq3_scene_try_open_facing_door(&s, 1) == 0, "鑰匙等級1 < 門需求2 → 開不了");
    CHECK(s.index_map[1*3+1] == 1, "失敗時門 tile 不變");
    CHECK(dq3_scene_try_open_facing_door(&s, 2) == 1, "鑰匙等級2 = 需求 → 開門");
    CHECK(dq3_scene_door_tier(&s, 1, 1) == 0, "開門後該格不再是鎖門");

    printf("== 開門 tile 變換(0x4977/0x4985)==\n");
    build(&s, 1, 0x62);   /* subid=0x62:低5bit=2(新 index)、高位 0x60 */
    CHECK(dq3_scene_open_door(&s, 1, 1) == 1, "開門回 1");
    CHECK(s.index_map[1*3+1] == (0x62 & 0x1f), "新 index = 高 byte & 0x1f");
    CHECK(s.hi_map[1*3+1] == (0x62 & 0xe0), "高 byte &= 0xe0(清事件位、留轉場位)");
    CHECK(dq3_scene_open_door(&s, 0, 0) == 0, "對非門開門 → 0");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
