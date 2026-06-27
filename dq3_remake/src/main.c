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
#include "dq3_spell.h"      /* 野外咒文施放:dq3_spells_known(列已學咒)*/
#include "dq3_menu.h"
#include "dq3_nameinput.h"
#include "dq3_tavern.h"
#include "dq3_zhuyin.h"
#include "dq3_save.h"
#include "dq3_status.h"
#include "dq3_cmdmenu.h"
#include "dq3_encounter.h"
#include "dq3_combat.h"
#include "dq3_shopdata.h"
#include "dq3_sub2.h"
#include "dq3_warp.h"
#include "dq3_owportal.h"
#include "dq3_progress.h"
#include "dq3_ship.h"
#include "dq3_rng.h"
#include "dq3_config.h"
#include "dq3_customglyph.h"   /* 自建字形(設/版)— 設定選單標籤 */
#include "dq3_scripted.h"
#include "dq3_item_use.h"

/* 可攜 setenv:Windows CRT 無 POSIX setenv,改 _putenv_s。只用於 debug dump 路徑。 */
#ifdef _WIN32
static int dq3_setenv(const char *k, const char *v) { return _putenv_s(k, v); }
#else
static int dq3_setenv(const char *k, const char *v) { return setenv(k, v, 1); }
#endif

/* 取船劇情(#2 真實 NPC 觸發,docs/50)。波魯多加 = CTY37(throne room,overworld (26,72));
 * 國王 = section0 (9,6) sub2 scripted-event NPC,byte4=26。對話分支(資料驅動,RE 確鑿):
 *   無胡椒 → rec26「怎麼了?我在等黑胡椒。」 / 給胡椒 → rec28「胡椒太好吃了…好想睡。」+ 授船。
 * 黑胡椒 = item 0x5c(rec93,glyph 比對確認);港口海格 = overworld (25,73)(城南鄰)。 */
#define DQ3_PORTOGA_CTY     37
#define DQ3_PORTOGA_KING_X   9
#define DQ3_PORTOGA_KING_Y   6
#define DQ3_PORTOGA_HARBOR_X 25
#define DQ3_PORTOGA_HARBOR_Y 73
#define DQ3_ITEM_PEPPER    0x5c   /* 黑胡椒(獻國王換船的劇情物)*/
/* 盜賊鑰匙:拿吉米之塔(精訊中文名;ナジミの塔)=CTY8,4F(sec3)老人 (9,9) sub2 對話 → 給盜賊鑰匙。 */
#define DQ3_NAJIMI_CTY      8
#define DQ3_NAJIMI_OLDMAN_X 9
#define DQ3_NAJIMI_OLDMAN_Y 9
#define DQ3_ITEM_THIEF_KEY  0x55
/* 魔法玉(攻略亦稱魔法球,item 0x58):雷貝村(CTY1)2F(sec1)老人,持盜賊鑰匙開門上樓對話給。
 * 用來炸誘人洞窟的牆。原版 runner flag 分支;remake 接:雷貝 2F sub2 老人 + 持鑰匙 → 給。 */
#define DQ3_LEVE_CTY        1
#define DQ3_LEVE_2F_SEC     1
#define DQ3_ITEM_MAGIC_BALL 0x58
/* 羅馬利亞皇冠線:CTY2 sec1 王座廳國王 (7,2) sub2。香巴尼塔(CTY10)打甘達特取金皇冠(0x33,
 * examine 既有系統可拿)→ 持皇冠對羅馬利亞國王 → 觸發讓位(攻略建議婉拒,保留皇冠)→ ROMALY 里程碑。 */
#define DQ3_ROMALY_CTY      2
#define DQ3_ROMALY_KING_SEC 1
#define DQ3_ROMALY_KING_X   7
#define DQ3_ROMALY_KING_Y   2
#define DQ3_ITEM_CROWN      0x33
/* 達瑪神殿=CTY49(取船後跨海到)。進神殿 → DHAMA 里程碑(轉職實際換職另為系統,docs/47 C11)。 */
#define DQ3_DHAMA_CTY       49
#define DQ3_PEPPER_CTY     15     /* 胡椒販售城:提示 NPC(4,15)「下方的店裡有賣黑胡椒」。
                                   * 但資料裡無一店進此貨(早期 build 斷鏈,docs/50)→ remake 補進該城道具店。 */
#define DQ3_PORTOGA_REC_WAIT 26   /* 「我在等黑胡椒」*/
#define DQ3_PORTOGA_REC_GOT  28   /* 「胡椒太好吃了…好想睡」(已收胡椒)*/
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

/* scripted_event 86 下降(ギアガの大穴;RE handler 0x783d / [0x5051]=3,docs/44):
 * 切到下層 overworld(field_under,懶載),置玩家於下層城 CTY77 入口附近,設 DQ3_FLAG_DESCENDED。
 * 原版觸發是 runner 劇情事件;remake 由 debug 口 / U 鍵代觸發。回 0 成功。 */
static int do_descent(const char *assets, dq3_scene **field_under, dq3_scene **cur,
                      int *layer, int *in_town, int *cur_cty, dq3_storyflags *flags)
{
    char err[256] = {0};
    if (!*field_under) { *field_under = dq3_field_load_map(assets, "DQ3UND.MAP", err, sizeof err);
        if (*field_under) load_field_hero(*field_under, assets); }
    if (!*field_under) { fprintf(stderr, "下降:下層 overworld 載入失敗: %s\n", err); return -1; }
    *layer = 1; *cur = *field_under; *in_town = 0; *cur_cty = -1;
    (*cur)->px = 84; (*cur)->py = 68;                 /* 下層城 CTY77 入口附近(可走,driven）*/
    if (flags) dq3_flags_set(flags, DQ3_FLAG_DESCENDED, 1);
    dq3_scene_apply_palette(*cur);
    fprintf(stderr, "scripted_event 86 下降 → 下層 overworld (%d,%d)\n", (*cur)->px, (*cur)->py);
    return 0;
}

/* 讀檔後還原玩家所在場景到原位置(隨時存讀檔 + 標題續玩共用,使用者需求)。
 * pos.in_town → 載入城鎮 CTY=cty section=sec;否則 overworld(layer 0=地表 / 1=下層,懶載下層)。
 * 還原後設玩家 (px,py)。回 0=成功。仿 do_descent 的指標傳遞。 */
static int restore_position(const char *assets, const dq3_save_pos *pos,
                            dq3_scene *field, dq3_scene **field_under, dq3_scene **town,
                            dq3_scene **cur, int *layer, int *in_town, int *cur_cty,
                            dq3_storyflags *flags)
{
    char err[256] = {0};
    if (pos->in_town && pos->cty >= 0) {                 /* 城鎮 */
        char ct[16]; int bn = (pos->cty >= 0 && pos->cty < 100) ? dq3x_map_blknum[pos->cty] : 1;
        dq3_scene *ns;
        sprintf(ct, "CTY%02d.DAT", pos->cty);
        ns = dq3_town_load(assets, ct, pos->sec, bn, err, sizeof err);
        if (!ns) { fprintf(stderr, "讀檔場景:CTY%d sec%d 載入失敗: %s\n", pos->cty, pos->sec, err); return -1; }
        if (*town) dq3_scene_free(*town);
        *town = ns; *cur = ns; *in_town = 1; *cur_cty = pos->cty; *layer = 0;
        load_field_hero(ns, assets);
    } else if (pos->layer == 1) {                        /* 下層 overworld(懶載)*/
        if (!*field_under) { *field_under = dq3_field_load_map(assets, "DQ3UND.MAP", err, sizeof err);
            if (*field_under) load_field_hero(*field_under, assets); }
        if (!*field_under) { fprintf(stderr, "讀檔場景:下層載入失敗: %s\n", err); return -1; }
        *cur = *field_under; *layer = 1; *in_town = 0; *cur_cty = -1;
        if (flags) dq3_flags_set(flags, DQ3_FLAG_DESCENDED, 1);
    } else {                                             /* 地表 */
        *cur = field; *layer = 0; *in_town = 0; *cur_cty = -1;
    }
    if (pos->px >= 0 && pos->px < (*cur)->map_w) (*cur)->px = pos->px;
    if (pos->py >= 0 && pos->py < (*cur)->map_h) (*cur)->py = pos->py;
    dq3_scene_apply_palette(*cur);
    fprintf(stderr, "讀檔場景還原:%s CTY%d sec%d (%d,%d)\n",
            pos->in_town ? "城" : (pos->layer ? "下層" : "地表"), pos->cty, pos->sec, (*cur)->px, (*cur)->py);
    return 0;
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
static void shop_modal(dq3_roster *roster, dq3_party *party, const dq3_items *items, const dq3_text *text, long *gold, const unsigned char *STOCK, int n);
static const unsigned char *shop_stock_for(int cty, int *n);
static void equip_modal(dq3_roster *roster, dq3_party *party, const dq3_text *text);
static dq3_inventory *g_equip_inv = NULL;   /* 裝備管理:背包來源 */
static dq3_items *g_equip_items = NULL; static int g_equip_items_ok = 0;  /* ITEM.DAT(category/can_equip)*/
static int  cmd_modal(dq3_scene *scene, dq3_roster *roster, dq3_party *party,
                      dq3_inventory *inv, dq3_dialogue *dlg, int dlg_ok, const dq3_text *sys);
static int item_modal(dq3_inventory *inv, const dq3_text *text, dq3_roster *roster, dq3_party *party);
static int g_item_world_eff = 0;   /* 野外道具選單選了蓋美拉翅膀/聖水 → 交 main 迴圈處理世界狀態 */
static unsigned g_sea_frame = 0;   /* 海面 palette cycling 幀計數 */
/* 晝夜系統(使用者確認原版機制 = 地表步數驅動,非時間):每 DN_PHASE_STEPS 步推進一相位
 * 白天→黃昏→黑夜→黎明→(循環)。相位 → palette 調暗(dq3_scene 內)。
 * 步數:使用者指定「白天→黑夜 = 120 步,中間有黃昏」→ 每相位 60 步(白天→黃昏→黑夜 = 2×60 = 120;
 * 全 4 相位循環 = 240 步)。原版確切步數計數器在多層 handler 鏈未逐指令定位,此為使用者指定值。 */
static int g_dn_step = 0;
#define DN_PHASE_STEPS 60                      /* 每相位步數(白天→黑夜 120 步,使用者指定)*/
static const char *DN_NAME[4] = { "白天", "黃昏", "黑夜", "黎明" };

/* B-7 耶進貝亞倉庫番(CTY76):3 石頭 NPC(b2==40)推到藍白地面目標 → 開密道取乾渴壺。
 * scene-scoped:離城重入(scene reload)→ 石頭 + 此旗標自動重置(使用者要求)。 */
static int g_sokoban_solved = 0;
#define SOKOBAN_CTY 76
static const int SOKOBAN_TGT[3][2] = { {4,5}, {5,5}, {6,5} };   /* 藍白地面 tile 91 位置 */
static void dhama_modal(dq3_roster *roster, dq3_party *party, const dq3_stats *gst, const dq3_text *text, dq3_inventory *inv);
static void tav_window(uint8_t *fb, int wx, int wy, int ww, int wh, uint8_t black, uint8_t frame, uint8_t bg);
static int  pal_near2(const dq3_color *p, int n, int r, int g, int b);

/* 自動存檔:把名冊/隊伍/道具/位置寫成 dq3_save.dat(remake 自有格式)。回 0=成功。
 * 路徑:DQ3_SAVE 環境變數,預設 "dq3_save.dat"(cwd;唯讀 cwd 時用 env 指可寫路徑)。 */
static const char *save_path(void) { return getenv("DQ3_SAVE") ? getenv("DQ3_SAVE") : "dq3_save.dat"; }

/* 多存檔 slot(至少 6 格,使用者需求)。slot 0 = F10 自動存檔(=save_path(),相容舊存檔);
 * slot 1..DQ3_SAVE_SLOTS = 手動存檔 dq3_saveN.dat(記錄點/F10 選格)。
 * 路徑基底沿用 save_path() 的目錄概念:slot0 直接用 save_path();slotN 用 "dq3_saveN.dat"
 * (若 DQ3_SAVE 指定了非預設路徑,slotN 以該路徑加序號後綴,維持可寫目錄一致)。 */
#define DQ3_SAVE_SLOTS 6
static const char *slot_path(int slot)
{
    static char buf[8][512];
    const char *base = save_path();
    if (slot <= 0) return base;                       /* slot 0 = 自動存檔(相容)*/
    if (slot > DQ3_SAVE_SLOTS) slot = DQ3_SAVE_SLOTS;
    {   /* base 去掉 ".dat" 後接 "N.dat";base 無 ".dat" 則直接接 "N.dat" */
        const char *dot = strrchr(base, '.');
        size_t stem = dot ? (size_t)(dot - base) : strlen(base);
        if (stem > 500) stem = 500;
        memcpy(buf[slot], base, stem);
        snprintf(buf[slot] + stem, sizeof buf[slot] - stem, "%d.dat", slot);
    }
    return buf[slot];
}

/* 把對話變數 {V} 的「主角/受話者名」設成隊長(無隊伍時用名冊首位=主角)。
 * 對話渲染遇 VAR0/VAR_ENT 等插值碼時即替換成此名(dq3_text_set_var_name)。 */
static void set_dialogue_hero(const dq3_roster *r, const dq3_party *p)
{
    int ri = (p->count > 0 && p->slot[0] >= 0 && p->slot[0] < r->count) ? p->slot[0]
           : (r->count > 0 ? 0 : -1);
    if (ri >= 0) dq3_text_set_var_name(r->list[ri].name, r->list[ri].name_len);
    else dq3_text_clear_vars();
}

/* 旅社住宿:治活著的隊員 cur_hp/cur_mp 到滿(不復活陣亡者),扣住宿費(inn_cost × 人數,對齊
 * RE 0x86f5 的 mul)。金錢不足則不住。回實際治療人數。 */
static int inn_rest(dq3_roster *r, const dq3_party *p, long *gold, int inn_cost)
{
    int s, healed = 0, members = p->count, cost = inn_cost * (members > 0 ? members : 1);
    if (*gold < cost) { fprintf(stderr, "旅社:金錢不足(需 %d,有 %ld)\n", cost, *gold); return -1; }
    for (s = 0; s < DQ3_PARTY_MAX; s++) {
        int ri = p->slot[s]; dq3_member *m;
        if (ri < 0 || ri >= r->count) continue;
        m = &r->list[ri].m;
        if (m->cur_hp > 0) { m->cur_hp = m->stat[DQ3_STAT_HP]; m->cur_mp = m->stat[DQ3_STAT_MP]; healed++; }
    }
    *gold -= cost;
    fprintf(stderr, "旅社:住宿(費 %d,治療 %d 人)→ 金錢 %ld\n", cost, healed, *gold);
    return healed;
}

/* 教會復活費 = level 表(★ 靜態 RE:handler 0x85ff,表 @ DGROUP 0x3c6c / file 0x19dac;
 * index = min(level,40)−1,word LE)。前版「level×10」是近似——Lv1-3 應同為 10(非 10/20/30)、
 * 高等差很大(Lv40 RE=1610 vs ×10=400)。表 = Lv1..Lv40 收費。 */
static const unsigned short DQ3_REVIVE_COST[40] = {
     10,  10,  10,  20,  30,  40,  50,  70,  90, 110,
    130, 150, 170, 200, 230, 260, 290, 330, 370, 410,
    450, 490, 530, 580, 630, 680, 730, 790, 850, 910,
    970,1030,1090,1160,1230,1300,1370,1450,1530,1610,
};
static int dq3_revive_cost(int level)              /* 對齊 0x8650:cmp lv,0x28; 上限 40;index=lv−1 */
{
    int lv = level < 1 ? 1 : (level > 40 ? 40 : level);
    return DQ3_REVIVE_COST[lv - 1];
}

/* 教會復活(蘇生,對齊 RE 0x85ff「誰需要復活」):陣亡隊員(cur_hp==0)→ 復原 cur_hp/cur_mp,
 * 收費 = dq3_revive_cost(level)(RE level 表)。金錢不足者跳過。回實際復活人數。
 * 註:原版另有解毒(0x849b,5 固定)/解詛咒(0x853c),remake 解狀態仍沿用免費小額處理。 */
static int church_revive(dq3_roster *r, const dq3_party *p, long *gold)
{
    int s, revived = 0;
    for (s = 0; s < DQ3_PARTY_MAX; s++) {
        int ri = p->slot[s]; dq3_member *m; int cost;
        if (ri < 0 || ri >= r->count) continue;
        m = &r->list[ri].m;
        if (m->status) {                               /* 教會解毒/解麻痺(免費,對齊原版小額)*/
            m->status = 0;
            fprintf(stderr, "教會:解除 Lv%d 隊員的異常狀態\n", m->level);
        }
        if (m->cur_hp != 0) continue;                  /* 只復活陣亡者 */
        cost = dq3_revive_cost(m->level);              /* ★ RE level 表(取代 level×10)*/
        if (*gold < cost) { fprintf(stderr, "教會:金錢不足,無法復活 Lv%d 隊員(需 %d)\n", m->level, cost); continue; }
        *gold -= cost; m->cur_hp = m->stat[DQ3_STAT_HP]; m->cur_mp = m->stat[DQ3_STAT_MP];
        revived++;
        fprintf(stderr, "教會:復活 Lv%d 隊員(費 %d)→ 金錢 %ld\n", m->level, cost, *gold);
    }
    return revived;
}

/* 存檔到指定 path(slot)。slot 0=自動存檔路徑、1..6=手動 slot。
 * in_town/layer/sec 記玩家所在場景,讀檔可完整還原回原位置(使用者需求)。 */
static int save_game_to(const char *path, const dq3_roster *r, const dq3_party *p,
                        const dq3_inventory *inv, const dq3_storyflags *flags, int cty, int px, int py, const dq3_ship *ship,
                        int in_town, int layer, int sec)
{
    dq3_save_pos pos; int rc;
    pos.cty = cty; pos.px = px; pos.py = py;
    pos.ship_owned = ship->owned; pos.ship_aboard = ship->aboard;       /* #2 船狀態 */
    pos.ship_px = ship->px; pos.ship_py = ship->py; pos.ship_layer = ship->layer;
    pos.in_town = in_town; pos.layer = layer; pos.sec = sec;            /* 場景還原 */
    pos.daynight = dq3_scene_get_daynight();                            /* v6:晝夜相位 */
    rc = dq3_save_write(path, r, p, inv, flags, pos);
    if (rc == 0) fprintf(stderr, "存檔 → %s(名冊%d 隊伍%d %s CTY%d sec%d (%d,%d))\n",
                         path, r->count, p->count, in_town ? "城" : (layer ? "下層" : "地表"), cty, sec, px, py);
    else fprintf(stderr, "存檔失敗(無法寫 %s)\n", path);
    return rc;
}

/* 消耗品使用(#3,docs/49)。對隊伍套效果並消耗道具:HEAL_HP 治第一個受傷隊員;
 * 其餘(回鎮/驅敵/解毒/解麻痺)回傳種類交呼叫端處理世界層狀態。回效果種類(DQ3_USE_*),
 * 道具不在背包或不可用回 DQ3_USE_NONE。 */
static int apply_item_use(dq3_inventory *inv, dq3_roster *r, dq3_party *p, int item_id)
{
    int kind = dq3_item_use_kind(item_id);
    if (dq3_inv_find(inv, item_id) < 0) return DQ3_USE_NONE;   /* 沒這個道具 */
    if (kind == DQ3_USE_HEAL_HP) {
        int i, healed = -1;
        for (i = 0; i < p->count; i++) {                       /* 第一個未滿且未陣亡的隊員 */
            dq3_member *m = &r->list[p->slot[i]].m;
            int h = dq3_item_use_heal(m, item_id);
            if (h > 0) { healed = h; fprintf(stderr, "藥草:隊員%d 回復 HP %d → %d/%d\n",
                                              i, h, m->cur_hp, m->stat[DQ3_STAT_HP]); break; }
        }
        if (healed <= 0) { fprintf(stderr, "藥草:無人需要治療\n"); return DQ3_USE_NONE; }
        dq3_inv_remove(inv, item_id);                          /* 消耗 */
        return DQ3_USE_HEAL_HP;
    }
    if (kind == DQ3_USE_CURE_POISON || kind == DQ3_USE_CURE_PARALYSIS) {
        int i, cured = -1;
        for (i = 0; i < p->count; i++)                /* 第一個有對應異常的隊員 */
            if (dq3_item_use_cure(&r->list[p->slot[i]].m, item_id)) { cured = i; break; }
        if (cured < 0) {                              /* 無人對應異常 → 不消耗(對齊原版)*/
            fprintf(stderr, "%s:目前無人需要解(無對應異常),未消耗\n",
                    kind == DQ3_USE_CURE_POISON ? "驅毒草" : "滿月草");
            return DQ3_USE_NONE;
        }
        dq3_inv_remove(inv, item_id);                 /* 消耗 */
        fprintf(stderr, "%s:隊員%d 解除%s\n", kind == DQ3_USE_CURE_POISON ? "驅毒草" : "滿月草",
                cured, kind == DQ3_USE_CURE_POISON ? "中毒" : "麻痺");
        return kind;
    }
    if (kind == DQ3_USE_PRAYER_RING) {         /* 祈禱之戒:回 MP + 每次使用 ~25.4% 損壞(#7c,RE file 0x5ad0/0x5af4)*/
        static dq3_rng prng; static int seeded = 0;
        int i, restored = -1, broke;
        if (!seeded) { dq3_rng_seed(&prng, 0x1357); seeded = 1; }
        for (i = 0; i < p->count; i++) {                       /* 回復第一個 MP 未滿的隊員 */
            int r2 = dq3_item_use_prayer_mp(&r->list[p->slot[i]].m);
            if (r2 > 0) { restored = r2;
                fprintf(stderr, "祈禱之戒:隊員%d 回復 MP %d → %d/%d\n", i,
                        r2, r->list[p->slot[i]].m.cur_mp, r->list[p->slot[i]].m.stat[DQ3_STAT_MP]); break; }
        }
        if (restored < 0) fprintf(stderr, "祈禱之戒:無人 MP 不足(仍會擲損壞判定,對齊原版)\n");
        /* 損壞判定(忠實 RE):RNG(256) ≤ 0x40 → ~25.4% 損壞消失。原版無「MP 滿則不用」閘。 */
        broke = (dq3_rng_next(&prng, 256) <= DQ3_PRAYER_BREAK_LE);
        if (broke) { dq3_inv_remove(inv, item_id);
            fprintf(stderr, "祈禱之戒:啊,戒指壞了。(損壞消失,~25.4%%)\n"); }
        return DQ3_USE_PRAYER_RING;
    }
    if (kind == DQ3_USE_AWAKEN) return kind;   /* 覺醒粉:位置相關,消耗延到 main(確認在諾阿尼魯)*/
    if (kind == DQ3_USE_RAINBOW) return kind;  /* 彩虹水滴:位置相關,消耗延到 main(確認在下層利姆達爾西北)*/
    if (kind == DQ3_USE_GAIA)   return kind;   /* 蓋亞之劍:武器不消耗,開火山交 main */
    if (kind == DQ3_USE_DRAIN)  return kind;   /* 乾渴壺:不消耗(可重用),吸海交 main */
    if (kind == DQ3_USE_FAIRYFLUTE) return kind; /* 妖精之笛:不消耗,魯比斯解詛咒交 main */
    if (kind == DQ3_USE_NONE) return DQ3_USE_NONE;
    /* 回鎮/驅敵:消耗在此,世界層效果由呼叫端依回傳種類處理。 */
    dq3_inv_remove(inv, item_id);
    return kind;
}

/* 海面 palette cycling(原版 DAC 動畫,docs/04/51):overworld slot2/10 在藍波間慢速循環。
 * sprite bank(slot16+)不動 → 主角膚色不受影響。城鎮不呼叫(無海動畫色)。 */
static void animate_sea(dq3_scene *s, unsigned frame)
{
    static const dq3_color WAVE[4] = { {0,85,223}, {0,105,240}, {24,130,255}, {0,70,205} };
    int p2 = (int)((frame >> 3) & 3), p10 = (p2 + 2) & 3;   /* 每 8 幀換格;slot10 相位差 2 */
    if (s->pal_count <= 10) return;
    s->pal[2] = WAVE[p2]; s->pal[10] = WAVE[p10];
    dq3_scene_apply_palette(s);
}

/* overworld 船 sprite overlay(docs/51):登船 → 玩家格畫船(依 facing);未登船 → 停泊格畫船。
 * 城鎮不畫(船只在 overworld);停泊船僅在同層才畫。在 dq3_scene_render 之後呼叫。 */
static void draw_ship_overlay(const dq3_scene *cur, const dq3_ship *ship, int in_town, int layer)
{
    if (in_town) return;
    if (ship->aboard)
        dq3_scene_draw_tile_at(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H,
                               cur->px, cur->py, dq3_ship_tile_for_facing(cur->facing));
    else if (ship->owned && ship->layer == layer)
        dq3_scene_draw_tile_at(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H,
                               ship->px, ship->py, DQ3_SHIP_TILE_DOWN);
}

/* 結局序列(索瑪戰勝後):設 ZOMA 里程碑(進度 → 9/9 = 破關)+ 啟動 ENDTXT 結局捲動。
 * *end_seq 設為 0 並開首段;主迴圈在對話翻完一段、玩家按 Enter 時逐段推進(見 advance 處)。 */
static void run_finale(dq3_storyflags *flags, dq3_dialogue *dlg, int dlg_ok,
                       const dq3_text *end_txt, int end_ok, int *end_seq)
{
    dq3_progress_set(flags, DQ3_MS_ZOMA);
    fprintf(stderr, "★ 打倒大魔王索瑪 —— 破關!(進度 %d/%d = 全主線完成)\n",
            DQ3_MS_COUNT, DQ3_MS_COUNT);
    /* E-11 時代結尾(そして伝説へ…):勇者受冊封為「洛特」,三件勇者專用裝備昇華為傳說的洛特裝備。
     * 精訊版 ITEM.DAT 本無洛特道具(已查證)→ remake 補上,圓「傳說的終章」副標、接通 DQ1/DQ2 洛特血脈。 */
    dq3_set_loto_blessed(1);
    dq3_flags_set(flags, 0x217, 1);
    fprintf(stderr,
        "★ そして伝説へ… 世界重歸光明。アレフガルド 的人們冊封勇者為『洛特』——\n"
        "  王者之劍化為【洛特之劍】(攻150)、光之鎧甲化為【洛特之鎧】(守95)、勇者之盾化為【洛特之盾】(守55),\n"
        "  將在這片大地代代相傳。多年以後黑暗再臨時,繼承洛特之血的勇者會再次握起它們……\n"
        "  ——獻給未發售的精訊《傳說的終章》:傳說,於此展開。\n");
    if (end_ok && dlg_ok && end_txt->n_records > 0) {
        *end_seq = 0;
        dq3_dialogue_open_text(dlg, end_txt, 0);
        fprintf(stderr, "結局捲動:啟動(共 %d 段 ENDTXT)\n", end_txt->n_records);
    }
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

/* 存檔/讀檔 slot 選擇器(6 格,使用者需求)。for_load=1 時只有「有資料」的 slot 可選
 * (空 slot 仍顯示但不可選)。回選中 slot(1..DQ3_SAVE_SLOTS);ESC 取消回 -1。
 * 標籤 = 「冒險之書 N」概念,以數字 glyph N 顯示 + 有資料時附 ● 標記(glyph 復用)。 */
static int slot_select(const dq3_text *text, int for_load)
{
    /* 「冒險之書」glyph(沿用既有字串若有;此處用數字 + 標題列)。標題:存=記錄/讀=繼續。 */
    static const uint16_t T_SAVE[4] = {533, 218, 145, 268};       /* 「冒險之書」(冒533 險218 之145 書268)*/
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg, gray; uint8_t yellow; dq3_menu m;
    uint8_t *fb = dq3_fb();
    int wx = 150, wy = 70, ww = 260, wh = 200, i;
    int has[DQ3_SAVE_SLOTS + 1];

    raw = dq3_load_file("DQ3.PAL", &rl);
    if (!raw) return -1;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = pal_near2(pal,pn,255,255,255); bg = pal_near2(pal,pn,16,16,32);
    gray = pal_near2(pal,pn,110,110,110); yellow = (uint8_t)pal_near2(pal,pn,255,255,0);

    for (i = 1; i <= DQ3_SAVE_SLOTS; i++) has[i] = dq3_save_exists(slot_path(i));

    dq3_menu_init(&m, wx + 40, wy + 40);
    for (i = 1; i <= DQ3_SAVE_SLOTS; i++) {
        uint16_t lab[1]; lab[0] = (uint16_t)i;     /* 數字 glyph = slot 號 */
        dq3_menu_add(&m, lab, 1);
    }
    m.cursor = 0;

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == 0x01) return -1;                 /* ESC 取消 */
        if (sc) {
            int sel = dq3_menu_input(&m, sc);
            if (sel >= 0) {
                int slot = sel + 1;
                if (for_load && !has[slot]) { /* 讀檔:空格不可選,忽略 */ }
                else return slot;
            }
        }
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        for (i = 0; i < 4; i++)                     /* 標題列:「冒險之書」*/
            dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+16+i*DQ3_GLYPH_PX, wy+12, T_SAVE[i], (uint8_t)white);
        dq3_menu_render(&m, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, (uint8_t)white, yellow);
        /* 每 slot 右側標「有資料/空」:有=黃點概念,用「●」無 glyph 則以數字色區分 —— 簡化:
         * 空 slot 標籤畫灰(覆蓋一次),提示不可載入。 */
        for (i = 1; i <= DQ3_SAVE_SLOTS; i++)
            if (!has[i])
                dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+40, wy+40+(i-1)*m.line_h, (uint16_t)i, (uint8_t)gray);
        dq3_present(); dq3_delay_ms(16);
    }
    return -1;
}

