/* dq3_scripted.h — scripted NPC(sub2)data-driven 對話/給物表(Step 2,docs/re-log-722)。
 *
 * 每個 sub2 NPC 的 handler(跳表 0x3bb4)結構同模式(gen_sub2_full 抽出,docs/data/sub2-handlers.md):
 *   test_flag(gate) → 選 rec(di−0xbb8)[+ 給物 [0x2593]]。
 * remake 以「前置道具/已持給物」對映 EXE 旗標,產生同樣 before/give/after 對白 + 給物。
 * byte4 為跳表鍵(全域);多數給予型 NPC byte4 唯一。複雜(取物換船/warp)另特例。
 */
#ifndef DQ3_SCRIPTED_H
#define DQ3_SCRIPTED_H

#define DQ3_SC_NOITEM 0xff

typedef struct {
    unsigned char  byte4;       /* sub2 跳表鍵 */
    unsigned char  cty;         /* 所在 CTY(0xff=不限,僅憑 byte4)*/
    unsigned char  give_item;   /* 給的道具 id(DQ3_SC_NOITEM=不給,純 require/檢查型 NPC)*/
    unsigned char  prereq_item; /* 給之前需持有的道具(DQ3_SC_NOITEM=無前置)*/
    unsigned char  require_item;/* 檢查型 NPC 需持有的道具(0x7c0c 檢查,不消耗;DQ3_SC_NOITEM=非檢查型)。
                                 * 持有→give_rec(成功反應)、未持→before_rec(需該物)。 */
    unsigned char  consume_prereq;/* 1=給物時消耗 prereq_item(transform:消耗 X 換 give_item;
                                 * 如精靈女王消耗夢幻紅寶石 0x59 換覺醒粉 0x5a)。0=保留 prereq(如鑰匙)。 */
    unsigned short milestone;   /* 給/成功時設的進度里程碑(0=無)*/
    unsigned short before_rec;  /* 前置未滿足 / 缺 require 物時的對白 rec */
    unsigned short give_rec;    /* 給予 / require 成功時的對白 rec */
    unsigned short after_rec;   /* 已給過的後話 rec(require 型不用)*/
} dq3_scripted;

/* 依 byte4(+cty)查表;無則回 NULL。 */
const dq3_scripted *dq3_scripted_get(int byte4, int cty);

#endif /* DQ3_SCRIPTED_H */
