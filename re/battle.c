/* battle.c — DQ3.EXE 戰鬥子系統反編譯(RE → C)
 *
 * 反編譯來源:tools/re_battle_dis.py(capstone,docker 內)反組譯 seg0 戰鬥函式群;
 * 對照 docs/13-exe-battle.md。real-mode large-model C 風格,忠實鏡射反組譯流程,
 * 以具名變數取代裸 DS:offset。求結構正確、可審閱,非位元精準(逐函式對 DOSBox
 * 驗證為後續步驟)。
 *
 * 反編譯對應(seg0 off):
 *   battle_check_encounter sub_bd97   battle_main          sub_bddf
 *   encounter_build_group  sub_a7d5   mob_draw_group       sub_b16f
 *   shp_load_sprite        sub_b19e   battle_enter_screen  sub_bfd1
 *   battle_setup_party     sub_c8c6   battle_command_loop  sub_c08b
 *   party_alive_count      sub_c5f6   battle_run_away      sub_c7d9
 */
#include "exe.h"
#include "battle.h"

/* ===== runtime / 既有函式 forward-declare(細節在他檔,本任務不重複反編譯)===== */
extern u16 rng_next(u16 bx_range);   /* sub_e6c9:擲 0..bx-1 → DX(全程式亂數)*/
extern u16 rng_raw(void);            /* sub_e6b9:取原始亂數 → AL */
extern void txt_draw(u16 msg_id);    /* lcall 111b:264 文字繪製器(di=msg_id)*/
extern void vid_flip(void);          /* sub_ed39 等繪製/翻頁類 */

/* ===========================================================================
 * 遭遇進入:sub_bd97(由 update_hud / sub_9530 每幀末呼叫)
 *
 * 主迴圈每幀 → draw_scene → update_player → update_hud(sub_9530)→ 末端 call
 * 此函式。地表(或符合條件城鎮)行走時,每步把遭遇計數器 [0x52f4] 減 1;歸 0 時
 * 依「是否被事件強制遭遇([0x4f46]&0x1000)」擲新的步數計數,並進入 battle_main。
 *
 * 步數來源:
 *   - 強制遭遇:rng(5)+1。
 *   - 一般:重擲 rng(0x12) 直到 >=0xa(10),再視 [0x526c](近城/特定地形)-2。
 * ======================================================================== */
int battle_check_encounter(void)
{
    u16 dx;

    /* 只在地表移動 / 或城鎮中 [0xd77]==0 的情況下推進遭遇計數 */
    if (g_scene_flag == SCENE_TOWN) {
        extern u8 town_no_encounter; /* DS:0x0d77 */
        if (town_no_encounter == 0)
            return 0;                /* je 0xbdde:城鎮且禁遇 → 不戰 */
    }

    if (--g_encounter_step != 0)     /* dec [0x52f4]; jne */
        return 0;                    /* 尚未到遭遇步數 */

    /* 計數歸 0 → 擲下次遭遇步數,並起戰 */
    if (g_event_flag & 0x1000) {     /* 強制遭遇事件 */
        dx = rng_next(5) + 1;        /* rng(5)+1 */
    } else {
        do { dx = rng_next(0x12); } while ((i16)dx < 0xa); /* 重擲到 >=10 */
        extern u8 g_near_special;    /* DS:0x526c */
        if (g_near_special != 1)
            dx -= 2;
    }
    g_encounter_step = (u8)dx;       /* [0x52f4] = 新計數 */

    battle_main();                   /* call 0xbddf */
    return 1;
}

/* ===========================================================================
 * 戰鬥主流程:sub_bddf
 *
 * 生成敵群(encounter_build_group)→ 若遭遇被取消([0x2517]!=0)即返回 →
 * 數存活我方(party_alive_count → bp);若 0 → 結束 → 攤平隊伍資料 →
 * 載入戰鬥背景 → 畫狀態框 / 敵名 → 進入指令/回合迴圈(battle_command_loop)→
 * 結算(勝:重繪場景 + 離場;逃:battle_run_away)。
 *
 * 反組譯顯示函式體含多個近乎重複的「戰鬥回合段」(本幀重入 / 多輪),此處取
 * 第一輪結構作代表(0xbddf..0xbe88),語意對齊;後段為回合重播與結算分支。
 * ======================================================================== */
