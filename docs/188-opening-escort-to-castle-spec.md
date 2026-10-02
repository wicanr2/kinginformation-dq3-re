# 188 — 開場連續演出勘誤與修正規格：家中 → 王城入口 → 國王

## 2026-10-02 最新結果：生日保留前文與四次捲動限定 CONFORMED

依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)，下方生日續頁規格
已完成 RE→DRAFT→原型審查→READY→正式實作→同狀態驗收。
正式從創角與第18次確認自然抵達末尾及四次捲動，六張完整640×350 RGB差異均為0，
沒有裁切或遮罩。其中五張是新增比較點，連同既有22張累計27張限定可見對拍。
FFFC保留前文，最後一字等待後四次捲動，自動返回，再交易出生場景。
捲動中不能提前確認，沒有RNG消耗；返回後自然抵達record83。時序只採平台規格近似。

最後冷啟動原版收據 `work/dosgolem-opening/issue4-birthday-pages-receipt.json`：36,603bytes，
SHA-256 `13c5e2632d3fd874f42b21ce061df16b7d3f6681846322ca8c59b79cb21c756f`。
與原型前冷啟動的66份PNG／bin全部相同；收據差異只在log artifact，38次IRQ1送達，
19次輸入與執行前一次固定0x1357均維持。原始圖像、腳本與archive留在本機。
正式紅測試仍保留：房間差172,261像素，第19次兩側不同狀態差198,652僅供診斷。
本段沒有把房間試作302像素的結果接入正式程式，也沒有寫死NPC步行影格。

schema/content為0.1.60／0.1.66，`opening_prelude.text_flow`提供具名保留文字原語與捲動參數。
實作為 `game/dialogue_retained.go`，正常輸入對拍為 `game/opening_retained_test.go`；
原始EXE parity、缺欄位／null／未知欄位／錯誤幾何拒絕見 `internal/gamepack/gamepack_test.go`。
自動合併台帳已追加本段有限語意，重生 sidecar `work/issue4-retained-reviewed-ida.json`：
849,255bytes，SHA-256 `f5858634c360f87a146348840d19b4383833d667c4475f4974cb61d75bc814e2`。
IDA9.4、原始EXE路徑／大小／hash／兩種bytes及原始位址均保留，舊sidecar不覆寫。

完整game362項頂層／34項子測試、11個internal套件及desktop main.go通過；正常新遊戲
到THE END85.55秒，主角／酒館與各段存讀檔通過。這只屬重製回歸。
標準批次38項選用SKIP，創角與本生日對拍另行嚴格通過，接受角色對拍因房間差異仍RED，
其餘35項未執行；沒有素材缺失。初次完整重播於末字揭露後立即按確認，尚未抵達
原版FFFC等待點；已修正驗證腳本以正式空白InputState等到等待，再用同一工具鏈乾淨重跑。
下一步為房間視野／外界圖塊／陰影的正式資料契約與動畫時間條件，不重新開啟已通過生日切片。

整批私有稽核入口 `work/issue4-retained-audit.py`，收據 `work/issue4-retained-evidence-receipt.json`：
21,932bytes，SHA-256 `3432cd48b3cb2894538189264d6b779f0ac6c303d6a57dc4a6574551dafc6713`。
769筆原始file bytes、全部當輪原始產物、正式來源hash及UID/GID1000通過；歷史root候選3213保留，
新增root及Markdown誤掛載目錄0。相關一次性Docker容器均清除，沒有新image或發行包。

## 同日較早證據：房間背景與陰影來源閉合，完整渲染仍 DRAFT

工作依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)，正式基準
`8cebf4a983e32cc55132c7830604fe4cb4e0b55b`。本批只新增唯讀原版觀測與有限語意索引，
沒有更改正式 Go 或資料包，schema/content 仍為0.1.59／0.1.65。
正式第18次輸入抵達record83等待，完整640×350 RGB仍差172,261像素；第19次兩側狀態不同，
198,652像素只供診斷。生日返回／初始clock的有限CONFORMED仍成立。

### 已審查的原始定位

輸入為 `assets_raw/DQ3.EXE`，115,282bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
工具為IDA Pro9.4，位址空間為IDA linear，MZ file=`linear−0xEC90`，DGROUP基底linear`0x24DD0`。
資料庫、EXE及上游dosgolem均未修改；每筆匯出保留原名、原始定位、file bytes、loaded bytes、xref、分級與出處。

| 原始定位 | 附加語意 | 等級與界線 |
|---|---|---|
| linear `0x11971..0x11991`／file `0x2ce1..0x2d01`，`sub_11900`之後的相鄰資料庫項目 | 房間視野為玩家減9、7，20×15格，未夾住地圖邊界 | strong；保留函式邊界，隔離畫布核對，尚未接入正式路徑 |
| linear `0x130f4..0x1311e`／file `0x4464..0x448e` | section+0x12經DH寫入DGROUP0B2D；家中值為71 | confirmed只限raw欄位及writer，與自然flow的raw0B2D=71閉合 |
| linear `0x11dd8..0x11ddc`、`0x11e47..0x11e4b`／file `0x3148..0x314c`、`0x31b7..0x31bb` | layer=0的X／Y界外圖塊讀取DGROUP0B2D | strong；不外推非零layer的0B57替代路徑 |
| `sub_1FC57` linear `0x1fc57..0x1fcc6`／file `0x10fc7..0x11036` | x原始byte+1、y+8、完整width／height的陰影；逐行旋轉AAAA遮罩，再做16位元AND | strong；只採dosgolem字組鎖存契約，未宣稱實機逐週期一致 |
| `sub_11ED0` linear `0x11ed0..0x11eee`／file `0x3240..0x325e` | NPC朝向與動畫位元選圖庫指標；bit7可略過動畫位元 | strong；自然繪圖時的索引與指標已取得，跨兩側的動畫時間仍待對齊 |
| `sub_21B98` linear `0x21b98..0x21bac`／file `0x12f08..0x12f1c` | 14條指令僅設定VGA暫存器，沒有額外24px清底 | strong；不從helper名稱推定畫布清除範圍 |

