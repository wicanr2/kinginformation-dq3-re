/* test_monster.c — 怪物數值 + SHP sprite 解碼驗證(對齊 docs/16 / d3mns_stats.json)。
 * 並坐實 #3 根因:id128 歐里狄加 / id129 九頭龍 = 空 sprite(缺圖 → 原版當機)。
 */
#include "dq3_monster.h"
#include <stdio.h>

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(int argc, char **argv)
{
    const char *assets = (argc>1)?argv[1]:".";
    char err[256]={0};
    dq3_monsters m;
    dq3_monster_stat s;
    dq3_monster_sprite spr;

    if (dq3_monsters_load(&m, assets, err, sizeof err)!=0){ fprintf(stderr,"load: %s\n",err); return 2; }

    printf("== 怪物數值(對 docs/16)==\n");
    dq3_monster_get_stat(&m,5,&s);
    printf("  id5 史萊姆: HP=%u exp=%u gold=%u\n", s.hp_base, s.exp, s.gold);
    CHECK(s.hp_base==6 && s.exp==4 && s.gold==2, "史萊姆 HP6/exp4/gold2");
    dq3_monster_get_stat(&m,2,&s);
    printf("  id2 金屬史萊姆: HP=%u exp=%u\n", s.hp_base, s.exp);
    CHECK(s.hp_base==3 && s.exp==4140, "金屬史萊姆 HP3/exp4140");
    dq3_monster_get_stat(&m,3,&s);
    CHECK(s.exp==32766, "金屬怪 exp32766");
    CHECK(dq3_monster_hp_max(&m,5)==9, "史萊姆 HP_max=base+rand=9");

    printf("== SHP sprite 解碼 ==\n");
    CHECK(dq3_monster_sprite_decode(assets,5,&spr,err,sizeof err)==0, "史萊姆 sprite 解碼成功");
    if (spr.w){
        int any=0,r,c; for(r=0;r<spr.h;r++)for(c=0;c<spr.w;c++)any|=spr.opaque[r][c];
        printf("  id5 sprite %dx%d 非透明像素=%d\n", spr.w, spr.h, any);
        CHECK(spr.w>0 && spr.h>0 && any, "史萊姆 sprite 尺寸合理且非空");
    }

    printf("== #3 根因:id128/129 缺 sprite ==\n");
    {
        int r128 = dq3_monster_sprite_decode(assets,128,&spr,err,sizeof err);
        int r129 = dq3_monster_sprite_decode(assets,129,&spr,err,sizeof err);
        printf("  id128 歐里狄加 decode rc=%d, id129 九頭龍 rc=%d\n", r128, r129);
        CHECK(r128<0 && r129<0, "id128/129 原版為空 sprite(缺圖,原版 blit 無 guard → 當機)");
    }

    printf("== #3 修正:復原 sprite 回退 ==\n");
    {
        int g128 = dq3_monster_sprite_get(assets,128,&spr,err,sizeof err);
        CHECK(g128==0 && spr.w>0 && spr.h>0, "id128 get 回退復原 sprite(歐里狄加)");
        int g129 = dq3_monster_sprite_get(assets,129,&spr,err,sizeof err);
        CHECK(g129==0 && spr.w>0 && spr.h>0, "id129 get 回退復原 sprite(五頭龍大王)");
        int g5 = dq3_monster_sprite_get(assets,5,&spr,err,sizeof err);
        CHECK(g5==0, "一般怪(史萊姆 id5)get 仍走原版 SHP");
    }

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    dq3_monsters_free(&m);
    return g_fail?1:0;
}
