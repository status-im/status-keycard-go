package signal

const (
	EventKeycardConnected = "keycard.connected"
)

func SendKeycardConnected(event interface{}) {
	send(EventKeycardConnected, event)
}

func SendEvent(typ string, event interface{}) {
	send(typ, event)
}
