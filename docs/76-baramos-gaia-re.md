# R-4 巴拉摩斯與蓋亞大洞窟 RE／實作記錄

日期：2026-07-28

## 結論

- 巴拉摩斯外城：CTY66，地表 `(43,130)`。
- boss 房：CTY65 sec0；CTY66 `(30,11)` → CTY65 `(10,13)`。
- boss 物件：CTY65 `(8,3)`，sprite37、sub2 handler70。
- formation：`ds:4EDF = 01 27 06 79 01`，怪物 `0x79` 一隻。
- 開戰／戰後對白：D3TXT06 rec85／rec86。
- 正常只戰一次。「打兩次」是精訊版巴拉摩斯打不死 bug 的重試現象。
- 勝利：remake milestone `0x213`；原版 story flag `0x29` 由 1 清為 0。

## IDA／raw byte 證據

IDA 呼叫鏈：

```
sub_1622A
  lea si, ds:4EDF
  call sub_1BE89
  cmp [2702],0
  ...
  mov bx,29h
  call sub_16EF4
```

formation loader `sub_1BF35`：

- `[si]` 是群數。
- `[si+1]`、`[si+2]` 是戰鬥背景／設定欄。
- 從 `[si+3]` 複製 `count*2` bytes 的 `{monster,count}`。

EXE SHA-256：
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。

## 阿里阿罕後續

巴拉摩斯勝利後回 CTY25 sec1 王座 `(9,7)`：

1. D3TXT01 rec98：國王讚揚。
2. D3TXT01 rec99：大魔王索瑪現身。
3. CLR story flag `0x4d`。

`0x29`、`0x4d` 都是「初值 1、事件完成後清除」，不是舊文件所寫的正向 SET。

## 自然下降

地表 `(54,129)` 依 flag `0x4d` 選城：

- 1：CTY71，洞窟封閉。
- 0：CTY72，地震後開放。

CTY72 的洞口環使用普通可走 tile 加 `hiMap subid=1`，指向第二筆 transition：

```
CTY72 → CTY77 sec0 @(16,9)
```

CTY77 的出口：

```
CTY77 → destSec 0xfe → DQ3UND @(85,67)
```

因此首次下降不是 runner event86，也不直接呼叫 remake `descend()`。

## 驗證

`game/baramos_test.go` 以 production `InputState` 驗證：

1. 命令窗「話す」→ rec85 → 單隻 `0x79`。
2. 勝利 → CLR `0x29` + rec86。
3. 回王座 → rec98/99 → CLR `0x4d`。
4. 地表走入洞窟載 CTY72。
5. 踩洞口 → CTY77。
6. 走 CTY77 `0xfe` 出口 → layer1 `(85,67)`。

KeyU 正式捷徑已移除。
