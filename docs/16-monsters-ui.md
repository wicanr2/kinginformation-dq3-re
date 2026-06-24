# 怪物資料線:sprite 圖鑑 + 數值結構 + 名稱 / UI 字串來源

把精訊版 DQ3 的怪物資料三條線串起來並驗證:**圖**(DQ3MNS.SHP sprite)、
**數值**(D3MNS.DAT 41-byte 記錄)、**字**(怪名 / 指令選單字串)。三者以同一個
`monster_id`(= sprite_id = D3MNS 記錄 index,0..129)對齊。

驗證基準:`references/game3.png`(藍天草原、6 隻藍色史萊姆、敵名「史萊姆 6」、
指令 戰 / 逃 / 防 / 道具)。結論:**史萊姆 = monster_id 5**,sprite 5 render 出的藍色
水滴狀怪物與 game3.png 完全一致;其數值(HP 6 / 經驗 4 / 金錢 2)、怪名(D3TXT00
記錄 0x25d = 史萊姆)逐項對上。

成果:
- `docs/monsters/monster_sheet.png` — 130 隻怪物 sprite 圖鑑;128/129(歐里狄加 / 五頭龍大王)
  原為未完成 boss(空 sprite),由復原 sprite 補完(bug #3,見 README + dq3_restored_sprites.c)。
- `docs/monsters/spr_NNN.png` — 單隻 sprite(spr_005 = 史萊姆,spr_002 = 金屬史萊姆…)。
- `docs/data/d3mns_stats.json` — 130 筆怪物數值 + 怪名。

工具(docker uv venv 內跑):
- `tools/mns_sprite.py` — SHP 解碼 + 圖鑑 render(`one N out.png` 出單隻、無參數出整張)。
- `tools/mns_stats.py` — D3MNS.DAT 數值表 → JSON + 對照表(自動併入 txt00 怪名)。
- `tools/re_mns_dis.py` — DQ3.EXE sprite 載入 / 解碼路徑反組譯偵察(capstone)。

---

## 1. 怪物 sprite:DQ3MNS.SHP 格式(已解,128/130 render)

`DQ3MNS.SHP`(1,377,210 bytes)= **offset table + 130 隻 plane-major 未壓縮 sprite**。

### 檔頭:offset table

```
+0  131 × u32 LE     ; 每筆 = 該 sprite 資料在檔內的絕對位移
                     ; sprite_id 資料 = file[ offs[id] : offs[id+1] ]
                     ; 131 筆 = 130 怪 + 1 sentinel(尾哨兵 = filesize)
```

`offs[0] = 0x20c`(= 131×4),offset table 自占檔首 0x20c bytes,單調遞增。

### 每隻 sprite 的資料區(由 `sub_b2af` 色彩 blit 反組譯確認)

```
+0  u16 w        ; 每 plane 每 row 的 byte 數(top bit 0x8000 為旗標,取 & 0x7fff)
                 ; 像素寬 = w * 8
+2  u16 h        ; row 數 = 像素高
+4  4 個 plane 連續(plane-major,未壓縮):
        每段 = h 行 × w byte,每 byte = 8 像素的「該 plane bit」(MSB-first)。
        VGA Map Mask 由 ah=0x08 遞減到 0x01 → 第 1 段是 bit-plane 3、末段 bit-plane 0。
        色號 = Σ (plane_bit << plane)(0..15),對 MNSBK.PAL 取色。
        色號 0 視為透明(背景)。
```

> 與 BLK tile(docs/04)同屬 VGA **4-bit planar** 家族,差別:
> - SHP 是 **plane-major**(plane3,plane2,plane1,plane0 各整段);BLK tile 也是 plane-major
>   但 plane0→3 順序。SHP 的 plane 順序由 Map Mask `ah` 高位先寫決定(plane3 先)。
> - SHP 每隻寬高可變(header 帶),BLK 固定 32×24。
> - 怪名先前推定 sprite0 表頭 `{16,123,..}` —— 該推定錯誤:那 8 bytes 是 **offset table
>   的兩個 u32**(this_off / next_off),不是 sprite 表頭。真正的 sprite 表頭是資料區的
>   前 4 byte `{w, h}`。

### 反組譯佐證(seg0 位址)

| 函式 | 作用 |
|---|---|
| `sub_b19e` | 確保 SHP 開檔(int21 AH=3Dh),seek `id*4` 讀 offset table 8 byte,算 `len = next - this`,把資料整段讀進遠段 `[0x2540]` 緩衝 |
| `sub_b2af` | 色彩 blit:讀 `[si]=w`、`[si+2]=h`,4 plane × h row × w byte,逐 byte `and es:[di]`(透明遮罩),row pitch 0x54,Map Mask `out 0x3c4` 逐 plane(ah 0x08→0x01)|
| `sub_b31a` / `sub_b37c` | AND-mask(陰影 / 透明)分支:對 mask plane 走 RLE(top 2 bit:0x40=run 0xff、0x80=run 0x00、0x00=literal run、0xc0=單一 literal)。色彩主體不走 RLE |

### 驗證

- 130 隻中 **128 隻** render 出可辨識怪物(史萊姆、金屬史萊姆、青蛙、烏鴉、甲胄、龍、惡魔…)。
- sprite 5(史萊姆)= 藍色水滴 + 兩眼 + 嘴,與 game3.png 的 6 隻史萊姆完全一致。
- sprite 0(怪影)= 大型粉紅魔物含紅眼;sprite 2(金屬史萊姆)= 史萊姆同形但金屬質感。

---

## 2. 怪物數值:D3MNS.DAT 欄位結構(由 EXE 反組譯定位)

`D3MNS.DAT` = 5330 bytes = **130 筆 × 0x29(41)byte**,無檔頭,index = monster_id。
loader `sub_17ee` 整檔讀進 DS:0x0d78(`d3mns_buf`)。

### 已確認欄位(對照 EXE 讀取點 + 已知 DQ3 數值)

| 偏移 | 型別 | 語意 | EXE 讀取點(seg0)| 佐證 |
|---|---|---|---|---|
| `+0x00` | u16 | **HP 基值** | `0xab43` `add ax,[bx+0xd78]` | 史萊姆=6、金屬史萊姆=3 |
| `+0x02` | u16 | **HP 隨機加成範圍**(當前 HP = base + rnd(range))| `0xab35` `[bx+0xd7a]` → RNG | — |
| `+0x04` | u16 | f04(→ 敵 struct +0x10;稀疏,多為 0,推定特性 / 抗性 / 特攻 id,非生攻擊)| `0xab63` `[bx+0xd7c]` | — |
| `+0x08` | u8 | f08(→ 敵 +0xb;隨怪物強度遞增,**推定 攻擊力**)| `0xab4e` `[bx+0xd80]` | — |
| `+0x09` | u16 | f09(→ 敵 +0xc;最平滑遞增 9→100+,**推定 防禦 / 等級基準**)| `0xab55` `[bx+0xd81]` | — |
| `+0x0b` | u16 | f0b(→ 敵 +0xe;**推定 敏捷 / 速度**)| `0xab5c` `[bx+0xd83]` | — |
| `+0x18` | u8 | **逃跑抗性 / 逃跑難度**(party 逃跑成功判定門檻)| `0xb490` `[bx+0xd90]` cmp RNG | sub_b48a「逃跑了!」分支 |
| `+0x1f` | u8 | f1f(→ 敵 +0x14)| `0xab6a` `[bx+0xd97]` | — |
| `+0x21` | u16 | **經驗值**(勝利時加到隊伍經驗總和 [0x24f6])| `0xbce4` `[bx+0xd99]` | 史萊姆=4、金屬史萊姆=4140、金屬怪=32766 |
| `+0x23` | u16 | **金錢**(加到金錢總和 [0x24fa])| `0xbcdc` `[bx+0xd9b]` | 史萊姆=2、貝荷瑪史萊姆=38 |
| `+0x25` | u8 | f25 | `0xc51e` `[bx+0xd9d]` | — |
| `+0x26` | u8 | f26 | `0xc529` `[bx+0xd9e]` | — |
| `+0x28` | u8 | **出現權重 / 群量基準**(遭遇點數預算分配,×2 累進到 [0x2513])| `0xab76` `[bx+0xda0]` | docs/13 已確定 |

> 怪物生成時(`sub_aafc` 區段 0xab2c)逐欄複製 D3MNS 記錄到「活躍敵人」struct,
> 是定位欄位語意的主要依據:`[rec+0x00] + rnd([rec+0x02])` → 敵 HP(struct +0x16)。

### 跨筆驗證(怪名取自 D3TXT00 記錄 0x258+id)

| id | 怪名 | HP | 經驗 | 金錢 | 對照 |
|---|---|---|---|---|---|
| 2 | 金屬史萊姆 | 3 | **4140** | 5 | 金屬系:HP 極低、經驗暴高 ✓ |
| 3 | 金屬怪 | 4 | **32766** | 10 | 液態金屬:經驗近上限 ✓ |
| 5 | **史萊姆** | 6 | 4 | 2 | game3.png 的史萊姆 ✓ |
| 6 | 大烏鴉 | 7 | 5 | 2 | 弱怪遞增 |
| 26 | (中段 boss) | 550 | 2200 | 0 | 劇情 boss:高 HP、無金錢 |

HP / 經驗 / 金錢 隨圖鑑順序整體遞增,與 DQ3 怪物配置一致 → 三欄定位正確。
f04/f08/f09/f0b 的精確語意(攻擊 / 防禦 / 敏捷)尚需對 DOSBox 戰鬥傷害逐筆校,
目前以「敵 struct 對應欄 + 單調趨勢」標記推定,不下死結論。

完整 130 筆 + raw_hex 在 `docs/data/d3mns_stats.json`。

---

## 3. 怪名 / 指令選單字串來源:D3TXT00.TXT(疑問已釐清)

docs/12 推測 handler 的訊息 ID(di=0xbXX / 0x1XX / 0x160…)落在「另一張 D3MNS UI
字串表」。**追查結果:沒有獨立的 D3MNS UI 字串檔——該表就是 `D3TXT00.TXT`**。

依據:
- 文字繪製器 `text_draw`(lcall 111b:264)從遠段 `[0x252e]`(`g_txt_seg`)逐取 2-byte 字碼。
- loader(re/loaders.c:146)確認 **`[0x252e]` = D3TXT00.TXT 整檔緩衝段**(`seg_txt`)。
- 故所有 `text_draw(msg_id)` 的 `msg_id` 直接索引 **D3TXT00.TXT 的記錄指標表**;
  D3TXT00 共 759 筆,涵蓋道具 / 武器 / 防具 / 咒文 / 地名 / 職業 / 系統訊息 / **怪名**。

### 怪名:D3TXT00 記錄 `0x258 + monster_id`

`battle_draw_names`(sub_c59b)對每隻敵人:`al = sprite_id`、`ax += 0x258` → 存 `[0x2591]`
(記錄選擇器)→ `text_draw(0x160)` 畫「{怪名} {隻數}」。實解前 12 筆:

| id | rec | 怪名 |
|---|---|---|
| 0 | 0x258 | 怪影 |
| 1 | 0x259 | 荷依米史萊姆 |
| 2 | 0x25a | 金屬史萊姆 |
| 3 | 0x25b | 金屬怪 |
| 4 | 0x25c | 貝荷瑪史萊姆 |
| **5** | **0x25d** | **史萊姆** |
| 6 | 0x25e | 大烏鴉 |
| 7 | 0x25f | 青蛙怪 |
| 8 | 0x260 | 人面蝶 |
| 9 | 0x261 | 泡泡史萊姆 |
| 10 | 0x262 | 獨角兔 |
| 11 | 0x263 | 大食蟻獸 |

完整 130 隻怪名已併入 `docs/data/d3mns_stats.json` 的 `name` 欄;
全表可在 `docs/script/txt00.txt`(記錄 0x258..0x2d8 區段)逐筆對照。

### 戰鬥 / 系統訊息(同檔)

| rec | 文字 |
|---|---|
| 0x160 | `{變數}出現了!` |
| 0x161 | `怪物們被打倒了!`(勝利,docs/13 di=0x161)|
| 0x169 | `{變數}們被打倒了!`(全滅,docs/13 di=0x169)|
| 0x174 | `得到{變數}點經驗。` |
| 0x15c | `{變數}逃跑了!`(逃跑成功)|

### 指令選單字串(欄位名)

野外 / 戰鬥指令選單字也取自 D3TXT00 短記錄(單詞):

| rec | 字 | | rec | 字 |
|---|---|---|---|---|
| 0xb6 | 道具 | | 0xb9 | 頭盒 |
| 0xb7 | 武器 | | 0xba | 盾牌 |
| 0xb8 | 甲胄 | | 0x1d5 | 下一頁 |

職業名 0x1c9..0x1d0(勇者 / 戰士 / 武鬥家 / 僧侶 / 魔法使者 / 賢者 / 商人 / 遊玩者)、
咒文名 0x79..0xb4(美拉 / 荷依米 / 魯拉…)、地名 0x1ef..0x202、狀態 0x24c..0x252
(死亡 / 痲痺 / 睡眠 / 混亂 / 中毒)亦全在 D3TXT00。

> docs/12 殘留「需 dump D3MNS UI 字串表才能對指令名」一項至此可結:**字串源 = D3TXT00**,
> 各 di 訊息 ID 對應的中文已可由 txt00 逐筆查出。commands.c 的指令動作(話す / 調べる /
> 移動 / 條件事件)對應的選單顯示字,即上述 D3TXT00 短記錄。

---

## 殘留 / 待釐清

- **SHP sprite 128 / 129**(最後兩隻,可能為最終 boss / 多段大圖)header `w` 帶 0x8000
  旗標(w_raw=0xfff8 / 0x18ac、h 異常),非標準 plane-major 版面,本工具略過;
  其餘 128 隻正常。需另解該旗標(疑為「多段 / 大尺寸 / 含 AND-mask 分頁」)版面。
- **D3MNS 欄位 f04 / f08 / f09 / f0b / f1f / f25 / f26 精確語意**:已定位 EXE 讀取點與
  對應的敵人 struct 欄,但「攻擊 / 防禦 / 敏捷 / 抗性」對位仍為推定,需對 DOSBox 戰鬥
  傷害 / 命中逐筆校(對齊傷害公式後即可定名)。
- **記錄 +0x06 / +0x0d..+0x17 / +0x19..+0x1e / +0x20 / +0x22 / +0x24 / +0x27** 諸欄
  目前無 EXE 讀取點命中(可能為 padding 或別處間接讀),原樣保留於 raw_hex。
- **怪物 sprite runtime palette**:MNSBK.PAL 靜態色把史萊姆存成接近藍(sprite 5 已正確
  顯藍);金屬系 / 部分怪可能仍有 EXE runtime DAC 改寫(同 docs/04 海面 cycle),未影響辨識。

## 咒文習得表(RE 抽出,docs/18 sub_db5f)

升級學咒 `sub_db5f`(file 0xeecf)讀 DGROUP `0x36f9` 起的指標表(per-系 stride 8 =
`{spellA,levelA,spellB,levelB}`);系基底 bx:`0`=勇者系、`8`=僧侶系、`0x10`=魔法系。
平行清單 `level[i]`/`spell[i]`:到 level[i] 級學會 spell[i],咒文名 rec = `spell_id + 0x79`。
清單長度由下一指標界定(原版越界即 #5 亂學咒 bug)。職業→系:勇者→勇者系、僧侶→僧侶系、
魔法使→魔法系、賢者→僧侶+魔法、其餘無咒。`tools/gen_spells.py` 抽成 `dq3_spells_table.c`
(勇者 18 / 僧侶 24 / 魔法 31 咒);`dq3_spell` 查詢、`dq3_status_render_spells` 渲染 じゅもん 畫面。
