# 原版實機操作:開機 → 創建主角 → 進阿里阿罕 → 露依達酒場創角

在 DOSBox(headless:Xvfb + xdotool + import 截圖)實際操作原版精訊 DQ3
(`assets_raw/DQ3.EXE`),把「開機到創建一名隊員」的完整流程跑通並逐畫面記錄。
重點在還原**新主角姓名輸入子系統**的精確按鍵操作(這是 headless 自動化最難的一關),
與露依達酒場/冒險者登錄所的結構。

> 與既有文件的關係:docs/29 是「自動推進入遊戲」的早期做法,docs/14/15 反組譯了姓名
> 輸入 handler。本文用原版 EXE 實機把整條流程跑通並校正按鍵語意,補上酒場/登錄所的
> 劇本結構(從遊戲自身資料 D3TXT00 解出)。截圖見 `docs/36_shots/`。

## TL;DR 操作流程

1. **開機**:DOSBox 跑 `dq3` → 開場 cutscene(單色巨龍圖)→ 標題畫面
   (DRAGON FIGHTER III 傳說的終章)→ 主選單「勇者鬥惡龍:▶遊戲開始 / 載入進度」。
2. **遊戲開始**:選「遊戲開始」→ 進**注音姓名輸入**(必填關卡,不能跳)。
3. **姓名輸入**(關鍵):grid 是 5×9 注音鍵盤,右側獨立「功能列」(英數/前進/後退/取消/完成)。
   - 用 `Down×3 Right×8` 到 **TOFN 格(cell43,視覺是 ⊙)** → `Space` 進功能列。
   - 功能列 row1 `Space` 切英數(英數模式打 1 字最省事)→ 回 grid 打字 → 再進功能列 → 完成。
4. **性別**:選單「▶男性 / 女性」。
5. **開場劇情**:黑底對話框旁白(主角生日)。
6. **進遊戲**:阿里阿罕主角故鄉的**室內房間**(紅地磚 + 白磚牆 + 床 + 家具 + NPC)。
7. **走到露依達酒場 / 冒險者登錄所** → 創角(職業 → 姓名 → 性別 → 登錄名冊)。

## DOSBox 環境(可重現)

- image `dq3-dosbox`(已存在);`assets_raw/` 掛 `/game`(完整原版 191 檔),repo 掛 `/work`。
- `tools/dosbox.conf`:`machine=svga_s3`、`nosound`、autoexec `mount c /game; c:; dq3`。
- 跑法(同步前景,有界,無 sentinel):
  ```bash
  timeout 160 docker run --rm -v "$PWD":/work -v "$PWD/assets_raw":/game \
    dq3-dosbox bash /work/tools/<script>.sh
  ```
- headless 看畫面:`import -window root out.png`(失敗回退 `xwd | convert`)→ `Read out.png`。
  **不開任何 GUI viewer**。送鍵前必 `xdotool windowfocus/windowactivate` 該 DOSBox 視窗。

## 逐步操作序列(按鍵 + 畫面 + 截圖)

| 步 | 按鍵 | 畫面變化 | 截圖 |
|---|---|---|---|
| 開機 | (等 ~12s) | 開場 cutscene:單色巨龍 + 勇者 | `36_shots/01_opening_cutscene.png` |
| 過 cutscene | `Return` ×數下 | 標題:DRAGON FIGHTER III 傳說的終章 ©1993 精訊 | `36_shots/02_title_splash.png` |
| 進主選單 | `Return` | 框「勇者鬥惡龍 / ▶遊戲開始 / 載入進度」(紅披風勇者立繪兩側) | `36_shots/03_title_menu.png` |
| 遊戲開始 | `Return` | **注音姓名輸入**:5×9 注音 grid + 右欄功能列;游標 raw0=ㄅ | `36_shots/04_name_input_zhuyin.png` |
| 進功能列 | `Down×3 Right×8` → `Space` | 游標到 ⊙(cell43)後焦點跳右欄,▶英數 | `36_shots/06_fnlist_focus.png` |
| 切英數 | `Space`(在 row1 英數) | grid 變英數鍵盤 `0 5 A F K…/ OK`,功能列首列改「注音」 | `36_shots/05_name_input_alnum.png` |
| 打 1 字 | `Up×3 Left×8` → `Space` | 回 raw0='0',名字列 +1 字 | — |
| 完成 | `Down×3 Right×8` `Space` → `Down×4` `Space` | 名字非空 + 選「完成」→ 離開姓名輸入 | — |
| 性別 | (自動進) | 選單「▶男性 / 女性」 | `36_shots/07_gender_select.png` |
| 開場 | `Return` ×~14 | 黑底劇情旁白(生日)→ 進城 | — |
| 進城 | — | **阿里阿罕室內房間**:紅地磚 / 白磚牆 / 紅藍床 / 黃櫥櫃 / 桌 / NPC / 主角 sprite | `36_shots/08_house_interior.png` |

