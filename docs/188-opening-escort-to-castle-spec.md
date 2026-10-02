# 188 — 開場連續演出勘誤與修正規格：家中 → 王城入口 → 國王

## 2026-10-02：家中序列與圖像選擇的來源核對及現行試作（DRAFT）

依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4) 接續 `c4d2a3c`。
正式程式仍為 schema0.2.0／content0.1.70。本節沒有套入家中 production 修正。
先前城鎮42狀態、record80、後三步及旗標交易的有限驗收維持，不把它外推為家中流程已符合。

原版再次由冷啟動、自然創角入口固定seed1357一次、37次正式輸入及74次IRQ1重生。
收據 `work/dosgolem-opening/issue4-home-entry-receipt.json`，SHA-256
`4063460ca46d743e1e8f7a3e34f6956d7e2306801e9180276988d5f4eaf00581`。
170個唯一產物的大小、雜湊及UID／GID1000全部核對，168個PNG／bin與前次母親完整返回收據相同。
只有一次測試seed設定，其他觀測唯讀；沒有座標、故事旗標或進入點注入。
dosgolem固定revision `2f44a68ebfc54b28fb15dd4a34510b0b04a5415d`、Go1.24.13及既有Docker image。
來源EXE仍為115,282bytes、SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
IDA Pro9.4；IDA linear−0xEC90=file，DGROUP基底linear0x24DD0。

| 原始定位及consumer | 有限結果 | 分級與邊界 |
|---|---|---|
| DGROUP3CBC／file0x19DFC，100D5→167C5→100F2 | 83返回後NPC0左2、下6、右7、轉左，主角全程(5,5)，之後才播81 | confirmed；同次自然16動作、原始序列bytes及8tick等待閉合，動畫相位未驗 |
| linear0x21DDC、21F79、21E3E、21E94 | 第一次Enter開始選圖，三次Enter選三輪，最後Enter確認結果；結果AX1、choices010101 | confirmed；只限本次正常五次Enter，不把其他選項或取消算成動態驗收 |
| DGROUP00DB／file0x1621B，22017→1EA8C | 選圖從DQ3LIN.BLS載入18個480bytes的masked frame | confirmed；唯讀檔名、實際開檔、BLS形狀及自然18次consumer閉合；完整RGB仍未通過 |
| DGROUP4348／file0x1A488，1F590→1F779 | 靜態視窗與六選項的共同選擇consumer；佈局模式4 | strong；runtime與EXE bytes一致，完整視窗／游標尚未試作驗收 |
| linear0x1E932..0x1E96C，DGROUP2B7A | 執行時生成480bytes步距表，原始迴圈200個word；此次只核對前66個及實際選項 | strong；直接writer bytes、自然偏移表及檔案索引閉合，不稱未抽樣槽位全部可用 |
| linear0x16F4B..0x16FCE，DGROUP09F1／09F2 | BIOS046C取初值，獨立generator選0..99圖號，取08C0三個值後生成三輪選項 | strong；IDA caller、取址、writer及直接triplet consumer閉合，clock初始條件與完整UI未驗 |
| linear0x16FDE..0x1702E／file0x834E..0x839E | 正確triplet與隨機triplet的前後順序由generator低bit選擇；triplet為floor(id/3)×3起的三個連續圖塊 | strong；原始bytes、100組table及本次三輪選項一致；不去重同組候選，不永久固定seed |
| linear0x1010B→10121→10130 | 下2／左2／下3／右6走到(9,10)才觸發；母親再到(10,11)，轉場後主角(8,38)、母親(8,37) | confirmed；本次同條正式接近重驗，其他方向與失敗gate仍未驗 |

DQ3LIN.BLS為46,086bytes，SHA-256
`f4a7218ec17147188607fc9b61f7e359278cbf9b716e5afc8776d77958cccd0c`。
原始6bytes表頭是byte-width4、高24、count96，實際檔案也正好6＋96×480。
本次原版的18個選項由runtime生成，不是EXE靜態資料，不得當成正式固定選項。
同理，單次顯示圖號13不是永久初值，也不能用創角seed替代另一個以BIOS clock初始化的generator。

重生入口：
`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --mother-home-original`。
收據驗證入口：`tools/verify_dosgolem_home_entry.py`，只在Docker內執行。
驗證涵蓋來源身份、seed條件、按下／放開、唯一產物、素材形狀、16動作、三輪選圖、結果確認及手動轉場鏈。
十種損壞收據另行驗證均拒絕，包括注入狀態、未先固定seed、錯誤版本、輸入／IRQ缺失、
重複產物、越界路徑、雜湊不符、來源不符及缺少選圖結果。
成功只表示原版收據通過核對，不表示remake對拍通過。

### 現行版本的可丟棄試作

`work/issue4-home-current-prototype.py`固定目前提交`c4d2a3c`，容器內從Git唯讀匯出到tmp。
正式17次創角輸入、18／19次確認，逐次觀測實際`InputState`，不呼叫事件helper或注入人物。
原提交在83返回後立刻播放81，沒有16動作，測試FAIL；隔離試作依runtime證據宣告17個人物狀態，
先移動原始NPC0到(10,10)並轉左，勇者全程(5,5)、另一NPC仍(8,3)，之後才播放81，測試PASS。
log為`work/issue4-home-current-baseline.log`與`issue4-home-current-prototype.log`。
原型未實作圖像選擇、手動接近及家中存讀檔，不能進入正式路徑。

選圖素材試作`work/issue4-home-sprite-prototype.py`使用原始1EF4h／84及每格4bytes公式定位。
第一次把透明區域當黑底的假說已推翻，原始失敗保存在`issue4-home-sprite-black-background-rejected.json`。
依BLS mask只核對不透明素材後，18個選項仍有7個色號差異，均在第三輪第二個圖塊的最左欄。
此結果保持RED，不遮掉七個像素、不當完整RGB，也不推測為已解出的游標或硬體問題。
透明背景、調色盤、完整視窗與正式UI仍待驗證。

