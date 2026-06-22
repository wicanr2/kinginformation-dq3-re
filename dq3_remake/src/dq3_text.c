/* dq3_text.c — 遊戲文字/對話渲染實作。 */
#include "dq3_text.h"
#include "dq3_assets.h"   /* dq3_font_glyph */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static uint8_t *slurp(const char *dir, const char *name, size_t *len)
{
    char path[2048]; FILE *f; long sz; uint8_t *b;
    snprintf(path, sizeof path, "%s/%s", dir, name);
    f = fopen(path, "rb"); if (!f) { *len = 0; return NULL; }
    fseek(f,0,SEEK_END); sz=ftell(f); fseek(f,0,SEEK_SET);
    if (sz<=0){fclose(f);*len=0;return NULL;}
    b=malloc((size_t)sz);
    if(!b||fread(b,1,(size_t)sz,f)!=(size_t)sz){fclose(f);free(b);*len=0;return NULL;}
    fclose(f); *len=(size_t)sz; return b;
}
static uint16_t rd16(const uint8_t *d, size_t o){ return (uint16_t)(d[o]|(d[o+1]<<8)); }

int dq3_text_load(dq3_text *t, const char *dir, const char *txt_name, char *err, int errcap)
{
    memset(t,0,sizeof *t);
    t->fon = slurp(dir, "D3TXT00.FON", &t->fon_len);
    if (!t->fon){ if(err)snprintf(err,errcap,"open D3TXT00.FON"); return -1; }
    t->txt = slurp(dir, txt_name, &t->txt_len);
    if (!t->txt){ if(err)snprintf(err,errcap,"open %s",txt_name); free(t->fon); t->fon=NULL; return -1; }
    if (t->txt_len < 2){ if(err)snprintf(err,errcap,"txt too small"); dq3_text_free(t); return -1; }
    t->n_records = rd16(t->txt,0)/2 - 1;   /* 指標表項數 -1(末項=結尾哨兵) */
    if (t->n_records < 0) t->n_records = 0;
    t->loaded = 1;
    return 0;
}

void dq3_text_free(dq3_text *t){ if(t){ free(t->fon); free(t->txt); t->fon=NULL; t->txt=NULL; t->loaded=0; } }

int dq3_text_record(const dq3_text *t, int rec, uint16_t *out, int max)
{
    size_t start, end, o; int k=0;
    if (!t->loaded || rec < 0 || rec >= t->n_records) return -1;
    start = rd16(t->txt, (size_t)rec*2);
    end   = rd16(t->txt, (size_t)(rec+1)*2);
    if (start > t->txt_len || end > t->txt_len || end < start) return -1;
    for (o = start; o + 1 < end && k < max; o += 2) {
        uint16_t v = rd16(t->txt, o);
        if (v == DQ3_TXT_END) break;
        out[k++] = v;
    }
    return k;
}

void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{
    uint8_t g[16][16]; int r,c;
    if (dq3_font_glyph(t->fon, t->fon_len, idx, g) != 0) return;
    for (r=0;r<16;r++) for(c=0;c<16;c++){
        int yy=y+r, xx=x+c;
        if (g[r][c] && yy>=0 && yy<fb_h && xx>=0 && xx<fb_w) fb[yy*fb_w+xx]=fg;
    }
}

int dq3_text_draw_record(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x0, int y0, int cols, int maxlines, int rec, uint8_t fg)
{
    uint16_t buf[2048];
    int n = dq3_text_record(t, rec, buf, 2048);
    int i, col=0, line=0;
    if (n < 0) return -1;
    for (i=0;i<n;i++){
        uint16_t v = buf[i];
        if (v==DQ3_TXT_NL || v==DQ3_TXT_NL2){ col=0; line++; if(maxlines&&line>=maxlines)break; continue; }
        if (v==DQ3_TXT_PAGE){ if(maxlines && line+1>=maxlines) break; col=0; line++; continue; }
        if (v>=0xffed){ col++; if(col>=cols){col=0;line++;} continue; }  /* 變數占位→空白 */
        if (v>=1476){ continue; }                                        /* 防越界 */
        dq3_text_draw_glyph(t, fb, fb_w, fb_h, x0+col*DQ3_GLYPH_PX, y0+line*DQ3_GLYPH_PX, (int)v, fg);
        col++;
        if (col>=cols){ col=0; line++; if(maxlines&&line>=maxlines)break; }
    }
    return line+1;
}
