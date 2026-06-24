/* dq3_battlescene.c — 互動戰鬥場景實作。 */
#include "dq3_battlescene.h"
#include "dq3_runtime.h"
#include "dq3_assets.h"
#include "dq3_monster.h"
#include "dq3_battle.h"
#include "dq3_combat.h"
#include "dq3_stats.h"
#include "dq3_text.h"
#include "dq3_packbg.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAXE 8
#define PARTY 4
#define GROUNDY ((int)(DQ3_SCREEN_H * 0.50))

/* xorshift32 PRNG(確定性,headless 可重現) */
static unsigned g_rng;
static unsigned rnd(void){ g_rng^=g_rng<<13; g_rng^=g_rng>>17; g_rng^=g_rng<<5; return g_rng; }
static int roll255(void){ return (int)(rnd() & 0xff); }

typedef struct {
    const char *dbg;           /* stderr 用 */
    uint16_t name[4]; int name_len; uint16_t cls;   /* 職業名 glyph + 單字職業 glyph */
    int hp, maxhp, mp, maxmp, level, atk, def, agi, defending;
    int weapon;                /* 武器 item code(#7a 雙擊判定)*/
    uint8_t equip[8];          /* 8 格裝備(#7b 抗魔掃描)*/
} member;

static int sky_idx, ground_idx, white_idx, red_idx, green_idx, black_idx;
static dq3_text g_txt; static int g_txt_ok;
static dq3_items g_items; static int g_items_ok;   /* ITEM.DAT 旗標(#7a/#7b)*/
static dq3_stats g_stats; static int g_stats_ok;   /* 成長/門檻表(#4/#5/#6 升級系統)*/
static uint8_t g_sky[DQ3_PACKBG_H][DQ3_PACKBG_W]; static int g_sky_ok;  /* packbg 天空 */

/* 指令標籤(glyph index 對照 docs/03 + glyph_unicode_map):戰鬥/逃跑/防禦/道具 */
static const uint16_t CMD_WAR[2]={107,207}, CMD_FLEE[2]={629,630}, CMD_DEF[2]={203,204}, CMD_ITEM[2]={402,1354};

static void draw_glyphs(uint8_t*fb,int x,int y,const uint16_t*g,int n,uint8_t fg){
    int i; if(!g_txt_ok)return;
    for(i=0;i<n;i++) dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,x+i*16,y,g[i],fg);
}
static void draw_number(uint8_t*fb,int x,int y,int val,uint8_t fg){
    uint16_t d[6]; int k=0,i; if(!g_txt_ok)return; if(val<0)val=0;
    if(val==0){ d[k++]=0; } else { int v=val; uint16_t tmp[6]; int tn=0;
        while(v>0&&tn<6){ tmp[tn++]=(uint16_t)(v%10); v/=10; } for(i=tn-1;i>=0;i--) d[k++]=tmp[i]; }
    draw_glyphs(fb,x,y,d,k,fg);
}

static int pal_near(const dq3_color *p,int n,int r,int g,int b){
    int i,best=0; long bd=-1;
    for(i=0;i<n;i++){ long dr=p[i].r-r,dg=p[i].g-g,db=p[i].b-b,d=dr*dr+dg*dg+db*db;
        if(bd<0||d<bd){bd=d;best=i;} }
    return best;
}
static void fillrect(uint8_t*fb,int x,int y,int w,int h,uint8_t c){
    int r,cc; for(r=0;r<h;r++){ int yy=y+r; if(yy<0||yy>=DQ3_SCREEN_H)continue;
        for(cc=0;cc<w;cc++){ int xx=x+cc; if(xx>=0&&xx<DQ3_SCREEN_W) fb[yy*DQ3_SCREEN_W+xx]=c; } }
}
static void rect_border(uint8_t*fb,int x,int y,int w,int h,uint8_t c){
    fillrect(fb,x,y,w,1,c); fillrect(fb,x,y+h-1,w,1,c);
    fillrect(fb,x,y,1,h,c); fillrect(fb,x+w-1,y,1,h,c);
}

