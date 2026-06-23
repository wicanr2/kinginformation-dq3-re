/* test_battle.c — 戰鬥結算 #1 修正驗證(巴拉摩斯打不死)。 */
#include "dq3_battle.h"
#include <stdio.h>

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(void)
{
    printf("== Bug #1:巴拉摩斯勝負判定 ==\n");

    /* 正常打死巴拉摩斯:敵 HP=0、我方在場 → 勝 */
    {
        uint16_t party[4]={45,38,40,33}; uint16_t enemy[1]={0};
        CHECK(dq3_battle_resolve(party,0,4,enemy,1)==DQ3_BATTLE_VICTORY, "打死巴拉摩斯(敵HP0)→ 勝");
    }
    /* ★ 關鍵場景:全隊被巴西魯拉吹出戰場(removed 全 1),巴拉摩斯仍活(HP>0) */
    {
        uint16_t party[4]={45,38,40,33}; uint8_t removed[4]={1,1,1,1}; uint16_t enemy[1]={500};
        dq3_battle_outcome fixed = dq3_battle_resolve(party,removed,4,enemy,1);
        dq3_battle_outcome buggy = dq3_battle_resolve_buggy(party,removed,4,enemy,1);
        printf("  全隊被吹飛+巴拉摩斯HP500:  修正=%d(2=敗)  原版bug=%d(1=勝)\n", fixed, buggy);
        CHECK(fixed==DQ3_BATTLE_DEFEAT, "修正:全隊被吹飛 → 判敗(正確)");
        CHECK(buggy==DQ3_BATTLE_VICTORY, "原版:全隊被吹飛 → 誤判勝(重現 #1)");
    }
    /* 全隊 HP 歸 0 同理判敗 */
    {
        uint16_t party[4]={0,0,0,0}; uint16_t enemy[1]={500};
        CHECK(dq3_battle_resolve(party,0,4,enemy,1)==DQ3_BATTLE_DEFEAT, "全隊HP0 → 敗");
    }
    /* 續戰:雙方都有人 */
    {
        uint16_t party[4]={45,0,0,0}; uint16_t enemy[2]={100,0};
        CHECK(dq3_battle_resolve(party,0,4,enemy,2)==DQ3_BATTLE_ONGOING, "雙方都有存活 → 續戰");
    }

    printf("== 傷害套用(uint16 不下溢)==\n");
    CHECK(dq3_battle_apply_damage(30,50)==0, "HP30 受50傷 → 0(不下溢)");
    CHECK(dq3_battle_apply_damage(100,30)==70, "HP100 受30傷 → 70");

    printf("== 物理傷害公式 ==\n");
    {   /* 反組譯真公式(file 0xc03e):dmg=(atk−def)/2 + rng(0..(atk−def)/4);會心×2 */
        int dmin = dq3_battle_phys_damage(40,20,0,0);    /* (40-20)/2=10 + 0 */
        int dmax = dq3_battle_phys_damage(40,20,255,0);  /* 10 + (5*255/256)=14 */
        int dcrit= dq3_battle_phys_damage(40,20,0,1);    /* 10 ×2 = 20 */
        printf("  atk40 def20: roll0=%d roll255=%d crit=%d\n", dmin, dmax, dcrit);
        CHECK(dmin==10, "一般 roll0 = (atk−def)/2 = 10");
        CHECK(dmax==14 && dmax>dmin, "一般 roll255 = 10 + (atk−def)/4 變異 = 14");
        CHECK(dcrit==20, "會心一擊 = 一般 ×2 = 20");
        CHECK(dq3_battle_phys_damage(10,80,0,0)==1 && dq3_battle_phys_damage(10,80,255,0)==0,
              "弱攻(atk≤def):roll0→1 點、roll255→miss(0)");
    }
    printf("== 逃跑判定 ==\n");
    CHECK(dq3_battle_flee_ok(50,0,0)==1, "高敏捷低抗性 roll0 → 逃成功");
    CHECK(dq3_battle_flee_ok(0,100,250)==0, "低敏捷高抗性 roll250 → 逃失敗");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail?1:0;
}
