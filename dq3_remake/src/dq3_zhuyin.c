/* dq3_zhuyin.c — 注音輸入法(組字 + 候選挑字)。 */
#include "dq3_zhuyin.h"

int dq3_zhuyin_lookup(int sh, int ji, int yu, int tone, uint16_t *out, int max)
{
    int key, lo = 0, hi = dq3_zh_nbucket - 1, n = 0, i;
    if (tone < 0) tone = 0;
    key = ((sh * 4 + ji) * 14 + yu) * 5 + tone;
    while (lo <= hi) {                       /* 二分查找 packed key */
        int mid = (lo + hi) / 2;
        if (dq3_zh_key[mid] == key) {
            int a = dq3_zh_off[mid], b = dq3_zh_off[mid + 1];
            for (i = a; i < b && n < max; i++) out[n++] = dq3_zh_pool[i];
            return n;
        }
        if (dq3_zh_key[mid] < key) lo = mid + 1; else hi = mid - 1;
    }
    return 0;
}

/* 注音盤格 → 注音符號 glyph(0xffff=無)。cell 41 = 一聲(無符號,畫橫線)。 */
static uint16_t cell_glyph(int cell)
{
    if (cell < 21) return (uint16_t)(65 + cell);          /* 聲母 */
    if (cell < 34) return (uint16_t)(86 + (cell - 21));   /* 韻母 */
    if (cell < 37) return (uint16_t)(99 + (cell - 34));   /* 介音 */
    switch (cell) {                                       /* 聲調 */
        case 37: return 103; case 38: return 104; case 39: return 105; case 40: return 102;
        default: return 0xffff;                           /* 41 = 一聲 */
    }
}

void dq3_zhuyin_init(dq3_zhuyin *z)
{
    z->sh = 0; z->ji = 0; z->yu = 0; z->tone = -1;
    z->cursor = 0; z->phase = DQ3_ZH_COMPOSE;
    z->ncand = 0; z->cand_cur = 0;
}

/* 選定一格 → 填組字 / 觸發查表。回 1=進入候選(PICK)。 */
static int compose_select(dq3_zhuyin *z, int cell)
{
    int tone;
    if (cell < 21) { z->sh = cell + 1; return 0; }            /* 聲母 */
    if (cell < 34) { z->yu = (cell - 21) + 1; return 0; }     /* 韻母 */
    if (cell < 37) { z->ji = (cell - 34) + 1; return 0; }     /* 介音 */
    /* 聲調(含一聲)→ 觸發查表 */
    switch (cell) { case 37: tone=1; break; case 38: tone=2; break;
                    case 39: tone=3; break; case 40: tone=4; break; default: tone=0; }
    z->tone = tone;
    z->ncand = dq3_zhuyin_lookup(z->sh, z->ji, z->yu, tone, z->cand, DQ3_ZH_MAXCAND);
    if (z->ncand > 0) { z->phase = DQ3_ZH_PICK; z->cand_cur = 0; return 1; }
    /* 查無此音 → 清聲調,維持組字讓使用者重選 */
    z->tone = -1;
    return 0;
}

int dq3_zhuyin_input(dq3_zhuyin *z, uint8_t sc)
{
    if (z->phase == DQ3_ZH_PICK) {
        switch (sc) {
            case 0x4b: if (z->cand_cur > 0) z->cand_cur--; else z->cand_cur = z->ncand - 1; return -1; /* ← */
            case 0x4d: z->cand_cur = (z->cand_cur + 1) % z->ncand; return -1;                          /* → */
            case 0x48: if (z->cand_cur >= DQ3_ZH_COLS) z->cand_cur -= DQ3_ZH_COLS; return -1;          /* ↑ */
            case 0x50: if (z->cand_cur + DQ3_ZH_COLS < z->ncand) z->cand_cur += DQ3_ZH_COLS; return -1;/* ↓ */
            case 0x01: dq3_zhuyin_init(z); return -1;          /* ESC 放棄本字 */
            case 0x1c: case 0x39: {                            /* Enter/Space 選定 */
                int g = z->cand[z->cand_cur];
                dq3_zhuyin_init(z);                            /* 清組字、回 COMPOSE */
                return g;
            }
            default: return -1;
        }
    }
    /* COMPOSE:注音盤 2D 導航 + 選格 */
    {
        int row = z->cursor / DQ3_ZH_COLS, col = z->cursor % DQ3_ZH_COLS, nc;
        int rows = (DQ3_ZH_CELLS + DQ3_ZH_COLS - 1) / DQ3_ZH_COLS;
        switch (sc) {
            case 0x48: row = (row + rows - 1) % rows; break;
            case 0x50: row = (row + 1) % rows; break;
            case 0x4b: col = (col + DQ3_ZH_COLS - 1) % DQ3_ZH_COLS; break;
            case 0x4d: col = (col + 1) % DQ3_ZH_COLS; break;
            case 0x01: dq3_zhuyin_init(z); return -1;          /* ESC 清組字 */
            case 0x1c: case 0x39: compose_select(z, z->cursor); return -1;
            default: return -1;
        }
        nc = row * DQ3_ZH_COLS + col;
        if (nc >= DQ3_ZH_CELLS) nc = DQ3_ZH_CELLS - 1;
        z->cursor = nc;
        return -1;
    }
}

void dq3_zhuyin_render(const dq3_zhuyin *z, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                       int x, int y, uint8_t fg, uint8_t curfg)
{
    int i, r, c;
    /* 組字列:已選 聲母 介音 韻母 聲調 */
    int cx = x;
    if (z->sh) { dq3_text_draw_glyph(t,fb,fb_w,fb_h,cx,y,(uint16_t)(64+z->sh),fg); cx+=DQ3_GLYPH_PX; }
    if (z->ji) { dq3_text_draw_glyph(t,fb,fb_w,fb_h,cx,y,(uint16_t)(98+z->ji),fg); cx+=DQ3_GLYPH_PX; }
    if (z->yu) { dq3_text_draw_glyph(t,fb,fb_w,fb_h,cx,y,(uint16_t)(85+z->yu),fg); cx+=DQ3_GLYPH_PX; }

    if (z->phase == DQ3_ZH_PICK) {
        /* 候選窗 */
        for (i = 0; i < z->ncand; i++) {
            int gr = i / DQ3_ZH_COLS, gc = i % DQ3_ZH_COLS;
            uint8_t col = (i == z->cand_cur) ? curfg : fg;
            dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + 8 + gc*DQ3_GLYPH_PX, y + 24 + gr*18, z->cand[i], col);
        }
        return;
    }
    /* 注音盤 */
    for (i = 0; i < DQ3_ZH_CELLS; i++) {
        int gr = i / DQ3_ZH_COLS, gc = i % DQ3_ZH_COLS;
        int px = x + gc*DQ3_GLYPH_PX, py = y + 24 + gr*18;
        uint8_t col = (i == z->cursor) ? curfg : fg;
        uint16_t g = cell_glyph(i);
        if (g != 0xffff) dq3_text_draw_glyph(t, fb, fb_w, fb_h, px, py, g, col);
        else { for (r = 7; r < 9; r++) for (c = 2; c < 12; c++) {  /* 一聲:橫線 */
            int xx=px+c, yy=py+r; if (xx>=0&&xx<fb_w&&yy>=0&&yy<fb_h) fb[yy*fb_w+xx]=col; } }
    }
}
