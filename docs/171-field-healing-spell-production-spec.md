# 野外補血咒文正式接線規格（2026-08-22）

## 目的與停止條件

目前正式主線已通過幽靈船與奧莉維亞海岬，但 remake 的野外咒文清單只接了
移動／場景效果，沒有原版的 HP 回復咒文。這使玩家只能依賴藥草，屬於正式玩家
路徑缺口。本切片只閉合「選施法者 → 選咒文 → 必要時選目標 → 扣 MP → 回復 HP」；
不以更多藥草、直接寫 HP 或測試專用入口替代。

完成條件：161～165 的原始 descriptor、施法者／目標選擇、MP 交易、單體／全體
分支、回復亂數、死亡 gate 與 HP 上限全部有同一輸入 EXE 的 writer-consumer 證據，
並由正式 `InputState` 測試覆蓋。逐 frame 動畫、PCM 實際播放時長與視窗像素位置若
沒有同狀態原版畫格，維持 V1／unknown，不阻塞 E2 runtime。

## 輸入與位址契約

- 輸入：`assets_raw/DQ3.EXE`
- 大小：115,282 bytes
- SHA-256：`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- 主工具：IDA Pro 9.4.0.260610，16-bit DOS loader
- 位址基準：本文未特別標示者皆為 IDA linear；`file = linear - 0xEC90`，
  `logical = linear - 0x10000`。
- 可重建匯出：`tools/ida_dump_field_heal.py`
- 本機 sidecar（gitignored）：`work/field-heal-ida-20260822.json`
- 註記方式：原始函式名、位址、bytes 與 operand 全部保留；下列語意是附加註解，
  不回寫或覆蓋原始 EXE。

## 證據 ledger

| 等級 | 原始定位 | 原始行為 | 可採用語意 |
|---|---|---|---|
| confirmed | `0x1C9C1..0x1C9DF` | `sub_1885F` 完成隊員選擇後才呼叫 `sub_1C9EE`；選擇結果在 `DS:0722` | 野外咒文先選施法者，不可由引擎替玩家自動挑第一位會施法者 |
| confirmed | `0x1C9EE..0x1CB33` | 以 `DS:0722` 解出角色 record，從角色的三段已習得咒文表建立清單，再把選中 spell record 交給 `sub_1CB3C` | 咒文清單屬於所選角色，不是全隊咒文聯集 |
| confirmed | descriptor file `0x1997B..0x19989`／IDA `0x2860B..0x28619` | 161=`03 1e 0b`、162=`05 55 0b`、163=`07 ff 0b`、164=`12 32 13`、165=`3e ff 13` | 三 byte 依序為 MP、base、flags；`0x0B` 是我方單體，`0x13` 是我方全體 |
| confirmed | `0x1CB61..0x1CB84` | `spellID=record-0x79`，查 `DS:37C3 + spellID*3`；MP 不足在任何目標或效果前返回 | MP cost 必須取 pack descriptor，MP 不足不改狀態 |
| confirmed | `0x1CB8C..0x1CBAC` | flags 同時具 bit `0x02`／`0x08` 才呼叫 `sub_188A9→sub_1885F`；取消旗標 `DS:0726==1` 直接離開 | 單體回復先選目標；取消不扣 MP，並結束這次咒文 modal |
| confirmed | `0x1CBAD..0x1CBF8` | 顯示施法序列後於 `0x1CBBC..0x1CBC0` 扣 `[caster+0x18]`，再由 `DS:38CC` 表派發 | 一旦確認合法目標，MP 先扣，再執行效果；效果無效不退款 |
| confirmed | `DS:38CC` entries 0..4／IDA `0x2869C..0x286A5` | 161～165 的五個 pointer word 都是 `0xCC4F` | 五個回復咒文共用 `sub_1CC4F`，不可假設每個 spell 有獨立 handler |
| confirmed | `0x1CC52..0x1CCA7` | 取 descriptor base；flags `0x08` 以 `DS:0741` 單次呼叫 `sub_146EB`，flags `0x10` 從隊員 1 起依 `DS:5077` 逐員呼叫 | 161～163 單體；164～165 全隊，順序為 active party order |
| confirmed | `0x14685..0x1469E` | 只檢查目標 `[member+0x38] & 0x80`；死亡時顯示失敗並返回 | 回復不能復活；死亡目標仍發生在已扣 MP 的 effect handler 內 |
| confirmed | `0x146FE..0x14729` | base `0xFF` 直接把 `[si+0x16]` 設為 `[si+0x2A]`；其他 base 令 `BX=base+10`，反覆 RNG 直到 `DX>=base`，把 DX 加到 HP 後以 max HP 截頂 | 163／165 補滿；161／162／164 均勻回復 `base..base+9`，最後 clamp max HP |
| confirmed | `0x1E6C9..0x1E6E6` | 以 `DIV BX` 的餘數回傳 `DX` | 上述 rejection loop 精確產生 `base..base+9`，不是戰鬥 `base/2 + rng(base/2)` |
| confirmed | `0x19834..0x19841` | `DS:259C` 以 1-based index 查 `DS:4F15` pointer table，回傳角色 record 至 `SI` | HP／MP writer 的 actor index 是 active party 1-based 索引 |

## Production 資料契約

版本專屬 record、MP、base、flags、範圍與公式不可新增成 Go 常數。game pack 必須
提供一份嚴格驗證的野外咒文定義，至少包含：

- `record_raw`
- 穩定的 `effect_id`
- `mp_cost`
- `base_amount`
- `target_scope`：`party_member`、`party_all` 或既有場景範圍
- `amount_formula`：本切片只接受 `base_to_base_plus_9` 或 `full_hp`
- 原始 descriptor bytes 與 `Evidence`

161～165 的 production 值必須逐 byte 等於上表。缺欄位、重複 record、未知
`effect_id`／scope／formula、MP 與 descriptor 不一致，一律失敗即關閉，不設 Go fallback。

## Runtime 狀態機

1. 玩家由正式命令窗選「咒文」，進入 active-party 施法者頁。
2. 選定角色後，只列該角色已習得且 game pack 宣告可在野外使用的咒文。
3. 施法者頁、咒文頁或單體目標頁取消都結束本次 command invocation；目標取消發生在
   MP writer 前，因此不扣 MP。
4. 選單確認時先檢查 caster 存活與 MP；不足則不扣 MP、不進 effect。
5. `party_member` 先進目標頁；目標取消直接結束且不扣 MP。
6. 確認目標後扣 MP，再執行 `heal_hp`。死亡或滿 HP 都不退款；滿 HP 只被 clamp。
7. `party_all` 不進目標頁，扣 MP 後依 active party 順序處理；死亡成員略過回復，
   其他成員各自取得一個獨立亂數結果。
8. 完成後關閉 modal；HP／MP 已由既有 save state 保存，不新增存檔欄位。

## 元件與正式路徑驗收

- game-pack：161～165 raw bytes parity；缺少／重複／不合法公式 fail closed；canonical hash 改變。
- 單體：同伴施法者→同伴目標；回復值落在 base..base+9 且 clamp；MP 精確扣除。
- 單體取消：目標頁取消不扣 MP。
- 死亡／滿 HP：死亡不回復但扣 MP；滿 HP 維持上限且扣 MP。
- 全體：164 各活人獨立回復 50..59；165 活人補滿；死人不復活；只扣一次 MP。
- 施法者：正式頁可選主角與同伴；清單不混入其他角色學會的咒文。
- 存檔：施法後 HP／MP encode/decode round-trip。
- campaign：奧莉維亞海岬後只經正式 `InputState` 使用補血咒文，繼續到下一個真實 blocker。

## 尚未宣稱

- `sub_15002`／`sub_21414` 的逐 frame 視覺與 PCM wall-clock 尚未做原版同狀態對拍，
  因此本切片只宣稱 E2 runtime，不宣稱 V3／音訊 timing exact。
- 本文件沒有證明所有 60 個咒文、所有狀態解除／復活咒文已接線；未列入 pack 的
  descriptor 仍須 fail closed，不能由本切片外推為完成。
