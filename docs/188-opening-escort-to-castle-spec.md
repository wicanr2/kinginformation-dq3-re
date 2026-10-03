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
