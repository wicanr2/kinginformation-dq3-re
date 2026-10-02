# DQ3 工作歷程

## 2026-10-03：Issue #4 家中正式流程、圖像選擇及手動出門

接續a2e477d，先閉合獨立圖像時鐘、原始問題表、預覽及Escape的正常輸入證據。
原版contract37次／74次IRQ1／170個唯一產物，navigation42次／84次／180個產物全部核對。
創角seed1357與圖像seed151B分開固定一次；後者來自原版本次自然BIOS值，不永久鎖定正式遊戲。
Escape實際保存當前選項並前進，推翻初步取消解讀；保留原始EXE的NOP比較與未驗用途。

可丟棄原型的兩條正常路徑、存讀檔及五個視窗通過後，docs/188審為有限READY，
才接入正式共用狀態機與schema0.3.0／content0.1.71的JSON。
母親原始record0先獨走至10,10，主角不跟隨，之後對話、選圖並交還控制。
13步正常移動接近9,10才接城鎮42個狀態；最後三步才交易旗標。
家中等待接近可從標題正常讀檔恢復，損壞或不相容存檔拒絕且不消耗現行狀態。

七個素材像素差異由原始34px游標框跨32px圖像格的XOR順序閉合，沒有修改原始圖片。
預覽、三輪及結果的完整336×192視窗均0差異；整張640×350各124人物像素仍RED。
實際Enter與確認鍵欄位不同，補齊性別與有限開場對話／選圖，正式Enter路徑及原有確認鍵比較都通過。
嚴格JSON、原始EXE／TXT oracle、BLS素材形狀、正常手動出門、城門旗標及存讀檔通過。
最後完整game370項頂層／45項子測試、11個internal套件及desktop建置通過。
正式新遊戲到THE END224.65秒，只屬remake可玩回歸；38項game及4項internal選用未執行，沒有素材缺失。

第一次完整回歸被4GiB容器OOM中止，Docker事件已核對；沒有測試斷言失敗。
保留中止JSON log，以同一image及4GiB硬上限、Go軟記憶體上限2GiB乾淨重跑。
其餘當輪型別、schema fixture及OPAQUE前導空白修正均在相同工具鏈重跑，沒有放寬原版比較。
原版完整圖像、收據及IDA database只留本機；完整V3、NPC動畫、城鎮水平視野與音訊仍未完成。
本輪未建立發行包，驗收、工具入口與下一切片見docs/188、docs/74及CONTEXT唯一狀態表。
384項本批程式／輸出UID／GID1000，歷史root-owned3213項與使用者scratch保持，Markdown目錄0。
DQ3一次性測試容器已清除；公開Issue只更新有限結果、限制及下一步，不附原版素材。

## 2026-10-02：Issue #4 母親接近入口與37步帶路的原版收據

接續ad00d38，在既有IDA9.4容器中找到010B入口的資料xref與DGROUP3BB4 raw54分派表，
再由11A6A目標cell writer→19530→196D2→1970B閉合，保留原始名稱、bytes、位址與分級。
最初正常方向鍵停在圖像選擇選單；補足進入、三次選圖與結果確認的五次Enter。
最初向左通路碰床，改為下2／左2／下3／右6；任意抓圖點暫時DS的錯誤斷言限回創角段。
480秒逾時改為660秒有界執行；分類錯誤另修為自然抵達1010B後才標其他相位。
舊產物先完整按hash歸檔，再刪除本情境舊檔重生；每張PNG都核對本次log生成記錄。
上述是探測與觀測問題，沒有當作產品缺陷、狀態注入或RND重擲。

正式來源唯讀、乾淨dosgolem 2f44a68、既有dq3-ebiten-test image。
原版37次正式輸入／74次IRQ1、seed1357一次：主角9,10才觸發母親；轉場後8,38／母親8,37。
37個主角自動步進逐格核對，抵達22,19／母親22,18轉下；停在城門對話，尚未觀測101D3／1020A。
原版收據1e5ab45d，464,340bytes，210份產物全部核對；前一生日116份PNG／bin由本次重生且相同。
九份IDA匯出3,720列原始指令核對，台帳自動合併分級，docs/66／192追加歷史反證回鏈。
完整私有收據與工具入口在docs/188；原版圖像、database與收據不入Git。

固定ad00d38的正常InputState人物反例FAIL，NPC仍5,4；可丟棄原型PASS，NPC先到10,10轉左再開81。
這只驗人物選擇、位置與順序；全域相位、正常接近、存檔及正式RGB未由原型驗收。
正式Go與JSON未改，schema0.1.62／content0.1.69及前次完整回歸保持。
新增--mother-entry-original重生模式只驗原版入口，不宣稱remake parity。
本輪持續以Issue #4追蹤；下一閘門為對話後／保存／恢復與有限READY，再修改正式路徑。

## 2026-10-01 — Issue #4：能力等待／確認正式對拍

使用者再次明確授權建立／更新Issue；主機gh回讀確認啟動Issue #1已關閉、創角Issue #4開啟，
續行記錄更新到#4，沒有重複建立。基線0a8aa6a的嚴格等待測試仍RED，15,468像素差異及提前
開始遊戲保留。IDA 9.4追caller後證實能力選項在DGROUP4080／linear28E50／file1A1C0、record434，
舊28BC6定位實為性別record556。原始名稱／位址不改，舊證據保留，docs/118、126同批回填勘誤。

原始完整record407／557／434、字模及frame XOR順序的隔離試作兩張差0，先完成DRAFT→READY審查，
再接入正式ngReview→ngConfirm與pack（schema0.1.57／content0.1.63）；試作碼收尾移除。
正式等待／確認兩張及既有18張全畫布RGB均0，冷啟動、16次真實IRQ1輸入重生後再次通過。
原版自然入口及重製首個輸入前各固定1357一次，七能力／目前HP、MP及最後356D仍相同。
ACK現在為創角必跑閘門；閒置／只按住方向不前進，新的按鍵只顯示確認，不重擲或接受角色。

標準game359項頂層／34項子測試通過；36項選用未執行，其中3項原版收據測試另行嚴格通過。
沒有素材缺失。正常正式輸入／主角與酒館命名／存讀檔至THE END（99.04秒），全部internal及desktop建置通過。
首次新增parity測試誤用IDA relocation bytes、預覽經驗值型別及歷史raw-window數量8造成驗證失敗；
依原始檔及新契約訂正後在相同工具鏈乾淨重跑，沒有弱化斷言。匯出工具同列IDA loaded與MZ file bytes，
另驗證歷史規格backlink；所有database／原始圖像／收據留在本機。

私有驗收work/dosgolem-opening/issue4-ack-verification-receipt.json為20,665bytes，SHA-256
6f6aba6327d33f7136e96584e20190229152f2298b0e99fdcfa31c8cccdef052。限定能力等待／確認CONFORMED，
接受角色後、出生點／母親、女性、候選字、閃爍、音訊與完整原版campaign仍待後續Issue工作。
本輪一次性容器收尾清空、輸出UID/GID1000，歷史root-owned3636項保持原狀；使用者scratch／Android libs保留。
commit／push及遠端回讀另回填Issue，不建立發行包。


## 2026-10-01 — Issue #4：性別畫面與固定種子Lv1交易

沿用遠端Issue #4及已授權的更新／commit／push，沒有重複建立Issue。由冷啟動IRQ1
英數命名「0」、完成及預設男性自然到生成入口；IDA 9.4先確認能力種子是DGROUP0B5A，
舊文件的BIOS tick→CS:701B屬預設名稱，不能混用。原版生成入口固定1357一次，重製
首個正式輸入前固定同值。七項能力、目前HP／MP及最後seed356D一致，交易對拍通過。

正式性別PNG初差3,176像素；以raw0x28BC6／record556／選項consumer及完整IRQ1畫面
完成DRAFT→隔離試作0差異→READY，再接入共用索引色renderer與pack，正式差異0。
五張完成前命名畫面同樣0；先前12張重跑後均0，合計18張完整640×350 RGB。
schema/content升為0.1.56／0.1.62；EXE／TXT parity、五項性別損壞契約拒絕通過。
原版EXE／dosgolem上游唯讀；完整素材、PNG／state、IDA及收據不入Git。

完整game360項頂層／35項子測試通過，33項選用擷取／額外收據未執行，沒有素材缺失。
正式新遊戲至THE END及各段存讀檔通過（91.52秒）。internal首次被六項舊schema測試
輸入擋住；更新該輸入後，使用相同容器／命令乾淨重跑全部internal及桌面main.go建置通過。
冷啟動重生入口`tools/verify_dosgolem_newgame.sh --creation`亦通過；完整參數見docs/113。

保留下一個產品阻塞點：原版能力面板`sub_1834E→sub_2111B`等待Enter，才返回caller並
畫確認提示；原版三個PC事件及正式按鍵已閉合。重製選性別後提前提示，再按Enter提前
開始遊戲。明確啟用的能力等待稽核保持RED，等待頁差15,468像素，沒有弱化斷言。
本批性別／能力交易CONFORMED不代表創角確認、出生點／母親或完整原版campaign完成。

新匯出入口`tools/ida_dump_creation_rng.py`同批掛入docs/113；有限range台帳自動附加
語意、推論等級、證據與未達confirmed警示，原始名稱／位址／bytes／xref及unknown保留。
首次ASCII文字編碼失敗改明定UTF-8，仍用同一IDA工具鏈重跑；Docker heredoc初漏`-i`
導致腳本未執行，補標準輸入後重跑；均分類為工具問題，不改遊戲規則。

最新私有驗收`work/dosgolem-opening/issue4-creation-verification-receipt.json`，19,276bytes，
SHA-256 `0581086322b07394b8c9dffa7e697c115b220670e6613de5f046b426f83b1775`；
精確原版／IDA收據hash與入口見docs/113。最後Docker與擁有權、工作樹及push核對回填Issue。

## 2026-10-01 — Issue #1：dosgolem 開場對拍

