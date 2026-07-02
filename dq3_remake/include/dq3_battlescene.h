/* dq3_battlescene.h — 互動戰鬥場景(遭遇 → 指令 → 回合 → 結算)。
 *
 * 整合 dq3_monster(怪群數值/sprite)、dq3_battle(#1 結算/傷害/逃跑)、runtime
 * (繪圖/輸入/palette)。指令選單 戰/逃/防/道具(對 game3.png / docs/13),
 * 回合:我方行動 → 敵方反擊 → #1-正確結算。HUD:怪群 sprite + 隊伍 HP 條 + 指令游標。
 */
#ifndef DQ3_BATTLESCENE_H
#define DQ3_BATTLESCENE_H

#include "dq3_roster.h"

/* 指定玩家隊伍(露依達酒場建立的名冊 + 編成):下一場起戰鬥改用此隊(姓名/職業/等級/數值)。
 * roster 非 const:勝利後把升級(經驗/等級/數值)回寫名冊,成長持久。傳 NULL 還原為內建範例隊。
 * 閉合「酒場創角 → つよさ → 上場戰鬥 → 升級成長」迴圈。 */
void dq3_battlescene_set_party(dq3_roster *roster, const dq3_party *party);

/* 光之珠(0x65)持有旗標:索瑪戰(怪 0x7c)開始時若為 1,自動使用光之珠 → 弱化索瑪(二階段)。
 * main.c 開索瑪戰前依背包設定。非索瑪戰無作用。用畢自動清 0。 */
void dq3_battlescene_set_light_orb(int has);

/* ★ remake 增強:戰鬥顯示敵人 HP + 下一輪預計動作(可開關;原版無此資訊)。 */
void dq3_battlescene_set_show_info(int on);
void dq3_battlescene_set_hurt_fx(int on);   /* 受傷視覺特效(震動+黃綠閃)開關 */

/* 跑一場戰鬥。
 *   monster_id/count:敵群(count 隻同種)。
 *   bg_page:戰鬥背景索引(隨地形;<0 = 讀 DQ3_BG_PAGE 環境變數,預設 0 草原)。
 *   script:headless 腳本(每字一回合指令:F=戰 R=逃 D=防 I=道具),NULL=互動。
 *   dump:非 NULL 時 headless,繪末幀寫 PPM。
 *   seed:RNG 種子(headless 確定性)。
 * 回傳結算(0 續/1 勝/2 敗;逃成功回 3)。 */
int dq3_battlescene_last_gold(void);  /* 上一場戰利金錢 */
int dq3_battlescene_run(const char *assets, int monster_id, int monster_count,
                        int bg_page, const char *script, const char *dump, unsigned seed);

#endif /* DQ3_BATTLESCENE_H */
