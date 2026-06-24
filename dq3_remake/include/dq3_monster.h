/* dq3_monster.h — 怪物資料(D3MNS.DAT 數值)+ sprite(DQ3MNS.SHP)。
 *
 * D3MNS.DAT = 130 筆 × 41(0x29)byte(docs/16),index = monster_id。
 *   +0x00 u16 hp_base   +0x02 u16 hp_rand   +0x08 u8 atk?   +0x09 u16 def?
 *   +0x0b u16 agi?      +0x18 u8 flee_resist +0x21 u16 exp   +0x23 u16 gold
 *   +0x28 u8 spawn_weight
 * DQ3MNS.SHP = 131×u32 offset table + 130 隻 plane-major sprite(docs/16):
 *   per sprite: u16 w_bytes(&0x7fff,×8=寬)、u16 h、4 plane(plane3→0)各 h×w_bytes,
 *   MSB-first;色 = Σ plane_bit<<plane;色 0 = 透明。palette = MNSBK.PAL。
 */
#ifndef DQ3_MONSTER_H
#define DQ3_MONSTER_H

#include <stdint.h>
#include <stddef.h>

#define DQ3_MONSTER_COUNT 130
#define DQ3_MONSTER_REC   41

typedef struct {
    uint16_t hp_base, hp_rand;
    uint8_t  atk;
    uint16_t def, agi;
    uint8_t  flee_resist;
    uint16_t exp, gold;
    uint8_t  spawn_weight;
} dq3_monster_stat;

typedef struct {
    uint8_t *raw; size_t raw_len;     /* D3MNS.DAT 整檔(owned) */
    int loaded;
} dq3_monsters;

int  dq3_monsters_load(dq3_monsters *m, const char *assets_dir, char *err, int errcap);
void dq3_monsters_free(dq3_monsters *m);
int  dq3_monster_get_stat(const dq3_monsters *m, int id, dq3_monster_stat *out);
/* 遭遇時的當前 HP = hp_base + rnd(hp_rand)(此處給確定性上限 base+hp_rand)。 */
uint16_t dq3_monster_hp_max(const dq3_monsters *m, int id);

/* ---- SHP sprite ---- */
#define DQ3_SHP_MAXW 160   /* 原版 sprite 寬多達 ~288;160 涵蓋多數含復原 boss(96/128)*/
#define DQ3_SHP_MAXH 160
typedef struct {
    int w, h;                                   /* 像素寬高 */
    uint8_t px[DQ3_SHP_MAXH][DQ3_SHP_MAXW];     /* palette 索引(MNSBK.PAL) */
    uint8_t opaque[DQ3_SHP_MAXH][DQ3_SHP_MAXW]; /* 1=畫,0=透明(色0) */
} dq3_monster_sprite;

/* 從 DQ3MNS.SHP 解第 id 隻 sprite。失敗回 <0(含尺寸超出 / 空 sprite)。 */
int dq3_monster_sprite_decode(const char *assets_dir, int id,
                              dq3_monster_sprite *out, char *err, int errcap);

/* 取得 sprite:先試 DQ3MNS.SHP;若為空(未完成 boss 128/129)回退復原資料。成功回 0。 */
int dq3_monster_sprite_get(const char *assets_dir, int id,
                           dq3_monster_sprite *out, char *err, int errcap);

#endif /* DQ3_MONSTER_H */
