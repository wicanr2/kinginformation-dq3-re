#!/usr/bin/env python3
"""抽取 DQ3.EXE DGROUP 靜態資料表 → 編入 RE/remake(讓 remake 不再依賴 DQ3.EXE)。

我們的反組譯原本只 RE 了「碼」(re/*.c),資料表只記了位址。本工具把這些表的「值」
一次抽出,產出:
  - docs/data/dgroup-tables.md          資料表目錄(offset/size/語意 + 值摘要)
  - dq3_remake/src/dq3_exedata.c/.h      內建 C 資料(remake 直接用,不讀 DQ3.EXE)

位址:HDR=0x1370、DGROUP seg=0x14dd → file base 0x16140;file = 0x16140 + DS_off。
"""
import struct, os, sys

EXE = "assets_raw/DQ3.EXE"
BASE = 0x16140
d = open(EXE, "rb").read()
def fo(ds): return BASE + ds
def u8(o): return d[o]
def u16(o): return struct.unpack_from("<H", d, o)[0]
def u32(o): return struct.unpack_from("<I", d, o)[0]

# ---- 表定義(DS off, 語意) ----
GROWTH_DS = 0x4366          # 8 職業 × 14 byte 成長表
THR_PTR_DS = 0x43d6         # 8 × u16 指標 → 各 44 entry u32 升級門檻
TERRAIN_DS = 0x4df6         # 256 byte tile→地形類別(戰鬥背景選頁)
EVENT_DS = 0x37c4           # tile→物件事件(3 byte/entry;確認/對話用)

NUM_CLASS = 8
GROWTH_ROW = 14
MAX_LEVEL = 43              # 門檻 lv0..43 = 44 entry
EVENT_N = 256              # 事件表 entry 數(以 tile index 為鍵)
TERRAIN_N = 256

def extract():
    growth = [list(d[fo(GROWTH_DS)+c*GROWTH_ROW: fo(GROWTH_DS)+c*GROWTH_ROW+GROWTH_ROW]) for c in range(NUM_CLASS)]
    # 門檻:指標表 → 各表
    ptrs = [u16(fo(THR_PTR_DS)+i*2) for i in range(NUM_CLASS)]
    thresh = [[u32(fo(p)+lv*4) for lv in range(MAX_LEVEL+1)] for p in ptrs]
    terrain = list(d[fo(TERRAIN_DS): fo(TERRAIN_DS)+TERRAIN_N])
    event = [list(d[fo(EVENT_DS)+t*3: fo(EVENT_DS)+t*3+3]) for t in range(EVENT_N)]
    return growth, ptrs, thresh, terrain, event

def gen_c(growth, thresh, terrain, event):
    os.makedirs("dq3_remake/src", exist_ok=True)
    h = []
    h.append("/* dq3_exedata.h — 自 DQ3.EXE DGROUP 抽出的靜態資料表(生成檔,勿手改)。")
    h.append(" * 由 tools/extract_dgroup_tables.py 產生;remake 改用內建表,不再讀 DQ3.EXE。")
    h.append(" * 目錄與語意見 docs/data/dgroup-tables.md。 */")
    h.append("#ifndef DQ3_EXEDATA_H\n#define DQ3_EXEDATA_H\n#include <stdint.h>\n")
    h.append("#define DQ3X_NUM_CLASS 8")
    h.append("#define DQ3X_GROWTH_ROW 14")
    h.append("#define DQ3X_MAX_LEVEL 43")
    h.append("extern const uint8_t  dq3x_growth[DQ3X_NUM_CLASS][DQ3X_GROWTH_ROW];")
    h.append("extern const uint32_t dq3x_thresh[DQ3X_NUM_CLASS][DQ3X_MAX_LEVEL+1];")
    h.append("extern const uint8_t  dq3x_terrain[256];   /* tile→地形類別 */")
    h.append("extern const uint8_t  dq3x_event[256][3];  /* tile→物件事件(3B) */")
    h.append("#endif")
    open("dq3_remake/include/dq3_exedata.h","w").write("\n".join(h)+"\n")

    c = []
    c.append("/* dq3_exedata.c — 自 DQ3.EXE DGROUP 抽出的靜態資料表(生成檔,勿手改)。 */")
    c.append('#include "dq3_exedata.h"\n')
    c.append("const uint8_t dq3x_growth[DQ3X_NUM_CLASS][DQ3X_GROWTH_ROW] = {")
    for row in growth:
        c.append("  {" + ",".join("0x%02x"%b for b in row) + "},")
    c.append("};\n")
    c.append("const uint32_t dq3x_thresh[DQ3X_NUM_CLASS][DQ3X_MAX_LEVEL+1] = {")
    for t in thresh:
        c.append("  {" + ",".join(str(v) for v in t) + "},")
    c.append("};\n")
    c.append("const uint8_t dq3x_terrain[256] = {")
    for i in range(0,256,16):
        c.append("  " + ",".join("0x%02x"%b for b in terrain[i:i+16]) + ",")
    c.append("};\n")
    c.append("const uint8_t dq3x_event[256][3] = {")
    for e in event:
        c.append("  {0x%02x,0x%02x,0x%02x},"%(e[0],e[1],e[2]))
    c.append("};")
    open("dq3_remake/src/dq3_exedata.c","w").write("\n".join(c)+"\n")

