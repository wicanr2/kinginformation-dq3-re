/* loaders.c — DQ3.EXE 素材載入函式反編譯(RE → C)
 *
 * 對應 DQ3.EXE seg0 各函式;位址見 docs/09-exe-loaders.md。
 * 風格:忠實鏡射反組譯流程,以具名變數/結構取代裸 DS:offset。
 * DOS 中斷以下方 shim 抽象(dos_open / dos_read / dos_write / dos_close /
 * dos_lseek):原程式為 real-mode `int 21h`,移植時換成 fopen/fread 等。
 *
 * 反編譯來源(file off → seg0):
 *   load_blk        0xfe64 (eaf4)   load_blk_page sub 0xff0e (eb9e)
 *   load_cty        0x43d1 (3061)   d3txt fon/txt 0x2b81/0x2ba7 (1811/1837)
 *   load_d3txt_map  0x2bdc (186c)   load_setup 0xaced (997d)
 *   load_player     0x2ae8 (1778)   save_player 0x2b1e (17ae)
 *   load_item       0x2b3b (17cb)   load_dragon 0x28ce (155e)
 *   open_world_maps 0x3f89 (2c19)   load_bls/blkbm/pal 0x1000e.. (eca0..)
 *   load_mns_shp    0xc50e (b19e)   load_packbg_page 0xda55 (c6e5)
 */
#include "assets.h"

/* ===== DOS int 21h shim(real-mode → 主機檔案 I/O)===== */
/* 回傳值約定鏡射 real-mode:open 失敗回 -1(原碼以 CF=1/jb 判斷)。      */
extern int  dos_open(const char *name, int mode);          /* AH=3Dh */
extern int  dos_creat(const char *name, int attr);         /* AH=3Ch */
extern int  dos_read(int h, void *buf, unsigned n);        /* AH=3Fh; n=0xffff 讀到 EOF */
extern int  dos_read_far(int h, uint16_t seg, uint16_t off, unsigned n); /* DS=seg 版 */
extern int  dos_write(int h, const void *buf, unsigned n); /* AH=40h */
extern void dos_close(int h);                              /* AH=3Eh */
extern long dos_lseek(int h, long pos, int whence);        /* AH=42h */

/* VGA planar 暫存器寫(out dx,ax);packbg 用)。runtime 段 0x111b 的對應。*/
extern void vga_out(uint16_t port, uint16_t val);

/* 主資料段內的素材緩衝(以具名陣列代替 DS:offset)。*/
extern uint8_t  blk_header_raw[6];   /* DS:0x2859 */
extern uint8_t  item_buf[];          /* DS:0x1f9  */
extern uint8_t  blkbm_buf[];         /* DS:0x2eea */
extern uint8_t  pal_buf[240];        /* DS:0x3232 dq3.pal */
extern uint8_t  mnsbk_pal_buf[480];  /* DS:0x3352 mnsbk.pal */
extern uint8_t  setup_hdr[6];        /* DS:0x63a  setup.dat 6B 表頭 */
extern uint8_t  save_buf[];          /* DS:0x4f29 dragon save 緩衝 */
extern uint8_t  player_buf[200];     /* DS:0x128  player.dat */
extern uint16_t setup_flags;         /* DS:0x000f setup 旗標位元 */
extern uint16_t mns_chunk;           /* DS:0x286b/0x2867/0x2869 setup 衍生 */
extern const uint8_t map_blknum[];   /* DS:0xa04 起:map_idx → BLK 檔號 */

/* ================================================================= */
/* BLK tile 圖庫載入(seg0:eaf4 / file 0xfe64)
 *
 * 流程:
 *   1. is_town_flag 決定開地表 dq3.blk 或城鎮 dq3?.blk(? = map_blknum 查號)。
 *   2. 讀 6-byte 表頭 → blk_header_raw(印證 docs/04 的 BLK 表頭 = 6B)。
 *   3. 把其後全部 tile body 讀進遠段 seg_blk(CX=0xffff 即讀到 EOF)。
 *   4. is_town_flag!=1 時,另載入兩段固定 BLKBM 區塊(load_blk_page,索引 0x12/0x13)。
 *   5. is_town_flag!=0(城鎮)時,改寫並開 blkbmN.dat,讀進 DS:0x308e。
 * 回傳 0 成功 / 0xff(-1)開檔失敗。
 */
