# 119 — 日夜 clock／NPC story flag audit（2026-08-10）

本文件是 `docs/71-npc-story-flag-visibility.md` 的現行盤點，不把「已有日夜機制」誤寫成
「21 個條件 NPC 旗標全部已接」。目標是把每個原版條件 flag 的
`入口／writer → game-pack state → NPC consumer → 玩家可見副作用` 分開記錄；沒有閉環的項目
維持 `unknown`，不以合理時序或攻略補值。

## 輸入與位址口徑

- 原始輸入：`assets_raw/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 原版靜態證據：IDA Pro 9.4（Docker 內，IDA linear `0x10000..0x2aee0`）；
  `docs/71` 的 SET 位址是 file offset，`[0x4f70]` 是 DGROUP logical data offset。
- runtime 證據：`dq3_remake_ebitan/game/game.go`、`worldmap.go`、各 pack event primitive；
  `docs/93` 的 raw clock／palette 對照與 `game/daynight_test.go`、
  `game/inn_daynight_test.go`、`game/encounter_gate_test.go`。
- 推論等級：`confirmed`＝原版 bytes／明確 writer-consumer 閉合；`strong`＝runtime 已有
  對應但同狀態玩家可見閉合仍待 V3；`unknown`＝只有 reader 或未找到 writer。

## 日夜機制先驗

### 2026-08-10 靜態 RE 勘誤（優先於本文件早期的「clock 未定位」描述）

IDA Pro 9.4 已在同一份 `DQ3.EXE`（115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`）閉合原版 clock／palette：

- `sub_1EE23`（IDA linear `0x1ee23`、logical `0xee23`、file `0x10193`）在
  `DGROUP:0x4f2d != 1` 且 `DGROUP:0x5051 != 1` 時遞增 `DGROUP:0x251d`。
- `0x251d==0x78` 寫 `0x526c=0`；`0x251d==0xf0` 歸零並寫 `0x526c=1`，有效範圍為
  `0..0xef`。因此 `[0x526c]` 是 120／240 tick 邊界的二值日／夜表 selector，而非
  現行 Go 四 phase 名稱本身。
- `sub_1EE76`（file `0x101e6`）以 `clock/0x14` 查 `DGROUP:0x25c5` 的 12-byte
  table（file `0x18705`：`01 00 00 00 01 02 03 04 04 04 03 02`），再由
  `DS:0x3232+index*0x30` 選 `DQ3.PAL` bank 並呼叫 `sub_20A3A`；`sub_1EE9B`
  做相同選擇供場景載入。`sub_1ECDC` 載入 `DQ3.PAL` `0xf0` bytes 與 `MNSBK.PAL`
  `0x1e0` bytes。

這一段把「原版值未定位」修正為 **原版靜態值已 confirmed，runtime 同狀態 palette V3
仍待**。舊段落保留作時間序列，不得再當目前結論。完整函式／table／推論等級見
[`docs/123`](123-static-battle-daynight-re.md)。

現行 engine 已有可到達的機制，這一輪不重寫：

1. `dnPhaseSteps=60`；phase 0/1/2/3 分別為日／黃昏／夜／黎明，240 個有效移動步完成一輪。
2. 移動呼叫 `advanceDaynight`；`isNight` 只在 phase 2 成立；旅社成功交易重設 phase／step，
   黑暗之燈使用原版 raw clock `0x008c` 並呼叫 `reloadTownDaynight`。
3. `applyDaynightPalette` 以 `DarkenPalette(worldPal, phase)` 更新地表、地下與城鎮 palette；
   `loadTownSceneSec` 依 phase 選日／夜 NPC 表，再以 `[0x4f70]` story flag 過濾每筆 record。
4. save/load、住宿付款失敗不改時鐘、夜間遭遇範圍與正式開場 trace 均有 component／input tests。

剩餘不是「沒有日夜」，而是原版 `DGROUP [0x526c]` day selector、`[0x251d]` clock 與 palette
register 的同狀態 V3 對拍，以及下表中尚未閉合的條件 NPC 事件。

Docker＋Xvfb `TestDumpDayNight -v`（`DQ3_ASSETS=/repo/assets_raw`）實測阿里阿罕 CTY00
sec0：白天 15 個 NPC、黑夜 6 個 NPC；產物雜湊為 `aliahan_day.png`
`47a75ce5be72f772da067ee12ee9ab5ad847896931fe2db94a523336c6ca3cc7`、
`aliahan_night.png` `d41adc903efa5cc2d708bce306a9581d0feb6664757c7f183cd06c0a61db7120`。
黑夜主要色票由 `(101,146,0)` 變為 `(42,64,0)`，但這只是 runtime darken 的可重現 V2
證據；沒有同一 raw 輸入的原版 PNG，不能升格為 palette V3。

