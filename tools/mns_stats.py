#!/usr/bin/env python3
"""解析 D3MNS.DAT 怪物數值表 → docs/data/d3mns_stats.json + 對照表。

記錄 = 130 筆 × 0x29(41) byte,index = monster_id。欄位偏移由 DQ3.EXE 反組譯確認:
  +0x00 u16 hp_base   (sub @0xab43: [rec+0] + rnd([rec+2]) → 當前 HP)
  +0x02 u16 hp_rand   (HP 隨機加成範圍)
  +0x04 u16 f04       (→ enemy[+0x10];推定 攻擊力 / atk)
  +0x08 u8  f08       (→ enemy[+0xb])
  +0x09 u16 f09       (→ enemy[+0xc])
  +0x0b u16 f0b       (→ enemy[+0xe])
  +0x1f u8  f1f       (→ enemy[+0x14])
  +0x21 u16 exp       (sub @0xbce4: 加到隊伍經驗總和 [0x24f6])
  +0x23 u16 gold      (sub @0xbcdc: 加到金錢總和 [0x24fa])
  +0x25 u8  f25       (sub @0xc51e)
  +0x26 u8  f26       (sub @0xc529)
  +0x28 u8  spawn_weight (遭遇點數預算分配;sub @0xab76 ×2 → [0x2513])
其餘 +0x06/+0x0d..+0x1e/+0x20/+0x22/+0x24/+0x27 語意未定,原樣保留。
"""
import struct, json, re

DAT = "assets_raw/D3MNS.DAT"
REC = 0x29
N = 130
NAME_BASE = 0x258   # 怪名 = D3TXT00 記錄 0x258 + monster_id(sprite_id)


def u16(r, o):
    return struct.unpack_from('<H', r, o)[0]


def load_names():
    """從已解碼的 docs/script/txt00.txt 取怪名(記錄 0x258 + id)。"""
    try:
        txt = open("docs/script/txt00.txt", encoding="utf-8").read()
    except FileNotFoundError:
        return {}
    recs = {}
    cur = None; buf = []
    for line in txt.splitlines():
        mm = re.match(r'^\[(\d+)\]\s?(.*)', line)
        if mm:
            if cur is not None:
                recs[cur] = "\n".join(buf).strip()
            cur = int(mm.group(1)); buf = [mm.group(2)]
        elif cur is not None:
            buf.append(line)
    if cur is not None:
        recs[cur] = "\n".join(buf).strip()
    return {i: recs.get(NAME_BASE + i, "") for i in range(N)}


def main():
    m = open(DAT, 'rb').read()
    names = load_names()
    rows = []
    for i in range(N):
        r = m[i*REC:(i+1)*REC]
        rows.append({
            "id": i,
            "name": names.get(i, ""),
            "hp_base": u16(r, 0x00),
            "hp_rand": u16(r, 0x02),
            "f04_atk?": u16(r, 0x04),
            "f08": r[0x08],
            "f09": u16(r, 0x09),
            "f0b": u16(r, 0x0b),
            "f1f": r[0x1f],
            "exp": u16(r, 0x21),
            "gold": u16(r, 0x23),
            "f25": r[0x25],
            "f26": r[0x26],
            "spawn_weight": r[0x28],
            "raw_hex": r.hex(),
        })
    with open("docs/data/d3mns_stats.json", "w") as f:
        json.dump(rows, f, ensure_ascii=False, indent=1)
    # 印表
    print(f"{'id':>3} {'hp':>4} {'hpR':>3} {'f04':>5} {'f08':>3} {'f09':>5} {'f0b':>5} {'exp':>6} {'gold':>4} {'wt':>3}  name")
    for x in rows:
        print(f"{x['id']:3d} {x['hp_base']:4d} {x['hp_rand']:3d} {x['f04_atk?']:5d} {x['f08']:3d} {x['f09']:5d} {x['f0b']:5d} {x['exp']:6d} {x['gold']:4d} {x['spawn_weight']:3d}  {x['name']}")


if __name__ == "__main__":
    main()
