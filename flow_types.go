package statuskeycardgo

type FlowType int
type runState int

const (
	GetStatus FlowType = iota
	RecoverAccount
	LoadAccount
	Login
	Sign
	ChangeCredentials
	UnpairThis
	UnpairOthers
	DeleteAccountAndUnpair
)

const (
	Idle runState = iota
	Running
	Paused
	Resuming
	Cancelling
)

const (
	FlowResult   = "keycard.flow-result"
	InsertCard   = "keycard.action.insert-card"
	CardInserted = "keycard.action.card-inserted"
)

type keycardFlow struct {
	flowType FlowType
	state    runState
	wakeUp   chan (struct{})
	storage  string
	params   map[string]string
}
