package signal

import "fmt"

const (
	EventKeycardConnected = "keycard.connected"
)

func SendKeycardConnected(event interface{}) {
	send(EventKeycardConnected, event)
}

func SendEvent(typ string, event interface{}) {
	fmt.Printf("sending event: %+v\n", typ)
	send(typ, event)
}
