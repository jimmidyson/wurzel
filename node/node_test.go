package node

import "testing"

func TestNode(t *testing.T) {
	v, err := Node()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Memory.Total == 0 {
		t.Errorf("could not get Node stats: %v", v)
	}
}

func BenchmarkNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Node()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}
