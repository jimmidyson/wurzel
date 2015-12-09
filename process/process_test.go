package process

import "testing"

func TestProcessIDs(t *testing.T) {
	v, err := ProcessIDs()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("could not get Process IDs: %v", v)
	}
}

func BenchmarkProcessIDs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ProcessIDs()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}

func TestProcesses(t *testing.T) {
	v, err := Processes()
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

func BenchmarkProcesses(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Processes()
		if err != nil {
			b.Errorf("error %v", err)
		}
	}
}