/* 標題主選單(新遊戲 / 繼續冒險)。顯示標題圖 + 2 項選單。
 * 回 0 = 新遊戲;1..6 = 繼續冒險選定的 slot(該 slot 有存檔)。
 * 「繼續冒險」→ slot_select(for_load=1) 選有資料的 slot;無任何存檔則「繼續冒險」灰化提示後回新遊戲。 */
static dq3_config *g_cfg = NULL;   /* main 設定;設定選單讀寫 + 存檔 */

/* 設定選單(title「設定」進入):切 RNG 模式(原版 DOS / 真實 REAL)→ 存 dq3.cfg。
 * 標籤 glyph 用原版字庫 + 自建「設/版」(dq3_customglyph)。←/→/Enter/Space 切換,ESC 存檔離開。 */
static void config_modal(const char *assets, const dq3_text *text)
{
    uint8_t *fb = dq3_fb(); dq3_color pal16[16];
    static const uint16_t T_SET[2]  = { DQ3_GLYPH_SHE, 1022 };   /* 設定 */
    static const uint16_t L_RNG[4]  = { 482, 427, 1397, 1157 };  /* 亂數模式 */
    static const uint16_t V_DOS[2]  = { 555, DQ3_GLYPH_BAN };    /* 原版 */
    static const uint16_t V_REAL[2] = { 577, 279 };              /* 真實 */
    static const uint16_t HINT[2]   = { 750, 312 };              /* 返回 */
    uint8_t white = 15, yellow = 14; int i, dirty = 0;
    if (!g_cfg) return;
    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc == 0x01 || sc == DQ3_SC_F10) break;               /* ESC/F10 → 存檔離開 */
        if (sc == 0x4b || sc == 0x4d || sc == 0x1c || sc == 0x39) {  /* ←/→/Enter/Space 切換 */
            g_cfg->rng_mode = (g_cfg->rng_mode == DQ3_RNG_DOS) ? DQ3_RNG_REAL : DQ3_RNG_DOS;
            dq3_rng_set_mode(g_cfg->rng_mode); dirty = 1;
        }
        if (load_and_decode_title(assets, "TITG.P", fb, pal16) == 0) dq3_set_palette(pal16, 16);
        {   /* 深色框背景(疊在標題上,讓設定文字可讀)*/
            int blk = pal_near2(pal16,16,0,0,0), frm = pal_near2(pal16,16,255,255,255), bgc = pal_near2(pal16,16,16,16,48);
            int bx = 150, by = 70, bw = 340, bh = 180;
            tav_window(fb, bx, by, bw, bh, (uint8_t)blk, (uint8_t)frm, (uint8_t)bgc);
            for (i = 0; i < 2; i++) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, bx+150+i*16, by+22, T_SET[i], yellow);
            for (i = 0; i < 4; i++) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, bx+40+i*16, by+80, L_RNG[i], white);
            { const uint16_t *v = (g_cfg->rng_mode == DQ3_RNG_REAL) ? V_REAL : V_DOS;
              for (i = 0; i < 2; i++) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, bx+40+(5+i)*16, by+80, v[i], yellow); }
            for (i = 0; i < 2; i++) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, bx+140+i*16, by+140, HINT[i], white);
        }
        dq3_present(); dq3_delay_ms(16);
        if (getenv("DQ3_DUMP")) { dq3_dump_ppm(getenv("DQ3_DUMP")); break; }   /* 一幀 dump 驗證 */
    }
    if (dirty) { dq3_config_save(g_cfg, dq3_config_path());
        fprintf(stderr, "設定:亂數模式 = %s,已存 dq3.cfg\n", g_cfg->rng_mode == DQ3_RNG_REAL ? "真實" : "原版"); }
}

static int title_menu(const char *assets, const dq3_text *text)
{
    /* glyph 全用原版真實 record 驗證過的(避免 unicode_map 反查錯,如「女」誤標「文」):
     * 「開始」=488,711(txt02 rec48「從現在開始」)、「繼續冒險」=1079,1094,533,218
     * (繼續 txt02 rec68、冒險之書 txt01 rec43 驗證)。 */
    static const uint16_t NEW[2]  = {488, 711};                 /* 開始(開488 始711)*/
    static const uint16_t CONT[4] = {1079, 1094, 533, 218};     /* 繼續冒險(繼1079 續1094 冒533 險218)*/
    static const uint16_t SET[2]  = {DQ3_GLYPH_SHE, 1022};      /* 設定(設=自建字形 / 定=1022)*/
    dq3_color pal16[16]; dq3_menu m; uint8_t *fb = dq3_fb();
    int has_any = 0, i, wx = 110, wy = 150;
    uint8_t white = 15, yellow = 14;

    for (i = 1; i <= DQ3_SAVE_SLOTS; i++) if (dq3_save_exists(slot_path(i))) has_any = 1;
    if (load_and_decode_title(assets, "TITG.P", fb, pal16) == 0) dq3_set_palette(pal16, 16);

    dq3_menu_init(&m, wx, wy);
    dq3_menu_add(&m, NEW, 2); dq3_menu_add(&m, CONT, 4); dq3_menu_add(&m, SET, 2);
    m.cursor = has_any ? 1 : 0;                       /* 有存檔 → 預設「繼續冒險」*/

    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (sc) {
            int sel = dq3_menu_input(&m, sc);
            if (sel == 0) return 0;                   /* 新遊戲 */
            if (sel == 1) {                            /* 繼續冒險 → 選 slot */
                if (!has_any) { /* 無存檔,忽略 */ }
                else {
                    int slot = slot_select(text, 1);
                    if (slot >= 1 && dq3_save_exists(slot_path(slot))) return slot;
                    /* 取消 → 回標題選單 */
                    if (load_and_decode_title(assets, "TITG.P", fb, pal16) == 0) dq3_set_palette(pal16, 16);
                }
            }
            if (sel == 2) {                            /* 設定 → 設定選單 */
                config_modal(assets, text);
                if (load_and_decode_title(assets, "TITG.P", fb, pal16) == 0) dq3_set_palette(pal16, 16);
            }
        }
        if (load_and_decode_title(assets, "TITG.P", fb, pal16) == 0) { /* 標題底圖每幀重繪 */ }
        dq3_menu_render(&m, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, white, yellow);
        dq3_present(); dq3_delay_ms(16);
    }
    return 0;
}

/* #4 remake 增強:對場景可見格中「有寶箱事件且一次性旗標已設(=取過)」者疊開過標記。
 * 原版寶箱取後不翻 tile(docs/31);此為使用者要求的「看得出開過」回饋,程式疊繪不動 BLK。
 * 只掃 20×15 可見窗(town/dungeon 才有事件層;overworld n_events=0 → 無事件即略過)。 */
static void overlay_opened_chests(const dq3_scene *s, const dq3_storyflags *fl)
{
    int cam_x = s->px - 10, cam_y = s->py - 7, x, y;     /* VIEW 20×15,half=10/7,同 scene 攝影機 */
    if (cam_x > s->map_w - 20) cam_x = s->map_w - 20;
    if (cam_y > s->map_h - 15) cam_y = s->map_h - 15;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;
    for (y = cam_y; y < cam_y + 15 && y < s->map_h; y++)
        for (x = cam_x; x < cam_x + 20 && x < s->map_w; x++) {
            int et, ep, ef;
            if (dq3_scene_tile_event_p2(s, x, y, &et, &ep, &ef)
                && (et == 0 || et == 1 || et == 3) && ep > 0 && ep < 0x90
                && dq3_flags_get(fl, ef))                /* 旗標已設 = 已取過 → 標記開過 */
                dq3_scene_mark_opened_tile(s, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H, x, y);
        }
}

