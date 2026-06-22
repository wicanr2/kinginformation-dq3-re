/* dq3_battle.h — 戰鬥結算(勝負判定),含 #1 巴拉摩斯打不死修正。
 *
 * #1 根因(docs/18):戰鬥某「累積值」上限夾制引用錯欄(+0x2336 應為 +0x2334),
 * 使巴拉摩斯戰的勝負判定走偏 —— 全隊被「巴西魯拉(空間追放)」吹出戰場(我方實際已
 * 無人)卻判我方贏,正常打死巴拉摩斯反而不結束。
 *
 * C 重寫以「正確結算」根治:先判我方是否全滅(含被吹飛/移出),再判敵方是否全滅。
 * 我方全滅 → 敗;敵方全滅 → 勝;否則續戰。順序確保「全隊被吹飛」必判敗,不會誤勝。
 */
#ifndef DQ3_BATTLE_H
#define DQ3_BATTLE_H

#include <stdint.h>

typedef enum { DQ3_BATTLE_ONGOING = 0, DQ3_BATTLE_VICTORY = 1, DQ3_BATTLE_DEFEAT = 2 } dq3_battle_outcome;

/* 存活數:HP>0 且未被移出(removed[i]==0)者。removed 可為 NULL(視為皆在場)。 */
int dq3_battle_alive(const uint16_t *hp, const uint8_t *removed, int n);

/* #1 修正版結算:我方全滅→敗(優先),敵方全滅→勝,否則續戰。 */
dq3_battle_outcome dq3_battle_resolve(const uint16_t *party_hp, const uint8_t *party_removed, int party_n,
                                      const uint16_t *enemy_hp, int enemy_n);

/* 重現原版 #1 bug:結算只看「敵方累積判定」而漏判我方全滅,全隊被吹飛時誤判勝。 */
dq3_battle_outcome dq3_battle_resolve_buggy(const uint16_t *party_hp, const uint8_t *party_removed, int party_n,
                                            const uint16_t *enemy_hp, int enemy_n);

/* 套用傷害(uint16,夾在 0,不下溢)。 */
uint16_t dq3_battle_apply_damage(uint16_t hp, int dmg);

/* 物理傷害(DQ 風格,roll=0..255 決定變異;crit=會心一擊忽略防禦)。
 * 註:精確 FC 公式待對 DOSBox 校準(docs/13),此為合理近似、確定性可測。
 *   一般:base = atk/2 - def/4;傷害 ∈ [base/2, base]。
 *   弱攻(base<1):傷害 ∈ [0, (atk+4)/6]。
 *   會心:atk/2(無視防禦)。 */
int dq3_battle_phys_damage(int atk, int def, int roll, int crit);

/* 逃跑是否成功:party_agi 對 enemy_flee_resist + roll(0..255)。 */
int dq3_battle_flee_ok(int party_agi, int enemy_flee_resist, int roll);

#endif /* DQ3_BATTLE_H */
