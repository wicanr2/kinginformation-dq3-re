# 67 — TIT*.P / FIRST.SCR 鑑定:attract·開場·結局美術全定位(2026-07-03)

> docs/65 戰役的 new-RE 項:原版錄影裡的「職業大立繪 attract 演出」素材位置。
> 結果:**一次全定位 —— 全部是獨立 PCX(TIT*.P,既有 DecodePCX 直接可解),無需再追資料流。**
> 方法 = rulebook 64 截圖 oracle:已破解碼器(dq3_pcx)解 20 檔 → 對原版錄影影格肉眼比對。
> dump 工具:`dq3_remake_ebitan/internal/dq3data/titdump_test.go`(env-guarded,輸出只准進 gitignored 目錄)。

## 鑑定表(20 檔 TIT*.P,全 640×350 PCX 4-plane RLE,自帶 16 色 palette)

| 檔 | 內容 | 對照 oracle |
|---|---|---|
| TITG | 標題(logo + ©1993 精訊) | docs/36 `02_title_splash`;影片 0:00 ✓(remake 已用) |
| TITF | 標題背景(無 logo) | 影片 ~4:30 過場 ✓ |
| TITB | 單色巨龍 vs 勇者(綠灰 mono) | docs/36 `01_opening_cutscene`「單色巨龍圖」✓ |
| TITD | 單色三人組立繪(戰士/勇者/僧侶) | 影片 ~0:45 暗場三人立繪 ✓ |
| TITE | 「1993」年份卡(黑底) | attract/開場演出張 |
| TITP | 太陽翼紋章(黑底) | attract/開場演出張(待影格對位) |
| **TITH** | **WARRIOR(戰士)** 男女立繪+STR/DEX/CON/INT/LUK 標籤(右版式) | attract 職業巡禮 |
| **TITI** | **PRIST(僧侶)**(左版式,黃綠光柱) | 影片 ~2:00 ✓ |
| **TITJ** | **MAGICIAN(魔法使者)**(右版式,藍光柱) | 影片 ~2:30 ✓ |
| **TITK** | **FIGHTER(武鬥家)**(左版式) | 影片 ~3:00 ✓ |
| **TITL** | **MERCHANT(商人)**(右版式) | 30s 取樣漏掉,補齊 |
| **TITM** | **WORTHY(賢者)**(左版式) | 影片 ~3:30 ✓(WORTHY=賢者,待核對清單 #10 解) |
| **TITN** | **PLAYER(遊玩者)**(右版式,小丑+兔女郎) | 影片 ~4:00 ✓ |
| **TITO** | **HERO(勇者)**(左版式,紅披風男女) | 職業第 8 張 |
| TIT3 | **THE END + 「TO BE CONTINUE DRAGONFIGHTER I」**(logo+城堡+雲海) | 結局畫面(oracle 步 56e 之後) |
| TIT0 | 山景+城堡剪影橫幅(疑結局/opening 卷軸景,右側含第二段畫面) | 待影格對位 |
| TIT1 / TIT2 / TITA / TITC | **待鑑定**(已 dump,scratch_re/titp/) | — |

- **職業畫面版式**:左右兩種鏡像版式交替(名稱花體英文字 + STR/DEX/CON/INT/LUK 標籤烤在圖裡;
  **能力條是執行期畫的**(影片中動態捲動),圖裡只有標籤與底線起點)。
- **職業順序(檔序)**:H 戰士→I 僧侶→J 魔法→K 武鬥→L 商人→M 賢者→N 遊玩→O 勇者;
  影片取樣順序(PRIST→MAGICIAN→FIGHTER→WORTHY→PLAYER)與檔序相容(I→J→K→(L漏)→M→N),
  推定 attract 依檔序輪播,精確起點/停留秒數待影格細對。

## FIRST.SCR(112,000 bytes = 640×350×4bpp 整,無 header raw)

- **row-interleaved 排列**：每掃描列依序放 plane0..3，各 80 bytes；由 `DecodeRowInterleaved4BPP`
  解出 640×350 雙人物原畫。
- 原版另設開場專用 16 色 DAC；palette 已由 DOSBox 能力確認／姓名畫面 oracle 固定並資料化在
  `interface.json.new_game_confirmation.palette`，不可直接套用 `DQ3.PAL`。

## Ebiten 接線狀態（2026-08-09）

`dq3_cht` game pack 已將 `TITH.P`→`TITO.P` 的八張卡以 `attract.frames` 宣告，manifest
保存每個 PCX 的大小與 SHA-256；標題閒置 1200 個 60Hz frame 後由共用引擎輪播，每張卡
停留 1200 frame。原始錄影每秒抽格在 `o0090..o0250` 閉合卡片順序與約 20 秒間隔，故
這組 timing 標為 **D2（影片 timing）**；PCX 檔案本身與 consumer 為 **D3**。任何確認／取消／
方向／觸控輸入都會中斷巡禮並交回原本主選單，無輸入注入或 debug shortcut。runtime 對拍：
[`title_attract_warrior.png`](../dq3_remake_ebitan/docs/img/title_attract_warrior.png)。

同一份 pack 也宣告 `FIRST.SCR` 為 `new_game_confirmation` raw screen；正式 `ngConfirm` renderer
先解碼背景，再疊 stat panels。runtime 對拍圖為
[`ng_confirm.png`](../dq3_remake_ebitan/docs/ng_confirm.png)，目前原始檔尺寸／解碼與配色維持 D2，runtime 背景達 V2，
stat panel 的逐像素幾何仍待同狀態 V3。

仍未宣稱逐像素 V3 的部分是原版能力條逐幀填充與淡入／淡出；目前先完成可重播的八卡
E2／V1 vertical slice，待取得同狀態 DOSBox 逐幀 oracle 後再升級 timing／動畫。

## 對戰役的意義

1. **attract/開場/結局美術層 = 21 張現成 PCX + 1 張 raw**；PCX 由
   `internal/dq3data/pcx.go`、FIRST.SCR 由 `DecodeRowInterleaved4BPP` 解碼，Ebiten 已接入
   attract 與能力確認背景，剩排程／能力條動畫與同狀態 V3 對拍。
2. oracle 待核對清單 #10 大半關閉:職業共 **8 張**、WORTHY=賢者、PLAYER=遊玩者、立繪=獨立 PCX。
3. 殘項:TIT1/TIT2/TITA/TITC 鑑定、attract 輪播節奏(秒/張)、能力條捲動速度與
   能力確認 stat panel 幾何 —— 可由現有影格 + dump 繼續閉合,無需新 RE。