對應已驗證腳本:`tools/dosbox_nameinput.sh`(姓名輸入 + 進城,本次以原版 EXE 重跑成功,
截到性別選單 → 室內房間)。

## 姓名輸入子系統:精確按鍵語意(本次校正)

這是整條流程最關鍵也最容易卡的一關。實機逐鍵截圖校正出以下**確切機制**(對 docs/15
反組譯):

- **grid = 1-D ring**:raw index 0..44,**Up = −9 / Down = +9 / Left = −1 / Right = +1,
  全部 mod 45**(雙軸都環繞)。`raw = row×9 + col`(row-major),`cell = col×5 + row`。
- **grid 與右側功能列是兩個獨立 widget**。**方向鍵只在 45 格 grid 內環繞,永遠到不了功能列**
  ——這是先前自動化打轉的根因。從 grid 跨進功能列**只有一條路**:走到 **cell43
  (=TOFN「切功能列」,視覺是 grid 最右欄倒數第二格 ⊙;raw35 = `Down×3 Right×8` 從 raw0)**
  按 `Space`,焦點才跳到右欄(▶ 停在第 1 列)。
- **功能列 5 列**(`win_nav(5)`,Up/Down 移列,Space 選):
  注音模式時 = `英數 / 前進 / 後退 / 取消 / 完成`;英數模式時首列變 `注音`(切回)。
  **完成 = 第 5 列**,從第 1 列 `Down×4` 到達。
- **`Space`(或 Enter)= 選取游標所在格**,**不是**「直接完成整個名字」。
- **完成放行條件**:選「完成」**且名字非空**(`[0x270a] ≥ 1`)才離開;名字空時按完成無效。
- **模式相關陷阱**:cell43(⊙)「進功能列」只在**注音模式**成立;在**英數模式**那一格
  是字元「∞」(會被當成一個字打進名字)。所以混模式盲走會打錯字、卡關。最穩做法:
  全程注音模式,或如下「切英數打 1 字」固定序列。

**最短可靠完成序列**(實測對原版 EXE 跑通,進到性別選單):
```
# 起點:姓名畫面,游標 raw0=ㄅ(新進姓名畫面固定在此)
Down×3 Right×8  Space     # raw0→raw35(TOFN ⊙)→ 進功能列(▶英數)
Space                     # 功能列 row1「英數」→ 切英數鍵盤
Up×3 Left×8     Space     # 英數 grid raw35→raw0('0')→ 打 1 字
Down×3 Right×8  Space     # raw0→raw35(TOFN)→ 進功能列
Down×4          Space     # 功能列 row1→row5「完成」→ 名字非空 → 離開姓名輸入
```
> ⚠ 用 `Space` 選格,不要用 `Return`(Return 在 grid 行為不同,會讓序列偏掉)。

## 露依達酒場 / 冒險者登錄所(從遊戲自身資料解出)

露依達的店在**阿里阿罕(CTY00)**內,是創建/管理同伴的地方。劇本(`D3TXT00`,
`docs/script/txt00.txt`)直接給出完整結構:

- **酒場入口**(rec 527):「＊『這裡是露依達的店,旅人們與同伴相會,分離的酒吧。』」
- **酒場主選單**(rec 528「客人有什麼吩咐呢?」→ rec 529):
  `找同伴參加 / 與同伴分離 / 觀看名單`(招募 / 解散 / 看名冊)。
