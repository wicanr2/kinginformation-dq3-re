# DQ3.EXE 野外指令子系統 + 對話流程(RE → C)

主迴圈狀態機(docs/10)在玩家按 Q / S / F1 時開「指令視窗」(`h_cmd_menu` = sub_9842),
選定指令後以 `[0x722]-1` 索引 **DS:0x3baa 的 12 項 near 指標表** 派發到各 field 指令動作。
本文件反編譯這 12 項指令動作,以及按 Enter(`h_confirm` = sub_7c43)推進的對話 / 事件流程,
並對照文字繪製器 111b:0264 的控制碼(docs/03)。

C 產出:[`re/commands.c`](../re/commands.c) + [`re/commands.h`](../re/commands.h)。
工具:`tools/re_cmd_dis.py`(docker capstone 內跑,反組譯本 slice 全部位址)。

```
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone \
           && python tools/re_cmd_dis.py all"     # 12 項派發表 + 各 handler + 對話 worker
           # python tools/re_cmd_dis.py 0xNNNN C  # 任意 file offset 反組譯 C 條
```

## DS:0x3baa 指令派發表(12 項 near 指標,實機讀出)

從主資料段 DS=0x14dd(file base 0x16140)讀出,與 docs/10 一致:

| sub-idx | near off | file off | 指令動作 | 反編譯位置 |
|---------|----------|----------|----------|-----------|
| 0 | 0x988d | 0x0abfd | 旗標 / 開關 + 條件式檔案 I/O(test [0xf]&0x18 → [0x286d]=2bit) | states.c `cmd_act0` |
| 1 | 0x98b9 | 0x0ac29 | 同群組(清 [0x286d],lcall 1053:16c) | states.c `cmd_act1` |
| 2 | 0x98cd | 0x0ac3d | test [0xf]&1 → [0x2784]=0,滑鼠 lcall 10b6:71/3a0 | states.c `cmd_act2` |
| 3 | 0x98e6 | 0x0ac56 | test [0xf]&1 → [0x2784]=1,lcall 10b6:385 | states.c `cmd_act3` |
| 4 | 0x9900 | 0x0ac70 | 重入主流程(kbd flush + call 0x7baa + 關窗 + jmp 0xa62c) | states.c `cmd_act4` |
| 5 | 0x504e | 0x063be | **道具 / 裝備 / 隊伍清單**(逐欄列出,空欄印不同訊息 ID) | states.c `cmd_item_list` |
| 6 | 0x5112 | 0x06482 | **話す(交談)** → sub_16dd 列名冊 / 對話對象頁 | commands.c `cmd_talk` |
| 7 | 0x5116 | 0x06486 | **旗標查詢 + 訊息**(查 flag 0x29,印 0xbd1/0xbd2) | commands.c `cmd_flag_msg` |
| 8 | 0x512f | 0x0649f | **移動 / 地圖切換動作** → sub_1a4c(座標 + tile 屬性解算) | commands.c `cmd_move_act` |
| 9 | 0x5133 | 0x064a3 | **調べる / 確認** → sub_7c50(= Enter 對話 worker chain) | commands.c `cmd_examine` |
| 10 | 0x5137 | 0x064a7 | **條件動作**(查 flag 0x29/0x4d → 訊息 + 開資訊視窗 0x2719) | commands.c `cmd_cond_a` |
| 11 | 0x51e9 | 0x06559 | **條件動作**(查 flag 0x29/0x2a/0x4d → 訊息 + 開資訊視窗) | commands.c `cmd_cond_b` |

> idx0..5 在 docs/10 / states.c 已就地反編譯,本文件接手 idx6..11 + 對話流程。
> 多數 entry 共用同一組工具:`game_flag_get`(sub_8279,bx=旗標索引 → al 0/1)、
> `game_flag_set`(sub_8264 / sub_824f)、`text_draw`(lcall 111b:264,di=訊息 ID)、
> `msg_window_reset`(sub_6372)、`msg_return_loop`(sub_6380 回主選單迴圈)、
> 以及選定後的「是 / 否確認 + 套用 + 重繪」worker `sub_7baa` 與動畫訊息序列 `sub_7c0c`/`sub_7bbe`。

