/* test_tavern.c — 酒場創角完整流程狀態機驗證(職業→姓名→性別→登錄→名冊)。 */
#include "dq3_tavern.h"
#include <stdio.h>

void dq3_text_draw_glyph(const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                         int x, int y, int idx, uint8_t fg)
{ (void)t;(void)fb;(void)fb_w;(void)fb_h;(void)x;(void)y;(void)idx;(void)fg; }

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

int main(void)
{
    dq3_stats st; dq3_roster r; dq3_party p; dq3_tavern tv;

    if (dq3_stats_load(&st, NULL, 1, NULL, 0) < 0) { printf("stats FAIL\n"); return 1; }
    dq3_roster_init(&r); dq3_party_init(&p);
    dq3_tavern_init(&tv, &r, &p, &st);

    printf("== 完整流程:職業→姓名→性別→登錄 ==\n");
    CHECK(tv.state == DQ3_TAV_CLASS, "起始 = 選職業");

    /* 選職業:游標移到 1(戰士)後 Enter */
    dq3_tavern_input(&tv, 0x50);           /* 下 → 職業1 戰士 */
    dq3_tavern_input(&tv, 0x1c);           /* Enter */
    CHECK(tv.state == DQ3_TAV_NAME && tv.pend_class == 1, "選戰士 → 進姓名輸入");

    /* 姓名:選 A(cell10)、B(cell11),再到 OK 完成 */
    tv.ni.cursor = 10; dq3_tavern_input(&tv, 0x1c);   /* A */
    tv.ni.cursor = 11; dq3_tavern_input(&tv, 0x1c);   /* B */
    CHECK(tv.ni.len == 2, "輸入 2 字");
    tv.ni.cursor = DQ3_NI_CELL_OK; dq3_tavern_input(&tv, 0x1c);  /* OK */
    CHECK(tv.state == DQ3_TAV_GENDER, "姓名 OK → 進性別選擇");

    /* 性別:Enter(男性=0)*/
    dq3_tavern_input(&tv, 0x1c);
    CHECK(tv.state == DQ3_TAV_ROSTER, "選性別 → 進名冊");
    CHECK(r.count == 1, "名冊登錄 1 人");
    {
        dq3_recruit *rc = &r.list[0];
        CHECK(rc->m.cls == 1 && rc->m.level == 1, "登錄者 = 戰士 Lv1");
        CHECK(rc->name_len == 2 && rc->name[0] == 15 && rc->name[1] == 16, "名字 = AB(glyph 15,16)");
        CHECK(rc->gender == DQ3_GENDER_MALE, "性別 = 男");
        CHECK(rc->in_party && p.count == 1, "自動入隊");
    }

    printf("== 名冊:Enter 切換入隊 / Space 新增 ==\n");
    dq3_tavern_input(&tv, 0x1c);           /* Enter → 離隊 */
    CHECK(!r.list[0].in_party && p.count == 0, "Enter → 離隊");
    dq3_tavern_input(&tv, 0x1c);           /* Enter → 再入隊 */
    CHECK(r.list[0].in_party && p.count == 1, "Enter → 再入隊");
    dq3_tavern_input(&tv, 0x39);           /* Space → 新增 */
    CHECK(tv.state == DQ3_TAV_CLASS, "Space → 回選職業(新增另一名)");

    /* 第二名:武鬥家(職業2)、名字 C、女 */
    dq3_tavern_input(&tv, 0x50); dq3_tavern_input(&tv, 0x50); dq3_tavern_input(&tv, 0x1c); /* 職業2 */
    tv.ni.cursor = 12; dq3_tavern_input(&tv, 0x1c);                      /* C */
    tv.ni.cursor = DQ3_NI_CELL_OK; dq3_tavern_input(&tv, 0x1c);          /* OK */
    dq3_tavern_input(&tv, 0x50); dq3_tavern_input(&tv, 0x1c);            /* 性別 女 */
    CHECK(r.count == 2 && r.list[1].m.cls == 2 && r.list[1].gender == DQ3_GENDER_FEMALE,
          "第二名 = 武鬥家 女");

    printf("== 第三名:注音輸入(Tab 切注音 → ㄕˋ → 是)==\n");
    dq3_tavern_input(&tv, 0x1c);                                /* 名冊 Enter(無妨)*/
    /* 用 Space 從名冊回職業選 */
    { int s; for (s = 0; s < 5 && tv.state != DQ3_TAV_CLASS; s++) dq3_tavern_input(&tv, 0x39); }
    dq3_tavern_input(&tv, 0x1c);                                /* 選職業0 勇者 → NAME */
    CHECK(tv.state == DQ3_TAV_NAME, "進姓名輸入");
    dq3_tavern_input(&tv, 0x0f);                                /* Tab → 注音模式 */
    CHECK(tv.name_mode == 1, "Tab → 注音模式");
    tv.zh.cursor = 16; dq3_tavern_input(&tv, 0x1c);            /* ㄕ */
    tv.zh.cursor = 39; dq3_tavern_input(&tv, 0x1c);            /* ˋ → 候選窗 */
    { int i, idx = -1; for (i = 0; i < tv.zh.ncand; i++) if (tv.zh.cand[i] == 399) idx = i;
      tv.zh.cand_cur = idx; }
    dq3_tavern_input(&tv, 0x1c);                                /* 挑「是」(399)→ 入名字 */
    CHECK(tv.ni.len == 1 && tv.ni.buf[0] == 399, "注音挑「是」→ 名字緩衝 = {399}");

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
