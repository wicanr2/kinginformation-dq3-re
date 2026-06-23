/* dq3_inventory.h — 道具欄 + 合成事件(含 #2 彩虹水滴修正)。
 *
 * 反組譯合成事件(file 0x77ce,docs/18/22):拉達多姆神聖祠堂「雨和太陽合而為一」劇情 ——
 * 消耗 太陽之石(0x72)+ 雲雨之杖(0x73)→ 在雲雨之杖格寫入合成品,設「已合成」旗標 0x139。
 * 原版 bug #2:合成品寫死 0x6b(銀寶珠),正確應為 0x75(彩虹水滴),否則架不了彩虹橋卡關。
 */
#ifndef DQ3_INVENTORY_H
#define DQ3_INVENTORY_H

#include <stdint.h>

#define DQ3_INV_SLOTS   16
#define DQ3_ITEM_NONE   0xff

/* 道具(item code,對 ITEM.DAT;0xff=空) */
#define ITEM_SUN_STONE     0x72   /* 太陽之石 */
#define ITEM_RAINCLOUD_ROD 0x73   /* 雲雨之杖 */
#define ITEM_SILVER_ORB    0x6b   /* 銀寶珠(原版 bug 誤產)*/
#define ITEM_RAINBOW_DROP  0x75   /* 彩虹水滴(正確合成品)*/

typedef struct { uint8_t slot[DQ3_INV_SLOTS]; } dq3_inventory;

void dq3_inv_init(dq3_inventory *inv);                 /* 全清空(0xff) */
int  dq3_inv_find(const dq3_inventory *inv, int item); /* 回 slot index,無回 -1 */
int  dq3_inv_has(const dq3_inventory *inv, int item);
int  dq3_inv_add(dq3_inventory *inv, int item);        /* 放第一個空格;滿回 -1 */
int  dq3_inv_remove(dq3_inventory *inv, int item);     /* 移除一個;回 slot,無回 -1 */

/* #2 合成事件:太陽之石 + 雲雨之杖 → 彩虹水滴(消耗兩者,在雲雨之杖格寫入成品)。
 * fixed!=0:產出 0x75(彩虹水滴,修正);fixed==0:重現原版 bug 產出 0x6b(銀寶珠)。
 * 回傳產出的 item code;材料不足回 -1。 */
int  dq3_synth_rainbow_drop(dq3_inventory *inv, int fixed);

#endif /* DQ3_INVENTORY_H */