static void render(uint8_t*fb, const dq3_monster_sprite*spr, const int*ehp,int en,
                   const member*party, int cursor, int show_menu, int monster_id)
{
    /* 版面對齊 references/game3.png:上方狀態列、中間綠地怪群、下方左指令/右敵名。
     * 天空(雲)0..GBAND_TOP、綠地帶 GBAND_TOP..GBAND_BOT、黑底 GBAND_BOT..底。 */
    /* 比例對齊 references/game3.png(640×350):狀態框 y8、packbg 場景 y80..246
     * (含內建天空+綠地)、綠地平線 ~y230、黑底 y246+(指令/敵名框)。 */
    const int FIELD_Y0 = 80, FIELD_Y1 = 246, GROUND_Y = 232;
    int i;
    memset(fb, (uint8_t)black_idx, (size_t)DQ3_SCREEN_W*DQ3_SCREEN_H);
    /* packbg 場景(含天空+綠地):88 row 映射到 y80..246 */
    if(g_sky_ok){
        int y,x;
        for(y=FIELD_Y0;y<FIELD_Y1;y++){
            int sr = (y-FIELD_Y0)*DQ3_PACKBG_H/(FIELD_Y1-FIELD_Y0); if(sr>=DQ3_PACKBG_H)sr=DQ3_PACKBG_H-1;
            for(x=0;x<DQ3_SCREEN_W && x<DQ3_PACKBG_W;x++)
                fb[y*DQ3_SCREEN_W+x]=g_sky[sr][x];
        }
    } else {
        for(i=FIELD_Y0;i<FIELD_Y1;i++)
            memset(fb+(size_t)i*DQ3_SCREEN_W,(uint8_t)(i<GROUND_Y?sky_idx:ground_idx),DQ3_SCREEN_W);
    }
    /* 怪群:站在綠地平線上(死的不畫)*/
    for(i=0;i<en;i++){
        int gx=(DQ3_SCREEN_W*(i+1))/(en+1)-spr->w/2, gy=GROUND_Y+8-spr->h, r,c;
        if(ehp[i]<=0) continue;
        for(r=0;r<spr->h;r++)for(c=0;c<spr->w;c++){
            int yy=gy+r,xx=gx+c;
            if(spr->opaque[r][c]&&yy>=0&&yy<DQ3_SCREEN_H&&xx>=0&&xx<DQ3_SCREEN_W)
                fb[yy*DQ3_SCREEN_W+xx]=spr->px[r][c];
        }
    }
    /* ── 上方隊伍狀態列(RE 座標系,docs/13):文字原點 X=0x13 byte=152px、Y=8px;
     *    4 欄,欄距 0xa byte=80px;字/行 16px。每欄:名 / H+HP / M+MP / 職業+等級。 */
    {
        int tx0=0x13*8, ty0=8, colpx=0xa*8;       /* =152px 起、80px/欄 */
        int sx=tx0-8, sy=ty0-6, sw=PARTY*colpx+8, sh=4*16+12;
        fillrect(fb,sx,sy,sw,sh,(uint8_t)black_idx);
        rect_border(fb,sx,sy,sw,sh,(uint8_t)white_idx);
        for(i=0;i<PARTY;i++){
            int cx=tx0+i*colpx, fg = party[i].hp>0?white_idx:red_idx;
            draw_glyphs(fb,cx,ty0,    party[i].name, party[i].name_len, (uint8_t)white_idx);
            dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,cx,ty0+16,22,(uint8_t)fg);      /* H,行高16 */
            draw_number(fb,cx+18,ty0+16, party[i].hp, (uint8_t)fg);
            dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,cx,ty0+32,27,(uint8_t)white_idx); /* M */
            draw_number(fb,cx+18,ty0+32, party[i].mp, (uint8_t)white_idx);
            dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,cx,ty0+48,party[i].cls,(uint8_t)white_idx);
            draw_number(fb,cx+18,ty0+48, party[i].level, (uint8_t)white_idx);
        }
    }
    /* ── 下方左:指令選單(角色名 + 2 欄 戰鬥/逃跑/防禦/道具 + ► 游標)── */
    if(show_menu){
        int mx=120,my=252,mw=150,mh=92;
        const uint16_t *col1[4]={CMD_WAR,CMD_FLEE,CMD_DEF,CMD_ITEM};
        int lead=0; while(lead<PARTY && party[lead].hp<=0) lead++;
        fillrect(fb,mx,my,mw,mh,(uint8_t)black_idx);
        rect_border(fb,mx,my,mw,mh,(uint8_t)white_idx);
        if(lead<PARTY) draw_glyphs(fb,mx+44,my+4, party[lead].name, party[lead].name_len, (uint8_t)white_idx);
        for(i=0;i<4;i++){
            int y=my+24+i*16;
            if(i==cursor) dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,mx+8,y,0x77,(uint8_t)white_idx); /* ★/▶ 游標 */
            dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,mx+28,y,col1[i][0],(uint8_t)white_idx);
            dq3_text_draw_glyph(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,mx+72,y,col1[i][1],(uint8_t)white_idx);
        }
    }
    /* ── 下方右:敵名 + 數量 ── */
    if(g_txt_ok){
        int ex=290,ey=262,ew=250,eh=36, alive=0;
        for(i=0;i<en;i++) if(ehp[i]>0) alive++;
        fillrect(fb,ex,ey,ew,eh,(uint8_t)black_idx);
        rect_border(fb,ex,ey,ew,eh,(uint8_t)white_idx);
        dq3_text_draw_record(&g_txt,fb,DQ3_SCREEN_W,DQ3_SCREEN_H,ex+12,ey+8,12,1,0x258+monster_id,(uint8_t)white_idx);
        draw_number(fb,ex+ew-40,ey+8, alive, (uint8_t)white_idx);
    }
}

