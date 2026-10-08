# 逐函式 byte-match 進度 (matching decompilation)

已提交GitHub的匹配來源與各自驗證範圍見[re/match來源清單](../re/match/README.md)。
研究候選、原始object重連與完整source build分開記錄；尚未匹配的來源不計完成。
新增SDK指令來源：[ctvmem_code.asm](../re/match/ctvmem_code.asm)及
[cmfdrv_code.asm](../re/match/cmfdrv_code.asm)，encoding常數來源為
[ctvmem_constants.asm](../re/match/ctvmem_constants.asm)與
[cmfdrv_constants.asm](../re/match/cmfdrv_constants.asm)。它們只還原已驗證instruction bytes，
ORG省略的資料缺口不計source coverage，完整driver source仍未完成。
[`tools/prepare_watcom16_asm.py`](../tools/prepare_watcom16_asm.py)從已驗證完整官方archive
建立r2 payload，新增Wasm，保留r1 inputs不變；沿用同一Dockerfile與固定runtime。
[`tools/sdk_instruction_source_manifest.json`](../tools/sdk_instruction_source_manifest.json)列已驗證
instruction範圍與省略data；[`tools/verify_sdk_instruction_sources.py`](../tools/verify_sdk_instruction_sources.py)
從提交的ASM／EQU source重建，核對實際OMF written mask、FIXUPP與每個原始instruction byte。
[`tools/review_sdk_data_regions.py`](../tools/review_sdk_data_regions.py)核對七個handler tables、
CTV DMA seed與CMF IRQ stack的原始初值／consumer，保留未知payload及平台契約推導等級。
CMF完整byte-layout的typed來源為[cmfdrv_data.json](../re/match/cmfdrv_data.json)，
[verify_cmf_source_module.py](../tools/verify_cmf_source_module.py)從ASM／EQU／JSON完整重建；
data只能落在兩個已審non-code區域，fields含未知語意，沒有機器指令data array。
CTV typed初值來源為[ctvmem_data.json](../re/match/ctvmem_data.json)，unknown payload及DMA seed
原樣保留；完整source recipe仍按本頁CTV byte-layout spec驗證，不能由保存值推定字段語意。
[verify_sdk_source_module.py](../tools/verify_sdk_source_module.py)共用嚴格type／range／written-mask
及全段byte比較，選CTVMEM或CMFDRV；原CMF CLI保留相容入口，不改原source或資料初值。

> 2026-10-08新增完整Goal：使用者指定「完成 dq3 matching decompliation」，並選定
> 主程式以C精確匹配，組語僅用於已確認的底層常式。最終由原碼乾淨重建整個EXE並
> 逐byte一致；`db`／原始機器碼拼接不算完成。這是本Goal的驗收標準，先前局部研究
> 與adapter僅是基礎證據。舊280函式清單須與IDA9.4完整清單重核，不能當作完整分母。
> 首個新C函式[`re/match/sub_5d49.c`](../re/match/sub_5d49.c)的完整7-byte PROC已匹配；
> 模組末端1-byte NOP位於compiler `ENDP`之外，配置仍未解。完整Goal保持未完成。
> 現行C本體4個／19 bytes，含兩個次級segment word reader及一個AX入參word store。
> 三個次級C來源共12 bytes已定位在原版CTVMEM／CMFDRV音效SDK；欄位、原始型別與DS脈絡unknown。
> 其餘7-byte C來源位於MZ入口code segment，產品角色仍未確認。主程式填表與搜尋C候選仍不匹配。
> C候選清單為[`tools/matching_c_manifest.json`](../tools/matching_c_manifest.json)，
> [`tools/run_matching_c_batch.py`](../tools/run_matching_c_batch.py)核對完整IDA函式邊界、
> 原始bytes與實際OMF重定位，再以新編artifact計算C覆蓋率。保留原始bytes另列，不能算完成。
> 完整導航清單由[`tools/ida_matching_inventory.py`](../tools/ida_matching_inventory.py)
> 匯出所有IDA code heads、函式chunks、typed xref、MZ relocation與舊清單差異。
> 自動函式邊界與資料／程式分類仍需審查，不由自動清單宣稱全程式已解讀。
> 首個C候選的compiler listing已證實NOP在`ENDP`之外；C批次現在以實際`PROC／ENDP`
> 核對函式本體，並另列完整module比較及post-ENDP bytes。模組padding配置尚未解決，
> 不由函式本體匹配宣稱模組或整檔原碼重建完成。
> 本Goal的驗收條件保存於[`tools/matching_goal_contract.json`](../tools/matching_goal_contract.json)，
> 尚無完整source-unit及build receipts的gate保持manual／未完成；不由既有測試名稱推定完成。
> [`tools/matching_goal_audit.py`](../tools/matching_goal_audit.py)核對原始bytes、清單與C
> 重編的新鮮度、兩份獨立artifact及listing正反對照，並明列仍未證實的完整Goal gates。
> [`tools/ida_matching_boundaries.py`](../tools/ida_matching_boundaries.py)在新database核對22個
> 末端、三個舊entry、實際IDA旗標及間接call operands；不改database邊界或原始名稱。
> 同一份新IDA查詢亦核對兩個次級code segment的word讀取常式；C候選
> [`sub_136f5.c`](../re/match/sub_136f5.c)與[`sub_14a8d.c`](../re/match/sub_14a8d.c)
> 保留未知DS脈絡與欄位用途，不由名稱宣稱模組分類或原版C型別。
> [`tools/review_matching_boundaries.py`](../tools/review_matching_boundaries.py)以fresh typed xref
> 及實際DOS service writer分級22個末端，另核對三個舊entry與首個C的SI table consumer。
> 分級不自動批准C source-unit範圍；table mapping保留DS條件與未知DI／欄位來源。
> [`tools/matching_cfg_candidates.py`](../tools/matching_cfg_candidates.py)沿原始typed flow／jump
> 建立17個跨界CFG候選，保留全部外部entry與共享指令、未解call-return及indirect effects；
> 不依auto owner截斷，也不補造被IDA抑制的返回邊。候選不自動當作可編譯C source unit。
> [`tools/fetch_watcom16.py`](../tools/fetch_watcom16.py)以固定官方release及完整archive hash
> 補足既有Watcom image缺少的16-bit wcc，供instruction-free register-ABI pragma對照。
> 原版compiler仍unknown，不以新工具能編C就反推原版工具鏈。
> [`tools/run_watcom16_abi.py`](../tools/run_watcom16_abi.py)以官方instruction-free ABI pragma
> 測BX／SI入參與AX存word、RNG core及BX／DX:AX控制；raw OMF／未知fixup保留並拒絕。
> Watcom source candidates由[`tools/watcom_matching_manifest.json`](../tools/watcom_matching_manifest.json)
> 定位，來源為[`sub_14ae6.c`](../re/match/sub_14ae6.c)、
> [`sub_32a3_watcom.c`](../re/match/sub_32a3_watcom.c)與
> [`sub_6fcf_watcom.c`](../re/match/sub_6fcf_watcom.c)。原始register side effects不由C簽名省略。
> OMF的signed16 implicit addend僅由明示producer contract啟用；保留raw word與signed值，
> 最終offset仍需0..FFFF。既有unsigned模式與未知placement／frame拒絕保持，不用自動wrap猜語意。
> [`tools/verify_watcom_signed_fixup.py`](../tools/verify_watcom_signed_fixup.py)以vendor WLINK與
> synthetic DATA正對照核對symbol-minus-two，並拒絕unsigned overflow、signed underflow、
> 越界placement、缺placement與未知frame。此控制不提供原版EXE的module layout。
> [`tools/probe_watcom_primary_codegen.py`](../tools/probe_watcom_primary_codegen.py)固定兩個主程式
> 原始範圍，比較C迴圈寫法及compiler的重排／loop選項。每個source／OBJ／fixup均保留，
> 不patch code，不把實驗候選自動加進正式coverage。
> [`tools/probe_turboc_primary_codegen.py`](../tools/probe_turboc_primary_codegen.py)沿用既有
> TC2.01 archive與DOSBox image，固定hash比較一般register locals及C暫存器偽變數；
> 未支援OMF group frame仍拒絕。它不證明原版compiler身分。
> [`tools/probe_sbcm_modules.py`](../tools/probe_sbcm_modules.py)以原版SBCM.LIB的OMF module／
> PUBLIC metadata建立導航候選，保留IDA原名及兩份binary hash。排除relocation的搜尋僅作線索，
> 不當作source match、完整重定位或低層ASM資格證據。
> [`tools/ida_matching_module_refs.py`](../tools/ida_matching_module_refs.py)從全段vendor link證據
> 匯出原版SDK module的原始functions／bytes／typed xrefs及啟動碼，別名僅作metadata，
> 不rename或修改database邊界，不由字串推定全程式compiler。
> [`tools/link_sbcm_module_controls.py`](../tools/link_sbcm_module_controls.py)以固定官方WLINK
> 重連原版OMF objects，核對整段code／data／gap bytes。這是module身分控制，source覆蓋增量為0。

## CTV byte-layout source spec：CONFORMED

READY審查後，ctv-source-module-r4／r5兩個新容器從repo ASM／EQU／typed JSON重建全部
2493 bytes，source／generated ASM／OBJ／MZ一致。CMF舊CLI亦在r7／r8回歸全部5296 bytes；
兩個已識別SDK module source合計7789 bytes，不用原始objects重連代替source。

原版CTVMEM.ASM已由固定SBCM.LIB whole-module與IDA9.4關係閉合。範圍linear22F60..2391D、
logical12F60..1391D、file142D0..14C8D，共2493 bytes，原EXE115282 bytes／SHA5178fdc8。
2258個instruction bytes維持既有source；data只准offset0003..00E3及0749..0754共235 bytes。

初值契約為兩個zero-ended header字串、eight zero bytes、原device/header初值、fixed ASCII版權文字、
14-entry API表、8-entry stream表、mutable state初值及11-byte unknown initialized payload。
API slot6必須保持6B06 literal，不能換成083A label；後者是E2平台契約衍生的初始化後target。
其他table entries以原instruction labels表達。各index gate與data references沿既有SDK data review。

