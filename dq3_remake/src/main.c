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
#include "dq3_inventory.h"
#include "dq3_roster.h"
#include "dq3_menu.h"
#include "dq3_nameinput.h"
#include "dq3_tavern.h"
#include "dq3_zhuyin.h"
#include "dq3_save.h"
#include "dq3_status.h"
#include "dq3_cmdmenu.h"
#include "dq3_encounter.h"
#include "dq3_combat.h"
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
    while (!dq3_should_quit()) { if (dq3_poll_scancode()==DQ3_SC_F10) break; dq3_present(); dq3_delay_ms(16); }
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
        fprintf(stderr, "scene %dx%d start player=(%d,%d) NPC=%d\n", s->map_w, s->map_h, s->px, s->py, s->n_npcs);
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
        if (sc == DQ3_SC_F10) break;        /* F10 離開(standalone 場景檢視)*/
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
    int oc;
    /* DQ3_BATTLE_PARTY=1:驗證「酒場建的隊伍接進戰鬥」— 建 2 名範例隊員(勇者/戰士)設為玩家隊。 */
    static dq3_roster vr; static dq3_party vp; static dq3_stats vs;
    if (getenv("DQ3_BATTLE_PARTY")) {
        uint16_t n1[2] = {106,187}, n2[2] = {107,144};   /* 「勇者」「戰士」當名 */
        int lv = getenv("DQ3_ST_LEVEL") ? atoi(getenv("DQ3_ST_LEVEL")) : 12;
        if (dq3_stats_load(&vs, assets, 1, NULL, 0) == 0) {
            dq3_roster_init(&vr); dq3_party_init(&vp);
            dq3_roster_create(&vr, &vs, 0, 0, n1, 2);   /* 勇者 */
            dq3_roster_create(&vr, &vs, 1, 0, n2, 2);   /* 戰士 */
            vr.list[0].weapon = 0x03; vr.list[0].armor = 0x21;  /* 勇者:銅劍(攻10)+皮甲冑(防8)*/
            vr.list[1].weapon = 0x0b;                            /* 戰士:鋼劍(攻28)*/
            { int i; for (i = 0; i < vr.count; i++) { dq3_member_init(&vr.list[i].m, &vs, vr.list[i].m.cls, lv);
                dq3_party_add(&vp, &vr, i); } }
            dq3_battlescene_set_party(&vr, &vp);
        }
    }
    oc = dq3_battlescene_run(assets, mon, count, -1, script, dump, seed);
    if (getenv("DQ3_BATTLE_PARTY"))   /* 驗證回寫:戰後印名冊隊員 等級/exp */
        { int i; for (i = 0; i < vr.count; i++)
            fprintf(stderr, "名冊回寫後:隊員%d Lv%d exp=%u\n", i, vr.list[i].m.level, vr.list[i].m.exp); }
    return oc < 0 ? 6 : 0;
}

/* ---- game 模式:地表↔城鎮↔戰鬥 串接狀態機 ----
 * 地表走動 → 步數遭遇隨機戰鬥;走到城鎮入口座標 → 進對應 CTY(load_cty 查表 0x748);
 * SPACE 亦可手動進/出城鎮(demo)。
 * 每次戰鬥/換場後重套目的場景 palette —— 實際運用 bug #8 修正(docs/28)。 */
static unsigned g_seed = 0x2468ace0u;
static unsigned grnd(void){ g_seed^=g_seed<<13; g_seed^=g_seed>>17; g_seed^=g_seed<<5; return g_seed; }