static int alive_party(const member*p){ int i,c=0; for(i=0;i<PARTY;i++)if(p[i].hp>0)c++; return c; }
static int alive_enemy(const int*hp,int n){ int i,c=0; for(i=0;i<n;i++)if(hp[i]>0)c++; return c; }
static int pick_alive_enemy(const int*hp,int n){ int i,t=alive_enemy(hp,n),k; if(!t)return -1;
    k=rnd()%t; for(i=0;i<n;i++)if(hp[i]>0){ if(k==0)return i; k--; } return -1; }
static int pick_alive_party(const member*p){ int i,t=alive_party(p),k; if(!t)return -1;
    k=rnd()%t; for(i=0;i<PARTY;i++)if(p[i].hp>0){ if(k==0)return i; k--; } return -1; }

/* 執行一回合;cmd:0戰 1逃 2防 3道具。回傳 0續 1勝 2敗 3逃成功。 */
static int do_turn(member*party, int*ehp, int en, int eatk, int edef, int eagi, int efree, int cmd)
{
    int i;
    for(i=0;i<PARTY;i++) party[i].defending=0;

    if(cmd==1){ /* 逃 */
        int lead=-1; for(i=0;i<PARTY;i++)if(party[i].hp>0){lead=i;break;}
        if(lead>=0 && dq3_battle_flee_ok(party[lead].agi, efree, roll255())) return 3;
        fprintf(stderr,"  逃跑失敗!\n");
    } else if(cmd==2){ /* 防 */
        for(i=0;i<PARTY;i++) if(party[i].hp>0) party[i].defending=1;
        fprintf(stderr,"  全員防禦。\n");
    } else if(cmd==3){ /* 道具:簡化為治療帶頭存活者 */
        int t=pick_alive_party(party);
        if(t>=0){ int heal=30; party[t].hp+=heal; if(party[t].hp>party[t].maxhp)party[t].hp=party[t].maxhp;
            fprintf(stderr,"  %s 使用藥草,回復 %d。\n", party[t].dbg, heal); }
    } else { /* 戰:每位存活成員攻擊;武器雙擊則打 2 次(#7a,dq3_combat)*/
        for(i=0;i<PARTY;i++){
            int hits, h;
            if(party[i].hp<=0) continue;
            hits = g_items_ok ? dq3_combat_num_attacks(&g_items, party[i].weapon) : 1;
            for(h=0; h<hits; h++){
                int e,crit,dmg;
                e=pick_alive_enemy(ehp,en); if(e<0) break;
                crit=(roll255()<8);
                dmg=dq3_battle_phys_damage(party[i].atk, edef, roll255(), crit);
                ehp[e]-= dmg; if(ehp[e]<0)ehp[e]=0;
                fprintf(stderr,"  %s 攻擊敵%d%s%s,造成 %d 傷害%s\n", party[i].dbg, e,
                        (hits>1)?(h==0?"(第1擊)":"(第2擊·飛鷹劍!)"):"",
                        crit?"(會心一擊!)":"", dmg, ehp[e]==0?" — 擊倒!":"");
            }
        }
    }
    if(alive_enemy(ehp,en)==0) return 1;   /* 勝 */

    /* 敵方反擊:每隻存活敵人攻擊;偶爾改放咒文(魔甲抗魔 #7b,dq3_combat)*/
    for(i=0;i<en;i++){
        int t,dmg;
        if(ehp[i]<=0) continue;
        t=pick_alive_party(party); if(t<0) break;
        if((roll255() & 3) == 0){           /* 1/4 機率放咒文(示範 #7b)*/
            int base = 24 + (int)(roll255() & 7);   /* 咒文基礎傷害 */
            dmg = g_items_ok ? dq3_combat_spell_damage(&g_items, party[t].equip, base) : base;
            if(party[t].defending) dmg/=2;
            party[t].hp = (int)dq3_battle_apply_damage((uint16_t)party[t].hp, dmg);
            fprintf(stderr,"  敵%d 詠唱咒文 → %s%s 受 %d 傷害%s\n", i, party[t].dbg,
                    (g_items_ok && dq3_combat_has_magic_resist(&g_items, party[t].equip))?"(魔法鎧甲抗魔·減半!)":"",
                    dmg, party[t].hp==0?" — 倒下!":"");
        } else {
            dmg=dq3_battle_phys_damage(eatk, party[t].def, roll255(), roll255()<8);
            if(party[t].defending) dmg/=2;
            party[t].hp = (int)dq3_battle_apply_damage((uint16_t)party[t].hp, dmg);
            fprintf(stderr,"  敵%d 攻擊 %s,造成 %d 傷害%s\n", i, party[t].dbg, dmg,
                    party[t].hp==0?" — 倒下!":"");
        }
    }
    /* #1 正確結算:我方全滅(含倒下)→ 敗 */
    if(alive_party(party)==0) return 2;
    return 0;
}

