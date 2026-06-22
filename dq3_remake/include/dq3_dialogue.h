/* dq3_dialogue.h — 對話/訊息視窗(D3TXT 記錄分頁渲染)。
 *
 * DQ 對話流程(docs/12):確認鍵 → 前方物件事件(dq3x_event[front_tile])→ 起對話;
 * 文字繪製器逐字模畫(4 行/頁),控制碼 0xfffe 換行、0xfffc 換頁(等玩家按鍵)、
 * 0xffed+ 動態插值(主角名/數值)。本模組做「視窗 + 分頁推進」。
 */
#ifndef DQ3_DIALOGUE_H
#define DQ3_DIALOGUE_H

#include <stdint.h>
#include "dq3_text.h"

#define DQ3_DLG_LINES 4

typedef struct {
    dq3_text txt;        /* 該場景的 D3TXTNN(owned 由 dq3_dialogue_load) */
    uint16_t buf[2048];  /* 目前記錄的 glyph 序列 */
    int len;
    int rec;             /* 目前記錄 */
    int pos;             /* 本頁起始 glyph index */
    int open;            /* 是否顯示中 */
} dq3_dialogue;

/* 載入該場景文字檔(如 "D3TXT01.TXT")。失敗回 <0。 */
int  dq3_dialogue_load(dq3_dialogue *d, const char *assets_dir, const char *txt_name,
                       char *err, int errcap);
void dq3_dialogue_free(dq3_dialogue *d);

/* 開始顯示記錄 rec(從頭)。回傳 0 成功。 */
int  dq3_dialogue_open(dq3_dialogue *d, int rec);
int  dq3_dialogue_is_open(const dq3_dialogue *d);

/* 推進到下一頁;若已到記錄結尾則關閉並回 1(done),否則回 0(還有)。 */
int  dq3_dialogue_advance(dq3_dialogue *d);

/* 把對話視窗 + 本頁文字繪進 framebuffer(底部訊息窗)。 */
void dq3_dialogue_render(dq3_dialogue *d, uint8_t *fb, int fb_w, int fb_h);

/* 設對話視窗用色(白字/暗底 palette 索引;set_palette 後呼叫)。 */
void dq3_dialogue_set_colors(int white_idx, int dark_idx);

#endif /* DQ3_DIALOGUE_H */
