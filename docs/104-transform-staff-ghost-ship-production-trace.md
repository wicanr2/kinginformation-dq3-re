# 變身杖交換／幽靈船世界狀態 production trace

更新：2026-08-02

## 結論

CTY54 handler44 不是舊 `scriptedTable` 所描述的「持有變身杖就立即換成船員骨頭」。原版有
兩組獨立的 Yes／No 對話，成功交易另包含世界位置、世界狀態與劇情旗標副作用。現行
game pack `dq3:event.greenland_transform_staff_exchange` 已取代舊硬編碼列；從新遊戲開始的
正式 `InputState` trace 已由沙曼歐莎假王勝利，正常住宿、魯拉、登船與航行抵達 CTY54，
走過拒絕及接受分支，並完成交換後 save/load。

本切片完成到「船員骨頭與幽靈船世界狀態成立」。幽靈船可見物件、登船入口、愛的回憶及
其後蓋亞之劍仍是下一個垂直切片，不因 world-state bit 已寫入就宣稱完成。

## 證據身分

- 輸入：`assets_raw/DQ3.EXE`
- 大小：115282 bytes
- SHA-256：`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- 主要工具：IDA Pro 9.4；資料庫位址空間為 IDA linear。
- 本載入映像換算：`file = IDA linear - 0xec90`。
- 原始資料：`CTY54.DAT`、`D3TXT06.TXT` records 56–62。

IDA 工作資料庫與完整匯出 sidecar 只保存在 `/tmp/dq3-cty54-ida/`，不進 Git。所有下列
位址仍保留原始定位；語意是附加註記。

## handler44 閉環

| 原始定位 | 原始行為 | 附加語意 | 等級 |
|---|---|---|---|
| linear `0x15c23`／file `0x6f93` | `BX=0x43; call sub_16F09` | 讀交換可用旗標；clear 時直接顯示 record61 | confirmed |
| linear `0x15c38`／file `0x6fa8` | 寫搜尋值 `0x62`，呼叫 inventory search | 尋找變身杖，保留命中 word 的 `SI` | confirmed |
| linear `0x15c48..0x15c65` | records56、57／58 與一次 choice call | 無杖時先問是否知道變身杖；兩支都不交易 | confirmed |
| linear `0x15c6e..0x15c89` | record59、choice、拒絕 record62 | 有杖時另開交換選擇；拒絕不消耗 | confirmed |
| linear `0x15c8b`／file `0x6ffb` | `mov word ptr [si],0x63` | 在原背包格把 `0x62` 換成船員骨頭 `0x63` | confirmed |
| linear `0x15c8f..0x15ca0` | `[0x5053]=150; [0x5055]=90; [0x4f44] |= 4` | 設成功後世界位置與 world-state bit `0x04` | confirmed |
| linear `0x15ca0` | `call sub_12C9C` | 依 world-state 重建世界物件；bit4 consumer 續呼叫 `sub_12D94` | confirmed |
| linear `0x15ca3..0x15ca9` | `BX=0x43; call sub_16EF4` | 清除交換可用旗標 | confirmed |
| linear `0x15ca9..0x15cb4` | 依序顯示 records60、61 | 成功後兩段原版對白 | confirmed |

`sub_12D94` 在 linear `0x12da0..0x12ddb` 讀取 `[0x5053/0x5055]`，依方向選
tile `0x33/0x35/0x37/0x39` 寫入地表；碰撞時會從原始候選座標表找替代位置並回寫。這證明
bit `0x04` 與 `(150,90)` 是玩家可見世界物件資料，不是可忽略的內部 bookkeeping；但
物件的完整移動／進入 consumer 尚未在本文件閉合。

## game-pack 與驗收

- schema `0.1.19` 新增有限 `choice_item_exchange_events`；selector、旗標、兩個 item ID、
  成功世界位置、world-state mask、七段文字及 Yes／No label 全在 JSON。
- Go 只保留具名有限狀態機；任何文字或命中道具在 choice 開啟期間消失都會在持久狀態
  mutation 前失敗即關閉。
- parity test 逐 byte 鎖定 CTY54 NPC、EXE 關鍵 writer，逐 word 鎖定 D3TXT06 records56–62。
- component tests 覆蓋無杖 Yes／No、持杖拒絕、成功原格換物、旗標／世界狀態／位置及後話。
- `TestOpeningProductionInputTrace` 覆蓋從新遊戲至此的正式玩家路徑與成功後 save/load。
- runtime V1 核對圖：[`交換提議`](img/greenland_transform_staff_offer.png)、
  [`Yes／No`](img/greenland_transform_staff_choice.png)。兩圖使用原始 CTY54 scene 與 pack
  文字；二選一視窗 geometry 仍是已知 V3 GAP，不以本圖宣稱原版同狀態 parity。

## 下一個 blocker

第一個未閉合的玩家 blocker 是 bit `0x04` 世界物件從 `(150,90)` 開始的可見 tile、移動規則、
碰撞／入口與幽靈船 CTY，接著才是愛的回憶取得和使用。下一輪必須沿
`writer → world object update → entry consumer → CTY／treasure → 玩家可見結果` 追完，
不可只因船員骨頭已存在就直接跳到後段事件。
