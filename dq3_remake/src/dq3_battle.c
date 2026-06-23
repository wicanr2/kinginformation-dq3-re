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

int dq3_battle_phys_damage(int atk, int def, int roll, int crit)
{
    /* 反組譯真公式(file 0xc03e,docs/13):
     *   atk > def:dmg = (atk−def)/2 + rng(0..(atk−def)/4);會心 → ×2。
     *   atk ≤ def(弱攻):rng;<0x80(50%)→ 1 點,否則 miss(0)。 */
    int diff, base, quarter, var, dmg;
    if (roll < 0) roll = 0; if (roll > 255) roll = 255;
    diff = atk - def;
    if (diff <= 0) return (roll < 0x80) ? 1 : 0;   /* 弱攻:50% 打 1、50% miss */
    base = diff / 2;                                /* (atk−def)/2 */
    quarter = diff / 4;                             /* (atk−def)/4 變異上限 */
    var = (quarter > 0) ? (quarter * roll) / 256 : 0;  /* rng(0..(atk−def)/4) */
    dmg = base + var;
    if (crit) dmg *= 2;                             /* [0x24b7]&0x10 會心 → ×2 */
    return dmg;
}

int dq3_battle_flee_ok(int party_agi, int enemy_flee_resist, int roll)
{
    /* 逃跑成功門檻:agi 越高越易、敵抗性越高越難。roll 0..255。 */
    int chance = 128 + party_agi - enemy_flee_resist * 2;
    if (chance < 16) chance = 16;          /* 至少有低機率 */
    if (chance > 240) chance = 240;
    return roll < chance;
}