static int run_game(const char *assets, const char *dump)
{
    char err[256] = {0};
    dq3_scene *field, *town = NULL, *cur;
    dq3_scene *field_under = NULL; int layer = 0;  /* 0=地表 1=下層 overworld(docs/43)*/
    int in_town = 0, enc, fx = 0, fy = 0, cur_cty = -1;   /* cur_cty:目前所在 CTY 號(#2 gate)*/
    int repel = 0;                                         /* #3 聖水:剩餘驅弱敵步數 */
    const int over_pool[4] = { 5, 6, 1, 0 };   /* 地表遭遇怪池(史萊姆系) */
    dq3_dialogue dlg; int dlg_ok = 0, dlg_rec = 1, cur_dlg_bank = 1;  /* 啟動載 D3TXT01(阿里阿罕 bank)*/
    dq3_text sys_txt; int sys_ok = 0;                                  /* 常駐 D3TXT00:系統訊息 + 道具名 */
    dq3_text end_txt; int end_ok = 0;                                  /* ENDTXT:結局序列(索瑪戰後)*/
    int end_seq = -1;                                                  /* 結局捲動段號(-1=未進行;0..n=當前 ENDTXT 段)*/
    int tsect = getenv("DQ3_SECT") ? atoi(getenv("DQ3_SECT")) : 0;  /* 城鎮 section(有事件者測試)*/
    dq3_inventory inv; dq3_storyflags flags;                        /* #2 合成:道具欄 + 劇情旗標 */
    dq3_ship ship; dq3_ship_init(&ship);                            /* #2 船:取船後跨海(docs/48)*/
    /* 不死鳥拉米亞(六珠 → 祠堂復活 → 飛行坐騎,青衫攻略;sprite=DQ3MAN.BLS entry 176)。
     * 飛行無視地形(山/海皆可越),落在可走格。 */
    dq3_charsprite phoenix_spr; int phoenix_ok = 0, phoenix_revived = 0, phoenix_aboard = 0;
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
    /* 不死鳥 sprite(DQ3MAN.BLS entry 176 = char44 拉米亞飛行 4 frame)*/
    phoenix_ok = (dq3_charsprite_load(&phoenix_spr, assets, "DQ3MAN.BLS", 176, err, sizeof err) == 0);
    fprintf(stderr, "[init] 不死鳥 sprite(DQ3MAN.BLS e176)載入 %s\n", phoenix_ok ? "OK" : err);
    /* 對話文字(阿里阿罕 = D3TXT01;之後每幀依 section bank 切換,docs/42)*/
    dlg_ok = (dq3_dialogue_load(&dlg, assets, "D3TXT01.TXT", err, sizeof err) == 0);
    /* 常駐系統/道具文字 = D3TXT00(系統訊息、道具名;與 NPC 對話 bank 分離,docs/42)*/
    sys_ok = (dq3_text_load(&sys_txt, assets, "D3TXT00.TXT", err, sizeof err) == 0);
    end_ok = (dq3_text_load(&end_txt, assets, "ENDTXT.TXT", err, sizeof err) == 0);  /* 結局文本 */
    enc = 6 + (int)(grnd() % 8);
    /* demo:身上帶兩材料,進祠堂「調べる」即可觸發 #2 合成(產彩虹水滴) */
    dq3_inv_init(&inv); dq3_flags_init(&flags);
    g_equip_inv = &inv; g_equip_items = &shop_items; g_equip_items_ok = shop_ok;   /* 裝備管理來源 */
    /* (移除早期 #2 合成測試的預塞太陽之石+雲雨之杖;現走真實取得鏈:
     *  太陽之石 CTY80 寶箱、雲雨之杖 精靈祠堂 CTY92。debug 仍可 item:0x72/0x73 補。) */
    /* 續玩:DQ3_LOAD 指定 slot(DQ3_LOAD=N 讀冒險之書 N,2..6;空/0/1 維持相容讀 slot 0 autosave)。
     * 讀檔後 restore_position 載入對應場景到原位置(隨時存讀檔回原位置,使用者需求)。
     * 互動式標題「繼續冒險」選 slot 走 run_game 前的 title_menu(下方)。 */
    int loaded_from_save = 0;
    const char *load_path = NULL;
    /* 互動標題主選單(非 dump、非 DQ3_DEBUG headless):新遊戲 / 繼續冒險選 slot。
     * 選「繼續冒險」slot → load_path;新遊戲 → 不讀檔。headless/dump 跳過,走 DQ3_LOAD env。 */
    if (!dump && !getenv("DQ3_DEBUG") && !getenv("DQ3_LOAD") && sys_ok) {
        int tslot = title_menu(assets, &sys_txt);
        if (tslot >= 1) load_path = slot_path(tslot);
    } else if (getenv("DQ3_LOAD")) {
        int lslot = atoi(getenv("DQ3_LOAD"));        /* DQ3_LOAD=N → slot N(2..6);0/1/非數字 → slot 0(相容舊語意)*/
        load_path = (lslot >= 2 && lslot <= DQ3_SAVE_SLOTS) ? slot_path(lslot) : slot_path(0);
    }
    if (load_path) {
        const char *lpath = load_path;
        dq3_save_pos pos;
        if (dq3_save_exists(lpath) && dq3_save_read(lpath, &roster, &party, &inv, &flags, &pos) == 0) {
            ship.owned = pos.ship_owned; ship.aboard = pos.ship_aboard;   /* #2 還原船狀態 */
            ship.px = pos.ship_px; ship.py = pos.ship_py; ship.layer = pos.ship_layer;
            dq3_scene_set_daynight(pos.daynight);                         /* v6:還原晝夜相位 */
            fprintf(stderr, "讀檔續玩 ← %s(名冊%d 隊伍%d,存檔位置 CTY%d (%d,%d)%s)\n",
                    lpath, roster.count, party.count, pos.cty, pos.px, pos.py,
                    ship.owned ? (ship.aboard ? " 船上" : " 有船") : "");
            /* 還原場景到存檔位置(field 已於上方載入)。 */
            if (restore_position(assets, &pos, field, &field_under, &town, &cur,
                                 &layer, &in_town, &cur_cty, &flags) == 0)
                loaded_from_save = 1;
        }
    }
    set_dialogue_hero(&roster, &party);   /* 對話 {V} 主角名初值(讀檔/新局後)*/

    /* ── DEBUG 口(DQ3_DEBUG 環境變數;腳本/headless 驗證用,免真的玩到那段)──
     * 指令以 ';' 分隔:descent(切下層 overworld)/ ascend / warp:CTY:X:Y(進城)/
     * flag:N(設旗標)/ item:N / gold:N。docs/46。例:DQ3_DEBUG="gold:1000;warp:2:20:10" */
    int debug_placed = 0;   /* debug 已定位玩家(warp/descent/ascend)→ 跳過開場 CTY00 載入 */
    { const char *dbg = getenv("DQ3_DEBUG");
      if (dbg && *dbg) { char buf[256]; char *tok; strncpy(buf, dbg, sizeof buf - 1); buf[sizeof buf - 1] = 0;
        for (tok = strtok(buf, ";"); tok; tok = strtok(NULL, ";")) {
            int a, b, c, dsec_dbg = 0;
            if (strcmp(tok, "descent") == 0) {
                do_descent(assets, &field_under, &cur, &layer, &in_town, &cur_cty, &flags); debug_placed=1;  /* scripted_event 86 */
            } else if (strcmp(tok, "ascend") == 0) {
                layer = 0; cur = field; in_town = 0; cur_cty = -1; dq3_scene_apply_palette(cur);
                debug_placed=1; fprintf(stderr, "[DEBUG] ascend → 地表\n");
            } else if (sscanf(tok, "warp:%d:%d:%d:%d", &a, &b, &c, &dsec_dbg) == 4 ||
                       sscanf(tok, "warp:%d:%d:%d", &a, &b, &c) == 3) {
                char ct[16]; int bn = (a >= 0 && a < 100) ? dq3x_map_blknum[a] : 1; dq3_scene *ns;
                int wsec = (sscanf(tok, "warp:%d:%d:%d:%d", &a, &b, &c, &dsec_dbg) == 4) ? dsec_dbg : 0;
                sprintf(ct, "CTY%02d.DAT", a); ns = dq3_town_load(assets, ct, wsec, bn, err, sizeof err);
                if (ns) { if (town) dq3_scene_free(town); town = ns; cur = town; in_town = 1; cur_cty = a;
                    load_field_hero(town, assets);
                    if (b < cur->map_w) cur->px = b; if (c < cur->map_h) cur->py = c;
                    dq3_scene_apply_palette(cur); debug_placed=1; fprintf(stderr, "[DEBUG] warp → CTY%d (%d,%d)\n", a, b, c);
                    if (getenv("DQ3_ORBSCAN")) {   /* 臨時:掃 event tile 給道具(0x40-0x8f)的位置 */
                        int tx, ty, et, ep;
                        for (ty = 0; ty < cur->map_h; ty++) for (tx = 0; tx < cur->map_w; tx++)
                            if (dq3_scene_tile_event_p2(cur, tx, ty, &et, &ep, 0) && ep > 0 && ep < 0x90)
                                fprintf(stderr, "  [itemscan] (%d,%d) type%d → 道具 0x%02x\n", tx, ty, et, ep);
                    } }
                else fprintf(stderr, "[DEBUG] warp CTY%d 失敗: %s\n", a, err);
            } else if (strcmp(tok, "party") == 0) {          /* 建測試隊伍(勇者/戰士/僧侶/魔法使者)*/
                static const int dcls[4] = {0, 1, 3, 4}; int pi;
                for (pi = 0; pi < 4 && roster.count < DQ3_ROSTER_MAX; pi++) {
                    uint16_t nm[1]; int ri; nm[0] = (uint16_t)(pi + 1);   /* 名 = glyph 數字 1-4 */
                    ri = dq3_roster_create(&roster, &gst, dcls[pi], DQ3_GENDER_MALE, nm, 1);
                    if (ri >= 0) dq3_party_add(&party, &roster, ri);
                }
                fprintf(stderr, "[DEBUG] party → 名冊%d 隊伍%d\n", roster.count, party.count);
            } else if (strcmp(tok, "equip") == 0) {          /* 直接開裝備管理(驗證)*/
                equip_modal(&roster, &party, sys_ok ? &sys_txt : &dlg.txt);
                if (getenv("DQ3_EQUIP_DUMP")) return 0;       /* dump 模式:equip_modal 已 dump,直接結束不被主 dump 覆寫 */
            } else if (sscanf(tok, "event:%i", &a) == 1) {   /* 直接跑 scripted_event */
                if (a == DQ3_SEVENT_DESCENT)
                    do_descent(assets, &field_under, &cur, &layer, &in_town, &cur_cty, &flags);
                else { int r = dq3_scripted_event_run(a, &inv, &flags, 1);
                    fprintf(stderr, "[DEBUG] event %d → result=%d\n", a, r); }
            } else if (strcmp(tok, "prog") == 0) {           /* 顯示主線進度階段(旗標流,#1)*/
                int st = dq3_progress_stage(&flags);
                fprintf(stderr, "[DEBUG] 進度階段 %d/%d = %s\n", st, DQ3_MS_COUNT, dq3_progress_name(st));
            } else if (sscanf(tok, "prog:%d", &a) == 1) {     /* 設里程碑 0..N 完成 */
                static const int seq[DQ3_MS_COUNT] = {DQ3_MS_START,DQ3_MS_THIEF_KEY,DQ3_MS_MAGIC_BALL,
                    DQ3_MS_ROMALY,DQ3_MS_DHAMA,DQ3_MS_SHIP,DQ3_MS_RAINBOW,DQ3_MS_DESCEND,DQ3_MS_ZOMA};
                int k; for (k = 0; k < DQ3_MS_COUNT && k <= a; k++) dq3_progress_set(&flags, seq[k]);
                fprintf(stderr, "[DEBUG] 進度設到里程碑 %d → 階段 %s\n", a, dq3_progress_name(dq3_progress_stage(&flags)));
            } else if (sscanf(tok, "opos:%d:%d", &a, &b) == 2) { /* overworld 定位玩家(當前層)*/
                if (!in_town && a < cur->map_w && b < cur->map_h) { cur->px = a; cur->py = b; debug_placed = 1;
                    fprintf(stderr, "[DEBUG] overworld 定位 (%d,%d) layer%d\n", a, b, layer); }
            } else if (strcmp(tok, "ship") == 0) {            /* 取船 + 登船於玩家當前格(#2)*/
                ship.owned = 1; ship.aboard = 1; ship.px = cur->px; ship.py = cur->py; ship.layer = layer;
                dq3_progress_set(&flags, DQ3_MS_SHIP);
                fprintf(stderr, "[DEBUG] 取得船 + 登船於 (%d,%d) layer%d\n", ship.px, ship.py, layer);
            } else if (sscanf(tok, "ship:%d:%d", &a, &b) == 2) { /* 船停泊於 (a,b),不登船 */
                ship.owned = 1; ship.aboard = 0; ship.px = a; ship.py = b; ship.layer = layer;
                dq3_progress_set(&flags, DQ3_MS_SHIP);
                fprintf(stderr, "[DEBUG] 船停泊 (%d,%d) layer%d\n", a, b, layer);
            } else if (strcmp(tok, "phoenix") == 0) {         /* 不死鳥拉米亞復活 + 搭乘(飛行坐騎)*/
                phoenix_revived = 1;
                if (!in_town) phoenix_aboard = 1;
                fprintf(stderr, "[DEBUG] 不死鳥拉米亞復活%s\n", phoenix_aboard ? " + 搭乘(飛行)" : "");
            } else if (strcmp(tok, "zomaseq") == 0) {         /* 索瑪前完整序列:六大魔人 → 怨靈122 → 殭屍123 → 索瑪124(杜 Ch56)*/
                int oc = 1, dm; const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                /* 六大魔人守衛(怪106 ×6;杜勝利 Ch56:龍王城六座石像=索瑪神殿大魔人守衛,盡破→現隱藏樓梯通索瑪)。
                 * 怪名 rec=id+600 反推:大魔人=rec706→id106。gate=flag 0x214:已破則跳過(不重打,隱藏樓梯已現)。 */
                if (!dq3_flags_get(&flags, 0x214)) {
                    fprintf(stderr, "★ 索瑪神殿:六大魔人守衛(怪106 ×6)\n");
                    for (dm = 1; dm <= 6 && oc == 1; dm++) {
                        fprintf(stderr, "  大魔人 %d/6\n", dm);
                        oc = dq3_battlescene_run(assets, 106, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                        dq3_scene_apply_palette(cur);
                    }
                    if (oc == 1) { dq3_flags_set(&flags, 0x214, 1); fprintf(stderr, "★ 六大魔人盡破 → 隱藏樓梯現,通索瑪\n"); }
                }
                if (oc == 1) {
                    fprintf(stderr, "★ 索瑪神殿前序列：巴拉摩斯怨靈(怪122)\n");
                    oc = dq3_battlescene_run(assets, 122, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                    dq3_scene_apply_palette(cur);
                }
                if (oc == 1) {
                    fprintf(stderr, "★ 巴拉摩斯僵尸(怪123)\n");
                    oc = dq3_battlescene_run(assets, 123, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                    dq3_scene_apply_palette(cur);
                }
                if (oc == 1) {
                    dq3_battlescene_set_light_orb(dq3_inv_find(&inv, 0x65) >= 0);
                    fprintf(stderr, "★ 大魔王索瑪(怪124)\n");
                    oc = dq3_battlescene_run(assets, 0x7c, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                    dq3_scene_apply_palette(cur);
                }
                fprintf(stderr, "[DEBUG] 索瑪序列 outcome=%d\n", oc);
                if (oc == 1) run_finale(&flags, &dlg, dlg_ok, &end_txt, end_ok, &end_seq);
            } else if (strcmp(tok, "zoma") == 0) {            /* 索瑪終戰(怪 0x7c)→ 勝則破關 */
                int oc; const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                dq3_battlescene_set_light_orb(dq3_inv_find(&inv, 0x65) >= 0);  /* 持光之珠 → 索瑪二階段弱化 */
                oc = dq3_battlescene_run(assets, 0x7c, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                dq3_scene_apply_palette(cur);
                fprintf(stderr, "[DEBUG] 索瑪戰 outcome=%d(1=勝 2=敗 3=逃)\n", oc);
                if (oc == 1) run_finale(&flags, &dlg, dlg_ok, &end_txt, end_ok, &end_seq);
            } else if (sscanf(tok, "boss:%i:%i", &a, &b) == 2 || sscanf(tok, "boss:%i", &a) == 1) {
                /* 通用 boss 戰:boss:<怪id>[:<勝利獎勵道具>]。真隊伍上場,勝利給獎勵 + log。
                 * (甘達特 26→皇冠 0x33、八頭大蛇 75 等 boss 劇情事件的可驗證機制。)*/
                int oc, reward = (sscanf(tok, "boss:%i:%i", &a, &b) == 2) ? b : -1;
                const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                oc = dq3_battlescene_run(assets, a, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                dq3_scene_apply_palette(cur);
                fprintf(stderr, "[DEBUG] boss 戰 怪0x%02x outcome=%d(1=勝 2=敗 3=逃)\n", a, oc);
                if (oc == 1 && reward >= 0) {
                    dq3_inv_add(&inv, reward);
                    fprintf(stderr, "★ 擊敗 boss 怪0x%02x → 獲得獎勵道具 0x%02x\n", a, reward);
                }
            } else if (strcmp(tok, "baramos") == 0) {  /* 巴拉摩斯本體 boss(怪121 HP1201,杜勝利 Ch44;下降前假最終王)*/
                const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                int oc;
                dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                fprintf(stderr, "★ 巴拉摩斯戰(怪121 HP1201)— 下降前主線大 boss\n");
                oc = dq3_battlescene_run(assets, 121, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                dq3_scene_apply_palette(cur);
                if (oc == 1) { dq3_flags_set(&flags, 0x213, 1); fprintf(stderr, "★ 擊敗巴拉摩斯 → 設旗標 0x213，可下降(索瑪現身)\n"); }
                else fprintf(stderr, "巴拉摩斯戰 outcome=%d(未勝)\n", oc);
            } else if (strcmp(tok, "finale") == 0) {          /* 直接驗破關→結局路徑(主線推到 9/9)*/
                run_finale(&flags, &dlg, dlg_ok, &end_txt, end_ok, &end_seq);
            } else if (sscanf(tok, "dlg:%d:%d", &a, &b) == 2) { /* 渲染任意對話 record(bank a,rec b)*/
                if (dlg_ok && a >= 0 && a <= 9) {
                    char be[128];
                    if (dq3_dialogue_set_bank(&dlg, assets, a, be, sizeof be) == 0) {
                        cur_dlg_bank = a; debug_placed = 1;   /* 跳過開場 CTY00 載入,否則迴圈把 bank 切回 1 */
                        if (dq3_dialogue_open(&dlg, b) == 0)
                            fprintf(stderr, "[DEBUG] 渲染 D3TXT0%d rec=0x%02x(%d)\n", a, b, b);
                        else fprintf(stderr, "[DEBUG] rec %d 開啟失敗\n", b);
                    } else fprintf(stderr, "[DEBUG] bank %d 載入失敗: %s\n", a, be);
                }
            } else if (sscanf(tok, "reclass:%d:%d", &a, &b) == 2) {  /* 達瑪轉職:隊伍 slot a → 職業 b */
                if (a >= 0 && a < party.count) {
                    dq3_member *m = &roster.list[party.slot[a]].m;
                    int oc = m->cls;
                    if (dq3_member_change_class(m, &gst, b) == 0)
                        fprintf(stderr, "★ 達瑪轉職:隊員%d 職業 %d→%d(Lv1,屬性減半)\n", a, oc, b);
                    else fprintf(stderr, "[DEBUG] 轉職失敗(勇者/不合法職業)\n");
                }
            } else if (strcmp(tok, "dhama") == 0) {           /* 開達瑪轉職選單(神官)*/
                dhama_modal(&roster, &party, &gst, sys_ok ? &sys_txt : &dlg.txt, &inv);
                dq3_scene_apply_palette(cur);
            } else if (sscanf(tok, "status:%d:%d", &a, &b) == 2) { /* 施加異常狀態:slot a,bit b(1毒2麻)*/
                if (a >= 0 && a < party.count) {
                    roster.list[party.slot[a]].m.status |= (uint8_t)b;
                    fprintf(stderr, "[DEBUG] 隊員%d status |= 0x%x(1=毒 2=麻)\n", a, b);
                }
            } else if (sscanf(tok, "hurt:%i", &a) == 1) {     /* 設隊長 cur_hp(測藥草用)*/
                if (party.count > 0) { roster.list[party.slot[0]].m.cur_hp = (uint16_t)a;
                    fprintf(stderr, "[DEBUG] 隊長 cur_hp=%d\n", a); }
            } else if (sscanf(tok, "use:%i", &a) == 1) {      /* 使用消耗品(#3)*/
                int k = apply_item_use(&inv, &roster, &party, a);
                if (k == DQ3_USE_RETURN_TOWN) {                /* 蓋美拉翅膀:回地表 */
                    layer = 0; cur = field; in_town = 0; cur_cty = -1; ship.aboard = 0;
                    dq3_scene_apply_palette(cur); debug_placed = 1;
                    fprintf(stderr, "蓋美拉翅膀:回到地表 (%d,%d)\n", cur->px, cur->py);
                } else if (k == DQ3_USE_REPEL) {               /* 聖水:驅弱敵 N 步 */
                    repel = DQ3_HOLY_STEPS;
                    fprintf(stderr, "聖水:弱敵迴避 %d 步\n", repel);
                } else if (k == DQ3_USE_AWAKEN) {              /* 覺醒粉:諾阿尼魯(CTY4)解催眠 */
                    if (in_town && cur_cty == 4) { dq3_inv_remove(&inv, 0x5a); dq3_flags_set(&flags, 0x31, 1);
                        fprintf(stderr, "★ 諾阿尼魯村:使用覺醒粉 → 全村催眠詛咒解除!(杜勝利 Ch11)\n"); }
                    else fprintf(stderr, "覺醒粉:此處用不上(需在諾阿尼魯村 CTY4),未消耗\n");
                } else if (k == DQ3_USE_GAIA) {                /* 蓋亞之劍:地表火山開通 */
                    if (!in_town) { dq3_flags_set(&flags, 0x32, 1);
                        fprintf(stderr, "★ 蓋亞之劍:小火山熔流而出,開通往尼羅肯特洞窟之路!(杜勝利 Ch40)\n"); }
                    else fprintf(stderr, "蓋亞之劍:需在地表小火山旁使用\n");
                } else if (k == DQ3_USE_DRAIN) {               /* 乾渴壺:地表四島礁吸海 */
                    if (!in_town) { dq3_flags_set(&flags, 0x33, 1);
                        fprintf(stderr, "★ 乾渴壺:吸乾海水,最終鑰匙祠堂顯現!(杜勝利 Ch27)\n"); }
                    else fprintf(stderr, "乾渴壺:需在地表四島礁附近使用\n");
                } else if (k == DQ3_USE_FAIRYFLUTE) {          /* 妖精之笛:魯比斯之塔 CTY82 解詛咒 */
                    if (in_town && cur_cty == 82) {
                        if (dq3_inv_find(&inv, 0x74) < 0) dq3_inv_add(&inv, 0x74);
                        dq3_flags_set(&flags, 0x34, 1);
                        fprintf(stderr, "★ 魯比斯之塔:妖精之笛解魯比斯詛咒 → 得精靈的守護 0x74\n");
                    } else fprintf(stderr, "妖精之笛:需在魯比斯之塔(CTY82)使用\n");
                } else if (k == DQ3_USE_RAINBOW) {             /* 彩虹水滴:下層利姆達爾西北架彩虹橋 */
                    if (!in_town && layer == 1) {
                        dq3_inv_remove(&inv, 0x75); dq3_flags_set(&flags, 0x35, 1);
                        dq3_progress_set(&flags, DQ3_MS_RAINBOW);
                        fprintf(stderr, "★ 彩虹水滴:雨和太陽合而為一,彩虹橋出現!通往終盤(杜勝利 Ch55)\n");
                    } else fprintf(stderr, "彩虹水滴:需在下層利姆達爾西北盡頭地表使用\n");
                }
            } else if (sscanf(tok, "flag:%i", &a) == 1) { dq3_flags_set(&flags, a, 1); fprintf(stderr, "[DEBUG] flag 0x%x set\n", a); }
            else if (sscanf(tok, "item:%i", &a) == 1) { dq3_inv_add(&inv, a); fprintf(stderr, "[DEBUG] item 0x%x\n", a); }
            else if (sscanf(tok, "gold:%d", &a) == 1) { gold = a; fprintf(stderr, "[DEBUG] gold=%d\n", a); }
            else if (sscanf(tok, "dn:%d", &a) == 1) { dq3_scene_set_daynight(a); if (cur) dq3_scene_apply_palette(cur); fprintf(stderr, "[DEBUG] 晝夜相位=%s\n", DN_NAME[a&3]); }
            else if (strcmp(tok, "sokoban") == 0) { g_sokoban_solved = 1; fprintf(stderr, "[DEBUG] 倉庫番強制解謎(密道開啟,測試用)\n"); }
            else fprintf(stderr, "[DEBUG] 未知指令: %s\n", tok);
        }
      }
    }

    if (dump && getenv("DQ3_DEBUG") && !getenv("DQ3_INPUT")) {
        /* DEBUG 口 + dump(無腳本輸入):直接渲染 debug 狀態並 dump;有 DQ3_INPUT 則走互動迴圈 */
        if (in_town) dq3_scene_npc_tick(cur);
        dq3_scene_render(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        overlay_opened_chests(cur, &flags);    /* #4:已取寶箱疊開過標記(remake 增強)*/
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "debug 場景 -> %s OK(layer=%d cty=%d (%d,%d))\n",
                                              dump, layer, cur_cty, cur->px, cur->py);
        if (dlg_ok) dq3_dialogue_free(&dlg);
        if (sys_ok) dq3_text_free(&sys_txt);
        if (town) dq3_scene_free(town);
        if (field_under) dq3_scene_free(field_under);
        dq3_scene_free(field);
        return 0;
    }
    if (dump && !getenv("DQ3_INPUT")) {
        /* headless demo:走到觸發戰鬥 → 進城鎮,沿途 log,dump 末幀(有 DQ3_INPUT 則走互動迴圈)*/
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
        overlay_opened_chests(cur, &flags);    /* #4:已取寶箱疊開過標記(remake 增強)*/
        /* demo:在城鎮開一段對話(D3TXT01 rec 1)疊在場景上 */
        if (dlg_ok && dq3_dialogue_open(&dlg, 1) == 0) {
            fprintf(stderr, "=== 對話 D3TXT01 rec1(疊在城鎮)===\n");
            dq3_dialogue_render(&dlg, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        }
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "game -> %s OK\n", dump);
        if (dlg_ok) dq3_dialogue_free(&dlg);
        if (sys_ok) dq3_text_free(&sys_txt);
        if (town) dq3_scene_free(town);
        if (field_under) dq3_scene_free(field_under);
        dq3_scene_free(field);
        return 0;
    }

    /* 互動開場:從阿里阿罕城鎮(CTY00 sec0)起步。debug 已定位 / 讀檔已還原場景則跳過。 */
    if (!debug_placed && !loaded_from_save) {
        dq3_scene *t0 = dq3_town_load(assets, "CTY00.DAT", 0, 1, err, sizeof err);
        if (t0) { town = t0; load_field_hero(town, assets); cur = town; in_town = 1; cur_cty = 0;
                  dq3_scene_apply_palette(cur);
                  fprintf(stderr, "開場:阿里阿罕城鎮(CTY00)— 西側櫃台建隊、C 開指令窗、B 商店、SPACE 出城\n");
                  /* 開場對白:媽媽送你出發(D3TXT01 阿里阿罕 rec21;主角是勇者歐里狄加的後代)*/
                  if (dlg_ok) dq3_dialogue_open(&dlg, 21); }
    }

    /* 互動:方向走動 + 隨機遭遇;SPACE 進/出城鎮;Enter 對話(前方有事件時)。 */
    while (!dq3_should_quit()) {
        uint8_t sc;
        /* 達瑪神殿:進城即達成 DHAMA 里程碑(轉職開放;實際換職另為系統)。 */
        if (in_town && cur_cty == DQ3_DHAMA_CTY && !dq3_progress_done(&flags, DQ3_MS_DHAMA)) {
            dq3_progress_set(&flags, DQ3_MS_DHAMA);
            fprintf(stderr, "★ 抵達達瑪神殿 → 轉職開放(DHAMA 里程碑)\n");
        }
        /* 對話檔跟著 section bank 切換(docs/42:bank=section header +0x17 → D3TXT0<bank>);
         * 進不同城/section 時 reload,讓 NPC 對話取對檔。不在對話顯示中才切(避免中途換檔)。*/
        if (in_town && dlg_ok && cur->dlg_bank >= 1 && cur->dlg_bank <= 9
            && cur->dlg_bank != cur_dlg_bank && !dq3_dialogue_is_open(&dlg)) {
            char be[128];
            if (dq3_dialogue_set_bank(&dlg, assets, cur->dlg_bank, be, sizeof be) == 0)
                cur_dlg_bank = cur->dlg_bank;
        }
        /* 不死鳥拉米亞:集滿六珠(綠藍紅紫黃銀 0x66-0x6b)→ 復活(六珠之力;原版不死鳥祠堂
         * 守護者觸發在此未發售 build 不完整 → remake 以「集齊六珠」當可靠復活條件)。overworld 按 y 起飛。 */
        if (!phoenix_revived) {
            int orbs = 0, oi; for (oi = 0x66; oi <= 0x6b; oi++) if (dq3_inv_find(&inv, oi) >= 0) orbs++;
            if (orbs >= 6) { phoenix_revived = 1;
                fprintf(stderr, "★ 六珠齊備(綠藍紅紫黃銀)→ 不死鳥拉米亞復活!(出城至 overworld 按 y 起飛)\n"); }
        }
        /* B-7 倉庫番:離開 CTY76 → 解謎旗標重置(石頭由 scene reload 自動回原位;使用者要求重入重置)。 */
        if (cur_cty != SOKOBAN_CTY) g_sokoban_solved = 0;
        /* B-6 新城鎮視覺:未建城(flag 0x216)→ 隱藏商店/居民(vis_flag 0x05),留老人 (16,2)空草原;
         * 帶商人來建城後重入 → 全顯(可買更多東西=成就感)。 */
        if (in_town && cur_cty == 83 && !dq3_flags_get(&flags, 0x216))
            dq3_scene_hide_unbuilt(cur, 0x05, 16, 2);
        /* NPC 隨機走動(docs/35 §九):城鎮每幀步進;對話中凍結不動。 */
        if (in_town && !(dlg_ok && dq3_dialogue_is_open(&dlg))) dq3_scene_npc_tick(cur);
        if (!in_town) animate_sea(cur, g_sea_frame++);   /* 海面 palette cycling(地表/下層)*/
        dq3_scene_render(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        overlay_opened_chests(cur, &flags);    /* #4:已取寶箱疊開過標記(remake 增強)*/
        draw_ship_overlay(cur, &ship, in_town, layer);   /* 船 sprite(docs/51)*/
        if (!in_town && phoenix_aboard && phoenix_ok)    /* 不死鳥坐騎 sprite 疊在玩家格(飛行中)*/
            dq3_scene_draw_charsprite_at(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H,
                                         cur->px, cur->py, &phoenix_spr, cur->facing & 3);
        if (dlg_ok && dq3_dialogue_is_open(&dlg))
            dq3_dialogue_render(&dlg, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        sc = dq3_poll_scancode();
        if (sc == DQ3_SC_F5) {              /* F5:隨時存檔 → 選 slot(使用者需求:隨時存讀)*/
            int slot = slot_select(sys_ok ? &sys_txt : &dlg.txt, 0);
            if (slot >= 1) {
                save_game_to(slot_path(slot), &roster, &party, &inv, &flags,
                             cur_cty, cur->px, cur->py, &ship, in_town, layer, in_town ? cur->section : 0);
                fprintf(stderr, "F5 隨時存檔:存入冒險之書 %d\n", slot);
            } else fprintf(stderr, "F5 存檔:取消\n");
            dq3_scene_apply_palette(cur);
        } else if (sc == DQ3_SC_F9) {       /* F9:隨時讀檔 → 選 slot → restore 回原位置 */
            int slot = slot_select(sys_ok ? &sys_txt : &dlg.txt, 1);   /* for_load=1:空格不可選 */
            if (slot >= 1 && dq3_save_exists(slot_path(slot))) {
                dq3_save_pos pos;
                if (dq3_save_read(slot_path(slot), &roster, &party, &inv, &flags, &pos) == 0) {
                    ship.owned = pos.ship_owned; ship.aboard = pos.ship_aboard;
                    ship.px = pos.ship_px; ship.py = pos.ship_py; ship.layer = pos.ship_layer;
                    dq3_scene_set_daynight(pos.daynight);                 /* v6:還原晝夜相位 */
                    restore_position(assets, &pos, field, &field_under, &town, &cur,
                                     &layer, &in_town, &cur_cty, &flags);
                    set_dialogue_hero(&roster, &party);
                    fprintf(stderr, "F9 隨時讀檔:載入冒險之書 %d → 回原位置\n", slot);
                }
            } else { fprintf(stderr, "F9 讀檔:取消 / 空 slot\n"); dq3_scene_apply_palette(cur); }
        } else if (sc == DQ3_SC_F10) {      /* F10:離開遊戲確認(Yes/No)→ 選 slot 存檔(使用者需求)*/
            if (confirm_quit(&dlg.txt)) {
                int slot = slot_select(sys_ok ? &sys_txt : &dlg.txt, 0);   /* 選格;ESC → 存 slot 0 autosave */
                save_game_to(slot >= 1 ? slot_path(slot) : slot_path(0),
                             &roster, &party, &inv, &flags, cur_cty, cur->px, cur->py, &ship,
                             in_town, layer, in_town ? cur->section : 0);
                fprintf(stderr, "F10 離開:存入%s\n", slot >= 1 ? "選定 slot" : "自動存檔(slot 0)");
                dq3_rt_set_quit();
            }
            dq3_scene_apply_palette(cur);   /* confirm 改過 DAC,還原場景色盤 */
        } else if (dlg_ok && dq3_dialogue_is_open(&dlg)) {
            if (sc == 0x1c || sc == 0x39) {
                dq3_dialogue_advance(&dlg);  /* Enter/Space 翻頁/關閉 */
                /* 結局捲動:一段 ENDTXT 翻完關閉後,自動接下一段,直到全部播完。 */
                if (end_seq >= 0 && !dq3_dialogue_is_open(&dlg)) {
                    end_seq++;
                    if (end_ok && end_seq < end_txt.n_records)
                        dq3_dialogue_open_text(&dlg, &end_txt, end_seq);
                    else {
                        end_seq = -1;
                        fprintf(stderr, "結局捲動:全部播畢 —— THE END\n");
                    }
                }
            }
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
                    set_dialogue_hero(&roster, &party);   /* 主角名可能改變 → 更新 {V} */
                    dq3_scene_apply_palette(cur);
                    fprintf(stderr, "露依達酒場:名冊%d 人、隊伍%d 人\n", roster.count, party.count);
                }
            }
            /* 話す:面向格上的 NPC → 顯示其對話(docs/42:dlg id=byte4,對話檔=section bank;{V}=主角名)。
             * 互動子型 (ctrl>>3)&7:0/1=對話、2=特殊 → 都顯示對話;3-7=設施(商店/旅社,另由 B 鍵等處理)。*/
            int talked = 0;
            {
                int fdx3 = (cur->facing==3)-(cur->facing==1), fdy3 = (cur->facing==0)-(cur->facing==2);
                int ni = dq3_scene_npc_at(cur, cur->px+fdx3, cur->py+fdy3);
                if (ni >= 0) {
                    int sub = (cur->npcs[ni].ctrl >> 3) & 7, b4 = cur->npcs[ni].b4;
                    /* 對話時 NPC 轉向面對勇者(原版行為):NPC 朝向 = 玩家朝向 ^ 2
                     * (上↔下、左↔右)。對話中 npc_tick 凍結 → 持續面向玩家。 */
                    cur->npcs[ni].ctrl = (uint8_t)((cur->npcs[ni].ctrl & ~3) | ((cur->facing ^ 2) & 3));
                    /* 子型 0/1 = 對話(byte4=對話 rec);2 = scripted-event NPC(byte4 索引 0x3bb4 跳表,
                     * 旗標條件對話 → 取主對話 rec,docs/42);3-7 = 設施(docs/40,byte4=設施索引)。 */
                    if (sub < 2 && cur_cty == DQ3_DHAMA_CTY &&
                        cur->npcs[ni].x == 6 && cur->npcs[ni].y == 4) {   /* 達瑪神官 → 轉職選單 */
                        dhama_modal(&roster, &party, &gst, sys_ok ? &sys_txt : &dlg.txt, &inv);
                        dq3_scene_apply_palette(cur);
                        talked = 1;
                    } else if (sub < 2 && cur_cty == 20 && cur->npcs[ni].x == 16 && cur->npcs[ni].y == 2 && dlg_ok) {
                        /* 提頓村(CTY20)牢房犯人:給綠寶珠 0x66(青衫攻略;runner handler byte4=35,rec38/39)。
                         * 提頓 = テドン(バラモス 手下滅村;白天廢墟、夜晚亡靈現身):原版閘在「夜晚進村 + 開牢門」
                         * (runner/region 事件,day-night doc §9 RE 結論)。日夜系統已實作 → 忠實還原夜 gated:
                         * **綠寶珠夜限定**(g_dn_phase==黑夜 才開牢門給珠);白天仍可讀屍體留書(物理遺書,rec29
                         * 「寶珠留給有緣人」)當提示,但不給珠。綠寶珠 → 雷亞姆蘭特祭壇(byte4=63-68 收)。 */
                        set_dialogue_hero(&roster, &party);
                        if (dq3_scene_get_daynight() != 2) {
                            dq3_dialogue_open(&dlg, b4);   /* 白天:只見留書,牢門深鎖(夜訪才開)*/
                            fprintf(stderr, "提頓村牢房(白天):牢門深鎖,只餘屍上留書;夜裡亡靈現身才開牢門\n");
                        } else if (dq3_inv_find(&inv, 0x66) < 0) {
                            dq3_inv_add(&inv, 0x66);
                            dq3_dialogue_open(&dlg, b4);   /* 夜:rec29 留書 + 開牢門取珠 */
                            fprintf(stderr, "★ 提頓村牢房(夜):亡靈現身開牢門,獲得綠寶珠 0x66 → 雷亞姆蘭特祭壇\n");
                        } else {
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "提頓村牢房:已取綠寶珠(rec%d)\n", b4);
                        }
                        talked = 1;
                    } else if (sub < 2 && cur_cty == 92 && cur->npcs[ni].x == 9 && cur->npcs[ni].y == 8 && dlg_ok) {
                        /* 精靈祠堂(CTY92)精靈:「服侍魯比斯的精靈」,持精靈的守護 0x74 → 換雲雨之杖 0x73
                         * (杜勝利 Ch54;原版 byte4=75 runner,[0x722]=81 region 觸發,bank9 rec28)。
                         * 雲雨之杖 → 神聖祠堂合成彩虹水滴(架彩虹橋通終盤)。 */
                        set_dialogue_hero(&roster, &party);
                        if (dq3_inv_find(&inv, 0x73) >= 0) {
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "精靈祠堂:已得雲雨之杖\n");
                        } else if (dq3_inv_find(&inv, 0x74) >= 0) {
                            dq3_inv_remove(&inv, 0x74); dq3_inv_add(&inv, 0x73);   /* 精靈的守護 → 雲雨之杖 */
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "★ 精靈祠堂:服侍魯比斯的精靈 → 換得雲雨之杖 0x73(彩虹材料)\n");
                        } else {
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "精靈祠堂:需先持精靈的守護(魯比斯之塔)\n");
                        }
                        talked = 1;
                    } else if (sub < 2 && cur_cty == 81 && cur->section == 1
                               && cur->npcs[ni].x == 4 && cur->npcs[ni].y == 3 && dlg_ok) {
                        /* 瑪依拉(CTY81)二樓道具店主:賣歐里空金屬 0x6d 換王者之劍 0x1c(杜勝利 Ch50)。
                         * 原版=賣金屬 22500G + 買王者之劍 35000G;合為「淨支出 12500G + 金屬 → 王者之劍」。
                         * (瑪依拉道具店 facility 正式店員入口 k!=b4 接線未定位,先掛此二樓對話 NPC。) */
                        set_dialogue_hero(&roster, &party);
                        if (dq3_inv_find(&inv, 0x1c) >= 0) {
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "瑪依拉道具店:已換得王者之劍\n");
                        } else if (dq3_inv_find(&inv, 0x6d) >= 0 && gold >= 12500) {
                            dq3_inv_remove(&inv, 0x6d); gold -= 12500; dq3_inv_add(&inv, 0x1c);
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "\u2605 \u746a\u4f9d\u62c9\u9053\u5177\u5e97:\u6b50\u91cc\u7a7a\u91d1\u5c6c 0x6d +12500G \u2192 \u738b\u8005\u4e4b\u528d 0x1c(\u8ce022500\u8cb735000)\n");
                        } else if (dq3_inv_find(&inv, 0x6d) >= 0) {
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "瑪依拉道具店:需 12500G 才能換王者之劍(金屬賣價已折抵)\n");
                        } else {
                            dq3_dialogue_open(&dlg, b4);
                            fprintf(stderr, "瑪依拉道具店:帶歐里空金屬(達姆杜拉)來換王者之劍\n");
                        }
                        talked = 1;
                    } else if (sub < 2 && cur_cty == 83 && cur->npcs[ni].x == 16 && cur->npcs[ni].y == 2 && dlg_ok) {
                        /* 新城鎮老人(RE handler 0x5aba:掃隊伍 [si+1]==6 找商人 → 寄存 + clear_flag 0x23 建城;
                         * 對話 rec0-5 我在這裡想/只要有商人/如何/是嗎/真的嗎/謝謝。杜勝利 Ch33-36)。
                         * 原版老人是 runner/region 事件(byte4 跳表無對應、無放置 NPC)→ remake 掛 CTY83 (16,2) NPC 位。
                         * 商人寄存 = 移出隊伍+名冊(裝備物品回阿里阿罕預存所,杜 Ch33);flag 0x216=建城完成。 */
                        int mi = -1, pi;
                        for (pi = 0; pi < party.count; pi++)
                            if (roster.list[party.slot[pi]].m.cls == 6) { mi = pi; break; }
                        set_dialogue_hero(&roster, &party);
                        if (dq3_flags_get(&flags, 0x216)) {
                            dq3_dialogue_open(&dlg, 5);       /* rec5「喔喔,非常謝謝你!」(已建城)*/
                            fprintf(stderr, "新城鎮老人:城鎮已建成(感謝商人同伴)\n");
                        } else if (mi >= 0) {
                            dq3_roster_remove(&roster, &party, party.slot[mi]);   /* 商人寄存:移出隊伍+名冊 */
                            dq3_flags_set(&flags, 0x216, 1);                       /* 建城完成 */
                            dq3_dialogue_open(&dlg, 5);                            /* rec5「喔喔,非常謝謝你!」*/
                            fprintf(stderr, "★ 新城鎮老人:商人同伴留下建城 → 寄存(回阿里阿罕預存所)+ 新城鎮建成(杜 Ch33-36)\n");
                        } else {
                            dq3_dialogue_open(&dlg, 1);       /* rec1「只要有商人的就好,」*/
                            fprintf(stderr, "新城鎮老人:想建城但缺商人(回阿里阿罕創商人同伴帶來,杜 Ch33)\n");
                        }
                        talked = 1;
                    } else if (sub < 2 && dlg_ok) {          /* 對話型 NPC */
                        set_dialogue_hero(&roster, &party);  /* {V} 主角名 */
                        if (dq3_dialogue_open(&dlg, b4) == 0) {
                            talked = 1;
                            fprintf(stderr, "話す NPC@(%d,%d) bank=D3TXT0%d rec=0x%02x\n",
                                    cur->npcs[ni].x, cur->npcs[ni].y, cur->dlg_bank, b4);
                        }
                    } else if (sub == 2 && cur_cty == DQ3_PORTOGA_CTY && dlg_ok &&
                               cur->npcs[ni].x == DQ3_PORTOGA_KING_X && cur->npcs[ni].y == DQ3_PORTOGA_KING_Y) {
                        /* 取船劇情:波魯多加國王(sub2 scripted NPC)。胡椒換船(docs/50)。
                         * 對話分支對齊資料:無胡椒→rec26(等胡椒);獻胡椒→rec28(吃胡椒)+ 授船。 */
                        set_dialogue_hero(&roster, &party);
                        if (ship.owned) {                        /* 已取船 → 仍顯示「吃胡椒」後話 */
                            dq3_dialogue_open(&dlg, DQ3_PORTOGA_REC_GOT);
                        } else if (dq3_inv_find(&inv, DQ3_ITEM_PEPPER) >= 0) {
                            dq3_inv_remove(&inv, DQ3_ITEM_PEPPER);   /* 獻上黑胡椒 */
                            ship.owned = 1; ship.aboard = 0;
                            ship.px = DQ3_PORTOGA_HARBOR_X; ship.py = DQ3_PORTOGA_HARBOR_Y; ship.layer = 0;
                            dq3_progress_set(&flags, DQ3_MS_SHIP);
                            dq3_dialogue_open(&dlg, DQ3_PORTOGA_REC_GOT);
                            fprintf(stderr, "★ 取船:獻黑胡椒給波魯多加國王 → 得船,停泊港口 (%d,%d)(SHIP 里程碑)\n",
                                    ship.px, ship.py);
                        } else {
                            /* 首訪授《國王的信》0x5b(杜 Ch16:拿給諾魯特→阿莎拉慕東洞 CTY62 開往東方通道→巴哈拉達)。
                             * flag 0x215 防重複;諾魯特 gate(dq3_scripted byte4=50)消耗此信。0x5b 真名=國王的信(非船,碼勘誤)。*/
                            if (!dq3_flags_get(&flags, 0x215)) {
                                dq3_inv_add(&inv, 0x5b); dq3_flags_set(&flags, 0x215, 1);
                                fprintf(stderr, "★ 波魯多加國王:授《國王的信》0x5b → 交諾魯特(阿莎拉慕東洞)開東方通道(杜 Ch16-17)\n");
                            }
                            dq3_dialogue_open(&dlg, DQ3_PORTOGA_REC_WAIT);   /* 等胡椒 */
                            fprintf(stderr, "波魯多加國王:等黑胡椒(無胡椒,未授船)\n");
                        }
                        talked = 1;
                    } else if (sub == 2 && cur_cty == DQ3_ROMALY_CTY && cur->section == DQ3_ROMALY_KING_SEC && dlg_ok &&
                               cur->npcs[ni].x == DQ3_ROMALY_KING_X && cur->npcs[ni].y == DQ3_ROMALY_KING_Y &&
                               dq3_inv_find(&inv, DQ3_ITEM_CROWN) >= 0) {
                        /* 羅馬利亞國王:持金皇冠 → 婉拒王位繼續(攻略建議保留皇冠不消耗)→ ROMALY 里程碑。 */
                        set_dialogue_hero(&roster, &party);
                        if (!dq3_progress_done(&flags, DQ3_MS_ROMALY)) {
                            dq3_progress_set(&flags, DQ3_MS_ROMALY);
                            fprintf(stderr, "★ 羅馬利亞國王:歸還金皇冠 → 婉拒王位,繼續冒險(ROMALY 里程碑)\n");
                        }
                        dq3_dialogue_open(&dlg, dq3_sub2_dialogue(b4));
                        talked = 1;
                    } else if (sub == 2 && cur_cty == 19 && b4 == 45 && dlg_ok) {
                        /* 八頭大蛇 boss 劇情(CTY19 洞穴,byte4=45,原版 flag 0x44 閘)。怪 75(0x4b)。
                         * remake:未討伐 → 開 boss 戰;勝利 → 設 flag 0x44 + 後話;已討伐 → 直接後話。 */
                        set_dialogue_hero(&roster, &party);
                        if (dq3_flags_get(&flags, 0x44)) {
                            dq3_dialogue_open(&dlg, dq3_sub2_dialogue(b4));
                        } else {
                            int oc; const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                            dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                            oc = dq3_battlescene_run(assets, 75, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                            dq3_scene_apply_palette(cur);
                            if (oc == 1) { dq3_flags_set(&flags, 0x44, 1);
                                if (dq3_inv_find(&inv, 0x69) < 0) dq3_inv_add(&inv, 0x69);  /* 戰勝得紫寶珠(青衫攻略:無姬大人=八頭大蛇)*/
                                if (dq3_inv_find(&inv, 0x14) < 0) dq3_inv_add(&inv, 0x14);  /* 草薙之劍 0x14(BBS:日邦格洞窟打八頭大蛇取草薙之劍)*/
                                fprintf(stderr, "★ 擊敗八頭大蛇(怪75)→ 獲得紫寶珠 0x69 + 草薙之劍 0x14 + flag 0x44\n"); }
                            else fprintf(stderr, "八頭大蛇戰 outcome=%d(未勝)\n", oc);
                            dq3_dialogue_open(&dlg, dq3_sub2_dialogue(b4));
                        }
                        talked = 1;
                    } else if (sub == 2 && cur_cty == 14 && b4 == 58 && dlg_ok) {
                        /* 甘達特盜賊巢穴(CTY14 sec1,巴哈拉達洞窟 boss 房,byte4=58)。原版=救人劇情過場:
                         * 守衛擋路「不能讓你們通過,看招吧」(D3TXT03 rec106-108)→ 打甘達特手下(怪27)→
                         * 牆上開關救出古布達/達妮亞 → 返回的甘達特(怪26)→ 結局(rec113-124)。
                         * EXE handler L0x5ee0 全程過場(對話+NPC動畫+warp),戰鬥為劇情強制戰。
                         * remake:flag 0x211=達妮亞救出(與 CTY15 古布達黑胡椒鏈 0x211 閉環,docs/boss-trigger-points)。 */
                        set_dialogue_hero(&roster, &party);
                        if (dq3_flags_get(&flags, 0x211)) {
                            dq3_dialogue_open(&dlg, dq3_sub2_dialogue(b4));   /* 已救出 → 後話 */
                            fprintf(stderr, "甘達特巢穴:已救出達妮亞(flag 0x211)\n");
                        } else {
                            int oc; const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                            dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                            fprintf(stderr, "★ 甘達特巢穴守衛:看招吧！→ 甘達特手下(怪27)\n");
                            oc = dq3_battlescene_run(assets, 27, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                            dq3_scene_apply_palette(cur);
                            if (oc == 1) {
                                fprintf(stderr, "★ 救出古布達/達妮亞 → 返回的甘達特(怪26)\n");
                                oc = dq3_battlescene_run(assets, 26, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                                dq3_scene_apply_palette(cur);
                            }
                            if (oc == 1) {
                                dq3_flags_set(&flags, 0x211, 1);
                                fprintf(stderr, "★ 擊敗甘達特 → 达妮亚获救(flag 0x211)，可回胡椒店找古布達\n");
                            } else fprintf(stderr, "甘達特巢穴戰 outcome=%d(未勝)\n", oc);
                            dq3_dialogue_open(&dlg, dq3_sub2_dialogue(b4));
                        }
                        talked = 1;
                    } else if (sub == 2 && cur_cty == 47 && b4 == 37 && dlg_ok) {
                        /* 勇氣神殿神父(RE handler logical 0x59e4,byte4=37;杜勝利 Ch29-30)。
                         * 原版:test_flag(0x13) 已證勇氣 → rec84「你們的勇氣」;否則 set [0xb34] + rec9「來的好」
                         * → region gate → rec10「那麼去吧」(接受)/ rec11「弱懦」。
                         * remake:接受 = set flag 0x13 —— **此 flag 同時驅動 owportal(dq3_owportal {82,165}→flag0x13→CTY75)**
                         * 把朗錫爾 overworld 入口轉到「神殿開放態」CTY75 → 通勇氣洞窟 CTY23(藍寶珠 0x67 已在 treasure 表)。
                         * 註:原版進洞強制單人([0x5077] 減員);remake 隨機遭遇僅地表、洞窟無戰 → solo 機制 moot,不偽造。 */
                        set_dialogue_hero(&roster, &party);
                        if (dq3_flags_get(&flags, 0x13)) {
                            dq3_dialogue_open(&dlg, 84);   /* 「你們的勇氣,…」(已接受)*/
                        } else {
                            dq3_flags_set(&flags, 0x13, 1);
                            dq3_dialogue_open(&dlg, 10);   /* 「那麼去吧,」→ 獨自進勇氣洞窟 */
                            fprintf(stderr, "★ 勇氣神殿神父:接受單獨戰鬥挑戰(flag 0x13)→ 朗錫爾 overworld portal 轉 CTY75 神殿開放 → 勇氣洞窟 CTY23(藍寶珠 0x67)\n");
                        }
                        talked = 1;
                    } else if (sub == 2 && cur_cty == 75 && b4 == 62 && dlg_ok) {
                        /* 勇氣神殿神父(神殿開放態,RE handler 0x608a,byte4=62):rec12「回來了嗎?」返回對話。
                         * 原版此處改隊伍人數 [0x5077](單獨戰鬥進/出);remake 洞窟無遭遇 → 純對話。 */
                        set_dialogue_hero(&roster, &party);
                        dq3_dialogue_open(&dlg, 12);       /* 「回來了嗎?」*/
                        fprintf(stderr, "勇氣神殿神父(CTY75 神殿開放):回來了嗎?\n");
                        talked = 1;
                    } else if (sub == 2 && cur_cty == 82 && b4 == 64 && dlg_ok) {
                        /* 不死鳥祠堂守護者(CTY82 sect17,byte4=64 祭壇 NPC):集六珠 0x66-0x6b(綠藍紅紫黃銀)
                         * → 復活不死鳥拉米亞(青衫攻略)。原版六祭壇位 byte4=63-68 僅 64 實際放置(build 不完整),
                         * remake 以此 NPC 當守護者統一判定。復活後按 y 在 overworld 起飛。 */
                        int orbs = 0, oi;
                        for (oi = 0x66; oi <= 0x6b; oi++) if (dq3_inv_find(&inv, oi) >= 0) orbs++;
                        set_dialogue_hero(&roster, &party);
                        if (phoenix_revived) {
                            fprintf(stderr, "不死鳥祠堂:拉米亞已復活(按 y 於 overworld 起飛)\n");
                        } else if (orbs >= 6) {
                            phoenix_revived = 1;
                            fprintf(stderr, "★ 不死鳥祠堂:六珠齊備(綠藍紅紫黃銀)→ 拉米亞復活!(出城至 overworld 按 y 起飛)\n");
                        } else {
                            fprintf(stderr, "不死鳥祠堂守護者:還缺寶珠(%d/6,需綠藍紅紫黃銀 0x66-0x6b)\n", orbs);
                        }
                        dq3_dialogue_open(&dlg, dq3_sub2_dialogue(b4));
                        talked = 1;
                    } else if (sub == 2 && dlg_ok && dq3_scripted_get(b4, cur_cty)) {
                        /* data-driven scripted 給物 NPC(dq3_scripted 表,Step 2,docs/re-log-722)。
                         * 前置道具未滿足→before_rec;持前置且未給→給物+里程碑+give_rec;已給→after_rec。 */
                        const dq3_scripted *sc = dq3_scripted_get(b4, cur_cty);
                        set_dialogue_hero(&roster, &party);
                        if (b4 == 25 && cur_cty == 15 && !dq3_flags_get(&flags, 0x211)) {
                            /* 古布達黑胡椒救人 gate(杜勝利 Ch19-20):需先到巴哈拉達洞窟打倒
                             * 甘達特手下(怪27)救出達妮亞(flag 0x211),古布達才回胡椒店免費給黑胡椒。
                             * 救人事件 = boss:27;boss:26;flag:0x211(洞窟雙 boss → 設 0x211)。 */
                            dq3_dialogue_open(&dlg, sc->before_rec);
                            fprintf(stderr, "古布達:還在巴哈拉達洞窟救達妮亞(打甘達特手下怪27)→ 黑胡椒未給\n");
                        } else if (sc->require_item != DQ3_SC_NOITEM) {
                            /* 檢查型 NPC(0x7c0c 檢查,不消耗):持物→success rec、否則→need rec。 */
                            if (dq3_inv_find(&inv, sc->require_item) >= 0) {
                                if (sc->milestone) dq3_progress_set(&flags, sc->milestone);
                                dq3_dialogue_open(&dlg, sc->give_rec);
                                fprintf(stderr, "scripted 檢查 NPC byte4=%d:持有 0x%02x → 反應 rec%d\n",
                                        b4, sc->require_item, sc->give_rec);
                            } else {
                                dq3_dialogue_open(&dlg, sc->before_rec);
                                fprintf(stderr, "scripted 檢查 NPC byte4=%d:缺 0x%02x → 需求 rec%d\n",
                                        b4, sc->require_item, sc->before_rec);
                            }
                        } else if (sc->prereq_item != DQ3_SC_NOITEM && dq3_inv_find(&inv, sc->prereq_item) < 0) {
                            dq3_dialogue_open(&dlg, sc->before_rec);
                        } else if (dq3_inv_find(&inv, sc->give_item) < 0) {
                            if (sc->consume_prereq && sc->prereq_item != DQ3_SC_NOITEM)
                                dq3_inv_remove(&inv, sc->prereq_item);   /* transform:消耗 prereq 換 give_item */
                            dq3_inv_add(&inv, sc->give_item);
                            if (sc->milestone) dq3_progress_set(&flags, sc->milestone);
                            dq3_dialogue_open(&dlg, sc->give_rec);
                            fprintf(stderr, "★ scripted NPC byte4=%d:%s道具 0x%02x(rec%d)\n",
                                    b4, sc->consume_prereq ? "換得" : "獲得", sc->give_item, sc->give_rec);
                        } else {
                            dq3_dialogue_open(&dlg, sc->after_rec);
                            fprintf(stderr, "scripted NPC byte4=%d:已給過 → 後話 rec%d\n", b4, sc->after_rec);
                        }
                        talked = 1;
                    } else if (sub == 2 && dlg_ok) {         /* scripted-event NPC:主對話 rec(旗標條件略)*/
                        int srec = dq3_sub2_dialogue(b4);
                        if (srec >= 0) {
                            set_dialogue_hero(&roster, &party);
                            if (dq3_dialogue_open(&dlg, srec) == 0) {
                                talked = 1;
                                fprintf(stderr, "scripted NPC@(%d,%d) byte4=%d → rec=0x%02x(旗標條件未模擬)\n",
                                        cur->npcs[ni].x, cur->npcs[ni].y, b4, srec);
                            }
                        }
                    } else if (sub >= 3) {                   /* 設施 NPC:走到店員 → 開該攤(docs/40)*/
                        const dq3_facility *fac = dq3_facility_at(cur_cty, cur->section, b4);
                        if (fac) {
                            talked = 1;
                            if ((fac->type == DQ3_FAC_WEAPON || fac->type == DQ3_FAC_ITEM) && shop_ok) {
                                const unsigned char *stock = &dq3_shop_itempool[fac->item_off];
                                int sn = fac->count;
                                unsigned char buf[16];     /* CTY15 道具店補進黑胡椒(早期 build 斷鏈補洞,docs/50)*/
                                if (cur_cty == DQ3_PEPPER_CTY && fac->type == DQ3_FAC_ITEM && sn < 15) {
                                    int z; for (z = 0; z < sn; z++) buf[z] = stock[z];
                                    buf[sn++] = DQ3_ITEM_PEPPER; stock = buf;
                                    fprintf(stderr, "(remake 補洞:本城道具店補進黑胡椒,docs/50)\n");
                                }
                                shop_modal(&roster, &party, &shop_items, sys_ok ? &sys_txt : &dlg.txt,
                                           &gold, stock, sn);
                                dq3_scene_apply_palette(cur);
                                fprintf(stderr, "設施:%s(CTY%d sec%d k%d,%d 品項)\n",
                                        fac->type==DQ3_FAC_WEAPON?"武器/防具店":"道具店",
                                        cur_cty, cur->section, b4, sn);
                            } else if (fac->type == DQ3_FAC_INN) {
                                inn_rest(&roster, &party, &gold, fac->inn_cost);   /* 治療 + 扣費(task1)*/
                                if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, 0x11a);  /* 旅社歡迎詞 */
                            } else if (fac->type == DQ3_FAC_CHURCH) {
                                int rev = church_revive(&roster, &party, &gold);   /* 蘇生陣亡隊員(task3)*/
                                /* 有復活 → 「誰需要復活」(rec 0x138);否則 → 「主持正義之人」(0x129)*/
                                if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, rev > 0 ? 0x138 : 0x129);
                            } else if (fac->type == DQ3_FAC_RECORD) {
                                /* 記錄點:開 6-slot 選單讓玩家選格存檔(使用者需求:至少 6 slot)。
                                 * ESC 取消 → 不存。選定 slot → 存檔 + 記錄對話。 */
                                int slot = slot_select(sys_ok ? &sys_txt : &dlg.txt, 0);
                                if (slot >= 1) {
                                    save_game_to(slot_path(slot), &roster, &party, &inv, &flags,
                                                 cur_cty, cur->px, cur->py, &ship,
                                                 in_town, layer, in_town ? cur->section : 0);
                                    fprintf(stderr, "記錄點:存入冒險之書 %d\n", slot);
                                    if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, 0xfd);   /* 記錄「旅行的經驗」*/
                                } else fprintf(stderr, "記錄點:取消存檔\n");
                                dq3_scene_apply_palette(cur);
                            }
                        }
                    }
                }
            }
            if (!talked)
            {
            int et2, ep2, ef2;
            /* 面向格優先(站在寶箱前調べる),其次本格 */
            int fdx2 = (cur->facing==3)-(cur->facing==1), fdy2 = (cur->facing==0)-(cur->facing==2);
            if (dq3_scene_tile_event_p2(cur, cur->px+fdx2, cur->py+fdy2, &et2, &ep2, &ef2) ||
                dq3_scene_tile_event_p2(cur, cur->px, cur->py, &et2, &ep2, &ef2)) {
                /* 寶箱 / 隱藏物品(docs/40):type 0/1/3 + param=道具 id。flag=一次性旗標,
                 * remake 用 flag bit 當「已取」標記(set=已取);取過再調べる回空。 */
                int is_item = (et2==0 || et2==1 || et2==3) && ep2 > 0 && ep2 < 0x90;
                /* 香巴尼之塔(CTY10)金皇冠正源(杜勝利 Ch6):原版要先打盜賊甘達特(怪26,6F)
                 * 才能取金皇冠。remake gate:取皇冠寶箱前,甘達特未敗(flag 0x210)→ 先 boss 戰;
                 * 勝利才放行寶箱。皇冠 → 還羅馬利亞國王(byte4=9/特例)→ ROMALY 里程碑。 */
                int gandata_gate = (is_item && ep2 == 0x33 && cur_cty == 10 && !dq3_flags_get(&flags, ef2));
                if (gandata_gate && !dq3_flags_get(&flags, 0x210)) {
                    int oc; const char *bs = getenv("DQ3_BATTLE_SCRIPT");
                    fprintf(stderr, "★ 香巴尼之塔:盜賊甘達特擋路!(杜勝利 Ch6)\n");
                    dq3_battlescene_set_party(party.count > 0 ? &roster : NULL, party.count > 0 ? &party : NULL);
                    oc = dq3_battlescene_run(assets, 26, 1, -1, bs ? bs : "FFFFFFFFFFFFFFFF", NULL, 1);
                    dq3_scene_apply_palette(cur);
                    if (oc == 1) { dq3_flags_set(&flags, 0x210, 1);
                        fprintf(stderr, "★ 擊敗甘達特(怪26)→ 可取金皇冠\n"); }
                    else fprintf(stderr, "甘達特戰 outcome=%d(未勝)→ 金皇冠未取\n", oc);
                }
                /* B-7 倉庫番 gate:CTY76 的乾渴壺 0x5e 在密道開啟前(三石未歸位)不可取。 */
                if (is_item && cur_cty == SOKOBAN_CTY && ep2 == 0x5e && !g_sokoban_solved) {
                    fprintf(stderr, "耶進貝亞倉庫番:密道未開 — 需先把三顆大石推到上方藍白地面(杜 Ch26)\n");
                    if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, 0xf3);   /* 暫借「空的」訊息 */
                } else if (is_item) {
                    /* 系統訊息 / 道具名走常駐 D3TXT00(sys_txt),非當前 section 對話 bank(docs/42)。 */
                    if (gandata_gate && !dq3_flags_get(&flags, 0x210)) {   /* 甘達特未敗 → 皇冠不可取 */
                        if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, 0xf3);
                    } else if (dq3_flags_get(&flags, ef2)) {    /* 已取過 → 空寶箱 */
                        if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, 0xf3);   /* 「可惜是空的。」*/
                    } else if (dq3_inv_add(&inv, ep2) >= 0) {   /* 入背包 + 標記已取 */
                        dq3_flags_set(&flags, ef2, 1);
                        fprintf(stderr, "寶箱: 獲得道具 0x%02x(旗標0x%02x 標記已取)\n", ep2, ef2);
                        if (sys_ok) dq3_dialogue_open_text(&dlg, &sys_txt, ep2 + 1);  /* 道具名 rec=code+1 */
                    } else if (sys_ok) {
                        dq3_dialogue_open_text(&dlg, &sys_txt, 0x13a);          /* 「行李滿了」*/
                    }
                } else if (et2 == 2) {        /* type-2 傳送(旅の扉/洞穴出口)→ warp 表 0x4ea0(docs/31/43)*/
                    int wc, wx, wy;
                    if (dq3_warp_get(ep2, &wc, &wx, &wy) == 0 && wc > 0 && wc < 100) {
                        char ct[16]; int bn = dq3x_map_blknum[wc]; dq3_scene *ns;
                        sprintf(ct, "CTY%02d.DAT", wc);
                        ns = dq3_town_load(assets, ct, 0, bn, err, sizeof err);
                        if (ns) {
                            if (town) dq3_scene_free(town);
                            town = ns; cur = town; in_town = 1; cur_cty = wc; load_field_hero(town, assets);
                            if (wx < cur->map_w) cur->px = wx;
                            if (wy < cur->map_h) cur->py = wy;
                            dq3_scene_apply_palette(cur);
                            fprintf(stderr, "傳送(旅の扉/洞):param=%d → CTY%d (%d,%d)\n", ep2, wc, wx, wy);
                        } else fprintf(stderr, "傳送載入失敗 CTY%d: %s\n", wc, err);
                    } else if (dlg_ok) dq3_dialogue_open(&dlg, 1);
                } else {
                    fprintf(stderr, "事件: type=%d param=0x%x\n", et2, ep2);
                    if (dlg_ok) { dlg_rec = (ep2 && ep2 < dlg.txt.n_records) ? ep2 : 1; dq3_dialogue_open(&dlg, dlg_rec); }
                }
            }
            }
        } else if (sc == 0x14 && in_town && cur_cty == 0) {  /* T:阿里阿罕酒場捷徑(開發用;正式入口=西側 LUIDA 櫃台)*/
            tavern_modal(assets, &roster, &party, &gst, &dlg.txt);
            set_dialogue_hero(&roster, &party);   /* 主角名可能改變 → 更新 {V} */
            dq3_scene_apply_palette(cur);
            fprintf(stderr, "離開酒場:名冊%d 人、隊伍%d 人\n", roster.count, party.count);
        } else if (sc == 0x2e) {            /* C:野外指令窗(命令)— 對話/咒文/狀況/道具/裝備/調查 */
            g_item_world_eff = 0;
            cmd_modal(cur, &roster, &party, &inv, &dlg, dlg_ok, sys_ok ? &sys_txt : &dlg.txt);
            dq3_scene_apply_palette(cur);
            if (g_item_world_eff == DQ3_USE_RETURN_TOWN) {       /* 蓋美拉翅膀 / 魯拉 / 烈米特:回地表 */
                layer = 0; cur = field; in_town = 0; cur_cty = -1; ship.aboard = 0;
                dq3_scene_apply_palette(cur);
                fprintf(stderr, "回到地表 (%d,%d)\n", cur->px, cur->py);
            } else if (g_item_world_eff == DQ3_USE_REPEL) {      /* 聖水:驅弱敵 */
                repel = DQ3_HOLY_STEPS;
                fprintf(stderr, "聖水:弱敵迴避 %d 步\n", repel);
            } else if (g_item_world_eff == DQ3_USE_AWAKEN) {     /* 覺醒粉:諾阿尼魯村(CTY4)解全村催眠 */
                if (in_town && cur_cty == 4) {
                    dq3_inv_remove(&inv, 0x5a);                  /* 消耗覺醒粉 */
                    dq3_flags_set(&flags, 0x31, 1);              /* 諾阿尼魯村已甦醒(原版催眠 flag 附近)*/
                    fprintf(stderr, "★ 諾阿尼魯村:使用覺醒粉 → 全村催眠詛咒解除!(杜勝利 Ch11)\n");
                } else {
                    fprintf(stderr, "覺醒粉:此處用不上(需在諾阿尼魯村 CTY4 使用),未消耗\n");
                }
            } else if (g_item_world_eff == DQ3_USE_GAIA) {       /* 蓋亞之劍:地表小火山旁 → 開通往尼羅肯特 */
                if (!in_town) { dq3_flags_set(&flags, 0x32, 1);
                    fprintf(stderr, "★ 蓋亞之劍:小火山熔流而出,開通往尼羅肯特洞窟之路!(杜勝利 Ch40)\n"); }
                else fprintf(stderr, "蓋亞之劍:需在地表小火山旁使用\n");
            } else if (g_item_world_eff == DQ3_USE_DRAIN) {      /* 乾渴壺:地表四島礁 → 吸海顯現祠堂 */
                if (!in_town) { dq3_flags_set(&flags, 0x33, 1);
                    fprintf(stderr, "★ 乾渴壺:吸乾海水,最終鑰匙祠堂顯現!(杜勝利 Ch27)\n"); }
                else fprintf(stderr, "乾渴壺:需在地表四島礁附近使用\n");
            } else if (g_item_world_eff == DQ3_USE_FAIRYFLUTE) { /* 妖精之笛:魯比斯之塔 CTY82 → 解詛咒得精靈的守護 */
                if (in_town && cur_cty == 82) {
                    if (dq3_inv_find(&inv, 0x74) < 0) dq3_inv_add(&inv, 0x74);
                    dq3_flags_set(&flags, 0x34, 1);
                    fprintf(stderr, "★ 魯比斯之塔:妖精之笛解魯比斯詛咒 → 得精靈的守護 0x74(→精靈祠堂換雲雨之杖)\n");
                } else fprintf(stderr, "妖精之笛:需在魯比斯之塔(CTY82)對魯比斯使用\n");
            } else if (g_item_world_eff == DQ3_USE_RAINBOW) {    /* 彩虹水滴:下層利姆達爾西北地表 → 架彩虹橋通終盤 */
                if (!in_town && layer == 1) {                    /* 下層 overworld */
                    dq3_inv_remove(&inv, 0x75);                  /* 消耗彩虹水滴 */
                    dq3_flags_set(&flags, 0x35, 1);              /* 彩虹橋已架旗標 */
                    dq3_progress_set(&flags, DQ3_MS_RAINBOW);    /* 推進 RAINBOW 里程碑 */
                    fprintf(stderr, "★ 彩虹水滴:雨和太陽合而為一,彩虹橋出現!通往終盤(杜勝利 Ch55)\n");
                } else fprintf(stderr, "彩虹水滴:需在下層利姆達爾西北盡頭地表使用\n");
            } else if (g_item_world_eff == DQ3_USE_RANARUTA) {   /* 拉那魯達:切換晝夜(白天↔黑夜)*/
                dq3_scene_set_daynight(dq3_scene_get_daynight() == 2 ? 0 : 2);
                g_dn_step = 0; dq3_scene_apply_palette(cur);
                fprintf(stderr, "★ 拉那魯達:晝夜切換 → %s\n", DN_NAME[dq3_scene_get_daynight()]);
            } else if (g_item_world_eff == DQ3_USE_DARK_LAMP) {  /* 黑暗之燈:變夜 */
                dq3_scene_set_daynight(2); g_dn_step = 0; dq3_scene_apply_palette(cur);
                fprintf(stderr, "★ 黑暗之燈:四周頓時暗了下來(變夜)\n");
            }
            g_item_world_eff = 0;
        } else if (sc == 0x30 && in_town) {  /* B:武器/防具商店捷徑(開發用;正式入口=走到店員 NPC)*/
            if (shop_ok) { int sn; const unsigned char *sk = shop_stock_for(cur_cty, &sn);
                shop_modal(&roster, &party, &shop_items, sys_ok ? &sys_txt : &dlg.txt, &gold, sk, sn); }
            dq3_scene_apply_palette(cur);
        } else if (sc == 0x16 && !in_town && phoenix_aboard) {  /* 不死鳥:降落於可走格 */
            if (dq3_scene_walkable(cur, cur->px, cur->py)) {
                phoenix_aboard = 0;
                fprintf(stderr, "不死鳥降落於 (%d,%d)\n", cur->px, cur->py);
            } else fprintf(stderr, "此處無法降落(非可走格,繼續飛行)\n");
        } else if (sc == 0x16 && !in_town && phoenix_revived) {  /* 不死鳥:起飛 */
            phoenix_aboard = 1;
            fprintf(stderr, "搭乘不死鳥拉米亞起飛(飛越山海)\n");
        } else if (sc == 0x16 && !in_town) {  /* U:下降(ギアガの大穴)/上升 */
            if (layer == 0) {
                /* 正式下降 = runner event 86(ギアガの大穴掉落,handler 0x783d)。原版觸發旗標為劇情動態、
                 * 無靜態 setter(docs/44 §RE 結論)→ 自製忠實 gate:需**巴拉摩斯已敗(flag 0x213)**,
                 * 對齊杜勝利 Ch44→46(打倒巴拉摩斯→索瑪現身→才下降アレフガルド)。debug `descent` token 不受此閘(測試用)。*/
                if (dq3_flags_get(&flags, 0x213))
                    do_descent(assets, &field_under, &cur, &layer, &in_town, &cur_cty, &flags);  /* scripted_event 86 */
                else
                    fprintf(stderr, "下降:ギアガの大穴尚未開啟 — 需先擊敗巴拉摩斯(索瑪現身,flag 0x213)\n");
            }
            else { layer = 0; cur = field; dq3_scene_apply_palette(cur); fprintf(stderr, "上升 → 地表\n"); }
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
            int moved;
            if (!in_town && phoenix_aboard) {        /* 不死鳥飛行:無視地形,飛越山海,只夾邊界 */
                int dx = (sc==0x4d)-(sc==0x4b), dy = (sc==0x50)-(sc==0x48);
                int nx = cur->px + dx, ny = cur->py + dy;
                cur->facing = (sc==0x48)?2:(sc==0x50)?0:(sc==0x4b)?1:3;
                moved = 0;
                if (nx >= 0 && nx < cur->map_w && ny >= 0 && ny < cur->map_h) {
                    cur->px = nx; cur->py = ny; moved = 1;
                    /* 飛到可走格上空按 Space 降落(見下);移動本身不落地 */
                }
            } else if (!in_town) {                   /* overworld:走船規則(航行/登船/上岸,docs/48)*/
                int r = dq3_ship_input(cur, &ship, sc, layer);
                moved = (r != DQ3_SHIP_BLOCKED);
                if (r == DQ3_SHIP_BOARD)          fprintf(stderr, "登船 → 可跨海\n");
                else if (r == DQ3_SHIP_DISEMBARK) fprintf(stderr, "上岸 → 船停泊 (%d,%d) layer%d\n", ship.px, ship.py, ship.layer);
            } else {
                /* B-7 倉庫番推石(CTY76):面向石頭 NPC(b2==40)→ 推到下一格(可走+無 NPC)+ 玩家跟進。
                 * 三石歸藍白地面 → g_sokoban_solved 開密道(乾渴壺 examine 才放行)。 */
                int sdx = (sc==0x4d)-(sc==0x4b), sdy = (sc==0x50)-(sc==0x48);
                int stx = cur->px+sdx, sty = cur->py+sdy;
                int sni = (cur_cty == SOKOBAN_CTY) ? dq3_scene_npc_at(cur, stx, sty) : -1;
                if (sni >= 0 && cur->npcs[sni].b2 == 40) {
                    int psx = stx+sdx, psy = sty+sdy;
                    cur->facing = (sc==0x48)?2:(sc==0x50)?0:(sc==0x4b)?1:3;
                    if (dq3_scene_walkable(cur, psx, psy) && dq3_scene_npc_at(cur, psx, psy) < 0) {
                        dq3_npc_move(cur, cur->npcs, sni, psx, psy);   /* 推石頭 */
                        cur->px = stx; cur->py = sty; moved = 1;
                        if (!g_sokoban_solved) {                       /* 三石歸位判定 */
                            int on = 0, k, ti;
                            for (k = 0; k < cur->n_npcs; k++) if (cur->npcs[k].b2 == 40)
                                for (ti = 0; ti < 3; ti++)
                                    if (cur->npcs[k].x == SOKOBAN_TGT[ti][0] && cur->npcs[k].y == SOKOBAN_TGT[ti][1]) on++;
                            if (on >= 3) { g_sokoban_solved = 1;
                                fprintf(stderr, "★ 耶進貝亞倉庫番:三石歸藍白地面 → 密道開啟!可取乾渴壺(杜 Ch26)\n"); }
                        }
                    } else moved = 0;   /* 石頭被牆/石擋,推不動 */
                } else {
                    moved = dq3_scene_input(cur, sc);   /* 對話中不在此分支(上面已攔)*/
                }
            }
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
                    if (dsec == 0xff || dsec == 0xfe) {  /* 出城:0xff→地表 / 0xfe→下層(docs/43)*/
                        int to_under = (dsec == 0xfe);
                        if (to_under && !field_under) {  /* 首次需下層 → 懶載 DQ3UND.MAP */
                            field_under = dq3_field_load_map(assets, "DQ3UND.MAP", err, sizeof err);
                            if (field_under) load_field_hero(field_under, assets);
                            else fprintf(stderr, "下層 overworld 載入失敗: %s\n", err);
                        }
                        if (town) { dq3_scene_free(town); town = NULL; }
                        layer = (to_under && field_under) ? 1 : 0;
                        cur = layer ? field_under : field; in_town = 0; cur_cty = -1;
                        if (dx && dx < cur->map_w) cur->px = dx;   /* spawn(0,0)=用既有位置 */
                        if (dy && dy < cur->map_h) cur->py = dy;
                        dq3_scene_apply_palette(cur);
                        fprintf(stderr, "出城 → %s overworld spawn(%d,%d)\n", layer ? "下層" : "地表", dx, dy);
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
                /* overworld 走到城鎮入口座標 → 進該 CTY(0x748 查表;依當前 layer 地表/下層)。
                 * 不死鳥飛行中飛越城鎮不進城(需先降落 y 再走入)。 */
                if (!in_town && !phoenix_aboard) {
                    int cidx = find_cty_at_map(cur->px, cur->py, layer);
                    /* 旗標條件 portal:同一 overworld 點依進度載不同城(docs/44 §7);override find_cty */
                    { int pc = dq3_owportal_resolve(cur->px, cur->py, &flags);
                      if (pc >= 0) { if (cidx != pc) fprintf(stderr, "portal:(%d,%d) 旗標條件 → CTY%d(原 %d)\n", cur->px, cur->py, pc, cidx); cidx = pc; } }
                    if (cidx >= 0) {
                        char cty[16]; int bn = dq3x_map_blknum[cidx];   /* 每CTY BLK號(0x0a04)*/
                        sprintf(cty, "CTY%02d.DAT", cidx);
                        town = dq3_town_load(assets, cty, 0, bn, err, sizeof err);
                        if (town) { load_field_hero(town, assets); cur = town; in_town = 1; cur_cty = cidx;
                            dq3_scene_apply_palette(cur);
                            fprintf(stderr, "入城:%s(%d,%d) → %s(BLK%d)事件數=%d%s\n",
                                    layer ? "下層" : "地表", cur->px, cur->py, cty, bn, town->n_events,
                                    cidx==DQ3_SHRINE_CTY ? "  [合成祠堂]" : ""); }
                        else fprintf(stderr, "入城失敗 %s: %s\n", cty, err);
                    }
                }
                if (moved && !in_town && repel > 0) repel--;   /* #3 聖水驅敵期間每步遞減 */
                if (moved && !in_town && ++g_dn_step >= DN_PHASE_STEPS) {  /* 晝夜:地表步數推進相位(原版機制=步數驅動)*/
                    g_dn_step = 0;
                    dq3_scene_set_daynight((dq3_scene_get_daynight() + 1) & 3);
                    dq3_scene_apply_palette(cur);              /* 立即重套(城鎮無 animate_sea 也生效)*/
                    fprintf(stderr, "晝夜:天色轉為%s\n", DN_NAME[dq3_scene_get_daynight()]);
                }
                if (moved) {                                   /* 中毒:每步扣 HP(不致死,留 1)*/
                    int pi2; for (pi2 = 0; pi2 < party.count; pi2++) {
                        dq3_member *pm2 = &roster.list[party.slot[pi2]].m;
                        if ((pm2->status & DQ3_STATUS_POISON) && pm2->cur_hp > 1) pm2->cur_hp--;
                    }
                }
                if (moved && !in_town && !ship.aboard && repel <= 0 && --enc <= 0) {  /* 船上/聖水期間不遇敵 */
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
    /* 腳本輸入 playthrough 結束 → dump 末幀(headless 驗證,docs/46)*/
    if (dump && getenv("DQ3_INPUT")) {
        dq3_scene_render(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        overlay_opened_chests(cur, &flags);    /* #4:已取寶箱疊開過標記(remake 增強)*/
        draw_ship_overlay(cur, &ship, in_town, layer);   /* 船 sprite(docs/51)*/
        if (!in_town && phoenix_aboard && phoenix_ok)    /* 不死鳥坐騎 sprite(末幀渲染)*/
            dq3_scene_draw_charsprite_at(cur, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H,
                                         cur->px, cur->py, &phoenix_spr, cur->facing & 3);
        if (dlg_ok && dq3_dialogue_is_open(&dlg)) dq3_dialogue_render(&dlg, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H);
        dq3_present();
        if (dq3_dump_ppm(dump) == 0) fprintf(stderr, "playthrough 末幀 -> %s(in_town=%d cty=%d (%d,%d))\n",
                                              dump, in_town, cur_cty, cur->px, cur->py);
    }
    if (dlg_ok) dq3_dialogue_free(&dlg);
    if (sys_ok) dq3_text_free(&sys_txt);
    if (town) dq3_scene_free(town);
    if (field_under) dq3_scene_free(field_under);
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
            dq3_setenv("DQ3_SHOP_DUMP", dump);
            { int sn; const unsigned char *sk = shop_stock_for(0, &sn);
              shop_modal(&roster, &party, &shi, &t, &g, sk, sn); }
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

/* 裝備一覽 modal:列隊伍各成員的 武器 / 防具(道具名 rec=code+1),ESC 離開。 */
/* slot(0=武器 cat&0x20 / 1=防具 cat&0x40)→ 背包中可裝備清單(首項 -1=卸下)。回 count。 */
/* slot(0武器/1鎧/2盾/3兜)→ 背包中該部位可裝備清單(首項 -1=卸下)。回 count。
 * 部位用 dq3_item_equip_slot(RE b4 高位),取代舊的 catmask(會把盾 0x60 誤判成鎧)。 */
static int equip_build_list(const dq3_inventory *inv, int cls, int slot, int *codes, int maxn)
{
    int n = 0, k;
    codes[n++] = -1;   /* 卸下 */
    if (!g_equip_items_ok) return n;
    for (k = 0; k < DQ3_INV_SLOTS && n < maxn; k++) {
        int c = inv->slot[k];
        if (c == DQ3_ITEM_NONE) continue;
        if (dq3_item_equip_slot(g_equip_items, c) == slot && dq3_item_can_equip(g_equip_items, c, cls))
            codes[n++] = c;
    }
    return n;
}

/* 裝備管理 modal:選隊員 → 選槽(武器/鎧/盾/兜 4 槽,RE b4 部位)→ 從背包換裝 / 卸下。
 * 裝備與背包分離(同商店模型);換裝時舊品回背包、新品出背包。inv/items 走 g_equip_*。 */
static void equip_modal(dq3_roster *roster, dq3_party *party, const dq3_text *text)
{
    /* 4 槽標籤 glyph(武器/鎧/盾/頭;兜字缺用頭 226;0xffff=單字終止)。 */
    static const uint16_t SLOT_LBL[4][2] = { {108,403}, {212,0xffff}, {238,0xffff}, {226,0xffff} };
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg, yellow; uint8_t *fb = dq3_fb();
    int wx = 24, wy = 30, ww = 320, wh = 230, i, j, s;
    int members[DQ3_PARTY_MAX], n = 0;
    int state = 0, msel = 0, ssel = 0, isel = 0;   /* 0=選人 1=選槽 2=選品;ssel 0..3 */
    int codes[32], ncodes = 0;
    for (i = 0; i < DQ3_PARTY_MAX; i++)
        if (party->slot[i] >= 0 && party->slot[i] < roster->count) members[n++] = party->slot[i];
    if (n == 0) return;
    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32); yellow = pal_near2(pal,pn,255,255,0);
    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        dq3_recruit *rc = &roster->list[members[msel < n ? msel : 0]];
        unsigned char *slots[4] = { &rc->weapon, &rc->armor, &rc->shield, &rc->head };
        if (sc == 0x01 || sc == DQ3_SC_F10) { if (state == 0) break; state--; }
        else if (sc == 0x50) { if (state==0) msel=(msel+1)%n; else if (state==1) ssel=(ssel+1)&3; else if (ncodes>0) isel=(isel+1)%ncodes; }
        else if (sc == 0x48) { if (state==0) msel=(msel+n-1)%n; else if (state==1) ssel=(ssel+3)&3; else if (ncodes>0) isel=(isel+ncodes-1)%ncodes; }
        else if (sc == 0x1c) {                       /* Enter */
            if (state == 0) { state = 1; ssel = 0; }
            else if (state == 1) { ncodes = equip_build_list(g_equip_inv, rc->m.cls, ssel, codes, 32); isel = 0; state = 2; }
            else if (state == 2 && g_equip_inv) {    /* 換裝 / 卸下 */
                unsigned char *slotp = slots[ssel];
                int newc = codes[isel];
                static const char *SN[4] = {"武器","鎧","盾","兜"};
                if (newc < 0) { if (*slotp != 0xff) { dq3_inv_add(g_equip_inv, *slotp); *slotp = 0xff; fprintf(stderr, "裝備:隊員%d %s 卸下\n", msel, SN[ssel]); } }
                else { if (*slotp != 0xff) dq3_inv_add(g_equip_inv, *slotp);   /* 舊品回背包 */
                       dq3_inv_remove(g_equip_inv, newc); *slotp = (unsigned char)newc;
                       fprintf(stderr, "裝備:隊員%d %s ← 道具0x%02x\n", msel, SN[ssel], newc); }
                state = 1;
            }
        }
        /* 渲染:每員 名字 + 4 槽(2×2:武器/鎧 上、盾/兜 下)*/
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        for (i = 0; i < n; i++) {
            dq3_recruit *r2 = &roster->list[members[i]];
            unsigned char *r2slots[4] = { &r2->weapon, &r2->armor, &r2->shield, &r2->head };
            int yy = wy + 6 + i * 52, cx = wx + 26;
            uint8_t nc = (i == msel) ? (uint8_t)yellow : (uint8_t)white;
            if (i == msel && state == 0)   /* ► 選人游標 */
                dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+8, yy, 0x29, (uint8_t)yellow);
            for (j = 0; j < r2->name_len; j++)
                dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, cx + j*DQ3_GLYPH_PX, yy, r2->name[j], nc);
            for (s = 0; s < 4; s++) {     /* 4 槽 2×2 */
                int col = s & 1, row = s >> 1;
                int sx = wx + 40 + col * 150, sy = yy + 14 + row * 16;
                uint8_t sc2 = (i==msel && state>=1 && ssel==s) ? (uint8_t)yellow : (uint8_t)white;
                int gx = sx;
                dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, gx, sy, SLOT_LBL[s][0], sc2); gx += DQ3_GLYPH_PX;
                if (SLOT_LBL[s][1] != 0xffff) { dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, gx, sy, SLOT_LBL[s][1], sc2); gx += DQ3_GLYPH_PX; }
                gx += DQ3_GLYPH_PX/2;
                if (*r2slots[s] != 0xff) dq3_text_draw_record(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, gx, sy, 6, 1, *r2slots[s]+1, sc2);
                else dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, gx, sy, 495, sc2);
            }
        }
        if (state == 2) {   /* 換裝子清單(右下疊窗)*/
            int lx = wx+150, ly = wy+wh-90, lw = 150, lh = 84;
            tav_window(fb, lx, ly, lw, lh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
            for (i = 0; i < ncodes && i < 4; i++) {
                int yy = ly + 8 + i*18;
                uint8_t cc = (i==isel) ? (uint8_t)yellow : (uint8_t)white;
                if (i==isel) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, lx+6, yy, 0x29, (uint8_t)yellow);
                if (codes[i] < 0) { dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, lx+22, yy, 495, cc); }  /* 卸下=「無」glyph 495 */
                else dq3_text_draw_record(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, lx+22, yy, 7, 1, codes[i]+1, cc);
            }
        }
        dq3_present(); dq3_delay_ms(16);
        if (getenv("DQ3_DUMP") && getenv("DQ3_EQUIP_DUMP")) { dq3_dump_ppm(getenv("DQ3_DUMP")); break; }   /* 一幀 dump 驗證 */
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

