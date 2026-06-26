/* dq3_scripted.c — scripted NPC 給物/對話表(Step 2)。資料源 docs/data/sub2-handlers.md
 * (gen_sub2_full)+ 逐 handler 反組譯(docs/re-log-722)。recs = di−0xbb8。 */
#include "dq3_scripted.h"
#include "dq3_inventory.h"   /* 道具 id */
#include "dq3_progress.h"    /* 里程碑 */

/* {byte4, cty, give_item, prereq_item, milestone, before_rec, give_rec, after_rec} */
static const dq3_scripted TABLE[] = {
  /* 拿吉米之塔 4F 老人:給盜賊鑰匙(handler L0x537f)。無前置;before 不觸發。 */
  {  12,  8, 0x55, DQ3_SC_NOITEM, DQ3_MS_THIEF_KEY, 54, 52, 53 },
  /* 雷貝 2F 老人:持盜賊鑰匙(開門)→ 給魔法玉(handler L0x521f)。 */
  {   7,  1, 0x58, 0x55,          DQ3_MS_MAGIC_BALL, 65, 60, 60 },
  /* CTY22 sect1 (3,4) sub2 NPC:給水槍 0x4b(handler L0x589f,flag 0x3c)。
   * rec45 招呼(已給/未給都先顯示)→ test_flag(0x3c) → 給 0x4b + rec46。 */
  {  31, 22, 0x4b, DQ3_SC_NOITEM, 0,                 45, 46, 45 },
  /* 龍女王城 CTY67 sect0 (14,24) 龍之女王:給光之珠 0x65(handler L0x5e02,sflag 0x4e)。
   * 對抗索瑪關鍵道具;無 tflag 條件(無前置)。rec71 統一(招呼=給予=後話)。
   * (sweep_sub2_npcs 權威清單確認 b4=52 為真 sub2 GIVE NPC;早期 count-prefix 解析 bug 曾誤判不存在。)*/
  {  52, 67, 0x65, DQ3_SC_NOITEM, 0,                 71, 71, 71 },
  /* CTY16 sect0 (33,24) 獲救劇情 NPC(handler L0x6646,flag 0xe5):與心愛的卡爾洛斯重逢答謝,
   * 給誘惑之劍 0x10(女性專用武器)。rec98 招呼 / rec99 給予(D3TXT04 bank)。flag-gate 觀感給予近似。
   * (byte4=35 give 0x66 handler 存在但無任何 NPC 引用 → 開發殘留,不接。)*/
  {  84, 16, 0x10, DQ3_SC_NOITEM, 0,                 98, 99, 98 },
  /* 巴哈拉達 CTY15 sect0 (5,24) 胡椒商古布達(handler L0x5719,flag 0x36):獲救後給黑胡椒 0x5c
   * (取船劇情材料)。rec118 給予/rec119 招呼/rec120 後話(D3TXT03 bank)。
   * (註:CTY15 道具店亦補進黑胡椒[docs/50 remake 補洞];此 NPC 為原版忠實取得管道。)*/
  {  25, 15, 0x5c, DQ3_SC_NOITEM, 0,                 119, 118, 120 },
  /* CTY64 sect0 (16,10) NPC(handler L0x5d70,flag 0x4b):「你們一定能粉碎大魔王陰謀」→
   * 給銀寶珠 0x6b(rec84 給予/rec85 招呼,D3TXT01 bank)。*/
  {  49, 64, 0x6b, DQ3_SC_NOITEM, 0,                 85, 84, 85 },
};
static const int N = (int)(sizeof TABLE / sizeof TABLE[0]);

const dq3_scripted *dq3_scripted_get(int byte4, int cty)
{
    int i;
    for (i = 0; i < N; i++)
        if (TABLE[i].byte4 == byte4 && (TABLE[i].cty == 0xff || TABLE[i].cty == cty))
            return &TABLE[i];
    return 0;
}
