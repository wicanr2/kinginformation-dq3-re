/* dq3_treasure.h — CTY 寶箱 / 隱藏物品表(自 CTY*.DAT examine 事件表萃取)。 */
#ifndef DQ3_TREASURE_H
#define DQ3_TREASURE_H

typedef struct {
    unsigned char cty, sec, sub;   /* CTY 號、section、事件 subid */
    unsigned char type;            /* 0=調査點/寶箱 1/3=寶箱給道具 */
    unsigned char item;            /* 道具 id(= ITEM.DAT index;名 = D3TXT00 rec id+1)*/
    unsigned char flag;            /* 一次性旗標 id([0x4f70] bit;set=未取)*/
} dq3_treasure;

extern const dq3_treasure dq3_treasures[];
extern const int dq3_treasures_len;
/* 某 CTY 的寶箱清單;回數量,*out 指向首筆(連續)。 */
int dq3_treasures_for(int cty, const dq3_treasure **out);

#endif
