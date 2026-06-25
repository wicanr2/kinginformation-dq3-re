/* dq3_locwarp.h — location scripted warp 資料(8 個 call 0xd1f9,docs/44 §5b)。
 *
 * 來源:DQ3.EXE 固定 5-byte struct {b0, cur_map(源CTY), dest, X, Y} + 反組譯 gate。
 * 語意:源迷宮/洞窟(cur_map)→ 城(dest)的 (X,Y) 落點,由 gate 放行。
 * gate_type:0=未明(可能 runner 條件)、1=test_flag(flag bit)、2=runner [0x722]==event。
 *
 * ★ 狀態:**資料已全解,但未 wiring** —— 這 8 個是 runner-driven scripted 事件,觸發在
 *   runner([0x722] 資料驅動,無靜態 setter,docs/47 A1)→ 純靜態追不到「何時/由哪格觸發」。
 *   待 runner 觸發系統解出後即可接(本表提供 dest/落點/gate 現成)。產生器:tools/gen_locwarp.py。
 */
#ifndef DQ3_LOCWARP_H
#define DQ3_LOCWARP_H

typedef struct {
    unsigned char src_cty;    /* 源 CTY(struct cur_map)*/
    unsigned char dest_cty;   /* 目的 CTY */
    unsigned char x, y;       /* 目的落點 tile */
    unsigned char gate_type;  /* 0=未明 1=flag 2=runner[0x722] */
    unsigned char gate_val;   /* flag id 或 runner event id */
} dq3_locwarp;

extern const dq3_locwarp dq3_locwarps[8];

#endif /* DQ3_LOCWARP_H */
