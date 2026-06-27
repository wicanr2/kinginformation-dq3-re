/* dq3_roster.h — 露依達酒場創角 + 冒險者名冊 + 隊伍編成(邏輯核心)。
 *
 * DQ3 開場在阿里阿罕「露依達的店」建立同伴:選職業 → 性別 → 取名 → 登錄名冊,
 * 再從名冊編入隊伍(主角 + 最多 3 名)。本模組是這套流程的資料核心;UI 繪製、
 * 注音姓名輸入移植(docs/15)、忠實初始擲值(RE 補)另做。
 *
 * 對應 RE:角色建立 modal sub_0854(docs/15;設角色結構 [0x5082]、隊伍人數 [0x5077]);
 * 登錄所對話 D3TXT00 rec 550+;隊員實體表 [bx+0x4f15](stride;名字 / 數值 / 職業 / 性別)。
 *
 * 初始數值:**忠實 rng 擲值已實作**(dq3_member_init_rng,RE sub_d9cc:Lv1 = base + rng(0..slope/2);
 * roster_create 走 create_rng() seed 0x4321,DOS/REAL 模式)。精訊版無性格系統(已確認)。
 */
#ifndef DQ3_ROSTER_H
#define DQ3_ROSTER_H

#include <stdint.h>
#include "dq3_stats.h"   /* dq3_member / dq3_stats */
#include "dq3_menu.h"    /* dq3_menu(職業選單)*/

#define DQ3_NAME_MAX    4    /* DQ3 名字最多 4 字(glyph index;0=空)*/
#define DQ3_ROSTER_MAX  16   /* 名冊容量 */
#define DQ3_PARTY_MAX   4    /* 隊伍:主角 + 3 同伴 */

enum { DQ3_GENDER_MALE = 0, DQ3_GENDER_FEMALE = 1 };

typedef struct {
    uint16_t name[DQ3_NAME_MAX];   /* glyph index(對 D3TXT00.FON)*/
    int      name_len;
    int      gender;               /* DQ3_GENDER_* */
    int      in_party;             /* 是否在現役隊伍中 */
    dq3_member m;                  /* 職業 / 等級 / 數值 / 經驗 */
    /* 裝備 4 槽(RE:ITEM.DAT b4 高位,精訊版 武器/鎧/盾/兜;0xff=無)。dq3_item_equip_slot 判部位。
     * 攻擊力 = 力量 + 武器 b0;守備 = 耐力/2 + (鎧+盾+兜 b1 總和)。飾品無乾淨部位編碼,不設槽。 */
    uint8_t  weapon;               /* 武器 item code */
    uint8_t  armor;                /* 鎧(身)item code */
    uint8_t  shield;               /* 盾 item code */
    uint8_t  head;                 /* 兜(頭)item code */
} dq3_recruit;

typedef struct {
    dq3_recruit list[DQ3_ROSTER_MAX];
    int count;
} dq3_roster;

typedef struct {
    int slot[DQ3_PARTY_MAX];       /* 名冊 index;-1=空 */
    int count;
} dq3_party;

/* 職業名(D3TXT00.FON glyph index)。供職業選單 / 名冊顯示。 */
extern const struct dq3_class_name { int len; uint16_t g[4]; } dq3_class_names[DQ3_NUM_CLASS];

/* 建立 8 職業選單(用 dq3_class_names 的 glyph 標籤)。 */
void dq3_roster_class_menu(dq3_menu *m, int x, int y);

/* 職業(0..7)+ 性別(0♂/1♀)→ DQ3MST.BLS sprite entry_base = (cls*2+gender)*4。
 * 對映經實機/charsheet 確認:c0=勇者♂、c1=勇者♀、c2=戰士♂…c15=遊玩者♀(docs/27)。 */
int dq3_class_sprite_entry(int cls, int gender);

void dq3_roster_init(dq3_roster *r);

/* 建立新冒險者並登錄名冊:職業 cls(0..7)、gender、name(glyph 陣列,len≤DQ3_NAME_MAX)。
 * 用成長表初始化 Lv1 數值(dq3_member_init(cls,1))。回名冊 index;名冊滿 / 參數錯回 -1。 */
int  dq3_roster_create(dq3_roster *r, const dq3_stats *st,
                       int cls, int gender, const uint16_t *name, int name_len);

/* 從名冊除名(若在隊伍中則一併移出)。回 0;index 不合法回 -1。 */
int  dq3_roster_remove(dq3_roster *r, dq3_party *p, int idx);

void dq3_party_init(dq3_party *p);

/* 把名冊 index 編入隊伍(滿 / 已在隊伍 / index 不合法 → -1)。回隊伍 slot。 */
int  dq3_party_add(dq3_party *p, dq3_roster *r, int roster_idx);

/* 把隊伍 slot 移出(回名冊)。回 0;slot 不合法 / 空回 -1。 */
int  dq3_party_remove(dq3_party *p, dq3_roster *r, int slot);

#endif /* DQ3_ROSTER_H */
