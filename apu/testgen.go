package apu

import (
	"math"
)

type testGen struct {
	v int32
}

const (
	amplitude float64 = 32767
)

func (t *testGen) GenerateSamples(buffer []int16) bool {
	for i := 0; i < len(buffer); i++ {
		val := amplitude * math.Sin(float64(t.v)*2*math.Pi/float64(freq))
		buffer[i] = int16(val)
		t.v += 620
	}
	return true
}
