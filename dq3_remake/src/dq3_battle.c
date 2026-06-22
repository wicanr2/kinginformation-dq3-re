/* dq3_battle.c — 戰鬥結算實作(#1 修正)。 */
#include "dq3_battle.h"

int dq3_battle_alive(const uint16_t *hp, const uint8_t *removed, int n)
{
    int i, c = 0;
    for (i = 0; i < n; i++)
        if (hp[i] > 0 && (!removed || !removed[i])) c++;
    return c;
}

dq3_battle_outcome dq3_battle_resolve(const uint16_t *party_hp, const uint8_t *party_removed, int party_n,
                                      const uint16_t *enemy_hp, int enemy_n)
{
    /* 先判我方全滅(含被吹飛/移出)→ 敗;杜絕「全隊被吹飛卻誤勝」(#1)。 */
    if (dq3_battle_alive(party_hp, party_removed, party_n) == 0)
        return DQ3_BATTLE_DEFEAT;
    if (dq3_battle_alive(enemy_hp, 0, enemy_n) == 0)
        return DQ3_BATTLE_VICTORY;
    return DQ3_BATTLE_ONGOING;
}

dq3_battle_outcome dq3_battle_resolve_buggy(const uint16_t *party_hp, const uint8_t *party_removed, int party_n,
                                            const uint16_t *enemy_hp, int enemy_n)
{
    /* 原版缺陷:漏判我方全滅,只要敵方累積判定成立就判勝 —— 全隊被吹飛時敵方仍在,
     * 但被吹飛使「在場我方=0」觸發了錯欄累積的結束分支,結算成我方勝。 */
    (void)party_hp;
    if (dq3_battle_alive(party_hp, party_removed, party_n) == 0)
        return DQ3_BATTLE_VICTORY;   /* ★ bug:全隊被吹飛 → 誤判勝 */
    if (dq3_battle_alive(enemy_hp, 0, enemy_n) == 0)
        return DQ3_BATTLE_VICTORY;
    return DQ3_BATTLE_ONGOING;
}

uint16_t dq3_battle_apply_damage(uint16_t hp, int dmg)
{
    int v = (int)hp - dmg;
    if (v < 0) v = 0;
    return (uint16_t)v;
}
