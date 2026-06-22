# 精訊 DQ3 修正版(RE 價值的證明)

既然 RE 已達 [byte-identical 重組(100% PASS)](17-build-toolchain.md),就有能力**精準修改原版**。
本文記錄用反組譯成果做出的「修正版」——把當年害人卡關的 bug 修掉,在 DOSBox 跑起來。

> 產生:`python3 tools/build_fixed_version.py` → `work/DQ3_fixed.EXE` + `work/dq3_fixed_game/`(可餵 DOSBox)。
> 各 bug 完整反組譯分析見 [`docs/18-bug-analysis.md`](18-bug-analysis.md)。

## 已修(binary patch 對照組,5/7;DOSBox 開機+進城+姓名畫面驗證無回歸)

| Bug | 修法 | 位置 / 手段 | 信心 |
|---|---|---|---|
| **#1 巴拉摩斯打不死** | EXE in-place,clamp 參照欄 `[bx+0x2336]→[bx+0x2334]`(2 處 `36→34`)| file 0xbe04 / 0xbe0a | 高(與青衫官方碼逐 byte 一致)|
| **#2 彩虹水滴誤拿黃寶珠** | EXE in-place,合成成品 item code `0x6b 銀寶珠→0x75 彩虹水滴` | file 0x77e9 | 高 |
| **#3 五頭龍/歐里狄加戰當機** | **SHP 補圖**:重繪 id128 歐里狄加 / id129 五頭龍大王 的 sprite(原本未完成填充值→blit 越界)| `tools/make_sprites.py` → `DQ3MNS.SHP` | 高 |
| **#4 勇者 MaxMP 成長偏低** | EXE 靜態 DGROUP 成長表(file 0x1a4a6):勇者 MP base `03→08`、slope `05→0a`(Lv43 MaxMP ~107→~217,可放比荷瑪拉)| file 0x1a4a8/9 | 高(DOSBox 驗證)|
| **#7a 隼劍只打一次** | EXE 同長度區段改寫(file 0xc1fa,25B):復用引擎既有但「沒接線」的 re-attack 路徑,把第二擊觸發條件改成「攻擊者武器==飛鷹劍 0x6e」| file 0xc1fa | 高 |

**EXE 共 25 bytes 變動**(同長度、不動 MZ header/reloc,佈局不變);SHP 補上 2 隻有效 sprite。

> #4 / #6 過程中**更正了 docs/18 兩處定位**:#4 成長表在 EXE DGROUP(非 dragon0.dat);#6 成長公式其實 16-bit 正確。
> #7a 的「第二擊機制本就存在、只是觸發條件接錯位元」是漂亮的反組譯發現。

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

## 未修(留 C 層 / SDL2;根因已定位)

- **#5 高等級升級錯亂**:8 職業門檻表各 44 entry(MAX_LEVEL=43),lv44 越界讀下職業 entry[0]=0 → 連升+學錯咒。修法 `clamp level≤43`,需 code cave —— 但 resident 碼**最大可用 cave 僅 8 byte**(clamp+trampoline 需 11~19B),安全優先不冒險。
- **#6 數值 255 溢位**:成長公式經核實為 16-bit 正確,靜態找不到可信 byte 截斷點(疑在顯示寬度或與 #5 互動),留 C 層釐清。
- **#7b 魔法鎧甲無抗魔**:咒文扣血需掃 8 格裝備找魔甲=迴圈,無同長度空間/安全 cave,硬塞會壞傷害公式。資料其實沒缺(魔甲 +6 bit 已置位),純粹程式從不讀取。
- **#7c 祈禱之戒**:反組譯確認破壞邏輯本就生效(file 0x5af4,`cmp al,0x40; ja 不壞`,~25% 觸發),**不需修**;青衫「永不壞」應為憑記憶的概略陳述。

> 這三項(#5/#6/#7b)正是「走正路 (b):從 C 重編」要在 **C source** 修對的——binary patch 對照組受 real-mode 同長度/cave 限制做不到,C 層沒有這些限制。