/* 隊伍職業索引(對 dq3_stats 成長表):勇者0 / 武鬥家2 / 僧侶3 / 魔法師4 */
static const int g_cls_idx[PARTY] = { 0, 2, 3, 4 };

/* 勝利結算:存活隊員獲經驗 → 升級(#5 夾43)→ 套成長(#4 MP/#6 uint16 不wrap)→ 更新戰鬥屬性。 */
static void apply_victory_exp(member *party, dq3_member *pm, uint32_t total_exp)
{
    int i;
    if (!g_stats_ok) return;
    fprintf(stderr, "=== 勝利!全隊獲得 %u 經驗 ===\n", total_exp);
    for (i = 0; i < PARTY; i++) {
        int gained;
        if (party[i].hp <= 0) continue;                       /* 陣亡者不獲經驗 */
        gained = dq3_member_gain_exp(&pm[i], &g_stats, total_exp);
        if (gained > 0) {
            party[i].level = pm[i].level;                     /* #5 */
            party[i].maxhp = (int)pm[i].stat[DQ3_STAT_HP];    /* #6 */
            party[i].maxmp = (int)pm[i].stat[DQ3_STAT_MP];    /* #4 */
            if (party[i].hp > party[i].maxhp) party[i].hp = party[i].maxhp;
            fprintf(stderr, "  ★ %s 升到 Lv%d!MaxHP=%d MaxMP=%d\n",
                    party[i].dbg, party[i].level, party[i].maxhp, party[i].maxmp);
        }
    }
}

