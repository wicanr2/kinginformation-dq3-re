/* dq3_assets.h — DQ3 原始素材的 C 解碼器(從 tools/*.py + re/loaders.c 移植)。
 *
 * 各格式皆已在 RE 階段驗證(見 docs/02 字型、docs/04 地圖/tile/palette、
 * docs/09 loaders、docs/title)。此處把純資料解碼搬成 C,輸出 indexed pixel +
 * palette,供 dq3_runtime 顯示。解碼器不碰 SDL、不碰檔案(buffer in / buffer out),
 * 方便單元測試與跨平台。
 */
#ifndef DQ3_ASSETS_H
#define DQ3_ASSETS_H

#include <stdint.h>
#include <stddef.h>
#include "dq3_runtime.h"   /* dq3_color */

/* ---- TIT*.P 標題畫面(ZSoft PCX,640×350,4 plane 1bpp,RLE)----
 * 解碼到 out(out_w*out_h indexed),16 色 palette(檔內 EGA palette)寫入 pal16。
 * 回傳 0 成功;<0:-1 magic、-2 格式、-3 尺寸超出、-4 OOM。 */
int dq3_pcx_decode(const uint8_t *d, size_t n,
                   uint8_t *out, int out_w, int out_h, dq3_color pal16[16]);

/* ---- DQ3.PAL / MNSBK.PAL(VGA 6-bit DAC palette)----
 * 解碼到 pal(最多 max 色),回傳實際色數;6-bit→8-bit:v8=(v6<<2)|(v6>>4)。
 * DQ3.PAL=80 色(地表/城鎮 tile)、MNSBK.PAL=160 色(怪物/背景)。 */
int dq3_pal_decode(const uint8_t *d, size_t n, dq3_color *pal, int max);

/* ---- D3TXT00.FON 遊戲內字型(16×16,row-major,MSB-first,每字 32 byte)----
 * 把第 idx 個字模解到 glyph[16][16](值 0/1)。回傳 0 成功;字數 = n/32。 */
int dq3_font_glyph(const uint8_t *d, size_t n, int idx, uint8_t glyph[16][16]);
int dq3_font_count(size_t n);   /* = n / 32 */

/* ---- BLK tile 圖庫(header 6B + count 個 384B tile;32×24,4-plane planar)----
 * 解析 header;tile 像素由 dq3_blk_tile 解到 px[24][32](值 0..15 = palette 索引)。 */
typedef struct {
    const uint8_t *body;   /* 指向第一個 tile(d+6) */
    int row_bytes;         /* header[0]=4 */
    int height;            /* header[1]=24 */
    int count;             /* header[2] */
    int tile_size;         /* 每 tile byte 數(= row_bytes*height*4 = 384) */
    size_t body_len;
} dq3_blk;

int dq3_blk_open(const uint8_t *d, size_t n, dq3_blk *blk);
int dq3_blk_tile(const dq3_blk *blk, int idx, uint8_t px[24][32]);

#endif /* DQ3_ASSETS_H */
