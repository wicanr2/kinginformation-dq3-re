/* test_dialogue.c — 驗證「話す NPC」資料路徑(無 SDL):
 *   1. CTY section header +0x17 = 對話 bank(docs/42)。
 *   2. NPC 記錄 byte4 = 對話 id;互動子型 (b3>>3)&7 < 3 = 對話型。
 *   3. dq3_dialogue 載 D3TXT0<bank> → open(byte4) 取得非空 glyph 序列。
 *   4. dq3_dialogue_set_bank 切換 bank(reload 文字檔)。
 * assets dir 由 argv[1] 指定(預設 "assets_raw")。 */
#include "dq3_dialogue.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int fails = 0;
#define CHECK(c, msg) do { if (c) printf("  [PASS] %s\n", msg); \
    else { printf("  [FAIL] %s\n", msg); fails++; } } while (0)

static unsigned char *slurp(const char *dir, const char *name, long *len) {
    char p[1024]; FILE *f; long n; unsigned char *b;
    snprintf(p, sizeof p, "%s/%s", dir, name);
    f = fopen(p, "rb"); if (!f) return NULL;
    fseek(f, 0, SEEK_END); n = ftell(f); fseek(f, 0, SEEK_SET);
    b = malloc(n); if (fread(b, 1, n, f) != (size_t)n) { fclose(f); free(b); return NULL; }
    fclose(f); *len = n; return b;
}
static int u16(const unsigned char *b, long o) { return b[o] | (b[o+1] << 8); }

int main(int argc, char **argv) {
    const char *dir = argc > 1 ? argv[1] : "assets_raw";
    long clen; unsigned char *cty = slurp(dir, "CTY00.DAT", &clen);
    int so, bank, np, base, cnt, i, talk_b4 = -1;
    dq3_dialogue dlg; char err[128];

    printf("== test_dialogue (assets=%s) ==\n", dir);
    if (!cty) { printf("  [SKIP] 找不到 %s/CTY00.DAT\n", dir); return 0; }

    so = u16(cty, 0);                          /* section0 base */
    bank = cty[so + 0x17];                      /* +0x17 = 對話 bank */
    CHECK(bank == 1, "CTY00 section0 對話 bank == 1 (D3TXT01)");

    np = u16(cty, so + 0);                       /* NPC 清單 +0 */
    if (np == 0xffff) np = u16(cty, so + 2);
    base = so + np; cnt = cty[base];
    for (i = 0; i < cnt; i++) {
        const unsigned char *r = cty + base + 1 + i * 7;
        if (((r[3] >> 3) & 7) < 3) { talk_b4 = r[4]; break; }   /* 第一個對話型 NPC */
    }
    CHECK(talk_b4 >= 0, "CTY00 找到對話型 NPC");

    if (dq3_dialogue_load(&dlg, dir, "D3TXT01.TXT", err, sizeof err) != 0) {
        printf("  [SKIP] 載 D3TXT01 失敗: %s\n", err); free(cty); return fails ? 1 : 0;
    }
    if (talk_b4 >= 0) {
        int ok = dq3_dialogue_open(&dlg, talk_b4);
        CHECK(ok == 0 && dlg.len > 0, "NPC byte4 對話記錄非空");
        printf("  NPC b4=0x%02x → D3TXT01 rec,glyph 數=%d\n", talk_b4, dlg.len);
    }

    /* set_bank 切到 D3TXT02(羅馬利亞)應 reload 成功且記錄數改變 */
    {
        int n1 = dlg.txt.n_records;
        int rc = dq3_dialogue_set_bank(&dlg, dir, 2, err, sizeof err);
        CHECK(rc == 0, "set_bank(2) reload D3TXT02 成功");
        CHECK(dlg.txt.n_records != n1 || n1 == 0, "reload 後記錄數更新");
        CHECK(dq3_dialogue_set_bank(&dlg, dir, 99, err, sizeof err) != 0, "set_bank 越界(99)被拒");
    }
    dq3_dialogue_free(&dlg); free(cty);

    printf(fails ? "== %d 項失敗 ==\n" : "== 全部通過 ==\n", fails);
    return fails ? 1 : 0;
}
