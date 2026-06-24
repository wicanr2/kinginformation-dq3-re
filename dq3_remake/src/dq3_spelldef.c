/* dq3_spelldef.c — 咒文施放 descriptor(可施放的攻擊 / 回復咒)。
 *
 * 公式來源:DQ3.EXE RE(docs/13,file 0xc22e)—— 實得值 = base/2 + rng(base/2),
 * 魔法鎧甲(裝備 0x2b)再減半(#7b)。此處只定 base 威力 / MP / 目標。
 * base / MP 數值:採 DQ3 標準值(BBS 咒文表 Jhon Horng-chinq 1994 + 通行 DQ3 數據佐證);
 * EXE 內玩家施法的靜態 descriptor 表尚未定位(docs/13 末「咒文系統 RE」),故威力以標準值代,
 * 待 descriptor 表 RE 完成可回填精確值(公式不變)。
 *
 * rec = 咒文名 D3TXT00 記錄(= spell_id + 0x79),對 dq3_spells_table.c。 */
#include "dq3_spell.h"
#include <stddef.h>

static const dq3_spell_def DEFS[] = {
    /* 攻擊系:火球(單) */
    { 121,  2,  20, DQ3_SK_DMG, DQ3_TG_ENEMY1 },   /* 美拉     */
    { 122,  6,  68, DQ3_SK_DMG, DQ3_TG_ENEMY1 },   /* 美拉米   */
    { 123, 10, 180, DQ3_SK_DMG, DQ3_TG_ENEMY1 },   /* 美拉宙瑪 */
    /* 火焰(一組) */
    { 124,  4,  20, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 吉拉     */
    { 125,  6,  48, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 比吉拉瑪 */
    { 126, 10, 110, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 比吉拉肯 */
    /* 爆炸(全體) */
    { 127,  5,  30, DQ3_SK_DMG, DQ3_TG_ALL },      /* 伊歐     */
    { 128,  8,  60, DQ3_SK_DMG, DQ3_TG_ALL },      /* 伊歐拉   */
    { 129, 15, 150, DQ3_SK_DMG, DQ3_TG_ALL },      /* 伊歐那順 */
    /* 冰雪 */
    { 130,  3,  15, DQ3_SK_DMG, DQ3_TG_ENEMY1 },   /* 希亞多   */
    { 131,  5,  36, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 希亞達可 */
    { 132,  8,  62, DQ3_SK_DMG, DQ3_TG_ALL },      /* 希亞達依恩 */
    { 133, 11, 100, DQ3_SK_DMG, DQ3_TG_ALL },      /* 馬希亞多 */
    /* 真空刃(僧,一組) */
    { 134,  4,  16, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 巴基     */
    { 135,  8,  40, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 巴基瑪   */
    { 136, 12,  90, DQ3_SK_DMG, DQ3_TG_GROUP },    /* 巴基克洛斯 */
    /* 回復(僧/勇) */
    { 161,  3,  50, DQ3_SK_HEAL, DQ3_TG_ALLY1 },   /* 荷依米   ~25-50 */
    { 162,  5, 170, DQ3_SK_HEAL, DQ3_TG_ALLY1 },   /* 比荷依米 ~85-170 */
    { 163,  7, 999, DQ3_SK_HEAL, DQ3_TG_ALLY1 },   /* 伯荷馬   全回復 */
    { 164, 18, 999, DQ3_SK_HEAL, DQ3_TG_ALLYALL }, /* 伯荷馬拉 全員全回復 */
};

const dq3_spell_def *dq3_spell_def_get(unsigned short rec)
{
    int i, n = (int)(sizeof DEFS / sizeof DEFS[0]);
    for (i = 0; i < n; i++) if (DEFS[i].rec == rec) return &DEFS[i];
    return NULL;
}
