# DOSBox oracle:自動推進入遊戲(姓名輸入破關 + 地表/城鎮主角捕捉)

remake 要做到「一模一樣」需以**原版 DOSBox 當 oracle 逐畫面比對**;phase ⑤ 與「真主角
sprite」都卡在「headless 自動化進不了遊戲」——精訊版開場強制注音姓名輸入。本文記錄已打通的
自動完成序列與首批 in-game 捕捉。

## 注音姓名輸入的自動完成(反組譯 + 實測)

反組譯 dispatcher `ni_dispatch`(sub_0d17,file 0x2087)+ `ni_alnum`(0x21fe)+ cell 分支
(0x22b2)定出機制:

- 鍵盤 grid 5 列 × 9 欄,游標 raw index `[0x26fe]`(0..44),`raw = row×9 + col`;
  `cell = col×5 + row`(欄優先);Up/Down=±9、Left/Right=±1(mod 45)、Space/Enter 選格。
- 模式旗標 `[0x26fc]`:bit0=注音、bit1=英數、bit2=功能列焦點、bit3=完成。
- **英數模式選 cell 44(raw 44 = 右下角)→ 直接 `or [0x26fc],8`(完成旗標)**;
  cell 43 → `or [0x26fc],4`(切功能列)。dispatcher 見 bit3 且名字非空(`[0x270a]≥1`)即定案。

**最短確定性按鍵序列**(注音起點 raw0):
1. 切英數:`Down×3 Right×8`(raw0→raw35=cell43)+ `Space`。
2. 打 1 字:`Up×3 Left×8`(raw35→raw0=字元'0')+ `Space`。
3. 完成:`Down×4 Right×8`(raw0→raw44=cell44)+ `Space` → 名字非空,定案進遊戲。

工具:`tools/dosbox_to_overworld.sh`(dq3-dosbox 容器內跑,Xvfb + xdotool,等待有上界)。
先前 `dosbox_ingame.sh` 卡在姓名畫面打轉,根因即「Space 只選格、未移到完成格」。

## 首批 in-game 捕捉(oracle 成立)

完成姓名後實測進到:
1. **起始室內房間**(阿里阿罕城內):紅地毯地板、白磚牆、桌、樓梯;**主角小 sprite** 居中。
   佈局與本專案 `dq3_town`(CTY)引擎渲染同型(紅地毯磚房),互為印證。
2. **城鎮戶外**(走出房間後):草原 / 水 / 樹 / 紅磚路 / 白磚紅頂建築 / 花叢;**2 個角色小
   sprite**(主角棕髮、同行金髮,皆藍衣),配色清楚。

→ oracle 流程打通:可程式化進到任意早期場景截圖,供 phase ⑤ 逐畫面比對與調色盤/sprite 校正。

## 主角 sprite 觀察(供「真主角 sprite」下一步)

- 地表/城鎮主角為**小 sprite**(目測 ~16×16–16×24),小於一個 tile(32×24);與名字畫面的
  大張立繪(`game2.png` 那種)不同,也與 `DQ3MNS.SHP` 怪物(較大)不同。
- 配色(DQ3.PAL 場景調色盤下):棕/黑髮帽、藍衣、膚色臉;同行角色金髮藍衣。
- 待辦:把此 sprite 對應到資產來源 —— 候選 `DQ3MAN.BLS`(已解空間格式 384B/32×24,docs/27,
  但場景 sprite 似 16 寬,需確認是否 16×24 子格 / 不同 stride)或 BLK tile 範圍;
  以本 oracle 畫面為基準反推正確子調色盤,取代 remake 佔位框。

## 對 bug #8(戰鬥後變黃/綠)的驗證價值

有了 oracle,可進一步:進一場戰鬥 → 勝利離場 → 截地表畫面,實機坐實 #8 的色偏與不恢復
(docs/28),並對照 remake「場景返回必重套 palette」確認免疫。
