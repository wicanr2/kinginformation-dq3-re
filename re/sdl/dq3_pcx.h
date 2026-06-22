/* dq3_pcx.h — TIT*.P 標題 PCX 解碼介面 */
#ifndef DQ3_PCX_H
#define DQ3_PCX_H

#include <stdint.h>
#include <stddef.h>
#include "dq3_runtime.h"   /* dq3_color */

/* 解碼 PCX(d,n)到 out(out_w*out_h packed 索引),16 色 palette 寫入 pal16。
 * 回傳 0 成功;<0 失敗(-1 magic,-2 格式,-3 尺寸,-4 OOM)。 */
int dq3_pcx_decode(const uint8_t *d, size_t n,
                   uint8_t *out, int out_w, int out_h, dq3_color pal16[16]);

#endif
