/* dq3_stats.h — 角色數值/升級系統(成長表 + 升級門檻),含 #4/#5/#6 修正。
 *
 * 資料來源:內建 dq3_exedata(自 DQ3.EXE DGROUP 一次抽出編入,生成檔)。remake 不依賴 DQ3.EXE;
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
#define DQ3_STAT_CAP    9999   /* #6:寬泛上限(dq3_stats_add_clamped 通用)*/
#define DQ3_STAT_PRIMARY_CAP 255   /* #6:力量/速度/耐力/聰明/運氣 byte 上限(原版超 255 wrap 歸 0)*/
#define DQ3_STAT_HPMP_CAP    999   /* #6:HP/MP 3 位數上限 */

/* 職業:0勇者 1戰士 2武鬥家 3僧侶 4魔法使者 5賢者 6商人 7遊玩者 */
/* 成長屬性欄 = 成長表 14 byte 的 7 個 (base,slope) 對,kind*2 = 列內 byte offset。
 * 屬性語意由三方交叉確認:成長表數值樣式(武鬥力量/速度最高、戰士聰明=0、遊玩運氣最高)
 * + 升級訊息(D3TXT00 rec191-197:力量/耐力/速度/最大HP/最大MP/聰明度/運氣)
 * + BBS 存檔記憶體佈局(docs/history/dq3-bbs-1994 §六)。 */
typedef enum {
    DQ3_STAT_HP   = 0,   /* 生命 */
    DQ3_STAT_MP   = 1,   /* 魔力 */
    DQ3_STAT_AGI  = 2,   /* 速度 すばやさ(+4)*/
    DQ3_STAT_STR  = 3,   /* 力量 ちから(+6)*/
    DQ3_STAT_INT  = 4,   /* 聰明 かしこさ(+8)*/
    DQ3_STAT_VIT  = 5,   /* 耐力 たいりょく(+A)*/
    DQ3_STAT_LUCK = 6    /* 運氣 うんのよさ(+C)*/
} dq3_stat_kind;
#define DQ3_STAT_COUNT 7

typedef struct {
    uint8_t  growth[DQ3_NUM_CLASS][14];                 /* 成長表原始列 */
    uint32_t thresh[DQ3_NUM_CLASS][DQ3_MAX_LEVEL + 1];  /* 升級門檻(累積經驗) */
    int      loaded;
    int      apply_bug4_fix;                            /* 勇者 MP 成長修正開關 */
} dq3_stats;

/* 載入成長表 + 門檻表(自內建 dq3_exedata)。apply_bug4_fix!=0 時套用 #4 勇者 MP 修正。
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

/* #6:依屬性別的上限(主屬性 255、HP/MP 999)。member 升級用此 clamp,杜絕 255 wrap。 */
int dq3_stat_cap_for(dq3_stat_kind k);

/* ---- 隊員實體:把上述零件整合成可玩的升級系統(#4/#5/#6 在此真生效)---- */
typedef struct {
    int      cls;          /* 職業 0..7 */
    int      level;        /* 1..43 */
    uint32_t exp;          /* 累積經驗(uint32,不溢位)*/
    uint16_t stat[DQ3_STAT_COUNT];  /* 7 屬性 uint16,順序同 dq3_stat_kind(HP MP 速 力 智 耐 運)= 上限值 */
    uint16_t cur_hp, cur_mp;        /* 目前 HP/MP(持久:戰鬥傷害保留、旅社/教會回復;cur_hp==0=陣亡)*/
} dq3_member;

/* 以職業+等級初始化隊員:exp=該級門檻、各屬性=成長目標值(growth_target)。 */
void dq3_member_init(dq3_member *m, const dq3_stats *st, int cls, int level);

/* 獲得經驗 → 跨門檻則升級,逐級套成長 delta(以 add_clamped #6 寫回,等級用 #5 修正版)。
 * 回傳本次升的等級數(0=未升)。#4 勇者 MP 由 st 載入時的成長表修正自動生效。 */
int  dq3_member_gain_exp(dq3_member *m, const dq3_stats *st, uint32_t add);

#endif /* DQ3_STATS_H */