void battle_main(void)
{
    int alive;

    g_escape_flag = 0;             /* [0x2517]=0 */
    /* 清回合暫存 [0x24fa]/[0x24f6]/[0x24f8]=0 */

    encounter_build_group();       /* sub_a7d5:查遭遇區 + 生敵群 */
    if (g_escape_flag != 0)        /* 遭遇取消(空群 / 無敵地形)*/
        return;

    alive = party_alive_count();   /* sub_c5f6 → bp = 存活我方數 */
    if (alive == 0) {
        g_battle_state = 0;        /* [0x2518]=0 */
        return;                    /* 全滅 / 無人可戰 → 退出 */
    }

    /* [0x24b4] = alive(>0 旗標);[0x24b3] = enemy_total([0x2500])*/
    {
        extern u8 g_has_party, g_has_enemy; /* DS:0x24b4 / 0x24b3 */
        g_has_party = (u8)alive;
        g_has_enemy = (u8)g_enemy_total;
    }

    battle_setup_party();          /* sub_c8c6:攤平隊伍 HP/MP/狀態 → [0x063a] */
    battle_enter_screen();         /* sub_bfd1:載入 packbg 戰鬥背景(page 0x13)*/
    battle_draw_status();          /* sub_c572:畫我方狀態框 */
    battle_draw_names();           /* sub_c59b:逐隻畫敵名 + 怪物群 */
    battle_command_loop();         /* sub_c08b:逐角色下指令 + 回合執行 */

    /* [0xd75]=0;再數一次存活 */
    alive = party_alive_count();
    if (alive != 0) {
        /* === 勝利 / 繼續:重繪場景並離場 === */
        battle_redraw_field();     /* sub_bcf2 */
        vid_flip();                /* call 0xed39 */
        vid_mode(0x47);            /* lcall 109c:47 */
        file_op_dx(0xdc);          /* lcall 1053:dc */
        battle_leave_screen();     /* sub_c03f:重載地表/城鎮 BLK */
        if (g_scene_flag == SCENE_OVERWORLD) {
            /* 地表:call 0x118b0 收尾 + 事件清理 */
        } else {
            /* 城鎮:call 0x124f6 */
        }
        g_battle_state = 0;
        return;
    }

    /* === 逃跑路徑(alive==0 經逃跑分支進來)=== */
    battle_leave_screen();         /* sub_c03f */
    battle_run_away();             /* sub_c7d9:還原地圖座標、退出戰鬥 */
    g_battle_state = 0;
}

/* ===========================================================================
 * 遭遇怪物群生成:sub_a7d5
 *
 * 1. 由玩家當前 tile 座標 [0x4f2f](X)/[0x4f31](Y)算「遭遇區編號」:
 *    地表時 region = [(Y&0xf0)+(X>>4) + 0x4966],與門檻 [0x52f5] 比較取大;
 *    海面 / 城鎮 / 邊界(Y>=0x12c…)走固定 region。region-1 << 5 索引遭遇表。
 * 2. 取遭遇表項(每項 0x20 bytes,複製 8 bytes 到 [0x2317])→ 候選怪物 id 表
 *    [0x231b](最多數隻,0xff 終止)。
 * 3. 對每個候選 id:讀 D3MNS 出現權重 [id*0x29 + 0x0da0];以「戰鬥點數預算」
 *    [0x2513]=0x26(38)為上限,擲亂數決定生成隻數,扣預算,寫入敵群表
 *    [0x2321](sprite id)/[0x231f](群隻數)。
 *
 * 反組譯主體在 0xa86d 起;本處鏡射「權重讀取(*0x29)→ 預算分配 → 群表填寫」核心。
 * ======================================================================== */
