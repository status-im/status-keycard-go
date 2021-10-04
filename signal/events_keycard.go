package signal

const (
	EventKeycardConnected = "keycard.connected"
)

func SendKeycardConnected(event interface{}) {
	send(EventKeycardConnected, event)
}
