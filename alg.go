package readmap

import (
	"runtime"
	"unsafe"
)

func ptrHash(p unsafe.Pointer) uintptr {
	// TODO: something sensible here
	//
	// runtime.PtrHash is a gross hack to get to the runtime's aeshash64.
	// It looks something like:
	//
	//	func PtrHash(i unsafe.Pointer) uintptr {
	//		return aeshash64(noescape(unsafe.Pointer(&i)), 0xabcd1234)
	//	}
	//
	// A real solution needs not only to handle the other non-amd64 hash
	// conditions, but also has to think about a potential moving-pointer
	// future. It's hard to think about because the runtime's map
	// implementation hasn't thought about it either.
	return runtime.PtrHash(p)
}
