package util

import "math/rand"

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1) //nolint: gosec // No se genera un token
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}
