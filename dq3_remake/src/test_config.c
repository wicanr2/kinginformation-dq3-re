/* test_config.c — 可攜設定檔 round-trip + 套用順序。 */
#include "dq3_config.h"
#include "dq3_rng.h"
#include <stdio.h>
#include <string.h>
static int g_fail=0;
#define CHECK(c,m) do{ if(c)printf("  [PASS] %s\n",m); else{printf("  [FAIL] %s\n",m);g_fail++;} }while(0)
int main(void){
    dq3_config c; const char *p="/tmp/dq3_test.cfg";
    printf("== 預設 ==\n");
    dq3_config_default(&c);
    CHECK(c.rng_mode==DQ3_RNG_DOS, "預設 rng=DOS 忠實");
    CHECK(c.music_enabled==1 && c.music_volume==70, "預設 music 開/音量70");
    printf("== 存→讀 round-trip ==\n");
    c.rng_mode=DQ3_RNG_REAL; c.music_enabled=0; c.music_volume=45; c.audio_backend=1;
    CHECK(dq3_config_save(&c,p)==0, "存檔成功");
    { dq3_config c2; dq3_config_default(&c2);
      CHECK(dq3_config_load(&c2,p)==4, "讀回 4 鍵(rng+music+music_vol+audio)");
      CHECK(c2.rng_mode==DQ3_RNG_REAL, "rng=real 持久");
      CHECK(c2.music_enabled==0 && c2.music_volume==45, "music 開關/音量持久");
      CHECK(c2.audio_backend==1, "音源後端持久(mt32)"); }
    printf("== 缺檔不變動 ==\n");
    { dq3_config c3; dq3_config_default(&c3);
      CHECK(dq3_config_load(&c3,"/tmp/dq3_nope.cfg")==0 && c3.rng_mode==DQ3_RNG_DOS, "缺檔→0、保預設"); }
    printf("\n%s (fail=%d)\n", g_fail?"FAIL":"ALL PASS", g_fail);
    return g_fail?1:0;
}
