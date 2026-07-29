# 92 — 達瑪神殿轉職：CTY17、個人物品與正式玩家流程

> 更新：2026-07-30。證據等級 D3／E3；畫面 V2。原版同狀態逐像素對拍仍待完成。

## 1. 本批關閉的錯誤

舊程式與文件把達瑪神殿當成 CTY49，且只記錄「轉職機制可能存在」。原版地表表中
CTY17 與 CTY49 同為 `(107,101,L0)`；地表 lookup 由低索引開始，實際進入的是
**CTY17**。CTY17 section0 `(10,10)` 的 sub2 NPC handler39 才是完整轉職入口。

另一個根本差異是物品所有權。原版每名角色有八個 item word；領悟之書必須由「所選
隊員」親自持有。舊 Go 共用背包即使持書，也不能作為轉賢者的 gate。

## 2. IDA Pro 9.4 證據鏈

本批優先使用 `/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4，於一次性 Docker 容器
分析既有 IDB；database、授權與輸出未加入 Git。

```text
CTY17 sec0 NPC (10,10), handler39
  → sub_15A9E (IDA 0x15a9e；logical 0x5a9e)
  → sub_10ABE (IDA 0x10abe；logical 0x0abe)
  → class selector sub_10C7F
  → member record writer 0x10be9..0x10c73
  → level／stats／items／spells／derived combat values
  → D3TXT06 rec45..55、66／67 的玩家可見流程
```

handler39 的 gate：

- 勇者 class0 禁止轉職。
- 所選角色最低 Lv20。
- 基本目標職業依序為 class `[1,2,3,4,6]`。
- class5 賢者只在來源 class7 遊玩者，或該角色八格物品中有 item `0x4a` 時出現。
- 全隊其他角色或主角背包中的 `0x4a` 不成立。

成功 writer：

- class 改為目標、EXP 清零、level 回到 1。
- 目前 HP／MP 與 `STR/AGI/maxHP/maxMP/INT/LUCK` 右移一位。
- **VIT 不減半**；舊 C remake 的「全能力減半」結論錯誤。
- 保留已學咒文，再依新職業 Lv1 擴充。
- 全部裝備解除，但物品仍留在該角色八格。
- 轉賢者時只消耗該角色持有的一本領悟之書；遊玩者無書時不消耗。

裝備面板也改為從所選角色自己的未裝備物品取候選；因此「給予裝備→同伴穿上→轉職
全卸裝→同伴重新穿上」可完整閉環，不會把卸下裝備困在無法操作的第二份清單。

## 3. 資料與引擎邊界

schema `0.1.7` 的 `reclass_events` 保存 CTY／section／NPC 座標、handler、Lv20 gate、
職業 raw ID、item `0x4a`、八格上限、文字與 evidence。D3TXT06 rec45–55、66–67
逐 word 存在 `texts.json`，並由原始檔 parity test 鎖定。

Go 只實作 `common:effect.reclass_original` 的有限狀態機與交易，不含 DQ3 專屬座標、
職業清單、文字或領悟之書 ID。對話的隊員名與職業名使用不同原版控制碼插值，不再把
所有變數錯誤折成主角名。

同批加入 `item_actions`：八格上限及「使用／給予／丟掉」標籤取自 D3TXT00 rec421。
本批正式接通「給予」；領悟之書可經 production 道具選單移到指定同伴。八格已滿時
交易失敗且不消耗。尚未取得完整 drop gate 證據，因此「丟掉」目前不執行推測性交易。

## 4. 驗收

- `TestDQ3DhamaReclassMatchesOriginalTextAndConfig`：JSON selector／gate 與
  D3TXT06 rec45–55、66–67 全 word parity。
- `TestReclassOriginalGatesAreMemberLocal`：勇者、Lv20、遊玩者與個人持書 gate。
- `TestReclassOriginalTransactionPreservesVITAndSpells`：精確減半欄、VIT、卸裝、
  消耗與舊咒文保留。
- `TestGiveItemUsesOriginalPersonalEightSlotCapacity`：八格滿時失敗不消耗。
- `TestOpeningProductionInputTrace`：從既有 boot trace 的領悟之書 checkpoint 正常
  離塔、航行至 CTY17、正式遭遇練到所選同伴 Lv20、以魔法鑰匙開門、rec421 給予、
  handler39 對話、選賢者、完成交易，再 save/load。

現行 runtime 圖：

- [`img/dhama_menu.png`](img/dhama_menu.png)：CTY17 賢者選項。
- [`img/dhama_reclass_success.png`](img/dhama_reclass_success.png)：原版 rec54 成功對話。

下一個連續 audit 由達瑪轉職讀檔 checkpoint 前往攻略步驟 22 提頓村，先閉合白天
CTY20 sec1《黑暗之燈》寶箱、地表使用變夜與其後仍可前往八頭大蛇洞窟；不可改用後段
事件 checkpoint 取代。
