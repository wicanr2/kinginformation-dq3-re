# 188 — 開場連續演出勘誤與修正規格：家中 → 王城入口 → 國王

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
