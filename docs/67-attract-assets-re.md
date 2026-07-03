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

- plane-major 排列(整面 4 plane 分離)解出**連貫的單色雙人立繪**(灰階試看);row-interleave 排列破碎 → **採 plane-major**。
- 內容:近景雙人物 mono 美術,屬開場演出族(「FIRST」= 開機第一張?)。真 palette 與在序列中的位置待影格對位。

## 對戰役的意義

1. **attract/開場/結局美術層 = 21 張現成 PCX + 1 張 raw**,解碼器已存在(`internal/dq3data/pcx.go`)——
   ebitan 實作 attract 演出**沒有素材障礙**,剩排程邏輯(輪播順序/秒數/能力條動畫)= 影格對位即可。
2. oracle 待核對清單 #10 大半關閉:職業共 **8 張**、WORTHY=賢者、PLAYER=遊玩者、立繪=獨立 PCX。
3. 殘項:TIT1/TIT2/TITA/TITC 鑑定、FIRST.SCR palette、attract 輪播節奏(秒/張)、
   能力條捲動速度 —— 全部可由現有影格 + dump 完成,無需新 RE。
