package main

// #cgo LDFLAGS: -shared
import "C"
import (
	"encoding/json"
	"fmt"
	"time"
)

var kctx *keycardContext

func main() {
	// example()
}

func example() {
	fmt.Printf("RUNNING EXAMPLE \n")
	res := Start()
	fmt.Printf("*** start %+v\n", C.GoString(res))
	time.Sleep(2)
	res = Select()
	fmt.Printf("*** select %+v\n", C.GoString(res))

	// res = Pair(C.CString("KeycardTest"))
	// fmt.Printf("*** select %+v\n", C.GoString(res))

	res = OpenSecureChannel(C.CString(`{"index":2, "key": "bff1d9e97be680a8a6f67b08057fac1c24f44a10070f009dc8472934edf51459"}`))
	fmt.Printf("*** OpenSecureChannel %+v\n", C.GoString(res))

	res = VerifyPin(C.CString(`{"pin": "123456"}`))
	fmt.Printf("*** VerifyPin %+v\n", C.GoString(res))

	res = GenerateKey()
	fmt.Printf("*** GenerateKey %+v\n", C.GoString(res))

	res = Stop()
	fmt.Printf("*** stop %+v\n", C.GoString(res))
	time.Sleep(10 * time.Second)
}

//export Start
func Start() *C.char {
	var err error
	kctx, err = startKeycardContext()
	if err != nil {
		return retValue("err", err.Error())
	}
	return retValue("ok", true)
}

//export Select
func Select() *C.char {
	if kctx == nil {
		l("select: not started")
		return retValue("error", "not started")
	}

	info, err := kctx.selectApplet()
	if err != nil {
		return retValue("error", err.Error())
	}

	return retValue("ok", true, "applicationInfo", info)
}

//export Stop
func Stop() *C.char {
	if kctx == nil {
		l("select: not started")
		return retValue("error", "not started")
	}

	if err := kctx.stop(); err != nil {
		return retValue("error", err.Error())
	}

	return retValue("ok", true)
}

//export Pair
func Pair(pairingPassword *C.char) *C.char {
	if kctx == nil {
		l("select: not started")
		return retValue("error", "not started")
	}

	pairingInfo, err := kctx.pair(C.GoString(pairingPassword))
	if err != nil {
		return retValue("error", err.Error())
	}

	return retValue("ok", true, "pairingInfo", pairingInfo)
}

//export OpenSecureChannel
func OpenSecureChannel(jsonParams *C.char) *C.char {
	if kctx == nil {
		l("select: not started")
		return retValue("error", "not started")
	}

	var params openSecureChannelParams
	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
		return retValue("error", err.Error())
	}

	err := kctx.openSecureChannel(params.Index, params.Key)
	if err != nil {
		return retValue("error", err.Error())
	}

	return retValue("ok", true)
}

//export VerifyPin
func VerifyPin(jsonParams *C.char) *C.char {
	if kctx == nil {
		l("select: not started")
		return retValue("error", "not started")
	}

	var params verifyPinParams
	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
		return retValue("error", err.Error())
	}

	err := kctx.verifyPin(params.Pin)
	if err != nil {
		return retValue("error", err.Error())
	}

	return retValue("ok", true)
}

//export GenerateKey
func GenerateKey() *C.char {
	if kctx == nil {
		l("select: not started")
		return retValue("error", "not started")
	}

	keyUID, err := kctx.generateKey()
	if err != nil {
		return retValue("error", err.Error())
	}

	return retValue("ok", true, "keyUID", keyUID)
}
