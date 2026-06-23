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

int dq3_inv_key_tier(const dq3_inventory *inv)
{
    int i, best = 0;
    for (i = 0; i < DQ3_INV_SLOTS; i++) {            /* 0x48c3:逐格掃 */
        int it = inv->slot[i];
        if (it >= ITEM_KEY_THIEF && it <= ITEM_KEY_FINAL) {  /* dl∈[0x55,0x57] */
            int tier = it - 0x54;                    /* sub dl,0x54 → 1/2/3 */
            if (tier > best) best = tier;            /* 取最大(=最高階鑰匙)*/
        }
    }
    return best;
}

void dq3_flags_init(dq3_storyflags *f)
{
    int i; for (i = 0; i < DQ3_FLAGS_BYTES; i++) f->bit[i] = 0;
}
int dq3_flags_get(const dq3_storyflags *f, int id)
{
    if (id < 0 || id >= DQ3_FLAGS_BYTES * 8) return 0;
    return (f->bit[id >> 3] >> (id & 7)) & 1;
}
void dq3_flags_set(dq3_storyflags *f, int id, int v)
{
    if (id < 0 || id >= DQ3_FLAGS_BYTES * 8) return;
    if (v) f->bit[id >> 3] |=  (uint8_t)(1 << (id & 7));
    else   f->bit[id >> 3] &= (uint8_t)~(1 << (id & 7));
}

int dq3_synth_shrine_examine(dq3_inventory *inv, dq3_storyflags *flags, int fixed)
{
    int r;
    /* 對齊 file 0x77ce 派發前提:旗標 0x139 已設則不再合成(避免重複)。 */
    if (dq3_flags_get(flags, DQ3_FLAG_RAINBOW_SYNTHED)) return -2;
    r = dq3_synth_rainbow_drop(inv, fixed);
    if (r >= 0) dq3_flags_set(flags, DQ3_FLAG_RAINBOW_SYNTHED, 1);  /* 0x77eb: set_flag(0x139) */
    return r;
}

int dq3_scripted_event_run(int event_id, dq3_inventory *inv, dq3_storyflags *flags, int fixed)
{
    /* 對齊 runner(file 0xabb2)的 id→handler 跳表(DS 0x3baa)。目前僅實作合成。 */
    switch (event_id) {
    case DQ3_SEVENT_RAINBOW_SYNTH:                 /* id 83 → handler 0x776c(0x77ce 尾段)*/
        return dq3_synth_shrine_examine(inv, flags, fixed);
    default:
        return -3;                                 /* 其餘 scripted-event 未實作 */
    }
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
