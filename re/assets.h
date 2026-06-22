/* assets.h — DQ3.EXE 素材載入層的型別與常數
 *
 * 由 DQ3.EXE(16-bit real-mode large model)反編譯。所有「DS:offset」對應
 * 主資料段 DS=0x14dd(file base 0x16140);此處抽象成具名 C 結構與全域。
 *
 * 座標換算:seg0 程式碼 file off = 0x1370(HDR) + seg0_off。
 *           主資料段欄位 file off = 0x16140 + ds_off。
 *
 * 素材載入通用樣式(real-mode DOS 呼叫):
 *   AH=3Dh 開檔(DS:DX=檔名, AL=0 唯讀) → BX=handle
 *   AH=3Fh 讀(BX=handle, CX=bytes, DS:DX=buffer;CX=0xffff 表「讀到 EOF」)
 *   AH=3Eh 關
 * 帶序號檔名(cty00 / dq3?.blk / blkbmN / d3txtNN / dragonN)是「模板」,
 * 載入時就地改寫檔名字串裡的數字位元組產生實際檔名。
 */
#ifndef DQ3_ASSETS_H
#define DQ3_ASSETS_H

#include <stdint.h>

/* ---- 檔名字串池(DS:0x003d 起,23 筆,8.3 FCB)---- */
/* 帶序號者就地改數字(見各 loader 註解)。pool offset 對照 docs/09。 */
extern char fn_setup[];    /* DS:0x3d  "setup.dat"   */
extern char fn_player[];   /* DS:0x47  "player.dat"  */
extern char fn_dragon[];   /* DS:0x52  "dragon0.dat" 模板,0x58 改數字 */
extern char fn_cty[];      /* DS:0x5e  "cty00.dat"   模板,0x61/0x62 改數字 */
extern char fn_d3txt_fon[];/* DS:0x68  "d3txt00.fon" */
extern char fn_d3txt_txt[];/* DS:0x74  "d3txt00.txt" 模板,0x79/0x7a 改數字 */
extern char fn_item[];     /* DS:0x80  "item.dat"    */
extern char fn_con_map[];  /* DS:0x89  "dq3con.map"  */
extern char fn_und_map[];  /* DS:0x94  "dq3und.map"  */
extern char fn_blk_town[]; /* DS:0x9f  "dq31.blk"    模板,0xa2 改數字 */
extern char fn_blk_over[]; /* DS:0xa8  "dq3.blk"     */
extern char fn_blkbm1[];   /* DS:0xb0  "blkbm1.dat"  模板,0xb5 改數字 */
extern char fn_blkbm[];    /* DS:0xbb  "blkbm.dat"   */
extern char fn_bls_mst[];  /* DS:0xc5  "dq3mst.bls"  */
extern char fn_bls_man[];  /* DS:0xd0  "dq3man.bls"  */
extern char fn_bls_lin[];  /* DS:0xdb  "dq3lin.bls"  */
extern char fn_pal[];      /* DS:0xe6  "dq3.pal"     */
extern char fn_mnsbk_pal[];/* DS:0xee  "mnsbk.pal"   */
extern char fn_d3mns[];    /* DS:0xf8  "d3mns.dat"   */
extern char fn_mns_shp[];  /* DS:0x102 "dq3mns.shp"  */
extern char fn_packbg[];   /* DS:0x10d "packbg.scr"  */
extern char fn_mbg[];      /* DS:0x118 "mbg.mcx"     */
extern char fn_ebg[];      /* DS:0x120 "ebg.mcx"     */

/* ---- BLK tile 圖庫 6-byte 表頭(docs/04)---- */
/* 程式碼端固定讀 CX=6 印證表頭為 6 bytes。 */
typedef struct {
    uint16_t row_bytes;  /* DS:0x2859  每 plane 每 row byte 數 = 4 → 寬 32px */
    uint16_t height;     /* DS:0x285b  高 = 24 row */
    uint16_t count;      /* DS:0x285d  tile 數量 */
} blk_header_t;          /* tile body 每格 = row_bytes*8 * height * 4bit/8 = 384B */

