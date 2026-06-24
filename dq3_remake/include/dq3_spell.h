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

#endif /* DQ3_SPELL_H */
