/* main.c — DQ3 remake 進入點。
 *
 * 模式:
 *   title(預設):顯示標題畫面(TIT*.P)。
 *   field:地表場景(dq3_field → scene),方向鍵走動 + 碰撞 + 攝影機跟隨。
 *   town:城鎮場景(dq3_town → scene);CTY/section/BLK 編號可由環境變數指定:
 *         DQ3_CTY=CTY00.DAT DQ3_SECT=0 DQ3_BLKN=1。
 *   battle:戰鬥場景(怪群 sprite + 天空/地面背景);DQ3_MON=怪 id、DQ3_MON_N=隻數。
 *
 * 用法:dq3_remake <assets_dir> [title|field|town] [titlefile]
 * headless 驗證(配 SDL_VIDEODRIVER=dummy):
 *   DQ3_DUMP=out.ppm        繪一幀 + dump,不開視窗。
 *   DQ3_WALK="RRDDLLUU"     field/town 模式先依序套用移動(R/L/U/D),再 dump 末幀。
 */
#include "dq3_runtime.h"
#include "dq3_assets.h"
#include "dq3_field.h"
#include "dq3_town.h"
#include "dq3_scene.h"
#include "dq3_monster.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ---- title 模式 ---- */
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
    while (!dq3_should_quit()) { dq3_poll_scancode(); dq3_present(); dq3_delay_ms(16); }
    return 0;
}

/* ---- 場景(field / town)共用驅動 ---- */
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

static int run_scene(dq3_scene *s, const char *dump)
{
    int tx, ty;
    dq3_scene_apply_palette(s);   /* 進場套 palette(修 bug #8 契約,docs/28) */

    if (dump) {
        const char *walk = getenv("DQ3_WALK");
        fprintf(stderr, "scene %dx%d start player=(%d,%d)\n", s->map_w, s->map_h, s->px, s->py);
        if (walk) {
            int moved = 0, blocked = 0; const char *p;
            for (p = walk; *p; p++) {
                uint8_t sc = walk_char_to_scancode(*p);
                if (!sc) continue;
                if (dq3_scene_input(s, sc)) moved++; else blocked++;
            }
            tx = s->px; ty = s->py;
            fprintf(stderr, "after walk \"%s\": moved=%d blocked=%d player=(%d,%d)\n",
                    walk, moved, blocked, tx, ty);
        }
        dq3_scene_render(s, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "scene -> %s OK\n", dump);
        return 0;
    }

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc) dq3_scene_input(s, sc);
        dq3_scene_render(s, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        dq3_delay_ms(16);
    }
    return 0;
}

/* ---- battle 模式:怪群 sprite 繪在天空/地面背景(對 references/game3.png)---- */
static int pal_nearest(const dq3_color *pal, int n, int r, int g, int b)
{
    int i, best = 0; long bd = -1;
    for (i = 0; i < n; i++) {
        long dr = pal[i].r - r, dg = pal[i].g - g, db = pal[i].b - b;
        long d = dr*dr + dg*dg + db*db;
        if (bd < 0 || d < bd) { bd = d; best = i; }
    }
    return best;
}

