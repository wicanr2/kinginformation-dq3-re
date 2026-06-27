# 原版精訊 DQ3 已知 bug / 異常行為(使用者實機觀察)

> 精訊版中文化做完了但有 bug+缺圖,當機/異常多。這裡記錄**使用者實機玩到的原版異常**,
> 供 remake 比對(remake 應修對、或刻意不復刻原版 crash)+ 反組譯時當線索。
> 7 個主 bug 見 docs/18;此處收原版 runtime 異常。

## ★ 金字塔內使用咒文 → 當機(使用者回報 2026-06-26)

- **現象**:原版一進入**金字塔(CTY13)**後,在裡面**使用咒文(野外/戰鬥?)就會當機**。
- **意義**:這是**原版固有 bug**(中文化做完但有 bug)。remake 不應復刻此 crash;若 remake 在金字塔
  施咒正常,即「修對了原版的 crash」(類似 7 主 bug 的處理態度)。
- **第一性原理 RE 分析(2026-06-27)**:
  - 咒文效果路徑已靜態解(`docs/re-log-spell-effect-dispatch.md`):cast handler 讀 descriptor
    `0x37c4`(base/target),效果分類 dispatcher `0xac60`(file 0xbfd0),效果套用 `0xaf33`→`0xc1f7`
    設 **battle-entity 動畫/sprite 欄位**(`[bx+0xd4f/0xd50/0xd52]`、`[si+0x4b]` sprite)。
  - **crash 根因(靜態可推的最可能類別)**:效果套用是**程序式動畫**,會碰場景/戰鬥的**逐圖資源**
    (sprite bank、palette、戰鬥背景 packbg)。金字塔=CTY13 為多 section 迷宮,其 section 載入的
    資源佈局與一般城鎮不同;施咒動畫存取某個在該 section **未載/越界/為 0xffff 的地圖相關指標** →
    crash。屬「程式碼正常、但被餵了該地圖未初始化的狀態」型 bug(對齊 #3 空 sprite 當機同類)。
  - **exact faulting 位址 = runtime-only**(rule 62 對「crash 在哪條指令 fault」有 runtime carve-out:
    那是執行期才確定的事)。**但 remake 不復刻,故無需定位**——見下「remake 處置」。
- **remake 處置 ✅(2026-06-27,已驗證免疫)**:
  - remake 咒文系統與**原版逐圖資源佈局完全解耦**:① 戰鬥施咒(`dq3_battlescene`)map-independent
    (不讀 CTY/section,背景用通用天空地面、非 packbg);② 野外「咒文」= 檢視畫面,不跑原版 overlay 動畫。
    ⇒ **結構上免疫**此 crash,不是靠 patch。
  - 驗證(headless,docker):`warp:13`(進金字塔)dump exit=0 不 crash;`DQ3_MON=14 battle` 施咒
    路徑(敵詠唱拉里荷→映射麻痺)完整跑完 exit=0。**remake 在金字塔施咒正常 = 修對了原版 crash。**
  - 註(2026-06-27 更新):**野外咒文效果已實作**(`field_cast_modal`:魯拉/烈米特→回地表、特黑洛斯→驅弱敵);
    與本 crash 無關。(舊註「尚未實作」已過時。)

## (待補)其他原版異常

- (BBS 史料/docs/21 已記:魔王打不死=#1、彩虹水滴變銀寶珠=#2、五頭龍空 sprite=#3 等)
