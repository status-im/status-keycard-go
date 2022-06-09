package statuskeycardgo

type FlowType int
type FlowParams map[string]interface{}
type FlowStatus map[string]interface{}
type runState int

type restartError struct{}
type giveupError struct{}

func restartErr() (e *restartError) {
	return &restartError{}
}

func (e *restartError) Error() string {
	return "restart"
}

func giveupErr() (e *giveupError) {
	return &giveupError{}
}

func (e *giveupError) Error() string {
	return "giveup"
}

const (
	GetAppInfo FlowType = iota
	RecoverAccount
	LoadAccount
	Login
	ExportPublic
	Sign
	ChangePIN
	ChangePUK
	ChangePairing
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
	FlowResult    = "keycard.flow-result"
	InsertCard    = "keycard.action.insert-card"
	CardInserted  = "keycard.action.card-inserted"
	SwapCard      = "keycard.action.swap-card"
	EnterPairing  = "keycard.action.enter-pairing"
	EnterPIN      = "keycard.action.enter-pin"
	EnterPUK      = "keycard.action.enter-puk"
	EnterNewPair  = "keycard.action.enter-new-pairing"
	EnterNewPIN   = "keycard.action.enter-new-pin"
	EnterNewPUK   = "keycard.action.enter-new-puk"
	EnterTXHash   = "keycard.action.enter-tx-hash"
	EnterPath     = "keycard.action.enter-bip44-path"
	EnterMnemonic = "keycard.action.enter-mnemonic"
)

const (
	ErrorKey         = "error"
	ErrorOK          = "ok"
	ErrorCancel      = "cancel"
	ErrorConnection  = "connection-error"
	ErrorUnknownFlow = "unknown-flow"
	ErrorNotAKeycard = "not-a-keycard"
	ErrorNoKeys      = "no-keys"
	ErrorHasKeys     = "has-keys"
	ErrorRequireInit = "require-init"
	ErrorPairing     = "pairing"
	ErrorUnblocking  = "unblocking"
	ErrorSigning     = "signing"
	ErrorExporting   = "exporting"
	ErrorChanging    = "changing-credentials"
	ErrorLoading     = "loading-keys"
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
	Paired       = "paired"
	NewPairing   = "new-pairing-pass"
	DefPairing   = "KeycardDefaultPairing"
	PIN          = "pin"
	NewPIN       = "new-pin"
	PUK          = "puk"
	NewPUK       = "new-puk"
	MasterKey    = "master-key"
	WalleRootKey = "wallet-root-key"
	WalletKey    = "wallet-key"
	EIP1581Key   = "eip1581-key"
	WhisperKey   = "whisper-key"
	EncKey       = "encryption-key"
	ExportedKey  = "exported-key"
	Mnemonic     = "mnemonic"
	MnemonicLen  = "mnemonic-length"
	MnemonicIdxs = "mnemonic-indexes"
	TXHash       = "tx-hash"
	BIP44Path    = "bip44-path"
	TXSignature  = "tx-signature"
	Overwrite    = "overwrite"
)

const (
	maxPINRetries = 3
	maxPUKRetries = 5
	maxFreeSlots  = 5
	defMnemoLen   = 12
)

const (
	masterPath      = "m"
	walletRoothPath = "m/44'/60'/0'/0"
	walletPath      = walletRoothPath + "/0"
	eip1581Path     = "m/43'/60'/1581'"
	whisperPath     = eip1581Path + "/0'/0"
	encryptionPath  = eip1581Path + "/1'/0"
)
