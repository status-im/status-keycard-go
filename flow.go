package statuskeycardgo

import (
	"errors"

	"github.com/status-im/status-keycard-go/signal"
)

type cardStatus struct {
	instanceUID string
	keyUID      string
	freeSlots   int
	pinRetries  int
	pukRetries  int
}

type KeycardFlow struct {
	flowType FlowType
	state    runState
	wakeUp   chan (struct{})
	pairings *pairingStore
	params   FlowParams
	cardInfo cardStatus
}

func NewFlow(storageDir string) (*KeycardFlow, error) {
	p, err := newPairingStore(storageDir)

	if err != nil {
		return nil, err
	}

	flow := &KeycardFlow{
		wakeUp:   make(chan (struct{})),
		pairings: p,
	}

	return flow, nil
}

func (f *KeycardFlow) Start(flowType FlowType, params FlowParams) error {
	if f.state != Idle {
		return errors.New("already running")
	}

	f.flowType = flowType
	f.params = params
	f.state = Running
	go f.runFlow()

	return nil
}

func (f *KeycardFlow) Resume(params FlowParams) error {
	if f.state != Paused {
		return errors.New("only paused flows can be resumed")
	}

	for k, v := range params {
		f.params[k] = v
	}

	f.state = Resuming
	f.wakeUp <- struct{}{}

	return nil
}

func (f *KeycardFlow) Cancel() error {
	prevState := f.state

	if prevState != Idle {
		return errors.New("cannot cancel idle flow")
	}

	f.state = Cancelling
	if prevState == Paused {
		f.wakeUp <- struct{}{}
	}

	return nil
}

func (f *KeycardFlow) runFlow() {
	var result FlowStatus
	var err error

	for {
		f.cardInfo = cardStatus{freeSlots: -1, pinRetries: -1, pukRetries: -1}
		result, err = f.connectedFlow()

		if _, ok := err.(*restartError); !ok {
			if result == nil {
				result = FlowStatus{ErrorKey: err.Error()}
			}
			break
		}
	}

	if f.state != Cancelling {
		signal.SendEvent(FlowResult, result)
	}

	f.state = Idle
}

func (f *KeycardFlow) pause(action string, errMsg string) {
	status := FlowParams{}

	if errMsg != "" {
		status[ErrorKey] = errMsg
	}

	if f.cardInfo.freeSlots != -1 {
		status[InstanceUID] = f.cardInfo.instanceUID
		status[KeyUID] = f.cardInfo.keyUID
		status[FreeSlots] = f.cardInfo.freeSlots
	}

	if f.cardInfo.pinRetries != -1 {
		status[PINRetries] = f.cardInfo.pinRetries
		status[PUKRetries] = f.cardInfo.pukRetries
	}

	signal.SendEvent(action, status)
	f.state = Paused
}

func (f *KeycardFlow) pauseAndWait(action string, errMsg string) error {
	if f.state == Cancelling {
		return errors.New("cancel")
	}

	f.pause(action, errMsg)
	<-f.wakeUp

	if f.state == Resuming {
		f.state = Running
		return nil
	} else {
		return errors.New("cancel")
	}
}

func (f *KeycardFlow) pauseAndRestart(action string, errMsg string) error {
	err := f.pauseAndWait(action, errMsg)

	if err != nil {
		return err
	}

	return restartErr()
}

func (f *KeycardFlow) requireKeys() error {
	if f.cardInfo.keyUID != "" {
		return nil
	}

	return f.pauseAndRestart(SwapCard, ErrorNoKeys)
}

func (f *KeycardFlow) closeKeycard(kc *keycardContext) {
	if kc != nil {
		kc.stop()
	}
}

func (f *KeycardFlow) connect() *keycardContext {
	kc, err := startKeycardContext()

	if err != nil {
		return nil
	}

	f.pause(InsertCard, "")
	select {
	case <-f.wakeUp:
		if f.state != Cancelling {
			panic("Resuming is not expected during connection")
		}
		return nil
	case <-kc.connected:
		if kc.runErr != nil {
			return nil
		}

		signal.SendEvent(CardInserted, FlowStatus{})
		return kc
	}
}

func (f *KeycardFlow) connectedFlow() (FlowStatus, error) {
	kc := f.connect()
	defer f.closeKeycard(kc)

	if kc == nil {
		return nil, errors.New(ErrorConnection)
	}

	if factoryReset, ok := f.params[FactoryReset]; ok && factoryReset.(bool) {
		err := f.factoryReset(kc)

		if err != nil {
			return nil, err
		}
	}

	err := f.selectKeycard(kc)

	if err != nil {
		return nil, err
	}

	switch f.flowType {
	case GetAppInfo:
		return f.getAppInfoFlow(kc)
	case RecoverAccount:
		return f.exportKeysFlow(kc, true)
	case Login:
		return f.exportKeysFlow(kc, false)
	case UnpairThis:
		return f.unpairThisFlow(kc)
	default:
		return nil, errors.New(ErrorUnknownFlow)
	}
}

func (f *KeycardFlow) getAppInfoFlow(kc *keycardContext) (FlowStatus, error) {
	return FlowStatus{ErrorKey: ErrorOK, AppInfo: toAppInfo(kc.cmdSet.ApplicationInfo)}, nil
}

func (f *KeycardFlow) exportKeysFlow(kc *keycardContext, recover bool) (FlowStatus, error) {
	err := f.requireKeys()

	if err != nil {
		return nil, err
	}

	err = f.openSCAndAuthenticate(kc)

	if err != nil {
		return nil, err
	}

	result := FlowStatus{KeyUID: f.cardInfo.keyUID}

	key, err := f.exportKey(kc, encryptionPath, false)
	if err != nil {
		return nil, err
	}
	result[EncKey] = key

	key, err = f.exportKey(kc, whisperPath, false)
	if err != nil {
		return nil, err
	}
	result[WhisperKey] = key

	if recover {
		key, err = f.exportKey(kc, eip1581Path, true)
		if err != nil {
			return nil, err
		}
		result[EIP1581Key] = key

		key, err = f.exportKey(kc, walletRoothPath, true)
		if err != nil {
			return nil, err
		}
		result[WalleRootKey] = key

		key, err = f.exportKey(kc, walletPath, true)
		if err != nil {
			return nil, err
		}
		result[WalletKey] = key

		key, err = f.exportKey(kc, masterPath, true)
		if err != nil {
			return nil, err
		}
		result[MasterKey] = key
	}

	return result, nil
}

func (f *KeycardFlow) unpairThisFlow(kc *keycardContext) (FlowStatus, error) {
	err := f.openSCAndAuthenticate(kc)

	if err != nil {
		return nil, err
	}

	err = f.unpairCurrent(kc)

	if err != nil {
		return nil, err
	}

	f.cardInfo.freeSlots++
	return FlowStatus{InstanceUID: f.cardInfo.instanceUID, FreeSlots: f.cardInfo.freeSlots}, err
}
