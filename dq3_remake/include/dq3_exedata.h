/* dq3_exedata.h — 自 DQ3.EXE DGROUP 抽出的靜態資料表(生成檔,勿手改)。
 * 由 tools/extract_dgroup_tables.py 產生;remake 改用內建表,不再讀 DQ3.EXE。
 * 目錄與語意見 docs/data/dgroup-tables.md。 */
#ifndef DQ3_EXEDATA_H
#define DQ3_EXEDATA_H
#include <stdint.h>

#define DQ3X_NUM_CLASS 8
#define DQ3X_GROWTH_ROW 14
#define DQ3X_MAX_LEVEL 43
extern const uint8_t  dq3x_growth[DQ3X_NUM_CLASS][DQ3X_GROWTH_ROW];
extern const uint32_t dq3x_thresh[DQ3X_NUM_CLASS][DQ3X_MAX_LEVEL+1];
extern const uint8_t  dq3x_terrain[256];   /* tile→地形類別 */
extern const uint8_t  dq3x_event[256][3];  /* tile→怪物/戰鬥事件(3B,docs/31)*/
#define DQ3X_WARP_N 128
extern const uint8_t  dq3x_warp[DQ3X_WARP_N][3];  /* 傳送目的地:地圖+座標 */
#define DQ3X_CTYLOC_N 100
extern const uint8_t  dq3x_cty_loc[DQ3X_CTYLOC_N][2];  /* 地表(X,Y)→CTY號(index) */
#endif
