package process

import "testing"

func TestIDs(t *testing.T) {
	v, err := IDs()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("could not get Process IDs: %v", v)
	}
}

func BenchmarkIDs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := IDs()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}

func TestList(t *testing.T) {
	v, err := List()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("could not get Processes: %v", v)
	}
	for _, vv := range v {
		if vv.Memory == nil {
			t.Errorf("could not get Processes memory: %#v", v)
		}
		if vv.MemoryEx == nil {
			t.Errorf("could not get Processes memory ex: %#v", v)
		}
	}
}

func BenchmarkList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := List()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}
