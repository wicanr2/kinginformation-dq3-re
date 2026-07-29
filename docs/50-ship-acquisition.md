# 波魯多加王室任務與取船原版證據

> 現行產品：`dq3_remake_ebitan/`。歷史 C remake 的補洞、debug warp 與推測旗標已移除；
> 本檔只保存原始 EXE／CTY／D3TXT 可重現的結論。

## 場景與事件入口

| 項目 | 原版值 | 證據 |
|---|---:|---|
| 外城／王宮 | CTY16／CTY37，地表入口 `(26,72)` | `ctyLoc` 與 CTY transition |
| 國王 NPC | CTY37 section0 `(9,6)`，sub2，byte4 handler `26` | CTY37 section0 NPC record |
| 白天入口 | CTY16 `(2,5)` → CTY37 `(13,29)` | CTY16 section0 transition |
| 船停泊位置 | world `(25,73)`，layer `0` | `DQ3.EXE` handler26 writer |

王宮在夜晚（`dnPhase=2`）由入口無法抵達國王；正式 production trace 先在 CTY16 旅店住宿，
回到白天後再進王宮。`TestPortogaRoyalMissionMapReachability` 鎖定白天原始入口至國王
相鄰交談格可達，不能以直接把玩家放到王座旁的截圖代替。

## handler26 閉合

入口為 logical `0x5771`、file `0x6ae1..0x6b75`：

1. 讀 story flag `0x37`。
2. flag set（首次晉見）：
   - 顯示 D3TXT04 rec24；
   - 經 file `0x7bbe` 給 item `0x5b`《國王的信》；
   - 經 file `0x8264` 清 flag `0x37`；
   - 顯示 rec25。
3. flag `0x37` clear：
   - 檢查 flag `0x38`；
   - 經 file `0x7c0c` 檢查 item `0x5c`《黑胡椒》；
   - 沒有時顯示 rec26，不消耗任何物品；
   - 有時把該背包格寫成 `0xff`，清 flag `0x2c`，寫入船座標 `(25,73)`，
     交給船隻 consumer，顯示 rec27。
4. 完成後顯示 rec28。

舊文件使用的 `0x215` 是 remake 自造旗標，不存在於這條原版交易，禁止再作 production
條件。原版新遊戲會先設 story bits 40–255，因此 `0x37`／`0x38` 初值為 set。

## 對話與物品

D3TXT04 rec24–28 依序是：

- 首次交付東方尋找黑胡椒的任務；
- 授予《國王的信》並說明交給諾魯特；
- 尚未帶回黑胡椒；
- 收取黑胡椒並授船；
- 取船後對話。

玩家可見文字只存在
`internal/gamepack/packs/dq3_cht/data/texts.json`；事件規則只引用穩定 text ID。
item `0x5b` 是《國王的信》，船是 vehicle state，不是 item。item `0x5c` 由巴哈拉達救人線
取得，不能再用歷史 C remake 的「放進商店」補洞代替原版流程。

## Go／Ebitengine 實作

`events.json` 的 `staged_vehicle_exchange_events` 保存 NPC selector、兩個原版旗標、授予／
需求 item、船種與停泊座標、清除旗標及五段 text ID。Go 的
`staged_vehicle_exchange.go` 只實作跨版本可重用的具名交易：

- 首次對話關閉後才給信並清任務旗標；
- 缺黑胡椒時不消耗；
- 成功時只消耗一件黑胡椒、清宣告旗標、授予停泊中的船；
- 完成後重複交談不會重複授船；
- 缺欄位、錯誤引用或未知 vehicle kind 於載入時失敗即關閉。

原始 handler／CTY／D3TXT parity、交易 component test、白天地圖可達性及從新遊戲開始的
正式 `InputState` trace 共同驗收；單獨呼叫 handler 或直接設定座標不算玩家流程證據。

## 畫面紀錄

| 檔案 | 現行內容 |
|---|---|
| `docs/img/cty37.png` | 白天王座廳、對話前 |
| `docs/img/portoga_king_letter_intro.png` | rec24 首次任務 |
| `docs/img/portoga_king_letter_item.png` | item `0x5b` 取得通知 |
| `docs/img/portoga_king_letter_grant.png` | rec25 授信後對話 |
| `docs/img/portoga_king_wait.png` | rec26 尚無黑胡椒 |
| `docs/img/portoga_getship.png` | rec27 收胡椒並授船 |
| `docs/img/ship_on_overworld.png` | 船停泊在 world `(25,73)` |

這些是 remake runtime 圖；原版同狀態 DOSBox 畫面仍需按 V2／V3 規則並列核對，不可只因
文字 record 相同便宣稱像素完全一致。
