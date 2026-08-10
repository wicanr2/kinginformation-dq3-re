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
| `0x0d` | `0xfab0`；後期條件 NPC | 未找到 pack 欄位、production writer 或既有事件 trace | **unknown**；先以 IDA caller／consumer 追出事件，不得猜設旗時機 |
| `0x0e` | `0x7830`；奧爾特加死亡／消失階段 | `ortegaDyingFlag` 在原版戰鬥後 set／clear，重載日夜場景 | **strong**；補同輸入 NPC frame／夜間 palette V3 對拍 |
| `0x10` | `0x56c8`；沙曼歐莎假王／拉之鏡 | `itemuse.go` 依勝敗 set／clear，`reloadTownDaynight` 消費 | **confirmed**；保留 `docs/103`／`docs/93` 證據 |
| `0x11` | `0x7563`；後期過場暫存可見性 | `phoenix.go` 只保留原版 return 前 clear；未找到 production set | **unknown（writer）**；不可把 clear-only 當完整事件 |
| `0x14` | `0x736a`；救援 boss／NPC present | `hostage_rescue` pack flags 由 guard/captive/boss/farewell 交易 set／clear | **confirmed**；補同狀態 movement／palette 對拍 |
| `0x16` | `0x716a`；諾魯德引導 passage | `guided_passage` animation pending／completed primitive 已 set／clear | **confirmed**；同狀態 NPC movement V3 仍待 |
| `0x17` | `0x156b`；母親／王座開場 | `motherEscort` set，謁見後 clear；正式 trace 已到達 | **confirmed**；逐格護送動畫仍是 V3 |
| `0x18` | `0x162a`；謁見後條件 NPC | 王座 handler set，NPC loader 依 flag 過濾 | **strong**；需原版同狀態畫面对拍 |
| `0x19` | `0x5622/0x5673/0x719c`；黑暗之燈後可見性 | `itemuse.go` 使用黑暗之燈後 set，日夜 palette／scene reload 已接 | **confirmed**；精確 palette register 仍待 V3 |
| `0x1a` | `0x760a`；龍之女王材料／中繼事件 | `docs/data/special-events-audit.md` 已確認 handler70 會用到；現行 pack／Go 沒有該事件 writer | **confirmed 原版／missing runtime**；先補 handler70 的 JSON evidence，再決定是否納入主線 |
| `0x1b` | `0x7610`；同一龍之女王材料／中繼事件 | 同 `0x1a`，尚無 production writer／consumer slice | **confirmed 原版／missing runtime**；不可用彩虹合成里程碑 `0x139` 代替 |
| `0x1c` | `0x7232`；諾魯德完成後 NPC | `guided_passage.CompletedFlagRaw` set；場景 loader 過濾 | **confirmed**；只補 V3 動畫／聲音 |
| `0x1f` | `0x6d4a`；階段 boss／NPC gate | `staged_boss` 第一／第二戰交易 set／clear | **strong**；完整 formation／掉落／音效仍是戰鬥長尾 |
| `0x20` | `0x7053`；階段 boss 後 NPC | `staged_boss` first victory set、second required/clear | **strong**；補同狀態 NPC／戰鬥 cue 對拍 |
| `0x21` | `0x6c8e`；沙曼歐莎變身前後 | `itemuse.go` 鏡事件成功 clear | **confirmed**；原版敗戰回復與 reload 已有 tests |
| `0x22` | `0x5735`；變身後可見性 | `itemuse.go` 成功 set，切白天並重載 | **confirmed**；palette／dialogue timing 仍待 |
| `0x23` | `0x6ea3/0x6f1d`；商人聚落／建城者 | `settlement_founder` completion、world entrance、男女 visibility 已接 | **confirmed**；CTY60/61 長尾另列 `docs/74` |
| `0x24` | `0x6c88`；條件 NPC | 未找到現行 pack 欄位或 production writer | **unknown**；依原版 caller→consumer 追查，不創造 Go 常數 |
| `0x25` | `0x73f0`；巴哈拉達救援後 NPC | `hostage_rescue.CompletionSetFlagsRaw` 可由 pack 設定 | **strong**；正式 trace 已過主流程，仍需同狀態 NPC V3 |
| `0x26` | `0x5336`；諾亞尼爾清醒居民 | `quest_item_chain.use.SetFlagsRaw=[0x26]`，使用 `0x5a` 後 reload | **confirmed**；`docs/86` 已閉合 |
| `0x27` | `0x6688`；羅馬利亞臨時王位 | `temporary_role.ActiveRoleFlagRaw=0x27`，角色圖／場景由 flag 派生 | **confirmed**；`docs/82` 已閉合 |

補充：`0x15` 是 handler40 依性別設定的建城者 visibility flag（見 `docs/102`），但不在
`docs/71` 由原版 SET literal 掃出的 21 個條件 NPC flag 清單；它已由 settlement pack
明確接線，不能反向把它塞回本表來改變原始清單數量。

## 結論與工作閘門

- 日夜核心是 **E3／機制已存在**，目前不應重寫 clock、住宿 reset 或 NPC loader。
- 21 個 flag 中，已有正式 writer／consumer 的事件可進入 V3 畫面／聲音對拍；`0x0d`、
  `0x11` writer、`0x24` 及龍之女王 `0x1a/0x1b` 必須先有 IDA caller／raw event evidence，
  才能新增 game-pack 欄位或 runtime handler。
- `0x13`（勇氣試煉 completed）不在 `docs/71` 的 21 個 NPC visibility flags；其 writer
  仍是 `docs/100` 明確標示的 `unknown`，不得混入本表或用「接受試煉即 set」補洞。
