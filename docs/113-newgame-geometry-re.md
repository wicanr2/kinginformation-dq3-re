# 開場／創角幾何反組譯與對拍

2026-10-01 現行狀態：命名導航與有限繪圖均已達下方限定 CONFORMED；
主選單、初始注音與兩條導航共 12 張正式全畫布 RGB 零差異。
早期 DRAFT／5,508、8,025 差異／「production 尚未變更」段落保留形成史，不能覆蓋現況。

本文件封存 `interface.json.new_game_geometry` 的來源與限制。初版的
`checkerboard_1px`、`solid_2px` 與「尚未 V3」段落都保留為時間序列；2026-08-12 的
現行 canonical 幾何是 `beveled_2px`、record 407 的十三個具名能力欄位與三層 raw EGA
backdrop，固定能力確認畫面已達 V3 靜態對拍。完整現況與殘差見
[`docs/126`](126-newgame-confirmation-v3-static-comparison.md)，不要以本文件早期段落覆蓋。

## 2026-10-01：dosgolem 正式主選單／初始命名差異（DRAFT）

工作依據：[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)。
原版由冷啟動、IRQ1 Enter 710,000,000／731,000,000 步取得主選單及初始注音畫面；
沒有遊戲狀態注入，也尚未進行能力擲骰。正式重製 `InputState` 兩次 Confirm 的
完整 640×350 RGB 比較分別有 5,508／8,025 個差異像素，範圍為介面區域，
背景立繪逐點一致。測試入口為
`TestDosgolemNewGameMenuAndNameComparison`（`DQ3_DOSGOLEM_NEWGAME_DIR` 指向原版收據目錄）。
不得裁切或遮罩差異後宣稱通過。