CTY00.DAT為7,546bytes，SHA-256
`ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836`。
section4的file基底為0x1383，+0x12=0x47；loader的取址、寫入與視野consumer已連接。
NPC loader在linear`0x131bd..0x13370`／file`0x452d..0x46e0`保留可見性、raw控制欄位、
圖庫快取槽的間接寫入；圖庫槽不能直接當成BLS原始人物ID。

### 陰影勘誤與完整畫布試作

`docs/94`舊句將`sub_1FC57`稱為清內容，本批追加勘誤。`sub_1FB36`只備份背景，
`sub_1FC57`做偏移陰影，字模consumer才處理文字畫布，三者不能互換。
原始共用結構使陰影範圍為x160、y246、352×96。

dosgolem的`internal/machine/machine.go:Read16/Write16`依序呼叫兩次byte讀寫，
`internal/machine/vga.go`每次讀取更新四個平面鎖存值；陰影的AND寫入共用最後讀取的值。
因此16px字組前8px的保留色彩來自後8px，不能用單純RGBA棋盤黑點取代。
這是本次固定revision執行器的渲染契約；[DOSBox Staging的VGA讀寫來源](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/hardware/video/vga_memory.cpp)
亦提供逐byte讀寫與鎖存運算的交叉核對，未用DOSBox圖片替代原版收據。

隔離副本固定來源`4903f533e66369a78467ebcba5125de7cc283cf5`，延續既有18次InputState。
它沿用上一批的視野／外界圖塊／共用框／clock試作，沒有更改原版輸入、seed或人物姿態。
此副本不代表現行production。完整畫布沒有裁切或遮罩：

| 試作 | 完整RGB差異 | 尚未匹配 |
|---|---:|---|
| 前批視野／色盤／共用字模框 | 2,034 | NPC、陰影、箭頭 |
| 純棋盤陰影 | 544 | 兩個NPC 261、箭頭41、框底242 |
| 依字組鎖存契約處理陰影 | 302 | 兩個NPC 261、箭頭41；框底242已消除 |

入口為本機`work/issue4-room-shadow-prototype.py`與`work/issue4-room-latch-prototype.py`，
須在既有`dq3-ebiten-test:20260822-r1`容器的來源副本執行，沿用Go編譯及有trap的Xvfb；
測試名稱為`TestIssue4RoomShadowPrototype`／`TestIssue4RoomLatchPrototype`。
日誌為`work/issue4-room-shadow-prototype-poses.log`／`work/issue4-room-latch-prototype.log`。
兩項測試保留完整RGB紅結果，預期退出1；不可把試作執行成功寫成原版對拍通過。

### 人物影格的有限動態觀測

同一次冷啟動在實際圖庫consumer取指標後讀取暫存器，沒有原版寫入：

| 步數／IDA linear | 原始暫存器 | 有限定位 |
|---|---|---|
| 1,222,142,894／`0x11ee8` | BX=0056、SI=50A0、DI=6C82、raw0004=1、seed356D | 母親所在格的快取索引43 |
| 1,222,161,235／`0x11ee8` | BX=0042、SI=3DE0、DI=7456、raw0004=1、seed356D | 床邊NPC所在格的快取索引33 |
| 1,222,323,332／`0x1e30b` | BX=0000、SI=0000、DI=0000、raw0004=0、seed356D | 勇者選第0影格 |

兩個NPC的差異分別為159、102像素，對應步行影格；勇者的102像素舊分類線索已訂正為床邊NPC。
本輪中途將母親索引口述為33亦已訂正，目的位址及原始事件保留供回查。
等待時raw0004又變為1，不能拿等待點的值倒推先前勇者畫出的影格。
正式測試目前只重播鍵序，未對齊兩側逐個輸入的動畫時間；不得把原版這次影格寫死在正式遊戲。
箭頭由原版glyph13在x336、y286呈現41個可見像素；顯示／空白相位尚未進入正式契約。

使用的原始素材身分：DQ3MAN.BLS 222,726bytes，SHA-256
`823f57e0724e36ac8ed1aa472f59d2e8fb059f171e05e3aafc457e386f77158d`；
DQ3MST.BLS 115,206bytes，SHA-256
`a1a48eaf6c13ae73472d5ff77769fa538c19e048c24496d20218a89f3230f244`；
DQ31.BLK 65,286bytes，SHA-256
`5996d95d743fb8e1fe8d3ad29c513f5241af2c3574cf9f652dce57ebd6ba3298`。
FON及PAL身分沿用下節已列輸入，原始素材不入Git。

### 收據與下一閘門

原版冷啟動入口仍為
`bash tools/verify_dosgolem_newgame.sh /home/anr2/cht/dosgolem --birthday-pages`。
最新收據`work/dosgolem-opening/issue4-birthday-pages-receipt.json`為31,719bytes，SHA-256
`b077e99c3b38f58b3326df8f8d25ed72b74d58f351a71d41374bc5063db4358c`。
舊35fc及本批29fd收據按內容hash保留。27張PNG／27份色號、38次IRQ1及16筆既有flow均未改變。
生成腳本留收據，SHA-256為`ffdb0e43eb7fce280c7208d551f44ebd6cad3ca31afe4b2cd187e9ce3024f563`。
本輪wrapper退出1，仍只因既有兩項完整RGB差異；同批創角8張及生日首頁2張通過。

已審查結論回填`tools/ida_dump_opening_handler54.py`的非破壞語意索引，
新exporter SHA-256為`e138dbb9fcae439e5dadda8dd14e121d4a2c738c93983d61978a0ffebd5ae656`。
在既有IDA9.4容器用以下file target重生，舊sidecar不覆寫：