- **冒險者登錄所**(rec 550)是**另一層**(樓上):
  「＊『這裡是冒險者的登錄所…把想成為同伴的人登錄起來,再到樓下的酒店把人找出來。』」
  → 兩段式:**登錄所建角色登錄入名冊**,**樓下酒場把名冊裡的人找來加入隊伍**。
- **8 職業**(rec 457-464):`勇者 戰士 武鬥家 僧侶 魔法使者 賢者 商人 遊玩者`。
- **性別**(rec 534/535):`男性 / 女性`。
- 名冊欄位(rec 531/465):`姓名 / 等級 / 職業 / 性別`(NO / 姓名 / 等級 / 性別)。

→ **創角流程 = 登錄所:選職業 → 取名(同主角的注音/英數輸入)→ 選性別 → 登錄入名冊**;
再到**樓下酒場「找同伴參加」**把該角色拉進隊伍。本專案 remake 的
`dq3_tavern`/`dq3_roster` 即還原此流程(職業→姓名→性別→登錄→名冊),與原版資料一致。

### 露依達店在阿里阿罕的位置(靜態解析)

- CTY00 = 阿里阿罕(docs/32「自報地名」txt01[9]「這裡是阿里阿罕」)。
- CTY00 多 section(docs/34/35):`section0` 是從地表進入看到的層,其餘 section
  (1,2,3,4…)= 室內房間 / 子區;`docs/maps/cty00_town.png` 把各 section render 出來
  (見 `36_shots/10_cty00_sections.png`):紅地磚 + 床的封閉房間 = 旅店 / 民居 / 酒場那類,
  含草地外圍的那層 = 城鎮戶外(建築門 = 紫色 tile)。
- **門/階梯/出城全在 section header `+0xc` 轉場表**(4-byte `{destCTY,destSec,X,Y}`,
  純靜態,docs/35)。CTY00 section0 有 4 道門(→ sec1/2/3/4)。露依達酒場是其中一棟建築門
  進去的內裝 section;**精確「哪道門 = 酒場」需逐 section 比對 NPC→劇本 rec 527/550 的綁定**
  (本次未做到逐格綁定,屬靜態 CTY-NPC 格式的後續工作)。

## 踩到的雷與 DOSBox 操作技巧

- **姓名輸入是 headless 最難關**:雙軸環繞的 grid + grid/功能列雙 widget,**盲走容易偏**。
  解法見上「最短可靠完成序列」,關鍵三點:(1)用 `Space` 選格;(2)新畫面游標固定 raw0;
  (3)進功能列只能走 cell43(⊙)。
- **開機時序敏感(~50% 整黑)**:DOSBox 啟動與首個 `Return` 的相對時序偶爾錯拍 → 整局黑屏
  (截圖 249 bytes = 全黑)。緩解:啟動 sleep 拉到 **14–15s**、Return 間隔 ≥1.5s;
  仍有失敗就**重跑**(本次多次重跑命中)。判活死靠**截圖檔案大小**:全黑=249;
  姓名畫面 ~11.6k;室內房間 ~5.5k;不同畫面大小不同,可當粗略狀態訊號。
- **滑鼠點選無效**:DOSBox 無 pointer integration,`xdotool mousemove/click` 的 root 座標
  對不到遊戲內 int33 命中區 → 點功能列「完成」無反應。**只能走鍵盤**。
- **有界紀律**:每送鍵固定截一張就停,`timeout` 包 docker;**禁止 `until sleep` 輪詢**;
  跑完 `docker kill` 收容器(本次收掉一個殘留容器)。
- **截圖看大小先篩**:先看 `stat -c %s` 大小變化判斷有沒有轉場,再 `Read` 該 PNG,省 token。

## 受阻點(誠實標記)

- **走出起始室內房間 → 阿里阿罕戶外 → 走到露依達店門**:**未在實機跑通**。
  起始房間是封閉建築,出鎮要踩到 section header `+0xc` 轉場表裡的**特定門 tile**;
  headless 盲走整片 42×43 / 多 section 找那一格,在時間上界內不可靠(密集掃格腳本超時)。