11-byte區域經新IDA最小查詢確認為byte data item、size11、非code／非unknown／非alignment，
原始flags1300。前方RET、後方exported dispatcher與normal known table targets分離，沒有直接
recorded code entry或xref。data角色採strong scoped static layout，purpose保持unknown；
不命名為signature，不用xref缺項證明全程式永不引用，也不猜額外runtime效果。
source只保留其原始u8初值，沒有從此data區域匯入機器指令作為原碼。

build輸入僅repo語意ASM／EQU／typed JSON，literal data只落兩個approved non-code ranges。
範圍、field kind、值域、gap／overlap、handler label與slot6初值違反即拒絕；全2493 bytes
與定hash original比較、兩個新容器source／OBJ／MZ相同後才CONFORMED。
READY只對此固定byte-layout source成立，不批准unknown字段語意、原始type／完整ABI、hardware
wall-clock或normal campaign。Go／pack／UI／save均無修改，原始binaries不作build inputs。

copyright起點為offset0039／file14309h，前方NUL屬header word；原始235 data bytes保持。
九個CTV負例及七個CMF負例通過。slot6若替換083A label原本只能在全段比較時拒絕，
現已在schema層要求literal6B06，禁止用推導target替代初值。未命名payload只保留原值，purpose未知。
完整byte-layout不升格fields／全部API／原硬體runtime或全EXE。instruction5148與data2641
保持，C overlap12不重算。下一步回主程式C、其他source-unit與完整layout。

本機收據在既有full-goal-r1下：

- `ctv-data-class-probe-r1.py`、`ctv-data-class-r1.json`：IDA minimal query、原始flags／xref及script hash。
- `ctv-source-module-r4/receipt.json`、`ctv-source-module-r5/receipt.json`：最終完整CTV source重建。
- `cmf-source-module-r7/receipt.json`、`cmf-source-module-r8/receipt.json`：common builder與CMF相容CLI。
- `ctv-source-verification-draft-r1.json`：seed-to-label的舊schema缺口，失敗記錄保留。
- `sdk-full-source-verification-r1.json`：兩個module各兩份同值及16負例拒絕。
- `sdk-full-source-publish-verification-r1.json`及`goal-audit-r11.json`：現行source／producer／schema、連結、UID核對，完整Goal仍未完成。

## CMF byte-layout source spec：CONFORMED

此spec先完成READY審查，再由cmf-source-module-r3／r4兩個新容器從repo ASM／EQU／typed JSON
完整重建5296 bytes；source／generated ASM／OBJ／MZ相同，全段materialization與bytes通過。
七個缺值／overlap／越界／落code／錯handler／未READY／字串注入負例拒絕，scope限byte-layout。

本spec只批准原版CMFDRV.ASM的byte-layout重建。原始EXE115282 bytes／SHA2565178fdc8，
SDK module整段來源與完整hash沿用下節7789-byte原版身分證據。IDA9.4 linear23920..24DD0，
logical13920..14DD0，file14C90..16140，整段5296 bytes。

code source2890 bytes保持既有逐指令驗證；另外2406 bytes只按已審data區段還原：

- offset0003..0009：FMDRV零結尾signature，原始字串／module metadata確認。
- offset0009..0229：header與mutable state的原始byte初值。字段用途未知，不補猜原始C型別或合理初值。
- offset0229..0285：五個near handler表，8／4／16／15／3 entries，由原始index writer、gate、CALL consumer與instruction head閉合。
- offset0285..0905：原始parameter/index data，含DI+285／305等consumer；個別字段用途未知，保留為typed byte初值。
- offset12D3..1337：50-word IRQ stack reserve，原始zero初值。IRQ writer把SS設CS並SP設1337，exclusive end及100-byte範圍閉合。

data來源只描述ASCII、u8／u16值及符號handler references；禁止跨入任何已批准instruction range。
build輸入僅repo ASM、EQU與typed JSON，不用原始EXE／OBJ／SDK bytes作build input。
原始EXE只由verifier讀取作比較。輸出CODE與data覆蓋整段，含symbol實際relocations，
逐全部5296 bytes比較，兩個新容器獨立source／OBJ／MZ一致後才能CONFORMED。

失敗模式為缺field、錯kind／值範圍、offset gap／overlap、錯handler label、落入code、
原始input hash不同或任一byte DIFF，一律拒絕。字段semantics、source language／compiler、
硬體wall-clock、driver所有API與正常campaign仍unknown，這些不由本layout spec升格。
此工作不改Go／pack、玩家路徑、UI或save。完整原版硬體行為另依成熟平台契約，不為data初值深挖timer。

evidence review：CMF header及arrays位於原入口JMP跨越的data區域，原data xrefs／indexed
consumers與five dispatch tables一致；兩個ASM code區段及100-byte stack獨立，沒有機器指令混入data。
typed-data DRAFT prototype已在cmf-data-prototype-r1從source重建完整5296 bytes相同。
READY只對此固定module與byte-layout生效，原始data含未命名字段，沒有宣稱全部產品語意已解。

目前Source bytes：CMF整段5296包含2890 instructions及2406 data；不重複新增既有2890 code。
CTV仍省略235 data bytes，含未命名11-byte payload，不能由CMF完成宣稱兩driver皆完成。
whole driver byte-layout重建通過不等於全EXE、原版正常campaign或全部data semantics已完成。

