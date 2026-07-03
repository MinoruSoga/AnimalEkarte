package main

import "testing"

func TestParseTotalCoverage(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    float64
		wantErr bool
	}{
		{
			name: "standard go tool cover -func output",
			in: "github.com/x/y/a.go:10:\tFoo\t100.0%\n" +
				"github.com/x/y/b.go:20:\tBar\t50.0%\n" +
				"total:\t(statements)\t83.4%\n",
			want: 83.4,
		},
		{
			name: "trailing blank lines and last total wins",
			in:   "total:\t(statements)\t10.0%\nsomething\ntotal:\t(statements)\t72.9%\n\n",
			want: 72.9,
		},
		{
			name:    "no total line",
			in:      "github.com/x/y/a.go:10:\tFoo\t100.0%\n",
			wantErr: true,
		},
		{
			name:    "malformed percent",
			in:      "total:\t(statements)\tNaN%oops\n",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTotalCoverage(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEvaluateRatchet(t *testing.T) {
	cases := []struct {
		name      string
		current   float64
		baseline  float64
		tolerance float64
		wantFail  bool
	}{
		{"baseline unrecorded warns only", 42.0, 0, 0.5, false},
		{"equal to baseline passes", 80.0, 80.0, 0.5, false},
		{"within tolerance passes", 79.6, 80.0, 0.5, false},
		{"exactly at floor passes", 79.5, 80.0, 0.5, false},
		{"below floor fails (refactor lowered coverage)", 79.4, 80.0, 0.5, true},
		{"large drop fails", 60.0, 80.0, 0.5, true},
		{"improvement passes", 85.0, 80.0, 0.5, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := evaluateRatchet(tc.current, tc.baseline, tc.tolerance)
			if res.fail != tc.wantFail {
				t.Fatalf("evaluateRatchet(%.1f, %.1f, %.1f).fail = %v, want %v (msg: %s)",
					tc.current, tc.baseline, tc.tolerance, res.fail, tc.wantFail, res.message)
			}
			if res.message == "" {
				t.Error("message must not be empty")
			}
		})
	}
}
