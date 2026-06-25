# RE 攻關日誌:咒文效果型別 dispatcher(monster-status 解鎖)

> 目標:找出「每個咒文的效果型別(傷害/回復/睡眠/麻痺/…)從哪來」,以接「怪物施狀態」。
> 線索:docs/13 提效果分類 dispatcher `0xbfd0`(依 `dl` 分流);docs/37 提 cast handler `0xad4c`
> (`[0x2511]=spell_id`)。**逐步記錄,成功失敗都留**——位址基準(file vs logical+0x1370)是常見坑。

位址換算:file = logical + 0x1370(resident 段)。DGROUP 資料 file = 0x16140 + DS_off。

---

## Step 1:反組譯 logical 0xbfd0(file 0xd340)→ ✗ 撲空

`dis_region 0xd340`:首 byte = `ret`(= 前一函式結尾),其後是 `int 0x33`(滑鼠)、
`[0x4f09]`、`lcall` 一堆螢幕/init 碼,**不是效果 dispatcher**。
⇒ 判斷:docs/13 的「0xbfd0」可能是 **file offset** 而非 logical(或我把 logical 當 file+0x1370 算反了)。
下一步試 file 0xbfd0(= logical 0xac60)。

## Step 2:反組譯 file 0xbfd0(logical 0xac60)→ ✓ 命中 dispatcher

```
L0ac60: cmp dl,0 ; jne →    ; dl==0:di=0x166訊息, [0x2511]=0x3b, call 0xb6fc
L0ac77: cmp dl,1 ; jne →    ; dl==1:[0x2511]=2, spell_id+0x79→[0x2591], call 0xafdb
L0ac97: (else)             ; [0x2511]=8, ..., call 0xb025
```
**這就是效果分類 dispatcher**:依 `dl`(效果型別)分流到不同 handler(0xb6fc/0xafdb/0xb025…)。
docs/13 的「0xbfd0」是 **file offset**(logical 0xac60),不是 logical 0xbfd0。坑:位址基準。
下一步:找 caller —— `dl` 從哪設(應是 per-spell 效果型別查表)。cast handler 線索 0xad4c。

## Step 3:反組譯 file 0xbebc(docs 說 cast handler 0xad4c)→ ✗ 認錯

file 0xbebc(logical 0xab4c)是 struct-copy + 怪物 spawn(讀 [bx+0xd80..] 怪資料、
call 0xfa39 RNG 挑選),**不是 cast handler**。docs/37 的「0xad4c」位址基準又混(file/logical)。
⇒ 放棄照 docs 位址,改**直接掃「誰 call dispatcher(logical 0xac60)」**反推 dl 設定處。

## Step 4:掃 E8 near-call → target == logical 0xac60(找 caller)

掃 E8 near-call → logical 0xac60:**0 處**。
⇒ dispatcher 非 near-call 進入 → (a)間接呼叫(跳表存函式指標)或(b)我認錯 dispatcher
(branch 設 [0x2511]=固定 spell id 0x3b/2/8,像「特定咒的多效果」而非通用分流)。
下一步:搜 0xac60 當 16-bit 資料 word(跳表入口)。

## Step 5:搜 word 0x ac60 在檔內(跳表/函式指標)

word 0xac60:檔內僅 1 處(file 0x15683,非 DGROUP)→ 非跳表入口。
⇒ file 0xbfd0 八成不是通用效果 dispatcher(branch 設固定 spell id 0x3b/2/8)。**換路徑**:
不從 docs 給的位址猜,改**搜「讀 descriptor 表 0x37c4 的指令」**——那是真 cast handler,
旁邊就會決定效果型別(base/target/效果)。

## Step 6:搜 code 引用 immediate 0x37c4(descriptor 表 = cast handler)

5 處引用,bytes 解讀:
- `8a 87 c4 37` = `mov al,[bx+0x37c4]`(讀 b0 base)— 3 處(0x9c89/0x9ce7/0xd4ad)
- `8a 97 c4 37` = **`mov dl,[bx+0x37c4]`**(讀進 **dl**!)— 2 處(0xa1bf/0xcc58)
⇒ **descriptor byte 直接進 dl = 效果型別來源**!下一步反組譯 0xa1bf 看 bx 索引 + dl 後續用法。

