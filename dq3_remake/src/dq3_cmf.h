/* dq3_cmf — 精訊 DQ3 CMF 變體音軌 parser + 排程器(驅動 dq3_opl2)。
 *
 * 軌來源:MBG.MCX 拆出的 track_NN.bin(見 tools/extract_cmf.py、docs/57)。
 * 事件格式:MIDI-like channel-voice + 尾隨 1-byte delta + running status:
 *   9x note vel delta(vel=0 視為 off)、8x note vel delta、bx ctrl val delta、cx prog delta、ex lo hi delta。
 * 本模組把事件依 tempo(tick→樣本)推進、轉成 OPL 暫存器寫入(note→fnum/block、program→instrument patch)。
 */
#ifndef DQ3_CMF_H
#define DQ3_CMF_H

#include "dq3_opl2.h"

typedef struct dq3_cmf dq3_cmf;

/* 載入一條軌(複製需要的資料);sample_rate 與 opl 一致。失敗回 NULL。 */
dq3_cmf *dq3_cmf_load(const unsigned char *track, int len, dq3_opl2 *opl,
                      int sample_rate, double tick_ms);
void dq3_cmf_free(dq3_cmf *p);

/* render n 個樣本(過程中按時間推進事件、寫 OPL)。loop!=0 時放完從頭。
 * 回傳實際仍在播放(1)或已結束(0,僅 loop==0 時會回 0)。OPL 由呼叫端 render 取樣;
 * 本函式只負責「在正確樣本位置寫 OPL 暫存器」,並呼叫 dq3_opl2_render 填 out。 */
int dq3_cmf_render(dq3_cmf *p, short *out, int n, int loop);

/* 總事件數 / 估計長度(秒),供工具印資訊。 */
int dq3_cmf_event_count(const dq3_cmf *p);
double dq3_cmf_length_sec(const dq3_cmf *p);

#endif