| 私有sidecar | target／額外末端 | bytes／SHA-256 |
|---|---|---|
| `work/issue4-room-section-reviewed-ida.json` | 0x443f／0x452d | 126,689／`fe03311344acdacb1f874cb0322123aea7c6554774b8bd1febf4d1a9a120305d` |
| `work/issue4-room-tile-reviewed-ida.json` | 0x30fa／0x3350 | 467,829／`e6af7820d9faac2c87e2500d5b2d540baee8d78cfb85a3af02ec245fbbb294da` |
| `work/issue4-room-window-reviewed-ida.json` | 0x10900／0x10b80 | 1,252,631／`612cde50f8cfa8fb3bca65926b3374c967b47c88bfa24414346ac4491421df43` |
| `work/issue4-room-glyph-clear-reviewed-ida.json` | 0x12f08 | 37,877／`72880c5a2d5c4497c3ac637fc2f179b5acb1ffaf1ecb4744bf0c852f01d5216b` |

`glyph-clear`僅是工作檔名，匯出語意明確標示沒有清底行為。
補充NPC／sprite／背景備份／writer候選的sidecar及原始腳本，統一由本機
`work/issue4-room-audit.py`稽核，產出`work/issue4-room-evidence-receipt.json`：6,045bytes，SHA-256
`a240c5cff84059c96c20e81fb92e6ff35198b883eb6e79881761b99f3cfc5273`。
稽核檢查完整圖像、IRQ1、既有flow、原始file bytes、推論警示、工具來源及UID/GID1000。
狀態為`draft_prototype_full_raster_mismatch`；歷史收據不可覆寫。

下一步先補共享文字流保留／捲動及兩側動畫時間條件，再審查有限資料包欄位與正式rendering入口。
不深挖PIT／ISR逐週期同步；時間採平台規格近似，畫面驗收明列實際採樣相位。
房間整體仍DRAFT，未接入production，也未宣稱正常原版campaign完成。
最近完整game／internal／桌面及新遊戲→THE END仍是下節的61.09秒批次，本批未重跑完整回歸。

## 2026-10-02：有限返回／初始 clock 修正（CONFORMED，限返回順序與初值；房間畫面仍 DRAFT）

下節完整續頁仍未 READY。本節只審查兩項可獨立實作的已證實行為：生日文字在
EOF 揭露完成後自動返回，與原版初始晝夜 clock；不把尚未閉合的文字保留／捲動、
母親姿態、箭頭、完整視野或房間像素一起升級。

原版 EXE 的 DGROUP `0x251d`／IDA linear `0x272ed`／file `0x1865d` 初值為
`1e 00`，即30。IDA Pro9.4 `sub_1EE76` linear `0x1ee76..0x1ee9a`／file
`0x101e6..0x1020a` 讀該 clock，除以20後索引既有色盤表；30選 bank0，0選 bank1。
最新 dosgolem 冷啟動19次IRQ1收據亦在生日、出生交易及房間等待觀測到 raw251d=30、
raw25d1=0x3232，seed仍0x356D；confirmed只限這條自然創角路徑及上述有限初值。
輸入 EXE 身份同下節；PAL為240bytes，SHA-256
`178a9ca809a33108d8e2a6430796eae8909d98e514fe543fc1923139110389c4`。
首輪新增flow收據已歸檔至 `work/dosgolem-opening/issue4-archive-a23584e6b08ece0ad065f3e26efe217c1d3ad5f96d4eaada9fb4664d1eff3165.json`，25,213bytes，SHA-256
`a23584e6b08ece0ad065f3e26efe217c1d3ad5f96d4eaada9fb4664d1eff3165`；前版收據已按hash歸檔。
只新增唯讀原始欄位觀測，19次輸入與固定一次seed不變。

隔離來源副本的試作以第18次正常確認、空白更新抵達record83首頁等待，seed0x356D，
沒有額外生日EOF鍵。房間試作再使用原版視野 `player−(9,7)`、CTY raw外界tile71、
共用字模框與初始clock30，完整RGB差異由153,932降為2,034像素；仍未通過。
試作來源 `work/issue4-birth-prototype.py`，紀錄 `work/issue4-birth-return-prototype.log`、
`work/issue4-birth-color-prototype.log`；不將試作中的診斷硬碼或未審查視野加入正式程式。
重生入口固定來源`4903f533e66369a78467ebcba5125de7cc283cf5`，避免正式修正破壞舊試作；
`work/issue4-birth-color-prototype-replay.log`再次以第18次輸入重現2,034像素差異。

有限契約審查：`opening_prelude.return_mode` 明確選擇具名原語
`automatic_after_reveal`或`confirm`，缺欄位／未知值失敗即關閉；本pack採前者。
內嵌`0xfffc`仍等待一次正式確認；末字完成同一hold後返回caller，不再多等一次確認。
場景交易移到生日返回後，使用已版控 `opening_escort` 的CTY、section及首frame玩家位置；
本pack值與原版caller寫入5、5一致，沒有新增Go座標或推導fallback。
`day_night_cycle.initial_clock`為必要整數欄位，範圍`[0,clock_ticks)`，本pack採30；
新遊戲初始化既有可存讀的phase／step，不改玩家之後的晝夜推進。
兩項須由原版EXE／TXT parity、正式第18次輸入狀態、正常流程回歸及存讀檔驗證。
房間完整RGB紅測試保留；READY不代表完整續頁CONFORMED。

正式實作及驗收：schema/content為0.1.59／0.1.65。第17次接受角色仍為生日黑底；
第18次確認後，只送正常空白InputState等待，末字完成既有hold後自動返回並交易房間，
抵達record83首頁等待、seed356D、clock30。內嵌等待、末字hold、早到確認不跳過末字及
EOF沒有額外確認，皆由component與正常玩家流程守門。出生落點不套用移動碰撞檢查，
原版是直接寫入床上的5、5；初版誤加該檢查的失敗日誌保留，修正後同工具鏈重跑。
兩項有限行為到E2／CONFORMED，不宣稱完整文字流或房間V3。