static int load_blk_page(uint16_t si, uint16_t bp, uint16_t ax); /* sub 0xff0e */

int load_blk(void)
{
    int h;
    const char *name;

    if (is_town_flag != 0) {
        /* 城鎮:dq3?.blk,? = (map_blknum[cur_map_idx] - 1) + '1' */
        uint8_t blknum = map_blknum[cur_map_idx];
        fn_blk_town[3] = (char)((blknum - 1) + '1');  /* DS:0xa2 = fn[3] */
        name = fn_blk_town;
    } else {
        name = fn_blk_over;                            /* "dq3.blk" */
    }

    h = dos_open(name, 0);
    if (h < 0)
        return -1;                                     /* mov al,0xff; ret */

    /* 6-byte 表頭 → blk_header_raw(DS:0x2859)*/
    dos_read(h, blk_header_raw, 6);

    /* tile body → 遠段 seg_blk,從目前檔位置讀到 EOF */
    dos_read_far(h, seg_blk, 0, 0xffff);
    dos_close(h);

    if (is_town_flag != 1) {
        /* 地表/部分城鎮:載入兩塊固定 BLKBM page(原碼用 [0x2665]=0 標記)*/
        load_blk_page(0x3c00, 0, 0x12);
        load_blk_page(0x3fc0, 0, 0x13);
    }

    if (is_town_flag != 0) {
        /* blkbmN.dat:N = (map_blknum[cur_map_idx]-1)+'1',改寫 DS:0xb5 = fn[5] */
        uint8_t blknum = map_blknum[cur_map_idx];
        fn_blkbm1[5] = (char)((blknum - 1) + '1');
        h = dos_open(fn_blkbm1, 0);                    /* 原碼未檢失敗 */
        dos_read_far(h, /*DS:0x308e 所在段*/ seg_blk, 0x308e, 0xffff);
        dos_close(h);
    }
    return 0;
}

/* sub 0xff0e:把 .bls 內第 (si 行起、page=ax) 的固定大小區塊讀進 seg_blspage:bp。
 * 偏移計算:base = 6 + (0x3c0 * bp) ;file_pos = ax*0xf00 + base(32-bit）。
 * 讀 0x3c0(960)bytes 進 DS=seg_blspage:SI。handle 取自 h_bls_man(DS:0x260b)。
 */
static int load_blk_page(uint16_t si, uint16_t bp, uint16_t ax)
{
    uint32_t base = (uint32_t)0x3c0 * (uint8_t)bp + 6;   /* +6 跳表頭 */
    uint32_t pos  = (uint32_t)ax * 0x0f00 + base;
    dos_lseek(h_bls_man, (long)pos, 0 /*SEEK_SET*/);
    dos_read_far(h_bls_man, seg_blspage, si, 0x3c0);
    return 0;
}

/* ================================================================= */
/* CTY 城鎮場景載入(seg0:3061 / file 0x43d1)
 *
 * 入口先掃一張 (target_tile-1 == [si]) 的查找表決定 map 索引 bp(略,見 0x43b7),
 * 命中後:設 is_town_flag=1,cur_map_idx=bp;把 bp 拆十進位改寫 cty 檔名數字
 * (DS:0x61/0x62 = fn[3]/fn[4]);開 ctyNN.dat,整檔讀進遠段 seg_cty;
 * 關檔後立刻呼叫 load_blk()(城鎮 tile 圖庫隨城鎮切換重載)。
 */