### 證據匯出與下一個正式閘門

私有IDA sidecar為`work/issue4-home-{picture,window,selector,window-range,box,table-writer,indirect-writer,options-init,offsets-init,triplet}-ida.json`。
十份共2,008筆唯一原始rows逐項核對EXE bytes及linear/file基準。
保留原函式、原始位址與xref type；附加台帳不覆寫原始語意。
匯出工具`tools/ida_dump_opening_handler54.py`補上DS相對取址及位移候選，均標unknown，不依operand位置猜讀寫。
零直接xref未用來宣稱沒有writer；選項writer由IDA資料庫的09F1原始operand定位到16F4B。

完整私有稽核入口`work/issue4-home-final-audit.py`，結果
`work/issue4-home-final-evidence-receipt.json`，SHA-256
`ea67ada60442d1f4c116dca7106466fb458bf8bf8ef4de9523ca43747024ce99`。
16FFA／17006的triplet規則已取得最小原始證據；下一步閉合clock初始條件、預覽與取消分支原型，及pack圖像選擇資料和有限狀態機，
再將家中母親順序、選圖、正常接近與同版本存讀檔一併審為READY並實作。
NPC全域相位、完整RGB、音訊與完整原版campaign仍未CONFORMED，不重開PIT／ISR逐週期研究。

## 2026-10-02：母親正式入口追查（DRAFT）

依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4) 接續 `ad00d38`。
本節只補足原版入口證據，不把尚未閉合的觸發寫入正式路徑。
來源仍為唯讀 `assets_raw/DQ3.EXE`，115,282bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；
IDA Pro9.4，linear−0xEC90=file，DGROUP基底linear0x24DD0。
CTY來源 `assets_raw/CTY00.DAT`，7,546bytes，SHA-256
`ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836`。

| 原始定位 | 有限語意 | 等級與邊界 |
|---|---|---|
| linear0x289F0／file0x19D60，raw `0B 01` | 指向原始 `sub_1010B` 的資料交叉參照；近位址010B，原始CS1000 | strong；保留原名、資料xref type1，不因直接code xref為零而判定無入口 |
| DGROUP3BB4／linear0x28984／file0x19CF4 | 兩位元組分派表；raw index54選到上述010B | strong；54是原始索引，不能混用歷史catalog的index+1標籤 |
| linear0x196D2..0x1970D／file0xAA42..0xAA7D | section+4的清單以DGROUP258C−1取handler，再乘2索引DGROUP3BB4，最後間接call `[SI]` | strong；caller為原始19574，待正常接近的動態閉環 |
| linear0x11B05..0x11B1A／file0x2E75..0x2E8A | 目標cell高位元組低5位非零，且tile attr沒有C8遮罩時，寫4F46位元0800及258C selector；之後提交玩家位置 | strong；caller119C8→11A6A，觸發仍受阻擋與其他gate影響 |
| CTY00 sec4 base file0x1383，handler file0x13B2 `36` | selector1對應raw handler54；(10,9)、(9,10)、(11,10)、(10,11)四格都是tile4／subid1 | strong；母親10,10所在tile21不同，不能把樓梯轉場當成唯一事件入口 |

新匯出 `work/issue4-mother-{entry,dispatch,selector,trigger,movement,tile,modal,modal-input}-ida.json`。
全部3,181列指令以原始MZ file bytes及linear/file換算逐列核對，沒有重新命名函式或改動原始資料。
稽核入口 `work/issue4-mother-entry-audit.py`；完整原版自然入口尚未核對前，這些語意維持strong。

### 正常輸入探測勘誤

第一次32次輸入的冷啟動收據有64次IRQ1送達，卻停在原版圖像選擇選單。
它沒有抵達母親入口，不能作為走近觸發或remake parity證據。
原始19530先檢查4F46位元4，呼叫21DDC；該流程先進入選單，再由21EC1呼叫21F79三次選圖，
最後21E3E等待結果確認。正常輸入需五次Enter；原始21F04..21F18的比較與NOP保持原樣。
不修改選單驗證、旗標、座標或程式進入點。

第二次探測漏了進入選單的確認，第一個方向鍵仍用來關窗；原先向左的通路還碰到床。
該批另在任意移動抓圖點讀到暫時CTY segment，原有「所有抓圖DS必為15ED」斷言不適用。
這是觀測腳本問題；命名契約仍限創角段，不把暫時DS中的raw欄位當作游標語意。
兩次失敗的log、影像與生成腳本依內容hash保留，沒有挑選亂數重擲。

修正探測從同一冷啟動、同一seed1357一次出發，保留原有19次創角／生日輸入，
再五次Enter及下2／左2／下3／右6，循CTY原始通路接近(9,10)。
共37次正式輸入；入口 `tools/dosgolem_newgame_probe.py` 的 `mother_approach` 情境。
重生命令 `bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --mother-entry-original`。
此模式只重生原版收據，結束成功不代表remake對拍通過。
原版、乾淨dosgolem來源與資產唯讀；只有收據輸出可寫。
結果仍須驗證74次IRQ1、所有產物hash、先前116份PNG／bin不變，以及selector→54→010B的動態鏈。
尚未據此修改production Go、JSON、存檔或完成聲明。

### 原版入口與城鎮帶路的有限閉合

