/* dq3_audio — remake 的 SDL2 音訊子系統:播放精訊自己的 SB FM 音樂(MBG.MCX 的 18 軌)。
 *
 * 架構:MBG.MCX(偏移表 + CMF 變體軌)→ dq3_cmf 解析 → dq3_opl2 FM 合成 → SDL audio callback。
 * IP:只播放使用者正版遊戲自身的音樂資料(MBG.MCX,gitignore、不散布;remake 又有原版 EXE 指紋閘)。
 *
 * 全部 API 在 headless / dummy audio 下安全 no-op(不出聲、不崩),game_tester 零回歸。
 */
#ifndef DQ3_AUDIO_H
#define DQ3_AUDIO_H

/* 初始化:讀 <assets>/MBG.MCX、開 SDL audio。回 0 成功(含「audio 開不起來但不致命」的降級)。
 * 必須在 SDL_Init(含 SDL_INIT_AUDIO)之後呼叫。 */
int  dq3_audio_init(const char *assets_dir);
void dq3_audio_shutdown(void);

/* 播放 MBG.MCX 第 track_id 軌(0..track_count-1);loop!=0 循環。track_id 同前則不重啟。 */
void dq3_audio_play(int track_id, int loop);
void dq3_audio_stop(void);

void dq3_audio_set_enabled(int on);     /* 關閉時靜音並停止排程 */
int  dq3_audio_enabled(void);
void dq3_audio_set_volume(int vol_0_100);
int  dq3_audio_track_count(void);
int  dq3_audio_current_track(void);     /* 目前軌;-1=無 */

/* 場景 → 建議軌號的對應(集中一處,日後可依 RE 選曲邏輯精校)。回 -1 = 不指定。 */
enum {
    DQ3_MUS_TITLE = 0, DQ3_MUS_FIELD, DQ3_MUS_TOWN, DQ3_MUS_DUNGEON,
    DQ3_MUS_BATTLE, DQ3_MUS_BOSS, DQ3_MUS_CASTLE, DQ3_MUS_SHIP,
    DQ3_MUS_ENDING, DQ3_MUS__COUNT
};
int  dq3_audio_scene_track(int scene_kind);   /* scene_kind ∈ DQ3_MUS_* → MBG.MCX 軌號 */
void dq3_audio_play_scene(int scene_kind, int loop);  /* 便捷:查表 + play */

#endif
