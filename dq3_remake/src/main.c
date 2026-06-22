/* main.c — DQ3 remake 進入點。
 *
 * 兩種模式:
 *   title(預設):顯示標題畫面(TIT*.P),證明 PCX 解碼 + 顯示管線。
 *   field:地表場景引擎(dq3_field),可方向鍵走動 + 碰撞,攝影機跟隨玩家。
 *
 * 用法:
 *   dq3_remake <assets_dir> [title|field] [titlefile]
 * headless 驗證(CI / 容器,配 SDL_VIDEODRIVER=dummy):
 *   DQ3_DUMP=out.ppm           只跑載入+解碼+繪一幀+dump,不開視窗。
 *   DQ3_WALK="RRDDLLUU"        field 模式下先依序套用移動(R/L/U/D),再 dump 末幀。
 */
#include "dq3_runtime.h"
#include "dq3_assets.h"
#include "dq3_field.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ---- title 模式(原地基里程碑)---- */
static int load_and_decode_title(const char *assets, const char *name,
                                 uint8_t *fb, dq3_color pal16[16])
{
    uint8_t *raw; size_t rawlen; int rc;
    char path[2048]; FILE *f; long sz;
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

static int run_title(const char *assets, const char *title, const char *dump)
{
    dq3_color pal16[16];
    if (load_and_decode_title(assets, title, dq3_fb(), pal16) != 0) return 2;
    dq3_set_palette(pal16, 16);
    if (dump) {
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "title %s -> %s OK\n", title, dump);
        return 0;
    }
    while (!dq3_should_quit()) {
        dq3_poll_scancode();
        dq3_present();
        dq3_delay_ms(16);
    }
    return 0;
}

/* ---- field 模式(地表走動)---- */
static uint8_t walk_char_to_scancode(char c)
{
    switch (c) {
        case 'U': case 'u': return 0x48;
        case 'D': case 'd': return 0x50;
        case 'L': case 'l': return 0x4b;
        case 'R': case 'r': return 0x4d;
        default: return 0;
    }
}

static int run_field(const char *assets, const char *dump)
{
    char err[256] = {0};
    const dq3_color *pal; int pcount;
    int tx, ty;
    dq3_field *f = dq3_field_create(assets, err, sizeof err);
    if (!f) { fprintf(stderr, "field create failed: %s\n", err); return 3; }

    pal = dq3_field_palette(f, &pcount);
    dq3_set_palette(pal, pcount);

    if (dump) {
        const char *walk = getenv("DQ3_WALK");
        dq3_field_get_player(f, &tx, &ty);
        fprintf(stderr, "field start player=(%d,%d)\n", tx, ty);
        if (walk) {
            int moved = 0, blocked = 0; const char *p;
            for (p = walk; *p; p++) {
                uint8_t sc = walk_char_to_scancode(*p);
                if (!sc) continue;
                if (dq3_field_input(f, sc)) moved++; else blocked++;
            }
            dq3_field_get_player(f, &tx, &ty);
            fprintf(stderr, "after walk \"%s\": moved=%d blocked=%d player=(%d,%d)\n",
                    walk, moved, blocked, tx, ty);
        }
        dq3_field_render(f, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "field -> %s OK\n", dump);
        dq3_field_destroy(f);
        return 0;
    }

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc) dq3_field_input(f, sc);
        dq3_field_render(f, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        dq3_delay_ms(16);
    }
    dq3_field_destroy(f);
    return 0;
}

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    const char *mode   = (argc > 2) ? argv[2] : "title";
    const char *dump   = getenv("DQ3_DUMP");
    int is_field = (strcmp(mode, "field") == 0);
    int rc;

    if (dq3_rt_init(is_field ? "DQ3 (精訊) — 地表 Field"
                             : "DQ3 (精訊) — 重製 Remake") != 0) return 1;
    dq3_set_assets_dir(assets);

    if (is_field) {
        rc = run_field(assets, dump);
    } else {
        const char *title = (argc > 3) ? argv[3]
                          : (strcmp(mode, "title") == 0 ? "TITG.P" : mode);
        rc = run_title(assets, title, dump);
    }

    dq3_rt_quit();
    return rc;
}
