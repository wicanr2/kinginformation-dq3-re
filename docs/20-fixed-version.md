# 精訊 DQ3 修正版(RE 價值的證明)

既然 RE 已達 [byte-identical 重組(100% PASS)](17-build-toolchain.md),就有能力**精準修改原版**。
本文記錄用反組譯成果做出的「修正版」——把當年害人卡關的 bug 修掉,在 DOSBox 跑起來。

> 產生:`python3 tools/build_fixed_version.py` → `work/DQ3_fixed.EXE` + `work/dq3_fixed_game/`(可餵 DOSBox)。
> 各 bug 完整反組譯分析見 [`docs/18-bug-analysis.md`](18-bug-analysis.md)。

## 已修(可立即驗證)

| Bug | 修法 | 位置 / 手段 | 信心 |
|---|---|---|---|
| **#1 巴拉摩斯打不死** | EXE in-place,clamp 參照欄 `[bx+0x2336]→[bx+0x2334]`(2 處 `36→34`)| file 0xbe04 / 0xbe0a | 高(與青衫官方碼逐 byte 一致)|
| **#2 彩虹水滴誤拿黃寶珠** | EXE in-place,合成成品 item code `0x6b 銀寶珠→0x75 彩虹水滴` | file 0x77e9 | 高 |
| **#3 五頭龍/歐里狄加戰當機** | **SHP 補圖**:重繪 id128 歐里狄加 / id129 五頭龍大王 的 sprite(原本是未完成填充值→blit 越界)| `tools/make_sprites.py` → `DQ3MNS.SHP` | 高 |

**EXE 僅差 3 bytes**(同長度、不動 MZ header/reloc,佈局不變);SHP 補上 2 隻有效 sprite。

> Bug #3 兩種修法:本專案用「**補回真實 sprite**」(忠實復原父子決鬥雙方);青衫官方碼則用
> 「把 boss 0x80/0x81 改接既有 sprite 0x7e/0x7d/0x7c」(快速避免當機,但顯示替代怪)。本修正版採前者。

## 交叉驗證:青衫攻略內附官方修正碼

青衫攻略 HTML 藏有「修改:DQ3.EXE」一節,列出當年社群實際使用的 8 段 EXE patch。
這些 patch 的 byte pattern 在 EXE 內**全部精準命中我們反組譯定位的位址**——等於 1990 年代的玩家
社群替我們的 RE 做了獨立交叉驗證(他們當年跑這些 patch 通關)。

## DOSBox 驗證

`work/dq3_fixed_game/`(原素材 + 修正 EXE/SHP)在 DOSBox 啟動:**開機 → 標題 DRAGON FIGHTER III →
進遊戲 → 城鎮走動** 全部正常,與原版開場一致,**修正未造成任何回歸**。
(九頭龍戰本身位於索瑪城深處,headless 自動化無法走到,故該場戰鬥的「不再當機」以
sprite 資料層 + blit 邏輯 + 青衫官方碼一致性 佐證;留待有存檔時實機補測。)

## 未修(留待 SDL2 重寫,屬 C 層 / 資料層)

- #4 勇者 MP+1:`dragon0.dat` 成長表勇者列 MP slope(需 dump 該檔)。
- #5 高等級升級錯亂:`sub_ecdb`(file 0xecdb)等級當索引越界查門檻/咒文表,需 clamp(code cave)。
- #6 數值 255 溢位:成長中間值 8-bit `mul`+`adc` wrap,需改型別 + clamp。
- #7 隼劍只打一次 / 魔甲無抗魔:雙擊與抗性**功能整段未實作**(裝備特效是 item-code 白名單,不讀 +5/+6 旗標),需補新碼。
- #7c 祈禱之戒:反組譯顯示破壞邏輯存在(~25% 觸發,file 0x5af4),**與青衫前提相左**,建議實機佐證後再定。
