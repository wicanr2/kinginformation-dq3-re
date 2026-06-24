/* dq3_spell.c — 咒文習得查詢實作。 */
#include "dq3_spell.h"

/* 把一系中 level 級前已學的咒(按等級)合併進 out(去重、保持等級序)。 */
static int merge_school(const dq3_spell_learn *s, int n, int level,
                        unsigned short *out, int cnt, int max)
{
    int i, j;
    for (i = 0; i < n; i++) {
        if (s[i].level > level) continue;
        for (j = 0; j < cnt; j++) if (out[j] == s[i].rec) break;   /* 去重 */
        if (j < cnt) continue;
        if (cnt >= max) break;
        out[cnt++] = s[i].rec;
    }
    return cnt;
}

/* 插入排序把 out 依「習得等級」排序需要原始等級;此處改為合併時即按系內等級序插入。
 * 兩系合併(賢者)時用簡單插入排序依等級。為此需查等級:用線性找回。 */
static int level_of(unsigned short rec)
{
    int i;
    for (i = 0; i < dq3_school_hero_n; i++)   if (dq3_school_hero[i].rec == rec)   return dq3_school_hero[i].level;
    for (i = 0; i < dq3_school_priest_n; i++) if (dq3_school_priest[i].rec == rec) return dq3_school_priest[i].level;
    for (i = 0; i < dq3_school_mage_n; i++)    if (dq3_school_mage[i].rec == rec)   return dq3_school_mage[i].level;
    return 99;
}

int dq3_spells_known(int cls, int level, unsigned short *out, int max)
{
    int cnt = 0, i, j;
    if (level < 1) level = 1;
    switch (cls) {
    case 0: cnt = merge_school(dq3_school_hero,   dq3_school_hero_n,   level, out, cnt, max); break; /* 勇者 */
    case 3: cnt = merge_school(dq3_school_priest, dq3_school_priest_n, level, out, cnt, max); break; /* 僧侶 */
    case 4: cnt = merge_school(dq3_school_mage,   dq3_school_mage_n,   level, out, cnt, max); break; /* 魔法使 */
    case 5: /* 賢者:僧侶 + 魔法兩系 */
        cnt = merge_school(dq3_school_priest, dq3_school_priest_n, level, out, cnt, max);
        cnt = merge_school(dq3_school_mage,   dq3_school_mage_n,   level, out, cnt, max);
        break;
    default: return 0;   /* 戰士/武鬥/商人/遊玩:無咒 */
    }
    /* 依習得等級排序(賢者合併後尤需)*/
    for (i = 1; i < cnt; i++) {
        unsigned short k = out[i]; int lk = level_of(k);
        for (j = i - 1; j >= 0 && level_of(out[j]) > lk; j--) out[j + 1] = out[j];
        out[j + 1] = k;
    }
    return cnt;
}
