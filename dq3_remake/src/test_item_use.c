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

    printf("\n%s (fail=%d)\n", g_fail ? "== FAIL ==" : "== ALL PASS ==", g_fail);
    return g_fail ? 1 : 0;
}
