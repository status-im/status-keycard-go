package main

/*
#cgo LDFLAGS: -L ../../build/libkeycard -lkeycard
#include "../../build/libkeycard/libkeycard.h"
extern char* HelloWorld();
*/
import "C"

import (
	"fmt"
)

func main() {
	res := C.HelloWorld()
	fmt.Printf("result: %+v\n", C.GoString(res))
}
