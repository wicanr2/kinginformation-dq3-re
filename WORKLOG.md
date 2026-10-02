# DQ3 工作歷程

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
