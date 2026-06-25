/* dq3_rng.h — DQ3 亂數產生器,雙模式。
 *
 * 模式(全域,dq3_rng_set_mode):
 *   DQ3_RNG_DOS  (預設·忠實):鏡射 DQ3.EXE file 0xfa39 / logical 0xe6c9 —
 *       state = rol16(state + 0x9018, 3); return state % n;
 *       16-bit state → 週期 ≤ 65536,維度小(會重現原版亂數規律/偏差)。
 *   DQ3_RNG_REAL (真實):xorshift32,週期 2³²-1,分布佳。當年 DOS 亂數維度太小,
 *       此模式給「不在乎完全還原、要更好隨機品質」的玩家。
 *
 * 兩模式共用同一 next(r,n) 介面;seed 同時播種 16-bit 與 32-bit state。
 * 預設 DOS → 既有確定性序列/測試不變(seed0 → next(100)=64)。
 */
#ifndef DQ3_RNG_H
#define DQ3_RNG_H

#include <stdint.h>

enum { DQ3_RNG_DOS = 0, DQ3_RNG_REAL = 1 };

typedef struct {
    uint16_t state;    /* DOS 模式 state = DS:[0xb5a] */
    uint32_t state32;  /* REAL 模式 xorshift32 state(非 0)*/
} dq3_rng;

/* 全域模式(影響所有 dq3_rng_next)。預設 DQ3_RNG_DOS。 */
void dq3_rng_set_mode(int mode);
int  dq3_rng_get_mode(void);

void dq3_rng_seed(dq3_rng *r, uint16_t seed);

/* 回傳 [0,n) 內值並更新 state;n==0 回 0。依全域模式走 DOS 或 xorshift32。 */
int  dq3_rng_next(dq3_rng *r, int n);

#endif /* DQ3_RNG_H */
