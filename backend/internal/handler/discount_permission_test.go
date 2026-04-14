package handler

import (
	"testing"
)

// TestFloatEquals は浮動小数等価判定の挙動を検証する（BUG-372）。
func TestFloatEquals(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want bool
	}{
		{"完全一致 - ゼロ", 0, 0, true},
		{"完全一致 - 整数", 10, 10, true},
		{"完全一致 - 小数", 10.5, 10.5, true},
		{"epsilon 内の差は等価とみなす", 10.0, 10.00001, true},
		{"epsilon 超える差は不一致", 10.0, 10.001, false},
		{"負値 - 等価", -5.0, -5.0, true},
		{"負値 - 不一致", -5.0, -5.5, false},
		{"異符号", 0.5, -0.5, false},
		{"極小差（浮動小数誤差想定）", 0.1 + 0.2, 0.3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := floatEquals(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("floatEquals(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
