package dq3data

import (
	"os"
	"path/filepath"
	"testing"
)

func findAsset(t *testing.T, name string) []byte {
	cands := []string{}
	if a := os.Getenv("DQ3_ASSETS"); a != "" {
		cands = append(cands, filepath.Join(a, name))
	}
	cands = append(cands, filepath.Join("..", "..", "..", "assets_raw", name))
	for _, p := range cands {
		if d, err := os.ReadFile(p); err == nil {
			return d
		}
	}
	t.Skipf("%s 不在(設 DQ3_ASSETS 或放 assets_raw)— 跳過", name)
	return nil
}

func TestOpenBLK(t *testing.T) {
	raw := findAsset(t, "DQ3.BLK")
	blk, err := OpenBLK(raw)
	if err != nil {
		t.Fatalf("OpenBLK: %v", err)
	}
	// 對拍 C 版格式(docs/04):row_bytes=4、height=24。
	if blk.RowBytes != 4 || blk.Height != 24 {
		t.Fatalf("header 不符 C:row_bytes=%d height=%d(期 4/24)", blk.RowBytes, blk.Height)
	}
	if blk.TileSize != 4*24*4 {
		t.Fatalf("tile_size=%d(期 384)", blk.TileSize)
	}
	if blk.Count <= 0 {
		t.Fatal("count<=0")
	}
	t.Logf("DQ3.BLK: %d tiles, tile_size=%d(32×24 4-bit planar)", blk.Count, blk.TileSize)

	// 解 tile:所有像素 index 必須 0..15;整庫至少有非空 tile(不是全 0)。
	anyNonZero := false
	for i := 0; i < blk.Count; i++ {
		px := blk.Tile(i)
		for row := 0; row < 24; row++ {
			for x := 0; x < 32; x++ {
				if px[row][x] > 15 {
					t.Fatalf("tile %d px[%d][%d]=%d 超出 4-bit", i, row, x, px[row][x])
				}
				if px[row][x] != 0 {
					anyNonZero = true
				}
			}
		}
	}
	if !anyNonZero {
		t.Fatal("全部 tile 都是空的(解碼疑有誤)")
	}
	// 決定性:同 tile 解兩次結果一致。
	if blk.Tile(0) != blk.Tile(0) {
		t.Fatal("解碼非決定性")
	}
}

func TestBLKPlaneSegmentsUseHighToLowColorBits(t *testing.T) {
	// 1 byte/row、1 row、1 tile；四個 segment 的最左像素依序應為
	// color bit3、bit2、bit1、bit0。這固定檔案排列，避免再次把可辨識圖案
	// 誤當成 palette 正確性的證明。
	raw := []byte{
		1, 0, // row bytes
		1, 0, // height
		1, 0, // count
		0x80, 0x40, 0x20, 0x10,
	}
	blk, err := OpenBLK(raw)
	if err != nil {
		t.Fatal(err)
	}
	px := blk.Tile(0)
	for x, want := range []uint8{8, 4, 2, 1} {
		if px[0][x] != want {
			t.Fatalf("segment pixel x=%d got color index %d, want %d", x, px[0][x], want)
		}
	}
}
