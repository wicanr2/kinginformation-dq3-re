package game

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCompanionEquippedSaleProductionInputTrace(t *testing.T) {
	traceOpeningProductionInputRoute(t, func(g *Game) {
		if len(g.companions) == 0 || g.companions[0].Items.Equipment()[2] < 0 {
			t.Fatal("正常換裝 checkpoint 缺少同伴穿戴盾牌")
		}
		code := g.companions[0].Items.Equipment()[2]
		defense := g.companions[0].Def(g.shop.items)
		// 圖片收據在開店前啟用持續繪圖，讓對話底圖取自當前場景。
		if os.Getenv("DQ3_SALE_RECEIPT_DIR") != "" && g.frame == nil {
			g.frame = ebiten.NewImage(ScreenW, ScreenH)
			g.renderFrame()
		}
		traceTalkFacility(t, g, facWeapon)
		send := func(in InputState) {
			t.Helper()
			in.DirHeld = -1
			if !in.Confirm && !in.Cancel {
				in.DirHeld = in.DirEdge
			} else {
				in.DirEdge = -1
			}
			if err := g.step(in); err != nil {
				t.Fatal(err)
			}
		}
		send(InputState{Cancel: true})
		if g.shop.stage != shopSellActor {
			t.Fatal("正式貨架取消未進出售選人")
		}
		for n := 0; n <= len(g.companions) && g.shop.sellActor != 1; n++ {
			send(InputState{DirEdge: 0})
		}
		send(InputState{Confirm: true})
		if g.shop.stage != shopSellItem || g.shop.sellActor != 1 {
			t.Fatal("正式出售入口未選同伴")
		}
		entries := g.shopSellEntries(1)
		target := -1
		for index, entry := range entries {
			if entry.code == code && g.companions[0].Items.IsWorn(g.companions[0].Items.Entries()[index]) {
				target = index
				break
			}
		}
		if target < 0 {
			t.Fatal("出售清單漏掉正常穿戴盾牌")
		}
		for n := 0; n < len(entries) && g.shop.sellCursor != target; n++ {
			send(InputState{DirEdge: 0})
		}
		before, seed := shopSaleSnapshot(g), g.prng.State()
		price, ok := g.shopSellPrice(code)
		if !ok || price <= 0 {
			t.Fatal("正常商店沒有盾牌售價契約")
		}
		send(InputState{Confirm: true})
		confirmed := shopSaleSnapshot(g)
		if g.shop.stage != shopSellConfirm || !reflect.DeepEqual(confirmed, before) || g.prng.State() != seed {
			logShopSaleSnapshotDiff(t, before, confirmed)
			t.Fatalf("進確認頁前不可移除物品、加錢或消耗亂數：stage=%d seed=%#x/%#x cursor=%d target=%d",
				g.shop.stage, g.prng.State(), seed, g.shop.sellCursor, target)
		}
		captureShopSaleFrame(t, g, "confirm")
		send(InputState{Cancel: true})
		if g.shop.stage != shopSellItem || !reflect.DeepEqual(shopSaleSnapshot(g), before) || g.prng.State() != seed {
			t.Fatal("出售確認取消必須保持持久狀態與亂數")
		}
		captureShopSaleFrame(t, g, "cancel")
		send(InputState{Confirm: true})
		send(InputState{Confirm: true})
		want := before
		want.Comps[0].itemStore.Remove(entries[target].position)
		want.Comps[0].Items = encodedItemStore(want.Comps[0].itemStore)
		want.HeroGold += price
		if !reflect.DeepEqual(shopSaleSnapshot(g), want) || g.prng.State() != seed {
			t.Fatalf("確認後只能清除同伴所選裝備並加一次錢：shield=%d gold=%d want=%d",
				g.companions[0].Items.Equipment()[2], g.heroGold, want.HeroGold)
		}
		if g.companions[0].Def(g.shop.items) != defense-g.shop.items.Defense(code) {
			t.Fatal("出售後仍保留盾牌防禦加成")
		}
		captureShopSaleFrame(t, g, "sold")
		traceCloseShop(t, g)
		if err := g.Save(); err != nil {
			t.Fatal(err)
		}
		saved := shopSaleSnapshot(g)
		loaded, err := newProductionTraceGame(g.assets)
		if err != nil {
			t.Fatal(err)
		}
		if err := loaded.Load(); err != nil || !reflect.DeepEqual(shopSaleSnapshot(loaded), saved) {
			t.Fatalf("正常交易存讀檔未保持：%v", err)
		}
		// Load 只還原資料；下一步沿目前遊戲實例讀回。
		if err := g.Load(); err != nil || !reflect.DeepEqual(shopSaleSnapshot(g), saved) {
			t.Fatalf("目前玩家讀回交易未保持：%v", err)
		}
		captureShopSaleFrame(t, g, "loaded")
		// 商店輸入不推進移動冷卻；先以正常空白幀等候。
		for n := 0; n < 60 && g.cd > 0; n++ {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatal(err)
			}
		}
		if g.cd > 0 || g.showTitle {
			t.Fatal("交易讀回後未回到可移動的玩家狀態")
		}
		x, y := g.px, g.py
		for dir := 0; dir < 4 && g.px == x && g.py == y; dir++ {
			// 撞到櫃台也會重設冷卻；每次換方向前正常等候。
			for n := 0; n < 60 && g.cd > 0; n++ {
				if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
					t.Fatal(err)
				}
			}
			if err := g.step(InputState{DirHeld: dir, DirEdge: dir}); err != nil {
				t.Fatal(err)
			}
		}
		if g.px == x && g.py == y {
			t.Fatal("讀回後無法以正式輸入移動到下一格")
		}
		captureShopSaleFrame(t, g, "next-step")
		if out := os.Getenv("DQ3_SALE_RECEIPT_DIR"); out != "" {
			receipt := struct {
				Scope        string `json:"scope"`
				Seed         uint64 `json:"seed_before_sale"`
				Item         int    `json:"item_code"`
				Price        int    `json:"price"`
				Defense      int    `json:"defense_before"`
				DefenseAfter int    `json:"defense_after"`
			}{"normal new-game prefix to companion equipped sale, cancel, save/load and next step; original shop dynamic parity not claimed", uint64(seed), code, price, defense, g.companions[0].Def(g.shop.items)}
			encoded, err := json.MarshalIndent(receipt, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, "receipt.json"), append(encoded, '\n'), 0644); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func shopSaleSnapshot(g *Game) saveState {
	s := g.snapshot()
	// Flags 由 map 匯出，集合比較須排序這份複本。
	sort.Ints(s.Flags)
	return s
}

func logShopSaleSnapshotDiff(t *testing.T, before, after saveState) {
	t.Helper()
	a, b := reflect.ValueOf(before), reflect.ValueOf(after)
	for i := 0; i < a.NumField(); i++ {
		if !reflect.DeepEqual(a.Field(i).Interface(), b.Field(i).Interface()) {
			t.Logf("交易snapshot差異 %s：before=%v after=%v", a.Type().Field(i).Name, a.Field(i).Interface(), b.Field(i).Interface())
		}
	}
}

func captureShopSaleFrame(t *testing.T, g *Game, name string) {
	t.Helper()
	out := os.Getenv("DQ3_SALE_RECEIPT_DIR")
	if out == "" {
		return
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	before, seed := shopSaleSnapshot(g), g.prng.State()
	g.renderFrame()
	if !reflect.DeepEqual(shopSaleSnapshot(g), before) || g.prng.State() != seed {
		t.Fatal("截圖不可修改交易狀態或亂數")
	}
	output, err := os.OpenFile(filepath.Join(out, name+".png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(output, &image.RGBA{Pix: append([]byte(nil), g.rgba...), Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
	closeErr := output.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("交易runtime PNG：%v/%v", err, closeErr)
	}
}
