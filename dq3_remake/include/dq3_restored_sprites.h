/* dq3_restored_sprites.h — bug #3 復原 boss sprite(128 歐里狄加 / 129 五頭龍大王)。
 * 原版 DQ3MNS.SHP 這兩格為空(未完成),遭遇即當機;remake 在 sprite 解碼失敗時回退本資料。
 * 實作見生成檔 dq3_restored_sprites.c(tools/gen_restored_sprites.py)。 */
#ifndef DQ3_RESTORED_SPRITES_H
#define DQ3_RESTORED_SPRITES_H

#include "dq3_monster.h"

/* 填入第 id 隻復原 sprite(僅 128 / 129)。成功回 0,非復原 id 或失敗回 <0。 */
int dq3_restored_sprite(int id, dq3_monster_sprite *out);

#endif /* DQ3_RESTORED_SPRITES_H */
