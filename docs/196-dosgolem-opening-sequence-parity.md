# 196 — dosgolem 開場順序對拍

日期：2026-10-01。工作依據：[GitHub Issue #1](https://github.com/wicanr2/kinginformation-dq3-re/issues/1)。

## 規格與證據審查

狀態：**CONFORMED（限六幕順序、前五幕色號及第六幕 129 個翻頁色號／RGB）**。
首次靜態第六幕對拍曾失敗並退回 DRAFT；經 IDA 補證與 READY 審查後接入七圖塊，
正式 bootstrap／InputState 對拍與存讀檔通過。淡入淡出逐相位與完整 campaign 不在此完成範圍。
目前程式 checkpoint 為 `0c6f780`；現行產品是 `dq3_remake_ebitan/`。

### 第六幕補證與 READY 審查（2026-10-01）

審查紀錄：第六幕格式／座標先達 **READY**，再經下方正式對拍進入限縮範圍的 CONFORMED。
IDA Pro 9.4 匯出入口為 `tools/ida_dump_boot_presentation.py`，sidecar 輸出
`/tmp/dq3-boot-ida.json`，原始定位、bytes、推論等級與輸入 hash 同列保留。
本節位址使用 IDA linear 與 MZ file offset：`file = linear - 0x10000 + 0x1370`；
DGROUP 資料使用 `file = 0x16140 + offset`。以下結論只適用上列固定 EXE。

| 原始定位 | 附加語意 | 推論等級／證據 |
|---|---|---|
| `sub_220F0`，IDA linear `0x220f0`／file `0x13460` | 載入 F 背景、DFP、動畫，再進 G 標題 | confirmed：caller／loader／自然開機讀檔 |
| `sub_222C4`，IDA linear `0x222c4`／file `0x13634` | 六圖塊按 byte 交錯四平面；OR 四平面作寫入遮罩，色號零透明 | confirmed：原始 consumer＋129 畫面逐點閉合 |
| `sub_22246`，IDA linear `0x22246`／file `0x135b6` | 第七圖塊按 row 的四平面連續，色號零仍覆寫 | confirmed：原始 consumer＋129 畫面逐點閉合 |
| DGROUP `0x5bde`／file `0x1bd1e`，原始 24 bytes `050002000d000900520009000000430008002b0052003100` | 六組 `(y, x_byte)`：`(5,2),(13,9),(82,9),(0,67),(8,43),(82,49)` | confirmed：動畫 loop 讀取＋129 畫面逐點閉合 |
| `sub_221A2`，IDA linear `0x221a2`／file `0x13512` | y=348，每次 -2，共 129 個位置至 92；第七圖塊 x_byte=9、y_offset=33；每位置等待 2 ticks | confirmed：loop、writer、consumer＋自然翻頁觀測 |
| `sub_224C4` 的 RET，IDA linear `0x224d5`／file `0x13845` | 翻頁完成觀測點；本次 load segment 下為執行器 `1319:0445` | confirmed：原始 bytes／執行期顯示頁；動畫 caller 返回 `1319:016b` |

`DFP.PIC`：24,435 bytes，SHA-256
`c3ad3b4ee95434f1601e519b1dfb16094184aa1393f2972136fac7e735299b00`。
七個 u32 offset 是 `0x1c,0xb81,0x126a,0x1c23,0x27d4,0x3199,0x3306`；
每段 u16 byte-width／height 後有 `width×height×4` bytes，另剩一個 byte 未被 consumer 讀取，
其語意 **unknown**，不猜成 terminator。資料不是 RLE 壓縮。

獨立解碼已與原版全部 129 個翻頁畫面逐色號一致。正式引擎將採具名有限疊圖契約：
素材引用、各圖塊平面排列／透明政策／座標及線性移動皆由 JSON 提供；不加入任意 JSON 程式碼。
正式對拍須由 bootstrap／InputState 依序走完，不能直接設定 opening index 或動畫位置。

時間採平台規格近似，不深挖 ISR：前五幕觀測 PIT divisor=65536，第六幕為 12428。
公式 `Hz=(315000000/264)/divisor`，見
[PIT 公開規格](https://wiki.osdev.org/Programmable_Interval_Timer)。前四幕淡入 45／停留 40／
淡出 45／空白 10 ticks，第五幕 45／100／45／0；第六幕淡入 45／動畫 258／停留 20／淡出 9。
以 60 次更新／秒的累積有理數取樣轉為邏輯幀，總幀數向上取整；取樣誤差小於一個更新幀。
這是 **hardware-spec approximation**，不聲稱 EXE 繪圖耗時或實機 wall-clock 完全一致。
淡入扣色階 `64,56,48,40,32,24,16,8,0`，淡出反序；各級等待 5 ticks，第六幕淡出為 1。
原始 consumer `sub_2271E`（IDA linear `0x2271e`／file `0x13a8e`）對各 DAC component
先 clamp 再扣除色階；`sub_22782` 以 `component×63/255` 轉成六位元色盤。

原版輸入 `assets_raw/DQ3.EXE`：115,282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
dosgolem 上游版本 `2f44a68ebfc54b28fb15dd4a34510b0b04a5415d`，Go 1.24.13，
Docker `dq3-ebiten-test:20260822-r1`。初次順序證據使用執行期指令步數與顯示卡畫面，
第六幕另按上節使用 IDA；沒有直接呼叫遊戲函式、修改 EXE／素材或注入狀態。

首次未修正的 dosgolem 執行 20,000,000 道指令，停在黑畫面；原版傳給 DOS 的
`TITA    .P  ` 無法對應已存在的 `TITA.P`。可丟棄工具副本將 `resolve` 的 basename
ASCII 空白去除後，原版自然執行並依序讀取下表。這是執行器工具修正，與 remake 修正分開。
平台依據是 [DOSBox Staging 的 DOS_MakeName](https://github.com/dosbox-staging/dosbox-staging/blob/main/src/dos/dos_files.cpp)：
該函式在解析 DOS 路徑時忽略 ASCII 空白。不是為 DQ3 建立檔名別名或假的圖片。

工具副本的原始 `internal/dos/files.go` SHA-256 為
`463a8a83315a3054af75d6220a8d5bcc1464021ea28bad15659e6b9d5a0ce69e`；
測試確認帶空白／大小寫／磁碟路徑的開檔與讀取，完整 `internal/dos` 回歸通過。
上游來源維持唯讀；正式收據必須明示此工具修正，不能寫成未修正上游已支援 DQ3。

| 順序 | 原版素材 | 內容 | 自然執行的擷取步數 | 推論等級 |
|---:|---|---|---:|---|
| 1 | `TITA.P` | 1989 | 20,000,000 | confirmed：執行讀檔與畫面 |
| 2 | `TITB.P` | 巨龍／勇者 | 100,000,000 | confirmed：執行讀檔與畫面 |
| 3 | `TITC.P` | 1990 | 200,000,000 | confirmed：執行讀檔與畫面 |
| 4 | `TITD.P` | 三人群像 | 300,000,000 | confirmed：執行讀檔與畫面 |
| 5 | `TITE.P` | 1993 | 400,000,000 | confirmed：執行讀檔與畫面 |
| 6 | `TITF.P`＋`DFP.PIC` | 背景與動態 DRAGON FIGHTER 標誌 | 500,000,000；另觀測全部 129 次翻頁 | confirmed：格式／consumer 與完整翻頁位置閉合 |

同次啟動第六幕讀取 `DFP.PIC`，後續讀取 `TITG.P` 及職業巡禮；沒有讀取 `TITP.P`。
本項 consumer 是原版自然開機至標題的讀檔／繪圖路徑，沒有把常式直接入口當成玩家路徑。
700,000,000 道指令的探測沒有缺檔，也沒有列出未實作 DOS 服務。

## remake 修正契約

- 共用引擎增加有限疊圖與有理數計時；schema `0.1.53 → 0.1.54`，欄位權威見
  [docs/84](84-game-pack-json-contract.md)。其他資料檔只有同步 schema，未改遊戲設定。
- 補 `opening_year_1990 → TITC.P`，第六張改成 `opening_logo → TITF.P`。
- 內容版號由 `0.1.59` 升至 `0.1.60`；canonical hash 由 loader 依內容重新計算。
- 舊 120 幀猜值由原始 tick 計數與 PIT 規格近似取代，六幕總幀數
  `462,462,462,462,627,208`；這些是可重生的換算結果，非實機精確時間。
- 正式 bootstrap 無輸入走完六張，正式 Confirm 仍交回標題／主選單。
- 不新增劇情旗標或存檔欄位；既有 pack identity 與存讀檔驗收維持。

勘誤：`docs/120` 的五張列表與 `TITP.P` 候選由此推翻；保留舊文與形成原因，
不能再把「曾經解碼一張圖片」當成該圖被原版開場消費的證據。

## 可重現驗證與限制

入口：`tools/verify_dosgolem_opening.sh`；工具副本建置與收據生成在
`tools/dosgolem_opening_probe.py`。原版與上游 dosgolem 都唯讀掛載，輸出位於既有
`work/dosgolem-opening/`，不入 Git。使用 `-dump-at` 的平面解碼 PNG／`.bin`；
`cmd/probe -shots` 保存線性原始資料，不能只因命名 `.png` 就當成平面畫面。

受影響驗收：game-pack schema／reference validation、實檔 SHA-256、前五張原版
4-bit 色號及第六幕 129 個完整畫面／RGB 與 production renderer 逐點比較、正常 bootstrap／InputState、
Confirm 跳過、主角存讀檔及 pack identity。

2026-10-01 正式工具重跑：前五張共 1,120,000 個色號與 production renderer 逐點通過；
第六張因缺少 `DFP.PIC` 疊圖失敗。不能改擷取時間或只取背景區域讓它通過。

追加驗證：七圖塊接入正式引擎後，前五幕 1,120,000 個色號及第六幕
28,896,000 個色號／RGB 全部通過，共 30,016,000 個比較像素。固定 500,000,000 步
的擷取仍保留；正式動畫驗收改為原版全部翻頁完成點，沒有裁切或省略失敗位置。
`work/dosgolem-opening/receipt.json` 保存原版 hash、上游 revision、兩項工具副本修正、
每次指令步數／IRQ0 ticks／PIT divisor／PNG與色號 hash；`comparison.json` 保存
原版收據 hash、資料包 canonical hash 與比較範圍。素材與圖片不入 Git。

IDA 可重現入口（唯讀原版、database 僅在一次性容器 `/tmp`）：

```bash
test -d /home/anr2/dq3 && test -d /tmp && timeout 120s docker run --rm \
  --network none --memory 3g --cpus 2 --pids-limit 192 -u "$(id -u):$(id -g)" \
  -v /home/anr2/dq3:/repo:ro -v /tmp:/out ida-pro-9.4-idapython:locked-v1 \
  bash -c 'idat -A -c -o/tmp/dq3-boot.i64 "-S/repo/tools/ida_dump_boot_presentation.py /out/dq3-boot-ida.json" /repo/assets_raw/DQ3.EXE'
```

腳本強制核對輸入 size／hash，自動合併受版控 file-offset 語意索引，同列保留
原始名稱／位址／bytes、推論等級及證據；UNKNOWN 醒目顯示，不改寫 database。

**尚未證實：**淡入淡出的逐相位 RGB、音效、精確 skip 時機、創角以後的 dosgolem
玩家路徑及完整 campaign。初次 RGB 比較有色盤亮度相位與 DAC 展開差異；即使色號完全
一致，也不升格整段開機播放 V3；第六幕的 129 個翻頁畫面只在同位置色號／RGB 範圍達 V3。
無輸入卡片與標誌合成不涉及亂數判定；不外推到需要固定 seed
的創角、遭遇或戰鬥。歷史 remake campaign E3 不等於已完成原版對拍。

完整回歸結果：全部 `internal/...` 通過；完整 `game` 有兩個失敗，修改前 `0c6f780`
在同一容器、同樣素材、同一測試命令亦重現相同失敗：
`TestOpeningProductionInputTrace` 在母親帶路後仍有對話的狀態不符合舊斷言；
`TestKandarTowerProductionRouteOpensThiefKeyDoor` 的 Lv1 fixture 遭遇 monster23 全滅。
這兩項不由本批開機修改造成，不能藉此猜改 production；另以
[Issue #2](https://github.com/wicanr2/kinginformation-dq3-re/issues/2) 追查。
目前不得寫成「完整 game 全綠」或重新宣稱現行 checkpoint campaign E3。