int dq3_battlescene_run(const char *assets, int monster_id, int monster_count,
                        int bg_page, const char *script, const char *dump, unsigned seed)
{
    dq3_color pal[256]; int pn;
    uint8_t *raw; size_t rlen;
    dq3_monster_sprite spr;
    dq3_monster_stat ms;
    dq3_monsters mons;
    int ehp[MAXE]; int en = monster_count;
    int eatk, edef, eagi, efree, ehpmax;
    char err[256]={0};
    int cursor=0, outcome=0, turn=0;
    dq3_member pm[PARTY];          /* 升級用隊員實體(#4/#5/#6)*/
    member party[PARTY] = {
        /* dbg,name,len,cls, hp,maxhp,mp,maxmp,lv,atk,def,agi,def, weapon, equip[8] */
        { "勇者",   {106,187},    2, 106, 35,35,  9, 9,1, 22,10,12, 0, 0x6e, {0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff} }, /* 飛鷹劍=雙擊(#7a) */
        { "武鬥家", {108,207,657},3, 108, 40,40,  0, 0,1, 26,14, 8, 0, 0x60, {0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff} },
        { "僧侶",   {109,704},    2, 109, 28,28, 10,10,1, 15,10,10, 0, 0x60, {0xff,0xff,0x2b,0xff,0xff,0xff,0xff,0xff} }, /* 魔法鎧甲=抗魔(#7b) */
        { "魔法師", {110,208,822},3, 110, 22,22, 11,11,1, 11, 6,11, 0, 0x60, {0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff} },
    };

    g_rng = seed ? seed : 0x1234567u;
    if(en<1)en=1; if(en>MAXE)en=MAXE;

    raw=dq3_load_file("MNSBK.PAL",&rlen);
    if(!raw){ fprintf(stderr,"load MNSBK.PAL failed\n"); return -1; }
    pn=dq3_pal_decode(raw,rlen,pal,256); free(raw);
    dq3_set_palette(pal,pn);
    sky_idx=pal_near(pal,pn,60,140,240); ground_idx=pal_near(pal,pn,40,150,40);
    white_idx=pal_near(pal,pn,255,255,255); red_idx=pal_near(pal,pn,230,40,40);
    green_idx=pal_near(pal,pn,40,220,40); black_idx=pal_near(pal,pn,0,0,0);

    /* 文字(敵名/指令/HP):D3TXT00.FON + D3TXT00.TXT */
    g_txt_ok = (dq3_text_load(&g_txt, assets, "D3TXT00.TXT", err, sizeof err) == 0);
    /* ITEM.DAT 旗標(#7a 雙擊 / #7b 抗魔) */
    g_items_ok = (dq3_items_load(&g_items, assets, err, sizeof err) == 0);
    /* 升級系統(成長/門檻表;apply_bug4_fix=1 → 勇者 MP 修正)+ 隊員實體 */
    g_stats_ok = (dq3_stats_load(&g_stats, assets, 1, err, sizeof err) == 0);
    { int i; for (i = 0; i < PARTY; i++)
        if (g_stats_ok) dq3_member_init(&pm[i], &g_stats, g_cls_idx[i], party[i].level); }

    /* 戰鬥背景:packbg page(隨地形;預設草原 terrain0 → page 0)。
     * DQ3_BG_PAGE 可指定其他地形頁(terrain*8,terrain3→0x19)。 */
    {
        int bgpage = bg_page;
        if (bgpage < 0) bgpage = getenv("DQ3_BG_PAGE") ? atoi(getenv("DQ3_BG_PAGE")) : 22;  /* 22=草原(對 game3.png) */
        g_sky_ok = (dq3_packbg_decode(assets, bgpage, g_sky, err, sizeof err) == 0);
    }

    if(dq3_monster_sprite_get(assets,monster_id,&spr,err,sizeof err)!=0){
        fprintf(stderr,"monster %d sprite: %s\n",monster_id,err); return -1; }
    if(dq3_monsters_load(&mons,assets,err,sizeof err)!=0){
        fprintf(stderr,"D3MNS: %s\n",err); return -1; }
    dq3_monster_get_stat(&mons,monster_id,&ms);
    ehpmax = ms.hp_base + ms.hp_rand; if(ehpmax<1)ehpmax=1;
    eatk = ms.atk>0?ms.atk:8; edef=(ms.def>0&&ms.def<255)?ms.def:4; eagi=ms.agi; efree=ms.flee_resist;
    { int i; for(i=0;i<en;i++) ehp[i]=ehpmax; }
    fprintf(stderr,"=== 遭遇 %s ×%d (HP%d atk%d def%d) ===\n", ms.exp?"怪物":"怪物", en, ehpmax, eatk, edef);

    /* ---- headless 腳本(script 驅動;dump 選擇性)---- */
    if(script){
        const char *p = script;
        while(p && *p && outcome==0){
            int cmd = (*p=='R'||*p=='r')?1:(*p=='D'||*p=='d')?2:(*p=='I'||*p=='i')?3:0;
            turn++;
            fprintf(stderr,"-- 回合 %d cmd=%c --\n", turn, *p);
            outcome=do_turn(party,ehp,en,eatk,edef,eagi,efree,cmd);
            p++;
            if(turn>50) break;
        }
        if(outcome==1) apply_victory_exp(party, pm, (uint32_t)en * ms.exp);   /* 勝利→經驗→升級 */
        if(dump){
            render(dq3_fb(),&spr,ehp,en,party,cursor,outcome==0,monster_id);
            dq3_present();
            if(dq3_dump_ppm(dump)==0)
                fprintf(stderr,"battle scene -> %s (outcome=%d: %s)\n", dump, outcome,
                        outcome==1?"勝":outcome==2?"敗":outcome==3?"逃":"續");
        } else {
            fprintf(stderr,"battle outcome=%d (%s)\n", outcome,
                    outcome==1?"勝":outcome==2?"敗":outcome==3?"逃":"續");
        }
        dq3_monsters_free(&mons); if(g_txt_ok)dq3_text_free(&g_txt);
        return outcome;
    }

    /* ---- 互動 ---- */
    while(!dq3_should_quit() && outcome==0){
        uint8_t sc;
        render(dq3_fb(),&spr,ehp,en,party,cursor,1,monster_id);
        dq3_present();
        sc=dq3_poll_scancode();
        if(sc==0x48){ cursor=(cursor+3)&3; }       /* 上 */
        else if(sc==0x50){ cursor=(cursor+1)&3; }  /* 下 */
        else if(sc==0x1c||sc==0x39){               /* Enter/Space:執行 */
            turn++;
            outcome=do_turn(party,ehp,en,eatk,edef,eagi,efree,cursor);
        }
        dq3_delay_ms(16);
    }
    if(outcome==1) apply_victory_exp(party, pm, (uint32_t)en * ms.exp);   /* 勝利→經驗→升級 */
    /* 結束:顯示末幀片刻 */
    render(dq3_fb(),&spr,ehp,en,party,cursor,0,monster_id);
    dq3_present();
    fprintf(stderr,"outcome=%d\n",outcome);
    dq3_monsters_free(&mons); if(g_txt_ok)dq3_text_free(&g_txt);
    return outcome;
}
