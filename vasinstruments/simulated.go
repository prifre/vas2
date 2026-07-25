package vasinstruments

import "math/rand"

// Returnera en slice []int32 istället för en fast array [6]int32
func SimulatedAeroTrak() []int32 {
	return []int32{
		rand.Int31n(1000),
		rand.Int31n(500),
		rand.Int31n(100),
		rand.Int31n(50),
		rand.Int31n(10),
		rand.Int31n(5),
	}
}

func SimulatedDustTrak() int32 {
	return rand.Int31n(1000)
}

func SimulatedPTrak() int32 {
	return rand.Int31n(1000)
}
