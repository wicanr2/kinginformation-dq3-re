# Ebiten Android 版操作界面規劃(觸控 UI,2026-07-02)

> 承 [docs/62](62-golang-ebiten-android-port-eval.md) 階段 6。目標:把桌面鍵盤操作的 DQ3 Ebiten port,
> 補上一套**手機可單手玩**的觸控界面,且不動到既有遊戲邏輯。
> 前提:移動/命令/戰鬥的控制模型已從 C remake 讀出(見下),觸控只是「另一個輸入來源」。

## 1. 遊戲實際的動作集(從 C remake 反推,非臆測)

| 情境 | 需要的抽象動作 | C 來源 |
|---|---|---|
| 地表/城鎮走動 | **4 向**(上下左右)| `dq3_scene_input`(scancode 0x48/0x50/0x4b/0x4d,只有 4 向)|
| 開野外命令窗 | **確定 A** | 面向 NPC 按確定對話;空地按確定開命令窗 |
| 野外命令窗(3×2)| 4 向移游標 + **A 選** + **B 關** | `dq3_cmdmenu`:對話/咒文/狀況/道具/裝備/調查(上下 %3、左右切欄)|
| 對話推進 | **A** | `dq3_dialogue` |
| 戰鬥指令 | 4 向 + **A** + **B** | `dq3_battlescene`:攻擊/咒文/防禦/道具/逃跑 |

**結論:整套遊戲只需 3 個實體控制 —— 十字鍵 + A(確定)+ B(取消)。** 這是經典 DQ 的 2 鍵 + 方向模型,
觸控化非常乾淨,不需要多按鈕手把。選單/戰鬥/對話全是同一組抽象動作在不同情境的重新解讀。

## 2. 核心架構決策:先抽象輸入,再綁定來源(不要讓觸控污染遊戲邏輯)

最容易做爛的地方是「在遊戲邏輯裡到處 `if 觸控…`」。正確做法是**一層抽象輸入狀態**,鍵盤與觸控都只是往裡面填值;
遊戲邏輯只讀抽象動作,不知道也不在乎來源。這也讓桌面/WASM/手機共用同一套邏輯(deep module,見 rulebook/70)。

```go
// internal/input/input.go — 唯一的輸入真相。遊戲邏輯只依賴這個介面。
type Dir int // DirNone / DirUp / DirDown / DirLeft / DirRight

type State struct {
	Dir      Dir  // 目前按住的方向(續走用)
	Confirm  bool // A:這一幀剛按下(edge)
	Cancel   bool // B:edge
	// 之後可加:MenuOpen、Skip(長按加速對話)…
}

// Source 產生一幀的 State。KeyboardSource + TouchSource 各實作一份;
// Combine(sources...) 做 OR 合併 → 桌面接鍵盤也能同時接觸控(除錯/模擬器方便)。
type Source interface{ Poll() State }
```

- 既有 `main.go` 目前直接讀 `ebiten.IsKeyPressed`;**第一步先把它抽成 `KeyboardSource`**,遊戲邏輯改讀 `input.State`。
  這步在桌面上零行為變化,但把「輸入來源」和「遊戲邏輯」解耦 —— 之後加觸控只是多一個 `Source`,不碰邏輯。
- `Confirm/Cancel` 用 **edge(剛按下)** 語意(對話推進、選單決定要防連發);`Dir` 用 **level(按住)** 語意(續走)。
  對照現有 `dq3_scene_input` 的格線移動 + `moveCooldown`,方向續走天然吻合。

## 3. 螢幕方向與版面

- **鎖橫向(landscape)**:DQ 是橫向卷軸,遊戲 framebuffer 是 640×350(≈1.83:1)。Android manifest 設
  `android:screenOrientation="sensorLandscape"`(在 `ebitenmobile bind` 產出的專案端設,不在 Go)。
- **邏輯座標固定 640×350**:`Layout(w,h)` 回傳 640×350,Ebiten 自動置中+letterbox。
  觸控座標經 Ebiten 也落在同一 640×350 邏輯系 → **控制元件用邏輯座標擺放,縮放由 Ebiten 統一處理**,不必自己算 DPI。
- **控制疊在畫面上(半透明 overlay)**,不佔遊戲畫面:左下十字鍵、右下 A/B,拇指自然落點。
  手機多為 2.0:1 以上的寬螢幕 → 640×350 會左右 pillarbox 出黑邊;控制**優先疊在遊戲畫面四角**(保證可達),
  黑邊只當視覺留白(phase 2 再考慮把控制外移到黑邊、擴大遊戲可視區)。

```
 ┌──────────────────────────────────────────────────────────┐
 │                                                            │  ← 640×350 遊戲畫面
 │                    (地表 / 城鎮 / 戰鬥)                     │
 │                                                            │
 │                                                            │
 │    ▲                                            ┌───┐      │
 │  ◀ ● ▶   ← 浮動十字鍵(左拇指)          B →   │ A │      │  ← A 決定(右拇指主鍵)
 │    ▼                                     ┌───┐  └───┘      │
 │            半透明,壓住時高亮             │ B │  ← B 取消   │
 └────────────────────────────────────────└───┘──────────────┘
                              (右下角另置小「☰ 選單/系統」鍵,可選)
```

## 4. 元件規格

### 4.1 浮動十字鍵(floating d-pad,左拇指)
- **浮動式**優於固定式:左半下緣任一點按下 → 十字鍵中心生在該點,方向 = 手指相對中心的角度。
  好處是適應不同手掌大小/持握,拇指不用去對準固定位置(手機 RPG 的主流做法)。
