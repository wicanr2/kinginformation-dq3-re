// Package rng 是 DQ3 忠實 DOS 亂數(移植 dq3_rng.c 的 DOS 模式:16-bit add 0x9018 + rol3 + mod)。
// 用於戰鬥擲值,使 roll 序列與原版/ C remake 位元一致(取代 math/rand)。
package rng

// RNG 是 16-bit DOS 亂數狀態(對 DS:[0xb5a])。
type RNG struct{ state uint16 }

// Seed 設種子。移植 dq3_rng_seed。
func (r *RNG) Seed(seed uint16) { r.state = seed }

// New 回一個已設種子的 RNG。
func New(seed uint16) *RNG { return &RNG{state: seed} }

// Next 回 [0,n) 的亂數。移植 dq3_rng_next(DOS 忠實:add 0x9018 → rol ax,1 ×3 → div n 取餘)。
func (r *RNG) Next(n int) int {
	if n <= 0 {
		return 0
	}
	s := r.state + 0x9018       // add ax,0x9018
	s = (s << 3) | (s >> 13)    // rol ax,1 ×3
	r.state = s                 // mov [0xb5a],ax
	return int(s % uint16(n))   // div bx → 餘數
}
