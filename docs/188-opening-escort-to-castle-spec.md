# 188 — 開場連續演出勘誤與修正規格：家中 → 王城入口 → 國王

目前家中正式流程以「2026-10-03 正式家中驗收」為準，城鎮及城堡攝影機見文末對應驗收節；下列較早DRAFT保留研究歷史。

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
# 2026-10-02 家中完整流程續作規格

## 2026-10-03 READY 審查

`work/issue4-home-production-prototype.py`以現行HEAD只讀快照建立隔離副本。
`TestOpeningHomeProductionPrototype`兩條正常輸入路徑通過：17次創角輸入、生日及房間確認、
17個原版人物狀態、母親對話自動返回、預覽及三輪選圖、交還控制、13次手動移動、
城鎮對話與最後旗標交易。Escape路徑依本輪動態收據等同選定當前圖像。
能力種子1357、獨立圖像種子151B均於首個輸入前固定一次。
家中及城鎮分別經正式Save／Load通過，沒有用讀檔替代初次正常玩家入口。

原版新收據SHA-256：contract為
`fbbb273d850f0db6deebd26a1301bbfac79e4764c5d11677f3571fbf97622074`，navigation為
`4777dacdcfc4803d0362e248e1cb34d293b586614cd910601528ec6852cfe6dd`。
兩者生成器封存SHA-256均為
`d354dd5eabe704265bb3583899a84c823bea3b13b39652c9adbe1e37642d67a5`。
來源工具、唯讀觀測、原始hash及輸入列全部留在本機manifest。

追加訂正：先前BLS的七個像素差異位於相鄰圖像首列，來源是34px游標框跨過32px圖像格。
依原始1F908／204B1在圖片之後施加XOR8即吻合，不需修改圖片或增加例外。
透明區域依1EA8C以原始背景tile0及sprite平面合成。數字欄先畫前導空白glyph12。
預覽、三輪選圖及結果的336×192視窗RGB均0差異。完整640×350各124差異，
視窗外人物動畫仍為RED。驗收輸出為 `work/issue4-home-production-visual-audit.json`，
只將已驗證的有限流程及視窗primitive審為READY，不宣稱全畫面、音訊或硬體wall-clock完成。

正式接入範圍：共用有限狀態機、嚴格JSON、完整96張BLS形狀與reference validation、
母親原始record0、正常接近入口及等待接近存檔。schema升為0.3.0，content為0.1.71。
新增契約與重生入口維持本文件及docs/84索引；後續須跑受影響測試、完整game／internal、
desktop build及正常新遊戲至THE END，再更新唯一目前狀態表。

本節沿用 Issue #4、dosgolem 2f44a68 與同一 DQ3.EXE。入口為正常創角17次輸入，
生日及房間各確認一次。原始 NPC record0 執行左2、下6、右7、原地向左，主角固定5,5。
先完成移動再播放record81，EOF後開啟圖像選擇。此鏈不得以直接事件或座標注入驗收。

圖像初始化的只讀收據為 `work/dosgolem-opening/issue4-home-contract-receipt.json`。
IDA linear16F4B讀BIOS046C，正常冷啟動讀值151Bh；16F56保存獨立種子，
1701D以16位加9014h再左旋3位。問題取低7位並拒絕大於99，
原始100組三個目標在DGROUP08C0、file16A00。16FDE混合目標及隨機三連圖像，
重複組合不排除。正常觀測問題13，選項為3,4,5,57,58,59／33,34,35,48,49,50／6,7,8,24,25,26。
此處151Bh是本次自然時鐘值，重製測試於首個輸入前固定獨立選圖種子一次，正式執行由時鐘取值。

21EC1先以1E829／1E82D顯示七個診斷數字，第一確認撤回左上80×112區域，
才開始三輪選圖。初始游標1；左／上遞減，右／下遞增，1至6環繞。
Escape的初步靜態解讀為回第一輪，已被本輪正常導航收據推翻。
1460001231的21FD2雖返回AX=FFFF、DGROUP0726低位=1，caller仍保存當前選項1，
1460004089的21F79已進第二輪。故此EXE的Escape等同確認當前選項。
三次選擇後顯示結果，再確認才還玩家控制。
原版21F0C／21F0D為NOP，第三輪選1即使不等於正解4仍返回成功，不猜補答錯處罰。

畫面資料取原始視窗DGROUP4348與consumer1F590：152,46、336×192、
record476頂列、record466重複10列、record467底列。矩形204B1的高度為BP+1、寬度DX。
圖片原點224,94，間距32；已選圖片272,182；游標223,93，34×26，XOR8。
圖像來源DQ3LIN.BLS實際96張，逐張480bytes；引擎先驗證header、完整長度與引用。
選單色盤bank0及index8=243,243,243。七個素材像素差異已由上述跨格游標閉合，完整人物動畫仍列為V3限制。

手動路徑下2、左2、下3、右6抵達9,10，才進原始handler54／1010B。
母親10,10先向下至10,11，主角保持9,10；再接既有42個城鎮狀態。
正式資料宣告原始NPC身分、接近條件與有限步驟；共用Go不寫DQ3座標、文字或raw handler fallback。
存讀檔需保留選圖結束後等待接近的狀態及母親位置，並可由正常輸入繼續。

來源位址為IDA9.4 linear，file=linear−EC90；原始bytes與hash見既有匯出。
新增只讀匯出為 `work/issue4-home-preview-ida.json`、`work/issue4-home-preview-consumer-ida.json`。
新來源模式為 `mother_home_contract`、`mother_home_navigation`，均由
`tools/dosgolem_newgame_probe.py`自行重生，沒有寫入問題或選項。
READY前需隔離原型通過正常輸入、Escape／環繞、接近入口及同版本存讀檔；完整RGB保留實測差異。

## 2026-10-03 正式家中驗收（有限 CONFORMED）

接續`a2e477d`，依上節READY接入schema0.3.0／content0.1.71。
正式人物順序、獨立圖像生成、預覽／三輪／結果、控制交還及13步手動接近通過。
家中不再於對話返回後直接轉場；主角正常抵達已驗事件格才接城鎮42個狀態與最後旗標交易。
選項、幾何、文字、色盤及原始actor引用只由pack提供。

| 驗證 | 實測結果與界線 |
|---|---|
| 原版來源核對 | contract37次輸入／74次IRQ1／170唯一產物；navigation42次／84次／180產物；全部大小、hash及生成記錄吻合 |
| 正式Enter玩家路徑 | 17次創角、生日／房間確認、17個母親狀態、三輪選圖與手動接近PASS；確認鍵分支另由原有來源狀態測試驗證 |
| Escape及四方向 | 保存第一輪當前選項、第二輪游標1→6→5→6→1，最後三輪結果010101 PASS；navigation原版的額外主流程Enter不列為整條等價輸入聲明 |
| 有限畫面 | 五個完整336×192視窗RGB差0；每張640×350全畫布仍124差異，完整V3保持RED |
| 同版本存讀檔 | 家中等待接近狀態由正常標題選單恢復，繼續走到城門；城鎮最後旗標及位置round-trip PASS |
| 失敗不消耗 | 不相容、錯場景、缺gate及越界存檔拒絕且狀態不變；標題讀檔失敗留在原選單；modal存檔拒絕且不覆寫冒險之書 |
| 嚴格資料與素材 | 12項損壞契約、5項archive／palette損壞拒絕；原始300bytes問題表、人物序列、視窗結構與三個原始record字模PASS |
| 完整重製回歸 | game370項頂層／45項子測試、internal138項頂層／193項子測試及全部11個套件、desktop PASS；正式新遊戲至THE END224.65秒，只屬remake可玩性 |

實際鍵盤Enter與確認鍵在`InputState`是不同欄位。正式測試現以Enter重播創角及開場；
性別與有限開場對話補齊Enter。預覽及結果依原始等待鍵consumer接受按鍵邊緣，
三輪選擇仍以確認／Escape保存選項，方向鍵只移動游標。沒有將Enter改成一般場景命令鍵。

### 實作與可重現入口

- 共用資料契約及拒絕測試：`internal/gamepack/opening_home.go`、`opening_home_test.go`；欄位索引在[docs/84](84-game-pack-json-contract.md)。
- 正式有限狀態機及繪圖：`game/opening_home.go`；正常玩家入口及存讀檔：`game/opening_home_test.go`。
- 素材形狀及存檔交易拒絕：`game/opening_home_asset_test.go`；原版PNG身份與全畫布／視窗分別計數：`game/opening_home_window_test.go`。
- 原版重生：在既有Docker工具鏈執行`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --mother-home-contract-original`；導航改用`--mother-home-navigation-original`。兩個模式僅重生並驗證原版，不能單獨表示remake通過。
- 正式比較：Docker內從`dq3_remake_ebitan/`編譯`go test -p 2 -c ./game`，以Xvfb執行`TestOpeningHomeNormalInput`及`TestDosgolemMotherArrivalStateComparison`。設定`DQ3_ASSETS`為唯讀素材，`DQ3_HOME_ORIGINAL_DIR`為原版收據目錄，`DQ3_MOTHER_FINISH_ORIGINAL`為既有完整返回收據；不提供來源時不產生畫面parity聲明。
- 本機驗收：`work/issue4-home-production-final.log`、`work/issue4-home-game.jsonl`、`work/issue4-home-internal.jsonl`、`work/issue4-home-production-visual-audit.json`；完整來源及圖像留在gitignored工作目錄。

完整回歸使用既有`dq3-ebiten-test:20260822-r1`、Go1.24.13及有界一次性Docker／Xvfb。
第一次在4GiB硬上限被OOM中止，`work/issue4-home-game-oom.jsonl`保留當次記錄；
Docker事件核對後設定`GOMEMLIMIT=2GiB`及`GOMAXPROCS=2`，同工具鏈完整乾淨重跑PASS。
game38項及internal4項選用測試未執行，沒有素材缺失跳過，沒有把環境中止當作產品缺陷。
最終私有來源／程式／輸出與擁有權稽核入口為`work/issue4-home-production-final-receipt.json`。
本批384項程式／工作輸出UID／GID1000；歷史root-owned3213項保持，Markdown目錄0。
一次性DQ3測試容器已清除，沒有新增image，使用者scratch及Android libs保持。

有限CONFORMED只涵蓋上述人物狀態、正式輸入、資料、存檔及五個視窗。
房間261及選圖124個人物像素、城鎮水平視野、音訊、其他接近方向與完整原版campaign仍未通過。
來源DS13／15／17的後續用途仍unknown；不依數值猜正式規則，也不重開PIT／ISR逐週期研究。

## 2026-10-03 NPC動畫續作（DRAFT）

接續`b7516c5`及Issue #4。問題限房間261與選圖124個人物像素，不重新實作已閉合的家中流程。
先由原始取圖consumer、DGROUP0004及其writer追查初值與作用範圍，再由dosgolem正常創角重生。
不指定人物影格，不以裁切或遮罩代替全畫布驗證，未達READY不修改正式動畫。
非破壞IDA匯出入口為`tools/ida_dump_npc_animation.py`，輸出
`work/issue4-npc-animation-ida.json`。工具保留原始位址、bytes、xref type及unknown候選；
未識別函式邊界的區間仍按原始位址匯出，不猜修database。
原版重生使用`tools/dosgolem_newgame_probe.py`的`mother_home_animation`情境，
在既有Docker工具鏈將來源掛為`/dosgolem`、repo唯讀、work可寫，設定
`DQ3_NEWGAME_PROBE_SCENARIO=mother_home_animation`後執行該Python入口。
輸出沿用`work/dosgolem-opening/issue4-home-animation-*`，37次原有輸入保持。
可丟棄動畫試作入口為`work/issue4-npc-animation-prototype.py`；固定`b7516c5`唯讀快照，
從原始六tick翻轉及已觀測頻率計算共用phase，再以正常Enter流程檢查影響。
試作值僅留隔離副本，尚未進production；原始phase條件、modal保持與存讀檔契約仍須審查。

### 六tick共用phase的有限來源閉合

原版EXE身份、IDA9.4及位址基準沿用本文件。`tools/ida_dump_npc_animation.py`的database
直接xref只列出DGROUP0004的writer；operand候選另找到缺直接xref的reader，不以空xref推定沒有讀取。

| 原始定位與bytes | 有限結論 | 推論等級 |
|---|---|---|
| linear1FE87..1FEA9，file111F7..11219，`83 3E 02 00 06`／`80 36 04 00 01` | word0002累計到6歸零，再翻轉byte0004低位 | confirmed；自然1160次翻轉全部間隔6tick，僅保留已選參數，不深挖ISR或硬體逐週期 |
| linear11EA7..11ECC，file3217..323C | 原始NPC bank乘8、ctrl低兩位乘2形成base frame | confirmed；原始bytes及取圖前後寄存器閉合，尚未外推全部CTY素材 |
| linear11EDA..11EE8，file324A..3258，`02 1E 04 00` | 非凍結NPC取圖加共用byte0004，再乘2取指標 | confirmed；自然18532個reader pair核對0／1增量，沒有各NPC獨立phase |
| linear11ED5..11EDA，file3245..324A，`F6 C2 80`／`75 04` | ctrl80h跳過共用phase加法 | strong；來源branch清楚，本次沒有凍結人物的動態樣本 |

第一份新增原版收據SHA-256為`d64fa77cf5d041fa0fa8c694aba57a816b6ca56ac0f84558ec78a5f73062a6e9`。
37次正式輸入、74次IRQ1及170個唯一產物核對；168個既有PNG／色號檔與contract收據完全相同。
來源原始圖片沒有換名取代oracle，動畫觀測唯讀。先前把1FEA9每次進入都當翻轉返回的稽核假設已訂正，
只有counter0002=0的返回才與1FEA4配對，沒有修改原版事件或放寬六tick斷言。

隔離共用clock原型的正式家中路徑及五個視窗仍PASS，完整畫面仍各124差異。
正常快速InputState共193個更新到選圖，原型實際phase與latch為1；原版背景取圖在tick8059、phase0。
這是不同時間條件，不能因兩者都是問題13就宣稱同動畫狀態；沒有將phase硬改成0。
後續重生已新增每個正常輸入的虛擬tick及整段翻轉原點，保留固定輸入及seed，不重擲。
來源驗證入口為`tools/verify_dosgolem_npc_animation.py`，同時核對既有家中收據與動畫／輸入clock。
正式動畫及pack仍保持`b7516c5`；目前只升級來源writer／consumer的有限證據，不升級完整畫面或動畫DRAFT。

### 完整啟動計數與原型反證

本節追加訂正前段有限觀測，不改寫第一份收據。新完整啟動收據SHA-256為
`0b57ba81e303f3d18dead8068775093074796f2e12b7bd669da5241a2b9e6023`；
37次正式輸入、74次IRQ1、170個唯一產物全部核對，168個既有PNG／bin仍與contract完全相同。
生成器封存SHA-256為`34bbda6bc1a88645b310c056bbb96cb06a6118efce1e75f4bf85e67eb0a22f9c`。
正常輸入順序及能力seed保持，新增唯讀輸入時鐘與完整啟動計數，沒有重擲或指定frame。

原始linear1FEA9共有12180次觀測，word0002從1逐次到5、0再循環，byte0004只在0翻轉。
2030次翻轉全部間隔六次遊戲計數。dosgolem虛擬Ticks在step481045303至481317467從757到761，
此間遊戲word0002卻只從1到2；因此不能用總Ticks除6推導相位。
啟動觀測的計時除數為65536，玩家NPC取圖及37次輸入時鐘均為12428。
驗證器改核對原始遞增序列、翻轉配對、各輸入時鐘及當次日誌一致性；
沒有增加「容許一次九tick」例外，也不深入ISR或硬體時鐘。所有18532個NPC取圖配對仍通過。
凍結人物動態樣本為0，ctrl80h分支維持strong，不能由此提升為confirmed。

| 同次房間繪圖的自然觀測 | 相位及取圖 | 結論 |
|---|---|---|
| step1222142883／1222142894，linear11ED0→11EE8，virtual tick6902 | base42→frame43，phase1 | NPC reader動態閉合 |
| step1222161224／1222161235，同consumer、tick6902 | base32→frame33，phase1 | 第二NPC同相位 |
| step1222323328／1222323332，linear1E2FF→1E30B，tick6904 | BX0→0，phase0 | 主角稍後取圖；同次畫面不能假定只取一個相位 |

隔離原型的家中兩條正常輸入與五個視窗仍PASS；完整選圖畫面各124，房間三圖各261，維持RED。
生日兩相、六張續頁及四次捲動仍PASS；第19次之後另一狀態的231差異只作診斷，不當房間同狀態驗收。
反證推翻「整張畫面只取一次共用phase」假說，原型不進正式Go、pack或存檔。
正式程式仍為schema0.3.0／content0.1.71，先進城鎮攝影機切片；動畫繪圖時序保持DRAFT。
相位writer與NPC reader已取得最小充分證據，不為這些像素重開PIT／ISR逐週期研究。

### 工具、附加語意與核對入口

- 完整來源重生：`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --mother-home-animation-original`。主機只控制Docker；900秒外層逾時，既有image、UID/GID1000、唯讀來源及素材。此模式只重生原版，不執行remake比較。
- 來源驗證：Docker內執行`python3 /repo/tools/verify_dosgolem_npc_animation.py --receipt /work/dosgolem-opening/issue4-home-animation-receipt.json --output /work/issue4-home-animation-source-audit.json`。核對六次遊戲計數、逐對取圖及37次正常輸入時鐘，不把虛擬Ticks視為遊戲計數。
- 已審查台帳：`tools/ida_npc_animation_ledger.json`。以原始EXE身份、IDA linear／file、bytes為key，逐項保存有限語意、consumer、推論等級及本節證據回鏈。`tools/ida_dump_npc_animation.py`自動合併，先驗證EXE與IDA bytes；unknown／strong均帶警示，原名及原始位址保持。
- 私有IDA重生：`work/issue4-npc-animation-reviewed-ida.json`及同名log，IDA9.4一次性database位於容器tmp，完成即清除。八筆台帳為六confirmed、二strong；新加主角reader仍為unknown候選，不由可讀反組譯自動提升。
- 私有原型與結果：`work/issue4-npc-animation-prototype.py`、`issue4-npc-animation-prototype.log`、`issue4-npc-animation-prototype-room.log`。來源快照固定b7516c5，原型值僅在隔離副本。
- 最小充分稽核：`work/issue4-npc-animation-audit.py`，輸出`work/issue4-npc-animation-final-audit.json`。核對既有三份來源、168個不變圖像、IDA原始rows、八筆台帳、shell／Python語法及十種損壞收據拒絕。損壞案例連同暫存日誌重新計算manifest，確認拒絕來自語意檢查；正式原版收據維持不變。

本批只修正來源驗證器與可重生證據工具，不提升正式動畫、全畫面、原版音訊或campaign完成度。

## 2026-10-03 正常城堡接近與視野（有限 CONFORMED，完整畫面 DRAFT）

Issue #4從已通過的母親返回繼續。修改前基準為`de52686`，schema0.4.0／content0.1.72。
探針保留原先38次冷啟動輸入，再從(21,17)送9次正常上鍵。
此路線只用於尋找下一個玩家阻塞點；不注入座標、場景、旗標、影格或文字進度。
正式原版EXE身份、dosgolem版本、Docker工具鏈及位址基準沿用本文件。

- 重生入口：`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --king-approach-original`。固定乾淨來源2f44a68，960秒外層逾時，原版及來源唯讀。
- 生成器：`tools/dosgolem_newgame_probe.py`的`king_approach`情境；唯讀觀測9個位置及原始DGROUP欄位，不預先將未知欄位命名。
- 來源核對：Docker內執行`python3 /repo/tools/verify_dosgolem_king_approach.py --output /work/issue4-king-approach-source-audit.json`。核對前38次輸入、前76次IRQ1、能力seed、母親事件及先前全部PNG／bin保持，再核對新增輸入、日誌及產物雜湊。
- 原版收據：`work/dosgolem-opening/issue4-king-approach-receipt.json`，SHA-256 `c8548e8459daa90e1980d3a9a6d617b5344edf3f921d2b8f06cf5812e28eeac6`。47次輸入、94次IRQ1、198個唯一產物已核對；174個舊PNG／bin不變。
- 隔離重製比較：`work/issue4-king-approach-prototype.py`固定de52686，從正常新遊戲及已通過的42個城鎮狀態續行9次上鍵。輸出`work/issue4-king-remake-approach.json`及9張PNG；不修改production規則。

隔離重製第1至8步由(21,16)至(21,9)，第9步正常轉入CTY25 section0的(15,30)。
原版也自然抵達CTY25 section0的(15,30)。定時擷取的第1步尚在更新，第2至9步位置吻合；
不把任意步數畫面當作完成影格。原版完整等待圖在linear2111B的caller返回0110:7E00，
對應IDA linear17DFB→17E00。該等待由1991D的閒置計數分支進入17DBB，含狀態窗，
並非每步地圖完成閘門。本批只有第9步自然觀測到該等待，不宣稱9個完成影格。

城堡視野有限READY依據：原始EXE的11971..11991讀玩家座標減9、7後呼叫11D8A，
原版與重製到同一進城位置，隔離攝影機原型使進城定時圖完整RGB差異從171566降至2775。
兩張圖片目視核對。CTY25.DAT為3756 bytes，SHA-256
`11d5c60377c6a98bbb9cfc9532652c5e23f4e5c939e769397e8fa5b2f22f9e2b`；
section0基底file0006，界外圖塊在file0018，值0。現行renderer將Y夾至17，
使主角顯示在底部；契約指定視野原點(6,23)，將主角置於視野圖塊座標(9,7)。
這些定位、資料與同位置試作共同支持本scene的D3幾何，不外推其他scene。

正式實作範圍只限CTY25 section0攝影機。`interface.scene_cameras`保存明示的scene與camera，
共用引擎只查引用及套用既有player_anchor原語。宣告須完整、唯一且帶D3，缺欄位／未知欄位
或重複scene拒絕。集合為空可表示尚無額外宣告，集合缺失拒絕。新版存檔依schema及pack hash拒絕舊資料。
原始EXE／DAT核對、完整正常輸入、攝影機參數、標題存讀檔及完整回歸是驗收範圍。

完整等待圖與攝影機原型仍差12726像素，包含閒置狀態窗及頂部圖塊。右上CTY tile高位含8000h，
語意及consumer尚未閉合，不猜補屋頂或影格。完整RGB、閒置狀態窗與謁見時序保持DRAFT。
11994返回時4F25／4F27已被renderer重用；本次值(6,38)不可解讀為視野原點。
只用11971..11991原始writer作幾何證據，不以任意時刻的raw欄位覆蓋已確認的運算。
有界IDA匯出與唯讀snapshot堆疊檢查分別在`work/issue4-king-map-wait-ida.json`、
`work/issue4-king-status-wait-ida.json`與`work/issue4-king-state-stack.json`；均保留原始定位與輸入身份。

| 原始定位與consumer | 推論等級與限定語意 |
|---|---|
| IDA9.4 linear11971..11991／file2CE1..2D01→11D8A，EXE身份沿用本節 | confirmed；玩家減9、7，僅CTY25 section0已有正常同位置與試作閉合 |
| CTY25 section0 file0006，file0018→既有section loader／界外consumer | confirmed；界外圖塊0，不外推其他section |
| IDA9.4 linear1991D→17DBB→17DFB→2111B，實際返回0110:7E00 | strong；原始閒置分支及狀態窗呼叫鏈。第9步自然等待caller為confirmed，正式UI與時間契約仍unknown |
| CTY25 row23 x24..26的raw8018／8004／8221，row24的raw801A | confirmed原始資料；8000h語意及consumer為unknown，不命名屋頂 |

IDA linear−EC90=file，DGROUP基底linear24DD0；工具、EXE大小與SHA-256沿用本文件來源契約。
唯讀堆疊檢查只解碼已保存的原版狀態，未恢復或執行該snapshot，不當冷啟動對拍。

第一次wrapper收尾解析失敗源於執行中編輯腳本，探針及獨立來源核對成功。
固定腳本後以同一Docker命令乾淨重跑退出0；保留第一次收據及產物的內容雜湊歸檔，不記為產品缺陷。

### 城堡攝影機正式驗收

上述幾何經有限READY審查後正式接入，schema0.5.0／content0.1.73，canonical hash
`sha256:27423313471ff2b262ae1805a6177899a1173f025c8bea99979b329e4e403064`。
`internal/gamepack/scene_camera_test.go`核對原始EXE anchor、CTY25 section0界外資料及八種損壞契約拒絕。
`game/opening_scene.go`只選pack引用；建立Game前由原生Town解碼器驗證CTY、section與界外圖塊，
無素材／未知引用／不符資料均拒絕。`internal/dq3data/townmap.go`新增原始`ExteriorTile`欄位，沒有版本fallback。
正常入口驗收在`game/scene_camera_test.go`，重用完整母親返回測試，再送原版九次上鍵。
47次正式輸入自然抵達(15,30)，camera原點(6,23)，Save及標題正常Load維持camera與旗標。
既有城鎮最後完整RGB差0保持。正式城堡PNG與隔離camera原型逐byte相同，SHA-256
`3e698b90d0ce5a81e2e5d0648d5e4f7b2c95461d6d004cf97d421b02c6ac29a1`。
完整RGB診斷仍定時2775、完整等待12726；僅幾何限定CONFORMED，不稱城堡整張V3或謁見完成。

game372項頂層／48子PASS、38選用SKIP，410個頂層完整清單由長主線及四批新程序完全覆蓋、無重複。
正常新遊戲至THE END130.22秒，僅屬重製回歸。internal142頂層／211子與11套件PASS、4選用SKIP；desktop建置PASS。
沒有素材缺失跳過。單程序完整game兩輪均在最後進城測試異常退出；第二輪Docker事件確認OOM。
排除長主線的一次大批執行亦晚段退出，不能算整批PASS；相同binary、工具鏈、容器限額與斷言分為四批後均PASS。
失敗log保留，未改遊戲規則、seed或斷言。internal首輪三個拒絕案例因fixture未宣告新必要集合而提前被拒絕，
fixture明示空集合後，原拒絕斷言保持，同工具鏈乾淨重跑PASS。

本機正式產物：`work/issue4-castle-camera-production.png`及`.json`；來源核對為`work/issue4-king-approach-source-audit.json`。
完整分批工具為`work/issue4-castle-camera-partition.py`，log及清單為`work/issue4-castle-camera-partition-{campaign,batch1,batch2,batch3,batch4}.jsonl`與各自`.selection.json`。
internal與建置為`work/issue4-castle-camera-final-internal.jsonl`、`work/issue4-castle-camera-desktop`。
最終私有稽核入口為`work/issue4-castle-camera-final-audit.py`，輸出`work/issue4-castle-camera-final-receipt.json`。
收據7853bytes，SHA-256 `c20eebe898042be5367e7dd0bfef13dea86909495de00511183cb085a82cd41e`；八種損壞來源均拒絕。
提交、推送與遠端Issue核對保存於`work/issue4-castle-camera-post-push-receipt.json`。
原始失敗log為`work/issue4-castle-camera-game.jsonl`、`issue4-castle-camera-final-game.jsonl`及`issue4-castle-camera-partition-rest.jsonl`。
本批新增工具、契約、測試及本機產物均由本節索引。下一切片是正常進城後閒置狀態窗與8000h圖塊consumer，再續行謁見。

### 2026-10-03 進城後圖塊與閒置狀態窗追查（DRAFT）

接續0900bd2，以Issue #4留言5960375932登記；以下追查起點為schema0.5.0／content0.1.73，最新正式驗收見本節末。
先以一次性IDA9.4資料庫追11971→11D8A的圖塊consumer，以及17DBB→18222的狀態窗內容。
有界匯出腳本為`work/issue4-castle-consumers-ida.py`，sidecar為`work/issue4-castle-consumers-ida.json`。
層欄位xref與取址候選追加匯出為`work/issue4-castle-layers-ida.py`及`.json`，不因直接xref缺writer就宣稱沒有寫入。
完整writer追加為`work/issue4-castle-layer-writers-ida.py`及`.json`。重生前保留manifest於`work/issue4-castle-prior-source.json`。
同條47次正常輸入追加`DQ3_LAYER_TILE`唯讀觀測，記錄右上四格原始BX、層selector及替代圖塊欄位，沒有改原版狀態。
隔離試作由`work/issue4-make-castle-layer-prototype.py`產生`work/issue4-castle-layer-prototype.py`，固定0900bd2重播正常輸入。
試作只驗證來源分支的可見影響，版本值暫留隔離副本，不進正式程式。輸出`work/issue4-castle-layer-prototype-approach.json`及九張`-north-NN.png`。
沿用本節EXE大小、SHA-256與IDA linear／MZ file換算；原始名字、bytes、xref及未證實警示保留。
追查時未將未知8000h語意或閒置時間契約放入正式程式；圖層原始資料與consumer閉合後依下方有限READY實作。
前輪原始c8548e84收據已依內容hash歸檔至`work/dosgolem-opening/issue4-archive-c8548e8459daa90e1980d3a9a6d617b5344edf3f921d2b8f06cf5812e28eeac6.json`；原先manifest另保留於`work/issue4-castle-prior-source.json`。

#### 圖層遮蔽來源與有限 READY

同條正常47次輸入重生退出0，新原版收據SHA-256
`16f0c0fc60f5f4aeb96e452eb0d0a6c25024c23a3c2662c90262a5722f50e918`。
198個唯一產物全部核對，先前196份PNG／bin逐項相同。最新獨立核對為`work/issue4-castle-layer-source-audit.json`。
原版右上四格在11E07讀到8018／8004／801A／801A，層selector00，替代欄位0B56=1B、0B57=46；
八個指令後於11E4B皆為801B，證實此正常進城狀態使用tile27。沒有改原版位置、旗標或layer。
隔離試作同條正常InputState的最後完整640×350 RGB差0，私有比較為`work/issue4-castle-layer-prototype-rgb.json`。
該圖位於閒置窗之前；不把它提升為包含閒置窗的完整等待畫面或九個完成影格。

| 原始定位與資料流 | 推論等級與範圍 |
|---|---|
| IDA9.4 linear13162..1317D／file44D2..44ED，section+15／+16→DGROUP0B56／0B57 | confirmed原始header與writer；本CTY值27／70。70的非零層動態畫面未抽樣，呈現資料為D2 |
| linear13135..1315A及11910..1195D，玩家格高byte與C0h→DGROUP2579 | strong；兩條原始取址與selector writer閉合，正常進城的selector00有動態確認 |
| linear11E07..11E4B／file3177..31BB→2D0A圖塊表 | confirmed零層分支：其他層替換為0B56；非零層不同類別替換0B57的靜態分支為strong |
| CTY25 section0，raw高byte的上兩bit | confirmed原始資料及零層遮蔽consumer，不將8000h命名屋頂；未核對其他層的轉換動畫或NPC遮蔽 |

typed契約：原生Town parser保留section+15／+16並把高byte上兩bit解為0..3圖層。
正式入口為`game/opening_scene.go`、`game/game.go`、`internal/dq3data/townmap.go`；
契約與原始資料核對見`internal/gamepack/scene_tile_layers_test.go`，正常玩家與來源拒絕見`game/scene_camera_test.go`。
新增表頭長度拒絕抽樣見`internal/dq3data/townmap_layers_test.go`。
pack新增必要陣列`interface.scene_tile_layers`，可明示空陣列；每筆CTY／section唯一，包含
`mode=player_cell_layer`、`base_layer`、`base_tile`、`other_tile`及分級證據。
本pack只宣告CTY25 section0、base_layer0、base_tile27、other_tile70。缺欄位、null、未知原語、
越界圖號／圖層、重複或無camera引用均拒絕；建立Game前核對實際CTY／section與兩個header圖塊。
另以正常loader的實際BLK count拒絕替代圖塊越界，不能把byte上限當成素材數量。
base_layer時，其他層的格子改畫base_tile；非base_layer時，非玩家所在層及界外畫other_tile。
保留原生移動／事件圖號，繪製選擇不改碰撞、事件、角色、亂數或圖塊資料。
只在宣告場景套用；不宣稱其他layer的轉換動畫、NPC遮蔽或整個城堡已V3。

資料欄位最低D2；此次玩家可見零層分支及base_tile達D3，其他tile70依原始header／consumer為D2，
只改繪製，不增加正式流程gate。缺失值不猜補，不把static strong寫成全分支動態confirmed。
schema0.6.0／content0.1.74；同版Save及正常標題Load須恢復目前格、hiMap與相同繪製規則。
驗收包含原始EXE／CTY parity、必要欄位及未知來源拒絕、正常47次進城、前閒置全RGB零差異、
同版本標題讀檔、後續正常主線回歸、完整game分批及internal／desktop。此有限規格審為READY。

閒置窗仍DRAFT。1991D的原始比較為FFh與12Ch，不能擴成FFFFh或12C300h；
時間域尚未審查，不據此猜秒數。18222已定位隊伍姓名、HP、MP、狀態／等級及職業consumer，
正式閒置window與時序須另閉合，這輪不以測試注入開窗取代正常閒置流程。

#### 讀檔畫面驗收勘誤

新增加的「讀檔後也須與進城原版PNG零差異」斷言連續兩次差304，未視為產品缺陷。
診斷`work/issue4-castle-layers-load-diagnostic.log`證實正常朝向1、讀檔0，walk皆0、layer皆0；
全畫布差異僅在(288,176)..(319,191)主角範圍。現行remake的saveState不保存朝向，
原版讀檔朝向尚未取得oracle；新斷言誤將不同朝向當同狀態，不能以調影格或注入朝向修成零差異。
回到證據審查後，保留47次正常進城的嚴格全RGB零差異，以及camera、layer宣告、完整hiMap、目前格與旗標的標題讀檔斷言。
讀檔完整RGB保持診斷304，不裁切／遮罩，正式收據明示`original_save_load_oracle=false`與`title_load_full_rgb_parity=false`。
這不是原版存讀檔完成聲明。讀檔後所有場景的完整原版畫面與朝向仍是待驗gate。
新的正式PNG／JSON與讀檔PNG為`work/issue4-castle-layers-production{.png,.json,-title-loaded.png}`，
乾淨通過log為`work/issue4-castle-layers-production-clean.log`；第一次失敗`work/issue4-castle-layers-production.log`與唯讀診斷log保留。
完整game由`work/issue4-castle-layers-partition.py`分為campaign與四批，產物同前綴的`.jsonl`及`.selection.json`，不覆寫前輪camera收據。

#### 零層圖塊遮蔽正式驗收（有限 CONFORMED）

schema0.6.0／content0.1.74，canonical hash為
`sha256:bb8a154a01619c2d0541eab4055fa2e2485d7babd9bef744ec322d9b1f25f9c8`。
正式正常47次InputState從新遊戲進城，前閒置640×350 RGB由2775差異降為0，與隔離試作逐byte相同。
正式PNG SHA-256為`4744e8a26b67e786313f4becc9f9ae1452811d71abf237885b048b620300153c`，已目視核對。
完整閒置等待圖由12726降為9951，仍RED；不把前閒置對拍擴成完整等待、謁見或整個城堡V3。
城鎮42狀態及全RGB零差異保持；同版標題讀檔保留camera、完整hiMap、layer、位置及旗標，朝向差異仍依上節限制記錄。

12種pack契約損壞、5種runtime來源損壞與原生表頭截斷拒絕通過。
來源稽核新增5種層觀測損壞，連同8種原有收據損壞均拒絕；負例同時修改log及metadata，避免只證明兩者不符。
三份IDA sidecar共逐列核對原始file bytes及MZ relocation，保留原名、位址、推論等級與未證實警示。
最小充分收尾入口為`work/issue4-castle-layers-final-audit.py`，輸出`work/issue4-castle-layers-final-receipt.json`；
核對198份原版產物、196份歷史PNG／bin不變、正式零差異與未知gate、完整測試清單、輸出UID／GID及主機衛生。
完整game373項頂層／53子PASS、38選用SKIP，411個頂層由campaign及四批覆蓋且無重複或遺漏。
internal145項頂層／223子與全部11套件PASS，4選用SKIP；桌面ELF建置通過。
正常新遊戲至THE END184.69秒，只屬remake回歸；原版完整campaign與音訊仍未CONFORMED。
internal收據為`work/issue4-castle-layers-internal.jsonl`，桌面產物為`work/issue4-castle-layers-desktop`。
首輪分批工具誤指不存在的`-final.test`，未開始測試；改為實際`-layers.test`後以相同容器與命令乾淨重跑。
隔離試作先有shell引用錯誤、再有helper回傳值編譯錯誤，修正工具後乾淨log為`work/issue4-castle-layer-prototype-clean.log`。
未改正式規則、seed或角色frame來消除驗證失敗。所有原始資料與私有圖像仍不進Git；沒有新發行包。
提交後稽核入口為`work/issue4-castle-layers-post-push-audit.py`與同前綴`-receipt.json`，核對遠端commit、Issue與清理狀態。
遠端結果留言5961178819；Issue #4保持OPEN，後續以其未完成項目為準。
下一切片為閒置狀態窗的正常開關、資料consumer及時間域。非零層轉換動畫、NPC遮蔽、原版讀檔朝向與謁見另保留待驗。

#### 閒置狀態窗生命週期追查（DRAFT）

接續bcc6ce0，以Issue #4留言5961286946登記。正式schema0.6.0／content0.1.74保持。
使用已驗證IDA Pro9.4 image及同一唯讀EXE，原始身份、linear／file換算沿用上節。
有界匯出入口為`work/issue4-idle-status-ida.py`，sidecar為`work/issue4-idle-status-ida.json`；
保留原名、raw bytes、MZ relocation、xref type及未證實警示，沒有覆寫函式或資料名稱。

| 原始定位 | 目前證據與未閉合範圍 |
|---|---|
| IDA linear1991D..1997C／fileAC8D..ACEC | confirmed有限正常入口：4F1F為FFh，三次差值300進19952；三組298／299／300與PIT ticks各遞增一。恢復0000／0007／0013的靜態指令為strong；19966觀測在pop0000之前，不宣稱已動態讀取pop後值 |
| linear17DBB..17E11／file912B..9181 | confirmed單人正常開關：三次寫入相同動態header，三次2111B等待及兩次17E11恢復；完整底圖恢復差異僅在主角動畫。其他隊伍人數維持strong |
| linear18222..182E3／file9592..9653 | confirmed本角色內容：姓名0、HP15、MP9、等級1及職業glyph；原生58bytes角色資料全程不變。四字上限、十byte欄距與異常狀態分支仍為strong |
| linear1F590..1F603／file10900..10973 | strong：window+0A的record401為左框、+0E的402依人數重複、+10的403為右框。此組是D3TXT00視窗record，不讀目前城堡對話bank |

新增`king_idle`情境沿原先47次正常輸入進城，再送兩次真實上鍵觀察關窗與續行。
入口`tools/dosgolem_newgame_probe.py`，私有前綴`work/dosgolem-opening/issue4-king-idle-*`，
保留唯讀`DQ3_IDLE_STATUS`及`DQ3_IDLE_WINDOW`觀測，不注入計數器、視窗、位置、朝向或旗標。
本情境不覆寫前輪`issue4-king-approach-*`收據。時間域、按鍵消耗及再次開窗以新的動態結果為準。

隔離布局原型入口為`work/issue4-idle-window-layout-prototype.py`，固定bcc6ce0。
重播正式新遊戲及47次輸入後，以已定位的原始record、姓名與數值consumer組合畫面。
輸出同前綴`.png`、`.json`及`.log`；乾淨重跑log為`-clean.log`。
這是layout-only原型，沒有正式idle狀態或timer，不能當成正常閒置生命週期已完成。
首次原型差1299，因誤讀目前城堡對話bank而缺少視窗frame／標籤；改讀原始D3TXT00後再核對。
第二次仍差1315。重查GUI composition與同狀態路由後，逐像素分類為1117個前景色、16個未覆蓋冒號及182個主角相位差異。
追加IDA窄匯出`work/issue4-idle-status-number-ida.py`及`.json`，linear219AA..219F3證實三位數欄位的前導空白glyph0C。
原始DGROUP25D6／file18716的正常色盤表為3C／3C／3C；沿用已審查開場前景色243／243／243，沒有從圖像手調座標或字距。
第三次原型完整RGB仍差182，全部在主角(288,179)..(319,191)；視窗及陰影區域逐像素差0。
最新執行log為`work/issue4-idle-window-layout-prototype-reviewed.log`，完整差異保留，`full_rgb_parity=false`。
原型執行PASS只表示組合與輸出成功。尚缺同影格oracle的完整零差異斷言撤回，不把剩餘主角差異遮罩或裁切掉。
獨立稽核入口`work/issue4-idle-window-layout-audit.py`及`.json`，核對三份IDA共1740列原始file bytes與MZ relocation、全畫布差異及布局區域。
正常色盤consumer另存`work/issue4-idle-status-palette-ida.py`及`.json`：linear1EF5E..1EF94，
25D5選三byte表項，更新六個色盤bank的index8，再提交25D1指向的色盤資料。
原版等待PNG的PLTE index8亦為243／243／243，與原始3C／3C／3C及已有開場前景色一致；其他健康色狀態尚未動態抽樣。

| 有限typed資料／行為 | 原始約束與尚待驗收 |
|---|---|
| 視窗起點、高度 | x=19個byte、y=238、height80；17DC5／17DCB與原始3EA4 |
| 動態寬度 | 隊伍人數×10個byte+4；每欄80px、基礎寬32px；17DD4..17DDB |
| 左框／每欄／右框 | D3TXT00 record401／402／403，1F590按實際隊伍人數重複body；不是三筆玩家對話 |
| 姓名 | x為視窗+32px，最多四glyph；215EE..2164F，每glyph正常步距16px |
| HP／MP／等級 | 同三位數欄位、右對齊、前導空白glyph12；219AA..219F3。末列有狀態時改畫status record，尚未動態抽樣 |
| 職業 | 182BE..182CB使用既有pack的class單字glyph，x為視窗+16px；不在引擎計算版本record或glyph |
| 關窗與恢復 | 兩次正常上鍵只關窗，位置仍15／30；完整底圖恢復差異僅在主角。靜態恢復計數／輸入不改稱動態pop後讀值 |
| 等待延遲 | 三次0000／000D差值門檻300，三組邊界各與PIT ticks遞增一；實際除數12428。正式60TPS換算仍須引用平台契約，不能用300個remake更新 |

重生入口為`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --king-idle-original`。
支援固定2f44a68執行器、正常49次輸入及唯讀閒置觀測；須沿用檔首Docker／權限／來源契約。
核對器`tools/verify_dosgolem_idle_status.py`與私有`work/issue4-king-idle-source-audit.json`僅核對原版，
不把layout-only原型升為remake生命週期parity。既有正式原版收據維持196份PNG／bin前綴對照。
首輪原版延長路線在900秒工具期限終止，最後送達92次IRQ1，沒有完整閒置收據。
失敗log已按hash保存為`work/dosgolem-opening/issue4-archive-b87d3dd30ec21ae7b65cbe7f88e3f2115c4057b48a4b0f786f1eaf3145313e8f.log`。
增加本情境期限至1200秒、外層1260秒，以相同輸入、seed、原版及image乾淨重跑；未調原版狀態。
乾淨生成log為`work/issue4-king-idle-clean.log`，本情境產物仍按hash歸檔，不覆寫前輪正式進城收據。
乾淨重跑退出0，原版收據`work/dosgolem-opening/issue4-king-idle-receipt.json`為507785bytes，SHA-256
`5cd11f15beb7e8c42f89780da825e5802b5cb38033b683e000bc06e32fb021d1`。
49次正常輸入、98次IRQ1及216份唯一產物通過；原先196份PNG／bin逐項不變。
兩次上鍵分別在2160000000、2200000000排入，實際make／break及窗口關閉均由原版自然送達。

| 順序與原版步數 | 觀測與有限分級 |
|---|---|
| waiting1，2138831661 | confirmed，2111B自然等待，位置15／30，單人 |
| restore2，2160026794 | confirmed，17E11恢復；第一個上鍵未移動 |
| waiting3，2197937916 | confirmed，正常再次開窗，角色資料不變 |
| restore4，2200026793 | confirmed，第二個上鍵未移動；角色資料不變 |
| waiting5，2237864789 | confirmed，正常第三次開窗；不外推其他輸入或隊伍人數 |

三次19952的0000／000D差值皆恰為300。每次前面都有298、299、300的19940觀測，
ticks與0000逐項各遞增一，實際PIT divisor12428保持。此結論只限三次原版閒置邊界，
不將所有遊戲counter改稱PIT ticks，也不證明原版硬體wall-clock或逐週期一致。
58bytes角色資料在35筆唯讀觀測中相同，姓名、職業、性別、等級、目前／最大HP及MP與創角收據吻合。
0007保留raw定位，不根據數值替它命名為掃描碼；19966在pop0000前，靜態stack恢復與動態讀值分開記錄。

完整640×350比較：第一次等待與前輪等待RGB差0；兩次restore相對前閒置圖或前次restore均差182，
全在主角(288,179)..(319,191)。第三次等待相對第一次差182；第五次等待相對第一次差1811，
範圍(0,0)..(382,167)，保留完整場景變化。沒有裁切、遮罩、固定人物frame或重擲seed。
獨立布局原型仍只證實視窗及陰影區域差0，完整畫面差182，不提升為正式UI或整張等待V3。

最小充分收尾入口`work/issue4-idle-status-final-audit.py`及`work/issue4-idle-status-final-receipt.json`，
12469bytes，SHA-256 `0e9ce1b1db36a3a1a64737094e171ad615102b3b6193f927b023e45e24d9ab1c`。
核對216份來源、196份前輪圖像、三份IDA原始1740列、完整原版五組圖像比較及布局原型。
16種損壞收據全部拒絕，包含timer間距、actor變動、位置、header、開窗／關窗順序及IRQ1；
負例同時修改暫存log與metadata並重算manifest，原始收據不改。Python／shell語法及UID/GID1000核對通過。

正式schema0.6.0／content0.1.74、Go及pack未改。正常單人開關、資料呈現及有限時間域已取得D3來源；
正式UI仍DRAFT，下一步審查具名狀態機的正常輸入／held-key所有權、平台規格換算及存讀檔暫態影響，
通過READY後才接入JSON與正式玩家路徑。其他健康色與status分支未動態抽樣，不以本收據宣稱完成。
遠端結果留言5961959766，Issue #4保持OPEN。提交後稽核入口為`work/issue4-idle-status-post-push-audit.py`及同前綴`-receipt.json`，核對遠端commit、Issue與容器清理。

#### 閒置狀態窗正式契約審查（有限 READY）

接續acc3819。正式入口只套用已有正常原版來源的CTY25 section0；其他場景不由此宣稱對拍完成。
新增`interface.field_idle_status`，共用引擎以具名等待／開窗／按鍵消耗狀態機處理，
不保存版本座標、record、遮罩、字串或硬體除數。布局引用既有party_hud的幾何、姓名及職業字模，
左右frame及每欄body以獨立text ID引用，寬度由原始record形狀與實際隊伍人數組合。
所有新巢狀欄位必填，缺失、null、未知欄位、無效來源或素材越界拒絕；最多人數沿既有HUD四欄容量。

平台前提依[DOSBox Staging timer.h](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/hardware/timer.h)的PIT_TICK_RATE1193182及
[timer.cpp](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/hardware/timer.cpp)的counter／時脈週期契約。
遊戲除數12428及300次閒置邊界由5cd11f15來源取得。固定60TPS的延遲為
ceil(300×60×12428／1193182)=188更新；只屬hardware-spec approximation，不宣稱原版wall-clock或逐週期一致。
一般輸入或其他modal重置等待；已開窗時凍結底圖，按鍵只關窗，方向held在該次按下未放開前持續消耗，
放開後才能接受下一次移動。原版兩次make到break間未移動提供本有限方向封包的D3來源，未外推鍵盤自動重複的精確時序。

存檔不新增視窗或計數欄位；同版本Save／標題Load保持位置、隊伍、HP／MP、條件與旗標，
restore清除凍結底圖、等待與按鍵暫態，再正常等待開窗。此為重製自身round-trip，不宣稱原版存檔格式或朝向parity。
開窗、等待與關窗只讀角色內容，不消耗RNG、HP、MP、金錢、物品或旗標。

狀態末列的DI起點是0x24C，即十進位588，依七筆mask優先序使用588..594。
前輪筆記若將其轉為596則為十進位轉換錯誤，原始bytes及定位保持，追加此勘誤。
`work/issue4-idle-status-strings-ida.json`閉合213C4→21286→252E的原始文字reader，
`work/issue4-idle-status-record-index-ida.json`確認未達3000的DI直接查record指標；兩份同名.py可重生。
原始D3TXT00的588..594依序是死亡、麻痺、睡眠、混亂、瑪荷、瑪努、中毒；私有原始字模圖為
`work/issue4-idle-status-glyphs.png`。mask、健康色五表及欄位consumer均為D2／strong資料，
只宣稱正常單人的D3動態呈現；其他健康色與狀態的原版動態抽樣仍未完成。

實作與驗收入口：`internal/gamepack/field_idle_status.go`及其`_test.go`保存typed契約、reference validation與原始EXE／TXT parity；
`game/field_idle_status.go`及其`_test.go`保存正式正常入口、source validation、按鍵消耗、三輪自然開關及Save／標題Load。
JSON欄位權威見[docs/84](84-game-pack-json-contract.md)。需保存完整640×350 PNG及差異，不用裁切或指定影格消除已知動畫差異。
schema0.7.0／content0.1.75；舊schema與hash的存檔明確拒絕，不自動遷移。
本有限契約經來源、原型、平台與存檔影響審查為READY；動畫與原版讀檔、音訊、完整campaign保持原有未完成狀態。

#### 正式閒置窗驗收（有限 CONFORMED）

正式schema0.7.0／content0.1.75，canonical hash為
`sha256:6fa52a4c7a71f5a8a1462ad4bd22674c05270238500b6770847bb37ddf604f17`。
正常新遊戲、創角、家中選圖、城鎮帶路及九次上鍵進城後，不直接呼叫開窗函式，
由正式InputState自然等待188更新，觀察三次開窗、兩次關窗。
兩個關窗上鍵及持續按住均不移動；放開後新按鍵可往下一格。
開窗期間完整畫布與RNG保持；完整存檔快照逐byte不變。
同一Game的Load及正常標題選單Load均清除等待／底圖／方向暫態，位置與遊戲進度保持。
此續行只證明重製可正常移動，尚未由原版謁見oracle驗收。

三次自然開窗的視窗及陰影區域RGB差0，第一張正式PNG與較早布局原型逐byte相同。
完整五張640×350圖的差異依序為182、182、979、979、1960；完整等待圖依序182、979、1960。
沒有指定動畫frame、裁切或遮罩。圖像差異仍在場景與人物，完整RGB對拍保持false。
較早camera測試的9951是尚未到延遲門檻的前閒置畫面對等待窗的診斷，不能再代表正式等待窗缺失。
正常47次輸入的前閒置完整RGB差0與讀檔304個主角差異保持各自限定範圍。
其他隊伍人數、健康色與異常狀態只有原始D2資料及component tests，缺原版動態抽樣。
時間精度維持hardware-spec approximation，不宣稱原版硬體wall-clock。

字型資源由`font_asset`引用、驗大小與hash並直接綁定renderer；不依賴當前場景對話bank。
十五種壞JSON契約、五種壞runtime來源、健康色正常／半血／四分之一／加總上限／死亡分支通過。
原始EXE／TXT parity覆蓋等待門檻、兩次SHR、健康色上限與死亡索引、palette index、前導空白、
三位數consumer、優先mask、視窗座標、原始record及glyph shape；不刪原生decoder。
列位置亦由JSON提供。linear18263／18270／1828D及182AF的原始`add dx,10h`，
分別閉合HP+16、MP+32及狀態／等級／職業+48；姓名沿既有text_inset_y。
全部位置必填，缺失、null、非字模步距或超出視窗均拒絕，不留共用Go座標fallback。

私有重跑入口為`work/issue4-run-idle-status-production.sh`及`work/issue4-run-idle-status-full-game.sh`，
分批工具`work/issue4-idle-status-partition.py`以同一binary清單覆蓋正常主線及四批新程序。
資料轉換器為`work/issue4-make-idle-status-pack.py`。
正式normal輸出`work/issue4-field-idle-production.json`與同前綴五張完整PNG，
log為`work/issue4-idle-status-contract-clean.log`及`work/issue4-idle-status-production-reviewed.log`。
最終獨立核對入口`work/issue4-field-idle-final-audit.py`及`work/issue4-field-idle-final-receipt.json`，
核對原版216份產物、196份前輪圖像、16種損壞收據、五份IDA raw bytes／MZ relocation、
完整正式PNG、分批完整清單與internal／desktop，以及UID/GID及Docker衛生。
資料轉換器從acc3819的乾淨pack重建，九份JSON與正式檔逐byte比較，避免手編資料無法重生。
最終收據22245bytes，SHA-256 `4135934fa5128573b0f641e4197907ac0fda5ef0514733ed89dd78bbb4c05b74`。
完整game414個頂層清單無重複或遺漏，376頂層／63子PASS、38選用SKIP；
internal147頂層／238子、全部11套件PASS、4選用SKIP；無素材缺失跳過，desktop為Linux x86_64 ELF。
正常新遊戲InputState至THE END226.04秒，只屬remake回歸；五份IDA共1901列原始bytes／relocation核對通過。
最近產品log為`work/issue4-idle-status-partition-{campaign,batch1,batch2,batch3,batch4}.jsonl`及其selection清單、
`work/issue4-field-idle-current-internal.jsonl`、`work/issue4-field-idle-current-desktop.log`。
提交後核對入口為`work/issue4-field-idle-post-push-audit.py`及同前綴`-receipt.json`。
所有原版素材、圖片、IDA database／授權與完整包留在本機；未建立新發行包。
遠端結果留言5962628881及Issue主文已更新，Issue #4保持OPEN。

#### 正常關窗後王座路線來源（DRAFT）

依Issue #4接續572e3bb，工作登記5962742201。正式schema0.7.0／content0.1.75維持。
沿用同一唯讀EXE、CTY25、固定dosgolem2f44a68及自然Lv1入口一次seed1357。
來源入口為`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --king-audience-original`。
`tools/dosgolem_newgame_probe.py`的`king_audience`情境保留原先49次輸入，
在第三次自然等待窗後追加上鍵關窗，再送35次上鍵與5次Enter；共90次正常鍵封包。
第一步行鍵距關窗鍵8000000個原版steps，其他步行／確認鍵間隔20000000；不由時序假設宣稱完成移動。
工具期限1800秒、外層1860秒；所有執行、建置及圖像擷取仍在一次性Docker，原版與dosgolem來源唯讀。

原始CTY25的section0北側樓梯資料指向section1的9,22，只作正常輸入路線的定位線索。
不以舊remake、現有獎勵測試或影片描述代替本次原版結果。
新增唯讀`DQ3_KING_AUDIENCE`觀測保留player原始位置、raw0B24／0B55／4F1F、文字DI、
返回堆疊、計數、人物前128bytes與64bytes旗標。不替暫時DS或raw0B55命名為目前section。
linear1991D、213C4及2111B用於候選循環／文字入口／等待位置；標籤ready與waiting仍須caller及動態結果核對。
每次新增輸入後保留定時完整PNG，另在候選自然等待或循環位置保留完整PNG／色號資料。
私有產物前綴`work/dosgolem-opening/issue4-king-audience-*`，不覆寫既有idle或approach來源。
來源生成PASS只能表示探測收尾成功；王座抵達、文字、旗標與獎勵、RGB及remake parity均待獨立核對。

#### 王座接近來源核對與攝影機審查（有限 READY）

原版來源正常結束，90次輸入／180次IRQ1、386份唯一產物全部核對，前輪214份PNG／bin逐byte不變。
收據`work/dosgolem-opening/issue4-king-audience-receipt.json`的SHA-256為
`325b5e7be2ceea0fa2a0fe0bbc291dcfed11c7c30d90cc3865f96d208defcc04`。
獨立入口為`tools/verify_dosgolem_king_audience.py`，輸出`work/issue4-king-audience-source-audit.json`。
生成腳本、Go probe原始碼與執行檔均留本機，前綴同原版收據；Go原始碼hash與當次metadata相符。
初次Go source由執行中容器另存，不在386項manifest內，仍以metadata的hash獨立核對。
收尾後producer自動保存Go source並納入manifest，後續同路線重生為387項；額外一項只為來源文字，
不改輸入、觀測或圖像。本次歷史386項收據及生成腳本保持，不重寫為新版產物數。
上樓後額外閒置窗消耗了一次上鍵，最後停在9,8，五次Enter未觸發handler56。
本收據只能證實正常上樓及接近，不能宣稱已完成謁見、獎勵或國王文字。
linear1991D的ready只代表候選循環取樣；兩個堆疊欄位保留raw words，不當作far return IP／CS。
213C4是簡單文字reader，不能代替真正國王對話21414的入口觀測。

第一個新視野差異出現在CTY25 section1的正常落點9,22。
來源新增第21次輸入為第20次步行，上樓前section0基底file0006，上樓後raw0B24為08C7，
與原始CTY25 section1基底file08C7相符；不從raw0B55推定section。
正常原版`audience-north-20.png`與乾淨572e3bb的相同落點完整640×350比較差86603像素。
舊重製clamp使camY=11、主角低96px；原版11971..11991讀4F33／4F35減9／7，故camY=15。
section1的界外圖塊由file0x08C7+0x12取得27，沿用130F4 loader與11DD8 consumer。
本有限場景的視野及界外資料為D3／confirmed，依相同來源與既有camera契約接入。

隔離入口`work/issue4-throne-camera-prototype-v2.py`從乾淨572e3bb、正式新遊戲與InputState重播，
只在私有副本增加section1的SceneCamera綁定，沒有控制人物影格、位置、旗標或重新擲seed。
同提交基準／原型輸出為`work/issue4-throne-camera-v2-{0,1}.json`及同前綴完整PNG與log。
完整差異由86603降為2159，仍保留柱子／圖層及人物差異，不宣稱整張RGB通過。
原型前一版後段遭SIGKILL，未產生完整狀態收據；縮小至首次上樓的同工具鏈重跑通過。
另一版少了docker標準輸入傳遞，未執行編輯；保留輸出，v2才是有camera變更與差異欄位的原型。

READY範圍：在既有`scene_cameras`新增CTY25 section1，anchor9／7、exterior27均留JSON。
沿用schema0.7.0，content升0.1.76；不同canonical hash的存檔照既有契約拒絕，不自動遷移。
共用引擎不加版本常數、fallback或新架構。驗收包括EXE／CTY parity、正常落點完整PNG、
正常Save及標題Load恢復camera、放開後新上鍵續行、既有城堡視野／閒置窗與desktop建置。
王座圖層、section1閒置窗、國王文字／獎勵及動畫另保留未完成狀態。

靜態國王順序的私有入口為`work/issue4-king-audience-rewards-ida.py/.json`與
`work/issue4-king-audience-text-ida.py/.json`，均由IDA9.4對唯讀同一EXE的一次性DB匯出。
linear1024C..102C4的原始sub_1024C先以DI0C06呼叫21414，再給六件物品、50金、clear17h／set18h。
21414的FFFC分支21558／2157E呼叫216C3，216D8與216FB輪詢21148；FFFF在21501分支retf。
record3078映射D3TXT01 record78，含九個FFFC等待；此次五次Enter沒有自然抵達該record。
原始6次物品writer16856搜尋全隊8格00FF空槽，金錢writer1895C寫4F37／4F39。
這是strong靜態順序，尚缺本次玩家可見交易閉合，不修改正式獎勵順序或猜補等待時長。
IDA輸出保留原始bytes、MZ relocation、位址與xref type；未更名原始定位，未深入ISR或PCM driver。

正式實作／驗收入口為`game/throne_camera_test.go`及`internal/gamepack/scene_camera_test.go`，
資料重建入口`work/issue4-make-throne-camera-pack.py`，比例驗證入口`work/issue4-run-throne-camera-production.sh`。
正式輸出為`work/issue4-throne-camera-production.json`、同前綴完整PNG／log及`.test`，
`work/issue4-throne-camera-gamepack.jsonl`保存全部gamepack契約結果，desktop與正常campaign採同前綴獨立log。
私有獨立稽核入口`work/issue4-throne-camera-final-audit.py`及同前綴`-final-receipt.json`，
核對完整來源、損壞拒絕、三份IDA原始bytes／relocation、同提交原型與正式PNG、資料重建及工作樹衛生。
前一版隔離正常探測為`work/issue4-king-audience-prototype.py`及`work/issue4-throne-camera-prototype.py`，
不得以無camera修改的前一版candidate輸出替代v2。

#### 王座攝影機正式驗收（有限 CONFORMED）

正式schema0.7.0／content0.1.76，canonical hash
`sha256:ccbf6f5f2add47996f6e48bcd35a81a5553428517a9b0bfbb25436edaf5e7469`。
正常新遊戲至首次上樓共70次輸入的等價按鍵，控制seed與picture BIOS條件沿前批一次設定。
每鍵在冷卻完成後送出及放開，不宣稱CPU steps與60TPS的wall-clock等價。
正常落點9,22、cam0,15與界外27通過；正式完整PNG逐byte等於隔離camera原型。
完整RGB仍2159，保留原版／正式兩張畫面，不裁切、遮罩或指定人物frame。
Save及正常標題Load恢復相同camera、座標、旗標、物品與金錢；後續新上鍵可到9,21。
只屬同版本重製round-trip，未外推原版存檔、自然等待窗或謁見parity。

全部gamepack86頂層／238子PASS，沒有SKIP；六項受影響game及正常主線另用新程序PASS。
正常InputState新遊戲至THE END81.38秒，只屬重製回歸；desktop為Linux x86_64 ELF。
舊城堡前閒置全RGB差0及三次等待窗182／979／1960、視窗區域差0保持。
本批按資料修改比例驗證，未重跑全部game及其他internal，不將前批全套清單算成本批結果。
九種壞來源收據全部拒絕；三份IDA sidecar原始file bytes／MZ relocation獨立核對通過。
轉換器從乾淨572e3bb重建全部九份JSON，與正式檔逐byte相同。
私有最終收據27698bytes，SHA-256
`34085fc4334bee169597bbf026dfdcdeee65850b1c08709840759163bee69ec4`。
本批只閉合camera；柱子圖層、王座閒置窗、國王文字／獎勵、動畫、原版讀檔及音訊保持未完成。

#### 王座柱子與等待窗續行（DRAFT）

接續61f60d3，依Issue #4續查2159個完整畫面差異及section1自然等待窗。
入口`work/issue4-throne-layers-prototype.py`與同前綴輸出保存乾淨提交的正常玩家路徑診斷。
先核對實際Scene圖號、palette及原版載入資料，不預設缺柱由圖層造成。
前輪自然final.state只作記憶體觀察，執行步數維持3100000000，不注入或續行、不擷取畫面；
原版圖像仍採當次冷啟動的完整PNG／bin。所有工作負載仍在一次性Docker，輸入唯讀。
原始CTY25 section1及載入後308E:0B87／0C1F／0B9B／0C33的四個柱子word均為000E。
此為dosgolem執行期seg:offset的memory-only核對；不與IDA linear或file offset混稱同一位址。
原始DQ31.BLK index14與來源首次上樓x128..159/y0..23的768個色號全同，圖塊本身沒有解碼缺口。
NPC表首項是byte count；先前以u16試讀得到1805的診斷已排除，依原生parser核對為13個7-byte record。
正式schema0.7.0／content0.1.76維持，未以DRAFT修改production path。

追加勘誤：前輪「柱子圖層差異」已推翻。獨立PNG解碼與原版bin逐色號差0，
完整RGB差2159只在人物圖格(3,2)、(3,3)、(16,3)、(15,4)、(10,6)、(9,7)，
各為474、474、428、406、206、171；四根柱子及其他底圖差0，不新增section1圖層來修假缺陷。
私有獨立入口`work/issue4-throne-png-bin-audit.py`與同前綴JSON保留完整畫布比較。

王座等待窗DRAFT沿用既有狀態機及資料契約，只新增CTY25 section1場景綁定。
原版90次輸入收據的第7次waiting／第8次restore均在9,22；17DBB、17DE5、18222、18259
使用與前層相同的動態header及58bytes角色資料。19940正常298／299／300邊界及19952入口閉合。
原始定位與EXE身份沿上節已核對IDA9.4，新增動態來源僅擴大正常單人場景適用範圍。
私有同提交試作入口`work/issue4-throne-idle-prototype.py`，產物同前綴baseline／candidate。
不控制位置、旗標、人物frame或亂數重擲；188更新仍為hardware-spec approximation。
審查與正常開關、完整PNG、後續新上鍵及標題存讀檔通過前，正式資料包維持。

#### 王座等待窗證據審查（有限 READY）

同提交試作只加入既有`field_idle_status.scenes`的section1綁定，自然等待窗與陰影區域差0，
完整640×350仍1772差異。基準未開窗，兩張完整圖及差異數值均保留，沒有指定人物frame。
原版第7次waiting、第8次restore保持位置9,22，動態header及actor與前層相同，
PIT／計數298、299、300及19952入口已閉合。`tools/verify_dosgolem_king_audience.py`
追加來源log一致性、同header／actor、門檻與正常關窗斷言；原始90次收據及產物不改。

READY只增加已證實場景引用，schema0.7.0維持、content升0.1.77。
不增加共用Go分支或版本常數；原有等待、模態凍結、方向鍵消耗、恢復與存檔暫態契約保持。
formal驗收需正常新遊戲至首次上樓、188更新門檻、同一自然等待窗與陰影RGB零差異、
完整等待及恢复PNG、關窗不移動、持續按住不移動、放開後新按鍵續行與正常標題存讀檔。
整張畫面、動畫、國王交易、原版讀檔與音訊維持未知，不能由局部視窗零差異外推。
正式入口為`game/throne_idle_status_test.go`，沿用`game/throne_camera_test.go`正常落點helper；
資料重建及比例驗證入口為`work/issue4-make-throne-idle-pack.py`與`work/issue4-run-throne-idle-production.sh`。
新增工具與稽核入口`work/issue4-throne-idle-final-audit.py`、同前綴收據均掛在本節。

首輪七項共用程序以137終止，最後一項未有結果；保留`work/issue4-throne-idle-process-137-*.log`，
改同一容器及binary逐項新程序。新測試首次把Save前／後respawn一起比較而失敗。
`work/issue4-throne-idle-snapshot-diagnostic.log`證實僅Save既有checkpoint更新，未有UI進度改變。
驗收分別比較自然開窗前後、明確Save後的關窗前後及標題Load完整快照，正式存檔規則不改。

#### 王座等待窗正式驗收（有限 CONFORMED）

schema0.7.0／content0.1.77，canonical hash
`sha256:9d2f7a1aa24b999258e8645258ae9f7af8a8c5169088a9bc8c5d495d96fdc5c0`。
正常新遊戲至9,22，自然188更新等待、開窗凍結、上鍵只關窗、持續按住不移動、放開後續行通過。
等待及關窗不改完整進度快照；明確Save既有respawn更新另取快照，正常標題Load完整恢復。
同一Game的Load清除暫態，重新等待可開窗；沒有修改存檔規則、人物frame、seed或正式輸入。
正式等待PNG逐byte等於隔離試作，視窗與陰影RGB差0，完整等待及恢復均1772，整張V3未完成。
同提交基準未開窗，完整差10143、視窗區域8458；首次上樓仍2159且只在人物圖格。

比例驗證為全部gamepack86頂層／238子、七項受影響game與desktop Linux x86_64 ELF，無SKIP。
正常InputState新遊戲至THE END129.64秒，只屬重製回歸；全部game及其他internal未重跑。
十二種壞原版收據拒絕，新增三項同時修改暫存log／metadata及manifest，驗證同header、門檻與位置。
九份JSON由乾淨61f60d3重建逐byte相同。原版386份產物與前輪214份PNG／bin保持，未重寫來源。
完整正式產物、逐項乾淨log與來源核對使用`work/issue4-throne-idle-*`前綴；
獨立收尾入口`work/issue4-throne-idle-final-audit.py`產生6041bytes收據，SHA-256
`131aa62782373f3c6b6bcca7ecc2c7c4be9bc30b8f17fb3ee4c63ac023dc002b`。
輸出UID/GID1000，既有root候選3213、Markdown目錄0保持；一次性容器皆已移除，沒有新image或發行包。
下一切片延長正常國王路線，先由dosgolem取得自然文字及獎勵交易；動畫、原版讀檔、音訊及完整主線保持待驗。

#### 正常國王文字與獎勵續行（DRAFT）

接續e56a8e3，正式content0.1.77維持。延長原版正常90次輸入，新增3102000000步的上鍵，
再於3130000000起每20000000步送九次Enter，終點3320000000；尚不由排程推定九次等待已完成。
前批最後按鍵3070000000且正常ready在3070527441，延長鍵只用來接近事件格，不注入位置或旗標。
所有既有capture及90次輸入保持，新情境`king_text`／前綴`work/dosgolem-opening/issue4-king-text-*`。
觀測入口為`tools/dosgolem_newgame_probe.py`，重生入口為
`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --king-text-original`；
唯讀追1024C原始handler、21414文字reader、216C3／216D8等待、10268返回及逐次交易定位。
保留原名、EXE hash與IDA linear基準，原版全畫面只由新情境冷啟動重生，不從磁碟狀態擷取。
來源獨立核對入口預定`tools/verify_dosgolem_king_text.py`及`work/issue4-king-text-source-audit.json`；
重製正常路線診斷入口為`work/issue4-king-text-remake-prototype.py`，不依DRAFT改正式獎勵順序。
工具期限1980秒、外層2040秒，沿用固定2f44a68執行器、原始資料唯讀、UID/GID1000及一次性Docker。
原始檔身份、seed、舊來源索引與IDA9.4證據沿上節；國王交易、畫面及存檔仍待有限READY審查。

乾淨e56a8e3的正常100次InputState基準先在9,8、無物品／金錢，再正常上鍵到9,7。
首個自然文字等待已得六件物品／50金並clear17h／set18h；九次Enter仍停同頁。
正式Poll只把Space映成Confirm，Enter僅設Enter；共用對話分支只接受Confirm或有prelude的Enter。
基準是實際Enter旗標，沒有用Confirm冒充鍵盤Enter。
私有`presented_deferred`試作只在乾淨副本沿用已解的404框／21414保留文字原語、既有陰影，
將獎勵移至文字EOF；DRAFT診斷暫用舊Go值，不能進production或commit。
原始window file19FAE的十個word為010B／0013／00EE／002C／0060／0194／0000／0000／0000／0000，
與已審查352×96、record404字模框相同。國王動態畫面與場景綁定仍待本次冷啟動閉合。

隔離`presented_deferred`試作從相同正常路線抵達9,7，第一頁至第九次等待維持無物品／金錢及17h旗標。
九次實際Enter可逐次續行，最後EOF才取得六件物品／50金及旗標切換。
結果只證明試作的正常路徑；原版本次來源及九頁完整PNG尚待核對，不提升為正式修正。
逐頁獨立比較入口為`work/issue4-king-text-compare.py`及同前綴JSON，不裁切或遮罩完整畫面。

#### 國王文字與延後交易（有限 READY）

本次冷啟動原版收據SHA-256為`b875807a2ae53cbdb9c02e43952e9e894e9230f76d64c469ffbea36c08b02f32`。
獨立來源核對通過100次正常輸入、200次IRQ1、435個產物，先前384份PNG／bin逐byte保持。
九次原生等待都在9,7；21414以DI0C06讀文字，九次Enter分別返回216C3等待。
10268文字返回時actor物品欄、金錢及旗標仍等於入口。
六次grant返回依序新增00、01、01、03、1F、1F，之後102B1金錢為0032，102B7 clear17h、102BD set18h，
102C3返回後1991D正常runner恢復。以上對明列EXE及此正常玩家路線為confirmed。
actor原始物品欄是+3A的八個word，原先801E出生裝備及最後00FF空槽保持，不把裝備混成背包。
呼叫端每次grant後直接設定下一物品，沒有依AX回傳跳過後續交易；滿欄失敗提示的動態畫面未驗。

乾淨e56a8e3基準九頁完整差6966／7390／7847／7394／7225／7641／7706／7242／7421，
視窗及陰影差6073／6497／6954／6501／6332／6748／6813／6349／6528。
隔離試作九頁完整均差893，視窗及陰影均0。完整圖保持，不以區域零差宣稱整張V3。

有限READY只新增共用有限`region_dialogue_reward_events`：scene、tile、所需／清除／設定旗標、
文字ID、既有presentation ID、shadow、依序物品與gold均由JSON提供。
事件在正式步行命中時開文字，禁止先給獎勵；九次等待接受實際Enter，EOF後才依序交易並恢復場景。
primitive不讀DQ3人物或座標常數，不執行JSON程式碼；未知引用、缺欄位、未審查證據拒絕。
`progress_flag_raw`只遷移現有重製存檔的0x200里程碑，另帶compatibility evidence，
不是原版旗標，不能用它證明原版進度；正式save格式及里程碑語意維持。
原版授予失敗不分支的caller沿用，未宣稱滿欄提示 parity；正式正常新遊戲具有充分空位。
schema升至0.8.0、content0.1.78；不同schema／hash存檔仍依現行契約拒絕。
驗收包括原始EXE／CTY／D3TXT parity、損壞契約拒絕、同條100輸入、九頁完整與視窗比較，
後續步行、一次性、正常Save／標題Load與同一Game Load清除pending，以及game／internal／desktop回歸。
正式入口預定`game/region_dialogue_reward.go`、`internal/gamepack/region_dialogue_reward.go`；
測試入口預定兩套件的`region_dialogue_reward_test.go`，欄位契約同步docs/84。

重建入口`work/issue4-make-king-text-pack.py`從乾淨e56a8e3資料包遷移九份JSON；
正式驗證入口`work/issue4-run-king-text-production.sh`在同一工具image與既有Go快取執行。
來源拒絕與收尾入口為`work/issue4-king-text-final-audit.py`及同前綴JSON。
正式`handler_raw`另核對實際CTY25圖格subid1指向原始56，文字逐word核對原始D3TXT01。

#### 國王正常謁見正式驗收（有限 CONFORMED）

正式schema0.8.0／content0.1.78，canonical hash
`sha256:82ca3334826bc909a65636e75a46895d4ee99b832cad1cd2cb0f430fd6029f2f`。
正式100次正常InputState至9,7，沒有位置、人物frame或旗標注入，沒有重新設定seed。
九頁自然等待接受實際Enter；各頁無物品／金錢交易，EOF後才依原版順序給六物品、50金及clear17h／set18h。
追加確認不重複發放，正常下鍵可到9,8；正式Save／標題Load完整快照保持。
從真正Save的謁見前狀態重新上鍵進入，再以同一Game Load取消pending，未完成文字不發獎勵。

正式九張完整PNG逐byte等於隔離試作；獨立解碼再與原版全640×350比較，完整均差893。
差異只在兩個NPC圖格，15,3為438、2,9為455；視窗及陰影均0。
因此只確認九頁文字窗、輸入與正常交易，整張V3、滿欄提示、原版存讀檔及音訊保持未知。
原版435產物及前批384份PNG／bin未改；十二種壞來源收據拒絕，含log／metadata／manifest同步改造的錯誤時序。
首輪負面核對發現text_return可被改成早於等待；verifier追加所有原始事件step嚴格遞增後，正常來源通過、反例拒絕。

全部game418頂層清單由21個有界程序覆蓋，378頂層／70子PASS、40選用SKIP。
全部internal149頂層／250子PASS、4選用SKIP及11套件通過，desktop為Linux x86_64 ELF，無素材缺失SKIP。
正常新遊戲至THE END75.37秒，只屬重製可玩回歸，不能提升原版campaign parity。
舊元件fixture停在標題及連按64次的舊假設造成失敗；訂正為實際保留文字等待。
自然EOF後測試額外Confirm誤開命令窗；helper在EOF直接返回後乾淨重跑，未改產品輸入規則。
容器Go快取及模組路徑一度不正確，改回既有`work/.gocache-test`／`.gopath-test`，未重建image。
相容里程碑來源分類由暫存user_report訂正為engine，最終包與全部測試重新建置驗證。

九份JSON從乾淨e56a8e3重建逐byte相同。既有`tools/ida_npc_animation_ledger.json`追加六筆原始定位，
保留前十筆與原名／bytes／位址，新增語意引用b875807a正常交易來源。
既有`tools/ida_dump_npc_animation.py`增加1024C..102C4有界範圍，自動附加分級語意。
本輪沿用先前IDA9.4原始匯出，新增分級以動態閉環審查，不宣稱重做database分析。
獨立收尾收據`work/issue4-king-text-final-receipt.json`為11805bytes，SHA-256
`9ec36ece96411c0229780666aaa4a14febaf13706dc49fea14f541e83de569d9`。
輸出UID/GID1000，既有root候選3213及Markdown目錄0保持；原版素材與圖像不加入Git。
下一原版切片從正常謁見完成checkpoint走回城鎮及酒館；人物動畫及音畫差異保持待驗。

#### 正常謁見後回程與登錄所入口（DRAFT）

接續0c47537及原版b875807a，正式schema0.8.0／content0.1.78保持。
本輪從正常100次謁見路線後追加步行，先核對王座9,22回程、城堡15,31出口及城鎮8,14登錄所入口。
`work/issue4-king-return-route.py`在乾淨提交副本重播正常100次InputState，無位置／旗標／frame注入，
取得離城及繞路方向；同前綴JSON／log只作原版輸入定位，不作原版oracle。
原版新情境`king_return`由`tools/dosgolem_newgame_probe.py`自行冷啟動重生，
前綴`work/dosgolem-opening/issue4-king-return-*`，保留前100次輸入及435份來源產物。
新增觀測只記錄正常runner、位置、原始scene欄位、camera、actor、gold及flags，不寫入原版狀態。
來源獨立核對入口為`tools/verify_dosgolem_king_return.py`及`work/issue4-king-return-source-audit.json`；
正式同條輸入診斷入口為`work/issue4-king-return-remake-probe.py`及同前綴收據／PNG。
獨立全畫布比較入口為`work/issue4-king-return-compare.py`；原始CTY幾何與圖層欄位保存在
`work/issue4-king-return-cty-raw.json`，只保留檔案偏移，不由數值推定未知語意。
既有IDA9.4匯出的原始camera／界外loader／consumer由
`work/issue4-king-return-static-audit.py`逐指令核對EXE，結果為同前綴JSON。
原有註記與限定範圍保持，新增場景在正常原版閉合前只列strong，不由共用renderer外推confirmed。
乾淨0c47537副本的185次正常輸入、87份完整PNG及狀態收據已產出；標題讀檔保留
位置、故事旗標、50金及六件物品。該結果不提升原版證據等級，原版回程仍在重播。
同一可丟棄探針的`DQ3_KING_RETURN_PROTOTYPE=camera|layers|both`分支只在隔離Go副本
測試原始欄位衍生的候選，不寫正式JSON，也不將DRAFT資料假標為已審查D3。
各分支的同條輸入及全畫布比較使用不同前綴；原始CTY與已有IDA匯出保持唯讀。
損壞收據核對入口為`work/issue4-king-return-source-negative.py`，只在容器tmp修改副本，
同步重算日誌manifest後檢查語意拒絕；結果為同前綴JSON，不改原始來源。
來源與畫面尚待閉合，不從路線規劃推定已抵達，也不據DRAFT修改正式資料或規則。

固定20,000,000指令間隔在下層城堡及城鎮可能跨過未完成繪圖，IRQ1已送達不代表原版
已處理該次步行。本批先保留固定時程結果，不將異步位置直接判成remake規則缺陷。
`tools/dosgolem_king_return_ready.py`改由原版100次正常謁見的既有checkpoint續行，
前綴`work/dosgolem-opening/issue4-return-ready-*`，先核對36次正常下鍵至離城。
每次只在1991D原版runner返回、按下及放開均送達、鍵盤佇列清空後續送下一鍵。
這是驗證fixture修正，正式引擎及pack仍保持。parent收據b875807a、checkpoint hash、
實際初始記憶體與虛擬時鐘另存；不重新固定seed，也不宣稱前100次重新冷啟動。
dosgolem原生磁碟snapshot v2保存CPU／memory／PIT／VGA／DOS，未保存鍵盤佇列及IRQ統計。
本續行限定parent全部200個邊緣已送達且無後續pending鍵的checkpoint；新的IRQ debug count從0記錄。
原生實作入口為dosgolem `internal/state/state.go`及`internal/machine/state.go`，不是記憶體patch。

追加勘誤：上述磁碟snapshot續行未通過，初始PIT觀測變成65536，轉場後停在自然等待窗。
失敗來源保存在`work/dosgolem-opening/issue4-return-ready-generation.py`、同前綴Go／log與PNG，
不提升為正式oracle。`tools/dosgolem_king_return_ready.py`現改回冷啟動，使用獨立
`work/dosgolem-opening/issue4-return-cold-ready-*`前綴；關閉逐指令文字trace，仍保留全部事件、
先前100次輸入與既有畫面，85次步行另逐鍵觀測。原版自然等待窗由正常Enter關閉，額外輸入全部記錄。
送步行鍵只在正常runner且等待計數自然重設後進行，不寫計數器、不強制影格、不重新設seed。

另一勘誤：固定時程收據的`origin_y`標籤讀的是1991D時的DGROUP4F27。
該欄位在11D8A繪圖consumer執行後已是暫態掃描座標，不能當攝影機原點。
既有原始收據及凍結Go來源保持；新探針改存raw4F25／raw4F27，另在11991呼叫11D8A前
記錄真正的origin與20×15幾何。新增camera結論只採該呼叫點，不從舊標籤推定。

現行冷啟動逐鍵來源使用`issue4-return-cold-ready-r2`前綴。
來源核對入口為`tools/verify_dosgolem_return_ready.py`，逐筆核對前100次輸入、200次IRQ1、
432份既有PNG／bin、85次步行及自然等待窗的額外Enter，不將額外輸入省略為185次。
攝影機只採11991呼叫前的觀測；1991D保留raw4F25／raw4F27。
第一版冷啟動逐鍵探針因每指令多餘觀測耗時而停止，僅保留診斷，不作oracle。
R2限制新欄位讀取在實際觀測點，並在既有100次謁見結束後停止舊觀測器。
等價正常InputState入口為`work/issue4-return-ready-remake-probe.py`，按來源實際packet順序重播。
每個原版等待窗在remake自然等待，缺窗即記第一個玩家blocker；不省略額外Enter。
該工具的`DQ3_KING_RETURN_PROTOTYPE=both`另在可丟棄測試副本加入CTY00 section0等待窗範圍，
連同既有camera／layers候選診斷完整路線；此分支的pack已變更，不將其hash稱為正式資料包hash。
逐鍵來源的協同損壞拒絕入口為`work/issue4-return-ready-source-negative.py`，
另核對繪圖前anchor、觀測階段及等待窗caller；原始檔案唯讀。
完整640×350比較入口為`work/issue4-return-ready-compare.py`，比對來源11991的攝影機原點，
不使用舊origin標籤；狀態、完整RGB與尚未閉合的人物時序分別保存。

R2追加勘誤：等待鍵在linear2111B入口立即送出，make發生於入口後2步。
既有IDA9.4原始匯出與EXE的file_bytes閉合2112C..21132：函式隨後以CLI清除DGROUP2856，
才STI並在21133輪詢。因此入口送鍵會被初始化清除，R2未產出有效回程oracle。
R3使用獨立`issue4-return-cold-ready-r3`前綴冷啟動，僅在21133、key_flag=0及
SS:SP+6／+8保留7E00:0110的等待caller時送Enter。
回程等待計數接受原生19955／19976清零後的start=0，或自然初始化後delta<=1，不寫計數器。
另修正收據生成器對第100次後輸入的deadline：後段沒有固定擷取時程，以完整make／break
在queued+11,000,000內送完核對；既有100次的deadline與畫面保持。
目前來源與remake仍DRAFT，以上是探針修正，不提升正式產品完成度。
remake探針把來源`return_events`與queued事件按原始step排序；觀測不計為玩家按鍵。
若正常runner先返回、稍後才開等待窗，步行PNG在等待前保存，不把關窗後畫面換作該步收據。

R3追加勘誤：21133送Enter可正常關閉等待窗，但start=0或delta<=1的探針條件過嚴，
只反覆關窗，沒有送出步行；不能將這個條件當原版ready契約。
現行R4以`issue4-return-cold-ready-r4`冷啟動，保留21133的初始化檢查，
步行只在原生1997C的RETN前、佇列清空且前鍵make／break全送達後排入。
1991D..1997C的既有IDA9.4匯出已與EXE file_bytes核對，包含300門檻、關窗與返回分支。
由原版完成等待檢查後再送鍵，不以探針重設計數器或猜測重設時點。
85次正常runner觀測仍在1991D；來源及產品仍須獨立核對才提升證據等級。
正式正常輸入測試入口為`game/king_return_test.go`的`TestDosgolemKingReturnNormalInput`，
需要`DQ3_KING_RETURN_ORIGINAL`及既有母親／王座收據環境；未提供時明示選用SKIP。
測試按來源排序85次步行、額外Enter及觀測，核對位置、raw section、camera、背包、金錢與旗標。
等待窗正文與完整畫布分別核對，另保存陰影差異；標題讀檔後再正常下一步，不宣稱原版存檔parity。

R4原版執行完成：正常186次輸入、372次IRQ1、85次runner及607份產物，末端8,2／raw111D。
收據首次核對因要求85次11991觀測而拒絕。原版一般步行走增量繪圖，11991完整重繪實際只有
15、36、74、85四次。此為verifier條件錯誤，沒有重跑或改寫來源；未觀測的camera欄位改保留null。
四次原生camera單獨核對，其他步只檢查既有EXE anchor契約，不稱原生camera量測。
同條輸入的未修改0c47537在ordinal74、CTY00 section0的5,22缺少自然等待窗，診斷已保存。
`both`候選186次正常輸入及標題讀檔PASS；`idle`候選只加入等待窗範圍，
用來隔離camera／layers的完整畫面差異，兩者都保持可丟棄設定，不作正式hash聲明。
`both`的85個位置／section／金錢／背包／旗標與四次camera均通過，但ordinal36仍13983完整RGB差異，
多個圖格反覆149像素，不能稱只有NPC差異。`camera`候選保留等待窗修正，只改登錄所攝影機，
用來與`idle`及`both`分開比較。城鎮layers候選未達READY，不進正式資料。

#### 回程場景證據審查（有限 READY）

來源收據SHA-256為`347600d38f062e111a08a852a0cdf2a8937b959827c37ff984f3dca44a732a75`。
原版186次輸入／372次IRQ1、607產物、85次正常runner，前432份PNG／bin逐byte保持；
16種協同損壞均拒絕。EXE、CTY00與位址基準沿上節的原始hash／file_bytes稽核。
四次11991觀測在15、36、74、85，分別包含下樓、離城、等待恢復及登錄所完整重繪。
其餘增量步行不假稱camera量測。兩側Lv1 seed1357只設定一次，後續NPC骰序不列精確一致閘門。

| 正式資料修改 | 原版入口及consumer | 已證實的有限範圍 |
|---|---|---|
| CTY00 section0等待窗 | ordinal74的5,22／raw000C；19952→17DBB→21133，正常Enter關窗 | 既有單人正文、300 tick門檻及不移動／不消耗；188 updates仍為hardware-spec approximation |
| CTY00 section0圖層 | 原始header file000C+15／+16為27／70；13162..1317D→0B56／0B57→11E07..11E4B | 既有player_cell_layer，base_layer0；正常回程包含非base的酒館內部視野 |
| CTY00 section0camera引用 | 36與74原生11991、anchor9／7、header界外0 | 與既有arrival_camera同值，提供layers validator的明示引用；renderer優先序保持 |
| CTY00 section2camera | 85的8,2／raw111D；11971..11991，CTY file111D+12為71 | 原生origin−1／−5、20×15；不clamp到11×15地圖邊界 |

未修改0c47537在ordinal74缺等待窗。只改等待窗的`idle`原型全程狀態通過，
ordinal74完整RGB差146268、ordinal85差200467。
`camera`只另修登錄所，74維持146268、85降至118；`both`再加入城鎮layers，
74降至1403、85保持118。85個原版位置／raw section／金錢／背包／旗標及四次原生camera均一致。
85個畫面沒有任何一張因layers候選而惡化。先前36的13983在所有分支相同，不能用它否定layers修正。

READY只修改既有JSON綁定，schema0.8.0維持、content升0.1.79，不新增共用Go版本資料或fallback。
camera／layers原始EXE／CTY parity、缺失／未知引用拒絕、正式186次InputState、正常標題讀檔與下一步，
以及完整game／internal／desktop需通過才列CONFORMED。乾淨0c47537重建入口為
`work/issue4-make-return-ready-pack.py`，九份JSON逐byte核對；不加入原版素材或私有PNG。

完整640×350仍不通過。目視74可見remake在黑色外層多畫三個NPC，另有NPC／主角相位差異；
85剩118只在主角圖格。36的13983跨多個圖格，原因保留未知。
本批限定圖塊、攝影機、等待窗與正常玩家狀態，NPC圖層可見性、動畫、原版音訊及原版存檔另追，
不由局部改善宣稱完整畫面V3或整款遊戲完成。
完整game回歸入口為`work/issue4-return-ready-game-regression.py`，以現行binary列出測試，
分批隔離Ebitengine程序，核對全部頂層終態與SKIP理由；原版明示收據測試另行嚴格執行。
導航診斷入口為`work/issue4-return-ready-navigation-diagnostic.py`；僅在乾淨副本追加失敗觀測，
不修改玩家狀態。收尾入口為`work/issue4-return-ready-final-audit.py`，輸出同前綴`final-receipt.json`，
核對來源、九份JSON乾淨重建、正式PNG、完整回歸、桌面ELF與UID／GID衛生。

#### 回程正式驗收（有限 CONFORMED）

schema0.8.0／content0.1.79，canonical hash為
`sha256:9757fa4135987cec259a1075e5ea12fe9270703c9bda471ff1fbca79bba81790`。
正常186次InputState、85個位置／raw section／金錢／背包／旗標及四次原生camera通過。
CTY00 section0等待窗自然開啟，正文及陰影RGB差0，正常Enter只關窗。
同版本Save、正常標題Load及後續下鍵移動通過；原版存檔及RNG持久化不由此宣稱parity。
正式收據為`work/issue4-return-ready-production.json`，log同前綴；88張runtime PNG包含
85次步行、等待、標題讀檔及後續一步。85張步行PNG與`both`試作逐byte相同，未遮罩或指定frame。
完整RGB仍為36的13983、74的1403、85的118，等待完整1095；完整畫面V3未通過。
酒館多畫的NPC、人物相位及城鎮入口多格差異保留未知，不能以圖塊規則自行補NPC filter。

完整game419頂層清單由21批覆蓋，374頂層／70子PASS、45選用SKIP；
本批原版回程明示來源測試另嚴格PASS。其餘選用來源或擷取未跑，沒有素材缺失SKIP。
internal149頂層／250子及11套件PASS、4選用SKIP；desktop Linux x86_64 ELF建置通過。
正常新遊戲InputState至THE END151.74秒，只屬remake回歸，無完整原版campaign聲明。

第一次game第14批卡在5,28，NPC暫擋0,29。只加觀測的乾淨診斷副本確認
`fieldIdle.open=true`，NPC停在相同動畫計數305；測試連續送空白輸入，無法關窗。
共用導航test helper改用正式Enter關窗後等待，不改正式凍結規則、NPC位置或計數。
原始失敗為`issue4-return-ready-game-batch-14-initial-0ca91015db05.jsonl`，診斷log同入口前綴。
同工具鏈重跑14及尚未執行的15..21通過，前13批同一1.79資料包PASS保持；resume log單獨保存。
桌面build後`file`指令缺件是檢查工具問題，改核對ELF header及machine62，沒有重建已成功binary。

九份JSON從乾淨0c47537重建逐byte相同，只改manifest與interface三個集合。
其他欄位排版保持，縮減diff前後的解碼JSON相同，canonical hash與既有PASS不變。
最終收尾收據`work/issue4-return-ready-final-receipt.json`，112587bytes，SHA-256
`3393c711ed6d9549b28613bf14ed97a205ba7681f74d0116cfb8a556ffc66c05`。
純排版前收據按8ae2579ddd60前綴保留；一次收尾掃描遇執行環境`.aws`暫時目錄消失，
只對該控制目錄容許消失，再以同命令乾淨重跑，原始輸入仍逐檔驗證，未忽略產品缺檔。
所有本批輸出UID／GID1000，既有root候選3213與Markdown目錄0保持。
遠端結果[5965833865](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5965833865)，
Issue #4保持OPEN，下一切片從正常登錄所入口續行；沒有新image或發行包。
Issue主文更新入口為`work/issue4-return-ready-update-body.py`，提交後Git與Docker清理核對
另存`work/issue4-return-ready-post-push-receipt.json`。原始素材、私有PNG及使用者scratch不加入Git。
正常原版重生入口為`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --king-return-ready-original`。
外層移除一次性「前綴已存在」限制，沿用共用產生器逐檔SHA-256歸檔，再重生的既有契約；
此修正不改本批已執行的frozen generation／Go，也沒有再次冷啟動原版。

#### 登錄所正常交談（DRAFT）

接續562208a及Issue #4，從CTY00 section2正常8,2接近原始NPC2,3。
舊docs/36只提供定位；首次對話、輸入owner、選單、命名、性別與登錄交易尚缺dosgolem動態閉環。
本切片不改正式引擎、pack或存檔。原版冷啟動保留先前186次正常輸入，額外方向及Enter均記錄。
新原版入口為`work/issue4-registry-entry-wrapper-draft.py`，現行私有前綴為`work/dosgolem-opening/issue4-registry-entry-r2`。
正常重製診斷入口為`work/issue4-registry-entry-remake-probe.py`，沿正式回程後正常InputState接近交談。
原版首次等待與重製入口是否一致，需各自完整畫布與狀態核對，不從內部modal存在判定parity。
生成器語法檢查入口為`work/issue4-registry-entry-generation-audit.py`，只展開來源、不執行原版。
直線六Left只到5,2，直接選2,4亦因櫃台不可走而失敗；R1原版探針已停止，未產出有效oracle。
前兩條重製診斷及初次Go換行生成錯誤均保留各自source／log，不列為產品缺陷。
重查spec閘門後沿正式traceTalkNPC候選，含中間為阻擋櫃台的兩格交談位置，正常到2,5面向上。
R3重製的13次正常輸入路線為Left×3、Down×4、Left×2、Up、Left、Up、Enter；
最後Up只面向櫃台，不移位。單次Enter在remake沒有開command／dialogue／tavern，是否不同須原版證據仲裁。
R2原版採此已定位路線冷啟動，保留原先186次輸入；source尚待核對，正式0.1.79保持。
獨立來源稽核入口為`work/issue4-registry-entry-validator-draft.py`，核對先前186次輸入、372次IRQ1、
604份PNG／bin及新增逐鍵事件，拒絕未對應事件的圖像；首次wait只記raw caller，不猜功能。
R3重製JSON沿用直線版本的`+8`計數，誤報194；實際13筆新增InputState合計199。
凍結Go及原始JSON保持，追加稽核記錄修正，不把元資料修正當成產品修改。
原版收據scope文字亦沿用回程說明，新增接近範圍由registry事件與獨立稽核限定。
原版鍵盤及登錄入口的有界IDA候選匯出為`work/issue4-registry-input-ida.py`及同前綴JSON，
沿用唯讀原始EXE與一次性database；未動態閉合的候選維持unknown。
鍵表、Enter／Space入口及後續consumer分別為`work/issue4-registry-key-table-ida.py`、
`issue4-registry-talk-entry-ida.py`與`issue4-registry-interact-consumer-ida.py`，各有同前綴JSON。
最小IDA探針為`work/issue4-registry-ida-minimal.py`，已核對9.4、原版hash及828個函式。
可丟棄Enter／Talk診斷為`work/issue4-registry-entry-remake-r4-prototype.py`，
只在乾淨副本讓Enter呼叫既有交談，沒有實作原版greeting，不作READY或正式修正。
完整狀態／RGB比較入口為`work/issue4-registry-entry-compare.py`，R3計數勘誤為
`work/issue4-registry-entry-remake-r3-count-audit.json`；完整PNG不裁切或遮罩。
原版重生的Docker控制入口為
`bash tools/verify_dosgolem_newgame.sh /tmp/dq3-dosgolem-2f44a68 --registry-entry-original`；
獨立拒絕稽核為`work/issue4-registry-entry-source-negative.py`，在容器/tmp建立私有鏡像，
協同修改待驗JSON、log與manifest，原版輸入保持唯讀。尚未有R2收據時不宣稱通過。
有界IDA原始bytes彙總為`work/issue4-registry-entry-static-audit.json`；
Enter是17C43、Space是17C83。17C43→18966→14D56→147BD使用原始1F7／726閘門，
14E0E先試正前方、櫃台再試兩格。此靜態鏈目前strong，不擴成全場景Enter的已確認規格。
R4正常199次InputState、13筆接近及Enter／Talk診斷PASS36.60秒，最後tavern_active=true／stage0。
它略過原版首次greeting的可能性仍待R2完整畫面仲裁；沒有因原型能跑就改正式程式。

##### 2026-10-03追加訂正及現行 R7 探針

上節R2執行中及R4八職業取消假設已過期，原始輸出保留。現行原版前綴為
`work/dosgolem-opening/issue4-registry-entry-r7`，由同一Docker控制入口重生。
R2在回程ordinal62耗盡繼承1980秒限制；R3缺二選一輪詢，R4漏問候內嵌等待，
R5漏姓名輸入owner，因此逐次停止。R6核對鍵盤ISR後確認Escape無法取消姓名，亦已停止。
這些前綴沒有有效登錄收據，不列為產品缺陷或原版parity。
R7只延長有界工具時限並修正觀測及正常按鍵，不改原版記憶體或前186次輸入。

原版身份仍為DQ3.EXE 115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
有界IDA9.4匯出沿用原始名稱、file bytes及linear位址；file=linear−EC90，DGROUP基底linear24DD0。
以下為strong靜態鏈，動態範圍由R7實際收據界定，不把取消路線當成創角交易驗收。

| 原始定位 | 消費端與有限語意 |
|---|---|
| linear106DF..106E7 | record550問候後呼叫1F63C二選一；原始TXT含兩次FFFC |
| linear10763..1076B | record554後呼叫10D17姓名輸入，先於職業 |
| linear10789..107A5 | record555／DGROUP3DEC=6，六職業選單，非八職業 |
| linear107BD..107C1 | 職業返回後才呼叫性別選單 |
| linear11096..11170 | 姓名輪詢210CA；方向、Enter／Space及64h滑鼠取消，沒有01h Escape分支 |
| linear2109F..210B9 | IRQ1 make原樣寫2856、break清零，沒有Escape轉64h |
| linear10DC8..10E4F | 五項功能列，選第四項寫726=1，返回10D9C，再到1081F取消文字 |
| linear10832..1083F | 登錄別人二選一只檢查722，選第二項後播560道別；Escape不等價於否定 |

RE入口：`work/issue4-registry-choice-wait-ida.py`、`issue4-registry-name-poll-ida.py`、
`issue4-registry-cancel-key-ida.py`、`issue4-registry-keyboard-irq-ida.py`，各有同前綴JSON／log。
每份匯出包含輸入hash、IDA9.4與位址基準；未閉合候選仍標unknown，解釋以本表限定。
歷史docs/36追加勘誤，舊「職業→姓名→性別」不作production規格。

R7接受問候後用Up、Left、Enter由raw0進功能列，再Down×3及Enter取消姓名。
取消文字後在二選一用Right明確選第二項，再Enter，最後關閉道別回到原生1991D。
新增按鍵只在原生21133／1F7B7／11096輪詢送出，上限32次；未知owner立即停止。
姓名觀測保留26FC mode、26FE cursor、270A長度及2710原始buffer，不把SI或stack猜為通用caller。
來源verifier須核對完整返回、所有前604份PNG／bin與IRQ序列，未完成返回即拒絕。

問候首頁可丟棄試作為`work/issue4-registry-entry-remake-r6-prototype.py`，
正常199次InputState及末端問候等待PASS19.46秒。只借用已審查共享文字原語與視窗陰影，
未實作二選一、姓名取消、後續職業、能力預覽與名冊交易，尚未原版RGB比較。
試作及R4 Enter／Talk均只留私有副本，不進production，不以能跑代替READY。
遠端續行記錄為[5966631605](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5966631605)，
其R6啟動狀態以上述R7與後續Issue留言為準。正式schema0.8.0／content0.1.79維持。

追加直接證據索引：`work/issue4-registry-window-reader-ida.py`及
`work/issue4-registry-window-fields-ida.py`匯出連續record的cursor／EOF、SI4080二選一及姓名視窗，
各有同前綴JSON／log。20770是既有音訊cue路徑，只保留入口定位，不深入硬體時序。
原始bytes獨立查核入口為`work/issue4-registry-entry-static-audit-r7.py`及同前綴JSON，
10份IDA9.4匯出、4680列有界指令、七筆strong附加語意，全部file bytes與原始EXE一致。
這只查核靜態鏈，尚未宣稱登錄正常路徑動態通過。

完整姓名取消的可丟棄試作入口為`work/issue4-registry-entry-remake-r8-prototype.py`。
沿正常186次回程後再送27次方向／Enter，沒有玩家狀態注入。R8在二選一繪圖發現
共享文字RGB243不在新遊戲專用lavender色盤，失敗source／Go／log保留。
R9入口為`work/issue4-registry-entry-remake-r9-prototype.py`，用目前場景色盤及已確認font index，
把文字前景由共享presentation提供，不用最近色號近似。此候選仍需原版完整畫布仲裁。
R9正常213次InputState、27筆登錄觀測PASS11.18秒，末端2,5面向上，無modal、名冊0，
金錢50、背包六件與裝備保持，取消過程RNG保持356D。此處只證明試作內部自洽。
所有版本專屬record與NPC值只存在可丟棄副本，不加入production或commit。
R7按鍵訂正另記於[5966709923](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5966709923)。
R10正常標題讀檔抽樣入口為`work/issue4-registry-entry-remake-r10-prototype.py`，
213次InputState取消、三次標題讀檔及正常下一步PASS60.08秒，名冊／背包／金錢／裝備與旗標保持。
私有試作稽核為`work/issue4-registry-entry-prototype-audit.py`及同前綴JSON，
綁定report、完整test／試作Go、binary、log與全部PNG。R9與R10的27張登錄PNG逐byte相同，
前12張接近PNG與未修改562208a診斷保持。稽核首次漏計回程等待圖一張，訂正數量後重跑；
沒有重生或改寫任何已執行試作。姓名完成、職業、性別、能力預覽及名冊交易仍未實作。

##### R7等待訂正與現行 R8

R8執行期間，登錄出生交易的有界靜態查核入口為
`work/issue4-registry-birth-ida.py`、同前綴JSON及log。
只追原始linear106DC..10D17的姓名、六職業、性別、能力預覽與登錄writer，
並保存DGROUP3DEC、3DF6、4F54及相關暫存欄位的原始bytes與xref type。
這項靜態匯出不能替代正常出生輸入或原版能力交易收據，推論等級保持strong或unknown。
獨立查核入口為`work/issue4-registry-birth-static-audit.py`及同前綴JSON，
對照全部539列原始file bytes、腳本與工具版本，再附加六職業索引及97-byte登錄交易的strong註記。
職業選項1..6讀DGROUP4F54的位元組1..6，得到raw class 1、2、3、4、6、7。
linear10924以DGROUP520B的暫存角色呼叫1D9CC及1DB5F；linear10A9F在確認後
將97 bytes複製到DGROUP530D＋97×slot，linear10816再寫4F64＋slot為1。
取消分支在1081F，沒有通過此writer。這些是靜態控制流證據，尚未宣稱出生畫面或能力亂數對拍。
後續出生探針及預檢入口為`work/issue4-registry-birth-probe.py`與
`work/issue4-registry-birth-generation-audit.py`。只在R8來源接受後啟動，
延續正常取消返回再進櫃檯，以一個英數字、第一職業與第一性別走能力預覽及登錄。
出生的測試seed1357在執行前固定，僅於自然1D9CC入口、SI520B及caller0960成立時設定一次。
原版與remake各自記錄此條件，不由全段骰序一致宣稱所有亂數parity。
`work/issue4-registry-birth-precompile.py`在容器/tmp使用獨立輸出目錄，
只建置生成的Go並在執行原版前停止；保存同前綴Go、JSON及log，不更動R8產物。

R7已正常抵達登錄所，十二次接近觀測後的Enter實際呼叫21414／record550，
位置2,5／raw111D，make／break為397／398。未修改remake同一第199次Enter無modal，
因此入口行為已有反證。R7未捕捉問候畫面或完成取消，不能當完整登錄收據。
探針誤把FFFC等待當作最後關窗21133。重新查閱規格閘門路由及既有IDA9.4直接證據後，
確認216C3自有輪詢：216D8→21148為可見箭頭，216FB→21148為隱藏箭頭，最後21133另屬一般等待。
R7已停止，原始Go、生成器、log與部分PNG保留；沒有將缺少事件的圖像補登成dosgolem收據。

現行前綴改為`work/dosgolem-opening/issue4-registry-entry-r8`，仍用上節同一Docker控制入口。
只新增原生216D8可見等待觀測與送鍵，保留21133／1F7B7／11096三種owner，
沒有將音訊或PIT driver重做為遊戲規格。既有原版語義來源為
`work/issue4-king-audience-text-ida.json`，工具9.4，輸入與本節相同EXE hash；
216C3..21710共31列file bytes已獨立對照原始檔一致。R8冷啟動已開始，仍待完整來源稽核。

R8若走到既定90億指令停止點，`work/issue4-registry-r8-state-audit.go`直接解碼
自行保存的final.state，輸出同前綴JSON，供CPU位置與原始記憶體診斷。
不呼叫LoadState、不補CRTC、不重畫原版圖片，也不把狀態檔當完整取消或畫面收據。
R8已自然跑滿90億指令，原生程式exit0，wrapper因沒有REGISTRY_DONE拒絕接受。
末尾狀態200117 bytes，SHA-256為428a49775779d6f4328913bd14cfaab961f1efa5d66e73a331fc511215d4dd6b，
直接解碼的CPU為CS:IP122B:08E4／IDA linear21A94、DS15ED、IF開、PIT12428、Ticks71310。
沒有還原或補畫此狀態。重查規格閘門路由後，必要有界IDA定位入口為
`work/issue4-registry-r8-blocker-ida.py`及同前綴JSON／log，範圍21A80..21AF0與直接xref。
取消及出生維持DRAFT，待此位置與caller閉合後才重跑。
21A94的原始指令為`7ef6`，讀DS:0005的有號比較未大於0時返回21A8C。
`work/issue4-registry-r8-blocker-flow-ida.py`及同前綴JSON／log只補21A01..21AA7
與現有INT1C向量所指1FE53..1FEB0的軟體counter writer，保留先前有界匯出。
此處追GUI能否續頁所需的counter，不研究DAC／PIT逐週期或音訊driver。
`work/issue4-registry-r8-counter-state-audit.go`及同前綴JSON以同一未還原狀態補出
DGROUP0000..001F、word0005與機器IRQ欄位，全部保存raw值供定位。
計時回呼旗標及舊向量的補充診斷入口為`work/issue4-registry-r8-callback-state-audit.go`及同前綴JSON。
只讀診斷另以原R8 binary從此state續行50萬指令，入口log為
`work/issue4-registry-r8-diagnostic-resume.log`，末尾state為同前綴final.state。
不送鍵、不重設seed、不輸出PNG；僅觀測IRQ0／INT1C與counter。
已知LoadState未保存CRTC且會重新計算IRQ0排程，這次續行只作執行器診斷，不能當同狀態正式對拍。
相同起點的writer定位log為`work/issue4-registry-r8-diagnostic-writer.log`，
只觀測1FE5F、1FE64、1FE87、1FE8B及1FEAC的暫存器，不改條件或產生畫面。
診斷續行確認IRQ0與原INT1C writer均到達，DS15ED；1FE87／1FE8B的CX=1，
1FEAC的CX=2。尚不能由一次snapshot的word0005=0宣稱計時器沒有更新。
新增最小callee匯出入口`work/issue4-registry-r8-callback-ida.py`及同前綴JSON／log，
只核對1FEBA..1FF10的暫存器保護與GUI callback，不延伸音訊硬體細節。
回呼1FF02明確寫CL=2；尚待其返回及色盤更新條件閉合，不能因此宣稱原版bug。
必要返回範圍與色盤呼叫的匯出入口為`work/issue4-registry-r8-callback-return-ida.py`及同前綴JSON／log。
訂正：1FF02..1FF27是七個CS原始word的左移，尚未證實為色盤更新。
保存的DGROUP0013為4002，1FED4的4000測試會直接進CL=2分支，counter是否循環不能解除此旗標。
旗標初始化的最小xref及1FF4D caller匯出入口為`work/issue4-registry-r8-flags-ida.py`及同前綴JSON／log。
原始word與未知語意保持，未使用改旗標、改CX或重設counter的方法製造正常收據。
原xref只列出IDA已解析為DGROUP的直接引用。另以
`work/issue4-registry-r8-flag-operands-ida.py`及同前綴JSON／log保留原始DS:13h的operand候選，
附xtype及DS register，不把候選當完整writer清單或推測的欄位名稱。
必要初始化writer匯出入口為`work/issue4-registry-r8-flag-writers-ida.py`及同前綴JSON／log，
限定21EC1..21F60、2206E..220B0及16F29..16F80。來源旗標語意尚未知，不稱防拷或色盤欄位。
初始化caller與實際輪詢邊界的追加匯出入口為`work/issue4-registry-r8-flag-entry-ida.py`
及同前綴JSON／log，限定21DDC、192ED、1351F與1FF02的必要範圍。
既有前段snapshot的零續行比較入口為`work/issue4-registry-r8-prefix-state-audit.go`。
只解碼原始CPU與DGROUP0007／0013，不還原機器、不產生PNG。
較短正常路線的隔離重製診斷入口為`work/issue4-registry-short-route-remake.py`
及同前綴private-test.go／test／log。從正常母親返回續行40次步行至登錄所，
原始CTY通路只用來定位候選，未取得原版動態收據前不宣稱此路線對拍通過。
原版較短來源的生成入口為`work/issue4-registry-short-probe.py`，展開至
`work/issue4-registry-short-checked-generation.py`，稽核為同前綴generation-audit.json。
保留原版38次母親返回，再以40次正常方向輸入進登錄所，取消後重新登錄。
旗標4000出現即拒絕來源；不改時鐘、CX或原版旗標，不外推先謁見路線已通過。
較短來源的獨立預編譯入口為`work/issue4-registry-short-precompile.py`與同前綴Go／JSON／log，
使用/tmp輸出，成功建置後即停止，正式原版尚未執行。
相同較短路線的取消候選為`work/issue4-registry-short-cancel-prototype.py`及同前綴產物。
隔離副本沿用R10私有原型，以正常輸入、完整PNG及標題存讀檔抽樣；尚未進正式Go或JSON。
追加勘誤：1991D是閒置檢查，1997C是其返回；實際場景讀鍵為19417→210CA。
先前ready標籤只表示探針的觀測邊界，不是GetKey函式名稱。既有IRQ與玩家位置證據保持。
零續行前段snapshot確認：生日1440000000時word0013=0000；母親返回、城堡、王座與
謁見末尾為0002；歷史回程5024000000及R8末尾為4002。各輸入狀態檔的SHA-256保留於
prefix-state-audit.log。原始1FEDC在word0007≥20000時測試低位是否為3或0Ch，
兩者均不符便於1FEFC設4000；1FF02寫CL=2，七個CS原始word再左移。
此鏈足以定位捲動受阻，不足以命名旗標的完整用途或宣稱同硬體時鐘parity。
較短路線的40次正常InputState診斷PASS2.30秒，經母親返回21,17，抵達登錄所8,2，
金錢及背包保持0。預編譯PASS；首次錯掛/source造成來源缺失，改正為既有/dosgolem後
同工具鏈重跑通過，原版未執行。現已啟動較短原版冷啟動，前綴為
`work/dosgolem-opening/issue4-registry-short-r1`，僅以正式IRQ1及原生邊界續行。
遠端工作登記見[5968196927](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5968196927)。
較短路線原版已停止，40個觀測與13次後段輸入均有原生產物，但未進登錄所。
第一個分歧為母親返回21,17後的正常下鍵：原版仍21,17，候選remake為21,18。
沒有record550、取消或出生；producer的DONE只表示探針結束，不等於來源合格。
不以較短路線宣稱先謁見gate，原因尚待writer／consumer閉合。
重查規格閘門路由後，必要移動定位入口為`work/issue4-registry-short-route-blocker-ida.py`
及同前綴JSON／log，限定1255B、11A6A與119B8的必要範圍。
第一個南側格的CTY00 raw word為0101；1F低位subid=1經196D2選取section的handler byte55，
DGROUP3BB4的第55個pointer為020B，原始定位為IDA linear1020B。
必要gate匯出入口為`work/issue4-mother-return-movement-gate-ida.py`及同前綴JSON／log，
限定1020B..1024C及196D2..1970E，尚未由handler名稱推定旗標規則。
追加靜態證據：1020B讀story flag17h，為1時開3E6E文字窗，呼叫文字consumer，
關窗後以4F1F=2呼叫194C3強制往北，最後清0B34。原始bytes及位址保持。
文字來源、顯示時序與完整副作用仍待自然動態閉合，本段為DRAFT。
窄任務入口為`work/issue4-mother-return-gate-probe.py`，產物前綴為
`work/dosgolem-opening/issue4-mother-return-gate-r2`。保留38次正常母親返回輸入，
只增加一次Down；觀察handler入口、文字完成、關窗、強制北行與返回，
不載入snapshot，不改座標、旗標或時鐘。生成、執行及稽核保持分開。
R1工具建置誤觸既有生成器的歸檔及清理段，沒有執行原版。
舊較短路線全部內容已在清理前依SHA-256歸檔，恢復入口為
`work/issue4-registry-short-r1-restore.py`及同前綴JSON；依原manifest逐檔驗大小與hash。
R2生成器限定獨立/tmp輸出，不再使用既有情境的輸出目錄。
必要文字上下文匯出入口為`work/issue4-mother-return-gate-context-ida.py`及同前綴JSON／log，
限定167F6、15002、194C3與16F09。原始record79候選為
`work/issue4-mother-return-gate-text-candidate.json`，需與原版consumer實際來源核對。
NPC writer最小補充入口為`work/issue4-mother-return-gate-npc-ida.py`及同前綴JSON／log，
僅121EF..122CE，不由167F6的名稱推定玩家可見效果。
來源獨立稽核入口為`work/issue4-mother-return-gate-audit.py`及同前綴JSON／log，
核對父收據、IRQ1順序、單次預設seed控制、10個自然事件邊界及完整PNG／bin。
原版稽核PASS：39次正常輸入／78 IRQ1，174份父PNG／bin保持，10個事件邊界。
1020B入口玩家21,18，10232文字完成；沒有額外Enter或文字等待，10245自然返回21,17。
全段故事旗標保持。原始RND繼續自然運行，不要求與remake閒置呼叫次數相同。

有限READY契約：原始CTY section的handler55格由共用`region_dialogue_return_events`選取，
僅story flag17h為1時生效，可重複觸發，不設額外里程碑。
DX=0／AX=1／CL=2經121EF只將原始NPC0轉向左方，不改位置或占位。
167F6等待counter大於7，然後15002開3E6E文字窗；DI0C07依既有文字selector契約
引用D3TXT01 record79，21個非EOF原始word，含獨立姓名插值與換行，沒有FFFC。
21414自然返回，1683A等待counter等於7，1F604關窗，4F1F=2經194C3作一般碰撞檢查後北行。
handler不設／清故事旗標、不給物品或金錢、不呼叫一般步進事件dispatch。
以上流程為原始bytes、writer／consumer與正常首次Down閉合的confirmed；其他入口方向未做動態V3。
以PIT除數12428及既有315000000／3280992 Hz契約，8／7ticks各向上取整為5個60TPS更新。
轉向及末尾顯示等待採hardware-spec approximation，不稱原版逐週期wall-clock。
資料包保存場景、handler、required flag、NPC record／朝向、退回方向、文字ID與兩段hold，
共用Go只作有限轉向→文字→末尾hold→碰撞移動。缺資料失敗，不增添版本fallback。
驗收須正常母親返回與一次Down、完整文字／關窗畫布、旗標及背包保持、重複觸發、
同版本正常存讀檔及下一次Up；旗標已清分支由原始cmp／ret及元件測試鎖定。
原版音訊、原版存讀檔及其他handler不在本次有限READY範圍。
正式接線入口為`game/region_dialogue_return.go`、`internal/gamepack/region_dialogue_return.go`，
資料由`work/issue4-mother-return-gate-pack.py`從原始TXT及既有視窗契約建立；
此工具及原版收據留本機，JSON／正式原語與測試納入版本庫。
正式正常測試先通過提示窗與退回，然後在同版本Load後發現必要NPC0被初始旗標80過濾。
原始CTY00 NPC0的Ctrl=0，重製npcStep的move bit未設；帶路後位置由已確認的最後arrival frame提供。
有限重製存檔恢復契約：當目的場景、帶路完成的set／clear flags與退回事件required flag均符合，
Load預先驗證原始靜止actor與最後arrival frame，重建被過濾的必要角色，保留原始record順序。
不新增存檔欄位、不猜補其他NPC；沒有改寫原始visibility flag。此項為engine D2相容恢復，
不宣稱原版Load、NPC朝向／動畫相位或任意場景存檔畫面已對拍。
正常玩家修正已通過文字窗及陰影RGB差0、退回、重複觸發、同版本標題Load與下一次Up。
完整提示畫面仍差13145，退回畫面13340，背景圖格與人物差異保留，不稱整張V3。
可重生入口為`tools/dosgolem_mother_return.py`，嚴格來源核對為
`tools/verify_dosgolem_mother_return.py`；使用`tools/verify_dosgolem_newgame.sh`的
`--mother-return-original`模式。正式正常測試以`DQ3_MOTHER_RETURN_ORIGINAL`指定合格收據。
乾淨562208a基準入口為`work/issue4-mother-return-gate-baseline.py`及同前綴private-test.go／test／log，
以相同正常母親返回加一次Down驗證原本走到21,18的分歧，預期精確斷言失敗才接受基準重現。
來源壞檔拒絕入口為`work/issue4-mother-return-gate-negative.py`及同前綴JSON／log，
只改/tmp副本，包含Go／metadata協同篡改；原版收據與PNG保持。
成熟模擬器交叉來源：[DOSBox Staging的CB_IRQ0](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/cpu/callback.cpp)。
該stub也只保存DS／AX／DX，沒有保存CX；不據先前猜測修改dosgolem BIOS增加CX保存。
已捕捉的13個狀態另由`work/issue4-registry-entry-r8-partial-audit.py`稽核，
同前綴JSON與observed-log.txt固定截至IRQ400的實際log，避免引用仍變動的log雜湊。
範圍只到首次問候等待，全640×350比較保留，不稱完整取消或成功登錄通過。
有限稽核PASS：13個原生觀測、前604份PNG／bin逐檔相同、正常第199次Enter問候550，
續頁為第200次輸入／IRQ399及400，觀測log固定至此。
候選完整RGB差依序為127、462、335、623、367、509、7596、6988、6540、310、636、549、508。
首次問候的完整文字視窗RGB差0，沒有以局部零差異提升完整V3。
位置、raw section及原版旗標／背包／金錢保持；出生、原版存讀檔與音訊不在此稽核範圍。

完整27狀態與RGB比較入口為`work/issue4-registry-entry-prototype-compare.py`，
來源先驗完整收據與產物，試作先驗frozen Go／binary／PNG，再比較完整640×350。
文字視窗統計只作有限定位，不遮罩完整差異，也不由取消路線宣稱出生交易或原版存讀檔通過。
先前比較及協同壞來源拒絕工具亦已指向R8，未取得合格收據前不執行接受聲明。
R7部分證據稽核入口為`work/issue4-registry-entry-r7-partial-audit.py`及同前綴JSON：
199次正常輸入、398次IRQ1、十二次原生ready，前604份PNG／bin逐檔size／SHA-256相同，
第199次Enter的make早於21414／record550呼叫，break晚於呼叫。沒有第13次問候PNG或DONE，
因此只保留正常接近與Enter反證，不稱為合格取消收據。遠端訂正見
[5966967383](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5966967383)。

## 2026-10-03 城鎮攝影機（有限 READY）

依Issue #4接續上述動畫反證，先閉合城鎮視野。來源EXE、CTY00身份與IDA9.4位址基準沿用本文件。
原始linear11971..11991／file2CE1..2D01讀玩家4F33／4F35，分別減9、7，
寫4F25／4F27，再以20×15呼叫11D8A。此原始renderer沒有地圖邊界clamp。
目的場景由已確認的10130及42個自然狀態限定CTY00 section0；CTY section基底file000C，
file001E的界外圖塊為0，沿用130F4..1311E→DGROUP0B2D→11DD8..11E4B的既有loader／consumer。
不把房間的界外71套到城鎮。以上為本目的場景的confirmed，不外推其他CTY。

`work/issue4-town-camera-prototype.py`固定b7516c5，在唯讀快照的隔離副本只改目的場景camera，
沒有改seed、人物frame、位置、旗標、輸入或舊動畫。由正常創角、家中選圖、13步手動接近，
重播原版42個城鎮狀態及單次城門確認，最後正常Save／Load通過。
完整640×350與來源9358ce6e的1020A PNG逐點比較RGB差0，沒有裁切、遮罩或指定frame。
私有原型log為`work/issue4-town-camera-prototype.log`；圖片及狀態收據同前綴。
第一次原型測試有區域變數命名衝突，修正測試名稱後同來源、同命令重跑，沒有修改產品規則。

有限READY契約：`opening_escort.arrival_camera`使用既有SceneCamera型別，所有欄位必填且D3。
只在pack的destination場景套用，包含帶路完成後及同版本讀檔；其他場景不由本切片宣稱符合。
共用Go只選scene camera與執行player_anchor；CTY／section、9／7與界外0全部放入JSON。
schema0.4.0／content0.1.72，同步全部資料schema；舊schema或hash的存檔明確拒絕，不自動遷移。
正式驗收須原始EXE／CTY parity、缺失／null／未知欄位／壞幾何及D2證據拒絕、
同條正常玩家狀態、完整PNG、讀檔後camera與PNG、後續玩家路線，以及完整game／internal／desktop。
動畫DRAFT及原版音訊、完整campaign保持，不以這張城鎮零差異提升其他畫面。

正式入口與測試索引：`game/opening_scene.go`選camera、`game/game.go`消費；
`internal/gamepack/arrival_camera_test.go`核對原始EXE／CTY與損壞契約，
`game/arrival_camera_test.go`從當次manifest核對最後PNG並嚴格全RGB比較，
`game/opening_escort_test.go`保留正常42狀態並由標題選單讀檔。
欄位入口為[docs/84](84-game-pack-json-contract.md)最新城鎮camera節。
原有八筆NPC台帳另追加linear11974／1197E兩筆anchor定位，保留原始bytes、consumer及9358ce6e來源。
現行十筆由`tools/ida_dump_npc_animation.py`自動附加，新的私有匯出為
`work/issue4-town-camera-reviewed-ida.json`，前一份八筆NPC匯出與稽核保持歷史身份。

### 城鎮攝影機正式驗收（有限 CONFORMED）

schema0.4.0／content0.1.72，canonical hash為
`sha256:80a1123aa00731b8471f86f5160e46b6de12d1986e75f1f4f553aca820abd890`。
同一b7516c5、同條正常InputState的基準完整RGB差81962，隔離原型及正式接入均0。
原版38次輸入／76次IRQ1完整返回收據9358ce6e保持；來源最後1020A PNG依當次manifest驗大小及hash，
正式比較完整640×350，不裁切、不遮罩、不指定動畫frame。兩張最後畫面已目視核對。
42個人物狀態、最後旗標交易、正常標題讀檔及camera資料維持；讀檔後原版NPC位置／圖像不由此宣稱parity。
家中兩條正常玩家路線及五個視窗RGB差0、完整各124限制維持。

十種攝影機契約損壞均拒絕，含合法0界外圖塊被省略；原始EXE anchor與CTY exterior parity通過。
完整game370項頂層／45子、internal140項頂層／203子、全部11個套件及desktop建置PASS。
正常新遊戲至THE END423.50秒，僅屬remake可玩回歸；game38及internal4項選用SKIP，無素材缺失。
首次internal六個失敗因fixture留在舊schema，提前被版本檢查拒絕。
fixture改引用現行SchemaVersion，未知欄位與其他拒絕斷言保持，同工具鏈乾淨重跑通過；
已通過的game未因fixture調整重跑。原始失敗log保留，不記為產品缺陷。

正式驗收log：`work/issue4-town-camera-production.log`、`issue4-town-camera-contract.log`、
`issue4-town-camera-game.jsonl`、`issue4-town-camera-internal-clean.jsonl`；失敗internal保存在原同前綴檔。
最小充分收尾入口為`work/issue4-town-camera-final-audit.py`，輸出`work/issue4-town-camera-final-receipt.json`；
核對176個原版產物、正式正常狀態、嚴格全RGB、原始台帳、程式／文件與輸出UID／GID及Docker衛生。
最後將原版1C掃描碼對應至實際`InputState.Enter`，同條玩家路徑、42狀態、全RGB零差異與標題讀檔重驗PASS；
log為`work/issue4-town-camera-enter.log`，沒有改產品規則或重新設定seed。
本切片新增工具、測試及私有稽核入口均在本文件索引，不建立另一份工作清單。
遠端結果留言5958611871；Issue #4保持OPEN。下一段從原版城門返回後，以正常輸入續行謁見。
房間261、家中124、動畫繪圖時序、原版音訊及完整campaign仍未完成，不深入ISR或硬體逐週期。


### 母親勸告與強制返回正式驗收（有限 CONFORMED）

schema0.9.0／content0.1.80，canonical hash `sha256:d2ea836e3df3c1c39e02481e73d2efcb3eae41340aa0cfe89debf2e06162fc4e`。
正式來源為`work/dosgolem-opening/issue4-mother-return-gate-r3-receipt.json`，SHA-256
`63ee434edb5f7c30d1046c6990bf6b69630cfd2311f9c9d329d31ceb3f6cfb25`。
追蹤中的產生器獨立冷啟動，保留174份父PNG／bin；39次正常輸入／78IRQ1及十個邊界嚴格通過。
R3的十份PNG／bin與先前獨立R2逐byte相同，沒有還原狀態或重設時鐘。
固定seed1357一次；原版自然RNG照常續行，不要求後段閒置骰序相同。

乾淨562208a以相同正常母親返回及首個Down精確重現21,18／21,17分歧。
正式正常InputState通過勸告、自動關窗、返回、重複觸發、旗標／物品／金錢保持，
同一實例Load取消pending、標題選單Load及後續Up至21,16。
原版字碼、handler、NPC轉向參數及退回方向parity，14種壞契約與七種壞資產拒絕。
18種協同損壞來源拒絕，包含被改Go／metadata、PNG／bin、父來源及注入狀態；原始產物不改。
勸告文字窗與陰影RGB差0；完整提示13145、返回13340。完整畫面保持RED，沒有遮罩或固定frame。
原版Load、音訊及其他方向動態對拍未知；兩段等待維持hardware-spec approximation。

完整game422頂層清單覆蓋，376頂層／77子PASS、46選用SKIP；本批原版正常勸告另嚴格PASS。
internal151頂層／264子PASS、4選用SKIP及11套件，desktop Linux x86_64 ELF建置通過。
正常新遊戲InputState至THE END266.49秒，只屬重製回歸；無素材缺失SKIP。
九份JSON從乾淨562208a重新建構，解碼資料與正式JSON一致；保留既有檔案格式。
第一次完整程序在主線中途停止但無斷言；原因未知，不能稱已證實OOM。
第二次有界程序因600秒上限停止，仍在後段航行，cgroup記錄max／oom／oom_kill均0。
同一最終binary的已完成261項結果保持，其餘分批補驗，campaign以獨立程序通過。
首次DISPLAY未就緒及舊negative fixture漏新必填集合，均訂正驗證工具後乾淨重跑。
原始失敗log保留，不寫成產品缺陷。

本輪完整回歸入口為`work/issue4-mother-return-gate-game-regression.py`，
結果`work/issue4-mother-return-gate-game-summary.json`及同前綴clean-batch JSONL。
文件同步入口為`work/issue4-mother-return-gate-update-docs.py`。
最小充分收尾入口為`work/issue4-mother-return-gate-final-audit.py`，
結果`work/issue4-mother-return-gate-final-receipt.json`；核對來源、正常輸入、損壞拒絕、
乾淨JSON重建、完整回歸、桌面格式與輸出擁有權。
正式正常PNG與結果為`work/issue4-mother-return-gate-production-r3`同前綴產物。
R1工具錯誤觸發舊生成器歸檔後，295份較短來源已逐檔恢復大小及SHA-256；原版未執行。
恢復收據`work/issue4-registry-short-r1-restore.json`保留，原始證據無損失。
登錄所完整取消與出生維持DRAFT，Issue #4保持OPEN；本輪沒有新發行包。

獨立收尾收據24,724bytes，SHA-256
`b837a0d89ae890a60086fef35b15440d7f017dd141a3dbdfec1eb47d85ab02cb`。
此收據固定本批來源、正常玩家結果、乾淨JSON重建及完整回歸；不把完整畫面RED改成通過。
同前綴`hygiene.json`保存擁有權檢查，`post-push-receipt.json`保存提交／遠端及Issue核對。
Issue原文、更新全文及結果留言保存在同前綴`prior-body.txt`、`body.txt`與`result-comment.txt`。
新檔UID／GID1000，既有root候選3213、Markdown目錄0保持；本批DQ3容器全部清除。

### 登錄所來源的原生輸入消費探針（DRAFT）

接續1b96eb8及Issue #4。正式pack維持0.1.80，原版EXE身份與IDA9.4位址基準沿用本文件。
既有`work/issue4-registry-key-table-ida.json`、`issue4-registry-input-ida.json`已包含19530與19417。
195D0的`or word ptr DS:13h,1`受4F46的0200及0004、counter52F8閘門限制，
前面還會呼叫19810。19810只清除各actor的38h bit10及4F46的0200。
此段目前為strong，沒有正常動態閉環；不能推定命令窗或戰鬥會設低位元。
不為未證實語意加入正常路線外的事件，也不修改原版旗標、CX或計時器。

舊回程探針每個方向鍵要求`Steps > queued+11000000`才接受返回。
這是觀測器額外等待，IRQ送達與固定指令間隔都不能單獨證明輸入已消費。
原生19417呼叫210CA；1941C讀AH方向掃描碼，接續碰撞、dispatch與1991D，最後1997C返回。
文字21133完成初始化後才輪詢；2113A只在非零key flag成立後清除；
21155只在21148看見非零flag時到達，2115D清除flag並返回。
此靜態來源足以建立只讀消費觀測候選，仍須原生log與GUI狀態驗證，不能把試跑當正式收據。

私有入口`work/issue4-registry-consumed-probe.py`，前綴
`work/dosgolem-opening/issue4-registry-consumed-r1`。保持前38次母親正常輸入與174份父PNG／bin。
後續方向及文字使用已記錄的謁見／回程／首次問候輸入序列；只在原生poll邊界排入IRQ1。
每個packet記錄排入、按下／放開、讀鍵消費、下一個自然poll、人物與完整raw狀態。
以實際消費及完整返回取代固定11000000指令等待；自然等待窗仍用明示正常Enter關閉。
生成只使用隔離tmp，保留frozen Go、binary、metadata與完整PNG；不覆寫既有來源。
旗標4000或來源狀態不符即拒絕；沒有成功取消或出生的完整證據前維持DRAFT。
工作登記為`work/issue4-registry-consumed-progress.txt`；後續稽核與限制仍追加於本入口。
宣告的完整掃描碼與產生器雜湊保存在`work/issue4-registry-consumed-plan.json`。
本探針目前只到正常姓名取消與返回，出生另需獨立正常輸入來源；不把DONE當來源接受。
R1已因路線不合格停止並保留全部產物。拒絕稽核入口為
`work/issue4-registry-consumed-rejection-audit.py`與同前綴JSON；不重命名為合格收據。
第一個Up捕捉到21,15，起點原為21,17。後段packet61在國王前返回field，
packet62的額外Enter重新進入record253二選一；後續Down等方向鍵被選單消費。
因此還不能由這批來源證實移除固定等待後的合法步行與登錄流程。
下一版須另觀測keyup後的正常空讀鍵與場景輸入恢復，並以原版已確認的謁見交易完成
判定轉入回程；不能照固定packet數多送確認。上述恢復條件目前為DRAFT。
R1拒絕稽核99,784bytes，SHA-256
`62256651ad897270a7c076e3d4f192717493cb8d1b6dfae60d33d29bd7ccbe9c`；174份父PNG／bin全部相同。
新的隔離入口為`work/issue4-registry-quiescent-probe.py`，前綴
`work/dosgolem-opening/issue4-registry-quiescent-r1`，計畫為`work/issue4-registry-quiescent-plan.json`。
keydown消費後另要求keyup已送達、1941C的AH=0及下一個自然field poll，
首三個Up逐項要求21,16、21,15、21,14；不符合即停止來源，不重設座標。
謁見以原版record78等待及clear17h／set18h／50金交易完成點進入回程，
不沿用前一固定100包中未必消費的確認鍵。是否形成合格同狀態來源仍待動態驗證。
Issue拒絕及續行登記為`work/issue4-registry-consumed-rejected-progress.txt`。

### 首次北行的殘留事件查證（DRAFT，追加勘誤）

前述首三個Up的單格期待未有原版依據，不作驗收契約。quiescent-r1在完整keyup及空讀鍵後，
仍自然從21,17移到21,16、播放DI0C07，再到21,15。它在錯誤期待處停止，未完成登錄所。
北側CTY00 section0的21,16原始word為0001，file069E，selector為0，沒有handler55。
因此「北側格觸發母親勸告」已收回；不能因record79出現就把北側地圖改成特殊格。
CTY00大小7546與SHA-256 ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836保持。

既有IDA9.4的11B0B在特殊格設DS:4F46 bit0800、11B11保存DS:258C選擇器。
196D2先清0800，再讀258C並經1970B呼叫handler。
11ADE..11B04的一般格移動未清此位元。帶路的194C3強制步進未呼叫19530分派。
首次北行可能消費帶路殘留的handler55，現階段為strong，待原生前後狀態閉合。
既有南行有限CONFORMED不擴張為北行或所有一般特殊格的parity。

獨立新版入口為`work/issue4-registry-quiescent-r2-probe.py`，前綴
`work/dosgolem-opening/issue4-registry-quiescent-r2`與同前綴計畫。
觀察初次空讀鍵的4F46／258C、19530、196D2、1020B、10232及10245，保留PNG／bin。
首三次北行以舊原生觀測21,15／21,14／21,13作來源一致性檢查，不稱已確認的remake規格。
國王clear17h／set18h／50金完成後才回程，不再多送確認；取消與出生仍須來源獨立稽核。
前兩版產生器、失敗log與拒絕收據保持，追加此勘誤，不重寫歷史證據。
正式1b96eb8正常首次Up的隔離診斷入口為`work/issue4-first-up-baseline.py`，
輸入為同前綴HEAD tar；僅記錄完整等待後的位置與事件，不把現行測試當原版oracle。

### 帶路殘留勸告事件（有限 READY）

quiescent-r2由原版冷啟動自然完成，202次正常輸入／404 IRQ1。
首次空讀鍵在21,17保留4F46=0800／258C=0001；第一次Up在21,16到196D2，
1020B入口已清0800，10232完成record79，10245到21,15。後兩次Up自然到21,14／21,13。
本段結論為confirmed，推翻前版「單格期待」及「北側地圖特殊格」說法。
固定dosgolem revision2f44a68與單次seed1357，無還原狀態、位置或計時器修改。
frozen Go SHA-256 ecc9714025c26789ad7ed32e7645833a96528393a6fe9ea1bf72fe2271b49405。
EXE身份及IDA9.4位址基準沿用上方；靜態writer為11B0B／11B11，consumer為19530／196D2／1970B。

有限契約：既有pack帶路路線的強制移動若經過已審查的region dialogue return特殊格，
保存該事件穩定ID，普通格不清除。下一個成功正常步進先由新落腳特殊格覆蓋選擇器，
再消費待執行事件；required flag、actor轉向、record、自動關窗及強制返回沿用已審查契約。
阻擋的步行不消費；事件分派先清pending，故下一次普通北行不重複勸告。
場景不符的待執行事件失效；不把未閉合的其他原始handler加入production。
版本raw值、文字與方向仍從CTY及game pack取得，不新增DQ3 Go常數或fallback。

重製存檔追加可省略的deferred事件穩定ID，保存尚未消費的正常檢查點。
Load在改Game前驗證ID、pack、scene與gate；已消費的存檔不重新生成事件。
此存檔延續契約為engine D2，原版存檔parity未知；原版硬體等待維持既有近似。
驗收為正式母親返回→首次Up→完整自動提示→21,15→第二次Up→21,14，
故事旗標、金錢及背包保持；另抽測既有Down、重複觸發、阻擋不消費及前後存讀檔。
完整640×350圖片逐項報告，不以視窗相同宣稱全圖V3；音訊及出生不在本切片完成範圍。

可重生的DRAFT全路線產生器為`tools/dosgolem_registry_quiescent_probe.py`，
有限首次北行獨立核對入口為`tools/verify_dosgolem_first_move.py`。
本次全路線也自然完成謁見50金、85回程、登錄問候、姓名取消及選否返回field2,5，
但完整取消、出生與remake接線仍待獨立審查，不由DRAFT的DONE直接宣稱CONFORMED。
正式正常測試入口為`game/deferred_region_return_test.go`；資料包版號與事件證據更新，schema不變。
本批回歸入口為`work/issue4-first-move-regression.py`，來源拒絕入口為
`work/issue4-first-move-negatives.py`，收尾入口為`work/issue4-first-move-final-audit.py`。
產物同前綴保存正常PNG、JSON、完整game清單及分批結果、internal與desktop。
完整主線的CTY10對話失敗只讀診斷入口為`work/issue4-first-move-campaign-diagnostic.py`，
使用HEAD tar加明確候選檔案的隔離副本，只追加失敗狀態，不在production加debug分支。

同入口的`-r2`至`-r5`診斷保存文字、戰鬥及含裝備容量的觀測，不覆寫失敗來源。
最終完整回歸入口為`work/issue4-first-move-regression-verified.py`，
結果為`work/issue4-first-move-game-verified-summary.json`及同前綴batch JSONL。
最終來源負面稽核為`work/issue4-first-move-negatives-final.py`與同前綴JSON。
最終正常對拍為`work/issue4-first-move-production-verified`同前綴PNG／JSON／log。
目前狀態同步入口為`work/issue4-first-move-update-docs.py`；
獨立收尾收據為`work/issue4-first-move-final-receipt.json`，
提交／推送與容器清理另記於`work/issue4-first-move-post-push-receipt.json`。

收回通用文字等待改動後的乾淨回歸入口為`work/issue4-first-move-regression-clean-final.py`，
結果為`work/issue4-first-move-game-clean-final-summary.json`及同前綴batch JSONL。
`verified`程序在後段航行由agent主動停止，exit143屬工具清理，沒有完整主線PASS。
最終正常對拍為`work/issue4-first-move-production-clean-final`同前綴PNG／JSON／log。

拉米亞最後步進的遭遇診斷為`work/issue4-first-move-campaign-diagnostic-r6.py`與同前綴log／private-test.go。
有界180秒堆疊證實測試在停泊點等待cd，正式Game仍在battle.active分支。
最終回歸入口為`work/issue4-first-move-regression-r5.py`，
結果`work/issue4-first-move-game-r5-summary.json`及同前綴batch JSONL；
最終正常對拍為`work/issue4-first-move-production-r5`同前綴PNG／JSON／log。

巴拉摩斯回程法力診斷為`work/issue4-first-move-campaign-diagnostic-r7.py`及同前綴產物。
Boss前已學rec175的預留角色MP27；Boss後CTY66 section0、9,29，存活施法者MP1／0，不能再施法。
測試改為保留該角色的全部當前MP，用正式攻擊／道具參戰，不猜固定回程施法次數或注入MP。
本次完整回歸為`work/issue4-first-move-regression-r6.py`，
結果`work/issue4-first-move-game-r6-summary.json`及同前綴batch JSONL；
正常對拍為`work/issue4-first-move-production-r6`同前綴PNG／JSON／log。

追加勘誤：全部MP預留未解除回程阻塞，已收回。r8時間序列入口為
`work/issue4-first-move-campaign-diagnostic-r8.py`及同前綴log／private-test.go。
進入CTY66時魔法使者MP241；沿途已選逃跑的遭遇仍在隊長逃跑結算前消耗同伴傷害咒文MP。
Boss前剩27，全部預留使Boss結束仍27，但回程遭遇又降到1，賢者57降到0。
因此測試修正逃跑時的同伴命令為普通攻擊，明示允許同伴施法的策略保持；原Boss預留契約保持。
最終回歸入口為`work/issue4-first-move-regression-r7.py`，
結果`work/issue4-first-move-game-r7-summary.json`及同前綴batch JSONL；
正常對拍為`work/issue4-first-move-production-r7`同前綴PNG／JSON／log。

r7在金皇冠寶箱容量滿時保留物品。測試取物前經正式選單丟棄藥草，騰出一格；
全隊持有權、容量、事件旗標與存讀檔斷言保持。完整回歸入口為
`work/issue4-first-move-regression-r8.py`，結果`work/issue4-first-move-game-r8-summary.json`及同前綴batch JSONL；
正常對拍為`work/issue4-first-move-production-r8`同前綴PNG／JSON／log。

追加勘誤：r8 全域保留 MP 停用了救治，r9 僅停用逃跑傷害咒文仍在前段 mon44 全滅。
兩種全域策略均已收回，既有戰鬥策略保持。r9 失敗保存於
`work/issue4-first-move-game-r9-batch-00.jsonl`，入口為 `work/issue4-first-move-regression-r9.py`。
portal 路由原先只計算主角聖水，漏掉同伴持有的三瓶；正式道具選單支援全隊持有者。
目前只將測試路由數量判斷改用既有 `countPartyItem`，保留最後一瓶策略與正常使用選單。
本輪回歸入口為 `work/issue4-first-move-regression-r10.py`，實際結果為同前綴 batch JSONL／log。
只有全部測試通過才產生 `work/issue4-first-move-game-r10-summary.json`；缺檔不能視為通過。
Issue 更新文字入口為 `work/issue4-first-move-party-water-progress.txt`。
同一 r10 binary 的正常對拍輸出為 `work/issue4-first-move-production-r10` 同前綴 JSON／PNG／log，
嚴格來源核對為 `work/issue4-first-move-source-r10.log`。

### 首次北行正式驗收（有限 CONFORMED）

schema0.9.0／content0.1.81，canonical hash `sha256:d6c7994ee7fa53d227a98263f12606a8f4109a87f4afea01cd6996ad7f269b5d`。
有限來源`work/dosgolem-opening/issue4-registry-quiescent-r2-first-up-receipt.json`，SHA-256
`c38bcc0947c5569383b4c1b03b9d0039dd59e6fea3397ee2cd924a1e1071fe1d`。
嚴格驗證41次正常輸入／82IRQ1、174份父PNG／bin、五個事件邊界及三個正常北行狀態；
14種協同壞來源拒絕，包含producer、Go／metadata、binary、runtime、log、PNG／bin及父來源。
正式冷母親返回後首次Up進21,16、自動警告後21,15，第二及第三Up到21,14／21,13。
首次警告完整640×350 RGB差0，三張北行195／9039／22192仍RED，沒有裁切、遮罩或固定frame。
pending Save／Load會重播一次，消費後Save／Load不重新生成；壞存檔拒絕不改Game。
存檔契約為engine D2，原版Save／Load與音訊未知。既有Down、重複返回及下一步保持。

完整game426頂層清單覆蓋，379頂層／86子PASS、47選用SKIP；原版首次北行與既有Down正常對拍另嚴格PASS。internal151頂層／264子PASS、4選用SKIP及11套件，desktop Linux x86_64 ELF建置通過。正常新遊戲InputState至THE END118.78秒，只屬重製回歸；無素材缺失SKIP。
完整回歸先暴露甘達特intro與隨機戰鬥同時active；主因為已處理特殊事件後仍進一般遭遇尾段。
有限READY與原始地址訂正見docs/85，當步return保持encounter counter及原始四敵編隊。
文字確認helper的通用等待改動沒有解決該失敗，已收回；保留原有retained region等待。
之後拉之鏡受阻，r5明確記錄全隊各8格、空位0；失敗保留物品與present flag。
測試玩家經正式丟棄選單移除未裝備且無場景用途的備品，取得一格後繼續；遊戲容量規則未改。
r6只讀180秒堆疊另定位到拉米亞停泊點的cd等待，Game仍在battle.active分支。
測試在最後一步抵達時遇敵卻只空等，現先用正式戰鬥選單完成遭遇，再等待與搭乘；不改遊戲規則。
失敗、中止與所有診斷保留，不以更換seed或注入狀態取得通過。
quiescent-r2完整取消來源維持DRAFT；本段CONFORMED僅涵蓋首次三個Up的狀態及警告畫面。
完整背景、NPC相位、音訊、原版存讀檔、登錄所取消／出生及完整原版主線仍待驗。
r5後續回歸的巴拉摩斯回程資源不足由r7定位，施法者存活但MP1／0。全部MP預留未解除，r6失敗後收回。
r8時間序列閉合沿途逃跑時同伴先施放傷害咒文的消費端。全域保留MP的r8與停用傷害咒文的r9都在前段全滅，兩者已收回。
portal路由只計算主角聖水，漏掉同伴持有者；測試改用既有全隊計數與正式持有者選單，原戰鬥策略保持。
沒有補寫角色能力、改速度或戰鬥規則。
金皇冠取物前也經正式丟棄選單騰出一格，隊伍持有權及存讀檔斷言保持。
Issue #4保持OPEN；沒有新發行包。

本輪收尾核對入口為 `work/issue4-first-move-final-audit.py`。
首份收據 `work/issue4-first-move-final-receipt.json` 保存文件索引補齊前的檔案雜湊，
最終索引補齊版本另存 `work/issue4-first-move-final-r2-receipt.json`，不覆寫首份。
production 增行核對為 `work/issue4-first-move-production-diff.patch` 與
`work/issue4-first-move-production-scan.json`；數字命中為既有格式解析／存檔驗證，
其他中文與ID命中為 gofmt 重新對齊的既有欄位註解，沒有新增玩家文字或 DQ3 fallback。
更新前遠端本文為 `work/issue4-first-move-pre-update-issue-body.txt`。
本輪 Issue 本文及結果留言分別由 `work/issue4-first-move-final-issue-body.txt` 與
`work/issue4-first-move-final-issue-comment.txt` 保存；推送核對見既有索引的 post-push 收據。

### 2026-10-04 登錄所完整取消來源審查（DRAFT）

接續07c01dd，現行production為content0.1.81。先審查quiescent-r2完整202次正常輸入／404IRQ1，
再以正常新遊戲至登錄所入口的正式InputState重播。首次三個Up的既有有限CONFORMED保持。
逐項核對queued、IRQ1按下／放開、原生消費、下一正常poll、record、畫面、人物與交易狀態。
姓名取消及選否後返回field的完整來源尚未接受；出生仍需另一條正常原版來源。
本輪登記文字為 `work/issue4-registry-cancel-review-progress.txt`。
獨立來源審查入口為 `work/issue4-registry-cancel-source-review.py`，
結果為 `work/dosgolem-opening/issue4-registry-quiescent-r2-cancel-source-r1-receipt.json`。
審查328份完整packet PNG／bin、全部404 IRQ1與164組queued／consumed／capture；
原版名冊完整不變性未觀測，保留unknown，不由單一hero record推定整個名冊。
正常重製診斷的乾淨輸入為 `work/issue4-registry-cancel-head-07c01dd.tar`。
可丟棄診斷測試為 `work/issue4-registry-cancel-baseline-private-test.go`，
輸出為 `work/issue4-registry-cancel-baseline-r1` 同前綴 JSON／PNG／log。
來源審查器初版把觀測階段當作輸入交易不變項；原生日誌顯示packet43播放record78後階段由0到1。
修正審查器，逐項鎖定實際階段序列後以相同容器重跑；沒有更改原版。
來源348622bytes，SHA-256 `0467c01bbf8218a0c21581cf200cc0b491284db53c72749ddb3631b79b2cf40e`。
可重現驗證與不覆寫重生入口為 `tools/verify_dosgolem_registry_cancel.py`，
在既有專用Docker內執行 `python3 /repo/tools/verify_dosgolem_registry_cancel.py <收據>`；
要重生時加入 `--emit` 並提供尚不存在、同名稱的輸出收據及完整唯讀原版來源。
驗證涵蓋正常取消、選否及返回行走，不含出生、完整名冊、存讀檔或音訊。
r1正常重播前149個位置、section、旗標與金錢一致，packet150原版Enter進record550，
remake只收到Enter欄位，沒有交談。鍵盤實際把Enter與Space的Confirm分開；這是目前操作綁定差異。
等價交談的可丟棄診斷為 `work/issue4-registry-cancel-baseline-r2-private-test.go`，
以現行正常命令窗送兩次Confirm，明示多一個確認，不宣稱raw-key一致。
結果為 `work/issue4-registry-cancel-baseline-r2` 同前綴JSON／PNG／log。
固定來源拒絕檢查入口為 `work/issue4-registry-cancel-negatives-r1.py`，
輸出同前綴JSON／log；來源、metadata、IRQ、末尾狀態、PNG／bin及父來源損壞必須拒絕。
正常出生DRAFT探針為 `work/issue4-registry-birth-r1-probe.py`，計畫為 `work/issue4-registry-birth-normal-r2-plan.json`，
原版產物前綴 `work/dosgolem-opening/issue4-registry-birth-normal-r2`。
沿用38次冷新遊戲與正常謁見／回程／問候，姓名鍵序取自已核對的原版命名流程。
只在原生poll送出鍵盤IRQ，選第一職業、第一性別及接受能力，之後選否返回行走。
在IDA linear10924／10A9F／10816／1081C唯讀觀察候選角色與名冊旗標，
只作writer定位，不將候選buffer當成正式已登錄角色。未完成前維持DRAFT。
初次builder因舊 `issue4-registry-birth-r1-probe-source.go` 已存在而拒絕，原版未啟動，舊檔保持。
確認新前綴未使用後，以同一工具鏈重跑。builder日誌為 `work/issue4-registry-birth-r1-builder.log`。
normal-r1因新增Go宣告在既有變數之前引用而編譯失敗，日誌另存builder-r2.log；原版未執行。
修正宣告順序，保留normal-r1凍結Go，使用未占用normal-r2來源；日誌另存builder-r3.log。

完整取消來源已獨立接受，驗證工具重建202次正常輸入／404IRQ1與328份packet產物一致；
13種隔離損壞全部拒絕，原始來源未改。接受的是原版來源，remake流程仍DRAFT。
r2等價交談在packet150進tavern.active、stage0，沒有原版record550問候；完整RGB差124152。
r1原始Enter綁定差29540另保留。兩者不是對拍通過，不用修改測試期待替代正式修正。
本輪結果與出生工作登記文字為 `work/issue4-registry-cancel-reviewed-progress.txt`。
來源工具的語法、UID／GID及既有root候選檢查為 `work/issue4-registry-cancel-hygiene-r1.json`。
本輪提交／推送、Issue與來源版本核對保存於 `work/issue4-registry-cancel-post-push-receipt.json`，
出生工作若尚在執行則明示記錄live與外層逾時，不當作完成收據。
normal-r2原版已自然結束：210次正常輸入／420IRQ1，172組packet，返回2,5。
獨立審查為 `work/issue4-registry-birth-normal-r2-source-review.py`，
結果為 `work/dosgolem-opening/issue4-registry-birth-normal-r2-source-r1-receipt.json`。
核對174父產物及前153組取消來源的狀態／PNG／bin保持；其後正常姓名、六職業、性別、能力等待及接受。
IDA linear10A9F..10ABD從DS520B向DS52FD+DX*61h複製97bytes；本次DX1，
10816之前slot1狀態0，1081C之後狀態1。候選128bytes只有前97bytes屬本次copy，
不把未直接觀測的目的record bytes或其他職業／性別路徑當已完成。
出生來源審查PASS：368685bytes，SHA-256
`de818064f9bcaf3f36dcd6a5d8b9259908781356a27c128cf1e8b41218c10ce3`。
可重現產生器為 `tools/dosgolem_registry_birth_probe.py`，與已執行私有producer逐byte相同。
獨立重建／不覆寫重生為 `tools/verify_dosgolem_registry_birth.py`，
在既有Docker內執行 `python3 /repo/tools/verify_dosgolem_registry_birth.py <收據>`，重生時加`--emit`。
產生器沿用 `/repo`唯讀、`/work`UID1000可寫與凍結 `/dosgolem`唯讀掛載；
`/work/dosgolem-opening`須已存在，且此normal-r2前綴未占用，不能在既有來源上覆寫。
出生來源陽性重建與13種損壞拒絕為 `work/issue4-registry-birth-normal-r2-negatives-r1.py`，
結果為同前綴JSON／log，所有原版及父來源保持唯讀。

### 登錄所下一個實作閘門（DRAFT）

已接受的原版取消0467c01b與單一戰士男性出生de818064提供正常輸入順序與交易邊界。
已證實：550兩次內嵌等待→是否登錄→554姓名；姓名完成後六職業→性別→能力等待→接受能力。
已證實：取消姓名走558再次詢問，選否560再返回行走；成功登錄走559再次詢問，選否560再返回。
已證實本次登錄：候選初始化、97bytes writer、slot1由0到1，hero／gold／story與位置保持。
限制：目的record bytes未直接觀察，其他職業／性別及名冊滿額路線未動態驗證；原版Save／Load與音訊未知。
正式修正前須把3E9C／3E6E、職業、性別與能力視窗的SI consumer對照成typed pack契約，
不以截圖目測或現有Tavern的40,40大框補值。原始資料→loader→狀態機→正常交談／出生→存檔→對拍一起驗收。
正常對拍明示使用現行命令窗的等價交談；raw-key綁定差異另記，不暗改全域操作。
姓名／職業／性別／能力與文字皆由pack提供；共用Go只處理具名狀態機，不能新增DQ3 raw ID或fallback。
目前正式流程仍DRAFT；證據足以完成此下一步，不需使用者補資訊或重選方向。
出生工具語法、執行版本一致及資料所有權收據為 `work/issue4-registry-birth-normal-r2-hygiene-r1.json`。
出生來源結果留言為 `work/issue4-registry-birth-normal-r2-result-comment.txt`；
遠端Issue本文的更新前快照及更新文本為同前綴`prior-issue-body.txt`／`issue-body.txt`。
最終提交、來源完整性、Issue與Docker清理記錄沿用 `work/issue4-registry-cancel-post-push-receipt.json`。

### 2026-10-04 登錄視窗consumer審查入口

非破壞匯出為 [`tools/ida_dump_registry_ui_contract.py`](../tools/ida_dump_registry_ui_contract.py)，
自動合併已分級的 [`tools/ida_registry_ui_ledger.json`](../tools/ida_registry_ui_ledger.json)。
原始EXE唯讀，由既有IDA9.4 image在一次性database執行；輸出路徑必須不存在。
本輪私有匯出為 `work/issue4-registry-ui-consumers-r1-ida.json`、`r2-ida.json`、`r3-ida.json`；
同前綴Python與log保留形成史。正式工具的審查後重生與有限READY契約記錄於下方。

### 登錄視窗與正常交易的有限 READY 契約

本段取代上方「視窗SI consumer尚未閉合」的目前待辦；歷史段落保持。
目前正式remake仍為schema0.9.0／content0.1.81，完整登錄流程尚未實作。
READY只涵蓋已接受的正常姓名取消路線及單一戰士男性出生路線。
其他職業／性別的動態畫面、滿額替換、原版存讀檔及音訊不隨本段升格。

輸入為`assets_raw/DQ3.EXE`，115282bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
IDA Pro9.4使用linear位址；MZ file=linear−EC90，DGROUP基底linear24DD0。
原始資料、一次性database及輸出保持私有；語意以原始定位為key附加。
審查後匯出為`work/issue4-registry-ui-reviewed-r1-ida.json`，2919168bytes，SHA-256
`6064fa977b2c6bcfc8335a4a68508c08821b853f430970c705ea459c5da3b798`。
工具SHA-256為`20a1694af4736df06ef1254c96de3796a12d355cea2d607915dbe2b7e3fbab02`；
ledger SHA-256為`a8e9c86a55e94cf06c0e7dfadc2b3e088118b21648a3d38ec337e312d5cbf4fb`。
1649筆原始指令中，26筆confirmed、6筆strong、1617筆unknown；未審查項目沒有批次升格。
ledger保留既有16筆NPC旁註並追加32筆登錄旁註，匯出自動核對EXE bytes及MZ relocation。

| 原始窗口定位 | SI consumer與像素投影 | 本段用途與限制 |
|---|---|---|
| DGROUP3E6E；linear28C3E；file19FAE | `15002→1F590`；x=152、y=238、width=352、height=96；record404 | 共同文字框。`21414`取x byte+2、y+16及保留行位置；不能換成80px高HUD |
| DGROUP42B8；linear29088；file1A3F8 | `1076B→10D17→1F590`；152,46,256,144；record451 | 姓名框，沿用已驗證的共用姓名盤與導航 |
| DGROUP3DD8；linear28BA8；file19F18 | `107A1→1F4E3`；344,46,128,128；record555 | 原始count=5由1078F覆寫6。職業先後序不得由舊八職業UI推定 |
| DGROUP3DF6；linear28BC6；file19F36 | `107BD→1F4E3`；344,46,96,64；record556 | 姓名及職業完成後才進性別 |
| DGROUP3DA8；linear28B78；file19EE8 | `107DD→1F590→1834E`；152,46,352,192；record407 | 能力與已裝備物品，等待一次正常按鍵後才進確認 |
| DGROUP3DC2；linear28B92；file19F02 | `107E7→1F590`；360,14,176,48；record557 | 能力接受詢問，沿用已驗證的框線／陰影原語 |
| DGROUP4080；linear28E50；file1A1C0 | `1F63C→1F4E3`；360,62,112,64；record434 | 共用Yes/No選單，不能把Enter同時用於能力等待與接受 |
| DGROUP3E9C；linear28C6C；file19FDC | 152,238,352,80；record401 | 場景HUD的既有來源，沒有證據可將它當登錄文字框 |

以上投影由`1F4E3/1F590`讀`[si+2/4/6/8]`、EGA每byte八個像素及
`1FC57/1FD30`的原始consumer導出，沒有以截圖估計窗口尺寸。
單列選單的游標由`1F956..1F964`讀`[si+18h/1Ah]`，y每項加16。
職業及性別起點為360,62，共同Yes/No為376,78；游標／滑鼠幾何目前為strong靜態結論。
`1F779`將選項初始化為1，方向鍵環繞選項，Enter與Space接受，Escape取消。
`1F908`的無滑鼠游標使用glyph11／12；存在滑鼠時的反白路線不在本段動態驗收。

| 狀態與交易 | 原始定位與推論等級 | 正式實作要求 |
|---|---|---|
| 問候及是否登錄 | 106DF／106E7，confirmed；取消與出生packet150至153 | 保留record550兩次inline wait，再開Yes/No；不能直接開職業窗 |
| 姓名 | 10763／1076B，confirmed；共用10D17的10D84長度gate | 播554後進姓名；空名選完成留在姓名盤，不回退為職業名 |
| 姓名取消 | 1076E..10775／1081F，confirmed；取消packet160至161 | 播558，重新詢問；不建角色、不消耗能力RNG、不改名冊 |
| 六職業 | 10789／1078F／107A1，confirmed；107B6映射為strong | 原始表DGROUP4F54／file1B094為0001020304060705；一基選項映射1,2,3,4,6,7。原版可見六項為戰士、武鬥家、僧侶、魔法使者、商人、遊玩者 |
| 性別及候選能力 | 107BD／107D1／107D4，confirmed限本次cursor1男性 | 完成職業後才生成候選；候選尚未進名冊 |
| 能力等待與接受 | 107E4／107E7／107EE，confirmed；出生packet167至169 | 新按鍵只從等待到確認；選是之後才commit，不重用同一輸入 |
| 名冊登錄 | 10811／10816及10A9F..10ABD，confirmed限本次DX1 | 複製97bytes後slot1狀態0→1。重製只加入roster，不自動加入companions；目的record bytes仍未直接動態觀察 |
| 再次詢問及退出 | 1082A／1083F／10847／1084C，confirmed | 成功559或取消558後詢問；選否播560，最後等待後擦框，返回原來2,5 |
| 首次詢問選否 | 106EA..10700，strong | 播551後走最後等待。此分支未有動態來源，不能與再次詢問的560合併宣稱已驗 |
| 名冊容量 | 10974／10706，strong | slot1..11計算registered與active，排除hero slot0。滿額替換支線未動態驗證，不猜刪除／替換流程 |

typed pack欄位及有限狀態機邊界見[docs/84](84-game-pack-json-contract.md)「登錄契約的READY設計」。
正式Go只查stable text ID、已驗證的class映射、共用窗口引用及具名狀態機；
原始record、CTY／section／handler、class raw值、游標幾何及可調容量都保留在pack。
保留文字須由共同caller開框時清空，後續record延續原始行位置，不能每段重新清除550。
候選能力沿用已驗證的Lv1規則；本次兩側只有冷啟動seed1357一致，長路線閒置RNG條件尚未對齊，
不能因種子數字相同就要求這名候選所有能力逐值一致，也不重新設種挑選數值。

正常驗收先重播已有38次冷新遊戲、謁見及回程，前149個位置／section／旗標／金錢保持。
交談使用現行命令窗的明示等價輸入，保留raw-key Enter差異；不暗改全域綁定。
取消路線須經姓名功能列選取消，再選否、正常下一步及同版本存讀檔。
出生路線須經正常姓名、第一職業、男性、能力等待與接受，再選否、名冊保存／讀回及樓下招募。
Load取消未接受候選與UI暫態，不增加名冊、不消耗額外RNG；存檔只屬engine契約，原版Load未知。
每一狀態保存完整640×350 PNG並逐項報告RGB差異；窗口局部相符不能宣稱完整V3。

本輪獨立核對入口為`work/issue4-registry-ui-contract-review-r1.py`，
結果為`work/issue4-registry-ui-contract-review-r1-receipt.json`。
新IDA查詢曾提供非指令邊界10AAA，工具拒絕；改為資料庫的10AA9／10AAB，沒有改原始定位。
較早讀取程式的不存在`game/party.go`歸為搜尋腳本問題，正式名冊入口位於`game/recruit.go`。
這些工具失敗不當作產品缺陷，現行remake的問候與順序分歧仍由前輪正常輸入收據保留。
獨立審查收據6832bytes，SHA-256
`9e5017de3c4b77504b47327e91c544377e106f92244cf74fdcdb02c14879f57f`。
審查器先漏掉相依module搜尋路徑，再將IRQ1事件陣列當計數；修正為正式工具目錄及陣列長度後，
以相同Docker命令乾淨重跑通過。原版及既有收據保持不變。
六種隔離損壞入口為`work/issue4-registry-ui-contract-negatives-r1.py`，結果為同前綴JSON；
1239bytes，SHA-256 `94e24ec3e042abcd4380326f3afa636356172476aa3c4c52260fbd4fc2cb14ae`。
拒絕覆寫的實際IDA日誌為`work/issue4-registry-ui-refuse-overwrite-r1.log`，exit2且已接受sidecar保持。
本輪Issue文字為`work/issue4-registry-ui-ready-r1-issue-comment.txt`；
工具、所有權與索引核對為`work/issue4-registry-ui-ready-r1-hygiene.json`，
提交／推送、Issue及清理收尾為`work/issue4-registry-ui-ready-r1-post-push-receipt.json`。

### 登錄實作檔案入口

`dq3_remake_ebitan/internal/gamepack/registration_test.go` 比較原始 EXE 視窗、職業表、TXT 記錄與 CTY NPC，並拒絕缺欄位、null、未知欄位及不合法引用。

`internal/gamepack/registration.go` 定義嚴格的登錄資料契約；`game/tavern.go` 執行具名登錄狀態；`game/indexed_window.go` 共用原版視窗 primitive。下列路徑均以 `dq3_remake_ebitan/` 為根。取消與正常出生驗證沿用上節 dosgolem 收據。

`dq3_remake_ebitan/game/registration_test.go` 鎖定空名、接受邊界、取消、滿額 fail-closed，並以正式冷開機輸入比較 dosgolem 取消與出生收據。使用私有收據時明確設定 `DQ3_REGISTRY_ORACLE_DIR` 與 `DQ3_REGISTRY_RECEIPT_DIR`；原版素材與產物維持不入 Git。

### 2026-10-04 正式登錄狀態鏈實作與驗收

沿用上一節有限READY，schema0.10.0／content0.1.82；canonical hash為
`sha256:140f2d39be09e8d092d40efb779b247f85222525e7ff5609285d90d1375644ea`。
正式handler由pack binding提供，姓名在職業之前，六職業為raw1,2,3,4,6,7。
空姓名不放行；性別選定後建立候選並顯示能力，下一個正常按鍵進接受詢問，接受後才加入名冊。
取消使用558再詢問，成功使用559再詢問；再次No為560，最終正常按鍵返回field2,5。
初次No551保持static strong。滿額替換尚未READY；滿額Yes不創建、刪除或重擲角色。

文字同一caller開框時清空，後續record延續原始行計數與捲動。
原始IDA9.4 linear10D9C `8d36b842`與10DA0 `e861e8`，分別取姓名窗DGROUP42B8與呼叫1F604撤窗。
職業／性別頁不得重畫姓名盤；正常原版packet165／166直接確認此效果，限定confirmed。
保留原始名稱、指令與位址；來源仍為6064fa97匯出，EXE身份與位址換算不變。
renderer沿用共用視窗primitive、當前場景palette bank及pack的字模前景色；不能把整個世界palette或標題覆寫用在此窗。

| 實際驗證 | 結果與限制 |
|---|---|
| 原版來源 | 取消0467c01b及出生de818064保持；dosgolem2f44a68，原版EXE5178fdc8，seed1357冷開機固定一次；沒有修改或注入原版 |
| 正常取消 | 母親正常checkpoint後164個queued packet，前149個field狀態及150..164登錄狀態通過；不寫roster／companions；來源含202次正常輸入／404IRQ1 |
| 正常單一出生 | 172個queued packet；165職業、166性別、167能力、168接受、169才roster1、172返回；姓名0、Warrior、male、Lv1、布衣及空槽一致；來源含210次／420IRQ1 |
| 等價輸入 | 原版Enter交談在remake多一次正式命令確認Talk；不宣稱raw key parity。英雄、位置、金錢、旗標與正常場景入口未注入 |
| Save／Load | 兩條正常路線返回後同版本往返通過；出生後正常樓下招募及再存讀通過。無效schema不改候選，有效Load清除候選並保留配置。原版Load仍unknown |
| 完整PNG | 共38張，未裁切、遮罩、固定人物或重擲；姓名296、問候／選單／職業／性別608、能力／接受263、返回677個RGB差異，完整V3仍RED |
| 能力亂數 | 原版HP11、remake13。初始seed相同仍未證明長路線亂數條件可比；目前不驗收完整向量或精確骰序，不把差值包裝成通過 |
| 標準回歸 | r7完整game430頂層清單，382頂層／86子PASS、48選用SKIP；正常新遊戲至THE END92.12秒，只屬remake可玩。無素材缺失SKIP |
| 資料與建置 | 全部11個internal套件，153頂層／281子PASS、4選用SKIP；新增17個損壞契約拒絕。desktop main.go建置為Linux x86_64 ELF，不是新發行包 |

目前等級為登錄狀態鏈有限E2／E3及畫面V2，沒有宣稱全視覺CONFORMED。
其他class／gender的動態證據、滿額替換、目的record bytes、原版樓下招募、原版Load、音訊與完整campaign仍未知。
下一切片延長原版正常樓下招募，不重開本批已閉合的視窗consumer。

可重播入口為`game/registration_test.go`的`TestRegistrationDosgolemNormalInputComparison`。
設`DQ3_MOTHER_FINISH_ORIGINAL=work/dosgolem-opening/issue4-mother-finish-receipt.json`，
`DQ3_REGISTRY_ORACLE_DIR=work/dosgolem-opening`及一個新的`DQ3_REGISTRY_RECEIPT_DIR`，均以容器的絕對路徑為準。
先用`go test -c`編譯，再在有界Xvfb及正確assets_raw工作目錄執行；工具鏈為dq3-ebiten-test:20260822-r1。
圖像、原版、database與執行封包留本機，不加入Git。

本機證據：`work/issue4-registry-impl-r6/registration-birth-false.json`為3238bytes，SHA256
`752619bd13f1a0ab54deee991114ff339a9d4a20620e0162afd155d72696f14a`；
出生`registration-birth-true.json`為5566bytes，SHA256
`615a519d1567262cd6818340e3eab181979e2b1468c3b57df3280b8c4f07643e`。
完整game收據`work/issue4-registry-impl-r7-full/receipt.json`為24763bytes，SHA256
`0d57da62b2ce8e7ab18d7a4e46a67751b4417895cce18ec56a3feb9201a82e81`。
r1／r2環境設定失敗、r3／r4未畫窗、r5姓名窗殘留及campaign舊商人操作的輸出保留，詳見WORKLOG。

### 2026-10-04 首次樓下招募選單的有限 READY

範圍是正常出生後下樓、正式交談、record527內嵌等待、record528及三項首次選單。
入隊後文字、分離、查看、原版Load與音訊不隨本段升格。完整RGB仍另行驗收。
原版EXE及位址換算沿用前節：115282bytes、SHA2565178fdc8，IDA9.4 linear，file=linear−EC90。

原版r4來源為234次正常輸入／468次IRQ1，196個queued／consumed／capture與392份完整packet PNG／bin。
174冷啟動父產物及前172個出生packet的344份PNG／bin逐byte保持。
本次直接觀察DS535E的97bytes，與出生10A9F候選前97bytes相同；slot1已登錄且首次選單前狀態1保持。
此目的record不變性限定本次Warrior male，不外推其他職業、完整名冊或存檔。
正式收據432468bytes，SHA256 `a85cad67ab6611291b82824973ec39de0e2192f4ad9808b0f362bd18a6e4fda9`；
較早r1來源1baabf39保留，r2增加共同對拍器所需的頂層EXE身份，不覆寫r1。

| confirmed 原始定位 | 正式行為 |
|---|---|
| IDA1036D `e8924c`、10370 `bf0f02` | 建共同文字框，開始record527，packet195為一次inline wait |
| IDA10378 `bf1002`、10380 `8d36323e`、10384 `e85cf1` | 問候EOF後直接record528，再由DGROUP3E32進1F4E3；packet196 count3、cursor1 |
| DGROUP3E32、linear28C02、file19F72；1F4E3與1F956 consumer | 原始byte座標43,30、寬18、高80投影344,30,144,80；游標360,46、行距16，由原始consumer導出 |
| 正常原生poll與IRQ1 | 出生field2,5經原有通路反向下樓，CTY00 sec0生於8,14，正常到2,18交談；英雄、金錢與故事旗標保持 |

pack新增必填`interface.recruitment_entry`。binding、問候text ID、有序具名選項、原始視窗、游標與點擊幾何皆由JSON提供。
Go只執行問候→首次選單，不新增DQ3座標、record或玩家文字。文字EOF及選單不共用同一個按鍵。
同版本Load清除招募UI暫態，錯誤存檔不改當前問候；此為engine存檔契約，原版Load仍unknown。
資料契約見[docs/84](84-game-pack-json-contract.md)的首次招募入口節。

重生產生器為 [tools/dosgolem_recruitment_entry_probe.py](../tools/dosgolem_recruitment_entry_probe.py)。
沿用repo與凍結dosgolem唯讀、work UID1000可寫及既有dq3-ebiten-test:20260822-r1；
正常從新遊戲開始，seed1357只設定一次，無座標、旗標、名冊或角色注入。
新產物前綴為issue4-recruit-normal-r4，已占用即拒絕，不在舊來源上覆寫。
[tools/verify_dosgolem_recruitment_entry.py](../tools/verify_dosgolem_recruitment_entry.py)獨立重建來源；
容器內執行 `python3 /repo/tools/verify_dosgolem_recruitment_entry.py /work/dosgolem-opening/issue4-recruit-normal-r4-source-r2-receipt.json`。
重生加`--emit`並提供尚不存在、同basename的輸出及完整原版父鏈。
驗證固定234／468／196、完整通路與record、IRQ按放鍵、消費階段、PNG／bin及出生前綴。

IDA重生入口為 [tools/ida_dump_recruitment_entry_contract.py](../tools/ida_dump_recruitment_entry_contract.py)，
自動合併 [tools/ida_recruitment_entry_ledger.json](../tools/ida_recruitment_entry_ledger.json)五筆已分級旁註。
以既有IDA9.4 image在一次性DB執行，EXE唯讀、輸出須不存在；不改原始名稱、位址或operand。
正常InputState入口為`TestRecruitmentEntryDosgolemNormalInputComparison`；需設定`DQ3_MOTHER_FINISH_ORIGINAL`、
`DQ3_RECRUIT_ENTRY_ORACLE_DIR`及新的`DQ3_RECRUIT_ENTRY_RECEIPT_DIR`，先go test -c，再從正確game目錄以有界Xvfb執行。

r2下樓計畫在3,5碰到家具，原版未到櫃檯。DRAFT導航審查0c3e6b69只接受出生目的bytes及既有父鏈，不接受招募。
r3改用已驗證反向通路，正常到首次選單並選擇入隊；packet198見slot1由1→2與record536。
其後record537／538出現4000計時旗標，探針拒絕。後段旗標仍待分類，不能稱為產品缺陷或最終入隊收據。
r4在首次選單自然停止，沒有放寬計時檢查或挪用DOSBox圖像。
PCM／硬體時序仍依AGENTS停止線，不深入driver或ISR。完成聲明只限首次招募入口；後續分支需另續原版來源。

### 首次招募入口正式實作與驗收

schema0.11.0／content0.1.83，canonical `sha256:1e30085e7a81e37446f20699051048b5e254ec30978d124ae5a96921ac75dc10`。
Go的`game/recruitment_entry.go`與`internal/gamepack/recruitment_entry.go`只處理共用狀態及嚴格引用；原始資料在pack JSON。
正常InputState從母親合法checkpoint重播全部196個packet，位置、場景、金錢、旗標及問候／首次選單通過。
同版本Save／Load、Load後正式下一步與壞Load不改問候通過，原版Load仍unknown。
47張完整入口PNG逐byte等於隔離試作，38張取消／出生逐byte等於前次正式輸出；兩張首次問候／選單各差295RGB。
有限狀態E2／E3及畫面V2，完整V3仍RED，不把後段未接受的r3交易升格。

完整game434清單，385頂層／86子PASS、49選用SKIP；正常THE END66.12秒，只屬remake回歸。
internal155頂層／293子PASS、4選用SKIP及11套件；12種壞pack拒絕。desktop main.go建置為Linux x86_64 ELF。
八種隔離壞原版來源拒絕且原始收據保持。正式IDA匯出1846499bytes、SHA25683c9ac67，五筆confirmed旁註按原始定位合併。

正式完整回歸為`work/issue4-recruit-entry-production-r1-full/receipt.json`及同目錄logs。
合併原版對拍程序在第三條路線被SIGKILL；前兩條未當作整批通過，失敗log保持。
以相同image及4GiB條件拆成獨立程序乾淨重跑，入口為`work/issue4-recruit-entry-production-r2-native-run.py`，
結果在`work/issue4-recruit-entry-production-r2-compare`。末次只補測試的Load後正式下一步，production Go／JSON與完整回歸相同。
來源拒絕入口為`work/issue4-recruit-entry-negatives-r1.py`及同前綴JSON；source CLI與IDA producer皆拒絕覆寫既有來源。
所有原版、PNG／bin、database、完整執行檔與使用者scratch留本機；沒有新發行包。Issue #4保持OPEN。
提交前檢查入口為`work/issue4-recruit-entry-final-audit.py`，末次收據為`work/issue4-recruit-entry-final-r2-receipt.json`。
新增Go掃描為`work/issue4-recruit-entry-added-go-scan.json`；命中限於測試證據、共用格式上限、schema版本與註解，無新增版本專屬production fallback。
提交、遠端與Issue核對收據保存於`work/issue4-recruit-entry-post-push-receipt.json`。


### 2026-10-04 入隊後文字與播放等待 DRAFT

正式程式保持da720aad49b1891dc2c32436706e4f034e7bb386，schema0.11.0／content0.1.83與1e30085e canonical不變。
本段取得原版前198個正常packet的有限證據及正式remake反證。完整入隊尚未READY，不修改正式Go／pack。

來源為原版assets_raw/DQ3.EXE，115282bytes，SHA2565178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c。
IDA Pro9.4採linear；file=linear−EC90，DGROUP基底24DD0；原生dosgolem PC仍以CS×16+IP+EF00對應。
dosgolem2f44a68、dq3-ebiten-test:20260822-r1、seed1357冷啟動固定一次；未還原state或注入座標、名冊、旗標或clock。
新的r5只把診斷guard改成記錄首次旗標並有界觀察20000000指令。公開正式r4 producer及其拒絕規則保持。
r5的DRAFT_TIMER_STOP不是完整入隊完成標記，不作正式來源接受。

| 有限原版觀察 | 證據定位及等級 |
|---|---|
| packet197進一名角色的清單，record530、cursor1、count1 | 前196個queue／consume／capture／IRQ及392份PNG／bin逐項保持；有限confirmed |
| packet198入隊，顯示record536內嵌等待 | slot1狀態1→2，目的97bytes保持；10415→10473、10491原operand[bx+4F64h]寫2及10422 record536；限定本次單一戰士男性confirmed |
| packet199正常Enter送達且消費，依序537／538 | 1042D／1044C原始record載入；此packet未完成，不加入已接受狀態鏈 |
| 4000第一個實際writer | step3776635393，previous_pc1FEFC、pc1FF02，DGROUP0013=4002、0007=4E20；對本次writer confirmed，旗標完整用途仍unknown |
| 有界觀察末尾 | step3796635394、PC208F3、counter4EC5，仍packet199／record538／474IRQ1；confirmed停點，未證實根因 |
| 音訊控制流 | 1043F呼叫2074E控制9，10444 BP20h經20770→20950，10454→208E2等待，10459呼叫20754控制10；pause／resume名稱只作strong，不當作已驗證平台API契約 |

4000首次寫入已閉合原始counter比較及writer；本次沒有經另一候選16F3B。
208E2先讀DS2898，再查22E10介面；清除完成狀態與返回條件保持原樣。
這足以否定「只解除4000 guard即可閉合入隊」的做法：解除診斷拒絕後原版仍在等待，未交還控制。
它尚不足以判定是dosgolem時基、完成回呼、音樂資料或原版本身的問題；不稱原版bug、防拷或PCM故障。

2092B的檔名DS0120原始bytes為ebg.mcx；20938..20946讀到DS253C+36B的資料段。
20975..2097E以BP−1E選四byte表，BP20h因此指向EBG.MCX第2號零基項。原始offset559，下一項6B4。
這條路線是事件音樂資料，不能沿用BP<1E的VCX PCM音效解釋。具體曲名與時長仍待已知CMF資料契約核對。
下一步只核對遊戲選曲參數、公開CMF／CT-MUSIC完成介面與dosgolem既有能力。
不得深入原版硬體driver／ISR，不清4000、寫DS2898、改clock、跳過等待或以DOSBox圖像替換dosgolem來源。

正式InputState的可丟棄診斷從母親合法checkpoint走完整198個packet。
packet197完整RGB差51535，packet198差56790；現行recruit stage1、dialogue=false、roster0、companions1。
交易數量吻合，但缺原版536文字及等待，仍留在舊清單。這是已觀察分歧，不把診斷PASS當入隊parity。
原版完整返回、下一次操作、原版Load及聲音仍未知；後續同版本存讀必須在READY後實作一併驗收。

公開重生入口為[tools/ida_dump_recruitment_join_wait_contract.py](../tools/ida_dump_recruitment_join_wait_contract.py)，
自動合併[tools/ida_recruitment_join_wait_ledger.json](../tools/ida_recruitment_join_wait_ledger.json)。
14筆旁註為13confirmed及1strong；已證實只限每筆列出的來源與範圍，未審查列仍unknown。
ledger保存原始file bytes，匯出按MZ relocation核對載入bytes；10454原始9AB2035310、IDA載入9AB2035320不能混用。
正式r1因far call的原始／載入bytes混用而拒絕；修正基準後r2完整重生，沒有改原版或放寬檢查。

獨立稽核入口為[tools/verify_recruitment_join_wait_contract.py](../tools/verify_recruitment_join_wait_contract.py)。
以既有image、repo及原版唯讀，work輸出UID1000，在容器內執行：

    python3 /repo/tools/verify_recruitment_join_wait_contract.py /repo/work/dosgolem-opening /repo/work/issue4-recruit-join-reviewed-r2-ida.json /repo/work/issue4-recruit-join-remake-draft-r2-compare/join-draft-r1.json

IDA重生沿用上節一次性DB及-S入口，輸出改用尚不存在的檔案；正式收據不覆寫。
診斷產生器為work/issue4-recruit-join-r5-builder.py，組合既有r3 DRAFT探針；凍結Go、binary與meta皆保存。
前198審查入口work/issue4-recruit-join-prefix-r1-review.py；有限收據424754bytes，SHA2568e49fb70b2b2a4428585a2391eac6cc56fef3dca30cc9206b93cd9741f0c2f0d。
完整r5.log為831714bytes，SHA2563973569b328d810dc0f0d1da368da170716f41424f94062044c36cfe4bc52134。
正式IDA收據712003bytes，SHA2561274a0628dfaf8f28465cbd064e16fe8d3de44ecf3173f2e9477bba589bb9658。
重製可丟棄入口work/issue4-recruit-join-remake-draft-r1.py；r1因DQ3_ASSETS未設而素材SKIP，沒有當成驗收。
相同binary加正確DQ3_ASSETS後r2正常輸入7.03秒，收據13327bytes、SHA2565563edb23e4860922013b9b238ad6dd01a479da07bec016043a7d5340c998132。
隔離壞來源及所有權檢查保存於work/issue4-recruit-join-review-r1.py及同前綴收據；八種損壞皆拒絕，原始輸入保持。
文件最後修訂後的八檔雜湊保存於work/issue4-recruit-join-review-r2-receipt.json；提交與Issue核對為work/issue4-recruit-join-post-push-receipt.json。
本輪未改production，不重跑既有完整game／internal；最近正式回歸仍由前節收據仲裁。原版、PNG、database及binary留本機。

### 入隊播放等待分支的只讀續驗

r6從同一正常冷啟動重播，seed1357及13項來源條件與r5保持，沒有增加AdLib或clock參數。
只在應用程式caller與等待入口讀原始狀態，首次等待後1000000指令自然停止；未寫入guest state。
獨立核對前198個packet、472IRQ1及396份完整PNG／bin逐byte保持。第199個packet仍未完成。

| 本次confirmed有限觀察 | 原始定位及狀態 |
|---|---|
| BP20h確實進EBG consumer | step2136382143、IDA linear10447，BP0020；2136382150進20959，DGROUP286D=0001 |
| cue後狀態 | linear1044C時DGROUP2898由0變1，2871保持0；不聲稱曲名、時長或聲音已驗收 |
| 等待分支 | step2137703544進linear208F5；286D=1、2898=1、2871=0，進22E10查詢，沒有走2091B分支 |
| 有界停點 | step2138703538、linear208FA；2898仍1、counter1915、474IRQ1，未返回玩家操作 |

IDA9.4限定應用程式wrapper匯出確認：linear22E14比較原operand`ds:5C02h`，非零時22E1D以`xchg al,ds:5C02h`交換FF；零時AX0返回。
因此22E10並非直接呼叫FMDRV的查詢服務，它讀應用程式提供的共享byte；此為原始bytes的confirmed結論。
linear22D47..22D50以DX=DGROUP、AX=5C02h、BX=1呼叫23920，原值存CS22D67／22D69。
23920的原始signature為FMDRV，入口跳24A2F。此狀態位址設定為strong，尚缺動態設定端與公開介面契約，不把driver內部writer猜成已證實。
5C02直接xref為空，register-based位址傳入已見，不能據此宣稱沒有writer。driver／ISR未展開。
dosgolem系統控制埠61h回保存值，與本次故障的因果未證實；此線索不作修法依據。

只讀產生器為`work/issue4-recruit-join-r6-builder.py`；獨立審查入口為`work/issue4-recruit-join-r6-review.py`。
收據`work/issue4-recruit-join-r6-review-receipt.json`，SHA256`ebc0ec6c2ed21cbc4947076e14bc8347800562c04c9dcf85a40a0c908028c78f`。
完整r6.log SHA256`8527809ee9a15855e056c82f7bc06b0d45a5b8f73c5e8139bee94b3ab2e74af9`；source及binary雜湊見該收據。
限定wrapper匯出入口為`work/issue4-recruit-join-interface-build-r1.py`及`-r2.py`，產生同前綴`interface-r1-ida.py`／`interface-r2-ida.py`與JSON。
r2 sidecar SHA256`99652386dd01214b98c05df48e5a733675d2b120cb94ed7e3ca9616ffbbf7841`，script SHA256`ff4a5c23677e261b2c06e30ea6b1852ca34fedb19df2253601618f5f2ced2a20`，輸入與位址基準沿上節。
最終核對為`work/issue4-recruit-join-r6-final-audit.py`及同前綴收據；Issue結果文字為`work/issue4-recruit-join-r6-result.txt`。
完整入隊仍DRAFT，正式程式與pack保持。下一步只核對狀態位址介面及dosgolem能力，再由原版正常重播返回，不清共享byte或跳過等待。

### 公開狀態介面與正常播放返回值

本輪命中原版對拍、平台規格及READY閘門，載入既有入口；不增加硬體driver或ISR逆向範圍。
[RBIL的SBFM介面紀錄](https://files.mpoli.fi/unpacked/software/texts/computer/inter56d.zip/interrup.p)
將BX1定義為DX:AX狀態byte位址，BX2為樂器表、BX4為驅動時脈除數、BX6為播放。
BX6的AX0表示成功，AX1表示已有音樂；BX9／10是暫停／續播。這份公開INT介面用於解讀參數，
DQ3使用內嵌FMDRV的far call，不能據此宣稱兩個driver的全部內部行為相同。

r8沿r6的相同正常冷啟動，原版、dosgolem版本、seed1357與13項來源條件保持。
執行前核對凍結Go含API、資料及等待三個observer，首次等待後20000000指令有界停止。
獨立審查前198個queue／consume／capture、472IRQ1及396份完整PNG／bin逐項保持。
沒有還原state、寫guest state、改clock或增加AdLib參數；完整入隊仍DRAFT。

| 本次confirmed有限觀察 | IDA linear原始定位與讀值 |
|---|---|
| 初始化確實傳入狀態位址 | 22D50，BX0001、DX:AX=15ED:5C02；22D55返回AX0 |
| 正常入隊先暫停再重設 | packet199，22DFC／22E01的BX9返回AX0；22DAD／22DB2的BX8返回AX0，5C02由FF轉00 |
| 時脈與樂器設定返回 | 22DCD／22DD2，BX4、AX308C返回AX0，PIT除數65536→12428；22DA0／22DA5，BX2、CX1返回AX0 |
| 短曲播放請求成功返回 | 22D8A，BX6、DX:AX=3999:058B；22D8F返回AX0，5C02由00轉FF；本次事件前32bytes另存收據 |
| 等待仍未完成 | 首次等待後20000000指令，2157703538停208EF，2898=1、5C02=FF、474IRQ1；packet199未完成 |

上述返回值排除本次播放請求被拒絕。等待期間ticks14491→14573，但5C02仍FF；計時器前進不代表播放完成。
沒有觀察到10459續播、record540或下一次玩家操作，不能據此寫完整入隊READY或宣稱已修正remake。
接著只讀核對中斷向量與有限入口，若發現平台缺口，先依公開契約補足，再正常重播取得返回來源。

r7先因組裝語法錯誤在執行前失敗，修正後又漏插API observer。核對凍結Go後明確停止容器，保留中斷產物。
這是診斷腳本問題，不列產品缺陷。r8改為插入前斷言及執行前preflight，通過後才啟動原版。
產生器為`work/issue4-recruit-join-r8-builder.py`；先讀r7的Go literal，再組合既有r6／r5，所有來源都保持。
preflight為`work/issue4-recruit-join-r8-preflight.json`，source SHA256`cef4fc99cf8db7dc654f3be11d29336b8ee9bdc7101a9f7ec18f7c8be6adad72`。
獨立審查入口`work/issue4-recruit-join-r8-review.py`；收據`work/issue4-recruit-join-r8-review-receipt.json`，
SHA256`351ec454af7a6e8f33f8b0dcd4dbd25a0e0e400956be2a39917354d3a7ae291b`。
原版log／meta／source／binary在`work/dosgolem-opening/issue4-recruit-normal-r8-*`，各自雜湊由收據保存。
r9入口為`work/issue4-recruit-join-r9-builder.py`與同前綴preflight；只增加向量讀取，不改CPU或DOS服務。
公開r4 guard、r5／r6驗證工具與已接受來源不變，不把新DRAFT產物放寬成正式來源。

原廠SDK第2版僅作文件查找：[封存入口](https://www.ardent-tool.com/sound/Sound_Blaster.html)。
本機`work/issue4-fmdrv-sdk-reference-ctsbk2.exe`，1170683bytes，
SHA256`14261354702f0221e95d73f0c2e5d7dccbba620992a28d9da82c9497137e997d`；只讀取文件，沒有執行封存程式。
其舊CMF介面指向前版手冊，沒有提供本次根因契約，不當DQ3 oracle。
擷取文件留在`work/issue4-fmdrv-sdk-reference-*`，原廠文件與SDK內容不加入Git。
工作進度文字`work/issue4-recruit-join-r8-progress.txt`已登記Issue #4留言5974169729。

### 中斷入口與PIC有限平台實驗

r8補核對實際EBG.MCX，3369bytes，SHA256`ad4e139c5154b1271b2b6a6129a37684b019ea1e36ac6270ef24cc74127d5824`。
packet199的3999:058B前32bytes與檔案offset58B逐byte相同，不沿用舊C或Go音樂decoder作格式oracle。
補驗入口為`work/issue4-recruit-join-r8-review-r2.py`，收據SHA256`9e28a96c99857aeef0c177197c934b78b56faa0f20ebed3ee5c6856bc5333d6d`。
`work/issue4-recruit-join-r8-negatives.py`拒絕改seed、Go來源、PNG、播放返回值、等待狀態及preflight來源共六種損壞。
原始產物保持；收據`work/issue4-recruit-join-r8-negative-receipt.json`，SHA256`68bd8f0995b1bda363a34daf6533224fc618423b0b593406f178e2de2e23214c`。
首輪負測試未替換雙引號的preflight路徑，被測到仍在讀真實preflight，測試拒絕認列；修正隔離路徑後乾淨重跑。

r9只增加中斷向量及有限入口讀取。前198個packet、472IRQ1、396份PNG／bin保持，API／資料／等待觀察與r8逐行相同。
初始化與packet199的1Ch向量為10F1:0043，70h仍為0080:01C0。
ticks14444至14499共56次確實進1Ch入口；首次播放等待後14473至14499仍進入，5C02保持FF。
這排除該範圍完全沒有送入1Ch的解釋，沒有證實它能完成音樂。
步前observer可能漏掉硬體IRQ08的第一條指令，零筆08入口記錄不能解讀為沒有IRQ08。
入口`work/issue4-recruit-join-r9-review.py`及收據`work/issue4-recruit-join-r9-review-receipt.json`，
SHA256`bb2bacbe702ab51dba599e47f6050aa2925123c02ef879e8fd3a05facfa02313`。

查到dosgolem原生machine僅保存PIC遮罩，沒有in-service狀態，20h的EOI只保存埠值。
依[Intel8259A第14頁](https://www.pcjs.org/documents/datasheets/intel/INTEL_8259A_PIC.pdf)，
fully nested模式在EOI前阻擋同級及較低優先中斷，非特定EOI清除最高優先的in-service bit。
本次小程式明確重現EOI前第二次IRQ0，屬已證實的平台契約缺口，與入隊等待的因果仍未知。

可丟棄r11／r12副本在主PIC補IRQ0／IRQ1 in-service及非特定EOI，讓BIOS timer stub發EOI。
三項有限契約測試通過：IRQ0須等EOI、IRQ0可搶先IRQ1及BIOS釋放IRQ0。
此副本沒有完整PIC命令、slave、讀取介面或snapshot狀態契約，不是可發布的dosgolem修正。
上游與原版唯讀；初始化即固定此能力，不在流程中清旗標、完成byte或調clock。
r10測試漏設定首個到期時間，原版未執行；保留產物，修正fixture後另存r11。

r11的20M觀察仍未返回。52份前198個PNG／bin與舊工具不同，其餘344份相同；不得把它冒充既有未修改來源。
步數、timer、部分人物相位及AX有差異；實際變更欄位與52份檔名由`work/issue4-recruit-join-r11-observation-receipt.json`保存，
SHA256`bf5c72ffd5358e349518e104a0cd3e8ea218c14a0af55fd329cfd08fed7926ac`。
本輪曾把總埠計數增加當作音樂寫入前進，已訂正：總埠包含新增EOI，不能證明OPL進度。

事件格式只用於選觀察上限。[AdPlug的CMF／MIDI解析器](https://github.com/adplug/adplug/blob/master/src/mid.cpp)
提供變長delta與事件處理參考；本機有界解析81個事件、461個delta tick、offset6B0的FF2F00結尾。
以本次BX4除數12428及公開PIT參考1193180Hz，推得約4.801713秒，等級為hardware-spec approximation。
這不設定production音訊時間，也不稱原版逐週期或wall-clock parity。
入口`work/issue4-recruit-join-ebg-duration-draft.py`；收據SHA256`7a9e4ec6bd0de88d5aef327bb36ced0ff30ad784e4b8e8f6550ca8079d551761`。

r12在執行前固定首次等待後100M觀察上限，同一PIC副本、seed與clock設定；前198個packet及396份PNG／bin逐項等於r11。
tick已超過格式推算曲長，仍未進10459續播或下一操作。2237707546停208EF，packet199、2898=1、5C02=FF、474IRQ1。
因此PIC有限補足在本次長觀察仍不足以閉合等待，不認列為remake修法或完整原版來源。
獨立審查`work/issue4-recruit-join-r12-review.py`及收據SHA256`40819d2d7ad61bc19746efbd6e5705b8a720c7ce07726daec405244b9ab0020d`。
r11／r12產生器、平台before／after logs、三份凍結Go與platform preflight均在`work/issue4-recruit-join-r11-*`／`-r12-*`。
平台r11 preflight SHA256`65fb2488088ebfa1a87a121b019c1e9bcafec83ea073db2d6cb5570e09bc8f70`；各份凍結來源hash由該收據列出。
Issue進度`work/issue4-recruit-join-r11-progress.txt`對應留言5974288002。
下一步從未改PIC的工具鏈只讀核對OPL、PIC／PIT讀取計數及應用程式wrapper；r13入口為`work/issue4-recruit-join-r13-builder.py`。
不展開硬體driver／ISR，不因局部平台測試綠色就升格完整入隊READY。正式remake、pack及最近完整回歸保持。

### 未修改工具鏈的長觀察與下一個有限切片

r13回到未修改PIC的2f44a68工具鏈，固定首次等待後100M上限；seed及13項來源條件保持。
前198個正常packet、472IRQ1及396份PNG／bin逐項等於r5／r8。沒有將PIC副本的52份相位差異帶回正式來源。
十個只讀樣本的OPL寫入計數均為1235；PIC20讀取0、21讀取1、PIT40讀取0、61讀取474、OPL388讀取65，均不前進。
本次等待未持續讀取這些硬體埠；不能藉調61、PIT latch或PIC狀態查詢猜修法。
應用程式wrapper的linear22E40／22E4A沒有觀察記錄，功能仍unknown；沒有跟進driver或ISR。
最後2237703538停208F3，packet199、2898=1、5C02=FF、474IRQ1；完整入隊仍DRAFT。
入口`work/issue4-recruit-join-r13-review.py`與收據`work/issue4-recruit-join-r13-review-receipt.json`，
SHA256`761367b3e3ad8aa2186801402d765a83a167e42fcf0aa018fd4cc672a6b29586`。
原版log／meta／source／binary在`work/dosgolem-opening/issue4-recruit-normal-r13-*`，雜湊由該收據保存。

下一切片收斂到已觀察的packet197：record530、選人窗口及正常取消。這是既有入隊差異中先遇到的畫面阻塞。
沿10974計數、109E5逐人文字consumer及3E14原始窗口建立有限DRAFT，再由正常輸入驗證取消與下一步。
僅當選人窗口的資料、引用、geometry及正常玩家路徑足夠READY，才修正式pack與renderer。
不把尚缺正常播放完成來源的199後入隊流程混入此READY，後段音樂完成介面保持unknown並登記Issue。
本輪只保存有限證據，正式Go／pack及最近完整Go回歸保持，不新增發行包。
收尾入口為`work/issue4-recruit-join-r13-final-audit.py`及同前綴收據；遠端結果與核對另存同前綴result／post-push產物。

## 2026-10-04 選人清單與正常取消有限 READY

範圍為正常出生一名戰士男性後，從首次招募選單選入隊、顯示清單、Esc取消、
繼續詢問選No、告別等待及返回場景。完整入隊的536文字及199後播放完成保持DRAFT。
空名冊、滿隊、其他職業、多人成員及分離／查看沒有正常原版驗收，本節不宣稱它們完成。

輸入為`assets_raw/DQ3.EXE`，115282bytes，SHA256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
工具為dosgolem2f44a68及IDA Pro9.4。IDA位址為linear，file=linear−EC90，DGROUP基底24DD0。
新增來源`work/dosgolem-opening/issue4-recruit-selection-cancel-r1-*`。
獨立稽核`tools/verify_dosgolem_recruitment_selection.py`核對201個queued／consumed／capture、
478IRQ1及402份完整640×350PNG／bin；前196個packet保持，197畫面與r5逐byte相同。
收據`issue4-recruit-selection-cancel-r1-source-r1-receipt.json`的SHA256為
`ee4e519a5d61ccd05dcf82c6372a40257770182d733e7641632ad2accbe88e2d`。
原版seed1357在冷啟動前固定一次；沒有狀態還原或角色、旗標、完成狀態注入。

| 原始定位 | 已審查契約 | 等級 |
|---|---|---|
| 103E1..10413；DGROUP3E14，file19F54 | 530追加選人問題；10974計數未入隊slot；窗口高=(人數+3)×16 | confirmed，正常197 |
| 1F4E3..1F58F；原始窗口words | 保存／陰影、531兩行header、532每人一行、533footer、109E5callback、活動框、原始cursor | confirmed，原始consumer及197 |
| 109E5..10A7F；215EE..21651；21929..2197E | 姓名最多4字；rowY78、步16；姓名X168、等級五位右對齊起點216、職業X312、性別X408 | confirmed，原始byte單位及197 |
| 1040E..10413→10469→103AE | Esc不搬移角色，接540及兩項選擇，初始Yes | confirmed，正常198 |
| 103AE..103D6 | No接541、等正常確認、撤窗返回；Yes回10378重播528後主選單 | No confirmed正常199..201；Yes由後續cf0730f1正常204包確認，先前strong保留為形成史 |

DGROUP3E14保留原始32bytes：
`09030f002e002c0070001302050014021502e5090500030011004e0000000a03`。
531／532／533的終止由213C4使Y再加16；x為VGA byte單位×8。
109E5讀slot內+1職業、+2性別、+15等級；原始姓名consumer保留raw位置與字模流程。
原始窗口cursor為136,78；flags3、x15、y46、width44。這些值由bytes／consumer導出，不由截圖猜測。
IDA sidecar為`work/issue4-selection-r2-ida.json`，producer為同前綴`.py`；原始bytes與MZ重定位分別保存。
既有14筆入隊等待旁註只沿用其原有推論等級，新selection語意列於本表，不覆蓋原始名稱。

READY實作契約：必填typed `recruitment_selection`提供窗口、frame文字、row anchors、number field、
職業／性別text映射、cursor及點選範圍。缺欄位、未知引用、未審查evidence與越界均拒絕。
共用Go只執行選人文字、動態清單與取消具名狀態；不加入DQ3 raw record或座標常數。
取消、540、541及最後確認不改名冊、隊伍、金錢、旗標或RNG；文字EOF與下一選擇不共用輸入。
清單及取消需正常InputState重播、全幀PNG、不遮罩比較、同版本存讀檔及正常下一步。
完整V3若仍有背景相位差異則保持RED；內部測試不能升格整款parity。

READY補充：540後Esc沿1F8EA..1F907只設0726取消byte，沒有改0722游標；103B9只讀0722。
因此依當前Yes／No游標返回，不能直接把Esc映射No。此分支為strong靜態閉合，原版動態未抽測。
Yes後528使用明示`continue_text_id`，不從問候陣列猜選最後一筆。

重生入口：在既有測試image內，以唯讀repo、唯讀frozen dosgolem、UID1000可寫work執行
`python3 /repo/tools/dosgolem_recruitment_selection_probe.py`。
它的預設前綴與原始來源相同，先在乾淨輸出overlay重生，不能覆寫上述歷史檔。
再執行稽核器，傳入source root、實際producer及新收據路徑；原始producer為本機cancel-r1-builder。

### 正式實作與有限驗收

schema0.12.0／content0.1.84，canonical hash為
`sha256:0bdb4ebfa99c9f36fb5e7f8dd3f6d5f2e1dc3892900bdadc174cf6be34e47fcc`。
`game/recruitment_selection.go`執行typed pack窗口、文字、欄位與正常取消狀態機；
`internal/gamepack/recruitment_selection.go`驗證必填欄位、文字來源與窗口／行數邊界。
原始EXE／D3TXT parity與14種壞契約拒絕通過。正常取消不消耗RNG或改名冊、金錢、旗標。
既有入隊交易保留；536後文字與音訊未接入本批，不稱完整招募完成。

來源r1已通過獨立稽核，但缺少共用runtime比較器所需的頂層EXE身份。
另存`issue4-recruit-selection-cancel-r1-source-r2-receipt.json`，不覆寫r1。
runtime採用r2，其SHA256為
`d59315c19fc5de86cd900b7becdb8c0d3487c18bd3ed72177f4a788a21be2682`。
只增加三個頂層身份欄位，原版log、meta、201個packet及402份PNG／bin保持。
原版工具body已由公開`tools/dosgolem_recruitment_selection_probe.py`獨立保存，與實際私有producer逐byte核對。
重生來源因producer身份不同須另行審查，不能直接替換已固定的runtime oracle。

IDA公開重生入口為`tools/ida_dump_recruitment_selection_contract.py`，
自動合併`tools/ida_recruitment_selection_ledger.json`的11筆附加語意，10confirmed、1strong。
審查sidecar為`work/issue4-selection-reviewed-r1-ida.json`，SHA256
`11906b346f8a8271b3fb68a63da264ec1c11d2d978d29974c1baa39ed7ebbbcb`。
825筆唯一指令的file bytes、MZ重定位、loaded bytes及原始名稱／函式邊界均核對。
間接callback未被IDA自動定義為函式時保留null原始身分，不虛構函式名。
七種來源損壞拒絕：seed、狀態注入、producer、輸入scan、名冊觀察、PNG、bin。
獨立收據`work/issue4-selection-evidence-review-r3-receipt.json`，SHA256
`2de4e29044e41006d7856f65e941603e783db264781b14a655f61f280a167d3d`。

正常InputState依201個packet重播；兩側seed1357都在執行前固定一次，長路線骰序不作逐次一致聲明。
位置、section、金錢、旗標、名冊與選單游標通過；Talk保留現行正式命令窗額外一次確認。
最終runtime入口`work/issue4-selection-normal-run-r3.py`，產物在`work/issue4-selection-normal-r3/`。
完整RGB差異為197清單295、198取消詢問295、199 No游標295、200告別295、201返回0。
清單舊差51535已降低；舊入隊路線198的56790不同於本次取消198，不能混用。
52張PNG保持；197..200差分皆SHA256
`4261092cd10bde7005a010c8b4a03246f89e759033185adfaaa9c31f888a9d70`，bbox為288,131至510,262。
差分入口`work/issue4-selection-diff-r2.py`及同前綴receipt，差分PNG留本機。
正常Save／Load、下一步及壞Load／有效Load元件驗證通過；原版存讀檔未驗。
有限狀態E2／E3、畫面V2。完整V3仍RED，本節不升格整款CONFORMED。

完整回歸入口`work/issue4-selection-full-r1.py`；r2前17批與同產品r4最後5批合併439唯一頂層測試。
389頂層／88子PASS、50選用SKIP；internal157頂層／307子PASS、4選用SKIP、11套件，沒有素材缺失SKIP。
正常新遊戲至THE END66.58秒，只屬remake可玩回歸。desktop為Linux x86_64 ELF。
合併入口`work/issue4-selection-combine-r1.py`，收據`work/issue4-selection-combined-regression-r1-receipt.json`。
最終舊原版路線入口`work/issue4-selection-old-oracles-r3.py`，另有取消、出生、首次招募及Load獨立程序。

收尾失敗均保留：缺asset掛載／固定依賴快取、runtime r1缺頂層身份、首份差分工具缺Pillow、
完整主線舊Join操作未等新問題、新Load元件fixture處於標題而未消費modal、標題存檔正規化的錯誤全文比較、
IDA原始函式null的錯誤審查假設、負例修改到未比較的出生writer欄位。
負例已改為修改真正的名冊狀態觀察，且斷言輸入確有改動。
出生回歸兩次SIGKILL；20秒堆疊診斷定位到選人畫面仍等待場景cooldown。
招募modal暫停場景cooldown，測試等待條件補上recruit未啟用，再乾淨重跑。
這些驗證腳本問題沒有以調seed、原版完成byte或正式規則製造通過。

本輪腳本索引：`work/issue4-selection-docs-r1.py`、`-docs-results-r1.py`、
`-evidence-review-r1.py`／r2／r3、`-full-r1.py`／r3／r4、`-old-oracles-r1.py`／r2／r3、
`-birth-diagnostic-r1.py`。原版、IDA database、PNG、binary及使用者scratch均不加入Git。
下一閘門為共用295像素差異與536文字的有限規格；199後正常播放完成仍unknown，不重開硬體driver／ISR。

最後乾淨重跑：`work/issue4-selection-old-oracles-r3/`四個獨立程序全部PASS，
取消／出生／首次招募的85張PNG與前一正式入口版本逐byte相同。
`work/issue4-selection-normal-run-r4.py`使用最終binary重跑201包；收據及52張PNG與r3逐byte相同。
最終正常驗收收據`work/issue4-selection-final-normal-r1-receipt.json`，
同版本存讀檔與正常下一步通過。測試等待條件修正未改pack、正式程式或畫面。
本輪收尾與遠端核對入口為`work/issue4-selection-final-audit-r1.py`及`-post-push-r1.py`，
收據與Issue正文／結果文字留在同前綴本機工作目錄。

## 2026-10-04 共用人物差異與正常Yes續行 DRAFT

本段保留來源執行前的DRAFT紀錄。正常Yes的後續接受結果見下方「正常Yes來源與有限狀態驗收」；人物時序仍DRAFT。

從e259412正常201包來源重算完整RGB差異。195..200的295個位置完全相同，
主角區182、櫃台NPC區106、右下NPC露出區7；194未開窗完整411，201關窗後0。
這是完整畫布差異的定位分析，沒有裁切或遮罩驗收。人物相位不可指定或挑選結果。
分析入口為`work/issue4-295-pixels-r1.py`及同前綴receipt；原版與runtime PNG都保持。

正常新分支入口為`work/issue4-recruit-yes-r1-builder.py`。它在同一冷新遊戲與seed1357下，
沿正常出生、下樓、Join、Esc、Yes、再Join、Esc、No、告別確認返回。
只在既有11ED0／11EE8及1E2FF／1E30B的人物取圖consumer觀察原始暫存器、counter0002與phase0004。
不改原版bytes、phase、時鐘、角色、名冊、旗標或亂數。前198個正常來源預期保持，須獨立稽核後接受。
新產物在`work/dosgolem-opening/issue4-recruit-yes-r1-*`；正常Yes尚未動態接受，不改先前strong等級。
稽核入口沿用`tools/verify_dosgolem_recruitment_selection.py --continue-yes`，
204包、484IRQ1與完整408份PNG／bin的斷言須與實際正常來源核對。
工具新增路線不可放寬既有201包來源；舊收據須逐byte保持，並另驗證損壞來源拒絕。
工作登記Issue #4留言5975028362。圖像時序仍DRAFT；PIT／ISR停止線與199後音樂unknown保持。

公開重生入口為`tools/dosgolem_recruitment_continue_probe.py`，完整body與本次執行的私有producer相同。
在既有image、唯讀repo／frozen dosgolem、UID1000可寫work執行；預設前綴相同，須使用乾淨輸出overlay。
接受入口為上述稽核器，傳入實際producer及新收據路徑；不覆寫既有來源。
獨立核對入口`work/issue4-recruit-yes-source-r1.py`，正式正常重播`work/issue4-recruit-yes-normal-r1.py`。

### 正常Yes來源與有限狀態驗收

原版冷來源自然完成204個queued／consumed／capture、484IRQ1與408份完整640×350PNG／bin。
原始EXE身份5178fdc8、dosgolem2f44a68及seed1357在執行前固定一次；前198個packet與396份圖像保持。
來源收據`issue4-recruit-yes-r1-source-r1-receipt.json`，SHA256
`cf0730f10e48673e2da6702c77a6e458269cfe0153216b8770b7d3889a08e829`。
199選Yes後重播528及三項選單，200再Join、201 Esc、202選No、203告別等待、204確認回field。
這只把Yes正常分支從strong提升為confirmed；540內Esc尚未動態抽測，仍strong。
所有名冊狀態1、slot1的97bytes、金錢、故事旗標及原版角色資料保持；原版存讀檔未驗。

原版人物取圖130組：87組NPC的base80／66加原始phase1為81／67，43組主角僅保留原始BX4→000A。
不把主角的原始BX數值當成已審查的frame型別。195僅一次NPC取圖；196..203沒有新人物取圖。
原始consumer與位址基準沿既有IDA9.4 sidecar，不修改原始名稱、bytes或database。
同次remake主角與可見NPC的walk均0；兩側沒有同動畫時序契約，不能聲稱同相位。
完整195..203各差295、返回204差411。舊No返回201差0只限那次phase，不是普遍的返回V3。
共用295的位置分組為主角182、櫃台NPC106、右下NPC7；沒有遮罩或指定影格。
人物時序保持DRAFT；六次原始遊戲計數與硬體Ticks不能直接互換，舊原型反證與停止線仍有效。

正式正常204包入口`TestRecruitmentContinueDosgolemNormalInputComparison`，
使用`DQ3_RECRUIT_YES_ORACLE_DIR`與`DQ3_RECRUIT_YES_RECEIPT_DIR`，母親來源與素材設定沿既有正常201包。
測試固定原版收據hash，只讀取兩側人物資料；來源不在時標選用SKIP，不能當parity通過。
runtime收據`work/issue4-recruit-yes-normal-r2/recruitment-continue.json`，SHA256
`c3dab244f5dd4df5fefb36ab9138e8474f71364ea47291abc49b9ecceaed8958`。正常Save／Load與下一步PASS；全域骰序未對齊，Talk保留額外正式命令確認。
原版與remake有限狀態E2／E3、畫面V2，完整V3仍RED；沒有修改正式Go、pack或存檔格式。

受影響回歸入口`work/issue4-recruit-yes-regression-r1.py`，四條原版正常路線與七個招募元件案例PASS。
137張既有PNG逐byte保持。新增唯讀人物metadata的r2重跑55張PNG與r1逐byte相同。
最近完整engine回歸仍e259412的439項／11internal／desktop及THE END66.58秒；本批不重跑未變產品全套。
十種隔離損壞拒絕，入口`work/issue4-recruit-yes-negatives-r1.py`，收據SHA256
`0865a4acbfda5abe066a284482262bf4b6324f02713062f387f3689440fc5c25`。
來源與公開producer逐byte相同，原版收據未覆寫；新路線不能替代先前No或完整入隊來源。
本輪工作與結果追加Issue #4，收尾入口`work/issue4-recruit-yes-final-r1.py`及同前綴收據。
下一閘門是人物取圖時序與536有限文字；不為295像素調clock、固定walk、重擲或深挖ISR。

## 2026-10-04 入隊536第一個文字等待 DRAFT

本段保留執行前範圍；下方續驗已接受正常198及199包，整段正式流程仍DRAFT。

接續7f828e6及Issue #4。人物counter、共用phase及讀取位置已有有限證據，
本輪不重試由總Ticks推導phase的已否定原型，也不深入PIT／ISR。
正常待機時remake也會換影格，但主角與每個NPC各自計數；先前「只在移動時換影格」的進度說法不完整。
本輪先閉合已列出的536文字阻塞。

原版來源入口為`work/issue4-recruit-text-r1-builder.py`，
沿公開選人探針的相同冷啟動、seed1357、正常出生及下樓輸入，改以Enter選人。
第198包在record536第一個inline wait停止；不送199、不探查音樂完成、不注入遊戲狀態。
來源稽核入口為`work/issue4-recruit-text-r1-review.py`，
核對196包父來源、完整198包與396份PNG／bin、472IRQ1、名冊1→2及97bytes原角色不變。
既有r5的198包有限來源保留；新產物以`work/dosgolem-opening/issue4-recruit-text-r1-*`另存。
party副本、插名consumer及正式文字對拍尚待審查，本節不升格READY或完整入隊。
正常重製診斷入口為`work/issue4-recruit-text-r1-remake.py`及同前綴runtime收據。
診斷只從正式InputState重播，未修正式引擎或pack，不把缺少536的現況稱為parity通過。

### 536續文、537與538的有限正常來源及原型

原始EXE為115282bytes、5178fdc8；dosgolem2f44a68與seed1357執行前固定一次。
198包來源SHA256`917a24f6ed74baec905cc11b18ac19461b965e0fdec7c461fb7b0f853ad39834`，472IRQ1與396份完整640×350PNG／bin。
199包來源SHA256`a6abd32b4bb67830e0a7ea4bbfc8f571eb474932b41055e0539d8f085f4148be`，474IRQ1與398份完整PNG／bin。
後者沿前者正常送一次Enter，前198包與396份圖像保持；沒有新狀態還原或內容注入。
第198包536停216D8內文等待，名冊slot1狀態1→2；第199包續寫536尾文、537、538後自然到208E2。
第199包標為`audio_wait`，不冒充鍵盤等待、播放完成或場景返回。
slot1原97bytes、金錢及故事旗標保持。actor觀察只有128bytes，涵蓋第二party槽前29bytes，
其中首byte改為slot index1，其餘已見28bytes相符；不能稱完整97bytes party副本已動態審查。

| 原始IDA linear定位 | 有限證據與等級 |
|---|---|
| 10415→10473..104C3；10A80..10A9E | 未入隊slot搜尋、97bytes載入與copy、slot狀態2；名冊交易與正常198已證實。完整party副本強推論，仍缺全97bytes動態觀察 |
| 10418..10425；DGROUP259C | byte5077轉CX後供536插名；正常記錄入口與畫面已證實，selector實際值尚待觀察，不因原型顯示正確就升格完整語意 |
| 21651..216AF；215EE..21650 | 原始FFFb名字consumer，259C非0／7時以值減1、乘2查4F15，再加3讀名欄位；四字上限、四byte步進。強推論，本輪角色與主角都名0，尚未動態區分兩者 |
| 1042A..10459 | 正常199依序537、暫停／cue32播放、538、208E2等待；只閉合到等待入口，不稱音訊或後續返回parity |

有界IDA入口`work/issue4-recruit-text-r1-ida.py`；sidecar SHA256
`e38e80f05c4dc7553205fbdd50b1a8dc31a90750aae7d7a021fdf5f4e4cd15d3`，157筆唯一rows的file原始bytes核對。
工具IDA9.4、linear−EC90為file，DGROUP基底24DD0；原名、原始xref type與定位保持。
sidecar的script hash指向共用base；實際composer SHA256`5480aa0c231b6ca129d52f1a9b079b14d4d5a080a86709c099d3eee9f0bcba12`，兩者由收尾收據分列，不混為同一腳本。
本輪不改database、既有semantic ledger或正式pack。

現行正式InputState診斷為`work/issue4-recruit-text-r2-remake.py`，以7f828e6未改引擎／pack走正常198包。
交易已搬移一人，但仍停舊選人狀態，沒有536內文等待；完整RGB差3028。
`r4-remake`只補536後差594。差異位置證明增加的299在人物區，文字沒有新增差異；
`work/issue4-recruit-text-r4-pixels-receipt.json`保存全幀差分與位置分組，分組不是遮罩驗收。
`r6-remake`保留caller完整底圖後198差295；`r8-remake`續到537／538及音樂等待，199也差295。
正常來源、runtime與原型PNG已目視核對；三份最終runtime各48張150..197舊PNG逐byte保持。
同版本Save／Load及Load後下一步PASS只屬remake診斷；不能代替原版音樂等待後返回。

原型raw record、暫態stage與底圖buffer只存在一次性隔離副本。尚未接production，沒有新增正式schema或資料fallback。
READY前須補完整party／插名動態審查與音訊完成閘門；正式內容將採typed pack與text ID，不能把原型數字帶入共用Go。
音樂driver／ISR與逐週期研究停止線保持；本輪未追查或偽造播放完成。

公開來源重生入口：`tools/dosgolem_recruitment_text_probe.py`為正常198包，完整body與實際r1 producer相同；
`tools/dosgolem_recruitment_audio_wait_probe.py`為正常199包，只將r2 producer的私有父工具路徑改為前一公開工具。
沿既有image、UID1000、唯讀repo／frozen dosgolem與可寫work，在乾淨輸出overlay執行Python工具。
原始前綴與接受來源相同，拒絕覆寫；新producer身份不同須另存並由稽核接受，不能冒用舊收據。
稽核入口為本機`work/issue4-recruit-text-r1-review.py`與`-r2-review.py`。
八種損壞拒絕：seed、producer、末包scan、末record、末phase、roster、actor、PNG；
`work/issue4-recruit-text-r2-negatives.py`先後跑有效隔離正對照，來源與已接受收據保持。
更新checker只加199 actor保持斷言，checker hash變更由獨立稽核記錄，不覆寫原接受收據。

收尾入口`work/issue4-recruit-text-final-r1.py`及同前綴收據；文件更新入口`-docs-r1.py`。
r1 Python不支援tar filter、r3錯用城鎮bank、r5多層逃脫、r7不唯一替換的失敗均保留。
r4的594與r6／r8的295是真實完整差分，沒有放寬到「容許594」或指定人物影格。
本輪不重跑未變的正式產品全套，最近完整回歸e259412與7f828e6受影響測試保持。
Issue #4與Goal保持進行中；原版、PNG、binary、IDA database及使用者scratch不加入Git。

## 2026-10-04 入隊角色副本與插名有限觀察 DRAFT

接續ab2f8a0及Issue #4，以`work/issue4-recruit-party-r1-builder.py`正常冷啟動至199包。
只附加104B5／104B9／104BE／104C3角色副本及21651／21697／2169A名字consumer的唯讀觀察。
輸入、seed1357與208E2停止點沿上一輪；接受收據前核對398份PNG／bin及199包不變。
不改狀態、不演播放完成；插名及完整副本的結論等待實際收據，正式整段仍DRAFT。

音訊資料診斷入口`work/issue4-recruit-party-audio-r1.py`與同前綴收據、MIDI。
依既有cue32入口及公開MIDI／driver契約，先核對461個delta tick與單次事件完整性。
Roland音色沿既有轉檔設定，屬合成近似；原始FM聲音、原版播放完成與完整返回仍未知。

既有`munt-smf2wav`映像已不存在；沿[tools/Dockerfile.munt](../tools/Dockerfile.munt)重建，
固定原廠`munt_2_8_2`的commit`3b05ec276f9e605af86b0eaef7f5eda43477a31f`。
本輪映像為`dq3-munt:2.8.2-r1`，替代舊未鎖版建置；只用唯讀自有ROM與MIDI、UID1000寫work，
不把ROM或render加入Git。建置與轉檔收據沿`work/issue4-recruit-party-munt-*`保存。


### 本輪有限來源與同名假設勘誤

正常冷源seed1357、dosgolem2f44a68、原EXE115282bytes／5178fdc8保持。
接受收據`work/dosgolem-opening/issue4-recruit-party-r2-source-r1-receipt.json`，SHA256
`d0f6428dbc6f66b0887c3991c4bd17cb00be2825df4a53e1cf5bc049d806ed32`。
完整199包、474IRQ1、398PNG／bin逐項等於a6abd32b；只有唯讀log新增24筆觀察。
來源r1因引用未組合的hook在執行前失敗；r2修正名稱空間後乾淨重生。原始、seed、clock及輸入不改。

| 原始定位與資料流 | 等級與實際觀察 |
|---|---|
| 104B5；DS:520B→DS:50E2，CX61h；104B9；104BC／104BE | 正常copy的97bytes逐byte相同；再把首byte改為名冊slot index1，104C3後保持。主角及其他兩槽97bytes保持；限定正常單人入隊confirmed |
| 1041E→DGROUP259C=2；21651→21697→2169A→215EE | 536兩處與538共三處FFFb，都取party指標50E2，加3到名字50E5；confirmed本路線 |
| 1042A→10CEE；10CFB..10D0E→DGROUP259C；537同一consumer | 10CEE沿4F15指標表掃描原operand[SI-2]為0的槽，再寫其一基索引。本次selector1，指標507F、名字5082，為主角；有限consumer confirmed，未觀察的重排分支仍strong |
| 1EBD8..1EC33 | 此helper讀[SI-2]作265D cache索引，依[SI+38]及角色欄位更新圖片cache。它沒有寫[SI-2]，不能把該欄猜成健康值或宣稱已閉合其初始化writer |

先前r8隔離原型把537也用入隊角色名字，已推翻。當次兩人同名，295差分不能證明名字角色綁定正確。
`work/issue4-recruit-party-remake-r1.py`補本路線主角綁定；兩人不同名元件PASS，正常199包及同版本存讀檔PASS。
50張runtime PNG逐byte等於前一文字原型，198與199完整RGB各295；不同名元件不替代原版正常來源。
本輪仍未接正式引擎或pack，不把有限原型列CONFORMED。

稽核為`work/issue4-recruit-party-r3-review.py`；r2 checker的「三段皆入隊角色」斷言被真實537指標拒絕，保留失敗。
`work/issue4-recruit-party-r3-negatives.py`對同一嚴格觀察斷言做前後有效對照及八種損壞拒絕。
八種為copy第97byte、名冊index、截短97bytes、主角變動、537錯誤selector、536錯誤指標、名字payload及缺consumer。
這是新增名字／copy條件的負測試，不冒稱重跑全部來源或PNG負測試。

537補查IDA9.4 sidecar`work/issue4-recruit-party-ida-r5.json`，SHA256
`017e8afdfa5cb233ed889a2fbcbd456fdc963fb7b2ac0fc42bed6c9bb59059f7`；
cache reader sidecar`work/issue4-recruit-party-prefix-ida-r1.json`，SHA256
`4dda020fcffa1723e52a0dbeab6be2655e85c4544910786c182fbb119afdd94d`。
輸入為/tmp/guest.exe工作副本，原路徑assets_raw/DQ3.EXE保持唯讀；hash相同，linear−EC90=file，DGROUP基底24DD0。
原始bytes／xref及名稱保持，sidecar仍記base script hash；實際composer為`work/issue4-recruit-party-ida-r1.py`／`-r2.py`，各hash由final收據分列。
啟動缺sidecar不認列成功；最小ASCII probe確認工具可用，原生IDA log指出中文source的ASCII解碼錯誤。
重跑使用HOME=/home/ubuntu、TVHEADLESS=1、LANG=C.UTF-8、LC_ALL=C.UTF-8、PYTHONUTF8=1、/tmp工作目錄及原生-L log。
沒有TTY時的TVHEADLESS依[Hex-Rays批次契約](https://hex-rays.com/blog/igor-tip-of-the-week-08-batch-mode-under-the-hood)。
無DB參數的最小已驗命令重生一次性DB；DB與原生log不加入Git。

### 單次短曲資料與完成時長近似

cue32原版callsite及EBG第2零基項沿既有證據。完整事件範圍(file)58B..6B3，81事件、461個delta tick，FF2F00結束。
原版EBG3369bytes，SHA256ad4e139c5154b1271b2b6a6129a37684b019ea1e36ac6270ef24cc74127d5824保持。
舊`tools/cmf_to_midi.py`用軌內word@2猜起點並讀尾隨單byte delta，本項得79事件／660ticks。
既有第20軌OGG11.054281秒，不能當正常入隊單次音源。舊轉檔與其他場景本輪保持，不外推所有音樂皆錯。
新公開工具[tools/convert_midi_event_stream.py](../tools/convert_midi_event_stream.py)明示範圍、來源hash、driver divisor、平台reference及program map。
它讀事件前VLQ與running status，拒絕未知事件、越界、缺終止及覆寫，不猜軌檔頭。
公開重生與獨立私有解析逐事件相同，SMF SHA256
`fcfbecab3e603ae170d6254b31e0d2550874452dec78db210e3981a3a928b2c6`；PPQ48、tempo499961us，單次4.801708771秒。
原版參數12428及[公開SBFM clock契約](https://files.mpoli.fi/unpacked/software/texts/computer/inter56d.zip/interrup.p)
用461×12428÷1193180得4.801713069秒。VLQ與事件前delta依[AdPlug原始解析器](https://raw.githubusercontent.com/adplug/adplug/master/src/mid.cpp)。
本推算是hardware-spec approximation；沒有深挖driver／ISR，不稱原版wall-clock或聲波confirmed。

本機`work/issue4-recruit-party-audio-public-r1.mid`與同前綴receipt保存公開重生身份，不冒用私有producer。
在容器內執行公開工具的參數：原始EBG.MCX、--start0x58b、--end0x6b3、固定上述hash、--clock-divisor12428、--reference-hz1193180。
program map沿既有Roland合成設定48,32,19,49,73,46,71,11；它是音色近似，不是原版FM樂器證據。
實際參數需以分開的CLI token傳入；--output及--receipt指定尚不存在的work檔案。

Munt原廠tag為munt_2_8_2，實測smf2wav1.9.3、library2.8.3；tag與component版號分列。
使用唯讀work/music/mt32rom、上述單次MIDI，UID1000及--network none寫work。
參數-m/rom、-imt32、-p44100、-t；r2明示--record-max-start-silence=-1保留完整SMF起始靜音。
r1預設會移除起始靜音，長4.795850秒，作診斷保留；沒有裁切或覆寫r1。
r2為211761個立體聲PCM frame，OGG4.801837秒，與規格推算差0.124ms，保留工具量化近似。
既有game-video映像完成Vorbis編碼、ffprobe及完整解碼；mean−19.7dB、max−9.9dB，沒有靜音產物。
本輪沒有實機人耳或原版波形驗收。MIDI、WAV、OGG與ROM只留本機，不進Git或公開Issue。

來源公開入口[tools/dosgolem_recruitment_party_probe.py](../tools/dosgolem_recruitment_party_probe.py)，body與r2私有producer逐byte相同。
新公開身份重跑必須另存、重新接受，不能冒用d0f6428d。仍用既有DQ3容器、唯讀repo／frozen dosgolem、UID1000寫work。
本輪收尾入口`work/issue4-recruit-party-final-r1.py`及同前綴receipt；docs更新入口`-docs.py`。
下一步將有限文字、copy、逐段名字、底圖與單次音訊契約審成READY，再接typed pack及正式引擎。
原版199後返回仍未知；完成閘門只能使用已明示的平台與音源契約近似，不重新展開driver／ISR。

### 入隊文字及單次短曲有限 READY 審查

範圍是既有正常入隊交易後的536內文等待、537主角名字、538入隊角色名字、caller底圖保留及單次音訊。
正常199包與完整97bytes來源d0f6428d已閉合。主角固定第一槽的現行引擎可用具名primary_actor綁定537；
未觀察的原版隊伍重排與[SI-2]初始化語意不納入本規格，不將該欄命名為健康值。
536／537／538字模逐word由D3TXT00原始record生成並保留Go decoder作oracle。
短曲使用EBG.MCX已核對的(file)58B..6B3、81事件、461ticks；事件前VLQ及running status採公開AdPlug契約。
時長461×12428÷1193180秒採hardware-spec approximation；靜音或無音訊裝置仍等待相同時長，按鍵不能提前結束。
FM模式沿既有OPL2音色近似合成正確事件；Roland模式使用已驗單次OGG，兩者均不循環。
暫停場景音樂在537結束後、538開始前；短曲完成後恢復原場景音樂。
等待後的原始IDA linear10459→1045E→10469返回，再由10398→103AE追加540、103B6開Yes／No。
這段控制流使用既有selection-reviewed sidecar的原始bytes及已驗正常取消分支；音訊後續尚無dosgolem動態來源。
正式實作可接既有540 Yes／No狀態機，不能宣稱這段原版動態返回或全幀V3通過。
底圖在caller首次繪圖保存，文字與選單延續，modal關閉或有效Load才清除；拒絕Load保留。
資料包保存文字ID、名字角色、控制碼、原始音源範圍、clock及render身份，缺值／未知引用拒絕。
加入角色只交易一次，後續按鍵與播放不重複搬移或消耗亂數；同版本Save／Load須保留角色資料並清UI暫態。
此有限規格READY；完整原版播放後返回、聲音波形及動畫仍未CONFORMED。欄位入口為[docs/84](84-game-pack-json-contract.md)。


### 正式入隊文字與單次短曲有限驗收

schema0.13.0／content0.1.85，canonical hash19f6124c5f03c9f14c2909b94be5ac48438df956ea8b59e82af58c8168c558f6。
正式入口為`game/recruitment_join.go`；typed pack、驗證與原始EXE／TXT／EBG parity在`internal/gamepack/recruitment_join.go`及其測試。
名字元件驗證實際繪圖ops依序為入隊角色、入隊角色、主角、入隊角色，不僅檢查context；正常來源兩人同名不能取代此驗證。
FM新有界event parser保留EOT、事件前VLQ及running status，以累積有理樣本時點渲染單次音源，不改其他MBG歷史parser。
Roland讀取`recruitment_join.ogg`並核對size／SHA；兩種音源均不循環，原場景player暫停後續播，不從頭換軌。
音源無裝置、關閉或降級仍保留289更新等待；輸入不能縮短。原版後續未接受，近似等級保持。
有效Load結束播放並清暫態，拒絕Load保留等待及畫布；角色交易不重複，RNG／金錢／旗標保持。
正常199包正式InputState、播放後540→No→告別確認、同版本存讀檔及下一步通過。後199步只屬remake正常路徑，沒有原版對拍收據。
入隊50張、No52張、Yes55張PNG逐byte保持；正式198／199各差295，由原先3028下降，完整V3仍RED。
game443項覆蓋，395頂層／91子PASS及48選用SKIP；internal161頂層、11套件PASS及4選用SKIP，沒有素材缺失SKIP。
正常新遊戲至THE END107.13秒；desktop Linux x86_64 ELF通過。沒有新發行包或跨schema存檔遷移。

本機入口：`work/issue4-recruit-join-production-r1-data.py`及`-tests.py`保存資料與正常輸入測試生成過程，不能重跑覆寫已存在輸出。
排版保持入口`-format.py`證明9份JSON整理前後逐項語意相同；canonical hash保持。
`-r2-run.py`以有界Xvfb、每項獨立程序完成全套443，避免連跑多份Game累積記憶體。
最終嵌入排版r6正常路線保持；只新增不同名實際ops斷言後r7再跑受影響測試，正式Go及pack語意未再改動。
runtime及log在`work/issue4-recruit-join-production-r2-full/`與最終`-r4-focused/`；統計入口`-r1-regression-summary.json`。
收尾入口`work/issue4-recruit-join-production-final-r1.py`，保存文件、binary、資料包、來源及PNG稽核。
初次建置型別錯誤、測試(file)103AE換算筆誤及素材相對掛載已修正；r1 focused第三個重型測試被記憶體限制終止，屬驗證環境。
全套結果統計一度預設SKIP47而拒絕，實際是48；只修統計腳本，不放寬任何產品斷言或重跑挑結果。
原版素材、OGG、ROM、PNG、binary及database不加入Git；私有staging短曲入口見docs/84。
下一步繼續共用人物動畫及後續正常節點，已閉合文字不重開；driver／ISR停止線維持。

### 2026-10-04 人物差異的原始影格與底圖核對

依Issue #4留言5976184810，續查正式78d84b1的共用295像素差異。
公開入口為[人物差異驗證工具](../tools/verify_dq3_recruitment_sprite_raster.py)。
工具只讀原始素材、既有dosgolem來源及正式完整PNG，不改影格、不產生替代畫面。
本節證實差異來源，不把正式動畫規格升為READY，也不宣稱畫面V3。

| 原始輸入 | 大小 | SHA-256 |
|---|---:|---|
| assets_raw/DQ3MST.BLS | 115206 | a1a48eaf6c13ae73472d5ff77769fa538c19e048c24496d20218a89f3230f244 |
| assets_raw/DQ3MAN.BLS | 222726 | 823f57e0724e36ac8ed1aa472f59d2e8fb059f171e05e3aafc457e386f77158d |
| assets_raw/CTY00.DAT | 7546 | ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836 |
| assets_raw/DQ31.BLK | 65286 | 5996d95d743fb8e1fe8d3ad29c513f5241af2c3574cf9f652dce57ebd6ba3298 |

原版使用已接受的cf0730f1來源，dosgolem2f44a68、原始EXE5178fdc8、seed1357固定一次。
204個packet及其408份完整PNG／bin、原始log hash再次核對。收據normal_inputs=242包含前綴38次輸入，不能混作204個packet。
260筆原始取圖觀察保持phase0004=1；原始IDA定位沿11ED0／11EE8與1E2FF／1E30B。
只引用原始暫存器及先前已審查sidecar，不將BX或SI逕自改名為素材frame型別。
本輪沒有重跑原版、調seed、重建database或改動dosgolem。

| 玩家可見人物 | 原始素材frame索引 | 資料定位 | 無modal的差異 |
|---|---|---|---:|
| 主角 | DQ3MST.BLS 4與5 | 正常男性勇者、原始向上方向；取圖consumer為IDA linear1E2FF／1E30B | 182 |
| 櫃台NPC | DQ3MAN.BLS 200與201 | CTY00.DAT(file)0x88的7bytes，sprite key29、raw direction0 | 106 |
| 右下NPC | DQ3MAN.BLS 26與27 | CTY00.DAT(file)0x8F的7bytes，sprite key7、raw direction1 | 123 |

上述frame索引屬於素材檔索引，不是IDA linear或原版DGROUP快取指標。
每份BLS子影格480bytes，像素384bytes與遮罩96bytes；file範圍由工具逐份列入收據。
每個人物的768個像素連同透明處，均與CTY原始tile4及DQ31.BLK底圖完整核對。
色盤取原版bin與PNG的色號配對；沒有為未出現的色號補值。
完整無modal194與204的三個人物區皆吻合原版影格1、remake影格0，差異合計411。
櫃台外觀也能匹配素材432／433；CTY key29的載入契約限定200／201，不能靠外觀唯一性命名。

正式樣本取自`work/issue4-recruit-join-production-r4-focused/yes/`，schema0.13.0及canonical19f6124c核對。
完整640×350畫布逐點檢查194..204共11張。195..203的差異均為主角182、櫃台106、右下露出7，合計295。
每個不同像素的兩側RGB都吻合原始兩影格與底圖。已知影格範圍外沒有差異，未解釋差異為0。
這證實本次差異由人物取圖影格造成，排除本組圖像的色盤、遮罩及底圖問題。
正式PNG仍差295／411，未裁切或遮罩驗收；逐點解釋不等於對拍通過。

工具在dq3-ebiten-test:20260822-r1內以Python3.11、唯讀repo與UID1000可寫work執行。
CLI依序傳入`--assets /repo/assets_raw`、`--original /repo/work/dosgolem-opening`、
`--source-receipt /repo/work/dosgolem-opening/issue4-recruit-yes-r1-source-r1-receipt.json`、
`--runtime /repo/work/issue4-recruit-join-production-r4-focused/yes`與不存在的`--output /work/<收據名>.json`。
工具拒絕覆寫；既有來源、素材或pack身份不同須重新審查，不能放寬hash冒用來源。

本機收據`work/issue4-sprite-raster-r2-receipt.json`，SHA-256
`1ed19f6b9eefbd03ca02c2617a5c86c3d4ec1493160638c933d309f26d43b77e`。
六種損壞隔離副本全部拒絕：人物範圍外差異、錯誤人物顏色、錯誤runtime相位、缺packet、錯誤差異數、損壞原始素材。
前後正對照都通過。負例入口`work/issue4-sprite-raster-r1-negatives.py`，最終負例收據`-r2-negatives-receipt.json`，
SHA-256`70b2a267549d33f733eaf9042e66a859b4e576a6b7f8bb6ccfa697bad6ca7e54`。
初次探索誤用16px分組當24px人物原點及未出現色號7；改回正式TileH24與原始色號配對後重跑。
首版工具把normal_inputs242誤當packet204而拒絕，只訂正欄位契約，不改原版輸入或證據。

下一閘門仍是六次遊戲計數與逐consumer取圖的可比時序。先前總Ticks及整張一次latch原型已被反證，不能再用。
正式Go、pack及存檔格式本輪保持。最近全套回歸仍為78d84b1，不因診斷工具新增而重跑未變產品全套。
本輪不深挖PIT／ISR。其他人物、畫面、原版播放返回與存讀檔仍未接受。

### 2026-10-04 觀看名單的清單與取消切片 DRAFT

Issue #4 的下一個正常節點為選單第三項觀看名單。沿用新遊戲、登錄戰士男性與正常下樓的196包前綴。
原版由[觀看名單探針](../tools/dosgolem_recruitment_view_probe.py)冷啟動重生，
以[來源稽核](../tools/verify_dosgolem_recruitment_view.py)接受199包、474筆IRQ1與398份640×350完整PNG／bin。
本機收據`work/dosgolem-opening/issue4-recruit-view-r1-source-r1-receipt.json`，SHA-256
`043b39b187bfb869252923fbff103261099ce16748d24f1245c0754d7f1eef56`。
EXE仍為115282 bytes、5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c；
dosgolem固定2f44a68，兩側seed1357在執行前固定一次，不挑選亂數結果。

| 原始定位 | 附加語意與等級 | 證據 |
|---|---|---|
| IDA linear103A8 → 10624，file1718 → 1994 | confirmed：主選單第三項進入觀看名單 | 原始call、動態packet199的10624入口、DS15ed |
| IDA linear10624 → 10974；10632／10635／1063F；DGROUP3E14 | confirmed：列出未入隊角色，使用既有選人視窗及動態列數 | 與入隊相同loader及consumer，packet199一列，raw slot1未變 |
| IDA linear10646 → 1F4E3；10649..1064E；103AE → 540 | strong：Esc取消清單後接繼續詢問 | 既有選人鍵盤consumer與原始caller；獨立View取消動態收據尚待稽核 |
| IDA linear10668..10692 → 1834E、210EC | DRAFT：選定角色後顯示詳細狀況，另有兩個正常讀鍵等待 | 原版205包探針已產生，尚未獨立接受與釐清全部消費順序 |

IDA9.4 sidecar在`work/issue4-view-r1-ida.json`及`-r2-ida.json`。
位址基準為IDA linear；MZ file=linear−EC90，DGROUP基底linear24DD0。
r2腳本SHA為f32537fe1f5cbaf13a07d92707e0cad13c453d91145f2327387d6bd76b82e65a；
393條指令的loaded bytes與MZ relocation預期逐條吻合。保留原始名稱、raw bytes與xref type，未rename。
未分級sidecar候選保持unknown，本節只附加以上已審查定位。

修正前正常InputState的197／198各差295，199清單差51477。
可丟棄原型使用既有pack選人視窗、只讀roster、方向鍵移動、Esc接540；199完整RGB降為295。
原型及修正前工具在`work/issue4-view-r1-baseline.py`與`-prototype.py`，PNG及收據各在同名前綴的`-receipts/`。
沒有裁切、遮罩、調seed或覆寫動畫影格。295仍屬共用人物影格差異，V3未通過。

取消原版生產器為[正常取消探針](../tools/dosgolem_recruitment_view_cancel_probe.py)，
接受工具為[取消來源稽核](../tools/verify_dosgolem_recruitment_view_cancel.py)。
兩者沿用完整199包，追加Esc、右、Enter與告別確認，要求角色、名冊、隊伍指標、金錢與旗標保持。
[詳細狀況探針](../tools/dosgolem_recruitment_view_detail_probe.py)只保存DRAFT來源，不宣稱其205包已接受。
本節不接受詳細狀況窗、K改名、空名冊訊息、所有職業／裝備、完整View或原版存讀檔。

#### 清單與取消有限 READY

獨立取消來源已接受：203個packet、482筆IRQ1、406份完整PNG／bin，normal_inputs241含前綴38次。
收據`work/dosgolem-opening/issue4-recruit-view-cancel-r1-source-r1-receipt.json`，SHA-256
`cf23fcf92bbd599feb8a2bf1a2b6d7092e2450d187cb36a570798e899eb355bd`。
前199包的輸入、消費、狀態與398份畫布逐項保持；200取消接540、201右移No、202接541等待、203返回1997C。
名冊狀態、slot1的97bytes、隊伍指標、主角128bytes、金錢與旗標全程不變。
因此上表Esc到540及No告別返回分支升為confirmed。metadata參數只容許輸出前綴換名，其他引數、kernel及原始輸入核對。
首次稽核把輸出檔名前綴也要求相同而拒絕；訂正為逐引數的明示前綴替換，未放寬其他來源驗證。

有限規格：View入口直接顯示roster清單，不追加入隊530問題；列數、姓名、等級、職業、性別、游標及外框沿用typed `recruitment_selection`。
方向鍵沿原始單欄選人consumer移動；Esc先恢復caller畫布，再追加既有540，No追加541並另等一次確認返回場景。
不搬移或新增任何角色，不消耗RNG或金錢、不改旗標。存讀檔沿既有同版本UI清理契約。
本有限分支READY，允許正式實作及正常203包驗收；選定角色後的詳細狀況仍DRAFT，不能宣稱完整View。

#### 正式清單與取消驗收

正式`game/recruit.go`直接用既有typed清單顯示roster，方向鍵以roster長度導覽，Esc接540。
沒有新增DQ3 raw ID、座標、文字、數值fallback或JSON欄位；schema0.13.0／content0.1.85與canonical19f6124c保持。
`TestRecruitmentViewDosgolemNormalInputComparison`從新遊戲以正式InputState重播203包。
逐包位置、場景、游標、名冊、隊伍與持久狀態通過；197..203的完整存檔snapshot與196相同，包含RNG。
同版本Save／Load後下一個正常移動通過。混合名冊／隊伍元件驗證方向鍵只在未入隊角色間循環，取消不交易角色。
詳細狀況未READY，確認角色仍保留舊行為，不納入此有限驗收。

| 正常樣本 | 完整RGB差異 | 結果 |
|---|---:|---|
| 197／198主選單 | 各295 | 與修正前保持 |
| 199觀看名單 | 51477降為295 | 正式選人視窗、姓名／等級／職業／性別吻合；仍有動畫差異 |
| 200取消／201 No游標／202告別 | 各295 | 完整畫布差異逐點等於既有招募人物差異 |
| 203場景返回 | 411 | 完整畫布差異逐點等於既有無modal人物差異 |

七張完整640×350的每一個不同像素與兩側RGB均比對；沒有新差異，也沒有裁切、遮罩或指定phase。
54張正式PNG留在`work/issue4-view-production-r1-full/view/`。
前196包對應47張runtime PNG逐byte保持；既有入隊50、取消52與Yes55張也全部保持。
獨立核對入口`work/issue4-view-production-r1-audit.py`，收據同名前綴`.json`，SHA-256
`cc7a6e46366bc5fd4019afdee7cd9c8724db12e52ec3508b47eb0223f70d7b02`。
有限狀態E2／正常路徑E3、畫面V2；完整V3仍RED，完整View仍DRAFT。

六種來源損壞全拒絕：錯scan、錯consumer、錯record、名冊改寫、缺PNG及bin損壞；前後正對照通過。
入口`work/issue4-view-cancel-negatives-r1.py`，收據`work/issue4-view-cancel-negatives-r1-receipt.json`，SHA-256
`d31162a2a9d7d1187b0acd44d6f94c6ff2c28598a9cae9e4acac54a0964a190f`。
父來源先完整接受並唯讀固定；負例只修改新取消尾端的隔離副本，不改原版、kernel或seed。

game445項覆蓋，398頂層／93子PASS、47選用SKIP；internal161頂層、11套件PASS、4選用SKIP。
新View及既有招募正常路線沒有SKIP，沒有素材缺失SKIP。正常新遊戲至結局`TestOpeningProductionInputTrace`79.95秒PASS。
desktop以`go build ... main.go`產生Linux x86_64 ELF；使用者`tmp_dump.go`保持，不能以根套件納入它。
本批診斷腳本參數／收據路徑及桌面入口筆誤均為驗證問題，已訂正後重跑，未列產品缺陷。
檢查入口`work/issue4-view-production-r1-checks.json`；容器全數清理，root-owned基線3213及零`.md`目錄保持。
原版素材、PNG、binary、database與私有收據未加入Git，沒有新發行包。
下一步獨立接受詳細狀況205包，釐清兩個讀鍵等待與窗口恢復；不深挖硬體driver／ISR，不重開已完成清單與取消。

### 2026-10-04 詳細狀況兩次等待勘誤與有限 READY

原版來源由[詳細狀況探針](../tools/dosgolem_recruitment_view_detail_probe.py)正常冷啟動重生，
由[詳細狀況來源稽核](../tools/verify_dosgolem_recruitment_view_detail.py)獨立接受。
收據`work/dosgolem-opening/issue4-recruit-view-detail-r1-source-r1-receipt.json` SHA-256
`e61060c7db7007c5a76e3790f4d55cc2c252619fd04c5cf572f615bc8fb69ee0`。
205包、486筆IRQ1、410份完整PNG／bin與243次正常輸入，seed1357固定一次；前199包保持。
原始EXE仍為assets_raw/DQ3.EXE、115282bytes、SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
dosgolem固定2f44a68，原版及上游唯讀，未注入角色或改CPU狀態。

先前Issue留言5976684795與進度訊息將第201包判成黑色撤窗，已推翻。
200與201的PNG逐byte相同，色號及640×350 RGB亦相同；PNG SHA-256
`f100e33c9ce81e604c560bc820745c5f5c9d29bfb5eb0733f96f10a2c2258593`。
唯讀控制流重播`work/issue4-view-detail-r2-flow-probe.py`與獨立核對`-r2-audit.py`證實：
原205包、IRQ1、名冊、詳細觀察及410份完整產物都保持；不再調查未成立的像素或硬體缺陷。
勘誤已追加至Issue留言5976740421，原錯誤形成史保留。

| 原始定位（IDA linear） | 玩家狀態與等級 |
|---|---|
| 10668／10671／1067B → 1834E | confirmed：選取名冊角色的97bytes複製到DGROUP520B，4F1D指向此副本，722=5；第一能力窗使用DGROUP3DA8 |
| 1834E → 18498 → 2111B，停21133 | confirmed：能力及裝備顯示後第一次讀鍵等待；第200包 |
| 2113A → 1849D → 184A1 → 184BF → 184A0 | confirmed限定本角色：已學咒文計數四byte為零，沒有咒文窗；第201包 |
| 10681 → 210EC，停21103 | confirmed：第二次讀鍵等待，能力窗完整保留；第201包 |
| 2110B → 1068E → 1F604 | confirmed限定Esc：第二次讀鍵後恢復能力窗底圖；第202包 |
| 103AE → 540；No → 541 → 1997C | confirmed：接繼續詢問、告別等待及場景返回；第202..205包 |

IDA9.4原始bytes／xref sidecar沿`work/issue4-view-r2-ida.json`，補充`work/issue4-view-detail-r1-ida.json`。
補充204條指令與MZ relocation預期逐byte相符，腳本SHA-256
`1b76c99d3d6bf35bbc275a15a370e38a98197919b35e53c5d4dfb0af98f12ec7`。
位址空間仍為IDA linear，file=linear−EC90，DGROUP基底linear24DD0。
保留原始名稱、範圍、bytes與xref；未分級helper語意不作證據。

有限規格READY：正常View確認後顯示所選未入隊角色，第一次新按鍵只前進到第二等待，畫面不變。
第二次Esc撤窗，再接既有540／541狀態機。Enter／Confirm／方向鍵也由已證實的raw讀鍵與非K分支接續，
只有未映射的AnyKeyEdge不在第二等待猜成關閉。K改名與非零咒文的詳細分支仍unknown，不實作猜測。
實作只讀角色，姓名、職業、性別、等級、當前HP／MP、七能力、攻防與經驗沿既有typed能力幾何及原始1834E欄位。
本輪限定沒有已學咒文、沒有異常狀態、裝備仍等於pack登錄初裝的角色；其他角色確認保持清單，不改持久狀態。
初裝形狀與cloth文字沿已驗證的game-pack及原始裝備oracle，不新增Go版號專屬值或fallback。
現行初裝沒有武器，攻擊值為力量；本輪不把此等式外推裝有武器的角色。
pack及存檔schema保持，modal暫態不寫入存檔；拒絕Load保持畫面與等待，有效Load清除UI後可正常續行。

可丟棄正常205包原型`work/issue4-view-detail-prototype-r1.py`已完成兩次等待及返回。
第200／201完整RGB差從43708／40204降為各145，第202..204各295，第205返回411。
兩側角色HP／MaxHP分別11與13，是既有未對齊的全域骰序輸入差異；不能改角色數字湊圖。
正式驗收須保留此正常路徑差異，另以原始97bytes解出的相同角色元件fixture核對能力欄位與完整畫布。
元件fixture只驗renderer，不替代正常205包與存讀檔。兩次等待間PNG須相同，持久snapshot含RNG全程不變。
原始資料→typed pack／角色→只讀能力renderer→正常InputState→Save／Load→下一步均需通過。
有限CONFORMED只涵蓋已驗分支；完整V3、改名、已學咒文、其他裝備與原版存讀檔仍未完成。
原版資產、PNG、binary、database與收據留本機，不公開散布。

來源稽核在既有`dq3-ebiten-test:20260822-r1`容器、唯讀repo及UID1000可寫work執行：
`python3 /repo/tools/verify_dosgolem_recruitment_view_detail.py /repo/work/dosgolem-opening /repo/tools/dosgolem_recruitment_view_detail_probe.py /repo/tools/dosgolem_recruitment_view_probe.py /work/<新收據名>.json`。
原版205包需由上述dosgolem探針本身重生；稽核拒絕覆寫與不同父來源，不把輔助執行器畫面登錄為dosgolem。

#### 正式詳細狀況有限驗收

正式入口`game/recruitment_view.go`重用受驗證的能力窗，顯示所選名冊角色的只讀副本。
第一次新按鍵只進第二等待，兩次畫面逐byte相同；第二次Esc撤窗再接540、No、541與返回。
不改角色、隊伍、金錢、旗標或RNG。未READY的角色保持原清單與游標，不猜裝備或咒文。
`TestRecruitmentViewDetailDosgolemNormalInputComparison`正常205包、同版本存讀檔及下一步PASS。
Save的snapshot本身沒有RNG欄位，現已另以`g.prng`逐包核對；較早「snapshot包含RNG」措辭在此訂正。
有效／拒絕Load在兩個等待點抽測，held方向不消費新按鍵，未知AnyKeyEdge不猜成K改名或關閉。
舊`TestRecruitViewReturnsToMenu`斷言已被原版推翻，改驗第二個名冊角色實際姓名及HP綁定與名冊保持。

正常200／201完整640×350 RGB各差145；其中138位於HP／MaxHP兩欄，兩側實際輸入為11與13。
另用原版97bytes解出的相同角色作renderer元件fixture，完整RGB差7，逐點等於既有露出的右下NPC差異。
原版與正式正常角色資料不改，元件PNG另命名；未裁切、遮罩或指定動畫phase，完整V3仍RED。
197..199及202..204各295，返回205為411；全部差異逐像素核對，沒有新增未解釋差異。
新正常樣本56張PNG，加一張獨立同角色元件PNG；五條既有路線258張逐byte保持，新路線前綴50張保持。
獨立全畫布入口`work/issue4-view-detail-production-r1-audit.py`，收據SHA-256 `208f0f0bf88c3e5aecfee546e4fdf9b243ec05246c156becd761610fe684cab5`。
本有限狀態與正常返回CONFORMED；完整View、K改名、已學咒文、其他裝備與原版Load仍unknown。

來源八種損壞全拒絕：錯第二consumer、缺第二等待、錯record、角色副本變動、名冊變動、缺PNG、bin損壞與第二能力畫面變動。
入口`work/issue4-view-detail-negatives-r1.py`，前後正對照PASS；收據SHA-256
`5990b54881cbfddfd411f4e76562196fa2ea55bcbe168a8bed3cad76d52057ab`。
來源接受收據保持e61060c7；新增同畫布斷言後checker hash另存負例收據，未覆寫已接受來源。

完整game449項由r7已過326項及r8其餘123項合併，402頂層／101子PASS、47選用SKIP。
r7與r8正式產品相同，r8只訂正被原版推翻的舊測試；不重跑已通過且未變的正常結局121.16秒。
全部11個internal套件、162頂層PASS、4選用SKIP及desktop Linux x86_64 ELF PASS，沒有素材缺失SKIP。
九份JSON與乾淨9f81ddb archive逐byte相同，schema0.13.0／content0.1.85／canonical19f6124c保持。
檢查入口`work/issue4-view-detail-r1-checks.json`，SHA-256 `a810fd7ead661b6aa8c6ee18e285793919345e2274dc916d34c924b0c3ce3d45`。
新Go只有具名UI狀態及typed引用，沒有新增DQ3 raw ID、座標、record、旗標或玩家句子。

建置型別與saveState欄位筆誤保留；元件fixture未安裝NPC交談繪圖器、標題未建角主角Load正規化屬驗證前提問題。
收尾初版只找data八份JSON而漏manifest，改核對pack完整九份；不改資料或放寬語意檢查。
可重生入口為`work/issue4-view-detail-production-r7-run.py`與`-r8-run.py`，日誌及PNG在各自`-full/`。
Docker一次性容器清理、UID1000寫入、root-owned基線3213、零.md目錄與使用者13項資料保持。
原版、PNG、binary、database及私有收據不加入Git；沒有新發行包。Issue #4與Goal保持進行中。

### 2026-10-04 K 改名入口 DRAFT

本切片接在已接受的詳細狀況第二等待。正式程式仍為 `28690cc`，
schema0.13.0／content0.1.85 保持。以下靜態結論尚未當作完整玩家路徑證實。

IDA Pro 9.4 的非破壞匯出在 `work/issue4-view-rename-r2-ida.json`，
SHA-256 `ad8fee33b365c1ebd6174387ba36fd1e9c554a08247b53384b5b34825ee3576c`。
321 筆原始指令與 MZ relocation 後的 bytes 全部符合；腳本 SHA-256
`7449809f01391a4faadaf9850c94ab36694f9916cf1d3f4358f63ac066d1c981`。
輸入為 `assets_raw/DQ3.EXE`，115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
位址空間為 IDA linear，file = linear − 0xEC90，DGROUP 基底 linear 0x24DD0。
原始名稱、bytes、xref 與範圍保留，database 及原版資料不加入 Git。

| 原始定位（IDA linear） | 分級與候選語意 |
|---|---|
| 10686 `80fc25`，1068B `e81100` | strong：第二等待收到 K 掃描碼 25h 時呼叫 1069F |
| 106A8 → 1885F；1885F..1886F | strong：重新依 DGROUP 5077 隊伍人數選人，單人隊伍將 DGROUP 722 設為 1 後直接返回 |
| 106B3..106BF | strong：依選擇索引取 `[bx+4F15h]` 角色指標，再呼叫 10D17 共用姓名輸入；不能以名冊的 520B 副本代替目標 |
| 10D17..10D45 | strong：清空新姓名狀態，開共用姓名窗及輸入器 |
| 10D84..10D99 | strong：完成要求非空姓名，重算並寫入 DGROUP 270E 字串長度 |
| 106C2..106DB | strong：取消則返回；成功將 DGROUP 270E 的 18 bytes 複製到選中角色指標 + 3 |
| 1068E..10695 | strong：改名返回後恢復能力窗底圖，再回到既有 caller |

尚待正常動態閉合：實際改名目標、成功與取消交易、返回畫面與後續存讀檔。
多角色隊伍的選人窗尚未 READY。不得因觀看的是未入隊角色，就推定改名也修改該角色。

兩份 DRAFT 正常來源分別由 `work/issue4-view-rename-r1-producer.py` 與
`work/issue4-view-rename-r2-producer.py` 在既有 `dq3-ebiten-test:20260822-r1` 容器執行。
沿用唯讀 dosgolem `2f44a68`、原版冷啟動及一次固定 seed1357，不注入角色或更改時鐘。
第一份停在 K 後的姓名輸入；第二份沿已驗證的出生姓名鍵序完成，再選 No 返回。
入口稽核為 `work/issue4-view-rename-r1-validator.py`；接受前須核對原先 201 包、
IRQ1、完整 640×350 PNG／bin 與角色資料。工作及來源索引見
[Issue #4 留言](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5977022721)。
原始圖像、執行檔與收據留本機；本節狀態為 DRAFT，尚未宣稱 remake parity。

#### 單人隊伍改名有限 READY

原版入口202包、480 IRQ1與404完整PNG／bin已接受，前201包保持。
入口收據 `work/dosgolem-opening/issue4-view-rename-normal-r1-source-r2-receipt.json`
SHA-256 `fe3df1165b4d7e27c1f5f2701ec59bfe3d078f9b20ea0f8ab9c4b2556d805f9c`。
相同姓名完成及正常No返回217包已接受，434完整PNG／bin、510 IRQ1；
首版交易收據 SHA-256 `631414a8c988b465f1c5768fd20c9fea6dee8cf04a1ebfa471da7295296ace9d`。
姓名功能列取消及No返回212包已接受，424完整PNG／bin、500 IRQ1；
收據 `work/dosgolem-opening/issue4-view-rename-normal-r4-source-r2-receipt.json`
SHA-256 `c139c899b83f915e19c6b395845f6d42adb207142ba77c699f70e1c49ae1d6dc`。
兩者前202包、名冊97bytes、旗標、金錢、位置、場景及全部角色指標保持。
稽核入口為 `work/issue4-view-rename-r1-validator.py` 與
`work/issue4-view-rename-tail-validator.py`，保留先前收據，不覆寫歷史。

上述10686至106DB的有限單人分支升為 confirmed：
入口游標5，1885F後改為1，106BE的DI為DGROUP507F，重新選中主角。
成功依序經106C2、106CB、106D8、106DA、1068E；18bytes目標從507F+3至507F+3+18。
取消在106C2觀察到DGROUP726=1，未經106CB／106D8／106DA，直接回1068E。
眼前能力頁的520B副本及未入隊535E角色不變。之後均接540、No、541及1997C。
姓名非空閘門沿同一10D17輸入器及既有已驗證的姓名契約，不新增文字或視窗猜值。

READY範圍為單人隊伍，在第二等待以具名Rename輸入開空姓名窗。
第一等待的K只消費第一次讀鍵，不同時開改名。姓名完成只修改主角姓名並同步對話姓名引用，
取消與空姓名完成不得寫角色；兩者返回既有540流程。
姓名窗保留正在觀看的能力頁與caller底圖，沿既有game-pack姓名幾何、文字與點陣renderer。
多角色隊伍尚未READY，保持第二等待，不猜選人窗。pack缺引用或設定一律拒絕。
暫態不存檔；有效Load清除改名UI，拒絕Load保持輸入，成功姓名需同版本存讀檔與下一步驗證。
有限交易還需不同姓名的正常來源、正式InputState路徑與全畫布驗收後才宣稱CONFORMED。

隔離入口診斷 `work/issue4-view-rename-prototype-r2.py` 保留兩側原本角色，
正式未修正入口差11926，共用姓名窗prototype差76；沒有角色注入、裁切、遮罩或調動畫相位。
原版與remake的HP輸入差異保持。驗證首版收據缺頂層原版身分欄位，另存補齊欄位的r2，
未放寬畫面驗證；該失敗屬收據格式問題。
不同姓名首版r3停在姓名輸入，原因是相鄰字格回功能列少一次左鍵；不列成完成來源。
修正r5以同一初始seed與明示鍵序重跑，未重擲或挑結果。取消來源不受此腳本問題影響。

#### 公開改名重生與驗證入口

[正常改名探針](../tools/dosgolem_recruitment_view_rename_probe.py)支援三條受驗證的正常輸入：
`--route 2`為相同姓名成功，`--route 4`為功能列取消，`--route 5`為不同姓名成功。
英數格 raw1 的字模為5，來自既有rec453欄優先語意表；不能把 raw index 1 當字模1。
不同姓名來源 `work/dosgolem-opening/issue4-view-rename-normal-r5-source-r2-receipt.json`
SHA-256 `0d43e2bc421854b87dc0dd2c2fd177c9eca4bdfaa2a1609bfa5e4099ef437cef`，
219包、514 IRQ1、438完整PNG／bin，前202包保持。主角只有原始97bytes的+7從0改5，
完整128bytes其他欄位、名冊、角色指標、旗標、金錢、位置及場景保持。
相同姓名新版來源SHA-256 `7cf247fd19a34022a4507fb404eee6325d4654281ba507cb60baeffe409c512d`，
補齊頂層原版身分欄位，既有631414a8首版保留。

在既有Docker image、UID1000、唯讀repo及dosgolem2f44a68、可寫work執行：
`python3 /repo/tools/dosgolem_recruitment_view_rename_probe.py --route 5 --prefix <新來源前綴>`。
`work/dosgolem-opening`沿用母親完成的已接受父收據9358ce6e，工具拒絕既有前綴，
seed1357在執行前固定一次。每個正常按鍵只在原生等待排入IRQ1；角色觀察唯讀。
不同姓名已由原版正常完整流程證實，不能以未完成的r3來源取代。

[正常改名來源稽核](../tools/verify_dosgolem_recruitment_view_rename.py)核對父來源、
producer及frozen Go／binary身分、完整按鍵序、IRQ1、每包等待點、全PNG／bin及角色交易：
`python3 /repo/tools/verify_dosgolem_recruitment_view_rename.py /work/dosgolem-opening <來源前綴> 5 <實際producer路徑> <新收據路徑>`。
原版收據由dosgolem自行重生；稽核本身不執行或修改遊戲，拒絕部分來源與覆寫。
公開契約在[docs/84](84-game-pack-json-contract.md)；原版素材、圖片、binary與收據不公開。

#### 2026-10-04 單人隊伍改名正式驗收

依上述原版三路證據，有限READY已接入正式typed pack及InputState，
schema0.14.0／content0.1.86，canonical hash
`sha256:e274124c2eebdb03a36a6da5e4a1110e1ea82a182699619e336bba2aeea25ec7`。
第二等待的Rename開既有空姓名輸入器，第一等待的K只消費第一次讀鍵。
非空成功只改主角姓名並同步對話引用；取消、空名與未READY多角色不寫入角色。
成功與取消接540、No、541及正常場景返回。設定、幾何、文字全由pack引用，沒有新增Go原版raw ID或玩家句子。
本節CONFORMED僅限單人、無已學咒文、無異常、pack初裝角色及受驗證的三條正常路線；狀態E2、流程E3、畫面V2。
多角色選人、已學咒文、其他裝備、原版Save／Load及完整View仍未知，完整V3仍RED。

正式三路217／212／219包均從新遊戲及正常玩家輸入重播，
逐包持久snapshot與另行RNG比較、名字窗游標／模式／長度、成功與取消交易、
同版本存讀檔及Load後下一步通過。有效Load清姓名暫態，拒絕Load保留；空名與容量閘門元件通過。
不同姓名正式字模為5，原版97bytes只有+7由0改5，未入隊角色保持。
正常測試每路分開程序，沒有座標注入、直接事件函式或debug入口。

新正常PNG同名68、取消63、異名70，共201張；各自前52張逐byte等於詳細狀況舊樣本。
六條既有路線315張PNG保持。全部640×350 RGB核對，未裁切、遮罩、指定phase或替換角色。
姓名窗完整差11926降76；69為HP數字、7為既有NPC影格，
差分位置及雙側RGB逐點屬於先前能力窗145差分，沒有新增未解釋差異。
能力200／201仍145，caller各295，同名返回217為411；取消212與異名219本次返回完整RGB零差異。
返回零差異僅對本次樣本成立，不外推動畫或完整流程。兩側正常HP11／13不改值湊圖。
全畫布入口 `work/issue4-view-rename-production-r1-audit.py`，收據
`work/issue4-view-rename-production-r1-audit.json` SHA-256
`3a2c2d6f7e146249c51d05fb73a510f19f4fd213ee1f27af5a6a3336c5d52e13`。

公開producer產生的完整Go與已執行的frozen Go，除輸出前綴外逐byte相同，
涵蓋bootstrap、正常按鍵、虛擬時間及唯讀observer；此核對沒有再次執行或注入原版。
入口 `work/issue4-view-rename-public-proof-r1.py`，收據SHA-256
`4523dbb9a3197c0b352926e018c0d303bf913fe851cd9463437bff941453f15a`。
公開validator再次接受異名來源，另存
`work/issue4-view-rename-public-r5-receipt.json` SHA-256
`f3830912c5da369f79a7abe41bd50f8df1d2d162db94e906e84cc0677c9aeb27`；不覆寫正式測試固定的0d43e2bc來源收據。
八種損壞全部拒絕：錯seed、原版身分、缺consumer、提前寫姓名、錯目標、缺PNG、錯bin及缺正常返回。
前後正對照相同；入口 `work/issue4-view-rename-negative-r1.py`，收據SHA-256
`1d0afca79213d0427ffd75314303d6e10ee702f283bc3214422d7265c47653ee`。

[IDA語意索引](../tools/ida_recruitment_selection_ledger.json)保留舊11筆，追加6筆confirmed，
逐筆保留輸入身分、IDA linear、file offset、bytes、consumer、有限範圍與動態證據。
匯出器自動合併到原指令，未改名覆蓋原始定位。
新匯出 `work/issue4-view-rename-r3-ida.json` SHA-256
`5d3c0c1dd0305243afd1f3b9da3f97a97498ebe76a23d1e30730cdd4d919964c`，
321筆relocation後bytes全符合，6筆新語意自動附註且confirmed；其餘未知保持警示。
輸入與位址基準同上，database只在容器/tmp，不加入Git。

九份JSON由乾淨28690cc archive及[遷移器](../tools/migrate_recruitment_rename_pack.py)重建，
逐byte等於現行pack，保留原有排版。首版json.dump擴展排版已訂正，typed資料與canonical hash保持。
重建收據 `work/issue4-view-rename-pack-rebuild-r2.json` SHA-256
`2743ba2ac9ab2ef3b1a3e5864443b371893d8d356cb60c44980983b42a08b2d3`。
最後排版版本的原始EXE parity、7種壞契約、pack hash與desktop建置再次通過，
入口 `work/issue4-view-rename-final-checks-r1.py`，收據在同名前綴目錄，SHA-256
`2271e06821f8ac65ea0e193cedff76871b8ceb4642cdc7b33a61bcb577c24132`。

完整game454項清單覆蓋，407頂層／107子PASS、47選用SKIP；
全部11個internal套件、164頂層PASS、4選用SKIP，沒有素材缺失SKIP。
正常新遊戲至THE END156.69秒及desktop Linux x86_64 ELF通過，只屬remake回歸。
完整結果在 `work/issue4-view-rename-production-r5-full/`；各正常PNG及日誌在r4-full，
可重播入口 `work/issue4-view-rename-production-r4-run.py`、`-r5-run.py`。
三路首次單程序連跑被終止，改分路程序；未有OOM事件證據，不稱確定記憶體溢位。
全套唯一X11初始化失敗以同一binary及新Xvfb -noreset重跑通過，正式產品未改。
原版fixture缺頂層hash、raw格誤當字模、漏掛/assets_raw及檢查器誤將既有.md檔當目錄均屬驗證問題，保留歷史。
原版、PNG、binary、資料庫及私有收據只留本機；UID1000、root-owned基線3213、零.md目錄及使用者13項資料保持。
沒有新image或發行包，Issue #4與Goal保持進行中。
# 2026-10-04 咒文詳細頁：證據審查與有限 READY

本節承接 Issue #4。正式前版 `f823b61` 的第一個正常阻塞是第三職業男性登錄的第170包：原版停在咒文頁，remake 已進接受選單。以下規格只閉合初裝、無異常角色的咒文顯示與等待；多角色改名、其他裝備、全流程 V3 及原版存讀檔保持未知。

| 證據 | 身分與範圍 |
|---|---|
| 原版輸入 | `assets_raw/DQ3.EXE`，115282 bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；`D3TXT00.TXT` 的完整 record 121–180、471–473 |
| 主要分析 | IDA Pro 9.4，image `ida-pro-9.4-idapython:locked-v1`；`work/issue4-view-spells-r2-ida.json` SHA-256 `06f755d9a8d3209e48312591716670a695e3de90468ac43c3bb8d05e0a699b20`；166 指令 bytes 與 MZ relocation 核對。IDA linear；file=linear−EC90；DGROUP 基底 linear24DD0 |
| 正常原版 | dosgolem `2f44a68ebfc54b28fb15dd4a34510b0b04a5415d`，冷啟動前固定 seed1357 一次。203包來源 SHA-256 `42ccf52f2b0e4a292059ede6f0c4cdcfde082db91aee446949fc5d0120c11caf`；209包來源 `3582f82a49c77ce8de770b8a64b7d0fbae78bc2c29e26a3464737ae848eae17f`，494 IRQ1、418 完整 PNG／bin |
| 前綴與交易 | 前165包及330 PNG／bin保持前版來源，209來源前203包保持203來源；第172包才寫名冊。角色 raw+2E..31=`01000100`；兩類指向同一咒文 index40，union232D 僅一格，實際文字 record161。hero、金錢、旗標及名冊97bytes於觀看與返回保持 |
| 原型審查 | `work/issue4-view-spells-prototype-r5/`，正式 InputState 走完209包與存讀檔續行。原版203與205完整能力圖相同。原型尚未接 production；其 raw ID／座標不得搬入共用 Go |

原始定位與語意保持附加：`184A1` 計算四類已知咒文；`184C0` 清除60格集合；`184D1..18526` 透過指標表與 `185E4` 合併；`1853B..18566` 計算唯一數與列數。`18567→1F590` 建窗；`18573..185C1` 依 index 遞增、每列四欄顯示 record=index+79h；`185D7→2111B` 等待；`185DC→1F604` 還原能力窗；`10681` 接第二次 caller 讀鍵。上述非零咒文的有限玩家流程為 confirmed／D3；60項完整 catalog 與多列公式為原始資料及 consumer 的 D2，不宣稱其他角色的正常動態 V3。

READY 契約：

- typed `character_spells` 放在版本化 interface，含原始窗口、header／row／footer 的穩定 text ID、四欄 catalog 的 raw record→text ID、文字起點與步距、額外兩列及證據。引擎唯讀取 `LearnedSpells`，去重後依 catalog 順序排列。未知 record 拒絕進頁，不同步、猜補或改寫角色。
- 原始窗口 DGROUP3F34 的24bytes為 `120313002e002c006000d7010100d801d901000000000000`。X=19字元、Y=46、width=44字元；header471／row472／footer473。每列16px，實際高度=(ceil(unique/4)+2)×16。非零角色必須顯示一頁；零咒文維持既有流程。文字首格168,62，欄距80、列距16。幾何與文字都放 JSON，共用 renderer 沿已驗證陰影／frame primitive。
- 登錄：能力等待→咒文等待→接受選單→登錄交易。任何新按鍵或點擊各消費一次等待； held key 不前進。咒文等待的 Esc 只關閉本頁，不取消或登錄角色。
- 觀看：名冊→能力等待→咒文等待→還原能力的獨立 caller 等待→540→No→541→場景。第一等待的 K 不改名；咒文等待的 K 只關頁；最後 caller 等待才沿既有單人改名契約處理 K。
- bootstrap 檢查完整 JSON 欄位、引用、glyph、畫布與最大容量；缺少契約即啟動失敗。共用引擎不含 DQ3 raw record、座標、字串或 fallback。暫態頁面不新增存檔欄位；有效 Load 清理 UI，拒絕 Load 保持 UI、持久狀態與 RNG。
- 驗收：原始 EXE／DAT parity、壞契約、去重／排序／多列／未知 record 元件測試；正常209包逐包位置／旗標、觀看 snapshot／RNG、同版 Save／Load 與下一步。完整640×350 RGB保留兩側正常能力差異，不裁切、遮罩、調 RNG 或指定動畫相位；只對新增咒文層與等待範圍判定 CONFORMED。

正式有限切片已通過正常209包，狀態 E2／玩家路徑 E3、畫面 V2。第170包完整RGB差由16485降548；169／170的差分位置與雙側RGB相同。觀看203／204／205各430，三張的完整差分相同。原始203及205完整能力圖相同，正式圖也相同；新增咒文層沒有新差異。完整V3仍RED，不能把不同能力或既有動畫當成零差異。正式45張PNG逐byte等於已驗原型，165..169的五張保留診斷前綴；全畫布稽核 `work/issue4-view-spells-formal-audit-r1.json` SHA-256 `4e568b0948bb8bb2a1db00c091d17739ff1361a935c7a8bd372c019bde2923e2`。

嚴格原版來源稽核收據 `work/dosgolem-opening/issue4-view-spells-class3-normal-r2-source-r2-receipt.json` SHA-256 `4dcb99c80e5cddc17940b09772c37169d5348cc40249996827029812365b8c08`。八種壞來源拒絕與前後正對照保持，稽核 `work/issue4-view-spells-audit-r6.json` SHA-256 `a309ff9a39e3743a05d28fe7a2f1eff93a161e834fad7f112a3b781e8d42d949`。來源接受不等於完整parity，receipt仍保留 `full_rgb_parity=false` 與原版Save／Load未驗。

原始語意索引 [ida_recruitment_selection_ledger.json](../tools/ida_recruitment_selection_ledger.json) 保留17筆並追加4筆有限confirmed，保留各原始linear／file位址、file bytes、consumer與來源範圍。沿已驗證IDA匯出流程自動合併，`work/issue4-view-spells-r3-ida.json` SHA-256 `5273a602c9084dcf089d709676112fa9f7692d0fbba14ba512445b9ff5121e6a`；166指令及MZ relocation全符合，4筆語意同列顯示。其他未審查項仍unknown；六份64bytes指標目標擷取只表示有界prefix，不能稱table長度。

正式typed契約、共用renderer、原始EXE／DAT parity及正式209包入口見 [docs/84](84-game-pack-json-contract.md) 的 `character_spells`；[公開資料遷移器](../tools/migrate_character_spells_pack.py)從乾淨 `f823b61` 的九份JSON重建schema0.15.0／content0.1.87，逐byte一致且保留排版。canonical `sha256:2d712e65f18ce7919ddccf970160d68d00b63a54a73088bcd473c38a9fa61d10`。資料重建收據 `work/issue4-view-spells-pack-rebuild-r2/receipt.json`；舊schema存檔仍拒絕，不自動遷移。

原型首次誤用酒館的 D3TXT01 record，已查明咒文框與咒名來自 D3TXT00；正式版只能用 pack 的已核對 glyph text ID。兩側 seed 相同但全域 RNG 呼叫序列未對齊，正常能力數值保持各自結果，完整畫面仍非 V3。

公開重生入口為 [dosgolem_character_spells_probe.py](../tools/dosgolem_character_spells_probe.py)，`--prefix` 必須為全新名稱，預設正常209包；`--first-ability` 停203包。原版工具鏈掛載與 seed 契約沿本文件既有 dosgolem 入口。來源稽核入口為 [verify_dosgolem_character_spells.py](../tools/verify_dosgolem_character_spells.py)，參數依序為原版產物目錄、前綴、實際 producer、全新 receipt；只有203包加 `--first-ability`。公開 producer 的完整 Go 與兩份已執行程式除輸出前綴外相同，證據 `work/issue4-view-spells-public-proof-r1.json`。原版素材、PNG、binary、database與私有收據保持本機。

完整回歸以 `work/issue4-view-spells-full-r2/game-receipt.json` 為準：game459覆蓋，412頂層／107子PASS、47選用SKIP；全部11個internal套件167頂層／341子PASS、4選用SKIP。正常THE END97.88秒及desktop Linux x86_64通過，沒有素材缺失SKIP。r1唯一過期的campaign登錄輸入已補上咒文確認，再以r2正常重跑；其餘通過項的production相同，保留原始r1收據。不外推原版完整流程。舊九路線516張、新45張PNG逐byte保持，收據 `work/issue4-view-spells-regression-images-r1.json` SHA-256 `46a4e1588cdca69c46abc274b8f5bb402079ce3096a98f4c745219dd73be8581`。

# 2026-10-04 其他裝備詳細頁與一般字母鍵：DRAFT

本輪工作依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5978134553)。由041caf9接手，先保留現行production checkpoint。

| 原始定位 | 推論等級與可查證結果 | 有界證據 |
|---|---|---|
| IDA linear1834E..1839D；file96BE..970D | strong：從所選角色+3A依序讀八個word，只顯示bit8000已裝備且低byte不是AF的項目；尚缺其他裝備正常動態樣本 | IDA9.4匯出140指令，原始bytes及MZ relocation逐項符合 |
| IDA linear1838C..18392 → 213C4 | strong：遮掉旗標後以item+1作文字record，consumer完成後Y加16；不改成四槽排序 | 追加350指令；原始物品來源及文字consumer |
| IDA linear18477 → 21929 | strong：畫面攻擊值讀角色+1C，不能一律以力量欄替代已武裝攻擊 | 原始指令與consumer；正式其他裝備renderer尚未READY |
| IDA linear10681..10689；file19F1..19F9 | strong：第二讀鍵後比較AH=25h，K進改名，其他鍵走1068E還原視窗 | 原始call、cmp與分支；正常A鍵來源正在重生 |

IDA輸入為assets_raw/DQ3.EXE，115282bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。工具IDA Pro9.4，位址空間linear；file=linear−EC90，DGROUP基底linear24DD0。
匯出 `work/issue4-view-equipment-r1-ida.json` SHA-256 `78b5d5592674068638ec525c4cc6e3870c7f831cb427e2a2a92d508c651f9276`；追加 `work/issue4-view-equipment-r2-ida.json` SHA-256 `f4f7d33181af4b5333fa6d7f2305bf5c25c1304f33c97cd6efbdcb5a65e8db9f`。database僅一次性容器/tmp，原始檔唯讀，既有21筆ledger保留。

裝備頁缺正常換裝、分離回名冊的原版來源。既有入隊原版199停短曲完成等待，維持硬體driver／ISR停止線；本輪不深入硬體、不解除裝備guard、不猜資料格式。

一般字母鍵已由新增元件紅測試 `TestRecruitmentViewDetailUnmappedFreshKeyCloses` 重現正式第二等待卡住。舊測試把AnyKeyEdge判為保持等待，該斷言尚缺原版動態支持，現列待勘誤。正式InputState的鍵盤入口已會產生AnyKeyEdge，修正範圍只在具名第二等待。原版尚未接受前不改production。

正常重生入口 [dosgolem_recruitment_view_key_probe.py](../tools/dosgolem_recruitment_view_key_probe.py) 沿用本文件的Docker與固定dosgolem2f44a68掛載，參數 `--prefix` 必須全新；正常209包只把206的Esc改A，冷啟動前seed1357固定一次，無狀態或裝備注入。
來源稽核入口 [verify_dosgolem_recruitment_view_key.py](../tools/verify_dosgolem_recruitment_view_key.py)，依序傳原版產物目錄、前綴、實際producer及全新receipt。重用既有嚴格咒文來源核對，另驗前205包、全部PNG／bin、第二讀鍵AH與非K分支。來源接受、READY與正式同狀態驗證分開記錄。

## 一般字母鍵：證據審查與有限 READY

正常A鍵來源已接受：`work/dosgolem-opening/issue4-view-close-key-normal-r1-source-r1-receipt.json`，SHA-256 `c0a7bc08fa22e5f144e5475b7aa5bb87c5599ea353ccea8c146bfa77008069fc`。工具為固定dosgolem2f44a68、DQ3.EXE115282bytes／5178fdc8，seed1357在執行前只固定一次。209包、494 IRQ1、418份完整PNG／bin，前205包與4dcb99c8來源逐byte相同。無角色注入、狀態還原、重擲或相位指定。

- confirmed：206的A掃描1Eh在2110B消費，10686、10689、1068E三點皆AX1E00；走非K關頁。之後540→No→541→原生field1997C。名冊97bytes、主角128bytes、隊伍、金錢、旗標保持。
- confirmed：正式修正前以同一209包與InputState重播，第一個新blocker在206。`work/issue4-view-close-key-normal-red-r1/game.log` 明確停在「second wait must close」，元件紅測試亦重現。
- READY引擎契約：rcViewClose消費新的AnyKeyEdge，清詳細頁並依pack既有text ID回到繼續詢問。K由既有具名改名分派優先處理；第一能力／咒文等待仍只消費一次輸入。held方向沒有新edge時保持等待。觸控既有契約保持。
- 不改pack、schema、geometry、文字或持久存檔格式。同版本有效Load清UI、拒絕Load保持UI沿既有契約；正式正常209包另驗snapshot／RNG、存讀檔與下一步。
- 範圍：初裝、無異常Class3男性正常A關頁與已有K／Esc分支回歸。其他裝備頁仍DRAFT；原版存讀檔、其他角色動態、全畫布V3與完整campaign未接受。正常動態證據限A，不冒稱逐鍵窮舉。

勘誤：前批「未知鍵保持第二等待」的元件assertion已由正常206及原始非K分支推翻。它源於未取得一般字母鍵動態來源，不是原版規格。本節追加勘誤並保留前批正常Esc／K收據，不改寫歷史。

公開新normal trace測試在 [character_spells_normal_test.go](../dq3_remake_ebitan/game/character_spells_normal_test.go) 的 `TestRecruitmentViewUnmappedKeyDosgolemNormalInputComparison`，與既有正常咒文測試共用整條production重播；原版收據hash與206角色等待閘門明確固定。

## 一般字母鍵：有限 CONFORMED

- 正式第二等待接AnyKeyEdge。第一等待仍分開消費，K維持具名分派優先，held方向保持。修正前正常209包的206紅測試，修正後同一來源通過；snapshot／RNG、同版本存讀檔及下一步保持。
- 正式A與既有Esc、三路K改名及受影響測試：37頂層、21子PASS、零SKIP；正常新遊戲至THE END112.88秒。桌面Linux x86_64 SHA-256 `cda5b6c0d97e80dece9fc0552d6df853614c64d2a9e4cfe9ace19104d7ba1777`。本輪不重跑無關已綠套件；前次完整game459及全部11個internal為041caf9。
- 新45張正式PNG逐byte等於前次Esc，八條舊路線561張保持。原版A／Esc的全209張PNG與bin相同；新正常203／204／205各430、206／207／208各7、209為411個完整RGB差異。能力與人物動畫差異保持，完整V3仍RED，不裁切、遮罩、重擲或指定phase。
- 八種壞來源全拒絕，前後正對照相同；既有Esc209包由新checker預設參數重新接受且狀態／輸入保持。收據 `work/issue4-view-close-key-audit-r1.json` SHA-256 `e41025909f14c3e0308d564277b9263dfe52e4e47ec908eb437a17ab308b4605`。
- 原始定位ledger保留21筆，追加10689非K分支一筆confirmed；不修改既有10686的K定位。IDA9.4重新匯出140條目並自動附註，SHA-256 `3516eec469517cf06b5f869f1d4d8b9d5ff00065b67a400dceb9e2437337d514`。其他裝備相關未閉合語意仍unknown／strong。
- game收據 `work/issue4-view-close-key-validation-r1/game-receipt.json` SHA-256 `79bdb7c00a38afa52d23835cc2c862cd67b48b70bcea596dc3659541e986b24c`；畫面稽核 `work/issue4-view-close-key-images-r1.json` SHA-256 `3cd67c8bfc40dcd5532f644cd08d5d0f4c55b4b5886e2359cd7817e44e3f4f22`。收尾 `work/issue4-view-close-key-final-audit-r1.json` SHA-256 `c103a2f852218a8b75a6d1eef7b287cfc9cf5cb04f1207dc65a5da97be47080a`。

九份JSON與041caf9逐byte相同，schema0.15.0／content0.1.87與canonical `sha256:2d712e65f18ce7919ddccf970160d68d00b63a54a73088bcd473c38a9fa61d10`保持。沒有存檔格式或交付變更。其他裝備頁未READY，原版Save／Load、音畫與完整原版campaign仍未知；正常THE END只證明remake可玩。

公開producer即本次實際執行的 [dosgolem_recruitment_view_key_probe.py](../tools/dosgolem_recruitment_view_key_probe.py)；來源meta固定producer／generator／runtime／完整Go及binary hash，原版執行不使用私有額外patch。兩個公開新工具由本節索引。原版素材、PNG、binary、database與私有收據不加入Git。

# 2026-10-04 空名冊觀看：DRAFT

工作依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5978311832)，從af7d148接手。正常姓名取消後選否、下樓、選觀看；不建立角色，不注入名冊或遊戲狀態。原版冷啟動前seed1357固定一次。

公開重生入口 [dosgolem_recruitment_empty_view_probe.py](../tools/dosgolem_recruitment_empty_view_probe.py)，沿本文件既有Docker、固定dosgolem2f44a68與唯讀原始資料掛載，參數 `--prefix` 必須全新。前164包沿已接受正常姓名取消來源，後接正式移動及觀看輸入。只附加唯讀入口、計數與文字觀察，不改原始程式。

IDA9.4原始bytes與MZ relocation核對121指令，匯出 `work/issue4-empty-view-r1-ida.json` SHA-256 `7bdc4ca6dbeba0fb8affe46c708126cd975f98e3ca65190046ad5206aeadc216`。輸入assets_raw/DQ3.EXE115282bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；linear位址，file=linear−EC90，DGROUP基底linear24DD0。既有22筆語意索引保持。

| 原始定位 | 推論等級 | 有界結論 |
|---|---|---|
| linear10624→10974→10627..1062F | strong | 計數後讀DGROUP5060；零時跳10696，不開名冊 |
| linear10974..109A9 | strong | 掃角色槽1..11，旗標1計入CH，低byte寫DGROUP5060；高byte不作推測 |
| linear10696..1069E→21414 | strong | 零名冊選record316並返回caller；文字bank與等待尚待正常動態來源 |
| linear103A8..103D6 | strong | View返回後選record540；No選541、獨立讀鍵，再還原caller |

目前remake空名冊仍進rcView，renderer沒有列表而Confirm沒有作用。原版來源接受前不改production；正常D3、紅測試、READY、資料契約與CONFORMED分開記錄。音樂完成仍遵守硬體時序停止線，不重新研究ISR。

## 空名冊觀看：證據審查與有限 READY

原版正常195包、466 IRQ1來源接受，`work/dosgolem-opening/issue4-empty-view-normal-r1-source-r2-receipt.json` SHA-256 `ff7a8abd5e93c867e5650f08a3af607feecf7013d7a781ff9a9ad2fd44eaeba7`。前164包與正常姓名取消來源0467c01b逐項相同，完整PNG／bin與174份母親父來源保持。實際producer即公開腳本，metadata固定generator／runtime／Go source及binary身份。

- confirmed：191經10624→10974→10627，DGROUP5060為零；名冊12槽旗標為020000000000000000000000，只保留hero。經1062F→10696→10699選record316，原生D3TXT00文字bank2826，停內文等待216D8。
- confirmed：192確認後完成316剩餘文字，1069E返回caller；103AE選540並開Yes／No。193選No，194選541並停獨立讀鍵21133，195返回field1997C、座標2,18。主角128bytes、金錢及旗標從150保持；View入口與返回名冊仍空。
- 原始record316保留完整glyph codes及FFFC內文等待，兩個FFFE換行。文字為「要先到冒險者的登錄所」及後續登錄指引；不將等待嵌入可讀字串。
- 修正前正式InputState正常195包第一次新blocker為191，`work/issue4-empty-view-red-r2/game.log` 明確重現空名冊缺內文等待；前188包仍沿正式新遊戲與正常姓名取消。
- READY：pack必填`empty_view_text_id`引用D3原始316。rcMenu的View在空roster時播放該record，EOF後以具名有限延續接既有AgainTextID；一次新輸入只消費內文等待，不能同時選Yes／No。無需新geometry、名冊交易或RNG呼叫。非空View沿原有路線。
- 新欄位使用schema0.16.0／content0.1.88；九份JSON同步，缺失、null、未知引用及非D3文字拒絕。舊schema或不同canonical hash存檔依既有策略拒絕，不自動遷移。
- 來源限定空名冊View；不外推空Join、空Leave、滿隊、原版存讀檔、音訊或完整campaign。完整RGB另記，不以state通過聲稱V3。

來源稽核入口 [verify_dosgolem_recruitment_empty_view.py](../tools/verify_dosgolem_recruitment_empty_view.py)，參數依序為產物目錄、前綴、實際producer與全新receipt。工具核對完整輸入IRQ、原生計數與文字、父來源、PNG／bin及field返回；原版產物留本機。

環境勘誤：初版checker把187問候誤標waiting，實際為inline_wait；按原版log修正後乾淨重跑。首版收據缺共用畫面checker要求的頂層original_sha256，導致驗證腳本形狀錯誤；保留r1，r2補齊身份且原版執行、輸入與PNG不變。這兩項不記為產品缺陷。

## 空名冊觀看：有限 CONFORMED

- 正式空View播放pack的316完整字碼，再接540。正常195包、逐包位置與旗標、空名冊、持久snapshot／RNG、同版本Save／Load及下一步全部通過。一次確認只結束內文等待；元件另驗held key不消費及告別獨立等待。非空View、A／Esc及三路K正常回歸保持。
- game463項完整覆蓋，416頂層／107子PASS、47選用SKIP；internal167頂層／344子PASS、11套件通過、4選用SKIP。沒有素材缺失SKIP。正常THE END130.07秒，desktop Linux x86_64 SHA-256 `f074a88692ea5d027bccc6f2a61e88c749fe7c6418af2c1a10e5a0b9d8635b1e`。game收據 `work/issue4-empty-view-full-r1/game-receipt.json` SHA-256 `0d46c863e2f5baff2e67be83bdb3df0b91c8a3335f7a18e77e87e57315329cd1`。
- 原有九路線606張PNG逐byte保持；新正常46張留本機，其中150..164等於既有正常取消。原版前164包與其PNG／bin同樣保持。新187..194完整RGB各295，195為411；有限狀態E2／流程E3、畫面V2，完整V3仍RED。取圖未遮罩或指定人物相位，不將前批人物歸因外推本批全部差異。
- 八種壞來源拒絕，正對照前後一致；收據 `work/issue4-empty-view-audit-r1.json` SHA-256 `467c47663e2c70e3feeadebdea14c9aacbef1726f4ee7a52b3aef89d66215188`。畫面收據 `work/issue4-empty-view-images-r1.json` SHA-256 `22b3fbca24dc58325c60b09c38971ebcb1d608a8e0d054a0d66fe58512391084`。
- 九份JSON從乾淨af7d148由[公開遷移器](../tools/migrate_recruitment_empty_view_pack.py)重建逐byte一致，保留排版；schema0.16.0／content0.1.88，canonical `sha256:096fe3a764b72142fade8b8842cc6e9bc085c7a12f9242420120dafba7bd5c50`。重建收據 `work/issue4-empty-view-pack-rebuild-r1/receipt.json` SHA-256 `d3089b8f85553a6b2c6575f73790678049d5e5cbf70a56419fecf196c8f55f4c`。舊schema存檔拒絕，不自動遷移。
- 原始定位索引保留22筆並追加1062F／10696兩筆有限confirmed；IDA9.4自動合併匯出121指令、bytes及MZ relocation核對，`work/issue4-empty-view-r2-ida.json` SHA-256 `21c8699354c3b548c15569157a6d6d84aa7dc9ddbf566cc78d9d4616b4f690cc`。計數高byte與其他空清單不外推。

此來源checker固定本段195包的Go及binary hash，只接受已審查checkpoint；完整重生採相同`issue4-empty-view-normal-r1`前綴及全新可寫輸出掛載，原有父來源以唯讀輸入提供。更改prefix、producer或工具鏈要另行接受，不能只自行改metadata hash。
圖片稽核兩次假失敗已查明：改名圖片位於子目錄，取消與創角路線又在157分岔。修正腳本後以真正正常取消150..164核對，保留完整606張非空路線檢查；不把分岔圖片當產品不一致。
沒有新發行包，原版Save／Load、空Join／Leave、滿隊、其他裝備、多角色改名、音畫與完整原版campaign仍未知。Issue及Goal保持進行中。

# 2026-10-04 空名冊Join與單人隊伍Leave：DRAFT

依 [Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5978582921)，從1217cbd續行。前一切片空View已CONFORMED，不把其record316外推所有空選單。
兩條正常冷啟動來源由[公開producer](../tools/dosgolem_recruitment_empty_probe.py)重生，`--action join`或`--action leave`，`--prefix`使用全新名稱。
沿本文件既有Docker、唯讀原始DQ3.EXE／資料與dosgolem2f44a68掛載，seed1357執行前固定一次；正常姓名取消、選No、下樓後選對應動作。觀察只讀名冊、計數、隊伍指標及文字bank，沒有狀態注入。

IDA Pro9.4窄查298條目，輸入assets_raw/DQ3.EXE115282bytes／SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`，linear位址；file=linear−EC90，DGROUP基底linear24DD0。
匯出 `work/issue4-empty-recruit-r1-ida.json` SHA-256 `a1f0f62a7d90ba457ea5fb22700ddb6f8a25cea7727bde70a7db70b34fc4ed29`，原始bytes與MZ relocation全符合，24筆既有語意索引保持。

| 原始定位 | 推論等級 | 有界結論 |
|---|---|---|
| linear10395→103D7..103EC | strong | Join先檢查DGROUP5077是否4；未滿先顯示530，再呼叫10974計數 |
| linear103EC..103F4→1046A..10472 | strong | 未入隊計數DGROUP5060為零時選316；返回caller後接540 |
| linear103A2→104C4..104CD→105BC..105C4 | strong | Leave先讀DGROUP5077；單人隊伍1直接選542並返回，沒有選人窗 |
| linear103AE..103D6 | confirmed於既有空View有限來源；本批待動態閉合 | 共用540 Yes／No、541獨立等待與場景返回；不先猜本批按鍵數 |

目前remake空Join只接問題後進空清單；單人隊伍Leave直接進空清單，renderer保留舊通用框。兩條正常原版來源接受前不改production。
本批只閉合玩家空清單阻塞、typed資料、正常UI與存讀檔。非空分離的角色交易與隊伍重排不由空隊伍樣本驗收；硬體driver／ISR停止線維持。

來源稽核入口[verify_dosgolem_recruitment_empty.py](../tools/verify_dosgolem_recruitment_empty.py)：依序傳產物目錄、`join`或`leave`、實際producer及全新receipt。固定本段已執行的Go／binary身份，不只相信metadata。冷啟動完整重生採同一已接受前綴、全新可寫輸出與唯讀父來源。
兩條193包／462 IRQ1來源已接受：Join `work/dosgolem-opening/issue4-empty-join-normal-r1-source-r1-receipt.json` SHA-256 `47d1a11585d4d3a3d5b6a04fc8923a3ab7adc305b1d7272a5d5823b73e3733cd`；Leave `work/dosgolem-opening/issue4-empty-leave-normal-r1-source-r1-receipt.json` SHA-256 `df05d0d4e61ff06c3ced99c9939c180ac9d8116305e31012f33c5350ba4d0ffb`。
兩側前188包與空View來源逐項相同，全部PNG／bin及174份母親父來源保持。分支入口、caller返回與告別後隊伍數1、名冊只hero、四個隊伍指標保持；主角128bytes、金錢及旗標從150保持。沒有重擲或狀態注入。

## 空Join與單人Leave：證據審查與有限 READY

- confirmed：Join189經103D7、103E1播放530，10974計數後103EC讀DGROUP5060為零，103F4跳1046A播放316並停內文等待。190的新確認才經10472返回caller103AE、播放540；191選No，192播放541，193獨立確認返回field1997C、2,18。
- confirmed：Leave190經104C4讀DGROUP5077為1，104CD跳105BC播放542，105C4返回caller103AE，同一包接540。542有兩個FFFE與FFFB主角姓名插值，沒有FFFC等待。191選No，192播放541，193獨立確認返回field；不新增原版沒有的讀鍵。
- 修正前正式正常輸入測試：Join189缺內文等待，Leave190缺續問；元件兩條同樣RED。既有空View195包仍GREEN。日志在 `work/issue4-empty-recruit-red-r1/`，來源與正式程式分開保存。
- READY資料契約：必填`empty_join_text_id`引用既有原始316，必填`empty_leave_text_id`引用新增原始542。530、316、542保留原始record界線及glyph控制碼。schema0.17.0、content0.1.89，九份JSON同步；缺失、null、未知引用及未審查文字拒絕。舊schema／canonical hash存檔沿既有拒絕策略。
- READY引擎：空roster Join先播PromptTextID，EOF以具名延續播EmptyJoinTextID，內文等待後才接AgainTextID。空companions Leave播放EmptyLeaveTextID，文字引擎綁定目前主角姓名，EOF直接接AgainTextID。與空View共用既有文字等待和續問狀態機，不嵌入任意JSON流程。無名冊交易、RNG、geometry或存檔結構改動。
- 範圍只限正常空名冊Join與單人Leave，未滿隊。非空分離、滿隊、其他裝備、原版存讀檔及完整V3仍未接受；不放寬既有guard。正式正常輸入另驗持久snapshot／RNG、同版本存讀檔與下一步。

乾淨pack重建與typed契約入口見[docs/84](84-game-pack-json-contract.md)的空加入與單人分離增補，公開工具為[migrate_recruitment_empty_pack.py](../tools/migrate_recruitment_empty_pack.py)。

## 空Join與單人Leave：有限 CONFORMED

- 空加入依pack先播530再316，獨立內文確認後540；單人分離播放542與目前主角姓名，EOF同包接540。兩條正式正常193包與等價InputState、持久snapshot／RNG、同版本Save／Load及下一步通過；既有空View195包、非空選人、A／Esc及三路K保持。未新增角色交易、RNG或geometry。
- game466覆蓋，419頂層／109子PASS、47選用SKIP；internal167頂層／350子PASS、11套件、4選用SKIP，沒有素材缺失SKIP。正常新遊戲至THE END253.74秒；desktop Linux x86_64 SHA-256 `8d99053101abc2bb3afd2214738ee0c7fbf0a11ebc6291c364faa5ef8602d1dc`。
- 十條舊路線652張PNG逐byte保持，新兩路各44張，正常150..188等於既有空View。兩路187..192完整RGB各295、193返回零差異。沒有遮罩、裁切、重擲或相位指定；有限狀態E2／流程E3、畫面V2，完整V3仍RED。不把不同返回時刻的零差異外推先前195包或其他人物動畫。
- 兩條來源各八種壞樣本拒絕，正對照前後一致。九份JSON由乾淨1217cbd與公開遷移器重建逐byte相同。schema0.17.0／content0.1.89，canonical `sha256:4b235d635e29011f232b212b533e5de32c29c32a8d094c236eeba68198b9b988`；舊schema及hash不符存檔拒絕，不自動遷移。沒有新發行包。
- 原始定位ledger保留24筆，追加1046A／105BC兩筆有限confirmed；IDA9.4重新匯出298條目，bytes及MZ relocation核對，自動附註26筆。輸入與linear／file／DGROUP基準沿本節READY；匯出 `work/issue4-empty-recruit-r2-ida.json` SHA-256 `a4deee0a4291fa5d00ff264e47aa4d78e951f91b711f85781dcfa31adf83ea11`。
- 完整game收據 `work/issue4-empty-recruit-retry-r1/game-receipt.json` SHA-256 `22201a3e574749d3de7eacd43cc9ab29921c6f153b6f9a0deea17d47e7a3a613`；來源稽核 `work/issue4-empty-recruit-audit-r1.json` SHA-256 `1ff3fb95298c0fd71b87a1b1c180b63b250badb9756cdb0d940205f98575d7f4`；畫面 `work/issue4-empty-recruit-images-r1.json` SHA-256 `1bc8adb7607b467dc3f9ad0972e3b67e2a9abd19a201673e3155803e37177bba`；乾淨pack `work/issue4-empty-recruit-pack-rebuild-r1/receipt.json` SHA-256 `744f87c07be4774a450e22acf0e0c34b19407030b9522d84f45261f8214c388f`。

驗證環境紀錄：首輪三個大型來源測試同時跑，3GiB容器的memory.events記錄oom_kill5，五項exit−9且無產品assertion。保留首輪log及收據，以同一binary、同一工具鏈逐一重跑五項全部PASS，重跑容器oom_kill0；其他已綠項目不重跑。首次編輯命令的換行escaping亦造成SyntaxError、尚未執行寫入；改檔案化容器腳本後完成，不記為產品缺陷。

公開producer與checker即本次實際原版執行與接受工具，由本節DRAFT索引；遷移器由docs/84及READY索引。原版素材、PNG、EXE、database與私有work不加入Git。下一步空清單Yes正常續行；非空分離、滿隊、其他裝備、多角色改名、原版Save／Load、音畫及完整原版campaign仍未知。Issue與Goal保持進行中。

# 2026-10-04 空Join與單人Leave的Yes續行：DRAFT

從c572ab3依[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5979001959)續行。正常新遊戲、姓名取消、下樓，選空Join或單人Leave，再選Yes、重選同一動作、No與告別確認；不注入角色、名冊或場景狀態。
原版固定dosgolem2f44a68，seed1357在冷啟動前只固定一次。來源與正式測試分開接受；完整RGB不裁切、遮罩或指定人物相位。

既有IDA9.4窄匯出 `work/issue4-empty-recruit-r2-ida.json` SHA-256 `a4deee0a4291fa5d00ff264e47aa4d78e951f91b711f85781dcfa31adf83ea11`，298條目、原始bytes／MZ relocation符合。輸入assets_raw/DQ3.EXE115282bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。linear位址；file=linear−EC90，DGROUP基底linear24DD0。

- strong：103B9比較DGROUP0722與2，非No經103C0(file1730，bytesEBB6)回10378(file16E8，bytesBF1002)，重新選record528，1037B呼叫21414後10384進選單consumer1F4E3。空清單正常Yes動態尚待閉合。
- confirmed於前一空No來源：Join的530／316及單人Leave的542至540，不包含本輪Yes或重複第二輪。26筆既有ledger保持，不因靜態相同就升格新分支。

公開正常重生入口[dosgolem_recruitment_empty_yes_probe.py](../tools/dosgolem_recruitment_empty_yes_probe.py)，在本文件既有Docker與唯讀原始素材／固定dosgolem掛載下，以`--action join`或`leave`及全新`--prefix`執行。沿已接受producer的正常前綴，只增加可見按鍵與只讀觀察；來源接受前不改production。

來源稽核入口[verify_dosgolem_recruitment_empty_yes.py](../tools/verify_dosgolem_recruitment_empty_yes.py)，依序傳原版產物目錄、`join`或`leave`、實際producer及全新receipt。固定本輪已執行的Go／binary身份，逐包核對實際IRQ、文字consumer、兩輪入口與返回、角色、PNG／bin及父來源；來源接受與正式parity分開記錄。

## 空清單Yes：證據審查與有限 READY

兩條正常196包／468 IRQ1來源已接受：Join `work/dosgolem-opening/issue4-empty-join-yes-normal-r1-source-r1-receipt.json` SHA-256 `f56419c0113c6f895b71f9e287550c1c7d9e477e76172ddabc48038f3d3c3777`；Leave `work/dosgolem-opening/issue4-empty-leave-yes-normal-r1-source-r1-receipt.json` SHA-256 `d8c6c47d82a6adb67a08dd236f9b175c4f217521938f41838c78b619d23274ef`。實際公開producer479f0829與兩側完整Go／binary身份固定；前190包、456 IRQ1及PNG／bin等於各自前一No來源，174份母親父來源保持。沒有重擲或状态注入。

- confirmed：191確認Yes，103B9讀DGROUP0722非2，103C0→10378重新選528，10384再次進原始三項選單。原生D3TXT00 bank2826，初始游標回1；不重播527問候、不退出場景或重新建立角色。
- confirmed：Join192再進103D7、先530，10974計數零，1046A再次316內文等待；193新確認後540。Leave192只移游標，193再次104C4讀隊伍數1、105BC選542主角姓名，EOF同包540。194選No、195選541獨立等待、196確認返回field1997C、2,18。
- confirmed：兩輪與Yes返回入口，名冊12旗標020000000000000000000000、DGROUP5077為1、四個隊伍指標穩定；主角128bytes、金錢及旗標從150保持。兩條actor與caller證據不外推非空分離、滿隊、原版Save／Load、音訊或完整V3。
- READY：沿既有ContinueTextID、具名rcAgain→文字→rcMenu，Yes只消費這次選擇；父問題EOF後停選單，游標重設0，不同一輸入再開Join或Leave。後續正常動作沿前一已CONFORMED空分支。共用文字保留既有已畫操作與捲動，不要求記憶體物件指標相同。正式輸入、全RGB、snapshot／RNG、同版本存讀檔與下一步分別驗收。
- 本輪沒有新資料欄位或規則，先驗既有schema0.17.0／content0.1.89及canonical4b235d63。若正式重播不符，回到證據與READY修正，不猜補production值。

新增元件測試首版以指標相等代替文字內容保留而RED。`appendRetainedRecord`明確建立新狀態物件，沿用原有ops；改成核對已畫操作前綴，保留 `work/issue4-empty-yes-baseline-r1/` 紅測試。三條既有空No正常測試保持GREEN。這是測試判準訂正，未動正式程式；正常Yes來源另行驗收。

## 空清單Yes：有限 CONFORMED

- 既有正式狀態機與pack符合兩條原版196包。191的Yes只重播528並重設父選單游標；Join192／193與Leave192／193分別等待或選擇，194 No、195告別獨立等待、196行走。持久snapshot／RNG、同版本Save／Load與下一步通過。正式Go與九份JSON不變，schema0.17.0／content0.1.89及canonical4b235d63保持；本輪未新增產品修正或發行包。
- 34受影響頂層／25子PASS、零SKIP；沒有素材缺失SKIP。最近完整game466及全部11個internal、正常THE END253.74秒、desktop為c572ab3，按比例未重跑無關套件。收據 `work/issue4-empty-yes-retry-r1/game-receipt.json` SHA-256 `a99f0d9bd92ab6eb3992b9e5e52f2ce458f26fb09f925f7e8f884b23ad4079f8`。
- 十一條重生舊路695張PNG逐byte保持，新兩路各47張，150..190等於各自No路線。兩路187..195完整RGB各295；196返回Join411、Leave0。全畫布未裁切、遮罩、重擲或指定phase；狀態E2／流程E3、畫面V2，完整V3仍RED。收據 `work/issue4-empty-yes-images-r1.json` SHA-256 `665b2b6ceab7ec806f8ac87c9d6213b5fc18864e18bd5776749fd12b664833a7`。
- 兩條來源各八種壞樣本拒絕，正對照前後相同；收據 `work/issue4-empty-yes-audit-r1.json` SHA-256 `f069ba2d53d3339520104545dda1b9dbb636fa5656118a62e127e9cc32eda1e7`。來源checker固定本節Go／binary，冷啟動重生沿相同已接受前綴及全新可寫輸出、唯讀父來源；不只改metadata身份來接受新來源。
- 原始定位ledger保留26筆，追加103C0有限confirmed；IDA9.4自動合併27筆、298條目bytes／MZ relocation全符合。匯出 `work/issue4-empty-yes-r1-ida.json` SHA-256 `7c2a987cc2a9e957dcd5cf1858cb25bcf886123fccc37abb34bac5f69712ea08`；輸入與linear／file／DGROUP基準沿本節READY，不改原始名稱與位址。

驗證環境訂正：改名三條路線在同一程序累積記憶體，3GiB容器記錄oom_kill1、exit−9；按既有已驗證方式分成run2／4／5三個程序，以同一binary全部PASS。一般關頁鍵漏設共用角色咒文來源環境變數而SKIP，補齊後實際PASS；重跑容器oom_kill0。保留首輪log與收據，不把環境或測試判準錯誤當產品缺陷。

下一步正常招募主選單Esc：既有IDA10387比較DGROUP0726取消flag，1038C跳103C2告別與獨立讀鍵；目前remake直接關閉。這仍為待正常來源的強推論，不先改production。原版Save／Load、非空分離、滿隊、其他裝備、多角色改名、音畫及完整原版campaign仍未知；Goal與Issue保持進行中。

# 2026-10-04 招募主選單 Esc：DRAFT

依 Issue #4 從 67f6ff8 續行。正常新遊戲、姓名取消、下樓、空 Join 後選 Yes，再在三項主選單按 Esc。前 191 包沿已接受空 Join Yes 來源；seed1357 在冷啟動前固定一次，不注入名冊、角色或時鐘。

公開來源入口 [dosgolem_recruitment_menu_cancel_probe.py](../tools/dosgolem_recruitment_menu_cancel_probe.py)，沿本文件既有 Docker 與 dosgolem2f44a68 唯讀掛載，以 `--prefix issue4-recruit-menu-esc-normal-r1` 在全新可寫來源輸出執行。只增加可見 Esc 與唯讀取消 flag 觀察，保留既有 producer／checker。

既有 IDA9.4 匯出 `work/issue4-empty-yes-r1-ida.json` SHA-256 `7c2a987cc2a9e957dcd5cf1858cb25bcf886123fccc37abb34bac5f69712ea08` 的 298 條目、bytes 與 MZ relocation 已核對。輸入 assets_raw/DQ3.EXE115282bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。linear 位址，file=linear−EC90，DGROUP 基底 linear24DD0。10387(file16F7,803E260701) 讀取消 byte0726；1038C(file16FC,7434) 指向103C2、541及獨立讀鍵。這是待本輪正常來源的 strong，不能只憑靜態分支升為 confirmed。

現行 rcMenu Cancel 直接 active=false。來源接受、正常紅測試與 READY 審查前不改正式程式。可能沿既有 FarewellTextID／rcFinalWait，不先猜補資料欄位。範圍僅上述正常空名冊返回；非空分離、滿隊、原版 Save／Load、音樂與完整 campaign 不外推。

## 招募主選單 Esc：證據審查與有限 READY

正常原版193包／462 IRQ1來源已接受，`work/dosgolem-opening/issue4-recruit-menu-esc-normal-r1-source-r1-receipt.json` SHA-256 `137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c`。前191包、458 IRQ1及完整PNG／bin與空Join Yes父來源逐項相同，174份母親父來源保持。實際公開producer、完整Go、binary及runtime身份固定；沒有狀態注入或重擲。來源核對入口 [verify_dosgolem_recruitment_menu_cancel.py](../tools/verify_dosgolem_recruitment_menu_cancel.py)，依序傳原版產物目錄、實際producer與全新receipt。

- confirmed：191選Yes回528主選單，192按Esc。10387／1038C讀取消byte0726為1，經103C2選541、原生文字bank2826、停21133獨立讀鍵。193新確認後103D6返回field1997C、座標2,18。之前189選Join時同欄為0；名冊、四個隊伍指標、主角128bytes、金錢及旗標保持。
- IDA9.4有界匯出478條目，原始bytes／MZ relocation核對，`work/issue4-recruit-menu-esc-r3-ida.json` SHA-256 `4d8f65947c016edc53cb21156ee3bf8a43cc621c20295010f4831cd1cd300ff6`。1F779..1F908是原始選單讀鍵consumer；1F8E1(file10C51,C606260700)與1F8F6(file10C66,C606260701)分別寫DS:726h零與一。IDA該段DS未知，直接xref不能列到DGROUP；以實際caller的DS15ed和上述動態值閉合，不因xref缺項稱沒有writer。writer的精確PC尚未動態觀測，該項仍strong。
- 修正前正常InputState的第一個新blocker在192，明確缺告別等待；元件三個游標均RED。保留 `work/issue4-recruit-menu-esc-red-r1/`。沒有產品assertion以外的失敗，oom_kill0。
- READY：rcMenu Cancel在typed selection存在時，追加其既有FarewellTextID，再以rcFinalWait獨立等待；本次Esc只啟動告別，不能同包關閉。下一個新按鍵才還原caller返回場景，held key不消費。無名冊交易、RNG呼叫、新文字、幾何或schema欄位。缺selection契約保持原狀，不猜fallback。
- 垂直鏈：原始541→現行pack文字引用／validator→共用文字consumer→正式InputState／完整RGB→snapshot／同版本Save／Load及下一步。原版Save／Load、非空分離、滿隊、音畫與全campaign保持未知；有限狀態與全RGB分開驗收。

## 招募主選單 Esc：有限 CONFORMED

- 正式主選單Cancel追加既有FarewellTextID，以rcFinalWait保留告別，不在本次Esc關閉。正常193包、逐包snapshot／RNG、同版本Save／Load與下一步通過，三個游標元件及held key邊界通過。有限E2／E3、畫面V2；原版完整流程未CONFORMED。
- 41個不同招募頂層、28子PASS，零SKIP；正常新遊戲至THE END95.45秒，合計42不同頂層，desktop Linux x86_64通過。最近完整game466及11個internal保持c572ab3，本輪按比例不重跑無關套件。
- 十三條既有路線789 PNG逐byte保持，新44 PNG留本機，150..191等於已接受Join Yes。187..192完整RGB各295，193返回0；目視192告別文字及底圖相同，295個差異位置與兩側RGB逐項等於父選單191的已知人物差異。沒有裁切、遮罩或指定人物相位；完整V3仍RED。
- 八種壞來源均拒絕，正對照前後相同，oom_kill0。保留27筆原始定位，新增10387／1038C的有限confirmed與1F8E1／1F8F6的strong。IDA9.4自動合併31筆，478條目原始bytes／MZ relocation核對；writer精確PC未知的警示保留。
- 九份JSON與67f6ff8逐byte相同，schema0.17.0／content0.1.89、canonical `sha256:4b235d635e29011f232b212b533e5de32c29c32a8d094c236eeba68198b9b988`保持。沒有新資料契約、存檔遷移或發行包。原版素材、PNG、EXE、database與私有work不加入Git。

主線驗證訂正：首輪全部41招募測試已綠，主線在正式寄放隊員後仍按舊假設連按Esc、立即導航。修正後第二次Esc正確進告別，因此導航鍵被告別等待消費。只在opening_input_trace_test.go補正常文字完成與獨立Confirm並確認UI關閉，保留首輪log，以正式新遊戲乾淨重跑THE END通過；正式產品不再修改。這證明remake主線可玩，不驗收原版非空分離。

工具環境紀錄：第一次READY命令有多餘加號，Python在寫入前拒絕；改為檔案化腳本完成。PNG稽核容器無Pillow，改用既有Go標準PNG解碼；首輪缺Go平行度限制造成fork失敗，補GOMAXPROCS2與-p2後乾淨重跑，原版／產品及圖片不變。沒有用環境假失敗調正式參數。

完整本機收據：

- 修正前RED：`work/issue4-recruit-menu-esc-red-r1/receipt.json`，SHA-256 `f4963c33cab6687995d6d563b949b35d034824f62d06860ea6f0198772ed8776`。

- 正常remake：`work/issue4-recruit-menu-esc-validation-r1/empty-menu-cancel/receipt.json`，SHA-256 `2622635de582e2035bd717b1a2086c02ff694c37b40cbdab0ead9fbec3b36e32`。

- 招募與主線：`work/issue4-recruit-menu-esc-retry-r1/game-receipt.json`，SHA-256 `2b9a97e767fbf0cd530b539d15ce447d65de5bd52aaadf39eb3b6119b767bb23`。

- desktop：`work/issue4-recruit-menu-esc-retry-r1/desktop-receipt.json`，SHA-256 `4a8baacc2f529c9c7fa064d2ea8fce02fd7afb5b50a7160dfd04ea5d5ab99411`。

- 來源負例：`work/issue4-recruit-menu-esc-audit-r1.json`，SHA-256 `638d6f7de1cf4539c7a20dcda7821a726794ec275ab4152d7aa4a48769232c39`。

- 畫面保持：`work/issue4-recruit-menu-esc-images-r1.json`，SHA-256 `f182c0d289288c5e402214e9985462979cb3908708bc7c73611f49c7f7fc78ba`。

- 全部RGB差異：`work/issue4-recruit-menu-esc-difference-r1.json`，SHA-256 `e6ea0ad6011540b3e89d4ba29d6e758e7520cc03d1e6babdc0f006ac3575859b`。

- IDA31筆：`work/issue4-recruit-menu-esc-r4-ida.json`，SHA-256 `ff1fd4972b334667d0545c060c6a45d29cb1069915102ec252fc5553e0da7f32`。

下一個正常玩家節點為原版F5／F6存讀檔；先驗可寫overlay及素材保護，再取得正常UI／恢復狀態來源。入隊播放後返回、非空分離、滿隊、音畫及完整原版campaign仍未知，硬體driver／ISR停止線維持。Goal與Issue保持進行中。


## 2026-10-04 正常 F5／F6 存讀檔 DRAFT

從已接受的招募主選單 Esc 正常193包來源續行。公開入口 `tools/dosgolem_save_load_probe.py` 使用固定2f44a68快照的既有 `DOS.Scratch`，原始assets_raw唯讀，暫存層位於本機work。先送194包F5，停在第一個原生完成或等待點，保存原生FileOps及完整PNG／bin。

- 已知：現行正式InputState沒有F5／F6；既有Save／Load與標題讀檔不能代替快捷鍵正常路徑。
- 未知：原版存檔選單、確認、檔名／格式、讀檔入口及恢復朝向。未達READY不更改正式行為或資料格式。
- 初始seed1357執行前固定一次；193包之前須與已接受來源逐項核對。沒有狀態注入、跳過等待或改時鐘。
- 存檔→移動→F6讀檔及完整狀態恢復尚未驗收。原始產物及SAV留本機，不加入Git。

續行探針 `tools/dosgolem_save_load_continue_probe.py` 依IDA9.4原始bytes建立DRAFT：253確認→No→252等待→field，或Yes→250選10槽→寫入。只送正常鍵，音樂等待有界觀察，未完成不跳過。IDA sidecar `work/issue4-save-f5-r1-ida.json` SHA256027cf5e7e453748fc933cbf02040e8aab5a28651432ecd9fdd24b6b828492bba，528條目，原始MZ bytes核對，語意尚待動態閉合。

完整往返DRAFT入口 `tools/dosgolem_save_load_roundtrip_probe.py`：F5 Yes→第一槽→252另按確認→正常左移→F6→第一槽讀取。193..200保存完整DGROUP4F29..57A5的2172bytes與原生caller／FileOps，沒有模擬器snapshot restore或記憶體注入。先前save探針錯把196的252等待預設為field；實際已寫player.dat200bytes與dragon0.dat2172bytes並到原生等待，該探針會以有界診斷退出。這是DRAFT終點設計缺口，不能誤稱產品或dosgolem失敗。

取消邊界DRAFT入口 `tools/dosgolem_save_load_cancel_probe.py`：原版113CF只檢查原始0722選中序號，不檢查0726取消byte。假說為Yes游標按Esc仍開選槽，再於選槽Esc取消並接252等待；以正常195／196兩次Esc查證，不將假說放入production。


### 正常存讀檔有限 READY

原版來源已接受：F5開窗194包／464IRQ1 `3350c30f`；No197包／470IRQ1 `08a3fead`；Yes存檔、行走與F6讀檔200包／476IRQ1 `0a88466e`；Yes游標Esc後選槽Esc197包／470IRQ1 `30a4f547`。獨立工具 `tools/verify_dosgolem_save_load.py` 與 `tools/verify_dosgolem_save_load_flow.py`。所有來源固定seed1357一次，沒有CPU記憶體注入或模擬器snapshot restore；F6是遊戲自身Load。前194包與完整PNG／bin一致。

輸入 `assets_raw/DQ3.EXE`115282bytes，SHA2565178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c。dosgolem2f44a68；IDA9.4 linear、file=linear−EC90，DGROUP基底24DD0。

| 原始定位 | 語意及推論等級 | 消費與驗收 |
|---|---|---|
| 113A9..113FB／file2719..276B | confirmed：F5先逐隊員顯示251，再253詢問；Yes或停在Yes的Esc進選槽；No接252獨立等待 | 正常194..197，未寫檔No與兩次Esc均返回2,18 |
| 1D8E9..1D94B／fileEC59..ECBB | strong：class→DGROUP43D6門檻指標、level*4讀門檻，減角色+32/+34累積經驗；251顯示剩餘經驗，不發放經驗 | 正常本例顯示29，128bytes主角保持；其他角色／等級尚未動態抽樣 |
| 11423..114D9／file2793..2849 | confirmed：模式0、選槽、更新200bytes索引，再保存2172bytes持久區；原生cue18呼叫及208E2等待 | 真正player.dat／dragon0.dat落地，196顯示252，197確認返回；不改硬體時鐘或跳過播放 |
| 114DA..1165F／file284A..29CF | confirmed有限：模式1只列已存槽，選第一槽native2172bytes讀取及場景重生 | 正常198左移至1,18，199 F6選槽，200返回2,18；200完整持久區與保存檔逐byte相同 |
| DGROUP3FBA／linear28D8A／file1A0FA；11660..11777 | strong幾何／confirmed十槽：x19byte、y30、寬42byte、height208；465兩行header、466每列、467footer；slot數10、rowY62／step16 | 正常195及199完整選槽；F6少量槽、空槽與其他槽仍需抽樣 |

正式實作契約：新增跨版本SaveMenu／LoadMenu抽象輸入並綁F5／F6，只有一般field開啟；已有modal先消費鍵。F5顯示每位成員剩餘經驗及確認，Yes游標Esc仍開選槽，No游標Esc視同No。存檔及選槽取消均追加252，再等新鍵返回。F6讀取成功直接返回，拒絕或不相容Load保留選槽且不消耗持久狀態。十槽、文字、視窗與原始控制碼均由typed pack提供。沿用現行JSON snapshot，第一槽保持既有savePath，其他槽使用獨立檔案；不換成原版binary、不新增自動遷移。

原版玩家既有PLAYER.DAT有十個非空槽。remake先前只有一個JSON存檔，初始存檔資料不等價；正式正常驗收須明示先前存檔fixture或保留這項比較限制，不以不同初始槽資料宣稱整張V3。原版2172bytes保存196／197與Load200完全相同；198／199僅4F33由2變1，Save前193..195只有三個先前備份欄位不同。朝向、日夜、其他槽、複數隊伍與完整聲波仍未知，不外推此單人白天來源。

驗收包括正常InputState整段、No／Esc不寫檔、不改RNG及snapshot、真正寫入與Load恢復、失敗保留UI、held edge、正常下一步、schema／引用拒絕、原始EXE／DAT parity及完整PNG差分。音訊時長採既有hardware-spec approximation，未聲稱原版wall-clock／波形。READY只允許已列契約，完整V3與campaign仍未CONFORMED。


### 2026-10-04 存檔音效與選槽 consumer 補核對 READY

IDA9.4 匯出 `work/issue4-save-sound-r1-ida.json` SHA25666e1431716b0cf2143d68973093628bdc5b80a246d1df41634111150791dd21e。原版linear1143F BP12h，20770比較BP<1Eh後走DGROUP253E的VOC表及22CF5；207CD才轉EBG。cue18因此使用既有VOC音效與來源sample duration完成閘門，沿docs/123、149公開規格契約，標為hardware-spec approximation，不稱為原版逐週期wall-clock。

姓名consumer linear215EE..21650匯出SHA256067899446fb91b2842df9271261ef5bab1da7ac111ae9bf2c2a5fdb38f375579，保留BP/DX/SI，限定四glyph，沒有更新DGROUP716；因此性別anchor為原始BP21+8+24 bytes，即X424。號碼以BP-4、五digit右對齊，X136是數字欄虛擬起點，實際1..10畫在X200及X184/200；不能誤把X136當作越界實畫。

本批實作入口為 `game/field_save_load.go`，typed契約與raw parity為 `internal/gamepack/field_save_load.go`／`field_save_load_test.go`；重建九份JSON使用 `tools/migrate_field_save_load_pack.py`。正式SaveMenu/LoadMenu只在一般field接收鍵盤邊緣，既有modal優先；第1槽沿用現行savePath，第2..10槽使用同一JSON格式的獨立檔名，無DOS binary import或資料格式變更。


### 2026-10-04 F6讀回色盤 DRAFT→有限READY 補證

正常200包讀回位置、角色、旗標、2172bytes持久區與remake JSON皆吻合，但完整畫面差45109像素。`work/issue4-field-save-load-rgb-r1.go`逐點統計只有12組顏色對應，沒有幾何或人物位置差異。原版紅235/52/0對remake255/73/0、灰186/195/195對207/215/215，逐項吻合DQ3.PAL bank1對bank0。沿docs/136既有已證實bank selector回查，不能調RGB百分比。

原版linear1160B寫DGROUP251D=120，11617比較已載入remembered-world Y DGROUP4F31與300；upper-world分支1161F寫526C=1、11625寫251D=0，1162B呼叫1EE9B重新選bank。日夜時鐘在4F29..57A5以外，原版DRAGON0.DAT不存它。正常上層F6因此必須重設clock0選bank1；本輪單人200包與palette observer閉合為confirmed／D3。下層world層號映射沿既有共用world parser，clock120與night分支僅原始bytes已證實／D2，沒有下層正常F6畫面來源；不宣稱其正常parity。

READY：`field_save_load.load_clock_rules`以具名world layer引用clock值及逐項證據。上層layer0→0；下層layer1→120保存已知静態值，後者列驗證限制。數值鎖在pack JSON，超過cycle、重複layer、缺失、null或未知layer拒絕，不提供Go數值fallback。F6才套用這個原版重畫契約；既有engine Save/Load及標題Load仍保留JSON時鐘，格式與舊schema拒絕策略保持。正常F6驗收比較2172bytes對應持久狀態，另核對clock0、palette bank1及完整PNG；不能要求原版未保存的clock等於remake JSON原始值。

元件測試另有兩個非產品原因：空集合的Go表示與JSON省略需以存檔bytes比較；不完整overworld fixture的curCty0在restore正規化為-1，改成合法metadata。原版200包後Go移動仍有既有cooldown，正常閒置至0後下一步才抽驗，不以座標注入驗收。


### 2026-10-04 空名冊讀回 READY 與元件RED

正常原版F6以AH3F直接覆寫完整4F29..57A5的2172bytes，名冊與所有角色slots亦在此區；不能只還原非空集合。remake JSON的Roster欄位同樣表示存檔當時名冊，空集合有明確意義。元件測試保存空名冊／空同伴後，建立讀檔前非空名冊／同伴；壞schema必須全部保持，有效F6必須還原空集合。r7於`work/issue4-field-save-load-validation-r6`捕捉RED：Load成功、UI已清、RNG保持，但存檔bytes不符。既有restore對companions先清nil，對roster只有len>0時賦值，因而保留讀檔前名冊。

有限READY：restore無條件清roster後再依已驗證snapshot重建；不改JSON格式、角色交易或RNG。此為engine完整還原驗證，正常來源仍限單人存檔→移動→F6；不把元件狀態注入稱為原版正常玩家流程。

### 2026-10-04 正常 F5／F6 有限 CONFORMED

正式冷啟動三條路線與六個元件測試均通過。原版來源四份保持，固定 seed1357 執行前設定一次；原版及remake逐包持久狀態、RNG、No／Esc不寫檔、native Save／Load恢復及正常下一步通過。empty roster元件RED修正後通過，該元件不是原版正常路線。

| 正常包 | 完整RGB差異 | 驗收界線 |
|---|---:|---|
| 194 F5經驗及確認 | 0 | 指定完整畫面V3 |
| 195 十槽 | 4064 | 原版十筆既有索引、remake空JSON，初始槽資料不同 |
| 196 存檔後252等待 | 0 | VOC duration完成後才接受確認，硬體規格近似 |
| 197 確認返回 | 0 | 不更動角色、旗標或RNG |
| 198 正常左移 | 356 | 人物動畫差異保持 |
| 199 F6選槽 | 32847 | 原版十筆、remake一筆，初始槽資料不同 |
| 200 native Load返回 | 0 | clock重設0、bank1；45109色盤差異已修正 |

No路線194..197全RGB均0；取消路線195差4064，其餘194／196／197均0。沒有遮罩、裁切、指定phase或重擲。完整RGB原圖與差分留在本機，未加入Git。有限E2／E3及指定畫面V3，不宣稱全流程V3或完整原版campaign。

schema0.18.0／content0.1.90，canonical `sha256:9d6325addef6db4d6f4c049f44fe517e61d2a0550724c67d5ae7ec0f649a3917`。九份JSON從乾淨1bef3d73556f9af9043825c28f32042982971ce6重建逐byte相同；schema／引用／raw EXE／DAT parity通過。完整game478覆蓋、431不同頂層／117子PASS、47選用SKIP；internal169頂層／363子、11套件PASS、4選用SKIP，沒有素材缺失跳過。正常新遊戲至THE END201.00秒、desktop Linux x86_64通過，oom_kill0。十四條舊路833張PNG逐byte相同。

| 本機收據 | SHA-256 |
|---|---|
| `work/dosgolem-opening/issue4-save-f5-normal-r1-source-r1-receipt.json` | `3350c30fa5c48fc690b144a6860cac155a4e02efb498649b7120a9bbd5f0973d` |
| `work/dosgolem-opening/issue4-save-f5-decline-r1-source-r1-receipt.json` | `08a3feadd29c2567a806d3bd0bda121214a5a57323bf993f969a0d2cbda963d2` |
| `work/dosgolem-opening/issue4-save-load-roundtrip-r1-source-r1-receipt.json` | `0a88466eab206d7a4b346b8229bbcebd10683044aa363cb346bb81e173e3d238` |
| `work/dosgolem-opening/issue4-save-load-cancel-r1-source-r1-receipt.json` | `30a4f5471277bb040d4dfcf62958098ac412e11d5d641235f9cf378c2fcbd32c` |
| `work/issue4-field-save-load-validation-r7/test-receipt.json` | `30cc21e32af3d12074634a2eeb96d99049c24991345e94f9d476874164de7275` |
| `work/issue4-field-save-load-full-r1/game-receipt.json` | `0a577a525ddf2575389dffffd73d3b081727af4c6d449bfda6f6ba596823c264` |
| `work/issue4-field-save-load-full-r1/internal.log` | `9b1c08c21c425c2869e4864c468281d167cebc8bf74b69ed91ff67a2bbed61c6` |
| `work/issue4-field-save-load-full-r1/desktop-receipt.json` | `74f5e43c622fbcffdb971b0465fbf6fa604ba8de234bcf1364b565d7378c95fc` |
| `work/issue4-save-load-source-negative-r2/receipt.json` | `2bf00c0b2bf3435a5b2d9c505c67855364a9842201bc13aabf37e59ca145f9d9` |
| `work/issue4-save-load-reviewed-r3-ida.json` | `79a9be3eed4814132970eda34dd41181fec8506f7618c9a28b487563d47d3435` |
| `work/issue4-field-save-load-pack-rebuild-r2/rebuild-receipt.json` | `a2471ac9cc46ae5492a86d47731ce42110bc12f4893ad3c040b124e9e8982fa6` |

公開原版產生器：[入口](../tools/dosgolem_save_load_probe.py)、[No／存檔續行](../tools/dosgolem_save_load_continue_probe.py)、[往返](../tools/dosgolem_save_load_roundtrip_probe.py)、[兩次Esc](../tools/dosgolem_save_load_cancel_probe.py)。獨立驗證：[入口checker](../tools/verify_dosgolem_save_load.py)、[流程checker](../tools/verify_dosgolem_save_load_flow.py)。用法沿各檔CLI，只能在Docker以固定dosgolem2f44a68快照執行，原始assets_raw唯讀、明示Scratch輸出；支援範圍為本節列出的正常單人路線。

公開實作：[共用狀態機](../dq3_remake_ebitan/game/field_save_load.go)、[正常路線與元件測試](../dq3_remake_ebitan/game/field_save_load_test.go)、[typed契約](../dq3_remake_ebitan/internal/gamepack/field_save_load.go)、[原始parity／壞契約](../dq3_remake_ebitan/internal/gamepack/field_save_load_test.go)。[九份JSON重建器](../tools/migrate_field_save_load_pack.py)的三個參數為乾淨1bef3d7 pack、原版素材目錄及既有glyph_unicode_map；詳細欄位見[docs/84](84-game-pack-json-contract.md)。

IDA入口：[非破壞匯出](../tools/ida_dump_field_save_load_contract.py)、[12筆原始位址ledger](../tools/ida_field_save_load_ledger.json)。以既有IDA9.4 locked-v1、UID1000、HOME=/home/ubuntu與image內建私有設定，唯讀原始EXE，工作副本及database置/tmp；`idat -A -o/tmp/view.i64 -S"/repo/tools/ida_dump_field_save_load_contract.py /work/output.json" /tmp/view.exe`，有界一次性Docker輸出必須非空、493條目、12筆分級註記且原始bytes／MZ relocation一致。不rename原始symbol，未審查定位保持unknown警示。授權與database不加入Git。

環境紀錄：同程序多個大型正常路線OOM已改為同binary分程序重跑；失敗紀錄保留。初輪internal少掛/assets_raw是掛載問題，補正後相同image／命令重跑。IDA reviewed-r2誤掛主機.idapro覆蓋image設定，靜默exit1且無sidecar；依正式工具契約移除該掛載，reviewed-r3相同匯出493條目通過。未將這些環境失敗寫成產品缺陷。

下一合法checkpoint是原版200包後正常行走，以及可比初始存檔資料的選槽UI。下層clock120只有D2；其他world／室內、複數隊伍／等級、完整槽資料、聲波、人物動畫、入隊播放後返回、非空分離及完整campaign仍未知。driver／ISR停止線維持，沒有新發行包。

### 2026-10-04 F6 返回後正常行走 DRAFT

依 Issue #4 從合法200包讀回場景續行。新入口 [正常行走探針](../tools/dosgolem_after_load_move_probe.py) 沿用已接受往返產生器，保留原版200包及所有輸入／IRQ1／PNG／bin，新增201左鍵與202右鍵，停在原生空讀鍵。另以唯讀觀察保存DGROUP251D時鐘與526C層狀態，不注入座標、朝向、動畫或時鐘。

Docker沿用dq3-ebiten-test:20260822-r1、固定dosgolem2f44a68唯讀快照及原始assets_raw唯讀；以 `python3 /repo/tools/dosgolem_after_load_move_probe.py --prefix issue4-after-load-move-r1` 建立獨立Scratch及本機收據。seed1357執行前固定一次，無模擬器snapshot restore。先驗證200包前綴與原生檔案保持，再審查兩步的座標、持久狀態與完整RGB；此DRAFT不修改正式Go或pack，也不預設動畫差異的原因。

獨立入口 [來源驗證器](../tools/verify_dosgolem_after_load_move.py) 先完整重驗已接受往返來源，再核對202包／480IRQ1、200包前綴、兩次按下／放開、2172bytes持久區、Scratch原生檔案及完整PNG／bin。參數依序為本機原版輸出目錄、producer路徑及尚不存在的收據輸出路徑；Go來源與binary雜湊固定於checker，來源未通過時不寫接受收據。

#### 正常行走來源有限 READY

`work/dosgolem-opening/issue4-after-load-move-r1-source-r2-receipt.json` SHA-256 `71a52768a26dabb174a3a96e53efe6aaebba78c7620b20cce39fe2f761548e10` 已接受。原版EXE、dosgolem版本與位址基準沿上節；新Go來源 `80dc4674639d14d390e8900c5775160500e675938054b0a43e6b69f872f3096a`、binary `b68bb2e0402b3e119e8d134149a4f236c517c2a994a109113034936e1c1b5e89`。新增240個正常輸入中的最後兩包，不改任何遊戲欄位；seed1357執行前固定一次。

- confirmed有限：200包讀回後，201左鍵到1,18、202右鍵回2,18；每包均在原生IDA linear1997C空讀鍵，raw0013沒有帶路flag。兩次按下／放開各送達IRQ1，總480次。
- confirmed有限：201完整2172bytes只改DGROUP4F33的低byte2→1；202整區與保存檔及200包逐byte相同。角色128bytes、旗標64bytes、金錢與場景保持；DGROUP251D時鐘兩步均0，526C低byte均1。
- confirmed有限：200包前綴的输入、IRQ1、狀態、PNG／bin及persistent保持；PLAYER200bytes、DRAGON2172bytes與native FileOps保持。沒有新增文字或檔案交易。
- 畫面線索：原版198與201同座標的完整224000色號逐byte相同，200與202的342個色號差異仍待remake逐點核對；不因此猜動畫規則。

初輪來源r1已通過原版稽核與八個負例，但只在meta內保存EXE身分。共用remake畫面核對要求頂層`original_sha256`，因此首輪在194報身分缺失；同一原版輸出由checker補齊四個頂層欄位，完整重驗另寫r2，不覆寫r1、不改原版或production。正常remake驗收入口為[共用F5/F6測試](../dq3_remake_ebitan/game/field_save_load_test.go)中的`TestFieldSaveLoadAfterLoadMoveDosgolemNormalInputComparison`，必須同時提供原版目錄及明示本機輸出，禁止素材缺失SKIP冒充通過。

#### 正常行走有限 CONFORMED 與畫面限制

正式新遊戲重播到202包已通過。201左移、202右移的完整JSON snapshot只按原版位置變化，其他欄位、clock0及RNG保持；202後engine Save/Load及正常下一步通過。正式Go與九份pack JSON逐byte等於e939db2；本輪新增原版來源、獨立checker與正常測試，未猜改產品規則。

原版packet在按鍵消費、整步完成及放開後停於空讀鍵。初輪一幀方向輸入在201仍有cooldown2，尚未完成整步。新測試只經正式DirHeld連續按住，到同一格界後放開，最多60更新，無坐標、動畫或clock寫入；當下輸入完成點可比，持續時長與CPU指令／TPS映射仍未知。這個有限狀態驗收不證明按鍵wall-clock或逐幀動畫完全一致，不以改cooldown讓測試通過。

完整RGB：200為0；201為356，完整差異範圍x291..541／y131..263，跨多處人物區；202為122，範圍x289..315／y180..191。未裁切／遮罩、指定phase或把差異當作通過。畫面V2，完整V3未知；差異的renderer／人物consumer尚未READY，留下一窄切片，不在本輪猜修。

10個命令、8不同頂層／4子PASS、零SKIP與OOM；147張既有F5/F6 PNG逐byte保持，新路線193..200與本輪roundtrip保持，兩個新步行PNG留本機。r2 checker的8種獨立壞來源全拒絕，完整父來源實驗正對照前後一致。最近完整game／internal／THE END／desktop仍為e939db2，未把本轮定向回歸升格全套。

| 本機收據 | SHA-256 |
|---|---|
| `work/dosgolem-opening/issue4-after-load-move-r1-source-r2-receipt.json` | `71a52768a26dabb174a3a96e53efe6aaebba78c7620b20cce39fe2f761548e10` |
| `work/issue4-after-load-move-validation-r3/test-receipt.json` | `776ae5245d91e76b06946ab489cfb18ca5cc1de39778d489f6f0fc3f83fbb0ca` |
| `work/issue4-after-load-move-validation-r3/after-load-move/receipt.json` | `5a5dd70d540a90245cf698ed7836ab9249e34de56d129eb8200887662bd23405` |
| `work/issue4-after-load-move-negative-r2/receipt.json` | `cb39e28857a70618e1a47092a42f568bf30030f0a37417ca140c51800c0d2428` |
| `work/issue4-after-load-move-rgb-r1.json` | `e3b8ae6d338dfa986aa70a61098256ccaced7565325edd74a4486dc34cf7c71d` |

重生入口為本節兩份已索引probe／checker；remake以`DQ3_AFTER_LOAD_MOVE_ORACLE_DIR`指定原版目錄，`DQ3_FIELD_SAVE_LOAD_RECEIPT_DIR`指定空輸出目錄，於Docker/Xvfb執行`TestFieldSaveLoadAfterLoadMoveDosgolemNormalInputComparison`，具備原版assets_raw才算驗收。seed原版1357執行前一次；remake PRNG於正常193checkpoint沿既有seed設定自然續行，兩步無亂數判定，不要求閒置過程的原版自然RND次數對齐。

### 2026-10-04 正常移動人物影格 DRAFT

沿 Issue #4 的合法202包來源，新增[人物影格唯讀探針](../tools/dosgolem_field_pose_probe.py)。在原版 IDA linear1E307 觀察原始 BX 影格表偏移、SI／DI、DGROUP0004與26F0；原生空讀鍵另保存0004、26AD、26B0、26D0／26D1。用法為同一有界 Docker 工具鏈執行 `python3 /repo/tools/dosgolem_field_pose_probe.py --prefix issue4-field-pose-normal-r1`，固定dosgolem2f44a68、原始EXE唯讀及新Scratch。只讀觀測，不改相位、CPU、時鐘或等待。

IDA9.4 非破壞匯出沿相同115282bytes EXE與5178fdc8完整雜湊。位址基準為IDA linear，file=linear−EC90，DGROUP基底24DD0。`work/issue4-field-pose-r5-ida.json`723條目、原始bytes與MZ relocation核對，SHA-256 `4fb6caf1682b9f9e9cff4304eedcc589bc5c1bb5100aa75af474cf3934813be2`。

- strong：119B8→1DC93的正常移動繪圖，1E057保存方向，1E1A4→1E2EF於1E2F1..1E307以角色、方向與DGROUP0004選取2B7A影格表。仍待正常動態consumer閉合。
- 訂正候選：主迴圈194BC→1EE23處理日夜時鐘及色盤，未當作人物renderer。r1匯出保留，沒有推測性改名。
- 未知：原版與remake的可比動畫時鐘與各consumer取圖時刻；本節不讓自然時鐘差異偽裝成V3通過，也不延伸ISR／PIT硬體研究。
- remake正常測試只新增實際hero walk／facing、更新數、cooldown及NPC影格收據，不修改正式Go、pack或測試狀態。原始PNG仍逐點核對，舊畫面必須保持。

[獨立人物來源checker](../tools/verify_dosgolem_field_pose.py)先完整重驗已接受202包父來源，再核對新probe／binary固定雜湊、全部原有DQ3事件、202包PNG／bin／persistent及原生存檔逐byte保持。三個CLI參數依序為原版輸出目錄、producer及尚不存在的接受收據；另核對1E307的原始BX偶數表偏移及0004低位。未完整通過不寫接受收據。

#### 人物影格 consumer 與完整差異已證實，動畫時鐘仍 DRAFT

新正常來源363f8f70已接受：202包／480IRQ1、94次1E307唯讀consumer觀測；全部原有DQ3事件、202包PNG／bin／persistent、native FileOps與原生SAV均和71a52768逐byte保持。沒有重新載入snapshot、遊戲狀態注入或時鐘寫入。原始步伐位元0004在201完成為1、202為0；最後consumer的BX分別0006／000C，原始BLS影格為3／6。原版方向為下、左、上、右，Go decoder重排為下、上、左、右，不混用方向碼。

[完整畫布診斷](../tools/verify_dq3_after_load_sprite_raster.py)唯讀核對原始BLS／CTY／BLK、完整PNG與remake實際影格收據。沒有產生替代圖、裁切、遮罩或修改phase；逐一檢查全部224000像素的差異，另核對三個完整人物含透明像素與底圖。

| 正常包 | 原版／remake hero步伐 | 英雄差異 | NPC14差異 | NPC15差異 | 未解釋差異 |
|---|---|---:|---:|---:|---:|
| 201 左移 | 1／0 | 127 | 106 | 123 | 0 |
| 202 右移 | 0／1 | 122 | 0 | 0 | 0 |

本節confirmed限於兩張自然取得的畫面及原始consumer。完整RGB仍356／122，完整V3未通過；原版與remake的可比動畫時鐘、取圖瞬間仍未知。畫面全部由實際原始影格解釋，不能據此決定正式動畫的週期、初相位或更新規則。本輪正式Go及九份JSON與e939db2保持，不新增猜測的production設定，不為phase-only樣本開硬體driver／ISR切片。下一個正常切片回到可比初始十槽metadata的F5/F6選槽UI。

10個命令、14筆PASS紀錄，涵蓋8不同頂層與4個子測試，零SKIP／OOM；200張既有完整PNG逐byte保持。本輪按比例重跑F5/F6與正常兩步，不冒稱全套；最近完整game478、11個internal、THE END201.00秒與desktop仍為e939db2。新source checker的9種損壞及raster的4種損壞均拒絕，兩者正對照前後一致；source完整父來源前後重驗，中間負例只快取已完整驗證的父資料，沒有拿快取代替父來源驗收。

IDA入口為[非破壞人物匯出](../tools/ida_dump_field_pose_contract.py)及[兩筆原始位址ledger](../tools/ida_field_pose_ledger.json)。同一locked-v1 image、UID1000／HOME=/home/ubuntu，原始EXE唯讀、工作副本與database置/tmp。`idat -A -o/tmp/view.i64 -S"/repo/tools/ida_dump_field_pose_contract.py /work/output.json" /tmp/view.exe`，723條目、2筆confirmed consumer自動附註、原始bytes與MZ relocation逐項核對；不改原始symbol，其他候選保留unknown警示。這個sidecar只證明已列consumer，不構成動畫時鐘READY。

完整畫布診斷CLI為 `python3 /repo/tools/verify_dq3_after_load_sprite_raster.py --assets /repo/assets_raw --original /work/dosgolem-opening --source-receipt /work/dosgolem-opening/issue4-field-pose-normal-r1-source-r1-receipt.json --runtime /work/issue4-field-pose-validation-r1/after-load-move --output /work/issue4-field-pose-raster-r2.json`。只能在既有有界Docker執行，原版與runtime輸入唯讀、輸出明示且不存在。支援範圍為本節固定來源與兩張正常完整畫面，其他角色／場景需新來源。

| 本機收據 | SHA-256 |
|---|---|
| `work/dosgolem-opening/issue4-field-pose-normal-r1-source-r1-receipt.json` | `363f8f7002fb97254fe6bd8fcac8d4725aad07401dccfb39e667be0576d5dad3` |
| `work/issue4-field-pose-validation-r1/test-receipt.json` | `8a8f1806ad2aa9303fae38831f364e1c4e9fa3720c7af53b7b53f0de07ac0a89` |
| `work/issue4-field-pose-validation-r1/after-load-move/receipt.json` | `f289c066b424a7886beb4d523a8adf815ea79df7d74bb1a52053f7c92b461d71` |
| `work/issue4-field-pose-validation-r1/preservation-receipt.json` | `1a6b40103ee08735182f4801a182b4f1aedabdb28207f757bf65bcb515b11ba4` |
| `work/issue4-field-pose-raster-r2.json` | `9c8a9a38eb6f24e2099f1bbf43105fa2a11d4a2b65c28377aabc25407a54b78c` |
| `work/issue4-field-pose-negative-r1/receipt.json` | `d6ed8912b0763b3cc76feded4c10117c743adb422ee9fa585b64d09199b4d128` |
| `work/issue4-field-pose-reviewed-r1-ida.json` | `680aaf8f92517c4f11f83cf2ed5e179694b7fd01ebbaefa1529ea1404f176819` |

環境紀錄：間接方向表11BA2沒有IDA自動函式邊界，r3停止且保留error sidecar；後續只匯出有界raw bytes並維持unknown。原始人物consumer的direct DS xref缺項，保留原運算元與動態DS15ED，不宣稱沒有reader。最初沿用147舊PNG計數，本輪實際含新兩步共200張，以實際清單核對全部相同。這些是分析／驗證限制，沒有為它們修改產品。

### 2026-10-04 初始十槽metadata可比對拍 DRAFT

依Issue #4下一切片，沿既有正常202包來源，準備等價的初始槽顯示資料。原始`PLAYER.DAT`200bytes，SHA-256 `a445a11f52a6711aba1433d9107d10284253e08d27aa6be2b82dc7aec91376dc`。每槽20bytes：姓名count word、四個32bit值、level byte18、gender byte19。原始姓名215EE限制四字，低wordFFFF時引用D3字模；其他值由21CD7→21CF0把CX:DX作CHINA.FON seek，再讀32bytes至DGROUP2734並交2121C。原版FileOps已保存正常195十槽讀取：22585、164596及101166三個位置、每次32bytes，成功且沒有缺檔。

IDA9.4有界sidecar `work/issue4-slot-metadata-r1-ida.json`260條目，SHA-256 `24318794177dae910fa944f1ba567f8bae9888423ee5ee3fa4857d53c71f72ee`。原始EXE115282bytes／5178fdc8完整hash沿上節，IDA linear、file=linear−EC90；保留原始名稱、運算元及MZ relocation。116D0讀+12h判空；116FE畫姓名；11707讀level；11711..1171B以gender減1引用534／535。fixture採原版1／2映射JSON0／1，不混用raw值。

`CHINA.FON`192200bytes，SHA-256 `42b769cc51857fc14f4e820f937c46a0d254c04374fe52870d59c6eb24f7bb49`；`D3TXT00.FON`47232bytes，SHA-256 `c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`。三個32bytes字模與既有D3字模106／144／303逐byte唯一匹配，不用Unicode近似或外加字型。

測試入口：[槽metadata fixture](../dq3_remake_ebitan/game/field_save_slot_metadata_test.go)及[正常F5/F6測試](../dq3_remake_ebitan/game/field_save_load_test.go)的`TestFieldSaveLoadComparableSlotsDosgolemNormalInputComparison`。在合法193checkpoint前置十個現行JSON檔，只更動檔案中的可見姓名／性別，保留可表示的level1；不改執行中的遊戲、RNG、phase或座標。之後仍由正常F5保存、正常F6讀回剛寫入的第一槽。這是既有存檔資料前置條件，不是DOS binary import，也不表示其他九槽的世界狀態等價。

先要求195十槽完整RGB零差異，再記錄199及其餘完整畫面；人物動畫差異照常保留。初始存檔metadata相同與完整存檔相同分開記錄，不能讓fixture製造完整V3或原版存檔互通聲明。若實際畫面推翻現行consumer設定，再回RE／READY修production。

#### 十槽顯示資料有限 READY／CONFORMED

原版正常195的FileOps、IDA原始116D0／116FE／11707／11711 consumer與三個逐byte唯一字模已閉合。confirmed範圍限十槽可見姓名、level1與raw gender1；其他九槽的完整世界狀態未知。測試在合法193checkpoint建立外部JSON前置資料，執行中的完整snapshot及RNG保持；之後以正式InputState完成F5保存、左移、F6讀回第一槽、兩步行走、同版本存讀檔及下一步。

| 正常包 | 可比槽資料完整RGB差異 | 有限結論 |
|---|---:|---|
| 194 F5問題 | 0 | 舊畫面保持 |
| 195 F5十槽 | 0 | 初始可見metadata可比，限定此完整畫面V3 |
| 196 告別等待／197返回 | 0／0 | 保存、完成閘門及新確認保持 |
| 198 左移 | 356 | 既有人物影格差異保持 |
| 199 F6十槽 | 123 | 全為窗外NPC15，完整V3仍RED |
| 200 讀回／201左移／202右移 | 0／356／122 | 正常持久狀態、clock與RNG保持，人物差異未消除 |

原先195的4064、199的32847不能當作產品缺陷：當時原版十槽與remake空JSON前置資料不同。本輪建立可比metadata後，195為0、199為123；這是驗證前置條件修正，沒有改正式renderer或把十槽資料永久寫入產品。`initial_slot_metadata_same_state=true`、`initial_storage_same_state=false`分開保存，未宣稱DOS存檔匯入或其他九槽完整等價。

[完整畫布checker](../tools/verify_dq3_slot_metadata_raster.py)固定接受來源363f8f70與本輪runtime58877b04，完整重驗416份來源產物及log。讀取兩張完整640×350 PNG；195逐點零差異。199核對原始CTY file8F..95的NPC15、DQ3MAN.BLS影格26／27、DQ31.BLK底圖及全部透明像素，再比較全224000像素，沒有任何人物範圍外差異。原版只讀邊界0004=1、remake實際walk=0；這個confirmed診斷不證明可比動畫時鐘。

重生測試以`DQ3_AFTER_LOAD_MOVE_ORACLE_DIR`指定已接受原版目錄、`DQ3_FIELD_SAVE_LOAD_RECEIPT_DIR`指定空本機輸出目錄，在既有有界Docker／Xvfb執行`TestFieldSaveLoadComparableSlotsDosgolemNormalInputComparison`。原始PLAYER、CHINA.FON、D3TXT00.FON缺失或hash不符均失敗，不以SKIP驗收。完整畫布CLI為 `python3 /repo/tools/verify_dq3_slot_metadata_raster.py --assets /repo/assets_raw --original /work/dosgolem-opening --source-receipt /work/dosgolem-opening/issue4-field-pose-normal-r1-source-r1-receipt.json --runtime /work/issue4-slot-metadata-validation-r1/after-load-move --output /work/issue4-slot-metadata-raster-r1.json`。原始資料與runtime唯讀，僅明示work輸出可寫；支援範圍為本節固定來源與兩張選槽圖。

兩批17命令／22筆PASS，涵蓋9不同頂層與5子測試，零SKIP／OOM。舊四條存讀檔路線200張完整PNG逐byte保持；可比路線53張中的51張與舊after-load-move相同，只有195／199因前置槽資料改變。四種損壞來源、runtime收據、runtime畫面及原始素材皆拒絕，正對照前後一致。正式Go與九份JSON保持e939db2，schema0.18.0／content0.1.90、canonical9d6325ad不變；最近完整game478、11個internal、THE END201.00秒及desktop仍為e939db2，不冒稱本輪全套重跑。

| 本機收據 | SHA-256 |
|---|---|
| `work/issue4-slot-metadata-validation-r1/test-receipt.json` | `e5a8f50fb12f9abe707cf51392bd80b77c2d0f732a0f6900a966c3a22de60882` |
| `work/issue4-slot-metadata-validation-r1/after-load-move/receipt.json` | `58877b0475c4618b650dc2e5ddc0099958912df146f7808d8c3659bcb6863ae6` |
| `work/issue4-slot-regression-r1/test-receipt.json` | `a2ac73052ce24ae864f37df36a675c9cca3b027fab5d5aea325204f8199a8f03` |
| `work/issue4-slot-regression-r1/preservation-receipt.json` | `6e4cd62929583968d923e62a5207a753e8942cbe8d6d7e5cc05320ea156d369b` |
| `work/issue4-slot-metadata-raster-r1.json` | `dc262e20c19ed736ec9a229a8db6359e4981f7ff7e337e1f76e6c9e9523b0d2d` |
| `work/issue4-slot-metadata-negative-r1/receipt.json` | `f2d17a1e8aed1a06805e3bf0ec0a16fa498cadddc2636e35ad52d002cb686181` |

環境／腳本紀錄：初版checker將全部manifest當作單層路徑，後改相對路徑仍誤把Scratch掛在原版輸出目錄內。連續兩次失敗後重查dosgolem路由與既有source checker，原生Scratch實際位於輸出目錄的同層；依既有契約修正，同一命令乾淨重跑通過。PNG計數初輪只glob頂層，後改rglob確認53張；收據首寫誤用唯讀/repo掛載，改明示/work後通過。這些是驗證腳本問題，未修改產品或原版產物。

下一合法垂直切片是正常入隊短曲播放完成後返回選單，從已接受的原生音樂等待入口續行，先證明完成與新按鍵再實作缺口。可比動畫時鐘、其他F6場景、複數隊伍、非空分離與完整campaign仍未知；不重跑已解釋的phase-only樣本，不深挖硬體driver／ISR，沒有新發行包。

### 2026-10-04 入隊短曲完成後正常返回 DRAFT

依Issue #4從正常199包、538返回與208E2音樂等待入口續行。[原版正常返回探針](../tools/dosgolem_recruitment_return_probe.py)沿已接受party來源d0f6428d的冷啟動、seed1357及輸入，保存199包完整畫面後繼續執行原始EXE。只在玩家層10459／1045E／10469／10398／103AE／103B6附加唯讀觀察；不改音樂完成旗標、CPU、phase、clock或等待。

播放器返回後若正常到540選單，再用既有實際Right／Enter選No、541告別的新Enter返回場景。先完整核對199包前綴與角色副本保持，再接受自然完成及後續收據；未知分支不猜補。本輪不追driver／ISR，硬體時長與合成音色近似沿既有READY契約，不宣稱逐波形或wall-clock parity。

沿固定dosgolem2f44a68、dq3-ebiten-test:20260822-r1，原版與repo唯讀，UID1000只寫既有work。執行入口為 `python3 /repo/tools/dosgolem_recruitment_return_probe.py`；輸出前綴issue4-recruit-return-r1，已有收據時拒絕覆寫。接受之前不改正式Go或pack。

2026-10-05 r1在原版自然出現raw0013 bit4000時，被沿用的登錄所probe guard停止。這個停止點不能證明音樂完成或工具能力缺口；較早2812行的入隊線索亦記錄此旗標未知。前199包全部DQ3事件與398份PNG／bin逐byte等於d0f6428d；私人前綴收據`work/issue4-recruit-return-prefix-r1.json`保存這個有限結果。

r2只把guard的適用範圍限制在原先前綴；199包後保留原生4000並唯讀記錄首次出現及有界進度，不清旗標。輸出改issue4-recruit-return-r2，r1 producer逐byte另存`work/issue4-recruit-return-r1-producer.py`。總指令上限2,500,000,001，沿原版工具相同時鐘，沒有調IRQ0參數；這是觀察界線，並非硬體wall-clock驗收。播放後控制流仍DRAFT，先取得自然返回或明確終態再判斷下一步。

[獨立正常返回checker](../tools/verify_dosgolem_recruitment_return.py)固定r2 source／binary／producer雜湊，完整重驗父來源、全部199包DQ3事件與398份PNG／bin保持，並要求完整自然返回、後續正常按鍵、No與告別、202包／480IRQ1及404份產物。資料尚未到齊或控制流不符時失敗，不把前綴收據當作播放完成。CLI依序為原版目錄、producer、尚不存在的輸出收據，另明示`--assets`唯讀原版目錄；仍只能在相同有界Docker執行。此checker於DRAFT備妥，尚未通過實際來源驗收。

### 2026-10-05 有界入隊返回觀察與停止線

r2在2,500,000,001指令上限停止，原生程式仍在等待，未觀察到10459之後的caller返回。199包入口計時器14472，終態17475，前進3003；IRQ1仍474、record仍538，沒有送出後續Right／Enter。這只證明本次有界執行未自然返回，不證明產品或dosgolem缺陷，也不宣稱永遠不返回。終態IF=false不能單獨證明計時中斷遺失。

兩次冷啟動的199包全部DQ3事件與各398份完整PNG／bin均等於父來源d0f6428d。r2 producer SHA-256為`a7eb2be1a4987af091fbb3ec52ef281cf385f8507df4d97181d6f753f1e3d092`，Go來源`aa3c7b92741325ccc69d7dd5c90f5a3ae2a36b5a4a557c9a56906a5a2a52b22f`，執行檔`0ec5657bb4473dc5d7d70dc7d57157822b89f038a7a1c292979cb50d038ab0ba`。目前公開producer輸出前綴為r2，上節r1是保留的歷史版本。兩側正常seed1357只固定一次；不宣稱跨流程骰序或角色隨機能力已等價。

本輪只重驗既有IDA Pro9.4 sidecar的18筆原始bytes，輸入DQ3.EXE大小115282，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。範圍為IDA linear208E2至20911，file=linear−EC90；DGROUP基底linear24DD0。可逐byte回查`test word ptr ds:286Dh,3`、`mov ax,ds:2898h`、零判定與far call `sub_22E10`。原始欄位及callee語意保持unknown，20912分支不在本次核對範圍；不把2898改名為播放中，也不推論callee沒有讀鍵盤。未重新分析driver／ISR或改CPU／時鐘。

公開checker對實際r2來源拒絕`normal return incomplete`，未建立`issue4-recruit-return-r2-source-r1-receipt.json`。目前只能驗證拒絕不完整來源，尚無自然完成來源的正對照，整段返回仍DRAFT。`TestRecruitmentJoinDosgolemNormalInputComparison`及`TestRecruitmentJoinLoadClearsAudioOnlyAfterValidation`通過，共2頂層／1子PASS、零SKIP／OOM，50張runtime PNG等於前輪。199之後的remake返回及存讀檔只屬內部回歸，不升格原版parity。

| 私人收據 | SHA-256 |
|---|---|
| `work/issue4-recruit-return-review-r1.json` | `7d1f80cf0152702a219f86b4e9f580b69ff338e5b465460ca3ca13f484cf731b` |
| `work/issue4-recruit-return-wait-review-r1.json` | `d016bd6522d9867992cda59ff3bfedf113190a60689ec53474899bd48b6517f5` |

正式Go／pack保持e939db2，不依逾時猜改播放等待。停止硬體driver、ISR、DAC、PIT細節追查，不重跑相同有界樣本或調IRQ參數。下一正常垂直切片為已接受137411c7的合法193checkpoint：F6選槽按Esc取消，確認返回及後續正常行走。入隊完成oracle、可比動畫時鐘、其他F6場景、複數隊伍、非空分離與完整原版campaign保留未知；沒有新發行包。

### 2026-10-05 正常 F6 選槽取消與行走 DRAFT

[正常取消探針](../tools/dosgolem_load_cancel_probe.py)從137411c7相同冷啟動、seed1357及193包正常路線續行。排定F6、十槽Esc、左移及右移，場景／選單條件不符便不送下一鍵；原生Scratch使用既有DOS.Scratch可寫overlay，原版素材唯讀。不讀回或注入checkpoint、不調CPU／時鐘；附加觀察只讀DGROUP251D、526C、0726與4F29..57A5持久區。

CLI為`python3 /repo/tools/dosgolem_load_cancel_probe.py --prefix issue4-load-cancel-r1`；沿固定dosgolem2f44a68、dq3-ebiten-test:20260822-r1，明示UID1000、有限資源與外層逾時。prefix既有輸出不得覆寫。先重驗193包全部事件與386份PNG／bin、再判定Esc返回、FileOps無存檔交易、持久區與世界時鐘；原版未知時不改正式產品。完整畫面與正常InputState／存讀檔驗收待來源接受後進行。

[獨立F6取消checker](../tools/verify_dosgolem_load_cancel.py)固定producer、生成Go及binary雜湊，沿父來源完整重驗，核對實際IRQ make／break、197包、394份完整PNG／bin及5份持久區。CLI依序為原版輸出目錄、producer、尚不存在的收據路徑；仍DRAFT，預期取消與行走條件需由本次原版實際來源審查，不以checker假設代替原版結果。

### 2026-10-05 F6 十槽 Esc 取消有限 READY

實際來源`work/dosgolem-opening/issue4-load-cancel-r1-source-r1-receipt.json`已接受，SHA-256 `e2c2e300b9c322389fa33992f1d61977ef9c16096ccfd3bea97e67bcabcee206`。正常197包／470IRQ1，38個冷啟動輸入加197包共235次正常輸入，seed1357執行前固定一次。前193包全部狀態、輸入／IRQ與386份完整PNG／bin保持，174份母親父產物亦保持；新來源全部394份packet PNG／bin及5份2172bytes持久區逐項核對。

本次confirmed範圍：原版已存在十槽的上層城鎮場景2,18，F6進十槽選單、初始游標1；Esc直接返回正常field，未新增文字或告別等待。新的左鍵到1,18，再右鍵到2,18；全部正常資料節點均為clock30、raw526C=1。Esc的raw0726從0到1保留原始欄位，不推論其他用途。2172bytes取消前後保持；左移只改座標低byte，右移完全回復。Scratch仍空，FileOps沒有後續dragon存檔讀回或建立／刪除，不能把未記錄AH40的FileOps單獨当作無寫入證據。

此有限READY允許驗證既有正式F6 Cancel是否立即關閉、不改存檔／snapshot／RNG／clock，以及新按鍵正常行走與存讀檔。輸入、位置和UI均不得注入；可比可見metadata沿上一輪合法外部JSON測試前置資料，不表示其他九槽世界狀態與原生DAT互通。下層場景、空槽、其他游標與動畫時鐘不在本次READY。[正常InputState驗收](../dq3_remake_ebitan/game/field_load_cancel_test.go)固定接受來源，完整畫面待本輪實測；不預設全RGB零差異或完整V3。

收據契約勘誤：r1把原版身份只放在meta內，既有`sourceCanvasDifference`要求外層`original_sha256`，首輪正常測試因此停在畫面稽核的身份檢查。原始EXE、cold run、PNG、持久區與遊戲程式均未改。checker補齊外層身份後對同一產物完整重驗，另存`issue4-load-cancel-r1-source-r2-receipt.json`，SHA-256 `426c7623d239af6f83fa9715de861ca932a9fa6bbc2a223bb7eac2f49aa484a7`，由正式測試固定引用；r1及首輪log保留。這是收據欄位問題，未作產品缺陷修正。

測試交易勘誤：r2正常F6取消與四張畫面均已產出，存讀檔斷言卻將Save之前的舊Respawn與Load後比較。單獨診斷證實只有Respawn從21,17變2,18，RNG保持；現行`saveTo`先寫當前復活checkpoint，再更新`g.respawn`。斷言移到Save成功後取得比較基準，驗證Load恢復真正已存狀態，不改存檔交易或忽略其他欄位。保留r2與單獨診斷log；這次修正驗證腳本的交易界線，不把預期Save副作用當產品缺陷。

[完整取消畫布診斷](../tools/verify_dq3_load_cancel_raster.py)固定本節來源與runtime收據，重驗399份來源產物及log，逐點檢查194..197四張完整640×350 PNG。右移只核對既有右向英雄6/7的原始BLS圖塊、透明像素及CTY／BLK底圖，沒有讀取或設定動畫counter；圖塊match不證明可比時鐘。CLI明示`--assets`、`--original`、`--source-receipt`、`--runtime`與不存在的`--output`，支援範圍限本節固定來源與四張完整畫面。

### 2026-10-05 F6 取消有限 CONFORMED

現行正式Cancel行為符合有限READY，不需要產品修改。正常InputState重播F6、Esc、左移與右移；取消期間完整snapshot／RNG及十份外部JSON存檔逐byte保持，clock30不重設，左右移動只改正常位置。Save成功後Load恢復完整已存snapshot與RNG，新的左鍵仍可續行。schema0.18.0／content0.1.90／canonical9d6325ad與正式Go／pack保持e939db2。

最終四命令全通過：`TestFieldLoadCancelDosgolemNormalInputComparison`、`TestFieldSaveLoadComparableSlotsDosgolemNormalInputComparison`、`TestFieldSaveLoadNoAndSlotCancelNeverWrite`、`TestFieldSaveLoadIndependentSlotsAndRejectedLoad`。4不同頂層／1子、5筆PASS、零SKIP／OOM。舊可比路線53張PNG逐byte保持，新取消路線48張包含44既有前綴保持。完整game／internal／THE END／desktop最近仍為e939db2，未冒稱本輪全套重跑。

正常194選槽、195取消返回及196左移的完整224000像素RGB均零差異，目視文字、游標、底圖與場景一致；這三張限定V3。197右移完整RGB差122，全部落於英雄原始6／7圖塊，完整人物含透明像素與底圖核對，沒有人物外差異，保持V2。原版consumer phase未在這份來源觀察，可比動畫時鐘仍unknown，不因raw圖塊吻合推論控制流或全流程V3。兩側世界clock30已核對，與未確認的動畫時鐘分開記錄。

獨立來源checker拒絕遺失最後IRQ1、取消後持久byte被改及最後PNG損壞三種來源，正對照前後一致。原版產物始終唯讀；沒有冷啟動重擲、模擬器restore、狀態注入、裁切、遮罩或替代PNG。r1來源、首輪log及交易診斷保留，不覆寫歷史收據。

| 本機收據 | SHA-256 |
|---|---|
| `work/dosgolem-opening/issue4-load-cancel-r1-source-r2-receipt.json` | `426c7623d239af6f83fa9715de861ca932a9fa6bbc2a223bb7eac2f49aa484a7` |
| `work/issue4-load-cancel-validation-r3/load-cancel/receipt.json` | `4ce86b0b73fae6a1439b315b98a4bda4f06ec71e5b3d3d79e00a12dd5b7c9bd9` |
| `work/issue4-load-cancel-negative-r1/receipt.json` | `30c4718f2b4f2921ec4be3213cd6e85fbbb3df6268609ebd2b34046864141776` |

重生正常測試使用`DQ3_LOAD_CANCEL_ORACLE_DIR`指定本節接受來源、`DQ3_LOAD_CANCEL_RECEIPT_DIR`指定空輸出，在相同有界Docker／Xvfb執行`TestFieldLoadCancelDosgolemNormalInputComparison`。素材或hash缺失失敗，不以SKIP驗收。完整畫布命令為`python3 /repo/tools/verify_dq3_load_cancel_raster.py --assets /repo/assets_raw --original /work/dosgolem-opening --source-receipt /work/dosgolem-opening/issue4-load-cancel-r1-source-r2-receipt.json --runtime /work/issue4-load-cancel-validation-r3/load-cancel --output /work/issue4-load-cancel-raster-r1.json`；既有收據不覆寫，重驗另用新輸出。

此有限CONFORMED只涵蓋十槽初始游標的上層F6取消與正常後續。下一切片是正常F5／F6第二槽存讀檔，從冷啟動創造可比的新保存狀態，避免拿既有其他槽的未知世界初值作對拍。入隊音樂返回、其他游標／場景、非空分離、多角色與完整campaign仍未知。停止已解釋影格樣本的時鐘重試，沒有新發行包。

### 2026-10-05 正常 F5/F6 第二槽 DRAFT

依Issue #4從75b296d續行。[正常第二槽探針](../tools/dosgolem_second_slot_probe.py)沿既有F5第一槽與後續行走builder，只改正常游標輸入：F5進十槽後按Down再Enter選第二槽，確認告別後左移，再F6按Down／Enter讀回第二槽，最後左右行走。選單／游標或場景條件不符就不排後續鍵。父194包、38個cold inputs與seed1357保留；clock／持久區／native caller只讀，不restore或注入狀態。

CLI為`python3 /repo/tools/dosgolem_second_slot_probe.py --prefix issue4-second-slot-r1`。沿固定dosgolem2f44a68與dq3-ebiten-test:20260822-r1，以UID1000、有限資源與外層逾時執行；原版與repo唯讀，只有既有work可寫，Scratch使用原先已審DOS.Scratch overlay。先查真正PLAYER／dragon1.dat交易與完整2172bytes、其他槽保持、Load世界時鐘及完整畫面，再審有限READY。第二槽不預設等於第一槽oracle；正式Go／pack暫不修改。既有輸出不覆寫，所有原版素材與收據留本機。

[第二槽獨立checker](../tools/verify_dosgolem_second_slot.py)固定本輪producer／Go／binary，完整重驗194包父來源與174母親產物，核對204包／484IRQ1、第二槽游標、真正dragon1.dat／PLAYER交易及完整2172bytes讀回。CLI依序為原版輸出目錄、producer、未存在的輸出收據，另以`--assets`明示唯讀原版目錄。checker預期條件仍DRAFT，需接受實際原版來源後審查READY；不把範本條件當已證實結果。

### 2026-10-05 第二槽正常存讀檔有限 READY

實際來源`work/dosgolem-opening/issue4-second-slot-r1-source-r1-receipt.json`已接受，SHA-256 `1a5a7c22c4599ce1bf4f6070d178a6d24ee49a8bbc4f374e84e39459a7a3f9d2`。正常204包／484IRQ1，38 cold inputs加204包共242次正常輸入，seed1357一次。前194包全部事件／388份完整PNG與bin及174母親產物保持；本次422份產物含408份packet PNG／bin、12份2172bytes持久區與2份原生存檔。

confirmed限本次上層單人場景：195選槽游標1，正常Down令196游標2，Enter於197保存，完成告別後198新Enter返回field。199左移到1,18，200 F6進選槽游標1，201 Down令游標2，202 Enter真正讀回第二槽及2,18；203左移、204右移可續行。世界clock193..201保持30，202..204為0，raw526C均1。

Scratch只含`player.dat`與`dragon1.dat`。PLAYER第二筆file20..39由本次主角metadata更新，file0..19及40..199逐byte等於原版，其他槽未被保存。dragon1.dat完整2172bytes的SHA-256為`9ef549b0ac969e8bb9a0138d4e339905c21f5abf8653ca718c09d4b4c5db5910`，與前一第一槽同狀態保存內容相同。原生202持久區完整等於保存檔；203只改座標低byte，204回復。FileOps真正create／read dragon1，沒有本輪dragon0交易；AH40另以原生caller及落地bytes核對，未單獨依FileOps推定寫入。

原始EXE115282bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；dosgolem2f44a68固定，observer位址為IDA linear。既有玩家層Save／Load caller11484、114C8、114D3、114D9、1157D、1158B、11591、1165F本次全部自然到達，raw0722均2、raw0726均0；197寫入CX087C／DX4F29，202讀回相同長度與DGROUP持久區。沿原先已審契約，不重新分析硬體driver或解讀未知欄位。

本次READY允許驗證第二槽正式InputState保存／讀回、其他JSON槽保持、snapshot／RNG及世界時鐘；可見初始metadata沿合法193checkpoint前外部資料條件，不注入執行中狀態。新保存後才比較第二槽恢復；其他既有槽完整世界不等價。[正式正常第二槽驗收](../dq3_remake_ebitan/game/field_save_second_slot_test.go)固定本次來源，完整畫面待實測，不以第一槽綠測試替代第二槽。音效duration保持既有hardware-spec approximation，沒有聲波或實機wall-clock聲明。

[第二槽完整畫布診斷](../tools/verify_dq3_second_slot_raster.py)固定本節原版與remake收據，重驗422份原版產物及log，逐點核對194..204十一張完整PNG。差異只以原始BLS／CTY／BLK與實際runtime人物資料解釋；選槽窗遮住的人物不作圖塊判斷，但所有224000像素仍比較，沒有人物外差異才接受。原版動畫counter未觀察，不以固定raw圖塊match宣稱動畫時序。CLI明示`--assets`、`--original`、`--source-receipt`、`--runtime`與不存在的`--output`，只支援本節固定來源，不替換／裁切／遮罩PNG。

### 2026-10-05 第二槽有限 CONFORMED

現行產品第二槽行為符合本次READY，沒有需要修正的正式Go／pack差異。正常F5選第二槽、保存完成與告別、左移、F6第二槽讀回及左右行走全部通過。十份JSON只有第二槽在197被替換，其他九槽逐byte保持；202讀回已保存的完整snapshot，依pack重設世界clock0，RNG保持。路線後Save成功／Load及新左鍵可續行，只作remake回歸，不外推原版額外存檔流程。

五項測試通過：`TestFieldSecondSlotDosgolemNormalInputComparison`、`TestFieldSaveLoadComparableSlotsDosgolemNormalInputComparison`、`TestFieldLoadCancelDosgolemNormalInputComparison`、`TestFieldSaveLoadIndependentSlotsAndRejectedLoad`、`TestFieldSaveLoadFailedWritePreservesModal`。補六張完整RGB0斷言後只重跑第一項，合計6命令／7筆PASS，涵蓋5不同頂層／1子，零SKIP／OOM。兩條舊路線48＋53張PNG逐byte保持；新第二槽55張含44既有前綴保持，補斷言前後全部55張相同。正式Go、九份pack及schema0.18.0／content0.1.90／canonical9d6325ad保持e939db2，沒有新包。

| 第二槽正常包 | 完整RGB差異 | 本次像素核對 |
|---|---:|---|
| 194..197、202、203 | 0 | 六張完整640×350零差異，正式測試鎖定；有限V3 |
| 198 | 411 | 英雄182、NPC14為106、NPC15為123 |
| 199 | 356 | 英雄127、NPC14為106、NPC15為123 |
| 200、201 | 各123 | 窗外NPC15完整圖塊；選槽窗與其他像素保持 |
| 204 | 229 | NPC14為106、NPC15為123；英雄及其他像素保持 |

完整BLS圖塊含透明像素、CTY／BLK底圖與原版色盤全部核對，差異外其餘224000像素一致；目視第二槽游標、保存後姓名「0」及讀回內容一致。這是各指定capture的confirmed raster explanation，不是原版動畫counter或週期的證明。世界clock30→0已證實，動畫時鐘未知；五張有差異畫面保持V2，不能宣稱全流程V3。

| 本機收據 | SHA-256 |
|---|---|
| `work/dosgolem-opening/issue4-second-slot-r1-source-r1-receipt.json` | `1a5a7c22c4599ce1bf4f6070d178a6d24ee49a8bbc4f374e84e39459a7a3f9d2` |
| `work/issue4-second-slot-validation-r2/second-slot/receipt.json` | `4578ab4aa889af13446c95e2a980fb9bdf28cceb33415e5bd12c54b71296dd1d` |
| `work/issue4-second-slot-raster-r1.json` | `0a3a6a68601791d8bc81ff7cc177d4dd97fd97020b778753e5e3fb92fbcbd737` |

重生正常測試以`DQ3_SECOND_SLOT_ORACLE_DIR`指定接受來源、`DQ3_SECOND_SLOT_RECEIPT_DIR`指定空輸出，在相同有界Docker／Xvfb執行`TestFieldSecondSlotDosgolemNormalInputComparison`。素材與來源hash缺失失敗，不以SKIP驗收。完整畫布CLI為`python3 /repo/tools/verify_dq3_second_slot_raster.py --assets /repo/assets_raw --original /work/dosgolem-opening --source-receipt /work/dosgolem-opening/issue4-second-slot-r1-source-r1-receipt.json --runtime /work/issue4-second-slot-validation-r2/second-slot --output /work/issue4-second-slot-raster-r2.json`；r1與r2 runtime收據逐byte相同，既有輸出不覆寫。

本次限定E2／流程E3及六張完整V3，只涵蓋本次新保存的上層單人第二槽；其他槽完整世界、空槽、下層／多角色及動畫時鐘未知。下一正常切片是合法checkpoint的Space命令窗開關、游標導覽及道具入口。現行CmdMenu註解引用歷史C，不能代替原版oracle；先從原版正常操作取得證據，再判斷正式產品缺口。入隊返回保留DRAFT及硬體停止線，完整原版campaign尚未完成。

三種損壞來源全部拒絕：其他槽metadata被改、讀回持久byte被改、最後一次IRQ1缺漏。每批完成後重新完整驗證未改動的正對照，接受來源仍為1a5a7c22。私人負例紀錄為`work/issue4-second-slot-negative-r1/receipt.json`及`work/issue4-second-slot-negative-r2/receipt.json`，逐byte相同，SHA-256為`d00995aa2e0d50e1f3f028edbf95752a4fe13be99e0884ddb54c51dbba980cf0`。

r1成功收據寫完後程序結束碼137，停止原因未確認，不能宣稱OOM、逾時或產品缺陷。r2維持相同image、UID1000、768MiB／2CPU／64pids及network none，只將外層上限由300秒改600秒並使用新輸出目錄，乾淨重跑結束碼0。兩輪歷史保留，沒有快取代替完整正對照，也沒有重跑原版或調整遊戲時鐘。

### 2026-10-05 正常 Space 命令窗 DRAFT

依Issue #4接續6595dc5。[正常命令窗探針](../tools/dosgolem_command_menu_probe.py)沿已接受193包冷啟動路線，將下一個F5改為正常Space，停在194包實際返回的可觀測輸入節點。原始EXE115282bytes及5178fdc8完整雜湊、dosgolem2f44a68、seed1357一次、原生DOS.Scratch overlay均保持；只讀世界clock及2172bytes持久區，不restore、注入狀態或調整時鐘。

CLI為`python3 /repo/tools/dosgolem_command_menu_probe.py --prefix issue4-command-menu-r1`，以既有dq3-ebiten-test:20260822-r1、UID1000、有限資源及外層逾時執行，原版與repo唯讀、既有work可寫。首先核對正常Space是否進命令窗、父193包事件／PNG／bin及持久區保持，獨立來源驗證後才進正常InputState比較。導覽、取消及道具入口尚待原版觀察，現行C移植註解、Go硬寫版面與元件測試不可當READY證據；有差異先補原版視窗consumer與資料契約，不猜改正式產品。

[獨立命令窗來源驗證](../tools/verify_dosgolem_command_menu.py)完整重驗父193包及174母親產物，核對194包／464IRQ1、38個cold inputs加194次正常輸入、Space與六選項初始游標，以及全部388份packet PNG／bin與兩份持久區。CLI依序指定原版輸出目錄、producer、尚不存在的接受收據；來源接受只證明原版本次正常開窗，未證明remake畫面或完整版面READY。候選IDA9.4非破壞sidecar在`work/issue4-command-menu-r1-ida.json`，261條目，SHA-256 `d78d5ff616212a0a5c2391ac68c5e6d17bcd8b8dcccbc3836989f86ba95c1d22`；保留原始bytes、MZ relocation、函式名及xref type，所有新語意仍unknown。未定DS的直接資料xref缺項，不當作沒有reader／writer。

原版開窗來源`work/dosgolem-opening/issue4-command-menu-r1-source-r1-receipt.json`已完整接受，SHA-256 `cdb0a4b69c83613681f6f3572665b1b1a8bc745fab5a8c959b23854c081af6fa`。193前綴全部事件／386份PNG／bin與174母親產物保持；194為正常Space、六選項、初始游標1，選單輪詢IDA linear1F7B7。2172bytes持久區與clock30保持，Scratch空。本次只允許用正式InputState.Confirm對照正常Space開窗，DRAFT可丟棄測試在`work/issue4-command-menu-test-draft.go`；完整Go版面、導覽與道具入口仍未READY。

[正常導覽探針](../tools/dosgolem_command_menu_navigation_probe.py)沿同一冷啟動重生194包，再送Down三次、Up、Left、Right兩次、Esc；回field後重新Space、Down、Right，只在原生游標5時Space進道具。停在206實際可觀測節點，未完整抵達不接受。CLI為`python3 /repo/tools/dosgolem_command_menu_navigation_probe.py --prefix issue4-command-navigation-r1`，相同有界Docker契約；附加觀察只讀DGROUP3D6C的64bytes、3E9C的24bytes與原始色號／選單欄位，不改時鐘或狀態。游標跨列／跨欄規則尚待此正常來源與IDA consumer審查，不能沿用C移植的繞回假設。
### 2026-10-05 正常指令窗有限 READY

路由重查命中復古 remake 的 spec-gated workflow，依 RE → DRAFT → 審查 → READY 進行。來源由[獨立正常導覽驗證](../tools/verify_dosgolem_command_navigation.py)接受；CLI依序指定原版輸出目錄、producer與新收據路徑。接受收據 SHA-256 `5b2a78c152e77f1ddbfd850af0cc129837bf206cd9a15731529927d21b12978a`，原版206包、488 IRQ1、244次正常輸入，前194包事件與完整PNG／bin保持。193..206的2172bytes持久區完全相同，clock30與seed1357一次保持，Scratch空。此來源只接受原版，不宣稱remake parity。

原始DQ3.EXE為115282bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。工具為IDA Pro9.4及固定dosgolem2f44a68。IDA linear轉file減EC90，DGROUP base為IDA linear24DD0／file16140。原始函式名、bytes、relocation與xref type留在私人sidecar；未知DS的空xref不表示沒有writer。新語意以下表為審查索引，不覆寫原始名稱。

| 原始定位 | 附加語意與證據 | 等級 |
| --- | --- | --- |
| IDA linear17C83、17CCB..17D18 | 5077隊伍人數寫3E9C的count；寬度為4+10×人數EGA bytes，1F590畫HUD、18222寫內容；3D6C交給1F4E3 | confirmed，正常194 raw及完整畫面閉合 |
| DGROUP3D6C／file19EAC／IDA linear28B3C | flags7、x19 bytes、y30、width24 bytes、height80、record400、count6、kind2；六筆x/y/callback為21/46/4E0E、21/62/8301、21/78/7E12、31/46/C9C1、31/62/372F、31/78/8966 | confirmed，原始64bytes與正常194..206 raw一致 |
| IDA linear1F779..1F8A4 | 游標1起始；Up/Down線性減加並在六項邊界繞回；Left/Right以count/2切換兩欄；Enter1C與Space39確認、Esc01取消 | 導覽與Esc confirmed，正常195..205閉合；Enter等價strong，仍需正常remake輸入測試，未有新增原版Enter動態收據 |
| IDA linear1F908、1F590、1F4E3 | kind2以(cursor−1)×6取游標座標；raw2784=FF時用glyph11；文字、shadow與flags2的checker XOR frame沿既有primitive | confirmed，正常raw258F=8／0727=5／2784=255及完整194畫面 |
| DGROUP3E9C／file19FDC／IDA linear28C6C；IDA linear18222、219D2、219AA | record401/402/403構成HUD；姓名起點x+4 bytes；HP/MP/level三位數右對齊，前置零寫glyph12；共用idle HUD契約已有相同consumer | confirmed，正常單人width14 bytes與數值畫面；健康／多角色延續原證據範圍，不擴大V3 |

私人IDA sidecar r1..r5條目數為261、397、102、63、43；SHA-256依序為`d78d5ff616212a0a5c2391ac68c5e6d17bcd8b8dcccbc3836989f86ba95c1d22`、`ac89ab26c36a09c97231addf2a8e67be62d9c3d1ecb7c2d020b64fb16dc378c3`、`3c6ec4f0d4219605ab21e4a78ae856e614569842334669ed5e713a950eacd3ea`、`6152b371ca8338b23cc234c339f8223c5537dd783073650769f7e1b6611087b5`、`30b15b79e12e3d358a13fb8eb8e10b8626aaaca952bbef1a4c33a6b654b868fa`。1C03F不再當作標籤writer；19842是另一個五選項handler，不混入本次六選項契約。

可丟棄試作只在正常194後暫時撤除Go UI圖層，使用原始record與上述raw contract重畫，未改輸入／狀態／動畫時鐘。r4完整640×350 RGB差異0，私人收據`work/issue4-command-menu-prototype-r4/command-open/receipt.json` SHA-256 `f0c37b975d30901cac781fe68a393700465f77c4f4745f3fade9122985cb9da6`。r1漏讀全域文字、r2編譯失敗及r3的16點前置空格差異均保留；不當作正式產品驗收。

READY範圍：正常Space開窗、按欄排列六指令、導航、Esc關窗／重開、同一HUD consumer與Enter/Space確認的引擎等價。`field_command_menu`保存raw window、文字引用、依原版順序的語意指令／游標／callback與色號，程式不新增DQ3座標、record、glyph fallback。引擎保留語意cursor供既有handler使用，原生序號由pack映射。觸控命中區依文字格與已證實錨點延伸，不聲稱原版滑鼠parity。HUD重用`field_idle_status`與`party_hud`已有的文字、動態寬度、健康色與狀態欄契約。

驗收需正常InputState從193合法checkpoint重播194..205，逐張完整RGB、游標、snapshot／PRNG／clock保持，Esc後F5/F6正式存讀檔及下一步可玩；原始EXE／D3TXT00 parity與缺值／未知引用拒絕測試。Enter分支只作remake等價測試，來源未動態觀測，不稱新增原版V3。新的型別入口為`dq3_remake_ebitan/internal/gamepack/field_command_menu.go`及其原始資料測試，欄位職責見docs/84；正式正常trace入口為`dq3_remake_ebitan/game/field_command_menu_test.go`。

道具206原版顯示穿戴衣服加六件背包，共七項，並保留inactive父指令窗；現行remake六項且父窗消失。選取與action writer尚未閉合，維持DRAFT與下一blocker，不把穿戴物直接塞入背包索引。新contract只承接已READY指令窗，不更改道具交易。音訊硬體時序、其他隊伍同狀態及完整campaign仍未驗收，原版素材與private sidecar不入Git。

[完整畫布核對器](../tools/verify_dq3_command_menu_raster.py)唯讀核對接受來源、426份原版產物與正式runtime收據，對194..205各比較完整224000點RGB，206保留已觀察的44816差異。CLI為`python3 /repo/tools/verify_dq3_command_menu_raster.py --original /work/dosgolem-opening --source-receipt /work/dosgolem-opening/issue4-command-navigation-r1-source-r1-receipt.json --runtime <正式command-menu輸出> --output <新收據>`。來源與正式輸出不裁切、遮罩、補畫或調整人物影格；道具RED不升格為完成。資料重建入口為[正常指令窗遷移器](../tools/migrate_field_command_menu_pack.py)，參數及九份JSON職責見docs/84。

### 2026-10-05 正常指令窗有限 CONFORMED

正式產品已依上節READY完成本次有限切片。原版record400完整框架與標籤、游標字模／色號及六筆順序由typed pack提供；共用狀態機上下線性繞回、左右切半欄，原生序號映射回既有語意指令。HUD重用已閉合idle狀態欄，移除舊的重複HUD繪圖。版本專屬raw位址及callback只在JSON與oracle中，不在共用renderer新增fallback。

從正常新遊戲路線抵達193後，以正式InputState重播Space、Down三次、Up、Left、Right兩次、Esc、Space、Down、Right、Space。正常194..205游標／模態與原版一致，十二張完整640×350 RGB差異全部0；開窗194由22519降0。兩側2172bytes／snapshot、PRNG與世界clock30保持，沒有restore、狀態／座標注入、時鐘調參、人物遮罩或替換PNG。206正常抵達道具入口後取消，再由正式F5保存、F6讀回及下一步行走通過。存讀檔比較以真正寫入snapshot為基準，包含Save更新Respawn。

schema0.19.0／content0.1.91，canonical `sha256:66224bc04ff5f7d412640c986c35e0aa5f4eb7f49d4b1344b2df2a47d778a773`。遷移器從乾淨6595dc5的0.18.0 pack生成兩份獨立副本，九份JSON逐byte相同並等於正式檔。原始EXE／TXT parity、十二種缺值／未知引用／越界壞契約及正式引擎測試通過。舊schema／hash存檔沿既有策略拒絕，不自動遷移。

完整R1保留三項舊策略失敗：Opening、Phoenix、Zoma測試用Left→Up尋找Examine，在原版順序下選到Equip。只訂正正常玩家測試為Up繞回第六項，再以同一image／命令／資源乾淨跑完整R2，產品程式不再改動。R2共484不同頂層覆蓋、488次呼叫，433不同頂層／119不同子PASS、51選用SKIP；internal171頂層／375子、11套件PASS、4選用SKIP。正常新遊戲至THE END190.25秒及Linux x86_64 desktop建置通過。cgroup memory max374、oom0／oom_kill0；必驗指令窗沒有SKIP。選用原版來源未啟用的SKIP不可當parity通過。

其後沿相同R2 binary另啟用四條最近正常來源：F6取消、讀回後行走、可比十槽資料及第二槽保存／讀回，共四不同頂層與兩子PASS、零SKIP／OOM。私人收據`work/issue4-command-recent-r1/receipt.json`的SHA-256為`7fd7a750ceb701ba66e0d48690c3869825679d7810d0bd44feb5321fb2c86953`。不覆寫完整R2的51項SKIP歷史，只記錄這四項已補驗。

前一完整存讀檔回歸目錄的1025張PNG在本次完整R2逐byte保持；本次57張含44既有前綴在R1／R2逐byte一致，正常trace收據也相同。完整畫布checker兩次報告逐byte一致。來源負例拒絕錯誤游標邊界、持久byte損壞、缺最後IRQ1；畫布負例拒絕runtime hash錯誤、缺195 PNG、損壞PNG。每組負例後完整正對照重驗成功，接受來源不變；負例的父來源快取只縮短破壞測試，未用於正對照驗收。

| 本機收據或產物 | SHA-256 |
| --- | --- |
| `work/dosgolem-opening/issue4-command-navigation-r1-source-r1-receipt.json` | `5b2a78c152e77f1ddbfd850af0cc129837bf206cd9a15731529927d21b12978a` |
| `work/issue4-command-full-r2/command-menu/command-menu/receipt.json` | `4293924f2b76a228dc80d91680ebcfcc989eefd765c589669e9fd07831276740` |
| `work/issue4-command-raster-full-r2.json` | `c02f15724d544bcc38749e8fe90865fc43c7217542a99ffa309595e5d2096375` |
| `work/issue4-command-full-r2/game-receipt.json` | `6dd3afdb43b61cf9134926dedbe27a9f9dc098e136f116439ca3a4a6961ae376` |
| `work/issue4-command-full-r2/game.test` | `06b4de048bbc086aceba2fe0e8fc3e0e4e8343aa39c1a6a6162429354db0729a` |
| `work/issue4-command-full-r2/internal.log` | `ec7a776a94079c50df22decf00749fa160a7f6b255a5dcca1943d5e5ae874899` |
| `work/issue4-command-full-r2/dq3-linux-x86_64` | `0dd87ef2f7cea49d9a3a7a6ce7ce3da8920b02ae36c068deb8d3cbcac9f2d1e6` |

重生正式正常trace時，以`DQ3_COMMAND_ORACLE_DIR`指定已接受的原版目錄、`DQ3_COMMAND_RECEIPT_DIR`指定空輸出，於本節Docker／Xvfb工具鏈執行`TestFieldCommandDosgolemNormalInputComparison`，素材缺失失敗，不能以SKIP驗收。完整畫布使用上節公開CLI，只支援明示版本與來源。收據、原版PNG／bin、IDA sidecar及原版素材留本機，不加入Git。

本次狀態E2／正常流程E3與指定十二張V3只適用單人上層指令窗；Enter原版動態未觀測，保留strong與remake等價測試。道具206仍全畫布RGB44816，原版七列含穿戴衣服且父指令窗保留，remake六列且父窗消失。穿戴／背包索引、action writer及交易副作用未閉合，下一切片仍DRAFT。觸控、其他隊伍、音畫時序、入隊返回及完整原版campaign沒有新增完成聲明。沒有新發行包。

### 2026-10-05 正常道具七列與取消 DRAFT

依Issue #4接續e94a4c7，首次阻塞為正常206道具清單，完整RGB差44816。[正常道具導覽探針](../tools/dosgolem_field_item_navigation_probe.py)從原版冷啟動重生已接受206包，seed1357一次；正常Space選第一件穿戴衣服，只開action後Esc，不執行使用／給予／丟棄。其後Down到背包、Up回穿戴列、上下繞回、Esc取消清單／父窗及正常右移，預計215包；實際節點或count不符即停止送鍵，不猜過渡。

CLI為`python3 /repo/tools/dosgolem_field_item_navigation_probe.py --prefix issue4-field-item-navigation-r1`。沿dosgolem2f44a68與dq3-ebiten-test:20260822-r1、UID1000、有限資源及外層逾時，repo／原版唯讀，既有work可寫。新增只讀觀察DGROUP062D／062F／2591、3FD8／4050視窗、原生游標與完整持久區，保留原始欄位與bytes；不restore、狀態注入或調時鐘。缺完成收據不能稱接受來源。

IDA9.4的r1因1372F沒有自動函式邊界而拒絕匯出，符合docs/158已有raw-window限制；不建立猜測function。r2改用原始1372F..13B18與1F779..1F8A5有界區間，456條目，SHA-256 `c52dda943d6935ed9127fa1a5e0131071a2cdb82d9604812dcaf160b7f93aef1`，私人`work/issue4-item-r2-ida.json`保留bytes／relocation／xref及unknown。現階段只可確認本EXE掃角色+3A的八個非空word；穿戴旗標、raw列與remake分離storage、完整窗口／action返回仍待正常來源和consumer審查。正式Go／pack不先改動。

首輪正常207開「如何」三項選單，208 Esc自然回field。與最初DRAFT預期回七列清單不符，沒有送209鍵，停止該已知無後續路由的研究容器；208前全部產物留存，未產生完成／接受收據。首輪producer保存在`work/issue4-item-navigation-producer-r1.py`，SHA-256 `53ae76a6b10fb0d03f6e53effb0653cb2df3b9021fe412551321158e9a6418ea`，不冒稱原版runner缺陷。

第二輪改成實際正常路線：208後Space重開父窗、Down／Right／Space進七列，再Down／Up／Up繞回第七列／Down繞回第一列，Esc取消清單後正常右移，預計218包。CLI使用新prefix `issue4-field-item-navigation-r2`，不覆寫r1。完整正常來源接受後，須訂正docs/158過度概括的取消階層；目前多人取消仍未知，單人動態結果不能外推。

[獨立正常道具來源驗證器](../tools/verify_dosgolem_field_item_navigation.py)固定本輪producer、Go source與binary，先完整重驗206包父來源，再核對218包／512 IRQ1、256次正常輸入、原始八格七件、游標繞回、兩次Esc、下一步、2172bytes及實際Scratch。CLI依序指定原版目錄、producer及新接受收據；預期契約仍DRAFT，以實際來源審查，不放寬未知條件或略過父來源。正式Go／pack與e94a4c7保持。

私人可丟棄render試作`work/issue4-item-render-draft-r2/`已由正常193路線重播至206，只重畫UI圖層，以該次原版固定七列與raw視窗核對完整640×350 RGB0。沒有產品storage／action聲明；初輪素材相對路徑錯誤保留，改用明示DQ3_ASSETS後沿相同image／命令／資源乾淨重跑。臨時Go test已移除。此結果只閉合目前206的render，不能用來假定所有狀態裝備永遠在前。

第二個可丟棄render試作`work/issue4-item-render-draft-r3/`的206物品窗及207「如何」窗完整RGB均0，先撤銷前一活動frame，再依原始4050結構與record421畫新窗口；只閉合UI圖層，不稱正常production action。IDA r3／r4／r5分別271／127／40條目，私人收據SHA-256為`2c8332225704b94070cf2ba75bc24cf164726845b5dd70a1dcc8d92258a47e2d`、`54933f9a0c9c9a53f8d370ac6c1b628924b4ca28d3c0c791ba3a5c4d350a63a4`、`58db34dd94ab4aaecfe1876ad574c2d6cd7ebb4b3e5577167c46e5a94326ee24`。18197→181B1只重算狀態與能力，不整理八格；139AB→13A62同owner給予保留raw word並向八格末端移動，包含空格。1885F單人直接選owner1，不顯示目標selector。尚未正常動態閉合的交易保持strong。

[正常同人給予順序探針](../tools/dosgolem_field_item_reorder_probe.py)沿218路線，再Space／Down／Right／Space進物品，正常選穿戴衣服→Down到給予→Space，單人自然跳過owner selector；若抵達真正等待點，再Enter返回並重新開物品清單，預計230包。CLI使用`--prefix issue4-field-item-reorder-r1`；新增只讀IDA linear13A62／13A87／13A9F／18197的原始八格及暫存器，不改raw word或裝備旗標。未取得接受來源前，不把「裝備永遠排前」或「無空格序列」寫入production。仍需證據審查與存檔表示契約，正式Go／pack保持。

### 2026-10-05 正常七列導覽來源接受，storage仍 DRAFT

正常218包／512IRQ1／256次輸入已由獨立checker完整接受，私人來源`work/dosgolem-opening/issue4-field-item-navigation-r2-source-r1-receipt.json` SHA-256 `28995c8dd702f83d70c51f3f09212dc556356ed563d5b0b9d3977ff4b0d60eb4`。462份產物包含436份完整PNG／bin及26份2172bytes持久區。前206包全部事件／PNG／bin／持久區與父來源5b2a78c1相同，174母親產物完整重驗。原版EXE、工具與位址基準沿上節；不restore、重擲、注入或調時鐘。

| 正常包 | confirmed玩家結果 |
| --- | --- |
| 206／207 | 七列第一列穿戴衣服，Space開「如何」三項，初始游標1 |
| 208 | 動作Esc關閉物品／父窗直接回field，不回物品清單 |
| 209..212 | Space、Down、Right、Space正常重開七列，游標重設1 |
| 213..216 | Down→2、Up→1、Up繞回7、Down繞回1 |
| 217／218 | 清單Esc直接回field，Right正常到3,18 |

193..217完整2172bytes保持，218只改座標word至3，clock全部30、raw526C1、Scratch空。3FD8動態width16bytes／height144px／count7，記錄418／419／420；4050動作窗width12bytes／height80px／record421／三項callback保持原始bytes。不能由此直接把七列序號當Go背包index。

重驗CLI為`python3 /repo/tools/verify_dosgolem_field_item_navigation.py /repo/work/dosgolem-opening /repo/tools/dosgolem_field_item_navigation_probe.py <新收據>`。父checker沿既有repo/work相對EXE路由，所以原版目錄使用此容器路徑；首次誤傳/work別名導致/assets_raw找不到，保留工具呼叫失敗，修正參數後同image／命令／資源乾淨重跑，未改產品或放寬比較。多人取消未知，docs/158以追加勘誤保留被推翻的舊斷言。

同人給予r1在220包被繼承的220包輸入guard拒絕，尚未送221，更未抵達交易writer；不是原版或dosgolem能力故障。實際計畫為230包，r2僅把這項guard改成明示230包，不調時鐘、指令上限或狀態，仍逐節點檢查再送鍵。r1冷啟動全部產物與原producer b69db537保留，r2使用新prefix `issue4-field-item-reorder-r2`。資料格式仍DRAFT，不能把未送出的交易算完成。

[正常同人給予來源驗證器](../tools/verify_dosgolem_field_item_reorder.py)固定r2 producer／Go source／binary，完整重驗218包父來源後，核對230包／536IRQ1、268次正常輸入、四筆只讀writer、完整畫布及2172bytes持久區。CLI為`python3 /repo/tools/verify_dosgolem_field_item_reorder.py /repo/work/dosgolem-opening /repo/tools/dosgolem_field_item_reorder_probe.py <新收據>`。只接受本EXE、工具版本、單人初始穿戴物給予自己這條路線；原版接受不等於remake parity。

### 2026-10-05 正常產品差異與資料方案待確認

正式Go／pack保持e94a4c7。可丟棄診斷測試以正式InputState從新遊戲重播至206..208，沒有狀態注入或重畫UI。206清單差44816、207動作窗差45281、208取消差57417；實際選中背包code0，原版第一件為穿戴布衣raw801E，取消後remake仍停在清單，原版回field。私人`work/issue4-item-normal-red-r2/item-open/normal-red.txt`保存結果。診斷測試PASS只表示已知RED條件重現，不能當parity通過；臨時Go test已移除。

七列／動作窗試作完整640×350 RGB均0，只有render閉合。正式資料仍把背包與裝備分開，沒有物理八格空格與順序，因此不能直接將試作接入產品。資料契約保持DRAFT，待使用者選擇「單一有序物品格，背包／裝備由它產生」或「保留背包／裝備，另加含空格的順序引用表」。兩者均保留原版物品格，需升級存檔契約；舊schema／hash維持明確拒絕，不自動遷移。未收到選擇前不實作相依的storage／save變更，原版證據驗證繼續。

候選寫入調查`work/issue4-item-storage-writes-r1.json`找到八個Go檔的26筆直接賦值，另有指標、slice及helper寫入，這份候選清單不代表完整寫入台帳。實作前須核對取得／購買／消耗／丟棄／給予／裝備／戰鬥／存讀檔入口，將所有修改集中到所選契約的具名操作，再審READY。多人取消、不同人給予、全部物品使用／丟棄與第四動作列仍需各自證據，不因目前單人來源而宣稱原版campaign完成。

來源checker首兩輪保留失敗：r1誤以為關閉清單會清raw count，原始3FD8+14h在219..222仍為7，開動作窗223才寫0；r2誤把沿用的文字觀察tag限定為父路線七筆，新交易自然新增packet225／DI0134的record308。重查spec閘門後，只依原始bytes及正常事件訂正兩項預期，父七筆與218包全部產物仍逐欄／逐byte核對，新增文字只接受這唯一一筆與writer前的時序。未改正式Go／pack、原版或dosgolem；沿同image／命令／資源完整乾淨重驗。

### 2026-10-05 正常單人穿戴物重排來源接受

本EXE115282bytes、SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`，dosgolem2f44a68、seed1357執行前固定一次。正常230包／536IRQ1／268次輸入來源獨立接受，SHA-256 `aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4`。498份產物包含460份完整PNG／bin與38份2172bytes持久區；前218包全部事件／產物逐byte保持，祖先來源及174母親產物完整重驗，沒有restore、狀態注入或時鐘調參。

| 正常包／定位 | confirmed輸入、狀態及副作用 |
| --- | --- |
| 219..224 | Space、Down、Right、Space打開七列，Space選第一件穿戴布衣，Down選給予 |
| 225／IDA linear13A62、13A87、13A9F | 單人Space自然選自己。原始八格`[801E,0000,0001,0001,0003,001F,001F,00FF]`改為`[0000,0001,0001,0003,001F,001F,00FF,801E]`；原始word／空格保持，所選word移至第八格 |
| 225／DI0134 | 自然顯示record308「要給誰」，交易writer已完成，停在正常等待點；既有文字七筆保持，新增只有這一筆 |
| 226／IDA linear18197 | Enter解除等待，重算能力consumer自然返回field；能力／金錢／旗標保持，沒有重新排列八格 |
| 227..230 | Space、Down、Right、Space正常重開物品，游標1；00FF空格跳過，布衣仍穿戴且顯示於最後 |

四筆writer的IDA linear／原始暫存器與只讀槽bytes均留在來源；DGROUP50B9..50C8對應持久dump偏移400..415，位址基準不混用。219..224完整2172bytes與218相同；225..230只將上述16bytes旋轉，其他全部保持。世界clock30、raw526C1、玩家3,18、Scratch空且沒有存檔交易。此confirmed只涵蓋單人第一件穿戴物給予自己，不外推不同owner、其他物品或完整campaign。

| 本機收據 | SHA-256／用途 |
| --- | --- |
| `work/dosgolem-opening/issue4-field-item-navigation-r2-source-r1-receipt.json` | `28995c8dd702f83d70c51f3f09212dc556356ed563d5b0b9d3977ff4b0d60eb4`，218包；r2重生逐byte一致 |
| `work/dosgolem-opening/issue4-field-item-reorder-r2-source-r3-receipt.json` | `aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4`，230包接受 |
| `work/issue4-item-normal-red-r2/receipt.json` | `fde66d178f0aabbca9e4ab32dcb08cf1a343fdc18a5782f8cc56a069ec6da781`，正式正常輸入RED診斷 |
| `work/issue4-item-render-draft-r3/receipt.json` | `5ae40fbeb03ed4028322119d40af8c36cb3f46150f0e7cb29be7739a8ecff64d`，兩張可丟棄render RGB0 |
| `work/issue4-item-reorder-negative-r3/receipt.json` | `3b2b51f8bcf39f3c8141ebb6c407369523a84c40174f2ff7c655ec6ed0140185`，四種破損來源均在指定檢查點拒絕 |

負例只破壞臨時副本：缺最後IRQ1、丟失穿戴flag、交易外持久byte、錯誤record308。最後公開checker SHA-256 `65c595dab8b5204baaab9d98b8ad598e8fce859d9060d57b0c2dbd41545b5917`與來源／負例收據一致；父來源快取只用於負例，正式接受完整重驗父鏈。r1負例沒有指定失敗點，不能取代r3；r2三例有指定失敗點，r3以最後版本再驗並加文字負例。原版產物、私人試作與IDA sidecar均不加入Git。

資料方案仍待使用者選擇，storage／save尚未READY；原版交易接受與render試作不代表正式remake已修正。後續由所選契約閉合正常物品入口、取消／重開、交易與存讀檔，再核對完整畫布。舊docs/158取消概括已保留並追加勘誤；現況只查CONTEXT，Issue #4保持OPEN。

負例後完整正對照r4再次接受，與r3收據逐byte一致，SHA-256仍aad971bb。導覽三種破損來源拒絕後的完整重驗收據r2亦與r1逐byte一致，SHA-256仍28995c8d。

#### 原始定位回填台帳

本台帳的不可變輸入為`DOS／assets_raw/DQ3.EXE`、115282bytes、SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`，位址空間為IDA linear。

| 原始定位 | 限定結論／等級 | 舊規格與回填 |
| --- | --- | --- |
| 1372F caller／1F779..1F8A5 selector consumer | 正常單人action及清單Esc直接回field，confirmed | docs/158第3點已勘誤，保留舊斷言及來源28995c8d |
| 13A62..13A9F writer／138F8、13919 reader／18197 consumer | 正常第一件穿戴物給自己，保留空格／word移至第八格，限定confirmed | docs/158追加230包來源aad971bb；禁止把compact inventory或裝備永遠在前當規格 |
| 13919在docs/159的同伴補給引用 | ownership reader不受影響，沒有新增多人動態證據 | 既有+3A八格讀取仍有效，物品使用／消耗與正式storage另依READY閘門 |
| 1885F在docs/78、docs/171及本檔舊改名段落 | 單人直接選owner1與既有靜態結論一致 | 咒文目標／取消、多人與改名流程不因本次給予來源而升級或重開 |

[物品證據回填檢查器](../tools/verify_field_item_evidence_backlinks.py)以完整EXE hash、原始IDA定位及必需勘誤marker核對docs/188→docs/158，缺檔、缺位址、缺marker或缺來源hash立即失敗。容器內CLI為`python3 /repo/tools/verify_field_item_evidence_backlinks.py /repo`。它只檢查回填關係，不取代正常原版來源及production驗收。

### 2026-10-05 物品寫入影響稽核 DRAFT

資料方案尚未收到選擇，八格storage／save仍未實作。[物品儲存候選稽核器](../tools/audit_item_storage_writes.go)使用容器內Go標準庫語法樹，搜尋`game/`的普通非測試Go檔，保留直接賦值、可能別名、所有解參照賦值、建構／snapshot及裝備setter入口。CLI為`go run -p 2 /repo/tools/audit_item_storage_writes.go /repo <新收據>`；沿dq3-ebiten-test:20260822-r1、UID1000、network none及work可寫，輸出拒絕覆寫。

首份`work/issue4-item-storage-ast-r1.json`記錄69份輸入hash與187筆語法候選。它不是完整型別或別名證明，複製值及唯讀入口也會入選；必須逐筆分類，不能拿候選數當已閉合寫入數。前一輪26筆直接賦值清單亦保留，不聲稱完整。

r2補入戰鬥`.items`及工具hash，為69份輸入／208筆；出售setter修正後r3為69份／207筆，Go1.24.13，工具SHA-256 `aa24aa8029bfcadf124f75d5d1c6df026a3b74b30b8458e5d7e759849f36f50d`。r3收據為`work/issue4-item-storage-ast-r3.json`，每份來源各記hash；候選減少是清除副本改成setter，不是少了一條持有權交易。

| 影響入口 | 語意審查／八格契約後續依賴 |
| --- | --- |
| `itemuse.go`、`panel.go` | 取得／購買、消耗、賣出、同人／跨人給予、丟棄、穿戴與退回舊裝；slice及指標別名不可漏掉 |
| `fieldspell.go`、`reclass.go` | 解參照清裝備、移除道具及多重賦值；直接欄位搜尋不足 |
| `battle.go` | `heroItems`、`actorItems`、同伴`.items`與戰後寫回；目前bag及穿戴物由不同集合重組，尚未保留物理空格順序 |
| `newgame.go`、`recruit.go`、`save.go`及角色建構入口 | 主角、入隊／名冊、臨時單人隊伍與村莊創始人的保存／還原；新契約必須涵蓋這些持有者 |
| 商店／酒館`.items` | 商品資料表引用，沒有持有權；不能把相同欄位名都當庫存寫入 |

稽核另外找到已由既有原版契約支持的獨立缺陷：同伴裝備出售清掉的是`equipActorSlots`回傳的array副本。正常新遊戲至羅馬利亞的正式輸入重現售款增加而盾仍穿戴。有限修正與驗收集中在[商店出售規格](182-shop-sell-runtime-spec.md)，不藉此決定八格資料格式，也不升格物品206..208或商店原版動態V3。

### 2026-10-05 使用者選定A，八格實作契約仍待READY審查

使用者已明確選「單一有序物品格，背包與裝備由它產生」，排除另加順序引用表。存檔升級，舊格式明確拒絕且不自動遷移。選擇紀錄已登記[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4#issuecomment-5984959363)，這是既有待確認決策的答案，不需要再次請求相同授權。

唯一可寫持有權集合為固定容量、有序的物理格；讀出的背包／裝備只是複本。空格保留，code0是合法物品；可見清單用非空格的物理位置，穿戴物仍在清單。取消不交易。自給先保留原始所選格，再將後續包括空格的全部物理格左移，所選格放到最後；依230包來源aad971bb。新增物品寫第一個空格，移除只清所選格，依docs/157的16856..16895與docs/158的writer，不做compact slice。

私有可丟棄核心試作在`work/issue4-item-store-draft-r1/slots.go`及`slots_test.go`，用Go標準庫表示`code`／`equipped_as`，不含文字、座標、record或DQ3容量常數。試作驗證來源八格重排、第一空格、移除不壓縮、滿格不寫、複本不能修改持有者及嚴格欄位解碼；尚未接入production。裝備切換、跨人穿戴物、戰鬥slot映射與完整save adapter仍需各自證據審查，不能因核心測試綠而宣稱八格遷移完成。

初始角色資料已有D3定位：`characters.json`新主角file1C2D..1C33將801E寫物理格0，登錄角色file1C94..1CC4先清八格00FF再寫801E至格0。新契約必須在JSON明示每個初始格，不從四件裝備順序推導；引擎沿ITEM metadata判讀類別，原始decoder保留作oracle。完整存檔需涵蓋主角、同伴、名冊、臨時單人隊伍及村莊創始人；snapshot／restore不能另保存可寫的bag或equipment真值。

審查已找到試作的缺口：[docs/147](147-cursed-equipment-church-service-spec.md)以同一EXE及ITEM hash閉合IDA linear18098..180A9的8000穿戴／4000詛咒writer、17F22鎖定consumer及17254..17266移除所有命中word的教會交易。僅保存code／equipped_as不足以保留原始狀態bits；試作必須修訂為完整保留每格原始狀態，再由pack契約及ITEM類別形成裝備檢視，不能把未知high bits丟掉或依名字重建。現階段四個試作測試通過只證明已知八格操作，不證明canonical表示已READY。

後續完整word試作在`work/issue4-item-store-draft-r2/words.go`及`words_test.go`，前版保留作已被取代的設計紀錄。新集合只保存有序`uint16`原始words，容量、空值、物品mask、穿戴mask及實際archive count必須明示傳入，不設引擎預設。背包／穿戴檢視只回傳複本與原始物理位置；裝備類別仍需由ITEM consumer導出。未知high bits保持原值，不由核心指定玩法語意。

5項試作PASS：正常八格重排／round-trip、第一空格與滿格失敗、4000詛咒及未解釋高bits保持、檢視／載入複本不能寫持有者、缺codec／越界物品／壞序列化word／非法位置拒絕。私有收據`work/issue4-item-store-draft-r2/receipt.json`逐檔hash，words.go SHA-256 `cec017687c71816d5f1ec6357ec89fb14b71fb555d8af37061952c82ee759650`。仍未接入正式引擎；consumer、JSON欄位、save拒絕gate及正常物品UI須整體審READY，不能把這5項稱為production完成。
