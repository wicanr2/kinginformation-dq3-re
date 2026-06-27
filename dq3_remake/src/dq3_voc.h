/* dq3_voc — 精訊 DQ3 數位音效(FVOC.VCX / NVOC.VCX)解碼。
 *
 * VCX bank 格式(RE 自素材,見 docs/57):
 *   開頭 = 連續遞增的 dword 偏移表(首筆 = 表大小);每筆指向一段 Creative VOC「sound data block」:
 *   byte0=0x01(type)、byte1-3=length、byte4=freq divisor(rate=1e6/(256-div))、byte5=codec(0=8-bit unsigned PCM)、其後 PCM。
 * 載入時把每個音效解成目標取樣率的 mono s16,供 dq3_audio 在 callback 直接混音。
 *
 * IP:只解碼使用者正版遊戲自身的音效資料(VCX,gitignore、不散布)。
 */
#ifndef DQ3_VOC_H
#define DQ3_VOC_H

typedef struct { short *pcm; int len; } dq3_sfx;     /* 已重取樣到目標 rate 的 mono s16 */

typedef struct {
    dq3_sfx *sfx;
    int      count;
} dq3_voc_bank;

/* 載入並解碼一個 VCX bank,全部重取樣到 target_rate。成功回 count(>0),失敗回 0。 */
int  dq3_voc_load(dq3_voc_bank *bank, const char *path, int target_rate);
void dq3_voc_free(dq3_voc_bank *bank);

#endif
