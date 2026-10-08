# Matching decompilation 研究工具鏈

本目錄支援局部 C／具語意 ASM 的重編與位元組比較。正式產品是
`dq3_remake_ebitan/` Go／Ebitengine remake；研究不另設 C／SDL 產品目標。
工作範圍、最新結果與本機收據入口見 [Issue #5](https://github.com/wicanr2/kinginformation-dq3-re/issues/5)
及 [docs/25](../../docs/25-match-progress.md)。原版 compiler 家族、版本與旗標仍待證實。

## 現行入口

| 入口 | 契約 |
|---|---|
| [IDA 探針](../ida_matching_probe.py) | 在 IDA9.4 一次性 database 匯出原始名稱、file／loaded bytes、typed xref與分級語意；缺邊界明列unknown |
| [MSC 探針](../run_matching_probe.py) | 單次DOSBox批次編譯、真實OMF檢查、NASM組譯與拒絕案例；C／ASM／保留原始bytes分開計算 |
| [OMF reader](../omf_matching_probe.py) | TIS OMF1.1的有限PUBDEF／SEGDEF／FIXUPP子集；未知placement、缺失段bytes及不支援的fixup拒絕，不設遮罩fallback |
| [Inertia 探針](../run_inertia_matching_probe.py) | 先核對MZ base及原始code，再對明確函式限時生成C。保留tail validation，生成C不代表byte-match |
| [RNG ABI 探針](../run_matching_rng_abi.py) | [Go局部probe](../dosgolem_matching_rng_abi.go)核對固定seed全集、暫存器、near return及memory delta；只讀dosgolem internal來源，不使用上游cmd/probe scratch |
| [MSC ABI 控制](../run_msc_abi_controls.py) | 編譯已知16／32-bit回傳與fastcall候選，與原版BX／DX契約分開記錄；原版compiler不由此定案 |
| [C／ASM adapter](../run_matching_rng_adapter.py) | [局部CPU探針](../dosgolem_matching_rng_adapter.go)與[轉接組語](../../re/match/rng_adapter.asm)核對固定輸入、暫存器／持久狀態及額外堆疊／旗標差異；數值通過不升格byte-match |
| [完整IDA清單](../ida_matching_inventory.py) | 所有code heads／function chunks、原始與loaded bytes、typed xref及舊清單差異；自動邊界仍需審查 |
| [完整Goal C批次](../run_matching_c_batch.py) | [候選清單](../matching_c_manifest.json)保存原始定位與DS placement；用compiler PROC／ENDP及實際OMF核對整個C函式，module padding分開記錄 |
| [完整Goal證據核對](../matching_goal_audit.py) | [契約](../matching_goal_contract.json)保存使用者選定的標準；核對完整清單、兩份C artifact及listing正反對照，manual gate保持未完成 |
| [邊界與間接入口](../ida_matching_boundaries.py) | fresh IDA讀出原始flag常數、22末端、三個舊entry與33個indirect call operands；不修database邊界 |
| [邊界分級](../review_matching_boundaries.py) | typed flow／實際DOS service writer與SI取址consumer分級，保留DS條件、未知DI／field與未批准source-unit範圍 |
| [多入口CFG](../matching_cfg_candidates.py) | 沿原始typed edges追到return／未解successor，保留外部entry與共用code，不自動併成C source unit |
| [Watcom16 register控制](../run_watcom16_abi.py) | 六個instruction-free C pragma控制；固定source path及-zld確保真實OBJ重現。實驗exact不自動加入正式coverage |
| [Watcom原始範圍來源](../watcom_matching_manifest.json) | 配合run_watcom16_abi.py的candidate模式，核對完整IDA範圍、source hash及原始bytes；主程式角色指標查詢／48-byte記錄writer及SDK AX store exact，fill／search仍DIFF |
| [WCC有號位移控制](../verify_watcom_signed_fixup.py) | vendor WLINK連結實際C object與synthetic DATA，保留map的frame bias；六無效位移拒絕，不當原版layout |
| [Watcom主程式codegen](../probe_watcom_primary_codegen.py) | 填表／搜尋126組C與loop／reorder flags，完整artifact保留；long-shift正對照實際發LOOP，原版候選全DIFF |
| [TC2.01主程式codegen](../probe_turboc_primary_codegen.py) | 固定本機archive／compiler hash，用既有DOSBox比較八組C；四DIFF／四group-frame REFUSED。生成C mtime固定以重現整個OBJ |
| [SBCM原始module](../probe_sbcm_modules.py) | 固定原版LIB hash，26個OMF metadata／LIDATA parser view與原始xref導航；排除fixup的候選不算match |
| [SDK完整身分控制](../link_sbcm_module_controls.py) | 官方WLINK重連原版CTVMEM／CMFDRV objects，7789 bytes與EXE一致；source coverage增量0，objects不作最終原碼 |
| [SDK原版IDA refs](../ida_matching_module_refs.py) | 固定whole-module proof後匯出原名、bytes、functions、typed refs及startup；保留library metadata與原始IDA定位 |
| [Wasm revision準備](../prepare_watcom16_asm.py) | 從已驗證完整archive clone r1 payload、加入官方Wasm，763檔；r1 inputs不修改 |
| [SDK指令source重建](../verify_sdk_instruction_sources.py) | [範圍manifest](../sdk_instruction_source_manifest.json)核對repo ASM／EQU、實際written mask與FIXUPP；5148 bytes exact，ORG data不計source |
| [SDK data審查](../review_sdk_data_regions.py) | 七個handler tables及IRQ stack閉合；CTV slot6原始6B06保留，E2結果由固定成熟模擬器source契約推導 |
| [完整CMF source/data](../verify_cmf_source_module.py) | code／EQU與[typed data](../../re/match/cmfdrv_data.json)重建5296 bytes，code/data範圍分離；field semantics仍unknown，不用原始objects |
| [完整SDK source/data](../verify_sdk_source_module.py) | common verifier選CTVMEM／CMFDRV，各重建2493／5296 bytes；CTV [typed data](../../re/match/ctvmem_data.json)保留E2 seed與unknown payload，舊CMF CLI相容 |

## 映像與輸入

| 映像 | 建置來源與支援 |
|---|---|
| `dq3-msc:bookworm-20261008-r1` | [Dockerfile.msc](Dockerfile.msc)：Debian基底digest與2026-10-01 snapshot固定；DOSBox0.74-3、NASM2.16.01、Python3.11。MSC binaries不寫入image |
| `dq3-inertia:py3147-c555363b-r2` | [Dockerfile.inertia](Dockerfile.inertia)：Python3.14.7、Inertia commit `c555363b810d3a6df786e5d6511d1bb28fa82333`及uv.lock固定；另補上游鎖檔缺少的Cython3.2.0，Linux x86_64 wheel URL及SHA-256固定。r2取代r1 |
| `dq3-watcom16:2.0-20261001-r1` | [Dockerfile.watcom16](Dockerfile.watcom16)以既有Python固定digest及[完整官方archive verifier](../fetch_watcom16.py)產生的payload建置；補足既有wcc386-only image。16-bit wcc／wdis／wlink／wlib與headers固定，instruction-free register ABI實測入口[run_watcom16_abi.py](../run_watcom16_abi.py)。不取代原版compiler身分證據 |
| `dq3-watcom16:2.0-20261001-r2` | 同Dockerfile／runtime加入固定官方Wasm，供MASM相容SDK source；[prepare_watcom16_asm.py](../prepare_watcom16_asm.py)保存revision來源，保留r1 C控制image |

原始遊戲放在`assets_raw/`。既有MSC候選工具位於gitignored的
`tools/build/msc/BIN/`、`LIB/`及`INCLUDE/INCLUDE/`；探針驗證CL／C1／C2／C3的固定雜湊。
這些binaries、OBJ、原版素材、IDA database與生成C只留本機。

Inertia source使用固定checkout。本機現行副本在
`work/matching-decomp-20261008-r1/inertia/`。Dockerfile要求上述commit；uv.lock之外的
Cython wheel以`--require-hashes`驗證。研究runner明示使用上游
`INERTIA_VEX_BACKEND=python` reference模式，未建置native Cython lifter。

## 建置與執行

Watcom16由[fetch_watcom16.py](../fetch_watcom16.py)在有network的受限Docker容器下載固定
release，完整archive hash通過後才建立payload。掛載前核對來源存在、形態及UID/GID。
在專案根目錄從已驗證payload建置revision：

```bash
timeout 120s docker build --network none -f tools/build/Dockerfile.watcom16 \
  -t dq3-watcom16:2.0-20261001-r1 \
  work/matching-decomp-20261008-r1/full-goal-r1/watcom16-source-r1/payload
```

控制可在一次性、network none、UID/GID、1GiB／1CPU／128 pids的容器執行
`python3 /repo/tools/run_watcom16_abi.py --output /out/watcom16-abi-new`。
repository唯讀掛到/repo，既有matching輸出掛到/out；每次output須是新名稱。
此image補16-bit支援，保留其他專案既有wcc386 image，沒有全域清理。

正式來源候選以同一runner加上
`--candidate-manifest /repo/tools/watcom_matching_manifest.json`
與 `--ida-inventory /repo/work/matching-decomp-20261008-r1/full-goal-r1/inventory-ida-r2.json`。
每次重編使用全新容器，固定compile目錄不供同容器連續啟動兩次。
候選的compiler_profile只可選runner明列的固定旗標組合：預設cdecl-size-reorder、
記錄writer使用watcall-speed-no-reorder。manifest不接受任意shell或compiler命令，
audit會核對實際command；不同profile不表示已辨識原版compiler。
有號位移控制執行 `python3 /repo/tools/verify_watcom_signed_fixup.py`
加 `--candidate-obj /out/<候選目錄>/sub_132a3.obj --output /out/<新控制目錄>`。
總審核加 `--watcom-receipt` 與 `--repeat-watcom-receipt`；RESOLVED只表示可重定位，
僅原始全範圍byte_exact才能計入C覆蓋。入口與現況見docs/25。

主程式codegen以 `python3 /repo/tools/probe_watcom_primary_codegen.py`
或 `python3 /repo/tools/probe_turboc_primary_codegen.py`，加上
`--output /out/<新目錄> --ida-inventory /out/inventory-ida-r2.json`。
前者使用dq3-watcom16 image，後者沿用dq3-msc image並從repo唯讀TC輸入啟動；
兩者均使用上方一次性容器限制。它們比較完整範圍，不提供production build或source-unit批准。

SDK metadata用 `python3 /repo/tools/probe_sbcm_modules.py`
加 `--output /out/<新目錄> --ida-inventory /out/inventory-ida-r2.json`。
身分控制用 `python3 /repo/tools/link_sbcm_module_controls.py`
加 `--module-inventory /input/receipt.json --output /out/<新控制目錄>`，原始module目錄唯讀掛到/input。
兩者用既有dq3-watcom16 image。IDA查詢在既有IDA9.4 image對新database副本執行
`-S"/repo/tools/ida_matching_module_refs.py /out/<新sidecar>.json /out/<控制目錄>/receipt.json"`。
完整scope、hash與位址基準見docs/25；原版library／OBJ／database僅留本機。

Wasm r2準備使用r1 image在Docker跑 `python3 /repo/tools/prepare_watcom16_asm.py`
加 `--original-source /repo/work/matching-decomp-20261008-r1/full-goal-r1/watcom16-source-r1`
及 `--output /out/<新r2來源目錄>`。再以相同Dockerfile、network none建置r2，context為新payload。
SDK source在r2 image跑 `python3 /repo/tools/verify_sdk_instruction_sources.py`
加 `--layout /out/sdk-module-layout-r1.json --output /out/<新build目錄>`。
每次獨立容器、repository唯讀、明確output可寫；verifier不允許DB／data／include／macro代替指令。
產出的MZ只有指令區段可用作局部oracle；原版data仍缺，不能作完整driver或交付包執行。

完整CMF recipe用r2 image執行 `python3 /repo/tools/verify_cmf_source_module.py`
加 `--layout /out/sdk-module-layout-r1.json --output /out/<新CMF來源重建目錄>`。
它只從repo code／EQU／typed JSON編譯，驗證全5296 bytes；原EXE僅作比較，原SDK objects不作build inputs。
data schema拒絕missing／overlap／落入code／錯handler與value range。這是module byte-layout，
不代表全EXE完成或可獨立啟動的遊戲封包，後續仍按完整Goal recipe與正常oracle驗收。

完整CTV recipe用同r2 image執行 `python3 /repo/tools/verify_sdk_source_module.py`
加 `--module CTVMEM.ASM --layout /out/sdk-module-layout-r1.json --output /out/<新CTV來源重建目錄>`。
CMF可選 `--module CMFDRV.ASM` 或保留舊CLI。data gate包含strict seed literal、合法handler head、
字符串及code/data範圍，完整byte比較不接受machine-code拼接。

從專案根目錄建置MSC映像，stdin context不含原版或compiler：

```bash
timeout 1500s env DOCKER_BUILDKIT=0 docker build --force-rm \
  --memory 2g --cpu-period 100000 --cpu-quota 200000 \
  -t dq3-msc:bookworm-20261008-r1 - < tools/build/Dockerfile.msc
```

確認固定Inertia checkout存在後，以它作build context：

```bash
docker build -f tools/build/Dockerfile.inertia \
  -t dq3-inertia:py3147-c555363b-r2 work/matching-decomp-20261008-r1/inertia
```

執行前確認每個host掛載來源存在且型態正確，檢查輸出UID/GID；輸出參數必須指向新目錄。
所有執行使用`--rm --network none`、資源限制及目前UID/GID。原版與compiler唯讀掛載，
只有指定研究輸出可寫。完整指令與位址換算見docs/25；收尾核對本輪容器及輸出擁有權。

## 歷史工具

`msc_compile.sh`、`tcc_compile.sh`、`rebuild.sh`、`match_check.py`、
`Dockerfile.turboc`及`Dockerfile.sdl`保留供歷史追溯。它們不是本輪的執行入口；
舊「最大段／masked match」不能代替PUBDEF定位、真實FIXUPP與完整bytes比較。
舊compiler鎖定與單函式C exact的勘誤保留在docs/19、docs/25及WORKLOG。
