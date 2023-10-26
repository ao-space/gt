package msquic

/*
#include "quic.h"
*/
import "C"

func init() {
	ok := C.Init()
	if !ok {
		panic("msquic init failed")
	}
}
