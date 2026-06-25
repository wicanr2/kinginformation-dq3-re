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
