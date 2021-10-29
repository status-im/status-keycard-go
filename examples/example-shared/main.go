package main

/*
#cgo LDFLAGS: -L ../../build/libkeycard -lkeycard
#include "../../build/libkeycard/libkeycard.h"
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.MkdirTemp("", "status-keycard-go")
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		return
	}

	defer os.RemoveAll(dir)

	pairingsFile := filepath.Join(dir, "keycard-pairings.json")

	res := C.KeycardInitFlow(C.CString(pairingsFile))
	fmt.Printf("result: %+v\n", C.GoString(res))

	if err != nil {
		fmt.Printf("error: %+v\n", err)
		return
	}

}