void encounter_build_group(void)
{
    extern u8  enc_cand[];     /* DS:0x231b  候選怪物 id(0xff 終止)*/
    extern u16 enc_budget;     /* DS:0x2513  戰鬥點數預算(初值 0x26)*/
    extern u16 enc_weight;     /* DS:0x250f  目前怪物出現權重 */
    extern u8  enc_thresh;     /* DS:0x2318  亂數門檻 */
    u16 i, n, count;

    /* 區域查表 + 取候選表(0xa7d5..0xa86d)略;結果:enc_cand[] 候選 id 串。*/

    enc_budget = 0x26;         /* [0x2513]=0x26 預算 */
    mob_group_count = 0;       /* [0x231f]=0 */

    for (;;) {                 /* loop @ 0xa89f:隨機取候選表一格 */
        u16 idx = rng_next(4); /* bx=4 → rng(4) */
        u8  id  = enc_cand[idx];
        if (id == 0xff) continue;          /* 空格重擲 */
        enc_cand[idx] = 0xff;              /* 取走 */

        /* D3MNS 出現權重:id*0x29 + 0x0da0(= d3mns_buf[id].spawn_weight)*/
        enc_weight = d3mns_buf[id * D3MNS_RECORD_SIZE + 0x28];
        if ((i16)enc_weight <= 0)
            goto next;                      /* 權重<=0:不出現 */

        /* 以預算 / 權重算最多可生成隻數 count = floor(budget/weight)*/
        count = 0;
        { i16 b = (i16)enc_budget;
          while (b - (i16)enc_weight >= 0) { b -= (i16)enc_weight; count++; }
          /* 反組譯為 sub bx,ax; jl; inc cx 迴圈 */
        }
        if ((i16)count <= 0)
            goto next;

        /* 擲門檻決定實際隻數:rng_raw() < [0x2318] → 1 隻;否則 rng(count)+1 隻 */
        if (rng_raw() < enc_thresh) {
            mob_sprite_id[mob_group_count * 2] = id; /* di+1 處存 1(隻數)*/
            enc_budget -= (u16)(enc_weight * 1);
            mob_group_count++;
        } else {
            u16 k = rng_next(count) + 1;
            mob_sprite_id[mob_group_count * 2 + 1] = (u8)k;
            enc_budget -= (u16)(enc_weight * k);
            mob_group_count++;
        }
    next:
        if (mob_group_count >= 8) break;    /* 群上限(迴圈次數約束)*/
    }
    (void)i; (void)n;
}

/* ===========================================================================
 * 怪物 sprite 顯示:sub_b16f(群)+ sub_b19e(單隻 blit)
 *
 * mob_draw_group:逐隻取 [0x2321] 的 sprite id,x 偏移累進到 [0x2477],
 * 呼叫 shp_load_sprite 載入該 sprite 並交 VGA planar 驅動貼出。
 *
 * shp_load_sprite:DQ3MNS.SHP handle 常駐 [0x249b](0 才開檔)。以 sprite id*4
 * seek 到 offset table 取 32-bit 資料位移,讀 8-byte 表頭到 [0x248f],再以表頭
 * 內位移 seek 到像素資料分頁讀入,交 seg 0x111b blit(4-bit planar,AND-mask
 * 透明)。
 * ======================================================================== */
void mob_draw_group(void)
{
    extern u16 mob_x_accum;   /* DS:0x2477 */
    u16 cx = mob_group_count; /* [0x231f] */
    u8 *id_p   = mob_sprite_id;        /* di = [0x2321] */
    u16 *x_p   = &mob_x_accum;         /* si = [0x2477] */

    *x_p = 0;
    while (cx--) {
        u16 id = *id_p;                /* al = [di] */
        shp_load_sprite(id);           /* call 0xb19e(si=x_offset 隨呼叫累加)*/
        /* x 累進:cx += [si]; [si+2] = cx(下一隻 x)*/
        id_p += 2;
        x_p  += 1;
    }
}

void shp_load_sprite(u16 id)
{
    /* handle 常駐:0 才開檔 DQ3MNS.SHP(int21 AH=3Dh)→ [0x249b] */
    if (h_d3mns == 0) {
        /* dx = "DQ3MNS.SHP"(DS:0x102);AH=3Dh,AL=0 */
        extern int dos_open(const char *, int);
        extern const char fn_mns_shp[];
        h_d3mns = (u16)dos_open(fn_mns_shp, 0);
    }

    /* seek 到 offset table 第 id 筆(id*4),讀 u32 資料位移 */
    {
        extern long dos_lseek(int, long, int);
        extern int  dos_read(int, void *, unsigned);
        long tab_pos = (long)id << 2;             /* shl dx,1 ×2 = id*4 */
        dos_lseek(h_d3mns, tab_pos, 0);
        dos_read(h_d3mns, &shp_hdr, 8);            /* 8-byte sprite 表頭 → [0x248f] */

        /* 以表頭內資料位移 seek 到像素區(高低字組 32-bit)*/
        {
            long data_pos = ((long)shp_hdr.data_hi << 16) | shp_hdr.data_lo;
            dos_lseek(h_d3mns, data_pos, 0);       /* [0x2493]/[0x2495] */
            /* 之後分頁讀像素 → VGA planar 驅動 seg 0x111b blit(後段略)*/
        }
    }
}

