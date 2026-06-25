/* dq3_item_use.c — 消耗品使用效果(#3)。id 對應 docs/49。 */
#include "dq3_item_use.h"

int dq3_item_use_kind(int item_id)
{
    switch (item_id) {
        case DQ3_ITEM_HERB:         return DQ3_USE_HEAL_HP;
        case DQ3_ITEM_ANTIDOTE:     return DQ3_USE_CURE_POISON;
        case DQ3_ITEM_CHIMERA_WING: return DQ3_USE_RETURN_TOWN;
        case DQ3_ITEM_HOLY_WATER:   return DQ3_USE_REPEL;
        case DQ3_ITEM_FULLMOON:     return DQ3_USE_CURE_PARALYSIS;
        default:                    return DQ3_USE_NONE;
    }
}

int dq3_item_use_heal(dq3_member *m, int item_id)
{
    int maxhp, healed;
    if (dq3_item_use_kind(item_id) != DQ3_USE_HEAL_HP) return -1;
    if (!m || m->cur_hp == 0) return 0;                 /* 陣亡不可用藥草(需教會復活)*/
    maxhp = m->stat[DQ3_STAT_HP];
    if (m->cur_hp >= maxhp) return 0;                   /* 已滿 */
    healed = DQ3_HERB_HEAL;
    if (m->cur_hp + healed > maxhp) healed = maxhp - m->cur_hp;
    m->cur_hp = (uint16_t)(m->cur_hp + healed);
    return healed;
}
