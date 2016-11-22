package readmap

import (
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestBasic(t *testing.T) {
	m := New()

	const outer = 75
	const inner = 250

	var keys [outer][inner]unsafe.Pointer
	var vals [outer][inner]unsafe.Pointer
	for i := 0; i < outer; i++ {
		for j := 0; j < inner; j++ {
			k := i * j
			v := -1 * i * j
			keys[i][j] = unsafe.Pointer(&k)
			vals[i][j] = unsafe.Pointer(&v)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < outer; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var j int
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic on i=%d, j=%d: %v", i, j, r)
				}
			}()
			for j = 0; j < inner; j++ {
				var v unsafe.Pointer

				if j > 0 {
					v = m.Get(keys[i][j/2])
					if got, want := v, vals[i][j/2]; got != want {
						t.Errorf("(i, j/2) = (%3d, %3d): got %x want %x", i, j/2, got, want)
					}
				}

				m.Insert(keys[i][j], vals[i][j])

				v = m.Get(keys[i][j])
				if got, want := v, vals[i][j]; got != want {
					t.Errorf("(i, j  ) = (%3d, %3d): got %x want %x", i, j, got, want)
				}

				v = m.Get(keys[i][j/3])
				if got, want := v, vals[i][j/3]; got != want {
					t.Errorf("(i, j/3) = (%3d, %3d): got %x want %x", i, j/3, got, want)
				}
			}
		}(i)
	}
	wg.Wait()

	r := (*readmap)(atomic.LoadPointer(&m.v))
	if len(r.keys) < outer*inner {
		t.Errorf("len(r.keys)=%d, need to store at least %d*%d", len(r.keys), outer, inner)
	}
	found := 0
	for i := 0; i < len(r.keys); i++ {
		if r.keys[i] != nil {
			found++
			if r.vals[i] == nil {
				t.Errorf("unexpected nil val for i=%d", i)
			}
		} else {
			if r.vals[i] != nil {
				t.Errorf("unexpected set val for i=%d", i)
			}
		}
	}
	if found != outer*inner {
		t.Errorf("got %d entries, want %d", found, outer*inner)
	}
}