static void load_field_hero(dq3_scene *s, const char *assets)
{
    dq3_scene_load_hero(s, assets, 0, NULL);    /* 勇者 = DQ3MST.BLS entry0(c0:黑髮金帶紅袍,對 oracle 確認)。
                                                 * ★ 主角/隊員在 DQ3MST.BLS(各職業 sprite),非 DQ3MAN.BLS(NPC/村民)*/
    dq3_scene_load_npc_sprites(s, assets);      /* NPC sprite(城鎮才有 NPC;地表 no-op)*/
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

/* load_cty 查表(DGROUP 0x748):某層 (px,py) → CTY 號,無則 -1。
 * 對齊 RE(file 0x43b7):Y==loc.Y 且 (X==loc.X 或 X==loc.X+1)→ entry index = CTY。
 * map:0=地面、1=下層;只配對同層 entry(dq3x_cty_loc[i][2]==map)。 */
static int find_cty_at_map(int px, int py, int map)
{
    int i;
    for (i = 0; i < DQ3X_CTYLOC_N; i++) {
        int lx = dq3x_cty_loc[i][0], ly = dq3x_cty_loc[i][1], lm = dq3x_cty_loc[i][2];
        if (lm != map) continue;                     /* 只配對同層 + 跳過空槽(0xff)*/
        if (py == ly && (px == lx || px == lx + 1)) return i;
    }
    return -1;
}
static int find_cty_at(int px, int py) { return find_cty_at_map(px, py, 0); }  /* 地面 */

static void tavern_modal(const char *assets, dq3_roster *roster, dq3_party *party,
                         const dq3_stats *st, const dq3_text *text);
static void status_modal_page(const dq3_roster *roster, const dq3_party *party, const dq3_text *text, int start_page);
static void shop_modal(dq3_roster *roster, dq3_party *party, const dq3_items *items, const dq3_text *text, long *gold);
static int  cmd_modal(dq3_scene *scene, dq3_roster *roster, dq3_party *party,
                      dq3_inventory *inv, dq3_dialogue *dlg, int dlg_ok);
static void item_modal(const dq3_inventory *inv, const dq3_text *text);
static void tav_window(uint8_t *fb, int wx, int wy, int ww, int wh, uint8_t black, uint8_t frame, uint8_t bg);
static int  pal_near2(const dq3_color *p, int n, int r, int g, int b);

/* 自動存檔:把名冊/隊伍/道具/位置寫成 dq3_save.dat(remake 自有格式)。回 0=成功。
 * 路徑:DQ3_SAVE 環境變數,預設 "dq3_save.dat"(cwd;唯讀 cwd 時用 env 指可寫路徑)。 */
static const char *save_path(void) { return getenv("DQ3_SAVE") ? getenv("DQ3_SAVE") : "dq3_save.dat"; }

static int autosave_game(const dq3_roster *r, const dq3_party *p, const dq3_inventory *inv,
                         int cty, int px, int py)
{
    dq3_save_pos pos; int rc;
    pos.cty = cty; pos.px = px; pos.py = py;
    rc = dq3_save_write(save_path(), r, p, inv, pos);
    if (rc == 0) fprintf(stderr, "自動存檔 → %s(名冊%d 隊伍%d CTY%d)\n", save_path(), r->count, p->count, cty);
    else fprintf(stderr, "autosave 失敗(無法寫 %s)\n", save_path());
    return rc;
}

/* F10 離開確認:渲染「離開遊戲嗎」+ 是/否 選單。回 1=是(離開)、0=否(繼續)。
 * 會改 DAC palette,呼叫端離開後須重套場景色盤。ESC = 否。 */
static int confirm_quit(const dq3_text *text)
{
    static const uint16_t TITLE[5] = {502, 488, 113, 689, 534};  /* 離開遊戲嗎 */
    static const uint16_t YES[1] = {399}, NO[1] = {678};         /* 是 / 否 */
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t yellow; dq3_menu m;
    uint8_t *fb = dq3_fb();
    int wx = 180, wy = 120, ww = 200, wh = 110, i;

    raw = dq3_load_file("DQ3.PAL", &rl);
    if (!raw) return 0;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = pal_near2(pal,pn,255,255,255); bg = pal_near2(pal,pn,16,16,32);
    yellow = (uint8_t)pal_near2(pal,pn,255,255,0);

    dq3_menu_init(&m, wx + 24, wy + 50);
    dq3_menu_add(&m, YES, 1); dq3_menu_add(&m, NO, 1);
    m.cursor = 1;                                  /* 預設「否」(較安全)*/

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == 0x01) return 0;                  /* ESC = 否(取消離開)*/
        if (sc) { int sel = dq3_menu_input(&m, sc); if (sel >= 0) return (sel == 0); }
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        for (i = 0; i < 5; i++)
            dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+16+i*DQ3_GLYPH_PX, wy+16, TITLE[i], (uint8_t)white);
        dq3_menu_render(&m, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, (uint8_t)white, yellow);
        dq3_present(); dq3_delay_ms(16);
    }
    return 0;
}

