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
