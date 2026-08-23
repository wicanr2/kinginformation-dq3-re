# 184 — 瑪依拉王者之劍特殊商店閉環

狀態：2026-08-23；D3／E2，正常玩家輸入與存讀檔已閉合。逐窗幾何與原版同狀態畫面仍非 V3。

## 原始輸入與工具

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/CTY81.DAT`：瑪依拉 section1 黑暗期 NPC 表含 `(4,3)`、sub2、handler raw72。
- IDA Pro 9.4，linear address；一次性 Docker、原始檔唯讀。可重生匯出為
  `tools/ida_dump_maira_special_shop.py`，未版控 sidecar 為
  `work/maira-special-shop-ida.json`。原始函式名與位址均保留，沒有改寫 IDA database。

## 已證實的入口到玩家效果

以下每列皆為 `confirmed`；位址是 IDA linear：

| 原始定位 | 原始證據 | 附加語意 |
|---|---|---|
| jump table `0x28a14` | bytes `15 63`，near offset `0x6315` | handler raw72 指向 `sub_16315` (`0x16315`)。 |
| `sub_16315` `0x16315..0x16333` | 寫 `DGROUP 0x0b62=1`、`0x0b34=1`，`BX=0` 呼叫 `sub_17034`，回來後清兩個暫態值 | 以 section facility index0 執行一次特殊商店；特殊模式只在此次呼叫期間有效。 |
| `sub_174D8` `0x174e5..0x17500` | 特殊模式下先 GET flag `0x134`；若未 set，再 GET flag `0x135`；後者 set 時 `INC CX` | 一般四項貨架之外，只有「`0x134` clear 且 `0x135` set」才顯示第五項。 |
| buy tail `0x176b8..0x176cc` | 特殊模式且選中 one-based `0x1d`，以 `sub_16EF4` clear flag `0x135` | 第五項是 raw `0x1c` 王者之劍；買下後成為一次性售罄。 |
| sell tail `0x17845..0x178a8` | 特殊模式且選中 one-based `0x6e`，售價固定 `0x57e4`；確認後清 item word、加錢，並以 `sub_16EF4` clear flag `0x134` | raw `0x6d` 歐里空金屬賣 22,500G；成功交易才解鎖王者之劍。 |

`sub_16EF4` 是 story flag CLEAR、`sub_16F09` 是 GET；其 mask／bit-array consumer 已在
`docs/123-static-battle-daynight-re.md` 閉合。兩個 flag 的原始靜態初值皆為 set，因而形成：

```text
CTY81 sec1 handler72
→ special facility0
→ 賣 raw0x6d／+22500G／clear 0x134
→ 再次開店時 raw0x1c 出現
→ 支付 ITEM.DAT 價格 35000G／取得道具／clear 0x135
→ 後續不再出現
```

這推翻舊 C/SDL prototype 的「持金屬並補淨額 12,500G 就直接換劍」作法。攻略只證實玩家
敘事順序，不能取代上述原版 writer／consumer。

## Remake 接線

- `events.json.special_shop_events[]` 保存 CTY／section／NPC、facility index、兩個 item raw ID
  與兩個 story flag raw ID；共用 Go 不含瑪依拉或 DQ3 專屬常數。
- `facilityForCty` 現以 `(cty, section, k)` 查找，避免 CTY81 section1 的 index0 被錯取成
  section0 旅店。
- 共用特殊商店 primitive 只做「成功賣出清 unlock flag」與「成功買入清 stock flag」；
  失敗、容量不足或金錢不足不改旗標。
- `storyBits` 已由既有冒險之書保存，因此解鎖與售罄狀態可跨 save/load。
- schema `0.1.48`／content `0.1.53`。

## 驗收與停止線

`TestMairaSpecialShopFormalInputSellUnlockBuyAndSaveRoundTrip` 從 CTY81 原始 NPC，以正式
`InputState` 開啟命令窗與交談，完成賣出、重進商店、購買及 save-state round-trip。
`TestMairaSpecialShopEventContract` 鎖定 pack selector 與 raw IDs；商店／facility targeted tests
一併通過。

本切片不宣稱商店所有文字、框線、動畫或聲音已達 V3，也不為攻略提及但沒有新玩家可見
discrepancy 的支線重新做全域 RE。王者之劍原版有限交易與正常玩家入口已足以完成 remake 功能 gate。
