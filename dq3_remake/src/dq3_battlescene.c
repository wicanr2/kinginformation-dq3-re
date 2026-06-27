/* dq3_battlescene.c — 互動戰鬥場景實作。 */
#include "dq3_battlescene.h"
#include "dq3_runtime.h"
#include "dq3_assets.h"
#include "dq3_rng.h"
#include "dq3_monster.h"
#include "dq3_battle.h"
#include "dq3_combat.h"
#include "dq3_stats.h"
#include "dq3_roster.h"
#include "dq3_spell.h"
#include "dq3_text.h"
#include "dq3_packbg.h"
#include "dq3_audio.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAXE 8
#define PARTY 4
#define GROUNDY ((int)(DQ3_SCREEN_H * 0.50))

/* xorshift32 PRNG(確定性,headless 可重現) */
static dq3_rng g_brng;                            /* 戰鬥亂數:走 dq3_rng(尊重 DOS/REAL 模式)*/
static int roll255(void){ return dq3_rng_next(&g_brng, 256); }

typedef struct {
    const char *dbg;           /* stderr 用 */
    uint16_t name[4]; int name_len; uint16_t cls;   /* 職業名 glyph + 單字職業 glyph */
    int hp, maxhp, mp, maxmp, level, atk, def, agi, defending;
    int status;                /* 異常狀態(DQ3_STATUS_*):中毒/麻痺,進出戰鬥同步 pm */
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
static const uint16_t CMD_SPELL[2]={429,430};   /* 咒文 */
#define BCMD_N 5   /* 戰/咒/逃/防/道具 */

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

/* 龍王城/索瑪神殿(下層最終地城)決戰 = **純黑背景**(原版 DQ3):索瑪 124、其前序列怨靈122/殭屍123、
 * 以及父親歐里狄加128 vs 五頭龍大王129 的決鬥,皆在黑暗中進行,不套天空/綠地 packbg。 */
static int dark_bg_monster(int id)
{ return id==122 || id==123 || id==124 || id==128 || id==129; }

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
    /* 背景三態:索瑪/最終地城 = **純黑**(line 86 已填黑,不畫天空綠地);否則 packbg(天空+綠地);
     * packbg 解碼失敗才退回純色天空/綠地 fallback。 */
    if(dark_bg_monster(monster_id)){
        /* 黑底:不疊任何天空/綠地(原版 DQ3 索瑪/龍王城決戰)。 */
    } else if(g_sky_ok){
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
            if(party[i].maxhp==0 && party[i].name_len==0) continue;   /* 缺席槽:整欄空白 */
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
        int mx=120,my=236,mw=150,mh=108;
        const uint16_t *col1[BCMD_N]={CMD_WAR,CMD_FLEE,CMD_DEF,CMD_ITEM,CMD_SPELL};
        int lead=0; while(lead<PARTY && party[lead].hp<=0) lead++;
        fillrect(fb,mx,my,mw,mh,(uint8_t)black_idx);
        rect_border(fb,mx,my,mw,mh,(uint8_t)white_idx);
        if(lead<PARTY) draw_glyphs(fb,mx+44,my+4, party[lead].name, party[lead].name_len, (uint8_t)white_idx);
        for(i=0;i<BCMD_N;i++){
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
    k=dq3_rng_next(&g_brng,t); for(i=0;i<n;i++)if(hp[i]>0){ if(k==0)return i; k--; } return -1; }
static int pick_alive_party(const member*p){ int i,t=alive_party(p),k; if(!t)return -1;
    k=dq3_rng_next(&g_brng,t); for(i=0;i<PARTY;i++)if(p[i].hp>0){ if(k==0)return i; k--; } return -1; }

/* 隊伍職業索引(對 dq3_stats 成長表):勇者0 / 武鬥家2 / 僧侶3 / 魔法師4(玩家隊伍時覆寫)。 */
static int g_cls_idx[PARTY] = { 0, 2, 3, 4 };

/* 敵方 AI(docs/37):本場怪的 D3MNS AI 欄位 + 我方平均等級 + 本場逃走數。 */
static dq3_monster_ai g_eai; static int g_eai_ok, g_eai_nspell, g_party_avglv, g_fled, g_ehpmax;
static int g_last_gold;                         /* 上一場勝利的戰利金錢 */

/* 戰鬥狀態修正(怪施 base==0 輔助/狀態咒效果;每場 reset;倍率用百分比整數)。
 * 依 docs/data/spell-effects-research.md:拜基魯多→敵攻↑、史卡拉/史克魯多→敵守↑、
 * 魯卡尼/魯加南→我守↓、瑪努莎→我方幻惑(物攻易失手)、瑪荷頓→我方封咒。倍率採 DQ 經典值。 */
static int g_eatk_pct=100, g_edef_pct=100, g_pdef_pct=100, g_party_blind=0, g_party_sealed=0;
static void reset_battle_mods(void){ g_eatk_pct=100; g_edef_pct=100; g_pdef_pct=100; g_party_blind=0; g_party_sealed=0; }
int dq3_battlescene_last_gold(void){ return g_last_gold; }

/* 咒文施放值:DQ3.EXE 公式(file 0xc22e)val = base/2 + rng(base/2)。 */
static int spell_value(int base)
{
    int half = base / 2;
    if (half < 1) half = 1;
    return half + (int)((roll255() * half) / 256);
}

/* 找會施法成員 + 最強可負擔的指定種類咒(kind);回傳成員 index、*pdef=咒文。無回 -1。 */
static int pick_caster_spell_kind(member *party, int kind, const dq3_spell_def **pdef)
{
    int i, best_i = -1; const dq3_spell_def *best = NULL;
    for (i = 0; i < PARTY; i++) {
        unsigned short recs[64]; int n, k;
        if (party[i].hp <= 0 || party[i].mp <= 0) continue;
        n = dq3_spells_known(g_cls_idx[i], party[i].level, recs, 64);
        for (k = 0; k < n; k++) {
            const dq3_spell_def *d = dq3_spell_def_get(recs[k]);
            if (!d || d->kind != kind || d->mp > party[i].mp) continue;
            if (!best || d->base > best->base) { best = d; best_i = i; }
        }
        if (best) break;   /* 取第一個能施法的成員 */
    }
    *pdef = best; return best_i;
}
static int pick_caster_spell(member *party, const dq3_spell_def **pdef)
{ return pick_caster_spell_kind(party, DQ3_SK_DMG, pdef); }

/* 手動選咒(互動):成員 ci 施放指定咒 d;NULL = 自動。 */
static int g_manual_cast_ci = -1;
static const dq3_spell_def *g_manual_cast_def = NULL;

/* 隊伍中 HP 比例最低(且未陣亡)且 < maxhp/4 的成員;無回 -1。 */
static int lowest_hurt_ally(const member *party)
{
    int i, t = -1; long worst = 1000;
    for (i = 0; i < PARTY; i++) {
        if (party[i].hp <= 0 || party[i].maxhp <= 0) continue;
        if (party[i].hp * 4 >= party[i].maxhp) continue;          /* 還夠血 */
        { long r = (long)party[i].hp * 100 / party[i].maxhp;
          if (r < worst) { worst = r; t = i; } }
    }
    return t;
}

/* 套用成員 ci 施放咒文 d 的效果(扣 MP + 傷害/回復;公式對齊 EXE 0xc22e)。 */
static void cast_spell_effect(member *party, int *ehp, int en, int ci, const dq3_spell_def *d)
{
    if (ci < 0 || !d) return;
    party[ci].mp -= d->mp; if (party[ci].mp < 0) party[ci].mp = 0;
    if (d->kind == DQ3_SK_HEAL) {
        int t = lowest_hurt_ally(party); if (t < 0) t = ci;
        int v = (d->base >= 999) ? party[t].maxhp : spell_value(d->base);
        party[t].hp += v; if (party[t].hp > party[t].maxhp) party[t].hp = party[t].maxhp;
        fprintf(stderr, "  %s 詠唱回復咒文 → %s 回復 %d HP。\n", party[ci].dbg, party[t].dbg, v);
    } else if (d->target == DQ3_TG_ENEMY1) {
        int e = pick_alive_enemy(ehp, en);
        if (e >= 0) { int v = spell_value(d->base); ehp[e] -= v; if (ehp[e] < 0) ehp[e] = 0;
            fprintf(stderr, "  %s 詠唱咒文 → 敵%d 受 %d 傷害%s\n", party[ci].dbg, e, v, ehp[e]==0?" — 擊倒!":""); }
    } else {
        int e, hit = 0;
        for (e = 0; e < en; e++) if (ehp[e] > 0) { int v = spell_value(d->base); ehp[e] -= v; if (ehp[e]<0)ehp[e]=0; hit++; }
        fprintf(stderr, "  %s 詠唱範圍咒文 → 波及 %d 隻敵人。\n", party[ci].dbg, hit);
    }
}

/* 執行一回合;cmd:0戰 1逃 2防 3道具 4咒文。回傳 0續 1勝 2敗 3逃成功。 */
static int do_turn(member*party, int*ehp, int en, int eatk, int edef, int eagi, int efree, int cmd)
{
    int i;
    eatk = eatk * g_eatk_pct / 100;          /* 敵攻倍率(拜基魯多)*/
    edef = edef * g_edef_pct / 100;          /* 敵守倍率(史卡拉/史克魯多)*/
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
    } else if(cmd==4){ /* 咒文 */
        if(g_party_sealed){                             /* 瑪荷頓 封咒:我方無法施咒 */
            fprintf(stderr,"  我方咒文被封(瑪荷頓),無法施放。\n");
            g_manual_cast_def = NULL; g_manual_cast_ci = -1;
        } else if(g_manual_cast_def){                   /* 手動選咒(互動)*/
            cast_spell_effect(party, ehp, en, g_manual_cast_ci, g_manual_cast_def);
            g_manual_cast_def = NULL; g_manual_cast_ci = -1;
        } else {                                        /* 自動:快倒先治療,否則最強攻擊咒 */
            const dq3_spell_def *hd; int hurt = lowest_hurt_ally(party);
            int hi = (hurt >= 0) ? pick_caster_spell_kind(party, DQ3_SK_HEAL, &hd) : -1;
            if (hi >= 0 && hd) cast_spell_effect(party, ehp, en, hi, hd);
            else {
                const dq3_spell_def *d; int ci = pick_caster_spell(party, &d);
                if(ci<0 || !d) fprintf(stderr,"  無人能施放咒文(MP 不足或非施法職)。\n");
                else cast_spell_effect(party, ehp, en, ci, d);
            }
        }
    } else { /* 戰:每位存活成員攻擊;武器雙擊則打 2 次(#7a,dq3_combat)*/
        for(i=0;i<PARTY;i++){
            int hits, h;
            if(party[i].hp<=0) continue;
            if(party[i].status & DQ3_STATUS_PARALYSIS){       /* 麻痺:本回合不能行動 */
                fprintf(stderr,"  %s 麻痺,無法行動。\n", party[i].dbg); continue; }
            hits = g_items_ok ? dq3_combat_num_attacks(&g_items, party[i].weapon) : 1;
            for(h=0; h<hits; h++){
                int e,crit,dmg;
                e=pick_alive_enemy(ehp,en); if(e<0) break;
                if(g_party_blind && roll255()<128){            /* 瑪努莎 幻惑:~50% 失手 */
                    fprintf(stderr,"  %s 揮空了(幻惑)。\n", party[i].dbg); continue; }
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

    /* 敵方回合:逐隻依真實 D3MNS AI(docs/37)— 逃跑 / 施咒 / 物攻。 */
    for(i=0;i<en;i++){
        int t,dmg;
        if(ehp[i]<=0) continue;
        /* 逃跑判定:我方平均等級 ≥ flee_thresh 且 rng ≤ flee_rate → 怪逃走(不給經驗)*/
        if(g_eai_ok && g_eai.flee_rate>0 && g_party_avglv >= g_eai.flee_thresh
           && roll255() <= g_eai.flee_rate){
            ehp[i]=0; g_fled++;
            fprintf(stderr,"  敵%d 逃走了!\n", i);
            continue;
        }
        t=pick_alive_party(party); if(t<0) break;
        /* 施咒 vs 物攻:rng(256) < cast_prob 且 會咒文 → 從 bitmask 隨機挑真實咒文施放 */
        if(g_eai_ok && g_eai_nspell>0 && roll255() < g_eai.cast_prob){
            int bits[48], nb=0, b, bit; const dq3_spell_def *d;
            for(b=0;b<48;b++) if(g_eai.spell_mask[b/8] & (0x80>>(b%8))) bits[nb++]=b;
            bit = bits[dq3_rng_next(&g_brng,nb)];        /* 均勻隨機(docs/37)*/
            { int srec = dq3_monster_spell_rec[bit];      /* bit→真咒名 rec(0x3930 remap)*/
              /* 異常咒(RE docs/re-log:base==0 + 名辨識):144 ラリホー睡 / 152 メダパニ混亂
               * → 玩家不能行動 → 映射麻痺。對活著的目標施加並帶回 pm(docs/47 #5)。 */
              if(srec==144 || srec==152){
                  party[t].status |= DQ3_STATUS_PARALYSIS;
                  fprintf(stderr,"  敵%d 詠唱%s → %s 陷入%s(麻痺)\n", i, srec==144?"拉里荷":"美達巴尼",
                          party[t].dbg, srec==144?"睡眠":"混亂");
                  continue;                               /* 本敵行動結束 */
              }
              /* 其餘 base==0 輔助/狀態咒(rec 143-160,docs/data/spell-effects-research.md):
               * 套對應戰鬥修正(取代 no-op);未模型化者詠唱不致傷。睡/混亂(144/152)已上面映射麻痺。 */
              if(srec>=143 && srec<=160){
                  const char *nm=0;
                  if(srec==151){ g_eatk_pct = g_eatk_pct*2>400?400:g_eatk_pct*2; nm="拜基魯多(敵攻擊上升)"; }
                  else if(srec==154||srec==155){ g_edef_pct+=50; if(g_edef_pct>300)g_edef_pct=300;
                          nm=(srec==154)?"史卡拉(敵守備上升)":"史克魯多(敵守備上升)"; }
                  else if(srec==146||srec==147){ g_pdef_pct-=25; if(g_pdef_pct<25)g_pdef_pct=25;
                          nm=(srec==146)?"魯卡尼(我方守備下降)":"魯加南(我方守備下降)"; }
                  else if(srec==158){ g_party_blind=1; nm="瑪努莎(我方陷入幻惑·物攻易失手)"; }
                  else if(srec==156){ g_party_sealed=1; nm="瑪荷頓(我方咒文被封)"; }
                  if(nm) fprintf(stderr,"  敵%d 詠唱%s\n", i, nm);
                  else   fprintf(stderr,"  敵%d 詠唱輔助/狀態咒(rec%d;效果未模型化,不致傷)\n", i, srec);
                  continue;
              }
            }
            d = dq3_spell_def_get(dq3_monster_spell_rec[bit]);  /* bit→真咒名(0x3930 remap)*/
            if(d && d->kind==DQ3_SK_HEAL){                /* 治療系怪:補一隻受傷的同伴 */
                int e, lo=-1, lv=1<<30;
                for(e=0;e<en;e++) if(ehp[e]>0 && ehp[e]<g_ehpmax && ehp[e]<lv){lv=ehp[e];lo=e;}
                if(lo<0) lo=i;
                { int v=(d->base>=255)?g_ehpmax:(d->base/2+(int)((roll255()*(d->base/2))/256));
                  ehp[lo]+=v; if(ehp[lo]>g_ehpmax)ehp[lo]=g_ehpmax;
                  fprintf(stderr,"  敵%d 詠唱回復咒文 → 敵%d 回復 %d。\n", i, lo, v); }
            } else {                                      /* 攻擊咒:打我方(魔甲抗魔 #7b)*/
                int base = d ? d->base : 24;
                dmg = g_items_ok ? dq3_combat_spell_damage(&g_items, party[t].equip,
                          base/2 + (int)((roll255()*(base/2))/256)) : base/2;
                if(party[t].defending) dmg/=2;
                party[t].hp = (int)dq3_battle_apply_damage((uint16_t)party[t].hp, dmg);
                fprintf(stderr,"  敵%d 詠唱咒文 → %s%s 受 %d 傷害%s\n", i, party[t].dbg,
                        (g_items_ok && dq3_combat_has_magic_resist(&g_items, party[t].equip))?"(魔甲抗魔·減半!)":"",
                        dmg, party[t].hp==0?" — 倒下!":"");
            }
        } else {
            dmg=dq3_battle_phys_damage(eatk, party[t].def * g_pdef_pct / 100, roll255(), roll255()<8);
            if(party[t].defending) dmg/=2;
            party[t].hp = (int)dq3_battle_apply_damage((uint16_t)party[t].hp, dmg);
            fprintf(stderr,"  敵%d 攻擊 %s,造成 %d 傷害%s\n", i, party[t].dbg, dmg,
                    party[t].hp==0?" — 倒下!":"");
        }
    }
    /* 中毒:回合末各中毒成員扣 HP(戰鬥內不致死,留 1;量為 remake 值)*/
    for(i=0;i<PARTY;i++) if(party[i].hp>0 && (party[i].status & DQ3_STATUS_POISON)){
        party[i].hp -= 3; if(party[i].hp<1) party[i].hp=1;
        fprintf(stderr,"  %s 中毒,受 3 毒傷。\n", party[i].dbg);
    }
    /* #1 正確結算:我方全滅(含倒下)→ 敗 */
    if(alive_party(party)==0) return 2;
    return 0;
}

/* 玩家隊伍覆寫(露依達酒場建立的)。NULL = 用內建範例隊。 */
static dq3_roster      *g_pl_roster = NULL;   /* 非 const:勝利後回寫升級 */
static const dq3_party *g_pl_party  = NULL;
static int g_pl_ri[PARTY];                    /* 各戰鬥槽 → 名冊 index(-1=缺席/範例)*/
void dq3_battlescene_set_party(dq3_roster *r, const dq3_party *p)
{ g_pl_roster = r; g_pl_party = p; }

/* 光之珠(0x65):索瑪戰持有時,第一回合自動使用 → 驅散黑暗結界(二階段)。
 * 設於 main.c 開索瑪戰前(inv 有 0x65 → 1)。非索瑪戰無效。 */
static int g_light_orb = 0;
void dq3_battlescene_set_light_orb(int has) { g_light_orb = has ? 1 : 0; }

/* 勝利結算:存活隊員獲經驗 → 升級(#5 夾43)→ 套成長(#4 MP/#6 uint16 不wrap)→ 更新戰鬥屬性。 */
static void apply_victory_exp(member *party, dq3_member *pm, uint32_t total_exp)
{
    int i;
    if (!g_stats_ok) return;
    fprintf(stderr, "=== 勝利!全隊獲得 %u 經驗 ===\n", total_exp);
    for (i = 0; i < PARTY; i++) {
        int gained;
        if (party[i].hp <= 0) continue;                       /* 陣亡者不獲經驗 */
        gained = dq3_member_gain_exp_rng(&pm[i], &g_stats, total_exp, &g_brng);  /* 忠實 rng 成長 */
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

/* 把升級後的隊員實體(經驗/等級/數值)回寫名冊,讓成長持久(閉合 創角→戰鬥→成長)。 */
static void writeback_roster(const dq3_member *pm)
{
    int i;
    if (!g_pl_roster) return;
    for (i = 0; i < PARTY; i++)
        if (g_pl_ri[i] >= 0 && g_pl_ri[i] < g_pl_roster->count)
            g_pl_roster->list[g_pl_ri[i]].m = pm[i];
}

/* 戰後把存活 HP/MP 同步進 pm.cur_hp/cur_mp(隨後 writeback_roster 持久化;陣亡=0)。 */
static void sync_cur_to_pm(const member *party, dq3_member *pm)
{
    int i;
    for (i = 0; i < PARTY; i++) {
        int hp = party[i].hp < 0 ? 0 : party[i].hp, mp = party[i].mp < 0 ? 0 : party[i].mp;
        if (hp > pm[i].stat[DQ3_STAT_HP]) hp = pm[i].stat[DQ3_STAT_HP];
        if (mp > pm[i].stat[DQ3_STAT_MP]) mp = pm[i].stat[DQ3_STAT_MP];
        pm[i].cur_hp = (uint16_t)hp; pm[i].cur_mp = (uint16_t)mp;
        pm[i].status = (uint8_t)party[i].status;        /* 戰鬥施加的異常帶回 */
    }
}

/* 互動選咒子選單:列領頭施法成員已學會的可施放咒文(名+MP),選定 → 設 g_manual_cast_*,回 1;
 * ESC 取消回 0。 */
static int spell_menu_select(member *party, const dq3_monster_sprite *spr, int *ehp, int en, int monster_id)
{
    int ci = -1, i, nd = 0, cur = 0;
    const dq3_spell_def *defs[64];
    for (i = 0; i < PARTY; i++) {
        unsigned short recs[64]; int n, k;
        if (party[i].hp <= 0 || party[i].mp <= 0) continue;
        n = dq3_spells_known(g_cls_idx[i], party[i].level, recs, 64);
        for (k = 0; k < n; k++) { const dq3_spell_def *d = dq3_spell_def_get(recs[k]);
            if (d && nd < 64) defs[nd++] = d; }
        if (nd > 0) { ci = i; break; }
    }
    if (ci < 0 || nd == 0) { fprintf(stderr, "  無人能施放咒文。\n"); return 0; }

    { const char *sd = getenv("DQ3_SPELLSEL_DUMP");   /* 視覺驗證:截選咒選單首幀 */
      if (sd) { int wx=16,wy=120,ww=110,wh=14+nd*16,r; if(wh>200)wh=200;
        render(dq3_fb(),spr,ehp,en,party,4,1,monster_id);
        fillrect(dq3_fb(),wx,wy,ww,wh,(uint8_t)black_idx); rect_border(dq3_fb(),wx,wy,ww,wh,(uint8_t)white_idx);
        for(r=0;r<nd;r++){ int yy=wy+6+r*16; if(yy+14>wy+wh)break;
            if(r==0)dq3_text_draw_glyph(&g_txt,dq3_fb(),DQ3_SCREEN_W,DQ3_SCREEN_H,wx+4,yy,11,(uint8_t)white_idx);
            dq3_text_draw_record(&g_txt,dq3_fb(),DQ3_SCREEN_W,DQ3_SCREEN_H,wx+20,yy,5,1,defs[r]->rec,
                (defs[r]->mp<=party[ci].mp)?(uint8_t)white_idx:(uint8_t)red_idx); }
        dq3_present(); dq3_dump_ppm(sd); fprintf(stderr,"spellsel(%d 咒) -> %s\n",nd,sd); return 0; } }

    while (!dq3_should_quit()) {
        uint8_t sc;
        int wx = 16, wy = 120, ww = 110, wh = 14 + nd * 16, r;
        if (wh > 200) wh = 200;
        render(dq3_fb(), spr, ehp, en, party, 4, 1, monster_id);
        fillrect(dq3_fb(), wx, wy, ww, wh, (uint8_t)black_idx);
        rect_border(dq3_fb(), wx, wy, ww, wh, (uint8_t)white_idx);
        for (r = 0; r < nd; r++) {
            int yy = wy + 6 + r * 16;
            if (yy + 14 > wy + wh) break;
            if (r == cur) dq3_text_draw_glyph(&g_txt, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H, wx + 4, yy, 11, (uint8_t)white_idx);
            dq3_text_draw_record(&g_txt, dq3_fb(), DQ3_SCREEN_W, DQ3_SCREEN_H, wx + 20, yy, 5, 1, defs[r]->rec,
                                 (defs[r]->mp <= party[ci].mp) ? (uint8_t)white_idx : (uint8_t)red_idx);
        }
        dq3_present();
        sc = dq3_poll_scancode();
        if (sc == 0x01) return 0;                                /* ESC 取消 */
        else if (sc == 0x48) cur = (cur + nd - 1) % nd;
        else if (sc == 0x50) cur = (cur + 1) % nd;
        else if (sc == 0x1c || sc == 0x39) {
            if (defs[cur]->mp <= party[ci].mp) { g_manual_cast_ci = ci; g_manual_cast_def = defs[cur]; return 1; }
            fprintf(stderr, "  MP 不足。\n");
        }
        dq3_delay_ms(16);
    }
    return 0;
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

    /* 戰鬥配曲:大型 boss(id≥106:大魔人/巴拉摩斯/索瑪/父親/九頭龍…)用 boss 曲,一般遇敵用戰鬥曲。
     * 返回主場景後由主迴圈每幀自動切回地表/城鎮曲。 */
    dq3_audio_play_scene(monster_id >= 106 ? DQ3_MUS_BOSS : DQ3_MUS_BATTLE, 1);

    /* ITEM.DAT 先載(裝備加成 #7a/#7b + 攻防加成需要;在玩家隊覆寫前)*/
    g_items_ok = (dq3_items_load(&g_items, assets, err, sizeof err) == 0);

    /* 玩家隊伍覆寫:若酒場已建隊,改用真實名冊成員(姓名/職業/等級/數值)。 */
    if (g_pl_roster && g_pl_party) {
        int si, pi = 0, j, k;
        for (si = 0; si < DQ3_PARTY_MAX && pi < PARTY; si++) {
            int ri = g_pl_party->slot[si];
            const dq3_recruit *rc;
            if (ri < 0 || ri >= g_pl_roster->count) continue;
            rc = &g_pl_roster->list[ri];
            party[pi].dbg = "隊員";
            party[pi].name_len = rc->name_len < 4 ? rc->name_len : 4;
            for (j = 0; j < party[pi].name_len; j++) party[pi].name[j] = rc->name[j];
            party[pi].cls   = dq3_class_names[rc->m.cls].g[0];      /* HUD 單字職業 glyph */
            party[pi].level = rc->m.level;
            party[pi].maxhp = (int)rc->m.stat[DQ3_STAT_HP];
            party[pi].maxmp = (int)rc->m.stat[DQ3_STAT_MP];
            /* 起始用持久 cur_hp/cur_mp(戰鬥傷害保留;clamp 防 stale)*/
            party[pi].hp = (int)rc->m.cur_hp > party[pi].maxhp ? party[pi].maxhp : (int)rc->m.cur_hp;
            party[pi].mp = (int)rc->m.cur_mp > party[pi].maxmp ? party[pi].maxmp : (int)rc->m.cur_mp;
            /* 攻擊力 = 力量 + 武器 b0;守備力 = 耐力/2 + (鎧+盾+兜 b1 總和)(ITEM.DAT,docs/22;4 槽)*/
            party[pi].atk = (int)rc->m.stat[DQ3_STAT_STR]
                            + (g_items_ok ? dq3_item_attack(&g_items, rc->weapon) : 0);
            party[pi].def = (int)rc->m.stat[DQ3_STAT_VIT] / 2;
            if (g_items_ok) party[pi].def += dq3_item_defense(&g_items, rc->armor)
                            + dq3_item_defense(&g_items, rc->shield) + dq3_item_defense(&g_items, rc->head);
            party[pi].agi   = (int)rc->m.stat[DQ3_STAT_AGI];
            party[pi].weapon = (rc->weapon==0xff)?0:rc->weapon;     /* #7a 雙擊判定用 */
            party[pi].defending = 0;
            for (k = 0; k < 8; k++) party[pi].equip[k] = 0xff;
            { int es=0;                                              /* #7b 抗魔掃描:4 槽都進 equip[] */
              if (rc->armor !=0xff) party[pi].equip[es++]=rc->armor;
              if (rc->shield!=0xff) party[pi].equip[es++]=rc->shield;
              if (rc->head  !=0xff) party[pi].equip[es++]=rc->head;
              if (rc->weapon!=0xff) party[pi].equip[es++]=rc->weapon; }
            g_cls_idx[pi] = rc->m.cls;
            g_pl_ri[pi] = ri;                  /* 記住回寫目標 */
            pi++;
        }
        for (k = pi; k < PARTY; k++) g_pl_ri[k] = -1;
        for (; pi < PARTY; pi++) {                                  /* 不足 4 名:缺席(整欄空白)*/
            party[pi].dbg = "(空)"; party[pi].name_len = 0; party[pi].cls = 13;  /* 13=空白 glyph */
            party[pi].hp = party[pi].maxhp = party[pi].mp = party[pi].maxmp = 0;
            party[pi].level = 0; party[pi].atk = party[pi].def = party[pi].agi = 0;
            party[pi].weapon = 0; g_cls_idx[pi] = 0;
        }
    }

    dq3_rng_seed(&g_brng, (uint16_t)(seed ? seed : 0x1234));
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
    /* (ITEM.DAT 已於玩家隊覆寫前載入)*/
    /* 升級系統(成長/門檻表;apply_bug4_fix=1 → 勇者 MP 修正)+ 隊員實體 */
    g_stats_ok = (dq3_stats_load(&g_stats, assets, 1, err, sizeof err) == 0);
    { int i; for (i = 0; i < PARTY; i++) {
        if (!g_stats_ok) continue;
        if (g_pl_roster && g_pl_ri[i] >= 0)
            pm[i] = g_pl_roster->list[g_pl_ri[i]].m;     /* 真實隊員:保留真 exp,升級才正確 */
        else
            dq3_member_init(&pm[i], &g_stats, g_cls_idx[i], party[i].level);
        party[i].status = pm[i].status;                 /* 進戰鬥帶入異常(overworld 中毒延續)*/
    } }

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
    /* edef:採真實怪物防禦。上限放寬到 <1000 讓合法高防(索瑪 300)通過;只把 0 / garbage
     * (空記錄 0xffff 等)退回 4。先前 <255 會誤把最終 boss 索瑪(def 300)clamp 成 4。 */
    eatk = ms.atk>0?ms.atk:8; edef=(ms.def>0&&ms.def<1000)?ms.def:4; eagi=ms.agi; efree=ms.flee_resist;
    { int i; for(i=0;i<en;i++) ehp[i]=ehpmax; }
    g_last_gold = 0;
    /* 敵方 AI(docs/37):讀本場怪 D3MNS AI 欄 + 算我方平均等級;g_fled 計逃走數 */
    g_eai_ok = (dq3_monster_get_ai(&mons, monster_id, &g_eai) == 0);
    g_eai_nspell = g_eai_ok ? dq3_monster_spell_count(&g_eai) : 0;
    g_ehpmax = ehpmax;
    g_fled = 0;
    reset_battle_mods();                 /* 每場戰鬥重置 buff/debuff/幻惑/封咒 */
    { int i, s=0, c=0; for(i=0;i<PARTY;i++) if(party[i].maxhp>0||party[i].level>0){ s+=party[i].level; c++; }
      g_party_avglv = c ? s/c : 1; }
    fprintf(stderr,"=== 遭遇 %s ×%d (HP%d atk%d def%d) ===\n", ms.exp?"怪物":"怪物", en, ehpmax, eatk, edef);

    /* 索瑪(0x7c)二階段:持有光之珠 0x65 → 戰鬥開始自動使用,驅散黑暗結界。
     * 原版語意:光之珠解除索瑪的黑暗保護 → 攻防大降、露出本體弱化形態。
     * 無光之珠 → 索瑪全力(atk 180)幾乎必殺隊伍(= 原版「需先取光之珠」的教學)。 */
    if (monster_id == 0x7c && g_light_orb) {
        int oatk = eatk, odef = edef;
        eatk = eatk / 3;                       /* 攻擊力大降:180→60(隊伍可承受)*/
        edef = edef / 4;                        /* 防禦大降:300→75(露出本體可受傷)*/
        fprintf(stderr, "★ 使用光之珠!驅散索瑪的黑暗結界 —— 索瑪由黑暗形態轉為弱化本體(二階段):"
                "攻 %d→%d 防 %d→%d\n", oatk, eatk, odef, edef);
        g_light_orb = 0;                       /* 用畢清旗標 */
    }

    if(getenv("DQ3_SPELLSEL_DUMP")){   /* 選咒選單視覺驗證 */
        spell_menu_select(party, &spr, ehp, en, monster_id);
        dq3_monsters_free(&mons); if(g_txt_ok)dq3_text_free(&g_txt); return 0;
    }
    /* ---- headless 腳本(script 驅動;dump 選擇性)---- */
    if(script){
        const char *p = script;
        while(p && *p && outcome==0){
            int cmd = (*p=='R'||*p=='r')?1:(*p=='D'||*p=='d')?2:(*p=='I'||*p=='i')?3:(*p=='M'||*p=='m')?4:0;
            turn++;
            fprintf(stderr,"-- 回合 %d cmd=%c --\n", turn, *p);
            outcome=do_turn(party,ehp,en,eatk,edef,eagi,efree,cmd);
            p++;
            if(turn>50) break;
        }
        if(outcome==1){ g_last_gold = ((en-g_fled)>0?(en-g_fled):0) * (int)ms.gold; apply_victory_exp(party, pm, (uint32_t)((en-g_fled)>0?(en-g_fled):0) * ms.exp);   /* 勝利→經驗+金錢→升級 */
            sync_cur_to_pm(party, pm); writeback_roster(pm); }                 /* 回寫名冊(含持久 HP)*/
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
        if(sc==0x48){ cursor=(cursor+BCMD_N-1)%BCMD_N; }   /* 上 */
        else if(sc==0x50){ cursor=(cursor+1)%BCMD_N; }     /* 下 */
        else if(sc==0x1c||sc==0x39){               /* Enter/Space:執行 */
            if(cursor==4){                          /* 咒文:先開手動選咒選單 */
                if(!spell_menu_select(party, &spr, ehp, en, monster_id)) continue;  /* 取消→回指令 */
            }
            turn++;
            outcome=do_turn(party,ehp,en,eatk,edef,eagi,efree,cursor);
        }
        dq3_delay_ms(16);
    }
    if(outcome==1){ g_last_gold = ((en-g_fled)>0?(en-g_fled):0) * (int)ms.gold; apply_victory_exp(party, pm, (uint32_t)((en-g_fled)>0?(en-g_fled):0) * ms.exp);   /* 勝利→經驗+金錢→升級 */
        sync_cur_to_pm(party, pm); writeback_roster(pm); }                 /* 回寫名冊(含持久 HP)*/
    /* 結束:顯示末幀片刻 */
    render(dq3_fb(),&spr,ehp,en,party,cursor,0,monster_id);
    dq3_present();
    fprintf(stderr,"outcome=%d\n",outcome);
    dq3_monsters_free(&mons); if(g_txt_ok)dq3_text_free(&g_txt);
    return outcome;
}