/* ===========================================================================
 * 戰鬥背景 / 場景切換
 *
 * battle_enter_screen(sub_bfd1):凍滑鼠 → 設文字游標 [0x3e70]=0x12 / [0x3e72]=0xf8
 * → file_op 準備 → 載入 packbg 戰鬥背景(call 0xc688 → load_packbg_page,page 0x13)
 * → 套用 → modal on([0x72a]=1)。即「戰鬥畫面 = packbg 第 0x13 頁的藍天草原背景」。
 *
 * battle_leave_screen(sub_c03f):設文字游標 [0x3e70]=0x13 / [0x3e72]=0xee → 重載
 * 地圖 BLK(call 0xeaf4 load_blk)→ 還原音效/鍵盤/滑鼠,回到地表 / 城鎮。
 * ======================================================================== */
void battle_enter_screen(void)
{
    extern u16 g_txt_cur_x, g_txt_cur_y; /* DS:0x3e70 / 0x3e72 */
    mouse_op(0x385);                  /* lcall 10b6:385 */
    /* int33h ax=3 存滑鼠座標 [0x2852]/[0x2854] */
    mouse_op(0x90);                   /* lcall 10b6:90 */
    g_txt_cur_x = 0x12;
    g_txt_cur_y = 0xf8;
    file_op_dx(0x16c);                /* lcall 1053:16c */
    file_op_dx(0x240);                /* lcall 1053:240,bp=0x13 = packbg page */
    vid_flip();                       /* call 0xed39 */
    vid_mode(0x47);                   /* lcall 109c:47 */
    /* [0x72a]=1 modal on */
    /* call 0xc688 → load_packbg_page(page 0x13)= 戰鬥背景圖 */
    extern void load_packbg_caller(void); /* sub_c688 */
    load_packbg_caller();
    mouse_op(0x3a0);                  /* lcall 10b6:3a0 */
}

void battle_leave_screen(void)
{
    extern u16 g_txt_cur_x, g_txt_cur_y;
    g_txt_cur_x = 0x13;
    g_txt_cur_y = 0xee;
    /* call 0xee9b */
    load_blk();                       /* call 0xeaf4:重載地表/城鎮 tile 圖庫 */
    mouse_op(0x71);                   /* lcall 10b6:71 */
    /* call 0xf612;[0xd75]=0;[0xd]=0;lcall 1053:377 */
}

/* ===========================================================================
 * HUD / 狀態列
 *
 * battle_draw_status(sub_c572):以隊伍人數 [0x5077] 算狀態框高度
 * (height = size*10 + 4),設視窗矩形 [0x3e9c..] → 畫框(sub_f590)→ 填內容(0x8222)。
 *
 * battle_draw_names(sub_c59b):逐隻取敵群 sprite id [0x2321],msg_id = id+0x258,
 * 經文字繪製器 lcall 111b:264 畫出怪名(di=0x160 行起);再 loop 敵總數 [0x2500]
 * 畫敵體(sub_b1fe);最後等同步號 [5]==6(VBL / 計時)收尾。
 *
 * battle_setup_party(sub_c8c6):4 名隊員,各從隊員結構 [bx+0x4f0d] 取:
 *   +0x1a/+0x1c/+0x1e/+0x20 → 4 個 word(推定 HP/MaxHP/MP/MaxMP)
 *   +0x24 → word(推定 攻擊或經驗)
 *   +0x01/+0x2e/+0x2f → 3 個 byte(等級 / 狀態)
 * 攤平寫進戰鬥工作緩衝 [0x063a];同時掃 +0x3a 起的 8 格裝備/狀態,
 * 設睡眠(0x4d→bit0x800)/中毒(0x11→bit0x4000)等狀態旗標到 +0x38 / +0x22。
 * ======================================================================== */
