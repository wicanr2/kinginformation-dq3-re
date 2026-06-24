/* dq3_encounter.h — 遭遇區查詢(docs/39)。位置→region→候選怪 + 戰鬥背景頁。 */
#ifndef DQ3_ENCOUNTER_H
#define DQ3_ENCOUNTER_H
int dq3_encounter_region(int x, int y);              /* 地表座標→region */
int dq3_encounter_pick(int region, unsigned rng);    /* region→候選怪 id(無回 -1)*/
int dq3_encounter_bgpage(int region, unsigned rng);  /* region→戰鬥背景頁 */
#endif
