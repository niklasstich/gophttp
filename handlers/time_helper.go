//go:build test

package handlers

import (
	"time"
)

func SetTimeFunc(f func() time.Time) {
	timeFunc = f
}
