/* dq3_text.h — 遊戲文字/對話渲染(D3TXT*.TXT + D3TXT00.FON)。
 *
 * 容器(docs/03):檔首 16-bit LE 指標表(ptr[0]=表大小);記錄 i = data[ptr[i]..ptr[i+1]],
 * 為一串 2-byte LE 值,0xffff 終止。值 <1476 = D3TXT00.FON 第 N 個 16×16 字模;
 * >=0xffed = 控制碼(0xfffe 換行、0xfffc 換頁、0xffed/0xfff5/0xfff6/0xfff9/0xfffa/0xfffb
 * = 動態插值占位 → 渲染為空白)。字型固定 D3TXT00.FON(不經 Big5/CHINA.FON)。
 */
#ifndef DQ3_TEXT_H
#define DQ3_TEXT_H

#include <stdint.h>
#include <stddef.h>

#define DQ3_GLYPH_PX  16
#define DQ3_TXT_NL    0xfffe
#define DQ3_TXT_NL2   0xfffd
#define DQ3_TXT_PAGE  0xfffc
#define DQ3_TXT_END   0xffff

/* 動態插值控制碼(docs/12/31;每個後接 +1 word 參數,須一併消耗)。
 * 渲染時依目前變數 context 替換為實字(主角名 / 數值 / 道具名);未設則略過。 */
#define DQ3_TXT_VAR_ENT  0xfffb   /* 實體/隊員名(依 [0x259c])*/
#define DQ3_TXT_VAR_ITEM 0xfffa   /* 道具名 */
#define DQ3_TXT_VAR_NUM  0xfff9   /* 數值(金額等,依 [0x2593])*/
#define DQ3_TXT_VAR7     0xfff6   /* 變數子字串 variant 7 */
#define DQ3_TXT_VAR0     0xfff5   /* 變數子字串 variant 0(城鎮對話最常見=主角名)*/
#define DQ3_TXT_VAR_IDX  0xffed   /* 索引字串(依 [0x249d]/[0x249f])*/

typedef struct {
    uint8_t *fon; size_t fon_len;     /* D3TXT00.FON(owned) */
    uint8_t *txt; size_t txt_len;     /* D3TXTNN.TXT(owned) */
    int n_records;
    int loaded;
} dq3_text;

/* 載入字型(固定 D3TXT00.FON)+ 指定 txt_name(如 "D3TXT01.TXT")。失敗回 <0。 */
int  dq3_text_load(dq3_text *t, const char *assets_dir, const char *txt_name, char *err, int errcap);
/* 只換文字檔(保留已載入的字型);供切換 D3TXT bank(對話分檔)。須先 load 過。失敗回 <0。 */
int  dq3_text_reload(dq3_text *t, const char *assets_dir, const char *txt_name, char *err, int errcap);
void dq3_text_free(dq3_text *t);

/* 取記錄 rec 的 2-byte 值序列(不含 0xffff 終止)到 out(最多 max);回傳值數,<0 失敗。 */
int  dq3_text_record(const dq3_text *t, int rec, uint16_t *out, int max);

/* 把記錄 rec 畫進 framebuffer 的文字視窗(retro-cjk:每字 16×16)。
 *   (x0,y0)=視窗內文起點;cols=每行最多字數;maxlines=最多行(0xfffc 換頁時停)。
 *   fg=字色 palette 索引。回傳畫出的行數。控制碼:換行/換頁/變數占位空白。 */
int  dq3_text_draw_record(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                          int x0, int y0, int cols, int maxlines, int rec, uint8_t fg);

/* 畫一個字模(idx)到 (x,y),set bit → fg。供其他模組(選單/HUD)用。 */
void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg);

/* ── 對話變數 {V} 替換 context(module-level;draw_record 渲染時取用)──
 * 渲染遇到插值控制碼時,消耗其 +1 參數並插入對應實字。未設者插值處留空(但仍正確消耗參數,
 * 不會把參數誤畫成字模)。name = glyph 索引序列(姓名輸入產出),供 VAR0/VAR_ENT/VAR7/VAR_IDX。 */
void dq3_text_set_var_name(const uint16_t *name_glyphs, int len);  /* 主角/受話者名 */
void dq3_text_set_var_num(long n);                                 /* VAR_NUM 數值 */
void dq3_text_clear_vars(void);

#endif /* DQ3_TEXT_H */
