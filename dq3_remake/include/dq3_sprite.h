/* dq3_sprite.h — DQ3MAN.BLS 角色 sprite 解碼(地表/城鎮主角 + NPC)。
 *
 * 格式(反組譯角色 blit sub_1ed0,docs/27):
 *   DQ3MAN.BLS = 6B 表頭 + 232 entry × 960 byte(stride 960)。
 *   每 entry(= 一個方向 frame)= 32×24,4-bit plane-major(plane3→0,384B 資料)
 *     + AND 遮罩(在資料 +0x180,96B 取 1 plane;bit=1 透明)。masked blit:
 *     dest = (dest AND mask) OR data。色盤 = 場景 DQ3.PAL 低 16 色。
 *   每角色 4 個連續 entry = 4 方向 frame。
 */
#ifndef DQ3_SPRITE_H
#define DQ3_SPRITE_H

#include <stdint.h>

#define DQ3_CHAR_W       32
#define DQ3_CHAR_H       24
#define DQ3_CHAR_FRAMES  4
#define DQ3_BLS_STRIDE   960
#define DQ3_BLS_MASKOFF  0x180

typedef struct {
    uint8_t px[DQ3_CHAR_FRAMES][DQ3_CHAR_H][DQ3_CHAR_W];      /* palette 索引 */
    uint8_t opaque[DQ3_CHAR_FRAMES][DQ3_CHAR_H][DQ3_CHAR_W];  /* 1=畫, 0=透明 */
    int loaded;
} dq3_charsprite;

/* 從 bls_name(主角=DQ3MST.BLS、NPC=DQ3MAN.BLS)載入 entry_base 起連續 4 個 frame。失敗回 <0。 */
int dq3_charsprite_load(dq3_charsprite *cs, const char *assets_dir, const char *bls_name,
                        int entry_base, char *err, int errcap);

#endif /* DQ3_SPRITE_H */
