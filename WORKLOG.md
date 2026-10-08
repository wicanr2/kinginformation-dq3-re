# DQ3 工作歷程

## 2026-10-06：Issue #4 正常屋內NPC圖層遮蔽

原版新遊戲正常422與八筆唯讀分支閉合後，修正renderer依既有scene_tile_layers只畫同層NPC。離開酒館後，兩個室內NPC不再畫到屋頂；位置、對話、碰撞與亂數沒有交易。schema0.32.0／content0.1.105，A八格/storage1/save2保持。

完整game519頂層覆蓋、523命令、468不同頂層／145子PASS，51既有選用SKIP；internal214頂層／531子、12套件PASS，4既有選用SKIP。正常THE END72.10秒、Go vet與Linux desktop PASS，必驗零SKIP，OOM／kill0。正常409／422的F5／F6、Load後下一步與兩路533張最終PNG一致。

422兩個完整人物格1536pixels與原版背景相同，整張RGB3390→2517，仍有2357未解釋的戶外人物差異。較早178..182也修正同一異層洩漏，原先全PNG保持假設已訂正；動畫時鐘、戶外NPC位置與完整原版campaign仍未知。323筆IDA raw定位及舊16註記保持，新增三筆限定confirmed；九JSON乾淨重建一致，七玩法JSON保持。

