/* dq3_spell.h — 咒文習得查詢:由職業 + 等級算出已學咒文(D3TXT00 名稱 rec)。
 * 資料 dq3_spells_table.c 由 DQ3.EXE 習得表抽出(tools/gen_spells.py)。 */
#ifndef DQ3_SPELL_H
#define DQ3_SPELL_H

typedef struct { unsigned char level; unsigned short rec; } dq3_spell_learn;

/* 三系習得清單(各按習得等級排序)。 */
extern const dq3_spell_learn dq3_school_hero[];   extern const int dq3_school_hero_n;
extern const dq3_spell_learn dq3_school_priest[]; extern const int dq3_school_priest_n;
extern const dq3_spell_learn dq3_school_mage[];    extern const int dq3_school_mage_n;

/* 取職業 cls 在 level 級已學會的咒文名 rec,填入 out(上限 max),回傳數量。
 * 職業→系:勇者→勇者系、僧侶→僧侶系、魔法使→魔法系、賢者→僧侶+魔法;其餘 0 咒。
 * 依習得等級排序,跨系去重。 */
int dq3_spells_known(int cls, int level, unsigned short *out, int max);

/* ---- 咒文施放 descriptor ----
 * 傷害/回復公式來自 DQ3.EXE RE(file 0xc22e):val = base/2 + rng(base/2),魔甲再減半(#7b)。
 * base 威力 / MP 消耗 / 目標 用 DQ3 標準值(BBS 咒文表佐證);詳見 dq3_spelldef.c 註解。 */
typedef enum { DQ3_SK_DMG = 0, DQ3_SK_HEAL = 1, DQ3_SK_REVIVE = 2 } dq3_spell_kind;
typedef enum { DQ3_TG_ENEMY1 = 0, DQ3_TG_GROUP = 1, DQ3_TG_ALL = 2,
               DQ3_TG_ALLY1 = 3, DQ3_TG_ALLYALL = 4 } dq3_spell_target;
typedef struct {
    unsigned short rec;       /* 咒文名 rec(對 dq3_spells_known)*/
    unsigned char  mp;        /* MP 消耗 */
    unsigned short base;      /* base 威力(傷害/回復;走 base/2+rng 公式)*/
    unsigned char  kind;      /* dq3_spell_kind */
    unsigned char  target;    /* dq3_spell_target */
} dq3_spell_def;

/* 查咒文 rec 的施放 descriptor;非可施放(輔助/未實作)回 NULL。 */
const dq3_spell_def *dq3_spell_def_get(unsigned short rec);

/* 怪物咒文 bit(0..47)→ 咒名 rec(EXE 0x3930 remap,docs/37)。 */
extern const unsigned short dq3_monster_spell_rec[48];

#endif /* DQ3_SPELL_H */
