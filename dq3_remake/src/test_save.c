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
    dq3_stats st; dq3_roster r; dq3_party p; dq3_inventory inv; dq3_storyflags fl;
    dq3_save_pos pos = { 5, 12, 8, 1, 0, 150, 179, 0,    /* +船:owned,泊(150,179)地表 */
                         1, 0, 2, 2,                      /* in_town=1,layer=0,sec=2,daynight=2(黑夜)*/
                         2, {{3,10,5,0},{7,20,8,1}} };    /* v8:魯拉去過2城(CTY3@(10,5)、CTY7@(20,8))*/
    uint16_t nm[2] = {15, 16};

    if (dq3_stats_load(&st, NULL, 1, NULL, 0) < 0) { printf("stats FAIL\n"); return 1; }
    dq3_roster_init(&r); dq3_party_init(&p); dq3_inv_init(&inv);
    dq3_roster_create(&r, &st, 1, DQ3_GENDER_MALE, nm, 2);     /* 戰士 AB */
    dq3_roster_create(&r, &st, 4, DQ3_GENDER_FEMALE, nm, 2);   /* 魔法使者 */
    dq3_party_add(&p, &r, 0); dq3_party_add(&p, &r, 1);
    dq3_inv_add(&inv, ITEM_SUN_STONE); dq3_inv_add(&inv, ITEM_KEY_MAGIC);
    dq3_flags_init(&fl); dq3_flags_set(&fl, 0x213, 1); dq3_flags_set(&fl, 0x39, 1);   /* 劇情進度旗標 */

    printf("== 寫存檔 ==\n");
    CHECK(dq3_save_write(path, &r, &p, &inv, &fl, pos) == 0, "寫檔成功");
    CHECK(dq3_save_exists(path), "存檔存在 + magic 正確");
    CHECK(!dq3_save_exists("/tmp/dq3_nonexistent_xyz.dat"), "不存在檔 → exists=0");

    printf("== 讀回 + 比對 ==\n");
    {
        dq3_roster r2; dq3_party p2; dq3_inventory inv2; dq3_save_pos pos2; dq3_storyflags fl2;
        memset(&r2, 0xcc, sizeof r2);   /* 先填髒值,確認真的被覆蓋 */
        CHECK(dq3_save_read(path, &r2, &p2, &inv2, &fl2, &pos2) == 0, "讀檔成功");
        CHECK(r2.count == r.count, "名冊人數一致");
        CHECK(r2.list[0].m.cls == 1 && r2.list[0].gender == DQ3_GENDER_MALE, "成員0:戰士 男 一致");
        CHECK(r2.list[1].m.cls == 4 && r2.list[1].gender == DQ3_GENDER_FEMALE, "成員1:魔法使者 女 一致");
        CHECK(r2.list[0].name[0] == 15 && r2.list[0].name[1] == 16, "名字一致");
        CHECK(r2.list[0].m.stat[0] == r.list[0].m.stat[0], "數值一致");
        CHECK(p2.count == 2 && p2.slot[0] == 0 && p2.slot[1] == 1, "隊伍一致");
        CHECK(dq3_inv_has(&inv2, ITEM_SUN_STONE) && dq3_inv_has(&inv2, ITEM_KEY_MAGIC), "道具一致");
        CHECK(pos2.cty == 5 && pos2.px == 12 && pos2.py == 8, "位置一致");
        CHECK(pos2.ship_owned == 1 && pos2.ship_px == 150 && pos2.ship_py == 179, "船狀態一致");
        CHECK(pos2.in_town == 1 && pos2.layer == 0 && pos2.sec == 2, "場景還原欄一致(in_town/layer/sec)");
        CHECK(pos2.daynight == 2, "v6:晝夜相位一致(黑夜)");
        CHECK(pos2.n_visited == 2 && pos2.visited[0].cty == 3 && pos2.visited[0].x == 10
              && pos2.visited[1].cty == 7 && pos2.visited[1].sec == 1,
              "v8:魯拉去過城鎮一致(存讀檔持久)");
        CHECK(dq3_flags_get(&fl2, 0x213) && dq3_flags_get(&fl2, 0x39), "v6:劇情旗標一致(0x213/0x39 持久)");
        CHECK(memcmp(&r, &r2, sizeof r) == 0, "名冊整塊 byte 一致");
    }

    printf("== 多存檔 slot 隔離(6 slot,使用者需求)==\n");
    {
        /* slot 路徑規則(對齊 main.c slot_path):base="/tmp/dq3_slot.dat" → slotN="/tmp/dq3_slotN.dat"。
         * 寫 6 個 slot 各帶不同 CTY,逐一讀回確認互不覆蓋。 */
        char sp[7][64]; int s, ok_iso = 1, ok_cnt = 0;
        for (s = 1; s <= 6; s++) {
            dq3_save_pos sps = { 0 };
            snprintf(sp[s], sizeof sp[s], "/tmp/dq3_slot%d.dat", s);
            sps.cty = 10 + s;                       /* 每 slot 不同 CTY 當指紋 */
            sps.px = s; sps.py = s * 2;
            if (dq3_save_write(sp[s], &r, &p, &inv, &fl, sps) == 0) ok_cnt++;
        }
        CHECK(ok_cnt == 6, "6 個 slot 全部寫檔成功");
        for (s = 1; s <= 6; s++) {
            dq3_roster rr; dq3_party pp; dq3_inventory ii; dq3_save_pos rp;
            if (dq3_save_read(sp[s], &rr, &pp, &ii, &fl, &rp) != 0 || rp.cty != 10 + s
                || rp.px != s || rp.py != s * 2) ok_iso = 0;
        }
        CHECK(ok_iso, "6 個 slot 各自獨立(讀回 CTY/座標指紋一致,互不覆蓋)");
        CHECK(dq3_save_exists(sp[3]) && dq3_save_exists(sp[6]), "slot 3 / slot 6 存檔皆存在");
        for (s = 1; s <= 6; s++) remove(sp[s]);
    }

    printf("== 防呆:壞檔 / 缺檔 ==\n");
    CHECK(dq3_save_read("/tmp/dq3_nonexistent_xyz.dat", &r, &p, &inv, &fl, &pos) == -1, "缺檔 → -1");
    { FILE *f = fopen(path, "wb"); fwrite("GARBAGE!", 1, 8, f); fclose(f); }
    CHECK(dq3_save_read(path, &r, &p, &inv, &fl, &pos) == -2, "壞 magic → -2");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