/* per-town 商店商品:直接取自 CTY*.DAT 設施表(dq3_shopdata,tools/gen_shop_data.py 自動產生)。
 * RE:NPC byte4 → section header +6 設施表 → block[type][count][item...](docs/40)。
 * remake 單一商店入口 → 合併該城「武器/防具店 + 道具店」品項;該城無商店時退回阿里阿罕(CTY0)。 */
static const unsigned char *shop_stock_for(int cty, int *n)
{
    static unsigned char buf[48];
    const unsigned char *w, *it; int nw, ni, k = 0, i;
    nw = dq3_shop_items(cty, DQ3_FAC_WEAPON, &w);
    ni = dq3_shop_items(cty, DQ3_FAC_ITEM, &it);
    for (i = 0; i < nw && k < 46; i++) buf[k++] = w[i];
    for (i = 0; i < ni && k < 46; i++) buf[k++] = it[i];
    if (k == 0) {                                   /* 該城無商店 → 退回阿里阿罕 */
        nw = dq3_shop_items(0, DQ3_FAC_WEAPON, &w);
        for (i = 0; i < nw && k < 46; i++) buf[k++] = w[i];
        ni = dq3_shop_items(0, DQ3_FAC_ITEM, &it);
        for (i = 0; i < ni && k < 46; i++) buf[k++] = it[i];
    }
    *n = k; return buf;
}