int load_cty(uint16_t map_idx)
{
    int h;
    is_town_flag = 1;
    cur_map_idx  = map_idx;

    /* bp(map_idx)→ 兩位十進位 ASCII,寫進 "cty00" 的 "00" */
    fn_cty[3] = (char)('0' + map_idx / 10);   /* DS:0x61 = fn[3] */
    fn_cty[4] = (char)('0' + map_idx % 10);   /* DS:0x62 = fn[4] */

    h = dos_open(fn_cty, 0);
    if (h < 0)
        return -1;

    dos_read_far(h, seg_cty, 0, 0xffff);      /* 整檔 → seg_cty */
    dos_close(h);

    load_blk();                               /* call 0xfe64 */
    return 0;
}

/* ================================================================= */
/* 遊戲文字字型/腳本(seg0:1811 / file 0x2b81)
 *
 * load_d3txt_font:d3txt00.fon 整檔 → seg_fon(DS:0x252c);
 * 緊接 load_d3txt_text:d3txt00.txt 整檔 → seg_txt(DS:0x252e)。
 * 原碼為同一常式依序載入兩檔(共用 DOS 開→讀→關)。
 */
int load_d3txt_font(void)
{
    int h = dos_open(fn_d3txt_fon, 0);
    if (h < 0) return -1;
    dos_read_far(h, seg_fon, 0, 0xffff);
    dos_close(h);
    return 0;
}

int load_d3txt_text(void)
{
    int h = dos_open(fn_d3txt_txt, 0);
    if (h < 0) return -1;
    dos_read_far(h, seg_txt, 0, 0xffff);
    dos_close(h);
    return 0;
}

/* 切換對話腳本(seg0:186c / file 0x2bdc)
 * 載入 d3txtNN.txt(NN = map number n,改寫 DS:0x79/0x7a = fn[5]/fn[6]),
 * 讀進 seg_txt:0x4e20(與 d3txt00 共段、偏移 0x4e20 分區)。
 * [0x2528] 為快取的目前 NN,相同則直接 return(避免重載)。
 */
int load_d3txt_map(uint8_t n)
{
    int h;
    extern uint16_t d3txt_cur_n;             /* DS:0x2528 */
    if (d3txt_cur_n == n) return 0;
    d3txt_cur_n = n;

    fn_d3txt_txt[5] = (char)('0' + n / 10);  /* DS:0x79 = fn[5] */
    fn_d3txt_txt[6] = (char)('0' + n % 10);  /* DS:0x7a = fn[6] */

    h = dos_open(fn_d3txt_txt, 0);
    if (h < 0) return -1;
    dos_read_far(h, seg_txt, 0x4e20, 0xffff);
    dos_close(h);
    return 0;
}

/* ================================================================= */
/* item.dat(seg0:17cb / file 0x2b3b):整檔 → DS:0x1f9。失敗回 -1。 */
int load_item(void)
{
    int h = dos_open(fn_item, 0);
    if (h < 0) return -1;
    dos_read(h, item_buf, 0xffff);
    dos_close(h);
    return 0;
}

/* ================================================================= */
/* player.dat 存檔(seg0:1778 / file 0x2ae8)
 *
 * 讀 200B(0xc8)save 區 → DS:0x128。開檔失敗 = 新遊戲:把 200B 清零並回 1。
 * 成功回 0。save_player 為對應建檔寫回(AH=3Ch creat → AH=40h write)。
 */
int load_player(void)
{
    int h = dos_open(fn_player, 0);
    if (h < 0) {
        /* rep stosb 0,200 → 清空存檔(新遊戲),回 1 */
        for (int i = 0; i < 0xc8; i++) player_buf[i] = 0;
        return 1;
    }
    dos_read(h, player_buf, 0xc8);
    dos_close(h);
    return 0;
}

void save_player(void)
{
    int h = dos_creat(fn_player, 0);          /* AH=3Ch, CX=0 一般屬性 */
    dos_write(h, player_buf, 0xc8);
    dos_close(h);
}

