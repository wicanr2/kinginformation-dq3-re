# RE 攻關日誌:`[0x722]` state machine(runner 觸發系統)

> 目標:逐行反組譯 `[0x722]` 的 state machine,搞清它到底是什麼、誰設它 → 決定 warp/portal/
> sub2 對話分支能不能接。**逐步記錄,成功失敗都留。** 位址:file = logical + 0x1370。

---

## Step 1:全掃 `[0x722]` 存取 → ★ 推翻「無 setter」

掃程式區所有指令(capstone),運算元含 `0x722`:
- **寫(setter):57 處**(大量 `mov word [0x722], 1`、少數 `=5/=ax/=bx/=si/=dx`、`inc/dec/add/sub`)
- **讀:165 處**(大量 `cmp [0x722], 1` / `cmp [0x722], 2`,也有 `cmp ...,3/4/9`)
- 其他:12 處(`push/pop [0x722]` 存堆疊、2 個 imul 假陽性)

**★ 重大更正**:docs/31/44/47 說「event id `[0x722]` 資料驅動、無靜態 setter」**是錯的**——
`[0x722]` 有 **57 個明確 setter**。且值域小(1/2/3/4/5/9),讀多為 `cmp ==1/2`。
⇒ `[0x722]` 比較像**狀態機的 state/mode 變數**(不是純資料驅動的 event id)。
之前「卡 runner」的前提需要重新檢視。

**最密集 cluster**:file 0x10a46–0x10e8f(logical 0xf6d6–0xfb1f)有極多 inc/dec/cmp/mov [0x722]
與 dx/ax/0 互動 → 疑為 **state machine 主體**。下一步逐行反組譯它。

## Step 2:反組譯密集 cluster(0xf6d6)→ 是「座標→region 命中測試」非事件分派

```
cmp cx,0xc0 / cmp dx,0x1a       ; 座標界檢
dx=(cx>>5)+1                    ; 座標 → cell 索引
cmp [0x722], dx                 ; 比對當前 [0x722] 與算出的 region
[0x71c]==3 分支:讀 [si+0x18/0x1a]=region{x,y}、[0x72c]=寬、[si+0x14]=高
迴圈掃 region 清單(6 byte/項 {x,y,..}),算命中 index=bp → cmp [0x722],bp; je 命中
```
⇒ **`[0x722]` = 當前 region/page 索引**(座標除算 + region 清單命中比對),**不是 runner event id**。
docs 把 `[0x722]` 當「event id」八成是**誤判**。下一步:找真正 runner 事件分派 `call [bx+0x3baa]`,
看 bx(event id)從哪來 —— 確認 [0x722] 與 event id 的關係。

## Step 3:找事件派發 → logical 0x9871 `call [bx+0x3baa]`

`call [bx+0x3baa]` 僅 1 處:L09871(file 0xabe1);`lea si,[0x3bb4]`(sub2 表)2 處:0x4ff2/0x96fe。
反組譯 runner 看 bx(event id)來源 + 與 [0x722] 關係。

## Step 4:✓ runner 派發確認 — bx = [0x722]

runner(函式 0xabb2 / L09842):
```
L09863: cmp byte [0x726],1 ; je → 跳過派發    ; [0x726]==1 gate
L0986a: mov bx, [0x722]                       ; ★ bx = [0x722]
L0986e: dec bx ; shl bx,1                      ; bx=([0x722]-1)*2
L09871: call [bx + 0x3baa]                     ; 派發 [0x722] 號 handler
```
⇒ **`[0x722]` 就是 runner 的 event/region 索引**。

## ★★ 核心結論(推翻舊 blocker)

`[0x722]` 同時被兩條路設定,**兩條都是靜態可決定的**:
1. **座標→region 命中測試**(Step 2,L0f6d6 cluster):玩家座標 → 除算 + region 清單比對 → 算出
   region index → 寫 [0x722]。region 清單是資料(可抽)。**玩家走到某區 → [0x722] 自動 = 該區號**。
2. **57 個明確 setter**(Step 1):故事推進直接 `mov [0x722], N`(進城/事件後等)。

⇒ **舊結論「[0x722] 無靜態 setter、純動態、追不到」是錯的**。它有 57 setter + 座標 region 規則,
**全部靜態**。所以 **warp/portal/sub2 的「何時觸發」是可以靜態 RE 出來並 wiring 的**——
不需要 DOSBox 動態!之前卡住是因為把 [0x722] 誤判成「不可知的動態 event id」。

## 決策點(待定)
要 wiring 的話,路徑:
- (A) 抽**座標→region 命中測試的 region 清單**([0x71c]/[0x72c]/region structs)→ 重建「玩家在哪區」。
- (B) 對每個 warp/portal handler,讀它 `cmp [0x722],N` 的 N → 對應到 region/事件。
- (C) sub2 對話分支:同樣 handler 內 `cmp [0x722]/旗標` → 選 rec(用戶推測 sub2 也吃 state machine,待驗)。