## `docs/71` 的 21 個條件 flag

| flag | 原版 SET（file）／用途線索 | runtime 現況 | 判定與下一步 |
|---|---|---|---|
| `0x0d` | `0xfab0`；後期條件 NPC | `story_flag_runtime_events.post_zoma_ending` 由 CTY80 handler74 正式冊封入口寫入並重載場景 | **IDA writer／pack／production writer confirmed；E2**；同狀態 map 動畫與 NPC 畫面仍非 V3 |
| `0x0e` | `0x7830`；奧爾特加死亡／消失階段 | `ortegaDyingFlag` 在原版戰鬥後 set／clear，重載日夜場景 | **strong**；補同輸入 NPC frame／夜間 palette V3 對拍 |
| `0x10` | `0x56c8`；沙曼歐莎假王／拉之鏡 | `itemuse.go` 依勝敗 set／clear，`reloadTownDaynight` 消費 | **confirmed**；保留 `docs/103`／`docs/93` 證據 |
| `0x11` | `0x7563`；後期過場暫存可見性 | `story_flag_runtime_events.phoenix_revival` 在六珠動畫完成時 SET，重放 map cell `0x80`／場景後 CLEAR | **IDA chain／pack／production transaction confirmed；E2**；逐幀畫面仍非 V3 |
| `0x14` | `0x736a`；救援 boss／NPC present | `hostage_rescue` pack flags 由 guard/captive/boss/farewell 交易 set／clear | **confirmed**；補同狀態 movement／palette 對拍 |
| `0x16` | `0x716a`；諾魯德引導 passage | `guided_passage` animation pending／completed primitive 已 set／clear | **confirmed**；同狀態 NPC movement V3 仍待 |
| `0x17` | `0x156b`；母親／王座開場 | `motherEscort` set，謁見後 clear；正式 trace 已到達 | **confirmed**；逐格護送動畫仍是 V3 |
| `0x18` | `0x162a`；謁見後條件 NPC | 王座 handler set，NPC loader 依 flag 過濾 | **strong**；需原版同狀態畫面对拍 |
| `0x19` | `0x5622/0x5673/0x719c`；黑暗之燈後可見性 | `itemuse.go` 使用黑暗之燈後 set，日夜 palette／scene reload 已接 | **confirmed**；精確 palette register 仍待 V3 |
| `0x1a` | `0x760a`；巴拉摩斯戰後世界／NPC 可見性交易 | `story_flag_runtime_events.boss_aftermath` 在 CTY65 固定編隊勝利後與 `0x1b` 同時 SET，重設 day selector／clock／palette | **IDA chain／pack／production transaction confirmed；E2**；世界圖重建的 V3 畫面仍待 |
| `0x1b` | `0x7610`；同一巴拉摩斯戰後世界／NPC 可見性交易 | 同一 `boss_aftermath` transaction；不再以彩虹合成里程碑 `0x139` 代替 | **IDA chain／pack／production transaction confirmed；E2**；同狀態 NPC 畫面仍非 V3 |
| `0x1c` | `0x7232`；諾魯德完成後 NPC | `guided_passage.CompletedFlagRaw` set；場景 loader 過濾 | **confirmed**；只補 V3 動畫／聲音 |
| `0x1f` | `0x6d4a`；階段 boss／NPC gate | `staged_boss` 第一／第二戰交易 set／clear | **strong**；完整 formation／掉落／音效仍是戰鬥長尾 |
| `0x20` | `0x7053`；階段 boss 後 NPC | `staged_boss` first victory set、second required/clear | **strong**；補同狀態 NPC／戰鬥 cue 對拍 |
| `0x21` | `0x6c8e`；沙曼歐莎變身前後 | `itemuse.go` 鏡事件成功 clear | **confirmed**；原版敗戰回復與 reload 已有 tests |
| `0x22` | `0x5735`；變身後可見性 | `itemuse.go` 成功 set，切白天並重載 | **confirmed**；palette／dialogue timing 仍待 |
| `0x23` | `0x6ea3/0x6f1d`；商人聚落／建城者 | `settlement_founder` completion、world entrance、男女 visibility 已接 | **confirmed**；CTY60/61 長尾另列 `docs/74` |
| `0x24` | `0x6c88`；條件 NPC | `story_flag_runtime_events.conditional_story_flag` 經 CTY44 handler33 正式話す入口，DI `0x0bff` 換算 record71 後交易 `0x3d→0x24/0x21` | **IDA handler／pack／production command path confirmed；E2**；原版完整後續 route 與 V3 對拍仍保留限制 |
| `0x25` | `0x73f0`；巴哈拉達救援後 NPC | `hostage_rescue.CompletionSetFlagsRaw` 可由 pack 設定 | **strong**；正式 trace 已過主流程，仍需同狀態 NPC V3 |
| `0x26` | `0x5336`；諾亞尼爾清醒居民 | `quest_item_chain.use.SetFlagsRaw=[0x26]`，使用 `0x5a` 後 reload | **confirmed**；`docs/86` 已閉合 |
| `0x27` | `0x6688`；羅馬利亞臨時王位 | `temporary_role.ActiveRoleFlagRaw=0x27`，角色圖／場景由 flag 派生 | **confirmed**；`docs/82` 已閉合 |

