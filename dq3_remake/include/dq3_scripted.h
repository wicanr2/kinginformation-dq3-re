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
    unsigned char  give_item;   /* 給的道具 id */
    unsigned char  prereq_item; /* 給之前需持有的道具(DQ3_SC_NOITEM=無前置)*/
    unsigned short milestone;   /* 給時設的進度里程碑(0=無)*/
    unsigned short before_rec;  /* 前置未滿足時的對白 rec */
    unsigned short give_rec;    /* 給予時的對白 rec */
    unsigned short after_rec;   /* 已給過的後話 rec */
} dq3_scripted;

/* 依 byte4(+cty)查表;無則回 NULL。 */
const dq3_scripted *dq3_scripted_get(int byte4, int cty);

#endif /* DQ3_SCRIPTED_H */