static int run_game(const char *assets, const char *dump)
{
    char err[256] = {0};
    dq3_scene *field, *town = NULL, *cur;
    int in_town = 0, enc, fx = 0, fy = 0, cur_cty = -1;   /* cur_cty:目前所在 CTY 號(#2 gate)*/
    const int over_pool[4] = { 5, 6, 1, 0 };   /* 地表遭遇怪池(史萊姆系) */
    dq3_dialogue dlg; int dlg_ok = 0, dlg_rec = 1;
    int tsect = getenv("DQ3_SECT") ? atoi(getenv("DQ3_SECT")) : 0;  /* 城鎮 section(有事件者測試)*/
    dq3_inventory inv; dq3_storyflags flags;                        /* #2 合成:道具欄 + 劇情旗標 */
    dq3_roster roster; dq3_party party; dq3_stats gst;              /* 露依達酒場:名冊 + 隊伍 */
    long gold = 0;                                                  /* 隊伍金錢(戰利累積)*/
    dq3_items shop_items; int shop_ok;                             /* ITEM.DAT(商店價/可裝備)*/

    dq3_roster_init(&roster); dq3_party_init(&party); dq3_stats_load(&gst, assets, 1, NULL, 0);
    shop_ok = (dq3_items_load(&shop_items, assets, NULL, 0) == 0);

    field = dq3_field_load(assets, err, sizeof err);
    if (!field) { fprintf(stderr, "field: %s\n", err); return 3; }
    load_field_hero(field, assets);
    /* 地表起點 = 阿里阿罕 overworld tile(原版初始 [0x4f2f]=0x99/[0x4f31]=0xae,file 0x13a0)
     * → region 1(史萊姆/大烏鴉系弱怪,docs/39);取代 pick_open_start 的任意空地。 */
    if (153 < field->map_w && 174 < field->map_h) { field->px = 153; field->py = 174; }
    cur = field; dq3_scene_apply_palette(cur);
    /* 對話文字(阿里阿罕 = D3TXT01) */
    dlg_ok = (dq3_dialogue_load(&dlg, assets, "D3TXT01.TXT", err, sizeof err) == 0);
    enc = 6 + (int)(grnd() % 8);
    /* demo:身上帶兩材料,進祠堂「調べる」即可觸發 #2 合成(產彩虹水滴) */
    dq3_inv_init(&inv); dq3_flags_init(&flags);
    dq3_inv_add(&inv, ITEM_SUN_STONE); dq3_inv_add(&inv, ITEM_RAINCLOUD_ROD);
    /* 續玩:DQ3_LOAD 且存檔存在 → 讀回名冊/隊伍/道具(位置另需載入對應場景,先記錄)。 */
    if (getenv("DQ3_LOAD") && dq3_save_exists(save_path())) {
        dq3_save_pos pos;
        if (dq3_save_read(save_path(), &roster, &party, &inv, &pos) == 0)
            fprintf(stderr, "讀檔續玩 ← %s(名冊%d 隊伍%d,存檔位置 CTY%d (%d,%d))\n",
                    save_path(), roster.count, party.count, pos.cty, pos.px, pos.py);
    }

    if (dump) {
        /* headless demo:走到觸發戰鬥 → 進城鎮,沿途 log,dump 末幀 */
        int steps, mon, oc;
        fprintf(stderr, "=== game: 地表起點 ===\n");
        for (steps = 0; steps < 12; steps++) {
            if (dq3_scene_input(cur, 0x4d)) {   /* 往右走 */
                if (--enc <= 0) {
                    int reg = dq3_encounter_region(cur->px, cur->py);   /* 位置→region(docs/39)*/
                    mon = dq3_encounter_pick(reg, grnd());              /* region→候選怪 */
                    if (mon < 0) mon = over_pool[grnd() % 4];           /* 空 region 後備 */
                    fprintf(stderr, "--- 第 %d 步:region 0x%x 遭遇怪 id%d(背景頁 %d)! ---\n", steps+1, reg, mon, field_bg_page(cur));
                    dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                    oc = dq3_battlescene_run(assets, mon, 1 + (int)(grnd()%3), field_bg_page(cur), "FFFFFFFF", NULL, grnd());
                    fprintf(stderr, "    戰鬥結束 outcome=%d,回地表(重套 palette)\n", oc);
                    dq3_scene_apply_palette(cur);   /* bug #8:回地表還原色盤 */
                    enc = 6 + (int)(grnd() % 8);
                    break;
                }
            }
        }
        /* demo:進真正的合成祠堂 CTY93(section 格式已修正,17×17 祠堂房間可正確載入)。 */
        { int bn = dq3x_map_blknum[DQ3_SHRINE_CTY]; char cty[16];
          sprintf(cty, "CTY%02d.DAT", DQ3_SHRINE_CTY);
          fprintf(stderr, "=== 進合成祠堂 %s(BLK%d,換場 + 重套 palette)===\n", cty, bn);
          town = dq3_town_load(assets, cty, 0, bn, err, sizeof err);
          if (town) { load_field_hero(town, assets); cur = town; cur_cty = DQ3_SHRINE_CTY;
              dq3_scene_apply_palette(cur);
              fprintf(stderr, "祠堂 %dx%d 事件數=%d\n", cur->map_w, cur->map_h, town->n_events); }
          else fprintf(stderr, "載入祠堂失敗: %s\n", err); }
        /* demo #2:在祠堂(CTY93)發 scripted-event 83 → 合成(gate 命中,runner→0x776c→0x77ce)*/
        if (cur_cty == DQ3_SHRINE_CTY) {
            int sr = dq3_scripted_event_run(DQ3_SEVENT_RAINBOW_SYNTH, &inv, &flags, 1);
            fprintf(stderr, "=== #2 祠堂(CTY%d)scripted-event 83:result=0x%02x flag0x139=%d ===\n",
                    DQ3_SHRINE_CTY, sr, dq3_flags_get(&flags, DQ3_FLAG_RAINBOW_SYNTHED));
        }
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
        /* NPC 隨機走動(docs/35 §九):城鎮每幀步進;對話中凍結不動。 */
        if (in_town && !(dlg_ok && dq3_dialogue_is_open(&dlg))) dq3_scene_npc_tick(cur);
        dq3_scene_render(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        if (dlg_ok && dq3_dialogue_is_open(&dlg))
            dq3_dialogue_render(&dlg, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        sc = dq3_poll_scancode();
        if (sc == DQ3_SC_F10) {             /* F10:離開遊戲確認(Yes/No)+ 自動存檔 */
            if (confirm_quit(&dlg.txt)) {
                autosave_game(&roster, &party, &inv, cur_cty, cur->px, cur->py);
                dq3_rt_set_quit();
            }
            dq3_scene_apply_palette(cur);   /* confirm 改過 DAC,還原場景色盤 */
        } else if (dlg_ok && dq3_dialogue_is_open(&dlg)) {
            if (sc == 0x1c || sc == 0x39) dq3_dialogue_advance(&dlg);  /* Enter/Space 翻頁/關閉 */
        } else if (sc == 0x1c) {            /* Enter:調べる本格 tile 事件(docs/31 真事件系統)*/
            int et, ep;
            /* #2 祠堂:站在祭壇 NPC(CTY93 (8,8))旁、面向它調べる才觸發 scripted-event 83
             * (祭司 NPC 座標靜態解,docs/32;runner 0xabb2 → handler 0x776c → 0x77ce 尾段)。 */
            {
                int fdx = (cur->facing==3)-(cur->facing==1);   /* 右+1 左-1 */
                int fdy = (cur->facing==0)-(cur->facing==2);   /* 下+1 上-1 */
                int faceX = cur->px + fdx, faceY = cur->py + fdy;
                int at_altar = (faceX==DQ3_SHRINE_NPC_X && faceY==DQ3_SHRINE_NPC_Y)   /* 面向祭壇 */
                            || (cur->px==DQ3_SHRINE_NPC_X && cur->py==DQ3_SHRINE_NPC_Y); /* 或站其上 */
            if (in_town && cur_cty == DQ3_SHRINE_CTY && at_altar) {
                int sr = dq3_scripted_event_run(DQ3_SEVENT_RAINBOW_SYNTH, &inv, &flags, 1);
                if (sr >= 0) {
                    fprintf(stderr, "祠堂:雨和太陽合而為一 → 彩虹水滴 0x%02x(旗標0x139 已設)\n", sr);
                    if (dlg_ok) dq3_dialogue_open(&dlg, 1);
                }
            }
            }
            /* 露依達酒場(CTY00 sec0 西側 LUIDA 櫃台):面向/站在櫃台店員調べる → 開創角。
             * 位置由腳本 rec49 +轉場 metadata 定位(docs/36),取代先前的 T 鍵暫代。 */
            {
                int fdx = (cur->facing==3)-(cur->facing==1), fdy = (cur->facing==0)-(cur->facing==2);
                int fx = cur->px+fdx, fy = cur->py+fdy;
                int dfx = fx-DQ3_LUIDA_X, dfy = fy-DQ3_LUIDA_Y, dpx = cur->px-DQ3_LUIDA_X, dpy = cur->py-DQ3_LUIDA_Y;
                int near_ctr = (dfx<0?-dfx:dfx)+(dfy<0?-dfy:dfy) <= 1 || (dpx<0?-dpx:dpx)+(dpy<0?-dpy:dpy) <= 1;
                if (in_town && cur_cty == DQ3_LUIDA_CTY && near_ctr) {
                    tavern_modal(assets, &roster, &party, &gst, &dlg.txt);
                    dq3_scene_apply_palette(cur);
                    fprintf(stderr, "露依達酒場:名冊%d 人、隊伍%d 人\n", roster.count, party.count);
                }
            }
            if (dq3_scene_tile_event(cur, cur->px, cur->py, &et, &ep)) {
                const char *tn = et==0?"調べる/寶箱":et==2?"傳送/門":(et==1||et==3)?"條件":"道具/其他";
                fprintf(stderr, "事件: type=%d(%s) param=0x%x\n", et, tn, ep);
                if (dlg_ok) { dlg_rec = (ep && ep < dlg.txt.n_records) ? ep : 1; dq3_dialogue_open(&dlg, dlg_rec); }
            }
        } else if (sc == 0x14 && in_town && cur_cty == 0) {  /* T:阿里阿罕酒場捷徑(開發用;正式入口=西側 LUIDA 櫃台)*/
            tavern_modal(assets, &roster, &party, &gst, &dlg.txt);
            dq3_scene_apply_palette(cur);
            fprintf(stderr, "離開酒場:名冊%d 人、隊伍%d 人\n", roster.count, party.count);
        } else if (sc == 0x2e) {            /* C:野外指令窗(命令)— 對話/咒文/狀況/道具/裝備/調查 */
            cmd_modal(cur, &roster, &party, &inv, &dlg, dlg_ok);
            dq3_scene_apply_palette(cur);
        } else if (sc == 0x30 && in_town) {  /* B:武器/防具商店(城鎮內;買→裝備領頭隊員)*/
            if (shop_ok) shop_modal(&roster, &party, &shop_items, &dlg.txt, &gold);
            dq3_scene_apply_palette(cur);
        } else if (sc == 0x39) {            /* SPACE:進/出城鎮 */
            if (!in_town) {
                town = dq3_town_load(assets, "CTY00.DAT", tsect, 1, err, sizeof err);
                if (town) { load_field_hero(town, assets); cur = town; in_town = 1; cur_cty = 0;
                    dq3_scene_apply_palette(cur);
                    fprintf(stderr, "城鎮 sect%d 事件數=%d\n", tsect, town->n_events); }
            } else {
                if (town) { dq3_scene_free(town); town = NULL; }
                cur = field; in_town = 0; cur_cty = -1; dq3_scene_apply_palette(cur);
            }
        } else if (sc == 0x48 || sc == 0x50 || sc == 0x4b || sc == 0x4d) {
            int moved = dq3_scene_input(cur, sc);   /* 對話中不在此分支(上面已攔)*/
            if (!moved && in_town) {
                /* 被擋:若面向鎖門且隊伍持足夠等級鑰匙 → 開門後重試(docs/35 §八)。*/
                int kt = dq3_inv_key_tier(&inv);
                if (dq3_scene_try_open_facing_door(cur, kt)) {
                    fprintf(stderr, "開鎖:鑰匙等級%d 開門\n", kt);
                    moved = dq3_scene_input(cur, sc);
                }
            }
            if (moved) {
                int dcty, dsec, dx, dy;
                /* 踩到轉場 tile(門/階梯/出城)→ section+0xc 轉場表 {destCTY,destSec,X,Y}
                 * (docs/35)。destSec==0xff → 出城回地表;destCTY!=當前 → 跨 CTY 載入;
                 * 同 CTY → 換 section。一律置玩家於目的 (X,Y)。 */
                if (in_town && dq3_scene_tile_transition(cur, cur->px, cur->py,
                                                         &dcty, &dsec, &dx, &dy)) {
                    if (dsec == 0xff) {               /* 出城回地表 */
                        if (town) { dq3_scene_free(town); town = NULL; }
                        cur = field; in_town = 0; cur_cty = -1; dq3_scene_apply_palette(cur);
                        fprintf(stderr, "出城:CTY%d → 回地表\n", cur_cty);
                    } else {                          /* 換 section / 跨 CTY */
                        int cross = (dcty != cur_cty && dcty < 100);
                        int ncty = cross ? dcty : cur_cty;
                        dq3_scene *ns = NULL;
                        if (ncty >= 0) {
                            char cty[16]; int bn = dq3x_map_blknum[ncty];
                            sprintf(cty, "CTY%02d.DAT", ncty);
                            ns = dq3_town_load(assets, cty, dsec, bn, err, sizeof err);
                        }
                        if (ns) {
                            if (town) dq3_scene_free(town);
                            town = ns; cur = town; cur_cty = ncty; load_field_hero(town, assets);
                            if (dx < cur->map_w) cur->px = dx;
                            if (dy < cur->map_h) cur->py = dy;
                            dq3_scene_apply_palette(cur);
                            fprintf(stderr, "門/階梯:→ CTY%d section%d spawn(%d,%d)%s\n",
                                    ncty, dsec, dx, dy, cross ? " [跨CTY]" : "");
                        } else {                      /* 載入失敗 → 回地表保底 */
                            if (town) { dq3_scene_free(town); town = NULL; }
                            cur = field; in_town = 0; cur_cty = -1; dq3_scene_apply_palette(cur);
                            fprintf(stderr, "轉場載入失敗 CTY%d sec%d → 回地表: %s\n", ncty, dsec, err);
                        }
                    }
                }
                /* 地表走到城鎮入口座標 → 進該 CTY(load_cty 查表 0x748)*/
                if (!in_town) {
                    int cidx = find_cty_at(cur->px, cur->py);
                    if (cidx >= 0) {
                        char cty[16]; int bn = dq3x_map_blknum[cidx];   /* 每CTY BLK號(0x0a04)*/
                        sprintf(cty, "CTY%02d.DAT", cidx);
                        town = dq3_town_load(assets, cty, 0, bn, err, sizeof err);
                        if (town) { load_field_hero(town, assets); cur = town; in_town = 1; cur_cty = cidx;
                            dq3_scene_apply_palette(cur);
                            fprintf(stderr, "入城:地表(%d,%d) → %s(BLK%d)事件數=%d%s\n",
                                    cur->px, cur->py, cty, bn, town->n_events,
                                    cidx==DQ3_SHRINE_CTY ? "  [合成祠堂]" : ""); }
                        else fprintf(stderr, "入城失敗 %s: %s\n", cty, err);
                    }
                }
                if (moved && !in_town && --enc <= 0) {
                    int reg = dq3_encounter_region(cur->px, cur->py);   /* 位置→region(docs/39)*/
                    int mon = dq3_encounter_pick(reg, grnd());
                    if (mon < 0) mon = over_pool[grnd() % 4];
                    dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                    dq3_battlescene_run(assets, mon, 1 + (int)(grnd()%3), field_bg_page(cur), NULL, NULL, grnd());
                    gold += dq3_battlescene_last_gold();   /* 戰利金錢入袋 */
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
    while(!dq3_should_quit()){ if(dq3_poll_scancode()==DQ3_SC_F10) break; dq3_present(); dq3_delay_ms(16); }
    dq3_text_free(&t); return 0;
}

/* 露依達酒場創角:完整流程狀態機(dq3_tavern:職業→姓名→性別→登錄→名冊)。
 * headless:DQ3_TAV_KEYS 餵按鍵(U/D/L/R 方向、E=Enter、S=Space)後 dump。 */
static void tav_window(uint8_t *fb, int wx, int wy, int ww, int wh, uint8_t black, uint8_t frame, uint8_t bg)
{
    int r, c;
    memset(fb, black, (size_t)DQ3_SCREEN_W * DQ3_SCREEN_H);
    for (r = 0; r < wh; r++) for (c = 0; c < ww; c++) { int yy=wy+r,xx=wx+c;
        if (yy>=0&&yy<DQ3_SCREEN_H&&xx>=0&&xx<DQ3_SCREEN_W) fb[yy*DQ3_SCREEN_W+xx]=bg; }
    for (c = 0; c < ww; c++) { fb[wy*DQ3_SCREEN_W+wx+c]=frame; fb[(wy+wh-1)*DQ3_SCREEN_W+wx+c]=frame; }
    for (r = 0; r < wh; r++) { fb[(wy+r)*DQ3_SCREEN_W+wx]=frame; fb[(wy+r)*DQ3_SCREEN_W+wx+ww-1]=frame; }
}

static int run_tavern(const char *assets, const char *dump)
{
    char err[256] = {0};
    dq3_color pal[256]; int pn;
    uint8_t *raw; size_t rl;
    dq3_text t; dq3_stats st; dq3_roster roster; dq3_party party; dq3_tavern tv;
    uint8_t *fb = dq3_fb();
    int white, black, frame, bg; uint8_t yellow;
    int wx = 40, wy = 60, ww = 220, wh = 200;

    raw = dq3_load_file("DQ3.PAL", &rl);
    if (!raw) { fprintf(stderr, "DQ3.PAL\n"); return 7; }
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal, pn, 255,255,255); black = pal_near2(pal, pn, 0,0,0);
    frame = pal_near2(pal, pn, 255,255,255); bg = pal_near2(pal, pn, 16,16,32);
    yellow = (uint8_t)pal_near2(pal, pn, 255,255,0);

    if (dq3_text_load(&t, assets, "D3TXT00.TXT", err, sizeof err) != 0) {
        fprintf(stderr, "text: %s\n", err); return 7; }
    dq3_stats_load(&st, assets, 1, NULL, 0);
    dq3_roster_init(&roster); dq3_party_init(&party);
    dq3_tavern_init(&tv, &roster, &party, &st);

    { const char *keys = getenv("DQ3_TAV_KEYS"); const char *p;
      if (keys) for (p = keys; *p; p++) { uint8_t sc = 0;
          switch (*p) { case 'U':sc=0x48;break; case 'D':sc=0x50;break; case 'L':sc=0x4b;break;
                        case 'R':sc=0x4d;break; case 'E':sc=0x1c;break; case 'S':sc=0x39;break; }
          if (sc) dq3_tavern_input(&tv, sc); } }

    if (getenv("DQ3_TAVERN_SCREEN") && atoi(getenv("DQ3_TAVERN_SCREEN")) == 3 && dump) {
        /* 注音輸入畫面驗證:組「ㄕ」+「ˋ」→ 候選窗(挑「是」)*/
        dq3_zhuyin z; const char *keys = getenv("DQ3_ZH_KEYS"); int phase2;
        dq3_zhuyin_init(&z);
        z.cursor = 16; dq3_zhuyin_input(&z, 0x1c);   /* ㄕ */
        z.cursor = 39; dq3_zhuyin_input(&z, 0x1c);   /* ˋ → 候選 */
        phase2 = (keys && keys[0]=='C');              /* DQ3_ZH_KEYS=C → 看候選窗;否則看注音盤 */
        if (!phase2) dq3_zhuyin_init(&z);             /* 重置看注音盤全貌 */
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_zhuyin_render(&z, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+12, wy+12, (uint8_t)white, yellow);
        dq3_present();
        if (dq3_dump_ppm(dump)==0) fprintf(stderr, "tavern(注音 phase=%d) -> %s OK\n", z.phase, dump);
        dq3_text_free(&t); return 0;
    }
    if (getenv("DQ3_TAVERN_SCREEN") && atoi(getenv("DQ3_TAVERN_SCREEN")) == 9 && dump) {
        /* F10 離開確認視窗 視覺驗證 */
        static const uint16_t TT[5] = {502,488,113,689,534}; dq3_menu cm; int i;
        int cwx=180,cwy=120,cww=200,cwh=110;
        dq3_menu_init(&cm, cwx+24, cwy+50);
        { uint16_t y1[1]={399}, n1[1]={678}; dq3_menu_add(&cm,y1,1); dq3_menu_add(&cm,n1,1); cm.cursor=1; }
        tav_window(fb, cwx, cwy, cww, cwh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        for (i=0;i<5;i++) dq3_text_draw_glyph(&t,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,cwx+16+i*DQ3_GLYPH_PX,cwy+16,TT[i],(uint8_t)white);
        dq3_menu_render(&cm,&t,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,(uint8_t)white,yellow);
        dq3_present();
        if (dq3_dump_ppm(dump)==0) fprintf(stderr,"tavern(F10 離開確認) -> %s OK\n", dump);
        dq3_text_free(&t); return 0;
    }
    if (getenv("DQ3_SHOP_SCREEN") && dump) {
        /* 商店畫面驗證:建一名隊員 + 金錢,渲染武器/防具店首幀(DQ3_SHOP_DUMP)*/
        static const uint16_t nm[2] = {106,187}; dq3_items shi; long g = 500;
        dq3_roster_create(&roster, &st, getenv("DQ3_ST_CLASS")?atoi(getenv("DQ3_ST_CLASS")):0, 0, nm, 2);
        dq3_party_add(&party, &roster, 0);
        if (dq3_items_load(&shi, assets, NULL, 0) == 0) {
            setenv("DQ3_SHOP_DUMP", dump, 1);
            shop_modal(&roster, &party, &shi, &t, &g);
        }
        dq3_text_free(&t); return 0;
    }
    if (getenv("DQ3_CMD_SCREEN") && dump) {
        /* 野外指令窗 視覺驗證(命令 + 6 指令 + ► 游標)*/
        dq3_cmdmenu cm; dq3_cmdmenu_init(&cm);
        cm.cursor = getenv("DQ3_CMD_CUR") ? atoi(getenv("DQ3_CMD_CUR")) : 0;
        tav_window(fb, 40, 60, 200, 96, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_cmdmenu_render(&cm, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, 52, 70, (uint8_t)white, yellow);
        dq3_present();
        if (dq3_dump_ppm(dump)==0) fprintf(stderr, "cmdmenu(cur=%d) -> %s OK\n", cm.cursor, dump);
        dq3_text_free(&t); return 0;
    }
    if (getenv("DQ3_ITEM_SCREEN") && dump) {
        /* 道具欄畫面 視覺驗證(填數個範例道具)*/
        dq3_inventory iv; dq3_inv_init(&iv);
        dq3_inv_add(&iv, ITEM_KEY_FINAL); dq3_inv_add(&iv, ITEM_RAINBOW_DROP);
        dq3_inv_add(&iv, ITEM_SUN_STONE); dq3_inv_add(&iv, 0x41 /*藥草*/);
        dq3_inv_add(&iv, 0x6e /*飛鷹劍*/);
        tav_window(fb, 30, 50, 290, 200, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_status_render_items(&iv, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, 30+14, 50+12, (uint8_t)white);
        dq3_present();
        if (dq3_dump_ppm(dump)==0) fprintf(stderr, "items -> %s OK\n", dump);
        dq3_text_free(&t); return 0;
    }
    if (getenv("DQ3_STATUS_SCREEN") && dump) {
        /* つよさ 能力值畫面驗證:建一名範例隊員(職業/等級可由 env 指定)*/
        dq3_recruit demo; int cls = getenv("DQ3_ST_CLASS") ? atoi(getenv("DQ3_ST_CLASS")) : 0;
        int lv = getenv("DQ3_ST_LEVEL") ? atoi(getenv("DQ3_ST_LEVEL")) : 20;
        static const uint16_t nm[3] = {106, 187, 0};   /* 「勇者」當範例名 */
        if (cls < 0 || cls > 7) cls = 0; if (lv < 1) lv = 1; if (lv > 43) lv = 43;
        memset(&demo, 0, sizeof demo);
        demo.name[0] = nm[0]; demo.name[1] = nm[1]; demo.name_len = 2; demo.gender = 0;
        demo.m.cls = cls; dq3_member_init(&demo.m, &st, cls, lv);
        tav_window(fb, 24, wy, 300, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        if (getenv("DQ3_ST_PAGE") && getenv("DQ3_ST_PAGE")[0]=='s')
            dq3_status_render_spells(&demo, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, 24+14, wy+12, (uint8_t)white);
        else
            dq3_status_render(&demo, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, 24+14, wy+12, (uint8_t)white);
        dq3_present();
        if (dq3_dump_ppm(dump)==0)
            fprintf(stderr, "status(cls=%d lv=%d HP=%u MP=%u) -> %s OK\n",
                    cls, demo.m.level, demo.m.stat[DQ3_STAT_HP], demo.m.stat[DQ3_STAT_MP], dump);
        dq3_text_free(&t); return 0;
    }
    if (dump) {
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_tavern_render(&tv, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+12, wy+12, (uint8_t)white, yellow);
        dq3_present();
        if (dq3_dump_ppm(dump)==0)
            fprintf(stderr, "tavern(state=%d roster=%d party=%d) -> %s OK\n", tv.state, roster.count, party.count, dump);
        dq3_text_free(&t); return 0;
    }
    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == DQ3_SC_F10) break;        /* F10 離開(standalone tavern demo)*/
        if (sc) dq3_tavern_input(&tv, sc);
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_tavern_render(&tv, &t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+12, wy+12, (uint8_t)white, yellow);
        dq3_present(); dq3_delay_ms(16);
    }
    dq3_text_free(&t); return 0;
}

/* 遊戲內露依達酒場 modal:用已載入的 font(text)跑 dq3_tavern 流程,ESC 離開。
 * 建立的隊員存進傳入的 roster/party(持久);會改 DAC palette,呼叫端離開後須重套場景色盤。 */
static void tavern_modal(const char *assets, dq3_roster *roster, dq3_party *party,
                         const dq3_stats *st, const dq3_text *text)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t yellow; dq3_tavern tv;
    uint8_t *fb = dq3_fb();
    int wx = 40, wy = 60, ww = 220, wh = 200;

    raw = dq3_load_file("DQ3.PAL", &rl);
    if (!raw) return;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = pal_near2(pal,pn,255,255,255); bg = pal_near2(pal,pn,16,16,32);
    yellow = (uint8_t)pal_near2(pal,pn,255,255,0);

    dq3_tavern_init(&tv, roster, party, st);
    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == 0x01 || sc == DQ3_SC_F10) break;  /* ESC / F10 離開酒場(回遊戲)*/
        if (sc) dq3_tavern_input(&tv, sc);
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_tavern_render(&tv, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+12, wy+12, (uint8_t)white, yellow);
        dq3_present(); dq3_delay_ms(16);
    }
}

/* 用字模畫十進位數(數字 glyph 0..9 = idx 0..9)。 */
static void draw_number_at(uint8_t *fb, const dq3_text *t, int x, int y, int v, uint8_t fg)
{
    uint16_t d[8]; int n = 0, i; if (v < 0) v = 0;
    if (v == 0) d[n++] = 0;
    else { int x2 = v; uint16_t tmp[8]; int tn = 0;
        while (x2 > 0 && tn < 8) { tmp[tn++] = (uint16_t)(x2 % 10); x2 /= 10; }
        for (i = tn-1; i >= 0; i--) d[n++] = tmp[i]; }
    for (i = 0; i < n; i++) dq3_text_draw_glyph(t, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, x + i*DQ3_GLYPH_PX, y, d[i], fg);
}

/* 武器/防具商店 modal:列早期裝備(名+價+能否裝備),Enter 購買→裝給領頭隊員,ESC 離開。
 * 道具名 = D3TXT00 rec=code+1;價/類別/可裝備職業由 ITEM.DAT(docs/22)。 */
static void shop_modal(dq3_roster *roster, dq3_party *party, const dq3_items *items, const dq3_text *text, long *gold)
{
    static const unsigned char STOCK[] = {0x01,0x03,0x0b,0x11,0x21,0x27,0x2b,0x41};
    int n = (int)sizeof STOCK, cur = 0, lead_ri;
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t yellow, redc; uint8_t *fb = dq3_fb();
    int wx = 24, wy = 40, ww = 300, wh = 220;

    lead_ri = (party->count > 0 && party->slot[0] >= 0 && party->slot[0] < roster->count) ? party->slot[0] : -1;
    if (lead_ri < 0) return;                                  /* 隊伍空 */
    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32);
    yellow = (uint8_t)pal_near2(pal,pn,255,255,0); redc = (uint8_t)pal_near2(pal,pn,255,80,80);

    { const char *sd = getenv("DQ3_SHOP_DUMP"); int i; dq3_recruit *rc = &roster->list[lead_ri];
      if (sd) {
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        draw_number_at(fb, text, wx+14, wy+10, (int)*gold, (uint8_t)yellow);
        for (i = 0; i < n; i++) { int yy = wy+34+i*18, code = STOCK[i];
            int ok = (code==0x41 || dq3_item_can_equip(items, code, rc->m.cls));
            if (i==0) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+4, yy, 11, yellow);
            dq3_text_draw_record(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+22, yy, 8, 1, code+1, ok?(uint8_t)white:redc);
            draw_number_at(fb, text, wx+22+7*DQ3_GLYPH_PX, yy, dq3_item_price(items, code), (uint8_t)white); }
        dq3_present(); dq3_dump_ppm(sd); fprintf(stderr,"shop(金%ld)-> %s\n",*gold,sd); return; } }

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        dq3_recruit *rc = &roster->list[lead_ri];
        int i;
        if (sc == 0x01 || sc == DQ3_SC_F10) break;
        else if (sc == 0x48) cur = (cur + n - 1) % n;
        else if (sc == 0x50) cur = (cur + 1) % n;
        else if (sc == 0x1c || sc == 0x39) {                  /* Enter/Space:購買 */
            int code = STOCK[cur], price = dq3_item_price(items, code);
            int cat = dq3_item_category(items, code);
            if (*gold >= price && (code == 0x41 || dq3_item_can_equip(items, code, rc->m.cls))) {
                *gold -= price;
                if (cat & 0x40) rc->armor = (unsigned char)code;        /* 防具 */
                else if (cat & 0x20) rc->weapon = (unsigned char)code;  /* 武器 */
                fprintf(stderr, "購買並裝備 道具0x%02x(花 %d,餘 %ld)\n", code, price, *gold);
            } else fprintf(stderr, "金錢不足或無法裝備。\n");
        }
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        draw_number_at(fb, text, wx+14, wy+10, (int)*gold, (uint8_t)yellow);   /* 金錢 */
        for (i = 0; i < n; i++) {
            int yy = wy + 34 + i * 18, code = STOCK[i];
            uint8_t fg = (i == cur) ? yellow : (uint8_t)white;
            int ok = (code == 0x41 || dq3_item_can_equip(items, code, rc->m.cls));
            if (i == cur) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+4, yy, 11, yellow);
            dq3_text_draw_record(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+22, yy, 8, 1, code+1, ok?fg:redc);
            draw_number_at(fb, text, wx+22+7*DQ3_GLYPH_PX, yy, dq3_item_price(items, code), (uint8_t)white);
        }
        dq3_present(); dq3_delay_ms(16);
    }
}