補充：`0x15` 是 handler40 依性別設定的建城者 visibility flag（見 `docs/102`），但不在
`docs/71` 由原版 SET literal 掃出的 21 個條件 NPC flag 清單；它已由 settlement pack
明確接線，不能反向把它塞回本表來改變原始清單數量。

## 結論與工作閘門

- 日夜核心是 **E3／機制已存在**，目前不應重寫 clock、住宿 reset 或 NPC loader。
- 21 個 flag 的**原版 SET literal → `[0x4f70]` API** 已由 IDA ledger 閉合；
  `sub_131BD` 的日／夜 NPC filter 是共同 consumer，但不是每一個 flag 都必然掛在 NPC record。
  `0x0d`、`0x11`、`0x24` 及巴拉摩斯戰後 `0x1a/0x1b` 的 runtime pack／事件 consumer 仍須
  以 caller／raw event evidence 補齊，才能新增 game-pack 欄位或 runtime handler。
- `0x13`（勇氣試煉 completed）不在 `docs/71` 的 21 個 NPC visibility flags；其 writer
  仍是 `docs/100` 明確標示的 `unknown`，不得混入本表或用「接受試煉即 set」補洞。

## 2026-08-11 靜態 consumer 補證（追加，不覆寫上文）

依使用者指定以 IDA Pro 9.4 重新追查 `seg016:off_28984`（`sub_196D2` 以 table index+1
dispatch）與五個旗標的 writer。原始 `DQ3.EXE`、CTY 與文字輸入仍唯讀，完整 raw／hash／
IDA 位址基準集中於 [`docs/123`](123-static-battle-daynight-re.md) 與
[`docs/data/dq3-static-battle-daynight-evidence.json`](data/dq3-static-battle-daynight-evidence.json)。

- `0x0d`：handler74 `sub_16346` 先 GET `0x0e1/0x136`，未完成時進 `sub_1E713`；該函式
  CLEAR `0x0f5`、SET `0x0d`，接著重載場景並跑 `sub_1E7C9`／`sub_1E7E8`／`sub_1E7F3` 的
  map 動畫。`sub_131BD` 對 CTY byte5 做 full-byte flag test，CTY02 有 raw `0x0d` record。
- `0x11`：handler69 `sub_16139` 在 `0x8e` 與 `0x8f..0x94` 都成立後，寫入目前 map cell
  byte `0x80`，SET `0x11`，再依序呼叫 `sub_134AB`、`sub_131BD`、`sub_11971`、
  `sub_12C9C`，返回前 CLEAR `0x11`。CTY70 有一筆 full-byte `0x11` record；這是暫時性
  過場旗標，不可把現行 Go 的 clear-only 當成完整事件。
- `0x1a`／`0x1b`：handler70 `sub_1622A` 的固定編隊 `DS:0x4edf` 戰鬥勝利後，重置 party、
  寫日 selector／clock、重載 palette，再同時 SET 兩旗標並以 `sub_12DDF` 重建 DQ3BLK／
  DQ3UND 世界地圖；generic `sub_131BD` 的 CTY25／CTY00 records 分別證實兩個 full-byte
  consumer 資料存在。
- `0x24`：handler33 `sub_15904` 顯示 rec `0x0bff`，CLEAR `0x3d`、SET `0x24`／`0x21`，
  經 `sub_15010` 關閉／重畫；CTY43 有 full-byte `0x24` record，仍未找到可安全命名的
  production 直接 GET caller。

