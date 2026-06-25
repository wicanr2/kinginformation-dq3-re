/* dq3_warp.h — scripted 傳送表 0x4ea0(type-2 examine 事件目的地,docs/31/35)。 */
#ifndef DQ3_WARP_H
#define DQ3_WARP_H
typedef struct { unsigned char dest, x, y; } dq3_warp;   /* dest=目的 CTY、(x,y)=落點 */
extern const dq3_warp dq3_warps[];
extern const int dq3_warps_len;
/* 取 warp[param];回 0 並填 dest_cty/x/y,空項回 -1。 */
int dq3_warp_get(int param, int *dest_cty, int *x, int *y);
#endif
