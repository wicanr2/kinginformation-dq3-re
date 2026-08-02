# 75 — 六寶珠／不死鳥拉米亞重新反組譯與 Ebiten 還原（2026-07-28）

## 結論

舊 C remake 與舊 RE log 的 `CTY82 sec17`、六珠齊自動復活、`y` 鍵起降及「精訊版
build 未完成」均不正確。本次以三份互相獨立的證據閉環：

1. 本機原版完整影片
   `dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4`
   約 41:40–42:30：逐座放珠、和兩名守護神交談、蛋動畫、出祠堂後走上鳥並飛行。
2. `CTY70.DAT`：17×20、spawn `(8,19)`、兩守護神 `(7,10)/(9,10)` handler69，
   中央條件 sprite `(8,8)`，六個 type0 祭壇 event。
3. `DQ3.EXE` file `0x7425..0x7599`：祭壇、對話、六幀動畫、world vehicle transaction。

## 六座祭壇

| handler | event flag | event 座標 |
|---:|---:|---:|
| 63 | `0x8f` | `(8,11)` |
| 64 | `0x90` | `(4,9)` |
| 65 | `0x91` | `(4,5)` |
| 66 | `0x92` | `(8,3)` |
| 67 | `0x93` | `(12,5)` |
| 68 | `0x94` | `(12,9)` |

`0x7446` 先檢查該祭壇旗標；若仍設，掃背包第一顆 `0x66..0x6b`，移除該道具，
清除 `0x12c+(item-0x66)` 與祭壇 `0x8f..0x94`，顯示 D3TXT04 rec96。
因此舊文件用「搜尋 SET 0x91..0x94」判斷 build 不完整，是把旗標極性看反。

祭壇未放珠是空盆 tile `0x94`；放珠後覆成紅珠 tile `0x95`。中央復活動畫才使用
`0x7b..0x80`，不能把 `0x7b` 當成六座祭壇覆圖。

## 守護神與復活

handler69（file `0x74a9`）：

- `flag 0x8e` 已清：rec94「傳說中的不死鳥拉米亞復活了」。
- 尚未放珠：rec92 後結束。
- 部分放珠：rec92 後，依 `0x12c..0x131` 逐顆用 rec93 列出道具名。
- 六座全完成：中央 `(8,8)` 依序畫 tile `0x7b..0x80`；清 `0x8e`，完成後清暫態
  `flag 0x11`，把拉米亞停泊座標設為 `(0x30,0xbd)=(48,189)` 並設 vehicle bit2。

D3TXT04 rec93/96 的 `VAR_NUM` 在 mode1 是道具名稱，不是十進位數字；Ebiten 對話器新增
`varNumItem` context。

## 六珠來源

| 道具 | 原版來源 | Ebiten 狀態 |
|---:|---|---|
| `0x66` 綠 | CTY20 夜間 `(16,2)` sub2 handler35；flag3e gate | 本批接上 |
| `0x67` 藍 | CTY23 sec2 寶箱 `(24,14)` | 已有 |
| `0x68` 紅 | CTY27 sec1 寶箱 `(1,12)` | 已有 |
| `0x69` 紫 | CTY19 八頭大蛇 monster75 勝利 | 已有 |
| `0x6a` 黃 | CTY83 sec0 寶箱 `(4,2)` | JSON／元件交易已閉合；自然 production trace 待完成（見 `docs/102`） |
| `0x6b` 銀 | CTY64 scripted handler49 | 已有 |

## 飛行

- 復活不會自動搭乘。玩家離開祠堂，走上 world tile `0xfd`／停泊座標才進 mode2。
- mode2 無視地形碰撞、飛過城鎮不進城、不推晝夜、不遇敵。
- 空中按一般 Confirm 嘗試降落；目前格 `attr & 0x64 != 0` 時拒絕，合法時鳥停在該格。
- 停泊在地表時 world renderer 使用 tile `0xfd`；搭乘飛行使用
  `DQ3MAN.BLS entryBase 176` 的方向／步行幀。
- save 需保存 owned/aboard/park X/Y，以及至少 64-byte story flags。讀舊 32-byte save 時
  保留前 256 flags，高位以原版初值補齊。

## Ebiten 驗證

`game/phoenix_test.go` 使用真實原版資產驗證：

- CTY70 尺寸、spawn、六 event、守護神與條件 sprite 身分。
- CTY20 夜間 NPC 經正式命令輸入交付綠寶珠。
- 正式命令窗「調查」放珠、守護神對話、rec92/93/94、六幀動畫 transaction。
- 出祠堂走上拉米亞、跨地形飛行、合法／非法 Confirm 降落。
- 64-byte 與舊 32-byte save 相容。

視覺證據：`docs/phoenix_six_orbs.png`、`docs/phoenix_revival_animation.png`、
`docs/phoenix_revived.png`、`docs/phoenix_flight.png`。