乾淨重跑已成功。原版收據 `work/dosgolem-opening/issue4-mother-approach-receipt.json`，
464,340bytes，SHA-256 `1e5ab45df66191749d0e2ef9f2dfe0261cf5f295c1c959d52fb6031e802c2536`。
dosgolem revision、image、Go版本與原始EXE身份沿用上節既有契約。
37次正常輸入、74次IRQ1、seed1357一次；沒有座標、旗標或對話狀態注入。
210份產物全部核對hash與擁有權；每張PNG另核對本次log的實際生成記錄，拒絕沿用舊圖。
前一生日收據的116份PNG／bin皆由本次重生且相同。
新增210份產物不是210張remake對拍通過；本批沒有正式remake圖像比較。

| 自然狀態／入口 | 原版結果 | 等級 |
|---|---|---|
| 第37次正式方向鍵，主角由(8,10)右移至(9,10) | 19530→196D2→1970B，selector1、SI3C20、AX010B→原始1010B | confirmed；只限這條正常接近路徑 |
| 10121，家中母親下移一步後 | 主角仍(9,10)，NPC0從(10,10)到(10,11) | confirmed；沒有主角家中自動跟隨16步 |
| 10130，原始轉場後 | CTY00 sec0，主角(8,38)，NPC0(8,37) | confirmed；原始section base從1383變000C |
| 37次194C3及其返回點 | 主角北2／西2／北7／東16／北10，逐格抵達(22,19)；NPC0在(22,18)轉向下 | confirmed；全部37個主角位置核對，不靠截圖推算 |
| 101C5及末尾抓圖 | 城門提示等待，主角(22,19)，母親(22,18) | confirmed；尚未觀測101D3／1020A，不宣稱對話後移動或旗標完成 |

因此本節上方入口表的raw54資料xref與selector分派，升為上述單一路徑的confirmed。
移動writer的其他阻擋／attr gate分支仍strong；四個事件格的分布是原始CTY事實，
沒有把另外三條接近方向也算作動態通過。
早期 `rec81後立即自動護送`、人物(8,3)及城門對話(21,19)的解釋已被本次原版反證推翻。
下方歷史表保留形成史；相關勘誤回鏈見[docs/66](66-original-flow-oracle.md)與[docs/192](192-opening-dialogue-v3-closure.md)。
母親順序、接近觸發與共同帶路必須一併重寫有限規格，不能把新版原型直接放進正式路徑。
存檔恢復、全域人物相位、正常接近的失敗gate及城門對話後仍是DRAFT閘門。
另已推翻docs/192把101BF的NPC動作誤作第一段文字consumer的解釋。
原始caller在101CE只有一個明確 `sub_21414` 呼叫；詞流／record連接須由下一次自然確認閉合。
不能依舊表把remake目前的兩次dialogue.Open直接升級為原版對話流程。

觀測腳本另外修正了一次分類錯誤：沒有用初始主角座標猜生日／其他場景；
改為只在自然抵達1010B後標記後續相位。舊產物先以內容hash完整歸檔，再移除本情境檔案後重生。
較早有分類錯誤的收據不作新圖像驗收。480秒逾時屬執行預算問題，改660秒後以同一工具鏈重跑。
本節動態狀態鏈與全數圖像重生已通過，不升級為remake CONFORMED。

非破壞附加台帳由 `tools/ida_dump_opening_handler54.py` 自動合併；
`work/issue4-mother-reviewed-ida.json` 保留原名、位址、兩種bytes、推論等級與本節回鏈。
審查入口 `work/issue4-mother-entry-audit.py`，結果 `work/issue4-mother-entry-evidence-receipt.json`，
核對九份IDA匯出的3,720列原始指令、自然分派、37步、210份產物與116份不變畫面。
審查收據25,856bytes，SHA-256 `946e279ccbdf06090f23e6127c823bcd0bc823e2c4250b40635b1ea0d2a17c1c`。
這些工作產物與原版資產留在本機，不加入Git。

有限可丟棄原型另驗證83與81之間的NPC0移動。固定 `ad00d38`，沿用現有typed frame，
由pack提供初始(5,4)、16個動作與主角不動的資料；末步轉左需明示方向。
正式17／18／19次InputState必須先播放83，再完成上述序列，最後才開81。
原型只核對人物選擇、位置及順序，不把既有動畫位元當作原版全域相位，也不處理後續接近／存檔。
入口 `work/issue4-mother-birth-prototype.py`，正式來源唯讀，原型與測試位於容器/tmp。

同一正常InputState測試在未改動的 `ad00d38` 失敗，NPC仍(5,4)，主角(5,5)；
隔離原型通過，NPC已(10,10)且轉左，另一NPC保持(8,3)，之後才開81。
log為 `work/issue4-mother-birth-baseline.log`／`issue4-mother-birth-prototype.log`。
這只支持人物狀態與對話順序的有限草案，沒有升級人物RGB、全域相位、接近事件或存檔。

## 2026-10-02：PIT 除數勘誤與 NPC0 移動序列（有限時間 CONFORMED，NPC DRAFT）

依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4) 接續 `de91dda`。
下節沿用預設 PIT 除數65536的18.2 Hz換算已被新收據推翻；字模、8／5閾值與完整生日畫布的結論保留。
原版唯讀觀測 `Ticks` 與 `PITDivisor()`，生日與房間全部301筆箭頭相位均為12428。
生日第一次顯示ticks5907、清除5915、再次顯示5920；8／5差值與來源閾值一致。
生日與房間各四次捲動提交的ticks分別6896..6899、7896..7899，每步差1。
原始逐字consumer仍為DGROUP0005至少1，定位linear0x214B9..0x214F8；不研究ISR逐週期。

