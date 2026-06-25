/* dq3_progress.h — remake 主線進度旗標流(#1)。
 *
 * 「事件 → set_flag → 解鎖下一關」的骨架。主線里程碑順序取自杜勝利攻略
 * (docs/history/dq3-walkthrough-credits.md)。原版用 runner/[0x722](靜態追不到、
 * 早期 build 可能未接全,docs/47),故 remake 自定一張里程碑旗標表串主線,不復刻 runner VM。
 *
 * 每個里程碑 = 一個劇情旗標(對 dq3_storyflags bit 陣列)。完成事件設旗標,
 * dq3_progress_stage 由已設旗標算出目前進度;gate(門/橋/船)查對應旗標放行。
 */
#ifndef DQ3_PROGRESS_H
#define DQ3_PROGRESS_H
#include "dq3_inventory.h"   /* dq3_storyflags */

/* 主線里程碑(順序 = 杜勝利攻略進度)。旗標 id 用 0x200+(避開 EXE 既有旗標 < 0x200)。 */
enum {
    DQ3_MS_START        = 0x200,  /* 阿里阿罕開場(見國王、酒場建隊)*/
    DQ3_MS_THIEF_KEY    = 0x201,  /* 取盜賊鑰匙(拿吉米之塔)→ 開 tier1 鎖門 */
    DQ3_MS_MAGIC_BALL   = 0x202,  /* 雷貝魔法球 → 炸誘人洞窟牆 */
    DQ3_MS_ROMALY       = 0x203,  /* 羅馬利亞皇冠線 */
    DQ3_MS_DHAMA        = 0x204,  /* 達瑪神殿轉職開放 */
    DQ3_MS_SHIP         = 0x205,  /* 波魯多加胡椒換船 → 可跨海 */
    DQ3_MS_RAINBOW      = 0x206,  /* 彩虹水滴合成(對 DQ3_FLAG_RAINBOW_SYNTHED)→ 架彩虹橋 */
    DQ3_MS_DESCEND      = 0x207,  /* 下降アレフガルド(對 DQ3_FLAG_DESCENDED)*/
    DQ3_MS_ZOMA         = 0x208,  /* 大魔王索瑪(終盤)*/
    DQ3_MS_COUNT        = 9
};

/* 目前進度階段(0..DQ3_MS_COUNT):連續已完成的里程碑數(第一個未完成處停)。 */
int  dq3_progress_stage(const dq3_storyflags *f);
/* 階段名(0..DQ3_MS_COUNT)。 */
const char *dq3_progress_name(int stage);
/* 標記某里程碑完成(set 對應旗標;部分鏡射既有 EXE 旗標)。 */
void dq3_progress_set(dq3_storyflags *f, int milestone_flag);
/* 某里程碑是否完成。 */
int  dq3_progress_done(const dq3_storyflags *f, int milestone_flag);

#endif