/* 武器/防具/道具商店 modal:列該城商品(名+價+能否裝備),Enter 購買→裝給領頭隊員,Tab 切買賣,ESC 離開。
 * 商品清單 per-town(shop_stock_for);道具名 = D3TXT00 rec=code+1;價/類別/職業由 ITEM.DAT(docs/22)。 */
static void shop_modal(dq3_roster *roster, dq3_party *party, const dq3_items *items, const dq3_text *text, long *gold, const unsigned char *STOCK, int n)
{
    int cur = 0, lead_ri, mode = 0;   /* mode 0=買 1=賣 */
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
        int sellc[2]; int nsell = 0;
        sellc[0] = rc->weapon; sellc[1] = rc->armor;        /* 賣出清單 = 已裝武器/防具 */
        nsell = 2;
        if (sc == 0x01 || sc == DQ3_SC_F10) break;
        else if (sc == 0x0f) { mode ^= 1; cur = 0; }        /* Tab:買 ↔ 賣 */
        else if (sc == 0x48) cur = (cur + (mode?nsell:n) - 1) % (mode?nsell:n);
        else if (sc == 0x50) cur = (cur + 1) % (mode?nsell:n);
        else if (sc == 0x1c || sc == 0x39) {                  /* Enter/Space:買 / 賣 */
            if (!mode) {                                       /* 買 */
                int code = STOCK[cur], price = dq3_item_price(items, code), cat = dq3_item_category(items, code);
                if (*gold >= price && (code == 0x41 || dq3_item_can_equip(items, code, rc->m.cls))) {
                    *gold -= price;
                    if (cat & 0x40) rc->armor = (unsigned char)code;
                    else if (cat & 0x20) rc->weapon = (unsigned char)code;
                    fprintf(stderr, "購買並裝備 道具0x%02x(花 %d,餘 %ld)\n", code, price, *gold);
                } else fprintf(stderr, "金錢不足或無法裝備。\n");
            } else {                                           /* 賣(半價、卸下)*/
                int code = sellc[cur];
                if (code != 0xff) {
                    int got = dq3_item_price(items, code) / 2; *gold += got;
                    if (cur == 0) rc->weapon = 0xff; else rc->armor = 0xff;
                    fprintf(stderr, "賣出 道具0x%02x(得 %d,餘 %ld)\n", code, got, *gold);
                } else fprintf(stderr, "沒有可賣的裝備。\n");
            }
        }
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        draw_number_at(fb, text, wx+14, wy+10, (int)*gold, (uint8_t)yellow);   /* 金錢 */
        /* 模式標題:買(道具 0x192) / 賣(賣 rec…)— 簡化以游標色區分,log 已標 */
        if (!mode) for (i = 0; i < n; i++) {                   /* 買清單 */
            int yy = wy + 34 + i * 18, code = STOCK[i];
            uint8_t fg = (i == cur) ? yellow : (uint8_t)white;
            int ok = (code == 0x41 || dq3_item_can_equip(items, code, rc->m.cls));
            if (i == cur) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+4, yy, 11, yellow);
            dq3_text_draw_record(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+22, yy, 8, 1, code+1, ok?fg:redc);
            draw_number_at(fb, text, wx+22+7*DQ3_GLYPH_PX, yy, dq3_item_price(items, code), (uint8_t)white);
        }
        else for (i = 0; i < nsell; i++) {                     /* 賣清單(武器/防具,半價)*/
            int yy = wy + 34 + i * 18, code = sellc[i];
            uint8_t fg = (i == cur) ? yellow : (uint8_t)white;
            if (i == cur) dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+4, yy, 11, yellow);
            if (code != 0xff) {
                dq3_text_draw_record(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+22, yy, 8, 1, code+1, fg);
                draw_number_at(fb, text, wx+22+7*DQ3_GLYPH_PX, yy, dq3_item_price(items, code)/2, (uint8_t)white);
            } else dq3_text_draw_glyph(text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+22, yy, 495, fg);  /* 無 */
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
 * 對話/調查 = 面向格事件→對話(祠堂/酒場特例仍在 Enter);咒文=野外施放、狀況=能力值、道具=item_modal、
 * 裝備=equip_modal(換裝/卸下)皆已實作。回傳選定指令或 -1。 */