既有22張完整640×350 RGB零差異；第18次兩側現在同record83等待，但正式房間完整RGB
仍差172,261像素。第19次remake仍在record83尾段、原版已進後續caller，198,652像素
僅作不同狀態的診斷。保留紅測試，不能拿隔離試作2,034像素當成正式結果。
新原版冷啟動wrapper確實退出1，原因只有兩項上述RGB差異，同批創角8張通過。
最終原版收據 `work/dosgolem-opening/issue4-birthday-pages-receipt.json`，25,213bytes，SHA-256
`35fccfec8e86c629a04d91a9c0f72e7eba112d5913ab78b55f2d181ba08d9c95`。
與首輪a235收據相比，27張原版PNG／色號、38次IRQ1與16筆唯讀flow均相同；生成腳本留收據。
執行日誌另有差異，完整收據各自保留，不把日誌等同於畫面或流程一致性。

全部11個internal套件、完整game362項頂層／36項子測試與desktop main.go建置通過。
標準批次35項選用跳過，其中創角／生日兩項另行嚴格執行；剩餘33項未跑，沒有素材缺失。
正常新遊戲→THE END為61.09秒，主角／酒館與各段存讀檔通過；只屬remake回歸。
日誌：`work/issue4-birth-return-internal.log`、`work/issue4-birth-return-game-final2.log`、
`work/issue4-birth-return-comparison-final2.log`、`work/issue4-birth-return-wrapper-final.log`。
整批私有稽核收據`work/issue4-birth-return-evidence-receipt.json`，8,159bytes，SHA-256
`b537cf20fbf3c1120dcc9855d6c0430401821fbf47fb54a0f581ed10231800f5`；
入口`work/issue4-birth-return-audit.py`核對54份圖像產物、原版事件、IDA bytes與來源hash。
狀態為`finite_state_conformed_full_raster_mismatch`，只列有限CONFORMED；本輪UID/GID1000，
歷史root候選3,213無新增本輪產物，Markdown誤掛載目錄0，DQ3一次性容器均已清理。

已分級caller匯出 `work/issue4-birth-return-reviewed-ida.json`，1,435,677bytes，SHA-256
`77b720a8d8b1d1956e9184c27b210027a02cf168282497f1d3acb47181cc3721`。
有界視野匯出 `work/issue4-birth-viewport-final-ida.json`，1,603,289bytes，SHA-256
`701a83cae83e03ad5357d3a46ccf6537066614a3a40a73ca07736c42bd4367ea`。
重生入口 `tools/ida_dump_opening_handler54.py` 可在既有IDA容器以target file0x2c70、
額外末端0x3000匯出相鄰資料庫項目；完整末指令使實際file範圍至0x3002／linear11C92。
同列保留原名／定位／bytes／分級／出處，不合併函式邊界；DS相對運算元沒有直接xref
不代表沒有writer。clock初值已confirmed，視野仍strong，不從工具標籤升級完整畫面。

## 2026-10-02：生日續頁與出生時序（DRAFT，正式紅測試已重現）

工作依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)，正式程式基準
`53fb722159b25e007ef6c0498187a366eacb14c0`。首頁22張既有驗收保持通過；本節記錄
首頁之後的新反證，不推翻下方首頁限定CONFORMED，也不以舊母親E3代替原版對拍。

從原版冷啟動重播既有17次IRQ1輸入，第18次Enter在步數1,220,000,000，
第19次Enter在1,340,000,000；全部19次make/break共38次均在擷取前送達。
種子只在自然Lv1入口預先固定0x1357一次，沒有入口跳轉、座標／旗標／文字狀態注入。
唯讀呼叫觀測只列原始IDA位址、暫存器與DGROUP欄位，沒有增加原版寫入。

| 原始定位／動態步數 | 附加語意 | 等級與限制 |
|---|---|---|
| `sub_21414` linear `0x21462..0x21467`／file `0x127d2..0x127d7` → linear `0x21558..0x2157b`／file `0x128c8..0x128eb` | record82的`0xfffc`在目前行之後等待，再返回同一文字consumer；不是清框重新開獨立文字頁 | strong；分支、row writer與consumer已取得，完整中間捲動畫面仍待核對 |
| linear `0x216c3..0x21726`／file `0x12a33..0x12a96`，等待返回在步數1,220,000,367 | 第18次正常Enter解除生日內嵌等待 | confirmed；自然呼叫觀測與原版完整畫布閉合，未外推其他按鍵 |
| linear `0x21501..0x21534`／file `0x12871..0x128a4` → `sub_219F4` | `0xffff`結束分支在row=3時捲動後返回，沒有另一個按鍵等待 | strong；IDA資料庫原始指令與caller；沒有為硬體timer再開driver研究 |
| linear `0x100ab`／file `0x141b`，步數1,221,840,427 | 生日record82返回caller，seed仍0x356D | confirmed；同一次正常Enter之後，沒有追加生日結束鍵 |
| linear `0x100b5..0x100c1`／file `0x1425..0x1431` → `sub_11900`，linear `0x100c4`／file `0x1434`，步數1,222,328,466 | 生日返回後才寫DGROUP4F33／4F35為5、5，再消費場景；下一段文字之前重建共用視窗 | confirmed；writer／consumer與純讀取動態觀測，不把生日前raw15、22解讀成已確認場景座標 |
| linear `0x100cd..0x100d0`／file `0x143d..0x1440`，下一次內嵌等待步數1,224,373,605 | 同一次第18次Enter後，房間已出現，record83停在其內嵌等待 | confirmed；D3TXT01 record83與完整原版畫面；未宣稱重製房間像素已一致 |
| linear `0x100d5`／file `0x1445`，步數1,342,465,423；linear `0x100e9`／file `0x1459`；linear `0x100fa`／file `0x146a` | 第19次Enter後record83返回，進NPC consumer與record81，再返回開場caller | strong；有自然呼叫觀測，但後續畫面的渲染正確性、母親移動與同狀態關係尚未閉合 |

