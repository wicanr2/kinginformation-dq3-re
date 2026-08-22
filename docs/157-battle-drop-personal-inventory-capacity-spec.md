# 157 — 戰後掉落與全隊個人八格容量規格（2026-08-22）

## 問題與停止線

商店切片改成原版每人八格後，正式新遊戲 trace 在第一次雷貝採買前仍顯示勇者持有 53 個
raw `0x41`（藥草）。來源不是商店，而是戰後結算對每次掉落直接 `append`。本切片閉合
「怪物掉落選擇 → 全隊空位搜尋 → 寫入／滿格」；不重新研究掉落機率表、戰鬥動畫或
record 345 動態姓名插值的精確 owner。

## 輸入、工具與位址契約

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4＋IDAPython；image `ida-pro-9.4-ver3:py312-x11-v3`。該 revision 只在既有
  私有 IDA image 補入 Ubuntu `libpython3.12`，可重建來源為
  `tools/Dockerfile.ida94-idapython`。
- exporter：`tools/ida_dump_battle_drop_inventory.py`。
- sidecar：`work/battle-drop-inventory-20260822.json`（研究輸出，不加入 Git）。
- 位址為 IDA linear；`logical=linear-0x10000`、`file=linear-0xec90`。原始名稱、bytes 與
  位址均保留，未 rename 或修改原始 EXE。

## 已證實的原版資料流

`sub_1C425`（IDA `0x1C425`，由 `sub_1C08B+0xA3` 呼叫）在勝利結算讀取 formation 第一隻
怪物，`0x1C509..0x1C537` 的掉落鏈為：

```text
0x1C509 mov al,[2321h]          ; 第一隻怪物 raw id
0x1C50E mov cl,29h / mul cl     ; D3MNS record stride
0x1C514 test [2518h],1          ; 強制掉落 gate
0x1C51C mov cl,[bx+0D9Dh]       ; record +0x25 threshold
0x1C520 call sub_1E6B9          ; rng
0x1C523 cmp al,cl / ja no_drop  ; AL <= threshold 才掉落
0x1C527 mov al,[bx+0D9Eh]       ; record +0x26 item raw id
0x1C52D mov [2591h],ax
0x1C530 mov [2593h],ax
0x1C533 inc [2591h]             ; 顯示用 one-based item record
0x1C537 call sub_1684E          ; 音效薄 wrapper，fall through 到八格 writer
```

`sub_16856`（IDA `0x16856`／file `0x7BC6`）不是全域無限背包 writer：

```text
0x16856 mov cl,[5077h]          ; 現行隊伍人數
0x1685C mov bx,0                ; DS:4F15 角色指標索引
0x1685F mov [259Ch],1           ; 1-based 角色序號
0x16867 mov si,[bx+4F15h]
0x1686B add si,3Ah              ; 個人八格 items 起點
0x1686E cmp word ptr [si],00FFh
0x16874/75 inc si / inc si
0x16876 inc dl
0x16878 cmp dl,8
0x1687D..83 下一角色，角色序號加一
0x16885 mov byte ptr [726h],1   ; 全隊皆滿，沒有 writer
0x1688C mov byte ptr [726h],0   ; 找到空位
0x16892 mov ax,[2593h]
0x16895 mov [si],ax             ; 只寫第一個空位
```

掉落 caller 隨後顯示 record `0x176`（「怪物留下寶箱」）與 `0x10B`（「發現物品」）；
若 `DS:0726==1` 再顯示 record `0x159`／345（「攜帶的東西太多，帶不走」）。record 345 的
原始 words 為
`01fc 018f fffb 0225 021c 0095 021b 017c fffe 014e 0165 01a7 0037 021c 00e4 0264 0039 ffff`。
`0xFFFB` 的精確角色插值 owner 本切片仍為 **unknown**，不得用勇者姓名猜值。

推論等級：掉落 threshold、item writer、全隊順序、每人八格、`0x00FF` 空值、滿格不寫與
`DS:0726` 分支均為 **confirmed**；record 345 的固定文字 words 為 **confirmed**，其
`0xFFFB` runtime 插值對象為 **unknown**。

## Remake 規格

1. 戰後掉落必須呼叫共用 `grantPartyItem`，依隊長至同伴順序尋找第一個有空位的個人物品欄。
2. 裝備與未裝備物品合計占用 game-pack `personal_inventory_slots`；缺契約時 fail closed。
3. 全隊滿格時不得寫入、不得覆蓋、不得讓任何 inventory 超過容量；經驗與金錢仍照常結算。
4. 成功才顯示現行取得物品提示；滿格 record 345 的玩家可見接線另需先閉合 `0xFFFB`
   插值 owner，不能在 Go 猜姓名。
5. 正式 trace 不可用清空 slice 或直接改數量修復；應讓早期藥草掉落自然停止於全隊容量。

## 驗收

- EXE parity：鎖定 `0x1C509..0x1C537` 及 `0x16856..0x16897` 的原始 bytes。
- component：隊長有空位、隊長滿而同伴有空位、全隊滿格、裝備占容量；滿格時 EXP／gold
  不受影響且沒有 inventory mutation。
- production：由標題開始重跑，第一次雷貝商店前任何角色都不得超過 pack 容量，再繼續到
  下一個玩家實際 blocker。
