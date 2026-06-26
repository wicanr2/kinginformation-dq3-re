/* dq3_scripted.c — scripted NPC 給物/對話表(Step 2)。資料源 docs/data/sub2-handlers.md
 * (gen_sub2_full)+ 逐 handler 反組譯(docs/re-log-722)。recs = di−0xbb8。 */
#include "dq3_scripted.h"
#include "dq3_inventory.h"   /* 道具 id */
#include "dq3_progress.h"    /* 里程碑 */

/* {byte4, cty, give_item, prereq_item, require_item, consume_prereq, milestone, before_rec, give_rec, after_rec} */
static const dq3_scripted TABLE[] = {
  /* ── 給予型(give_item != NONE)── */
  /* 拿吉米之塔 4F 老人:給盜賊鑰匙(handler L0x537f)。無前置;before 不觸發。 */
  {  12,  8, 0x55, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0, DQ3_MS_THIEF_KEY, 54, 52, 53 },
  /* 雷貝 2F 老人:持盜賊鑰匙(開門,保留)→ 給魔法玉(handler L0x521f)。 */
  {   7,  1, 0x58, 0x55,          DQ3_SC_NOITEM, 0, DQ3_MS_MAGIC_BALL, 65, 60, 60 },
  /* CTY22 sect1 (3,4):給水槍 0x4b(handler L0x589f,flag 0x3c)。rec45 招呼 → 給 0x4b + rec46。 */
  {  31, 22, 0x4b, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0, 0,                 45, 46, 45 },
  /* 龍女王城 CTY67 sect0 (14,24) 龍之女王:給光之珠 0x65(handler L0x5e02)。對抗索瑪關鍵道具。 */
  {  52, 67, 0x65, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0, 0,                 71, 71, 71 },
  /* CTY16 sect0 (33,24) 獲救劇情 NPC(L0x6646):與卡爾洛斯重逢答謝 → 給誘惑之劍 0x10(女性專用)。 */
  {  84, 16, 0x10, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0, 0,                 98, 99, 98 },
  /* 巴哈拉達 CTY15 sect0 (5,24) 胡椒商古布達(L0x5719):獲救後給黑胡椒 0x5c(取船材料)。 */
  {  25, 15, 0x5c, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0, 0,                 119, 118, 120 },
  /* CTY64 sect0 (16,10)(L0x5d70):「粉碎大魔王陰謀」→ 給銀寶珠 0x6b。 */
  {  49, 64, 0x6b, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0, 0,                 85, 84, 85 },

  /* ── transform(消耗 prereq 換 give_item)── */
  /* 精靈之村 CTY5 sect0 (17,7) 精靈女王(L0x5567 je 分支):持夢幻紅寶石 0x59 →
   * 消耗換覺醒粉 0x5a(原版 `mov [si],0x5a` 就地替換)+ rec96;無紅寶石 → rec90(杜勝利 Ch10)。 */
  {  16,  5, 0x5a, 0x59,          DQ3_SC_NOITEM, 1, 0,                 90, 96, 90 },

  /* 雪地草原 CTY54 sect0 (8,2) 老人(L0x5c20 je 分支,flag 0x43):持變身杖 0x62 →
   * 消耗換船員之骨 0x63(原版 `mov[si],0x63` 就地替換)+ rec59(杜勝利 Ch37)。無變身杖 → rec56。 */
  {  44, 54, 0x63, 0x62,          DQ3_SC_NOITEM, 1, 0,                 56, 59, 60 },

  /* ── 檢查型(give_item==NONE,require_item != NONE;持物→success rec、否則→need rec,不消耗)── */
  /* CTY62/63 sect0(L0x5daf,flag 0x4c):需持船 0x5b → 反應 rec87,否則 rec89(渡海 gate)。
   * cty=0xff:byte4=50 在 CTY62(×2)/CTY63 皆同義(全 CTY 僅此處用 b4=50)。 */
  {  50, 0xff, DQ3_SC_NOITEM, DQ3_SC_NOITEM, 0x5b, 0, 0,               89, 87, 87 },
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