原始輸入：`assets_raw/DQ3.EXE`，115,282bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；
`assets_raw/D3TXT01.TXT`，6,420bytes，SHA-256
`4d0f78b20f123a986feb9af62845183213adee6881e953825eb74678abce8271`。
record82的file範圍`0x1426..0x1462`、record83 `0x1462..0x14ac`均含一個`0xfffc`與末端`0xffff`；
record81 `0x1402..0x1426`沒有`0xfffc`。這些是TXT file offset，不能和EXE或IDA位址混用。
IDA Pro9.4、MZ file=`linear−0xEC90`，DGROUP base linear`0x24dd0`，
runtime physical=`linear−0xef00`。新sidecar `work/issue4-birthday-pages-text-ida.json`：
769,130bytes，SHA-256 `690857dc7e26685c3fd0f05998b2d902c0a8307fa95d91b2d9cde2bc90416140`，
同列保存原名、定位、MZ file bytes、loaded bytes、xref與分級，沒有rename或修改原檔。

唯讀flow收據 `work/dosgolem-opening/issue4-archive-f8752bdf2c4e8d333c53b05bd591ccd211d34d6f16c516fde8643db0f06790df.json`，
24,493bytes，SHA-256 `f8752bdf2c4e8d333c53b05bd591ccd211d34d6f16c516fde8643db0f06790df`。
原始19次輸入、IRQ1、flow、生成腳本原文與所有圖像雜湊都保留；後續重生不覆寫唯一歷史收據。
執行器唯讀revision為`2f44a68ebfc54b28fb15dd4a34510b0b04a5415d`，
image為`dq3-ebiten-test:20260822-r1`，Go1.24.13；檔名／BIOS色盤修正仍在一次性工作副本內。

正式紅測試只透過`InputState`與空白更新等待文字穩定；不得因對話剛關閉便比較暫態黑畫面。
第18次輸入後，重製仍停在生日buffer pos15，完整640×350 RGB差異170,238像素；
第19次輸入後，重製才進record83，差異198,874像素。後者與原版已到後續caller的狀態不同，
只作診斷，不當成兩側相同record的繪圖parity。首頁兩張仍零差異，19次收據的創角8張也通過。
原版房間與目前重製房間還有視野原點、外部底色、文字布局與框線差異；截圖只是定位線索，
正式geometry仍須由場景／視窗writer-consumer閉合後才可放進JSON。

本節尚未READY：需補齊共享文字流的保留／捲動範圍、出生場景視野consumer與後續NPC事件，
再以可丟棄試作核對第18次輸入的完整畫布，審查有限game-pack契約。
正式修正需同時移除多餘生日EOF等待、維持可見文字順序、在生日返回後才交易出生場景，
並由正常輸入驗證下一個等待點；未知不可在Go或JSON中猜補。
目前沒有更改production或schema/content，仍為0.1.58／0.1.64。
重生入口：`bash tools/verify_dosgolem_newgame.sh /home/anr2/cht/dosgolem --birthday-pages`；
它必跑同批創角與續頁正式比較，未修正前預期退出1。既有首頁入口`--opening`維持綠色。

上述有界入口已實際冷啟動重跑，退出1來自已記錄的兩個續頁差異，創角8張通過。
與前一次唯讀flow冷啟動相比，27張原版PNG與色號位元組的雜湊、38次IRQ1及16筆flow全部相同。
最終原版收據`work/dosgolem-opening/issue4-birthday-pages-receipt.json`為24,493bytes，SHA-256
`7fd546d5b73a65933c5ddd24a34988bef4ca49df18d01074e082e692dce0417c`。
整批私有證據`work/issue4-birthday-pages-evidence-receipt.json`為2,999bytes，SHA-256
`e24ffc33a19413e15e91718c2ef3ff3d2d8f3fe7d943b1c6e67482ab7b7d8e32`，狀態為`known_mismatch`。
原版素材、圖片、state、database與完整收據不入Git或公開Issue附件。

已審查的返回／出生交易亦回填自動匯出索引，保留先前strong與unknown收據作歷史。
新增`work/issue4-birthday-pages-reviewed-ida.json`為1,380,185bytes，SHA-256
`82803df5acaa0c31986f670b6fa2566603062ba39cae63834d66b409e21f138d`。
confirmed只限表中自然輸入後的返回、5／5交易與record83等待；捲動畫面與房間渲染仍未升級。

## 2026-10-01：接受角色後的生日旁白（CONFORMED，首頁限定）

工作依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)，目前主線基準
`f3e3e1189ba58bb4b2b6f0de4fdb7ee60f83ada2`。本節先閉合接受角色後的首頁；
本文件下方母親帶路的歷史 E3，不等於 dosgolem 已完成同狀態對拍。

冷啟動延續 16 次已驗收創角輸入，再以第 17 次 IRQ1 Enter 接受男性、英數姓名 `0`。
原版只在自然 Lv1 入口固定 seed `0x1357` 一次，未寫入座標、對話或旗標。
私有原版收據 `work/dosgolem-opening/issue4-opening-receipt.json`：17,832 bytes，SHA-256
`0a83f01a50aae94aedb3e9dd20485c2f2f6490548990dc59035233e4cc57dad5`。
第 1,111,000,000 與 1,191,000,000 步的 PNG／色號位元組完全相同；
中間一張多出等待箭頭。不以圖片檢視工具曾呈現的黑畫面推論文字消失。

