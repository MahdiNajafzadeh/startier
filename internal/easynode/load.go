package easynode

import "time"

func Load[T any](v *T) *T {
	for {
		if v != nil {
			return v
		}
		time.Sleep(time.Millisecond * 100)
	}
}
