/* dq3_save.c — 存檔 / 讀檔。 */
#include "dq3_save.h"
#include <stdio.h>
#include <string.h>

#define DQ3_SAVE_MAGIC "DQ3SAVE4"   /* v4:member 加 status 欄位(中毒/麻痺)*/
#define MAGIC_LEN 8

/* 各結構大小寫進檔頭,讀檔時比對 → 防跨版本誤讀。 */
static void sizes(uint32_t s[4])
{
    s[0] = (uint32_t)sizeof(dq3_roster);
    s[1] = (uint32_t)sizeof(dq3_party);
    s[2] = (uint32_t)sizeof(dq3_inventory);
    s[3] = (uint32_t)sizeof(dq3_save_pos);
}

int dq3_save_write(const char *path, const dq3_roster *r, const dq3_party *p,
                   const dq3_inventory *inv, dq3_save_pos pos)
{
    uint32_t s[4]; FILE *f = fopen(path, "wb");
    if (!f) return -1;
    sizes(s);
    fwrite(DQ3_SAVE_MAGIC, 1, MAGIC_LEN, f);
    fwrite(s, sizeof s, 1, f);
    fwrite(r, sizeof *r, 1, f);
    fwrite(p, sizeof *p, 1, f);
    fwrite(inv, sizeof *inv, 1, f);
    fwrite(&pos, sizeof pos, 1, f);
    fclose(f);
    return 0;
}

int dq3_save_read(const char *path, dq3_roster *r, dq3_party *p,
                  dq3_inventory *inv, dq3_save_pos *pos)
{
    char magic[MAGIC_LEN]; uint32_t s[4], want[4];
    FILE *f = fopen(path, "rb");
    if (!f) return -1;
    if (fread(magic, 1, MAGIC_LEN, f) != MAGIC_LEN || memcmp(magic, DQ3_SAVE_MAGIC, MAGIC_LEN) != 0) { fclose(f); return -2; }
    if (fread(s, sizeof s, 1, f) != 1) { fclose(f); return -2; }
    sizes(want);
    if (memcmp(s, want, sizeof want) != 0) { fclose(f); return -3; }   /* 結構大小不符 */
    if (fread(r, sizeof *r, 1, f) != 1 || fread(p, sizeof *p, 1, f) != 1 ||
        fread(inv, sizeof *inv, 1, f) != 1 || fread(pos, sizeof *pos, 1, f) != 1) { fclose(f); return -2; }
    fclose(f);
    return 0;
}

int dq3_save_exists(const char *path)
{
    char magic[MAGIC_LEN]; FILE *f = fopen(path, "rb"); int ok;
    if (!f) return 0;
    ok = (fread(magic, 1, MAGIC_LEN, f) == MAGIC_LEN && memcmp(magic, DQ3_SAVE_MAGIC, MAGIC_LEN) == 0);
    fclose(f);
    return ok;
}