| 原始定位與 consumer | 有限結論 | 等級與證據 |
|---|---|---|
| IDA linear `0x10077..0x1009a`／file `0x13e7..0x140a`，`sub_20A07` | 創角返回後清畫面，再開共用視窗 | strong；database caller、四平面清零 writer；自然原版首頁黑底，該首頁已 confirmed |
| linear `0x100a3..0x100a6`／file `0x1413..0x1416`，`DI=0x0c0a` → `sub_21414` | 消費 D3TXT01 record82；房間尚未顯示 | confirmed；原始 record 與冷啟動畫面閉合 |
| linear `0x100b5..0x100c1`／file `0x1425..0x1431` | 上述文字返回後才寫 `DGROUP4F33/4F35=5,5` 並呼叫 `sub_11900` | strong；本批未動態驗收下一個畫面，不能宣稱出生點已對拍 |
| linear `0x28c3e`／file `0x19fae`／DGROUP `0x3e6e` → `sub_15002` → `sub_1F590` | raw `0b 01 13 00 ee 00 2c 00 60 00 94 01`：原點 `(152,238)`、352×96、record404；框線由完整字模內容產生 | strong；與原版首頁核對後才升級，不能把 raw 外界直接畫成一條矩形 |
| linear `0x214f8..0x214fe`／file `0x12868..0x1286e` | 一般字模之後 `SI+=2`、`BP+=3`，橫向步距 24px | strong；IDA bytes、writer／consumer，待試作全畫布核對 |
| linear `0x215b7..0x215cb`，`sub_21651` 保存並還原 SI、`BP+=名字長度×3` | `0xfff5` 是獨立的姓名插值控制碼，不吞掉下一個 word；下一個 `1` 是「16歲」的十位字模 | strong；生日 raw record、插值 consumer；舊 `+1參數` 說法在此不成立 |

輸入 `assets_raw/DQ3.EXE`：115,282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；
`D3TXT01.TXT`：6,420 bytes，SHA-256
`4d0f78b20f123a986feb9af62845183213adee6881e953825eb74678abce8271`。
IDA Pro 9.4、位址空間為 IDA linear，MZ file=`linear−0xEC90`；匯出同列保留
IDA loaded bytes、MZ file bytes、原名、推論等級與出處，原檔不修改。
重生入口為 `tools/ida_dump_opening_handler54.py`，target file `0x13a0`／`0x12784`。
第一份 caller／直接 callee sidecar `work/issue4-birthday-caller-ida.json`：1,352,787 bytes，
SHA-256 `83cacdd7cccb1838072f9a3b14af19be4cce239f0e9d25e36104fe1be52ac785`。

原始生日首頁與目前正式輸入重播的差異為 167,706 個 RGB 像素，遍及整張畫面。
重製提前顯示房間、以矩形代替 record404、採 16px 步距，且錯吞十位字模。
可丟棄繪圖試作已通過完整 640×350 RGB 比較，差異 0，沒有裁切、遮罩或更換原版收據。
試作紀錄 `work/issue4-birthday-prototype.log` 保留；這仍不是正式修正。
因此將黑底、record82文字 ID、record404完整框線 ID、raw版面、24px步距與獨立變數碼
契約審查為首頁限定 READY：新增 `interface.opening_prelude`，透過具名文字引用和
有限的 `glyph_step_x`／`variable_code_words` 原語呈現，不嵌入任意程式。
正式引擎只讀契約；缺欄位、引用、框線字模形狀、顏色或版面一律拒絕載入。
採用的動態姓名控制碼 `0xfff5` 本身只佔一個 word，姓名與一般字模都前進 24px；
其他正式對話目前仍沿用既有解析，本批不把尚未重驗的整體文字規則宣稱為原版 parity。
箭頭動畫、續頁、房間與母親仍待後續收據，正式首頁已由17次正常輸入與完整RGB比較限定 CONFORMED。


### 首頁正式驗收與下一閘門

提交前另核對 IDA 匯出工具的導覽標籤：舊 sidecar 的頂層 `unknown` 標籤沿用
歷史模板 `file 0x147b`，但實際 target、原始 bytes、xref 與分級索引均為
`file 0x13a0`／IDA linear `0x10030`。工具現改為依指定 target 產生標籤；
舊收據與整批驗收雜湊保留，不覆寫歷史。新增私有補充收據
`work/issue4-birthday-reviewed-final-ida.json`：1,369,167 bytes，SHA-256
`f9f2ae6ec39d7db880b7e028507a3eb31947025e1f58a46d45dd6845cdc27cea`。
本次勘誤只修導覽中繼資料，不擴大原有語意證據等級。

正式兩張生日首頁零差異，連同既有20張共22張通過；DRAFT繪圖替換已移除。
`bash tools/verify_dosgolem_newgame.sh /home/anr2/cht/dosgolem --opening` 已在正確image
由原版冷啟動重生、17次IRQ1 make/break共34次送達後，再通過創角與生日首頁正式比較。
中間擷取標為 `birthday-wait`，不能以舊診斷名稱 `home` 當成房間證據。
最終原版收據18,022bytes，SHA-256
`8bed7d6d21c7fb7f401c4d1f17c7c8bc45429c7a21e380306cdb5a157879f2bb`；
最終整批私有驗收19,286bytes，SHA-256
`c0a5cdb30fe9d61fe5c5be5ba947b0bfd1a88cb0ceb7d59dba5666ce6e967ab9`，
路徑 `work/dosgolem-opening/issue4-birthday-verification-receipt.json`。
生成腳本原文與hash已同時納入收據，先前收據以內容hash保存，不改寫形成史。

最終有限語意sidecar `work/issue4-birthday-reviewed-ida.json`，1,369,130bytes，SHA-256
`220bd5fc93d3e068a8fff4f4c1999246e17f18b2ffe830b58a126f0862af2422`。
自動合併台帳、保留原名／定位／兩種bytes、警示未確認語意，並檢查docs/94／84回鏈。
confirmed範圍只包含raw window前12bytes；相鄰欄位與下一結構不能跟著升級。

schema/content為0.1.58／0.1.64，EXE／TXT parity與33項旁白損壞／巢狀缺欄位拒絕通過。
全部11個internal套件、desktop main.go、標準game359項頂層／34項子測試通過；
37項選用在標準批次跳過，其中4項原版收據已另行嚴格通過，沒有素材缺失。
正常新遊戲→THE END為102.55秒，主角／酒館與各段存讀檔通過；這只證明重製回歸。

