/* dq3_opl2 — 精簡 OPL2 (YM3812) FM 音源核心(暫存器層介面)。
 *
 * 用途:播放精訊 DQ3 從 MBG.MCX 抽出的 CMF 變體音軌(由 dq3_cmf 解析後寫 OPL 暫存器)。
 * 非 cycle-accurate;為「聽感接近 AdLib/聲霸卡 FM」的近似實作(2-op FM + 4 波形 + ADSR + fnum/block)。
 * 需要更精確音色時可換 DBOPL/Nuked-OPL,介面相同(write reg/val + render)。
 *
 * 9 個 2-operator channel。暫存器配置同 YM3812:
 *   0x20-0x35 AM/VIB/EGT/KSR/MULT、0x40-0x55 KSL/TL、0x60-0x75 AR/DR、0x80-0x95 SL/RR、
 *   0xA0-0xA8 F-num low、0xB0-0xB8 KeyOn/Block/F-num high、0xC0-0xC8 FB/CNT、0xE0-0xF5 waveform。
 */
#ifndef DQ3_OPL2_H
#define DQ3_OPL2_H

typedef struct dq3_opl2 dq3_opl2;

dq3_opl2 *dq3_opl2_new(int sample_rate);
void dq3_opl2_free(dq3_opl2 *o);
void dq3_opl2_reset(dq3_opl2 *o);
void dq3_opl2_write(dq3_opl2 *o, int reg, int val);     /* 寫一個 OPL 暫存器 */
void dq3_opl2_render(dq3_opl2 *o, short *out, int nsamples); /* 產生 nsamples 個 mono s16 */

#endif
