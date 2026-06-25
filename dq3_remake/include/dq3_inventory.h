/* dq3_inventory.h — 道具欄 + 合成事件(含 #2 彩虹水滴修正)。
 *
 * 反組譯合成事件(file 0x77ce,docs/18/22):神聖祠堂「雨和太陽合而為一」劇情 ——
 * 消耗 太陽之石(0x72)+ 雲雨之杖(0x73)→ 在雲雨之杖格寫入合成品,設「已合成」旗標 0x139。
 * 地點(青衫攻略):下層世界**利姆達爾鎮往南航行的小島神聖祠堂**(非拉達多姆;拉達多姆只是
 * 太陽之石產地)。詳見 docs/32 世界地點圖。
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

/* 鑰匙(鎖門 docs/35 §八;等級 = code − 0x54,愈高開愈多門)。 */
#define ITEM_KEY_THIEF     0x55   /* 盜賊的鑰匙(等級 1)*/
#define ITEM_KEY_MAGIC     0x56   /* 魔法鑰匙(等級 2)*/
#define ITEM_KEY_FINAL     0x57   /* 最終鑰匙(等級 3)*/

typedef struct { uint8_t slot[DQ3_INV_SLOTS]; } dq3_inventory;

/* 劇情旗標位元集(對 DQ3.EXE flag store;set_flag/sub_8264)。0x139=「彩虹水滴已合成」。 */
#define DQ3_FLAG_RAINBOW_SYNTHED 0x139
#define DQ3_FLAG_DESCENDED       0x13a     /* 已下降至下層世界(アレフガルド;對 EXE [0x5051] 語意)*/
#define DQ3_FLAGS_BYTES 64                 /* 覆蓋旗標 id 0..511 */
typedef struct { uint8_t bit[DQ3_FLAGS_BYTES]; } dq3_storyflags;

void dq3_flags_init(dq3_storyflags *f);
int  dq3_flags_get(const dq3_storyflags *f, int id);
void dq3_flags_set(dq3_storyflags *f, int id, int v);

void dq3_inv_init(dq3_inventory *inv);                 /* 全清空(0xff) */
int  dq3_inv_find(const dq3_inventory *inv, int item); /* 回 slot index,無回 -1 */
int  dq3_inv_has(const dq3_inventory *inv, int item);
int  dq3_inv_add(dq3_inventory *inv, int item);        /* 放第一個空格;滿回 -1 */
int  dq3_inv_remove(dq3_inventory *inv, int item);     /* 移除一個;回 slot,無回 -1 */

/* 持有的最高鑰匙等級(0=無、1=盜賊、2=魔法、3=最終)。
 * 鏡射 DQ3.EXE 0x48c3:掃道具格,item∈[0x55,0x57] → tier=item−0x54,取最大。
 * 註:EXE 逐隊員各 8 格;remake 用單一道具欄,語意等同「隊伍是否持某級鑰匙」。 */
int  dq3_inv_key_tier(const dq3_inventory *inv);

/* #2 合成事件:太陽之石 + 雲雨之杖 → 彩虹水滴(消耗兩者,在雲雨之杖格寫入成品)。
 * fixed!=0:產出 0x75(彩虹水滴,修正);fixed==0:重現原版 bug 產出 0x6b(銀寶珠)。
 * 回傳產出的 item code;材料不足回 -1。 */
int  dq3_synth_rainbow_drop(dq3_inventory *inv, int fixed);

/* #2 祠堂合成事件(in-game 觸發,對齊 file 0x77ce 的條件與副作用)。
 * 在祠堂「調べる」時呼叫:已合成過(flag 0x139)→ 回 -2 不重複;持有兩材料 → 合成
 * (fixed 同上)+ 設旗標 0x139,回成品 code;材料不足 → 回 -1。 */
int  dq3_synth_shrine_examine(dq3_inventory *inv, dq3_storyflags *flags, int fixed);

/* scripted-event 派發(鏡射 DQ3.EXE runner file 0xabb2:event id 在 [0x722],
 * `dec bx; shl bx,1; call [bx+0x3baa]` → handler 跳表)。原版 event id 為資料驅動
 * (地圖/劇本 → [0x631] → [0x722]);彩虹水滴合成 = scripted-event 83。
 * 目前僅實作 id 0x53;其餘回 -3(未實作)。 */
#define DQ3_SEVENT_RAINBOW_SYNTH 0x53   /* 祠堂「雨和太陽合而為一」合成事件 id(handler 0x776c)*/
#define DQ3_SEVENT_DESCENT       0x56   /* 下降至下層世界(ギアガの大穴;handler 0x783d,docs/44)。
                                         * 場景效果(切下層 overworld + 設旗標)在 main.c do_descent。*/
#define DQ3_SHRINE_CTY           93     /* 合成祠堂 = CTY93(下層,利姆達爾CTY86南方;cty_loc+alefgard 比對,docs/32)*/
/* 祠堂祭司 NPC 座標(CTY93 section0 NPC 清單唯一一隻,17×17;靜態解,三證據:CTY 資料 +
 * 劇本 txt08「神聖的祠堂」+ 地理 docs/32)。調べる此格 → 合成 scripted-event 83。 */
#define DQ3_SHRINE_NPC_X         8
#define DQ3_SHRINE_NPC_Y         8

/* 露依達酒場(登記隊員)= 阿里阿罕 CTY00 sec0 西側 LUIDA 建築的櫃台店員。
 * 定位來源(腳本+地圖 metadata,非人工猜):D3TXT01 rec49「鎮上西方有路依達酒吧」+
 * sec0 轉場表 門(8,14)→sec2(2F 預存所)+ sec0 NPC 表(西側 (8,17) 櫃台店員)。docs/36。 */
#define DQ3_LUIDA_CTY            0
#define DQ3_LUIDA_X              8
#define DQ3_LUIDA_Y              17
int  dq3_scripted_event_run(int event_id, dq3_inventory *inv, dq3_storyflags *flags, int fixed);

#endif /* DQ3_INVENTORY_H */