- **量化成 4 向**(遊戲只有 4 向):把角度切成上/下/左/右四象限;死區半徑內回 `DirNone`(避免抖動誤走)。
  可留一個 `eightWay` 旗標當未來擴充,但預設 4 向對齊 `dq3_scene_input`。
- 續走:只要手指還在十字鍵範圍且方向不變,持續回該 `Dir` → 配合現有 `moveCooldown` 一格一格走。

### 4.2 A / B 按鈕(右拇指)
- **A(決定/對話/推進)**:右下最大、最順手的鍵。**B(取消/返回)**:A 左上方稍小。
- 動作用 edge:手指按下當幀送一次 `Confirm/Cancel = true`(`inpututil` 風格的 just-pressed)。
- 情境化標籤(可選,提升可讀性):命令窗時 A 顯「決定」、B 顯「取消」;對話時 A 顯「▼」。

### 4.3 選單/系統鍵(可選)
- 因為「A 開命令窗」已涵蓋主要選單,**選單鍵非必要**;若要快速開系統選單(存檔/設定),
  在右上角放一顆小 `☰`。MVP 可先不做。

### 4.4 尺寸 / 安全區 / 回饋
- **實體尺寸**:按鈕直徑目標 ≥ 9mm(依 DeviceScaleFactor 換算像素);十字鍵作用半徑更大。
- **安全區**:控制與螢幕邊緣留 inset(避開瀏海、手勢導覽列、圓角)。`ebitenmobile` 可取安全區 insets。
- **回饋**:壓住時提高 alpha / 微放大;Ebiten 無原生震動,haptic 之後走 mobile bind 端補(非 MVP)。
- **透明度**:未壓 ~35%、壓住 ~70%,不擋畫面又看得見。

## 5. Ebiten 實作草圖(`internal/input/touch.go`)

```go
// 每幀:掃所有觸點,依落點分區,填 State。多點觸控是必須(十字鍵 + A 同時按)。
func (t *TouchSource) Poll() State {
	var st State
	ids := ebiten.AppendTouchIDs(nil)          // 目前所有觸點
	for _, id := range ids {
		x, y := ebiten.TouchPosition(id)       // 邏輯座標(640×350)
		switch t.zoneOf(x, y) {                // 左下=dpad、右下=A/B 群
		case zoneDpad:
			if t.dpadOrigin[id] == nil {       // 首次觸該區 → 定浮動中心
				t.dpadOrigin[id] = &pt{x, y}
			}
			st.Dir = quantize4(x, y, *t.dpadOrigin[id], deadzone)
		case zoneA:
			if isJustPressed(id) { st.Confirm = true }  // edge:用 inpututil.AppendJustPressedTouchIDs 判
		case zoneB:
			if isJustPressed(id) { st.Cancel = true }
		}
	}
	t.gcReleased(ids)                          // 放開的觸點清掉 dpadOrigin
	return st
}
```

整合:`Game.Update()` 改成 `st := g.input.Poll(); route(st)`,其中 `g.input = Combine(Keyboard{}, Touch{})`;
`route` 把 `st.Dir` → 現有的移動/選單游標,`st.Confirm/Cancel` → 對話/命令窗/戰鬥決定。
繪製:`Game.Draw()` 疊一層 `touchOverlay.Draw(screen)`(半透明十字鍵 + A/B),桌面可用旗標關掉。

## 6. 分階段建置

1. **抽象輸入層**:抽 `KeyboardSource` + `input.State`,遊戲邏輯改讀抽象動作(桌面零行為變化)。← 先做,解耦
2. **觸控 MVP**:浮動十字鍵 + A/B(4 向、edge 確定/取消)、半透明 overlay、多點觸控。地表能走、能對話、能開命令窗。
3. **情境化**:選單/戰鬥沿用同一抽象動作即通;補 A/B 情境標籤、死區/靈敏度調整。
4. **打磨**:安全區 inset、DeviceScaleFactor 尺寸、壓感回饋、（可選)系統鍵/震動/自訂佈局。
5. **打包驗證**:`ebitenmobile bind` → `.aar` → Android Studio,真機/模擬器驗多點觸控 + 橫向鎖 + 安全區。

## 7. 待決選項(需你拍板)

- **十字鍵型式**:浮動(適應手掌,推薦)vs 固定位置(所見即所得)。
- **方向數**:4 向(對齊原版,推薦)vs 8 向(斜走,但遊戲邏輯只吃 4 向,得再對映)。
- **控制擺位**:疊在遊戲畫面四角(簡單,MVP)vs 外移到 pillarbox 黑邊(畫面不被遮,但要改 Layout 映射)。
- **存檔命名等文字輸入**:用遊戲內建 glyph 選字盤(照 C remake)vs 叫系統輸入法(IME)。DQ 傳統走選字盤。

## 8. 與既有程式的接點

- 現有 `main.go` 的 `ebiten.IsKeyPressed(...)` → 收斂進 `KeyboardSource`(第一階段唯一改動點)。
- `Scene` 抽象(地表/城鎮共用)已就位 → 觸控不需區分場景,填同一組 `input.State` 即可。
- 命令窗/對話/戰鬥的 Go 版尚未移植(階段 4)→ 觸控層與它們透過 `input.State` 解耦,**兩邊可並行開發**,
  介面(抽象動作)先定好即可。
