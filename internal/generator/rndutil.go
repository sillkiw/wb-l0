package generator

import (
	mrand "math/rand"
	"time"
)

var rng = mrand.New(mrand.NewSource(time.Now().UnixNano()))

func pick[T any](arr []T) T { return arr[rng.Intn(len(arr))] }

func randDigits(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('0' + rng.Intn(10))
	}
	return string(b)
}

func randHex(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hex[rng.Intn(len(hex))]
	}
	return string(b)
}
