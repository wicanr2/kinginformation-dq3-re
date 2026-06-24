/* dq3_monster.c — 怪物數值 + DQ3MNS.SHP sprite 解碼。 */
#include "dq3_monster.h"
#include "dq3_restored_sprites.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static uint16_t rd16(const uint8_t *d, size_t o) { return (uint16_t)(d[o] | (d[o+1] << 8)); }
static uint32_t rd32(const uint8_t *d, size_t o) {
    return (uint32_t)d[o] | ((uint32_t)d[o+1]<<8) | ((uint32_t)d[o+2]<<16) | ((uint32_t)d[o+3]<<24);
}

static uint8_t *slurp(const char *dir, const char *name, size_t *len)
{
    char path[2048]; FILE *f; long sz; uint8_t *b;
    snprintf(path, sizeof path, "%s/%s", dir, name);
    f = fopen(path, "rb"); if (!f) { *len = 0; return NULL; }
    fseek(f,0,SEEK_END); sz=ftell(f); fseek(f,0,SEEK_SET);
    if (sz<=0){fclose(f);*len=0;return NULL;}
    b=malloc((size_t)sz);
    if(!b||fread(b,1,(size_t)sz,f)!=(size_t)sz){fclose(f);free(b);*len=0;return NULL;}
    fclose(f); *len=(size_t)sz; return b;
}

int dq3_monsters_load(dq3_monsters *m, const char *dir, char *err, int errcap)
{
    memset(m,0,sizeof *m);
    m->raw = slurp(dir, "D3MNS.DAT", &m->raw_len);
    if (!m->raw) { if(err) snprintf(err,errcap,"open D3MNS.DAT"); return -1; }
    if (m->raw_len < (size_t)DQ3_MONSTER_COUNT*DQ3_MONSTER_REC) {
        if(err) snprintf(err,errcap,"D3MNS.DAT too small"); free(m->raw); m->raw=NULL; return -1;
    }
    m->loaded = 1;
    return 0;
}

void dq3_monsters_free(dq3_monsters *m){ if(m){free(m->raw); m->raw=NULL; m->loaded=0;} }

int dq3_monster_get_stat(const dq3_monsters *m, int id, dq3_monster_stat *o)
{
    const uint8_t *r;
    if (!m->loaded || id<0 || id>=DQ3_MONSTER_COUNT) return -1;
    r = m->raw + (size_t)id*DQ3_MONSTER_REC;
    o->hp_base = rd16(r,0x00); o->hp_rand = rd16(r,0x02);
    o->atk = r[0x08]; o->def = rd16(r,0x09); o->agi = rd16(r,0x0b);
    o->flee_resist = r[0x18];
    o->exp = rd16(r,0x21); o->gold = rd16(r,0x23); o->spawn_weight = r[0x28];
    return 0;
}

int dq3_monster_get_ai(const dq3_monsters *m, int id, dq3_monster_ai *o)
{
    const uint8_t *r; int i;
    if (!m->loaded || id<0 || id>=DQ3_MONSTER_COUNT) return -1;
    r = m->raw + (size_t)id*DQ3_MONSTER_REC;
    o->cast_prob = r[0x0d];          /* docs/37 */
    o->flee_thresh = r[0x17];
    o->flee_rate = r[0x18];
    for (i=0;i<6;i++) o->spell_mask[i] = r[0x0e + i];
    return 0;
}

int dq3_monster_spell_count(const dq3_monster_ai *ai)
{
    int i, b, c=0;
    for (i=0;i<6;i++) for (b=0;b<8;b++) if (ai->spell_mask[i] & (0x80>>b)) c++;
    return c;
}

uint16_t dq3_monster_hp_max(const dq3_monsters *m, int id)
{
    dq3_monster_stat s;
    if (dq3_monster_get_stat(m,id,&s)!=0) return 0;
    return (uint16_t)(s.hp_base + s.hp_rand);
}

/* SHP:131×u32 offset table + 每隻 {u16 w_bytes, u16 h, 4 plane plane-major}。 */
int dq3_monster_sprite_decode(const char *dir, int id, dq3_monster_sprite *out,
                              char *err, int errcap)
{
    size_t n; uint8_t *d; uint32_t off, nxt; uint16_t wb, h; int W, s, r, b, bit;
    #define FAIL(m) do{ if(err)snprintf(err,errcap,"%s",m); free(d); return -1;}while(0)
    d = slurp(dir, "DQ3MNS.SHP", &n);
    if (!d) { if(err)snprintf(err,errcap,"open DQ3MNS.SHP"); return -1; }
    if (id<0 || id>=DQ3_MONSTER_COUNT) FAIL("id range");
    if ((size_t)(id+1)*4+4 > n) FAIL("offset table oob");
    off = rd32(d,(size_t)id*4); nxt = rd32(d,(size_t)(id+1)*4);
    if (off+4 > n || nxt <= off) FAIL("empty/bad sprite");   /* 空 sprite(如未完成 boss)*/
    wb = (uint16_t)(rd16(d,off) & 0x7fff); h = rd16(d,off+2);
    W = wb*8;
    if (W<=0 || h<=0 || W>DQ3_SHP_MAXW || h>DQ3_SHP_MAXH) FAIL("size out of range");
    memset(out,0,sizeof *out); out->w=W; out->h=h;
    {
        size_t base = off+4; size_t plane_sz = (size_t)wb*h;
        if (base + plane_sz*4 > n) FAIL("sprite data oob");
        for (s=0;s<4;s++){
            int pl = 3-s;   /* plane3 先(Map Mask 0x08→0x01)*/
            size_t b0 = base + (size_t)s*plane_sz;
            for (r=0;r<h;r++)
                for (b=0;b<wb;b++){
                    uint8_t v = d[b0 + (size_t)r*wb + b];
                    for (bit=0;bit<8;bit++)
                        if (v & (0x80>>bit)) out->px[r][b*8+bit] |= (uint8_t)(1<<pl);
                }
        }
        for (r=0;r<h;r++)
            for (b=0;b<W;b++)
                out->opaque[r][b] = out->px[r][b] ? 1 : 0;  /* 色0=透明 */
    }
    free(d);
    return 0;
    #undef FAIL
}

int dq3_monster_sprite_get(const char *dir, int id, dq3_monster_sprite *out,
                           char *err, int errcap)
{
    if (dq3_monster_sprite_decode(dir, id, out, err, errcap) == 0) return 0;
    /* 原版空 sprite(未完成 boss 128/129)→ 回退復原資料(bug #3 顯示而非當機)*/
    if (dq3_restored_sprite(id, out) == 0) return 0;
    return -1;
}