首頁只宣稱E2／限定V3可見畫面。重製仍沿用先前內部場景／座標預載，完整出生狀態、
生日續頁累積／捲動、箭頭、母親、音訊與全campaign均未達原版CONFORMED。
下一個原版輸入先重生生日續頁，不能用後段checkpoint或直接事件呼叫跳過。
Docker一次性容器已清理、UID/GID1000、原版及上游唯讀；歷史root候選3636項不改動，
沒有`.md`誤掛載目錄或本批root產物，也沒有建立新image／發行包。

日期：2026-08-24。這份文件取代 `docs/66`、`docs/74` 與 `docs/187` 中「handler54 在
CTY00 sec0 `(8,38)` 立即播 rec80 並交回操作」的舊結論。舊證據保留，但舊解釋已推翻。

## 1. 問題與證據等級

| 斷言 | 等級 | 證據 |
|---|---|---|
| rec81 後母親與主角在家中逐格走到右下樓梯 | confirmed | 本機原版影片 `4:53–5:12`、CTY00 家中畫面 |
| 樓梯後主角仍由劇本控制，自動走到王城入口 | confirmed | 同影片 `5:12–5:20`；主角在沒有玩家方向輸入的連續鏡頭中由家門移動至城門 |
| rec80、rec79 在城鎮路線 `(21,19)` 依序顯示，兩段關閉後仍自動走到王城入口 | confirmed | 同影片約 `5:17–5:24`；IDA `0x101bf..0x101da` 證實兩個連續文字 consumer，中間沒有 movement call；見 `docs/192` |
| `sub_1010B` 涵蓋轉場後序列，而非只到家門 | confirmed | IDA Pro 9.4、`DQ3.EXE` SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；IDA linear `0x1010b..0x1020b`／file `0x147b..0x157a` |
| file `0x154d..0x1565` 三段 `DGROUP 0x4f1f=1,2,2 → sub_194C3` 是轉場後可見自動移動 consumer | strong | 同一函式內位於場景切換之後、旗標交易之前；與完整影片順序吻合。`sub_194C3` 內部逐欄語意尚未全部命名 |
| `arrival_frames` 的逐 tile 路線 | strong／可見近似 | 由原始 CTY00 sec0 的可通行格，取 `(8,38)` 至城門前 `(21,9)` 的合法連續路徑，再與影片的轉折及終點核對；尚未宣稱每一格與原版內部路點表逐項相等 |
| 王座 rec78 時主角應站在國王正下方且可見 | confirmed | 原版影片約 `5:27` 起的同狀態畫面；`TestOpeningKingAudienceRendersHero` 以正常 renderer 與透明狀態作同 tile 像素差，證實 remake 已畫勇者。先前把 runtime PNG 中的角色認成大臣，不能列為產品缺陷 |

IDA sidecar 由 `tools/ida_dump_opening_handler54.py` 可重建；暫存 `.i64` 與 JSON 不加入 Git。
工具位址與 file offset 必須並列，不把導航名稱 `sub_1010B` 當作語意證據。

## 2. 推翻舊結論的理由

舊分析只閉合：

```text
rec81 → 家中逐格帶路 → CTY00 sec0 (8,38) → flags／rec80
```

但完整函式與玩家畫面實際是：

```text
rec81
→ 家中母親＋主角逐格帶路
→ 樓梯／場景轉換至 CTY00 sec0 (8,38)
→ 主角單人自動行走至 (21,19)
→ rec80（母親指路）→ rec79（含主角姓名的催促）
→ set flag0x17／clear flag0x50
→ 繼續自動行走至王城入口 (21,9)
→ 交回玩家操作
→ 玩家進王城、上樓
→ 站在國王正下方觸發 rec78，主角仍可見（既有 renderer 已符合）
```

因此 `(8,38)` 只是第二階段起點，不是整個 opening runner 的完成點。舊測試只驗 transaction
與最終主線可達性，無法證明控制權交還時機或逐格畫面。

## 3. game-pack 契約

`interface.opening_escort` 改為兩階段有限序列，不在 Go 內新增 DQ3 座標、record 或旗標：

- `cty/section/frames`：既有家中 leader＋player frames。
- `destination{cty,section}`：樓梯後目的場景。
- `arrival_frames[]`：轉場後只含 `player{x,y}` 與 `hold_frames` 的完整 tile 序列；第一格必須
  等於目的場景落點，後續每格必須四方向相鄰且可行走。
- `dialogue_frame_index`／`dialogue_records[]`：指定非首尾路點與原版有序對話；本 pack 為
  index 36、`(21,19)`、records `[80,79]`。最後一段關閉後才重播剩餘 frames。
- `set_story_flags`／`clear_story_flags`：分別保存 `0x17`／`0x50`；缺值或未知引用 fail closed。

共用 Go 只負責有限序列、方向／步行幀、場景切換、對話與旗標 effect primitive。它不得知道
CTY00、`(8,38)`、王城入口、rec80、`0x17` 或 `0x50`。

本次再勘誤後 schema 為 `0.1.53`，DQ3 content 為 `0.1.59`。

## 4. runtime 規格

1. rec81 關閉後鎖住一般玩家輸入，啟動家中 frames。
2. 家中 frames 完成後載入 `destination`，不得先開 rec80、不得先交易 completion flags。
3. 逐格執行 `arrival_frames`；每個可見 tile 更新 facing／walk，期間一般玩家輸入仍被鎖住。
4. 抵達 `dialogue_frame_index` 才依序開 `dialogue_records[]`；最後一段關閉後才套用
   completion flags，並繼續剩餘 `arrival_frames`，到最後一格才結束 opening。
