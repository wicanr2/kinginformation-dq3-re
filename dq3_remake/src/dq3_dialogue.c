/* dq3_dialogue.c — 對話/訊息視窗實作。 */
#include "dq3_dialogue.h"
#include "dq3_runtime.h"
#include "dq3_assets.h"
#include <string.h>

/* 訊息窗位置(底部,DQ 風格);內文 16×16 字模,4 行。 */
#define WIN_X 24
#define WIN_Y 244
#define WIN_W (DQ3_SCREEN_W - 48)
#define WIN_H 96
#define PAD   12
#define COLS  ((WIN_W - 2*PAD) / DQ3_GLYPH_PX)

static int g_white = 15, g_dark = 0;   /* palette 索引(由場景 palette 決定;預設高亮/黑) */

int dq3_dialogue_load(dq3_dialogue *d, const char *dir, const char *txt_name, char *err, int errcap)
{
    memset(d, 0, sizeof *d);
    if (dq3_text_load(&d->txt, dir, txt_name, err, errcap) != 0) return -1;
    return 0;
}
void dq3_dialogue_free(dq3_dialogue *d){ if(d){ dq3_text_free(&d->txt); d->open=0; } }

int dq3_dialogue_open(dq3_dialogue *d, int rec)
{
    int n = dq3_text_record(&d->txt, rec, d->buf, 2048);
    if (n < 0) return -1;
    d->len = n; d->rec = rec; d->pos = 0; d->open = 1;
    return 0;
}
int dq3_dialogue_is_open(const dq3_dialogue *d){ return d->open; }

/* 從 pos 掃一頁(最多 DQ3_DLG_LINES 行 / 到 0xfffc 換頁 / 結尾),回傳「下一頁起點」。 */
static int scan_page(const dq3_dialogue *d, int pos, int *out_end_is_eof)
{
    int i = pos, col = 0, line = 0;
    *out_end_is_eof = 0;
    while (i < d->len) {
        uint16_t v = d->buf[i];
        if (v == DQ3_TXT_PAGE) { i++; return i; }           /* 換頁:下一頁從此後 */
        if (v == DQ3_TXT_NL || v == DQ3_TXT_NL2) { col=0; line++; i++; if(line>=DQ3_DLG_LINES) return i; continue; }
        if (v >= 0xffed) { i++; col++; if(col>=COLS){col=0;line++;} if(line>=DQ3_DLG_LINES) return i; continue; }
        i++; col++;
        if (col >= COLS) { col = 0; line++; if (line >= DQ3_DLG_LINES) return i; }
    }
    *out_end_is_eof = 1;
    return i;
}

int dq3_dialogue_advance(dq3_dialogue *d)
{
    int eof; int next;
    if (!d->open) return 1;
    next = scan_page(d, d->pos, &eof);
    if (eof) { d->open = 0; return 1; }   /* 本頁已到結尾 → 關閉 */
    d->pos = next;
    return 0;
}

void dq3_dialogue_render(dq3_dialogue *d, uint8_t *fb, int fb_w, int fb_h)
{
    int i, col=0, line=0, x0=WIN_X+PAD, y0=WIN_Y+PAD, eof;
    int end = scan_page(d, d->pos, &eof);
    if (!d->open) return;
    /* 視窗底 + 白框 */
    { int r,c; for(r=0;r<WIN_H;r++)for(c=0;c<WIN_W;c++){ int yy=WIN_Y+r,xx=WIN_X+c;
        if(yy>=0&&yy<fb_h&&xx>=0&&xx<fb_w) fb[yy*fb_w+xx]=(uint8_t)g_dark; } }
    { int c; for(c=0;c<WIN_W;c++){ fb[WIN_Y*fb_w+WIN_X+c]=(uint8_t)g_white; fb[(WIN_Y+WIN_H-1)*fb_w+WIN_X+c]=(uint8_t)g_white; }
      int r; for(r=0;r<WIN_H;r++){ fb[(WIN_Y+r)*fb_w+WIN_X]=(uint8_t)g_white; fb[(WIN_Y+r)*fb_w+WIN_X+WIN_W-1]=(uint8_t)g_white; } }
    /* 本頁文字 */
    for (i=d->pos; i<end && i<d->len; i++) {
        uint16_t v = d->buf[i];
        if (v==DQ3_TXT_PAGE) break;
        if (v==DQ3_TXT_NL || v==DQ3_TXT_NL2){ col=0; line++; if(line>=DQ3_DLG_LINES)break; continue; }
        if (v>=0xffed){ col++; if(col>=COLS){col=0;line++;} continue; }   /* 變數占位空白 */
        if (v<1476)
            dq3_text_draw_glyph(&d->txt, fb, fb_w, fb_h, x0+col*DQ3_GLYPH_PX, y0+line*DQ3_GLYPH_PX, (int)v, (uint8_t)g_white);
        col++;
        if (col>=COLS){ col=0; line++; if(line>=DQ3_DLG_LINES)break; }
    }
}

/* 由場景設定對話視窗用色(白字/暗底);呼叫端在 set_palette 後調。 */
void dq3_dialogue_set_colors(int white_idx, int dark_idx){ g_white=white_idx; g_dark=dark_idx; }
