# 變化之杖交換的隊伍持有者驗收規格（2026-08-22）

## 問題

怪力魔獎勵已依原版共用 writer 合法寫入第一個有空位的同伴。格陵蘭交換 production 已以
`hasPartyItem` gate 並用 `replacePartyItem` 在實際 owner 原格把變化之杖換成船員骨頭；但
完整 campaign 的拒絕、成功與 save/load 斷言仍使用 hero-only `hasItem`，因此把合法的同伴
持有誤報成交易失敗。

## 證據與推論等級

- `docs/158-field-item-owner-selection-drop-spec.md`：IDA Pro 9.4 已證實 `DS:062D` owner 與
  rec421 個人物品操作，推論等級 **confirmed**。
- `docs/161-common-party-item-grant-writer-spec.md`：`sub_16856` 依隊長至同伴、每人八格寫入，
  推論等級 **confirmed**。
- `docs/162-party-item-gate-and-consume-spec.md`：已閉合 required-item 的全隊 gate 與
  `replacePartyItem` 單一 owner 原格替換。
- `docs/168-samanosa-mirror-reward-capacity-transaction.md`：monster89 的 `0x62` 獎勵也使用
  相同 party writer；不可再假設落在勇者。

本切片沒有新增原版語意，因此不重跑 IDA；只讓測試與 production 已證實的 ownership 契約
一致。原始輸入雜湊與 IDA 位址契約沿用上述文件。

## 規格與驗收

1. 拒絕交換後，以 `hasPartyItem(required)` 證明實際 owner 的杖仍存在，世界旗標不變。
2. 接受交換後，以 `hasPartyItem` 證明 required 消失、granted 出現在同一角色原格。
3. save/load 後仍以全隊持有驗證，不把道具搬回勇者。
4. 新 component test 專門把變化之杖放在同伴，鎖定 offer gate、成功對話、原格替換與勇者
   背包不變。
5. 完整 production trace 必須由 monster89 的實際獎勵 owner 直接完成交換，再繼續到幽靈船。