5. 玩家進王城與上樓仍走正式 portal／方向輸入，不自動傳送到王座。
6. 王座 region 觸發 rec78 時，player renderer 必須在 `(aliahanKingX, aliahanKingY+1)` 對應的
   畫面 tile 留有非背景 sprite pixels。2026-08-24 的像素差異測試已證實既有 renderer 符合；
   此項保留為回歸 gate，不另改產品程式。

## 5. 驗收 gate

- schema／reference validation：destination、arrival frames、record 與 flags 全部合法。
- component：家中與轉場後每格相鄰；rec80 前不交回操作、不提早交易旗標。
- production trace：標題 → 正式創角 → 家中 → 自動抵達城門 → 玩家進城 → 王座 rec78。
- visual：至少輸出家中中段、家門落點、城鎮自動行走中段與 rec80 路點；rec78 王座由既有
  runtime PNG 加像素差異測試守門。
- oracle：與原版影片同狀態目視核對；播放器裁切不作遊戲 viewport 幾何證據。
- save/load：rec80 結束後與首次謁見後各做 round-trip；演出中途存檔不在本批範圍，必須
  fail closed 或禁止存檔，不猜 resume 語意。

2026-08-24 已通過 schema、component、王座像素差異及正式創角 production trace 的開場段；
完整 trace 繼續抵達後續主線，最後在既有「低危區練至 Lv20」500000-step 上限失敗，與本切片
無關。因此本開場切片恢復 E3；逐 tile 原版內部路點與逐幀 timing 仍只到上述證據等級，不稱 V3。

## 生日續頁保留與捲動規格，2026-10-02

DRAFT 建於本輪原型前，現經下列證據審查為 READY；範圍限 `opening_prelude` 的生日 record82。入口仍為正式創角，
第18次確認延續原有文字；捲動完成後才允許出生交易。房間、record83、箭頭與
NPC 動畫另有未完成閘門，不藉此擴大對拍聲明。

來源仍為上述原始 EXE 與 D3TXT01 雜湊。IDA Pro 9.4 匯出
`work/issue4-birthday-pages-text-ida.json`，769,130 bytes，SHA-256
`690857dc7e26685c3fd0f05998b2d902c0a8307fa95d91b2d9cde2bc90416140`。
這是 database 的 caller／consumer 匯出，保留原名、bytes、xref 與位址基準。

| 原始定位 | 行為 | 證據等級 |
|---|---|---|
| IDA linear `0x2149a..0x214ba` | 換行重設水平游標；四行滿後呼叫 `sub_219F4`，不自動依 columns 換行 | strong；consumer，待自然畫面閉合 |
| IDA linear `0x21558..0x21593` | FFFC 先推進一行，再等待確認；確認後沿用原畫面 | strong；consumer 與自然流程觀測 |
| IDA linear `0x219fe`／file `0x12d6e` | `B9 04 00`，四次捲動 | strong；原始 bytes 與四次自然呼叫觀測 |
| IDA linear `0x21a0a..0x21a28`／file `0x12d7a..0x12d98` | 寬40個 VGA bytes，即320px；讀取60列，來源 y+20，目的 y+16 | strong；視窗 consumer 與 copy writer |
| IDA linear `0x21a71`／file `0x12de1` | `BB 04 00`，每次清除底部4列；文字區320×64，每次上移4px | strong；四平面 clear writer |
| IDA linear `0x21501..0x2152f` | EOF 最後一字等待後捲動，再自動返回，沒有額外確認 | 已有 EOF confirmed；四次可見捲動待本輪核對 |

具名有限流程 `retained_rows` 保留繪圖操作，依原始控制碼換行。字模以16×16
不透明寫入，姓名沿用已審查的24px步距。`text_flow` 必填 mode、scroll_step_pixels、
scroll_steps、scroll_hold_frames 與 evidence；幾何由既有 Window 的 inset 與行数導出。
驗證要求 content height 等於行數×字模高度、四次位移合計一行、引用與 D3 證據有效。
此具名流程只搭配已審查的 `automatic_after_reveal`。捲動每步等待3個60TPS更新，沿用既有平台時序近似，屬 hardware-spec approximation；
不聲稱與原版硬體 wall-clock 逐週期一致。

驗收用同一種子0x1357、同一19次IRQ輸入的 dosgolem 冷啟動收據，新增 EOF 前及
四次捲動後擷取。重製經17次創角輸入與第18次確認，僅用正式空白 InputState 等待。
逐張比較完整640×350 RGB，不裁切、不遮罩；另外檢查捲動中不交易出生、不消耗RNG，
返回後抵達原有下一節點並由既有 production trace 驗證存讀檔。原型通過才審為 READY。

### READY 審查結果

固定來源 `e2d6cc25467a8e3d4896a7f68ead4f3110d4588b` 的私有原型由正常 InputState
抵達全部六個比較點，完整RGB均為0差異；已目視核對續頁末尾與第四次捲動。
`work/issue4-retained-prototype.py`／`work/issue4-retained-prototype.go`／
`work/issue4-retained-test.go` 與 `work/issue4-retained-prototype.log` 保留重生方法與結果。
原型前原版收據已保存為 `work/dosgolem-opening/issue4-archive-fe874811378a8f88223dc55b8cc4361ac391105e86e5e0c5b5017f59c36f2984.json`：36,603bytes，SHA-256
`fe874811378a8f88223dc55b8cc4361ac391105e86e5e0c5b5017f59c36f2984`。
既有54張PNG／bin內容均與上一批一致，38次IRQ1送達，兩側種子執行前固定一次0x1357。
沒有修改輸入、注入座標、文字或故事狀態。新擷取只增加既有consumer邊界的唯讀觀測。
上表可見幾何、FFFC保留及四次捲動升為 confirmed，僅限此生日流程；時序仍為平台近似。
四次捲動位於step1221370959／1221491508／1221612130／1221732757，consumer均為
IDA linear0x21A8B。完成時1221840427的畫布與第四次擷取相同，caller為linear0x100AB。
正式實作接入前的房間紅測試仍為172,261差異，本段驗收不會取消該閘門。
