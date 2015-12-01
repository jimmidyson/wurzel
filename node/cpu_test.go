package node

import "testing"

func TestCPUInfo(t *testing.T) {
	v, err := CPUInfo()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("could not get CPU Info")
	}
	for _, vv := range v {
		if vv.ModelName == "" {
			t.Errorf("could not get CPU Info: %v", vv)
		}
	}
}

func BenchmarkCPUInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CPUInfo()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}
