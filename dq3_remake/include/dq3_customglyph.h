/* dq3_customglyph.h — 自建字形(D3TXT00.FON 缺的 CJK)。
 * 位元資料由 tools/gen_custom_glyphs.py 從 CJK 字型 rasterize(同原版 16×16 MSB-first 格式)。
 * 用途:remake 自有 UI 詞(如「設定」的「設」)在原版字庫沒有時補上,不依賴原版 record。 */
#ifndef DQ3_CUSTOMGLYPH_H
#define DQ3_CUSTOMGLYPH_H
#include <stdint.h>

/* glyph 陣列中以 (DQ3_CG_BASE + 索引) 引用自建字形;dq3_text_draw_glyph 會 fallback 到此模組。
 * 0xf000 遠高於原版字庫 1476 字,不衝突。 */
#define DQ3_CG_BASE 0xf000
/* 索引(對齊 dq3_customglyph.c 產生順序:tools/gen_custom_glyphs.py 的字元順序)。 */
enum { DQ3_CG_SHE = 0,   /* 設 */
       DQ3_CG_BAN = 1 }; /* 版 */
#define DQ3_GLYPH_SHE (DQ3_CG_BASE + DQ3_CG_SHE)   /* 「設」在 glyph 陣列中的值 */
#define DQ3_GLYPH_BAN (DQ3_CG_BASE + DQ3_CG_BAN)   /* 「版」 */

extern const uint16_t dq3_customglyph_bits[][16];
extern const int dq3_customglyph_count;

/* 在 (x,y) 畫自建字形 idx(16×16,字身 row0..13),色 fg。idx 越界則不畫。 */
void dq3_customglyph_draw(int idx, uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg);

#endif
