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

## 映像與輸入

| 映像 | 建置來源與支援 |
|---|---|
| `dq3-msc:bookworm-20261008-r1` | [Dockerfile.msc](Dockerfile.msc)：Debian基底digest與2026-10-01 snapshot固定；DOSBox0.74-3、NASM2.16.01、Python3.11。MSC binaries不寫入image |
| `dq3-inertia:py3147-c555363b-r2` | [Dockerfile.inertia](Dockerfile.inertia)：Python3.14.7、Inertia commit `c555363b810d3a6df786e5d6511d1bb28fa82333`及uv.lock固定；另補上游鎖檔缺少的Cython3.2.0，Linux x86_64 wheel URL及SHA-256固定。r2取代r1 |

原始遊戲放在`assets_raw/`。既有MSC候選工具位於gitignored的
`tools/build/msc/BIN/`、`LIB/`及`INCLUDE/INCLUDE/`；探針驗證CL／C1／C2／C3的固定雜湊。
這些binaries、OBJ、原版素材、IDA database與生成C只留本機。

Inertia source使用固定checkout。本機現行副本在
`work/matching-decomp-20261008-r1/inertia/`。Dockerfile要求上述commit；uv.lock之外的
Cython wheel以`--require-hashes`驗證。研究runner明示使用上游
`INERTIA_VEX_BACKEND=python` reference模式，未建置native Cython lifter。

## 建置與執行

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
