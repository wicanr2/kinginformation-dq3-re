/* dq3_locwarp.h — location scripted warp 資料(8 個 call 0xd1f9,docs/44 §5b)。
 *
 * 來源:DQ3.EXE 固定 5-byte struct {b0, cur_map(源CTY), dest, X, Y} + 反組譯 gate。
 * 語意:源迷宮/洞窟(cur_map)→ 城(dest)的 (X,Y) 落點,由 gate 放行。
 * gate_type:0=未明(可能 runner 條件)、1=test_flag(flag bit)、2=runner [0x722]==event。
 *
 * ★ 狀態(2026-06-27 更正):這 8 條洞窟/迷宮→城的連接在遊戲內**可走** —— 由 remake 既有的
 *   type-2 examine warp(0x4ea0,旅の扉/洞穴出口)+ section +0xc 轉場表(門/階梯/出城)覆蓋。
 *   本表是同連接的另一條 EXE codepath(0xd1f9)的冗餘資料,**不需單獨 wiring**(故 main.c 無消費端)。
 *   保留供對照/未來若要走 0xd1f9 精確 gate 用。產生器:tools/gen_locwarp.py。
 *   (舊註「runner 純靜態追不到」已過時:[0x722] runner 已靜態攻克,見 docs/re-log-722。)
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
