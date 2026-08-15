package lstep

import (
	"net/url"
	"testing"
)

func TestListLstepCsvImportsQuery_ToLimit(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want int
	}{
		{name: "default", raw: "", want: 20},
		{name: "custom", raw: "50", want: 50},
		{name: "lower boundary", raw: "1", want: 1},
		{name: "upper boundary", raw: "100", want: 100},
		{name: "zero falls back", raw: "0", want: 20},
		{name: "too large falls back", raw: "101", want: 20},
		{name: "non numeric falls back", raw: "abc", want: 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values := url.Values{}
			if tc.raw != "" {
				values.Set("limit", tc.raw)
			}
			got := newListLstepCsvImportsQuery(values).toLimit()
			if got != tc.want {
				t.Errorf("toLimit() = %d, want %d", got, tc.want)
			}
		})
	}
}
