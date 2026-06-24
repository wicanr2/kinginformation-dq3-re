/* test_roster.c — 露依達酒場創角 + 名冊 + 隊伍編成驗證。 */
#include "dq3_roster.h"
#include <stdio.h>

/* stub:roster 測試不繪製(dq3_menu.c 連結需要)*/
void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{ (void)t;(void)fb;(void)fb_w;(void)fb_h;(void)x;(void)y;(void)idx;(void)fg; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

static const char *CLS[] = {"勇者","戰士","武鬥家","僧侶","魔法使者","賢者","商人","遊玩者"};

int main(void)
{
    dq3_stats st;
    dq3_roster r;
    dq3_party p;
    uint16_t nm[3] = {1, 2, 3};

    if (dq3_stats_load(&st, NULL, 1, NULL, 0) < 0) { printf("stats load FAIL\n"); return 1; }
    dq3_roster_init(&r);
    dq3_party_init(&p);

    printf("== 創角:各職業 Lv1 數值 ==\n");
    {
        int c, ok = 1;
        for (c = 1; c < DQ3_NUM_CLASS; c++) {   /* 1..7(0=主角另生);建同伴 */
            int idx = dq3_roster_create(&r, &st, c, c & 1, nm, 3);
            dq3_recruit *rc = &r.list[idx];
            printf("  %s: idx=%d Lv%d HP=%u MP=%u 力=%u\n", CLS[c], idx, rc->m.level,
                   rc->m.stat[DQ3_STAT_HP], rc->m.stat[DQ3_STAT_MP], rc->m.stat[DQ3_STAT_B4]);
            if (idx < 0 || rc->m.cls != c || rc->m.level != 1 || rc->m.stat[DQ3_STAT_HP] == 0) ok = 0;
        }
        CHECK(ok, "7 職業同伴建立成功、Lv1、HP>0");
        CHECK(r.count == 7, "名冊 7 人");
    }

    printf("== 參數防呆 ==\n");
    CHECK(dq3_roster_create(&r, &st, 99, 0, nm, 3) == -1, "非法職業 → -1");
    CHECK(dq3_roster_create(&r, &st, 1, 5, nm, 3) == -1, "非法性別 → -1");

    printf("== 隊伍編成(主角 + 最多 3)==\n");
    {
        CHECK(dq3_party_add(&p, &r, 0) == 0, "加第 1 人 → slot0");
        CHECK(dq3_party_add(&p, &r, 1) == 1, "加第 2 人 → slot1");
        CHECK(dq3_party_add(&p, &r, 2) == 2, "加第 3 人 → slot2");
        CHECK(dq3_party_add(&p, &r, 3) == 3, "加第 4 人 → slot3");
        CHECK(dq3_party_add(&p, &r, 4) == -1, "第 5 人 → 隊伍滿 -1");
        CHECK(dq3_party_add(&p, &r, 0) == -1, "重複加同人 → -1");
        CHECK(r.list[0].in_party && r.list[3].in_party, "in_party 旗標正確");
        CHECK(p.count == 4, "隊伍 4 人");
    }

    printf("== 移出隊伍 / 名冊除名 + 參照修正 ==\n");
    {
        CHECK(dq3_party_remove(&p, &r, 1) == 0, "移出 slot1");
        CHECK(!r.list[1].in_party && p.count == 3, "slot1 釋放、in_party 清");
        CHECK(dq3_party_add(&p, &r, 4) >= 0, "騰出後可再加第 5 人");
        /* 除名 idx0(在隊伍中)→ 應自動移出 + 後面參照往前補 */
        {
            int before = r.count;
            dq3_roster_remove(&r, &p, 0);
            CHECK(r.count == before - 1, "名冊除名 → 人數 -1");
        }
    }

    printf("== 職業+性別 → DQ3MST entry 對映 ==\n");
    CHECK(dq3_class_sprite_entry(0, DQ3_GENDER_MALE) == 0,   "勇者♂ → entry0");
    CHECK(dq3_class_sprite_entry(0, DQ3_GENDER_FEMALE) == 4, "勇者♀ → entry4");
    CHECK(dq3_class_sprite_entry(1, DQ3_GENDER_MALE) == 8,   "戰士♂(class1) → entry8");
    CHECK(dq3_class_sprite_entry(2, DQ3_GENDER_MALE) == 16,  "武鬥家♂(class2) → entry16");
    CHECK(dq3_class_sprite_entry(4, DQ3_GENDER_MALE) == 32,  "魔法使者♂ → entry32");
    CHECK(dq3_class_sprite_entry(7, DQ3_GENDER_FEMALE) == 60,"遊玩者♀ → entry60");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
