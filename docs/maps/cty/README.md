# CTY 全場域地圖(各 section 標號)

tools/render_all_cty.py 把每個 CTYxx.DAT 的每個非空 section 渲成全圖,
標 CTY號 / sec號 / 尺寸 / BLK號。每 CTY 用 map_blknum(DGROUP 0x0a04)選對的 tile 圖庫。

- 一檔一 CTY:CTY00.png … CTY93.png(縱向堆疊該 CTY 的所有 section)。
- 城鎮地表層 = sec0;室內房間 / 樓層 = sec1,2,3…(docs/34/35)。
- 地名對照見 ../cty_name_fill.md、../../32-world-locations.md。

用途:肉眼瀏覽全場域、辨識地點(如露依達酒場在哪個 CTY/section)、對照 remake 渲染。
