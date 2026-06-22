/* dq3_runtime.c — DOS 硬體 runtime 的 SDL2 / stdio 實作。
 * 把原版 far-call runtime 的語意換成 SDL2;內部複雜度藏在此檔。
 */
#include "dq3_runtime.h"
#include <SDL.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static SDL_Window   *g_win;
static SDL_Renderer *g_ren;
static SDL_Texture  *g_tex;             /* 邏輯解析的 streaming texture */
static uint8_t       g_fb[DQ3_SCREEN_W * DQ3_SCREEN_H];   /* indexed framebuffer */
static dq3_color     g_pal[256];
static uint32_t      g_pal32[256];      /* 預乘 ARGB8888 快取 */
static int           g_have_sdl;        /* 是否已初始化 SDL(headless 可不開) */
static int           g_quit;
static char          g_assets[1024] = ".";

int dq3_rt_init(const char *title)
{
    if (SDL_Init(SDL_INIT_VIDEO | SDL_INIT_TIMER) != 0) {
        fprintf(stderr, "SDL_Init: %s\n", SDL_GetError());
        return -1;
    }
    g_have_sdl = 1;
    g_win = SDL_CreateWindow(title ? title : "DQ3",
                             SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED,
                             DQ3_SCREEN_W * DQ3_SCALE, DQ3_SCREEN_H * DQ3_SCALE,
                             SDL_WINDOW_SHOWN);
    if (!g_win) { fprintf(stderr, "CreateWindow: %s\n", SDL_GetError()); return -1; }

    g_ren = SDL_CreateRenderer(g_win, -1, SDL_RENDERER_ACCELERATED);
    if (!g_ren) g_ren = SDL_CreateRenderer(g_win, -1, SDL_RENDERER_SOFTWARE);
    if (!g_ren) { fprintf(stderr, "CreateRenderer: %s\n", SDL_GetError()); return -1; }

    /* nearest-neighbor:底圖整數放大保持銳利(retro-cjk-hires-canvas)。 */
    SDL_SetHint(SDL_HINT_RENDER_SCALE_QUALITY, "0");

    g_tex = SDL_CreateTexture(g_ren, SDL_PIXELFORMAT_ARGB8888,
                              SDL_TEXTUREACCESS_STREAMING,
                              DQ3_SCREEN_W, DQ3_SCREEN_H);
    if (!g_tex) { fprintf(stderr, "CreateTexture: %s\n", SDL_GetError()); return -1; }
    return 0;
}

void dq3_rt_quit(void)
{
    if (g_tex) SDL_DestroyTexture(g_tex);
    if (g_ren) SDL_DestroyRenderer(g_ren);
    if (g_win) SDL_DestroyWindow(g_win);
    if (g_have_sdl) SDL_Quit();
    g_tex = NULL; g_ren = NULL; g_win = NULL; g_have_sdl = 0;
}

uint8_t *dq3_fb(void) { return g_fb; }

void dq3_set_palette(const dq3_color *pal, int count)
{
    int i;
    if (count > 256) count = 256;
    for (i = 0; i < count; i++) {
        g_pal[i] = pal[i];
        g_pal32[i] = 0xFF000000u
                   | ((uint32_t)pal[i].r << 16)
                   | ((uint32_t)pal[i].g << 8)
                   |  (uint32_t)pal[i].b;
    }
}

void dq3_present(void)
{
    void *pixels; int pitch, row, i;
    if (!g_tex || !g_ren) return;
    if (SDL_LockTexture(g_tex, NULL, &pixels, &pitch) != 0) return;
    for (row = 0; row < DQ3_SCREEN_H; row++) {
        uint32_t *d = (uint32_t *)((uint8_t *)pixels + row * pitch);
        const uint8_t *s = &g_fb[row * DQ3_SCREEN_W];
        for (i = 0; i < DQ3_SCREEN_W; i++) d[i] = g_pal32[s[i]];
    }
    SDL_UnlockTexture(g_tex);
    SDL_RenderClear(g_ren);
    SDL_RenderCopy(g_ren, g_tex, NULL, NULL);   /* nearest 放大到視窗 */
    SDL_RenderPresent(g_ren);
}

int dq3_dump_ppm(const char *path)
{
    FILE *f = fopen(path, "wb");
    int i, n = DQ3_SCREEN_W * DQ3_SCREEN_H;
    if (!f) return -1;
    fprintf(f, "P6\n%d %d\n255\n", DQ3_SCREEN_W, DQ3_SCREEN_H);
    for (i = 0; i < n; i++) {
        dq3_color c = g_pal[g_fb[i]];
        fputc(c.r, f); fputc(c.g, f); fputc(c.b, f);
    }
    fclose(f);
    return 0;
}

/* ---- 輸入:SDL keysym → BIOS scancode(方向 + 選單 + ESC)---- */
static uint8_t sdl_to_bios_scancode(SDL_Keycode k)
{
    switch (k) {
        case SDLK_UP:     return 0x48;
        case SDLK_DOWN:   return 0x50;
        case SDLK_LEFT:   return 0x4b;
        case SDLK_RIGHT:  return 0x4d;
        case SDLK_RETURN: return 0x1c;   /* 確定 */
        case SDLK_SPACE:  return 0x39;   /* 選單 / 動作 */
        case SDLK_ESCAPE: return 0x01;   /* 取消 / 返回(離開另由 F10,見 rules/esc) */
        default:          return 0;
    }
}

uint8_t dq3_poll_scancode(void)
{
    SDL_Event e; uint8_t sc = 0;
    if (!g_have_sdl) return 0;
    while (SDL_PollEvent(&e)) {
        if (e.type == SDL_QUIT) g_quit = 1;
        else if (e.type == SDL_KEYDOWN) {
            SDL_Keycode k = e.key.keysym.sym;
            /* F10 = 離開(本階段標題畫面用;ESC 只送取消 scancode,不離開) */
            if (k == SDLK_F10) g_quit = 1;
            else {
                uint8_t s = sdl_to_bios_scancode(k);
                if (s) sc = s;
            }
        }
    }
    return sc;
}

int dq3_should_quit(void) { return g_quit; }

void dq3_delay_ms(uint32_t ms) { if (g_have_sdl) SDL_Delay(ms); }
uint32_t dq3_ticks_ms(void)    { return g_have_sdl ? SDL_GetTicks() : 0; }

void dq3_set_assets_dir(const char *dir)
{
    if (dir) { strncpy(g_assets, dir, sizeof(g_assets) - 1); g_assets[sizeof(g_assets)-1] = 0; }
}

uint8_t *dq3_load_file(const char *name, size_t *len)
{
    char path[2048];
    FILE *f; long sz; uint8_t *buf;
    snprintf(path, sizeof(path), "%s/%s", g_assets, name);
    f = fopen(path, "rb");
    if (!f) { if (len) *len = 0; return NULL; }
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    if (sz < 0) { fclose(f); if (len) *len = 0; return NULL; }
    buf = (uint8_t *)malloc((size_t)sz);
    if (buf && fread(buf, 1, (size_t)sz, f) != (size_t)sz) { free(buf); buf = NULL; }
    fclose(f);
    if (len) *len = buf ? (size_t)sz : 0;
    return buf;
}