三種壞來源全拒絕，正對照相同。native／raster checker的未證實假設失敗已保留並縮回實際證據範圍，未調seed或圖片。證據與READY／有限CONFORMED在docs/188，唯一現況表CONTEXT。依授權commit＋push並更新[Issue #4](https://github.com/wicanr2/kinginformation-dq3-re/issues/4)，Issue／Goal維持進行中；下一步正常422的戶外NPC及後續玩家路徑。沒有新包，使用者十三項scratch保持。收尾擁有權稽核維持3213個既有root-owned項目，沒有新增root-owned輸出或.md目錄；本批一次性Docker容器已全部清除。

## 2026-10-05：Issue #4 正式木棒使用與新鍵返回

依docs/188有限READY，修正健康單人主角使用第一列code0檜木棒的兩行原生消息與返回。
原版233包／542IRQ1／507產物接受，父230包／498產物保持，固定seed一次，無注入或restore。
A八格、穿戴與持久snapshot保持，remake RNG不變，正式新Enter、F5／F6及下一步通過。
完整232 RGB18419→106，原始NPC14完整BLS201／200與BLK背景解釋全部差異；完整V3與動畫時鐘未知。
三路207張舊PNG與新路81張前綴保持；九JSON由兩份乾淨1549181重建一致。
schema0.23.0／content0.1.95，canonicald28a7781，save_version2保持，無新包。

game495頂層清單完整覆蓋，444不同頂層／141子PASS、51選用診斷SKIP；internal196頂層／426子、12套件PASS、4選用診斷SKIP。七項必驗零SKIP，正常THE END118.69秒、go vet及desktop PASS，OOM0。
兩類checker各三負例拒絕，重複正對照一致；IDA9.4新三筆原始位址語意自動合併，627筆bytes／名稱／xref保持。

來源綁定、code+1識別與輸出目錄的測試錯誤已訂正；中止兩個含錯誤測試的容器後乾淨重跑。
原始code0 reader推翻早先code1假設，索引與docs/188保留勘誤。JSON排版過度重寫已改回保持排版的遷移工具。
不把這些失敗列為產品缺陷，不改原版或放寬驗收。完整證據及收據見docs/188／84，唯一現況表CONTEXT。

正常233返回場景後重開物品清單，觀察「丟掉」的問題、物品交易與返回。先取得dosgolem原版證據，再依RE→READY修正；已閉合的A、225與木棒使用不重開。 依既有授權commit＋push及更新Issue #4，Issue／Goal仍進行中。
原版素材與完整產物不提交；使用者十三項未追蹤資料保持，Docker批次清理後核對擁有權。

## 2026-10-05：Issue #4 正式225原生給予提示

接續6923b29，依docs/188 READY修正單人自給消息。正式程式保留交易前完整底圖，窗口、frame、
字距、陰影及顯示hold由pack提供；自然顯示後等待新Enter。A八格與save_version2保持，
暫態底圖不序列化。schema0.22.0／content0.1.94，canonical9ac94eed，九JSON兩份乾淨重建一致。

正常194..230逐包八格與原版相同，226返回、230重開、F5／F6與下一步通過。
194..207完整RGB0，225由30140降106，全部由NPC14完整原始BLS201／200及BLK背景解釋；
未解釋差異0，完整225 V3與動畫時钟仍unknown。207張正式路線PNG只有225改變，其餘206保持。
正式兩側完整225 PNG已目視核對，沒有裁切、遮罩、替圖或修改動畫相位。

game493頂層清單完整覆蓋，442不同頂層／141子PASS、51選用診斷SKIP；internal194頂層／425子、12套件PASS、4選用診斷SKIP。正常THE END94.65秒，必驗三路與新確認鍵元件零SKIP，go vet及desktop PASS。
checker兩次正對照一致，壞PNG、錯動畫parity宣稱及NPC外新增像素三負例正確拒絕。
完整收據、canonical與重生工具集中docs/188／84；現況表CONTEXT，下一原版切片為道具使用入口。

首輪21414測試預期bytes誤填，依既有IDA原始sidecar訂正；internal r2漏帶兩項oracle變數，
r3補齊後乾淨完整重跑。公開checker首patch文件anchor缺失而無寫入，訂正實際anchor後完成。
保留失敗收據，沒有因此改原版、產品規則或放寬驗收。

所有工作依既有授權登記Issue #4並commit＋push。原版素材、PNG、binary及IDA資料庫不提交，
十三項使用者未追蹤資料及.claude保持；一次性Docker容器批次後清理。沒有新發行包，
原版完整campaign／音畫／多人保持unknown，Issue與Goal繼續進行。

## 2026-10-05：Issue #4 第225步原生提示RE與READY

接續3832de2，依Issue下一項調查給予提示。原版保留指令窗、舊物品清單與操作窗，再開DGROUP3E6E消息窗口印record308；remake先清父窗再開一般對話。IDA9.4已閉合15002／21414／139C2／13A62..13A9F／2111B順序，547筆指令sidecar保持原始定位與bytes。正常新冷啟動230包、536 IRQ1及498份產物逐byte等於既有接受來源，新增13筆只讀窗口／字模觀測沒有改遊戲狀態或時鐘。

原生renderer試作沿runtime224完整底圖、原始FON、frame404／正文308與既有VGA字組陰影。全畫布30140差異降為106，全部由NPC14完整32×24原始圖塊解釋，runtime BLS201、原版200，其他畫布一致。動畫時鐘保持未知，不裁切、遮罩或改相位；試作尚未接入正式產品。限定READY、完整來源收據及工具入口集中docs/188，下一切片正式接入後驗收194..230、226返回、230重開、F5／F6及新一步。

IDA首兩輪要求自動function而失敗，沿既有明示callback範圍乾淨重跑後成功；沒有推測性rename或function boundary。producer import初輪少掛/work，補上既有輸出掛載後正常執行；原版只冷重播一次。首次遞迴核對所有祖先來源超過60秒外層逾時而無收據，改固定已接受父收據hash並核對兩側全部498份產物後0.81秒通過。試作缺PIL時沿既有PNG decoder完成；首輪猜相位方向失敗，完整原始雙影格唯一匹配後訂正。

公開checker兩次正對照逐byte保持，壞PNG、壞probe source及錯字模座標三負例均由指定斷言拒絕。負例fixture先因跨filesystem硬連結、後因缺IDA sidecar而失敗；補齊同一checker重跑通過，不列產品錯誤。三筆限定confirmed原始位址語意追加既有索引，r4自動匯出547筆並保持原名、位址及bytes。

本輪只提交公開producer／checker、READY與目前狀態。正式Go／九JSON、schema0.21.0／content0.1.93與最近THE END／desktop仍為3832de2；不重跑未變更的產品測試，不宣稱runtime或全流程原版驗收。全部工作依授權登記Issue，私有PNG、EXE、Go原版執行器binary與IDA database不提交。一次性Docker容器批次後核對清理，使用者十三項未追蹤資料及.claude保持。

## 2026-10-05：Issue #4 A正式持有權、存檔與230步重排

接續7f01143，依使用者A決定及docs/188兩份READY，所有角色與戰鬥改用唯一有序Store，移除可寫背包／裝備副本。物理八格、空格、code0及完整旗標保持；商店、原野、裝備、NPC、教會、轉職、創始人與戰鬥交易接入實際持有者。所有五類save owner保存完整words，save_version2拒絕全部舊存檔與無pack身分資料，壞owner在任何restore前失敗。schema0.21.0／content0.1.93、canonical e2cf0b0e；兩份乾淨pack獨立重建九JSON相同，沒有新發行包。

IDA9.4及dosgolem七target有界常式證實只在advanced class清高bits且移除所有書；其他轉職保留words，free source持書亦清除。rec171第一個詛咒word不要求wear，AND00FF清所有高bits。docs/92、181追加勘誤保留舊證據，有限confirmed writer不外推原版正常轉職或音畫。

正常dosgolem230來源aad971bb逐包完整words及正式InputState194..230通過，第一列穿戴布衣自給後physical7，空格保持；226 Enter返回、230重開、正式F5／F6與下一步通過。194..207完整RGB0，225提示仍30140，其他411／229動畫差異不升格V3。完整收據與重生入口集中docs/188；下一切片是225原生窗口／底圖證據，不重開已完成的A核心、pack、save與206／207。

完整game r2覆蓋492頂層清單、496次呼叫，唯一主線失敗是測試持有舊view。此前target及完整r1的錯部位／九格fixture、穿戴重複計數及選中穿戴物等失敗均保留，按實際ITEM／Store契約訂正。補驗r2後段選死亡角色丟棄失敗，r3核對死亡角色限制，後段減員全滅保留；r5因CTY79白天表無設施而停下；r6在城內使用黑暗燈遭既有地表限定gate拒絕；r7改在進城前使用，經鑰匙與section轉場，再經教會／旅店兩次恢復全隊，正常THE END 79.71秒。合計440不同頂層／141子PASS、51選用SKIP；沒有改產品gate、HP或PRNG讓測試通過。

最後正式實作r4另重編三條必驗正常路線，三頂層PASS、零SKIP；新增八列版面邊界負例後internal192頂層／415子、12套件PASS及4選用診斷SKIP，go vet及Linux desktop通過。這是完整覆蓋與受影響補驗，不把補驗說成同一binary完整重跑。1083／1087舊PNG保持，206物品入口及三張出售順序圖受本批影響；出售V3仍未知。正式驗收收據oom／oom_kill0，root-owned基線3213及零.md目錄保持，使用者13項未追蹤資料不加入提交。

全部工作登記Issue #4，依既有commit＋push授權提交；實際commit與遠端核對由Issue末次留言記錄。Docker使用既有image、UID1000及有界--rm容器，批次結束檢查並清理，不上傳私有PNG／素材／IDA資料庫。既有十三項未追蹤資料及本輪另出現的`.claude/`均保留，未納入提交。最終稽核收據由docs/188索引；Issue與Goal保持進行中。

## 2026-10-05：Issue #4 正常物品導覽、八格重排與資料決策閘門

接續e94a4c7。以dosgolem2f44a68、seed1357執行前固定一次，正常七列導覽218包／512IRQ1／256次輸入來源28995c8d接受；單人第一件穿戴布衣給自己230包／536IRQ1／268次輸入來源aad971bb接受。前218包全部事件及產物逐byte保持，174母親祖先來源完整重驗。交易前後只有DGROUP50B9..50C8八格變更，空格與801E穿戴word保持，clock30／能力／金錢／旗標不變；Enter自然返回，再正式重開清單，布衣列在最後。完整證據、分級及五份公開工具入口集中docs/188。

正式正常InputState診斷重現206清單差44816、207動作窗差45281、208取消差57417；remake選背包code0，原版第一列穿戴布衣，remake取消回清單而原版回field。可丟棄render試作206／207完整RGB0，只證明UI圖層；臨時Go test已移除，正式Go／九份pack保持e94a4c7。最近完整game／internal／THE END／desktop仍沿該版本，不因診斷測試PASS聲稱parity完成。

IDA9.4非破壞匯出保留原始bytes／位址／xref／分級，不重新命名或猜函式邊界。首輪1372F沒有自動函式邊界，改用有界raw window；只追玩家可見八格、選擇、writer及consumer，沒有硬體driver／ISR新切片。docs/158取消概括保留舊斷言並追加勘誤，單人初始穿戴物重排升為限定confirmed，多人／其他動作未知。新定位回填護欄及不可變EXE hash／來源索引通過。

正常導覽r1在208觀察到Esc直接回field，與DRAFT預期回清單不符，未送後續鍵便停止已知無路由的研究容器，未建立接受收據。給予r1在220被繼承輸入guard阻止，未送221／未抵達writer；r2只把guard改為實際計畫230包，沒有調時鐘或指令上限。來源checker前兩輪分別錯假定raw count關窗清零與沒有新增文字事件；依原始3FD8及record308訂正契約，保留失敗。這些都分類為驗證工具問題，不當作產品或模擬器故障。

導覽三種破損來源拒絕後完整正對照收據r2與r1逐byte一致。重排最終四種負例逐一在指定檢查點拒絕：缺IRQ1、穿戴flag損壞、交易外持久byte、錯record308；原始產物保持，父來源快取只用於負例。其後完整正對照r4與r3逐byte一致，來源仍aad971bb。試作與RED診斷均零OOM，不能拿UI試作當正式交易驗收。

現行背包／裝備分開保存，丟失物理八格順序。已依grilling共同決策規則請使用者選擇單一有序物品格，或保留現有資料加順序引用表。兩者均保留空格／穿戴位置且需升級存檔契約；舊schema／hash明確拒絕、不自動遷移。資料契約仍DRAFT，沒有選擇前不實作相依storage／save；收到選擇後核對全部寫入入口並審READY，再修正常物品垂直鏈。Issue及Goal保持進行中。

主機gh已授權更新Issue。一次edit-last誤改舊留言，依更新前全文還原並逐則核對185則歷史留言保持，後續特定留言編輯只用明示ID；最終Issue／commit／push讀回另外保存於私人收據。十三項使用者資料保持未追蹤，原版素材、影片、IDA database、試作與所有私人收據不加入Git。Docker root-owned基準3213及零.md目錄保持；所有本輪工作使用有界--rm容器，清理與提交結果在Issue末次讀回核對。沒有新發行包。

## 2026-10-05：Issue #4 正常指令窗、導航與HUD

接續6595dc5。原版正常Space194完整RGB差22519，現行指令窗幾何與導航沿用歷史C移植，不能當原版規格。dosgolem固定2f44a68、seed1357一次，正常206包／488IRQ1來源5b2a78c1，前194包與174母親產物保持，193..206完整2172bytes與clock30保持。IDA9.4只追玩家可見raw視窗、caller、游標consumer及HUD，五份非破壞sidecar保留bytes／位址／分級；未深入硬體driver／ISR。

可丟棄原型完整194 RGB0後先審有限READY，才接入正式game-pack與共用primitive。修正原始record400框架／標籤、游標、線性六項上下繞回與半欄切換，HUD重用已閉合idle契約。schema0.19.0／content0.1.91、canonical66224bc0，九份JSON從乾淨6595dc5兩次重建逐byte相同。正常InputState194..205十二張完整640×350 RGB0；snapshot／RNG／世界clock保持，實際道具入口及取消後F5／F6與下一步通過。

正常trace首輪存檔斷言用了Save前舊Respawn；改讀實際保存的snapshot，未改產品交易。完整R1三項舊測試使用Left→Up選Examine，在原版導航下實際選Equip；依證據改成Up繞回第六項，再用同一image、資源與命令乾淨跑完整R2。首輪失敗保留，不能寫成產品回歸或ALSA故障。

完整R2 game484頂層覆蓋、488次呼叫，433不同頂層／119子PASS、51選用SKIP；internal171頂層／375子、11套件PASS、4選用SKIP。正常新遊戲到THE END190.25秒、Linux x86_64 desktop建置PASS；cgroup memory max374、oom0／oom_kill0，不能聲稱所有記憶體事件為零。1025張舊PNG逐byte保持，指令窗57張含44前綴在兩輪完整測試保持。必驗正常指令窗零SKIP。

其後沿相同R2 binary補驗最近四條正常來源，F6取消、讀回後行走、可比十槽與第二槽共四頂層／兩子PASS、零SKIP／OOM，來源與正式產物均不覆寫。完整R2的選用SKIP數保留，補驗收據7fd7a750另存docs/188索引。

原版來源三種負例與畫布收據三種負例全部拒絕，完整正對照前後一致。來源5b2a78c1、正式trace4293924f、畫布c02f1572均留本機；公開重生入口、原始定位及詳細hash集中docs/188。root-owned基線3213保持，沒有`.md`目錄，背景容器工作結束即移除。使用者未追蹤資料未加入提交，沒有新發行包。

有限CONFORMED只涵蓋正常指令窗／導航／HUD及指定十二張畫面。Enter只經靜態strong與remake等價測試，未有本輪原版Enter動態收據。下一blocker道具206完整RGB44816：原版穿戴衣服加六件背包共七列並保留父窗，remake六列且父窗消失。action與storage對應未閉合，維持DRAFT，不把裝備直接加入背包。入隊返回、人物動畫及完整原版campaign未知，Issue與Goal保持進行中。

## 2026-10-05：Issue #4 正常第二槽存讀檔

接續75b296d。dosgolem正常F5第二槽保存、告別、左移、F6第二槽讀回與左右行走204包／484IRQ1，來源1a5a7c22。194包全部事件／388份PNG／bin與174母親產物保持。PLAYER只改第二筆metadata，其他九筆不變；dragon1完整2172bytes保存與原生讀回一致，clock30→0。原版只跑一次，不重擲、不restore或注入狀態。

有限READY後新增正常InputState驗證，第二槽JSON實際寫入／讀回，其他九槽、snapshot／RNG及世界時鐘保持，Save後Load與新一步通過。5不同頂層／1子PASS；另補六張完整RGB0斷言，只重跑受影響正常測試，共6命令／7筆PASS、零SKIP／OOM。兩條舊路線101張PNG保持，新55張含44既有前綴保持，補斷言前後55張相同。正式Go／pack保持e939db2，最近完整game／internal／THE END／desktop未變。

六張完整畫面194..197、202、203 RGB0；198／199／200／201／204仍差411／356／123／123／229，完整原始人物圖塊、透明像素與底圖核對全部解釋，沒有其他畫布差異，動畫counter與時序仍未知。公開工具、有限READY及收據集中docs/188。下一切片是正常Space命令窗開關、導覽及道具入口；入隊返回及完整campaign未知，沒有新發行包。

三種負例為其他槽metadata被改、讀回持久byte被改及IRQ1缺漏，全部拒絕，完整正對照前後一致。r1已寫完成功收據，但程序結束碼137，原因未確認；不當作OOM或產品缺陷。r2沿用相同工具與資源，外層上限由300秒改600秒，乾淨重跑正常結束碼0，收據逐byte一致。負例收據SHA-256為d00995aa2e0d50e1f3f028edbf95752a4fe13be99e0884ddb54c51dbba980cf0，歷史紀錄保留。

## 2026-10-05：Issue #4 正常 F6 十槽取消與行走

接續8cd1ed5。dosgolem冷啟動正常F6、Esc取消、左右行走197包／470IRQ1，來源426c7623。前193包事件／386份PNG／bin與174母親產物保持；完整2172bytes取消前後保持，左移只改座標低byte、右移回復，clock30及raw526C=1保持，Scratch空且沒有原生存檔交易。

有限READY後新增正常InputState驗證。取消立即返回、不改外部十槽資料／snapshot／RNG，左右行走、Save成功後Load與新一步通過。四項回歸共4頂層／1子PASS，零SKIP／OOM。194選槽、195返回及196左移完整RGB0；197右移差122，完整原始英雄6／7圖塊、透明像素與底圖核對全部解釋，動畫時序仍未知。舊路線53張PNG保持，新取消路線48張含44既有前綴保持。

首輪收據缺外層原版hash；補齊契約後重驗同一原版產物，r1／r2與log均保留。第二輪的Save／Load斷言把Save前舊Respawn當基準；只讀診斷證實唯一差異是Save預期更新復活點，將比較基準移至Save成功後。未把兩種驗證問題當產品缺陷，正式Go／pack保持e939db2。三種壞來源拒絕，正對照前後一致；收據、公開工具與有限CONFORMED集中docs/188。

下一切片為正常F5／F6第二槽存讀檔。入隊短曲返回保留unknown及硬體停止線，Issue與Goal進行中；最近完整game／internal／THE END／desktop未變，沒有新發行包。

## 2026-10-05：Issue #4 入隊返回有界觀察與停止線

接續5559d72。正常入隊父來源199包、474IRQ1及398份PNG／bin在兩次冷啟動中逐byte保持。r1在原生raw0013 bit4000出現時被沿用的登錄guard停止，分類為probe適用範圍問題。r2只將guard限制在前199包，保留原生旗標，以同一工具時鐘跑滿2,500,000,001指令；計時器前進3003，仍未觀察到自然返回，沒有後續鍵盤輸入。

核對既有IDA9.4的18筆玩家層等待bytes；原始運算元、位址與unknown分級保持，未追2898 writer或22E10 callee。遵守硬體driver／ISR停止線，不用清旗標、改時鐘或加等待特例補出結果。完成checker拒絕不完整來源，沒有完成收據；有限觀察不證明產品或模擬器缺陷。

兩項remake回歸含一子測試通過，零SKIP／OOM，50張runtime PNG逐byte保持。正式Go／pack仍e939db2，最近完整game／internal／THE END／desktop不變。本輪工具、DRAFT與私人收據索引在docs/188；下一正常切片為合法193checkpoint的F6 Esc取消及行走，Issue與Goal保持進行中，沒有新發行包。

## 2026-10-04：Issue #4 可比十槽顯示資料與正常 F5/F6

接續c01e44a。原始PLAYER的20bytes槽資料、CHINA.FON檔案位置與D3字模逐byte閉合，三個字模106／144／303唯一匹配。IDA9.4保留原始位址與260條目sidecar，raw gender1減1對應JSON0；沒有新增產品字型、資料格式或猜測設定。

合法193checkpoint只建立外部JSON測試前置資料，不改執行中snapshot／RNG。正常F5選槽195完整RGB0；F6選槽199完整差123，全為窗外NPC15原始26／27步伐影格。原先4064／32847來自不可比初始槽資料，這次修正驗證條件，正式Go及九份pack仍等於e939db2。其他九槽世界狀態、完整F6 V3與動畫時鐘仍未知。

兩批17命令／22筆PASS，9不同頂層／5子測試，零SKIP／OOM。200張舊PNG保持；可比路線53張中的51張保持，只改195／199。完整畫布checker四種壞證據拒絕，正對照前後一致。收據、hash、RE／READY／CONFORMED與公開重生入口集中docs/188。

驗證腳本的Scratch相對目錄、遞迴PNG計數及唯讀掛載寫入三類問題已按既有契約修正，保留失敗紀錄；未當作產品缺陷。最近完整game478、11個internal、正常THE END201.00秒及desktop仍為e939db2。本輪不冒稱全套重跑，沒有新發行包。

下一步正常入隊短曲播放完成後返回選單。Issue #4與Goal保持進行中，人物phase-only差異不重複開切片，不深入硬體driver／ISR。

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

## 2026-10-03：Issue #4 正常王座接近與攝影機修正

接續572e3bb，工作登記5962742201，進度留言5963166706。命中dosgolem、對拍、IDA9.4及規格閘門路由。
原版正常冷啟動90次輸入／180IRQ1、386唯一產物核對PASS，先前214份PNG／bin逐byte不變。
固定乾淨dosgolem2f44a68，EXE／原始資料唯讀，seed1357沿自然Lv1入口一次固定。
原版正常上樓至9,22；上樓後閒置窗消耗一個上鍵，最後9,8，五次Enter沒有觸發謁見。
追加來源工具與嚴格核對器，沒有把送出九十次按鍵提升為完成國王交易。
1991D的堆疊欄位只保留raw words，不猜far return；213C4不是國王21414入口，限制回填docs/188。

IDA9.4三份私有匯出核對原始bytes／MZ relocation、函式邊界及xref。
handler56的文字呼叫在六件物品、50金及clear17h／set18h前；record78有九個FFFC等待。
此為strong靜態順序，缺自然玩家交易，不直接改正式獎勵或猜補時序，未深入ISR／音訊driver。

首次上樓的原提交clamp造成主角低96px，完整RGB差86603。
隔離camera原型只增加section1綁定，完整差降至2159，保留柱子圖層及人物差異。
有限READY後新增正式JSON camera；schema0.7.0／content0.1.76，所有值由既有typed契約消費。
canonical hash `sha256:ccbf6f5f2add47996f6e48bcd35a81a5553428517a9b0bfbb25436edaf5e7469`。
正常上樓、完整PNG、同版本Save／標題Load恢復camera與旗標、後續上鍵移動通過。
正式PNG逐byte等於原型；完整RGB仍2159，不稱整張V3。不同pack hash的存檔照原契約拒絕。

gamepack86頂層／238子、六項受影響game、desktop Linux x86_64 ELF通過，沒有SKIP。
正常新遊戲InputState至THE END81.38秒通過，只屬重製回歸。
本批未重跑全部game及其他internal，前批結果保留其版本界線。
九種來源收據損壞均拒絕，原始收據未改；轉換器從乾淨572e3bb重建九份JSON與正式檔完全相同。
私有收尾`work/issue4-throne-camera-final-audit.py`產生27698bytes收據，SHA-256
`34085fc4334bee169597bbf026dfdcdeee65850b1c08709840759163bee69ec4`。
所有新增工具、測試、私有RE與驗收入口索引在docs/188，現況表與docs/74同步。

自動核准審查曾拒絕主機shell wrapper，理由是可能在主機執行dosgolem。
改用明確的一次性Docker命令後核准，沒有退回主機執行工作負載；此為環境／命令審查問題。
較早後段原型遭SIGKILL且未產生收據，不當作產品缺陷或完成證據；縮小到首次上樓正常重跑。
一次編輯命令缺Docker標準輸入選項，沒有執行修改；另存v2並核對source／差異欄位後重跑。
新輸出UID/GID1000，既有root候選3213及Markdown目錄0維持，未建新image或發行包。
下一切片閉合王座圖層與section1閒置窗，再以正常輸入觸發謁見；Issue #4保持OPEN。

## 2026-10-03：Issue #4 王座自然等待窗與柱子判讀勘誤

接續61f60d3，工作登記5963310715，證據勘誤及試作登記5963410940。
命中dosgolem、GUI對拍、規格閘門及文件職責路由，更新README前另讀現況聲明規範。
原始CTY25四個柱子word、載入後記憶體與DQ31.BLK index14一致，Scene圖號及palette亦正確。
獨立PNG解碼與原版bin逐色號差0；2159完整RGB差異全部在六個人物圖格，四柱及底圖差0。
撤回前輪「柱子圖層差異」，保留舊斷言及新反證；沒有新增圖層來修不存在的缺陷。
前輪final.state只作零續行的記憶體讀取，不由缺CRTC的disk state重生畫面。

原版90次／180IRQ1／386產物及前輪214份PNG／bin保持。第7次waiting與第8次restore在9,22。
正常298／299／300邊界、19952入口、17DE5動態header及58bytes角色資料閉合且與前層相同。
強化既有來源核對器，不重寫或注入原版收據。十二種壞來源拒絕，新增三項同步改暫存log／metadata／manifest。
乾淨61f60d3同提交試作只加既有場景綁定，未改位置、旗標、frame或重擲seed。
基準未開窗完整RGB差10143、視窗8458；試作正常開窗視窗及陰影差0、完整1772。

有限READY後正式JSON新增王座scene，schema0.7.0／content0.1.77，無新增production Go分支或版本常數。
canonical hash `sha256:9d2f7a1aa24b999258e8645258ae9f7af8a8c5169088a9bc8c5d495d96fdc5c0`。
正常70次輸入至上樓，自然188更新等待、凍結、關窗方向鍵消耗、按住不移動、放開後續行通過。
完整等待／恢復均1772，等待窗及陰影RGB差0，正式等待PNG逐byte等於試作；不宣稱整張V3。
自然開關完整快照保持，明確Save既有respawn更新另取基準，正常標題Load完整恢復並清除UI暫態。

首輪七項共用圖形程序137終止，最後一項未有結果；原log按hash保留，改同限制／工具鏈逐項新程序。
新測試首次比較Save前後快照失敗，診斷只差Save既有respawn更新；訂正測試基準後乾淨重跑，未改產品存檔。
一次唯讀搜尋遇到歷史`.go`目錄，改為只讀普通檔案；沒有修改該目錄。
gamepack86頂層／238子、七項受影響game、desktop Linux x86_64 ELF PASS，沒有SKIP。
正常新遊戲至THE END129.64秒，只屬重製回歸；未重跑全部game／其他internal，不挪用前批全套結果。
九份JSON從乾淨61f60d3重建逐byte相同；新測試、工具與私有收據入口掛回docs/188。
獨立收尾`work/issue4-throne-idle-final-audit.py`產生6041bytes收據，SHA-256
`131aa62782373f3c6b6bcca7ecc2c7c4be9bc30b8f17fb3ee4c63ac023dc002b`。
CONTEXT唯一狀態表、PROJECT_MEMORY、docs/74、docs/84及README同步；原始素材與私有圖像不進Git。
輸出UID/GID1000，既有root候選3213、Markdown目錄0保持。一次性Docker均清除，未建新image或發行包。
下一切片為正常國王文字及獎勵交易來源，Issue #4保持OPEN；動畫、原版讀檔、音訊與完整原版主線待驗。

## 2026-10-03 Issue #4 正常謁見文字與延後獎勵

接續e56a8e3及已授權Issue／commit／push。重讀規格閘門路由、docs/74與docs/188，
沿用唯一CONTEXT狀態表。原版由dosgolem2f44a68冷啟動延伸至100次正常輸入／200IRQ1，
435個產物核對，前批384份PNG／bin保持。來源SHA-256為b875807a2ae53cbdb9c02e43952e9e894e9230f76d64c469ffbea36c08b02f32。
九次原生等待、文字返回、六次物品writer、50金、clear17h／set18h及正常runner閉合。

乾淨基準提前發獎且實際Enter停在首頁。隔離試作改用已審查保留文字原語、EOF後交易，
九頁視窗與陰影差0、完整均893。有限READY後正式新增共用region_dialogue_reward原語，
入口、handler、文字、旗標與獎勵移入JSON，source validator核對實際CTY及完整D3TXT字碼。
schema0.8.0／content0.1.78，hash82ca3334826bc909a65636e75a46895d4ee99b832cad1cd2cb0f430fd6029f2f。
重製0x200里程碑維持，另以engine D2證據標記，不冒稱原版旗標。

正式正常100輸入、九次Enter、EOF後交易、一次性、後續下鍵、Save／標題Load通過。
真正Save的謁見前狀態重新進入後，同一Game Load取消pending且不給未完成獎勵。
九張正式完整PNG逐byte等於試作，獨立全RGB各893，只在15,3的438及2,9的455個NPC像素。
視窗及陰影差0；完整畫面、滿欄提示、原版存讀檔及音訊不宣稱parity。

舊fixture未退出標題、連按64確認不足，以及自然EOF後多送Confirm開命令窗，均定位為測試問題。
訂正helper後同工具鏈乾淨重跑，沒有改產品輸入。容器快取路徑修回既有工具設定；保留失敗log。
獨立負面收據發現原先缺事件step單調檢查；強化verifier後十二種壞來源全部拒絕。
產生器尚存舊user_report分類造成JSON重建差異，訂正為engine後九份JSON逐byte通過。

最終資料包完整game418頂層清單由21程序覆蓋，378頂層／70子PASS、40選用SKIP。
全部internal149頂層／250子PASS、4選用SKIP及11套件，desktop Linux x86_64 ELF通過。
正常InputState至THE END75.37秒，只屬重製回歸，無素材缺失SKIP。
既有位址索引追加六筆confirmed語意，原始IDA9.4定位與前十筆保持，來源均引用本次正常閉環。
工具、私有產生器、驗證與收尾入口掛回docs/188；CONTEXT、PROJECT_MEMORY、docs/74、docs/84、README同步。
收尾收據11805bytes，SHA-256 9ec36ece96411c0229780666aaa4a14febaf13706dc49fea14f541e83de569d9。
輸出UID/GID1000，root候選3213、Markdown目錄0保持；原始素材與私有PNG未加入Git。
本輪沒有新發行包；Issue #4保持OPEN，下一切片為正常謁見後離城與酒館來源。
提交前掃描：新增production Go沒有版本raw ID；命中只在原始CTY格式遮罩與資料驗證範圍。
game40與internal4個SKIP均已抽核理由，沒有素材缺失。一次性Docker容器均清除，未新增image。

## 2026-10-03 Issue #4 正常回程、酒館圖塊層與登錄所攝影機

接續0c47537，沿已授權Issue／commit／push及active Goal續行。重查規格閘門、dosgolem與文件職責路由，
沿用CONTEXT唯一現況表。原版由固定dosgolem2f44a68冷啟動，正常186次輸入／372IRQ1，
85次步行觀測、607產物，前批432份PNG／bin逐byte保持。來源SHA-256為
347600d38f062e111a08a852a0cdf2a8937b959827c37ff984f3dca44a732a75。
Lv1 seed1357只設定一次；後續等待及NPC骰序保持自然執行，沒有狀態或frame注入。

固定20m指令間隔跨過27..41m城鎮繪圖，IRQ送達不等於步行已消費，不能據此修移動規則。
舊1991D的origin_y實為繪圖後暫態raw4F27，追加勘誤，原收據保持。
原生磁碟snapshot的PIT觀測器未完整恢復，診斷停在等待窗；沒有有效原版收據。
R2在2111B入口送Enter被2112C..21132初始化清除；R3雖正常關窗，但計數條件阻擋步行。
R4改在21133初始化後關窗、1997C原生返回後送步行，沒有改遊戲計數器。
實際11991完整重繪只有15、36、74、85四次；最初verifier要求85次的條件已訂正，未冷重跑來源。
相關原始Go／log／PNG保留，僅R4通過的來源347600d3作本批oracle。

未修改0c47537正常回程在74的CTY00 section0、5,22缺自然等待窗。
隔離idle、camera、both候選逐項比較，位置／section／背包／50金／旗標保持。
登錄所camera令85完整差異由200467降至118；城鎮layers令74由146268降至1403，85張沒有惡化。
36的13983在所有候選相同，撤回以它否定layers的初判，保留原因未知。
有限READY後正式JSON加入城鎮等待窗、玩家格圖塊層、登錄所camera及validator所需場景引用。
schema0.8.0／content0.1.79，hash9757fa4135987cec259a1075e5ea12fe9270703c9bda471ff1fbca79bba81790。
正式引擎沿用既有原語，版本數值不新增為Go fallback。存檔格式維持，不同hash仍明確拒絕。

正式186次InputState、85個玩家狀態與四次原生camera通過。
單人等待窗正文及陰影RGB差0，正常Enter只關窗；標題存讀檔與後續下一步通過。
88張runtime PNG中的85張步行與已審查both候選逐byte相同；完整酒館1403、登錄所118、城鎮13983、等待1095仍RED。
酒館多畫NPC及人物相位、城鎮入口多格差異、原版音訊與存讀檔未閉合，不能稱整張V3。

完整game第14批首輪卡在5,28等待NPC讓路。只加觀測的乾淨副本確認等待窗開啟，NPC被凍結。
導航test helper以正常Enter關窗後等待，不改正式規則或注入位置。保留初次失敗hash0ca91015db05，
同工具鏈重跑14及未執行15..21通過，前13批同一1.79 pack的PASS保持。
完整清單419頂層，374頂層／70子PASS、45選用SKIP；本批原版回程來源測試另嚴格PASS。
internal149頂層／250子PASS、4選用SKIP及11套件通過，全部SKIP理由已核對，沒有素材缺失。
正常新遊戲InputState至THE END151.74秒，只屬remake可玩回歸。
desktop建置完成後因缺file命令而失敗，另查ELF64 little endian／machine62通過，不誤記產品缺陷。

16種協同損壞原版來源拒絕，原始收據未改。九份JSON從乾淨0c47537重建逐byte相同。
只重寫受影響集合以保留其他JSON格式，解碼資料及canonical hash保持。
一次收尾掃描遇執行環境.aws控制目錄消失，只對該控制目錄處理競態，再以同命令乾淨重跑。
收尾收據112587bytes，SHA-256為3393c711ed6d9549b28613bf14ed97a205ba7681f74d0116cfb8a556ffc66c05。
工具、測試、私有重建／診斷／稽核入口均掛回docs/188；CONTEXT、PROJECT_MEMORY、docs/74、docs/84及README同步。
遠端結果留言5965833865，Issue #4保持OPEN；下一切片從正常登錄所入口續行交談與登錄。
輸出UID／GID1000，既有root候選3213與Markdown目錄0保持，沒有新image或發行包。
提交前diff check、Go格式、Python／shell語法及新增production Go版本raw ID檢查通過；
本批Go變更全為測試與原始格式oracle，原版素材、私有PNG與使用者scratch不加入Git。
提交／推送及最後Docker清理結果保存至`work/issue4-return-ready-post-push-receipt.json`，入口同掛docs/188。
原版重生工具移除外層只准執行一次的前綴限制，沿用共用產生器的逐檔hash歸檔契約；
已接受來源及frozen Go保持，沒有因包裝修正再冷重跑原版。


## 2026-10-03 Issue #4 母親勸告與強制返回

接續562208a，沿已授權Issue／commit／push及active Goal續行。
重查規格閘門路由，沿用CONTEXT唯一現況表；dosgolem為正式原版oracle。
登錄所R8只接受首次問候13狀態；完整取消與出生受阻，沒有製造完整來源。
較短正常路線第一個Down即分歧：原版21,17，原提交remake21,18。
IDA9.4保留1020B原始定位及bytes，閉合CTY handler55、flag17h、NPC0轉左、
record79、自動關窗與194C3碰撞檢查後北行；不追硬體逐週期。

窄任務R1工具誤觸舊生成器的歸檔清理段，沒有執行原版；295份證據由先前hash歸檔完整恢復，
逐檔大小及SHA-256相同。R2與追蹤中的正式R3改在獨立tmp生成，既有來源保持。
R3冷啟動39次正常輸入／78IRQ1，174份父PNG／bin保持，十個原生邊界及單次seed1357通過。
正式來源63ee434e；完整PNG、色號bin與R2逐byte一致。沒有狀態注入、重擲或人工挑frame。

有限READY後新增共用轉向→文字→hold→返回狀態機，版本資料與字碼放入JSON。
schema0.9.0／content0.1.80，hash d2ea836e3df3c1c39e02481e73d2efcb3eae41340aa0cfe89debf2e06162fc4e。
正常Down、自動關窗、返回、重複觸發與旗標／金錢／背包保持通過。
Load後原NPC0被初始visibility flag80過濾，依既有pack最後arrival frame與原始靜止actor作有限恢復。
不新增存檔欄位或改旗標；此項是engine D2，沒有宣稱原版存檔或朝向parity。
同一Game Load清除pending、標題讀檔與下一次Up通過。
文字窗與陰影RGB差0；完整提示13145、返回13340保持RED。

14種壞契約、七種壞資產、18種協同壞來源拒絕；原始來源不改。
九份JSON從乾淨562208a重建語意相同，canonical hash與正常PNG收據一致。
首次圖形DISPLAY未就緒及舊negative fixture缺新集合皆屬驗證工具問題，訂正後乾淨重跑。
第一次整套程序中止未留下斷言，原因未知；第二次600秒逾時，仍在後段航行，沒有OOM事件。
保留同一最終binary已完成的261項，再以獨立程序補完其餘清單與campaign。
完整game422頂層清單覆蓋，376頂層／77子PASS、46選用SKIP；本批原版正常勸告另嚴格PASS。
internal151頂層／264子PASS、4選用SKIP及11套件，desktop Linux x86_64 ELF建置通過。
正常新遊戲InputState至THE END266.49秒，只屬重製回歸；無素材缺失SKIP。
最小充分收尾入口與私有產物均掛回docs/188，現況與欄位契約同步。
原版素材、PNG、RE database與使用者scratch不加入Git；沒有新image或發行包。
完整背景、NPC相位、原版音訊／存讀檔及登錄所取消／出生仍待驗，Issue #4保持OPEN。

文件同步腳本首次有縮排錯誤，尚未執行任何寫入；修正語法後同命令重跑通過。
獨立收尾收據24,724bytes，SHA-256
`b837a0d89ae890a60086fef35b15440d7f017dd141a3dbdfec1eb47d85ab02cb`。
新來源與輸出UID／GID1000，root候選3213及Markdown目錄0保持；本批一次性DQ3容器全部清除。
Python／shell語法、Go格式及diff check通過；新增production Go無DQ3專屬raw ID／座標／record／flag。
提交／推送與Issue結果保存在`work/issue4-mother-return-gate-post-push-receipt.json`，入口同掛docs/188。

## 2026-10-03 Issue #4 原生輸入消費來源探針（DRAFT）

接續1b96eb8，上一輪已改正式狀態並完成推送，屬實際進展。新輪主機auth與Issue #4重讀通過，
既有未追蹤scratch及登錄DRAFT工具保留；正式schema0.9.0／content0.1.80維持。
重查規格閘門、現況及dosgolem能力入口；R8完整來源仍未接受，沒有重啟已終止的原版程序。

重用已存在的IDA9.4匯出，確認195D0低位元writer受4F46的0200、0004及52F8閘門限制。
19810只清actor的38h bit10與4F46的0200，未有玩家可見動態閉環；不能推定為命令窗或戰鬥。
原始符號、位址、bytes與strong等級保留，不追硬體driver或以CX／旗標patch解來源。
舊探針每鍵另等11000000指令，目前只是假說此等待會推過原版計時門檻。
1941C的AH方向碼、2113A及21155的正常讀鍵分支已由原始控制流定位，建立只讀消費候選。
新的`work/issue4-registry-consumed-probe.py`在隔離tmp生成，凍結Go及binary。
冷啟動前38鍵保持，之後仍走已記錄的謁見、回程、首次問候及正常姓名取消路線。
IRQ排入、按下／放開、消費、自然poll與完整raw狀態均另行記錄，不修改時鐘或位置。
來源前綴`issue4-registry-consumed-r1`，plan與工作登記同掛docs/188。
遠端工作留言5969155742；完整取消來源、出生及正常remake對拍仍待驗。

原生消費R1已拒絕並主動停止，192個觀測保留。首個Up捕捉21,15，起點21,17。
packet61返回國王前field，packet62多送Enter重開record253二選一；後續方向都在選單內。
這次來源沒有到登錄所，不能只看read-key及poll便接受路線。計時門檻假說仍未證實。
前174份父PNG／bin逐檔大小及hash保持；Go／binary／producer凍結身份一致。
拒絕稽核99,784bytes，SHA-256 62256651ad897270a7c076e3d4f192717493cb8d1b6dfae60d33d29bd7ccbe9c。
原版容器因已確認不合格由agent停止，exit137屬主動清理，沒有把它寫成產品失敗。
下一版觀測keyup後的正常空讀鍵，首三步不符即拒絕；按已證實的謁見交易恢復field後進回程。
正式Go、pack與已推送1b96eb8保持；新的私有入口與限制同掛docs/188。

追加勘誤：quiescent-r1完整keyup及空讀鍵後仍在首次Up播放record79，從21,16自動到21,15。
首三次Up的單格期待沒有原版依據。北側21,16的CTY00 word0001、selector0，沒有特殊事件。
「北側格觸發母親勸告」已收回，不把原生記錄出現文字直接當成地圖事件來源。
既有IDA9.4顯示特殊格移動設4F46 bit0800及258C，強制帶路不作一般dispatch；
殘留事件待新版原生觀測閉合，維持strong，不改正式規則。
新版私有quiescent-r2另記錄上述raw狀態及handler入口，計畫與完整產物同掛docs/188。
Issue工作與勘誤留言5969651364；前兩版失敗來源與拒絕收據保持。
乾淨1b96eb8隔離副本的正常母親帶路最後RGB差0；首次Up完整等待2200更新後仍為21,16，
沒有pending或文字，0金及兩個故事旗標保持。此為重製診斷，不稱原版parity。
私有入口為work/issue4-first-up-baseline.py、同前綴test及log。
診斷腳本首次生成因引號巢狀語法失敗，未產生檔案；修正後同容器命令乾淨重跑通過。

## 2026-10-04 Issue #4 首次北行殘留事件與特殊遭遇分派

quiescent-r2自然完成202次正常輸入／404IRQ1，首次北行前4F46=0800／258C=1。
196D2先清事件位元再進handler55，record79後強制北行到21,15；後兩次Up到21,14／21,13。
由有限READY實作deferred事件穩定ID，普通強制帶路保持、成功一般步進消費、阻擋不消費、新事件覆蓋。
同版本存檔保存未消費ID並在Load前檢查pack／scene／gate及必要actor，已消費不重新生成。
存檔為engine D2，原版未知；版本內容仍從pack取得，沒有新增Go raw ID或fallback。
schema0.9.0／content0.1.81，hash `sha256:d6c7994ee7fa53d227a98263f12606a8f4109a87f4afea01cd6996ad7f269b5d`。

有限來源c38bcc09嚴格通過41正常輸入／82IRQ1、174父PNG／bin保持與14種協同損壞拒絕。
同一最終binary正式冷母親返回與首三次Up、前後Save／Load、既有Down／重複勸告通過。
首次警告完整RGB差0，北行完整195／9039／22192、Down13145／13340仍RED；不宣稱完整V3。
重製新存檔不等於原版存檔對拍，音訊與後段亂數內部骰序未知。

完整回歸首輪CTY10對話確認64次失敗，三版只讀診斷定位bossIntro及battle.active同時成立。
先前通用文字等待調整未解決阻塞，已收回；直接原因為隨機遭遇搶走確認鍵。
verified程序在後段航行由agent主動停止以回復原測試helper，未通過整段主線；exit143不寫成產品失敗。
IDA9.4既有19574→196D2後19577 RET跳過一般encounter尾段，有限READY後對boss事件當步return。
既有自然觸發測試增加seed1357／counter1邊界，修正前active=true／counter17失敗，修正後無battle且counter1。
原始handler14 file19D10 word5477→IDA linear15477→file67E7；舊file167E7基準錯置，docs/85追加勘誤。
沒有改遭遇機率、編隊或正式seed；四敵編隊保持。

後續拉之鏡寶箱未給物品、present flag仍true；r4背包列表不足以排除容量，r5確認全隊含裝備各8格、空位0。
只調整測試玩家補給策略，經正式選單丟棄未裝備且無pack場景用途的備品；保留遊戲容量與交易規則。
原始失敗、八版只讀診斷及原提交campaign PASS199.32秒均留本機，未覆寫來源。
r6的180秒堆疊落在拉米亞停泊點cd空等，Game仍在battle.active分支；最後一步遇敵後測試未處理modal。
沿用既有正式戰鬥選單處理遭遇，再等待冷卻並搭乘，設與登船helper相同8192有界檢查。
clean-final舊程序也由agent主動停止，exit143屬工具清理；沒有把測試死等寫成產品缺陷。
r5主線168.04秒在巴拉摩斯回程缺rec175法力；r7記錄預留角色Boss前MP27，CTY66 sec0、9,29存活施法者只剩1／0。
曾試作全部MP預留，r6仍失敗，已收回；r8時間序列確認沿途逃跑時同伴先施法，每次仍消耗12MP。
CTY66入場MP241、Boss前27、全部預留時Boss後仍27，但回程普通遭遇降到1，賢者57降到0。
全域保留MP的r8與僅停用逃跑傷害咒文的r9均在前段mon44全滅，兩者已收回。
路由原聖水檢查只計算主角背包，漏掉牧師三瓶；改用既有countPartyItem與正式道具持有者選單。
原戰鬥與保留最後一瓶策略保持；不改速度、技能、MP、seed或戰鬥規則。
r7新路線在金皇冠寶箱容量滿時未取得物品，取物前經正式選單丟棄藥草騰出一格，全隊持有權與存讀檔斷言保持。
完整game426頂層清單覆蓋，379頂層／86子PASS、47選用SKIP；原版首次北行與既有Down正常對拍另嚴格PASS。internal151頂層／264子PASS、4選用SKIP及11套件，desktop Linux x86_64 ELF建置通過。正常新遊戲InputState至THE END118.78秒，只屬重製回歸；無素材缺失SKIP。
完整回歸入口work/issue4-first-move-regression-r10.py，正常對拍production-r10，證據與工具同掛docs/188。
CONTEXT、PROJECT_MEMORY、docs/74、docs/84、docs/85與README同步；來源與私有產物不加入Git。
Issue留言5970305881／5970492139／5970591714／5970708867登記容量、拉米亞及法力診斷，Issue保持OPEN，完整取消及出生來源仍DRAFT。
收尾稽核、提交／推送與Docker清理另記同前綴final-receipt及post-push-receipt；沒有新image或發行包。

## 2026-10-04 Issue #4 登錄所取消來源與正常分歧

接續07c01dd。取消來源獨立審查202次正常輸入／404IRQ1、164組queued／consumed／capture，
328份完整packet PNG／bin及174份父產物保持。來源348622bytes，SHA-256
0467c01bbf8218a0c21581cf200cc0b491284db53c72749ddb3631b79b2cf40e。
新增可重生且不覆寫的來源稽核工具；獨立重建一致，13種隔離損壞全部拒絕，原版未修改。
初次來源審查器將觀測階段視為不變項；packet43讀record78後由0到1，修正審查器再乾淨重跑。
搜尋器遇到同名.go目錄而停止，改為只讀普通檔；不刪除或更改既有目錄。

乾淨07c01dd的正常新遊戲前149個位置、section、旗標與金錢吻合；原始Enter在packet150缺少問候，
完整畫面差29540。操作綁定使Enter與Confirm分開，不把裸Enter當作現行命令窗確認。
明示等價交談經正式命令窗送兩次Confirm，仍直接進tavern.active／stage0、dlg=false，完整差124152。
兩個已失敗的可丟棄診斷、JSON與PNG留本機；未加入production測試或修改期待以製造通過。
原版550問候／詢問→554姓名→558取消再詢問→560選否返回行走已閉合，
完整名冊不變性、出生、原版存讀檔與音訊保留unknown。

出生探針先因舊前綴凍結Go已存在拒絕，之後因Go宣告順序編譯失敗；兩次均未啟動原版。
保留舊檔與失敗日誌，修正宣告順序並核對未占用normal-r2前綴，同一專用工具鏈開始冷原版蒐證。
來源唯讀、seed1357只固定一次；只有正常IRQ1輸入與IDA原始位址的唯讀觀察。
正常出生未完成前，取消／出生的正式流程仍DRAFT，production保持0.1.81。
CONTEXT唯一現況、PROJECT_MEMORY、docs/74及docs/188同步。Issue留言5970984356／5971266398已登記。
所有本輪入口與私有產物見docs/188；出生有界工作尚在執行，收尾前核對並清除已終止的DQ3容器。

同一出生工作已自然結束，normal-r2共210次正常輸入／420IRQ1、172組packet。
獨立審查及已保存的公開工具重建一致：344份packet PNG／bin、174父產物與前153組取消來源保持。
原版在姓名完成後開六職業，再選性別，顯示能力等待，接受後才呼叫10A9F。
本次DX1，97bytes由DS520B複製到DS535E；10816之前slot1旗標0，1081C之後1。
直接觀察的是候選128bytes與slot旗標；目的record bytes未直接擷取，不外推其他職業／性別。
出生來源368685bytes，SHA-256 de818064f9bcaf3f36dcd6a5d8b9259908781356a27c128cf1e8b41218c10ce3。
13種隔離損壞拒絕通過，原版與父來源未修改；producer與執行私有版本逐byte相同。
產生器與獨立驗證工具受版控，原版資產、source收據及PNG留本機。
取消稽核提交6b44d35已推送；出生證據與當前閘門追加同一docs/188，不把來源通過寫成remake完成。
production仍0.1.81；下一步完成視窗consumer與READY契約，再修正正式登錄流程。

## 2026-10-04 登錄視窗consumer與有限READY契約

依Issue #4從c810d49續行。命中復古remake／對拍路由，載入規格閘門及文件職責；
IDA工作另沿用use-ida-pro-9-4技能、專用README與工具契約。
既有IDA9.4 locked-v1以UID1000在一次性Docker建立database，原始EXE及工具來源唯讀。
r1／r2／r3有界匯出分別核對窗口、選單入口與完整caller／writer／consumer；形成史不覆寫。
正式工具重生6064fa97、2919168bytes、1649筆原始指令；26confirmed／6strong／1617unknown。
新增32筆登錄旁註與既有16筆索引自動合併，保留原名、位址、運算元、bytes及推論等級。

共同文字窗為DGROUP3E6E的352×96；3E9C的352×80是HUD，不能互換。
姓名、六職業、性別、能力及Yes/No窗皆由SI consumer導出，不依截圖目測。
原始職業窗count5由1078F覆寫6，class映射1,2,3,4,6,7。
首次詢問選否走551，取消558／成功559後選否走560；首次選否仍static strong。
正常取消0467c01b及單一出生de818064再次完整重建通過，形成有限READY契約。
typed registration、正式狀態機、保留文字、候選交易、Load暫態及正常驗收方法寫入docs/188／docs/84。
CONTEXT唯一現況表、PROJECT_MEMORY及docs/74同步；下一批直接進正式實作，不重開已閉合consumer。

獨立審查收據9e5017de、6832bytes；六種損壞拒絕且原始來源保持。
正式IDA工具面對既存輸出exit2，原sidecar保持；工具AST、輸出UID／GID及容器清理另核對。
工具失敗保留：不存在party.go、非指令邊界10AAA、漏module搜尋路徑、將IRQ事件陣列當數字比較。
修正查詢及審查器後用相同隔離工具鏈乾淨重跑，沒有改原版或production製造通過。
production保持schema0.9.0／content0.1.81及d6c7994e canonical hash；正常問候與順序差異仍待修正。
本批沒有Go／pack修改，不重跑無改動的game全套，前輪r10仍是最近production回歸。
私有來源、PNG、IDA database與完整交付不加入Git；Issue保持OPEN，Goal繼續，沒有新發行包。

## 2026-10-04 Issue #4 正式登錄狀態鏈

依6faf4e7有限READY完成typed registration、問候與詢問、姓名先於六職業、性別、能力預覽／接受、再詢問與最終返回。
接受後才寫名冊，不自動入隊；取消與有效Load清除候選，壞Load保持候選。schema0.10.0／content0.1.82，canonical140f2d39。
正常取消／單一戰士男性出生、後續正常樓下招募與同版本存讀通過；38張完整PNG保留296／608／263／677差異。
能力HP原版11、remake13，長路線亂數條件未對齊，完整V3不升格。首次No551維持static strong；滿額替換fail closed。
完整game430項覆蓋，382頂層／86子PASS、48選用SKIP；正常THE END92.12秒，只屬remake回歸。
全部internal153頂層／281子PASS、4選用SKIP及11套件；desktop main.go ELF64 x86_64建置通過。

保留所有失敗輸出。r1 pack測試漏掛 /assets_raw，r2正常對拍漏設定母親來源，均為驗證設定問題。
r3／r4捕捉到保留文字前景色不在場景索引表時整個選單未畫出；r5按pack前景色建立索引後姓名差296。
r5職業／性別錯誤保留姓名窗；原版10D9C／10DA0與正常packet165／166證實撤窗，r6完整差降為608。
r5完整campaign商人仍用職業先行的舊操作；r7改為正常問候、姓名、六職業映射、能力接受，完整回歸通過。
兩次desktop腳本誤用不存在cmd/dq3，後用main.go完成；image缺file工具，以Python讀ELF header確認架構，未當成產品缺陷。
末次只加嚴loader的字模邊界、word對齊與共用陰影一致性，資料和有效玩家流程保持；重跑internal與受影響登錄／desktop。

CONTEXT唯一狀態表、PROJECT_MEMORY、docs/74、docs/84、docs/188與README穩定摘要同步。
全部工作依Issue #4，已登記實作進度5972241036。原版／圖片／database／封包及使用者scratch均不入Git。
最後收尾收據記錄提交、push、遠端Issue核對及Docker清理；Issue保持OPEN，Goal繼續，沒有新發行包。

## 2026-10-04 Issue #4 首次樓下招募入口

沿120f322正常出生來源延長原版下樓與交談。r2錯把有家具的路線當直線，停於3,5；保留全部來源。
改用已驗證登錄所反向通路及實際CTY通路，r3到首次選單並選擇入隊；後段4000計時旗標觸發拒絕，不宣稱招募完成。
r4在首次三項選單正常停止：234次輸入／468IRQ1、196packet、392PNG／bin；174父產物及出生前172packet保持。
正式來源a85cad67，目的slot1的97bytes等於出生writer。本次狀態1不變性只限單一Warrior male。
新增有限READY、strict recruitment_entry、原始問候527／528與529選單、JSON binding／geometry、Load UI清除與正常InputState。
完整game434覆蓋，385頂層／86子PASS、49選用SKIP；正常THE END66.12秒，只屬remake回歸。
internal155頂層／293子PASS、4選用SKIP及11套件；12種壞契約、八種壞來源拒絕；desktop ELF64 x86_64通過。
取消／出生／首次招募與Load後正式下一步在獨立程序PASS；85完整PNG保持，首次兩圖各295RGB，完整V3未完成。
IDA9.4沿既有image在一次性DB匯出，原始EXE唯讀；五筆confirmed旁註保留原名、位址與bytes，自動合併至匯出。

保留形成史及失敗：初版Python命名空間與Go未使用變數在原版啟動前拒絕；r2路線錯誤；r3後段旗標拒絕。
prototype首輪收據缺共同canvas比較所需的頂層EXE身份，另存r2不覆寫首份後通過。
文件指令反引號曾被shell展開，改以完整檔案輸入修正；Python slim不含git，改用既有Ebitengine image收斂JSON差異。
正式合併三條native案例第三條被SIGKILL；相同image與4GiB下拆成獨立程序重跑，未改seed、原版或期待值。
一次容器自動核准審查逾時，依工具指示重試一次成功；不是憑證或產品阻塞。
所有工作登記Issue #4；首次進度5972746542。提交、push、Issue及Docker清理記錄由本輪收尾收據保存。
沒有新發行包；原版、圖片、database、binary與使用者scratch不入Git。下一切片為入隊後文字與4000旗標分類，不重開本批入口。


## 2026-10-04 Issue #4 入隊後文字與播放等待診斷

接續da720aa。命中原版對拍、spec gate及平台規格路由，載入dosgolem、IDA9.4與硬體停止線。
r5由正常冷新遊戲沿既有路線到入隊；前198個packet與396份PNG／bin獨立審查，前196個逐項保持。
packet197進record530清單；packet198名冊狀態1→2與536內嵌等待，目的97bytes保持。
packet199已送達／消費並進537／538；counter20000時1FEFC寫4000，20000000指令後仍停208F3播放完成等待。
因此4000為本次長等待後的writer結果；尚未證實播放未結束的根因，也沒有完整入隊來源。
既有IDA資料庫有界匯出與公開重生工具確認BP20h進EBG事件曲長資料路線，不沿用VCX PCM解釋或分析driver／ISR。
新增旁註保留原始定位、file／loaded bytes、xref與分級；完整後段仍DRAFT，正式schema／content／hash保持。

正常remake可丟棄副本在197完整RGB差51535、198差56790；選人後缺536文字與等待且仍在舊清單。
r1未設DQ3_ASSETS造成素材SKIP，正確設定後相同binary正常198輸入7.03秒通過診斷，這不是parity通過。
公開IDA首份r1因far call relocation基準混用拒絕，修正原始file bytes及載入比較後r2重生成功。
工具呼叫兩次JavaScript語法錯誤、唯讀Python查詢语法及不存在machine目錄均在查詢前拒絕；修正後用同容器重跑，未改產品。
所有失敗與私有產物保持，來源工具維持正式r4 guard；未清旗標、改clock、重設seed或注入角色。
工作與結果登記Issue #4，進度留言5973523662；現況、下一閘門与证據入口同步CONTEXT、PROJECT_MEMORY及docs/74、docs/188。
沒有production修改，不重跑無改動的完整Go回歸，沒有新發行包。最終稽核、commit／push及容器清理記錄於本輪收據。
下一步核對事件短曲的播放完成介面，取得正常返回來源後才完成READY及正式修正。Goal繼續，Issue保持OPEN。

### 同輪只讀續驗

r6保持相同冷啟動、seed及13项條件，在caller和等待入口只讀原始狀態；前198個packet及396份PNG／bin逐byte保持。
BP20h實際進20959 EBG consumer，等待時286D1、2898=1、2871=0，正常進208F5音樂查詢，排除另一等待分支。
22E10原始bytes讀共享byte5C02並交換FF，22D47..22D50傳狀態位址給FMDRV；後者只有strong靜態設定證據。
原版有界續跑仍未返回，未改正式Go或pack；沒有展開硬體driver／ISR，沒有清旗標或完成byte。
只讀重播、獨立審查及兩份有界IDA匯出皆通過，輸入與私有日志保留。文件patch錨點失敗後核對工作樹，整批未寫入，修正錨點後重試。
索引及雜湊追加docs/188，唯一目前狀態表同步CONTEXT、PROJECT_MEMORY與docs/74。完整入隊仍DRAFT，Goal繼續。

### 2026-10-04 公開FM介面與有界平台驗證

r8核對初始化BX1傳15ED:5C02，入隊暫停、重設、時脈、樂器與播放均AX0返回。前198個packet及396份產物保持。
r9確認1Ch入口56次，API觀察與r8逐行相同；沒有展開硬體driver／ISR。r8原始EBG位址及六種壞證據另行核對。
依Intel8259A重現同級IRQ0在EOI前再次進入，建立可丟棄主PIC副本。三項契約測試通過，完整PIC及snapshot不受支援。
r11前198個來源有52份圖片／bin與原工具不同，另存並明示為平台實驗，不替換原已接受來源。
按公開CMF格式解析81個事件、461delta tick，估計4.801713秒，只用於固定較長觀察窗。
r12首次等待後100M仍未自然返回，PIC有限補足不足以閉合入隊；原版完成介面仍unknown。
r13回到未修改工具鏈，396份產物保持，十次觀察OPL寫入1235及指定硬體讀取計數不前進，未觀察到兩個未知wrapper。

保留失敗與勘誤：r7組裝語法及漏observer、r10fixture未安排首個IRQ、負測試未隔離雙引號preflight路徑，修正後另存或乾淨重跑。
總埠計數曾被誤判為音樂進度，已在對話及證據文件訂正，增加量包含EOI。沒有把局部測試當原版parity或產品修正。
所有來源、腳本、私有SDK參考、收據與分級索引追加docs/188；工作登記Issue #4，進度5974169729、5974288002與5974433783。
正式schema／content／hash及Go保持。下一批從原版record530選人窗口與正常取消閉合有限READY，音樂長等待保持DRAFT。
本輪無production改動，不重跑相同完整Go回歸。提交、push、Issue核對與容器清理由r13收尾收據保存；沒有新發行包。

## 2026-10-04 Issue #4 選人清單與正常取消修正

接續9a13fb5，沿dosgolem正常冷路線在197選人之後改送Esc、Right選No、Enter及告別確認。
原版201個packet／478IRQ1／402PNG及bin完整稽核，前196保持，197等於既有入隊來源。
IDA9.4同image有限窗口與文字consumer匯出825筆指令，11筆分級旁註自動合併，原始名字及位址保留。
有限READY後實作typed selection、追加530、原始動態清單、540繼續詢問、541告別及獨立確認回場。
schema0.12.0／content0.1.84，canonical0bdb4ebf；所有資料與文字在pack，沒有新增Go DQ3 fallback。
201個正式InputState、Save／Load與下一步通過；清單51535降295，取消與告別各295，返回0。
52張runtime PNG保持，四張差分相同；完整V3仍RED，536後文字／音訊仍DRAFT。
14種壞pack與七種損壞來源拒絕，825筆file／loaded bytes及原始定位核對。
完整game以r2前17批及同產品r4後5批合併439頂層，389頂層／88子PASS、50選用SKIP。
internal157頂層／307子PASS、4選用SKIP及11套件；desktop ELF64 x86_64。正常THE END66.58秒，只屬remake回歸。

失敗形成史見docs/188最新驗收節，私有log保持。兩次出生回歸SIGKILL及20秒堆疊定位到測試誤等modal暫停的場景cooldown。
修正測試等待條件，不改正式規則；最終正常舊路線以old-oracles-r3另跑。
原版r1收據缺共用runtime身份，另存r2 d59315c1；r1保持ee4e519a。首份差分缺Pillow改用既有stdlib decoder。
Load元件fixture標題未消費、全文snapshot正規化、IDA null函式假設及負例改錯紀錄均修正後重跑。
正常輸入的來源與完整畫面未換seed或挑結果；硬體driver／ISR及完整音樂等待不在本批。
CONTEXT唯一表、PROJECT_MEMORY、docs/74、docs/84、docs/188及README同步。Issue保持OPEN，Goal繼續，沒有新發行包。
收尾檢查root-owned仍3213、無.md目錄，新增檔與文件UID1000；原版、PNG、database及使用者scratch不入Git。
提交／推送、遠端Issue及最終Docker清理結果由本輪收尾收據另存。

最終old-oracles-r3的四個獨立程序PASS，85張PNG等於前次正式入口版本。
normal-run-r4以最終binary重播201包，52張PNG及收據逐byte等於r3；存讀檔／下一步通過。
出生程序的兩次SIGKILL已由堆疊定位與等待條件修正收斂，沒有當成正式產品缺陷或刪除失敗輸出。
收尾稽核首次把輸出指向唯讀repo掛載；所有檢查已通過但收據寫入拒絕，修正為可寫work後重跑。

## 2026-10-04 Issue #4 共用人物差異與正常Yes續行

上一輪e259412屬正式修正，已推送；本輪沿目前工作樹與遠端Issue #4續驗。
完整差分將295定位為主角182、櫃台NPC106及右下NPC7；文字與框線未出現新差異。
正常冷原版改走Esc後Yes、再Join、Esc、No、告別確認；204packet／484IRQ1／408PNG及bin，cf0730f1。
前198packet與396份圖像保持，名冊、slot1的97bytes、金錢、旗標不變；未注入狀態或改clock。
只讀130組人物取圖，87組NPC加共用phase1；主角43組保留原始BX4→000A，不推定型別。
modal196..203不重新取圖；remake本次walk0。新返回204完整差411，前次No返回0不外推全部相位。
新增正式204包驗收與來源固定hash；Yes→528正常分支confirmed，540內Esc仍strong，完整V3及時序DRAFT。
正常InputState、同版本Save／Load與下一步PASS；四條既有正常路線、七個招募元件與137PNG保持。
十種損壞來源拒絕；來源稽核新增profile及場景不變性，既有201包d59315c1輸出保持。
公開正常Yes producer保存實際執行body；原版／PNG／binary留本機，工具與主題文件可重生入口同步。
正式engine與pack不變，0.12.0／0.1.84及0bdb4ebf；最近完整回歸e259412保持，不重跑相同產品全套。
Issue進度5975028362；結果、commit／push及Docker清理另由本輪收尾收據保存，Issue與Goal保持進行中。

## 2026-10-04 Issue #4 入隊536／537／538有限正常文字

依7f828e6及Issue #4續行，已查現行遠端Issue、工作樹與Docker狀態。
停止重試已否定的總Ticks相位原型，正常冷來源分別停536 inline wait及538後音樂等待入口。
198／199包、472／474IRQ1、396／398完整PNG及bin接受，917a24f6／a6abd32b；正常199保留前198包。
名冊狀態1→2，原角色97bytes、金錢、旗標保持；party只有前29bytes動態觀察，完整副本及插名selector仍需審查。
IDA9.4非破壞範圍與原始bytes核對；sidecar保存base script hash，實際composer另列，不誤稱同一腳本。
正式產品正常198缺536及文字等待，完整差3028；全域bank536隔離原型594，caller底圖保持後198／199各295。
正常Save／Load、Load後下一步僅屬remake診斷；原版199仍音樂等待，不宣稱完整入隊、音訊或V3。
八種損壞來源拒絕，前後正對照PASS；三份最終runtime各48張舊PNG保持。
失敗保留：r1 tarfile filter不支援、r3使用目前城鎮文字bank、r5多層字串逃脫、r7不唯一替換。
r4的594定位為新同伴提前覆蓋主角；修正生成與底圖原型後r6／r8乾淨重跑，不調seed、clock或人物phase。
公開198／199來源工具與docs/188索引保存。正式引擎／pack、schema及最近完整回歸保持，沒有新發行包。
本輪Issue進度5975257608；結果、commit／push及容器清理由同前綴final收據追加。
下一步補插名／party動態證據與音訊完成閘門，再將整段文字及底圖列READY後實作。


## 2026-10-04 Issue #4 入隊副本、不同名字與單次短曲

依ab2f8a0與遠端Issue續行，199包／474IRQ1／398完整PNG及bin保持，新來源d0f6428d接受。
24筆只讀觀察閉合97bytes copy與名冊index覆寫。536／538取新角色；537本路線取主角，推翻同名原型的錯誤綁定。
新隔離原型正常199包、存讀檔、不同名元件PASS，50張runtime PNG保持，仍差295。正式產品與pack未修改。
新條件八種損壞拒絕及有效對照通過；r1 producer hook引用錯誤、r2 checker同名假設被拒絕，保留失敗。
IDA9.4最小sidecar及有限reader核對。前幾次缺輸出，原生log定位ASCII讀中文錯誤；指定UTF-8後乾淨重跑。
10CEE讀[SI-2]選名字，1EBD8只讀該值作圖片cache索引；沒有猜成健康值或稱初始化writer已閉合。
cue32原始81事件／461ticks；舊轉檔79／660且OGG11.054281秒。新公開單次MIDI逐事件相同，SMF時長4.801708771秒。
Munt舊image缺失，修原Dockerfile固定原廠commit重建dq3-munt:2.8.2-r1；library2.8.3及smf2wav1.9.3實測。
APT建置完成；獨立網路診斷HTTP／HTTPS200，另一次APT探針下載完成但外層60秒超時，不列產品或建置失敗。
Munt起始靜音設定另存r2，211761frames／OGG4.801837秒，完整解碼與非靜音PASS。音色與時長屬近似，原版播放後返回仍未知。
公開來源、MIDI工具與Dockerfile可重生，原版、ROM、音源及database未入Git；入口見docs/188。
本輪Issue進度5975544582，結果、commit／push及容器清理由final收據追加；Goal與Issue保持進行中。
下一步收斂有限READY並接typed pack與正式引擎，不再追硬體driver／ISR。正式產品未變，不重跑同一全套回歸。


## 2026-10-04 正式入隊文字與單次音訊

接續1d7d899與Issue #4。有限READY將536兩處新角色、537主角及538新角色名字接入typed pack，保留caller底圖與內文等待。
原版caller10459→10469返回、10398→103AE追加540；此段後續仍無動態來源，依靜態控制流與平台規格近似實作。
schema0.13.0／content0.1.85，canonical19f6124c。81事件／461ticks在FM與Roland後端單次播放，等待289更新，不可按鍵縮短。
播放後續播原場景並進540 Yes／No。正常角色交易只一次，播放期間有效／拒絕Load、存讀檔及下一步通過。
15種壞pack契約拒絕，有界事件parser壞資料拒絕；EXE／TXT／EBG原始資料、FM單次樣本時長及非靜音核對。
正式正常199、No201、Yes204通過；入隊50、No52、Yes55PNG保持，正式198／199差3028降295。
game443覆蓋，395頂層／91子PASS、48選用SKIP；internal161頂層／11套件PASS、4選用SKIP；desktop Linux x86_64 ELF。
正常THE END107.13秒只屬remake回歸，原版全流程與音畫V3未完成。
型別比較與原始位址測試筆誤修正；素材相對掛載及第三個連續重型測試的記憶體限制改為每項隔離程序乾淨重跑。
九份JSON保留既有排版，前後語意及canonical相同。r5全套、r6最終嵌入排版與r7實際名字ops斷言的受影響測試由各自收據保存。
Issue進度5976023866；提交、推送與遠端結果由final收據追加。本機音源與使用者scratch保持，不建立新發行包。
現況見CONTEXT，原始定位與重生入口見docs/188，欄位及私有音源staging見docs/84；後續不重開已完成文字或深入driver／ISR。

## 2026-10-04 正常招募人物差異的原始影格核對

接續78d84b1與Issue #4留言5976184810。新增唯讀工具`tools/verify_dq3_recruitment_sprite_raster.py`，入口掛入docs/188。
核對已接受cf0730f1來源的204包、408份PNG／bin及log身份，沿正式r4-focused的11張正常完整畫面診斷。
原始BLS兩影格連同CTY／BLK底圖逐點吻合；主角182、櫃台106、右下123合計411，開窗後右下露出7，合計295。
全部完整畫布差異都可由原始影格解釋，這組樣本的色盤、遮罩及底圖吻合。正式PNG差異保持，沒有替代圖片或遮罩驗收。
六種損壞副本拒絕，前後正對照通過；輸出拒絕覆寫，工具hash、語法與UID/GID核對。
初版工具把normal_inputs242誤當204包而拒絕，已依原始欄位訂正；原版、輸入與素材保持。
本機收據`work/issue4-sprite-raster-r2-receipt.json`、`-r2-negatives-receipt.json`及`-r3-final-audit.json`。
收尾首次把root-owned的.md檔列作.md目錄，改分開find查詢；既有root-owned3213保持，.md目錄為0。
Docker檢查需主機權限，依既有隔離規則續做；一次性容器已清理，其他專案容器保持。
本批正式Go、pack與存檔格式保持，最近全套回歸仍78d84b1；動畫時序DRAFT與完整V3未通過。
依已授權commit + push保存工具與證據，提交及遠端結果追加Issue；原版素材、PNG、binary與使用者13項scratch不加入Git。

## 2026-10-04 正常觀看名單清單與取消

- 依Issue #4正常路線續驗。IDA9.4與dosgolem證實第三項列未入隊名冊；原版199包043b39b1及203包cf23fcf9接受，輸入／IRQ1／完整PNG與bin、97bytes、指標與旗標核對。
- evidence先DRAFT；可丟棄原型清單51477→295，取消來源接受後有限READY，正式沿typed清單接入方向鍵與Esc→540→No→541→返回。不新增pack欄位、raw設定或存檔欄位；schema0.13.0、content0.1.85、canonical19f6124c保持。
- 正式203包、snapshot與RNG不變、Save／Load及下一步PASS。七張全畫布差異逐點等於既有295／411人物差異；舊157張及前綴47張PNG保持。54張新PNG及原版素材留本機，不加入Git。
- 六種壞來源全部拒絕，前後正對照通過。詳細狀況205包只保留DRAFT，兩個讀鍵等待待閉合；確認仍舊行為，不宣稱完整View或V3。
- 全game445檢查，398頂層／93子PASS、47選用SKIP；internal161頂層、11套件PASS、4選用SKIP。正常結局79.95秒；desktop Linux x86_64 ELF PASS，無素材缺失SKIP。
- 診斷腳本曾誤用sourceCanvasDifference參數及把目錄當收據；訂正後同工具鏈重播。desktop曾誤指定不存在的cmd/dq3，根套件又納入使用者tmp_dump.go；核對main.go後以明確入口建置，不改使用者資料。這些是驗證問題。
- 路由入口與新公開producer／validator均掛入docs/188，欄位沿用說明在docs/84。私有工具、檢查、負例與核對收據使用issue4-view前綴，入口詳見docs/188。容器内沒有rg，改用grep檢查production新增值，沒有新增DQ3設定fallback。
- Docker一次性容器全清理，root-owned基線3213、零.md目錄保持。保護13項資料；commit／push結果保存`work/issue4-view-r1-post-push.json`。下一步詳細狀況，維持driver／ISR停止線，沒有新發行包。

## 2026-10-04 詳細狀況兩次等待與正常返回

- Issue #4來源205包／486IRQ1／410完整PNG及bin接受，e61060c7；唯讀控制流重播保持全部產物。200與201完全同圖，訂正先前黑色撤窗誤判，停止無根據的像素／硬體調查。
- 先正常隔離原型，再docs/188有限READY，正式確認分支接只讀能力窗、兩次新按鍵與Esc→540→No→541→返回。無咒文、無異常及pack初裝是本輪邊界，其他角色與未映射鍵不猜補。
- 正常205包、持久snapshot、另行RNG比較、Save／Load及下一步PASS；正常200／201完整差43708／40204降145。原始相同97bytes角色元件完整差7，138為正常HP11／13數字輸入差異；未改角色湊圖。
- 八種壞來源全部拒絕。五路線258張舊PNG及新路線前綴50張保持，全畫布逐點核對，完整V3仍RED。
- 完整game449項合併r7已過326及r8餘123，402頂層／101子PASS、47選用SKIP；internal162頂層、11套件PASS、4選用SKIP；正常THE END121.16秒與desktop Linux x86_64 PASS，無素材缺失SKIP。
- r7舊「View確認回主選單」測試失敗，與原版衝突，r8改驗所選第二個名冊角色實際欄位。正式產品保持，已通過測試不重跑。型別／欄位筆誤、標題fixture繪圖器與Load正規化、收尾漏manifest均分類為驗證問題。
- 九份JSON與乾淨9f81ddb archive相同，schema0.13.0／content0.1.85／canonical19f6124c保持。來源、驗收、負例與本機工具入口全部掛docs/188；typed能力引用見docs/84。
- 收尾核對 `a810fd7ead661b6aa8c6ee18e285793919345e2274dc916d34c924b0c3ce3d45`，全畫布 `208f0f0bf88c3e5aecfee546e4fdf9b243ec05246c156becd761610fe684cab5`。根擁有權基線3213、零.md目錄與UID1000保持。提交推送與容器清理結果保存`work/issue4-view-detail-r1-post-push.json`，並追加Issue #4。
- 原版素材、圖片、binary、database、使用者13項scratch未加入Git，沒有新發行包。下一步K改名與已學咒文頁、人物動畫及後續原版流程，Goal保持進行中。

## 2026-10-04 觀看詳細狀況的單人隊伍K改名

- Issue #4正常K入口202包、同名217包、取消212包、異名219包已接受。重新依隊伍選主角507F，能力副本與未入隊名冊不變。全部來源沿dosgolem2f44a68、一次seed1357與正常IRQ1，不注入角色。
- IDA9.4原始bytes及正常writer／consumer閉合後，docs/188有限READY接typed pack與正式InputState；第一等待K不跨讀鍵，第二等待開空姓名窗。空名拒絕、取消不寫入，成功只修改主角姓名並同步對話；接540／No／541正常返回。
- 三路正常重播、持久snapshot、另行RNG、UI欄位、存讀檔與下一步通過。姓名窗完整RGB11926降76，69為既有HP數字、7為既有NPC影格；差分位置與雙側RGB落在既有145差分。完整V3仍RED，沒有改角色、裁切、遮罩或指定phase。
- 六路線315張PNG與三新路線前綴156張保持，新正常201張留本機；取消及異名本次返回零差異，同名返回411不變。公開producer完整Go除前綴外等於已執行來源，公開validator重新接受；八種壞來源全拒絕。
- schema0.14.0／content0.1.86／canonical e274124c；九份JSON由乾淨28690cc與公開遷移器重建且逐byte相同。首版排版擴展已修正，typed資料保持；末版原始EXE parity、壞契約及desktop再次通過。舊schema存檔仍拒絕，不自動遷移。
- 完整game454覆蓋，407頂層／107子PASS、47選用SKIP；internal164頂層、11套件PASS、4選用SKIP；正常THE END156.69秒與desktop Linux x86_64 PASS，沒有素材缺失SKIP。同binary的X11環境失敗以Xvfb -noreset重跑通過；三路改獨立程序，沒有OOM事件不推定原因。
- IDA索引保留11筆追加6筆confirmed，r3匯出321指令與原始bytes／relocation全符合。來源、公開工具、負例、全畫布與本機重生入口掛docs/188，欄位掛docs/84；唯一狀態表更新CONTEXT。
- root-owned3213及零.md目錄保持；最後檢查首版把既有root-owned.md檔誤當目錄，修正判斷後乾淨重跑。13項使用者資料、原版、PNG、binary、database及私有work不加入Git，Docker容器收尾清理。沒有新發行包，完整View、多角色改名、咒文與原版campaign未完成；Issue與Goal保持進行中。
## 2026-10-04 登錄與觀看的咒文詳細頁

- 依 Issue #4，正常第三職業男性203／209包原版來源接受。冷啟動dosgolem2f44a68、seed1357固定一次；209包494 IRQ1與418完整PNG／bin，前165及前203包保持。第172包才寫名冊；兩種已知咒文指向同一index40，union去重後顯示record161。來源4dcb99c8，八種壞來源拒絕且前後正對照保持。
- 第一個正式玩家差異在登錄170：原版咒文等待，前版直接顯示接受選單。IDA9.4原始166指令bytes與xref、四類union／動態窗口／文字consumer／獨立等待與還原，加正常209包閉合後，docs/188審成有限READY。原型誤用酒館D3TXT01的record已查明並訂正；正式版只用pack的D3TXT00 glyph引用。
- typed `character_spells`、共享renderer及正式登錄／觀看等待已接入。完整60項catalog與63份record逐項核對；去重排序不改LearnedSpells，未知record拒絕，pack副本只在renderer改動態高度。12種壞契約、held key、新按鍵／點擊、三次觀看等待、有效／拒絕Load及狀態保持通過。
- 正式209包逐包位置／旗標、候選人在接受前不寫名冊、觀看持久snapshot／RNG、Save／Load及正常下一步通過。登錄170完整RGB差16485降548，169／170差分位置與雙側RGB相同；觀看203／204／205各430且完整差分相同。新增咒文層無新差異，能力數值與動畫保持各自結果，完整V3仍RED。正常45張PNG與原型逐byte相同，九條舊路線516張保持。
- schema0.15.0／content0.1.87／canonical `2d712e65f18ce7919ddccf970160d68d00b63a54a73088bcd473c38a9fa61d10`。九份JSON由乾淨f823b61及公開遷移器重建逐byte一致，保留排版；三份框架可讀文字沿已有frame glyph語意修正，原始glyph_codes保持。舊schema存檔仍拒絕，不自動遷移。
- game459完整覆蓋，412頂層／107子PASS、47選用SKIP；internal167頂層／341子PASS、11套件及4選用SKIP；正常THE END97.88秒與desktop Linux x86_64通過，沒有素材缺失SKIP。r1的唯一失敗是campaign仍只送兩次登錄確認，僧侶新咒文頁未關；查明後只改正式InputState測試操作，r2重跑該路線。其餘458項沿用同一production的r1證據，不重跑已通過項。
- 首筆原始bytes測試把far call誤寫成near call，改依IDA的原始file bytes與MZ relocation。直接UI元件fixture缺出生能力與已訪城鎮，Load啟動舊存檔補遷移；逐欄診斷後用正式出生／起點紀錄建立有效fixture，正常209包不注入。一次Docker掛載遺漏、PNG診斷缺PIL、tar祖先目錄檢查、收尾誤要求成功log含THE END均分類為驗證工具問題，修正後乾淨重跑；不寫成產品缺陷。
- IDA索引保留17筆並追加4筆有限confirmed；r3自動合併匯出166指令與原始bytes／relocation相符。公開producer完整Go除前綴外等於已執行來源；公開validator嚴格審查窗口／文字／等待／還原與返回。工具入口掛docs/188，typed欄位與遷移掛docs/84，唯一狀態表更新CONTEXT。
- 關鍵本機收據：來源`4dcb99c8`；來源負例`a309ff9a`；完整畫面`4e568b09`；舊路線與新45張`46a4e158`；IDA匯出`5273a602`；完整回歸`work/issue4-view-spells-full-r2/game-receipt.json` SHA-256 `c49013df8954a2e36f66613cf866a072d019ce0a7273702a5ef8385349e53d88`；摘要`work/issue4-view-spells-final-summary-r1.json`。
- 原版素材、圖片、binary、database與私有work維持本機；使用者13項資料保留。提交／推送、root擁有權基線與Docker清理在本輪收尾收據及Issue追加。沒有新發行包，其他裝備、多角色改名、其他角色咒文動態、人物動畫及原版完整流程仍待驗；Issue與Goal保持進行中。

## 2026-10-04 詳細頁一般字母鍵關頁

- 依Issue #4從041caf9接手。先追其他裝備：IDA9.4的140＋350條目核對物品欄順序、bit8000、文字consumer與攻擊欄位。正常換裝後返回名冊仍缺原版來源，維持DRAFT與硬體driver／ISR停止線。
- 同一入口發現第二次等待忽略一般字母鍵。正常A鍵209包原版來源c0a7bc08接受，前205包保持；元件與正常路線均先取得206卡住紅測試，再經docs/188有限READY接正式AnyKeyEdge。舊「未知鍵保持」assertion追加勘誤，保留既有Esc／K證據。
- 正式A／Esc／三路K、37受影響頂層／21子測試、零SKIP；正常THE END112.88秒及desktop通過。九份JSON保持；八種壞來源拒絕，舊209來源重新接受。原始定位索引保留21筆，追加非K一筆confirmed並由IDA自動附註。
- 舊561張與新45張PNG保持；原版A／Esc全209PNG/bin相同，203..209全畫布差依序430、430、430、7、7、7、411，完整V3仍RED。收據與公開重生入口見docs/188；本機收尾c103a2f8。
- 圖片稽核首次命令引用錯誤、收尾稽核首次收據誤寫唯讀/repo，均分類為腳本／掛載環境問題。修正後在同一image、同一輸入乾淨重跑；正式產品未因這些失敗改動。Docker image inspect的sandbox socket拒絕亦屬環境，實際工作沿既有image非root容器執行。
- 原版DQ3.EXE保持5178fdc8，root-owned基線3213、零.md目錄；使用者13項未追蹤資料保留，未納入提交。沒有新發行包，所有原版素材與私有產物留本機。提交、推送及容器清理的最終核對以Issue #4本批留言與本機post-push收據為準。

## 2026-10-04 空名冊觀看

- 從af7d148依Issue #4續行，原版dosgolem仍2f44a68。入隊播放後等待沒有新執行器能力，維持硬體driver／ISR停止線。本輪選已列工作中的空名冊View，不重開已完成開場或改音樂時計。
- 正常新遊戲、姓名取消、選No、下樓觀看共195包／466 IRQ1。前164包及完整PNG/bin保持；191原生零名冊分支選D3TXT00 record316，192內文確認後540、No→541獨立等待→195field。來源ff7a8abd，冷啟動前seed1357固定一次，無狀態／名冊注入。
- 正式InputState修正前第一新blocker191，正常紅測試明確重現。docs/188有限READY後接具名EOF延續與pack文字引用；空名冊不開空白清單，不改RNG或持久角色。正常195包、存讀檔及下一步通過，非空View與A／Esc／三路K保持。
- schema0.16.0／content0.1.88，canonical096fe3a7。九份JSON由乾淨af7d148與公開遷移器重建逐byte一致，保留排版；原始EXE／DAT parity與schema/reference負例通過。舊schema存檔仍拒絕，沒有自動遷移或新發行包。
- 完整game463覆蓋、416頂層／107子PASS、47選用SKIP；internal167頂層／344子PASS、11套件通過、4選用SKIP，沒有素材缺失SKIP。正常THE END130.07秒與desktop Linux x86_64通過。收據work/issue4-empty-view-full-r1/game-receipt.json為0d46c863。
- 九條舊路線606張完整PNG保持，新46張留本機；正常取消150..164保持。187..194完整RGB各295、195為411，沒有遮罩、裁切或指定相位；完整V3仍RED。八種壞來源拒絕且正對照前後一致；收據467c4766。圖片22b3fbca，乾淨pack重建d3089b8f。
- IDA9.4以原始115282bytes／5178fdc8輸入，保留22筆原始定位語意，追加1062F／10696兩筆有限confirmed；r2自動合併匯出121條目，bytes及MZ relocation全符合，收據21c86993。計數高byte、其他空清單與完整流程不外推。
- 環境／腳本失敗：容器socket需host權限；首次文件讀取誤用非現行檔名與.go目錄；checker將187誤標waiting，實際inline_wait；r1收據缺共用圖像checker的頂層原版身份。按相同原版來源修正，保留r1、r2收據及乾淨紅測試，不作產品缺陷。
- 圖片稽核先漏改名子目錄，再誤用創角前綴代替取消前綴；查明157分岔後，以實際正常取消150..164重跑失敗部分，606張既有路線檢查保持。沒有為使圖片通過調整正式程式、輸入或相位。
- 唯一現況表更新CONTEXT，現行計畫更新docs/74；新公開producer/checker/migration由docs/188及docs/84索引。所有原版素材、PNG、database與私有work留本機；使用者13項資料保留。提交、推送、擁有權及Docker清理以本輪收尾收據與Issue最終留言為準。

## 2026-10-04 空加入與單人分離

- 從1217cbd依Issue #4續行，既有空View已閉合。窄查IDA9.4 caller、名冊計數及consumer298條目；seed1357在冷啟動前固定一次，正常姓名取消、選No、下樓後選對應動作，沒有狀態或名冊注入。
- 原版兩路各193包／462IRQ1，前188與空View逐項相同。Join530後零名冊選316內文等待；Leave單人直接542主角姓名、EOF同包540。No→541獨立讀鍵→field2,18；名冊、隊伍指標、主角、金錢及旗標保持。來源47d1a115與df05d0d4。
- 修正前正常Join189／Leave190及元件RED；docs/188有限READY後實作兩個必填typed文字引用與具名EOF延續，保留原始record界線及控制碼。正常193包、snapshot／RNG、同版本存讀檔及下一步通過；schema0.17.0／content0.1.89，canonical4b235d63，舊schema存檔仍拒絕。
- 完整game466覆蓋、419頂層／109子PASS、47選用SKIP；internal167頂層／350子PASS、11套件、4選用SKIP，無素材缺失SKIP。正常THE END253.74秒、desktop PASS。十條舊路652張PNG逐byte保持，新兩路各44張；187..192全RGB各295、193返回零差異，完整V3仍RED。
- 各八種壞來源拒絕且正對照保持；乾淨1217cbd九份JSON由公開遷移器逐byte重建。保留24筆原始定位並追加1046A／105BC有限confirmed，IDA自動合併26筆及原始bytes／MZ relocation核對。完整收據與hash見docs/188最新CONFORMED。
- 環境失敗：首次命令換行escaping在執行寫入前SyntaxError，改檔案化Docker腳本乾淨重跑。首輪三個大型來源測試共用3GiB發生5次OOM；memory.events、exit−9、日誌留存，使用同一binary逐一重跑受影響五項全PASS，重跑容器oom_kill0。不將這些記為產品缺陷，不改程式為了驗證通過。
- README穩定摘要補正過期schema／content，唯一現況表、docs/74、PROJECT_MEMORY與WORKLIST更新。新公開producer、checker、遷移器均有docs/188或docs/84入口；無新文件或交付目錄。下一步空清單Yes正常續行，非空分離、滿隊與音畫仍待來源，維持driver／ISR停止線。
- 使用者13項scratch及本機資產保留。提交、推送、擁有權與Docker清理以本輪收尾收據及Issue最終留言為準；Goal與Issue不宣稱完成。
- 收尾腳本首次將收據寫到唯讀`/repo/work`而失敗；檢查項目已跑完，未改產品檔案。輸出改為明確可寫掛載`/work`後，以同一容器工具鏈乾淨重跑收尾檢查。

## 2026-10-04 空清單Yes正常續行

- 從c572ab3依Issue #4續行。兩條正常冷啟動來源各196包／468 IRQ1，前190包及PNG/bin等於各自空No；seed1357在執行前固定一次，無狀態或名冊注入。Joinf56419c0、Leaved8c6c47d。
- 191原生103C0→10378重播528及10384選單，游標歸首；再選同動作→194 No→195告別→196正常行走。名冊、隊伍指標、主角、金錢及旗標保持。有限READY後正式InputState、snapshot／RNG、同版本存讀檔與下一步通過；既有狀態機吻合，沒有產品規則或資料變更。
- 34受影響頂層／25子PASS、零SKIP。十一條重生舊路695張PNG保持，新兩路各47張；187..195完整RGB各295，196 Join411、Leave0。完整V3仍RED。九份JSON與c572ab3逐byte相同，schema0.17.0／content0.1.89、canonical4b235d63保持；沒有新發行包或存檔遷移。
- 兩條各八種壞來源拒絕，前後正對照相同；原始定位保留26筆並追加103C0有限confirmed，IDA9.4自動合併27筆、298條目bytes及MZ relocation核對，收據7c2a987c。完整來源、驗收、負例與圖片收據見docs/188本節CONFORMED。
- 測試判準訂正：首版新增元件誤把文字保留等同物件指標相等；查明appendRetainedRecord新建狀態並沿用ops，改核對操作前綴，保留紅測試。三條既有空No保持；不為此改production。
- 環境失敗：觀察尚空日誌誤取末行引發IndexError，後加空值guard；程式搜尋遇名稱以.go結尾的目錄，已有輸出的直接函式足以定位真因。改名三條同程序累積記憶體OOM1，改既有分程序方式，同一binary全PASS；A鍵漏設共用來源環境變數SKIP，補齊後實際PASS，重跑oom_kill0。沒有重新啟動活躍來源或調正式參數。
- 本輪按比例只重跑受影響招募驗證，最近完整game466、全部11個internal、正常THE END253.74秒與desktop為c572ab3，不冒稱本輪全套重跑。唯一現況表、docs/74、PROJECT_MEMORY及WORKLIST更新；新producer／checker在docs/188索引。
- 下一個具體玩家路徑為招募主選單Esc，原始取消flag分支與告別已定位但正常動態仍待驗；其他原版流程、音畫與完整V3未知，driver／ISR停止線維持。使用者13項資料、原版素材與私有work保留；提交、推送及Docker清理以收尾收據和Issue最終留言為準。

## 2026-10-04 招募主選單Esc告別等待

2026-10-04 招募主選單Esc已修正：正常193包／462 IRQ1來源137411c7證實先顯示541告別、獨立等待，再以新按鍵返回。修正前192包RED，有限READY後正式InputState、snapshot／RNG、同版本存讀檔及下一步通過。
本輪41招募頂層／28子PASS、零SKIP，正常新遊戲至THE END95.45秒與desktop PASS；42個不同頂層含主線。789張既有PNG逐byte保持，新Esc44張；187..192全RGB各295，193返回零差異，完整V3仍RED。
主線舊測試將主選單Esc視為立即關閉，已補上正常告別文字及獨立確認輸入；保留首輪失敗，正式產品只修改rcMenu Cancel。九份JSON保持，schema0.17.0／content0.1.89、canonical4b235d63。最近完整game466與11個internal仍為c572ab3，未冒稱本輪全套重跑。
唯一現況表在CONTEXT，來源、READY／CONFORMED與公開工具在docs/188，Issue #4進行中。下一步原版正常F5／F6存讀檔：先建立明確可寫overlay，不改原始素材，再驗證存檔、移動、讀檔及恢復。非空分離、入隊播放後返回、音畫與完整原版campaign仍未知；driver／ISR停止線維持。沒有新發行包。

- 原版正常來源137411c7、193包／462 IRQ1，192 Esc後541獨立等待、193確認返回。修正前正常192與三個游標元件RED；先寫READY，正式Go只改rcMenu Cancel引用既有typed告別。新公開producer／checker由docs/188索引，原版收據與PNG留本機。
- 41招募／28子PASS、零SKIP；主線舊導航假設失敗保留，補正式告別確認後THE END95.45秒、desktop通過。八種壞來源拒絕，789張舊PNG與九份JSON保持。31筆IDA分級、478條目原始bytes／MZ relocation核對。完整hash及環境失敗分類見docs/188，不把有限綠色測試升格原版全流程完成。
- 使用者13項資料保留；容器均已清除，無新映像或發行包。提交、推送與擁有權核對以本輪收尾收據及Issue最終留言為準。下一步正常F5／F6原版存讀檔，維持音訊硬體停止線。

## 2026-10-04 正常 F5／F6 存讀檔與重畫

2026-10-04 正常 F5／F6 存讀檔已接入正式玩家入口。F5 經驗提示、確認、十槽選擇、VOC 完成等待及告別返回，F6 選槽、讀回及上層時鐘重設已通過有限對拍。空名冊讀回殘留現有角色的問題已修正。
dosgolem 四份正常來源接受，固定 seed1357 一次，193 包前綴保持；存檔→左移→F6 共200包／476 IRQ1，原生讀回完整2172bytes。兩側 RNG 保持，沒有模擬器 restore 或遊戲狀態注入。
F6 完整畫面差45109已降0。F5 的194、196、197及F6的200均完整RGB零差異；選槽195差4064、行走198差356、選槽199差32847。原版既有十槽資料與remake初始空JSON不等價，選槽畫面不宣稱V3。
現行schema0.18.0／content0.1.90，canonical `sha256:9d6325addef6db4d6f4c049f44fe517e61d2a0550724c67d5ae7ec0f649a3917`。九份JSON從乾淨1bef3d7重建逐byte相同。舊schema或不同hash存檔拒絕，不自動遷移。
完整game478頂層覆蓋：431個不同頂層及117子PASS、47選用SKIP；internal169頂層／363子、11套件PASS、4選用SKIP，無素材缺失跳過。正常新遊戲至THE END201.00秒、desktop Linux x86_64通過，oom_kill0。
十四條舊路線833張完整PNG逐byte保持。有限狀態E2／流程E3、指定畫面V3；全流程V3與完整原版campaign仍未知。下一步原版F6返回後正常行走及可比初始槽資料；下層F6、複數隊伍／其他等級、人物動畫、入隊播放後返回及非空分離保留未知。沒有新發行包。

- 四份原版來源、八種壞來源全拒絕及正對照保持，12筆IDA分級語意／493原始條目核對。正式垂直鏈為EXE／TXT→typed pack→正式F5/F6→UI／VOC完成→獨立JSON槽→Load／下一步。來源、腳本、hash與有限驗收入口見docs/188。
- 先前196終點誤設field、JSON nil／empty比較及不合法overworld測試metadata均保留並分類為驗證腳本問題；F6 bank與空名冊殘留是本輪實際產品修正。OOM及缺/assets_raw掛載分別重跑，未為環境失敗猜改遊戲規則。
- 使用者13項資料保留；提交、推送、root基線及Docker清理以收尾收據與Issue最終留言為準。

## 2026-10-04 F6後正常行走與驗證契約

2026-10-04 原版F6後正常行走來源已延長到202包／480IRQ1，接受收據71a52768。200包前綴、完整PNG／bin、native FileOps及原生存檔保持；201左移只改2172bytes中的座標低byte，202右移回復保存區，時鐘均0。
正式InputState持續按住方向直到整步完成，再放開；逐包完整snapshot、RNG、時鐘、engine同版本存讀檔及下一步通過。輸入完成點可比，按鍵時長與CPU／TPS映射未精確對拍，不能稱為完整時序parity。
本輪8個不同頂層／4子PASS、10命令、零SKIP／OOM；147張既有F5/F6 PNG逐byte保持，新路線200前保持。200完整RGB0，201差356、202差122，完整V3仍未通過。
正式Go與九份JSON和e939db2保持，schema0.18.0／content0.1.90、canonical9d6325ad不變。最近完整game478與11個internal、THE END201.00秒及desktop仍為e939db2；本輪沒有冒稱全套重跑，也沒有新發行包。
下一步可比初始十槽資料的選槽對拍，以及201／202人物畫面差異的consumer證據。下層／室內F6、複數隊伍／其他等級、入隊後返回、非空分離、音畫及完整原版campaign仍未知。

- 初輪收據身分欄位缺失與一幀輸入未完成行走均保留，修正checker欄位及正常按住／放開重播，沒有改正式cooldown、座標或動畫。r2來源八個負例拒絕，正對照前後一致。完整hash、PNG與差異範圍見docs/188。
- 使用者13項資料保持；提交、推送與Docker清理以收尾收據及Issue結果為準。

## 2026-10-04 F6後人物影格consumer與完整像素來源

2026-10-04 F6後兩步的人物影格consumer已由正常202包／480IRQ1來源363f8f70核對。原版新舊全部事件、PNG／bin、2172bytes持久區及原生存檔逐byte保持；94次只讀觀測，不改時鐘或相位。
原始BLS／CTY／BLK逐點解釋201的356像素：英雄127、NPC14為106、NPC15為123；202的122全為英雄步伐影格。兩張畫面沒有其他差異，完整RGB仍356／122，動畫時鐘對應未知，完整V3未通過。
本輪10命令、8不同頂層／4子PASS、零SKIP／OOM；200張既有完整PNG逐byte保持。正式Go與九份JSON仍等於e939db2，schema0.18.0／content0.1.90、canonical9d6325ad保持。最近完整game／internal／THE END／desktop仍為e939db2。
新producer、來源checker、完整畫布診斷及兩筆原始位址ledger均由docs/188索引，IDA9.4自動附註723條目。未改產品動畫或打包。
下一個正常切片是可比初始十槽metadata的F5/F6選槽UI；人物動畫的可比時鐘仍DRAFT。其他F6場景、複數隊伍、入隊後返回、非空分離、音畫及完整原版campaign仍未知。

- 原版來源363f8f70、完整像素9c8a9a38、IDA兩筆680aaf8f；受影響測試及200張舊PNG保持，收據入口見docs/188。正常consumer與靜態bytes有分級，不將畫面線索升格動畫時鐘規格。
- IDA間接方向表入口11BA2沒有自動函式邊界，首輪停止並保留error sidecar；後續只匯出有界raw bytes並維持unknown，不猜補函式。圖片保存抽查初用147舊計數，本輪包含新增兩步共200張，按實際清單修正；沒有產品失敗。
- 使用者13項資料、原版素材與private work保持；提交、推送、負例及容器清理以最終收據／Issue結果為準。

## 2026-10-05 同伴裝備出售持有權修正與八格A選擇

2026-10-05 同伴出售裝備清除副本的錯誤已修正。正常新遊戲、出售取消／確認、單次售款、防禦更新、存讀檔及下一步通過。schema0.19.0／content0.1.91與canonical66224bc0保持，沒有新包。

完整game485頂層覆蓋、489次執行，434不同頂層／121子PASS、51選用SKIP；internal171頂層／375子、11套件PASS及4選用SKIP。正常THE END163.93秒、Linux desktop PASS；逐項重跑零OOM，1082張既有PNG及5張新出售PNG保持。

使用者已選A單一有序物品格，不再等待架構選擇。完整word試作5項PASS，保留空格、穿戴／詛咒及未知高位元；尚未READY或接入production，存檔尚未升級。206..208畫面仍RED，商店確認／取消PNG空窗未驗。唯一現況表CONTEXT、出售docs/182、八格docs/188；Issue／Goal進行中。

- 原版契約沿docs/182：IDA linear1789A..1789E所選word清00FF後才加錢；正式只改既有裝備setter，沒有新raw ID／文字／pack fallback。主角／同伴5頂層／2子targeted PASS，正常NPC出售、取消不交易、單次售款135、防禦85→78、Save／Load及下一步保持。
- 正常RED r5 fef3927b、最終GREEN r5 32187900、完整r2 1536a816；元件／完整原版bytes及有限E2／E3集中docs/182。確認／取消PNG空窗未驗，不宣稱商店動態V3。
- 首輪完整並行OOM分類為環境失敗，逐項同資源重跑通過。驗證旗標集合排序、Load標題狀態、撞櫃台cooldown及截圖啟用時機均保留勘誤，不當產品缺陷。
- 語法樹稽核69份輸入／207候選，不代表完整型別或別名證明。使用者A決定已登記Issue #4；私有完整word試作5項PASS，原版詛咒4000及未知高bits保留，尚未production接線。來源與下一READY gate見docs/188。
- 使用者13項資料保留，原版EXE hash5178fdc8保持；root基線3213、零.md目錄，最終提交／推送／Issue與容器清理以收尾收據為準。

## 2026-10-05 A 共用物理格核心

依使用者選定的 A 實作 `internal/itemstore`。唯一集合保存有序完整 words；背包／裝備檢視附物理位置且是複本。寫入採複製後提交，避免 Store 值複製後透過共用 slice 改到另一持有者。取得寫第一空格、移除留下空格、自給旋轉包含空格的全部後續格；跨人交易先驗雙方與旗標 gate，失敗不寫。

- 有限READY、原始位址、邊界、工具及執行入口集中docs/188，JSON入口docs/84。容量／mask／metadata完全外供，沒有DQ3 production fallback。換裝只改格內flags；舊詛咒裝備阻止替換，教會primitive移除所有詛咒words，玩家資格與費用留在consumer。
- 核心快照版本1只保存words。缺契約、越界、null、舊欄位、重複key、大小寫變體與trailing JSON拒絕。一般Marshal不會靜默丟失private fields，普通Unmarshal不能繞過archive／pack契約。
- 核心r2為12頂層／4子PASS、零FAIL／SKIP，Go1.24.13、97.5% statement coverage，go vet通過。實際EXE／ITEM hash及原始指令核對；已接受dosgolem aad971bb收據與224／225／230持久區逐檔hash核對，再比較八格交易與核心round-trip。沒有新正式InputState、PNG或V3。
- IDA9.4有界r1／r2匯出366／77條，來源保持。轉職10C42..10C5E的高byte清除受新職業條件限制，尚缺正常原版oracle，不在核心猜補全域解除裝備。sidecar與腳本留既有work，不加入Git。
- 正式Game／Member、pack、戰鬥及全部save adapter仍DRAFT；schema0.19.0／content0.1.91、canonical66224bc0保持，存檔尚未升級。此前928c7db出售修正與完整回歸保持，206..208仍RED。提交／推送、使用者13項資料、原版素材、擁有權與Docker清理由收尾收據核對。

## 2026-10-05 A 物品編碼與初始八格接入pack

資料包升至schema0.20.0／content0.1.92，canonical0839ecc9。characters JSON保存完整編碼、128筆實際ITEM部位／詛咒metadata及兩個角色明示八格；舊equipment初值移除，既有裝備預覽由words導出。正式boot核對實際archive count、完整record shape及逐筆原始decoder，新的啟動拒絕測試通過。Game／Member的可寫bag／equipment與完整save仍待遷移。

- READY／有限CONFORMED及原始bytes、等級、工具、CLI集中docs/188，欄位契約docs/84。初始words另與aad971bb接受收據第一個正常packet核對；原始IDA9.4創角primary十筆writer rows保持原始strong流程限制，不外推整個創角或campaign。
- target r3為13頂層／40子PASS、零SKIP。首輪D1未拒絕是validator缺口，補D2/D3 gate；登錄OR指令定位及非裝備fixture為測試問題，依原始bytes與metadata訂正，不放寬資料。D3初值gate及正常初始八格另驗。
- 從乾淨146b549的pack產生兩份獨立副本，九JSON逐byte一致並等於正式檔；重建器先驗EXE／ITEM完整hash及原始instruction bytes，原版資產維持唯讀。來源／產物hash留既有work收據。
- 逐項完整game為486頂層覆蓋、490次執行，435不同頂層／121子PASS、51選用SKIP；internal188頂層／412子、12套件PASS及4選用SKIP。正常新遊戲至THE END141.62秒、Linux desktop建置通過，oom／oom_kill0；1087張既有PNG逐byte保持，現行來源重編game.test與完整binary一致。選用SKIP不作原版parity。
- 206..208物品畫面仍RED，沒有新增物品UI／正常給予／完整word save的驗收，也沒有新發行包。下一主要切片接所有物品持有者，再閉合正常218／230及存讀檔；已完成核心與pack不重開。
- root基線3213保持、零.md目錄，UID/GID1000抽查；使用者13項資料不stage，原版／database／PNG不加入Git。提交、推送、Issue精確讀回及Docker容器清理由最終收據核對。

## 2026-10-05 單人道具action Esc返回場景

接續3be415f，依Issue #4及docs/188有限READY修正獨立取消分支。正式產品只改單人、已選主角、action Cancel，關閉panel及父指令並清暫存選取；多人及目標取消保持既有路由。A完整持有權／save遷移仍DRAFT，不藉取消切片宣稱已完成。

- 正常新測試在未修正程式208重現留於清單。最終r3由正常新遊戲到194..218，取消不交易、重開、清單取消、右移、正式F5／F6及讀檔後左移PASS。受影響回歸22頂層／4子PASS，共23不同頂層／4子、零SKIP／OOM；go vet ./game通過。最近完整game／internal／THE END與Linux build仍3be415f，不冒稱本輪全套重跑。
- 修正後208完整差57417降411，217亦411；原始主角182、NPC14為106、NPC15為123完整圖塊解釋全部差異，其他畫布差異0。目視確認已返回場景。動畫時序未知、未V3，不改相位或圖像。206／207的清單與選取仍RED，A唯一集合／word save尚未接線。
- 新25張完整PNG與44張前綴留本機；57張既有指令窗PNG及44張新路線前綴逐byte保持。公開核對器由docs/188索引，重驗原版462份產物、runtime收據與PNG hashes，核對全畫布及原始CTY／BLS／BLK。兩種損壞輸入在指定檢查點拒絕，之後完整正對照逐byte相同。
- 首輪修正後將RGB0當功能gate的測試假定已依原始圖塊診斷分級勘誤；r2讀檔後行走過早失敗，依既有cooldown與field save測試補自然閒置輸入。正式產品兩次訂正均保持，歷史收據不覆寫。
- schema0.20.0／content0.1.92／canonical0839ecc9保持，沒有新包。使用者13項資料、原始素材及root基線由最終收據核對；提交／推送、Issue讀回與Docker清理記錄於收尾收據。

## 2026-10-05 正常物品丟掉：原生結果、穿戴拒絕與零價gate

2026-10-05 正常單人物品丟掉已限定驗收：成功只清所選物理word、保留空格及其他順序；顯示原始277姓名／物品與兩行結果。穿戴801E拒絕並顯示272，不消耗。兩者保留實際操作底圖，等新Enter返回場景。ITEM零價拒絕gate已補齊；A單一Store、storage_version1與save_version2保持。schema0.24.0／content0.1.96，canonical44ce09cb。

dosgolem正常261包、299輸入、598IRQ1、591產物接受，233前綴507份保持；seed1357執行前固定一次，無注入／restore。完整2172bytes只交易物理格0／1。正常231..261、正式F5／F6及下一步通過，RNG保持。241／260全640×350差122由英雄完整MST6／7解釋；250差106由NPC14完整MAN201／200解釋。未解釋差異0，沒有遮罩、裁切或改相位，窗口／文字V2；動畫時鐘及完整V3仍未知。291張舊runtime PNG、新路81張前綴及231..233保持。

完整game499頂層清單覆蓋，448不同頂層／141子PASS、51選用診斷SKIP；internal198頂層／427子、12套件PASS、4選用診斷SKIP。11項必驗零SKIP；正常THE END105.14秒、go vet及Linux desktop PASS，正式收據OOM0。九JSON兩份乾淨3dc5b47重建一致，兩類checker各三負例拒絕。沒有新包，完整原版campaign、音畫與未測分支保持未知。正常261返回場景後開啟狀況命令，核對第一個玩家可見結果與返回。先取得dosgolem原版證據，再依RE→READY修正；已閉合的A、225、木棒使用與本輪丟掉不重開。 Issue／Goal保持進行中，唯一現況表在CONTEXT。

Issue #4登記本輪與中途結果，原版來源／限定READY／CONFORMED及完整收據集中docs/188。失敗與訂正：2GiB測試OOM、Go觀察工具變數拼錯、checker的last_record及mount錯誤、元件physical index夾具與THE END零價鑰匙舊策略，均保留私有收據；沒有重擲或改原版。原先刪錯格假說未重現，正式A選取已是物理位置。两個受舊腳本影響的測試容器明确中止後清除，最終正常／完整回歸同binary，無OOM。

game binary `9fc8bf356f38781e226c2dc57f19225270a47134a296e7ff6844dda9c9941de2`；desktop `22561e266e21285ce891ac685775b1a868389aa360bdceeec25ac0b8db00adfa`。來源三負例及畫布三負例拒絕；291舊PNG、81新前綴及231..233保持，九JSON兩次重建一致。IDA647指令及原八筆annotation保持，新三筆自動合併。

正常261返回場景後開啟狀況命令，核對第一個玩家可見結果與返回。先取得dosgolem原版證據，再依RE→READY修正；已閉合的A、225、木棒使用與本輪丟掉不重開。完整原版campaign與動畫時鐘未知，不宣稱整款完成。依既有授權commit並push，確切SHA與遠端結果由Issue #4結果留言及git log核對；本輪沒有發行包。Docker／root-owned及保護檔案收尾見本批衛生收據，未清理其他專案。


## 2026-10-05 狀況首選單正常對拍

正常健康單人狀況入口已修正，先顯示原始三列選單，保留命令窗與底圖；上下單欄繞回、Esc返回及左右下一步通過。首列保留既有詳細窗E2入口，全體及排序缺證據時不交易狀態；詳細內容及其他角色未由本輪驗收。A單一八格、storage_version1及save_version2保持，schema0.25.0／content0.1.97，canonical694b740c。
dosgolem正常273包、311輸入／622IRQ1／627產物接受，261前綴591份保持；seed1357執行前固定一次，無注入／restore。262..271完整2172bytes保持，272只變player X，273完整返回baseline，clock30保持。正式262..273、snapshot／RNG、F5／F6及下一步通過；264..270首選單文字／外框／游標RGB0，完整640×350仍差142，剩餘影格與陰影差異未新增raw phase驗收。273完整RGB0，不外推其他畫面或動畫時鐘。
完整game501頂層覆蓋、450不同頂層／141子PASS、51選用診斷SKIP；internal200頂層／438子、12套件PASS、4選用診斷SKIP。13項必驗零SKIP，正常THE END107.22秒、go vet及Linux desktop PASS，同一game binaryec80bdec，OOM0。403張舊PNG保持，九JSON兩份乾淨771ec63重建相同；兩类checker各三負例拒絕，正對照重複相同。IDA241筆原始指令保持，原11筆annotation保持，新取址語意strong自動合併。沒有新包，完整原版campaign與音畫仍未知。
從正常273返回場景後重開指令，選狀況首列，核對詳細窗的第一個玩家可見結果與返回。先取得dosgolem原版證據，再依RE→READY修正；已閉合的A、給予、木棒使用、丟掉及首選單不重開。 Issue／Goal保持進行中，唯一現況表在CONTEXT。

先依Issue #4的771ec63正常261checkpoint重生原版；264首次顯示三列，remake直接詳細窗為RED。依docs/188有限READY接入schema0.25.0／content0.1.97的狀況selector，加入原始EXE／TXT parity與資料拒絕檢查，正常冷啟動至273並補F5／F6及下一步。實際畫布與收據見docs/188／84，README只更新穩定摘要。

失敗分類與訂正：inspect建立既有scratch後被prefix guard拒絕；IDA18338沒有函式邊界，以原始有界範圍匯出；272測試未送DirHeld，診斷證實PX3、clock30／RNG保持，補正常輸入與自然冷卻後同命令乾淨重跑通過。候選sidecar的歷史模板標籤已校正。沒有改原版、重擲、state注入或猜補行走規則，失敗收據保留。

來源及畫布checker六負例拒絕、重複正對照相同；403舊PNG保持、九JSON兩份乾淨771ec63重建一致。最新IDA241條原始定位及原11筆annotation保持，新增1830B strong語意自動合併。完整與正常game binary為`ec80bdec35e9f16e9859d8aad2b2287fe2bf34df859049e00bea0c45b093057f`，desktop為`dc41893a8727832a42a2ce3e160b417c1db6c8f73da091f08f8eb0e4fd2ca235`。依既有授權commit及push，確切SHA以git log與Issue #4結果留言核對。Docker清理、root-owned基線及保護檔案由本輪衛生收據核對，沒有新發行包，未清理其他專案。


## 2026-10-05 最終資料契約審查與驗證

提交前將選項數量改由資料包提供，拒絕空選單；DQ3原版三項保持。這是共用引擎契約修正，沒有新增版本專屬常數。前述runtime-r2、full-r1及internal-r1為修正前實際收據，保留歷史。最終使用同一game binary `23afdb032c958cfa8baaa04bbf599736b0a7be9654a33259b2a5d65c0b78e058`，正常13項必驗零SKIP，完整501頂層覆蓋、450不同頂層／141子PASS、51選用診斷SKIP；internal200頂層／439子、12套件PASS、4選用診斷SKIP。正常THE END136.94秒、go vet及Linux desktop PASS，OOM／oom_kill0。desktop為`39d985152d516cde6fd89297247cfa5167a0e8f5721947d15e9145ac9fd782b4`，14603000 bytes。

最終runtime-r3與r2的527張PNG及狀況收據逐byte相同；403張上一丟掉切片PNG保持的證據仍成立。公開畫布checker再核對r3，首選單文字／外框／游標RGB0，完整畫布142餘差保留。其中20像素位於NPC14與子窗陰影交界，其餘122像素位於NPC15區域；未新增raw phase驗收。取消後左右下一步與存讀檔通過，不宣稱詳細窗、全體、排序或動畫時鐘完成。

internal-r2已通過，後續vet沿用既有輸出檔名被排他建立保護拒絕，屬驗證腳本問題。獨立vet-r2在同一工具鏈乾淨重跑通過，舊收據未覆寫。原11筆annotation、241條IDA原始定位與bytes保持；新增1830B取址語意仍為strong。31個變更檔擁有權1000:1000，新增production Go沒有版本專屬raw ID／座標／玩家文字，root-owned基線3213與零.md目錄保持。

| 最終本機收據 | SHA-256 |
| --- | --- |
| `work/dosgolem-opening/issue4-field-status-normal-r3-source-r1-receipt.json` | `a30e50edd8ab3a3f8f8f95532b84aff4773d5a19879385ab31158e4d57b9b9aa` |
| `work/issue4-status-runtime-r3/game-receipt.json` | `4084759689a1a7398a3cee7f80472b97b407a7db17974fdd763759e694b78255` |
| `work/issue4-status-full-r2/game-receipt.json` | `7a12e9e4efe7d518f50f4b85e0b09ad9bbc328ed02544cbc4966743f742c3ea1` |
| `work/issue4-status-internal-r2.json` | `adad6bc837987b78df702555e33c7de0f288105fea72bb48f7d2ff4b0757f43b` |
| `work/issue4-status-vet-r2.json` | `eb467c08af06cbb26d4cd884bf992cefeb996d3a2f3cd4c503bd8668ebdd67c3` |
| `work/issue4-status-runtime-r3/desktop-receipt.json` | `f878e8c60d214cded32875e4908d8fe8c7053acd8802bfda4418d8ede7ab66f8` |
| `work/issue4-status-runtime-r3/ITEM_STATUS/status-receipt.json` | `6cc603ded9c0815562397c50bb453cc89dbf8a3a46427d30500df244d7e9b02e` |
| `work/issue4-status-raster-r2.json` | `18e11e0b15e06f9ecea7961db5280e4fe77278db0d76b0c7fa3318b48ab81307` |
| `work/issue4-status-final-test-counts-r2.json` | `4dd07723398a74895e6960214a59f65682a579e78c7da31f70bd3af27eadc428` |
| `work/issue4-status-aux-r1.json` | `86777b9d30f97c4de06035f69fae479ed96f30cd6ab28fd0ab9015497d7418e6` |
| `work/issue4-status-r6-ida.json` | `135458368e6cacaf1e7a628c982d10f1d9c7e382d189ceb419e1bdb5883b6c0d` |
| `work/issue4-status-precommit-hygiene-r1.json` | `fc2fe5a836fdde296764eee4a2ded94719bcc80fb27d3c9c9a62fab7876cecea` |

工具入口與執行契約沿本節既有索引，唯一目前狀態表在CONTEXT。Issue #4保持OPEN，下一切片為正常273後狀況首列的詳細窗與返回；完整原版campaign仍未知。沒有新發行包。

## 2026-10-05 Issue #4 正常詳細狀況頁與新鍵返回

完整game503頂層覆蓋、452不同頂層／141子PASS、51選用診斷SKIP；internal200頂層／448子、12套件PASS、4選用診斷SKIP。14項新舊必驗零SKIP，正常THE END174.26秒、最終Go vet與Linux desktop PASS，同一game binary01d19989、OOM0。九JSON由94e1276乾淨重建相同，三種壞來源拒絕且正對照重生相同；IDA395筆原始定位／bytes與原12筆annotation保持，新18498新讀鍵限定confirmed自動合併。沒有新發行包。

正常280來源edfad334接受，seed1357一次，318輸入／636IRQ1／648產物，273前綴保持。詳細頁23054→0，274..277與280完整RGB0；278／279仍差351／356，保留未驗畫面。原生窗口、數字幾何、Store Worn及fresh-key返回接正式路徑，八格／snapshot／RNG及正常F5／F6後下一步通過。無狀態注入、restore、裁切或動畫覆寫。

完整證據、工具入口、失敗分類與收據集中docs/188，資料契約docs/84，唯一目前狀態表CONTEXT。比較助手的meta身份、far call bytes及跨掛載fixture問題均訂正後在同一工具鏈重跑，未放寬原版比較。舊runtime527張PNG、395筆IDA原始定位與原12筆annotation保持。原版素材與使用者十三項未追蹤資料保留，不加入Git。Issue #4保持OPEN，下一切片為正常280後「看全體的情形」首個結果與返回。

## 2026-10-05：Issue #4 正常單人全體狀況頁

接續9a19513，依Issue指定全體狀況切片，dosgolem冷啟動正常288、seed1357一次，父280及648項產物保持。326鍵／652 IRQ1／672項產物核對，來源fac03247；無狀態注入或restore。IDA9.4閉合窗寬writer、原生橫向文字、姓名與HP/MP／金錢consumer、新鍵返回及bit3略過陰影。

DRAFT先重現正式第二列沒有結果。prototype取錯金錢欄位、FON-only來源與stats欄位訂正後，剩餘342像素全部位於金錢窗陰影；原版1FC57分支補證後r4完整RGB0，再審READY。正式typed summary及共用有限renderer接入，所有版本文字、raw window、欄距、容量與數值原點保存在JSON。schema0.27.0/content0.1.99、canonical7891ab6a；save_version2/storage_version1保持。

正式正常281..285、287..288完整640×350 RGB0，286返回中間畫面差122保留，不強設動畫、遮罩或裁切。新Enter消耗後返回field，snapshot、八格與RNG保持，正式F5/F6及Load後下一步通過。完整game505頂層覆蓋、454不同頂層／141子PASS、51選用診斷SKIP；509次執行／599 PASS記錄。internal202頂層／462子、12套件PASS、4選用診斷SKIP。七項指定正常／受影響路線零SKIP，正常THE END92.23秒、Go vet及最終Linux desktop PASS，同一game binaryf4e518b7、OOM0。

上一完整game的1688張PNG逐byte保持，九JSON由9a19513乾淨重建一致。公用來源checker正對照全部資料保持，壞PNG、缺IRQ與壞probe source均拒絕；只訂正公開標題後再核對最終checker身分。IDA605列原名、位址、bytes及relocation保持，13筆舊annotation保持，新增三筆限定confirmed自動合併。原版r1未執行便遇到Go cache唯讀；r2補明確cache掛載後一次完成。稽核768MiB被終止後在2GiB核對同一來源，沒有重擲。這些均保留為工具／環境紀錄。

證據、READY、CONFORMED與私人收據索引集中docs/188；JSON契約docs/84，唯一狀態表CONTEXT。下一切片正常288後的重新排序入口。多人全體V3、動畫時鐘、聲波與完整原版campaign保持未知；沒有新發行包。工作依授權登記Issue #4、commit＋push。使用者十三項未追蹤資料及.claude保持，原始素材、圖像、binary、私人收據與IDA database不加入Git；root-owned基線3213、零.md目錄及本輪Docker清理核對。

## 2026-10-05 單人重新排序兩頁與fresh-key返回

依Issue #4續行正常288，原版單人第三列以DI020A顯示TXT00/522波魯多加王的信。r1只預定終末等待而停在294的inline216D8；r2按實際狀態以295新Enter續頁、296另一新Enter返回，297／298左右行走。固定seed1357一次、無注入／restore或重擲，正常298、336鍵／672 IRQ與702產物接受，父288保持。持久2172bytes僅行走X改變；引用信件原因unknown，不另寫合理提示。

一次性prototype十張全RGB0，READY後正式接入typed FieldStatusReorder與confirm型保留行EOF等待；窗口、文字、捲動與指示資料在pack。schema0.28.0／content0.1.100、canonical757ef211，save2/storage1及A八格保持。正式正常289..298完整640×350 RGB0，fresh-key不滲透，snapshot／RNG、F5/F6與Load後下一步通過。

完整game507頂層覆蓋、456不同頂層／141子PASS、51原有選用診斷SKIP，511次執行／601 PASS記錄。internal204頂層／476子、12套件PASS、4原有選用診斷SKIP；七項指定路線零SKIP，正常THE END123.74秒、Go vet與最終Linux desktop PASS，正式收據OOM0，同一game binarybbcca34a。1827張前輪完整PNG、727列IDA原始定位及16筆舊註記保持，新18694限定confirmed自動合併；九JSON乾淨bdc955f重建一致，公開checker正對照與三負例通過。

首輪未使用import、r2未註冊文字版型、收尾誤用前前輪PNG數與Go gate包含JSON均保留訂正，回查規格與同一產物後乾淨重跑；未放寬產品驗收或修改原版。證據與重生工具入口docs/188，資料契約docs/84，唯一現況表CONTEXT。下一切片正常298後未學咒文單人咒文入口。多人排序、動畫、音畫及完整原版campaign保持未知，Issue／Goal進行中，沒有新包。使用者十三項資料保持；本批UID/GID1000、root-owned3213及零.md目錄保持，測試與IDA一次性容器已清理。

## 2026-10-06 單人空咒文入口與新按鍵返回

依Issue #4從正常298續行。dosgolem一次固定1357冷啟動至304，342鍵／684 IRQ／720產物，父來源702份保持；單人自選索引1，人物原始+30／+31均0，DI0106顯示TXT00/262，保留命令背景，fresh21133後返回field1997C並可行走。探針固定基址診斷誤名保留勘誤，以人物指標快照及AX分支驗證，無重擲、注入或restore。

可丟棄正常重播首次301差52013，READY後資料進field_spell_entry JSON，共用訊息primitive保留背景及新鍵消費。正式299..304六張完整RGB0，snapshot／八格／MP／旗標／RNG保持，正常F5／F6及讀檔後下一步通過。schema0.29.0／content0.1.101、canonical0e9d617d，save2／storage1保持。

完整game509頂層覆蓋、458不同頂層／141子PASS、51原有選用診斷SKIP，513次執行／603 PASS記錄；internal206頂層／489子、12套件PASS、4原有選用診斷SKIP。五項指定路線零SKIP，正式THE END65.99秒、vet與Linux desktop PASS，OOM0。前輪1976張PNG保持；510列IDA原始定位及17筆舊註記保持，兩筆有限confirmed追加且自動匯出。九JSON乾淨317014c重建一致，公開checker正例與三負例通過。存讀檔測試最初未計既有存檔點／Load時鐘交易，vet沿用輸出名稱及PNG負例改到長度均屬驗證工具問題，保留失敗後按實際契約乾淨重跑，沒有修改產品規則或原版。

本批UID/GID1000，root-owned3213及零.md目錄保持；原版、圖像、database與binary留本機，不新增image或交付包。來源／READY／CONFORMED與工具入口docs/188，JSON契約docs/84，唯一現況表CONTEXT。下一正常切片為304後單人裝備入口與取消；其他施法、多人、動畫與完整原版campaign仍未知，Issue／Goal進行中。


## 2026-10-06 正常裝備入口原版314與首次RED

- Issue #4由正常304接續，固定seed1357一次冷啟動到314；352按鍵／704IRQ／750產物與父720前綴核對。四次Esc逐槽跳過，312返回、313／314行走；完整2172bytes只改左移座標。来源234fc67e，沒有注入、restore、檔案writer或未實作服務。
- 可丟棄正常InputState診斷305..307完整RGB0，308差39570，remake仍顯示選人窗；snapshot／RNG與前空咒文六張保持。診斷35.78秒PASS、零SKIP／OOM，不能稱裝備parity。
- 官方IDA9.4 sidecar697b06e3保留335函式指令／468唯讀decode及原始null函式身分。callback未自動成函式、MZ relocation比較、builder空scratch與INPUT缺kind、checker1GiB137、audit副本scratch均保留勘誤；同原版來源與公開checker正例／三負例後續通過，OOM0。
- 新增三個公開重生工具，更新CONTEXT唯一狀態表、PROJECT_MEMORY、docs/74與docs/188；正式Go、schema0.29.0/content0.1.101、save2/storage1保持，沒有新image或交付包。
- final audit efd326f4：來源、首個RED、公開正負例、IDA、所有者及root-owned3213／零.md目錄核對。十三項使用者檔案保留；本輪一次性Docker與Xvfb已終了，提交與推送的具體SHA由Issue結果留言記錄。
- 下一閘門仍DRAFT：列表篩選／physical slot、窗口dynamic consumer、攻擊／防禦及穿戴writer先閉合，再審READY並修正正式裝備流程；Issue／Goal保持進行中。

### 2026-10-06 正常339單人裝備穿戴／卸下限定修正

- 原版固定1357一次正常339，377鍵／754IRQ／825產物，父314保持；IDA9.4原始單人入口、物理候選、四槽窗口及writer-consumer達有限READY。
- 正式入口直接選主角，完整候選保留物理格、重複及已穿戴，末列卸下只清旗標。原生順序與版面、文字、資格進pack；schema0.30.0/content0.1.102，A八格、save2/storage1保持。
- 35步完整word、snapshot、RNG與返回能力通過，卸下後F5／F6及下一步通過；20張完整RGB0，15張仍有動畫餘差。原版完整campaign與所有動畫未知，沒有新包。
- 完整game511頂層覆蓋、460不同頂層／141子PASS、51原有選用診斷SKIP，515次執行／605 PASS記錄；internal210頂層／504子、12套件PASS、4原有選用診斷SKIP。七項指定路線零SKIP，正式THE END107.10秒、vet與Linux desktop PASS，OOM0。2131張舊PNG逐byte保持，2321張本輪PNG；九JSON從乾淨d830251重建一致，公開checker正例／三負例通過。IDA390列原始定位，335舊列及468候選保持；沒有新包。
- 元件編譯、文字控制碼解析及IDA稽核暫存路徑失敗已保留docs/188勘誤，後續同工具鏈重跑通過；沒有修改接受來源或挑選亂數。
- 原始檔hash、UID/GID1000、root-owned3213與零.md目錄保持。先前逾時掃描容器已明確停止，完整回歸Xvfb及一次性容器終了；十三項使用者資料保持。提交／推送SHA與全部遠端結果由Issue #4記錄。
- 下一正常來源從339穿戴另一件甲胄，再核對狀況與穿戴後存讀檔；Issue／Goal持續。

### 2026-10-06 正常366第二件甲胄與穿戴後第二槽存讀檔

- 原版404鍵／808IRQ／908產物與父339的825份保持，來源377d6d30。physical5甲胄穿戴、防禦力8、詳細狀況、原生2172bytes第二槽保存／完整讀回及下一步閉合。固定1357一次，無注入／restore。
- 正式正常27步word／snapshot／RNG、其他九槽與讀回通過；五張完整RGB0，其餘22張差106..356保持。新來源工具與正常回歸加入索引docs/188，A八格、save2/storage1、正式產品Go及九JSON保持485f3c2。
- 驗證準備型別／檔案形態與runpy搜尋路徑錯誤保留；正式整批r1共用目錄失敗、r2舊路線OOM，均不列產品缺陷。r3沿既有獨立程序契約，三項PASS／零SKIP／OOM。來源正例及三負例、vet與所有者檢查通過。最近完整game／internal／THE END／desktop沿前checkpoint，沒有冒稱全套重跑。
- root-owned3213及零.md目錄保持，十三項使用者資料未提交；原始素材、影像、database與binary留本機，沒有新image／包。Docker清理及提交／推送SHA由Issue #4結果留言與handshake記錄。
- 下一正常切片：正常366後把已穿戴physical5甲胄換到另一物理格，確認舊格清穿戴旗標、新格設旗標與狀況及返回；先dosgolem來源，再審有限READY。Issue／Goal保持進行中。

### 2026-10-06 正常381同部位甲胄換穿

- 原版419鍵／838IRQ／953產物與父908保持，來源d13ac425。372清physical5旗標、設physical4旗標，其餘持久區保持；374攻8守8、詳細consumer與返回閉合，無注入／restore或新存檔writer。
- 新增來源工具與正常回歸，共用366前綴，不改正式產品Go／JSON。三項獨立程序PASS、零SKIP／OOM；vet與來源正例／三負例通過，各217張舊PNG、原366報告與新15張診斷保持。12張完整RGB0，3張仍差351／229／122，不稱全畫布V3。
- 381後remake正常F5／F6及下一步通過，沒有原版換穿後存檔樣本；最近完整game／internal／THE END／desktop仍485f3c2，本輪未重跑全套。
- UID/GID1000、root-owned3213與零.md目錄保持，原版資料／影像／database留本機，十三項使用者資料未提交；沒有新image或包。Docker／遠端收尾與提交SHA由Issue結果留言及handshake記錄。
- 下一來源：從正常381開道具清單，核對第四動作列的第一個結果與返回；先dosgolem來源，再依證據審READY，不預設效果。Issue／Goal保持進行中。

### 2026-10-06 Issue #4 正常394道具三動作與第四列待辦勘誤

接續0e41f5a，沿使用者已確認A單一八格。Issue開始留言6002178566先登記本輪DRAFT；既有207來源是三項，故先查證而不預設第四項。原版新冷啟動394包／432鍵／864IRQ1／992產物，source c4878843；父381全部事件及953產物保持，seed1357一次、無注入／restore。三Down確認1→2→3→1，Esc返回，左右行走；2172bytes除393位置低byte保持，原生保存檔不改。公開producer／checker及正式測試由docs/188索引。

來源checker首輪草稿預設公開producer尚未建立，補同byte入口後同image乾淨重驗；正例完全重現，PNG CRC、缺IRQ、變producer三負例拒絕。讀大型收據256MB容器曾終止，1GB同命令重讀通過。remake草稿r1／r2的3GB程序OOM保留；manifest單次讀取試作未解決主因，未進正式測試。6GB r3 heap診斷找到約99.56%配置在WritePixels，before381存活heap近2.98GB，loaded394 RSS超過3GB。依量測採有界4GB容器及原始畫布helper，394、共用381、裝備範圍三項獨立程序PASS，零SKIP／OOM，vet通過。

正式394八格／snapshot／RNG／clock、全部十JSON槽保持、F5／F6與Load後下一步通過；新增13張全畫布中382..393十二張RGB0，394差122保持。每路232張舊PNG及原381報告不變，新13張與診斷逐byte同。138份正式Go／JSON保持485f3c2，A八格及save2/storage1不變；最近完整回歸沿485f3c2，未聲稱本輪全套或campaign完成。final-audit-r1保存所有者、畫面與收據核對，root-owned基線3213、零markdown目錄。

修正目前計畫：健康單人此路線無第四項，下一正常394選「調查」首個結果與返回。歷史計畫與多人未知留存。原版資產、影像、binary、profiling及研究收據只在本機work，無新發行包；使用者十三項未追蹤資料保持。驗收後commit＋push及Issue結果的實際ID由work/issue4-item-action-count-handshake-r1.json記錄。全部一次性容器完成後清除，其他專案容器不動。Issue／Goal繼續。

## 2026-10-06 Issue #4：徒步空結果調查與新鍵返回

原版正常401／404來源接受，394及401全部前綴保持；修正前缺原始訊息，401完整差25761。依docs/188有限READY，新增typed field_examine與原始264／265、共享3E6E窗口和新鍵返回；所有版本資料留九JSON，schema0.31.0/content0.1.103、canonical0612be3e，A八格/save2/storage1保持。共用訊息helper可按原始兩record換行，正常394延續helper保留原nil分支。乘船249潛水分支未審，不套徒步264；scope測試鎖定排除shipAboard。

最終runtime-r2正常404與scope兩項零SKIP／OOM；395..402八張完整RGB0，403／404仍差356／351。十JSON槽保持、正式F5／F6及Load後行走通過。原版本輪沒有保存404的新樣本，此保存驗證只屬remake內部。245張舊394 PNG保持；九JSON由3aff82a乾淨重建相同，來源IRQ／DI／producer三負例拒絕。IDA259列定位／bytes／xref與19舊annotation保持，新18C93／18C9B／18CAC三筆confirmed由sidecar自動附註；root-owned3213與零markdown目錄基線保持，輸出UID/GID1000。

最終完整game516頂層覆蓋、465不同頂層／141子PASS、51既有選用診斷SKIP；internal212頂層／520子、12套件及4選用SKIP PASS。正常新遊戲至THE END65.16秒，vet及Linux desktop PASS，OOM0。full-r2與runtime-r2為最終foot guard；full-r1已自然PASS後發現乘船範圍缺口，保留舊收據並重新跑最終版本。source欄位、IRQ負例tag、far-call file／loaded bytes及大日誌讀取的草稿問題均已分類和訂正，沒有當作產品缺陷。

下一批無對象對話原版406首個260來源接受，SHA7406970c、444鍵／888IRQ1／1028產物，404前綴保持；只屬source-only，正式對話未修。r1草稿不存在的變數造成編譯失敗，r2來源與獨立checker通過。下一步延續406取得返回／行走，補caller與READY；原版完整campaign、動畫與音畫仍未知。

Issue #4持續更新；本次以「fix: restore native empty examine response and fresh-key return」提交並推送origin/main，實際提交身分由git log與Issue最終留言回查。僅提交程式、JSON、工具與既有文件，不加入原版／database／影像／binary／私有探針及13項使用者scratch。一次性--rm容器已自然清除，無殘留專案容器。驗收與重生入口統一見docs/188及CONTEXT；沒有新發行包，Goal保持進行中。

## 2026-10-06 Issue #4：無對象對話260、新鍵返回與正常409

原版406及409來源接受，固定seed1357一次、447鍵／894IRQ1／1037產物，406全前綴保持。IDA9.4以命令首callback14E0E、無NPC與櫃台、14E7F/14E82 DI0104→15023共同訊息窗→2111B新鍵等待→14E85返回閉合。原始EXE／TXT及24bytes窗口固定hash；原22筆annotation保持，兩筆confirmed新增；重新有界399條原始定位／bytes／xref保持。依docs/188 RE→DRAFT→review→READY後新增typed field_talk、record260及原生presentation，正式selectCommand無對象else走共用訊息primitive，有NPC／櫃台／故事分支不改。A唯一八格/save2/storage1保持，schema0.32.0/content0.1.104、canonicalc4285a9e，未打包。

正常409、十槽保持、RNG／snapshot、正式F5／F6 roundtrip及Load後行走PASS。406全RGB25799→228，全為英雄122及NPC14為106；405／407／408／409差351／0／229／122，原始BLS／BLK完整診斷未解釋像素0。NPC15在406被原生訊息窗全遮擋，完整768 overlay像素相同，未猜hidden frame。255張舊404 PNG及新409前綴保持；全套normal409再生同PNG。九JSON由c452046乾淨重建相同；producer／IRQ1／native DI三負例拒絕。

過強405完整V3斷言、診斷PLTE未用色／NPC15遮擋、負例scratch路徑與下一DRAFT checker語法均已訂正，原版與正式renderer未因腳本問題修改。512MB摘要誤讀大來源退出137，按既有3GB稽核契約重跑；正式4GB正常與full有memory.max壓力4897／5189，但OOM與kill均0，保留實際限制。

最終完整game518頂層覆蓋、522命令、467不同頂層／145子PASS、51既有選用SKIP；internal214頂層／531子、12套件及4既有選用SKIP PASS。正常THE END72.52s、Go vet與Linux desktop PASS。最終收據與完整重生入口見docs/188；root-owned3213／Markdown目錄0基線保持，修改及輸出UID/GID1000。所有本批--rm容器自然清除，未動其他專案容器。

下一批原版410 Enter首結果來源62e40319接受，448鍵／896IRQ1／1040產物；409全前綴保持，ready/1997C、3,18、不開命令窗、持久區及clock0保持。仍source-only，下一步從合法409以正式Enter驗remake及後續操作，未審行為不進production。原版完整campaign、動畫時鐘及音畫仍未知。Issue #4各階段已更新。

本批提交標題為「fix: restore native no-target talk response and fresh-key return」；提交與推送完成身分以git log及Issue最終回讀為準，尚未核對前不作已推送聲明。僅納入本批程式／JSON／工具／既有文件，13项使用者scratch及原版素材、database、binary與影像不入Git。Goal保持進行中。


### 2026-10-06 NPC自動移動、四圖層常式與正常458收尾

接續97b5c9e屋內NPC圖層切片。Issue #4已登記原版續行、局部證據、主線失敗查證與測試路線修正。唯一現況表已更新CONTEXT；詳細原始輸入、地址基準、推論等級、READY與有限CONFORMED集中docs/188，舊docs/35錯誤按追加勘誤回填，docs/84保存schema0.33.0契約。README僅更新穩定現況。

修正NPC viewport／同圖層／逐格順序、RND10、商值轉向與地形低byte，所有參數從pack提供。原版正常458與1184產物接受；1672入口返回、499scan及四圖層component相同。正式409／422／458與存讀檔通過，842 PNG與NPC實作基準保持。458全畫布1751差異仍保留，完整NPC骰序／動畫與原版campaign未知。九JSON乾淨重建一致，A八格/storage1/save2保持，沒有新包。

完整清單523頂層由526命令基準及10受影響補驗覆蓋，51既有選用SKIP；internal216／546、12套件及4選用SKIP通過。正式新遊戲THE END82.48秒、Go vet與desktop通過，最終必驗零SKIP／OOM0。基準唯一主線失敗已追到測試玩家策略及持有者／容量假設；全面旅行治療試作撤回，正式補給、復活、給予與魯拉後r16通過。沒有改產品戰鬥數值、物品規則或種子。r1..r15全部失敗保留，不挑重擲結果。

圖形3GiB補驗OOM三次，保留final-r2；同工具鏈與命令調4GiB乾淨final-r3通過。vet混用參數、tmp_dump.go及IDA匯出引用舊動畫來源均為工具／runner錯誤，修正後重生，不歸產品缺陷。436筆IDA原始定位／bytes／xref與原19註記保持，新註記逐筆來源正確；三類壞來源拒絕，正對照同fe39794a。最終稽核3e0fd436、證據稽核a130bc54，詳細本機收據入口見docs/188。

root-owned3213與Markdown目錄0基線保持，新工具與輸出UID/GID1000；13項使用者scratch及所有原版素材／影像／database／binary不提交。一次性容器收尾清理，未重建image、未動其他專案資源。本批提交標題「fix: match native NPC automatic movement rules」；提交／推送身分以git log及Issue最終回讀為準。Goal維持進行中，下一切片由合法458追加原版右移一鍵。

## 2026-10-08 Issue #5：局部 matching 的首批工具與實測

使用者要求先登記GitHub工作項目再開工，已建立
[Issue #5](https://github.com/wicanr2/kinginformation-dq3-re/issues/5)。本輪驗證局部matching
能否減少對拍定位工作，正式Go／game-pack與正常458基準保持，Issue #4仍暫停。
唯一目前狀態表為CONTEXT，研究入口docs/25；舊C exact及compiler鎖定的反證追加於
docs/17、docs/19、docs/135，不重寫歷史證據。

- 原版`assets_raw/DQ3.EXE`115282bytes、SHA256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`前後相同。沒有原地解殼、patch或修改database。
- IDA9.4重生既有NPC六範圍436指令；新探針另匯出入口／RNG／兩個NPC函式216指令與828自動函式導覽清冊，BX有界RNG另存原始範圍。動畫計數缺自動邊界保持unknown；四個target的Hex-Rays都拒絕16-bit。
- 新OMF有限reader按PUBDEF／SEGDEF定位並解析FIXUPP；RNG兩個外部符號DS offset由原始operand明示供給。缺placement、未知PUBDEF、far fixup、checksum與修改ASM常數五負例全部拒絕。
- 五個既有C樣本加兩個RNG候選全部新編。最終`msc-r3/receipt.json`SHA256 `43c62d234021022cd87070079b4a464113c86ed6b341eacbcbb9df724ade0180`：四DIFF、三REFUSED、C exact0；同一次DOSBox批次0.920594秒，未宣稱舊流程或整體對拍加速倍數。
- NASM2.16.01從`re/match/sub_e6b9.asm`七條指令重新組譯16bytes，與IDA及原版完全相同。scaffold雜湊相同，但其餘115266bytes保留原始輸入，C重編bytes0，沒有完整原碼完成聲明。
- MSC工具映像以固定Debian digest／snapshot重建，compiler只在runtime唯讀掛載。Inertia固定commit c555363b、Python3.14.7、uv.lock，另以官方wheel hash補上游漏列的Cython3.2.0；r2取代r1，使用明示Python reference lifter。
- 環境與腳本失敗均保留：首次360／600秒建置逾時；Inertia r1缺native lifter、r2缺Cython、r3快取權限。只掛特定可寫快取，保持source唯讀及非root。r4探針誤套底層base1000，CLI實為10000；r5 loader smoke缺註冊。修正後r6才作反編譯能力證據，不將這些失敗寫成產品缺陷。
- `inertia-r6/receipt.json`SHA256 `68e85d15426941097e0556d7b420aee43ebf388d00fe8758d4e1fa41abb88c6a`：RNG／有界RNG／NPC mover三案各54.87／55.53／54.22秒，16份中間C層，全部exit4、validation failed、merge gate hold。RNG可見未初始化DS及缺回傳值；未採用、未編譯中間C，也未修改上游語意來製造通過。
- 最終`msc-final-audit.json`、`inertia-final-audit.json`及`final-audit.json`核對來源新鮮度、正確位址、原版hash、四工具語法與索引正對照、16份中間artifact hash；新輸出root-owned0／Markdown目錄0，UID/GID1000。

完整命令、限制與重生入口見docs/25及tools/build/README。README只保存現行工具契約，
移除過期C／SDL目標；工作歷程保存在本節。下一步限縮RNG／BX有界RNG的caller、
暫存器ABI與compiler候選辨識，原版compiler及整體加速仍unknown，Issue #5保持開啟。

本批只提交研究工具、具語意ASM與文件，標題為
`research: establish DQ3 matching probes and semantic RNG assembly baseline`；提交／推送身分
以Git log及Issue最終回讀為準。所有--rm研究容器已清除，只保留MSC與Inertia r2現行映像；
被取代r1已刪，所需IDA私有image保留，未動其他專案資源。使用者13項未追蹤資料與
原版／database／OBJ／binary／中間C均不提交，沒有新發行包。

### 2026-10-08 Issue #5續行：RNG寄存器ABI與46-byte source重編

使用者再次授權開工，沿docs/25已呈現的RNG／BX有界RNG與compiler控制分支。
正式Go／game-pack、schema0.33.0/content0.1.106及Issue #4暫停狀態保持。
遠端主機gh auth首輪自動核准審查逾時未執行，按工具指示重試一次成功，沒有憑證阻塞。

IDA9.4從唯讀5178fdc8原版建立一次性database，重生核心33及有界49個direct caller窗口。
四NPC呼叫點直接MOV BX為4／4／20／10，consumer使用DX／DL與後續AL商值。
確認核心16bytes與有界30bytes的全部原始指令；新增`sub_e6c9.asm`不含db，兩函式46bytes
由NASM重編exact，全scaffold仍同hash，但其餘115236bytes原樣保留，C exact0。

新增dosgolem局部Go probe與Docker-only runner，只複製上游internal非測試來源與go.mod，
不使用dirty的cmd/probe/main.go，也沒有回寫上游。來源revision a9714ebd、selected source
canonical hash334a21511c2603ef606b889c960d17ff5099bad7e3b11ba67aa5ca703eddf2b0；
Go1.26.7，2GiB／2CPU，原版與來源唯讀，輸出UID/GID1000。

最終cpu-r3固定控制輸入：核心、BX10、BX0各65536seed與84邊界，每側196692次局部呼叫。
共2885416 CPU指令、0.568192秒；乾淨建置18.708801秒。原版與46-byte source scaffold
完整暫存器／高半部／segments／IP／SP／狀態word及memory delta相同。
核心已定義旗標另用opcode模型驗證，DIV未定義旗標沒有猜期望值。
seed1357／BX10為AX02bf、DX0007、state1b7d，錯把AX當餘數的預定負對照拒絕。
這是明示direct-entry、注入與CPU重入的局部證據，不取代正常玩家路徑或全局骰序。

已知C控制樣本實測：candidate MSC從stack argument BEEF取值，16-bit返回AXBEEF／DX2468，
32-bit返回DXBEEF／AX5f77，BX10保持。`_fastcall`候選報C2054／C2061，沒有OBJ。
只限制這顆候選及本組flag，不由此排除原版語言或其他compiler／pragma。
原先一般C介面無法單靠BP frame omission閉合BX／DX與商值consumer，相關勘誤在docs/25。

初次probe誤把WatchWrites視為每次寫入，seed00e4的低byte未變因而失敗；核對observer實作
後改驗memory delta，沒有改引擎。DOSBox先建空ERR檔亦曾被誤判compiler failure，改讀FAIL
內容後相同來源／compiler乾淨重跑；兩次失敗與輸出保留。單檔gofmt父目錄不可寫，改由
容器stdout格式化寫回同檔，未擴大掛載或chown。

cpu-r3收據SHA256042238a26df44e313b98279ea96da1157bd106542d14602f6d6cdf165889e47d。
五筆ABI語意由受版控`ida_rng_abi_ledger.json`保存原始位址／bytes／consumer／分級來源；
IDA重新建database自動合併五筆，舊NPC註記保持。根目錄、現況表與工具入口沿用既有職責，
詳細重生命令與本機收據在docs/25，下一步以這兩個exact fixture核對ABI-aware adapter／指紋。

本批只提交診斷Go／Python、語意ASM、reviewed ledger與文件；提交標題為
`research: close RNG register ABI and source-assemble bounded helper`，提交／推送身分以Git log
及Issue最終回讀為準。新Go raw IDs／位置均屬原版診斷與控制測試，不進production fallback。
使用者13項scratch、所有EXE／OBJ／database／生成binary與上游來源不提交，沒有新發行包。

### 2026-10-08 Issue #5續行：C／ASM adapter 與副作用界線

使用者以「繼續囉／go」接續已展示的RNG局部ABI、adapter與候選compiler研究。
先核對主機git狀態、唯一現況表與遠端Issue #5。Issue保持OPEN，現行遠端紀錄仍為722e778；
本輪不改Issue #4暫停狀態。知識路由命中retro remake、IDA9.4、compiler分流與文件職責，
沿用docs/25及既有matching研究根，沒有新增同義研究目錄。

新增Docker-only `run_matching_rng_adapter.py`、局部dosgolem Go probe與具名adapter組語。
診斷C從明示stack參數取上限，以DX:AX回傳餘數／商值；adapter保存其他暫存器，BX0直接返回。
OMF沿用實際PUBDEF／SEGDEF／FIXUPP，只接受已證實的DS:0B5A外部配置，沒有遮罩比較。
原版／compiler／dosgolem唯讀，只有既有研究輸出可寫；全部容器為目前UID/GID、network none
及有資源限制的--rm程序。

- 原版115282 bytes、SHA2565178fdc8保持，正式Go與pack沒有差異。既有IDA producer及兩份semantic ledger hash、新五註記原始bytes與196692案例收據核對通過；核心33及有界49 caller全部為xref type17。
- 原始兩個RNG常式由NASM重新組譯46 bytes exact；C候選56 bytes，adapter31 bytes。原版有界常式30 bytes，C raw比較DIFF，C exact仍0，其餘115236原始bytes不宣稱原碼重建。
- 最終固定輸入為BX10／BX0各65536 seed及80邊界。原版、exact ASM與C adapter共131152組，全部暫存器、高半部、段暫存器、返回位置與持久state相同；exact ASM全部觀測結果相同。
- 65616次非零呼叫都有額外stack modified-byte事件。第二輪另加入返回後完整128-byte scratch stack比較，這些呼叫的最終stack均不同；raw flags亦均不同。DIV算術旗標不猜硬體期望值，BX0的完整stack／state及旗標保持。
- 預定seed1357的交換AX／DX、誤傳CX、移除BX0 gate及DS offset0B5B四案全部拒絕。篡改編譯artifact亦在建置前拒絕，沒有改Inertia被拒絕的中間C。
- 三次DOSBox單候選compile為0.715865／0.766146／0.715652秒；r2／r3的OBJ、重定位C code及全部ASM artifact逐byte相同。最終CPU0.711280秒、乾淨Go建置9.127850秒；原版983920指令、exact ASM983920、C adapter3149248。
- 工具自動回報四類錯誤的register/state mismatch或division exception，取代人工逐項核對。但沒有同工作人工工時或完整玩家路線成本基準，整體加速仍unknown，沒有加速倍數。
- 候選CL／C2／C3含C5.10標記，位置分別為component file0x745d／0x2ccbe／0x1d6a9。CL亦含FORTRAN與Quick C字串，不能由全部banner推執行路徑或原版compiler。只修正Dockerfile未閉合的歷史斷言註解，建置指令不變，沒有重建重複image。
- 最終audit核對producer／parser／Go／ASM／IDA新鮮度、dosgolem選定source canonical334a2151、原版ledger與收據、索引正對照、UID/GID1000。Python語法、Go vet、gofmt、git diff --check通過；研究根root-owned與Markdown目錄皆0。
- 新Go原始定位全部位於tools局部probe，沒有production fallback。容器未安裝rg，依共用搜尋規則改以grep核對；正式game與pack的git diff保持空白。

新收據在`work/matching-decomp-20261008-r1/rng-adapter-compile-r3/`與`rng-adapter-cpu-r3/`。
CPU receipt SHA256 `9d702555c10aeccef610832acf2cf76d9cda8b660c161a886444707e1233246a`，
compile receipt SHA256 `b5c4f94f0f25409f127cdae35e5cd3f6a201bb57683085c19df641fbf6feede4`。
原版、compiler、OBJ、database、生成binary與先前protected scratch保持本機，不加入Git。
沒有正式遊戲修改或新包，因此本輪只驗研究工具，不將既有THE END與PNG描述為新回歸。

前段RNG ABI研究的預定獨立提交尚未執行，本輪將其未提交工具／ledger與已閉合adapter一併
提交，標題為`research: verify RNG register ABI and C assembly adapter limits`。提交／推送身分
以Git log為準。全部本批容器已清除，其他專案容器保留；現況與下一步已更新CONTEXT、
PROJECT_MEMORY與docs/74。下一步限於既有NPC caller的AX／DX雙結果consumer局部比較，
不擴大成完整EXE原碼重建，也不變更remake完成閘門。

### 2026-10-08完整matching Goal：範圍確認、完整導航與首個C PROC

使用者新增持續Goal「完成 dq3 matching decompliation」，並回答採主程式C精確匹配，
ASM僅限已確認底層常式；全EXE須由原碼乾淨重建且逐byte相同，db／raw code拼接不算完成。
這是使用者明示的新範圍，取代上節局部研究的停止線。原版／素材權利邊界、Docker-only、
正式Go／game-pack與Issue #4暫停保持。原版compiler與source-unit分類仍未知。
上一Goal前工作已推送c9e65ef，本輪屬實際進展，Goal保持active而未宣稱完成。

知識路由命中retro remake、IDA9.4、compiler／runtime分流、worklist資料與文件職責。
套用grilling核對使用者驗收標準，先讀現行狀態／docs/74與docs/24／25，再查主機Issue #5。
Issue仍OPEN，遠端內容仍為722e778首批紀錄；本Goal未授權改寫Issue，未送遠端留言。

- 新`ida_matching_inventory.py`以IDA9.4從原版5178fdc8建立一次性database，核對MZ relocation後的loaded bytes及全部file bytes。最終inventory-ida-r2匯出828自動函式、29979個指令位置、81854唯一code bytes；未映射與重疊code bytes皆0。自動邊界只是導航，不能直接當C source units。
- 舊280 entry與新IDA共有277，三個舊entry不在新入口，新清單另有551。另有22個非terminal末端及4455個函式外指令位置。已將這些列入完整Goal，而非只沿用最容易通過的280分母。
- 首個C為`re/match/sub_5d49.c`，原始sub_15D49，IDA linear15D49..15D50、logical5D49..5D50、file70B9..70C0。完整操作為DS:0B60寫word1及near return；欄位產品用途unknown，原始table xref289E2保留，沒有命名成未證實旗標。
- 新C批次從唯讀source、固定MSC candidate與IDA明確範圍編譯；單一PUBDEF／SEGDEF及真實FIXUPP把外部word配置到DS:0B60，來源C沒有inline ASM、db或machine-code array。
- `/Ox`第一輪7指令bytes相同，但整個CODE segment多90 NOP，判DIFF。`/Os`兩輪額外生成`xor ax,ax; call __chkstk`，未支持code fixup而REFUSED。回查compiler分流與實際listing後，`/Gs`控制移除該helper，整個module仍8 bytes而原版函式7 bytes。
- compiler `/Fc` listing明確標示PROC offset0..7，ENDP之後才是offset7的NOP。以compiler提供的獨立範圍逐行核對完整原始OMF bytes，包括ENDP外的NOP，再將實際fixup套在同一函式。沒有用RET、NOP內容、原版長度或遮罩猜切點。
- 最終C-batch-r6／r7：完整C PROC1個、7 bytes精確匹配；兩份OBJ、重定位函式／module bytes及listing相同。完整module比較仍DIFF，post-ENDP padding配置unresolved。partial scaffold其餘115275 bytes仍保留原始輸入，不能當整檔完成。
- `matching_goal_contract.json`保存使用者選定的六個驗收gate。完整source-unit覆蓋、全部主程式C、ASM底層分類、data/layout、乾淨整檔重建及兩次重建／runtime都尚未證實。manual gate明列將來所需證據，不用測試名稱、0錯誤或partial hash相同取代。
- 新evidence audit核對所有原始指令、source／producer／semantic ledger新鮮度、compiler PROC及兩次重編；listing的正對照與missing-ENDP／修改post-ENDP byte兩負例均通過。自動清單、source C與完整module的界線分開記錄。
- 全清單第一輪資料已解碼，但IDA環境的ASCII write_text在輸出中文annotation時失敗，空sidecar／log保留。修正為明示UTF-8，用同一image與唯讀原版建立新database乾淨重跑。這是匯出器環境契約問題，沒有更動原版bytes或game語意。

工具與收據沿用gitignored `work/matching-decomp-20261008-r1/full-goal-r1/`，沒有建立新的
同義研究根。新來源與契約均由docs/25及tools/build/README索引；docs/24追加舊分母與
module選段勘誤，保留歷史收據；CONTEXT、PROJECT_MEMORY及唯一plan docs/74更新為完整Goal。
本輪只提交C source、IDA／compiler／audit工具、JSON契約及文件；原版、database、OBJ、COD、
private EXE及protected scratch不提交。提交標題為
`research: establish full matching scope and first exact C procedure`，commit／push以Git log為準。
下一切片先審22個邊界與三個舊entry，再從已閉合資料流還原下一批C及完整layout。

最終verification核對原版hash保持，Python語法與git diff --check通過，正式Go／pack的diff
仍空白；研究根root-owned與Markdown目錄皆0，來源與輸出UID/GID1000。本批IDA／MSC
一次性容器已清除，未動其他專案資源。完整inventory SHA256
`84592c1f36706c1033050e5afd0b6fd7b9cd7773548473141d57c02af6bf0652`，
C receipt SHA256 `9fc31e9f71c52f72885f14ce303eceee451b0989b37fc69ab0f4d60ca5da5e01`，
goal-audit-r2 SHA256 `daecf4458e37b4782aada753a88dccc89c5b3028be98db8b6b1f944be04b97e4`。

### 2026-10-08完整matching續行：邊界分級、間接callback與15-byte C

上一Goal輪有實際進展，0702882已推送；本輪接續使用者已選的完整C／已確認底層ASM標準。
Goal維持active，沒有縮小成局部compatible adapter。先核對工作樹、目前記憶／CONTEXT／
README／docs/74與docs/25；原版115282 bytes／5178fdc8、正式Go／pack與protected scratch保持。
路由命中retro remake、IDA9.4及平台規格優先，沿用既有研究根和工具image。

新增fresh `ida_matching_boundaries.py`與offline `review_matching_boundaries.py`：

- 22候選末端逐項核對原始bytes及typed xref。17個有實際fl_F type21到physical next，不能以自動owner截斷；2個有相鄰MOV AH4C／INT21的DOS service末端；3個因直接callee的FUNC_NORET抑制post-call flow，保持strong。實際flag常數由IDA工具輸出，FUNC_NORET為1，未從數字或名稱猜。
- sub_192F0的callee193E3以194BF→19405形成原始事件循環。16346→1E713→1E7F3→192BC連到DOS清理與退出。保留callee-based限制，不由flag單獨推無返回，亦不把unreachable RET或多entry自動併成C單位。
- 舊182BB在182B8的far-call operand內、1A660在1A65D的MOV內、1A753在1A751 near-call operand內；偏移3／3／2 bytes，fresh查詢皆無incoming xref。排除這三個舊source-function入口；若有獨立重疊入口證據再重開。
- DOS AH4C契約引用RBIL61的標準service，不為平台語意另開RE。核對的是本案實際writer、callsite及轉跳，不深入hardware timer／ISR。
- 首個C的table word在IDA289E2／file19D52為495D。實際取址是14FF2的DS:3BB4，不是鄰近命令表的DS:3BAA。startup192AB／192AE carrier支持DGROUP linear24DD0；在此DS條件下，[DI+4]=47經BL、清BH、SHL BX、SI加BX及零word gate到14FFF CALL[SI]，選中logical5D49。
- DI來源、欄位+4產品意義、table完整合法範圍與runtime DS未驗，保留strong conditional static mapping。沒有猜成NPC handler字段、flag名稱或正常玩家證據。export另保留全部33個indirect call原始operand與typed屬性，未修改database名稱／邊界。

新增C sub_136f5／sub_14a8d，各為DS-relative word0033／000C→AX及near return。
原始定位分別為IDA236F5..236F9／24A8D..24A91，load-image logical136F5／14A8D，
file14A65／15DFD。source使用已驗16-bit unsigned表示，但不宣稱原版C型別、signedness、
DS來源或模組用途；兩個C單位仍屬次級segment未分類的局部資料操作。

C-batch-r9／r10以相同最終producer與fresh boundary-ida-r2來源重編：三個PROC共15 bytes
匹配，兩個新4-byte module亦完整相同；首個7-byte module仍有ENDP後1-byte NOP未配置。
兩份OBJ、實際fixup後的body／module及listing逐byte相同。既有goal audit改為每source unit
分別驗missing-ENDP及post-ENDP／PROC-byte mutation，共六個負例全部拒絕，正對照通過。
沒有為了增加C數量採用raw bytes、inline ASM、mask或未閉合callee prototype。

新來源與原始指令索引由docs/25及tools/build/README收錄；CURRENT表、PROJECT_MEMORY、
docs/74、docs/24及goal contract更新為C3／15 bytes與17／2／3邊界結果。
完整Goal的所有六gate仍未證實，source-unit範圍尚未批准，不用C15 bytes或partial hash冒稱完成。
下一切片從17個實際flow edge建立保留多entry的CFG候選，逐一閉合shared tail及函式外code。

本輪只提交C、query／review工具與文件，private EXE／OBJ／COD／database及13項scratch
不提交。提交標題為`research: resolve boundary evidence and match three C procedures`，
commit／push以Git log為準；Docker清理與最後驗證另列收尾收據。

收尾：C／boundary tools語法、git diff --check及正式Go／assets／dist-all空diff通過。
removed-flow、偽造legacy incoming xref及錯table index三個獨立反例均拒絕；所有新輸出
UID/GID1000，研究根root-owned／Markdown目錄皆0。本批IDA／MSC容器已清除。
fresh boundary SHA256 `b7f6a4d00bfbb959d3e04b856c4025368147f6a3e5c4363c34dce9278f03f7d1`，
boundary review SHA256 `2cd054c149e746f5ac4a3477cfd1ab56c014f1d96966ee2615e037461c62afc2`，
C receipt SHA256 `2d89337301487276e5d6ea268973782f98294f90ccf88674f149e48d3ddbbb4d`，
goal audit SHA256 `00016d5ae21fc300160e9d0595573d3f58dea720b91b73ba2474c574841ef9ee`。

### 2026-10-08完整matching續行：多entry CFG與Watcom16 register ABI

上一Goal輪有實際進展，7cd8ccd已推送；本輪接續使用者已選的完整C精確匹配標準。
先核對worktree與現況入口，知識路由沿用retro remake／IDA9.4，再依compiler實際缺件
載入官方Watcom16 ABI／object dependency契約。所有程序在Docker，原始EXE與repo輸入唯讀。

新matching_cfg_candidates工具跟IDA真實flow／jump，不把call targets當local edge，不補造
被抑制return，也不依auto owner截斷。17 entry共991唯一head，31 shared，候選皆未批准。
第一組10000..10030為14指令／48 bytes，保留10000／1000A／10014／1001E四外部entry；
11900與1196B亦共用suffix。正對照與10007 mid-instruction edge拒絕通過，避免對錯範圍寫C。

主機已有fd2-watcom-matching image僅含wcc386，確認沒有16-bit backend，未因啟動失敗猜缺件。
為16-bit支援建立明確revision，保留其他專案32-bit image。fetch工具下載固定官方
2026-10-01-Build archive129081693 bytes，全檔SHA256a961f3e0通過，762個payload逐檔核對。
外層timeout124發生於完整成功輸出後收尾；獨立hash、size及per-file payload皆通過，原下載
container已停止，沒有以timeout重啟成功下載或宣稱compiler壞掉。

dq3-watcom16:2.0-20261001-r1由固定python3.13-slim OCI digest及已驗payload建置。
支援wcc16／wcc386／wdis／wlink／wlib與h，source／命令入口tools/build/README及docs/25。
compiler實測Version2.0 beta Oct1 2026 64-bit host／16-bit target，正式原版compiler仍unknown。
private compiler、原版、OBJ、source archive、database均不提交，image未發布。

六個C控制只使用instruction-free pragma parm／value／modify exact，不用`=` assembly body：

- BX／SI echo分別編89D8C3／89F0C3，證明可直接指定原版常見register參數，非stack adapter。
- AX store編A33200C3，與原版sub_24AE6完整4 bytes相同；仍是compiler控制，沒有source-unit ledger批准，正式coverage不增加。
- _rotl core24 bytes、有界BX／DX:AX51 bytes、local shift-or core40 bytes，全與原版16／30 bytes不同。保存register、CL旋轉、volatile重讀與雙DIV均保留，不以數值／ABI等價稱exact。
- 首輪無-oi生成真正_rotl call，未知code fixupREFUSED；依官方inline契約加-oi後才解析成功，沒有硬patch原碼或放寬unknown relocation。
- OBJ重現核對兩次失敗均保留：第一次absolute command source path不同，改relative path仍因THEADR的absolute path及dependency timestamp不同。逐record定位後，固定container內/tmp/watcom16-compile並使用官方-zld移除dependency metadata；不mask、strip或canonicalize OBJ。
- 最終watcom16-abi-r7／r8六份完整OBJ、source、resolved code及actual fixup相同，compile每批約0.03秒；不將此當全流程加速倍數。

正式C仍3／15 bytes，既有46 ASM只是局部證據，全部六Goal gate保持未證實、Goal active。
下一批以已驗register-ABI C介面，為多entry主程式資料流寫C候選；source-unit／callee未閉合
時先補證據，不用新的selector、stack bridge或重排entry填洞。正常Go／pack與Issue #4暫停保持。
本批新增source／工具／Dockerfile與穩定索引，沒有新遊戲包、Release或公開原版資料。
提交標題`research: preserve multi-entry CFG and establish 16-bit register ABI compiler`，commit／push以Git log為準。

收尾：CFG四entry正對照／mid-instruction反例、六compiler控制、整OBJ／code／fixup重現、
工具語法及git diff --check通過，正式Go／assets／dist-all diff仍空白。
research根root-owned／Markdown目錄0，payload與輸出UID/GID1000。新Watcom16 --rm容器
全部清除。重建後檢查懸空image，僅見其他專案IDA私有來源4ac62de83339，未刪；
新revision無被取代image，原wcc386-only與其他專案資源保持。
Watcom16最終receipt SHA256
`c5dd97f7e7bf0c0f6ac73074e11f7e185cf0bfc170598f473b6b35758f3dca4d`，
本機驗證收據為full-goal-r1/watcom16-verification-r1.json，CFG收據為cfg-candidates-r1.json。

## 2026-10-08：暫存器 C 來源與 WCC 位移控制

沿用retro remake／IDA9.4知識路由與完整matching契約，將AX入參word store移入明確C來源
及Watcom manifest。原始sub_24AE6的IDA linear24AE6..24AEA、logical14AE6..14AEA、
file15E56..15E5A保持；compiler整個module為A33200C3，沒有inline machine instruction。
DS／module、欄位用途及原版型別仍unknown。正式C由三個15 bytes增為四個19 bytes。

主程式sub_132A3填表及sub_16FCF搜尋也有完整來源與原始範圍清單，仍為DIFF。
填表C17 bytes先INC BX再以symbol-2寫，DEC／JNE與原版ADD／LOOP不同；搜尋C24 bytes
重新分配AH／BL／CX並保存DX，未匹配原版15-byte LODSB loop與隱含register效果。
不把RESOLVED、數學等價或pragma介面當作原碼匹配，不新增正式C覆蓋。

實際WCC OMF raw FFFE由wdis的symbol-2及vendor WLINK synthetic DATA獨立核對。
vendor map frame265F減2產生operand265D；指定原版DS265D則候選operand265B。
兩種配置分開記錄，不由synthetic frame bias推定原版layout。新控制工具保留raw addend與
producer-scoped signed16，預設unsigned及未知frame拒絕維持。六無效位移全部拒絕，
包含越界placement但相加後落回16-bit範圍的反例。

兩次MSC c-batch-r13／r14與全新容器的Watcom watcom-source-r7／r9重編OBJ、code及fixup
一致。r8在同容器固定compile目錄已存在時失敗，保留收據後以新容器重跑；未改來源或flags。
總審核goal-audit-r6核對四個19 bytes、完整IDA輸入、source與producer hash及實際重定位，
六完整Goal gates保持未證實。偽造exact、過期repeat manifest與重複unit三負例皆拒絕；
MSC六listing負例亦保持。具體範圍、工具入口與收據索引在docs/25及tools/build/README。

原版115282 bytes與SHA2565178fdc8保持，正式Go／pack、Issue #4暫停及私有資產保持。
本批只提交C來源、研究工具、清單與現況文件。研究輸出root-owned及Markdown目錄均0，
一次性容器已清除，沒有新image或發行包。下一步追主程式LOOP／LODSB與多entry的C codegen。

## 2026-10-08：主程式 loop codegen 與現成 compiler 邊界

前輪d8fa757為實際進展，依CONTEXT及docs/74接續。路由命中compiler/runtime分流，
載入compiler-runtime-helper-triage及文件職責。原版、既有TC套件與repo唯讀掛載，
所有搜尋、下載、compiler、解析與驗證在有界UID1000 Docker容器。

新增兩個C codegen probe並立即掛到docs/25與tools/build/README。
Watcom八種填表寫法乘九組正式flags為72案，六種搜尋寫法乘九組flags為54案，全部DIFF。
關閉重排恢復store順序，count先宣告恢復CX／BX初始化順序，-ot可發ADD BX,2，仍以
DEC／JNE而非LOOP計數；搜尋最短20 bytes仍未匹配原版15 bytes。
保留每份原始C、OBJ、compiler command、code與真實fixup，不採prefix或masked match。

為避免將loop最佳化選項誤讀成LOOP指令，核對官方固定release的commit e2856866。
原猜8086目錄404，依官方listing改i86；54份codegen C source共935938 bytes逐檔
Git blob SHA1及SHA256通過，快照留work，不加入Git。Do4CXShift有兩個M_LOOP emitter。
新可讀C long-shift正對照實際產生15-byte module與LOOP，證明工具能發該指令；
不從有限source snapshot推定整個backend不支持C迴圈，也不把控制樣本計入C覆蓋。

沿用本機TC2.01 archive及compiler固定hash，以既有dq3-msc image的DOSBox交叉控制。
四個一般register-locals候選因實際frame1／target2 fixup尚未支持而REFUSED；
四個_CX／_BX原生C偽變數候選皆19-byte DIFF，含INC兩次、DEC、OR與JNE。
原版compiler保持unknown，TC2.01身分僅對本機控制工具成立。

初次batch未用CALL，compiler及DONE已完成但停DOS prompt，90秒timeout收據保留。
沿既有已驗runner改call go.bat後同image乾淨退出。之後兩次OBJ不相同，code相同；
逐OMF record定位唯一COMENT classE9 source file time差異。固定實際生成C的mtime至
2026-10-01 UTC後完整OBJ相同，沒有strip、mask或patch object。

最終Watcom primary-codegen-r3／r4的126案加long-shift控制，以及Turbo C
turboc-primary-r4／r5八案，完整source／OBJ／code／fixup或拒絕結果相同。
verification逐案重讀object解析、核對原版hash及54份官方source，語法與輸出UID通過。
正式C仍四個19 bytes，完整Goal六gate仍未證實、active。下一步回原版compiler／ABI
及source-unit歸屬，不重跑這134個已排除組合，不由零match推定所有C compiler皆不可能。

正常Go／game pack及十三項使用者未追蹤資料保持，沒有新image、遊戲包或Release。
本批只提交兩個公開probe、索引與現況，原版與compiler binaries、source快照及OBJ不提交。
research根root-owned及Markdown目錄0，一次性容器全部清除；git diff --check通過。

額外直接核對16份最終TC object的非空E9 payload，時間／日期為0000／5D41，
與固定source mtime相符；不依兩個容器碰巧同時執行推定可重現。首個檢查器未區分短E9
record而unpack失敗，修正為另存短record、不解讀time後乾淨通過。
最新goal-audit-r7再次重核四個19 bytes與六個未完成gates；所有私有收據由docs/25索引。

## 2026-10-08：原版 Creative SDK module 與來源歸屬

上一輪4842c6e已完成codegen控制，屬實際進展。本輪依新現況回到原版工具鏈分流，
載入retro-toolchain-runtime-fingerprinting與技能reference，沿用IDA9.4優先契約。
原版EXE115282 bytes／5178fdc8、SBCM.LIB35840 bytes／01b242cb保持；全部工作在有界UID1000 Docker。

原版SBCM.LIB確為page16 OMF，26個`.ASM` module metadata完整校驗。
新probe保留原始module、PUBLIC、external與fixup，九個排除重定位的導航候選不計source match。
LIDATA最初REFUSED，獨立USE16 decoder保留原始OBJ並建立parser view；重複資料的fixup仍拒絕。
CMFDRV有226個未寫入bytes，沒有用parser預填零冒充完整match。

官方WLINK從原始OMF重連CTVMEM2493及CMFDRV5296 bytes，完整code／data／gap均等於EXE。
兩次完整vendor MZ相同，合計7789 bytes只算module身分，source coverage增量0。
本輪commentary曾將合計誤寫7779，已按2493＋5296更正，不改個別長度或原始bytes。
source、data語意、code／data分類與ABI仍需還原，原始OBJ不作最終source build輸入。

fresh IDA保留兩段80自動函式、2175 code heads、原始names／bytes／chunks，無MZ relocation。
driver入口17／19 typed far-call refs及DSP／timer／PIC I/O與PUBLIC身分一致。
第一次腳本的ida_bytes.BADADDR不適用，error sidecar保留；改idc.BADADDR後新DB副本重跑。
所有原始定位保持，SDK名稱只作外部metadata，不改正式database或名稱。

三個已match C helpers的歸屬閉合：sub_236F5在CTVMEM，sub_24A8D／sub_24AE6在CMFDRV，
共12 bytes。更新manifest推論範圍，runtime DS／欄位／原始型別仍unknown。
兩份MSC及兩個新Watcom容器重編source／OBJ／code／fixup一致，C仍四個19 bytes。
其餘7-byte C在MZ入口code segment，產品角色仍未知。goal-audit-r8六gate仍未證實。
原版Press X已找到另一段DOS consumer，無MSC版本／runtime map證據；docs/19追加勘誤保留歷史。

library截斷、checksum、LIDATA expansion、repeated-data fixup及self-fixup越界五負例拒絕。
source與producer新鮮度、原始names／chunks／bytes及完整vendor MZ核對通過；收據在docs/25。
原版compiler與其他八個SDK候選仍未閉合；下一步兩個已識別driver的可讀ASM／data source spec，
不重跑134組，不把linked object當原碼完成。正常Go／pack、原版、十三項scratch及Issue #4暫停保持。

使用者要求matching產生的正確程式碼存GitHub。已核對四C與兩ASM來源在origin/main；
新增re/match/README列驗證範圍、重建工具與DIFF候選限制，掛入docs/25與CONTEXT索引。
本批只提交公開source／工具／manifest／文件，原版binary、OBJ、SDK及IDA資料庫不提交。
Docker一次性容器清除，研究root-owned與Markdown目錄0，git diff --check通過。

發布入口核對另以NASM重新組譯兩個既有語意ASM來源，完整16／30 bytes仍等於原版；
四個C19 bytes以新manifest的兩次收據重核。README所有source及重建連結存在，新增三工具
語法、索引及UID/GID通過，收據source-publish-verification-r1由docs/25索引。
第一次README patch同檔重複operation而全批未寫入，合併operation後完成，不屬產品問題。

## 2026-10-08：SDK 語意指令來源與原版資料省略

前輪9c8d1ee已識別SDK及發布source入口，屬實際進展。依唯一現況表回到兩driver source，
路由命中RE→spec並載入retro-remake-spec-gated-workflow；原版及repo分析輸入唯讀，全部Docker。
fresh IDA匯出operand text與四個未覆蓋區域，未改原始names／bytes／function boundaries。
CTVMEM2258及CMFDRV2890 bytes為原始instructions，其餘235／2406 bytes仍待data／code role審查。

先做可丟棄DRAFT ASM prototype。NASM2.16不能用猜測的load decorator且register encoding不同；
沿已驗證完整官方archive取Wasm，控制樣本符合原版MASM shape。
前兩prototype因implicit operands與absolute memory語法失敗，後兩因displacement及accumulator
encoding DIFF，全部保留。語意source沒有opcode byte emission；normal external absolute EQU
經真正WLINK fixup保留原來16-bit width，XCHG對稱operand次序保留原operand註解。
REP與segment prefix、LOOP／MUL／DIV／string隱含operands依IDA及compiler實測訂正。

prototype-r6整段7789 bytes相同，但含未審data literals，DRAFT及source增量0保持。
正式新增的四份SDK instruction／EQU source移除所有data literals，以ORG省略未知區域。
原始objects不作build輸入，source無DB／DW／DD／INCBIN／include／macro，原始IDA位置及operands保留。
verifier從repo文字source組譯、真正link，actual OMF written mask恰等於declared指令區段；
逐每個原始byte核對5148 bytes。driver資料2641 bytes、完整ABI與全source未完成，產生MZ不能作完整driver。

新prepare_watcom16_asm從immutable r1 clone官方payload加入Wasm，763檔逐hash通過，
同固定runtime及Dockerfile形成dq3-watcom16:2.0-20261001-r2，沒有host runtime或原版inputs。
r1 C控制image保留，不重抓成功archive或改其他專案image。Wasm binary hash7e216ab5固定。
最終sdk-source-build-r3／r4在新容器重編，完整source／OBJ／MZ／FIXUPP receipts相同。
七種raw data／label後data／TIMES／include／INCBIN／macro／word emission負例拒絕。

12 bytes與三個C helper重疊，新唯一instruction source為5136 bytes。既有C19及RNG ASM46保持，
局部所有唯一source-instruction reproduction為5201 bytes，不升格完整module、source-unit gate或campaign。
使用者要求正確source存GitHub，新增四ASM／EQU、manifest及verifier，README與docs/25立即掛索引。
完整driver spec仍DRAFT；下一步四個data gaps、typed data與dispatcher／buffer證據，主程式C仍未完成。

原版SDK／EXE、prototype內data、OBJ／MZ／IDA database／vendor payload均不提交。
正常Go／game pack、十三項scratch及Issue #4暫停保持；語法、連結、source／written-mask
驗證與git diff --check通過。一次性container清除，研究root-owned及Markdown目錄0。

## 2026-10-09：SDK 指令 source 發布收尾

工作跨日後依新的environment日期更新唯一現況表，不改原Goal選擇日期。
最終兩份source receipts相同，5148 instruction bytes與12-byte C重疊已核對；
goal-audit-r9維持四C19 bytes及六個未完成gates。source-publish收據核對新script語法、
README／索引連結、全source UID/GID及研究0 root-owned／0 Markdown目錄。
Docker批次容器全部清除；懸空image只有其他專案私有IDA來源4ac62de83339，保留。
新Wasm r2沒有被取代的本專案image，r1仍供已鎖定C控制使用。沒有新遊戲包。
正確partial instruction source與其明示限制提交GitHub，原版data及prototype不提交。

首個725522b提交前cached diff檢查抓到新ASM行尾空白，orchestration未在exit2後停止，
後續仍提交及push，已向使用者說明。只修本輪四source的行尾空白，sdk-source-build-r5／r6
新容器重編source／OBJ／MZ／fixup一致，5148 bytes保持；新source hash由publish-verification-r2核對。
補提交前明示檢查exit code後才提交，不amend、不覆寫已推送歷史。

## 2026-10-09：CMF 完整 source/data byte-layout 與 CTV DMA seed

前輪0d5fe58新增5136唯一instruction source，屬實際進展。本輪依CONTEXT／docs/74追四data gaps，
載入RE→spec閘門；DMA／DSP writer需求再命中retro-platform-spec-first，依成熟契約停止硬體時序RE。
原版SDK／EXE與repository分析輸入唯讀，UID1000、bounded Docker、明確source／output寫入。

seven table consumer閉合：CTV API14／stream8，CMF status8／controller4／channel16／API15／negative3。
資料值、consumer位置與handler原始IDA頭一致。CTV API slot6初值6B06超界，未將它改成合理pointer。
原始init把CS:91轉實體DMA目的地，連續送E2＋06及E2＋6B；直接xref不足以看到DMA writer。
固定DOSBox Staging commit d8271efb、整個soundblaster.cpp snapshot SHA314dfb06保存本機，
從實際source解析reset AA／count0與E2 table得到DMA3A／08、word083A，與原2373A比較值一致。
記為platform-contract derivation，不稱原版實機、DAC waveform或wall-clock parity，不開新timer slice。

CMF資料角色最小充分：signature、原始header/state初值、五個near handler tables、
parameter/index bytes及IRQ reserve。SS=CS／SP1337的writer與原100 zero bytes確定50-word stack幾何。
字段用途與原始型別未知者保留unknown，初值只取定hash的原始bytes，不猜0或C remake。
typed-data DRAFT prototype完整5296 bytes等於原版；docs/25先完成layout-only READY審查，再正式source。

新增cmfdrv_data.json及verifier。build輸入僅repo ASM／EQU／typed JSON，不用原始OBJ或EXE code，
EXE只在verification比較；data只能寫入兩個approved non-code範圍，不能跨code、overlap或缺值。
normal WAsm／WLink重建全部5296 bytes，actual segment written mask完全填滿，含真實symbol relocations。
最終cmf-source-module-r3／r4兩個新容器source／generated ASM／OBJ／MZ receipts一致，2406 data bytes新還原。
缺field、overlap、value越界、落code、錯handler、未READY及字串注入七負例拒絕，scope限完整byte-layout。
SDK data reviewer也重新從pinned platform source解析，沒有只信receipt推導欄位。

CMF spec轉CONFORMED，但原data semantics／完整driver ABI／hardware runtime仍unknown。
CTV235 bytes仍待審，11-byte未命名payload不自動當已完成source；主程式C與全EXE六gate仍未完成。
既有C19／RNG46／SDK instruction5148及12-byte重疊保持，新增2406是data不是額外instructions。
正確source按使用者要求提交GitHub，原版assets、vendor source snapshot、OBJ／MZ／IDA資料庫不提交。
正常Go／game pack、Issue #4暫停及十三項scratch保持；本輪source／schema與最終diff核對後commit＋push。

收尾cmf-source-publish-verification-r1重核最終r3／r4 source hashes及七負例，README／索引／語法／
UID/GID通過，research root-owned及Markdown目錄0；goal-audit-r10維持四C19 bytes及六個未完成gates。
source group和原data區段完全分離，未使用原始objects；CMF data2406 bytes不重算instructions。
route再查仍命中spec gate及platform-spec-first，snapshot hash與派生等級保留。一次性container清除，
其他專案新增的wizardry／wolong執行container保持，沒有新image或發行包。

## 2026-10-09：CTV 完整 source/data 與 SDK builder 合併

上一輪efacf96完整CMF byte-layout source5296 bytes已通過，屬實際進展。本輪接CTV235 data，
路由仍命中RE→spec，讀取當前真相及直接證據，全部bounded UID1000 Docker。
新IDA minimal query確認236A9的11-byte區域為data、byte array、非code／unknown／align；
前RET與後exported dispatcher分離、known table不指向此區。purpose仍unknown，僅strong static layout。
typed初值不猜修、不把缺xref當永不使用，原始fields/type及runtime硬體parity不升格。

先於docs/25審CTV layout READY，新增ctvmem_data.json：兩個header字串、初值、固定copyright、
API14／stream8表、mutable state及unknown11-byte payload，data範圍只准0003..00E3及0749..0754。
第一次copyright起點錯放0038，strict string guard拒絕；依原file14309／module0039調整field分段，
全部235原始bytes不變。E2 slot6仍literal6B06，不替換成推導target083A。

既有CMF builder抽到common verify_sdk_source_module，profiles鎖CTVMEM／CMFDRV範圍與source。
舊verify_cmf_source_module保留CLI及render_data相容，不重複兩個流程。共同written mask、
全段byte compare及data schema保持，ASM／EQU／typed JSON是唯一build input，原EXE只comparison。
初次common patch同檔Delete/Add重複operation被工具拒絕，未套入；拆為單次file edit後完成。
後一次文件patch漏context prefix全批拒絕，修正後完成，未修改原source語意或原始binary。

CTV完整2493 bytes重建成功，兩份最終r4／r5 source／generated ASM／OBJ／MZ相同；
CMF舊CLI最終r7／r8回歸5296 bytes相同，合計兩SDK7789 bytes已由完整source恢復。
九CTV與七CMF共16負例通過。seed-to-label原schema先未拒絕而whole-byte compare才會拒絕，
保留draft收據後加strict literal slot guard，現在schema即拒絕，不改初值。
CTV spec轉CONFORMED，scope限byte-layout；purpose／原types／完整ABI／原硬體runtime仍unknown。

source已依使用者要求保存repo，README／docs/25／build入口同步；原SDKobjects、EXE、probe DB、
產物與私有data snapshots不提交。原C19／RNG46／SDK instruction5148及C重疊12保持；
unique所有known source bytes為7789＋7＋46=7842，不能當全EXE完成或已審source-unit比例。
下一步主程式C、其他frame/source-unit與完整layout，六Goal gates仍未證實、active。
正常Go／pack、Issue #4暫停與十三項scratch保持，沒有新image或發行包。

收尾sdk-full-source-publish-verification-r1核對兩份CTV及CMF最終source／producer hash、schema
與README／build連結、UID/GID，語法通過，research root-owned／Markdown目錄均0。
goal-audit-r11維持四C19 bytes及六個未完成gates，不由module完成聲稱原版campaign或全EXE。
一次性source／IDA容器清除，其他專案pto2／supabase等container保持。

## 2026-10-09：主程式角色指標查詢 C 精確匹配

依目前狀態從 SDK source 返回主程式 C。路由命中 compiler-runtime-helper-triage，
保留原版 compiler unknown；使用 use-ida-pro-9-4 技能，沿用已驗證 py312 image並核對新sidecar。
原始 EXE／database 唯讀，DB 工作副本與產物只放 ignored work，所有分析與重編在有界 UID1000 Docker。

選原始 sub_19834，linear19834..19842、logical9834..9842、fileABA4..ABB2，共14 bytes。
純 C 宣告 volatile index與word table；instruction-free pragma指定SI返回及保存暫存器。
初次私有prototype用-of已匹配；正式既有-ofr runner同樣完整匹配，沒有改producer或加codegen特例。
新source sub_9834.c納入Watcom manifest，不使用ASM指令、code array、raw bytes或code patch。

fresh IDA9.4完整chunk及56近caller保留；抽樣10B1D、139D6、13E5A的返回SI consumer。
139D2索引writer與139D9加3A／掃八word閉合；115C2、105F9及10613為間接table writers。
不由table直接xref零命中宣稱沒有writer；186A5取址指令沒有IDA function boundary，
仍列unresolved target，未將它當完整初始化函式。原始C型別、table完整範圍與所有索引合法性unknown。
角色指標語意與docs/171一致，本輪沒有新增正常campaign或遊戲畫面收據。

正式watcom-source-r12／r13在兩個新容器重編，完整14-byte module、source、OBJ與兩個OMF fixup同值。
goal-audit-r12核對來源新鮮度、完整IDA範圍、實際OBJ及兩份重編，五個C共33 bytes，
其中主程式21、SDK12；既有fill17／search24仍DIFF，六Goal gate仍未證實，Goal active。
SDK7789、主程式21及RNG46的unique已知source bytes為7856，只屬局部重建統計。

正確C與manifest、原始定位／hash／限制寫入來源索引與docs/25，依使用者要求提交GitHub。
原版EXE／SDK／OBJ／code.bin／IDA資料庫及全部使用者scratch不提交；Go／pack與Issue #4保持。

primary-table-source-publish-r1核對完整來源／OBJ／code hash、兩份重編、來源索引與連結、
UID/GID及五C33 bytes。研究root-owned與Markdown目錄均0，git diff --check通過。
本輪一次性容器全部清除，其他專案container保持；沒有新image或發行包。

## 2026-10-09：主程式物品視窗記錄 writer 的48-byte C

上一輪a4749bf新增14-byte角色指標查詢並已推送，屬實際進展。本輪沿主程式C，
路由仍命中compiler-runtime-helper-triage，保留原版compiler unknown，不改既有產品或驗收標準。
全部分析／compiler／IDA在bounded UID1000 Docker；原EXE及正式DB唯讀，副本放ignored work。

sub_137F9為完整單一chunk48 bytes，SI記錄pointer、BX word入參。最初C prototype42 bytes，
缺AX/CX保存，ADD AX,2被size最佳化成兩次INC；未計source match，也未拿等價結果冒充exact。
官方Watcom pragma語法與實際callee對照顯示default convention影響保存集合：watcall保存CX，
void指定非AX value register後亦保存AX；速度最佳化發ADD，停用重排保留原始讀寫順序。
新的純C prototype完整48 bytes匹配，沒有內嵌組語、code array、raw bytes或opcode patch。

fresh IDA9.4保留兩個近caller、原始chunk／bytes／operand／typed xref。兩caller均寫DS:072F=43、
0731=46、0733=BX，SI取DS:3FD8後呼叫；caller auto owner缺失保持，不改函式邊界。
177FE呼叫視窗consumer1F590，再由1FC57讀幾何。新1F690取座標與row count判定選人，
七個writer word均有原始consumer。用途與docs/188正常物品窗口證據一致，原始struct未知。
靜態consumer閉合不新增正常玩家路徑或畫面V3聲明。

新增sub_37f9.c及manifest compile_profile。runner固定兩組明列旗標，舊profile不改，
新watcall-speed-no-reorder使用-ot/-of/-ecw，其餘DOS／small／8086旗標維持。
未知profile拒絕，audit核對實際compiler_command及兩份獨立收據的command一致性。
正式watcom-source-r14／r15兩個新容器重編，source／完整OBJ／48-byte module／三FIXUPP相同；
其餘既有source不變，fill17與search24仍DIFF。goal-audit-r13核對六C81 bytes，主程式69／SDK12。
兩種command竄改及未審profile三負例拒絕；六完整Goal gate仍未證實，Goal active。

使用者要求正確原碼放GitHub，本輪C／profile／重建工具與證據索引一起提交。
SDK7789＋主程式69＋RNG46=7904已知unique source bytes，僅局部來源統計，不報全程式完成比例。
原binary、vendor工具、OBJ、code.bin、DB與十三項scratch不提交，Go／pack、Issue #4保持。
下一步帶呼叫的主程式C與near／far symbol及frame定位，再回完整data／MZ／layout。

primary-record-source-publish-r1另核對新source hash、兩份48-byte完整module、三profile負例、
README／build連結與新增Python語法；既有來源的完整OBJ／code／fixup及compiler command均未變。
研究root-owned與Markdown目錄均0，UID/GID1000及git diff --check通過。
本輪一次性容器全部清除，其他專案container保持；沒有新image或發行包。

## 2026-10-09：主程式近／遠呼叫 C 與MZ relocation

上一輪e7160e7新增48-byte視窗記錄writer並已推送，屬實際進展。本輪依唯一現況追帶呼叫C，
路由命中compiler-runtime-helper-triage；原版compiler與完整source-unit分類仍unknown。
全部分析／C／ASM／WLINK／IDA在bounded UID1000 Docker；原binary及正式DB唯讀，副本只放ignored work。

先以可讀C的near／far CALL及三次word store作合成control，沒有DQ3輸入，source增量0。
TIS OMF 1.1與actual WCC確認F5／T2，near為self-relative location1，far為segment-relative location3。
第一個WLINK prototype的AUTO grouping把far CALL改成push CS／near CALL等長序列，
依官方FARCALLS契約改用explicit FARGROUP及NOFARCALLS，不patch輸出或當成原版compiler證據。
兩種object order提供forward／backward near displacement，addresses由actual map取得，保留DS frame bias。
最初兩ASM object的dependency mtime不同；沿既有Wasm -zld禁用該記錄，完整OBJ重新比較，沒有mask。

新增omf_call_fixups及verify_watcom_call_fixups。DS／near／far symbol互斥，地址MZ-relative，
caller不跨64KiB，near target同frame，call只准已審zero addend，另外回報far segment word的MZ位置。
15非法frame／target／mode／location／placement／addend／overlap／opcode反例全部拒絕。
最終call-fixup-control-r7／r8兩個新容器的C／ASM source、完整OBJ、兩完整MZ與relocations相同。
原版caller整合只准已審初始CODE frame0；不由其他IDA segment範圍猜其CS frame，擴充須另匯ledger。

純C sub_3016以四個近call及return重建原sub_13016全部13 bytes；-oc保留CALL／RET，不內嵌指令。
純C sub_ee19把DS:25D1 word傳DX，far CALL109C:007A後near return，共10 bytes相同。
原farcalleeIDA20A3A的ES=DS／AX1012／BX0／CX16／INT10及RETF保留，硬體語意不由自動註解升格。
較早98B9條件式C candidate仍DIFF，未加入正式source manifest或coverage。
private診斷曾誤取far wrapper的file offset，立即改以MZ header+logical重算10189..10193；
正式source／manifest／IDA與原始bytes均使用重核範圍，不把誤取資料作證據。

fresh IDA9.4七函式匯出保留原始names／bytes／chunks／typed xrefs；1300D是JZ entry、135AF是near CALL，
不能都稱caller。caller寫DS:256A，首callee寫DS:0B24，後三callee讀其state；callee source仍未還原。
1C026原near caller與DX入參及BIOS register consumer已核對，原始C prototype／完整ABI仍unknown。
來源用保守clobber宣告，不由C簽名假定原callee保存暫存器。

新增cdecl-size-calls固定profile、typed call_placements及共用resolver整合；原兩profile與DS parser不改。
原caller位置、CS0 frame、原typed near／far xref及原MZ segment relocation table均由audit核對。
最終watcom-source-r18／r19兩新容器source／完整OBJ／整個module／FIXUPP一致，far module+7亦match MZ。
goal-audit-r15為八C104 bytes，主程式五個92、SDK三個12；六個完整Goal gate仍未證實，active。
五audit反例拒絕，其中協同位移與CS alias即使bytes相同仍失敗，不以byte偶合取代original定位。
SDK7789＋主程式92＋RNG46=7927已知unique source bytes，僅局部統計，不報完成比例。

正確兩C與重建／驗證工具、原定位／hash／限制按使用者要求提交GitHub。原EXE／vendor objects／
OBJ／MZ／DB／原素材與十三scratch不提交，Go／pack、Issue #4保持，沒有新發行包。

primary-call-source-publish-r1重核final r18／r19及control r7／r8的source／producer新鮮度、
完整OBJ／MZ一致、15＋5負例、原source objects保持、MZ positions、索引／連結與Python語法。
研究root-owned及Markdown目錄0，全部source UID/GID1000；git diff --check通過。
本輪一次性container全清，其他專案container保持；沒有新image或發行包。

## 2026-10-09：frame ledger與typed MZ metadata

上批4768928的near／far C來源已提交並推送，remote main已核對。C仍八個104 bytes，屬實際進展。
本輪新ABI loop controls均DIFF；既有C source與compiler-generated ASM的symbolic prototype以三種語意變換
生成17-byte counted loop相同。build不讀原EXE、沒有raw code arrays；最初linker group frame bias差16，
保留WCC F5 relocation政策後重建匹配。不把prototype當正式C coverage。
依共同決策規則使用grilling，已展示具體prototype及保留主程式C／全EXE exact的兩個工具鏈選項；
自訂編譯階段分支待使用者回答，沒有採用或寫入正式build。不是再次詢問已選C／ASM標準。

獨立進行original frame證據。原DB工作副本在目前py312 image的input identity為None，輸出不成立；
原DB hash保持，未推定遊戲問題。fresh原EXE loader輸入hash／IDA9.4／bytes重新核對後有效。
新增ida_matching_frames：保留原始names、linear／file／MZ-relative與selector base，兩份完整JSON相同。
19 segment、1199 original direct far CALL均有MZ segment word及typed xref；target frame aliases0。
runtime CS／DS、indirect targets及source-unit邊界仍unknown，不放寬原caller初始CODE限制。

原MZ header4976 bytes、1232 relocation targets全部排序且唯一，raw pairs均採64KiB location windows。
表頭宣告115280、實檔115282，最後u16=0；purpose及DOS實際讀入行為unknown，不猜修page欄位或刪尾字。
新增mz_layout typed JSON及verify_mz_layout_source，header／file-end只由metadata fields產生，
原EXE在生成後作comparison。兩個新容器完整4976＋2 bytes同值，12結構／range／位置負例拒絕；
render在original input path無效時結果保持。原program code／body沒有生成或拼接。
source byte-layout metadata符合Microsoft EXE.INC的前14-word格式，額外3word維持unknown。
已知unique source bytes7927＋4978=12905，只屬局部source reproduction，完整Goal仍active、六gate未證實。
C104與SDK7789保持，正常Go／pack、Issue #4與十三scratch不改，沒有新image或發行包。

frame-and-mz-source-publish-r1核對兩份完整frame exports、MZ typed source／encoder新鮮度、
兩份header／file-end及12負例，source renderer無original path依賴，syntax／索引／連結／UID通過。
goal-audit-r16仍為八C104與六未完成gates；symbolic compiler prototype未採用、不計正式coverage。
研究root-owned及Markdown目錄0；本輪一次性container清除，其他專案container保持。
提交前git diff --check通過，source／metadata／重建script存GitHub，原EXE／DB／artifact不提交。

## 2026-10-09：底層顯示region語意ASM來源

上輪744cb2b的frame／MZ來源已推送，屬實際進展。自訂compiler階段仍待使用者回答，
未採用prototype，不改C104。獨立處理已允許的底層服務，路由命中compiler/runtime分流。
原binary與repo輸入唯讀，fresh IDA／組譯／link／比較均bounded UID1000 Docker。

新ida_matching_video匯出原seg005 linear209CE..20B60全部209 heads／401指令bytes、16 auto functions，
與未入function的page1／single-DAC／RET entry。20B4A共用20B0A尾段，所有entry及原names保留。
原CS frame109C、entry offsetE；source region不是已確認original object邊界。MZ relocations0。
唯一byte20B5F為IDA data／align、非unknown、value0、無xref，layout保存，不由缺xref稱永不使用。

每個unit為BIOS INT10、VGA port I/O、BDA或incoming ES memory service，低層資格confirmed。
DOSBox-X官方INT10 code作palette／DAC／page／refresh介面來源；原operand／register setup才是原版證據。
DS25B4與25F1保留raw offset及unknown field，不替換成推測用途；沒有game-state writer或遊戲規則。
不開retrace cycle／wall-clock研究，不稱完整原版顯示或timing parity。

DRAFT semantic ASM prototype401 bytes完全match，Wasm方向編碼符合原instruction；沒有code byte arrays。
先於docs/25審READY，再保存video_bios_code.asm／typed alignment JSON與verify_video_source。
ASM禁止DB／DW／DD／INCBIN／import／macro；唯一DB0只由reviewed JSON生成在最後的non-code item。
actual OMF code segment全written402，source／generated ASM／完整OBJ／MZ兩份同值，11非法source／layout負例拒絕。
region spec轉CONFORMED，scope限低層資格與完整byte-layout，object identity／完整ABI／hardware runtime仍unknown。
原始EXE只comparison，original objects不進build，不把產生的MZ當可正常啟動的完整driver／遊戲。

已知unique source bytes12905＋402=13307，C仍8／104；完整Goal六gate未證實，active。
按使用者要求保存正確source、typed data、tools與證據索引到GitHub；原EXE／DB／artifact不提交。
正常Go／pack、Issue #4及十三scratch保持，沒有新image或發行包。

video-source-publish-r1重核兩份source／generated ASM／OBJ／MZ、401＋1 range與11負例、
fresh IDA input／producer hash、Python syntax、source索引／連結／UID/GID；研究root-owned及Markdown目錄0。
goal-audit-r17保持八C104及六未完成gates。一次性容器清除，未動其他專案containers。
提交前git diff --check通過，僅source／typed data／tools與證據記錄提交GitHub，原binary／DB／artifact不提交。
