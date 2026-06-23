/* main.c — DQ3 remake 進入點。
 *
 * 模式:
 *   title(預設):顯示標題畫面(TIT*.P)。
 *   field:地表場景(dq3_field → scene),方向鍵走動 + 碰撞 + 攝影機跟隨。
 *   town:城鎮場景(dq3_town → scene);CTY/section/BLK 編號可由環境變數指定:
 *         DQ3_CTY=CTY00.DAT DQ3_SECT=0 DQ3_BLKN=1。
 *   battle:互動戰鬥(dq3_battlescene);DQ3_MON=怪 id、DQ3_MON_N=隻數、DQ3_BATTLE_SCRIPT、DQ3_SEED。
 *   game:地表↔城鎮↔戰鬥串接(走動隨機遭遇、SPACE 進出城鎮、換場重套 palette)。
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
#include "dq3_battlescene.h"
#include "dq3_text.h"
#include "dq3_dialogue.h"
#include "dq3_exedata.h"
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

/* ---- battle 模式:互動戰鬥(dq3_battlescene)---- */
static int run_battle(const char *assets, const char *dump)
{
    int mon   = getenv("DQ3_MON")   ? atoi(getenv("DQ3_MON"))   : 5;   /* 預設史萊姆 */
    int count = getenv("DQ3_MON_N") ? atoi(getenv("DQ3_MON_N")) : 3;
    const char *script = getenv("DQ3_BATTLE_SCRIPT");                  /* headless 指令序列 */
    unsigned seed = getenv("DQ3_SEED") ? (unsigned)strtoul(getenv("DQ3_SEED"),0,0) : 0x1234567u;
    int oc = dq3_battlescene_run(assets, mon, count, -1, script, dump, seed);
    return oc < 0 ? 6 : 0;
}

/* ---- game 模式:地表↔城鎮↔戰鬥 串接狀態機 ----
 * 地表走動 → 步數遭遇隨機戰鬥;SPACE 進/出城鎮(demo 觸發,精確 tile→CTY 待 load_cty 表 RE)。
 * 每次戰鬥/換場後重套目的場景 palette —— 實際運用 bug #8 修正(docs/28)。 */
static unsigned g_seed = 0x2468ace0u;
static unsigned grnd(void){ g_seed^=g_seed<<13; g_seed^=g_seed>>17; g_seed^=g_seed<<5; return g_seed; }

static void load_field_hero(dq3_scene *s, const char *assets)
{
    dq3_scene_load_hero(s, assets, 16, NULL);   /* 金髮勇者 */
}

/* 地形 → 戰鬥背景頁(packbg index)。terrain 取自內建 dq3x_terrain(DGROUP 抽出),
 * 對映到 packbg 陣列中視覺相符的背景(草原/丘陵/洞窟…)。remake 不依賴 DQ3.EXE。 */
static int field_bg_page(const dq3_scene *s)
{
    /* terrain → packbg 索引(草原22=藍天綠地對 game3.png / 丘陵40 / 山20 / 洞窟30) */
    static const int T2BG[8] = { 22, 40, 20, 30, 40, 20, 22, 22 };
    int tile = s->index_map[s->py * s->map_w + s->px];
    int terr = dq3x_terrain[tile & 0xff];
    return T2BG[terr & 7];
}

