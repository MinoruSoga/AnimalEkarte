package config

import (
	"fmt"
	"time"
	_ "time/tzdata" // tzdata をバイナリに埋め込み、zoneinfo を持たない実行環境でも JST を解決できるようにする
)

const JapanTimeZone = "Asia/Tokyo"

// JST is the process-wide cached JST *time.Location, loaded once at package
// init. tzdata is embedded via the blank import above, so LoadLocation cannot
// fail in practice; a failure here indicates a broken build and should panic
// immediately rather than be silently re-derived differently at each call site.
var JST = func() *time.Location {
	loc, err := time.LoadLocation(JapanTimeZone)
	if err != nil {
		panic(fmt.Sprintf("failed to load %s timezone: %v", JapanTimeZone, err))
	}
	return loc
}()

func ConfigureTimeZone() error {
	loc, err := time.LoadLocation(JapanTimeZone)
	if err != nil {
		return fmt.Errorf("load %s timezone: %w", JapanTimeZone, err)
	}
	time.Local = loc
	return nil
}