### idx6..11 各 handler 語意(反組譯佐證)

- **idx6 cmd_talk(話す)** `sub_5112` 整段 `call sub_16dd; ret`。`sub_16dd`(file 0x2a4d)
  以 `[0x722]` 為頁內序號,逐欄掃結構(stride 0x14)把對話 / 名冊對象畫到視窗
  (`lcall 111b:779` 畫序號 / 屬性、`lcall 111b:43e` 畫框、`lcall 111b:214` 畫字串,
  空欄印 di=0x1d4),末端 `win_nav([0x3fba], [0x3fce])` 讓玩家選對象。對應 DQ
  「話す → 選對象」清單頁。
- **idx7 cmd_flag_msg** `sub_5116`:`msg_reset → game_flag_get(0x29) ? di=0xbd1 : 0xbd2
  → text_draw(di) → msg_return_loop`。純條件訊息輸出。
- **idx8 cmd_move_act** `sub_512f` 整段 `call sub_1a4c; ret`。`sub_1a4c`(file 0x2dbc)
  前段 `lcall 1053:240/dc`(資源)+ 畫面捲動 + 等鍵;後段(0x2dda 起)是
  **前方目標座標 = [0x4f33]+dx / [0x4f35]+dy → 邊界檢查([0xb28]×[0xb2a])→
  查 tile 屬性表 [tile*2+0x308e] → 設 [0x4f2d]=3(切城鎮 / 傳送)或更新移動 / 牆面
  旗標([0x258a] / [0x4f46])**。是地圖移動 / 出入口 / 傳送的核心解算(屬性位元細節
  屬地圖 / 碰撞 slice)。
- **idx9 cmd_examine(調べる)** `sub_5133` 整段 `call sub_7c50; ret`。`sub_7c50` 即
  **Enter 對話 / 事件 worker chain 的共用入口**(見下)。指令選單的「調べる」與按
  Enter 走同一條對話推進鏈,只是入口不同。
- **idx10 cmd_cond_a / idx11 cmd_cond_b** `sub_5137` / `sub_51e9`:查 `game_flag_get`
  (0x29 / 0x4d / 0x2a)分流,印對應訊息(di=0xbe5..0xc1b)後 `call 0x2719`
  (開資訊 / 狀態視窗)或進動畫迴圈(沿 [0x251d] 逐格推進 + `sub_7baa` 套用 + 設旗標)。
  是劇情條件觸發的場景動作(機關 / 階梯 / 特定事件)。

## 對話 / 事件流程:Enter(sub_7c43)→ worker chain

Enter handler `h_confirm`(states.c)在非特定模式下設 `[0x1f7]=1`(動作執行中)後進入
worker chain。chain 的共用入口是 `sub_7c50`(file 0x8fc0,亦即 idx9 調べる 的目標):

```
examine_chain (sub_7c50):
  confirm_worker();                 ; sub_9cd6  讀前方物件 → 跑事件 / 起對話
  if ([0x726] != 0) {               ; 上一段有後續才續段
      confirm_stage2();             ; sub_60c6
      if ([0x726] != 0)
          confirm_stage3();         ; sub_5b2d
  }
  lcall 1104:123                    ; 等鍵
  [0x1f7] = 0                        ; 動作結束
```

`[0x726]`(`g_action_result`)= 每段結果碼,非 0 才接續下一段,構成「逐段推進對話 / 事件」。

### confirm_worker(sub_9cd6)= 前方物件互動核心

反組譯(file 0xb04b 起,迴圈頭在 0xb037):掃前方 / 周邊格,以前方 tile `[0x2511]`
查物件事件表 `[tile*3 + 0x37c4]`:

```
for (g_obj_loop=1; g_obj_loop<=cx; g_obj_loop++) {
    u16 ft = [0x2511];                       // 前方 tile / 物件
    u8  ev = *(u8*)(ft*3 + 0x37c4);          // 物件事件碼(0xff = 無)
    if (ev != 0xff) {
        bp = ev; if (bp==0) bp=1;
        map_obj_walk();                      // sub_fa39 取對象資料
        [0x2593] = dx;                       // 對話 / 數值參數
        event_run(...);                      // sub_ba71 跑事件 / 起對話
        [0x2702]++;                          // 命中計數
    } else {
        event_run_else();                    // sub_bae8
        [0x2702]++;
    }
}
if ([0x2702] == 0) event_none();             // sub_ba41「沒有任何反應」
```

`sub_60c6`(stage2)以 bp=0x92/0x93/0x94 選角色旗標 + `game_flag_get` 分支,跑多段對話 /
道具獲得(di=0xc14..0xc18,`lcall 1053:240` 資源、設進度旗標 [0x12c..] 群);`sub_5b2d`
(stage3)讀選定對象 `[0x24a5]→[bx+0x4f15]` 隊伍 / 對象結構,複製名字(0x3a 起 8 欄)、
搬 [0x5058] 區塊、call `sub_1960`/`sub_481b`/`sub_452d`/`sub_2ce1`(視窗 / 重繪),
是對話結束後的隊伍 / 狀態更新(如加入同伴 / 道具入袋)。

> 事件命中後,**對話本體經 `text_draw(msg_id)` 逐字繪製**:讀 `[0x252e]` 遠段的
> 2-byte 字碼序列(D3TXT 記錄),控制碼按 docs/03 換行 / 換頁 / 插值。**「按 Enter 翻頁」
> 即由換頁碼 0xfffc 在繪製器內 `lcall 1104:123` 等鍵實現**(見下)。

## 文字繪製器 111b:0264 控制碼(程式碼 ↔ docs/03 雙向印證)

`text_draw`(入口 DI=msg_id)從 `[0x3e70]/[0x3e72]` 取視窗左上座標(x+2、y+0x10),
y 再加 `g_txt_line*0x10`(逐行下移)。取記錄起點 SI 後,**從遠段 `[0x252e]` 逐取 2-byte 字碼** `bx`,分支:

| bx 值 | commands.h 名 | 反組譯行為 | docs/03 語意 |
|-------|--------------|-----------|-------------|
| `0xffff` (-1) | `TXT_END` | 收尾;若 `[0x259e]!=1` 等鍵後 `retf` | 記錄結束 |
| `0xfffe` (-2) | `TXT_NEWLINE` | dx+=0x10;`[0x259b]==3` → 滿頁等待,否則 line++ | 換行 |
| `0xfffd` (-3) | `TXT_NEWLINE2` | 同 -2 路徑 | 換行(變體) |
| `0xfffc` (-4) | `TXT_PAGEFEED` | `lcall 111b:6` 翻頁 + `lcall 1104:123` **等鍵** | 換頁 / 段落分隔 |
| `0xfffb` (-5) | `TXT_VAR_NAME` | 跳 0x1290a 插值 | 動態插值(主角 / 角色名) |
| `0xfffa` (-6) | `TXT_VAR_ITEM` | 跳 0x12913 插值 | 動態插值(道具名) |
| `0xfff9` (-7) | `TXT_VAR_NUM` | 跳 0x1291e 插值(`[0x2593]` 數值) | 動態插值(金額 / 數值) |
| `0xfff5` (-0xb) | `TXT_VAR_B` | 跳 0x12927 | 動態插值(變數 B) |
| `0xfff6` (-0xa) | `TXT_VAR_A` | 跳 0x1293e | 動態插值(變數 A) |
| `0xffed` (-0x13) | `TXT_VAR_ED` | 跳 0x12955 | 動態插值 / 控制(最低控制碼) |
| 其他(<0xffed) | — | 畫 16×16 字模,bp(x)+=字寬,si+=2,line 計數推進 | glyph index → D3TXT00.FON 字模 |

要點:

