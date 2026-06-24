/* dq3_roster.c — 露依達酒場創角 + 名冊 + 隊伍編成。 */
#include "dq3_roster.h"
#include <string.h>

/* 職業名 glyph 索引(D3TXT00.FON;反查 glyph_unicode_map)。 */
const struct dq3_class_name dq3_class_names[DQ3_NUM_CLASS] = {
    {2, {106, 187, 0, 0}},      /* 勇者 */
    {2, {107, 144, 0, 0}},      /* 戰士 */
    {3, {108, 207, 657, 0}},    /* 武鬥家 */
    {2, {109, 704, 0, 0}},      /* 僧侶 */
    {4, {110, 208, 210, 187}},  /* 魔法使者 */
    {2, {111, 187, 0, 0}},      /* 賢者 */
    {2, {112, 194, 0, 0}},      /* 商人 */
    {3, {113, 705, 187, 0}},    /* 遊玩者 */
};

void dq3_roster_class_menu(dq3_menu *m, int x, int y)
{
    int c;
    dq3_menu_init(m, x, y);
    for (c = 0; c < DQ3_NUM_CLASS; c++)
        dq3_menu_add(m, dq3_class_names[c].g, dq3_class_names[c].len);
}

int dq3_class_sprite_entry(int cls, int gender)
{
    if (cls < 0 || cls >= DQ3_NUM_CLASS) cls = 0;
    if (gender != DQ3_GENDER_FEMALE) gender = DQ3_GENDER_MALE;
    return (cls * 2 + gender) * 4;   /* DQ3MST.BLS char = cls*2+gender,entry=char*4 */
}

void dq3_roster_init(dq3_roster *r) { memset(r, 0, sizeof *r); }
void dq3_party_init(dq3_party *p)
{
    int i;
    p->count = 0;
    for (i = 0; i < DQ3_PARTY_MAX; i++) p->slot[i] = -1;
}

int dq3_roster_create(dq3_roster *r, const dq3_stats *st,
                      int cls, int gender, const uint16_t *name, int name_len)
{
    dq3_recruit *rc;
    int i, idx;
    if (r->count >= DQ3_ROSTER_MAX) return -1;
    if (cls < 0 || cls >= DQ3_NUM_CLASS) return -1;
    if (gender != DQ3_GENDER_MALE && gender != DQ3_GENDER_FEMALE) return -1;
    if (name_len < 0) name_len = 0;
    if (name_len > DQ3_NAME_MAX) name_len = DQ3_NAME_MAX;

    idx = r->count++;
    rc = &r->list[idx];
    memset(rc, 0, sizeof *rc);
    rc->weapon = rc->armor = 0xff;     /* 無裝備(0xff;memset 後 0 = 木棒,須清)*/
    rc->gender = gender;
    rc->name_len = name_len;
    for (i = 0; i < name_len; i++) rc->name[i] = name ? name[i] : 0;
    rc->in_party = 0;
    dq3_member_init(&rc->m, st, cls, 1);   /* Lv1 初始(成長表 base;rng 擲值待 RE 精校)*/
    return idx;
}

int dq3_roster_remove(dq3_roster *r, dq3_party *p, int idx)
{
    int i, s;
    if (idx < 0 || idx >= r->count) return -1;
    if (p) {                                          /* 在隊伍中則先移出 */
        for (s = 0; s < DQ3_PARTY_MAX; s++) if (p->slot[s] == idx) dq3_party_remove(p, r, s);
    }
    for (i = idx; i < r->count - 1; i++) r->list[i] = r->list[i + 1];  /* 往前補位 */
    r->count--;
    if (p) {                                          /* 補位後修正隊伍中 > idx 的參照 */
        for (s = 0; s < DQ3_PARTY_MAX; s++) if (p->slot[s] > idx) p->slot[s]--;
    }
    return 0;
}

int dq3_party_add(dq3_party *p, dq3_roster *r, int roster_idx)
{
    int i;
    if (roster_idx < 0 || roster_idx >= r->count) return -1;
    if (r->list[roster_idx].in_party) return -1;       /* 已在隊伍 */
    if (p->count >= DQ3_PARTY_MAX) return -1;          /* 隊伍滿 */
    for (i = 0; i < DQ3_PARTY_MAX; i++) {
        if (p->slot[i] < 0) {
            p->slot[i] = roster_idx;
            r->list[roster_idx].in_party = 1;
            p->count++;
            return i;
        }
    }
    return -1;
}

int dq3_party_remove(dq3_party *p, dq3_roster *r, int slot)
{
    int ri;
    if (slot < 0 || slot >= DQ3_PARTY_MAX || p->slot[slot] < 0) return -1;
    ri = p->slot[slot];
    if (ri >= 0 && ri < r->count) r->list[ri].in_party = 0;
    p->slot[slot] = -1;
    p->count--;
    return 0;
}
