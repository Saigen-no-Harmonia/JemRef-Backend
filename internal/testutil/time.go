package testutil

import "time"

// TimePtr time.Time型のポインタを返却する
func TimePtr(t time.Time) *time.Time {
	return &t
}
