/* main.c — DQ3 remake 第一階段里程碑:顯示標題畫面(TITG.P)。
 *
 * 證明地基通:runtime shim(dq3_runtime)+ 資產解碼(dq3_assets)+ 顯示管線
 * 能在現代跨平台環境運作。流程:載入 TIT*.P → PCX 解碼成 indexed + 16 色
 * palette → 寫進 framebuffer → present(nearest 放大)。
 *
 * 用法:
 *   dq3_remake <assets_dir> [TITG.P]
 * headless(CI/容器)驗證:設環境變數 DQ3_DUMP=out.ppm,只跑載入+解碼+dump,
 *   不開視窗(配 SDL_VIDEODRIVER=dummy);或加 SDL_VIDEODRIVER=offscreen 跑視窗路徑。
 */
#include "dq3_runtime.h"
#include "dq3_assets.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int load_and_decode_title(const char *assets, const char *name,
                                 uint8_t *fb, dq3_color pal16[16])
{
    uint8_t *raw; size_t rawlen; int rc;
    char path[2048];
    FILE *f; long sz;

    /* 直接讀檔(headless 路徑不依賴 runtime 的 assets dir 設定) */
    snprintf(path, sizeof(path), "%s/%s", assets, name);
    f = fopen(path, "rb");
    if (!f) { fprintf(stderr, "open %s failed\n", path); return -1; }
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    if (sz <= 0) { fclose(f); return -1; }
    raw = (uint8_t *)malloc((size_t)sz);
    if (!raw || fread(raw, 1, (size_t)sz, f) != (size_t)sz) {
        fclose(f); free(raw); return -1;
    }
    fclose(f); rawlen = (size_t)sz;

    rc = dq3_pcx_decode(raw, rawlen, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, pal16);
    free(raw);
    if (rc != 0) { fprintf(stderr, "pcx decode rc=%d\n", rc); return rc; }
    return 0;
}

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    const char *title  = (argc > 2) ? argv[2] : "TITG.P";
    const char *dump   = getenv("DQ3_DUMP");
    dq3_color pal16[16];

    if (dq3_rt_init("DQ3 (精訊) — 重製 Remake") != 0) return 1;
    dq3_set_assets_dir(assets);

    if (load_and_decode_title(assets, title, dq3_fb(), pal16) != 0) {
        dq3_rt_quit();
        return 2;
    }
    dq3_set_palette(pal16, 16);

    /* headless 驗證:present 一次後 dump PPM,離開。 */
    if (dump) {
        dq3_present();
        if (dq3_dump_ppm(dump) == 0)
            fprintf(stderr, "decoded %s -> %s OK\n", title, dump);
        dq3_rt_quit();
        return 0;
    }

    /* 互動:顯示標題,F10 / 關窗離開(ESC 只送取消 scancode,見 rules/esc)。 */
    while (!dq3_should_quit()) {
        dq3_poll_scancode();
        dq3_present();
        dq3_delay_ms(16);     /* ~60fps;日後改 frame_wait(sub_ee23)等價節拍 */
    }

    dq3_rt_quit();
    return 0;
}