新原版收據 `work/dosgolem-opening/issue4-birthday-pages-receipt.json`，111,787bytes，SHA-256
`1c41e8bc2566ad659cfea40ab52fb984f4d0e128bfca34703707b55819801644`。
來源EXE、位址基準、IDA9.4及字模身份沿用下節；dosgolem固定revision
`2f44a68ebfc54b28fb15dd4a34510b0b04a5415d`，Go1.24.13，既有image。
上游 `cmd/probe/main.go` 有未提交修改，保持原狀；改用只讀的乾淨隔離副本
`/tmp/dq3-dosgolem-2f44a68`。入口為
`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --birthday-pages`。
觀測僅讀取原版欄位，19次正式輸入、38次IRQ1及一次固定seed0x1357維持。

有限草案只修正生日／房間共用文字時間參數。依
[DOSBox Staging timer契約](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/hardware/timer.cpp)，
頻率是PIT輸入除以已觀測除數。沿用PIT輸入315000000／264 Hz，頻率改為315000000／3280992 Hz。
固定60TPS向上取整：8tick為5更新、5tick為4更新、1tick為1更新。
`wait_indicator`頻率、`window.glyph_hold_frames`與`text_flow.scroll_hold_frames`均由JSON保存。
時間維持hardware-spec approximation；不宣稱同硬體wall-clock、NPC動畫或整段campaign parity。
驗收需正常17／18次輸入、生日兩相與六張續頁完整RGB、原始bytes、pack嚴格驗證、後續正常玩家流程及存讀檔回歸。
可丟棄原型固定 `de91dda`，入口 `work/issue4-clock-prototype.py`／`issue4-clock-prototype-run.sh`。

另外，原版第19次Enter解除record83後，先等待7tick，再以DX0消費DGROUP3CBC三位元組序列，之後才播放record81。
序列raw `02 01 00 06 00 00 07 03 00 01 01 02 FF`，EXE file0x19DFC、IDA linear0x28A8C。
16個動作依序左2、下6、右7、轉左；NPC0由(5,4)到(10,10)，主角全程(5,5)。
每動作等待8個已觀測tick，除數亦12428。record81 EOF及開場caller返回時主角仍(5,5)。
目前remake直接接record81，再把(8,3)的人物與主角一起移動；此順序與人物選擇有原版反證。
後續linear0x1010B帶路入口的正式觸發仍unknown；NPC全域相位的初始化亦尚未閉合。
這部分保持DRAFT，不以固定人物影格或猜測觸發補入正式路徑。

### 有限時間契約審查

可丟棄JSON原型的生日顯示／清除／再次顯示及原有六張續頁完整RGB均零差異，創角嚴格PASS。
房間仍261個人物差異，第19次不同狀態仍196,698，沒有掩蓋或固定影格。
原版118份產物全部核對，與前一份7ca7da23收據的116份PNG／bin完全相同，19次輸入及38次IRQ1亦相同。
301筆箭頭觀測中，定時兩相均相差8／5tick；兩個正常確認清除事件可提前終止相位，不套用定時閾值。
原始consumer的逐字 `3D 01 00` 在file0x12863，捲動 `3D 00 00`／`7E F6` 在file0x12E01／0x12E04，皆等1tick。
四份IDA原始bytes已逐列核對，時間與NPC來源審查收據
`work/issue4-clock-source-evidence-receipt.json`，30,744bytes，SHA-256
`f5087b5a90bd88974440e492de3da9ab11b304fbb477583dd4f484772c33a1a5`。
只將生日／房間共享文字的時間契約升為READY；schema欄位不變，content改0.1.69。
本切片不修改存檔格式或NPC流程；回歸必須驗證正常下一節點及既有存讀檔。

### 正式時間驗收

已將有限READY接入原有JSON欄位，schema0.1.62／content0.1.69。
正式正常17／18次輸入的生日兩相及六張續頁完整RGB零差異，創角嚴格PASS；限定28張保持。
原始EXE glyph／scroll閾值bytes及FON parity、必填契約拒絕測試通過。
`game/opening_retained_test.go`另以自然收據核對生日／房間實際除數、8／5tick及四次捲動每步1tick，
避免再次以未觀測的BIOS預設頻率代替遊戲參數。沒有新增production Go版本常數或存檔格式。
完整房間兩相及穩定畫面仍261個人物像素差異；第19次不同狀態196,698仍僅診斷。
完整game365項頂層／34項子測試、11個internal與desktop main.go通過。
正常新遊戲至THE END88.40秒、主角／酒館及各段存讀檔通過，只屬remake回歸。
標準38項選用SKIP，創角／生日嚴格PASS，房間RED，其他35項未執行；沒有素材缺失。

驗證log `work/issue4-clock-production.log`／`issue4-clock-full.log`，私有稽核入口
`work/issue4-clock-final-audit.py`；正式收據 `work/issue4-clock-final-evidence-receipt.json`，
9,414bytes，SHA-256 `93d75e9a3c6f03cb63d03eab9b42c05f161f4605e7818e5ce6c9920552307b33`。
本節時間換算有限CONFORMED，仍標hardware-spec approximation；不升級整段wall-clock或NPC流程。
所有新增輸出UID／GID1000，歷史root／Markdown目錄候選3213保留；本輪一次性Docker容器已清除。
沒有新image、發行包或原版素材入Git；上游未提交修改與使用者scratch／Android libs保持。

### NPC序列證據與下個閘門

追加勘誤：下表較早寫「1010B正式觸發未知」已由本文件最新母親入口節解出。
原版是正式走近觸發；城鎮37步抵達(22,19)，不沿用舊影片近似路線。
其餘未驗的全域相位、對話後與存檔仍未知，沒有因入口已解出而升級。

