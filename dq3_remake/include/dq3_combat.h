/* dq3_combat.h — 戰鬥道具特效(#7a 隼劍雙擊 / #7b 魔甲抗魔),含 C 層修正。
 *
 * 資料來源:使用者 DQ3.EXE 旁的 ITEM.DAT(128 筆 × 7 byte),不入 git。
 *   record[5](+5):武器旗標,bit7(0x80)= double-hit(飛鷹劍 0x6e:+5=0xc0)。
 *   record[6](+6):防具旗標,bit2(0x04)= 抗魔法(魔法鎧甲 0x2b:+6=0xd4、天使袍 +6=0x1c)。
 *
 * 精訊版這兩個特效「資料齊全、程式從不讀取」(docs/18/22):攻擊碼無攻擊次數變數、
 * 咒文扣血不掃裝備抗魔。C 層補回(通用旗標偵測,非寫死單一 item code):
 *   #7a 攻擊次數 = (武器 +5 bit7) ? 2 : 1。
 *   #7b 咒文傷害:受方任一裝備 +6 bit2 置位 → 減半。
 */
#ifndef DQ3_COMBAT_H
#define DQ3_COMBAT_H

#include <stdint.h>

#define DQ3_ITEM_COUNT   128
#define DQ3_ITEM_STRIDE  7
#define DQ3_EQUIP_SLOTS  8
#define DQ3_ITEM_EMPTY   0xff

#define DQ3_WPN_DOUBLE_HIT  0x80   /* item +5 bit7 */
#define DQ3_ARM_MAGIC_RESIST 0x04  /* item +6 bit2 */

typedef struct {
    uint8_t item[DQ3_ITEM_COUNT][DQ3_ITEM_STRIDE];
    int     loaded;
} dq3_items;

int dq3_items_load(dq3_items *it, const char *assets_dir, char *err, int errcap);

/* ITEM.DAT 數值(RE docs/22):b0 攻 / b1 防 / 價(b2+b3*256)/ 類別 b4 / 可裝備職業 b6(0xff=全/0x80=勇者)。 */
int dq3_item_attack(const dq3_items *it, int code);
int dq3_item_defense(const dq3_items *it, int code);
int dq3_item_price(const dq3_items *it, int code);
int dq3_item_category(const dq3_items *it, int code);
int dq3_item_can_equip(const dq3_items *it, int code, int cls);

/* 裝備部位(RE:ITEM.DAT b4 高位,精訊版 4 槽)。回 0=武器 1=鎧 2=盾 3=兜;非裝備回 -1。
 * 公式:b4<0x20 → 非裝備;否則 (b4>>5)−1(0x2_武器/0x4_鎧/0x6_盾/0x8_兜,def 遞增佐證)。
 * 飾品(戒指/手環)在精訊版無乾淨部位編碼(b4=0x00 或同道具 0x18),不列裝備槽。 */
enum { DQ3_EQ_WEAPON = 0, DQ3_EQ_BODY = 1, DQ3_EQ_SHIELD = 2, DQ3_EQ_HEAD = 3, DQ3_EQ_SLOTS = 4 };
int dq3_item_equip_slot(const dq3_items *it, int code);

/* 道具旗標(code 越界回 0) */
uint8_t dq3_item_flag5(const dq3_items *it, int code);
uint8_t dq3_item_flag6(const dq3_items *it, int code);

/* #7a:依武器決定攻擊次數(通用 +5 bit7)。 */
int dq3_combat_num_attacks(const dq3_items *it, int weapon_code);

/* #7b:受方 8 格裝備是否含抗魔裝(任一 +6 bit2)。 */
int dq3_combat_has_magic_resist(const dq3_items *it, const uint8_t equip[DQ3_EQUIP_SLOTS]);

/* #7b:咒文最終傷害 — 受方抗魔則減半,否則原值。 */
int dq3_combat_spell_damage(const dq3_items *it, const uint8_t equip[DQ3_EQUIP_SLOTS], int base);

#endif /* DQ3_COMBAT_H */