- **一頁 4 行(`[0x259b]` 0..3)**:`cmp [0x259b],3` 為滿頁判定 —— 滿頁即停下等玩家按鍵翻頁,
  正是 DQ 對話框「顯示數行 → 按鍵續頁」行為。換頁碼 `0xfffc` 直接 `lcall 1104:123` 等鍵。
- **`[0x259e]`(`g_txt_nowait`)= 1** 時不等鍵(系統 / UI 訊息一次畫完);0 時逐頁等鍵
  (對話)。各 field 指令印 UI 訊息前多會設 `[0x259e]=1`。
- **`[0x252e]`(`g_txt_seg`)= 字碼來源遠段**:選單 / UI 訊息指向 D3MNS 字串區,劇情對話指向
  D3TXT 記錄區。這解釋 handler 的訊息 ID(di=0xbXX / 0x1XX,遠超 D3TXT00 的 759 筆)為何
  落在另一張 **UI / 選單字串表**,與 D3TXT 劇情記錄分屬不同段。

## 主資料段變數補充(DS=0x14dd)

| DS off | commands.h 名 | 含義 |
|--------|--------------|------|
| `0x3baa` | `g_cmd_menu_disp[]` | 指令選單二級派發表(12 項 near 指標,[0x722]-1 索引) |
| `0x252e` | `g_txt_seg` | 文字繪製器字碼來源遠段(選單 D3MNS / 對話 D3TXT) |
| `0x259b` | `g_txt_line` | 視窗內當前行號(0..3;==3 滿頁等鍵) |
| `0x259e` | `g_txt_nowait` | 1=自動推進不等鍵 / 0=逐頁等鍵 |
| `0x2593` | `g_txt_varnum` | 動態插值數值 / 對話參數暫存 |
| `0x3e70`/`0x3e72` | `g_txt_x`/`g_txt_y` | 文字視窗左上座標(繪字起點) |
| `0x2511` | `g_front_tile` | 玩家前方 tile / 物件索引(物件表 [ft*3+0x37c4]) |
| `0x2702` | `g_event_hit` | 互動命中事件計數(0=沒有反應) |
| `0x259c` | `g_obj_loop` | 物件 / 隊伍掃描迴圈索引 |
| `0x24a5` | `g_sel_obj` | 視窗導航選定的物件 / 對象 |
| `0x726` | `g_action_result` | 對話 / 事件每段結果碼(≠0 續段) |
| `0x1f7` | `g_action_active` | Enter 動作執行中旗標 |

旗標索引(各 handler `game_flag_get/set` 用):`0x29`=指令 / 場景可用閘、`0x4d`=條件動作第二旗標、
`0x2a`/`0x2c`=事件旗標、`0x92..0x94`=角色相關旗標(stage2)。

## 殘留 / 待釐清

- **指令中文名精確對位**:idx6=話す、idx9=調べる 已由「呼叫鏈 + 物件表互動」確認;
  idx8(移動 / 地圖切換)、idx10/11(條件事件)語意明確但「呪文 / 装備 / 作戰 / 搜尋」等
  典型 DQ 指令名需對 D3MNS UI 字串表(`[0x252e]` 指向段)逐 ID render 才能一一定名。
  目前指令名以「動作語意」標記,非逐字對 DQ 選單字。
- **UI 訊息 ID → 文字**:handler 的 di=0xbXX/0x1XX 索引 D3MNS 字串表(非 D3TXT),
  尚未 dump 該表;dump 後可把每個 di 對到實際中文,精確命名各指令分支。
- `sub_60c6`/`sub_5b2d` 的角色旗標 0x92..0x94 與隊伍結構欄位語意(加入同伴 / 道具入袋)
  需對照 player.dat / 隊伍資料結構進一步確認。
- 文字繪製器 VAR_*(0xfffb..0xffed)各插值來源(主角名 / 道具名 / 金額 / 變數 A/B)
  的填值常式(0x1290a..0x12955 各分支)細節屬繪字底層,本 slice 只記控制碼語意。
```
