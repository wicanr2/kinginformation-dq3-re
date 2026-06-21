#!/usr/bin/env python3
"""CHINA.FON run-based 字模對齊核心。

關鍵事實(已用程式驗證,見 docs/02-font-format.md):
  CHINA.FON 的字模多半以 32-byte 同相位連續排列;字模之間偶有「1-byte 插入」,
  使後續字模的起始 offset 奇偶相位翻轉(gap 由 32 變 33/65/97…)。先前誤判為
  「左右半對調混淆」,實為 1-byte 對齊漂移——deobf_rows(b)[y] 對 y>=1 完全等於
  rows16(b-1)[y](往前 1 byte 讀)。因此「去混淆」= 讀對相位,沒有真正的混淆。

對齊解法:run-based。維持一條「目前相位」的 run,每個 glyph 位置:
  - 候選 = {維持相位 b, 翻相位 b+1};
  - 用 validity() 評分,row15==0 且 row14 近空 的「乾淨底部留白」是強訊號;
  - 連續性偏好:同相位(維持 run)給予 bonus,只有翻相位明顯較佳才翻;
  - 對不上(索引表/雜湊區)就 +1 滑動找回 run。
"""
import struct

RAW = "assets_raw"


def load():
    data = open(f"{RAW}/CHINA.FON", "rb").read()
    offs = [struct.unpack_from('<I', data, i * 4)[0] for i in range(22)] + [len(data)]
    return data, offs


def rows16(blob, b):
    return [(blob[b + y * 2] << 8) | blob[b + y * 2 + 1] for y in range(16)]


def _pixels(rows):
    return [[(rows[y] >> (15 - x)) & 1 for x in range(16)] for y in range(16)]


def validity(rows):
    """字模『像不像一個對齊乾淨的正字』的分數,越高越好。

    組成:
      + 筆畫 4-連通性(正字筆畫連續)
      - 孤立雜點懲罰(錯位的字會散出孤立點)
      + 底部留白獎勵(row15 全空、row14 近空 = 對齊乾淨的最強訊號)
      - 底部溢墨重罰(row15 有墨 = 相位錯了,筆畫溢進行距留白)
      - 密度/變化異常輕罰(排除梳齒表格、過疏過密)
    """
    px = _pixels(rows)
    ink = sum(sum(r) for r in px)
    if ink == 0:
        return -50.0
    # 4-連通 edge 計數
    nb = 0
    iso = 0
    for y in range(16):
        for x in range(16):
            if not px[y][x]:
                continue
            c = 0
            if y + 1 < 16 and px[y + 1][x]:
                c += 1
            if y - 1 >= 0 and px[y - 1][x]:
                c += 1
            if x + 1 < 16 and px[y][x + 1]:
                c += 1
            if x - 1 >= 0 and px[y][x - 1]:
                c += 1
            nb += c
            if c == 0:
                iso += 1
    score = (nb / ink) - 1.6 * (iso / ink)

    r14 = bin(rows[14]).count('1')
    r15 = bin(rows[15]).count('1')
    # 底部留白:強訊號
    if r15 == 0:
        score += 0.9
    else:
        score -= 0.25 * r15            # 行距留白被筆畫溢入 = 相位錯
    if r14 == 0:
        score += 0.5
    elif r14 <= 2:
        score += 0.15
    else:
        score -= 0.10 * (r14 - 2)

    # 頂列也常為留白(字身 row0..13,但很多字頂端 1 列空)。輕微獎勵頂端乾淨。
    if rows[0] == 0:
        score += 0.10

    # 密度約束(字身 16x14)
    body = rows[:14]
    dens = sum(bin(x).count('1') for x in body) / (16 * 14)
    if dens < 0.05 or dens > 0.62:
        score -= 1.5
    # 垂直梳齒(整欄同值)= 表格/索引區
    colconst = sum(1 for c in range(16) if len({(rr >> (15 - c)) & 1 for rr in body}) == 1)
    if colconst > 8:
        score -= 1.0
    # 列多樣性過低 = 重複 pattern
    if len(set(body)) < 7:
        score -= 0.8
    return score


def is_glyph_strict(rows):
    """硬門檻:用來判斷一個候選 frame 是否『夠像字』可進 run。"""
    if rows[15] != 0:
        return False
    if bin(rows[14]).count('1') > 3:
        return False
    body = rows[:14]
    dens = sum(bin(x).count('1') for x in body) / (16 * 14)
    if not (0.06 <= dens <= 0.60):
        return False
    if len(set(body)) < 7:
        return False
    colconst = sum(1 for c in range(16) if len({(rr >> (15 - c)) & 1 for rr in body}) == 1)
    if colconst > 8:
        return False
    return True


