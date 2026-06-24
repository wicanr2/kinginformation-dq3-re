/* dq3_tavern.c — 露依達酒場創角流程狀態機。 */
#include "dq3_tavern.h"

/* 性別標籤 glyph(rec556)*/
static const uint16_t GENDER_M[2] = {775, 674};   /* 男性 */
static const uint16_t GENDER_F[2] = {234, 674};   /* 女性(234=女,OCR 誤標「文」)*/

static void build_gender_menu(dq3_menu *m, int x, int y)
{
    dq3_menu_init(m, x, y);
    dq3_menu_add(m, GENDER_M, 2);
    dq3_menu_add(m, GENDER_F, 2);
}

void dq3_tavern_init(dq3_tavern *tv, dq3_roster *roster, dq3_party *party, const dq3_stats *st)
{
    tv->state = DQ3_TAV_CLASS;
    tv->roster = roster; tv->party = party; tv->st = st;
    tv->roster_cursor = 0;
    tv->pend_class = 0; tv->pend_gender = 0; tv->last_created = -1;
    dq3_roster_class_menu(&tv->class_menu, 52, 80);    /* 視窗內預設位置 */
    build_gender_menu(&tv->gender_menu, 52, 96);
    dq3_nameinput_init(&tv->ni);
    dq3_zhuyin_init(&tv->zh);
    tv->name_mode = 0;
}

int dq3_tavern_input(dq3_tavern *tv, uint8_t sc)
{
    switch (tv->state) {
    case DQ3_TAV_CLASS: {
        int sel = dq3_menu_input(&tv->class_menu, sc);
        if (sel >= 0) { tv->pend_class = sel; dq3_nameinput_init(&tv->ni); dq3_zhuyin_init(&tv->zh);
                        tv->name_mode = 0; tv->state = DQ3_TAV_NAME; }
        break;
    }
    case DQ3_TAV_NAME:
        if (sc == 0x0f) { tv->name_mode ^= 1; break; }   /* Tab:英數 ↔ 注音 切換 */
        if (tv->name_mode == 1) {                         /* 注音輸入 */
            int g = dq3_zhuyin_input(&tv->zh, sc);
            if (g >= 0 && tv->ni.len < DQ3_NI_NAME_MAX)   /* 挑定一個國字 → 加入名字緩衝 */
                tv->ni.buf[tv->ni.len++] = (uint16_t)g;
        } else {
            dq3_nameinput_input(&tv->ni, sc);            /* 英數輸入 */
            if (tv->ni.done) { tv->gender_menu.cursor = 0; tv->state = DQ3_TAV_GENDER; }
        }
        break;
    case DQ3_TAV_GENDER: {
        int sel = dq3_menu_input(&tv->gender_menu, sc);
        if (sel >= 0) {
            tv->pend_gender = sel;
            tv->last_created = dq3_roster_create(tv->roster, tv->st, tv->pend_class,
                                                 sel, tv->ni.buf, tv->ni.len);
            if (tv->last_created >= 0) dq3_party_add(tv->party, tv->roster, tv->last_created);
            tv->roster_cursor = (tv->last_created >= 0) ? tv->last_created : 0;
            tv->state = DQ3_TAV_ROSTER;
        }
        break;
    }
    case DQ3_TAV_ROSTER: {
        int n = tv->roster->count;
        if (n <= 0) { tv->state = DQ3_TAV_CLASS; break; }
        switch (sc) {
        case 0x48: tv->roster_cursor = (tv->roster_cursor + n - 1) % n; break;   /* 上 */
        case 0x50: tv->roster_cursor = (tv->roster_cursor + 1) % n; break;       /* 下 */
        case 0x1c: {                                                             /* Enter:切換入隊 */
            dq3_recruit *rc = &tv->roster->list[tv->roster_cursor];
            if (rc->in_party) {
                int s; for (s = 0; s < DQ3_PARTY_MAX; s++)
                    if (tv->party->slot[s] == tv->roster_cursor) { dq3_party_remove(tv->party, tv->roster, s); break; }
            } else dq3_party_add(tv->party, tv->roster, tv->roster_cursor);
            break;
        }
        case 0x39:                                                              /* Space:新增另一名 */
            tv->class_menu.cursor = 0; tv->state = DQ3_TAV_CLASS; break;
        default: break;
        }
        break;
    }
    }
    return tv->state;
}