/* 道具欄 modal:列出持有道具(名稱 = D3TXT00 rec=code+1),ESC 離開。 */
/* 對隊伍套用一個道具(野外つかう):heal=治第一個受傷者;cure=解第一個對應異常者。
 * 用到 → 消耗並回 1;不可用(滿/無異常/非野外道具)回 0。 */
static int field_use_item(dq3_inventory *inv, dq3_roster *r, dq3_party *p, int code)
{
    int kind = dq3_item_use_kind(code), i;
    if (kind == DQ3_USE_HEAL_HP) {
        for (i = 0; i < p->count; i++) {
            dq3_member *m = &r->list[p->slot[i]].m;
            if (dq3_item_use_heal(m, code) > 0) {
                dq3_inv_remove(inv, code);
                fprintf(stderr, "野外つかう:藥草 → 隊員%d HP %d/%d\n", i, m->cur_hp, m->stat[DQ3_STAT_HP]);
                return 1;
            }
        }
        return 0;
    }
    if (kind == DQ3_USE_CURE_POISON || kind == DQ3_USE_CURE_PARALYSIS) {
        for (i = 0; i < p->count; i++)
            if (dq3_item_use_cure(&r->list[p->slot[i]].m, code)) { dq3_inv_remove(inv, code); return 1; }
        return 0;
    }
    if (kind == DQ3_USE_PRAYER_RING) {            /* 祈禱之戒:回 MP + 每次使用 ~25.4% 損壞(#7c)*/
        static dq3_rng prng; static int seeded = 0;
        int restored = -1, broke;
        if (!seeded) { dq3_rng_seed(&prng, 0x1357); seeded = 1; }
        for (i = 0; i < p->count; i++) {
            int r2 = dq3_item_use_prayer_mp(&r->list[p->slot[i]].m);
            if (r2 > 0) { restored = r2;
                fprintf(stderr, "野外つかう:祈禱之戒 → 隊員%d MP %d/%d\n", i,
                        r->list[p->slot[i]].m.cur_mp, r->list[p->slot[i]].m.stat[DQ3_STAT_MP]); break; }
        }
        if (restored < 0) fprintf(stderr, "野外つかう:祈禱之戒(無人 MP 不足,仍擲損壞)\n");
        broke = (dq3_rng_next(&prng, 256) <= DQ3_PRAYER_BREAK_LE);   /* RE 0x5af9 cmp al,0x40 → ~25.4% */
        if (broke) { dq3_inv_remove(inv, code); fprintf(stderr, "野外つかう:啊,祈禱之戒壞了。\n"); }
        return 0;                                  /* 無世界層效果(回復/損壞已就地處理)*/
    }
    if (kind == DQ3_USE_DARK_LAMP) {              /* 黑暗之燈:變夜(不消耗,可重用)*/
        fprintf(stderr, "野外つかう:黑暗之燈 → 變夜\n");
        return kind;
    }
    if (kind == DQ3_USE_RETURN_TOWN || kind == DQ3_USE_REPEL) {   /* 需世界狀態 → 回傳碼交 main */
        dq3_inv_remove(inv, code);
        fprintf(stderr, "野外つかう:%s\n", kind == DQ3_USE_RETURN_TOWN ? "蓋美拉翅膀(回地表)" : "聖水(驅弱敵)");
        return kind;
    }
    return 0;
}

