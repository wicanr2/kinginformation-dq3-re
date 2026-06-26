/* dq3_item_use.h — 消耗品使用效果(#3)。
 *
 * 道具 id → 效果種類:id 對應由商店貨架交叉驗證(docs/49,RE 確鑿:ITEM.DAT 價格
 * 全中經典 DQ3 值)。效果「種類」是 RE 事實;需數值的量(藥草治療量、聖水步數)
 * 在未定位的 use-effect 路徑——與本專案咒文威力表同處境(docs/13:312),故採經典 DQ3
 * 值並於 docs/49 標註來源,不自由臆造。
 */
#ifndef DQ3_ITEM_USE_H
#define DQ3_ITEM_USE_H
#include "dq3_stats.h"

/* 消耗品 id(商店交叉驗證,docs/49)。 */
#define DQ3_ITEM_HERB        0x41   /* 藥草:回 HP */
#define DQ3_ITEM_ANTIDOTE    0x42   /* 驅毒草:解毒 */
#define DQ3_ITEM_CHIMERA_WING 0x43  /* 蓋美拉翅膀:回最後城鎮 */
#define DQ3_ITEM_HOLY_WATER  0x44   /* 聖水:暫時驅弱敵 */
#define DQ3_ITEM_FULLMOON    0x45   /* 滿月草:解麻痺 */

/* 效果種類。 */
enum {
    DQ3_USE_NONE = 0,        /* 不可使用(裝備/鑰匙/劇情物)*/
    DQ3_USE_HEAL_HP,         /* 回復 HP(藥草)*/
    DQ3_USE_CURE_POISON,     /* 解毒(驅毒草)— 待狀態系統(docs/47 #5)*/
    DQ3_USE_CURE_PARALYSIS,  /* 解麻痺(滿月草)— 待狀態系統 */
    DQ3_USE_RETURN_TOWN,     /* 回最後城鎮(蓋美拉翅膀)*/
    DQ3_USE_REPEL,           /* 暫時驅弱敵(聖水)*/
    DQ3_USE_AWAKEN,          /* 覺醒粉 0x5a:在諾阿尼魯村(CTY4)用 → 解全村催眠(杜勝利 Ch11)。
                              * 位置相關:消耗延到 main 確認在 CTY4 才執行。 */
    DQ3_USE_GAIA             /* 蓋亞之劍 0x0f:在阿莎拉慕小火山旁用 → 熔流開通往尼羅肯特之路
                              * (杜勝利 Ch40)。武器,不消耗;設「火山開通」旗標。 */
};

/* 道具 id → 效果種類。 */
int dq3_item_use_kind(int item_id);

/* 藥草固定治療量(經典 DQ3 值,docs/49)。 */
#define DQ3_HERB_HEAL   30
/* 聖水驅敵步數(經典 DQ3 值)。 */
#define DQ3_HOLY_STEPS  64

/* 對隊員套 HEAL_HP:cur_hp += amount,封頂於 max(stat[HP]);陣亡(cur_hp==0)不可用藥草。
 * 回實際回復量;不可治(已滿/陣亡)回 0;非治療道具回 -1。 */
int dq3_item_use_heal(dq3_member *m, int item_id);

/* 解狀態道具:驅毒草→清 POISON、滿月草→清 PARALYSIS。清掉回 1;該道具非解狀態 / 無對應
 * 異常回 0。對隊員 m 套用。 */
int dq3_item_use_cure(dq3_member *m, int item_id);

#endif
