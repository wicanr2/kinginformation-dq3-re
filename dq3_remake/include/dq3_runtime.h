/* dq3_runtime.h — DOS 硬體 runtime → SDL2 shim(adapters at the edges)
 *
 * 原版 DQ3.EXE 把硬體存取集中在一套手寫組語 runtime(far-call,見 docs/06):
 *   VGA planar 繪圖(0x111b)、Sound Blaster(0x129c)、鍵盤(0x1104)、
 *   滑鼠(0x10b6)、檔案 I/O(0x1053)、BIOS 視訊(0x109c)、tick 計時。
 * 現代跨平台 remake 不模擬 real-mode;改在「邊界」用 SDL2 / stdio 重寫等價語意。
 * 遊戲邏輯(由 re/*.c 反編譯結果移植來的 C)只呼叫這層窄介面,不再 far-call。
 *
 * 設計原則(deep module + adapters at edges,見 rules/70):
 *   - 對外介面窄:framebuffer(indexed)+ palette + 輸入事件 + 計時 + 檔案。
 *   - 內部複雜度(planar↔packed、scancode 對應、nearest 放大、headless dump)藏實作。
 *
 * 顯示模型對應原版 VGA:
 *   原版用 mode 0x10(640×350 16 色 EGA planar)畫標題;遊戲畫面亦在此族解析。
 *   shim 用單張 indexed framebuffer(每 byte 一像素 = palette 索引),
 *   present 時套 palette → ARGB8888 texture → nearest 放大到視窗
 *   (retro-cjk-hires-canvas:底圖整數放大保持銳利,中文字日後用 16×16 點陣畫在此畫布)。
 */
#ifndef DQ3_RUNTIME_H
#define DQ3_RUNTIME_H

#include <stdint.h>
#include <stddef.h>

/* 邏輯畫布 = 原版 VGA 解析(標題 640×350)。 */
#define DQ3_SCREEN_W 640
#define DQ3_SCREEN_H 350

/* 內部畫布放大倍率(nearest)。視窗 = 邏輯畫布 × SCALE。 */
#ifndef DQ3_SCALE
#define DQ3_SCALE 2
#endif

/* RGB palette 條目(原版 6-bit VGA DAC,載入時擴成 8-bit)。 */
typedef struct { uint8_t r, g, b; } dq3_color;

/* ---- runtime 生命週期 ---- */
int  dq3_rt_init(const char *window_title);   /* 建 SDL window/renderer/texture */
void dq3_rt_quit(void);

/* ---- 顯示:indexed framebuffer ---- */
uint8_t *dq3_fb(void);                          /* 邏輯畫布像素緩衝(W*H,每 byte=索引) */
void     dq3_set_palette(const dq3_color *pal, int count);
void     dq3_present(void);                     /* indexed→RGB,nearest 放大,翻頁 */
/* 把目前 framebuffer + palette 直接寫成 PPM(headless 驗證,不需視窗)。 */
int      dq3_dump_ppm(const char *path);

/* ---- 輸入:DOS BIOS scancode 語意,來源是 SDL 事件 ---- */
/* 對應原版主迴圈讀鍵介面(docs/08 kbd_read_scancode):0x48/0x50/0x4b/0x4d 方向、
 * 0x1c Enter、0x39 Space、0x01 ESC…。回傳 0=本幀無新按鍵。 */
uint8_t dq3_poll_scancode(void);
int     dq3_should_quit(void);                  /* 視窗關閉 / 要求離開 */
void    dq3_rt_set_quit(void);                  /* 遊戲層確認離開後呼叫(F10 → Yes 流程)*/
#define DQ3_SC_F1   0x3b                          /* F1 = 按鍵說明視窗 */
#define DQ3_SC_F10  0x44                         /* F10 = 離開請求 scancode(遊戲層 Yes/No + autosave)*/
#define DQ3_SC_F5   0x3f                          /* F5 = 隨時存檔(選 slot)*/
#define DQ3_SC_F9   0x43                          /* F9 = 隨時讀檔(選 slot,回原位置)*/

/* ---- 計時:取代 BIOS tick / frame_wait(sub_ee23)---- */
void     dq3_delay_ms(uint32_t ms);
uint32_t dq3_ticks_ms(void);

/* ---- 檔案 I/O:取代 seg 0x1053 的 DOS AH=3Dh/3Fh/3Eh,改 stdio ---- */
/* 從 assets 目錄讀整檔到 malloc buffer;*len 回傳長度,失敗回 NULL(呼叫端 free)。 */
uint8_t *dq3_load_file(const char *name, size_t *len);
void     dq3_set_assets_dir(const char *dir);

#endif /* DQ3_RUNTIME_H */
