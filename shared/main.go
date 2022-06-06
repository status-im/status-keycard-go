package main

// #cgo LDFLAGS: -shared
// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"unsafe"

	skg "github.com/status-im/status-keycard-go"
	"github.com/status-im/status-keycard-go/signal"
)

func main() {}

var globalFlow *skg.KeycardFlow

func retErr(err error) *C.char {
	if err == nil {
		return C.CString("ok")
	} else {
		return C.CString(err.Error())
	}
}

func jsonToParams(jsonParams *C.char) (skg.FlowParams, error) {
	var params skg.FlowParams

	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
		return nil, err
	}

	return params, nil
}

//export KeycardInitFlow
func KeycardInitFlow(storageDir *C.char) *C.char {
	var err error
	globalFlow, err = skg.NewFlow(C.GoString(storageDir))

	return retErr(err)
}

//export KeycardStartFlow
func KeycardStartFlow(flowType C.int, jsonParams *C.char) *C.char {
	params, err := jsonToParams(jsonParams)

	if err != nil {
		return retErr(err)
	}

	err = globalFlow.Start(skg.FlowType(flowType), params)
	return retErr(err)
}

//export KeycardResumeFlow
func KeycardResumeFlow(jsonParams *C.char) *C.char {
	params, err := jsonToParams(jsonParams)

	if err != nil {
		return retErr(err)
	}

	err = globalFlow.Resume(params)
	return retErr(err)
}

//export KeycardCancelFlow
func KeycardCancelFlow() *C.char {
	err := globalFlow.Cancel()
	return retErr(err)
}

//export Free
func Free(param unsafe.Pointer) {
	C.free(param)
}

//export KeycardSetSignalEventCallback
func KeycardSetSignalEventCallback(cb unsafe.Pointer) {
	signal.KeycardSetSignalEventCallback(cb)
}
