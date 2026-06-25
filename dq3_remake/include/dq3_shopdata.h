/* dq3_shopdata.h — CTY 設施表(商店/旅社/教會/記錄點),自 CTY*.DAT 萃取。 */
#ifndef DQ3_SHOPDATA_H
#define DQ3_SHOPDATA_H

enum { DQ3_FAC_INN=0, DQ3_FAC_WEAPON=1, DQ3_FAC_ITEM=2, DQ3_FAC_CHURCH=3, DQ3_FAC_RECORD=4 };

typedef struct {
    unsigned char cty, sec, k;   /* CTY 號、section、設施索引(NPC byte4)*/
    unsigned char type;          /* DQ3_FAC_* */
    unsigned char count;         /* 商店品項數(非商店=0)*/
    unsigned short item_off;     /* 商店品項在 dq3_shop_itempool 的起始 */
    unsigned short inn_cost;     /* 旅社住宿費 raw(type=INN)*/
} dq3_facility;

extern const dq3_facility dq3_facilities[];
extern const int dq3_facilities_len;
extern const unsigned char dq3_shop_itempool[];
extern const int dq3_shop_itempool_len;

/* 取某 CTY 第一個指定類型商店的品項;回品項數,*items 指向 pool。找不到回 0。 */
int dq3_shop_items(int cty, int type, const unsigned char **items);

/* 依 (cty, sec, k=NPC byte4) 精確查設施;找不到回 NULL。供「面向店員 → 開該攤」。 */
const dq3_facility *dq3_facility_at(int cty, int sec, int k);

#endif
