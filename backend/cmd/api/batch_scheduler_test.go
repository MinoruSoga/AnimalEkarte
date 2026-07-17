package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHourlyTick(t *testing.T) {
	t.Run("次の時刻の0分に切り上げる", func(t *testing.T) {
		now := time.Date(2026, 7, 10, 14, 23, 45, 0, time.UTC)
		next := hourlyTick(now)
		assert.Equal(t, time.Date(2026, 7, 10, 15, 0, 0, 0, time.UTC), next)
	})

	t.Run("ちょうど0分でも次の時刻へ進む", func(t *testing.T) {
		now := time.Date(2026, 7, 10, 14, 0, 0, 0, time.UTC)
		next := hourlyTick(now)
		assert.Equal(t, time.Date(2026, 7, 10, 15, 0, 0, 0, time.UTC), next)
	})
}

func TestDailyAt2AM(t *testing.T) {
	t.Run("当日02:00より前なら当日02:00", func(t *testing.T) {
		now := time.Date(2026, 7, 10, 1, 30, 0, 0, time.UTC)
		next := dailyAt2AM(now)
		assert.Equal(t, time.Date(2026, 7, 10, 2, 0, 0, 0, time.UTC), next)
	})

	t.Run("当日02:00以降なら翌日02:00", func(t *testing.T) {
		now := time.Date(2026, 7, 10, 2, 30, 0, 0, time.UTC)
		next := dailyAt2AM(now)
		assert.Equal(t, time.Date(2026, 7, 11, 2, 0, 0, 0, time.UTC), next)
	})

	t.Run("ちょうど02:00なら翌日02:00", func(t *testing.T) {
		now := time.Date(2026, 7, 10, 2, 0, 0, 0, time.UTC)
		next := dailyAt2AM(now)
		assert.Equal(t, time.Date(2026, 7, 11, 2, 0, 0, 0, time.UTC), next)
	})
}

func TestRunScheduled(t *testing.T) {
	t.Run("ctx.Doneでループを終了する", func(t *testing.T) {
		var calls int32
		ctx, cancel := context.WithCancel(context.Background())
		immediateEachMS := func(now time.Time) time.Time { return now.Add(time.Millisecond) }

		done := make(chan struct{})
		go func() {
			runScheduled(ctx, "test-task", immediateEachMS, func(context.Context) error {
				atomic.AddInt32(&calls, 1)
				return nil
			})
			close(done)
		}()

		// 数回発火させてからキャンセルする
		time.Sleep(20 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("runScheduled did not return after ctx cancellation")
		}
		assert.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(1), "少なくとも1回はtaskが発火するはず")
	})

	t.Run("taskがエラーを返してもループを継続する", func(t *testing.T) {
		var calls int32
		ctx, cancel := context.WithCancel(context.Background())
		immediateEachMS := func(now time.Time) time.Time { return now.Add(time.Millisecond) }

		done := make(chan struct{})
		go func() {
			runScheduled(ctx, "failing-task", immediateEachMS, func(context.Context) error {
				n := atomic.AddInt32(&calls, 1)
				if n <= 2 {
					return errors.New("boom")
				}
				return nil
			})
			close(done)
		}()

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&calls) >= 3
		}, 2*time.Second, time.Millisecond, "エラーを返しても後続の発火が続くはず")

		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("runScheduled did not return after ctx cancellation")
		}
	})

	t.Run("taskがpanicしてもループを継続し次周期で再実行される", func(t *testing.T) {
		var calls int32
		ctx, cancel := context.WithCancel(context.Background())
		immediateEachMS := func(now time.Time) time.Time { return now.Add(time.Millisecond) }

		done := make(chan struct{})
		go func() {
			runScheduled(ctx, "panicking-task", immediateEachMS, func(context.Context) error {
				n := atomic.AddInt32(&calls, 1)
				if n == 1 {
					panic("boom")
				}
				return nil
			})
			close(done)
		}()

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&calls) >= 3
		}, 2*time.Second, time.Millisecond, "panicしても後続の発火が続くはず(プロセス/goroutineが落ちない)")

		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("runScheduled did not return after ctx cancellation")
		}
	})
}
