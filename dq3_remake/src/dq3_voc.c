/* dq3_voc — 見 dq3_voc.h。VCX bank 解碼 + 重取樣到目標 rate。 */
#include "dq3_voc.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* 解一個 VOC sound-data block(type 1)→ 配置 s16,重取樣到 target_rate。回 1 成功。 */
static int decode_block(const unsigned char *p, long avail, int target_rate, dq3_sfx *out)
{
    long blen; int div, codec, src_rate, src_n, i;
    const unsigned char *pcm;
    if (avail < 6 || p[0] != 0x01) return 0;            /* 只處理 type 1 sound data */
    blen = (long)p[1] | ((long)p[2] << 8) | ((long)p[3] << 16);   /* 含 div+codec 兩 byte */
    div = p[4]; codec = p[5];
    if (codec != 0) return 0;                            /* 只支援 8-bit unsigned PCM */
    src_rate = (div < 256) ? (int)(1000000L / (256 - div)) : 11025;
    if (src_rate <= 0) src_rate = 8000;
    src_n = (int)(blen - 2);                             /* PCM byte 數 */
    if (src_n <= 0) return 0;
    if (6 + src_n > avail) src_n = (int)(avail - 6);     /* 防越界 */
    pcm = p + 6;
    {
        long dst_n = (long)src_n * target_rate / src_rate;
        short *s = (short *)malloc((dst_n > 0 ? dst_n : 1) * sizeof(short));
        if (!s) return 0;
        for (i = 0; i < dst_n; i++) {
            long si = (long)i * src_rate / target_rate;
            int u = (si < src_n) ? pcm[si] : 128;        /* unsigned 8-bit,中心 128 */
            s[i] = (short)((u - 128) * 200);             /* → s16(留 headroom 供混音)*/
        }
        out->pcm = s; out->len = (int)dst_n;
    }
    return 1;
}

int dq3_voc_load(dq3_voc_bank *bank, const char *path, int target_rate)
{
    FILE *f = fopen(path, "rb");
    unsigned char *d; long len, i; long *offs; int n = 0, cap = 64, k;
    memset(bank, 0, sizeof *bank);
    if (!f) return 0;
    fseek(f, 0, SEEK_END); len = ftell(f); fseek(f, 0, SEEK_SET);
    d = (unsigned char *)malloc(len);
    if (!d || fread(d, 1, len, f) != (size_t)len) { free(d); fclose(f); return 0; }
    fclose(f);

    offs = (long *)malloc(cap * sizeof(long));
    i = 0;
    while (i + 4 <= len && n < cap) {
        unsigned long v = (unsigned long)d[i] | ((unsigned long)d[i+1] << 8)
                        | ((unsigned long)d[i+2] << 16) | ((unsigned long)d[i+3] << 24);
        if (n > 0 && (long)v <= offs[n-1]) break;
        if ((long)v >= len || (long)v < (n + 1) * 4) break;
        offs[n++] = (long)v; i += 4;
    }
    if (n <= 0) { free(offs); free(d); return 0; }

    bank->sfx = (dq3_sfx *)calloc(n, sizeof(dq3_sfx));
    bank->count = 0;
    for (k = 0; k < n; k++) {
        long end = (k + 1 < n) ? offs[k+1] : len;
        long avail = end - offs[k];
        if (decode_block(d + offs[k], avail, target_rate, &bank->sfx[k]))
            bank->count = k + 1;          /* 連續成功計數;失敗的留空(pcm=NULL)*/
    }
    free(offs); free(d);
    return bank->count;
}

void dq3_voc_free(dq3_voc_bank *bank)
{
    int i;
    if (!bank || !bank->sfx) return;
    for (i = 0; i < bank->count; i++) free(bank->sfx[i].pcm);
    free(bank->sfx);
    memset(bank, 0, sizeof *bank);
}
