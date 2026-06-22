/* sdl_main.c — DQ3 SDL2 port PoC(最小可跑里程碑)。
 *
 * 證明「解碼資產 -> SDL 顯示」管線通:用已解的 PCX loader 載入標題畫面
 * (TIT*.P,640x350 4bpp + 自帶 16 色 palette),在 SDL2 視窗顯示,
 * 並走最小事件迴圈(ESC / 關窗離開,方向鍵 scancode 已接 runtime)。
 *
 * 這不是完整 remake;是 remake 的地基:確認 runtime shim(dq3_runtime)+
 * 資產解碼(dq3_pcx)+ 顯示管線可在現代跨平台環境運作。
 *
 * 用法: dq3_sdl <assets_dir> [TITG.P]
 *   無參數時預設 assets_dir=. 、檔名=TITG.P。
 *   headless(CI/容器)可設 SDL_VIDEODRIVER=dummy + 設 DQ3_DUMP=out.ppm 只驗解碼。
 */
#include "dq3_runtime.h"
#include "dq3_pcx.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void dump_ppm(const char *path, const uint8_t *fb, const dq3_color *pal)
{
    FILE *f = fopen(path, "wb");
    int i, n = DQ3_SCREEN_W * DQ3_SCREEN_H;
    if (!f) return;
    fprintf(f, "P6\n%d %d\n255\n", DQ3_SCREEN_W, DQ3_SCREEN_H);
    for (i = 0; i < n; i++) {
        dq3_color c = pal[fb[i] & 0x0f];
        fputc(c.r, f); fputc(c.g, f); fputc(c.b, f);
    }
    fclose(f);
}

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    const char *title  = (argc > 2) ? argv[2] : "TITG.P";
    const char *dump   = getenv("DQ3_DUMP");
    uint8_t *raw; size_t rawlen;
    dq3_color pal16[16];
    uint8_t *fb;
    int rc;

    /* headless 解碼驗證模式:只跑 loader+decoder,dump PPM,不開視窗。 */
    if (dump) {
        FILE *f; long sz; char path[2048];
        snprintf(path, sizeof(path), "%s/%s", assets, title);
        f = fopen(path, "rb");
        if (!f) { fprintf(stderr, "open %s failed\n", path); return 2; }
        fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
        raw = (uint8_t *)malloc((size_t)sz);
        if (fread(raw, 1, (size_t)sz, f) != (size_t)sz) { fclose(f); return 2; }
        fclose(f); rawlen = (size_t)sz;
        {
            static uint8_t fbuf[DQ3_SCREEN_W * DQ3_SCREEN_H];
            rc = dq3_pcx_decode(raw, rawlen, fbuf, DQ3_SCREEN_W, DQ3_SCREEN_H, pal16);
            if (rc != 0) { fprintf(stderr, "pcx decode rc=%d\n", rc); return 3; }
            dump_ppm(dump, fbuf, pal16);
            fprintf(stderr, "decoded %s (%lu bytes) -> %s OK\n",
                    title, (unsigned long)rawlen, dump);
        }
        free(raw);
        return 0;
    }

    if (dq3_rt_init("DQ3 (精訊) — SDL2 port PoC") != 0) return 1;
    dq3_set_assets_dir(assets);

    raw = dq3_load_file(title, &rawlen);
    if (!raw) { fprintf(stderr, "load %s failed\n", title); dq3_rt_quit(); return 2; }

    fb = dq3_fb();
    rc = dq3_pcx_decode(raw, rawlen, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, pal16);
    free(raw);
    if (rc != 0) { fprintf(stderr, "pcx decode rc=%d\n", rc); dq3_rt_quit(); return 3; }
    dq3_set_palette(pal16, 16);

    while (!dq3_should_quit()) {
        dq3_poll_scancode();      /* 方向鍵已轉 BIOS scancode,日後接主迴圈 */
        dq3_present();
        dq3_delay_ms(16);         /* ~60fps;日後改 frame_wait 等價節拍 */
    }

    dq3_rt_quit();
    return 0;
}