/* 野外道具選單:游標選取 + Enter 使用。heal/cure 就地套用;蓋美拉翅膀/聖水回傳效果碼
 * (DQ3_USE_RETURN_TOWN/REPEL)交 main 處理世界狀態。回 0=無世界效果。inv 非 const(會消耗)。 */
static int item_modal(dq3_inventory *inv, const dq3_text *text, dq3_roster *roster, dq3_party *party)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t *fb = dq3_fb();
    int wx = 30, wy = 50, ww = 290, wh = 200;
    int cursor = 0, world_eff = 0;
    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return 0;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32);
    while (!dq3_should_quit()) {
        int codes[DQ3_INV_SLOTS], n = 0, i, sx, sy;
        uint8_t sc;
        for (i = 0; i < DQ3_INV_SLOTS; i++) if (inv->slot[i] != DQ3_ITEM_NONE) codes[n++] = inv->slot[i];
        if (n == 0) cursor = 0; else if (cursor >= n) cursor = n - 1; else if (cursor < 0) cursor = 0;
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_status_render_items(inv, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+14, wy+12, (uint8_t)white);
        if (n > 0) {   /* 游標:選中道具左側畫實心方塊(排版同 render_items:2 欄、起 y+20+0、row 17)*/
            int col = cursor & 1, rowi = cursor >> 1, px, py, dx, dy;
            sx = (wx+14) + col * (8 * DQ3_GLYPH_PX); sy = (wy+12+20) + rowi * 17;
            px = sx - 11; py = sy + 2;
            for (dy = 0; dy < 12; dy++) for (dx = 0; dx < 8; dx++) {
                int xx = px+dx, yy = py+dy;
                if (xx>=0 && xx<DQ3_SCREEN_W && yy>=0 && yy<DQ3_SCREEN_H) fb[yy*DQ3_SCREEN_W+xx] = (uint8_t)white;
            }
        }
        dq3_present(); dq3_delay_ms(16);
        sc = dq3_poll_scancode();
        if (sc == 0x01 || sc == DQ3_SC_F10) break;
        else if (sc == 0x50 && cursor + 2 < n) cursor += 2;      /* 下 */
        else if (sc == 0x48 && cursor - 2 >= 0) cursor -= 2;      /* 上 */
        else if (sc == 0x4d && cursor + 1 < n) cursor += 1;      /* 右 */
        else if (sc == 0x4b && cursor - 1 >= 0) cursor -= 1;      /* 左 */
        else if (sc == 0x1c && n > 0) {                          /* Enter 使用 */
            int k = field_use_item(inv, roster, party, codes[cursor]);
            if (k == DQ3_USE_RETURN_TOWN || k == DQ3_USE_REPEL || k == DQ3_USE_AWAKEN || k == DQ3_USE_GAIA || k == DQ3_USE_DRAIN || k == DQ3_USE_FAIRYFLUTE || k == DQ3_USE_RAINBOW || k == DQ3_USE_DARK_LAMP) { world_eff = k; break; }  /* 交 main */
        }
    }
    return world_eff;
}

