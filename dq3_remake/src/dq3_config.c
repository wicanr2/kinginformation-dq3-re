/* dq3_config.c — 可攜設定檔(key=value)。見 dq3_config.h。 */
#include "dq3_config.h"
#include "dq3_rng.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void dq3_config_default(dq3_config *c)
{
    c->rng_mode = DQ3_RNG_DOS;   /* 預設忠實 */
    c->music_enabled = 1;
    c->music_volume = 70;
    c->audio_backend = 1;   /* 預設 MT-32(無 MT-32 音檔時自動退回 SB FM)*/
    c->combat_info = 1;     /* 預設開:戰鬥顯示敵人 HP + 預計動作 */
    c->combat_hurt_fx = 1;  /* 預設開:我方受傷 震動+黃綠閃光 */
}

const char *dq3_config_path(void)
{
    const char *p = getenv("DQ3_CONFIG");
    return (p && *p) ? p : "dq3.cfg";
}

/* 去頭尾空白(就地)。 */
static char *trim(char *s)
{
    char *e;
    while (*s == ' ' || *s == '\t') s++;
    e = s + strlen(s);
    while (e > s && (e[-1]=='\n' || e[-1]=='\r' || e[-1]==' ' || e[-1]=='\t')) *--e = 0;
    return s;
}

int dq3_config_load(dq3_config *c, const char *path)
{
    FILE *f = fopen(path, "r");
    char line[256]; int got = 0;
    if (!f) return 0;
    while (fgets(line, sizeof line, f)) {
        char *s = trim(line), *eq, *key, *val;
        if (*s == '#' || *s == 0) continue;
        eq = strchr(s, '=');
        if (!eq) continue;
        *eq = 0; key = trim(s); val = trim(eq + 1);
        if (strcmp(key, "rng") == 0) {
            c->rng_mode = (val[0]=='r' || val[0]=='R') ? DQ3_RNG_REAL : DQ3_RNG_DOS;
            got++;
        } else if (strcmp(key, "music") == 0) {
            c->music_enabled = (val[0]=='1' || val[0]=='o' || val[0]=='O' || val[0]=='y') ? 1 : 0;
            got++;
        } else if (strcmp(key, "music_vol") == 0) {
            int v = atoi(val); c->music_volume = v < 0 ? 0 : (v > 100 ? 100 : v);
            got++;
        } else if (strcmp(key, "audio") == 0) {
            c->audio_backend = (val[0]=='m' || val[0]=='M') ? 1 : 0;   /* mt32 / sb */
            got++;
        } else if (strcmp(key, "combat_info") == 0) {
            c->combat_info = (val[0]=='1' || val[0]=='o' || val[0]=='O' || val[0]=='y') ? 1 : 0;
        } else if (strcmp(key, "combat_hurt_fx") == 0) {
            c->combat_hurt_fx = (val[0]=='1' || val[0]=='o' || val[0]=='O' || val[0]=='y') ? 1 : 0;
            got++;
        }
    }
    fclose(f);
    return got;
}

int dq3_config_save(const dq3_config *c, const char *path)
{
    FILE *f = fopen(path, "w");
    if (!f) return -1;
    fputs("# DQ3 remake 設定(可攜;取代 env)\n", f);
    fprintf(f, "rng=%s\n", c->rng_mode == DQ3_RNG_REAL ? "real" : "dos");
    fprintf(f, "music=%d\n", c->music_enabled ? 1 : 0);
    fprintf(f, "music_vol=%d\n", c->music_volume);
    fprintf(f, "audio=%s\n", c->audio_backend ? "mt32" : "sb");
    fprintf(f, "combat_info=%d\n", c->combat_info ? 1 : 0);
    fprintf(f, "combat_hurt_fx=%d\n", c->combat_hurt_fx ? 1 : 0);
    fclose(f);
    return 0;
}