def gen_doc(growth, ptrs, thresh, terrain, event):
    names = "勇者 戰士 武鬥家 僧侶 魔法使者 賢者 商人 遊玩者".split()
    L = []
    L.append("# DGROUP 靜態資料表目錄(DQ3.EXE)\n")
    L.append("我們的反組譯先前只 RE 了**碼**(`re/*.c`),資料表只記位址。本目錄把 DGROUP 的")
    L.append("**靜態資料表一次抽出編目**(offset/size/語意/值),由 `tools/extract_dgroup_tables.py`")
    L.append("生成內建 C(`dq3_remake/src/dq3_exedata.c`),**remake 不再執行期讀 DQ3.EXE**。\n")
    L.append("> 位址慣例:`file = 0x16140 + DS_off`(HDR 0x1370 + DGROUP seg 0x14dd×16)。\n")
    L.append("| 表 | DS off | file off | size | 語意 | 用途 |")
    L.append("|---|---|---|---|---|---|")
    L.append("| 成長表 | 0x4366 | 0x%05x | 8×14 | 每職業 HP/MP/… base+slope | dq3_stats(#4 MP / 成長)|"%fo(GROWTH_DS))
    L.append("| 升級門檻指標 | 0x43d6 | 0x%05x | 8×u16 | → 各職業 44×u32 累積經驗 | dq3_stats(#5 clamp)|"%fo(THR_PTR_DS))
    L.append("| 地形表 | 0x4df6 | 0x%05x | 256 | tile→地形類別(0/1/3…) | 戰鬥背景選頁(packbg)|"%fo(TERRAIN_DS))
    L.append("| 物件事件表 | 0x37c4 | 0x%05x | 256×3 | tile→事件碼(0xff=無)| 確認/對話/觸發(sub_9cd6)|"%fo(EVENT_DS))
    L.append("")
    L.append("## 成長表(0x4366,8 職業 × 14 byte)")
    L.append("列格式:`HPb HPs MPb MPs b4b b4s b6b b6s b8b b8s bAb bAs c d`(docs/23)")
    for i,row in enumerate(growth):
        L.append("- %d %s: `%s`"%(i,names[i]," ".join("%02x"%b for b in row)))
    L.append("")
    L.append("## 升級門檻(0x43d6 指標 → 各 44 entry)")
    L.append("指標(DS off):"+", ".join("0x%04x"%p for p in ptrs)+";每職業 44 entry(lv0..43)。")
    L.append("- 勇者 lv1/10/43 門檻 = %d / %d / %d(範例)"%(thresh[0][1],thresh[0][10],thresh[0][43]))
    L.append("")
    L.append("## 地形表(0x4df6,tile→地形)")
    nz = [(t,v) for t,v in enumerate(terrain[:48])]
    L.append("前 48 tile:`"+" ".join("%d"%v for t,v in nz)+"`(地形類別,戰鬥背景 page=[0xd73]+terrain*8,docs/13)")
    L.append("")
    L.append("## 物件事件表(0x37c4,tile→事件 3 byte)")
    L.append("每 entry 3 byte;`sub_9cd6` 以前方 tile 查 `[tile*3+0x37c4]`,事件碼決定對話/觸發。")
    L.append("非空範例:"+", ".join("tile%d=%02x%02x%02x"%(t,event[t][0],event[t][1],event[t][2]) for t in range(6)))
    L.append("")
    L.append("## 再生方式")
    L.append("```\ntools/dockrun.sh tools/extract_dgroup_tables.py\n```")
    L.append("→ 重生 `dq3_remake/{include,src}/dq3_exedata.*` 與本目錄。版權:數值平衡表依 decomp")
    L.append("慣例納入可重編譯來源;大型創作素材(sprite/劇本/音樂)仍為外部 asset。\n")
    os.makedirs("docs/data", exist_ok=True)
    open("docs/data/dgroup-tables.md","w").write("\n".join(L))

def main():
    growth, ptrs, thresh, terrain, event = extract()
    gen_c(growth, thresh, terrain, event)
    gen_doc(growth, ptrs, thresh, terrain, event)
    print("生成 dq3_remake/src/dq3_exedata.c + include/dq3_exedata.h + docs/data/dgroup-tables.md")
    print("成長表 勇者列:", " ".join("%02x"%b for b in growth[0]))
    print("地形表 tile0..7:", terrain[:8])
    print("事件表 tile0:", event[0])

if __name__ == "__main__":
    main()
