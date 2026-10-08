# 逐函式 byte-match 進度 (matching decompilation)

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
