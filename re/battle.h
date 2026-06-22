/* battle.h — DQ3.EXE 戰鬥子系統宣告(遭遇 / 回合 / 指令 / 怪物資料 / sprite)
 *
 * 反編譯來源:tools/re_battle_dis.py(capstone,docker 內)反組譯 seg0 戰鬥函式群;
 * 對照 docs/13-exe-battle.md。real-mode large-model C 風格,求結構正確、可審閱,
 * 非位元精準。
 *
 * 位址慣例:sub_XXXX 的 XXXX = seg0 邏輯 offset(file off = 0x1370 + off)。
 * 全域沿用主資料段 DS=0x14dd 的 offset 標註(file base 0x16140)。
 *
 * 本檔只放本任務(battle.c)自宣告的型別與全域;runtime far-call wrapper 與既有
 * 素材載入 / 玩法函式請 include exe.h / gameplay.h(唯讀,不在此重複)。
 */
#ifndef DQ3_BATTLE_H
#define DQ3_BATTLE_H

#include <stdint.h>

#ifndef DQ3_EXE_H           /* 若未先 include exe.h,補基本型別 */
typedef uint8_t  u8;
typedef uint16_t u16;
typedef int16_t  i16;
#endif
#ifndef DQ3_GAMEPLAY_H
typedef int8_t   i8;
#endif

#ifndef far
#define far
#endif

/* =================================================================
 * 怪物參數表 D3MNS.DAT(loader sub_17ee → 整檔讀進 DS:0x0d78)
 *
 * 檔案大小 5330 = 130 筆 × 41(0x29)bytes,無檔頭(整檔即記錄陣列)。
 * 戰鬥怪物群生成(sub_a86d)以 monster_id*0x29 索引,讀「出現權重」欄位
 * (記錄相對 +0x28,即每筆最後一個 byte;table base DS:0x0da0 = 0x0d78+0x28)。
 *
 * 各欄位完整語意尚待逐筆對照(HP/攻防/經驗/金錢仍為推定),目前可確定:
 *   - 記錄定長 41 bytes、共 130 筆。
 *   - +0x28 = 出現權重 / 最小群量(>0 才可作為遭遇候選)。
 * 結構欄位以位移宣告,語意推定見註解(對 DOSBox 逐筆驗證為後續步驟)。
 * ================================================================= */
#define D3MNS_RECORD_SIZE   0x29   /* 41 bytes/筆 */
#define D3MNS_RECORD_COUNT  130    /* 5330 / 41 */
#define D3MNS_LOAD_OFF      0x0d78 /* DS 內載入位址(sub_17ee 目的)*/
#define D3MNS_WEIGHT_BASE   0x0da0 /* = 0x0d78 + 0x28,出現權重欄位表基底 */

typedef struct {
    /* 0x00..0x27:推定含名稱索引 / HP / MP / 攻擊 / 防禦 / 速度 / 經驗 / 金錢 /
     * sprite 編號等(逐欄語意待對 D3TXT 怪物名與 DOSBox 戰鬥數值驗證)。 */
    u8  fields[0x28];
    u8  spawn_weight;    /* +0x28  出現權重 / 群量基準(sub_a86d 讀此欄,>0 才出現)*/
} d3mns_record;

extern u8 d3mns_buf[];   /* DS:0x0d78  D3MNS.DAT 整檔(d3mns_record[130])*/

/* =================================================================
 * 怪物 sprite 圖集 DQ3MNS.SHP(loader sub_b19e;handle 常駐 DS:0x249b)
 *
 * 檔首 32-bit offset table(每筆 u32,LE);實測首個 offset = 0x20c → 131 筆
 * (130 怪 + 1 sentinel)。sub_b19e 流程:
 *   1. handle 為 0 才開檔(int21 AH=3Dh "DQ3MNS.SHP"),存 DS:0x249b。
 *   2. seek 到 sprite_id*4(offset table)→ 讀 8-byte sprite 表頭到 DS:0x248f。
 *   3. 表頭 = { u16 w?, u16 h?, u32 data_pos? };以表頭內 [0x2493]/[0x2495] 再 seek
 *      到像素資料,分頁讀入後交 VGA planar 驅動(seg 0x111b)blit(4-bit planar)。
 *
 * 表頭欄位語意(寬/高/資料位移)為推定,以實測 sprite0 = {16,123,..} 佐證。
 * ================================================================= */