static int run_battle(const char *assets, const char *dump)
{
    char err[256] = {0};
    dq3_color pal[256]; int pn;
    uint8_t *raw; size_t rlen;
    dq3_monster_sprite spr;
    int mon = getenv("DQ3_MON") ? atoi(getenv("DQ3_MON")) : 5;   /* 預設史萊姆 */
    int count = getenv("DQ3_MON_N") ? atoi(getenv("DQ3_MON_N")) : 6;
    int sky, ground, groundy = (int)(DQ3_SCREEN_H * 0.62);
    uint8_t *fb = dq3_fb();
    int i;

    /* MNSBK.PAL → palette */
    raw = dq3_load_file("MNSBK.PAL", &rlen);
    if (!raw) { fprintf(stderr, "load MNSBK.PAL failed\n"); return 5; }
    pn = dq3_pal_decode(raw, rlen, pal, 256); free(raw);
    dq3_set_palette(pal, pn);

    if (dq3_monster_sprite_decode(assets, mon, &spr, err, sizeof err) != 0) {
        fprintf(stderr, "monster %d sprite: %s\n", mon, err); return 6;
    }

    sky    = pal_nearest(pal, pn, 60, 140, 240);    /* 天空藍 */
    ground = pal_nearest(pal, pn, 40, 150, 40);     /* 地面綠 */

    /* 背景:上天空、下地面 */
    for (i = 0; i < DQ3_SCREEN_H; i++) {
        uint8_t c = (i < groundy) ? (uint8_t)sky : (uint8_t)ground;
        memset(fb + (size_t)i * DQ3_SCREEN_W, c, DQ3_SCREEN_W);
    }

    /* 怪群:橫向排開,站在地面線上方 */
    if (count < 1) count = 1; if (count > 8) count = 8;
    for (i = 0; i < count; i++) {
        int gx = (DQ3_SCREEN_W * (i + 1)) / (count + 1) - spr.w / 2;
        int gy = groundy - spr.h - 8;
        int r, c;
        for (r = 0; r < spr.h; r++)
            for (c = 0; c < spr.w; c++) {
                int yy = gy + r, xx = gx + c;
                if (!spr.opaque[r][c]) continue;
                if (yy >= 0 && yy < DQ3_SCREEN_H && xx >= 0 && xx < DQ3_SCREEN_W)
                    fb[yy * DQ3_SCREEN_W + xx] = spr.px[r][c];
            }
    }

    if (dump) {
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "battle mon=%d x%d -> %s OK\n", mon, count, dump);
        return 0;
    }
    while (!dq3_should_quit()) { dq3_poll_scancode(); dq3_present(); dq3_delay_ms(16); }
    return 0;
}

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    const char *mode   = (argc > 2) ? argv[2] : "title";
    const char *dump   = getenv("DQ3_DUMP");
    int rc;

    if (dq3_rt_init("DQ3 (精訊) — 重製 Remake") != 0) return 1;
    dq3_set_assets_dir(assets);

    if (strcmp(mode, "field") == 0 || strcmp(mode, "town") == 0) {
        char err[256] = {0};
        dq3_scene *s;
        if (strcmp(mode, "field") == 0) {
            s = dq3_field_load(assets, err, sizeof err);
        } else {
            const char *cty = getenv("DQ3_CTY");
            const char *sect = getenv("DQ3_SECT");
            const char *blkn = getenv("DQ3_BLKN");
            s = dq3_town_load(assets, cty ? cty : "CTY00.DAT",
                              sect ? atoi(sect) : 0,
                              blkn ? atoi(blkn) : 1, err, sizeof err);
        }
        if (!s) { fprintf(stderr, "%s load failed: %s\n", mode, err); rc = 3; }
        else {
            /* 主角 sprite:entry 與 facing→frame 對映可由 env 調(對 oracle 鎖定) */
            const char *he = getenv("DQ3_HERO_ENTRY");
            const char *fo = getenv("DQ3_FACING_ORDER");  /* 如 "0,2,1,3" */
            int entry = he ? atoi(he) : 16;   /* 預設金髮勇者(entry16,對 oracle 室內主角) */
            int order[4] = {0,1,2,3}; const int *op = NULL;
            if (fo && sscanf(fo, "%d,%d,%d,%d", &order[0],&order[1],&order[2],&order[3]) == 4) op = order;
            if (dq3_scene_load_hero(s, assets, entry, op) != 0)
                fprintf(stderr, "hero sprite load failed (entry %d) — 用佔位框\n", entry);
            rc = run_scene(s, dump);
            dq3_scene_free(s);
        }
    } else if (strcmp(mode, "battle") == 0) {
        rc = run_battle(assets, dump);
    } else {
        const char *title = (argc > 3) ? argv[3]
                          : (strcmp(mode, "title") == 0 ? "TITG.P" : mode);
        rc = run_title(assets, title, dump);
    }

    dq3_rt_quit();
    return rc;
}
