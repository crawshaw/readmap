package readmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type Map struct {
	v     unsafe.Pointer // *readmap
	insMu sync.Mutex
}

func New() *Map {
	m := new(Map)
	r := &readmap{
		keys: make([]unsafe.Pointer, primes[0]),
		vals: make([]unsafe.Pointer, primes[0]),
	}
	atomic.StorePointer(&m.v, unsafe.Pointer(r))
	return m
}

func (m *Map) Get(k unsafe.Pointer) unsafe.Pointer {
	r := (*readmap)(atomic.LoadPointer(&m.v))
	i, found := r.find(k)
	if found {
		return atomic.LoadPointer(&r.vals[i])
	}
	return nil
}

func (m *Map) Insert(k, v unsafe.Pointer) {
	if k == nil {
		panic("readmap: nil key")
	}
	if v == nil {
		panic("readmap: nil value")
	}

	m.insMu.Lock()
	defer m.insMu.Unlock()

	r := (*readmap)(atomic.LoadPointer(&m.v))
	for {
		i, toofar := r.findNew(k)
		if !toofar {
			atomic.StorePointer(&r.vals[i], v)
			atomic.StorePointer(&r.keys[i], k)
			return
		}
		newlen := nextprime(len(r.keys))
	newmap:
		rNew := &readmap{
			keys: make([]unsafe.Pointer, newlen),
			vals: make([]unsafe.Pointer, newlen),
		}
		for vali, k := range r.keys {
			if k == nil {
				continue
			}
			i, toofar := rNew.findNew(k)
			if toofar {
				newlen = nextprime(newlen)
				goto newmap
			}
			atomic.StorePointer(&rNew.keys[i], k)
			atomic.StorePointer(&rNew.vals[i], r.vals[vali])
		}
		r = rNew
		atomic.StorePointer(&m.v, unsafe.Pointer(r))
	}
}

type readmap struct {
	keys []unsafe.Pointer
	vals []unsafe.Pointer
}

func (r *readmap) find(k unsafe.Pointer) (i int, found bool) {
	n := ptrHash(k)
	for _, step := range steps {
		n = (n + step) % uintptr(len(r.keys))
		v := atomic.LoadPointer(&r.keys[n])
		if v == k {
			return int(n), true
		}
		if v == nil {
			return int(n), false
		}
	}
	return -1, false
}

func (r *readmap) findNew(k unsafe.Pointer) (i int, toofar bool) {
	n := ptrHash(k)
	for _, step := range steps {
		n = (n + step) % uintptr(len(r.keys))
		v := atomic.LoadPointer(&r.keys[n])
		if v == k {
			panic("key already in map")
		}
		if v == nil {
			return int(n), false
		}
	}
	return -1, true
}

var steps = [...]uintptr{0, 1, 2, 4, 8, 16}
var primes = [...]int{23, 47, 97, 211, 431, 863, 1709, 3607, 7309, 14731, 29501, 60293, 134639, 297889}

func nextprime(p int) int {
	for _, prime := range primes {
		if prime > p {
			return prime
		}
	}
	panic("out of primes")
}
