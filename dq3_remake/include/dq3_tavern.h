/* dq3_tavern.h — 露依達酒場創角完整流程狀態機(職業→姓名→性別→登錄→名冊)。
 *
 * 串接 dq3_menu(職業/性別選單)、dq3_nameinput(英數姓名)、dq3_roster(名冊/隊伍)。
 * 對應 RE:角色建立 modal sub_0854(docs/15;姓名後接「▶男性/女性」性別選單)。
 * 性別 glyph:男性=[775,674]、女性=[234,674](rec556;234 為「女」OCR 誤標「文」)。
 */
#ifndef DQ3_TAVERN_H
#define DQ3_TAVERN_H

#include <stdint.h>
#include "dq3_roster.h"
#include "dq3_menu.h"
#include "dq3_nameinput.h"
#include "dq3_zhuyin.h"

typedef enum {
    DQ3_TAV_CLASS = 0,   /* 選職業 */
    DQ3_TAV_NAME,        /* 輸入姓名(英數)*/
    DQ3_TAV_GENDER,      /* 選性別 */
    DQ3_TAV_ROSTER       /* 名冊 / 隊伍編成 */
} dq3_tav_state;

typedef struct {
    int state;
    dq3_menu      class_menu;
    dq3_menu      gender_menu;     /* 男性 / 女性 */
    dq3_nameinput ni;
    dq3_zhuyin    zh;              /* 注音輸入(name_mode=1 時用)*/
    int           name_mode;       /* 0=英數 / 1=注音(Tab 0x0f 切換)*/
    int           roster_cursor;   /* 名冊畫面選中列 */
    int           pend_class, pend_gender;
    int           last_created;    /* 最近登錄的名冊 index(-1=無)*/

    dq3_roster      *roster;       /* 參照 */
    dq3_party       *party;
    const dq3_stats *st;
} dq3_tavern;

void dq3_tavern_init(dq3_tavern *tv, dq3_roster *roster, dq3_party *party, const dq3_stats *st);

/* 處理一個 scancode,推進狀態機。回傳當前狀態(dq3_tav_state)。
 * Space(0x39)在名冊畫面 = 新增另一名(回職業選);Enter 在名冊 = 切換該員入隊/離隊。 */
int  dq3_tavern_input(dq3_tavern *tv, uint8_t scancode);

/* 繪製當前畫面(依 state 分派)。 */
void dq3_tavern_render(const dq3_tavern *tv, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                       int x, int y, uint8_t fg, uint8_t curfg);

#endif /* DQ3_TAVERN_H */
