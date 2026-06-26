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
#define DQ3_ITEM_PRAYER_RING 0x48   /* 祈禱之戒:回 MP,每次使用 ~25.4% 機率損壞(#7c)*/

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
    DQ3_USE_GAIA,            /* 蓋亞之劍 0x0f:在阿莎拉慕小火山旁用 → 熔流開通往尼羅肯特之路
                              * (杜勝利 Ch40)。武器,不消耗;設「火山開通」旗標。 */
    DQ3_USE_DRAIN,           /* 乾渴壺 0x5e:在四島礁附近地表用 → 吸乾海水顯現最終鑰匙祠堂
                              * (杜勝利 Ch27)。不消耗(可重用);設「祠堂顯現」旗標。 */
    DQ3_USE_FAIRYFLUTE,      /* 妖精之笛 0x77:在魯比斯之塔(CTY82)用 → 解魯比斯詛咒,得精靈的守護 0x74
                              * (杜勝利 Ch53)。不消耗;精靈的守護 → 精靈祠堂 CTY92 換雲雨之杖。 */
    DQ3_USE_PRAYER_RING,     /* 祈禱之戒 0x48:野外用 → 回復一名隊員 MP,每次使用 ~25.4% 機率損壞消失
                              * (#7c)。RE file 0x5ad0 通用回復 handler MP 分支(+0x18 MP / +0x2c MaxMP):
                              * MP = min(MP + rng[bp,bp+10], MaxMP);0x5af4 破壞檢查 al≤0x40(~25.4%)
                              * → rec0x188「戒指壞了」+ remove。bp(回復力)由 dispatcher 傳,ITEM.DAT 無
                              * 使用威力欄、bp 未靜態釘死 → 回復量採 classic 值(同藥草 DQ3_HERB_HEAL 處置);
                              * 破壞機率為 RE 精確值。青衫「永不壞」前提經 RE 推翻(損壞邏輯本就存在)。 */
    DQ3_USE_RAINBOW          /* 彩虹水滴 0x75:在下層利姆達爾(CTY86)西北盡頭地表用 → 架彩虹橋通終盤
                              * (杜勝利 Ch55「雨和太陽合而為一出現彩虹橋」)。消耗;設「彩虹橋已架」旗標。
                              * 註:合成 bug #2 已修(產 0x75 非銀寶珠 0x6b);此為彩虹水滴的「用途」接線。 */
};

/* 道具 id → 效果種類。 */
int dq3_item_use_kind(int item_id);

/* 藥草固定治療量(經典 DQ3 值,docs/49)。 */
#define DQ3_HERB_HEAL   30
/* 聖水驅敵步數(經典 DQ3 值)。 */
#define DQ3_HOLY_STEPS  64
/* 祈禱之戒 MP 回復量(classic 值;RE handler 為 rng[bp,bp+10],bp 未靜態釘死,見 enum 註)。 */
#define DQ3_PRAYER_MP   30
/* 祈禱之戒每次使用損壞門檻:RNG(256) ≤ 此值 → 損壞(RE file 0x5af9 cmp al,0x40;0x41/0x100≈25.4%)。 */
#define DQ3_PRAYER_BREAK_LE  0x40

/* 對隊員套 HEAL_HP:cur_hp += amount,封頂於 max(stat[HP]);陣亡(cur_hp==0)不可用藥草。
 * 回實際回復量;不可治(已滿/陣亡)回 0;非治療道具回 -1。 */
int dq3_item_use_heal(dq3_member *m, int item_id);

/* 解狀態道具:驅毒草→清 POISON、滿月草→清 PARALYSIS。清掉回 1;該道具非解狀態 / 無對應
 * 異常回 0。對隊員 m 套用。 */
int dq3_item_use_cure(dq3_member *m, int item_id);

/* 祈禱之戒:對隊員 m 回復 MP(cur_mp += DQ3_PRAYER_MP,封頂 stat[MP])。回實際回復量;
 * MP 已滿回 0。損壞判定(RNG)由呼叫端處理(需 inventory 移除)。 */
int dq3_item_use_prayer_mp(dq3_member *m);

#endif
