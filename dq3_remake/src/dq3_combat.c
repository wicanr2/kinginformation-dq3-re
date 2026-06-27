/* dq3_combat.c — 戰鬥道具特效實作(#7a/#7b)。 */
#include "dq3_combat.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int dq3_items_load(dq3_items *it, const char *assets_dir, char *err, int errcap)
{
    char path[2048]; FILE *f; long sz; uint8_t *d = NULL;
    #define FAIL(m) do { if (err) snprintf(err, errcap, "%s", m); free(d); return -1; } while (0)
    memset(it, 0, sizeof *it);
    snprintf(path, sizeof path, "%s/ITEM.DAT", assets_dir);
    f = fopen(path, "rb");
    if (!f) FAIL("open ITEM.DAT failed");
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    if (sz < DQ3_ITEM_COUNT * DQ3_ITEM_STRIDE) { fclose(f); FAIL("ITEM.DAT too small"); }
    d = (uint8_t *)malloc((size_t)sz);
    if (!d || fread(d, 1, (size_t)sz, f) != (size_t)sz) { fclose(f); FAIL("read ITEM.DAT"); }
    fclose(f);
    memcpy(it->item, d, DQ3_ITEM_COUNT * DQ3_ITEM_STRIDE);
    it->loaded = 1;
    free(d);
    return 0;
    #undef FAIL
}

uint8_t dq3_item_flag5(const dq3_items *it, int code)
{
    if (code < 0 || code >= DQ3_ITEM_COUNT) return 0;
    return it->item[code][5];
}

uint8_t dq3_item_flag6(const dq3_items *it, int code)
{
    if (code < 0 || code >= DQ3_ITEM_COUNT) return 0;
    return it->item[code][6];
}

/* ITEM.DAT 7-byte 結構(RE,docs/22):b0 攻、b1 防、b2+b3*256 價、b4 類別/部位、
 * b5 武器旗標、b6 可裝備職業 bitmask。 */
/* 洛特裝備昇華(E-11 時代結尾):破索瑪→勇者受冊封為「洛特」,三件勇者專用裝備化為傳說的洛特裝備
 * (精訊版 ITEM.DAT 本無洛特道具,remake 補上以圓「傳說的終章」標題)。王者之劍→洛特之劍、
 * 光之鎧甲→洛特之鎧、勇者之盾→洛特之盾;數值對齊 DQ3 lore(ロト裝備)再昇華為結局獎勵。 */
static int g_loto_blessed = 0;
void dq3_set_loto_blessed(int v) { g_loto_blessed = v ? 1 : 0; }
int  dq3_loto_blessed(void) { return g_loto_blessed; }

int dq3_item_attack(const dq3_items *it, int code)
{ int a = (code<0||code>=DQ3_ITEM_COUNT)?0:it->item[code][0];
  if (g_loto_blessed && code==0x1c) a = 150;          /* 王者之劍 120 → 洛特之劍 150 */
  return a; }
int dq3_item_defense(const dq3_items *it, int code)
{ int d = (code<0||code>=DQ3_ITEM_COUNT)?0:it->item[code][1];
  if (g_loto_blessed) { if (code==0x3f) d = 95;       /* 光之鎧甲 70 → 洛特之鎧 95 */
                        else if (code==0x40) d = 55; }/* 勇者之盾 35 → 洛特之盾 55 */
  return d; }
int dq3_item_price(const dq3_items *it, int code)
{ return (code<0||code>=DQ3_ITEM_COUNT)?0:(it->item[code][2] | (it->item[code][3]<<8)); }
int dq3_item_category(const dq3_items *it, int code)
{ return (code<0||code>=DQ3_ITEM_COUNT)?0:it->item[code][4]; }
int dq3_item_equip_slot(const dq3_items *it, int code)
{ int b4 = dq3_item_category(it, code);
  return (b4 < 0x20) ? -1 : ((b4 >> 5) - 1); }   /* 0x2_武器/0x4_鎧/0x6_盾/0x8_兜;非裝備 -1 */
/* 職業 cls(0..7)可否裝備此道具(b6 bitmask;0xff=全職業、0x80=勇者專用)。 */
int dq3_item_can_equip(const dq3_items *it, int code, int cls)
{ if(code<0||code>=DQ3_ITEM_COUNT||cls<0||cls>7) return 0;
  return (it->item[code][6] & (0x80>>cls)) ? 1 : 0; }

int dq3_combat_num_attacks(const dq3_items *it, int weapon_code)
{
    return (dq3_item_flag5(it, weapon_code) & DQ3_WPN_DOUBLE_HIT) ? 2 : 1;
}

int dq3_combat_has_magic_resist(const dq3_items *it, const uint8_t equip[DQ3_EQUIP_SLOTS])
{
    int i;
    for (i = 0; i < DQ3_EQUIP_SLOTS; i++) {
        int c = equip[i];
        if (c != DQ3_ITEM_EMPTY && (dq3_item_flag6(it, c) & DQ3_ARM_MAGIC_RESIST))
            return 1;
    }
    return 0;
}

int dq3_combat_spell_damage(const dq3_items *it, const uint8_t equip[DQ3_EQUIP_SLOTS], int base)
{
    if (base < 0) base = 0;
    return dq3_combat_has_magic_resist(it, equip) ? base / 2 : base;
}
