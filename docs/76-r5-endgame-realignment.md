# 76 — R-5 終盤原版資料校正

本批以 IDA Pro 9.4、CTY67/90 原始資料與本機完整遊戲影片交叉驗證，撤回舊 C remake
及早期文件中未由原版證明的終盤設定。

## 龍女王與光之珠

- CTY67 section0 NPC `(14,24)`，sub2 handler `52`。
- IDA `sub_15E02`：給 item `0x65` 後 CLEAR story flag `0x4e`、SET story flag `0x19`，
  並播放 D3TXT06 rec71。
- Ebiten 舊碼只執行 generic give-item；R-5 已補兩筆 story flag，並驗證存檔 round-trip。

## 彩虹橋

- IDA `loc_14243` 僅接受 `[0x4f2f]=0x7f`、`[0x4f31]=0x175`。`0x100` 是地下世界層，
  因此實際玩家座標是 `(127,117)`。
- 成功後 OR 原版 world state `[0x4f44]` bit `0x40`，並將 `(126,117)` 寫為 tile `0x53`。
- 舊碼的「地下世界任何位置皆可使用」及 remake flag `0x35` 均不成立。新版持久化
  `worldState`，舊存檔若有 `flag0x35` 會一次性遷移。

## CTY90 與索瑪

- CTY90 才是索瑪城；CTY79/80 是結局／拉達多姆變體。
- CTY90 section4 的 `(19,29)` 是 `attr=0 + hiMap subid1` 的 plain-tile transition，
  目的為 section5 `(12,55)`。舊 loader 只對 CTY72 開放這種轉場，造成終盤不可自然到達；
  R-5 已依 CTY90 原始 transition table 接通。
- section5 的六個 events 是 item `0x4c`、flags `0xef..0xf4` 的世界樹之葉寶箱。
  本機影片約 3190–3210 秒也顯示大魔人為一般隨機遭遇，常以兩隻成組出現。
- IDA formation bytes：
  - `DS 0x4EFA` → monster `0x7a`（巴拉摩斯怨靈）
  - `DS 0x4EFF` → monster `0x7b`（巴拉摩斯殭屍）
  - `DS 0x4EF0` → monster `0x7c`（索瑪）
- handler80 `sub_164CD` 在索瑪戰前播放 D3TXT07 rec72。舊 remake 的
  `6×monster106 + flag0x214` 是攻略文字推測，不是原版固定事件，已移除。
- 正式入口改為 CTY90 section5 handler80 NPC 的「話す」，不再依賴 KeyZ。

## 驗證範圍

- 正確座標使用彩虹水滴前後 tile mutation 與存檔還原。
- CTY90 section4 plain transition 經 production `tryTransition()` 進 section5。
- section5 面向 handler80 NPC，以正式命令窗「話す」啟動 monster `0x7a`。
- 固定序列 `0x7a → 0x7b → rec72 → 0x7c`。
- 移除 T/R/Z、Enter 直進城與 Cancel 直出城後門；露易達酒館保留正式 NPC 入口。

R-5b 已完成王座隱藏樓梯顯現、歐里狄加天橋旗標／對白過場，以及索瑪戰後返回
拉達多姆國王冊封再進 ending，詳見 `docs/77-r5b-castle-aftermath.md`。歐里狄加
formation 0x80 的逐回合固定動畫、非戰鬥傳送咒文選單與無除錯完整通關仍待完成，
不得把目前的旗標閉環擴張宣稱為整個 remake 已完成。
