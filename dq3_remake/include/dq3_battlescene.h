/* dq3_battlescene.h — 互動戰鬥場景(遭遇 → 指令 → 回合 → 結算)。
 *
 * 整合 dq3_monster(怪群數值/sprite)、dq3_battle(#1 結算/傷害/逃跑)、runtime
 * (繪圖/輸入/palette)。指令選單 戰/逃/防/道具(對 game3.png / docs/13),
 * 回合:我方行動 → 敵方反擊 → #1-正確結算。HUD:怪群 sprite + 隊伍 HP 條 + 指令游標。
 */
#ifndef DQ3_BATTLESCENE_H
#define DQ3_BATTLESCENE_H

/* 跑一場戰鬥。
 *   monster_id/count:敵群(count 隻同種)。
 *   script:headless 腳本(每字一回合指令:F=戰 R=逃 D=防 I=道具),NULL=互動。
 *   dump:非 NULL 時 headless,繪末幀寫 PPM。
 *   seed:RNG 種子(headless 確定性)。
 * 回傳結算(0 續/1 勝/2 敗;逃成功回 3)。 */
int dq3_battlescene_run(const char *assets, int monster_id, int monster_count,
                        const char *script, const char *dump, unsigned seed);

#endif /* DQ3_BATTLESCENE_H */
