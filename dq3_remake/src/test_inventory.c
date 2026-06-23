/* test_inventory.c — #2 彩虹水滴合成事件驗證(原版 bug vs 修正)。 */
#include "dq3_inventory.h"
#include <stdio.h>

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(void)
{
    dq3_inventory inv;

    printf("== Bug #2:彩虹水滴合成(file 0x77ce)==\n");

    /* 修正版:太陽之石 + 雲雨之杖 → 彩虹水滴(0x75) */
    dq3_inv_init(&inv);
    dq3_inv_add(&inv, ITEM_SUN_STONE);
    dq3_inv_add(&inv, ITEM_RAINCLOUD_ROD);
    {
        int r = dq3_synth_rainbow_drop(&inv, 1);
        printf("  修正版合成 → item 0x%02x\n", r);
        CHECK(r == ITEM_RAINBOW_DROP, "修正:合成品 = 彩虹水滴 0x75");
        CHECK(dq3_inv_has(&inv, ITEM_RAINBOW_DROP), "道具欄取得彩虹水滴");
        CHECK(!dq3_inv_has(&inv, ITEM_SUN_STONE) && !dq3_inv_has(&inv, ITEM_RAINCLOUD_ROD),
              "太陽之石 + 雲雨之杖 已消耗");
    }

    /* 原版 bug:同樣材料卻產出 0x6b(銀寶珠)→ 架不了彩虹橋卡關 */
    dq3_inv_init(&inv);
    dq3_inv_add(&inv, ITEM_SUN_STONE);
    dq3_inv_add(&inv, ITEM_RAINCLOUD_ROD);
    {
        int r = dq3_synth_rainbow_drop(&inv, 0);
        printf("  原版bug合成 → item 0x%02x\n", r);
        CHECK(r == ITEM_SILVER_ORB, "原版:誤產銀寶珠 0x6b(重現 #2 卡關)");
        CHECK(!dq3_inv_has(&inv, ITEM_RAINBOW_DROP), "原版拿不到彩虹水滴");
    }

    /* 材料不足 → 不合成 */
    dq3_inv_init(&inv);
    dq3_inv_add(&inv, ITEM_SUN_STONE);
    CHECK(dq3_synth_rainbow_drop(&inv, 1) == -1, "缺雲雨之杖 → 不合成");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail?1:0;
}
