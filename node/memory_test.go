package node

import "testing"

func TestMemory(t *testing.T) {
	v, err := Memory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Total == 0 {
		t.Errorf("could not get Memory stats: %v", v)
	}
}

func BenchmarkMemory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Memory()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}

func TestSwap(t *testing.T) {
	v, err := Swap()
	if err != nil {
		t.Errorf("error %v", err)
	}
}

func BenchmarkSwap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Swap()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}