- 因此「實機走進酒場 → 登錄所創角畫面」這段以**靜態資料**(D3TXT00 rec 527/550、職業表、
  cty00_town.png section render)補全,**未逐格實機操作**。要實機跑通建議改用
  `dosbox_warp_cty.sh` 的手法:**patch 新遊戲起始 section/CTY**,直接 spawn 到酒場內裝,
  或解 CTY00 各 section 的 NPC→rec 綁定鎖定酒場 section,再針對性走門。

## 再生 / 工具

```
# 進遊戲(姓名輸入 oracle,原版 EXE):assets_raw 掛 /game
timeout 160 docker run --rm -v "$PWD":/work -v "$PWD/assets_raw":/game \
  dq3-dosbox bash /work/tools/dosbox_nameinput.sh     # → 性別選單 → 室內房間
# 解酒場/登錄所劇本
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/v && . /tmp/v/bin/activate && python tools/text_to_utf8.py"
grep -nE '露依達|登錄所|職業|勇者' docs/script/txt00.txt    # rec 527/550/457-464
```

本次新增的探索腳本(過程記錄,非全部成功):`tools/dosbox_tavern_*.sh`、
`dosbox_name_*.sh`、`dosbox_boot_check.sh`、`dosbox_houseexit.sh`。
代表性截圖:`docs/36_shots/`。

## CTY00 內裝 section 調查(remake 渲染,定位酒場用)

逐 section 渲染(`DQ3_SECT=1..5`)+ NPC dump:

| sec | 大小 | 內容 | NPC 數 |
|---|---|---|---|
| 0 | 42×43 | 阿里阿罕**戶外**(草地+建築門 tile)| 24 |
| 1 | 16×12 | 室內房間(床+桌)| 4 |
| 2 | 11×15 | 室內房間(直向)| 4 |
| 3 | 9×8 | 小室 | 2 |
| 4 | 13×13 | 室內房間 | 3 |
| **5** | 31×32 | **阿里阿罕城(金柱+王座,城堡內裝)** | 13 |

**未解(faithful 觸發點)**:哪一間是露依達酒場、酒場主 NPC 座標——**從 tile 佈局無法可靠分辨**
(室內都是紅地磚房間)。酒場主是**特殊功能 NPC**(踩到/調べる → 開酒場/登錄所選單,非一般對話),
其派發在城鎮 examine `0x9f7f` 的寫死座標 / 商店 dispatch,需逐一 RE 該分支(或 DOSBox 實機踩門確認)。
remake 現以 **T 鍵**(阿里阿罕內)開酒場 modal 作可達的暫代;換成真實店門待此 RE。

> 下一步(擇一):① RE `0x9f7f` 內 shop/酒場/登錄所的 NPC 型別或座標 dispatch;
> ② DOSBox warp patch 直接 spawn 進酒場 section 後讀當下 `[0x256a]`(section)+ 觸發 tile 座標反推。

## ★ 露依達酒場定位(腳本 + 地圖 metadata,非人工猜)— 已解

| 來源 | 結論 |
|---|---|
| **腳本** D3TXT01 rec49 | 「鎮上**西方**,有個路依達酒吧,在那裡可以尋找同伴」→ 酒場在 CTY00 sec0 **西側** |
| **轉場表** sec0 +0xc | **門(8,14)→CTY0.sec2** → 上 2F **預存所**(門在 x=8 西側,符合) |
| **NPC 表** sec0 +2 | 西側 (8,17) 櫃台店員(2F 門 (8,14) 正下方)= 1F 登記隊員店員 |

**結論**:
- **酒場 1F(登記隊員)= CTY00 sec0 西側 LUIDA 建築**,跟 (8,17) 櫃台店員調べる開創角。
- **2F(預存所)= CTY0 sec2**,從 sec0 門 (8,14) 上去。
- 地圖左上角磚牆排成 **LIH**(開發者彩蛋,見 repo README)。

remake:`DQ3_LUIDA_CTY/X/Y`(`dq3_inventory.h`);run_game 在 CTY00 對 (8,17) 櫃台調べる →
`tavern_modal` 創角流程(取代 T 鍵暫代,T 鍵保留為開發捷徑)。全程由腳本 + metadata 推導,未靠肉眼。