| 原始定位與 consumer | 有限語意 | 等級與限制 |
|---|---|---|
| IDA linear0x100D5..0x100F2／file0x1445..0x1462 | record83返回→7tick停頓→DX0／SI3CBC→sub_167C5→視窗→record81 | confirmed；正常第19次Enter的動態順序與原始caller閉合 |
| DGROUP3CBC／linear0x28A8C／file0x19DFC | 三位元組repeat／direction／mode序列，FF結束；左2、下6、右7、轉左 | confirmed；13個raw bytes、parser與16個自然結果一致 |
| linear0x167C5..0x16822／file0x7B35..0x7B92 | sub_167C5解讀序列，sub_167F6依序NPC writer、場景renderer、主角consumer、page flip，再等待counter大於7 | confirmed；每步原始consumer與Ticks差8，不宣稱同硬體wall-clock |
| linear0x121EF..0x122CD／file0x355F..0x363D | DX選DGROUP0B66+8×DX；低2位方向先寫，mode2只轉向，其餘按方向更新X／Y及佔位 | strong；原始writer及16個自然結果閉合，mode1尚未玩家路徑驗收 |
| CTY00 section4，file0x139C起的day表 | 第一原始列(5,4)，flag50h，對應可見runtime NPC0；(8,3)是另一可見人物 | confirmed；raw CTY與同次runtime slot；不合併人物 |
| linear0x100FA..0x1010A／file0x146A..0x147A | record81 EOF、7tick、恢復視窗、caller返回；主角仍(5,5)，NPC0(10,10) | confirmed；唯讀原版同次觀測，尚未宣稱自由移動入口 |
| linear0x11ED0..0x11F4E／file0x3240..0x32BE，DGROUP0004 | NPC取圖讀全域相位，ctrl80h凍結分支；當前remake每NPC計数不同 | strong；原始reader與直接xref，初始相位與更新契約尚未READY |
| linear0x1010B起／file0x147B起 | 後續帶路函式存在，先檢查flag50h；沒有直接caller xref | unknown；正式玩家觸發尚未閉合，不以零xref宣稱未使用 |

輸入 `assets_raw/CTY00.DAT`，7,546bytes，SHA-256
`ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836`。
其餘沿用本節EXE身份與IDA9.4位址基準。sidecar保留原名、原始位址、loaded及file bytes、xref type與unknown原始分級；
本表附加審查語意，不把自動函式名當證據，不修改資料庫名稱。

| 私有sidecar | 大小 | SHA-256 | 已核對raw rows |
|---|---|---|---|
| `work/issue4-npc-consumer-ida.json` | 89,448 | `c83c03ce5243fbda2485927ba2b28a6541fe165e09395e3ccc8b64eb15cd871c` | 58 |
| `work/issue4-npc-sequence-ida.json` | 659,974 | `a576cd2c866f02ed2dbf48ce62de44001f1fb120c888d76c83e4d6b6bb873dc2` | 368 |
| `work/issue4-npc-follow-ida.json` | 1,506,253 | `3b5738dd5fbca63a3e02253d69087ba9680944f20066f19b2e3a82b89451a4b1` | 539 |

匯出入口 `tools/ida_dump_opening_handler54.py`，target file依序0x3240、0x7B35、0x147B，仍在一次性IDA Docker執行。
下一切片先找0x1010B的entry selector及正常觸發，補齊83返回後的NPC狀態鏈與有限pack規格，
再修正正式人物選擇與文字順序。全域相位不以固定walk值近似；硬體ISR／PIT逐週期細節依停止線不追。

## 2026-10-02：等待箭頭相位（生日限定CONFORMED，人物仍RED）

依[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)接續5029ccf。
本輪起點的完整房間302個RGB差異中，箭頭佔41，人物佔261。本節只閉合共用文字等待指示。

| 原始定位 | 附加語意 | 等級與限制 |
|---|---|---|
| IDA linear0x216C5／file0x12A35，`BD 2A 00` | 字模橫座標為42個VGA bytes，即336px | confirmed；來源bytes與完整生日兩相閉合 |
| linear0x216C8..0x216CB／file0x12A38..0x12A3B，`BX=13`→sub_211B6 | 16×16不透明寫入箭頭；DX沿用文字consumer的當前行 | confirmed；不將首次換頁暫態當完整首頁 |
| linear0x216E6／file0x12A56，`3D 08 00` | 顯示後等待DGROUP0005至少8，再寫入字模12 | confirmed；只證實來源閾值與可見清除，實機wall-clock仍為近似 |
| linear0x21709／file0x12A79，`3D 05 00` | 空白等待至少5，再回到箭頭繪製 | confirmed；來源bytes與自然相位，不研究ISR逐週期 |
| linear0x21710..0x21718／file0x12A80..0x12A88 | 確認輸入後亦先寫字模12，再返回文字流程 | confirmed；正常輸入後續頁、六張畫面閉合 |

IDA Pro9.4匯出`work/issue4-wait-arrow-ida.json`，282,146bytes，SHA-256
`20ebafb740eca58d0c030652ea1dd3e591efb2ddb0a79615ae993bf92634be47`。
重生入口`tools/ida_dump_opening_handler54.py`，target file0x12A33、末端0x12A97。
輸入`assets_raw/DQ3.EXE`，115,282bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；
位址空間為IDA linear，file=linear−0xEC90，runtime physical=linear−0xEF00。
147筆原始file bytes已核對，保留原名、位址、xref與分級，沒有rename原檔。
字模來源`assets_raw/D3TXT00.FON`，47,232bytes，SHA-256
`c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`。

正式`opening_prelude.wait_indicator`必填有限契約：mode、x、visible_glyph、hidden_glyph、
visible_ticks、hidden_ticks、rate_numerator、rate_denominator、evidence與timing_evidence。
引擎只知道當前行的兩相字模閃爍，60TPS更新幀數由資料包頻率向上取整。
沿用既有315000000／17301504 Hz時基，8／5 ticks分別約27／17個更新。
時間只標hardware-spec approximation，來源為
[DOSBox Staging timer實作](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/hardware/timer.cpp)
的65536除數與timer delay契約；不宣稱原版逐週期時序一致。