def _best_phase_here(blob, b, look=40):
    """在 [b, b+look) 內找『局部 validity 最大且合格』的對齊起點,回傳該 offset 或 None。
    用於 run 進入點:避免接受 first-strict 而落在真相位 ±1/±2 的旁邊。"""
    L = len(blob)
    best = None
    best_s = -1e9
    for d in range(look):
        bb = b + d
        if bb + 32 > L:
            break
        r = rows16(blob, bb)
        if not is_glyph_strict(r):
            continue
        s = validity(r)
        if s > best_s:
            best_s = s
            best = bb
    return best


def extract_section(data, offs, si, debug=False, cont_bonus=0.0):
    """run-based 抽取一個 section,回傳 [(byte_off, rows) ...]。

    策略:
      1. 滑動找到第一個合格區段,在小窗內取『validity 局部極大』為 run 起點
         (避免落在真相位旁邊 ±1/±2)。
      2. run 內:候選 = {維持相位 b, 翻相位 b+1};各做 1-step lookahead——
         不只看本字 validity,也看『接著 +32 的下一字』validity,合成總分。
         維持相位給連續性 bonus;取總分較高且合格者輸出,b 推進到該位置 +32。
      3. 兩相位皆不合格(掉出 run)→ +1 滑動,直到在小窗內重新找到 run 起點。
    """
    base = offs[si]
    blob = data[base:offs[si + 1]]
    L = len(blob)
    out = []
    b = 0
    CONT_BONUS = cont_bonus      # 維持相位獎勵(僅用於打破近似平手)
    in_run = False
    miss = 0

    def frame_score(bb, lookahead=True):
        """本字 validity(+若 lookahead,加上接著 +32 下一字的 validity)。"""
        if bb + 32 > L:
            return -1e9, None
        r = rows16(blob, bb)
        s = validity(r)
        if lookahead and bb + 64 <= L:
            nb = bb + 32
            rn = rows16(blob, nb)
            ns = validity(rn)
            # 翻相位的下一字也可能再翻,取 {nb, nb+1} 較佳者當 lookahead
            if nb + 33 <= L:
                rn1 = rows16(blob, nb + 1)
                ns = max(ns, validity(rn1))
            s += 0.5 * ns
        return s, r

    while b + 32 <= L:
        if not in_run:
            bb = _best_phase_here(blob, b)
            if bb is None:
                b += 1
                continue
            b = bb
            in_run = True
            miss = 0
            # 不 continue;直接落入 run 處理本位置

        # run 內:在小範圍 {b-1..b+2} 重新對齊(吸收 1-byte 插入造成的相位漂移,
        # 以及上一字落點 ±1/±2 帶來的累積偏移)。維持相位 b 給連續性 bonus,
        # 其餘偏移要明顯較佳才奪標。各候選帶 1-step lookahead。
        cands = []
        last = out[-1][0] - base if out else -10**9   # 上一字落點(相對),避免回頭重疊
        for d in (0, 1, -1, 2, -2, 3):
            bb = b + d
            if bb < 0 or bb <= last:        # 不得回退到已輸出位置之前(避免重疊/倒退)
                continue
            s, r = frame_score(bb)
            if r is None or not is_glyph_strict(r):
                continue
            bonus = CONT_BONUS if d == 0 else 0.0
            cands.append((s + bonus, bb, r))

        if cands:
            _, bb, rr = max(cands, key=lambda c: c[0])
            out.append((base + bb, rr))
            b = bb + 32
            miss = 0
        else:
            miss += 1
            b += 1
            if miss > 80:
                in_run = False
                miss = 0
    # 局部精修:run 過程可能累積 ±1/±2 相位漂移;對每個輸出字在 ±2 內找 validity 局部極大,
    # 並 snap 過去(不得越過前後鄰字造成重疊)。修正「整段慢慢偏掉」的累積誤差。
    out = _refine(blob, base, out)

    if debug:
        print(f"sec{si:2}: secstart={base} 抽出 {len(out)} 字")
    return out


def _refine(blob, base, out, win=2, passes=2):
    L = len(blob)
    rels = [o - base for o, _ in out]
    for _ in range(passes):
        changed = False
        for i in range(len(rels)):
            cur = rels[i]
            lo = (rels[i - 1] + 1) if i > 0 else 0
            hi = (rels[i + 1] - 1) if i + 1 < len(rels) else L - 32
            best = cur
            best_s = validity(rows16(blob, cur))
            for d in range(-win, win + 1):
                bb = cur + d
                if bb < lo or bb > hi or bb < 0 or bb + 32 > L:
                    continue
                s = validity(rows16(blob, bb))
                if s > best_s + 1e-9:
                    best_s = s
                    best = bb
            if best != cur:
                rels[i] = best
                changed = True
        if not changed:
            break
    return [(base + r, rows16(blob, r)) for r in rels]
