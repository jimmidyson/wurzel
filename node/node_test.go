package node

import "testing"

func TestInfo(t *testing.T) {
	v, err := Info()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Memory.Total == 0 {
		t.Errorf("could not get Node stats: %v", v)
	}
}

func BenchmarkInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Info()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}
