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

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail?1:0;
}
