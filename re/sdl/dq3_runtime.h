/* dq3_runtime.h — DOS runtime -> SDL2 shim 介面(adapters at the edges)
 *
 * 原版 DQ3.EXE 把硬體存取集中在一套手寫組語 runtime(far-call,見 docs/06):
 *   VGA planar 繪圖(0x111b)、Sound Blaster(0x129c)、鍵盤(0x1104)、
 *   滑鼠(0x10b6)、檔案 I/O(0x1053)、BIOS 視訊(0x109c)。
 * 現代 SDL2 port 不模擬 real-mode;改在「邊界」用 SDL2/stdio 重寫等價語意。
 * 遊戲邏輯(re/*.c 反編譯出的 C)日後改成呼叫這層,而非 far-call wrapper。
 *
 * 設計原則(deep module + adapters at edges):
 *   - 對外介面窄:framebuffer 4bpp + palette + 輸入事件 + 計時 + 檔案。
 *   - 內部複雜度(planar->packed、scancode 對應、HiDPI 放大)藏在實作。
 */
#ifndef DQ3_RUNTIME_H
#define DQ3_RUNTIME_H

#include <stdint.h>
#include <stddef.h>

/* ---- 顯示:邏輯畫布 = 原版 VGA 解析(預設 640x350,4bpp packed)---- */
#define DQ3_SCREEN_W 640
#define DQ3_SCREEN_H 350

/* 內部畫布放大倍率(retro-cjk-hires-canvas:nearest 放大,中文可用 24x24)。
 * 視窗 = 邏輯畫布 * SCALE。底圖以 nearest-neighbor 放大保持銳利。 */
#ifndef DQ3_SCALE
#define DQ3_SCALE 2
#endif

/* RGB 調色盤條目(原版每色 6-bit VGA DAC,載入時擴成 8-bit)。 */
typedef struct { uint8_t r, g, b; } dq3_color;

/* runtime 生命週期 */
int  dq3_rt_init(const char *window_title);   /* 建 SDL window/renderer/texture */
void dq3_rt_quit(void);

/* framebuffer:4bpp packed(每 byte 一像素,值 0..255 索引 palette)。
 * 遊戲層寫這塊 buffer,呼叫 present 才上螢幕。 */
uint8_t *dq3_fb(void);                          /* 邏輯畫布像素緩衝(W*H) */
void dq3_set_palette(const dq3_color *pal, int count);
void dq3_present(void);                          /* packed->RGB,nearest 放大,翻頁 */

/* ---- 輸入:DOS scancode 語意,但來源是 SDL 事件 ---- */
/* 回傳 BIOS scancode(0=無),對應 docs/08 主迴圈讀鍵介面 kbd_read_scancode。 */
uint8_t dq3_poll_scancode(void);
int  dq3_should_quit(void);                      /* 視窗關閉 / ESC(由 host 決定) */

/* ---- 計時:取代 BIOS tick / frame_wait(sub_ee23)---- */
void     dq3_delay_ms(uint32_t ms);
uint32_t dq3_ticks_ms(void);

/* ---- 檔案 I/O:取代 seg 0x1053 的 DOS AH=3Dh/3Fh/3Eh,改 stdio ---- */
/* 從 assets 目錄讀整檔到 malloc buffer;*len 回傳長度,失敗回 NULL。 */
uint8_t *dq3_load_file(const char *name, size_t *len);
void     dq3_set_assets_dir(const char *dir);

#endif /* DQ3_RUNTIME_H */