void battle_draw_status(void)
{
    extern u16 win_rect[]; /* DS:0x3e9c.. 視窗矩形 */
    u16 size = g_party_size;          /* [0x5077] */
    /* [0x3e9e]=0x13;[0x3ea0]=8;[0x3ea8]=size;height = size*0xa + 4 → [0x3ea2] */
    win_rect[1] = 0x13;
    win_rect[2] = 8;
    win_rect[6] = size;
    win_rect[3] = size * 0xa + 4;
    /* call 0xf590(畫框)+ call 0x8222(填狀態)*/
}

void battle_draw_names(void)
{
    u16 cx = mob_group_count;         /* [0x231f] */
    u8 *p  = mob_sprite_id;           /* [0x2321] */
    /* [0x259b]=0 行計數歸零 */
    while (cx--) {
        u16 id  = *p;
        u16 msg = id + 0x258;         /* 怪名 D3TXT 訊息 ID = sprite_id + 0x258 */
        extern u16 g_txt_msg;         /* DS:0x2591 */
        g_txt_msg = msg;
        txt_draw(0x160);              /* di=0x160:文字繪製器畫怪名 */
        p += 2;
    }
    /* loop [0x2500] 隻畫敵體 sub_b1fe;等 [5]==6 */
}

void battle_setup_party(void)
{
    int m;
    /* call 0x8197 前置;cx=4 名 */
    for (m = 0; m < 4; m++) {
        extern u16 party_ptr_tab[];   /* DS:0x4f0d 起,每名一個結構指標 */
        u8 *src = (u8 *)(uintptr_t)party_ptr_tab[m];
        u8 *dst = &g_party_buf[m * 0xb]; /* 連續攤平(每名約 0xb bytes:5 word+3 byte)*/

        /* src+0x4d = 0(清旗標);複製 5 個 word(+0x1a/1c/1e/20/24)*/
        *(u16 *)(dst + 0) = *(u16 *)(src + 0x1a);
        *(u16 *)(dst + 2) = *(u16 *)(src + 0x1c);
        *(u16 *)(dst + 4) = *(u16 *)(src + 0x1e);
        *(u16 *)(dst + 6) = *(u16 *)(src + 0x20);
        *(u16 *)(dst + 8) = *(u16 *)(src + 0x24);
        dst[10] = src[1];                /* 等級 / 名稱索引 */
        /* 另複製 +0x2e/+0x2f 兩 byte(狀態)*/

        /* 掃 +0x3a 起 8 格裝備:0x4d→睡眠(or 0x800);0x11 且 +1&0x80→中毒(0x4000)*/
        {
            u8 *s = src + 0x3a;
            u16 sleep = 0, poison = 0;
            int k;
            for (k = 0; k < 8; k++, s += 2) {
                if (s[0] == 0x4d) sleep  = 0x800;
                if (s[0] == 0x11 && (s[1] & 0x80)) poison = 0x4000;
            }
            *(u16 *)(src + 0x38) |= sleep;
            *(u16 *)(src + 0x22) = (*(u16 *)(src + 0x22) & 0xbfff) | poison;
        }
    }
}

/* ===========================================================================
 * 回合 / 指令
 *
 * party_alive_count(sub_c5f6):掃 [0x5077] 名隊員(指標表 [bx+0x4f15]),
 * 對未設 0x80 旗標者,若 [di+0x16](HP)!=0 → 計數 bp++;並設 0x80 旗標。
 * 回傳 bp = 存活人數(0 → 全滅,主流程結束)。
 *
 * battle_command_loop(sub_c08b):本回合逐角色下指令。
 *   - call 0xbf7f(前置)→ call 0xc169(開「戰/逃/防/道具」指令選單,二級派發)。
 *   - [0x2517]==2 → 逃跑成功,跳結算([0x2517]:0=續行 / 1=防禦或無事 / 2=逃跑)。
 *   - 逐角色 loop([0x2515] 計數):畫指令游標、依選擇 call 對應動作。
 *   - 行動結束依 [0x24b4](我方還有人)/[0x24b3](敵方還有)決定續行或結算:
 *       敵方歸 0 → di=0x161 訊息(勝利)+ 給經驗 / 金錢(sub_c954 / sub_c425)。
 *       我方歸 0 → di=0x169 訊息(全滅)+ 等鍵。
 *
 * battle_run_away(sub_c7d9):逃跑收尾——還原戰前地圖座標 [0x4f48..0x4f52] 回
 * [0x4f2f..0x256c],重設 [0x4f2d]=1、重繪、退出戰鬥([0x2518]|=8)。
 * ======================================================================== */
