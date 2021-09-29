package signal

const (
	// EventSignRequestAdded is triggered when send transaction request is queued
	EventKeycardConnected = "keycard.connected"
)

func SendKeycardConnected(event interface{}) {
	send(EventKeycardConnected, event)
}
