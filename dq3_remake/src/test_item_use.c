/* test_item_use.c — 消耗品效果分類 + 藥草治療封頂/陣亡保護(#3,docs/49)。 */
#include "dq3_item_use.h"
#include <stdio.h>

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(void)
{
    dq3_member m;

    printf("== 效果分類(id 商店交叉驗證)==\n");
    CHECK(dq3_item_use_kind(DQ3_ITEM_HERB) == DQ3_USE_HEAL_HP, "藥草 0x41 → HEAL_HP");
    CHECK(dq3_item_use_kind(DQ3_ITEM_ANTIDOTE) == DQ3_USE_CURE_POISON, "驅毒草 0x42 → CURE_POISON");
    CHECK(dq3_item_use_kind(DQ3_ITEM_CHIMERA_WING) == DQ3_USE_RETURN_TOWN, "蓋美拉翅膀 0x43 → RETURN_TOWN");
    CHECK(dq3_item_use_kind(DQ3_ITEM_HOLY_WATER) == DQ3_USE_REPEL, "聖水 0x44 → REPEL");
    CHECK(dq3_item_use_kind(DQ3_ITEM_FULLMOON) == DQ3_USE_CURE_PARALYSIS, "滿月草 0x45 → CURE_PARALYSIS");
    CHECK(dq3_item_use_kind(0x6e) == DQ3_USE_NONE, "飛鷹劍(裝備)→ NONE");
    CHECK(dq3_item_use_kind(0x55) == DQ3_USE_NONE, "盜賊鑰匙(劇情)→ NONE");

    printf("== 藥草治療 ==\n");
    m.stat[DQ3_STAT_HP] = 50; m.cur_hp = 10;
    CHECK(dq3_item_use_heal(&m, DQ3_ITEM_HERB) == 30 && m.cur_hp == 40, "10→40(回 30)");

    m.cur_hp = 40;   /* 40 + 30 = 70 > 50 → 封頂回 10 */
    CHECK(dq3_item_use_heal(&m, DQ3_ITEM_HERB) == 10 && m.cur_hp == 50, "封頂:40→50(只回 10)");

    m.cur_hp = 50;   /* 已滿 */
    CHECK(dq3_item_use_heal(&m, DQ3_ITEM_HERB) == 0 && m.cur_hp == 50, "已滿 → 回 0");

    m.cur_hp = 0;    /* 陣亡 */
    CHECK(dq3_item_use_heal(&m, DQ3_ITEM_HERB) == 0 && m.cur_hp == 0, "陣亡 → 藥草無效(回 0)");

    m.cur_hp = 10;
    CHECK(dq3_item_use_heal(&m, DQ3_ITEM_HOLY_WATER) == -1, "對非治療道具 → -1");

    printf("== 解狀態(驅毒草/滿月草)==\n");
    m.status = DQ3_STATUS_POISON;
    CHECK(dq3_item_use_cure(&m, DQ3_ITEM_ANTIDOTE) == 1 && m.status == 0, "驅毒草解中毒");
    CHECK(dq3_item_use_cure(&m, DQ3_ITEM_ANTIDOTE) == 0, "無中毒再用驅毒草 → 0(不消耗)");
    m.status = DQ3_STATUS_POISON | DQ3_STATUS_PARALYSIS;
    CHECK(dq3_item_use_cure(&m, DQ3_ITEM_FULLMOON) == 1 && m.status == DQ3_STATUS_POISON, "滿月草只解麻痺(留中毒)");
    CHECK(dq3_item_use_cure(&m, DQ3_ITEM_HERB) == 0, "藥草非解狀態 → 0");

    printf("== 祈禱之戒(#7c:回 MP + ~25%% 損壞)==\n");
    CHECK(dq3_item_use_kind(DQ3_ITEM_PRAYER_RING) == DQ3_USE_PRAYER_RING, "祈禱之戒 0x48 → PRAYER_RING");
    m.stat[DQ3_STAT_MP] = 50; m.cur_mp = 10;
    CHECK(dq3_item_use_prayer_mp(&m) == 30 && m.cur_mp == 40, "MP 10→40(回 30)");
    m.cur_mp = 40;   /* 40+30=70>50 → 封頂回 10 */
    CHECK(dq3_item_use_prayer_mp(&m) == 10 && m.cur_mp == 50, "封頂:40→50(只回 10)");
    m.cur_mp = 50;   /* 已滿 */
    CHECK(dq3_item_use_prayer_mp(&m) == 0 && m.cur_mp == 50, "MP 滿 → 回 0");
    /* 損壞門檻 = RNG(256)≤0x40,機率 = 0x41/0x100 ≈ 25.4%(忠實 RE,不需修正)*/
    CHECK(DQ3_PRAYER_BREAK_LE == 0x40, "損壞門檻 0x40(RE file 0x5af9 cmp al,0x40)");

    printf("\n%s (fail=%d)\n", g_fail ? "== FAIL ==" : "== ALL PASS ==", g_fail);
    return g_fail ? 1 : 0;
}