#define SHP_OFFTAB_ENTRY_BYTES 4    /* u32/筆 */

typedef struct {
    u16 w;          /* +0  寬(像素 / planar word 數,推定)*/
    u16 h;          /* +2  高(列數,推定)*/
    u16 data_lo;    /* +4  像素資料位移低字(seek 用)*/
    u16 data_hi;    /* +6  像素資料位移高字 */
} shp_sprite_hdr;   /* 8 bytes,讀進 DS:0x248f */

extern u16 h_d3mns;          /* DS:0x249b  DQ3MNS.SHP 檔 handle(0=未開)*/
extern shp_sprite_hdr shp_hdr; /* DS:0x248f  最近讀入的 sprite 表頭 */

/* 怪物群顯示(sub_b16f):逐隻 sprite_id 取自 [0x2321],x 累進存 [0x2477],
 * 計數 [0x231f];每隻呼叫 sub_b19e blit。*/
extern u16 mob_group_count;  /* DS:0x231f  本場敵群隻數 */
extern u8  mob_sprite_id[];  /* DS:0x2321  各隻 sprite 編號(stride 2:id,count)*/

/* =================================================================
 * 戰鬥工作緩衝(由隊伍/敵群資料攤平到固定 DS 區,供回合運算)
 * ================================================================= */
extern u8  g_party_size;     /* DS:0x5077  我方隊伍人數(多處作迴圈次數)*/
extern u8  g_party_buf[];    /* DS:0x063a  4 名隊員戰鬥資料(sub_c8c6 攤平)*/
extern u16 g_enemy_total;    /* DS:0x2500  敵方總數(指令迴圈用)*/
extern u8  g_cmd_count;      /* DS:0x2515  本回合待下指令的我方角色數 */
extern u16 g_battle_state;   /* DS:0x2518  戰鬥狀態 / 結果位元(bit2=逃跑離場…)*/
extern u8  g_escape_flag;    /* DS:0x2517  0=戰鬥續行 / 1=遭遇取消(無敵)/ 2=逃跑成功 */
extern u8  g_encounter_step; /* DS:0x52f4  遭遇步數倒數計數器 */
extern u16 g_mode_flag;      /* DS:0x4f3b  模式旗標(>=2:Space/Enter 改走 sub_434f)*/
extern u16 g_event_flag;     /* DS:0x4f46  事件旗標(bit0x1000 = 強制遭遇)*/

/* =================================================================
 * 函式(本任務反編譯)
 * ================================================================= */

/* 遭遇進入(由 update_hud / sub_9530 每幀末呼叫)*/
int  battle_check_encounter(void);  /* sub_bd97  步數倒數,歸 0 擲新計數 + 起戰 */
void battle_main(void);             /* sub_bddf  戰鬥主流程(生群→繪→回合→結算)*/

/* 遭遇怪物群生成 + 區域查表 */
void encounter_build_group(void);   /* sub_a7d5  由玩家 tile 查遭遇區 + 生成敵群 */

/* sprite / 顯示 */
void mob_draw_group(void);          /* sub_b16f  逐隻取 SHP sprite blit */
void shp_load_sprite(u16 id);       /* sub_b19e  seek offset table → 讀表頭 → 載入像素 */

/* 戰鬥背景 / 場景切換 */
void battle_enter_screen(void);     /* sub_bfd1  載入 packbg 戰鬥背景(page 0x13)*/
void battle_leave_screen(void);     /* sub_c03f  還原地表/城鎮(重載 BLK)*/

/* HUD / 狀態列 */
void battle_draw_status(void);      /* sub_c572  畫我方狀態框(隊伍人數定高度)*/
void battle_draw_names(void);       /* sub_c59b  逐隻畫敵名(文字繪製器 111b:264)*/
void battle_setup_party(void);      /* sub_c8c6  攤平隊伍戰鬥資料到 DS:0x063a */

/* 回合 / 指令 */
int  party_alive_count(void);       /* sub_c5f6  數存活隊員(回傳 bp;設 0x80 旗標)*/
void battle_command_loop(void);     /* sub_c08b  逐角色開指令選單(戰/逃/防/道具)*/
void battle_redraw_field(void);     /* sub_bcf2  戰後/逃跑回場狀態列重繪 */
void battle_run_away(void);         /* sub_c7d9  逃跑:還原地圖座標 + 退出戰鬥 */

#endif /* DQ3_BATTLE_H */