/* 遊戲內 つよさ 能力值 modal:顯示在隊伍中的隊員數值,←/→ 切換,Space/Enter 能力↔咒文,ESC 離開。
 * start_page:0=能力 1=咒文。 */
static void status_modal_page(const dq3_roster *roster, const dq3_party *party, const dq3_text *text, int start_page)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t *fb = dq3_fb();
    int wx = 24, wy = 50, ww = 300, wh = 210, cur = 0, n = 0, i, page = start_page;
    int members[DQ3_PARTY_MAX];

    for (i = 0; i < DQ3_PARTY_MAX; i++)
        if (party->slot[i] >= 0 && party->slot[i] < roster->count) members[n++] = party->slot[i];
    if (n == 0) return;   /* 隊伍空,不開 */

    raw = dq3_load_file("DQ3.PAL", &rl);
    if (!raw) return;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = pal_near2(pal,pn,255,255,255); bg = pal_near2(pal,pn,16,16,32);

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == 0x01 || sc == DQ3_SC_F10) break;            /* ESC / F10 離開 */
        if (sc == 0x4b) cur = (cur + n - 1) % n;              /* ← 上一名 */
        else if (sc == 0x4d) cur = (cur + 1) % n;             /* → 下一名 */
        else if (sc == 0x39 || sc == 0x1c) page ^= 1;         /* Space/Enter:能力↔咒文 切換 */
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        if (page == 0)
            dq3_status_render(&roster->list[members[cur]], text, fb,
                              DQ3_SCREEN_W, DQ3_SCREEN_H, wx+14, wy+12, (uint8_t)white);
        else
            dq3_status_render_spells(&roster->list[members[cur]], text, fb,
                              DQ3_SCREEN_W, DQ3_SCREEN_H, wx+14, wy+12, (uint8_t)white);
        dq3_present(); dq3_delay_ms(16);
    }
}

