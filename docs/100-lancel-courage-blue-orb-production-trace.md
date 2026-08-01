# 蘭西爾勇氣試煉／藍寶珠 production trace

> 2026-08-02。範圍：提頓綠寶珠合法 checkpoint → 蘭西爾 handler37 → 單人勇氣洞窟
> → CTY23 藍寶珠 → CTY75 handler62 復隊。本文件不宣稱已找到 story flag `0x13` writer。

## 結論

schema `0.1.15` 新增有限的 `temporary_solo_challenges`；DQ3 pack content `0.1.16` 將
NPC selector、世界落點、暫時模式 raw mask、返回目的地、文字與洞窟提示外部化。
CTY23 sec2 的藍寶珠 `{type=1,item=0x67,present_flag=0xad}` 同步遷入
`treasure_events`。Go 只保存「暫時單人→返回復隊」具名狀態機，不知道 DQ3 專屬 ID。

同一條從 boot 開始的 `TestOpeningProductionInputTrace` 已用正式輸入完成：航行、蘭西爾
住宿切回白天、最終鑰匙開門、神官對話與 Yes、四人變單人、地表步行進 CTY23、洞窟轉場、
調查取得藍寶珠、試煉途中 save/load、走原始出口回 CTY75、神官對話、三名同伴原樣復隊，
以及復隊後 save/load。狀態為 E3／V2；尚無同狀態 DOSBox 截圖，所以不是 V3。

## 非破壞性 IDA 9.4 證據

- 輸入：`assets_raw/DQ3.EXE`，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：`/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4，映像
  `ida-pro-9.4-ver2:civ1-py312-v1`。
- 原檔唯讀掛載；分析副本、`.i64`、sidecar 與輸出只在 `/tmp/dq3-courage-ida/`，未回寫、
  移動或重新命名原始檔，也未加入 Git。
- sidecar 語意以原始位址為 key，標示 `[PROVEN]`／`[UNKNOWN]`；自動名稱、指令與 bytes
  保留。production parity test 另鎖定 file `0x6dce..0x6e0a` 與 `0x73fa..0x740d` raw bytes。

| 推論等級 | 原始位置 | 已閉合語意 |
|---|---|---|
| proven | handler37 `(logical) 0x59e4..0x5a98`；caller `(IDA linear) 0x289ce` | 讀 flag `0x13`；未完成時播 D3TXT06 rec9，Yes 播 rec10、No 播 rec11；保存 active count `DGROUP 0x5077→0x5057`、強制 count=1、寫 world `(82,165)`、座標 `(25,7)` 並 OR `DGROUP 0x4f46` bit `0x80`。 |
| proven | handler62 `(logical) 0x608a`；caller `(IDA linear) 0x28a00` | 播 rec12，還原 `0x5057→0x5077`，AND clear mode bit `0x80`，重載正常場景。 |
| proven | handlers85／86 `(logical) 0x6685/0x668c` | CTY23 sec2 raw handlers只播 D3TXT07 rec66／67；不寫完成旗標。 |
| proven | world loader `(logical) 0x25fe` | `(82,165)` 的 flag `0x13` 讀分支為 clear→CTY75、set→CTY47；舊 C remake 與舊 `owPortal` 的方向相反。 |
| unknown | flag `0x13` writer | handler37、handler62、handlers85／86 與 CTY23 event table 都沒有 writer；不得因攻略或舊 C 程式自行在接受挑戰或取珠時設定。 |

原始影片 `dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4`
約 38:30–42:00 顯示四人進神殿、接受後只剩勇者、取得藍寶珠、返回神官後四人重新出現。
影片可證玩家可見順序與隊伍變化，不用來證明 flag writer。

## 原始資料 anchors

- CTY47 sec0：spawn `(36,55)`；scripted NPC `(25,16)`, handler37。
- CTY75 sec0：spawn `(11,7)`；scripted NPC `(26,7)`, handler62。
- CTY23 sec2：`46×24`、spawn `(20,9)`；event0 raw `01 67 00 ad`。
- D3TXT06：rec9／10／11／12／84；D3TXT07：rec66／67。JSON `glyph_codes` 與原 record
  包含 `0xffff` 結尾逐 word parity。

## 實作與驗證

- `temporary_solo_challenge.go`：對話、Yes/No、隊伍暫存、單人地表落點、返回復隊及
  overlapping entrance 的完成旗標讀分支。
- `save.go`：持久化 active event ID 與完整離隊同伴 records；洞窟中與復隊後均 round-trip。
- `warp.go`：刪除舊 C remake 猜測的反向 `(82,165)` portal；pack primitive 接管此入口。
- `temporary_solo_challenge_test.go`：拒絕不交易、接受／復隊、未知 writer 不被合成、存讀檔。
- `gamepack_test.go`：EXE raw anchors、CTY47／75 NPC、CTY23 event 與兩個文字 bank parity。
- runtime PNG：`docs/img/lancel_courage_trial_prompt.png`、
  `docs/img/courage_cave_blue_orb_obtained.png`，均已目視確認無透明、裁切或錯誤場景。

已通過：

```text
go test ./internal/... ./game -count=1 -timeout 20m
go build ./game ./internal/...
go build -o /tmp/dq3-remake .  # 在排除使用者空白 scratch tmp_dump.go 的容器工作副本
```

## 尚未關閉

- flag `0x13` writer 仍為 unknown；後續遇到需要完成態 rec84 的正式流程時，必須再次從 writer
  追到 consumer，不可把「取得藍寶珠」自動等同於設定該 flag。
- 原版與 remake 的同狀態畫面對拍仍缺，維持 V2。
- 二選一小視窗目前沿用既有共用 renderer；其幾何尚未完成原版結構的 D3 閉合，不得把現行
  `430/220/120/68` 提升為 game-pack canonical 設定。本輪只消除新事件的重複 hardcode。
- 下一個 campaign audit 從復隊且持藍寶珠的合法 checkpoint 繼續，依原版順序先找第一個
  實際 blocker；目前流程表的下一候選是海盜村紅寶珠。