原版首次生日箭頭擷取仍在可見頁換入前，只有箭頭，必須保留為暫態診斷。
正式完整首頁採下一次顯示與首次清除兩個已識別相位；不挑選相似圖片或更換人物影格。
前一批捕捉超過240秒，已清除容器；本批沿用image與資源限制，內外逾時改為360／420秒。
原版原始資料與dosgolem來源仍唯讀，19次輸入及種子設定不改。
原型來源固定5029ccf，入口`work/issue4-wait-indicator-prototype.py`，結果保存於
`work/issue4-wait-indicator-prototype.log`。生日顯示、清除、再次顯示及原有六張續頁全RGB零差異。
房間兩相均261差異，範圍只在人物，沒有遮罩或指定影格；完整房間仍RED。
自然相位確認字模、當前行與清除控制流，可見部分升為confirmed，8／5閾值來自原始指令。
上述有限契約審為READY；定時精度維持hardware-spec approximation，不外推人物或後續事件。
完整原版收據`work/dosgolem-opening/issue4-birthday-pages-receipt.json`，83,763bytes，SHA-256
`54a0c21f80559c355b0e0dbaf9f726af59739021c7cfcfaf33a833468fcfbe51`，80份產物含39張PNG／bin、log與生成腳本。
首次生日顯示步數1102060165保留為transient；完整顯示1103628229、清除1103025099。
301筆相位觀測只讀取DGROUP欄位，種子仍0x356D，沒有新增遊戲狀態寫入。

### 正式驗收

正式來源`game/dialogue_retained.go`與`internal/gamepack/gamepack.go`／`opening_scene.go`。
嚴格解碼與原始EXE／FON parity在`opening_scene_test.go`；生日兩相、再次顯示、續頁清除與六張既有捲動由
`game/opening_retained_test.go`正常InputState核對，沒有文字狀態注入。
房間完整RGB兩相及穩定畫面均261差異，範圍(289,131)..(412,168)只在人物；箭頭41差異消除。
第19次不同狀態的196,698仍只作診斷。生日新增一張獨立顯示畫面，累計限定28張；房間未加入通過數。
schema/content為0.1.62／0.1.68，位置、字模、閾值、頻率與證據全部放在JSON，沒有Go版本 fallback。

完整game365項頂層／34項子測試、11個internal、desktop main.go通過；
正常新遊戲至THE END63.35秒，主角／酒館與各段存讀檔通過，只屬重製回歸。
標準38項選用SKIP，創角與生日嚴格PASS、房間RED，餘35項未執行，沒有素材缺失。
測試記錄`work/issue4-wait-indicator-production.log`與`work/issue4-wait-indicator-full.log`。
附分級IDA sidecar `work/issue4-wait-arrow-reviewed-ida.json`，289,977bytes，SHA-256
`61e18140a63a00217a468af5c2eb646fa4dbf23a190576bd6abdbe5f0b6bc8a3`。
147筆原始file bytes再核對；原版名字／位址不改，先前未知收據保留，新增語意由REVIEW_LEDGER自動附註。

私有稽核入口`work/issue4-wait-indicator-audit.py`，收據`work/issue4-wait-indicator-evidence-receipt.json`，
28,655bytes，SHA-256 `f802ab3a1024cbb25ebcb758be3dcb128150776fc0e5a05b14f0afd5f1e09ce3`。
核對80份原版產物、兩份IDA共294筆raw rows、正式來源hash、正常19次輸入與固定種子、完整RGB和回歸。
原版與dosgolem來源唯讀，本輪UID/GID1000；歷史root候選3213與Markdown目錄候選4保留。
本輪一次性容器已清除，未建立新image或發行包，原版圖片／資料／database不入Git。
下一切片為NPC繪製時的相位及record83返回後更新，不以固定人物影格近似。

## 2026-10-02 上一輪結果：有限房間呈現已正式接入，完整RGB仍RED

本輪工作與有限READY審查依[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)。
固定現行5a2101d的隔離原型302差異，正式版本亦重現302：場景人物261、文字框箭頭41。
完整640×350 RGB全部計入，未裁切／遮罩或指定人物影格；原有172,261差異已大幅縮小。
陰影框底差異消除，背景、字模框與文字亦已目視核對；完整房間仍DRAFT／RED，沒有增加已通過對拍張數。
record83保留前文及EOF捲動、81自動返回後正常接回帶路；NPC更新與母親狀態仍待原版閉合。
第19次兩側不同狀態的196,698差異只供診斷，不當same-state驗收。

正式來源`game/opening_scene.go`、`game/dialogue.go`與`game/game.go`；
契約與原始資料parity在`internal/gamepack/opening_scene.go`及`opening_scene_test.go`。
`game/opening_scene_test.go`使用17次創角、第18／19次確認與正式空白InputState，
不注入場景／人物；驗證呈現不洩漏到帶路。所有版本值與83／81文字引用移入JSON，
移除Go中的開場record陣列，schema/content為0.1.61／0.1.67，欄位入口見[docs/84](84-game-pack-json-contract.md)。

生日六張全RGB仍零差異，創角嚴格通過，累計限定27張保持。原版新收據37,150bytes，
SHA-256 `52ad27c06e8821d3f115f42dbbdf4794d7782b16256e5901623b4dfec9a6eaa6`。
原始66份PNG／bin與13c5e263收據相同，另加兩份黑頁觀測；19次輸入與38次IRQ1相同。
生成腳本SHA-256 `de8b9e3b27d9d977e554d38f2667ff0714c54fbba757f273beaec9557ec36c88`。
黑頁觀測推翻「人物前已完成完整背景」假設，原始收據與下方勘誤保留。