/* ---- 繪製 ---- */
static void draw_num(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h, int x, int y, int v, uint8_t fg)
{
    uint16_t d[6]; int k = 0, i;
    if (v < 0) v = 0;
    if (v == 0) d[k++] = 0;
    else { int x2 = v; uint16_t tmp[6]; int tn = 0;
        while (x2 > 0 && tn < 6) { tmp[tn++] = (uint16_t)(x2 % 10); x2 /= 10; }
        for (i = tn - 1; i >= 0; i--) d[k++] = tmp[i]; }
    for (i = 0; i < k; i++) dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + i*DQ3_GLYPH_PX, y, d[i], fg);
}

static void render_roster(const dq3_tavern *tv, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                          int x, int y, uint8_t fg, uint8_t curfg)
{
    int i, j;
    for (i = 0; i < tv->roster->count; i++) {
        const dq3_recruit *rc = &tv->roster->list[i];
        int ly = y + i * 18, cx = x + 16;
        uint8_t col = (i == tv->roster_cursor) ? curfg : fg;
        if (i == tv->roster_cursor) { int r,c; for(r=4;r<12;r++)for(c=0;c<(r<8?r-3:12-r);c++){
            int xx=x+2+c,yy=ly+r; if(xx>=0&&xx<fb_w&&yy>=0&&yy<fb_h) fb[yy*fb_w+xx]=curfg; } }
        for (j = 0; j < rc->name_len; j++) { dq3_text_draw_glyph(t,fb,fb_w,fb_h,cx,ly,rc->name[j],col); cx += DQ3_GLYPH_PX; }
        cx += 8;
        { const struct dq3_class_name *cn = &dq3_class_names[rc->m.cls]; int g;
          for (g=0; g<cn->len; g++){ dq3_text_draw_glyph(t,fb,fb_w,fb_h,cx,ly,cn->g[g],col); cx += DQ3_GLYPH_PX; } }
        cx += 8; draw_num(t, fb, fb_w, fb_h, cx, ly, rc->m.level, col); cx += 2*DQ3_GLYPH_PX + 6;
        if (rc->in_party) { int r,c; for(r=4;r<12;r++)for(c=0;c<8;c++){
            int xx=cx+c,yy=ly+r; if(xx>=0&&xx<fb_w&&yy>=0&&yy<fb_h) fb[yy*fb_w+xx]=curfg; } }
    }
}

void dq3_tavern_render(const dq3_tavern *tv, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                       int x, int y, uint8_t fg, uint8_t curfg)
{
    switch (tv->state) {
    case DQ3_TAV_CLASS:  dq3_menu_render(&tv->class_menu, t, fb, fb_w, fb_h, fg, curfg); break;
    case DQ3_TAV_NAME:
        if (tv->name_mode == 1) {                          /* 注音:名字緩衝列 + 注音盤/候選 */
            int i; for (i = 0; i < tv->ni.len; i++)
                dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + i*DQ3_GLYPH_PX, y, tv->ni.buf[i], fg);
            dq3_zhuyin_render(&tv->zh, t, fb, fb_w, fb_h, x, y + 20, fg, curfg);
        } else {
            dq3_nameinput_render(&tv->ni, t, fb, fb_w, fb_h, x, y, fg, curfg);
        }
        break;
    case DQ3_TAV_GENDER: dq3_menu_render(&tv->gender_menu, t, fb, fb_w, fb_h, fg, curfg); break;
    case DQ3_TAV_ROSTER: render_roster(tv, t, fb, fb_w, fb_h, x, y, fg, curfg); break;
    }
}
