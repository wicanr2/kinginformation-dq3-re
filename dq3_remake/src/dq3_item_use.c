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
        case DQ3_ITEM_PRAYER_RING:  return DQ3_USE_PRAYER_RING;  /* 祈禱之戒:回 MP + ~25% 損壞(#7c)*/
        case DQ3_ITEM_DARK_LAMP:    return DQ3_USE_DARK_LAMP;    /* 黑暗之燈:變夜(晝夜系統)*/
        case 0x5a:                  return DQ3_USE_AWAKEN;   /* 覺醒粉:諾阿尼魯解催眠(杜勝利 Ch11)*/
        case 0x0f:                  return DQ3_USE_GAIA;     /* 蓋亞之劍:阿莎拉慕火山開通(杜勝利 Ch40)*/
        case 0x5e:                  return DQ3_USE_DRAIN;    /* 乾渴壺:四島礁吸海顯現祠堂(杜勝利 Ch27)*/
        case 0x77:                  return DQ3_USE_FAIRYFLUTE; /* 妖精之笛:魯比斯之塔解詛咒得精靈的守護(Ch53)*/
        case 0x75:                  return DQ3_USE_RAINBOW;  /* 彩虹水滴:下層利姆達爾西北架彩虹橋通終盤(Ch55)*/
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

int dq3_item_use_prayer_mp(dq3_member *m)
{
    int maxmp, restored;
    if (!m) return 0;
    maxmp = m->stat[DQ3_STAT_MP];
    if (m->cur_mp >= maxmp) return 0;                   /* MP 已滿 → 回 0(呼叫端可改選他人)*/
    restored = DQ3_PRAYER_MP;
    if (m->cur_mp + restored > maxmp) restored = maxmp - m->cur_mp;
    m->cur_mp = (uint16_t)(m->cur_mp + restored);
    return restored;
}

int dq3_item_use_cure(dq3_member *m, int item_id)
{
    int kind = dq3_item_use_kind(item_id), bit = 0;
    if (!m) return 0;
    if (kind == DQ3_USE_CURE_POISON)         bit = DQ3_STATUS_POISON;
    else if (kind == DQ3_USE_CURE_PARALYSIS) bit = DQ3_STATUS_PARALYSIS;
    else return 0;
    if (!(m->status & bit)) return 0;        /* 無對應異常 → 不消耗 */
    m->status &= (uint8_t)~bit;
    return 1;
}
