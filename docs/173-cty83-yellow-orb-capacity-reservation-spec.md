# 173 — CTY83 黃寶珠容量預留規格（2026-08-22）

## 玩家阻塞點

場外回復咒文改走原版施法流程後，正式主線不再固定消耗藥草。從 CTY64 銀寶珠
checkpoint 返航 CTY83 時，途中戰鬥掉落會把剛釋出的個人物品格再次填滿。建城者
record42／43 已正常顯示；第一次重播在未預留容量時得到
`party item=false, present=true`，符合共同 writer 的滿格失敗。

這個結果精確符合原版共同物品 writer 的滿格失敗交易，不能把它解釋為黃寶珠事件
未接線，也不能藉由增加容量或直接寫入背包繞過。

## 已讀取的直接證據

- [`docs/102`](102-merchant-settlement-world-entrance-re.md)：CTY83 section0、tile `(4,2)`
  subid2 的原始 event bytes 是 `01 6a 00 4a`。handler48 只顯示 record42，並在
  present flag `0x4a` set 時追加 record43；真正取得道具的是共同 examine consumer。
- [`docs/161`](161-common-party-item-grant-writer-spec.md)：共同 writer 依勇者至同伴順序，
  每人掃八個 word；全隊滿格時不寫物品。成功後，caller 才可推進不可逆交易。
- [`docs/170`](170-cty38-spare-equipment-supply-capacity-spec.md)：正式測試只能優先丟棄
  藥草；仍不足時，只可丟棄 ITEM.DAT `EquipSlot>=0` 且目前未穿戴的備品。

## 本輪 IDA Pro 9.4 複核

- 輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4＋IDAPython；一次性 Docker image
  `ida-pro-9.4-ver3:py312-x11-v4`。
- exporter：`tools/ida_dump_party_item_grant_callers.py`；sidecar：
  `/tmp/dq3-yellow-orb-ida-20260822/party-item-grant-callers.json`。
- 位址空間：IDA linear；`logical=linear-0x10000`、`file=logical+0x1370`。
- 匯出方式：保留原始函式名、位址、call bytes 與附近 raw window；沒有 rename、patch
  或修改原始 EXE／database。

複核結果為 **confirmed**：`sub_18C0F` 是寶箱／場景取得物 caller，與劇情獎勵共同
落入 `sub_1684E`／`sub_16856`；writer 仍是逐角色八格、第一個 `0x00ff` 成功寫入，
全隊滿格則設失敗狀態而不寫。黃寶珠的 CTY event 與 handler48 交易已由 `docs/102`
閉合；本輪不把 raw caller 自動升格為其他事件語意。

## Production trace 規格

1. 從銀寶珠合法 save/load checkpoint 以正式魯拉、登船與航行抵達 CTY83。
2. 以正式談話完成 handler48 的 record42／43，不合併成給物事件。
3. 調查前要求全隊至少一個合法空格：先從持有者以正式道具選單丟藥草；若仍不足，
   再丟 ITEM.DAT 證明為裝備、且不在任何 equipped slot 的備品。
4. 不丟劇情物品、不改 pack 容量、不直接修改 inventory／story flag。
5. 由正常調查命令觸發 CTY83 type1 event；成功後才可斷言取得 `0x6a` 且 clear `0x4a`。
6. 立即 save/load，並由同一路徑繼續前往六寶珠祭壇。

所有取得與 save/load 斷言必須使用全隊 consumer，不得只查勇者。加入容量預留後的
第二次重播得到 `hero item=false, present=false`；這不是遺失，而是共同 writer 已合法把
黃寶珠寫入同伴。測試已改查 `hasPartyItem`，與 IDA 證實的隊伍順序一致。

## 驗收界線

- E3：完整 boot production trace 通過黃寶珠交易、save/load 並抵達下一主線節點。
- 這次容量修正不提升畫面等級；原版同狀態 V2／V3 仍依 `docs/102` 標示為未閉合。
- 若正式丟棄安全集合後仍無空格，測試必須失敗即關閉，另立新的玩家阻塞規格。