int party_alive_count(void)
{
    extern u16 party_ent_tab[]; /* DS:0x4f15 起,每名實體指標 */
    int cl = g_party_size;      /* [0x5077] */
    int bp = 0, i;

    for (i = 0; i < cl; i++) {
        u8 *di = (u8 *)(uintptr_t)party_ent_tab[i]; /* [bx+0x4f15] */
        if (*(u16 *)(di + 0x38) & 0x80)
            continue;                                /* 已計過 */
        if (*(u16 *)(di + 0x16) != 0)                /* HP!=0 → 存活 */
            bp++;
        *(u16 *)(di + 0x38) |= 0x80;                 /* 設已處理旗標 */
    }
    return bp;
}

void battle_command_loop(void)
{
    extern u8 g_turn_done;     /* DS:0x2507 */
    extern u8 g_cmd_sel[];     /* DS:0x0cf6 起,各角色指令選擇暫存 */
    extern u8 g_has_party, g_has_enemy; /* DS:0x24b4 / 0x24b3 */
    u16 cx;

restart:
    g_escape_flag = 0;         /* [0x2517]=0 */
    g_turn_done   = 0;         /* [0x2507]=0 */
    /* [0x728]=0 */
    /* call 0xbf7f 前置 */
    /* call 0xc169:開指令選單(戰/逃/防/道具),結果寫 [0x2517] / 各角色選擇 */

    if (g_escape_flag == 2)    /* 逃跑成功 → 結算 */
        goto resolve;

    /* 逐角色執行所選指令(loop [0x2515] = g_cmd_count)*/
    for (cx = g_cmd_count; cx > 0; cx--) {
        /* 每角色:畫指令游標、依 g_cmd_sel[] 分派攻擊/防禦/道具/逃 */
        if (g_turn_done == 1) goto resolve;
        if (g_escape_flag == 2) goto resolve;

        if (g_has_party != 0) {
            if (g_has_enemy == 0) {
                /* === 敵全滅:勝利 === */
                file_op_dx(0x240);              /* bp=0x22 */
                txt_draw(0x161);                /* di=0x161:勝利訊息 */
                /* call 0xc954(結算)+ call 0xc425(經驗 / 金錢)*/
                return;
            }
            /* 敵仍在 → 重開下一輪指令 */
            goto restart;
        }
    }

    /* 我方全滅分支:di=0x169 全滅訊息(隊伍人數==1 例外)+ 等鍵 */
    if (g_party_size != 1)
        txt_draw(0x169);
    kbd_set_mode(0x1104 >> 8, 0xdb); /* lcall 1104:db 等鍵(示意)*/
    return;

resolve:
    /* call 0xc954 結算 + 收尾繪製 */
    return;
}

void battle_redraw_field(void) { /* sub_bcf2:狀態列 / 場景重繪(細節見 docs)*/ }

void battle_run_away(void)
{
    /* 還原戰前地圖座標(戰鬥開始前備份在 [0x4f48..0x4f52])*/
    g_scene_flag = SCENE_TOWN;        /* [0x4f2d]=1 */
    /* [0x4f2f]=[0x4f48];[0x4f31]=[0x4f4a];[0x256a]=[0x4f50];[0x256c]=[0x4f52] */
    /* [0x4f33]=[0x4f4c];[0x4f35]=[0x4f4e];清本幀位移 [0x2572]/[0x2574];[0x4f1f]=0xff */
    battle_redraw_field();            /* call 0xbcf2 */
    g_battle_state |= 8;              /* [0x2518]|=8:標示「逃跑離場」*/
    mouse_op(0x3a0);                  /* lcall 10b6:3a0 */
}
