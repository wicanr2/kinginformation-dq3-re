/* dq3_spelldef.c — 咒文施放 descriptor(生成檔,tools/gen_spells.py)。
 * MP/base/目標由 DQ3.EXE DS:0x37c3 抽出(3 byte/咒:b0=MP、b1=base、b2=flags);
 * 公式 base/2+rng(base/2)(docs/13 file 0xc22e)、魔甲減半 #7b。
 * spell_id = rec - 0x79。請勿手改,改 tools/gen_spells.py。 */
#include "dq3_spell.h"
#include <stddef.h>

static const dq3_spell_def DEFS[] = {
    { 121,  2,  10, DQ3_SK_DMG, DQ3_TG_ENEMY1 },  /* 美拉 */
    { 122,  6,  80, DQ3_SK_DMG, DQ3_TG_ENEMY1 },  /* 美拉米 */
    { 123, 12, 180, DQ3_SK_DMG, DQ3_TG_ENEMY1 },  /* 美拉宙瑪 */
    { 124,  4,  20, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 吉拉 */
    { 125,  6,  35, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 比吉拉瑪 */
    { 126, 12, 100, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 比吉拉肯 */
    { 127,  5,  20, DQ3_SK_DMG, DQ3_TG_ALL },  /* 伊歐 */
    { 128,  9,  60, DQ3_SK_DMG, DQ3_TG_ALL },  /* 伊歐拉 */
    { 129, 18, 140, DQ3_SK_DMG, DQ3_TG_ALL },  /* 伊歐那順 */
    { 130,  3,  30, DQ3_SK_DMG, DQ3_TG_ENEMY1 },  /* 希亞多 */
    { 131,  6,  50, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 希亞達可 */
    { 132,  9,  70, DQ3_SK_DMG, DQ3_TG_ALL },  /* 希亞達依恩 */
    { 133, 12, 100, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 瑪希亞多 */
    { 134,  4,  15, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 巴基 */
    { 135,  6,  40, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 巴基瑪 */
    { 136,  9, 100, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 巴基克洛斯 */
    { 137,  8,  80, DQ3_SK_DMG, DQ3_TG_ENEMY1 },  /* 萊丁 */
    { 138, 30, 200, DQ3_SK_DMG, DQ3_TG_ALL },  /* 基加丁 */
    { 139,  7, 255, DQ3_SK_DMG, DQ3_TG_ENEMY1 },  /* 沙基 */
    { 140,  7, 255, DQ3_SK_DMG, DQ3_TG_GROUP },  /* 沙拉基 */
    { 141,  1, 255, DQ3_SK_DMG, DQ3_TG_ALL },  /* 梅干得 */
    { 142, 24, 255, DQ3_SK_DMG, DQ3_TG_ALL },  /* 多拉哥拉姆 */
    { 161,  3,  30, DQ3_SK_HEAL, DQ3_TG_ALLY1 },  /* 荷依米 */
    { 162,  5,  85, DQ3_SK_HEAL, DQ3_TG_ALLY1 },  /* 比荷依米 */
    { 163,  7, 255, DQ3_SK_HEAL, DQ3_TG_ALLY1 },  /* 比荷瑪 */
    { 164, 18,  50, DQ3_SK_HEAL, DQ3_TG_ALLYALL },  /* 比荷瑪拉 */
    { 165, 62, 255, DQ3_SK_HEAL, DQ3_TG_ALLYALL },  /* 比荷瑪順 */
};

const dq3_spell_def *dq3_spell_def_get(unsigned short rec)
{
    int i, n = (int)(sizeof DEFS / sizeof DEFS[0]);
    for (i = 0; i < n; i++) if (DEFS[i].rec == rec) return &DEFS[i];
    return NULL;
}

/* 怪物咒文 bit(0..47)→ 咒名 rec(EXE 0x3930 remap,docs/37)。 */
const unsigned short dq3_monster_spell_rec[48] = {
    121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
    133, 134, 135, 136, 139, 140, 141, 144, 145, 146, 147, 148,
    149, 150, 152, 154, 155, 156, 157, 158, 161, 162, 163, 164,
    169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180,
};