起始 HEAD `0c6f780`，未追蹤使用者檔案全數保留。主機 gh 登入成功，遠端原先沒有 Issue；
自動核准審查首次拒絕建立，使用者再明確授權後建立並回讀
[Issue #1](https://github.com/wicanr2/kinginformation-dq3-re/issues/1)。

dosgolem 上游 `2f44a68` 從真實 EXE 入口探測，原版輸入 hash 與工具版本見
[docs/196](docs/196-dosgolem-opening-sequence-parity.md)。未修正探測黑畫面由空白補齊的
DOS 檔名造成；可丟棄副本修正後 DOS tests PASS，原版自然執行至六張卡片與標題／巡禮。
確認 remake 漏 1990 卡且誤用候選紋章，規格審查達 READY。

環境勘誤：容器登入 shell 覆寫 Go PATH，改 `bash -c`；Xvfb-run 作為 PID1 只剩
啟動等待，依實際程序狀態停止後改有界 bash／Xvfb＋trap，相同開場測試及五張擷取 PASS。
第一次 `-shots` 的原始色號檔誤命名 `.png`，後改 `-dump-at` 由 dosgolem 平面擷取重生。
這些是工具／驗證環境問題，不記為 remake 缺陷。

首次正式第六幕比較失敗，證實 TITF 只有背景，原版另有 DFP 動態疊圖；規格退回 DRAFT，
以 IDA 9.4 追原始 caller、writer、directory、兩種平面 consumer、座標及翻頁。
獨立解碼與原版全部 129 次翻頁一致後達 READY，再接入共用引擎與資料包。
schema `0.1.54`／content `0.1.60` 的正式無輸入序列、原版前五幕色號及第六幕 129 個
位置的色號／RGB 共 30,016,000 個比較像素通過；時間採公開 PIT 公式近似，沒有深入 ISR。

全部 internal、受影響開機／存讀檔與桌面建置通過。完整 game 有母親帶路斷言及 Lv1
盜賊鑰匙 fixture 全滅兩項失敗；以修改前 `0c6f780` 在同一容器與實際素材重現相同失敗，
另登記 [Issue #2](https://github.com/wicanr2/kinginformation-dq3-re/issues/2)，不能冒稱完整回歸全綠。
本批未建立發行包，未提交原版素材或 IDA database。commit／push 與最後清理核對於 Issue #1 記錄。

## 2026-10-01 — Issue #3：同伴持冠還冠 gate

Issue #2 正式重播在同伴持冠並存讀檔後，晉見國王仍走任務對話。保留原有 docs/82，
以 IDA 9.4 一次性 database 重查 handler9 → sub_1689C → 命中 SI → caller 間接清格，
確認原版逐角色掃八格。新增可重現匯出入口與分級索引；來源 hash／位址與 READY 審查
見 docs/82，不提交 EXE 或 database。第一份暫存 sidecar 為空，未採用；明定 UTF-8、
增加工具日誌後重生並核對非空／schema／hash，根因未獨立驗證。

新增同伴持有兩件的 component 在修正前失敗，gate 由主角計數改為全隊計數後通過，
只消耗一件且不動無關持有者。既有任務、王位、辭位與存檔派生測試通過；同輪正式
新遊戲 trace 已經同伴持冠、隊伍存讀檔、還冠、王位／辭位，再繼續至日邦格紫寶珠。
除整段主線外，其餘 game、全部 internal 與 desktop main.go 建置通過；原版 dosgolem
動態同狀態路線仍未取得，整段主線由 Issue #2 繼續重驗，不能宣稱 E3 或完整 parity。

回歸計數：381 項 game 通過、32 項 opt-in 跳過，沒有素材缺失。31 項是可選畫面擷取；
另 1 項為未指定收據目錄的開場對拍，重新指定既有收據後獨立通過（3.84 秒），
前五幕色號與第六幕 129 次翻頁色號／RGB 維持一致。這不是重新執行原版 campaign。

測試工具環境：高速正式路線會密集建立原生音訊播放器，曾出現 SIGKILL 與音訊堆疊逾時，
前者未證實 Docker OOM。測試專用 adapter 停止輸出聲音，但仍轉送原始 VOC duration 與
完成等待，不修改 production 音訊；這些回歸沒有做人耳驗收。編譯診斷曾誤用不存在的
companionsToSave 方法，改回 compsToSav；更換回復 closure 後移除未使用宣告，乾淨重跑。

工作樹未追蹤使用者檔案保留；本輪未建新發行包。還冠局部修正已由
`fc78bb5d5f8bc97f91ad0b34fc257ce7834dc638` 提交並推送，
[Issue #3](https://github.com/wicanr2/kinginformation-dq3-re/issues/3) 已關閉。還冠三項元件再次
通過，已重生並目視核對顯字 PNG；路徑、hash 與元件範圍見 docs/82。

## 2026-10-01 — Issue #2：正式主線回歸重驗

起始母親失敗是舊斷言仍期待 rec80 結束就移動；現行原始開場規格為同格切換 rec79。
依正常輸入核對記錄、座標與旗標後訂正斷言。盜賊鑰匙路線的測試背包超過個人八格，
改以合法容量購入與正式使用聖水，路線回歸通過。兩者均不改 production 規則。

整段主線每次自新遊戲入口開始，執行前固定 remake 種子 `0x1357`，未重設種子挑選結果。
回歸陸續越過同伴持冠／還冠、轉職後正式重新裝備、兩場八頭大蛇與紫寶珠、建城商人、
拉之鏡與怪力魔、變化杖交換、幽靈船愛的回憶／存讀檔及奧莉薇亞海岬旗標交易。
測試策略改用正式商店關閉、教會復活／解毒、旅店、隊伍容量與道具給予、回復咒文、
攻擊增益及既有避敵資源；事件或數值沒有注入。缺陷中的同伴還冠另由 Issue #3 追蹤。

幽靈船前舊十格補給斷言已與後續野外回復規格衝突，訂正追加於 docs/170，保留歷史。
海岬原航路未啟用避敵資源，且以怪物編號 80 作逃跑門檻，遇 monster78 全滅。
原始 D3MNS.DAT（5330 bytes、SHA-256
`48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`）record78、
file offset3198 的 cast probability 為40，唯一 mask bit18；依 docs/180 已審 consumer
為全體死亡後施術怪 HP 歸零的 `sacrifice_death`。採正式避敵與能力判定後海岬事件通過。

返航診斷再顯示三名同伴死亡、勇者 MP 僅1；出港前旅店不能代替復活，幽靈船去程也仍
有舊怪物編號策略。正在同一容器、同一種子乾淨重驗正式教會與航行資源策略，尚未
THE END，不宣稱整體 E3 或原版同狀態對拍。一次編譯使用不存在的復活輔助函式名稱，
核對既有入口 `traceReviveDeadAtChurch` 後更正；這是測試程式錯誤。

後續勘誤：此 checkpoint 只有勇者已學特黑洛斯，同伴剩餘 MP 不能作避敵資源。取得
愛的回憶後經正式魯拉回港補給，海岬、蓋亞之劍與存讀檔通過。無轉場步行加入與一般
步行相同的有界 NPC 等待，越過回港通道的暫時阻塞。尼羅肯特仍在避敵資源耗盡後遇
monster91／92 全滅；補給時另抓到備品整理把有場景用途的蓋亞之劍丟棄，已依資料包
效果引用保留。保留魯拉的戰鬥策略也錯誤封住同伴全部 MP，正限定只保留勇者成本，
讓同伴使用已學回復及傷害咒文；仍由相同固定種子乾淨重驗，不改 production。

選用大型畫廊在魔法球測試設定停止，沒有跑到還冠段；不據此猜改遊戲。限定同伴還冠
元件擷取通過。初張 PNG 是逐字顯示初始空白幀，依資料包等待長度推進文字後重取；
不是原版同狀態畫面。此限制已登記 Issue #2／#3。

最終回歸：實際地表進洞入口原先沒有避敵，補上正式資源使用、同伴戰鬥 MP 與位置
診斷後，尼羅肯特／銀寶珠、六珠、巴拉摩斯及下層世界均通過。彩虹橋原斷言只查勇者
聖水，改用全隊物品或既有可負擔咒文，正式主線由標題抵達 THE END（148.88 秒）。
種子 0x1357 在第一個輸入前斷言固定，沒有重新設種或挑選結果；各段存讀檔通過。

同一版本其餘標準 game 回歸 356 項頂層／382 項含子測試通過，加上主線為
357 項頂層／383 項含子測試；31 項明示選用畫面擷取跳過，沒有素材缺失。
指定既有原版收據的六幕／129 翻頁比較通過（9.09 秒），全部 internal 與 desktop
main.go 建置通過。未重複執行已通過且未再改動的長主線。

本機收據為 `work/dosgolem-opening/issue2-regression-receipt.json`，3392 bytes、SHA-256
`b12899ff50e98242e7c9303b3503df2defde4dd2d7698bdf4762246fb374c7cb`，保存來源檔 hash、
工具 image、種子、輸入、範圍及日誌雜湊。正式 THE END PNG 已目視核對，14386 bytes、
SHA-256 `250a981f8f56aa3130e83c99d1b330a7474a5432b132203e2056d23c804fbfc9`。
原版開場僅重用既有 dosgolem 收據比較；創角後原版／完整音畫與人耳驗收仍未知。
這些結果不代表完整 remake 或 dosgolem campaign 對拍完成，沒有新發行包。

Docker 衛生：確認 docs/105-gaia-sword-re.md 與 docs/106-gaia-sword-idapro.md 為
UID0、完全空白的歷史誤掛載目錄後，只移除這兩個目錄。未遞迴更改持有者；使用者
scratch、Android libs、原版素材與資料庫均未改動或納入提交。另確認並移除同樣完全
空白的 game/world.go 誤掛載目錄；全專案仍有 3636 個歷史 root-owned 項目，未廣域修復。

## 2026-08-26 — README 歷史自評（2026-10-01 移存）

以下逐字保留原首頁的當時自評與來源，不能覆蓋目前 CONTEXT／docs/74／遠端 Issue。
相對連結因兩文件同位於根目錄而保持有效。

目前狀態（2026-08-26）：指定的六項玩家功能與四項殘餘 polish 已接入；最新原版／remake
game test 找到開場連續演出在家門轉場後缺少「主角自動走到王城入口」的玩家可見段落，
現已依 [`docs/188`](docs/188-opening-escort-to-castle-spec.md) 改為 game-pack 兩階段序列。
王座 rec78 的勇者經像素差異測試確認既有 renderer 已正確繪出，先前截圖判讀不列為缺陷。
現行工作樹為 schema `0.1.53`／content `0.1.59`。本輪依玩家實測把正常音樂固定為
直接 OGG（FM 僅保留明確診斷開關）、將注音組字移回原版下方面板、加入 `ㄨㄤˇ` 可選「王」
的明示相容別名、補母親逐格帶路，以及 NPC 待機兩幀與移動面向同步；見
[`docs/187`](docs/187-opening-player-path-polish.md)。同版公開 patch 與本機完整版已重包於
`dist-all/v0.1.36/`；四個公開檔已發布，四個完整版不公開，詳見 [`docs/194`](docs/194-release-v0.1.36.md)。
發行後工作樹另修正戰後固定誤回地表曲及音樂 OFF→ON 不立即恢復的 runtime 路由；
castle／town／dungeon／field 與 title／battle／ending 現共用同一場景 selector，見
[`docs/195`](docs/195-audio-transition-runtime-polish.md)。此修正尚未重包為新公開 release。
本機完整版已另以 `v0.1.38-local` 同 checkpoint 重包至 `dist-all/v0.1.38-local/full/`；
這不是公開 release，亦未上傳 GitHub。

### 自我評估：現行 remake 與原版的差距

**整體 remake 完成度：約 90%；相對原版仍有約 10% 差距。** 這是截至 2026-08-26、以
玩家可見成果加權的工程自評，不是「已解讀 85% executable bytes」，也不代表逐像素 parity。
固定計算方式為「玩家流程 35%＋資料／規則 25%＋畫面 20%＋聲音 10%＋平台交付 10%」：

| 評估面向 | 目前分數 | 權重 | 加權貢獻 | 主要依據 |
|---|---:|---:|---:|---|
| 玩家流程與事件串接 | 96% | 35% | 33.6% | 新遊戲至 `THE END` 的正常輸入 trace 已達 E3；開場 escort 已補正，尚未再跑完整人工回歸 |
| 資料、規則與存檔 | 93% | 25% | 23.25% | 藥草、聖水、祈禱之戒的參數、持有者／目標選擇與消耗順序已由 IDA 閉合並遷入 JSON；未知 helper 仍不冒稱 exact |
| 畫面與動畫 | 78% | 20% | 15.6% | BLK 位平面、同格 rec80→rec79、20×4 分頁與 16×16 glyph cell 已依 EXE／D3TXT／影片閉合；rec78 證據圖已排除逐字初始幀假缺字，但同頁原版畫格與全場景動態仍以 V2 為主 |
| 音樂、音效與時序 | 84% | 10% | 8.4% | OGG 與關鍵 cue 已接線；七個正式場景軌通過完整解碼抽樣，戰後依 castle／town／dungeon／field 恢復，音樂 OFF→ON 立即恢復當前場景；跨平台人耳、EBG 事件 cue 與逐動作停頓尚未完整對拍 |
| 三平台交付 | 94% | 10% | 9.4% | v0.1.36 的 Linux／Windows／macOS 公開 patch 與本機完整版同 checkpoint 封裝完成；macOS 尚缺真機驗收，Android 不在本輪三平台範圍 |
| **合計** |  | **100%** | **90.25%（對外取整為 90%）** | **剩餘差距約 10%，主要集中於全場景動態／聲音 V3 與真機驗收** |

本專案目前可合理稱為「主要玩家流程已重製、核心資料與事件大多接線完成」，但不能稱為
「逐畫面、逐幀、逐聲音完全等同原版」。下表保留各維度的證據與限制，避免總百分比掩蓋
證據強弱：

| 維度 | 現況 | 與原版仍有的差距 | 是否阻塞一般 remake 完成 |
|---|---|---|---|
| 主線可玩性 | 先前乾淨正式輸入 trace 已由新遊戲抵達 `THE END`，主要事件、戰鬥、載具與結局具正常入口 | 本輪完整 trace 已通過新修正的開場，後段因測試用低危區練至 Lv20 的 500000-step 上限停止；這是 trace 策略／隨機遭遇穩定性問題，不是已重現的玩家流程死路 | 否；但下一次發版前仍應用可重現 checkpoint 抽驗 |
| 開場與設定資料 | 家中母親帶路、轉場後自動行走、同格 rec80→rec79 與旗標交易已依 IDA／D3TXT／影片接成 game-pack 序列；BLK 位平面勘誤後背景色號已對齊 | `arrival_frames` 尚未證明每個內部路點／timing exact；逐 glyph timer 採硬體規格近似，不宣稱 DOS wall-clock 逐週期一致 | 不阻塞流程 E3；文字 record／cell layout 已閉合，整體動態畫面仍阻塞 V3 |
| 規則與資料 | 主線必要的道具、怪物 action、抗性、formation、商店、咒文、掉落、日夜及事件交易多數達 D2/D3→E2 | 藥草、聖水、祈禱之戒已移除舊近似並接正式選人；未被玩家路徑使用的 helper 仍保留證據限制 | 否；未知項不得冒稱 exact |
| 畫面與操作 | 原始資產、中文字型、HUD、選單、NPC、角色與主要場景均可由正式 runtime 顯示；能力確認固定畫面已有 V3 靜態對拍；BLK 位平面修正後 rec80 背景色號已對齊，opening 文字 record／cell layout 已達 V3；rec78 dump 現固定輸出完整當頁 | 多數其他畫面仍是 V1／V2；rec78 尚缺同頁原版畫格，不能把 viewport／palette 整體升格；逐窗、戰鬥、attract／ending timing 亦未全面 V3 | 不阻塞功能 release；阻塞「視覺忠實完成」聲明 |
| 音樂與音效 | 正常產品直接播放 OGG；關鍵戰鬥 VOC cue 與完成等待已接線；v0.1.36 完整版七軌通過完整解碼抽樣；發行後工作樹另閉合戰後場景恢復與音樂 OFF→ON 熱恢復，見 [`docs/193`](docs/193-audio-sampling-polish-20260825.md)、[`docs/194`](docs/194-release-v0.1.36.md) 與 [`docs/195`](docs/195-audio-transition-runtime-polish.md) | 跨平台實際裝置的人耳切換、EBG 事件 cue、逐動作停頓尚未完整對拍；DAC／PIT／DMA wall-clock 依公開硬體規格作可重現近似，不宣稱逐週期一致 | 否；真機人耳屬 V3 |
| 存檔與資料包 | pack 具 schema、reference validation、content hash；主要事件有 save/load transaction | 不是每個可選支線與每個演出中間 frame 都有 round-trip；演出中途存檔語意未證實時採失敗即關閉 | 否；正常 checkpoint 已涵蓋主線 |
| 平台交付 | Linux／Windows／macOS 已依現行 release checkpoint `0.1.53/0.1.59` 建立 v0.1.36 公開 patch 與本機完整版；Android 有較早 checkpoint 產物 | macOS 真機、Android host audio／真機仍未完成，Android 不屬本輪桌面 release | 不阻塞三平台桌面 release |

綜合判定：**功能層可維持 campaign E3；資料／規則主要為 E2；rec80→rec79 的文字
record／cell layout 已閉合，但整體視覺與聲音仍不能標 V3。**
目前沒有已知的必要主線功能缺口，但仍有可選的原版忠實度、測試穩定性與跨平台同版交付工作。
新發現只有在會改變玩家體驗或交付 gate 時才重新開啟實作；純硬體逐週期、未使用 helper 或
沒有玩家可見差異的完整反編譯不再列為 remake 完成條件。詳細逐畫面矩陣見
[`docs/74`](docs/74-ebiten-remake-completion-plan.md)，開場勘誤與證據限制見
[`docs/188`](docs/188-opening-escort-to-castle-spec.md)，本輪實際抽樣與畫面差異見
[`docs/189`](docs/189-opening-sampled-parity-20260824.md)，rec80→rec79 與逐 glyph timer 證據見
[`docs/192`](docs/192-opening-dialogue-v3-closure.md)。

## 2026-10-01 — Issue #4 原版新遊戲入口與 DRAFT 視窗驗證

使用者明確授權建立及持續更新 Issue 後，建立
[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)；前兩個探測／差異項目已完成，
創角到母親、正式修正及完整驗收保持待辦。Issue #1～#3 未重開。

原版冷啟動到 750,000,001 步，以第 710,000,000／731,000,000 步排入 Enter，
四次 IRQ1 make/break 均記錄實際送達。獨立 dosgolem 副本依 BIOS 公開契約補 AX=1013h，
色盤頁面測試修改前失敗、修改後及完整 DOS／machine 回歸通過；上游與 EXE 唯讀。
磁碟 state 缺 CRTC 等顯示狀態，續跑亂圖分類為執行器限制；正式畫面只由冷啟動重生。
原版、工具來源、Go 版本、binary、實際輸入與產物 hash 保存於私有
`work/dosgolem-opening/issue4-keylog-receipt.json`，未進 Git。

IDA 9.4 非破壞匯出與逐筆語意索引閉合 raw window、record475/451/452/456、字模、
陰影、框線及初始命名游標。`docs/113` 保存輸入 hash、位址基準、推論等級與勘誤：
旋轉只每列執行一次，平面回跳略過旋轉；word AND 保留第二 byte latch。
兩頁試作全畫布 RGB 差異均為零，正式 renderer 仍有 5,508／8,025 像素差異。
新增獨立試作／正式對拍測試與有界 Docker 控制入口；沒有變更 production 或 game-pack。

驗證：Python 編譯、Go 編譯、腳本語法、工具紅綠測試、試作與正式失敗分離檢查通過；
正式 InputState 主線再次到 THE END（144.84 秒），共 14 項受影響測試通過。
起初的 100 秒外層逾時小於既知整段主線耗時，分類為驗證腳本範圍／逾時設定問題；
同一 image 及測試命令改為有界 300 秒後乾淨重跑通過，未寫成產品缺陷。
另修正紅測試診斷比對漏寫 `0x` 的腳本斷言，再重生收據；不是 BIOS 修補失敗。

下一閘門為有限 pack 契約 DRAFT→READY 審查、正式 renderer 修正及創角後對拍；
目前沒有能力擲骰、出生點、母親、完整 campaign 音畫或新發行包的原版完成聲明。
Docker 使用一次性 `--rm`、UID/GID1000、唯讀輸入及有界資源；本批產物持有者已抽查。
工作歷程與提交僅含上述工具、測試和現況文件，保留使用者 scratch／Android libs。
有界控制入口的完整 `--prototype` 執行已通過；原版重生與正式 RED 亦分開重驗。
每次重生前以內容 hash 保留上一批正式收據／畫面，入口及權利分類仍由 docs/113 指定。
本批相關執行中／停止 Docker 容器均清空，沒有建立新 image；全專案歷史 root-owned
候選仍為 3636 項，沒有 `.md` 誤掛載目錄或本批新 root-owned 產物，未廣域修復。

## 2026-10-01 — Issue #4 命名導航與功能切英數

原版冷啟動以 IRQ1 四方向實際輸入確認：左右沿全格盤循環，raw0 左移到44，
raw44 右移到0；上、左到 raw35 才進功能列。模式旗標實際為1→5→2，切英數後
保留 raw35。IDA 9.4 database 重查兩個模式 caller、欄數／格數 writer、導航、
模式交易與游標 consumer，固定輸入／原始位址／bytes／分級與 READY 審查見 docs/113。

首次左、左、Enter 探測錯將 raw43 聲調標成功能格，實際為「查無此字」；保留舊收據
並追加勘誤。IRQ1 因 IF 延後送達，不要求 queued+1；契約改成完整 make/break 在下張
收據前送達，記錄真實步數，不追逐硬體時序。正確功能探測又反證了腳本猜測的
raw0／mode1；以觀察值修正斷言，在同一工具鏈乾淨重跑。這些均與產品差異分開記錄。

實作前兩項獨立回歸為紅：正式導航得到 raw8／raw44，模式測試重設 raw0。
READY 後只改共用全格盤運算與功能切英數的游標保留，沒有新增版本值或資料格式。
修正後六次方向／四次功能原版狀態對拍與 12 項受影響頂層測試通過；正式主線的
主角及酒館三人以功能列切英數，不再用 Toggle，姓名、性別及各段存讀檔通過，
正常輸入抵達 THE END（69.58 秒，執行前固定 remake 0x1357，未重設種子）。
其餘 game 356 項頂層／26 項子測試通過，35 項選用擷取／收據在標準批次跳過，
其中本切片導航已另跑；沒有素材缺失，全部 internal 與 desktop main.go 建置通過。

有界控制入口追加 `--navigation`，完整兩次原版冷啟動重生及正式狀態比較回傳0。
更新已審查語意索引並重生非破壞 sidecar；未知快照保留，沒有批次升格其他語意。
正式命名／英數 PNG 已重生並目視核對，存於私有工作目錄；全畫布視窗／字色／
游標差異仍未修正，導航 CONFORMED 不代表 V3、原版能力擲骰或母親 parity。
收據 `work/dosgolem-opening/issue4-navigation-verification-receipt.json` 保存 image ID、
來源 hash、原版／重製輸入範圍、seed、日誌與權利邊界，hash 見 docs/113。

所有本批來源與產物持有者已核對；一次性 Docker 容器無執行中／停止殘留。
歷史 root-owned 3636 項保持原樣，沒有 `.md` 目錄或新 root-owned 產物。
使用者 scratch／Android libs、原版資產、dosgolem 上游與 IDA database 未納入提交，
沒有新 image 或發行包。提交／推送及精確 HEAD 另回填 Issue #4；
下一步仍為有限 pack 規格與正式 renderer，原版創角／出生點／母親待閉合。

## 2026-10-01 — Issue #4 正式主選單與命名字盤繪圖

依原版 writer、完整字模記錄及冷啟動收據，先在 docs/113 審查有限契約至 READY，
再接入正式 NewGameFlow。共用引擎只執行不透明字模、EGA word-latch 陰影與 XOR，
版面、文字引用與色號在 pack；schema/content 升為0.1.55／0.1.61。
聲母／介音／韻母的既有組字順序抽成共用輸出，34／21 是刪除原行後原值搬移，
不是新設定或另一份 fallback；新繪圖器沒有 DQ3 座標、record 或 flag 常數。
JSON 保留既有格式，沒有混入整批無關格式化。

正式 InputState 的主選單、初始注音、六次方向及四次功能輸入共12張，
完整640×350 RGB 均零差異；兩個有界入口從原版冷啟動重生後再次通過。
原版收據逐筆 hash、IRQ1、模式與 cursor 均核對，沒有裁切／遮罩／狀態注入。
背景及共享色盤保持原樣；主選單、注音與英數 runtime PNG 已目視核對。
EXE／TXT 直接 parity、12個無效 pack 案例及完整 internal 通過。
完整 game 359項頂層／28項子測試通過，33個選用擷取／額外收據未執行，
沒有素材缺失跳過。正常主線含主角、酒館三人功能列命名與各段存讀檔，
由新遊戲抵達 THE END（109.47秒）；desktop main.go 建置通過。
主線 seed 在首個輸入前固定0x1357，未重設或重擲；原版尚未能力擲骰，不能宣稱 RND 對拍。

docs/113 將本表限定狀態標成 CONFORMED，docs/84 登記欄位及新檔入口；
PROJECT_MEMORY、CONTEXT、docs/74、WORKLIST 同步現況。私有收據
`work/dosgolem-opening/issue4-raster-verification-receipt.json` 為13,128bytes，
SHA-256 `4f9fe517d6d0f23fa158f5f3bd6a55d5900620dd11069e876a76d59d21aa8e73`。
候選字、原版能力擲骰、出生點／母親及完整 campaign 尚未閉合，Issue #4 保持開啟。
提交標題為「閉合主選單與命名字盤正式畫面對拍（#4）」；精確 commit／push 由 Issue 回填。

一次性 Docker 容器均採 --rm、UID/GID1000、有界資源與唯讀原始輸入，
本批來源及產物擁有權已驗證。歷史root-owned仍3636項，沒有.md目錄或新root-owned產物；
未廣域修復。沒有新image或發行包，使用者scratch／Android libs保持原樣。
本批相關執行中／停止容器收尾已清空；原版資產、PNG／state／database未納入提交。


## 2026-10-01 — Issue #4 接受角色後生日首頁

由上一合法checkpoint的16次創角輸入延伸第17次IRQ1 Enter，原版先清畫面並顯示黑底
生日旁白；重製提前顯示房間。紅測試完整畫布差異167,706像素。IDA Pro9.4查明
raw DGROUP3E6E／linear28C3E／file19FAE的record404框線、24px字距，以及0xfff5
不吞下一word的姓名插值；舊解析把「16歲」顯示為「6歲」。

先將證據、固定輸入與分級寫入docs/188 DRAFT；隔離繪圖試作完整RGB差異0後審查READY，
才把文字／完整字模框、黑底、顏色、版面、字距及控制碼長度放入opening_prelude。
試作已移除，正式正常輸入兩張首頁零差異，連同既有20張共22張通過；首頁限定CONFORMED。
一般對話維持先前解析，本批不外推全遊戲變數語意。schema/content為0.1.58／0.1.64。
原始文字glyph_codes直接來自D3TXT；value沿用可讀字模索引，不作未考訂續頁的字形權威。

補上33項旁白損壞／巢狀缺欄位拒絕與原始EXE／TXT parity。全部11個internal、桌面main.go
建置與完整game通過：359頂層／34子測試；37項選用跳過，其中4項原版收據另行嚴格通過，
沒有素材缺失。正常新遊戲／主角與酒館命名／各段存讀檔至THE END為102.55秒，僅屬重製回歸。
有界控制入口新增--opening；冷啟動重生17次原版IRQ1後，創角等待／確認與生日首頁均通過。
生成腳本原文納入私有收據；等待圖由錯誤診斷名稱home改為birthday-wait並重生，未換名假造收據。

原版首頁檢視工具曾呈現不完整畫面，後以PNG實際RGB及色號位元組證實兩張完全相同，
沒有把顯示工具問題寫成產品缺陷。新增測試最初漏匯入fmt，修正後在同一容器命令乾淨重跑；
圖片解析最初假設RGB PNG，改依真實indexed PNG格式核對。這些是驗證工具問題，與產品紅測試分開。

docs/94保留舊框線形成史並追加勘誤；IDA匯出同列合併原名、位址、原始MZ bytes與loaded bytes，
附有限分級與警示，自動核對docs/94／84入口。confirmed只涵蓋raw window前12bytes。
PROJECT_MEMORY、CONTEXT唯一現況表、docs/74、WORKLIST同步生日首頁及下一閘門；
README只修正開發版與有限驗收摘要，將舊重包成果明列為歷史checkpoint。
整批私有驗收work/dosgolem-opening/issue4-birthday-verification-receipt.json：19,286bytes，
SHA-256 c0a5cdb30fe9d61fe5c5be5ba947b0bfd1a88cb0ceb7d59dba5666ce6e967ab9。
提交／推送的精確HEAD回填Issue #4；Issue保持開啟，下一步生日續頁，再到房間與母親。
重製既有內部場景預載尚未同原版出生時序閉合，不將首頁零差異宣稱為全狀態或整段campaign parity。

提交前Go掃描命中兩處原始控制碼勘誤註解與共用換行validator；初次稽核錯將所有hex命中
都視為違規，依專案允許的parser／validator／證據註解逐項分類後乾淨重跑通過，沒有新增production fallback。
所有來源與產物UID/GID1000；一次性--rm容器收尾已清空，沒有新image／發行包。
歷史root候選仍3636項、沒有.md誤掛載目錄或本批新root產物；未廣域修復。
原版EXE／TXT／FON、dosgolem上游、使用者scratch／Android libs保持原樣；
圖片、state、原版素材、IDA database／授權與完整私有收據未納入Git或公開Issue附件。

最後核對匯出工具，發現頂層unknown導覽標籤仍寫歷史模板file0x147b；實際target、bytes、
xref與逐項分級均正確。改為由指定target產生file0x13a0／linear0x10030標籤後，以同一IDA
容器重生補充收據work/issue4-birthday-reviewed-final-ida.json（1,369,167bytes；SHA-256
f9f2ae6ec39d7db880b7e028507a3eb31947025e1f58a46d45dd6845cdc27cea）。
舊收據及整批驗收雜湊保持原樣，docs/188追加勘誤；沒有改動遊戲行為或提升語意等級。

## 2026-10-02 — Issue #4 生日續頁正式紅測試與出生返回證據

接續53fb722，從未修改原版冷啟動重播17次IRQ1，再於1,220,000,000與1,340,000,000步
送兩次正常Enter；19次make/break共38次在擷取前送達。種子在自然Lv1入口固定1357一次，
沒有座標、旗標、文字或入口注入。追加純讀取IDA已定位caller與等待事件，不改原版狀態。

原版第18次Enter後，birthday consumer返回linear100AB，才寫DGROUP4F33／4F35=5、5，
經11900與15002後在家中record83的內嵌等待停下。Go仍停生日pos15，多要求一個EOF按鍵；
完整640×350 RGB差異170,238像素。第19次原版已走向後續NPC／record81，Go才進record83，
差異198,874像素；兩側record狀態不同，只記診斷，不宣稱同record繪圖parity。

IDA Pro9.4新sidecar非空、輸入115282bytes／SHA及位址基準均驗證。FFFC在共用consumer內
等待後續寫；FFFF於row3捲動後返回，沒有另一個按鍵等待。docs/188追加帶原始bytes定位、
動態步數與confirmed／strong分級的DRAFT，docs/192追加反向入口；生日前raw15、22沒有被
猜成場景座標。共享文字捲動與出生視野writer／consumer尚需補證，不能把截圖目測寫進pack。

工具新增birthday_continue情境與--birthday-pages有界入口；正式測試由原版收據的19次輸入
與正常空白InputState重播，等待文字穩定後比較，沒有debug狀態捷徑。實際wrapper冷啟動
重跑退出1為已知續頁紅測試，創角8張保持通過；與前一次冷啟動的27張PNG／色號、IRQ1與
16筆唯讀flow逐項一致。既有主選單／導航／創角／生日首頁22張仍零差異；Go測試可建置，
Python與shell語法檢查通過。本輪未改production或schema/content，未重跑完整game／internal／
桌面；最近完整驗收仍為53fb722生日首頁切片，不把局部驗收升格為campaign原版對拍。

最初新增情境漏列收據cursor檢核分支，原版已正常完成而Python收尾報KeyError；修正後
在同一容器與固定輸入乾淨重跑，保留第一次產物。第一版比較在Go剛關閉文字時取到暫態
黑畫面；改用正常空白更新完成下一段轉移後，差異數改為198,874並保留兩份日誌。
映像Entrypoint欄位查詢與技能symlink容器路徑問題均已修正；沒有新增image或退回主機分析。

原版最終收據24,493bytes／SHA-256
7fd546d5b73a65933c5ddd24a34988bef4ca49df18d01074e082e692dce0417c；整批私有證據
work/issue4-birthday-pages-evidence-receipt.json為2,999bytes／SHA-256
e24ffc33a19413e15e91718c2ef3ff3d2d8f3fe7d943b1c6e67482ab7b7d8e32，明列known_mismatch。
歷史flow收據按內容hash保存，重生不覆寫唯一證據。CONTEXT、PROJECT_MEMORY、docs/74與
WORKLIST同步目前DRAFT及下一步；README原有「續頁尚未完成」摘要仍正確，未加入流水帳。

所有本批產物與修改檔UID/GID1000；一次性Docker容器完成後清理，歷史root候選3213，
沒有.md誤掛載目錄，不廣域修復。原版EXE／TXT／FON、dosgolem唯讀上游及使用者scratch／
Android libs保持原樣；圖片、state、database與完整收據不納入Git或公開Issue附件。
本批提交／推送及精確HEAD回填Issue #4，仍保持開啟；下一步補場景視野與文字捲動證據，
試作與READY審查後修正正式續頁及出生交易，不能用額外按鍵掩蓋差異。

已動態審查的返回／出生交易回填自動語意索引，confirmed限於第18次自然Enter後的caller、
5／5交易與record83等待，不提升房間渲染；舊strong側錄保留。新reviewed sidecar
work/issue4-birthday-pages-reviewed-ida.json為1,380,185bytes／SHA-256
82803df5acaa0c31986f670b6fa2566603062ba39cae63834d66b409e21f138d，原始bytes／分級／
出處與反向入口檢查通過。沒有以工具內名稱或database本身代替證據。

## 2026-10-02 — Issue #4：生日返回、出生交易與初始clock正式修正

基準4903f53，遠端Issue #4保持開啟；沿用使用者明確授權的Issue更新及commit／push。
IDA9.4資料庫原始bytes與dosgolem自然19次IRQ1／38次送達閉合兩項有限行為：生日EOF
不再等待確認、初始DGROUP251D word為30。原版clock30選bank0，舊remake零值選bank1。
在隔離來源副本先驗證第18次正常確認可到record83，再以來源審查為有限READY；
共用視野／外界圖塊／字模框與正確色盤試作仍差2,034像素，沒有加入正式來源。

正式pack新增必要return_mode與initial_clock，schema/content0.1.59／0.1.65；缺值、null、
未知原語與clock越界拒絕，保留EXE／TXT decoder與原始bytes parity。
生日EOF末字完成既有hold後自動返回，內嵌等待仍需正常確認；出生場景改在返回後交易，
使用既有opening_escort的CTY／section／首frame玩家位置，不新增Go版本座標或fallback。
初始clock寫入既有可存讀phase／step，不鎖住正式遊戲或覆寫既有存檔。

首輪完整回歸抓到舊測試仍要求生日之前已進房間；改成生日等待→一次確認→自動返回。
局部母親fixture明示建立返回後場景，不冒充正式路線。修正時誤加出生落點移動碰撞檢查，
床上5、5遭拒；依原版直接writer移除多餘檢查，保留失敗日誌後同容器乾淨重跑。
另補末字hold後才返回，避免最後一字尚未可見便切場景；再次完整重跑。

11個internal套件、game362項頂層／36項子測試與desktop main.go通過；
正常新遊戲、主角／酒館及各段存讀檔至THE END（61.09秒）。35項選用標準批次跳過，
其中創角／生日兩項另行嚴格執行，剩餘33項未跑，沒有素材缺失。
既有22張完整RGB零差異，第18次正式輸入到record83等待、seed356D、clock30。
房間正式完整RGB仍差172,261像素；第19次兩側不同狀態，198,652像素只供診斷。
冷啟動wrapper已重生並退出1，原因只有兩項已知RGB紅測試，同批創角8張通過。

最終原版收據25,213bytes／SHA-256
35fccfec8e86c629a04d91a9c0f72e7eba112d5913ab78b55f2d181ba08d9c95；
首輪a235與上輪7fd收據均按內容hash保留。來源、有限CONFORMED、strong視野與
unknown完整文字流分開記錄於docs/188；目前表、工作清單與README同步現行程式，
不把試作或綠色內部測試當成房間V3。所有PNG、色號、收據、原版素材、IDA資料庫留本機。

沿用原有Docker images，原版及dosgolem上游唯讀；一次性容器均清理，其他專案容器不動。
本批修改與產物UID/GID1000，工作樹protected scratch／Android libs保留，無新image或發行包。
提交／推送與回讀結果回填Issue #4；下一步閉合文字保留／捲動、視野／框線、母親及箭頭。

收尾稽核：首輪a235與最終35fc的27張PNG／27份色號、38次IRQ1及16筆flow均相同；
所有56份當輪產物hash核對通過，執行日誌差異分開保存。IDA原始file bytes與完整指令末端
核對通過；新增Go無版本專屬record、flag或座標。隔離試作固定4903f53再次重生2,034像素。
私有整批收據work/issue4-birth-return-evidence-receipt.json為8,159bytes，SHA-256
b537cf20fbf3c1120dcc9855d6c0430401821fbf47fb54a0f581ed10231800f5，
狀態finite_state_conformed_full_raster_mismatch。歷史root候選3,213，本輪新增root產物0、
Markdown誤掛載目錄0；本輪DQ3容器無殘留，其他專案資源未變更。

## 2026-10-02：Issue #4房間陰影來源及唯讀人物影格

正式基準8cebf4a，schema/content0.1.59／0.1.65。本批在Issue #4登記後，沿用既有
dosgolem與IDA9.4容器，沒有修改正式Go或資料包。路由命中逆向重製、dosgolem、IDA及
規格閘門；收尾再查同一路由，未把DRAFT試作加入production。

IDA閉合CTY section+0x12經DH寫入DGROUP0B2D及layer0視野的界外consumer。
家中CTY00 sec4值71與自然flow相同。原始定位、分級與可重生target回填docs/188及
IDA匯出索引，舊sidecar與原始binary保留。writer候選查詢最初因直接import會執行main，
第二次因IDA預設ASCII讀取中文失敗；改成只載入定義及明示UTF-8後在相同容器重跑成功。
兩次都是工具腳本問題，沒有寫成remake缺陷或退回主機分析。

共用視窗的1FB36為背景備份，1FC57為偏移陰影；docs/94舊「清內容」解釋追加勘誤。
21B98只有14條VGA設定指令，沒有額外24px清底。陰影試作固定4903f53來源副本及
第18次正常輸入，完整RGB由2,034降為544；依dosgolem字組鎖存契約處理後，框底242像素
差異消失，完整RGB剩302。兩個NPC步行影格共261，箭頭41，紅測試保留，沒有裁切或遮罩。

唯讀probe新增raw0004／raw26ad及取圖庫指標後的57筆暫存器事件，未增加原版寫入。
母親原始快取索引43、床邊NPC33、勇者第0影格；NPC繪圖時動畫位元1，勇者繪圖時已為0。
中途把母親索引口述為33與勇者102像素分類線索均已訂正；原始目的位址證明後者是床邊NPC。
正式比較仍需補兩側動畫時間條件，不將固定seed這次觀測到的影格永久鎖進遊戲。

同一入口重生19次IRQ1冷啟動，38次make/break送達。新收據31,719bytes／SHA-256
b077e99c3b38f58b3326df8f8d25ed72b74d58f351a71d41374bc5063db4358c；
54份PNG／色號、38次IRQ1與16筆既有flow和35fc歷史收據相同。中間29fd亦按hash保留。
wrapper退出1只因正式房間172,261像素與第19次不同狀態198,652像素；同批創角8張及
生日首頁2張通過。沒有重跑完整game／internal／desktop；最近完整回歸仍是前批61.09秒。

四份已分級IDA匯出與其餘sidecar、生成腳本及完整PNG由work/issue4-room-audit.py稽核。
7,357筆原始file bytes、分級警示、來源hash與UID/GID通過；私有整批收據6,045bytes，
SHA-256 a240c5cff84059c96c20e81fb92e6ff35198b883eb6e79881761b99f3cfc5273，
狀態draft_prototype_full_raster_mismatch。原版素材、圖片、state、database與完整收據不入Git。

CONTEXT唯一目前表、PROJECT_MEMORY、docs/74與WORKLIST更新試作前沿，README維持穩定摘要。
修改檔與本批產物UID/GID1000；歷史root候選3,213無新增，本批root產物及.md誤掛載目錄0。
DQ3／IDA一次性容器收尾為空，未更動其他專案容器、使用者scratch或Android libs。
本批提交及推送回填Issue #4，仍保持開啟；下一步共享文字流及動畫時間條件，再審查READY。

## 2026-10-02：Issue #4生日保留文字與四次捲動正式修正

從e2d6cc2接續Issue #4已登記的生日文字切片。路由命中dosgolem、IDA9.4與規格閘門；
實作前及結論回填前重查spec gate與文件職責。原版consumer證明FFFC先推進行再等待，
確認後保留前文；EOF於最後一行分四次複製320×60、清除底部4列，完成後才返回caller。
原始EXE／TXT身份、IDA linear／file定位及bytes保留在docs/188，不重新命名或修改binary。

probe只新增EOF前及四次捲動邊界擷取，不改19次IRQ輸入或執行前一次種子1357。
固定e2d6cc2的隔離原型六張全640×350 RGB差異0，先追加DRAFT、審查READY，再接入正式
retained_rows引擎。pack新增必填text_flow，schema/content0.1.60／0.1.66；參數、D3來源及
拒絕規則見docs/84。正式確認不能跳過揭露／捲動；捲動中不交易出生、不消耗RNG。
速度沿用3個60TPS更新的硬體規格近似，不宣稱原版wall-clock逐週期一致。

最後冷啟動以同一入口重生33張PNG／33份色號，66份圖像與原型前完全相同，38次IRQ1送達。
正式生日六張0差異、創角8張及生日首頁2張通過；生日六張中的stable與首頁重複，新增5張，
累計限定27張可見畫面。房間正式仍差172,261，第19次不同狀態198,652只作診斷，wrapper退出1。
未接入房間試作302差異，也未把NPC影格或箭頭相位寫死。Issue #4仍開啟，下一步為房間契約。

component、原始EXE parity與巢狀缺欄位／null／未知欄位／錯誤幾何拒絕通過。完整game362頂層／
34子測試、11個internal與desktop main.go通過。正常新遊戲至THE END85.55秒，主角／酒館及
各段存讀檔通過，僅屬重製回歸。標準38項選用SKIP，三項對拍另跑，創角與生日捲動PASS、
接受角色的房間RED；其餘35項未跑，沒有素材缺失。

初次原型容器呼叫的自動核准審查逾時，重試一次成功，沒有退回主機。正式首次編譯缺少
原始文字decoder import，補齊後同命令重跑通過。首次全程trace在最後一字顯示後立即按確認，
尚未等到FFFC；修正驗證腳本只送空白InputState等到正式等待，再於同容器乾淨重跑。
稽核腳本原先要求日誌包含只在失敗時印出的THE END訊息；改核對實際終點斷言及trace成功，
沒有弱化遊戲驗收。這些驗證／工具問題與房間產品差異分開記錄。

最後原版收據36,603bytes／SHA-256
13c5e2632d3fd874f42b21ce061df16b7d3f6681846322ca8c59b79cb21c756f；
原型前fe874與早期b077均按hash保留。新分級IDA sidecar849,255bytes／SHA-256
f5858634c360f87a146348840d19b4383833d667c4475f4974cb61d75bc814e2；
自動附加有限語意，769筆原始file bytes核對通過，不覆寫舊sidecar。
整批私有稽核work/issue4-retained-evidence-receipt.json為21,932bytes／SHA-256
3432cd48b3cb2894538189264d6b779f0ac6c303d6a57dc4a6574551dafc6713。
本機圖片、原版素材、state、database及完整收據不入Git或公開附件。

CONTEXT唯一目前表、PROJECT_MEMORY、docs/74、WORKLIST與README摘要同步。
git diff --check通過；既有image無rg，限定新增Go行的等價regex掃描只命中測試oracle，
新正式原語沒有版本專屬ID／座標／record／旗標／顏色常數。UID/GID1000，歷史root候選3213保留，
本批root產物及.md誤掛載目錄0。dosgolem、IDA與Xvfb容器均已清除，其他專案資源不變。
使用者scratch／Android libs保留，未建立新image或發行包；正式切片提交及推送連結回填Issue #4。

## 2026-10-02：Issue #4有限房間呈現正式修正，完整對拍仍RED

基準5a2101d，主機gh回讀OPEN的Issue #4後登記續行，審查留言5946174168。
現行來源隔離原型302差異後，依docs/188將有限視野／界外圖塊／共用文字框／字組陰影審查READY。
接入正式開場入口與pack，版本0.1.61／0.1.67，移除Go開場record陣列。
完整房間由172,261降至302，人物261／箭頭41仍RED，沒有裁切或指定人物影格。
record83保留前文、83／81自動返回再接帶路；原版NPC更新、母親狀態及時間相位仍未閉合。

只新增步數1222142894的唯讀觀測，擷取是黑色顯示頁，訂正「已完成背景」假設；不作背景oracle。
新原版52ad27c0與原13c5e263的66份PNG／bin及19次輸入／38次IRQ1相同，新增兩份黑頁觀測。
生日六張及創角嚴格通過，累計限定27張保持；第19次196,698是不同狀態診斷。
正式結果見work/issue4-room-production2.log、差異分類見work/issue4-room-production-diff.log。
初次完整回歸受舊schema fixture及trace連按／等待整串關閉影響；修正驗證腳本後，
同一工具鏈乾淨重跑work/issue4-room-full4.log，全game364頂層／34子測試、11個internal與桌面通過。
正常新遊戲→THE END106.51秒，主角／酒館及各段存讀檔通過；標準38項選用SKIP，無素材缺失。

私有稽核work/issue4-room-production-evidence-receipt.json為26,625bytes，SHA-256
63a3f70a88a450ac2822f172556ff3531f0fb3e8554e271f6fb31045aa9e78ef。
原始產物、來源、IDA9.4 raw bytes及UID/GID通過；歷史root候選3213與Markdown目錄4保留。
原版／dosgolem上游唯讀、保留使用者scratch與Android libs，一次性Docker清除，沒有新image／發行包。
依使用者既有授權commit＋push，commit hash與遠端現況由Issue #4末次結果回讀保存。
下一步NPC／箭頭相位與83返回後原版狀態；Goal仍active，不稱完整remake或campaign parity完成。

## 2026-10-02：Issue #4 等待箭頭兩相正式修正

從5029ccf接續剩餘人物261／箭頭41的完整房間RED。IDA9.4追sub_216C3與caller，
證實當前文字行、不透明字模13／12、橫座標336px、顯示8及清除5 ticks，確認後先清除再返回。
原版首次生日箭頭在顯示頁換入前，僅當暫態診斷；下一次完整顯示才列比較。
第一次捕捉逾時240秒，分類為驗證環境；沿用原image與資源，改為360／420秒有界重跑，取得完整收據54a0c21f。
19次輸入、38次IRQ1與舊68份PNG／bin不變，種子各固定1357一次，新增五張相位圖含首次暫態。

固定5029ccf原型生日兩相與六張續頁全RGB零差異，房間兩相均261，審有限READY後才接入正式流程。
必填wait_indicator提供所有版本值與雙份行為／平台證據，schema/content0.1.62／0.1.68，
拒絕缺欄位、null、未知原語、無效字模／幾何、零頻率、無界時長及未審查來源。
正式生日兩相與六張續頁PASS，累計限定28張；房間三張全RGB仍261人物差異，未遮罩或指定影格。
8／5 ticks換算27／17個60TPS更新僅標hardware-spec approximation，不追硬體ISR逐週期。

完整game365頂層／34子測試、11個internal與桌面通過；正常新遊戲到THE END63.35秒，
主角／酒館及各段存讀檔通過，只屬重製回歸。標準38項選用SKIP，原版三項另行兩PASS、一RED，沒有素材缺失。
私有稽核f802ab3a核對80份產物與294筆IDA原始rows，UID/GID1000，歷史root候選3213／Markdown目錄4保持。
本批一次性容器清除，沒有新image、發行包或原版公開附件；所有證據入口與限制見docs/188。
下一步追NPC繪製相位與83返回後NPC／母親狀態，完整原版campaign仍未CONFORMED。

## 2026-10-02：Issue #4 實際PIT除數與NPC0原版序列

接續de91dda的房間人物對拍，先核對NPC reader的全域相位，再補拍record83返回後16個自然動作。
發現生日／房間PIT除數12428，推翻先前預設65536的18.2Hz換算。原始8／5及1tick consumer閾值維持。
新原版1c41e8bc收據118份產物全部核對，116份PNG／bin與前次7ca7da23相同，19次輸入及38次IRQ1不變。
正常Lv1入口seed1357只固定一次，不注入角色位置／旗標，不挑圖片或人物影格。
上游cmd/probe/main.go已有未提交修改，保留；Docker內建立固定2f44a68的乾淨只讀副本再跑。

有限DRAFT與固定de91dda的JSON原型證實生日兩相及六張續頁零差異，審為READY後正式套用。
正式JSON箭頭顯示／清除5／4更新，逐字與每步捲動1更新；schema0.1.62維持，content0.1.69。
新增正常原版觀測守門，核對兩個場景實際除數、8／5tick、四次捲動每步1tick及JSON換算。
原始EXE glyph／scroll bytes測試保留，沒有production Go版本常數或fallback。
完整game365項頂層／34項子測試、11個internal、desktop main.go通過；正常新遊戲至THE END88.40秒、主角／酒館及各段存讀檔通過。
標準38項選用SKIP，創角及生日嚴格PASS，房間RED，餘35項未跑；沒有素材缺失。

原版NPC0由(5,4)到(10,10)後轉左，再顯示record81；主角與caller返回時仍(5,5)。
現行remake選(8,3)人物並於81之後帶動主角的順序已有反證。後續0x1010B正式觸發未知，NPC保持DRAFT。
原始名稱、位址、raw bytes與推論等級見docs/188；三份新IDA sidecar共965筆file bytes核對，沒有rename。
不重開硬體ISR／PIT逐週期研究，不以RNG內部呼叫數一致作完成閘門。

首次原型控制命令未獲Docker socket存取，未執行工作負載；改用已授權的host Docker控制後同image完成。
來源稽核曾把正常確認清除當定時清除；依原始21718分支分類後重跑，301觀測與閾值通過。
自動核准審查拒絕包含本機路徑、逆向細節及雜湊的Issue留言。改為不含上述資訊的工作狀態留言後成功，未上傳原版素材或圖片。
工作留言 https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5951152098 ，Issue保持開啟。
私有來源審查f5087b5a及正式驗收入口見docs/188；提交、推送及清理結果另行回填。

正式私有收據work/issue4-clock-final-evidence-receipt.json為9,414bytes，SHA-256
93d75e9a3c6f03cb63d03eab9b42c05f161f4605e7818e5ce6c9920552307b33。
終點由TestOpeningProductionInputTrace的THE END狀態斷言驗證；標準執行不開選用PNG，不能要求其附加log。
輸出UID/GID1000、歷史root／Markdown目錄候選3213保持；本批DQ3容器已清除，沒有新image或發行包。


## 2026-10-02：Issue #4 母親城鎮42狀態與最後交易修正

接續ed27d71，原版冷啟動38次正式輸入、76次IRQ1、seed1357一次，自然續跑至1020A。
176個唯一產物hash／大小通過，168個既有PNG／bin重生且相同。前輪210列實際170檔、40重複，
原始收據不覆寫；生成器已拒絕重複。原版EXE、dosgolem上游及使用者未提交資料均保留。

DRAFT經原始EXE／CTY／TXT與完整動態consumer審查為有限READY，再實作城鎮切片。
以原始NPC record0識別母親；pack宣告42個兩者位置／方向，單一record80引用穩定文字ID，
FFFC確認後走到21,17，最後交易set17h／clear50h。版本為schema0.2.0／content0.1.70，
缺actor、必要欄位、null、未知引用或不合法移動均失敗即關閉；舊schema存檔拒絕，無自動遷移。
修改前同條正常輸入在城鎮第一步FAIL；修正版42狀態與同版本存讀檔PASS。家中流程未一併提升。

完整回歸暴露測試策略與斷言問題：取船後與尼羅肯特返港未住宿、聖水容量未扣既有存量、
分階段Boss未用已學的封咒／致盲／防禦且同回合重複安排治療。
測試改走正式旅店及戰鬥選單，治療策略考慮同回合已排命令；正式遊戲數值、規則及seed保持。
幽靈船、海岬、蓋亞之劍、雲雨之杖與彩虹合成的隊伍獎勵仍被部分測試只查勇者背包，
已依現行writer修正為全隊檢查，旗標、道具存在／消耗及存讀檔驗收保留，六珠消耗檢查亦涵蓋全隊。

最終完整game367頂層／34子測試、11個internal與desktop main.go PASS，正常新遊戲InputState至THE END65.74秒。
標準38項選用SKIP；原版創角與生日捲動另行嚴格PASS，完整房間與續行仍RED，無素材缺失。
一批原版比較忘記啟用選用參數而SKIP，未當驗收；啟用後同容器重跑得到上述PASS／RED。
房間三圖各261、家中續行198049、城鎮返回83387像素差異，保留完整畫布、不裁切或遮罩。
城鎮水平視野差32px，動畫、音訊、家中NPC順序／圖像選擇／正常接近及完整原版campaign仍未完成。
本輪只閉合共同城鎮checkpoint的有限狀態，不宣稱兩側38次同輸入或整個開場V3。

docs/188保存證據與READY範圍，docs/84保存契約；docs/192、189、66、74追加原始位址勘誤與回鏈。
所有本機收據入口見docs/188。Issue工作與進度留言為5952414514、5953241717，Issue保持OPEN。
Docker、來源擁有權及提交／推送收據於本批收尾另行核對；沒有新image、發行包或原版公開附件。

本批收尾已核對原版176個唯一產物、42狀態收據、程式／資料／文件及失敗即關閉的勘誤回鏈。
私有稽核入口為work/issue4-arrival-final-audit.py，輸出work/issue4-arrival-final-evidence-receipt.json。
所有本批變更與輸出UID/GID1000；指定find檢查得既有root候選3213、Markdown目錄0，未新增root檔。
衛生腳本曾錯用較早Markdown目錄4的假設；以同容器指定find重跑核對實際0，分類為驗證腳本問題。
本批DQ3一次性容器已清除，dosgolem上游cmd/probe/main.go使用者修改保持，沒有新增image或交付物。
提交、推送、Issue最新狀態與Docker清理記錄由work/issue4-arrival-post-push-receipt.json保存。

## 2026-10-02：Issue #4 家中原版收據、選圖來源及現行版本試作

接續c4d2a3c，已先回讀Issue #4並登記續行，留言5954641898。
新增mother_home_entry情境的唯讀選圖觀測，原版自然冷啟動、seed1357一次、37次正式輸入／74次IRQ1。
收據4063460c的170個唯一產物全部核對，168個PNG／bin與前次完整返回收據相同。
核對母親16動作、三輪選圖／結果確認、13次手動接近及轉場鏈；不宣稱remake parity。

IDA9.4補足視窗、游標、BLS consumer及動態選圖writer，十份sidecar共2,008筆唯一原始rows核對。
直接xref缺少DS相對欄位，新增保留原operand及取址候選的匯出，不猜讀寫。
定位16F4B從BIOS046C初始化獨立generator；09F1與2B7A是runtime資料，不能當EXE靜態表。
初次來源驗證誤把這兩區當靜態EXE bytes，依原始writer更正分類後同原版收據重跑通過，屬驗證腳本問題。
DQ3LIN.BLS實際46,086bytes、96個480bytes圖塊；單次圖號13／18選項不寫成正式固定初值。

固定現行c4d2a3c，以正式17／18／19次InputState測試，原提交在母親移動前先播81而FAIL。
可丟棄人物試作17狀態PASS，母親NPC0先到10,10轉左，勇者保持5,5，再播81。
沒有改production Go、pack或存檔，選圖、正常接近與家中存讀檔仍須一併READY後實作。
素材原型把透明區域當黑底的假說已保留反證；僅核對不透明素材仍有7色號差異，維持RED，完整RGB未驗。

新增原版收據驗證入口tools/verify_dosgolem_home_entry.py，十種損壞收據拒絕通過；shell／Python語法通過。
現行原提交FAIL與隔離試作PASS、來源／IDA raw bytes及圖像RED的私有稽核為
work/issue4-home-final-evidence-receipt.json，SHA-256 ea67ada60442d1f4c116dca7106466fb458bf8bf8ef4de9523ca43747024ce99。
重生模式為tools/verify_dosgolem_newgame.sh的--mother-home-original；所有工作均在一次性Docker執行。
16FFA／17006的triplet原始規則已取得；下輪閉合clock初始條件、預覽及取消分支原型，再接完整家中玩家流程。
全域人物相位、音訊及原版campaign仍未完成。
清理、提交、推送與Issue回讀結果另存本機收尾收據，不把本輪來源工作算成正式remake完成。

## 2026-10-03：Issue #4 家中正式流程接入

補記已推送b7516c5的有限切片。依docs/188的READY，母親17個人物狀態先於record81，
預覽、三輪選圖、結果及13步正常手動接近接入正式玩家流程。獨立圖像種子、100組問題表、
實際96張BLS、視窗、色盤及游標XOR由原始資料閉合，全部版本值放進JSON。
schema0.3.0／content0.1.71；Escape保存當前選項，不猜答錯處罰。
家中標題正常讀檔、城鎮42狀態與最後旗標交易通過；五個336×192視窗RGB差0，
完整畫面仍各124個人物差異。第一次完整回歸被容器OOM中止，按Docker事件分類為環境，
同工具鏈設定GOMEMLIMIT2GiB乾淨重跑通過。game370頂層／45子、internal138頂層／193子及11套件、
desktop建置通過；正常新遊戲至THE END224.65秒，只屬remake可玩回歸。
來源、正式PNG與最終私有核對入口見docs/188；Issue留言5957335599，沒有新發行包或公開原版附件。

## 2026-10-03：Issue #4 共用NPC相位反證與城鎮camera修正

接續b7516c5，遠端工作登記5957380305。dosgolem固定乾淨2f44a68，
原版完整啟動37次正常輸入／74次IRQ1，能力seed1357只設定一次。
新來源0b57ba81的170個唯一產物全部核對，168個既有PNG／bin與contract相同。
12180次遊戲計數、2030次六計數翻轉、18532個NPC reader pair及37次輸入時鐘全部閉合。
虛擬Ticks757→761期間遊戲計數只進一次；驗證器核對原始遞增序列，沒有增加九tick例外。
十種語意損壞收據在暫存副本拒絕，原始來源保持。非凍結reader為confirmed，凍結動態樣本0，分支維持strong。

IDA9.4一次性database保持原始名字、位址與bytes。新八筆已分級台帳由匯出自動合併。
首次核對發現八個loaded／file差異，全部對應原始MZ relocation；補足兩種bytes及原始relocation對照後，
同image重生、423列raw核對通過。這是驗證工具問題，不當成EXE或產品缺陷。
私有匯出、台帳、來源驗證器與負向案例入口全數掛回docs/188。

房間同次繪圖NPC在tick6902取phase1，主角於tick6904取phase0。
單次全畫面phase原型仍房間261、選圖124，生日保留文字仍PASS；假說已推翻，原型未進production。
最小充分writer／reader證據已取得，動畫時序保持DRAFT，不追ISR或硬體逐週期。

轉到同Issue既有城鎮視野工作，登記5958225007。原始renderer為玩家減9、7、不clamp，
目的CTY00 section0的原始界外圖塊0。固定b7516c5同條正常創角、選圖及接近的完整城鎮最後PNG，
修改前81962像素差異，隔離camera原型0差異；沒有改人物frame、seed、旗標或輸入。
原型初次區域變數名衝突，修正測試後同來源重跑，分類為原型工具問題。
有限READY後正式新增必填arrival_camera，schema0.4.0／content0.1.72。
所有版本值由JSON提供，destination場景及標題正常讀檔選相同camera；其他場景不外推。
十種契約拒絕、EXE／CTY parity、42個正常人物狀態與最後完整640×350 RGB差0通過。
家中兩條正常玩家路徑、五視窗差0及完整124限制維持。完整回歸及提交／推送在本批收尾記錄。

正式完整game370頂層／45子、38項選用SKIP通過，正常新遊戲至THE END423.50秒。
internal首次失敗來自六個錯誤資料測試的舊schema fixture，提前在版本gate被拒絕。
fixture改使用現行SchemaVersion，原路徑、長度、未知欄位及事件拒絕斷言保持；
同工具鏈乾淨重跑internal140頂層／203子及11套件PASS，4項選用SKIP，desktop建置PASS。
沒有素材缺失跳過。已通過的完整game沒有因fixture修改重跑。
兩張最後640×350 PNG已目視核對；正式最後畫布零差異只限本狀態，不提升房間、家中、音訊或原版campaign。
遠端結果留言5958611871及Issue主文已更新，Issue保持OPEN。私有完整收尾入口為
work/issue4-town-camera-final-audit.py及issue4-town-camera-final-receipt.json，索引在docs/188。
最後以實際InputState.Enter重播原版掃描碼，42狀態、最後完整RGB零差異與標題讀檔再次PASS，
log為work/issue4-town-camera-enter.log；只有測試輸入映射與註解調整，已通過的產品回歸保持。

## 2026-10-03：Issue #4 正常進城攝影機修正

接續de52686，工作登記5958958481及來源進度5959336233。dosgolem固定乾淨2f44a68，
保留38次正常母親返回輸入，再送九次上鍵；能力seed1357在自然入口只固定一次。
第二次乾淨來源c8548e84有47次輸入、94次IRQ1、198個唯一產物，174個原先PNG／bin逐項相同。
原版自然抵達CTY25 section0的(15,30)，沒有座標、旗標、文字或影格注入。
定時抓圖包含更新中狀態，不能宣稱九個完成影格。唯一第九步等待由閒置狀態窗caller進入，
IDA9.4原始堆疊、caller與consumer閉合；4F25／4F27在renderer返回後重用，不能把當時(6,38)當視野。
第一次wrapper因執行中編輯腳本而解析失敗，探針與獨立核對成功；固定腳本後同命令乾淨重跑退出0。
第一次來源依內容雜湊完整歸檔，未覆寫歷史或當產品缺陷。

原始EXE的玩家減9、7及CTY25 section0界外圖塊0提供最小充分幾何證據。
固定de52686的正常進城基準定時圖差171566，隔離camera原型降至2775，完整等待圖由170381降至12726。
有限READY後正式接入scene_cameras，schema0.5.0／content0.1.73；九份JSON同步，所有版本值保留在pack。
引擎只查scene引用並消費既有player_anchor。建立Game前以原生Town parser驗證CTY、section與exterior，
缺素材、未知引用或資料不符均拒絕。parser保留section+12原始欄位，production Go沒有新增版本ID／座標／旗標或玩家文字。
八種契約拒絕、原始EXE／DAT parity、無素材與三種來源錯誤拒絕通過。
正式47次InputState自然進城，Save及正常標題Load維持camera、位置與旗標；城鎮最後全RGB差0保持。
正式PNG與隔離camera原型逐byte相同；城堡完整等待圖仍12726差異，只閉合camera幾何。
閒置狀態窗、8000h圖塊consumer、謁見、音訊及原版完整campaign保持待驗，不猜屋頂或指定動畫影格。

完整單程序game兩輪均於最後進城測試異常退出，第二輪Docker監看記錄明確OOM事件。
兩輪正常新遊戲至THE END已通過，但不能將異常退出整批記PASS。排除主線的大批亦晚段退出。
依環境路由，用相同binary、容器4GiB／2CPU、GOMEMLIMIT2GiB及原斷言，將長主線與其餘四批分成新程序。
核對410個頂層完整清單聯集相同且無重複；game372頂層／48子PASS、38選用SKIP。
正常新遊戲InputState至THE END130.22秒，僅屬重製可玩回歸。最後正常進城及標題讀檔6.57秒PASS。
internal首輪三個拒絕案例因fixture缺少新必要集合而提前拒絕；明示空集合後保留原斷言，乾淨重跑PASS。
最終internal142頂層／211子及11套件PASS、4選用SKIP；desktop建置PASS，沒有素材缺失跳過。

收據驗證器另核對新增18次IRQ的實際掃描碼、送達時間及完整等待PNG／bin。
八種損壞來源在暫存副本全部拒絕。Python／shell語法、Go格式與git diff --check通過。
私有稽核入口work/issue4-castle-camera-final-audit.py及結果issue4-castle-camera-final-receipt.json，
連同正式PNG、JSON、所有分批清單、失敗log、IDA及來源工具均掛回docs/188。
輸出UID/GID1000，歷史root候選3213、Markdown目錄0保持。本批一次性容器均已結束，未建立新image或發行包。
提交、推送、遠端Issue更新與最後Docker清理另存work/issue4-castle-camera-post-push-receipt.json；Issue保持OPEN。

## 2026-10-03：正常進城圖層遮蔽修正

依Issue #4留言5960375932接續0900bd2。命中原版畫面對拍、規格閘門、IDA及文件職責路由。
原版EXE／CTY唯讀，以IDA9.4追section表頭writer、玩家格selector與圖塊consumer；原始定位及等級見docs/188。
原版冷啟動47次正常輸入／94次IRQ1重生198份唯一產物，196份前輪PNG／bin逐項不變。
四格八筆唯讀觀測閉合正常零層的替代圖塊，沒有注入位置、旗標、layer或影格。
固定0900bd2的隔離圖層原型從正常新遊戲重播，前閒置完整640×350 RGB由2775降為0。
依有限READY接入scene_tile_layers；schema0.6.0／content0.1.74，版本資料與圖塊引用留在JSON。
原生Town parser保存section+15／+16並解typed layer；共用renderer只選已宣告規則，繪製不改碰撞或事件。
建立Game前核對實際CTY／section、header值及正常loader的實際BLK count，未知或越界拒絕。
12種pack契約損壞、5種runtime來源損壞、表頭截斷及原始EXE／CTY parity通過。

正式47次InputState前閒置完整RGB差0，PNG與隔離原型逐byte相同並目視核對；完整等待圖仍9951差異。
城鎮42狀態及全RGB零差異保持；標題Save／Load維持camera、layer、完整hiMap、位置與旗標。
新讀檔零差異斷言兩次失敗304，診斷朝向1／0且差異全在主角。現行saveState沒有朝向，原版讀檔oracle未知。
重查同狀態路由後，保留正常47次嚴格零差異與讀檔狀態斷言；保留完整讀檔PNG及304差異，明示原版oracle及讀檔parity均false。
未注入朝向、裁切或遮罩來消除差異；未改正式存檔行為。此新斷言的同狀態假設失效，不記成已證實的原版讀檔產品缺陷。
隔離試作曾有shell引用與helper回傳值編譯錯誤；分批runner曾誤指不存在的binary，均屬工具問題。
修正工具後以相同隔離流程乾淨重跑，未改正式規則或seed。

完整game373頂層／53子PASS、38選用SKIP；411頂層由長主線與四批全新程序完整覆蓋，無重複或遺漏。
正常新遊戲至THE END184.69秒，只屬remake回歸；internal145頂層／223子、全部11套件PASS，4選用SKIP；desktop ELF通過。
13種原版收據損壞均拒絕，新layer負例同時修改log與metadata；三份IDA sidecar逐列核對file bytes與MZ relocation。
最小充分稽核work/issue4-castle-layers-final-audit.py輸出issue4-castle-layers-final-receipt.json，9902bytes，SHA-256
`9174a2ce5ae1df4a0c91d66677850c5308dedb1cbff7b66368cfba3b5faba3eb`；全部私有產物入口掛回docs/188。
Python／shell語法、Go格式及git diff --check通過；新增production Go無版本raw ID／座標／旗標／玩家文字，原生格式與驗證域除外。
現況同步CONTEXT唯一表、PROJECT_MEMORY、docs/74、docs/84及README；遠端結果留言5961178819，Issue保持OPEN。
輸出UID/GID1000；既有root候選3213、Markdown目錄0維持。本批一次性容器已移除，未建立image或新發行包。
提交、推送、遠端Issue核對及Docker清理另存work/issue4-castle-layers-post-push-receipt.json。
下一切片閉合閒置狀態窗後正常續行謁見；非零層轉換、NPC遮蔽、原版讀檔朝向、音訊與完整原版campaign保持待驗。

## 2026-10-03：Issue #4 閒置狀態窗原版來源與布局試作

接續bcc6ce0，工作登記5961286946，進度5961717243。命中dosgolem、畫面對拍、GUI還原、規格閘門、IDA及平台規格路由。
以既有IDA9.4 image及同一唯讀EXE追正常閒置caller、動態視窗header、隊伍內容、三位數前導空白與色盤consumer。
三份有界sidecar共1740列核對file bytes與MZ relocation，保留原名、定位、xref type及分級警示；沒有改原版或database原始身份。

新增king_idle來源情境與--king-idle-original重生入口，沿正常47次進城再送兩次上鍵。
首輪900秒工具逾時，最後92次IRQ1，沒有完整閒置收據。原始失敗log按hash歸檔，分類為工具期限不足。
只將本情境期限改1200秒、外層1260秒，以同輸入、seed、原版、來源及image乾淨重跑，退出0。
原版收據5cd11f15為49次輸入、98次IRQ1、216份唯一產物；196份前輪PNG／bin逐項相同。
觀測到三次自然開窗、兩次上鍵關窗。兩次鍵都消耗於關窗，位置維持15／30，58bytes角色資料在35筆觀測保持。
三次入口計數差值皆300，前面各有298／299／300，PIT ticks與遊戲計數各遞增一；實際除數12428保持。
只驗有限遊戲邊界，不深入ISR或宣稱原版硬體wall-clock一致。靜態pop與動態pop前觀測分開保存。

固定bcc6ce0的隔離布局試作從正常新遊戲及47次正式InputState續行，再組合DRAFT布局，不接正式timer。
首次誤讀目前城堡對話bank差1299；第二次差1315。重查GUI與同狀態契約後，定位為前景色、未覆蓋冒號與主角動畫。
依原始D3TXT00 frame、前導空白glyph及正常3C色盤重跑，視窗與陰影區域RGB差0，整張仍差182。
原型PASS只表示組合及輸出成功，不當成完整RGB或正常閒置生命週期完成。未裁切、遮罩或指定角色frame。
原版兩次restore的完整圖比較均差182，全部在主角；第三次等待差182，第五次等待差1811，完整差異保留。

來源驗證器核對輸入、IRQ1、角色內容不變、完整開關順序、三輪PIT邊界、原始header及所有產物。
16種損壞收據全部拒絕；暫存log與metadata同步修改並重算manifest，正式來源保持。
私有收尾work/issue4-idle-status-final-audit.py輸出issue4-idle-status-final-receipt.json，12469bytes，SHA-256
`0e9ce1b1db36a3a1a64737094e171ad615102b3b6193f927b023e45e24d9ab1c`；所有新入口掛回docs/188。
Python／shell語法通過，輸出UID/GID1000；既有root候選3213、Markdown目錄0保持。
正式Go、pack與存檔未改，schema0.6.0／content0.1.74維持，前輪完整產品回歸不重跑。
CONTEXT唯一狀態表、PROJECT_MEMORY、docs/74與docs/188同步，README穩定產品摘要維持。
本批一次性容器均已結束，未建立新image或發行包。提交／推送及遠端Issue結果留在同前綴post-push收據。
下一步審查閒置狀態機的輸入所有權、平台時間換算及存讀檔暫態，再經READY接入正式JSON與玩家路徑。
結果留言5961959766及Issue主文已更新，正式閒置UI仍未勾選完成。

## 2026-10-03：Issue #4 正式閒置隊伍窗

接續acc3819，工作登記5962141169。命中dosgolem、GUI還原、對拍、規格閘門與平台時序路由。
沿用原版49次正常輸入／98IRQ1的5cd11f15收據，不注入位置、旗標或人物影格。
IDA9.4補足原始文字reader兩份窄匯出，五份sidecar共1901列核對file bytes與MZ relocation。
原始DI24Ch應為十進位588，前輪596的轉換錯誤在docs/188追加勘誤，原始定位與bytes保持。

有限READY後接入正式等待／開窗／按鍵消耗狀態機；只宣告已有來源的正常城堡場景。
所有版本文字、字型、列位置、狀態mask、健康色及時間參數由pack提供。
字型驗證後直接綁定renderer，HP／MP／狀態列依原始add dx指令遷入JSON，不留引擎版本座標。
schema0.7.0／content0.1.75；九份JSON由acc3819乾淨pack重建，逐byte相同。
300個原版tick依公開PIT契約換為188個60TPS更新，精度維持hardware-spec approximation，未深入ISR。

正常新遊戲、創角、家中選圖、城鎮42狀態及九次上鍵進城後，三次自然開窗與兩次關窗通過。
方向鍵只關窗，持續按住不移動；放開後新按鍵可續行。開窗凍結畫布與RNG。
完整存檔快照逐byte不變；同一Game的Load及正常標題Load清除等待、底圖與按鍵暫態。
三次視窗及陰影RGB差0；完整五圖依序182、182、979、979、1960，差異保持。
完整等待第一圖從較早缺窗9951降至182；前閒置全RGB零差異與讀檔304個主角差異仍各有獨立限定範圍。
第一張正式PNG與布局試作逐byte相同，但本次由正常等待狀態機生成。
其他隊伍、健康色及status動態、謁見、原版讀檔、音訊與完整原版campaign未驗，不宣稱完整V3。

十五種壞JSON、五種壞來源、五個健康色分支及原始EXE／TXT parity通過。
初次原生資料測試誤取EXE immediate偏移，按IDA原始bytes改正後同工具鏈乾淨重跑；分類為驗證腳本問題。
完整game414頂層清單分長主線及四批新程序，376頂層／63子PASS、38選用SKIP。
internal147頂層／238子及11套件PASS、4選用SKIP；desktop Linux x86_64 ELF通過，無素材缺失跳過。
正常新遊戲InputState至THE END226.04秒，只屬remake可玩回歸。
十六種原版收據損壞均拒絕；負例同時改暫存log與metadata並重算manifest，來源保持。

私有收尾work/issue4-field-idle-final-audit.py輸出issue4-field-idle-final-receipt.json，22245bytes，
SHA-256 `4135934fa5128573b0f641e4197907ac0fda5ef0514733ed89dd78bbb4c05b74`；新工具與產物索引掛回docs/188。
Go格式與git diff --check通過。新增production Go掃描僅命中typed CTY引用與legacy_record來源驗證，沒有版本raw ID／座標／旗標或玩家文字。
輸出UID/GID1000；既有root候選3213、Markdown目錄0維持。一次性容器已結束，沒有新image或發行包。
遠端結果留言5962628881及Issue主文已更新，Issue保持OPEN；提交／推送與最後容器清理另存work/issue4-field-idle-post-push-receipt.json。
下一切片從關窗後正常續行謁見，先取得原版輸入、狀態、畫面與副作用證據。
