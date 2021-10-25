package signal

/*
#include <stddef.h>
#include <stdbool.h>
#include <stdlib.h>
extern bool KeycardServiceSignalEvent(const char *jsonEvent);
extern void KeycardSetEventCallback(void *cb);
*/
import "C"
import (
	"encoding/json"
	"unsafe"

	"github.com/ethereum/go-ethereum/log"
)

// KeycardSignalHandler is a simple callback function that gets called when any signal is received
type KeycardSignalHandler func([]byte)

// storing the current signal handler here
var keycardSignalHandler KeycardSignalHandler

// All general log messages in this package should be routed through this logger.
var logger = log.New("package", "keycard-go/signal")

// Envelope is a general signal sent upward from node to RN app
type Envelope struct {
	Type  string      `json:"type"`
	Event interface{} `json:"event"`
}

// NewEnvelope creates new envlope of given type and event payload.
func NewEnvelope(typ string, event interface{}) *Envelope {
	return &Envelope{
		Type:  typ,
		Event: event,
	}
}

// send sends application signal (in JSON) upwards to application (via default notification handler)
func Send(typ string, event interface{}) {
	signal := NewEnvelope(typ, event)
	data, err := json.Marshal(&signal)
	if err != nil {
		logger.Error("Marshalling signal envelope", "error", err)
		return
	}
	// If a Go implementation of signal handler is set, let's use it.
	if keycardSignalHandler != nil {
		keycardSignalHandler(data)
	} else {
		// ...and fallback to C implementation otherwise.
		str := C.CString(string(data))
		C.KeycardServiceSignalEvent(str)
		C.free(unsafe.Pointer(str))
	}
}

// SetKeycardSignalHandler sets new handler for geth events
// this function uses pure go implementation
func SetKeycardSignalHandler(handler KeycardSignalHandler) {
	keycardSignalHandler = handler
}

// KeycardSetSignalEventCallback set callback
// this function uses C implementation (see `signals.c` file)
func KeycardSetSignalEventCallback(cb unsafe.Pointer) {
	C.KeycardSetEventCallback(cb)
}
