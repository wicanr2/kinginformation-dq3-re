# 83 — 旅店交易與羅馬利亞北行 blocker

## 原版旅店證據

DQ3.EXE file `0x86f5` 是旅店 handler：

- `0x86fb` 讀 CTY facility block `+1` 的 raw 單價。
- `0x8700 call 0xd966`；該函式以 `[0x5077]` 為隊伍筆數，逐筆讀角色指標
  `[0x4f15]`，跳過死亡旗標 `+0x38 & 0x80`，只對目前 HP `+0x16 != 0` 者增加 `bp`。
- `0x8705 mul cl` 將 raw 單價乘 `bp`，結果存 `[0x2593]`。
- `0x8736..0x8754` 先檢查金幣再扣款；不足時不進恢復交易。
- `0x87cb` 再逐筆跳過死亡旗標，把目前 HP/MP `+0x16/+0x18` 寫成上限
  `+0x2a/+0x2c`。旅店不復活死亡角色。

Ebiten `openFacility(facInn)` 已按上述語意實作，component tests 鎖定：

1. 費用只乘存活人數。
2. 勇者與存活同伴全部補滿 HP/MP。
3. 死亡同伴維持死亡。
4. 金幣不足時不扣款、不部分恢復。

## Production route 的 blocker 與閉合結果

由羅馬利亞王座 checkpoint 正常回城、出城並步行往 CTY10 的診斷路線顯示：

- 舊旅店只補勇者，確實使同伴跨區累積傷害；此項已修正。
- 修正後抵達羅馬利亞旅店時曾只有三人存活，因此原版公式收 `3×3=9G`，不是四人
  `12G`；死亡者不能靠住宿恢復。
- 正式使用聖水後，路線仍會在效果結束後遇到 monster 17；全滅樣本是單隻怪，不是群量
  爆增。
- 全滅時勇者仍為 Lv1。現行 trace 走最短任務路徑且對 id≥10 一律嘗試逃跑，沒有重現
  原版 RPG 的前期練級循環。

2026-07-29 已完成前兩項，不採用削弱 monster 或注入資源：

1. 正式 trace 用國王 50G 買 2 株藥草，保留復活／住宿預算；在阿里阿罕入口同一
   encounter region 移動、戰鬥、回城，經正式教會與旅店練到 hero Lv4
   (`EXP=324`)，再繼續跑到羅馬利亞。
2. 教會 handler 已有三選單／選人／yes-no／扣款／復活交易。40 級費用由
   `DQ3.EXE` file `0x19dac` 證實並存入 `dq3_cht` JSON pack；file
   `0x85ff..0x8696` 的 level clamp、付款、清死亡與 HP/MP 回滿均有 component test。
3. 後續 CTY10 audit 已閉合至 E3「自然進入 Kandar 戰」：正式 trace 由此 checkpoint
   北行、開 sec3 盜賊鑰匙門並踏 sec5 trigger；見 `docs/85`。本文件不再維護下一步。

尚未證明 Lv4 隊伍能合法取勝；戰力、裝備與補給 audit 由 `docs/74`／`docs/85` 接續。
