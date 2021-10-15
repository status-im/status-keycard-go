package main

// #cgo LDFLAGS: -shared
import "C"
import (
	"fmt"

	statuskeycardgo "github.com/status-im/status-keycard-go"
)

func main() {}

//export HelloWorld
func HelloWorld() *C.char {
	res := statuskeycardgo.HelloWorld()
	return C.CString(fmt.Sprintf("shared lib: %s", res))
}
