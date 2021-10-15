package statuskeycardgo

type FlowType int
type runState int
type stepState int

const (
	GetAppInfo FlowType = iota
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
	SwapCard     = "keycard.action.swap-card"
)

const (
	ErrorKey         = "error"
	ErrorOK          = "ok"
	ErrorCancel      = "cancel"
	ErrorConnection  = "connection-error"
	ErrorUnknownFlow = "unknown-flow"
)

const (
	AppInfo     = "application-info"
	InstanceUID = "instance-uid"
	KeyUID      = "key-uid"
)

type keycardFlow struct {
	flowType FlowType
	state    runState
	wakeUp   chan (struct{})
	storage  string
	params   map[string]interface{}
}