/* ================================================================= */
/* dragonN.dat(seg0:155e / file 0x28ce)
 *
 * N = ([0x722] - 1) + '0',改寫 DS:0x58 = fn[6]("dragon0" 的 '0')。
 * 讀 CX = (0x57a5 - 0x4f29) bytes(save 結構大小)→ DS:0x4f29。
 */
int load_dragon(void)
{
    extern uint16_t dragon_slot;              /* DS:0x722 */
    int h;
    unsigned sz = 0x57a5 - 0x4f29;            /* save 緩衝大小 */

    fn_dragon[6] = (char)('0' + (dragon_slot - 1)); /* DS:0x58 = fn[6] */

    h = dos_open(fn_dragon, 0);
    if (h < 0) return -1;
    dos_read(h, save_buf, sz);
    dos_close(h);
    return 0;
}

/* ================================================================= */
/* 世界地圖(seg0:2c19 / file 0x3f89)
 *
 * 只「開檔並快取 handle」,不一次讀入(地圖很大,後續按頁 seek 讀)。
 * dq3con.map → h_con_map(DS:0x1f1);dq3und.map → h_und_map(DS:0x1f3)。
 * [0x1f1]!=0 表已開啟,直接 return。
 */
void open_world_maps(void)
{
    if (h_con_map != 0) return;
    h_con_map = (uint16_t)dos_open(fn_con_map, 0);
    h_und_map = (uint16_t)dos_open(fn_und_map, 0);
}

/* ================================================================= */
/* .bls 聲音/精靈分頁庫(seg0:eca0 / file 0x1000e)
 *
 * 只開檔快取 handle(分頁 seek 讀,見 load_blk_page / 0xfff2)。
 * dq3mst.bls → h_bls_mst(DS:0x260b);dq3man.bls → h_bls_man(DS:0x2609)。
 * 任一開檔失敗回 -1。
 */
int load_bls_banks(void)
{
    int h;
    h = dos_open(fn_bls_mst, 0);
    if (h < 0) return -1;
    h_bls_mst = (uint16_t)h;
    h = dos_open(fn_bls_man, 0);
    if (h < 0) return -1;
    h_bls_man = (uint16_t)h;
    return 0;
}

/* blkbm.dat(seg0:ecc0 / file 0x10030):整檔 → DS:0x2eea。 */
int load_blkbm(void)
{
    int h = dos_open(fn_blkbm, 0);
    if (h < 0) return -1;          /* 注意:原碼此處未檢 CF,handle 直接用 */
    dos_read(h, blkbm_buf, 0xffff);
    dos_close(h);
    return 0;
}

/* ================================================================= */
/* 調色盤(seg0:ecdc / file 0x1004c)
 *
 * 先 lcall 0x109c:0xc4(BIOS 視訊準備),再:
 *   dq3.pal   → 240B(0xf0) DS:0x3232,緩衝指標存 [0x25d1]。
 *   mnsbk.pal → 480B(0x1e0) DS:0x3352。
 * 之後把 [0x25d5]=0、call 0x102ce(套用 palette → DAC),lcall 0x109c:0x8d/0xa7。
 * (docs/04:dq3.pal=80色×3B=240;mnsbk.pal=160色×3B=480,印證讀取長度。)
 */
int load_palettes(void)
{
    int h;
    /* vga_prepare();  // lcall 0x109c:0xc4 */

    h = dos_open(fn_pal, 0);
    if (h < 0) return -1;
    dos_read(h, pal_buf, 0xf0);    /* 240 bytes = 80 色 */
    dos_close(h);

    h = dos_open(fn_mnsbk_pal, 0);
    if (h < 0) return -1;
    dos_read(h, mnsbk_pal_buf, 0x1e0); /* 480 bytes = 160 色 */
    dos_close(h);

    /* palette_apply_flag=0; apply_palette_to_dac();  // call 0x102ce */
    /* vga_set_mode(); vga_unblank();                  // lcall 0x109c:0x8d / 0xa7 */
    return 0;
}