/* 野外指令窗(命令):rec400 的 6 指令 對話/咒文/狀況/道具/裝備/調查;選定即派發。
 * 對話/調查 = 面向格事件→對話(祠堂/酒場特例仍在 Enter);道具/裝備 尚未實作。回傳選定指令或 -1。 */
/* 道具欄 modal:列出持有道具(名稱 = D3TXT00 rec=code+1),ESC 離開。 */
static void item_modal(const dq3_inventory *inv, const dq3_text *text)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t *fb = dq3_fb();
    int wx = 30, wy = 50, ww = 290, wh = 200;
    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32);
    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == 0x01 || sc == DQ3_SC_F10) break;
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_status_render_items(inv, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+14, wy+12, (uint8_t)white);
        dq3_present(); dq3_delay_ms(16);
    }
}

static int cmd_modal(dq3_scene *scene, dq3_roster *roster, dq3_party *party,
                     dq3_inventory *inv, dq3_dialogue *dlg, int dlg_ok)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t yellow; uint8_t *fb = dq3_fb();
    int wx = 40, wy = 60, ww = 200, wh = 96, sel = -1;
    dq3_cmdmenu cm; dq3_cmdmenu_init(&cm);

    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return -1;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32); yellow = (uint8_t)pal_near2(pal,pn,255,255,0);

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        int r = sc ? dq3_cmdmenu_input(&cm, sc) : -1;
        if (r == -2) return -1;                   /* ESC 取消 */
        if (r >= 0) { sel = r; break; }
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_cmdmenu_render(&cm, &dlg->txt, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+12, wy+10, (uint8_t)white, yellow);
        dq3_present(); dq3_delay_ms(16);
    }
    if (sel < 0) return -1;
    switch (sel) {
    case DQ3_CMD_STATUS: status_modal_page(roster, party, &dlg->txt, 0); break;
    case DQ3_CMD_SPELL:  status_modal_page(roster, party, &dlg->txt, 1); break;
    case DQ3_CMD_TALK: case DQ3_CMD_EXAMINE: {
        int fdx = (scene->facing==3)-(scene->facing==1), fdy = (scene->facing==0)-(scene->facing==2);
        int fx = scene->px+fdx, fy = scene->py+fdy, et, ep;
        if (dq3_scene_tile_event(scene, fx, fy, &et, &ep) ||
            dq3_scene_tile_event(scene, scene->px, scene->py, &et, &ep)) {
            if (dlg_ok) { int rec = (ep && ep < dlg->txt.n_records) ? ep : 1; dq3_dialogue_open(dlg, rec); }
        } else fprintf(stderr, "%s:前方無人 / 無物\n", sel==DQ3_CMD_TALK?"對話":"調查");
        break; }
    case DQ3_CMD_ITEM:  item_modal(inv, &dlg->txt); break;   /* 共享道具欄(鑰匙/合成品…)*/
    case DQ3_CMD_EQUIP: fprintf(stderr, "裝備:裝備系統尚未實作\n"); break;
    }
    return sel;
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
            int entry = he ? atoi(he) : 0;    /* 預設勇者 = DQ3MST.BLS entry0(c0);見 load_field_hero */
            int order[4] = {0,1,2,3}; const int *op = NULL;
            if (fo && sscanf(fo, "%d,%d,%d,%d", &order[0],&order[1],&order[2],&order[3]) == 4) op = order;
            if (dq3_scene_load_hero(s, assets, entry, op) != 0)
                fprintf(stderr, "hero sprite load failed (entry %d) — 用佔位框\n", entry);
            dq3_scene_load_npc_sprites(s, assets);   /* NPC sprite(town 模式)*/
            rc = run_scene(s, dump);
            dq3_scene_free(s);
        }
    } else if (strcmp(mode, "battle") == 0) {
        rc = run_battle(assets, dump);
    } else if (strcmp(mode, "game") == 0) {
        rc = run_game(assets, dump);
    } else if (strcmp(mode, "text") == 0) {
        rc = run_text(assets, dump);
    } else if (strcmp(mode, "tavern") == 0) {
        rc = run_tavern(assets, dump);
    } else {
        const char *title = (argc > 3) ? argv[3]
                          : (strcmp(mode, "title") == 0 ? "TITG.P" : mode);
        rc = run_title(assets, title, dump);
    }

    dq3_rt_quit();
    return rc;
}
