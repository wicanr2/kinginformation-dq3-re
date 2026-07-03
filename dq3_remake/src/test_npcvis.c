/* test_npcvis.c — NPC story-flag 可見性過濾驗證(docs/71)。
 *
 * 對真實 CTY00.DAT section0(阿里阿罕)載入 NPC 兩次:
 *   (1) 不設 dq3_scene_set_npc_vis(NULL,預設)→ 應得未過濾總數 24(docs/71 ground truth)。
 *   (2) dq3_npc_vis_init + dq3_scene_set_npc_vis → 應得起始顯示數 15。
 * 用法:dq3_npcvis_test [assets_dir](預設 "."）。
 */
#include "dq3_town.h"
#include "dq3_scene.h"
#include "dq3_sprite.h"
#include "dq3_runtime.h"
#include <stdio.h>

/* ---- stub:本測試不觸發 sprite/palette 呼叫,僅供連結 dq3_scene.c/dq3_town.c ---- */
void dq3_set_palette(const dq3_color *pal, int count) { (void)pal; (void)count; }
int  dq3_charsprite_load(dq3_charsprite *cs, const char *a, const char *b, int e, char *err, int cap)
{ (void)cs; (void)a; (void)b; (void)e; (void)err; (void)cap; return -1; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(int argc, char **argv)
{
    const char *assets = (argc > 1) ? argv[1] : ".";
    char err[256] = {0};
    dq3_scene *s;
    uint8_t npc_vis[32];

    printf("== CTY00 section0(阿里阿罕)NPC 過濾前總數 ==\n");
    dq3_scene_set_npc_vis(NULL);   /* 明示不過濾 */
    s = dq3_town_load(assets, "CTY00.DAT", 0, 1, err, sizeof err);
    if (!s) { fprintf(stderr, "load CTY00 sec0 failed: %s\n", err); return 2; }
    printf("  n_npcs(未過濾) = %d\n", s->n_npcs);
    CHECK(s->n_npcs == 24, "未過濾總數 = 24(docs/68/71 ground truth)");
    dq3_scene_free(s);

    printf("== 套用 story-flag 可見性過濾(docs/71 新遊戲初值)==\n");
    dq3_npc_vis_init(npc_vis);
    dq3_scene_set_npc_vis(npc_vis);
    s = dq3_town_load(assets, "CTY00.DAT", 0, 1, err, sizeof err);
    if (!s) { fprintf(stderr, "load CTY00 sec0 (filtered) failed: %s\n", err); return 2; }
    printf("  n_npcs(起始顯示) = %d\n", s->n_npcs);
    CHECK(s->n_npcs == 15, "起始顯示數 = 15(docs/71 ground truth,對齊 ebitan)");
    dq3_scene_free(s);
    dq3_scene_set_npc_vis(NULL);   /* 還原,避免影響同進程後續(此測試無後續,保守起見）*/

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
