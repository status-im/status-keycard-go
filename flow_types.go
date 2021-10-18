package statuskeycardgo

type FlowType int
type FlowParams map[string]interface{}
type FlowStatus map[string]interface{}
type runState int

type restartError struct{}

func restartErr() (e *restartError) {
	return &restartError{}
}

func (e *restartError) Error() string {
	return "restart"
}

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
	ErrorNotAKeycard = "not-a-keycard"
)

const (
	AppInfo     = "application-info"
	InstanceUID = "instance-uid"
	KeyUID      = "key-uid"
)
