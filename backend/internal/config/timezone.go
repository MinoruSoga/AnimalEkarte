package config

import (
	"fmt"
	"time"
	_ "time/tzdata" // tzdata をバイナリに埋め込み、zoneinfo を持たない実行環境でも JST を解決できるようにする
)

const JapanTimeZone = "Asia/Tokyo"

func ConfigureTimeZone() error {
	loc, err := time.LoadLocation(JapanTimeZone)
	if err != nil {
		return fmt.Errorf("load %s timezone: %w", JapanTimeZone, err)
	}
	time.Local = loc
	return nil
}
