/* dq3_packbg.h — 戰鬥背景(PACKBG.SCR)解碼。
 *
 * 反組譯 load_packbg_page(file 0xda55)+ 選頁 wrapper(0xd9f8):
 *   - 背景隨地形不同:terrain = tileattr 表 [tile+0x4df6];page = [0xd73] + terrain*8
 *     (terrain==3 → page 0x19);[0xd73] 初值 0。
 *   - 每頁:file 偏移 page*0x3d80 + page;讀 0x6e00 = 88 row × 4 plane × 0x50(80)byte,
 *     **640 寬 row-interleaved planar**(每 row plane0/1/2/3 各 80 byte,MSB-first);
 *     VGA 寫 mode 為 replace + AND-into-0xff,等價直接取 buffer bits。
 *   - 調色盤:MNSBK.PAL 子區塊,16 色,offset [0xd73]*0x30([0xd73]=0 → 前 16 色)。
 * 已驗:page 0(草原)解出乾淨 640×88 藍天白雲。
 */
#ifndef DQ3_PACKBG_H
#define DQ3_PACKBG_H

#include <stdint.h>

#define DQ3_PACKBG_W    640
#define DQ3_PACKBG_H    88

/* 解第 page 頁的天空背景到 out(88×640 indexed,索引 0..15 對 MNSBK.PAL 子調色盤)。
 * 失敗回 <0。 */
int dq3_packbg_decode(const char *assets_dir, int page,
                      uint8_t out[DQ3_PACKBG_H][DQ3_PACKBG_W], char *err, int errcap);

#endif /* DQ3_PACKBG_H */