/* ---- 遠堆積段控制代碼(runtime 分配的素材緩衝段)---- */
/* 原始為 DS:offset 存的段值;C 端以「素材大緩衝」抽象。 */
extern uint16_t seg_fon;     /* DS:0x252c  d3txt00.fon 字模緩衝段 */
extern uint16_t seg_txt;     /* DS:0x252e  d3txt00.txt / d3txtNN.txt 緩衝段 */
extern uint16_t seg_blk;     /* DS:0x2532  目前 BLK tile body 緩衝段 */
extern uint16_t seg_blspage; /* DS:0x2534  .bls 分頁讀取目的段 */
extern uint16_t seg_cty;     /* DS:0x2536  cty NN.dat 緩衝段 */
extern uint16_t seg_packbg;  /* DS:0x2542  packbg.scr 緩衝段 */

/* ---- 載入狀態 ---- */
extern uint16_t cur_map_idx;  /* DS:0x256c  目前地圖/城鎮索引(查 [bx+0xa04] 得檔號)*/
extern uint16_t is_town_flag; /* DS:0x4f2d  0=地表 1/2/3=城鎮分類(決定 dq3.blk vs dq3?.blk)*/

/* 快取的常駐檔案 handle(.bls / .map 不一次讀入,保持開啟分頁 seek 讀)*/
extern uint16_t h_bls_mst;    /* DS:0x260b */
extern uint16_t h_bls_man;    /* DS:0x2609 */
extern uint16_t h_con_map;    /* DS:0x1f1  dq3con.map handle */
extern uint16_t h_und_map;    /* DS:0x1f3  dq3und.map handle */
extern uint16_t h_packbg;     /* DS:0x1f5  packbg.scr handle */
extern uint16_t h_d3mns;      /* DS:0x249b dq3mns.shp 共用 handle(?) */

/* ---- 各 loader(對應 DQ3.EXE seg0 位址,見 docs/09)---- */
int  load_setup(void);          /* seg0:9981  setup.dat → 6B 表頭旗標 */
int  load_player(void);         /* seg0:177c  player.dat → save 200B(無檔則清零=新遊戲)*/
void save_player(void);         /* seg0:17ce  player.dat 建檔寫 200B */
int  load_dragon(void);         /* seg0:14be  dragonN.dat → DS:0x4f29 save 緩衝 */
int  load_cty(uint16_t target_tile); /* seg0:3047  ctyNN.dat → seg_cty,再呼叫 load_blk */
int  load_d3txt_font(void);     /* seg0:171b  d3txt00.fon → seg_fon */
int  load_d3txt_text(void);     /* seg0:179d  d3txt00.txt → seg_txt */
int  load_d3txt_map(uint8_t n); /* seg0:182c  d3txtNN.txt → seg_txt:0x4e20 */
int  load_item(void);           /* seg0:17cb  item.dat → DS:0x1f9 */
void open_world_maps(void);     /* seg0:2c19  開 dq3con.map + dq3und.map(只快取 handle)*/
int  load_blk(void);            /* seg0:eaf4  dq3.blk 或 dq3?.blk → seg_blk;同時載 blkbmN */
int  load_blkbm(void);          /* seg0:ecc0  blkbm.dat → DS:0x2eea */
int  load_bls_banks(void);      /* seg0:ec9e  開 dq3mst.bls + dq3man.bls(快取 handle)*/
int  load_palettes(void);       /* seg0:ecdc  dq3.pal(240B)+ mnsbk.pal(480B)*/
int  load_mns_shp(void);        /* seg0:b19e  dq3mns.shp → 計數表 */
void load_packbg_page(uint16_t page); /* seg0:c6e5  packbg.scr 分頁 planar 載入 */

#endif /* DQ3_ASSETS_H */
