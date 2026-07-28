package reservation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTimeToHHMM は PostgreSQL time 型が返す各種文字列表現を "HHMM" に正規化する挙動を検証する。
func TestTimeToHHMM(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "HH:MM:SS 形式（GORM が返す典型形）", in: "09:30:00", want: "0930"},
		{name: "HH:MM 形式", in: "09:30", want: "0930"},
		{name: "HHMM 形式（既に正規化済み）", in: "0930", want: "0930"},
		{name: "コロンなし・4文字ちょうど", in: "1830", want: "1830"},
		{name: "4文字未満はそのまま返す", in: "9:3", want: "93"},
		{name: "空文字はそのまま返す", in: "", want: ""},
		{name: "コロン除去後に4文字未満", in: "1:2", want: "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeToHHMM(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestPtrTimeToHHMM は nil / 非nil ポインタの両分岐を検証する。
func TestPtrTimeToHHMM(t *testing.T) {
	t.Run("nil ポインタは nil を返す", func(t *testing.T) {
		assert.Nil(t, ptrTimeToHHMM(nil))
	})

	t.Run("非nilポインタは正規化済みの新しいポインタを返す", func(t *testing.T) {
		s := "09:30:00"
		got := ptrTimeToHHMM(&s)
		require.NotNil(t, got)
		assert.Equal(t, "0930", *got)
		assert.NotSame(t, &s, got, "新しいポインタが返されること（元の変数を再利用しない）")
	})
}
