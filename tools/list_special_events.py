#!/usr/bin/env python3
"""列全遊戲「走過去按空白觸發」的特殊事件格(boss / 機關 / 劇情觸發點)。

第一性原理(使用者洞見):boss 不是會走動的 NPC,是地圖上固定一格放著 sprite 的物件,
玩家走過去按空白鍵(examine)就觸發。所以「boss 正式觸發點」== 某 CTY/section/座標上
kind=special 的事件格。

權威來源 = docs/data/npc_dialogue.json(已 RE 解碼的全 CTY NPC dump,欄位:
  sec, bank, x, y, sprite(sprite entry), sub, kind(talk/facility/special), dlg(對話 record), text)。
這是反組譯產物,比手刻 parser 可靠(手刻易踩 section 偏移錯位)。kind=special = examine 觸發事件;
dlg = 該事件的對話/事件 record 號(boss 事件如八頭大蛇 CTY19 sec1 (35,12) dlg=45,對上 byte4=45)。

注意:special 同時涵蓋 boss、機關(按鈕/石門)、give 道具劇情 NPC——dlg 號是事件 id,
不保證等同 dq3_scripted 的 byte4(多數一致,但需個案以對話內容/攻略確認)。

用法:
  tools/list_special_events.py              # 列全部 special 事件
  tools/list_special_events.py CTY44        # 只列某 CTY
  tools/list_special_events.py --text       # 附對話 text(辨識 boss vs 劇情)
"""
import json, sys, re, os

J = os.path.join(os.path.dirname(__file__), '..', 'docs', 'data', 'npc_dialogue.json')
d = json.load(open(J, encoding='utf-8'))

# 分類狀態(持久化,避免重複盤點;見 special-events-classified.json)
CLS_PATH = os.path.join(os.path.dirname(__file__), '..', 'docs', 'data', 'special-events-classified.json')
classified = {}
if os.path.exists(CLS_PATH):
    classified = json.load(open(CLS_PATH, encoding='utf-8')).get('events', {})

want_cty = None
show_text = '--text' in sys.argv
todo_only = '--todo' in sys.argv      # 只列未分類 / review(尚需處理的),跳過已定性的
for a in sys.argv[1:]:
    if a.upper().startswith('CTY'):
        want_cty = a.upper()

rows = []
for cty, recs in d.items():
    if want_cty and cty.upper() != want_cty:
        continue
    cn = int(re.sub(r'\D', '', cty))
    for r in recs:
        if isinstance(r, dict) and r.get('kind') == 'special':
            rows.append((cn, r.get('sec'), r.get('x'), r.get('y'),
                         r.get('sprite'), r.get('dlg'), r.get('text', '')))

rows.sort(key=lambda r: (r[0], r[1] or 0, r[5] or 0))
# 去重(同 cty+dlg 一個事件)+ 套分類
seen_key = set()
shown = []
for cn, sec, x, y, sp, dlg, t in rows:
    key = 'CTY%d:%d' % (cn, dlg)
    if key in seen_key:
        continue
    seen_key.add(key)
    cat = classified.get(key, '?未分類')
    if todo_only and cat not in ('?未分類', 'review'):
        continue
    shown.append((cn, sec, x, y, sp, dlg, t, cat))

from collections import Counter
catc = Counter(classified.get('CTY%d:%d' % (cn, dlg), '?未分類')
               for cn, _, _, _, _, dlg, _ in
               [(r[0], r[1], r[2], r[3], r[4], r[5], r[6]) for r in
                {(r[0], r[5]): r for r in rows}.values()])
hdr = '=== kind=special 事件%s,共 %d(分類:%s)===' % (
    '(--todo:僅未分類/review)' if todo_only else '',
    len(shown), ' '.join('%s=%d' % (k, v) for k, v in sorted(catc.items())))
print(hdr)
for cn, sec, x, y, sp, dlg, t, cat in shown:
    line = '  [%-13s] CTY%-2d sec%s (%2s,%2s) sprite=%-3s dlg=%-3s' % (cat, cn, sec, x, y, sp, dlg)
    if show_text and t:
        line += '  ' + t.replace('\n', ' ').replace('/', '')[:42]
    print(line)
