package handler

import "time"

func localTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(time.Local)
}

func localTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := localTime(*t)
	return &v
}