## Step 7:反組譯 logical 0xa1bf(mov dl,[bx+0x37c4] 上下文)

```
L0a1b4: mov ax,[0x2511]        ; spell_id
L0a1b7: shl bx,1 ; add bx,[0x2511]  ; bx = spell_id*3(3-byte stride ✓)
L0a1bd: mov dl,[bx+0x37c4]     ; dl = b0 = BASE(★ 不是獨立效果型別!)
L0a1c3: mov bp,dx              ; bp = base
L0a1c5: mov al,[bx+0x37c5]     ; al = b1 = target
L0a1d1: and al,0x18 ; cmp al,0x10 ; je  ; 依 target scope(單/群/全)分流
L0a1d8: call 0xaf33            ; 效果 handler(bp=base)
```
★ **修正假設**:dl = base,非獨立效果型別。descriptor 只有 base/target,**無「效果型別」byte**。
⇒ 傷害/回復/狀態的區分**不在 descriptor** → 在 spell_id 範圍 或 b1 其他 bit 或 handler 內。
下一步:反組譯效果 handler 0xaf33,看怎麼判「這咒是傷害/回復/睡眠」。

## Step 8:反組譯效果 handler 0xaf33

```
L0af33: (詠唱訊息:di=0x14c/d/e 訊息 id + text display + [0x259e]=1)
L0af79: pop bp ; cmp bp,0 ; je 0xc2f2 ; jne → jmp 0xc1f7  ; ★ base>0 → 套效果 0xc1f7;base==0 → 0xc2f2
```
⇒ **base 值本身分流效果**:base==0 = 狀態/輔助咒(走 0xc2f2)、base>0 = 傷害/治療(走 0xc1f7)。
與 Step6 的 base 觀察一致(damage 10-180、heal 15-200、即死 255、**狀態/輔助 base=0**)。
下一步:追 base==0 路徑(0xc2f2 / caller 後續)看「哪個狀態(睡/麻/幻)」怎麼定 → per-spell。

## Step 9:追 base>0 套效果 0xc1f7(傷害 vs 回復怎麼分)+ base==0 路徑

0xc1f7 在設 battle-entity 動畫/狀態欄位 `[bx+0xd4f/0xd50/0xd52]`、`[si+0x4b]` sprite 等,
深入多 handler(0xc733/0xd9b2…)。程式是**程序式效果套用**,非單一查表。

---

## 結論(9 步 trace 後):效果型別是「程序式」,無乾淨查表

**確立事實**:
1. cast handler 讀 descriptor `0x37c4`(`bx=spell_id*3`):**只有 b0=base、b1=target**,
   **無獨立「效果型別」byte**(b2 經 Step6 證實同家族不同,非效果型別)。
2. 效果由 **base 值分類**(程式 `cmp bp,0` 等):
   - **base==0** → 狀態/輔助咒(走非傷害路徑)
   - **base==255** → 即死/特殊(沙基ザキ系)
   - **base 1–254** → 傷害 / 回復(再依 target / spell_id 細分)
3. 套用是程序式(0xaf33 詠唱 → 0xc1f7 設 battle-entity 欄位),散在多 handler + per-spell-id。

**對 monster-status 的意義**:
- ✗ 沒有「spell→status」乾淨表可抽 → 不能純 data-driven。
- ✓ 但 **base-value 分類**是可用產物:base==0 那批就是狀態/輔助咒(含睡 ラリホー/幻 マヌーサ)。
- 務實接法:用 base==0 + 該咒名(glyph 比對)鎖定**特定狀態咒 id**,enemy 放該咒 → 映射到
  remake 的麻痺(睡≈不能行動)。比「復刻整套程序式效果」務實。

**這趟 RE 沒白費**:排除了「有效果型別 byte 可抽」的假設(否則會做錯一張表),並定出 base 分類規則。
