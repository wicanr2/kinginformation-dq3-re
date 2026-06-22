/* dq3_blk.c — BLK tile 圖庫解碼(地表 DQ3.BLK / 城鎮 DQ3N.BLK)。
 * 對應 docs/04 + tools/tile_lib.py:header 6B(row_bytes=4, height=24, count),
 * 每 tile 384B = 32×24,4-bit planar(4 plane 各 96B,row×4byte,MSB-first)。
 */
#include "dq3_assets.h"

int dq3_blk_open(const uint8_t *d, size_t n, dq3_blk *blk)
{
    if (n < 6 || !blk) return -1;
    blk->row_bytes = d[0] | (d[1] << 8);
    blk->height    = d[2] | (d[3] << 8);
    blk->count     = d[4] | (d[5] << 8);
    blk->tile_size = blk->row_bytes * blk->height * 4;   /* 4 planes */
    blk->body      = d + 6;
    blk->body_len  = n - 6;
    if (blk->count <= 0 || blk->tile_size <= 0) return -2;
    if ((size_t)blk->count * blk->tile_size > blk->body_len) return -3;
    return 0;
}

int dq3_blk_tile(const dq3_blk *blk, int idx, uint8_t px[24][32])
{
    const uint8_t *tile;
    int plane, row, bx, bit, planesz;
    if (!blk || idx < 0 || idx >= blk->count) return -1;
    tile = blk->body + (size_t)idx * blk->tile_size;
    planesz = blk->row_bytes * blk->height;   /* 96 */
    for (row = 0; row < 24; row++)
        for (bx = 0; bx < 32; bx++) px[row][bx] = 0;
    for (plane = 0; plane < 4; plane++) {
        int base = plane * planesz;
        for (row = 0; row < blk->height && row < 24; row++) {
            int ro = base + row * blk->row_bytes;
            for (bx = 0; bx < blk->row_bytes && bx < 4; bx++) {
                uint8_t byte = tile[ro + bx];
                if (!byte) continue;
                for (bit = 0; bit < 8; bit++)
                    if (byte & (0x80 >> bit))
                        px[row][bx * 8 + bit] |= (uint8_t)(1 << plane);
            }
        }
    }
    return 0;
}
