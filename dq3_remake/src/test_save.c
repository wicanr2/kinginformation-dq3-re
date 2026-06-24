/* test_save.c — 存檔/讀檔 round-trip 驗證(寫 → 讀 → 比對一致)。 */
#include "dq3_save.h"
#include <stdio.h>
#include <string.h>

void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{ (void)t;(void)fb;(void)fb_w;(void)fb_h;(void)x;(void)y;(void)idx;(void)fg; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(void)
{
    const char *path = "/tmp/dq3_test_save.dat";
    dq3_stats st; dq3_roster r; dq3_party p; dq3_inventory inv;
    dq3_save_pos pos = { 5, 12, 8 };
    uint16_t nm[2] = {15, 16};

    if (dq3_stats_load(&st, NULL, 1, NULL, 0) < 0) { printf("stats FAIL\n"); return 1; }
    dq3_roster_init(&r); dq3_party_init(&p); dq3_inv_init(&inv);
    dq3_roster_create(&r, &st, 1, DQ3_GENDER_MALE, nm, 2);     /* 戰士 AB */
    dq3_roster_create(&r, &st, 4, DQ3_GENDER_FEMALE, nm, 2);   /* 魔法使者 */
    dq3_party_add(&p, &r, 0); dq3_party_add(&p, &r, 1);
    dq3_inv_add(&inv, ITEM_SUN_STONE); dq3_inv_add(&inv, ITEM_KEY_MAGIC);

    printf("== 寫存檔 ==\n");
    CHECK(dq3_save_write(path, &r, &p, &inv, pos) == 0, "寫檔成功");
    CHECK(dq3_save_exists(path), "存檔存在 + magic 正確");
    CHECK(!dq3_save_exists("/tmp/dq3_nonexistent_xyz.dat"), "不存在檔 → exists=0");

    printf("== 讀回 + 比對 ==\n");
    {
        dq3_roster r2; dq3_party p2; dq3_inventory inv2; dq3_save_pos pos2;
        memset(&r2, 0xcc, sizeof r2);   /* 先填髒值,確認真的被覆蓋 */
        CHECK(dq3_save_read(path, &r2, &p2, &inv2, &pos2) == 0, "讀檔成功");
        CHECK(r2.count == r.count, "名冊人數一致");
        CHECK(r2.list[0].m.cls == 1 && r2.list[0].gender == DQ3_GENDER_MALE, "成員0:戰士 男 一致");
        CHECK(r2.list[1].m.cls == 4 && r2.list[1].gender == DQ3_GENDER_FEMALE, "成員1:魔法使者 女 一致");
        CHECK(r2.list[0].name[0] == 15 && r2.list[0].name[1] == 16, "名字一致");
        CHECK(r2.list[0].m.stat[0] == r.list[0].m.stat[0], "數值一致");
        CHECK(p2.count == 2 && p2.slot[0] == 0 && p2.slot[1] == 1, "隊伍一致");
        CHECK(dq3_inv_has(&inv2, ITEM_SUN_STONE) && dq3_inv_has(&inv2, ITEM_KEY_MAGIC), "道具一致");
        CHECK(pos2.cty == 5 && pos2.px == 12 && pos2.py == 8, "位置一致");
        CHECK(memcmp(&r, &r2, sizeof r) == 0, "名冊整塊 byte 一致");
    }

    printf("== 防呆:壞檔 / 缺檔 ==\n");
    CHECK(dq3_save_read("/tmp/dq3_nonexistent_xyz.dat", &r, &p, &inv, &pos) == -1, "缺檔 → -1");
    { FILE *f = fopen(path, "wb"); fwrite("GARBAGE!", 1, 8, f); fclose(f); }
    CHECK(dq3_save_read(path, &r, &p, &inv, &pos) == -2, "壞 magic → -2");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
