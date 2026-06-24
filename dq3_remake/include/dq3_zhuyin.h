/* dq3_zhuyin.h — 注音輸入法(組字 + 同音候選挑字),對齊 docs/15 ni_zhuyin/compose。
 *
 * 原版用 EXE 內建注音引擎(seg 0x11c4);remake 改用「注音→字模候選表」(dq3_zhuyin_table.c,
 * 由 D3TXT00.FON 1476 字模的 unicode 經 pypinyin BOPOMOFO 反建,721 音節桶 / 1338 字)。
 * 組字:選 聲母 → (介音) → 韻母 → 聲調 → 查表得同音字 → 候選窗用方向鍵挑一個 → 加入名字。
 *
 * 注音符號 glyph(D3TXT00.FON):聲母 ㄅ-ㄙ 65-85(code1-21)、韻母 ㄚ-ㄦ 86-98(1-13)、
 * 介音 ㄧㄨㄩ 99-101(1-3)、聲調 ˊ103/ˇ104/ˋ105/˙102(一聲無符號=tone0)。
 */
#ifndef DQ3_ZHUYIN_H
#define DQ3_ZHUYIN_H

#include <stdint.h>
#include "dq3_text.h"

/* 生成表(dq3_zhuyin_table.c)*/
extern const int      dq3_zh_nbucket;
extern const uint16_t dq3_zh_key[];
extern const uint16_t dq3_zh_off[];
extern const uint16_t dq3_zh_pool[];

/* 查同音字模:聲母1-21/介音0-3/韻母1-13/聲調0-4 → 候選 glyph 寫進 out(最多 max)。回候選數。 */
int dq3_zhuyin_lookup(int sh, int ji, int yu, int tone, uint16_t *out, int max);

#define DQ3_ZH_CELLS   42      /* 21聲 + 13韻 + 3介 + 5聲調(含一聲)*/
#define DQ3_ZH_COLS    9
#define DQ3_ZH_MAXCAND 32

enum { DQ3_ZH_COMPOSE = 0, DQ3_ZH_PICK };

typedef struct {
    int sh, ji, yu, tone;          /* 組字緩衝(0=未選;tone:-1 未選 / 0-4)*/
    int cursor;                    /* 注音盤游標 0..DQ3_ZH_CELLS-1 */
    int phase;                     /* COMPOSE / PICK */
    uint16_t cand[DQ3_ZH_MAXCAND];
    int ncand, cand_cur;
} dq3_zhuyin;

void dq3_zhuyin_init(dq3_zhuyin *z);

/* 處理 scancode(方向 + Enter)。挑定一個字 → 回該 glyph(>=0)並清組字回 COMPOSE;否則回 -1。 */
int  dq3_zhuyin_input(dq3_zhuyin *z, uint8_t scancode);

/* 繪製:組字列(已選注音)+ 注音盤(游標)/ 候選窗(PICK 時)。 */
void dq3_zhuyin_render(const dq3_zhuyin *z, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                       int x, int y, uint8_t fg, uint8_t curfg);

#endif /* DQ3_ZHUYIN_H */
