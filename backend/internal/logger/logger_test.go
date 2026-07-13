package logger

import (
	"sync"
	"testing"
)

// TestDefault_ConcurrentFallbackInitIsRaceSafe は、Init() 未実行の状態で複数 goroutine が
// 同時に Default() を呼んでもフォールバック初期化が sync.Once で保護され、
// -race 検出が発生しないことを検証する（BE-refactor.md B-4）。
func TestDefault_ConcurrentFallbackInitIsRaceSafe(t *testing.T) {
	var wg sync.WaitGroup
	const goroutines = 50
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			if l := Default(); l == nil {
				t.Error("Default() returned nil")
			}
		}()
	}
	wg.Wait()

	if defaultLogger == nil {
		t.Fatal("defaultLogger was not initialized by fallback")
	}
}
