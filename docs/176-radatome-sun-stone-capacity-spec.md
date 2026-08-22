# 176 — 拉達多姆太陽之石容量交易規格（2026-08-22）

## 玩家阻塞點

正式主線已通過龍女王光之珠並下降至阿列夫加特，從 CTY79 正常進入 CTY80 section2。
調查太陽之石寶箱後得到 `party item=false, present=true`，是原版共同寶箱 writer 在全隊
滿格時的失敗結果。

## 已讀取證據

- `docs/data/quest-items.md` 與 `docs/re-log-722-state-machine.md`：太陽之石 `0x72` 位於
  CTY80 section2，後續與雲雨之杖合成彩虹水滴。
- `events.json` 的 `dq3:event.radatome_sun_stone` 保存 CTY80 section2、type1、item
  `0x72`、present flag `0xc2`，consumer 為共同 pack treasure transaction。
- [`docs/31`](31-event-system.md)／[`docs/35`](35-script-format.md)：type1／3 事件依隊伍
  每人八格尋找 `0xff` 空位；滿格顯示行李滿且不完成取得。
- [`docs/161`](161-common-party-item-grant-writer-spec.md)：IDA Pro 9.4 已證實場景／寶箱
  caller `sub_18C0F` 落入共同 writer。
- 本輪 fresh IDA Pro 9.4＋IDAPython sidecar：
  `/tmp/dq3-yellow-orb-ida-20260822/party-item-grant-callers.json`；輸入
  `assets_raw/DQ3.EXE` 115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；位址為
  IDA linear，原始名稱與 bytes 均保留。

推論等級：CTY selector、item／flag 與共同 writer 為 **confirmed**；本輪滿格失敗為
**confirmed runtime**。太陽之石原版逐框畫面與聲音不由本規格提升為 V3。

## Production trace 規格

1. 從光之珠合法路徑自然下降，經正式飛行、城門與 transition 抵達 CTY80 section2。
2. 調查前以正式道具選單預留一格：先丟藥草，其次只丟 ITEM.DAT 證明為裝備且未穿戴
   的備品。
3. 不丟劇情物品、不擴大背包、不直接寫 inventory／flag。
4. 以正常調查命令觸發 type1 event；成功條件是全隊持有 `0x72` 且 flag `0xc2` clear。
   不得只查勇者，因共同 writer 可合法寫入同伴。
5. 太陽之石持有者位置保持原樣，後續彩虹合成 consumer 必須從全隊找到並消耗它。

## 驗收界線

- 完整 boot production trace 必須越過 CTY80 並走到下一個實際玩家 blocker。
- 每名角色仍受 pack 八格容量；安全丟棄集合不足時失敗即關閉。
- 本切片不修改 production engine、寶箱資料、合成規則或終盤戰鬥。
