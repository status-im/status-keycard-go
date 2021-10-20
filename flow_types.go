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
	EnterPairing = "keycard.action.enter-pairing"
	EnterPIN     = "keycard.action.enter-pin"
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
	AppInfo      = "application-info"
	InstanceUID  = "instance-uid"
	FactoryReset = "factory reset"
	KeyUID       = "key-uid"
	FreeSlots    = "free-pairing-slots"
	PINRetries   = "pin-retries"
	PUKRetries   = "puk-retries"
	PairingPass  = "pairing-pass"
	PIN          = "pin"
)

const (
	maxPINRetries = 3
	maxPUKRetries = 5
	maxFreeSlots  = 5
)