CTV表slot6的6B06初值已保留。原版初始化把CS:0091地址換算為DMA目的地，
輸入06／6B至DSP E2；依[固定DOSBox Staging來源](https://github.com/dosbox-staging/dosbox-staging/blob/d8271efbccc7d0d6c0db60fb897ca937a80c7c6d/src/hardware/audio/soundblaster.cpp#L2069)
的reset state及E2 table，推導DMA輸出3A／08，即083A。它與原版2373A的比較值一致；
此結果屬mature-emulator platform-contract derivation，不冒稱原版實機或wall-clock parity。
因此slot6不是應改成有效code pointer的壞資料，原始literal及可變slot角色仍照實保存。
地址取用及DMA間接writer解釋了CPU直接xref沒有對應完整writer的現象。
CMF timer／IRQ的標準行為也依平台契約，這輪不深入逐週期硬體RE。

本機收據在既有full-goal-r1下：

- `dsp-e2-platform-r1/receipt.json`與完整source快照：固定commit、SHA256及source解析出的E2結果。
- `sdk-data-review-r2.json`：七handler tables、CTV mutable seed、IRQ stack及推論等級。
- `cmf-data-prototype-r1/receipt.json`：READY之前的可丟棄typed-data prototype。
- `cmf-source-module-r3/receipt.json`、`cmf-source-module-r4/receipt.json`：最終完整CMF source／data／fixup重建。
- `cmf-source-module-verification-r1.json`：七data-layout負例與原始objects未作build輸入。
- `cmf-source-publish-verification-r1.json`及`goal-audit-r10.json`：最終source hash／producer／schema、七負例、source入口與UID核對，完整Goal未完成。

build與驗證入口見上方[CMF source verifier](../tools/verify_cmf_source_module.py)。初值來自
已定hash的原始SDK／EXE bytes，未命名字段保留unknown，不以0或C remake猜補。

## SDK 語意指令來源重建

本節為最新結果，完整EXE與兩個driver的資料source仍未完成。原始EXE、library、IDA9.4
與linear／logical／file基準保持下節原版SDK身分契約。

新來源只保存已驗證語意指令。其code區段如下，ORG跨過的235／2406 data bytes不寫入OMF：

| SDK source | IDA linear指令範圍 | 原始指令bytes | 資料尚未還原 |
|---|---|---|---|
| ctvmem_code.asm | 22F60..22F63、23043..236A9、236B4..2391D | 959個指令、2258 bytes | 22F63..23043及236A9..236B4共235 bytes |
| cmfdrv_code.asm | 23920..23923、24225..24BF3、24C57..24DD0 | 1216個指令、2890 bytes | 23923..24225及24BF3..24C57共2406 bytes |

每個instruction label保存原始IDA名稱、linear／logical／file定位及原operand註解。
SDK library已由完整re-link與PUB／I/O／caller閉合身分，屬低層音效支援；這次ASM範圍
不含主程式game code。欄位用途、data角色、driver完整ABI仍未升格。
source內沒有DB／DW／DD／INCBIN／macro注入，原始objects不作source build輸入。
verifier核對實際OMF written mask恰等於declared instruction ranges，再用完整真正linker fixups
核對每個原始byte；linker為ORG holes產生的zeros從未算進source coverage。

NASM2.16控制樣本的MOV／SUB／ADD暫存器encoding為89／29／01，與原版8B／2B／03不同。
不能用新版manual或不存在的`{load}` decorator猜修。完整固定官方archive實際含Wasm，
其MASM模式控制樣本8BC3／2BC0／03C3與原版shape一致；來源clone加入已驗證Wasm形成
`dq3-watcom16:2.0-20261001-r2`，原r1 compiler image及payload保留。
Wasm binary SHA256為 `7e216ab56214fe36f80fa60cc757d05f4508683ef28e992969556945fe28f2cd`。
同一固定runtime／Dockerfile，763 payload files逐檔核對，沒有host runtime或新未鎖版依賴。

編碼契約的幾個必要控制已實測：

- 原始16-bit displacement即使值為0／1／3，仍需保留寬度。數值、forward EQU與WORD PTR cast會縮短；外部absolute EQU symbol經WLINK正常重定位，保留原來16-bit欄位。
- AX accumulator的ADD／SUB／CMP／AND與83 sign-extended形式同長但opcode不同；同樣以正常absolute EQU qualifier保留原encoding，不寫opcode bytes。
- XCHG兩個register的語意對稱，Wasm canonical encoding需交換來源operand次序；原operand保留註解，兩邊的machine bytes核對相同。
- IDA列出的LOOP／MUL／DIV／string隱含operands不直接輸出，明示operand及REP／segment prefix完整保存。

最初兩份prototype語法錯誤，後兩份長度／encoding DIFF均保留。完整prototype-r6的語意指令
加未審查data literals可重建兩段7789 bytes，但phase保持DRAFT、正式source增量0。
公開來源移除全部未審查data，只保留已match instructions；source-build-r3／r4兩份
完整source／OBJ／MZ／FIXUPP receipts相同。七種byte/data/include/macro注入負例拒絕。
完整driver的code／data spec仍DRAFT；指令來源重建通過不升格driver功能、正常campaign或全EXE。

source指令共5148 bytes，其中12 bytes已由三個C helper覆蓋，新增唯一指令來源5136 bytes。
原四C19 bytes與兩RNG ASM46 bytes保持；全部已驗證的唯一指令來源5201 bytes只是局部
reproduction統計，不是完整source-unit、原版semantic或完整Goal完成比例。

本機收據在 `work/matching-decomp-20261008-r1/full-goal-r1/`：

- `sdk-module-layout-r1.json`：新IDA operand text及四個缺口的原bytes／items／xref。
- `asm-encoding-controls-r1/`、`wasm-controls-r1/`、`wasm-width-control-r1/`、`wasm-absolute-control-r1/`：具體assembler正反控制。
- `sdk-asm-prototype-r1`至`r6`：可丟棄DRAFT的語法／encoding差異與最後全段prototype。
- `sdk-instruction-source-r1/receipt.json`：移除全部data後逐指令匹配。
- `sdk-source-build-r3/receipt.json`、`sdk-source-build-r4/receipt.json`：最終公開source的兩次乾淨重編。
- `sdk-source-guard-verification-r1.json`、`sdk-instruction-source-verification-r1.json`：七負例與12-byte重疊核對。
- `sdk-source-publish-verification-r1.json`及`goal-audit-r9.json`：最終source／manifest／producer、連結、UID及完整Goal未完成核對。
- `sdk-source-build-r5/receipt.json`、`sdk-source-build-r6/receipt.json`及`sdk-source-publish-verification-r2.json`：行尾空白修正後的現行source hash與兩次完整重編，instruction bytes保持5148。
- `watcom16-asm-source-r1/payload/source-manifest.json`：Wasm r2的完整固定payload。

下一步審四個data gaps的dispatcher／table／buffer證據，建立typed data及完整module spec。
主程式C與其餘source-unit／layout／完整EXE仍未完成；正確指令來源按使用者要求提交GitHub。

## 原版 SDK module 身分與 C 來源歸屬

本節保存上一檢查點。原版 `assets_raw/DQ3.EXE` 為115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
附帶 `assets_raw/SBCM.LIB` 為35840 bytes，SHA-256
`01b242cb99193d006e23b73babe122887e59713410f88798283a9120df1d0683`。
原版查詢使用IDA9.4；linear基準0x10000，file=`linear-0x10000+0x1370`。

SBCM是實際OMF library，page16、dictionary起點file8200、五個512-byte dictionary blocks，
26個module均有`.ASM` THEADR名稱。這是原始metadata，不等於主程式編譯器或全部來源語言。
兩個原先未支援module使用LIDATA；bounded USE16 decoder保留原始OBJ，再建立LEDATA parser view。
重複資料後若有FIXUPP仍拒絕，encoded／decoded大小、hash與原始record位置分開記錄。
CMFDRV還有226個未寫入bytes，導航不得自行補0聲稱原碼匹配。

| module metadata | 原版完整範圍 | 完整身分證據 | IDA導航 |
|---|---|---|---|
| CTVMEM.ASM／CTVM_VOICE_DRV | linear22F60..2391D；logical12F60..1391D；file142D0..14C8D，2493 bytes | 原始OMF全段沒有FIXUPP；官方WLINK重連的完整code／data等於原版 | 34自動函式、959指令位置，入口有17個typed far-call refs，包含SDK wrapper及原版20577／20582／205A0 |
| CMFDRV.ASM／SBFM_CMF_DRV | linear23920..24DD0；logical13920..14DD0；file14C90..16140，5296 bytes | 官方WLINK從原始OMF重連，包含兩個LIDATA及未寫入區域的完整段等於原版 | 46自動函式、1216指令位置，入口有19個typed far-call refs |

兩段合計7789 bytes，SDK原始object重連的兩次完整MZ亦相同。
推論等級為confirmed scoped original-module identity；它們是Creative音效支援module，
原始PUBLIC metadata、DSP／timer／PIC I/O及caller證據一致。
這些數字包含module內data，不當作已還原ASM code或source coverage。
可讀原碼、code／data分類及全部driver ABI仍待還原；不以原版OBJ作最終source build輸入。
硬體wall-clock與DAC／PIT逐週期時序不在這次module身分查詢內。

兩段均沒有原版MZ relocation。CTVMEM之後的三個alignment bytes不在2493-byte module證據內，
最終layout仍須獨立核對。另八個有排除fixup搜尋命中的SDK module保持hypothesis，
它們的外部symbol／group frame未閉合，不能由導航候選升格whole-segment parity。
沒有命中也不證明module未被使用，ctvdsk包含未寫入區域，仍未批准來源／歸屬。

三個已匹配C函式的原始bytes及IDA names／chunks保持：

- sub_236F5位於CTVMEM.ASM，C完整4 bytes。
- sub_24A8D與sub_24AE6位於CMFDRV.ASM，各完整4 bytes。

正式C仍四個19 bytes，這三個12 bytes屬於SDK driver，不當作主程式產品邏輯已完成。
剩餘sub_15D49的7 bytes位於MZ入口code segment；table47條件式入口保持，但產品角色未知。
兩份更新manifest重編保持原code，欄位語意、原始型別與runtime DS不因module歸屬而升格。

舊「Press X」字串的compiler論據亦須限縮：fresh IDA在linear20288有data offset xref，
LEA DX至linear200AB後CALL2030A，屬另一段原版DOS錯誤訊息consumer。
沒有直接MSC版本／runtime map證據；這段consumer與Creative driver身分均不能證明主程式用MSC5.x。
舊近RET、LOOP與C小函式匹配也不足以決定整個程式的語言或memory model，歷史論述見docs/19勘誤。

本機收據在 `work/matching-decomp-20261008-r1/full-goal-r1/`：

- `sbcm-modules-r3/receipt.json`與原始26個OBJ：metadata、LIDATA parser view及九個導航候選。
- `sdk-linker-control-r3/receipt.json`、`sdk-linker-control-r4/receipt.json`：原版objects的官方重連，不是原碼重建。
- `sdk-module-refs-r4.json`：fresh IDA原名、80函式、2175指令位置、typed xref、startup及字串consumer。
- `sdk-module-verification-r1.json`：五負例、兩次完整vendor MZ及原版輸入身分。
- `sdk-ownership-verification-r1.json`：IDA與原inventory的names／bytes／chunks保持及三C歸屬。
- `source-publish-verification-r1.json`：GitHub來源入口、四C19 bytes、兩ASM46 bytes新編、索引及連結核對。
- `c-batch-r15/receipt.json`、`c-batch-r16/receipt.json`、`watcom-source-r10/receipt.json`、`watcom-source-r11/receipt.json`及`goal-audit-r8.json`：更新SDK歸屬後的正式C重編與四個19 bytes核對。

第一次IDA匯出因BADADDR常數所在module不同而留下error sidecar，改用idc.BADADDR後由原工作庫
新副本重跑；原始EXE、library、正式database、names與function boundaries均未修改。
下一步為兩個已證實音效module建立可讀ASM／data source spec與code／data分類，再做source rebuild。
主程式C、其他SDK frame及完整layout／EXE六gate繼續保持未完成，不重跑前輪134組。

## 主程式迴圈 codegen 控制

本節保存上一檢查點，正式C覆蓋仍四個19 bytes。原始 `assets_raw/DQ3.EXE` 為115282 bytes，
SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
原版證據使用IDA9.4 inventory，linear基準0x10000，file=`linear-0x10000+0x1370`。
兩個候選範圍為sub_132A3的linear132A3..132B4／logical32A3..32B4／file4613..4624，
以及sub_16FCF的linear16FCF..16FDE／logical6FCF..6FDE／file833F..834E。

| 控制 | 實際結果 | 界線 |
|---|---|---|
| Watcom填表，8種C寫法×9組正式flags | 72案均DIFF，16種resolved code，最短17 bytes | 關閉重排及先宣告count可還原初始化與store順序；-ot可發原版ADD BX,2，但仍以DEC／JNE計數，完整18 bytes不等於原版17 bytes |
| Watcom搜尋，6種C寫法×9組flags | 54案均DIFF，24種resolved code，最短20 bytes | register配置、LODSB及LOOP均未完整匹配；短於原候選24 bytes不等於exact |
| Watcom可讀C的long-shift正對照 | 整個15-byte module包含JCXZ、SHL／RCL與LOOP | confirmed compiler emission；沒有原版unit身分，不計C覆蓋，也不證明任意C迴圈可用LOOP |
| TC2.01一般register locals，4組flags | 四案REFUSED，實際fixup為frame method1／target2 | group-frame語意尚未支持，不假裝resolved或用raw operand當匹配 |
| TC2.01 `_CX`／`_BX`偽變數，4組flags | 四案均DIFF，完整19 bytes相同 | 原生C extension認得這些名稱；仍生成兩次INC、DEC、OR及JNE，不是原版ADD／LOOP |

全部來源為C或instruction-free register pragma，沒有inline instruction或machine-code array。
Watcom的`-ol`是loop最佳化選項；本輪126案沒有因這個名稱產生原版LOOP。
這些結果只排除列出的source／flags／compiler組合，不證明所有C compiler均無法匹配。
原版編譯器、原始型別、主程式與底層unit歸屬仍unknown。

官方Watcom release `2026-10-01-Build` 指向固定commit
`e28568669775a7119f381f3b37d21745d8afcfb6`。54份i86及共用Intel C codegen來源共935938 bytes，
逐檔Git blob SHA1與SHA256核對，保存在 `watcom-codegen-source-r1/`。
其中[Do4CXShift](https://github.com/open-watcom/open-watcom-v2/blob/e28568669775a7119f381f3b37d21745d8afcfb6/bld/cg/intel/i86/c/i86enc.c#L177)
在lines244／252發出M_LOOP；可讀C long-shift控制亦實際產生該指令。
來源snapshot不是完整compiler，不能從54份檔案的搜尋缺項宣稱整個compiler沒有其他emitter。
第一次猜8086路徑404後依官方目錄改為i86，保留既有tag收據，沒有重抓成功輸入。

TC2.01沿用本機 `tools/build/tc201.zip`，SHA256
`2c87f988605ae9ed70e5fef35b9854de87e36ccdb021caec35ab2424ca4b5553`，
compiler `Disk2/TCC.EXE` 的SHA256為
`19650666dcaa03e3f68efd9beeb57821ba4ed4d84c88a52b0e989bbfc97e07ba`，
並與archive內原檔逐byte核對。沿用dq3-msc image中的DOSBox，未另建image或使用舊host wrapper。
初次batch少CALL，compile及DONE已完成但留在DOS prompt，90秒逾時；新runner使用
`call go.bat`後正常退出。此失敗屬shell腳本，不列compiler或產品缺陷。
其後兩份object的COMENT classE9只有生成C的DOS file time不同，code相同但OBJ不相同。
固定實際生成C的mtime為2026-10-01 UTC後，兩次完整OBJ相同；不strip、mask或patch object。

本機收據位於 `work/matching-decomp-20261008-r1/full-goal-r1/`：

- `primary-codegen-r3/receipt.json`、`primary-codegen-r4/receipt.json`：126案及long-shift正對照，完整source／OBJ／code／fixup重編一致。
- `turboc-primary-r4/receipt.json`、`turboc-primary-r5/receipt.json`：八案的完整source／OBJ／code／拒絕結果一致。
- `watcom-codegen-source-r1/manifest.json`：官方固定commit與54份source hash。
- `turboc-metadata-C000-r1.json`、`turboc-metadata-C004-r1.json`：舊object的原始差異record。
- `primary-codegen-verification-r1.json`：兩套compiler逐案重新讀object核對，原版hash、source snapshot及UID通過。
- `turboc-source-time-control-r1.json`：16份OBJ的實際非空E9 time／date為0000／5D41；短E9 record另保留，不當成時間欄位。
- `goal-audit-r7.json`：正式來源與原始範圍重核，仍四個19 bytes，完整Goal未完成。

正式C覆蓋沒有增加，完整Goal六gate仍未證實。下一步回到原版主程式的compiler／ABI與
source-unit歸屬證據；不重跑這134個已排除組合，也不由局部compiler控制猜原版語言。

## 暫存器 C 來源與重定位核對

本節保存上一檢查點；最新結果見上節。輸入為 `assets_raw/DQ3.EXE`，115282 bytes，
SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
IDA9.4 linear基準為0x10000，file=`linear-0x10000+0x1370`。

| 原始範圍 | 可讀 C 結果 | 證據等級與界線 |
|---|---|---|
| sub_24AE6，IDA linear24AE6..24AEA，logical14AE6..14AEA，file15E56..15E5A | AX寫DS-relative word0032後near return，整個4-byte module為`A33200C3`，精確匹配 | confirmed scoped instruction及compiler bytes；DS、欄位用途、原始型別與module分類unknown |
| sub_132A3，IDA linear132A3..132B4，logical32A3..32B4，file4613..4624 | stride-two填13 bytes；C亦17 bytes，但先INC BX兩次，再以symbol-2寫入，以DEC／JNE計數 | DIFF。原版先寫入、ADD BX,2，再LOOP；指令順序、encoding及flags不同，不計入coverage |
| sub_16FCF，IDA linear16FCF..16FDE，logical6FCF..6FDE，file833F..834E | 六byte搜尋C為24 bytes，原版15 bytes | DIFF。C保存DX並重新分配target／index／counter；原始AH、AL、BL、SI、CX效果尚未完整匹配 |

pragma僅宣告暫存器介面，沒有inline instruction、`db`或原始code array。
來源清單核對full IDA inventory、原始bytes及完整function chunk；其C型別只代表已驗16-bit
compiler表示，不推定原版語言或產品語意。填表／搜尋來源保持研究候選，不作已完成source unit。

WCC在填表object的F5/T2 fixup留下raw word `FFFE`。wdis列為symbol-2。
unsigned parser依舊拒絕overflow；只對明示producer contract啟用signed16，不自動猜sign或wrap。
正對照以vendor WLINK連結實際C object與synthetic DATA：PUBDEF為265D，vendor map中的frame
offset為265F，實際operand為265D，證實265F-2。此frame bias不是原版DS layout證據。
指定原版DS offset265D時，候選operand為265B；與vendor fixture的配置分開記錄。
六負例包含unsigned overflow、signed underflow、缺placement、負placement、越界但結果仍在
16-bit範圍的placement及未知frame，全部拒絕。TIS格式來源為
[OMF1.1規格](https://openwatcom.org/ftp/devel/docs/omf.pdf)；signed解釋限於這次WCC／WLINK控制。

兩份獨立MSC重編 `c-batch-r13`／`r14` 及兩個全新容器的Watcom重編
`watcom-source-r7`／`r9` 核對實際OBJ、code與fixup一致。
同容器第二次使用固定compile目錄的 `watcom-source-r8` 失敗收據保留；新容器重跑不改來源或flags。
正式C覆蓋為MSC三個15 bytes加Watcom一個4 bytes，共四個19 bytes。
第一個MSC module的ENDP後NOP配置仍未解，六個完整Goal gates均未完成。

本機收據在 `work/matching-decomp-20261008-r1/full-goal-r1/`：

- `signed-reloc-r3/receipt.json`及vendor MZ／map／log：producer-scoped正反控制。
- `watcom-source-r7/receipt.json`、`watcom-source-r9/receipt.json`：三個完整C來源的兩次重編。
- `c-batch-r13/receipt.json`、`c-batch-r14/receipt.json`：parser更新後MSC重編。
- `goal-audit-r6.json`：兩套compiler的實際bytes、來源與重定位總核對。
- `c-source-verification-r1.json`：偽造exact、過期repeat manifest與重複unit三負例、語法與UID核對。

下一步追主程式LOOP／LODSB及多入口的compiler codegen與source-unit歸屬。
原版compiler保持unknown；不以數學等價或register宣告代替精確原碼重建。

## 多入口CFG與register-ABI工具鏈續行

本節保存前一檢查點，最新結果見上節。

17個已審跨界entry以IDA原始flow／jump追到991個唯一指令位置，31位置由多entry共用。
每個候選保留code／data外部xref、call dependency、原始owner與實際return，不補造被抑制
的call-return。此範圍沒有未映射successor，不代表callee effects或source-unit歸屬已完成。
所有候選的 `source_unit_extent_approved` 仍為false。

第一組IDA linear10000..10030共14個指令／48 bytes，保留入口10000、1000A、10014及1001E；
不能把四個原始entry合成一個有新selector參數的C函式。另一組11900／1196B亦共用suffix。
正對照核對四入口與14指令，偽造跳到10007的edge在原始指令中被拒絕。
候選及測試收據為 `full-goal-r1/cfg-candidates-r1.json`／`cfg-verification-r1.json`。

原版SI／BX暫存器參數不能由一般MSC stack介面近似。本輪先核對既有
`fd2-watcom-matching:2.0-20261001-r1`，確認它只有wcc386，沒有16-bit wcc。
沿用官方同release，完整下載129081693-byte archive並核對SHA256
`a961f3e02ce27bcd88428345dc57483457a623b6e718b6cee52c48d24ab8065f`，再逐檔核對762項
compiler／header payload。外層下載程序收尾觸及timeout124，但獨立完整hash與payload清冊
均通過，沒有重啟成功下載或把timeout當compiler缺陷。

新revision `dq3-watcom16:2.0-20261001-r1` 提供官方wcc16／wcc386／wdis／wlink／wlib及h。
來源與建置契約見上方fetch script及tools/build/README；原版compiler仍unknown。
暫存器宣告依[官方Open Watcom guide](https://open-watcom.github.io/open-watcom-1.9/cguide.html)
的16-bit parm／value／modify exact規則，只指定介面，沒有`=` inline instruction或machine-code array。

| C控制 | 原始編譯結果 | 限制 |
|---|---|---|
| BX／SI word echo | compiler分別發`89D8C3`／`89F0C3`，register→AX後near return | compiler控制樣本，不登記成原版function |
| AX word store | `A33200C3`，與原版sub_24AE6完整4 bytes相同 | 仍是實驗控制，沒有approved source-unit ledger，不新增正式C覆蓋 |
| 三次_rotl core | -oi後24 bytes，包含CX保存、CL rotate與volatile重讀 | 與原版16 bytes不匹配 |
| BX／DX:AX bounded RNG | -oi後51 bytes，包含保存暫存器與兩次DIV | 與原版30 bytes不匹配，C不進production |
| local／shift-or core | 數學等價C40 bytes | 與原版16 bytes不匹配，不用數值等價充作exact |

最終六控制使用flags `-bt=dos -ms -0 -os -oi -s -ofr -ecc -zld`。
最初未加-oi，標準_rotl為真正library call，未知code fixup仍REFUSED；依官方inline契約加-oi後
再乾淨重跑。兩次OBJ核對亦揭露compiler THEADR的absolute source path及dependency timestamp；
實際OMF差異保留。以固定 `/tmp/watcom16-compile` 及官方-zld修正建置環境，沒有遮罩、
刪除或canonical化OBJ bytes。最終 `watcom16-abi-r7`／`r8` 六份OBJ、code與fixup完全相同。

正式C覆蓋仍3個／15 bytes，低階ASM仍只保留既有局部證據，完整Goal六gate未完成。
下一步以已驗instruction-free register ABI，為保留多entry的主程式資料流建立C候選；
若source-unit或callee語意未閉合，先補證據，不用新呼叫介面改動原版entry／stack／side effects。

## 邊界與間接入口續行

原始輸入、IDA9.4及linear／logical／file基準保持上節規則。
fresh database的 `boundary-ida-r2.json` 與 `boundary-review-r1.json` 已核對：

| 原候選 | 審查結果 | 等級與限制 |
|---|---|---|
| 17個跨邊界末端 | 實際IDA `fl_F` type21連到physical next instruction，跨出原自動函式範圍 | confirmed scoped static flow；需閉合完整CFG與所有外部entry，不直接合併成C函式 |
| `sub_102C4`／`start`末端 | 原始相鄰`MOV AH,4Ch; INT21h`選擇DOS terminate service，局部末端有效 | confirmed service selection；不由最後一個terminal block宣稱整個函式永不返回，亦不自動歸屬後方RET |
| 3個抑制post-call flow | callee實際`FUNC_NORET`為1。`sub_192F0`末端call `sub_193E3`，後者以原始`0x194bf→0x19405`形成事件循環；`sub_16346→sub_1E713→sub_1E7F3→0x192bc`連到DOS清理／退出區塊 | strong scoped callee interpretation；flag自身不是證據，indirect／nonlocal transfer與source-unit範圍仍待審 |
| 三個舊entry | IDA linear182BB落在182B8 far call的第3 byte；1A660落在1A65D MOV的第3 byte；1A753落在1A751 near call的第2 byte。fresh查詢皆無incoming xref | confirmed current-input boundary；排除舊函式入口，若有獨立入口證據再開，不宣稱全程式沒有重疊指令 |

DOS AH4C的標準行為依[RBIL Release61](https://fd.lod.bz/rbil/interrup/dos_kernel/214c.html)，
不為標準API語意另開遊戲RE切片。本案只核對實際AH writer、callsite與原始轉跳。
其他17個跨邊界流程是已確認關係，完整source-unit本體仍未批准。

首個C的間接入口現在有條件式靜態閉合：startup原始carrier
`0x192ab`的`B8DD14`經MZ relocation成IDA segment24DD，`0x192ae`將其寫DS。
在DS為此DGROUP的條件下，`0x14feb`讀`[di+4]`至BL，清BH、SHL BX一次，
`0x14ff2`取DS:3BB4到SI，`0x14ff6`加BX，讀word並以零值gate決定是否`CALL [SI]`。
table base為IDA linear28984，entry289E2是第47個word，file19D52的495D指向logical5D49。
DI物件來源、欄位`+4`產品意義、table完整合法範圍與實際runtime DS均未驗，
故只標strong conditional static mapping，不把`+4`命名為NPC handler或故事旗標。

新增C來源 `sub_236F5`／`sub_24A8D` 分別讀DS-relative word0033／000C至AX後near return。
對應source為上方已索引的 `sub_136f5.c`／`sub_14a8d.c`，各4 bytes；
field與DS／module脈絡保持unknown，C的unsigned只是已驗16-bit表示，不宣稱原版型別。
完整C本體目前3個／15 bytes；第一個7-byte module仍多1-byte NOP，兩個新4-byte module
則完整匹配。這不升格完整主程式、底層分類或全EXE重建。

最新C批次為 `full-goal-r1/c-batch-r9`／`c-batch-r10`，三份OBJ、重定位body／module
及listing逐byte相同。fresh exporter亦保留全部33個原始indirect calls及typed operands。
後續從17個實際flow edge建立跨entry CFG候選，不靠助記符或auto owner猜C函式範圍。

## 完整Goal起始基準

本節保存完整Goal啟動時的基準；最新續行結果見上節，唯一目前狀態在CONTEXT。
輸入為 `assets_raw/DQ3.EXE`，115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
IDA9.4 linear基準 `0x10000`，file=`linear-0x10000+0x1370`。

| 層次 | 已核對的目前狀態 | 尚未完成 |
|---|---|---|
| 完整原始導航 | 新IDA database匯出828個自動函式、29979個code heads、81854個唯一原始code bytes；無未映射／重疊的code bytes | 自動邊界不是source unit；22個末端可能切斷活躍控制流，4455個指令位置在函式外，程式／資料與模組歸屬仍需審查 |
| 舊清單對照 | 舊280個entry與IDA共有277；三個舊entry不在新函式入口，新清單另有551個entry | 不用280或828直接宣稱完整主程式分母；先閉合source-unit範圍與底層分類 |
| 主程式C | `sub_15D49`，IDA linear`0x15d49..0x15d50`、logical`0x5d49..0x5d50`、file`0x70b9..0x70c0`；讀得懂的C word assignment與near return，完整7 bytes匹配 | DS:0B60產品用途未知；入口目前只有table data xref `0x289e2`，不從名稱猜旗標用途；其他主程式C尚未完成 |
| compiler範圍 | 固定MSC候選 `/c /AS /Os /Gs /Fc` listing的`PROC..ENDP`為offset0..7，真實OMF fixup將外部word配置到DS:0B60；函式本體7 bytes相同 | 完整module有offset7的NOP，raw module8 bytes不等於原版7-byte函式；final module padding placement未知 |
| 組語 | 既有46 bytes的原始指令精確匹配與局部ABI收據保持 | 在完整Goal計算前仍須核對每個ASM unit是否屬於使用者允許的已確認底層常式 |
| 整檔 | 7-byte C partial scaffold與原版hash相同，其餘115275 bytes原樣保留 | 保留區域不算source recovery；完整source/data/layout、真正乾淨整檔重建與runtime收據均不存在 |

`/Ox`第一輪的7個函式bytes相同，整段多NOP，保留為DIFF。
`/Os`兩輪都額外呼叫compiler的`__chkstk`，因未支持該code fixup而REFUSED。
回查compiler／runtime分流入口後，`/Gs`控制樣本移除了這條額外呼叫，但仍有模組NOP。
compiler自身listing的`ENDP`位於NOP之前，提供了新的獨立函式範圍證據；
因此後續依PROC／ENDP核對整個函式，並保留完整module的DIFF及未解padding，
沒有以RET、NOP內容、原版長度或遮罩猜切點。

首個C當時兩份乾淨重編為 `full-goal-r1/c-batch-r6`／`c-batch-r7`，OBJ、實際重定位的函式／
module bytes及listing相同。compiler版本身分仍只適用掛載候選，原版compiler保持unknown。
新full-inventory首輪因IDA環境使用ASCII寫出而失敗，空sidecar與log保留；
明示UTF-8後同一工具與唯讀原版乾淨重跑，`inventory-ida-r2.json`才是有效來源。

本機收據根 `work/matching-decomp-20261008-r1/full-goal-r1/`：
`first-c-ida.json`、`inventory-ida-r2.json`、`c-batch-r7/receipt.json`及`goal-audit-r1.json`。
完整Goal六個gate保持未證實，evidence audit PASS不代表Goal完成。
當時下一步先審22個非terminal邊界與三個舊entry，建立主程式source-unit及間接入口對照，
再依原始資料流還原下一批C；不能把自動函式碎片直接編成獨立C函式。

已有唯讀原版、MSC工具與上述research根時，使用新輸出名稱重跑C批次：

```bash
timeout 160s docker run --rm --network none --memory 1g --cpus 1 --pids-limit 128 \
  --user "$(id -u):$(id -g)" \
  -v "$PWD":/repo:ro -v "$PWD/tools/build/msc":/msc:ro \
  -v "$PWD/work/matching-decomp-20261008-r1/full-goal-r1":/out \
  --workdir /tmp --entrypoint python3 dq3-msc:bookworm-20261008-r1 \
  /repo/tools/run_matching_c_batch.py --output /out/c-batch-new \
  --ida-evidence /repo/work/matching-decomp-20261008-r1/full-goal-r1/first-c-ida.json
```

掛載前先核對host路徑存在、檔案／目錄形態及UID/GID。
同一容器可執行 `matching_goal_audit.py --inventory /repo/work/matching-decomp-20261008-r1/full-goal-r1/inventory-ida-r2.json`
`--c-receipt /repo/work/matching-decomp-20261008-r1/full-goal-r1/c-batch-new/receipt.json`
`--repeat-c-receipt /repo/work/matching-decomp-20261008-r1/full-goal-r1/c-batch-r7/receipt.json`
`--output /out/goal-audit-new.json`；要求兩份來源與producer相同，輸出不得覆寫。

> 2026-10-08：依 [Issue #5](https://github.com/wicanr2/kinginformation-dq3-re/issues/5)
> 啟動局部 matching 的對拍加速實驗。下方 MSC 5.x「已鎖定」與固定 codegen 成因均為
> 歷史判讀，尚未由精確 compiler／linker 版本及完整重定位閉合。五個既有 OMF 產物
> 先以唯讀方式核對為 raw exact 0／5；本輪再新編七個 C 候選，沒有 exact。
> 具語意 ASM 已重編兩個函式共46 bytes並完全匹配。原版是否有殼仍未知。
> 新探針入口為 [`tools/ida_matching_probe.py`](../tools/ida_matching_probe.py)，
> 以 IDA Pro 9.4 匯出原始定位、bytes、typed xref 與既有分級語意。
> 研究輸出位於 gitignored `work/matching-decomp-20261008-r1/`。
> prototype 不改正式 Go／game-pack；完整 EXE 原碼重建不因本實驗成為 remake 完成閘門。
> [`re/match/sub_e6b9.asm`](../re/match/sub_e6b9.asm) 保存 IDA 核對的七條指令，
> 供 NASM 重新組譯實驗。它不含 `db`，ASM 與 C 覆蓋率分開計算。
> [`tools/omf_matching_probe.py`](../tools/omf_matching_probe.py) 依
> [TIS OMF 1.1](https://openwatcom.org/ftp/devel/docs/omf.pdf) 解析真實 PUBDEF／SEGDEF／FIXUPP。
> 它只支援本實驗的單函式 USE16 code 與明示外部符號 DS offset，其他 fixup 拒絕解析；
> 不使用位元組遮罩或「最大段」猜選函式。
> [`tools/run_matching_probe.py`](../tools/run_matching_probe.py) 在 MSC 研究容器內執行
> 單次 DOSBox 批次編譯、真實 OMF 檢查、具語意 ASM 組譯與壞來源拒絕，輸出本機收據。
> [`tools/run_inertia_matching_probe.py`](../tools/run_inertia_matching_probe.py) 明示核對
> Inertia 的 MZ 載入基準，對 RNG／有界 RNG／NPC mover 做限時實測；生成 C 不自動升格為 match。
> 續行ABI研究沿用本入口；[`re/match/sub_e6c9.asm`](../re/match/sub_e6c9.asm)
> 保存BX有界RNG的原始13條指令。`ida_matching_probe.py`同時匯出直接caller的原始
> bytes／typed xref與前後窗口，caller語意在審查前仍保持unknown。
> [`tools/dosgolem_matching_rng_abi.go`](../tools/dosgolem_matching_rng_abi.go) 是Docker內的
> 明示局部測試，固定列舉seed與邊界輸入；核對原版／source-assembled ABI、寫入端與保留
> 暫存器，不取代正常玩家路徑，也不深追全流程亂數呼叫序列。
> [`tools/run_matching_rng_abi.py`](../tools/run_matching_rng_abi.py) 只複製唯讀dosgolem的
> internal非測試Go來源與go.mod，在Docker暫存module建置；不讀取上游cmd/probe scratch，
> 來源清冊／hash與控制輸入收據另存明確研究輸出。
> [`tools/run_msc_abi_controls.py`](../tools/run_msc_abi_controls.py) 編譯已知的16／32-bit
> 回傳C控制樣本與`_fastcall`關鍵字候選；只核對掛載MSC候選的ABI，不反推原版語言。
> 已審查的局部ABI註記保存於[`tools/ida_rng_abi_ledger.json`](../tools/ida_rng_abi_ledger.json)，
> 以原始IDA位址、file offset與bytes為key；IDA探針自動合併其推論等級、consumer與局部收據來源。

## 2026-10-08 首批實測

### 續行：C／ASM adapter 實測

[`tools/run_matching_rng_adapter.py`](../tools/run_matching_rng_adapter.py) 分兩個 Docker
階段執行：`compile` 使用既有 MSC 映像，`cpu` 使用既有 Go 映像及唯讀 dosgolem。
[`re/match/rng_adapter.asm`](../re/match/rng_adapter.asm) 是本專案撰寫的診斷轉接，
將 BX 放到 C 堆疊參數，保存其餘暫存器，零上限直接返回。
[`tools/dosgolem_matching_rng_adapter.go`](../tools/dosgolem_matching_rng_adapter.go)
核對固定 seed 全集、邊界、四種錯誤轉接，以及暫存器、持久狀態、旗標與額外堆疊寫入。
prototype 不進正式程式；數值通過與原始 bytes exact 分開記錄。

原版仍是 `assets_raw/DQ3.EXE`，115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
原始定位與 bytes 取自 IDA9.4 的 `rng-abi-r1/ida-rng-abi-reviewed.json`，
核心為 IDA linear `0x1e6b9..0x1e6c9`，有界常式為 `0x1e6c9..0x1e6e7`。
file 與 logical 的換算保持下節規則；diagnostic adapter 位於 dosgolem logical `0x0400`，
C code 位於 logical `0x0500`，不將這兩個測試位址登記成原版函式。

| 比較層 | 實測結果 | 證據界線 |
|---|---|---|
| 原版／source ASM | 兩函式重新組譯46 bytes exact；有界常式131152組同狀態呼叫的全部觀測結果一致 | confirmed local instruction reproduction；其餘115236原始bytes沒有原碼重建聲明 |
| C／ASM adapter | 56-byte C加31-byte adapter；原版有界常式30 bytes。65536個seed各測BX10與BX0，加80個邊界，全部暫存器、高半部、段暫存器、返回位置及持久state相同 | confirmed scoped register/state equivalence；明示direct-entry、注入與CPU重入，非正常玩家流程 |
| 記憶體／旗標 | 65616次非零上限呼叫留下額外堆疊內容；返回後完整128-byte scratch stack均與原版不同。同批raw flags也全部不同，BX0不交易state或stack，旗標相同 | full memory equivalence未通過；DIV算術旗標不指定猜測的硬體期望值，不由raw差異宣稱C數值規則錯誤 |
| 失敗定位 | 交換AX／DX、誤傳CX、移除零上限gate及DS offset改成0B5B四案全部拒絕；零上限負例觸發除零 | 明示預定固定seed1357，無重擲或挑選結果；工具直接回報register/state mismatch或division exception |
| C exact | 新候選raw bytes為DIFF，C exact仍0 | 沒有以數值成功、暫存器轉接或忽略stack宣稱byte-match |

候選compiler的固定CL／C2／C3分別在 `(file)0x745d`、`0x2ccbe`、`0x1d6a9`
含 `C 5.10` 字串。CL同時含FORTRAN及Quick C的banner，所以這些是已核對component的
版本標記，不能將所有可列印字串當作實際編譯路徑，也不能由此定案DQ3的compiler。
完整字串、file offset與component hash保存在compile收據，原版compiler仍unknown。
修正Dockerfile的舊指紋斷言只涉及註解，映像執行契約與建置指令保持。

局部成本現在可重測：三次單候選DOSBox編譯各約0.716／0.766／0.716秒，
兩份獨立編譯的OBJ、重定位C code與全部ASM artifact相同。
這個工具可自動辨識上表四類錯誤，省去人工逐暫存器及state核對。
沒有相同工作的人工工時或完整玩家路線成本基準，故不給加速倍數，也不宣稱整體對拍加速。
setup包含撰寫adapter、追查ABI與乾淨Go建置，不能只拿不到一秒的CPU時間當端到端成本。
最終CPU三方比較0.711280秒，乾淨Go建置9.127850秒；局部CPU步數為原版983920、
exact ASM983920及C／ASM adapter3149248，不能把匹配原碼等同更快的執行。
首輪只觀測modified-byte事件，第二輪另補返回後完整stack bytes；兩輪舊收據均保留。

最終輸出沿用 `work/matching-decomp-20261008-r1/`：
`rng-adapter-compile-r3/compile-receipt.json`、`rng-adapter-cpu-r3/receipt.json`、
`producer-meta.json`與`final-audit.json`。後兩份位於CPU輸出目錄。
CPU收據SHA-256 `9d702555c10aeccef610832acf2cf76d9cda8b660c161a886444707e1233246a`；
compile收據SHA-256 `b5c4f94f0f25409f127cdae35e5cd3f6a201bb57683085c19df641fbf6feede4`。
producer與OMF parser的新鮮度、唯讀dosgolem選定來源、原版ABI ledger、既有收據hash、
獨立重編、工具索引正對照及UID/GID均由`audit`階段核對。

重跑前先確認每個host掛載存在、形態及UID/GID；沿用既有研究根，輸出使用全新名稱：

```bash
timeout 120s docker run --rm --network none --memory 1g --cpus 1 --pids-limit 128 \
  --user "$(id -u):$(id -g)" \
  -v "$PWD":/repo:ro -v "$PWD/tools/build/msc":/msc:ro \
  -v "$PWD/work/matching-decomp-20261008-r1":/out \
  --workdir /tmp --entrypoint python3 dq3-msc:bookworm-20261008-r1 \
  /repo/tools/run_matching_rng_adapter.py compile --output /out/rng-adapter-compile-new \
  --ida-evidence /repo/work/matching-decomp-20261008-r1/rng-abi-r1/ida-rng-abi-reviewed.json

timeout 330s docker run --rm --network none --memory 2g --cpus 2 --pids-limit 128 \
  --user "$(id -u):$(id -g)" -e PYTHONDONTWRITEBYTECODE=1 \
  -v "$PWD":/repo:ro -v /home/anr2/cht/dosgolem:/dosgolem:ro \
  -v "$PWD/work/matching-decomp-20261008-r1":/out \
  --workdir /tmp --entrypoint python3 hr-go-ebiten:1.26.7-2.9.9-r1 \
  /repo/tools/run_matching_rng_adapter.py cpu --output /out/rng-adapter-cpu-new \
  --compiled /repo/work/matching-decomp-20261008-r1/rng-adapter-compile-new
```

同一Go容器可執行 `audit --compiled /repo/work/matching-decomp-20261008-r1/rng-adapter-compile-new`
`--repeat-compiled /repo/work/matching-decomp-20261008-r1/rng-adapter-compile-r3`
`--cpu /out/rng-adapter-cpu-new`。重編來源及producer須保持相同，不能覆寫final-audit。
下一步限於既有NPC caller的AX／DX雙結果consumer局部比較，原版全局骰序、compiler與
完整EXE原碼重建仍unknown；不擴大remake完成閘門。

### 續行：BX／DX局部ABI閉合

目前具語意ASM覆蓋兩個已知函式、46bytes；C exact保持0。原版整體compiler與
全流程對拍加速仍未知。正式Go／game-pack及Issue #4暫停狀態保持。

| 原始定位 | 已閉合的局部契約 | 等級與來源 |
|---|---|---|
| `sub_1E6B9`，IDA linear`0x1e6b9..0x1e6c9`，file`0xfa29..0xfa39` | DS:0B5A加9018h後做三次一位左旋，更新word並留在AX；保留其餘通用／段暫存器，near return令SP增加2 | confirmed local ABI；IDA原始bytes、全部65536seed、定義明確的旗標模型及原版／ASM同狀態比較 |
| `sub_1E6C9`，IDA linear`0x1e6c9..0x1e6e7`，file`0xfa39..0xfa57` | BX為unsigned上限；BX非零時先更新狀態，再以DIV BX把商留AX、餘數留DX。BX=0返回DX=0並保留AX／狀態，不消耗亂數 | confirmed local ABI；BX10／BX0各65536seed、84邊界與保留暫存器比較。DIV未定義旗標不補硬體期望值 |
| NPC callers，IDA linear`0x12040/0x12062/0x12071/0x120ad` | 前置直接MOV BX為4／4／20／10，caller消費DX或DL；`0x12083`另消費AL商值。沒有把上限當成stack argument | confirmed scoped static caller-consumer；四個窗口與完整NPC函式原始bytes／typed xref。既有商值轉向不重開 |

新IDA匯出包含RNG核心33及有界RNG49個直接caller窗口；128窗口的診斷上限沒有截斷。
所有原始名稱、位址、bytes保留，不由名稱猜測型別。
五筆審查註記在`ida_rng_abi_ledger.json`記錄輸入hash、原operand／bytes、consumer、
推論等級與動態收據；重新建database後自動合併五筆，舊NPC註記保持。

`rng-abi-r1/cpu-r3/receipt.json`共196692個固定局部呼叫／每側，原版與source-assembled
執行共2885416指令，約0.568秒；乾淨建置約18.71秒。沒有新遊戲暖機，故這是明示的
direct-entry、記憶體／暫存器注入與CPU重入，不能稱正常玩家驗收，也不控制production RNG。
使用dosgolem `a9714ebdab2ad6b529f81225472680f2b11f2842`的internal非測試來源；
來源canonical SHA256 `334a21511c2603ef606b889c960d17ff5099bad7e3b11ba67aa5ca703eddf2b0`。
上游`cmd/probe/main.go`的使用者修改未複製、未改動。

固定seed1357、BX10的實際結果為狀態／核心AX=`1b7d`，有界AX=`02bf`、DX=`0007`。
預先定義的「把AX當餘數」負對照被拒絕。原版／重編46bytes scaffold的全檔hash相同，
但其餘115236bytes仍保留原始輸入，沒有完整原碼聲明。

候選MSC的已知C控制樣本，BX10與stack argument BEEF故意不同：

| 控制樣本 | 實際執行結果 | 證據界線 |
|---|---|---|
| `unsigned abiword(unsigned limit)` | 從stack取得BEEF，AX=BEEF、DX保持2468；8-byte code，5指令 | 該掛載候選的標準C ABI，不代表原版compiler |
| `unsigned long abipair(unsigned limit)` | 從stack取得BEEF，DX=BEEF、AX=5F77；14-byte code，8指令 | 該候選的DX:AX長值回傳；不能由原版DX餘數反推原版C return type |
| `_fastcall`單參數候選 | compiler報C2054／C2061，未產生OBJ | 此關鍵字／旗標組合不支援；未排除其他pragma、compiler或手寫ASM |

這項證據推翻「只調BP frame omission即可讓原先一般C介面匹配有界RNG」的充分性。
目前確定的是局部寄存器契約不同，沒有證明原版全程用哪個語言。
當時下一步以這兩個精確fixture作ABI-aware C／ASM adapter與compiler指紋的控制基準，
不修改被Inertia語意驗證拒絕的C來製造成功。
此局部adapter實驗現已由上節閉合；它通過register/state比較，但不是完整memory或byte-match。

原版／研究輸出均維持現行位置。新收據根為
`work/matching-decomp-20261008-r1/rng-abi-r1/`：`ida-rng-abi-reviewed.json`、
`assembly-receipt.json`、`cpu-r3/receipt.json`、`cpu-r3/producer-meta.json`、
`msc-abi-controls-r2/receipt.json`及`final-audit.json`。
首輪observer與compiler marker失敗保留：WatchWrites只報值變動，不計相同值重寫；
DOSBox IF重導向會先建立空ERR檔，應讀FAIL內容。修正的是探針契約，沒有修改引擎或遊戲。

已有唯讀輸入與新輸出目錄時，局部CPU實驗可按下列命令重跑。Go容器沿用既有
`hr-go-ebiten:1.26.7-2.9.9-r1`；先核對dosgolem來源及全部host掛載存在，輸出不得覆寫：

```bash
timeout 330s docker run --rm --network none --memory 2g --cpus 2 --pids-limit 128 \
  --user "$(id -u):$(id -g)" -e PYTHONDONTWRITEBYTECODE=1 \
  -v "$PWD":/repo:ro -v /home/anr2/cht/dosgolem:/dosgolem:ro \
  -v "$PWD/work/matching-decomp-20261008-r1/rng-abi-r1":/out \
  --workdir /tmp --entrypoint python3 hr-go-ebiten:1.26.7-2.9.9-r1 \
  /repo/tools/run_matching_rng_abi.py --output /out/cpu-new --assembly-root /out \
  --dosgolem-root /dosgolem --compiler-controls /out/msc-abi-controls-r2
```

本輪目標是驗證局部 matching 能否減少對拍的人工定位成本。正式 Go／game-pack、
正常458 checkpoint、schema0.33.0/content0.1.106均保持，Issue #4仍暫停。
下方2026-06歷史嘗試的原版 compiler 身分及單函式 C exact 聲明，以本節勘誤為準。

| 範圍 | 結果與推論等級 |
|---|---|
| 原始輸入 | `assets_raw/DQ3.EXE`，115,282 bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 入口與殼 | IDA9.4入口linear `0x19299`、logical `0x9299`、file `0xa609`，22筆原始指令可回查。常見packer標記零命中；殼的存在仍unknown，本切片沒有另做解殼 |
| IDA證據 | `ida-probe-r3.json`：入口、RNG與兩個NPC函式共216筆指令；828筆自動函式清冊僅供導航。`ida-npc.json`另重生既有六範圍436筆指令及分級註記 |
| Hex-Rays限制 | 四個明確target都回報 `16-bit functions cannot be decompiled`。這是IDA反編譯能力限制，沒有取消其bytes／xref主要證據地位 |
| 真實OMF | 按PUBDEF／SEGDEF選段，解析FIXUPP。RNG兩筆外部 `_g_rng` 引用以明示DS offset `0x0b5a`還原；此名稱是候選C的符號，不是原版符號 |
| 新編C | 七候選全部產生OBJ；四組完成重定位後比較，全部DIFF。另三組因未知符號placement或未寫入段尾拒絕。C exact為0；缺資料沒有猜補 |
| C迭代耗時 | 最終`msc-r3/receipt.json`保存單次DOSBox批次；七檔約0.92秒，cycles固定100000。這是本輪工具耗時，沒有等價的舊流程實測，不宣稱整體對拍加速倍數 |
| 具語意ASM | `sub_1E6B9`七條指令，linear `0x1e6b9..0x1e6c9`、logical `0xe6b9..0xe6c9`、file `0xfa29..0xfa39`。NASM2.16.01重新組譯16bytes與原版完全相同，instruction reproduction為confirmed；不含`db` |
| 拒絕案例 | 最終MSC探針核對缺placement、未知PUBDEF、不支援far fixup、壞OMF checksum及修改ASM常數五案，全部拒絕 |
| 整檔scaffold | SHA-256相同，但只有16bytes由具語意ASM重編；C重編bytes為0，其餘115,266bytes保留原始輸入。完整原碼重建仍unknown |
| Inertia | commit `c555363b810d3a6df786e5d6511d1bb28fa82333`、Python3.14.7及uv.lock固定，另以hash固定Cython3.2.0。CLI實際MZ base為`0x10000`，RNGlinear為`0x1e6b9`；原始file bytes核對後才做生成C及語意審查 |
| Inertia實測 | 修復環境及探針後的`inertia-r6/receipt.json`，三案都抵達反編譯並產出中間C；各約54.87／55.53／54.22秒，全部exit4、validation failed、merge gate hold。可採用C函式為0，中間C沒有送入重編或production |

舊文件的「Microsoft C 5.x已鎖定」保持為hypothesis。候選OBJ的`MS C`浮水印屬於
候選工具產物，不能反推原版compiler。RNG原版的三次`rol ax,1`與候選C的
`mov cl,3; rol ax,cl`不同；nested `_rotl(...,1)`及shift候選分別為18與28bytes，
原版為16bytes。這些差異已在明示fixup後核對，沒有把未知位址遮罩成exact。

Inertia環境與位址勘誤：r1缺native Cython lifter，r2的Python reference模式仍缺
上游uv.lock漏列的Cython runtime；r3修復依賴後遇到source內固定快取路徑的權限問題。
只對`.inertia_decomp_cache`掛明確可寫輸出，source維持唯讀，沒有改成root執行。
r4另暴露探針錯誤：底層DOSMZ default為linear`0x1000`，CLI會將paragraph選項轉成
linear`0x10000`。r4的三個目標因此錯移，屬驗證腳本問題，不是反編譯能力證據。
修正後smoke與實驗都使用CLI自己的`_build_project`，不再套用底層default。
所有失敗與各自收據保留，沒有覆寫成成功樣本。

r5的CLI loader smoke另缺必要backend／DOS SimOS註冊；補齊初始化後的r6才是
實際反編譯能力樣本。r6在RNG與有界RNG的postprocess發現未初始化的DS carrier，
並回報stack facts已分類卻沒有materialize；中間C亦可見unsigned-return函式缺回傳值。
NPC mover同樣由語意驗證拒絕。保留這些工具限制，不修改中間C掩蓋失敗，
也不把它們當成原版D2／D3資料來源。

本輪已證實有限工具迭代可重跑，但沒有證明整體對拍加速。下一步限縮為已匯出的
RNG／BX有界RNG之caller、暫存器輸入／輸出與compiler候選辨識；先解開原碼／ABI契約，
再增加C候選。完整EXE原碼重建仍不作本輪完成聲明，Inertia的三個中間C只供診斷。

本機收據位於`work/matching-decomp-20261008-r1/`：`ida-probe-r3.json`、
`ida-npc.json`、`ida-rng-wrapper.json`、`omf-parser-receipt-r2.json`、
`semantic-asm-r1/receipt.json`、`msc-r3/receipt.json`、`msc-final-audit.json`、
`inertia-r6/receipt.json`及`inertia-final-audit.json`。原版、database、OBJ、binary與
生成C均不加入Git。工具來源與入口保存於上述受版控腳本及`tools/build/README.md`。

最終MSC實驗可在相同容器中重跑。輸出目錄必須是新路徑；原始輸入與MSC目錄唯讀，
研究輸出才可寫，先確認每個host來源存在及UID/GID：

```bash
timeout 190s docker run --rm --network none --memory 1g --cpus 2 --pids-limit 128 \
  --user "$(id -u):$(id -g)" \
  -v "$PWD":/repo:ro -v "$PWD/tools/build/msc":/msc:ro \
  -v "$PWD/work/matching-decomp-20261008-r1":/out \
  --workdir /out --entrypoint python3 dq3-msc:bookworm-20261008-r1 \
  /repo/tools/run_matching_probe.py --output /out/msc-new \
  --ida-evidence /out/ida-probe-r3.json --msc-root /msc
```

Inertia實驗必須把特定快取目錄掛成可寫輸出，其他source唯讀。先建立並核對
`work/matching-decomp-20261008-r1/inertia-cache-r4/`的UID/GID；新輸出例如`inertia-new`：

```bash
timeout 260s docker run --rm --network none --memory 3g --cpus 2 --pids-limit 128 \
  --user "$(id -u):$(id -g)" -e PYTHON_JIT=1 -e PYTHONHASHSEED=0 \
  -e PYTHONDONTWRITEBYTECODE=1 -v "$PWD":/repo:ro \
  -v "$PWD/work/matching-decomp-20261008-r1":/out \
  -v "$PWD/work/matching-decomp-20261008-r1/inertia-cache-r4":/opt/inertia/.inertia_decomp_cache \
  --workdir /tmp --entrypoint /usr/bin/nice dq3-inertia:py3147-c555363b-r2 \
  -n 10 /opt/venv/bin/python /repo/tools/run_inertia_matching_probe.py --output /out/inertia-new
```

## 2026-06 歷史嘗試

正路 (b):把 seg0 函式逐一寫成「用 **MSC 5.1** 編出 byte-identical 原版機器碼」的 C。
本篇記錄第一批 leaf 函式的 byte-match 嘗試、可重複的 workflow、各函式 match% 與殘差成因,
以及下一批的建議。

> 前置:編譯器已鎖定 **Microsoft C 5.x small/near model**(指紋見 `docs/17`、`docs/19`)。
> 整檔已用 nasm 以 db 形式 byte-identical 重組(`docs/17` §2,sha256 相同)。本篇是更高一層:
> 不用 db 固定原版 bytes,而是**從 C 重新編出**相同機器碼。

## 結論先講

- 本批挑了 5 個最簡單的 leaf 函式(純全域 / 旋轉 / port out / memcpy / 填表迴圈)。
- **沒有一個達到 100% byte-identical**;但對「無 local、無迴圈」的直線函式,**指令選擇與結構已逼近原版**
  (`sub_e91e` frameless、長度精確相同、指令集完全一致,僅首對 `mov dx`/`mov ax` 載入順序相反)。
- 卡點全部歸類清楚(見「殘差分析」),都是 **MSC 5.1 codegen 的固定選擇**(旋轉編碼、intrinsic 引數求值順序、
  memcpy 策略、**有 local 必開 BP frame**),不是 C 寫法錯誤。
- 最關鍵的單一卡點:**這顆 MSC 5.1 build 對任何「含 local 變數」的函式都會發 `push bp; mov bp,sp` 並把
  變數配到 stack/SI/DI**,與原版 280 函式全為 frameless + `LOOP` + BX 定址的指紋衝突。解掉這個,迴圈類函式可大幅提升。

## 可重複 workflow

```bash
# 0) 一次性:建一顆帶 capstone 的 python image (host 不污染)
#    Dockerfile: FROM python:3.12-slim; RUN pip install capstone   → tag dq3-recap

# 1) 讀原版該函式反組譯
docker run --rm -v "$PWD":/w -w /w dq3-recap python3 tools/re_match.py orig <seg0_off_hex> <size>

# 2) 寫等價 C → re/match/sub_XXXX.c

# 3) MSC 5.1 編 (DOSBox in docker)
tools/build/msc_compile.sh re/match/sub_XXXX.c /c /AS /Ox   # → re/match/sub_XXXX.as.msc.obj

# 4) 抽 OBJ code bytes 與原版逐 byte 比對
docker run --rm -v "$PWD":/w -w /w dq3-recap python3 \
    tools/re_matchbatch.py <seg0_off_hex> <size> re/match/sub_XXXX.as.msc.obj
#  或整批:re_matchbatch.py manifest tools/re_match_manifest.json
```

`tools/re_matchbatch.py` 同時報兩個數字:
- **raw match%**:逐 byte 完全相同的比率。
- **masked match%**:把「16-bit 絕對位址運算元」(全域 / 陣列的 linker 指派位址,屬 data reloc 差異,
  原版位址 ≠ 我們的 linker 位址)遮罩後的比率 — 用來看「指令序列 / opcode / 結構」是否一致,
  排除掉「位址不同」這種與 codegen 無關的差異。

> 注意:`msc_compile.sh` 用 DOSBox,單次約 30–60s 且序列化(同時間只跑一個 DOSBox)。
> 產出檔名只依 model tag(`/AS`→`.as.`),**不含優化 flag**,故換 flag 重編會覆蓋同名 obj。

## 本批結果

| 函式 | seg0 | size | raw% | masked% | 用的 flag | 狀態 |
|---|---|---|---|---|---|---|
| `sub_e6b9` | 0xe6b9 | 16 | 31% | 62% | `/AS /Ox` + `#pragma intrinsic(_rotl)` | 頭尾同碼,旋轉編碼殘差 |
| `sub_e91e` | 0xe91e | 20 | 70% | 70% | `/AS /Ox` + `#pragma intrinsic(outpw)` | frameless、長度相同、指令集全同,僅首對載入順序 |
| `sub_edab` | 0xedab | 18 | 0% | 23% | `/AS /Ox` + `#pragma intrinsic(memcpy)` | memcpy 策略不同(見下) |
| `sub_e96d` | 0xe96d | 21 | 2% | 7% | `/AS /Ox` | 迴圈,BP frame + SI/DI 配置殘差 |
| `sub_32a3` | 0x32a3 | 17 | 0% | 9% | `/AS /Ox`(`/Gs` 無差) | 迴圈,BP frame + SI/DI 配置殘差 |

對齊最好的兩個:

```
sub_e91e (VGA GC 暫存器初始化, 4 x outpw):
  orig: bace03 b80300 ef b80508 ef b80700 ef b808ff ef c3   (mov dx,3ce; mov ax,3; out; ...)
  ours: b80300 bace03 ef b80508 ef b80700 ef b808ff ef c3   (mov ax,3; mov dx,3ce; out; ...)
        ^^^^^^^^^^^^^^  只有首對 mov 順序相反;其後 14 bytes 完全相同

sub_e6b9 (LCG 亂數前進):
  orig: a1 5a0b | 05 1890 | d1c0 d1c0 d1c0 | a3 5a0b | c3   (mov ax,[g]; add; rol×3; mov [g]; ret)
  ours: a1 0000 | 05 1890 | b103 d3c0       | a3 0000 | c3   (mov ax,[g]; add; mov cl,3+rol cl; mov [g]; ret)
        頭 (mov ax,[g]; add ax,0x9018) 與尾 (mov [g],ax; ret) 同碼;
        位址 5a0b vs 0000 屬 data reloc;只有中間旋轉編碼不同。
```

## 殘差分析(逐一歸類,均為 MSC 5.1 codegen 固定選擇)

1. **旋轉編碼(`sub_e6b9`)**
   原版三個 `rol ax,1`(`d1c0 d1c0 d1c0`,6 bytes,each = D1 /0 立即 1 形式)。
   MSC 5.1 對 C 旋轉慣用語 `(x<<3)|(x>>13)` **不認得**,會展成 `shl/shr cl + or`(更長)。
   用 `_rotl` intrinsic 才接近,但 intrinsic 一律 `mov cl,n; rol ax,cl`(`b103 d3c0`,4 bytes),
   即使 `_rotl(x,1)` 也不發 `d1c0`。
   → **MSC 5.1 的 `_rotl` 不會發 1-bit 立即旋轉**;原版那三個 `rol ax,1` 極可能來自 **inline asm**
   或更晚版本 MSC codegen。這顆 5.1 純 C 無法重現該 6-byte 形式。

2. **intrinsic 引數求值順序(`sub_e91e`)**
   `outpw(port, val)` intrinsic 在 `/Ox` 下先把 `val` 載入 ax、再把 `port` 載入 dx(右到左 cdecl 求值),
   原版相反(先 `mov dx,port` 再 `mov ax,val`)。其後三次 out 因 dx 不變、只重載 ax,**完全對齊**。
   → 只差首對兩條 `mov` 的順序;intrinsic 求值順序非 C 端可控(試 `/Os` 反而退化成真的 call)。

3. **memcpy 策略(`sub_edab`)**
   原版 = `push es; push ds; pop es; mov si,[src]; lea di,[dst]; mov cx,0x30; rep movsb; pop es`
   (純 byte 複製,不存 si/di,顯式 ES=DS)。
   MSC `memcpy` intrinsic = `push si/di` + word/byte 拆分(`shr cx,1; rep movsw; adc cx,cx; rep movsb`)。
   → 兩種不同 memcpy 實作;原版那種 frameless + 不存 si/di + 純 movsb,像手寫或 `movmem` 風,
   非 MSC 5.1 memcpy intrinsic。

4. **BP frame + 暫存器配置(`sub_e96d`, `sub_32a3`)— 最關鍵卡點**
   原版迴圈函式:**無 BP frame**、計數放 **CX 用 `LOOP`**、index/指標放 **BX/DI**、累加值放 **AX**。
   這顆 MSC 5.1 build:**只要函式有 local 變數就一定發 `push bp; mov bp,sp; sub sp,N`**,
   且把 register 變數配到 **SI/DI** 並以 `dec+jne` 取代 `loop`,結尾還把變數 spill 回 frame。
   試過 `register` 關鍵字、`/Gs`(無 stack probe)均無法去掉 frame。
   → 這與原版 280 函式 **全 frameless + `LOOP` 184 次 + BX 定址** 的指紋直接衝突。
   推測原版的 frame omission 來自某個尚未找到的 flag(或不同 build 的 MSC 5.x)。
   **解掉這點,所有迴圈類 leaf 函式可望大幅提升甚至 100%。**

5. **資料絕對位址(全函式共通,非 codegen 差異)**
   全域 / 陣列的 16-bit 絕對位址(如 `0xb5a`、`0x265d`、`0x289a`)由 linker 指派,與原版不同。
   單獨編譯(`/c`)的 obj 此處為 `0000` 並帶 reloc record。這屬 **data relocation 差異**,
   待整體 link、佈局對齊原版 DGROUP 後可一併消除;`re_matchbatch.py` 的 masked% 已把這類位元組排除。

## 下一批建議

1. **優先攻破 frame omission**:這是迴圈類函式 byte-match 的總開關。可試方向:
   - 比對原版某個「確定有 local」的函式,逆推它用的 MSC flag 組合;
   - 試 `CL` 其他優化 flag(`/Ol` loop opt、`/Oa`、`/Oc`)或不同 MSC 5.x sub-version;
   - 確認 docs/17 §7 統計的 frameless 是否其實對應「source 無 local」的函式子集。
2. **直線無 local 函式**先做(最好控):VGA / port I/O 初始化序列(`out` 系列)、純全域指派、
   常數表寫入。`sub_e91e` 已證實 frameless + intrinsic 路可行,只差求值順序。
3. **避開 intrinsic 求值順序問題**:找只有「單一 out / 單一 memcpy」或載入順序無歧義的函式。
4. **旋轉 / 特殊指令**:若原版大量用 `rol r,1`,評估是否原始碼即含 inline asm;此類保留 db 形式
   (見 docs/17 §6 後續 1),不強求純 C。
5. **data reloc 收斂**:待累積數個函式後,做一次「整批 link + 對齊 DGROUP 佈局」實驗,
   把絕對位址也對上,屆時 raw% 才能反映真實 byte-identical。

## 受阻誠實揭露

- 本批 **0 個達 100%**。最接近 `sub_e91e`(frameless、長度相同、指令集全同,僅首對載入順序)、
  `sub_e6b9`(頭尾同碼)。
- 主因是 **MSC 5.1 codegen 固定選擇**(旋轉編碼、intrinsic 求值順序、memcpy 策略、**有 local 必開 frame**),
  非 C 判讀錯誤。
- workflow(讀原版 → 寫 C → MSC 5.1 編 → 抽 obj 比對 + masked%)**已可重複**,
  工具 `tools/re_matchbatch.py` + `tools/re_match_manifest.json` 一鍵整批。
- 推進 100% 的單一最高槓桿:**找回 MSC 5.x 的 frame-pointer omission 行為**(見「下一批建議」1)。