同格式掃描 `assets_raw/CTY*.DAT` 89 檔／768 筆 NPC，命中數為
`0x0d:26`、`0x11:1`、`0x1a:6`、`0x1b:25`、`0x24:4`。因此目前可把「原版 writer →
`[0x4f70]` → generic NPC consumer」標為 `confirmed`；仍不能把事件語意、Go pack writer、
直接 `sub_16F09` caller 或玩家可見結果標成完成，五項 runtime 欄維持 `unknown`／fail closed。

## 2026-08-11 攻略對讀與事件身份勘誤

本輪把本機攻略 `docs/66-original-flow-oracle.md`、`docs/75-phoenix-orbs-re.md`、
`docs/76-baramos-gaia-re.md`、`docs/77-r5b-castle-aftermath.md` 與原始 CTY 事件逐格對讀：

| flag | 攻略／CTY 路線定位 | IDA 靜態定位 | 目前 remake 影響 |
|---|---|---|---|
| `0x0d` | step 56e，索瑪後回到拉達多姆／國王；CTY80 sec1 handler74 | `sub_16346 → sub_1E713`，SET `0x0d`，重載場景及 map animation；CTY79／80 有後期居民 records | Go ending 使用自有 `cleared/lotoBlessed`，未寫原版 `0x0d`；終盤 NPC 可見性仍非 exact runtime |
| `0x11` | step 42，六珠祭壇／拉米亞復活；CTY70 handler69 | `sub_16139` 寫 map cell `0x80`，SET、重建 NPC／場景後返回前 CLEAR | Go 只有自訂不死鳥動畫與 clear；暫時場景的 map mutation／NPC visibility 尚未閉合 |
| `0x1a`／`0x1b` | step 44，CTY65 巴拉摩斯；step 45 返回阿里阿罕。龍之女王是 step 43 的 CTY67 handler52 | `sub_1622A` 固定編隊戰後同時 SET 兩旗標，重置 clock／palette／世界 map；CTY00／25 有對應 full-byte NPC records | Go 沒有兩旗標 production writer；巴拉摩斯戰後居民／世界狀態 parity 仍缺 |
| `0x24` | step 35，沙曼歐莎／拉之鏡前後；CTY43 有條件居民，CTY44 handler33 觸發相鄰交易 | `sub_15904` 顯示 rec `0x0bff`、CLEAR `0x3d`、SET `0x24`／`0x21` | Go 已有 `0x21/0x22/0x10/0x42` 鏡事件，但沒有 `0x24` writer；條件居民可見性仍缺 |

因此「五個旗標的原版 writer 與 generic consumer 已確認」不等於「五個旗標的 Go 玩家路徑
已完成」。本節取代先前把 `0x1a/0x1b` 稱為龍之女王材料中繼的現行解讀；舊說法保留於
`docs/data/special-events-audit.md` 的歷史勘誤段，並標示為已推翻。

## 2026-08-11 E2 runtime 接線勘誤（現行狀態）

上方「runtime unknown／missing」文字是本輪接線前的時間序列，不能覆蓋本節。依同一份
IDA 9.4 evidence，四條已能以固定 selector／caller 接到正式 engine path 的 transaction
已遷入 `events.json.story_flag_runtime_events`（schema `0.1.29`）：

- handler74：CTY80 國王冊封完成時寫 `0x0d`、清 `0xf5` 並重載場景。
- handler69：CTY70 六珠動畫完成時寫 transient `0x11`，把已證實的中央 map byte `0x80`
  以 save bit 保存，重載／存檔還原後再清 transient；拉米亞停放位置同樣只讀 pack。
- handler70：CTY65 巴拉摩斯固定編隊勝利後同時寫 `0x1a/0x1b`、清 `0x29`，重設日夜
  phase／step 並重套 palette。
- handler33：CTY44 `(13,27)` 正式「話す」入口先開原版 `DI=0x0bff`（`di-0xbb8=71`）
  對應 record，再依原始順序清 `0x3d`、寫 `0x24/0x21` 並重載條件 NPC。

`game/story_flags_test.go` 以 pack contract、正式 `InputState` command trace、Phoenix
save/load map replay、Baramos aftermath 及 ending writer 鎖定上述交易；Docker component
tests 已通過。這把五旗標的「原版 writer→pack→Go consumer」從 unknown 收斂為 E2，但不把
原版完整路線、逐幀動畫、palette register 或同狀態 DOSBox 畫面宣稱為 V3；仍未證實的後續
語意維持 evidence ledger／fail-closed。
