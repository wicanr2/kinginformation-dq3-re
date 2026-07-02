/* dq3_config.h — 可攜設定(跨 Windows/macOS/Android/Linux)。
 *
 * 為什麼不用 env:env 只在命令列/CI 方便;Windows 捷徑、macOS Finder、Android/iOS
 * 由系統啟動的 app 拿不到自訂 env。正式設定走「設定檔(key=value 純文字)」——
 * 檔案 I/O 各平台都有,路徑日後可換成各平台 app-data 夾。
 *
 * 套用順序(後者覆寫前者):內建預設 → 設定檔 → env(僅 dev/CI override)。
 * 設定檔格式:每行 `key=value`;`#` 開頭為註解。目前鍵:`rng=dos|real`。
 */
#ifndef DQ3_CONFIG_H
#define DQ3_CONFIG_H

typedef struct {
    int rng_mode;        /* DQ3_RNG_DOS / DQ3_RNG_REAL */
    int music_enabled;   /* 0/1:精訊 SB FM 音樂(MBG.MCX)開關 */
    int music_volume;    /* 0..100 */
    int audio_backend;   /* 0=SB FM 即時合成、1=MT-32 預錄 WAV */
    int combat_info;     /* 0/1:戰鬥顯示敵人 HP + 下一輪預計動作(remake 增強)*/
    int combat_hurt_fx;  /* 受傷特效時長 ms(0=關;100~2000,預設 500;原版特效 RE file 0x13846)*/
} dq3_config;

/* 預設值(rng=DOS 忠實)。 */
void dq3_config_default(dq3_config *c);

/* 從設定檔載入(找到的鍵覆寫 c;檔不存在 → 不變動)。回讀到的鍵數。 */
int  dq3_config_load(dq3_config *c, const char *path);

/* 寫設定檔(供日後遊戲內設定選單存檔)。回 0=成功。 */
int  dq3_config_save(const dq3_config *c, const char *path);

/* 預設設定檔路徑(env DQ3_CONFIG 可覆寫;否則 "dq3.cfg")。 */
const char *dq3_config_path(void);

#endif /* DQ3_CONFIG_H */