工具缺口已分開處理：dosgolem 副本的 BIOS `AX=1013h` 色盤頁面服務依
[RBIL 契約](https://fd.lod.bz/rbil/interrup/video/101013.html) 與
[DOSBox Staging 實作](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/ints/int10_pal.cpp)
補足，兩種頁面模式與 DAC 映射的紅綠測試及完整 DOS／machine 回歸通過。
上游唯讀。磁碟檢查點未保存 CRTC 等顯示狀態，續跑畫面異常不能當作 remake 反證；
本次正式畫面比較只使用冷啟動生成的收據。

新增非破壞 IDA 9.4 匯出入口
[`tools/ida_dump_newgame_entry_windows.py`](../tools/ida_dump_newgame_entry_windows.py)，
每筆保留原始函式名、IDA linear、MZ file offset、bytes、推論等級、證據及輸入 hash。
未審查條目一律醒目列 unknown；不能以這份匯出直接命名 production JSON 欄位。
輸入仍為本文固定 SHA-256 的 `assets_raw/DQ3.EXE`。目前僅完成差異重現，
外框、標題清底與游標 writer 的修正規格尚未 READY，production 尚未變更。

原版收據生成來源為
[`tools/dosgolem_newgame_probe.py`](../tools/dosgolem_newgame_probe.py)；
有界 Docker／Git 控制入口為
[`tools/verify_dosgolem_newgame.sh`](../tools/verify_dosgolem_newgame.sh)：
`bash tools/verify_dosgolem_newgame.sh` 重生冷啟動收據並跑正式比較；
`bash tools/verify_dosgolem_newgame.sh /home/anr2/cht/dosgolem --prototype`
另驗 DRAFT，不能當作 production 通過。它只沿用既有工具 image 與工作目錄，不建立新目錄。
工具副本的色盤頁面驗證為
[`tools/dosgolem_palette_contract_test.go`](../tools/dosgolem_palette_contract_test.go)。
來源掛 `/repo` 唯讀、dosgolem 掛 `/dosgolem` 唯讀、既有 `work/` 掛 `/work` 可寫，
在 `dq3-ebiten-test:20260822-r1` 一次性 Docker 容器執行，預設無網路並以目前 UID/GID
寫入；Go 快取指向既有 `work/.gocache-test`／`.gopath-test`。新階段產物皆使用
`work/dosgolem-opening/issue4-*`，不覆寫 Issue #1 收據；工具副本保留在容器暫存區，
不改 dosgolem 上游。

### 2026-10-01 追加：writer 閉合與試作結果

這次追加保留上方正式版差異及早期尺寸斷言；**試作通過不代表 production 已修正**。
`TestDosgolemNewGameWindowPrototype` 必須明確設定
`DQ3_DOSGOLEM_WINDOW_PROTOTYPE=1`，與正式對拍測試分開執行。
主選單及初始注音命名均為完整 640×350 RGB **0 差異**；試作只改 test buffer，
沒有改正式 Game 狀態、production Go renderer 或 game-pack。

| 原始定位（IDA linear／MZ file／DGROUP） | 附加語意與等級 | caller／consumer 與證據 |
|---|---|---|
| `0x10006`／`0x1376` | `confirmed`：主選單 raw window 取址 | `lea si,ds:3D4E` → `0x1000A` 呼叫 `sub_1F4E3`；冷啟動 Enter 主選單 |
| `0x28B1E`／`0x19E8E`／`0x3D4E` | `confirmed`：主選單 `(flags=3,x_byte=28,y=150,width_byte=22,height=64,record=475)` | `sub_1F4E3` 讀 `[si+2/4/6/8/0A]`；原始 30 bytes 保留在私有 sidecar |
| `0x1FC57..0x1FCC5`／`0x10FC7..0x11035` | `confirmed`：陰影起點 raw x+1 byte、y+8；逐列旋轉 `0xAAAA` 的四平面 word AND | `sub_1F4E3`／`sub_1F590` → writer；原版與試作的立繪交界亦逐點吻合 |
| `0x1FD30..0x1FE10`／`0x110A0..0x11180` | `confirmed`：上／下 16 列、左／右 2 byte 邊帶 XOR；每列只旋轉一次 | `1FDE4` 執行 `ror bh,1`，平面回跳 `1FDE6`；原先直條試作判讀已推翻，棋盤格控制流與全畫布比較提供反證 |
| `0x211B6..0x2121B`／`0x12526..0x1258B` | `confirmed`：16×16 不透明字模；前景色讀 DS:`0x258F` | `sub_213C4` → `sub_211B6` → `sub_21B98`；record475/451/452/456 的框線及空白 glyph 不可省略 |
| `0x1126F..0x112B2`／`0x25DF..0x2622` | `confirmed`：命名游標 XOR 平面 8/4；兩 byte、15 列 | `sub_10E55`／`sub_11087→sub_1123C`；姓名 `(248,62)` 與字盤 `(168,94)` 初始游標吻合 |

上述等級只適用本文固定 EXE 與這兩個靜態玩家狀態。完整 palette 初始化 writer 仍為
`unknown`；冷啟動收據可直接觀察 palette index8=`(255,223,255)`、框線 XOR index5，
`FIRST.SCR` 保留的 index8 則為黑，不能直接沿用背景 palette 當前景色。
能力擲骰、游標閃爍相位、其他命名操作、性別及母親流程尚未對拍。

word AND 的 latch 依據為 dosgolem `Machine.Read16/Write16` 與
[DOSBox-X 的 VGA_UnchainedVGA_Handler](https://dosbox-x.com/doxygen/html/vga__memory_8cpp_source.html)：
兩次 byte 讀取完成後 latch 保留第二 byte，兩次寫回共用該 latch。逐像素 AND 的早期
試作剩 22 個立繪交界差異；保留 word 讀寫語意後為零。這是可重播執行器／成熟模擬器
契約，不擴大聲稱所有實機顯示卡的匯流排行為逐週期一致。

非破壞匯出工具已追加上述逐筆語意索引；其餘指令醒目保持 `unknown`。直接 xref 未涵蓋
這些 DS 相對取址，保留 `entry_range` 與原始運算元，不能用空 xref 判斷沒有 caller。
私有 `dq3-newgame-cursor.json` 審查前快照為 877,844 bytes、SHA-256
`4826be95b82b6e4a2132cdd546d5426edd2935a2479a68a6904857abb801ba94`，工具 IDA Pro 9.4；
語意索引的後續重生快照另記 hash，不覆寫這份形成史。
審查後非破壞重生的 `dq3-newgame-reviewed.json` 為 1,080,283 bytes、SHA-256
`79f6b2c82be70d6f0d4dc86372ec3568171f269c55d55ca9899eea297ca8a133`；
每筆補有輸入路徑、大小、hash、原始指令大小，已確認與未知條目分級顯示。
追加 raw window 與 DS 相對運算元索引後的 `dq3-newgame-reviewed-v2.json` 為
1,080,574 bytes、SHA-256 `519186d2f3320e1de0e1c9db5bf3b9e4556e4f246cf133464bbea99aa52311ca`；
匯出自動附上已確認語意及 palette 初始化 writer 未知的限制。
收據重生工具會先把上一批 `issue4-keylog-*` 以內容 hash 存為同目錄的
`issue4-archive-<SHA256>.<副檔名>`，供回查形成史；不移動原始輸入、不加入 Git。

下一閘門是把已驗證原始 record／window／色盤規則寫成有限且可驗證的 pack 欄位，完成
DRAFT → READY 審查後接入正式 renderer；再由正式對拍入口重跑。不可把測試中的 raw
數值複製為 production Go 常數，也不可把試作 buffer 合成當成正式 UI 完成。

### 2026-10-01 追加：命名導航與功能切換規格（DRAFT → READY → CONFORMED）

此切片由 [Issue #4 的導航紀錄](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5928028713)
追蹤。輸入沿用上述 `assets_raw/DQ3.EXE` 路徑、115,282 bytes 與固定 SHA-256；
IDA Pro 9.4 位址仍為 linear，MZ file offset=`linear−0xEC90`。
執行器為 dosgolem `2f44a68ebfc54b28fb15dd4a34510b0b04a5415d` 的有界唯讀來源副本，
Docker image 為 `dq3-ebiten-test:20260822-r1`。所有收據在既有 `work/dosgolem-opening/`，
原版圖像、色號、狀態及 IDA database 不加入 Git。

| 原始定位 | 附加語意、等級與來源 |
|---|---|
| `sub_11087`，linear `0x11087..0x11172`／file `0x23F7..0x24E2` | `strong`：英數 `sub_10E8E` 與注音 `sub_10F5B` 共用導航；IDA database 匯出含 caller xref type17。方向分支見下列原始指令 |
| linear `0x110C5..0x110DE`／file `0x2435..0x244E` | `strong`：上移先減 DGROUP `0x2702` 欄數，負值加 DGROUP `0x2704` 格數；`2b0e0227`、`030e0427`，保留原始 `word_274D2`／`word_274D4` |
| linear `0x110E8..0x11102`／file `0x2458..0x2472` | `strong`：下移加欄數，達格數則減格數；`030e0227`、`2b0e0427` |
| linear `0x11109..0x11125`／file `0x2479..0x2495` | `confirmed`：左移 `dec cx`（`49`），負值加全格數；冷啟動 raw0→raw44，沒有列內環繞 |
| linear `0x1112D..0x1114A`／file `0x249D..0x24BA` | `strong`：右移 `inc cx`（`41`），達格數則減全格數；不限制同一列 |
| `word_274CE`，linear `0x274CE`／DGROUP `0x26FE` | `confirmed`：原版 raw 游標。原版 IRQ1 左鍵、上鍵、左鍵及 PNG／唯讀 runtime 觀察；實際 DS=`0x15ED`，不用探測器固定的 `ds:` 別名 |
| `word_274CC`，linear `0x274CC`／DGROUP `0x26FC` | `confirmed`：本切片觀察注音=1、功能焦點=5、英數=2；功能切換後 raw35 保留。上、左、Enter、Enter 冷啟動實驗與 PNG；其他旗標不外推 |

typed input 為既有 `InputState.DirEdge` 的下／上／左／右與 Confirm／Enter。
只修正格盤導航及功能列第零項模式切換：raw 範圍 `0..44`；上下加減欄數後在全格數內
環繞；左右加減一後在全格數內環繞。從 raw0 上、左到 raw35，才對應欄優先語意
cell43；Enter 進五列功能，Enter 切英數時保留 raw35。模式切換不能偷偷選取字元、
改姓名、觸發完成或消耗亂數。既有 Tab／情境鍵的便利操作不是原版按鍵證據。

垂直鏈為原始 `sub_10F5B` 的欄數／格數 writer → `sub_11087` → raw 游標 →
`sub_1123C`／`sub_1126F` 畫面位置，以及 raw→語意 cell→功能入口。
重製經既有 pack 的 45 格文字／幾何與共用 `NameInput`，同時影響正式新遊戲及酒館命名。
沒有新增資料欄位、raw ID、座標或文字；存檔格式不變，游標是 modal 暫態。
驗收須涵蓋原版四方向邊界、功能切換的實際觀察值、重製兩個正式入口、產生姓名後的
存讀檔與既有正常主線。錯誤方向及越界輸入維持既有保護，不新增版本預設。

探測勘誤：首次左、左、Enter、Enter 的 raw43 是語意 cell39 的聲調，實際出現
「查無此字」並返回注音盤。該次 `function-cell`、`function-focus`、`alnum` 檔名有誤，
原始收據與內容 hash 存檔保留；不能當作功能／英數 oracle。新探測用
`DQ3_NEWGAME_PROBE_SCENARIO=name_navigation` 的正確候選名稱，功能入口另用
`name_function_mode`；生成入口仍是 `tools/dosgolem_newgame_probe.py`。
第一次正確功能探測的觀察斷言仍猜 raw0／mode1，已被原版 raw35／mode5 反證，
歸為驗證腳本問題；不改原版狀態，修正斷言後以同一工具鏈乾淨重跑。

停止線：此規格不包含候選空窗規則、能力擲骰、游標閃爍相位或逐像素繪圖修正；
正式畫面仍有上方 5,508／8,025 差異，不能因導航通過升格為 V3。
能力擲骰前尚未比較 `RND()`，不宣稱原版 campaign／完整 remake 完成。
形成史：上述 DRAFT 時 production 尚未依此段變更；等待乾淨收據及審查。

**READY 審查（正式實作前）：**同一容器、同一輸入乾淨重跑已通過。
`issue4-name-receipt.json` 為 9,671 bytes、SHA-256
`b5467fe3940cc8b0bf4a700be0d31b1988a8d168da749835732153e5aaf8f16a`，
實際游標依序 `0→44→0→36→0→44→43`，覆蓋四方向邊界；候選探測保留於收據，
不納入本次候選規則驗收。`issue4-mode-receipt.json` 為 6,893 bytes、SHA-256
`4f6758d9c7012e33e3d95e3605de03258b738a8a3462950de66d3a733d129017`，
實際 `raw0/mode1→raw36/1→raw35/1→raw35/5→raw35/2`。兩者固定原版輸入 hash、
腳本 hash、Go 1.24.13、執行器 pin、實際 IRQ1 make/break、圖像／色號 hash；
沒有未實作服務、狀態注入或亂數比較。已目視核對原版右環繞及英數 raw35 PNG。

新增英數 writer 的 IDA database 匯出 `dq3-newgame-navigation-unknown.json` 為
1,119,759 bytes、SHA-256 `c09ce10f1129a79fa05534d0321dfa7aa432a0b9c6740f08e1f0f36eed0f0da3`，
仍逐筆醒目標未知；審查只提升本段已閉合導航／模式交易，其他條目不批次升格。
`sub_10DC8` linear `0x10DE8/0x10DED/0x10E4F` 的原始 bytes
`8326fc260e`／`830efc2602`／`8326fc260b` 只改模式及焦點；
`sub_10E8E` linear `0x10E8E..0x10EB3` 不寫 raw 游標，直接畫 record453 並呼叫
`sub_11087`。反向切注音由 `sub_10F5B` linear `0x10F80`（`c706fe260000`）重設 raw0；
本批保留這個既有行為，不外推雙向都保留游標。四方向表列初始 `strong` 的項目，
現在由原始 writer／consumer 與四方向正式 PNG／runtime 狀態閉合，限定此 45 格盤為 `confirmed`。
右、上、下及英數焦點已無會影響此修正的未知；純導航演算法及既有功能切換可以正式實作。
正式 renderer 仍另為 DRAFT。

驗收入口：`TestDosgolemNameInputNavigationComparison` 讀實際原版觀察值及檔案 hash，
以正常 `NewGame`／`InputState` 對照六個方向與四個功能輸入；
`TestNameInputModeFunctionPreservesRawCursor` 另隔離模式重設回歸。
完整 `TestOpeningProductionInputTrace` 必須讓主角及酒館登錄三人以功能列切換英數，
不用 Toggle 捷徑，再正常輸入姓名、存讀檔並走到 THE END。這是玩家流程回歸，
後段原版亂數／campaign parity 不因這條重製 trace 通過而成立。

**CONFORMED（只限本段導航／功能切英數）：**修正前正式對拍呈紅燈：原版左移 raw44，
重製卻為 raw8；原版上、左為 raw35，重製為 raw44；隔離模式測試另重現 raw35 被重設0。
正式修正只改共用 `niMoveRaw` 全格盤運算與功能切英數保留既有 raw 游標，沒有新增
pack 值、資料格式、raw ID、座標或玩家文字。修正後六次方向及四次功能狀態均吻合原版，
不選字、不完成、不消耗 RNG。受影響 12 項頂層測試及兩個原版收據子測試通過。

正式主線的主角與酒館三人改用上、左、Enter、Enter 進英數，不使用 Toggle；
姓名、性別、各段存讀檔及正常玩家路線到 THE END 通過（69.58 秒，執行前固定
remake `0x1357`，未重新設種）。其餘 game 為 356 項頂層及 26 項子測試通過；
35 項選用擷取／收據測試在該標準批次跳過，其中本切片導航收據已另行明確執行。
沒有素材缺失，全部 internal 與 desktop `main.go` 建置通過。
已重生並目視核對正式導航／英數 PNG；格位狀態一致，外框／字色／游標繪圖仍有
已知差異，沒有 V3 聲明。

可重現控制入口追加 `--navigation`：
`bash tools/verify_dosgolem_newgame.sh /home/anr2/cht/dosgolem --navigation`，
順序冷啟動重生兩條原版收據，再執行正式狀態比較；整條命令回傳0。
最新原版導航收據為 9,677 bytes、SHA-256
`0f4a2c60ea17f5686a17a53dab609d281fe526ca21ccc7ed33bc1ec2928072a3`；
功能收據為 6,899 bytes、SHA-256
`3f45aba19f45c82f293b3be0ef781c0d9aa4a337027aa752d75076d62355d14d`。
READY 時的兩份收據已依內容 hash 在同一目錄存檔，沒有覆寫形成史。
生成工具現在連同原版日誌存檔，延伸情境的 750,000,000 步圖像標為 `initial`，
避免與最後 state 的時點混淆；不把磁碟 state 當作正式畫面來源。

審查後 IDA 匯出 `dq3-newgame-navigation-reviewed.json` 為 1,121,400 bytes、SHA-256
`4288bba262aa9fd8cc5518227c79d3ba019922a16f33d7f17957a0a55b57f6d5`，
12 筆已審查導航／模式定位由受版控索引自動附註，仍保留每筆原名、位址、bytes、
推論等級與來源；其他項目醒目保留 unknown。
驗收收據 `work/dosgolem-opening/issue4-navigation-verification-receipt.json` 為
4,890 bytes、SHA-256 `11ba6c16ca3d5b9383acbfb60f947f2b86e7d7841d66266b3b6d5d9fb1a3551f`，
保存實際 image ID、程式／測試／工具 hash、原版收據 hash、seed、輸入、日誌及範圍。
所有本批檔案為 UID/GID1000:1000，沒有新增誤掛載 `.md` 目錄；
3,636 個歷史 root-owned 項目沒有廣域改動，一次性容器均已清理。

下一個最小閘門仍為上方繪圖 DRAFT 的有限 pack 欄位與正式 renderer，
不是重新研究已閉合導航；原版能力擲骰、出生點與母親保持 unknown，Issue #4 保持開啟。

### 2026-10-01：正式主選單／命名字盤繪圖契約（DRAFT → READY）

本切片沿用上方固定 EXE／TXT／FON、IDA 9.4 位址契約及冷啟動收據，
處理主選單、初始注音、六次方向導航、raw35 功能焦點及英數盤的正式完整畫布。
typed contract 是 `new_game_geometry.raster`：三個具名視窗 role（menu／header／mode）
各引用既有 raw window ID 與完整文字 text ID；另有注音／英數 grid text ID、grid origin、
menu／function cursor 與 hit rect、陰影位移、框線邊帶寬高、字模色號、框線／游標 XOR
遮罩及有限 palette overrides。它描述資料與引用，不承載繪圖命令串或任意 JSON code。
新欄位採 schema `0.1.55`／content `0.1.61`，缺值、未知 window／text、越界或非法
控制字一律由 loader 拒絕；初始化需確認字模與背景素材完整，不能猜預設值。

證據審查：menu raw `0x28B1E` 的 30 bytes 及 record475 已由原版與試作全畫布閉合；
header raw `0x29088`／record451、grid452/453、mode raw `0x290C4`／record456 依
原始 caller→writer 及冷啟動 PNG 閉合。字盤 origin `(152,94)`、姓名 `(248,62)`、
格盤 `(168,94)`；word shadow offset `(8,8)`，frame 邊帶16px；font index8、frame XOR5、
游標 XOR12／16×15，palette index8=`(255,223,255)`。上述為限定狀態 `confirmed`。
完整 palette 初始化 writer 與閃爍相位仍未知，不由靜態圖外推。
function raw `0x290A6` 的 cursor words +24/+26=`41/94`，經 `sub_1F908` 為 `(328,94)`，
row step16，selected glyph11／blank12；menu cursor `(240,166)`。`1F93F` 的原始
`mov ax,[si+6]; sub ax,4` 為 row 區域寬度，byte 投影8，menu144px／function64px；
同列兩個 cursor 皆由既有 pack `choice_cursor` 字模提供。
模式焦點 PNG 證實 grid cursor 移除、功能 glyph11 出現、下方 mode window 關閉；
切英數後 grid raw35 游標恢復，下方 mode window 不再出現。

實作固定順序：唯讀背景複製 → word-latch 陰影 → 完整 record 不透明字模 →
flags bit2 控制四邊 XOR → menu glyph cursor；命名則 header → 模式 grid →
姓名字元／姓名 XOR cursor → grid XOR 或 function glyph cursor，注音且無功能焦點才畫
mode window／既有組字欄。title palette 使用複製後有限 override，不修改背景或 pack。
正式 `InputState` 與姓名狀態機保持前批導航規格，不加入 debug 入口。

候選字窗、性別及能力確認維持既有獨立繪圖路徑，不把尚未閉合的候選行為猜進本契約；
酒館沿用既有場景色盤／繪圖，抽測其正式命名與存讀檔，不把 title palette 外推到城鎮。
共享姓名／格盤 anchor 的一像素及標題位置勘誤由原始字模／writer 修正，舊幾何斷言保留
在本文形成史。存檔格式不變，pack hash／版本照既有流程更新。

驗收：原始 EXE 的三個 raw window 與 cursor bytes、原始 TXT 完整 records 維持 Go decoder
parity oracle；reference／邊界／缺值拒絕；正式 NewGame／InputState 在完整640×350 RGB
比較 menu、初始 name、導航、功能 focus 與 alnum，不裁切或遮罩；背景及 palette 不變測試；
姓名／酒館／存讀檔及正常主線，完整 game／internal／desktop。已審查證據足以描述上述
輸入、順序、邊界、輸出與驗收，**此限定契約 READY，准予正式實作**。
輸入 `D3TXT00.TXT` 18,680 bytes、SHA-256
`38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b`；
`D3TXT00.FON` 47,232 bytes、SHA-256
`c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`。
原始資產與 PNG／state／IDA database 仍僅存本機；本切片沒有原版能力 RNG、母親或完整
campaign 完成聲明。新增通用繪圖入口 `game/indexed_window.go` 與原始資料驗證
`internal/gamepack/newgame_raster_test.go` 均由本段及 docs/84 掛入索引。

### 2026-10-01 追加：正式繪圖限定 CONFORMED

上述 READY 後，正式 `NewGameWithPack` 編譯索引色繪圖設定；`NewGameFlow.draw`
透過同一正式 menu／name 狀態繪圖。沒有用試作 buffer 或直接切入事件。
完整字模記錄、視窗、色號、陰影、框線與游標設定均由資料包提供，
schema `0.1.55`／content `0.1.61`；缺引用、字模、邊界或 D3 證據均拒絕載入。
既有組字映射抽成共用輸出，沒有另增一份聲母／介音／韻母常數表。

| 正式玩家狀態 | 全畫布 RGB 差異像素 |
|---|---:|
| 主選單／初始注音命名 | 0／0 |
| 六次方向：左、右、上、下、左、左 | 每一步 0，共六張 |
| 上、左、Enter 功能焦點、Enter 切英數 | 每一步 0，共四張 |

比較入口仍為 `TestDosgolemNewGameMenuAndNameComparison` 與
`TestDosgolemNameInputNavigationComparison`；後者現在也逐步比較畫面，
原版 artifact hash 與模式／游標觀察仍先驗證。沒有裁切、遮罩或挑選重擲。
完整有界 `production` 及 `--navigation` 入口已從冷啟動重生並返回0；
每次均保留舊收據。主選單、注音及英數 PNG 已目視核對，背景與共享色盤未被繪圖修改。

`TestNewGameRasterOriginalDataParity` 直接讀固定雜湊的 EXE／TXT，
核對三組 raw window、文字引用、五筆完整記錄與兩組 cursor／hit 寬度；
12 項損壞契約案例均被拒絕。完整 game 359 項頂層／28 項子測試通過，
33 項選用擷取／額外收據未執行（含歷史試作及既有開機收據），沒有素材缺失跳過。
全部 `internal/...` 及 desktop `main.go` 建置通過。正常新遊戲主線含主角、
酒館三人正式功能列命名與各段存讀檔，抵達 THE END（109.47 秒）。
重製 seed 在首個玩家輸入前固定 `0x1357`，沒有中途重設；原版全部收據在能力擲骰前，
所以此處**沒有原版 RND 對拍**。

私有收據 `work/dosgolem-opening/issue4-raster-verification-receipt.json`：
13,128 bytes，SHA-256 `4f9fe517d6d0f23fa158f5f3bd6a55d5900620dd11069e876a76d59d21aa8e73`。
保存固定原始輸入、dosgolem revision、Go／IDA／Docker 版本、實作來源 hash、
12 組原版／正式 PNG hash、日誌及範圍；原版／重製皆沒有狀態注入。
所有原版圖像、state、database 與完整收據仍只在本機，不加入 Git 或公開包。
本次 CONFORMED 限於上表狀態；候選字、性別、能力亂數、出生點、母親、
閃爍相位、音訊與完整原版 campaign 仍待各自閉合。下一步由 Issue #4 繼續追原版創角後路線。

## 輸入與工具

以下保存 2026-08 的形成史；當時 image 的 IDAPython 限制不能覆蓋上方
2026-10-01 由 `ida-pro-9.4-idapython:locked-v1` 重生的 sidecar 與目前工具入口。

- 輸入：`assets_raw/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4（`/home/anr2/ida_94_official/dist`，在一次性 Docker image
  `ida-pro-9.4-ver3:latest` 執行）；database 與批次 listing 留在 `/tmp`，未加入 Git。
- IDA listing：`/tmp/dq3-ida-batch.asm`。IDA linear range `0x10000..0x2aee0`，
  `seg0`／logical offset 與 file offset 不可混用；本文件的 `linear:` 標籤就是 IDA
  linear address。
- 可重生輸出：listing header 同時記錄輸入 SHA-256、format、base address 與 entry
  point；IDAPython 在此 image 缺少 host Python 3.14 library，故本批只採用 IDA
  auto-analysis／IDC 匯出，不宣稱有 IDAPython sidecar。

## caller → writer → consumer

### 姓名／注音輸入

`sub_10854`（logical `0x0854`）呼叫 `sub_10d17`（`0x0d17`）建立姓名 modal；
`sub_10d17` 先 `lea si,byte_29088` 呼叫 `sub_1f590`，再依 `sub_10dc8`／
`sub_10e8e`／`sub_10f5b` 分派功能列、英數與注音盤。

| IDA linear raw 結構 | raw 欄位 `(flags,x,y,width,height)` | consumer／用途 |
|---|---|---|
| `0x29088` | `(1,19,46,32,144)` | 姓名主窗；`sub_1f590` → `sub_1fb36` EGA frame writer |
| `0x290a6` | `(1,41,78,12,112)` | 五列功能列；`sub_10dc8` → `sub_1f779` navigation |
| `0x290c4` | `(1,19,190,12,48)` | 注音盤／切換窗；`sub_10f5b` → `sub_1f590` |
| `0x290e0` | `(1,19,190,34,48)` | 組字候選窗；`sub_112e4` → `sub_1f590` |

`sub_10f5b` 明確寫入 `word_274d2=9`、`word_274d6=0x15`、`word_274d8=0x5e`；
`sub_1123c` 將欄位換成 `x=0x15+2*column`、`y=0x5e+0x10*row`，所以玩家可見
姓名盤是 9 欄、16px 水平／垂直步距。`sub_11172` 同樣把游標位置交給滑鼠／
`sub_1123c` consumer；這不是依英數字串長度猜出的排列。

`sub_1f590` 讀 `byte_29088+0x0a=0x1c3`（D3TXT00 rec451），其 raw 起點
`(0x13,0x2e)` 使標題 anchor 為 `(233,46)`；`sub_10e55` 的 `di=0x1f+2*len`、
`dx=0x3e` 使姓名列文字 anchor 為 `(249,62)`，游標每字 16px。rec452／453 的
逐列 glyph、五列功能文字與 rec456 組字提示已另存 `new_game_labels`；raw35
（欄優先 `cell=col*5+row` 的 cell43）才設定功能列焦點，不能將 45 格壓成舊的 38 格。

### 性別／能力確認

`sub_10854` 在姓名完成後依序呼叫 `sub_1f4e3`（性別選擇）、`sub_1f590` on
`byte_28b78`（能力主面板）、`sub_1f590` on `byte_28b92`（提示／選項），並由
`sub_1f63c` 重繪背景；因此三組 raw window 不能合併成一個全螢幕 modal。

| IDA linear raw 結構 | raw 欄位 `(flags,x,y,width,height)` | 截圖中量到的 640×350 rect |
|---|---|---|
| `0x28b78` | `(3,19,46,44,192)` | 左能力 `x=159,y=52,w=145,h=98`、裝備 `x=159,y=148,w=145,h=82`、右能力 `x=303,y=52,w=194,h=178`（最終 V3 靜態 pack） |
| `0x28b92` | `(3,45,14,22,48)` | `confirm_prompt: x=367,y=20,w=162,h=34` |
| `0x28bc6` | `(3,43,46,12,64)` | `confirm_choice: x=367,y=68,w=98,h=50` |

截圖來源（均為 DOSBox 原版 1024×768 擷取；左上 640×350 是邏輯畫布）：

- `docs/36_shots/04_name_input_zhuyin.png`，SHA-256
  `e8b4e4dd4975d45815b86275585d73ae085754a0b218aa2a6b2e36748cfdff7a`；姓名主窗
  框線量測為 `(159,52)..(400,181)`，功能列分隔在 `x=319/320`，注音模式窗為
  `(159,196)..(240,229)`。
- `docs/36_shots/07_gender_select.png`，SHA-256
  `abbdfc0b3fba317e9a0823e9fdc4ac7b483faebea6c039597e25a96d7e178288`；性別 raw
  window 的同狀態框線 anchor 取 `(344,46)`，文字起點 `(368,62)`，列距 16px。
- `dosbox/v3_01_afterstats.png`，SHA-256
  `0b55c114a44a9ff5b83f1601425b82cc607ffb1ef3d92c5f123dd047e73b8695`；能力面板
  三分割與提示／選項 rect 由調色盤 index 13 的水平／垂直線段量測，不以目前
  remake screenshot 反推。

## JSON 與推論等級

`interface.json.new_game_geometry` 同時保存像素 rect、文字／格盤 anchor、共用外框
RGB、`border_pattern` 與上述 `raw_windows`。`new_game_labels` 另保存 rec451–456 的逐列 45 格與五列
功能文字；
`NameInput` 以 `raw35 → cell43 → function list`、第五列完成的正式輸入路徑消除舊的
直接完成捷徑。整個物件目前標為 `D2`：raw window writer、格距公式、字模 record、
外框 RGB 與 `bh=0xaa` 邊線遮罩已交叉吻合，但尚未為每一個 panel 建立同一輸入序列的
逐幀 DOSBox capture，也尚未閉合 `confirm_choice` 的藍色選擇圖樣。因此 production 可
使用這些幾何與字模，但不能宣稱創角畫面 V3。

能力確認右欄的欄位語意另經 `dosbox/v3_01_afterstats.png` 與 record 407 逐 glyph
解碼校正：六列是「運氣點數／最大HP／最大MP／攻擊力／守備力／經驗」，不是早期
renderer 使用的「速度／HP／MP」。`agility` 只屬詳細狀況窗 role；欄位與 glyph
對映見 [`docs/112`](112-newgame-labels-re.md)。這項修正仍是 D2／V2，不會把尚未閉合
的框線 pattern 提升為 V3。

若後續證據推翻任一矩形，應保留本文件與 JSON 的舊斷言，追加勘誤及新 sidecar；不
得移動 `raw_windows` 位址或只留下改名後的語意。

## 2026-08-10 勘誤：靜態 confirm_choice 已另立資料契約

上文保留的是本文件原先「藍色選擇圖樣尚未閉合」時的歷史狀態。本輪以同一正式姓名／
性別輸入重跑 DOSBox 穩定畫面後，將可見但不涉及逐幀時序的部分拆成三個 pack 欄位：
`confirm_choice_backdrop`（`x=360,y=62,w=112,h=64`、`RGB(0,85,223)`、奇偶 phase）、
`confirm_choice_content`（`x=376,y=78,w=80,h=32` 的黑色內容區）及
`confirm_choice_frame`（`x=367,y=68,w=98,h=50`、`checkerboard_frame_2px`、
lavender／膚色 accent）。能力確認右欄六列的實際 glyph 起點也校正為
`y=126,142,158,174,190,206`；舊 JSON 的 `y=96` 會把文字畫入選擇框，已由
`interface.json` 與 parity test 修正。完整 raw／像素證據與推論等級見
[`docs/118`](118-newgame-choice-backdrop-re.md)。

這項勘誤只把穩定畫面的幾何、色彩與可重現 frame primitive 接入 runtime（D2／V2）；
palette register 切換、游標閃爍、能力條淡入及逐幀動畫仍保留原文件的 V3 限制，不能把
靜態 PNG 對拍擴大解讀成整段創角演出已完成。

## 2026-08-12 現行勘誤：固定確認頁 V3 靜態完成

前段最後一句的「D2／V2」是當時的暫時狀態，已由較強的同輸入畫面比較取代。現在的
`interface.json` 是唯一 canonical owner：

- `stats_name=(200,46)`、`stats_hero=(216,62)`、`stats_sex=(168,78)`、
  `stats_sex_value=(248,78)`、`stats_cloth=(200,158)`；
- 13 個 `stats.*.{label,value}` 取代一維 row anchor，值欄保存 `digits`；
- 左／裝備面板以 `frame_edge_widths.right=1` 描述 shared seam，右面板寬度為 194；
- creation／prompt／choice raw backdrop 依 `draw_order=0/1/1` 在 foreground 前後合成。

這些值由 `sub_10854 → sub_1834E`、raw windows、EGA writer 與同狀態 PNG 共同約束。
完整 hash、AE=1,474、反證與 V3 範圍在 [`docs/126`](126-newgame-confirmation-v3-static-comparison.md)。
逐幀游標、palette register 和整段創角 timing 仍是獨立的未完成工作。
