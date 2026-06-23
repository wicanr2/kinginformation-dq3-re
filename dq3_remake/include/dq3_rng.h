/* dq3_rng.h — DQ3.EXE 亂數產生器(鏡射 file 0xfa39 / logical 0xe6c9)。
 *
 * 反組譯(L0e6d2):
 *   if (n==0) return 0;
 *   state = rol16(state + 0x9018, 3);   // [0xb5a] += 0x9018; rol ax,1 ×3
 *   return state % n;                    // dx:ax / bx → 餘數 dx
 * 與 dq3_battlescene 的 xorshift 不同:NPC 移動 / 原版邏輯要的是「這一套」確定性序列。 */
#ifndef DQ3_RNG_H
#define DQ3_RNG_H

#include <stdint.h>

typedef struct { uint16_t state; } dq3_rng;   /* state = DS:[0xb5a] */

void dq3_rng_seed(dq3_rng *r, uint16_t seed);

/* 回傳 state%n(更新 state);n==0 回 0。鏡射 0xe6c9 的 add 0x9018 / rol3 / div。 */
int  dq3_rng_next(dq3_rng *r, int n);

#endif /* DQ3_RNG_H */