完整game364項頂層／34項子測試、11個internal及desktop main.go通過；
正常新遊戲至THE END106.51秒，主角／酒館及各段存讀檔通過，只屬重製回歸。
標準38項選用SKIP，原版三項另行兩PASS、一房間RED，剩35項未執行，沒有素材缺失。
初次回歸的fixture仍寫舊schema；trace仍連按確認，且舊helper等待整串關閉而越過同更新的下一段。
已依正式狀態機改成等待當段openingIdx轉移，只在文字停點確認，再以同一容器乾淨重跑。

私有稽核入口`work/issue4-room-production-audit.py`，收據`work/issue4-room-production-evidence-receipt.json`：
26,625bytes，SHA-256 `63a3f70a88a450ac2822f172556ff3531f0fb3e8554e271f6fb31045aa9e78ef`。
保存原始source／sidecar、正式PNG與完整回歸、冷啟動及固定seed條件。
UID/GID1000；歷史root候選3213及Markdown目錄4保留，新增輸出擁有權通過。
一次性Docker均隨工作清除，沒有新image或發行包。
下一步為兩側NPC繪圖相位、箭頭相位及83返回後的原版狀態，不寫死這次人物影格、不深挖PIT／ISR。


## 房間正式呈現規格草案，接續5a2101d

狀態 READY，限下列資料與呈現原語；完整房間仍 DRAFT。沿用下方原始EXE／CTY／FON／PAL身份、IDA9.4 sidecar及自然19次IRQ收據。
先閉合正常出生場景的背景producer、共用字模框與文字consumer，再處理人物動畫及箭頭。
正式原版第一個房間等待仍以完整640×350 RGB守門，不裁切或遮罩；302差異原型不算通過。

預定資料契約 `interface.opening_scene_presentation` 指定CTY／section、player_anchor視野、
界外圖塊、共用presentation引用、文字ID序列與具名VGA字組陰影。全部版本數值在JSON，
引擎只處理有限原語。攝影機只在明確指定的opening場景及文字序列期間使用此契約，
未審查的其他場景仍列原版oracle待辦，不把家中證據外推。資料缺失或引用錯誤拒絕載入。

勘誤：步數1222142894、IDA linear0x11EE8的新增擷取實際為全黑顯示頁，
推翻本草案先前「已完成完整背景」的假設。這份PNG保留為觀測限制，不能用作背景驗收。
原始圖塊consumer亦有穿插NPC繪圖的分支，不能假設兩側renderer有相同階段。
本輪以現行來源隔離重播最終record83等待，完整畫布比較視野player−9/7、界外圖塊71、
共用文字框與陰影。人物、箭頭及其他差異全部計入，不以中途背景或遮罩代替最終畫面。

共用原始視窗由record404產生，原點152/238、352×96、字距24、文字前景243；
陰影由原始byte x+1與y+8，按固定dosgolem revision的16位元AND／鎖存契約呈現。
record83及81的原始詞流引用pack文字，EOF與FFFC沿用同一已審查consumer；
原版83返回後的NPC更新與母親移動仍須獨立閉合，不能只因文字返回就稱整段開場完成。
驗收包括schema／原始資料parity、正式InputState、完整RGB、後續節點及存讀檔回歸。

審查結果：固定現行來源5a2101d的隔離副本，正式17次創角與第18次確認後等待record83，
完整640×350 RGB差異302，與舊副本一致。沒有設定人物姿態或清除差異。
有限契約採`presentation_id`引用已審查生日樣式，`text_ids`依原始caller提供83、81；
`camera`採player_anchor與9/7、layer0界外71；`shadow`採vga_word_latch_and與8/8。
CTY／section必須與既有出生入口相同，文字引用及控制碼必須驗證，所有欄位必要且拒絕null／未知。
Go不保存這些版本數值。缺資料拒絕載入，不把NPC更新、步行影格或箭頭相位一同升級。
record81與83共用consumer只授權保留／捲動與返回原語；母親演出狀態仍未對拍閉合。

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
## 2026-10-02 母親城鎮帶路續跑與限定實作規格

本節追加訂正，不取代前輪收據。Issue #4 是遠端工作入口。本輪先將城鎮抵達後的
位置、NPC、文字與旗標交易列為 DRAFT，再以原版冷啟動續跑審查。家中自動接管、
圖像選擇選單、全域動畫相位與完整畫面對拍仍待修正，不把本節升格為整個開場完成。

## 原版續跑與證據審查

- 重生工具：`tools/dosgolem_newgame_probe.py`，環境
  `DQ3_NEWGAME_PROBE_SCENARIO=mother_finish`。沿用 dosgolem `2f44a68` 與既有隔離工具鏈，
  完整冷啟動 38 次正常輸入、76 次 IRQ1，固定測試種子一次，沒有其他遊戲狀態注入。
