# 教會解毒、毒氣與場景毒傷的 IDA 證據閉環（2026-08-22）

## 範圍與證據契約

本文件只回答角色中毒、怪物毒氣、場景毒傷與教會解毒的原版資料流，並訂正
「解詛咒也是角色 `+0x38` flag」的舊說法。它不是所有角色狀態的完整命名表。

- 輸入：`assets_raw/DQ3.EXE`
- 大小：115282 bytes
- SHA-256：`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- 主要工具：IDA Pro 9.4，16-bit DOS database 與交叉參照
- 位址：本文 `IDA` 是 IDA linear；`logical` 是第一個程式 segment 的邏輯位址；
  `file` 是原始 EXE offset。三者不可混用。
- 匯出：`tools/ida_dump_church_status.py`；保留原始 `sub_*` 名稱、位址、bytes、xref
  與推論等級，不 rename、不寫回原始 binary。

## 已證實：怪物 bit38 會分派到毒氣 action 0x32

`sub_1A973` 的怪物行動 dispatcher 先以 `D3MNS +0x0d` 決定是否進入
`sub_199DC`。`sub_19AD6` 對 `D3MNS +0x0e..+0x13` 六個 byte 做 popcount，均勻抽中
一個 set bit，將 bit index 留在 `DH`。

`sub_199DC` 並未把所有 bit index 直接當作玩家咒文 ID：

```text
bit 0..15   → action=bit
bit 16..17  → action=bit+2
bit 18..47  → action=DGROUP 0x3930[bit-18]
```

DGROUP `0x3930` 的 30-byte 原始表為：

```text
14 17 18 19 1a 1b 1c 1d 1f 21 22 23 24 25 28 29
2a 2b 30 31 32 33 34 35 36 37 38 39 3a 3b
```

因此 bit38（`18 + 20`）確定 remap 為 action `0x32`。非特殊戰鬥模式下，
`sub_199DC` 以 `action-0x14` 索引 DGROUP `0x394e` 的 word pointer table；第 30 筆是
logical `0xa307`，也就是下節毒氣 writer。推論等級：**confirmed**（mask selector、
原始 remap bytes、間接 table consumer 與 handler 入口已閉合）。

這推翻 docs/37 舊有「48-bit mask 全部直接對映 `dq3_spelldef`」的敘述。接線前的 Go
確曾讓 bit38 落回物理攻擊；本切片落地時的 schema `0.1.35` 已用 pack
`monster_actions` 接手 bit38。
其他未閉合 bit 仍不可因 `MonsterSpellRec` 的歷史名稱表而冒稱已知咒文語意。

## 已證實：毒氣 writer

IDA `0x1a307`（logical `0xa307`）先顯示 D3TXT00 rec375（「{角色}放出毒氣。」），
再逐一處理隊員：

```text
0x1a327  test word ptr [di+38h], 0c0h  ; 已中毒或死亡則跳過
0x1a32e  call sub_1E6B9                ; rng(256)
0x1a331  cmp al, 64h
0x1a333  ja  skip                      ; roll <= 100 才成功
0x1a335  or byte ptr [di+38h], 40h     ; 寫入持久中毒位元
0x1a339  mov di, 173h                  ; D3TXT00 rec371「{角色}中毒了。」
```

推論等級：**confirmed**（正式 action table → handler → 狀態 writer → 玩家可見訊息）。

## 已證實：正常移動每步 1 HP 毒傷

正式場景 movement caller `sub_193E3` 呼叫 `sub_19530`，後者在每次合法移動中逐員呼叫
`sub_1964E`。該 consumer：

```text
0x19654  mov ax,[si+38h]
0x19659  test ax,80h                    ; 死亡不處理
0x19685  test ax,40h                    ; 中毒？
0x1968a  test word_272EF,3              ; 船／飛行等移動狀態時跳過
0x19692  add byte ptr [bx+662h],1       ; 本步傷害累積 +1
...
0x196af..0x196cc                        ; 扣 HP；歸零時寫死亡 bit0x80
```

同一個每步 damage byte 也承接地形傷害；它在 `sub_19530` 開頭逐步清零，因此毒傷不是
累積到若干步才一次扣除，而是每個符合條件的移動步各 1 HP。`sub_1EF95` 是受傷 palette
flash helper，不是毒傷公式。推論等級：**confirmed**（正式 movement caller、status
consumer、damage accumulator 與死亡 writer 已閉合）。

## 已證實：教會解毒

教會 dispatcher `sub_17068` 的第一項進入 `sub_1712B`（file `0x849b`）：

```text
0x1715f  mov ax,[di+38h]
0x17162  test ax,40h                    ; 沒中毒則 rec0x12a
0x17176  mov word ptr ds:2593h,5        ; 固定費用 5G
...                                    ; 確認與付款；失敗不清狀態
0x171a3  and word ptr [di+38h],0bfh     ; 成功清 bit0x40
0x171a8  mov bp,21h                     ; 音效 cue
0x171b0  mov di,12fh                    ; 成功訊息
```

推論等級：**confirmed**（入口、狀態 reader、費用、付款 gate、writer 與訊息已閉合）。

## 解詛咒訂正與停止線

`sub_171CC`（file `0x853c`）不是讀角色 `+0x38`：它掃描 `+0x3a` 起八個 item word 的
`bit0x4000`，沒有命中時顯示 rec0x12b；費用是角色 level `+0x15 × 100`；付款後把每個
命中 word 寫成 `0x00ff`，再呼叫裝備重算。這足以推翻「poison/curse 共用 +0x38 flags」。

上述是本文件當時的停止線。後續 `docs/147` 已由 `sub_17ED9` 閉合來源：ITEM
`+4/+5` little-endian word 與 `0x0e00` 相交時，裝備 writer OR item word `bit0x4000`；
五筆原始 ITEM record、同部位換裝拒絕及教會 consumer 均為 **confirmed**，不再是 unknown。

## Remake 接線現況與剩餘 gate

當時 schema `0.1.35`／content `0.1.40` 已完成下列 E2 垂直鏈；後續 `0.1.37`／`0.1.42`
又加入驅毒草與解詛咒交易，分別見 `docs/138`、`docs/147`。`docs/149` 切片當時再升為
schema `0.1.38`／content `0.1.43`；工作樹現行版本以根 README 與 manifest 為準，後續
音效欄位不改變本段 poison transaction：

- game-pack `monster_actions` 將 bit38 指向 poison condition，逐名處理存活且未中毒成員；
- 戰鬥結束回寫 `Game`／`Member`，save JSON round-trip 保留 condition；
- 正常移動按 pack 每步 1 HP，船／飛行只抑制毒傷；
- 教會按 pack 固定 5G 清除指定成員的 poison；
- 原始 EXE bytes／文字 glyph parity、schema/reference validation 與 component tests 通過。

這證明該有限鏈為 E2；後續 `TestOpeningProductionInputTrace` 已以正式驅毒草
選人交易越過 CTY23 單人回程；後續正式補給路線也已跨過 monster89、幽靈船、愛的回憶、
CTY36、蓋亞之劍與銀／黃寶珠；`docs/144..146` 的隊伍／補給策略訂正後，最新完整重播已
由標題抵達 `THE END`。CTY23 證據見 `docs/139`，補給及怪物 action 訂正見
`docs/140`～`docs/146`；不得由測試直接清 poison。

解詛咒的現行證據與 E2 接線見 `docs/147`。所有玩家可見文字與服務設定仍須留在
game-pack JSON；不能把 DQ3 record、費用或句子新增成 production Go 常數。主線 E3 不代表
教會逐畫面／cue timing 已達 V3。