/* ================================================================= */
/* dq3mns.shp 怪物 sprite 表頭(seg0:b19e / file 0xc50e)
 *
 * h_d3mns([0x249b])為 0 才開檔(只開一次、常駐 handle)。dq3mns.shp 為
 * 分頁 sprite 庫,後續按 sprite 編號 seek 讀(此函式只負責確保檔已開啟)。
 */
int load_mns_shp(void)
{
    if (h_d3mns == 0) {
        int h = dos_open(fn_mns_shp, 0);
        h_d3mns = (uint16_t)h;
    }
    return 0;
}

/* ================================================================= */
/* packbg.scr 全螢幕封裝背景(seg0:c6e5 / file 0xda55)
 *
 * 分頁 planar 背景:page 大小 0x3d80(15744)B;file_pos = page*0x3d80 + page。
 * (原碼 mul di by 0x3d80 後 add dx,di:即 *0x3d81?——保留為 0x3d80*page+page。)
 * 開檔一次快取 h_packbg(DS:0x1f5)。seek 到該頁後:
 *   設定 VGA Graphics Controller(port 0x3ce):mode/bitmask;
 *   讀 0x6e00 bytes 進遠段 seg_packbg:0(由 [0x4f09] 指目的 DI);
 *   再透過 0x3c4 Sequencer Map Mask 逐 plane 解交錯(後段略)。
 * 印證:packbg 為 4-bit planar 全螢幕圖,與 BLK/SHP 同 planar 家族(docs/04)。
 */
void load_packbg_page(uint16_t page)
{
    long pos;
    if (h_packbg == 0)
        h_packbg = (uint16_t)dos_open(fn_packbg, 0);

    pos = (long)page * 0x3d80 + page;        /* mul di,0x3d80; add dx,di */
    dos_lseek(h_packbg, pos, 0);

    /* VGA GC:read mode / set/reset / bit mask = 0xff(整 byte)*/
    vga_out(0x3ce, 0x0003);                  /* GC index3 = 0(data rotate)*/
    vga_out(0x3ce, 0x0805);                  /* GC mode */
    vga_out(0x3ce, 0x0007);                  /* GC color don't care */
    vga_out(0x3ce, 0xff08);                  /* GC bit mask = 0xff */

    /* 讀 0x6e00 bytes 進 seg_packbg(目的 DI=[0x4f09])*/
    dos_read_far(h_packbg, seg_packbg, /*off=[0x4f09]*/ 0, 0x6e00);

    /* 後續 0x3c4 Sequencer Map Mask 逐 plane 寫 VRAM(解交錯)— 見 0xdab0 起 */
}

/* ================================================================= */
/* setup.dat 設定(seg0:997d / file 0xaced)
 *
 * 讀 6-byte 表頭 → setup_hdr(DS:0x63a),依 byte[0] 設旗標:
 *   ==1 → setup_flags |= 2 ;否則 |= 4 並:
 *         [0x286b]=word[0x63b], [0x2867]=byte[0x63d], [0x2869]=byte[0x63e]。
 *   byte[5](0x63f)!=0 → setup_flags |= 0x20。
 * 開檔失敗:[0x286f]=0x80(預設值),return。
 */
int load_setup(void)
{
    int h = dos_open(fn_setup, 0);
    if (h < 0) {
        extern uint16_t setup_default;       /* DS:0x286f */
        setup_default = 0x80;
        return -1;
    }
    dos_read(h, setup_hdr, 6);
    dos_close(h);

    if (setup_hdr[0] == 1) {
        setup_flags |= 2;
    } else if (setup_hdr[0] != 0) {
        extern uint16_t setup_w0, setup_b3, setup_b4; /* 0x286b/0x2867/0x2869 */
        setup_flags |= 4;
        setup_w0 = *(uint16_t *)&setup_hdr[1];  /* word[0x63b] */
        setup_b3 = setup_hdr[3];                /* byte[0x63d] */
        setup_b4 = setup_hdr[4];                /* byte[0x63e] */
    }
    if (setup_hdr[5] != 0)                       /* byte[0x63f] */
        setup_flags |= 0x20;
    return 0;
}
