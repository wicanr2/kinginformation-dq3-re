/* dq3_inventory.c — 道具欄 + 合成事件實作。 */
#include "dq3_inventory.h"

void dq3_inv_init(dq3_inventory *inv)
{
    int i; for (i = 0; i < DQ3_INV_SLOTS; i++) inv->slot[i] = DQ3_ITEM_NONE;
}

int dq3_inv_find(const dq3_inventory *inv, int item)
{
    int i; for (i = 0; i < DQ3_INV_SLOTS; i++) if (inv->slot[i] == (uint8_t)item) return i;
    return -1;
}
int dq3_inv_has(const dq3_inventory *inv, int item) { return dq3_inv_find(inv, item) >= 0; }

int dq3_inv_add(dq3_inventory *inv, int item)
{
    int i; for (i = 0; i < DQ3_INV_SLOTS; i++) if (inv->slot[i] == DQ3_ITEM_NONE) { inv->slot[i] = (uint8_t)item; return i; }
    return -1;
}
int dq3_inv_remove(dq3_inventory *inv, int item)
{
    int s = dq3_inv_find(inv, item); if (s >= 0) inv->slot[s] = DQ3_ITEM_NONE; return s;
}

int dq3_synth_rainbow_drop(dq3_inventory *inv, int fixed)
{
    int s_sun, s_rod, result;
    /* 對齊 file 0x77ce:找太陽之石 → 消耗(=0xff);找雲雨之杖 → 該格寫入成品。 */
    s_sun = dq3_inv_find(inv, ITEM_SUN_STONE);
    s_rod = dq3_inv_find(inv, ITEM_RAINCLOUD_ROD);
    if (s_sun < 0 || s_rod < 0) return -1;       /* 材料不足 */
    inv->slot[s_sun] = DQ3_ITEM_NONE;            /* 消耗太陽之石 */
    result = fixed ? ITEM_RAINBOW_DROP : ITEM_SILVER_ORB;  /* #2:0x75 修正 / 0x6b bug */
    inv->slot[s_rod] = (uint8_t)result;          /* 在雲雨之杖格寫入成品 */
    return result;
}
