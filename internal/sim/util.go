package sim

import "math/rand"

// RandFloat returns a random float64 in [min, max)
func RandFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
