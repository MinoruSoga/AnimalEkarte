package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// MemoryProfile はメモリ使用量をプロファイルする
func MemoryProfile(filename string) error {
	f, err := os.Create(filename) // #nosec G304
	if err != nil {
		return fmt.Errorf("could not create memory profile: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close error: %v\n", closeErr)
		}
	}()

	// GC を実行してから計測
	runtime.GC()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}

	fmt.Printf("Memory profile written to %s\n", filename)
	return nil
}

// CPUProfile は CPU 使用率をプロファイルする
func CPUProfile(filename string, duration time.Duration) error {
	f, err := os.Create(filename) // #nosec G304
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close error: %v\n", closeErr)
		}
	}()

	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("could not start CPU profile: %w", err)
	}
	defer pprof.StopCPUProfile()

	fmt.Printf("CPU profiling for %v...\n", duration)
	time.Sleep(duration)
	fmt.Printf("CPU profile written to %s\n", filename)
	return nil
}

// GoroutineProfile はゴルーチン数をプロファイルする
func GoroutineProfile(filename string) error {
	f, err := os.Create(filename) // #nosec G304
	if err != nil {
		return fmt.Errorf("could not create goroutine profile: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close error: %v\n", closeErr)
		}
	}()

	if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
		return fmt.Errorf("could not write goroutine profile: %w", err)
	}

	fmt.Printf("Goroutine profile written to %s\n", filename)
	fmt.Printf("Active goroutines: %d\n", runtime.NumGoroutine())
	return nil
}

// PrintMemoryStats はメモリ統計情報を出力
func PrintMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("\n=== Memory Stats ===")
	fmt.Printf("Alloc (allocated heap objects): %v MB\n", m.Alloc/1024/1024)
	fmt.Printf("TotalAlloc (total allocated): %v MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("Sys (system memory): %v MB\n", m.Sys/1024/1024)
	fmt.Printf("NumGC (garbage collections): %v\n", m.NumGC)
	fmt.Printf("Goroutines: %v\n", runtime.NumGoroutine())
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: profile [memory|cpu|goroutine|stats]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "memory":
		if err := MemoryProfile("profile_memory.pprof"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "cpu":
		duration := 10 * time.Second
		if len(os.Args) > 2 {
			if d, err := time.ParseDuration(os.Args[2]); err == nil {
				duration = d
			}
		}
		if err := CPUProfile("profile_cpu.pprof", duration); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "goroutine":
		if err := GoroutineProfile("profile_goroutines.pprof"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "stats":
		PrintMemoryStats()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
