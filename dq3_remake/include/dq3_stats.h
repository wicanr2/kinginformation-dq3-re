/* dq3_stats.h — 角色數值/升級系統(成長表 + 升級門檻),含 #4/#5/#6 修正。
 *
 * 資料來源:使用者自己的 DQ3.EXE 靜態映像(不把版權數值入 git),file offset 見
 * docs/23 / tools/re_stat_patch.py:
 *   成長表  file 0x1a4a6(DS:0x4366),8 職業 × 14 byte/列。
 *   門檻表  指標表 file 0x1a516(DS:0x43d6)→ 8 職業各 44 entry × u32 累積經驗。
 *
 * 公式(忠實對齊 sub_d9cc):target = base + (slope×level)/2,全程 16-bit。
 * MP/各屬性實得增量 = rng % (target − current),下限 1(此處只算確定性的 target 上限)。
 *
 * 三個數值 bug 的修正(青衫無官方碼,binary patch 受限,C 層為單行/型別修正):
 *   #4 勇者 MaxMP 成長偏低:勇者列 MP base 3→8、slope 5→10(拉大 target/delta)。
 *   #5 高等級(Lv44+)升級錯亂:查門檻前 clamp level≤43,杜絕越界讀相鄰職業 entry[0]=0
 *      造成的連升。binary patch 因 code cave 不足(≤8B)做不到,C 層為單行 if。
 *   #6 屬性 255 溢位:屬性一律 uint16 + 寫回前明確 clamp,不靠自然 wrap。
 */
#ifndef DQ3_STATS_H
#define DQ3_STATS_H

#include <stdint.h>

#define DQ3_NUM_CLASS   8
#define DQ3_MAX_LEVEL   43     /* 門檻表 lv0..43,共 44 entry */
#define DQ3_STAT_CAP    9999   /* #6:屬性顯式上限(取代自然 wrap) */

/* 職業:0勇者 1戰士 2武鬥家 3僧侶 4魔法使者 5賢者 6商人 7遊玩者 */
/* 成長屬性欄(列內 base/slope 對):0 HP, 1 MP, 2 b4, 3 b6, 4 b8, 5 bA */
typedef enum {
    DQ3_STAT_HP = 0, DQ3_STAT_MP = 1, DQ3_STAT_B4 = 2,
    DQ3_STAT_B6 = 3, DQ3_STAT_B8 = 4, DQ3_STAT_BA = 5
} dq3_stat_kind;

typedef struct {
    uint8_t  growth[DQ3_NUM_CLASS][14];                 /* 成長表原始列 */
    uint32_t thresh[DQ3_NUM_CLASS][DQ3_MAX_LEVEL + 1];  /* 升級門檻(累積經驗) */
    int      loaded;
    int      apply_bug4_fix;                            /* 勇者 MP 成長修正開關 */
} dq3_stats;

/* 從 DQ3.EXE 讀成長表 + 門檻表。apply_bug4_fix!=0 時套用 #4 勇者 MP 修正。
 * 失敗回 <0(err 可帶訊息)。 */
int dq3_stats_load(dq3_stats *st, const char *assets_dir, int apply_bug4_fix,
                   char *err, int errcap);

/* 成長目標值(確定性上限):target = base + (slope×level)/2,16-bit。 */
uint16_t dq3_stats_growth_target(const dq3_stats *st, int cls, dq3_stat_kind kind, int level);

/* 由累積經驗求等級。
 *   fixed!=0:#5 修正版 — level 夾在 [1,43],不越界。
 *   fixed==0:重現原版 bug — 不 clamp,越界讀相鄰職業門檻 → 連升(以 cap 防無限迴圈,
 *            回傳值 > 43 即代表發生 runaway)。 */
int dq3_stats_level_for_exp(const dq3_stats *st, int cls, uint32_t exp, int fixed);

/* #6:屬性安全累加(uint16 + 顯式 clamp 到 DQ3_STAT_CAP,不 wrap)。 */
uint16_t dq3_stats_add_clamped(uint16_t cur, int delta);

#endif /* DQ3_STATS_H */
