/* test_combat.c — 戰鬥道具特效單元測試:#7a 隼劍雙擊 / #7b 魔甲抗魔。
 * 對比「原版(從不讀旗標)vs C 修正」。回傳 0=全通過。
 * 用法:dq3_combat_test <assets_dir>(需含 ITEM.DAT)。
 */
#include "dq3_combat.h"
#include <stdio.h>

static int g_fail = 0;
#define CHECK(cond, msg) do { \
    if (cond) printf("  [PASS] %s\n", msg); \
    else { printf("  [FAIL] %s\n", msg); g_fail++; } } while (0)

#define FALCON_SWORD 0x6e
#define MAGIC_ARMOR  0x2b
#define PLAIN_SWORD  0x60   /* 任一非雙擊武器(下方檢查其 +5 bit7 為 0) */

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    char err[256] = {0};
    dq3_items it;

    if (dq3_items_load(&it, assets, err, sizeof err) != 0) {
        fprintf(stderr, "load ITEM.DAT failed: %s\n", err); return 2;
    }

    printf("== Bug #7a:隼劍(飛鷹劍 0x6e)雙擊 ==\n");
    printf("  飛鷹劍 +5=0x%02x  普通劍 0x%02x +5=0x%02x\n",
           dq3_item_flag5(&it, FALCON_SWORD), PLAIN_SWORD, dq3_item_flag5(&it, PLAIN_SWORD));
    CHECK(dq3_item_flag5(&it, FALCON_SWORD) & DQ3_WPN_DOUBLE_HIT,
          "飛鷹劍 +5 bit7 = double-hit 旗標(資料齊全)");
    CHECK(dq3_combat_num_attacks(&it, FALCON_SWORD) == 2, "修正:飛鷹劍打 2 次");
    CHECK(!(dq3_item_flag5(&it, PLAIN_SWORD) & DQ3_WPN_DOUBLE_HIT)
          && dq3_combat_num_attacks(&it, PLAIN_SWORD) == 1, "非雙擊武器打 1 次");
    /* 原版:攻擊碼無攻擊次數變數、不讀 +5,等同永遠 1 次 */
    printf("  (原版從不讀 +5 → 飛鷹劍也只打 1 次,即 bug)\n");

    printf("== Bug #7b:魔法鎧甲(0x2b)抗魔 ==\n");
    {
        uint8_t with_armor[DQ3_EQUIP_SLOTS] = {0xff,0xff,MAGIC_ARMOR,0xff,0xff,0xff,0xff,0xff};
        uint8_t no_armor[DQ3_EQUIP_SLOTS]   = {0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff};
        printf("  魔甲 +6=0x%02x (bit2=%d)\n",
               dq3_item_flag6(&it, MAGIC_ARMOR),
               (dq3_item_flag6(&it, MAGIC_ARMOR) & DQ3_ARM_MAGIC_RESIST) ? 1 : 0);
        CHECK(dq3_item_flag6(&it, MAGIC_ARMOR) & DQ3_ARM_MAGIC_RESIST,
              "魔法鎧甲 +6 bit2 = 抗魔旗標(資料齊全)");
        CHECK(dq3_combat_has_magic_resist(&it, with_armor), "偵測到裝備魔甲 → 抗魔");
        CHECK(!dq3_combat_has_magic_resist(&it, no_armor), "未裝魔甲 → 無抗魔");
        CHECK(dq3_combat_spell_damage(&it, with_armor, 80) == 40, "修正:抗魔咒文傷害減半 80→40");
        CHECK(dq3_combat_spell_damage(&it, no_armor, 80) == 80, "無抗魔咒文傷害不變 80");
        printf("  (原版咒文扣血從不掃裝備 → 魔甲不減傷,即 bug)\n");
    }

    printf("\n%s (%d failures)\n", g_fail ? "== 有測試未通過 ==" : "== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