static int run_game(const char *assets, const char *dump)
{
    char err[256] = {0};
    dq3_scene *field, *town = NULL, *cur;
    int in_town = 0, enc, fx = 0, fy = 0;
    const int over_pool[4] = { 5, 6, 1, 0 };   /* 地表遭遇怪池(史萊姆系) */
    dq3_dialogue dlg; int dlg_ok = 0, dlg_rec = 1;
    int tsect = getenv("DQ3_SECT") ? atoi(getenv("DQ3_SECT")) : 0;  /* 城鎮 section(有事件者測試)*/

    field = dq3_field_load(assets, err, sizeof err);
    if (!field) { fprintf(stderr, "field: %s\n", err); return 3; }
    load_field_hero(field, assets);
    cur = field; dq3_scene_apply_palette(cur);
    /* 對話文字(阿里阿罕 = D3TXT01) */
    dlg_ok = (dq3_dialogue_load(&dlg, assets, "D3TXT01.TXT", err, sizeof err) == 0);
    enc = 6 + (int)(grnd() % 8);

    if (dump) {
        /* headless demo:走到觸發戰鬥 → 進城鎮,沿途 log,dump 末幀 */
        int steps, mon, oc;
        fprintf(stderr, "=== game: 地表起點 ===\n");
        for (steps = 0; steps < 12; steps++) {
            if (dq3_scene_input(cur, 0x4d)) {   /* 往右走 */
                if (--enc <= 0) {
                    mon = over_pool[grnd() % 4];
                    fprintf(stderr, "--- 第 %d 步:遭遇怪 id%d(背景頁 %d)! ---\n", steps+1, mon, field_bg_page(cur));
                    oc = dq3_battlescene_run(assets, mon, 1 + (int)(grnd()%3), field_bg_page(cur), "FFFFFFFF", NULL, grnd());
                    fprintf(stderr, "    戰鬥結束 outcome=%d,回地表(重套 palette)\n", oc);
                    dq3_scene_apply_palette(cur);   /* bug #8:回地表還原色盤 */
                    enc = 6 + (int)(grnd() % 8);
                    break;
                }
            }
        }
        fprintf(stderr, "=== 進城鎮 CTY00(換場 + 重套 palette)===\n");
        town = dq3_town_load(assets, "CTY00.DAT", tsect, 1, err, sizeof err);
        if (town) { load_field_hero(town, assets); cur = town; dq3_scene_apply_palette(cur);
            fprintf(stderr, "城鎮 sect%d 事件數=%d\n", tsect, town->n_events); }
        dq3_scene_render(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        /* demo:在城鎮開一段對話(D3TXT01 rec 1)疊在場景上 */
        if (dlg_ok && dq3_dialogue_open(&dlg, 1) == 0) {
            fprintf(stderr, "=== 對話 D3TXT01 rec1(疊在城鎮)===\n");
            dq3_dialogue_render(&dlg, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        }
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "game -> %s OK\n", dump);
        if (dlg_ok) dq3_dialogue_free(&dlg);
        if (town) dq3_scene_free(town);
        dq3_scene_free(field);
        return 0;
    }

    /* 互動:方向走動 + 隨機遭遇;SPACE 進/出城鎮;Enter 對話(前方有事件時)。 */
    while (!dq3_should_quit()) {
        uint8_t sc;
        dq3_scene_render(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        if (dlg_ok && dq3_dialogue_is_open(&dlg))
            dq3_dialogue_render(&dlg, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        sc = dq3_poll_scancode();
        if (dlg_ok && dq3_dialogue_is_open(&dlg)) {
            if (sc == 0x1c || sc == 0x39) dq3_dialogue_advance(&dlg);  /* Enter/Space 翻頁/關閉 */
        } else if (sc == 0x1c) {            /* Enter:調べる本格 tile 事件(docs/31 真事件系統)*/
            int et, ep;
            if (dq3_scene_tile_event(cur, cur->px, cur->py, &et, &ep)) {
                const char *tn = et==0?"調べる/寶箱":et==2?"傳送/門":(et==1||et==3)?"條件":"道具/其他";
                fprintf(stderr, "事件: type=%d(%s) param=0x%x\n", et, tn, ep);
                if (dlg_ok) { dlg_rec = (ep && ep < dlg.txt.n_records) ? ep : 1; dq3_dialogue_open(&dlg, dlg_rec); }
            }
        } else if (sc == 0x39) {            /* SPACE:進/出城鎮 */
            if (!in_town) {
                town = dq3_town_load(assets, "CTY00.DAT", tsect, 1, err, sizeof err);
                if (town) { load_field_hero(town, assets); cur = town; in_town = 1; dq3_scene_apply_palette(cur);
                    fprintf(stderr, "城鎮 sect%d 事件數=%d\n", tsect, town->n_events); }
            } else {
                if (town) { dq3_scene_free(town); town = NULL; }
                cur = field; in_town = 0; dq3_scene_apply_palette(cur);
            }
        } else if (sc == 0x48 || sc == 0x50 || sc == 0x4b || sc == 0x4d) {
            int moved = dq3_scene_input(cur, sc);   /* 對話中不在此分支(上面已攔)*/
            if (moved) {
                int et, ep;
                /* 踩到 type-2 傳送/門/階梯 → 查 warp 表 0x4ea0 目的地(docs/31)*/
                if (in_town && dq3_scene_tile_event(cur, cur->px, cur->py, &et, &ep) && et == 2
                    && ep < DQ3X_WARP_N) {
                    int dmap = dq3x_warp[ep][0], dx = dq3x_warp[ep][1], dy = dq3x_warp[ep][2];
                    fprintf(stderr, "傳送門:warp[%d] → map=%d (%d,%d)\n", ep, dmap, dx, dy);
                    /* demo:離開城鎮回地表(精確 map→場景對映待 RE load_cty/世界表)*/
                    if (town) { dq3_scene_free(town); town = NULL; }
                    cur = field; in_town = 0; dq3_scene_apply_palette(cur);
                }
                if (moved && !in_town && --enc <= 0) {
                    int mon = over_pool[grnd() % 4];
                    dq3_battlescene_run(assets, mon, 1 + (int)(grnd()%3), field_bg_page(cur), NULL, NULL, grnd());
                    dq3_scene_apply_palette(cur);   /* bug #8:戰後還原地表色盤 */
                    enc = 6 + (int)(grnd() % 8);
                }
            }
        }
        dq3_delay_ms(16);
        (void)fx; (void)fy;
    }
    if (dlg_ok) dq3_dialogue_free(&dlg);
    if (town) dq3_scene_free(town);
    dq3_scene_free(field);
    return 0;
}

/* ---- text 模式:對話視窗渲染(CJK 字模)---- */
static int pal_near2(const dq3_color *p,int n,int r,int g,int b){
    int i,best=0; long bd=-1;
    for(i=0;i<n;i++){ long dr=p[i].r-r,dg=p[i].g-g,db=p[i].b-b,d=dr*dr+dg*dg+db*db;
        if(bd<0||d<bd){bd=d;best=i;} }
    return best;
}
static int run_text(const char *assets, const char *dump)
{
    char err[256]={0};
    dq3_color pal[256]; int pn;
    uint8_t *raw; size_t rl;
    dq3_text t;
    const char *txt = getenv("DQ3_TXT") ? getenv("DQ3_TXT") : "D3TXT01.TXT";
    int rec = getenv("DQ3_REC") ? atoi(getenv("DQ3_REC")) : 1;
    int white, black, frame, bg;
    uint8_t *fb = dq3_fb();
    int wx=40, wy=200, ww=DQ3_SCREEN_W-80, wh=130;

    raw=dq3_load_file("DQ3.PAL",&rl);
    if(!raw){ fprintf(stderr,"DQ3.PAL\n"); return 7; }
    pn=dq3_pal_decode(raw,rl,pal,256); free(raw); dq3_set_palette(pal,pn);
    white=pal_near2(pal,pn,255,255,255); black=pal_near2(pal,pn,0,0,0);
    frame=pal_near2(pal,pn,255,255,255); bg=pal_near2(pal,pn,16,16,32);

    if(dq3_text_load(&t,assets,txt,err,sizeof err)!=0){ fprintf(stderr,"text: %s\n",err); return 7; }
    fprintf(stderr,"%s 記錄數=%d,畫 rec %d\n", txt, t.n_records, rec);

    /* 背景 + 對話視窗框 */
    memset(fb,(uint8_t)black,(size_t)DQ3_SCREEN_W*DQ3_SCREEN_H);
    { int r,c; for(r=0;r<wh;r++)for(c=0;c<ww;c++){ int yy=wy+r,xx=wx+c;
        if(yy>=0&&yy<DQ3_SCREEN_H&&xx>=0&&xx<DQ3_SCREEN_W) fb[yy*DQ3_SCREEN_W+xx]=(uint8_t)bg; } }
    { int c; for(c=0;c<ww;c++){ fb[wy*DQ3_SCREEN_W+wx+c]=(uint8_t)frame; fb[(wy+wh-1)*DQ3_SCREEN_W+wx+c]=(uint8_t)frame; }
      int r; for(r=0;r<wh;r++){ fb[(wy+r)*DQ3_SCREEN_W+wx]=(uint8_t)frame; fb[(wy+r)*DQ3_SCREEN_W+wx+ww-1]=(uint8_t)frame; } }

    dq3_text_draw_record(&t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+12, wy+12,
                         (ww-24)/DQ3_GLYPH_PX, (wh-24)/DQ3_GLYPH_PX, rec, (uint8_t)white);

    if(dump){ dq3_present(); if(dq3_dump_ppm(dump)==0) fprintf(stderr,"text -> %s OK\n",dump);
        dq3_text_free(&t); return 0; }
    while(!dq3_should_quit()){ dq3_poll_scancode(); dq3_present(); dq3_delay_ms(16); }
    dq3_text_free(&t); return 0;
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
    } else if (strcmp(mode, "game") == 0) {
        rc = run_game(assets, dump);
    } else if (strcmp(mode, "text") == 0) {
        rc = run_text(assets, dump);
    } else {
        const char *title = (argc > 3) ? argv[3]
                          : (strcmp(mode, "title") == 0 ? "TITG.P" : mode);
        rc = run_title(assets, title, dump);
    }

    dq3_rt_quit();
    return rc;
}