- 原始 EXE：`assets_raw/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 原始文字：`assets_raw/D3TXT01.TXT`，6420 bytes，SHA-256
  `4d0f78b20f123a986feb9af62845183213adee6881e953825eb74678abce8271`。
- 本機收據：`work/dosgolem-opening/issue4-mother-finish-receipt.json`，SHA-256
  `9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f`；
  `work/issue4-mother-finish-evidence-receipt.json` 保存核對結果。176 個唯一產物的
  大小與雜湊全部吻合，先前 168 個 PNG／色號檔維持相同內容。
- 已證實：IDA linear `10130` 的主角 `(8,38)`、原始 NPC record0 `(8,37)`，
  接著 37 次 caller 返回的位置形成北2、西2、北7、東16、北10。
  `101C5` 主角 `(22,19)`、母親 `(22,18)` 朝下。
- 已證實：IDA linear `101CE` 只有一次 `DI=0C08 → sub_21414`。
  `21441` 的 36 個 raw words 與原始 record80 完整相同，包含一個 `FFFC` 及最後 `FFFF`。
  record79 不屬於這次帶路對話。
- 已證實：IDA linear `101D3 → 101E6 → 101EF → 101F8 → 1020A`，
  確認後主角依序 `(21,19)`、`(21,18)`、`(21,17)`；母親仍在 `(22,18)` 朝下。
  到 `101F8` 旗標尚未交易；`1020A` 的 DGROUP `4F72` 從 `00` 變成 `01`，
  `4F7A` 從 `FF` 變成 `7F`，`0B34` 歸零。
  此處 flag17h 是高位優先 bit map 的 bit0，flag50h 是 bit7。

位址使用 IDA Pro 9.4 loaded linear，file offset = linear − EC90；上列動態事件
由 dosgolem 只讀觀察自然入口。原始 bytes 與非破壞性 IDA 索引沿用本文件既有證據鏈。

## READY 範圍

上述城鎮抵達後的有限狀態鏈已達 D3，限定實作可進 production：

1. pack 宣告原始 NPC record 身分、每步兩者位置、母親方向、單一文字 ID。
2. 原始 record number 留作資料 oracle；共用引擎只消費穩定文字 ID。
3. 對話關閉後仍保留母親旗標，完成最後三步才交易旗標。
4. 正式新遊戲 InputState trace 必須走到下一個可操作節點，驗證存讀檔及後續謁見。

本輪不重新證實既有 hold_frames、家中序列或角色動畫。既有每步等待只保留為未完成的
呈現近似，不能以本節位置證據宣稱時序或 RGB 已達 V3。runtime 圖片必須如實保留差異。

前輪 `mother_approach` manifest 有 210 列，實際是 170 個唯一檔案，重複 40 列。
原始收據維持不變；本輪生成器改為拒絕重複檔名，並確認 PNG 由當次執行寫出。

### 限定實作與驗收

現行schema0.2.0／content0.1.70。原始NPC記錄索引在可見性過濾後仍保留，目的場景
以pack引用選母親；42步逐項宣告兩者位置及母親方向。強制帶路可穿過原版普通碰撞格，
但轉場提交前仍驗界內與NPC身分。未知actor／文字引用回報錯誤，不跳過交易或猜補。
文字由穩定ID與已審查共享呈現開啟，單一record80的FFFC等待後自動EOF；最後三步完成才交易旗標。
既有家中演出、每步hold_frames與動畫仍為未驗證近似，這些範圍未升級CONFORMED。

| 驗證 | 結果與界線 |
|---|---|
| 修改前ed27d71同條正常輸入 | 第一個城鎮步進FAIL：主角7,38／母親8,37；原版應8,37／8,36 |
| EXE／CTY／TXT parity | 42個兩者狀態、原始NPC record0、單一文字及旗標引用PASS；必填及null／未知欄位拒絕PASS |
| 正式玩家入口到城鎮checkpoint | seed1357於首個InputState前固定一次；42個原版觀測值逐項PASS，城門確認一次，未注入位置／旗標 |
| 同版本存讀檔 | 主角21,17、場景與旗標PASS；原版存檔格式與NPC動畫不由此宣稱parity |
| 完整remake回歸 | 367項game頂層／34子測試、11個internal與desktop main.go PASS；正式新遊戲至THE END65.74秒 |
| 嚴格原版畫面 | 創角與生日捲動PASS；接受角色的完整房間三圖各261、家中續行198049像素，RED |
| 城鎮返回完整640×350 RGB | 83387像素差異；可見水平視野差32px與人物動畫差異，RED；不裁切或遮罩 |

原版38次輸入含五次圖像選擇確認及13次家中接近；正式remake仍沿用錯誤家中自動接管。
本輪state比較限定共同城鎮checkpoint，不宣稱兩側38次等價輸入、整個開場E3／V3或音訊一致。
來源重生入口：`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --mother-finish-original`。
這個入口只重生原版，成功不代表remake通過。remake正式比較入口為
`TestDosgolemMotherArrivalStateComparison`，在既有Docker／Xvfb驗證環境設定
`DQ3_MOTHER_FINISH_ORIGINAL=/work/dosgolem-opening/issue4-mother-finish-receipt.json`，素材仍唯讀。

本機收據：`work/dosgolem-opening/issue4-remake-mother-arrival-receipt.json`、
`work/issue4-arrival-baseline.log`、`work/issue4-arrival-full.log`、
`work/issue4-arrival-retained.log`、`work/issue4-arrival-visual-receipt.json`。
`work/issue4-arrival-final-audit.py`核對來源、收據、文件回鏈及輸出擁有權，
輸出`work/issue4-arrival-final-evidence-receipt.json`；私有圖像與原版檔不加入Git。

### 結論回填索引

原始鍵為DOS／上述SHA-256的DQ3.EXE／IDA9.4 linear，file=linear−EC90。
MOTHER-FINISH-38-INPUTS是舊規格勘誤標記；本機稽核逐項檢查標記及docs/188回鏈存在。

| 原始鍵 | 新證據語意與等級 | 回填文件 |
|---|---|---|
| 101BF／101CE／21441 | NPC轉身後單一record80，全部36words含EOF；confirmed | docs/192、docs/189、docs/74；舊雙record已推翻 |
| 101D3／101E6／101EF／101F8／1020A | 三步後主角21,17，最後set17h／clear50h；confirmed | docs/192、docs/189、docs/66、docs/74；舊對話後未知與提前旗標已勘誤 |

原始位址與歷史表保留；不用較早V3聲明提升本輪畫面。家中入口、動畫與原版存檔仍分開追蹤。

提交／推送與遠端Issue收尾收據：`work/issue4-arrival-post-push-receipt.json`；
容器及擁有權稽核：`work/issue4-arrival-hygiene-receipt.json`。未新增公開原版附件。
