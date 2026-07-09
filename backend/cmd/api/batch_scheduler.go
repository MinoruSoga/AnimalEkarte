package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/logger"
)

// runScheduled は nextFire が返す時刻まで待機して task を実行するループを ctx がキャンセルされるまで繰り返す。
// task のエラーはログに残すのみでループは継続する（LSTEP-BE-004/014, FEAT-383 バッチの既存挙動を維持）。
func runScheduled(ctx context.Context, name string, nextFire func(now time.Time) time.Time, task func(context.Context) error) {
	for {
		now := time.Now()
		next := nextFire(now)
		timer := time.NewTimer(next.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := task(ctx); err != nil {
				logger.Error(name+" failed", slog.String("error", err.Error()))
			}
		}
	}
}

// hourlyTick は次の時刻の0分を返す（毎時0分起動のバッチ用）。
func hourlyTick(now time.Time) time.Time {
	return now.Truncate(time.Hour).Add(time.Hour)
}

// dailyAt2AM は次の 02:00（now と同一ロケーション）を返す。
// now が当日02:00以降であれば翌日02:00を返す。
func dailyAt2AM(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