/* 達瑪神官轉職選單:選隊員 → 選新職業 → dq3_member_change_class(勇者不可轉/不可轉成勇者)。 */
static void dhama_modal(dq3_roster *roster, dq3_party *party, const dq3_stats *gst, const dq3_text *text, dq3_inventory *inv)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t yellow; uint8_t *fb = dq3_fb();
    int wx = 40, wy = 50, ww = 220, wh = 180, i;
    dq3_menu mem_menu, cls_menu;
    int state = 0, chosen = -1;
    if (party->count <= 0) return;
    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32); yellow = (uint8_t)pal_near2(pal,pn,255,255,0);
    dq3_menu_init(&mem_menu, wx+24, wy+28);
    for (i = 0; i < party->count; i++) {
        dq3_recruit *rc = &roster->list[party->slot[i]];
        dq3_menu_add(&mem_menu, rc->name, rc->name_len);
    }
    dq3_roster_class_menu(&cls_menu, wx+24, wy+28);
    while (!dq3_should_quit()) {
        uint8_t sc = dq3_poll_scancode();
        if (state == 0) {                                  /* 選隊員 */
            int r = sc ? dq3_menu_input(&mem_menu, sc) : -1;
            if (r == -2) break;
            if (r >= 0) { chosen = r; state = 1; cls_menu.cursor = 0; }
        } else {                                           /* 選職業 */
            int r = sc ? dq3_menu_input(&cls_menu, sc) : -1;
            if (r == -2) { state = 0; }
            else if (r >= 0) {
                dq3_member *m = &roster->list[party->slot[chosen]].m;
                int was_yusha = (m->cls == 7);   /* 遊人免領悟之書 */
                /* 賢者(class 5)需領悟之書,或本身是遊人(免書;杜勝利 Ch21)。
                 * 領悟之書真碼 = 0x4a(★2026-06-27 碼勘誤:原寫 0x40 實為勇者之盾,認錯道具)。
                 * 取得 = 加爾那之塔 CTY18 sec1 寶箱(攻略反證:CTY18 = 力量種子+鐵頭盔+領悟之書)。 */
                if (r == 5 && !was_yusha && dq3_inv_find(inv, 0x4a) < 0) {
                    fprintf(stderr, "達瑪:轉賢者需《領悟之書》(或遊人免書)— 加爾那之塔 CTY18 取得\n");
                    break;
                }
                if (dq3_member_change_class(m, gst, r) == 0) {
                    fprintf(stderr, "★ 達瑪轉職:隊員%d → 職業%d%s(Lv1,屬性減半)\n", chosen, r,
                            r == 5 ? "(賢者)" : "");
                    if (r == 5 && !was_yusha && dq3_inv_find(inv, 0x4a) >= 0) {
                        dq3_inv_remove(inv, 0x4a);   /* 領悟之書 0x4a 轉一次賢者後消耗(原版語意)*/
                        fprintf(stderr, "  《領悟之書》已用(消耗)\n");
                    }
                } else fprintf(stderr, "達瑪:勇者不可轉職 / 不可轉成勇者\n");
                break;
            }
        }
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        dq3_menu_render(state ? &cls_menu : &mem_menu, text, fb, DQ3_SCREEN_W, DQ3_SCREEN_H,
                        (uint8_t)white, yellow);
        dq3_present(); dq3_delay_ms(16);
    }
}

/* 野外可施放的咒文(base==0 工具咒,docs/data/spell-effects-research.md)。
 * eff = 回給主迴圈的世界效果碼;mp = 消耗(classic 值,FC/SFC 標準,精確未逐一 RE)。 */
static const struct { unsigned short rec; int mp; int eff; } FIELD_SPELLS[] = {
    { 172, 8, DQ3_USE_RETURN_TOWN },   /* 魯拉 ルーラ:飛行回城(remake → 回地表)*/
    { 173, 6, DQ3_USE_RETURN_TOWN },   /* 烈米特 リレミト:出迷宮/塔 → 回地表 */
    { 176, 2, DQ3_USE_REPEL },         /* 特黑洛斯 トヘロス:暫時驅弱敵 */
    { 177, 8, DQ3_USE_RANARUTA },      /* 拉那魯達 ラナルータ:切換晝夜 */
};
#define N_FIELD_SPELLS ((int)(sizeof FIELD_SPELLS / sizeof FIELD_SPELLS[0]))

/* 隊伍中是否有人會 rec 這個咒(且回該人 index;無回 -1)。 */
static int party_knows_spell(const dq3_roster *r, const dq3_party *p, unsigned short rec)
{
    int i, j; unsigned short recs[64];
    for (i = 0; i < p->count; i++) {
        const dq3_member *m = &r->list[p->slot[i]].m;
        int n = dq3_spells_known(m->cls, m->level, recs, 64);
        for (j = 0; j < n; j++) if (recs[j] == rec) return i;
    }
    return -1;
}

/* 野外咒文施放 modal:列隊伍會的工具咒(魯拉/烈米特/特黑洛斯)→ 游標選 → 扣 MP → 回世界效果碼。
 * 回 DQ3_USE_RETURN_TOWN / DQ3_USE_REPEL / 0(取消或無可施放 / MP 不足)。 */
static int field_cast_modal(dq3_roster *roster, dq3_party *party, const dq3_text *sys)
{
    dq3_color pal[256]; int pn; uint8_t *raw; size_t rl;
    int white, black, frame, bg; uint8_t *fb = dq3_fb();
    int wx = 60, wy = 60, ww = 200, wh = 120, cursor = 0;
    int avail[N_FIELD_SPELLS], navail = 0, k;
    for (k = 0; k < N_FIELD_SPELLS; k++)
        if (party_knows_spell(roster, party, FIELD_SPELLS[k].rec) >= 0) avail[navail++] = k;
    if (navail == 0) { fprintf(stderr, "咒文:目前無人會野外咒文\n"); return 0; }
    raw = dq3_load_file("DQ3.PAL", &rl); if (!raw) return 0;
    pn = dq3_pal_decode(raw, rl, pal, 256); free(raw); dq3_set_palette(pal, pn);
    white = pal_near2(pal,pn,255,255,255); black = pal_near2(pal,pn,0,0,0);
    frame = white; bg = pal_near2(pal,pn,16,16,32);
    while (!dq3_should_quit()) {
        uint8_t sc; int i;
        tav_window(fb, wx, wy, ww, wh, (uint8_t)black, (uint8_t)frame, (uint8_t)bg);
        for (i = 0; i < navail; i++) {
            int yy = wy + 14 + i * 20;
            if (sys) dq3_text_draw_record(sys, fb, DQ3_SCREEN_W, DQ3_SCREEN_H, wx+30, yy, 6, 1,
                                          FIELD_SPELLS[avail[i]].rec, (uint8_t)white);
            if (i == cursor) {     /* ► 游標 */
                int dx, dy, px = wx + 14, py = yy + 2;
                for (dy = 0; dy < 12; dy++) for (dx = 0; dx < 8; dx++) {
                    int xx = px+dx, ry = py+dy;
                    if (xx>=0 && xx<DQ3_SCREEN_W && ry>=0 && ry<DQ3_SCREEN_H) fb[ry*DQ3_SCREEN_W+xx] = (uint8_t)white;
                }
            }
        }
        dq3_present(); dq3_delay_ms(16);
        sc = dq3_poll_scancode();
        if (sc == 0x01 || sc == DQ3_SC_F10) return 0;            /* ESC/F10 取消 */
        else if (sc == 0x50 && cursor + 1 < navail) cursor++;     /* 下 */
        else if (sc == 0x48 && cursor - 1 >= 0) cursor--;          /* 上 */
        else if (sc == 0x1c) {                                     /* Enter 施放 */
            int fk = avail[cursor], mi = party_knows_spell(roster, party, FIELD_SPELLS[fk].rec);
            dq3_member *m = (mi >= 0) ? &roster->list[party->slot[mi]].m : NULL;
            if (!m) return 0;
            if (m->cur_mp < FIELD_SPELLS[fk].mp) {
                fprintf(stderr, "咒文:MP 不足(需 %d,隊員%d 有 %d)\n", FIELD_SPELLS[fk].mp, mi, m->cur_mp);
                continue;                                          /* 留在選單 */
            }
            m->cur_mp = (uint16_t)(m->cur_mp - FIELD_SPELLS[fk].mp);
            fprintf(stderr, "野外施咒:rec%d(隊員%d 扣 MP %d → %d)→ 效果碼 %d\n",
                    FIELD_SPELLS[fk].rec, mi, FIELD_SPELLS[fk].mp, m->cur_mp, FIELD_SPELLS[fk].eff);
            return FIELD_SPELLS[fk].eff;
        }
    }
    return 0;
}

static int cmd_modal(dq3_scene *scene, dq3_roster *roster, dq3_party *party,
                     dq3_inventory *inv, dq3_dialogue *dlg, int dlg_ok, const dq3_text *sys)
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
    case DQ3_CMD_STATUS: status_modal_page(roster, party, sys, 0); break;   /* 道具/咒文名走 D3TXT00 */
    case DQ3_CMD_SPELL:  g_item_world_eff = field_cast_modal(roster, party, sys); break;  /* 野外咒文施放(魯拉/烈米特/特黑洛斯)→ 世界效果碼;清單檢視走「狀況」Space */
    case DQ3_CMD_TALK: case DQ3_CMD_EXAMINE: {
        int fdx = (scene->facing==3)-(scene->facing==1), fdy = (scene->facing==0)-(scene->facing==2);
        int fx = scene->px+fdx, fy = scene->py+fdy, et, ep;
        if (dq3_scene_tile_event(scene, fx, fy, &et, &ep) ||
            dq3_scene_tile_event(scene, scene->px, scene->py, &et, &ep)) {
            if (dlg_ok) { int rec = (ep && ep < dlg->txt.n_records) ? ep : 1; dq3_dialogue_open(dlg, rec); }
        } else fprintf(stderr, "%s:前方無人 / 無物\n", sel==DQ3_CMD_TALK?"對話":"調查");
        break; }
    case DQ3_CMD_ITEM:  g_item_world_eff = item_modal(inv, sys, roster, party); break;  /* 世界效果碼存全域 */
    case DQ3_CMD_EQUIP: equip_modal(roster, party, sys); break; /* 裝備名 = D3TXT00 */
    }
    return sel;
}

/* ── 版權保護閘 ────────────────────────────────────────────────────────────
 * 本重製版受版權保護:須搭配使用者**合法持有的精訊原版 DQ3.EXE** 才能執行。
 * remake 執行期本身**不讀** DQ3.EXE(資料已反組譯編入 dq3_exedata)——此檢查純為確認
 * 「原版在手」:沒有正版執行檔就不給跑。檢查 <assets>/DQ3.EXE(或 DQ3_ORIGINAL_EXE 指定路徑)
 * 的檔案大小 + FNV-1a 全檔指紋是否為精訊原版未改檔。 */
#define DQ3_ORIG_SIZE 115282u
#define DQ3_ORIG_FNV  0x7d64946eu
static int verify_original_exe(const char *assets)
{
    char path[1024]; const char *env = getenv("DQ3_ORIGINAL_EXE");
    FILE *f; long sz; uint32_t h = 0x811c9dc5u; int c;
    if (env && *env) snprintf(path, sizeof path, "%s", env);
    else snprintf(path, sizeof path, "%s/DQ3.EXE", assets);
    f = fopen(path, "rb");
    if (!f) {
        fprintf(stderr,
            "\n✗ 找不到精訊原版 DQ3.EXE(%s)。\n"
            "  本重製版受版權保護,須搭配您【合法持有】的精訊原版執行檔才能啟動。\n"
            "  請把原版 DQ3.EXE 放進素材夾,或以環境變數 DQ3_ORIGINAL_EXE=/path/DQ3.EXE 指定。\n\n",
            path);
        return -1;
    }
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    while ((c = fgetc(f)) != EOF) h = (h ^ (uint32_t)(unsigned char)c) * 0x01000193u;
    fclose(f);
    if ((uint32_t)sz != DQ3_ORIG_SIZE || h != DQ3_ORIG_FNV) {
        fprintf(stderr,
            "\n✗ DQ3.EXE 指紋不符(size=%ld fnv=0x%08x;精訊原版應為 size=%u fnv=0x%08x)。\n"
            "  須為精訊原版未經改檔的執行檔。\n\n",
            sz, (unsigned)h, DQ3_ORIG_SIZE, (unsigned)DQ3_ORIG_FNV);
        return -1;
    }
    fprintf(stderr, "✓ 精訊原版 DQ3.EXE 驗證通過(版權保護:確認正版在手)。\n");
    return 0;
}

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    const char *mode   = (argc > 2) ? argv[2] : "title";
    const char *dump   = getenv("DQ3_DUMP");
    const char *rngmode = getenv("DQ3_RNG");   /* dev/CI override(正式設定走設定檔)*/
    dq3_config cfg; int rc;

    /* 設定套用順序:預設 → 設定檔(跨平台,取代 env)→ env(僅 dev/CI override)。 */
    dq3_config_default(&cfg);
    dq3_config_load(&cfg, dq3_config_path());
    g_cfg = &cfg;                              /* 設定選單讀寫 */
    if (rngmode) cfg.rng_mode = (rngmode[0]=='r'||rngmode[0]=='R') ? DQ3_RNG_REAL : DQ3_RNG_DOS;
    dq3_rng_set_mode(cfg.rng_mode);
    if (cfg.rng_mode == DQ3_RNG_REAL)
        fprintf(stderr, "[RNG] 真實亂數模式(xorshift32,週期 2^32)\n");

    /* 版權保護:確認使用者合法持有的精訊原版 DQ3.EXE 在手,否則不啟動。 */
    if (verify_original_exe(assets) != 0) return 3;

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
    } else if (strcmp(mode, "config") == 0) {       /* 設定畫面驗證(DQ3_DUMP 一幀)*/
        dq3_text t; char cerr[256];
        if (dq3_text_load(&t, assets, "D3TXT00.TXT", cerr, sizeof cerr) == 0) { config_modal(assets, &t); dq3_text_free(&t); }
    } else {
        const char *title = (argc > 3) ? argv[3]
                          : (strcmp(mode, "title") == 0 ? "TITG.P" : mode);
        rc = run_title(assets, title, dump);
    }

    dq3_rt_quit();
    return rc;
}
