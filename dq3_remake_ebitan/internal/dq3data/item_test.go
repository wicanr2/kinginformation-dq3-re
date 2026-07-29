package dq3data

import "testing"

func TestItems(t *testing.T) {
	it, err := OpenItems(findAsset(t, "ITEM.DAT"))
	if err != nil {
		t.Fatalf("OpenItems: %v", err)
	}
	// 銅劍 id3 攻10(勇者初始武器,gameplay 實證);皮甲冑 id0x21 防8(勇者初始防具)
	if it.Attack(3) != 10 {
		t.Fatalf("銅劍 id3 攻應 10,得 %d", it.Attack(3))
	}
	if it.Defense(0x21) != 8 {
		t.Fatalf("皮甲冑 id0x21 防應 8,得 %d", it.Defense(0x21))
	}
	// 銅劍是武器(EquipSlot 0)、有價
	if it.EquipSlot(3) != 0 {
		t.Fatalf("銅劍應武器部位 0,得 %d", it.EquipSlot(3))
	}
	if it.EquipSlot(50) != 3 {
		t.Fatalf("皮帽子 code50/b4=0x60 應為兜部位 3,得 %d", it.EquipSlot(50))
	}
	if it.EquipSlot(58) != 2 {
		t.Fatalf("青銅盾 code58/b4=0x80 應為盾部位 2,得 %d", it.EquipSlot(58))
	}
	if it.Price(3) <= 0 {
		t.Fatalf("銅劍應有價,得 %d", it.Price(3))
	}
	// 全 128 項可查、有一批有價
	priced := 0
	for c := 0; c < ItemCount; c++ {
		if it.Price(c) > 0 {
			priced++
		}
	}
	if priced == 0 {
		t.Fatal("無任何有價道具(疑解析錯)")
	}
	t.Logf("128 道具解析 ✓;銅劍 攻10 價%dG、皮甲 防8;有價 %d 項", it.Price(3), priced)
}
