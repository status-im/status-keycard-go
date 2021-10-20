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
		return f.recoverAccountFlow(kc)
	default:
		return nil, errors.New(ErrorUnknownFlow)
	}
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

func (f *KeycardFlow) factoryReset(kc *keycardContext) error {
	// on success, remove the FactoryReset switch to avoid re-triggering it
	// if card is disconnected/reconnected
	delete(f.params, FactoryReset)
	return errors.New("not implemented")
}

func (f *KeycardFlow) selectKeycard(kc *keycardContext) error {
	appInfo, err := kc.selectApplet()

	f.cardInfo.instanceUID = tox(appInfo.InstanceUID)
	f.cardInfo.keyUID = tox(appInfo.KeyUID)
	f.cardInfo.freeSlots = bytesToInt(appInfo.AvailableSlots)

	if err != nil {
		return restartErr()
	}

	if !appInfo.Installed {
		return f.pauseAndRestart(SwapCard, ErrorNotAKeycard)
	}

	if requiredInstanceUID, ok := f.params[InstanceUID]; ok {
		if f.cardInfo.instanceUID != requiredInstanceUID {
			return f.pauseAndRestart(SwapCard, InstanceUID)
		}
	}

	if requiredKeyUID, ok := f.params[KeyUID]; ok {
		if f.cardInfo.keyUID != requiredKeyUID {
			return f.pauseAndRestart(SwapCard, KeyUID)
		}
	}

	return nil
}

func (f *KeycardFlow) pair(kc *keycardContext) error {
	if f.cardInfo.freeSlots == 0 {
		return f.pauseAndRestart(SwapCard, FreeSlots)
	}

	if pairingPass, ok := f.params[PairingPass]; ok {
		pairing, err := kc.pair(pairingPass.(string))

		if err == nil {
			return f.pairings.store(f.cardInfo.instanceUID, toPairInfo(pairing))
		} else if isSCardError(err) {
			return restartErr()
		}

		delete(f.params, PairingPass)
	}

	err := f.pauseAndWait(EnterPairing, "")

	if err != nil {
		return err
	}

	return f.pair(kc)
}

func (f *KeycardFlow) initCard(kc *keycardContext) error {
	//NOTE: after init a restart of the flow is always needed
	return errors.New("not implemented")
}

func (f *KeycardFlow) openSC(kc *keycardContext) error {
	if !kc.cmdSet.ApplicationInfo.Initialized {
		return f.initCard(kc)
	}

	pairing := f.pairings.get(f.cardInfo.instanceUID)

	if pairing != nil {
		err := kc.openSecureChannel(pairing.Index, pairing.Key)

		if err == nil {
			appStatus, err := kc.getStatusApplication()

			if err != nil {
				// getStatus can only fail for connection errors
				return restartErr()
			}

			f.cardInfo.pinRetries = appStatus.PinRetryCount
			f.cardInfo.pukRetries = appStatus.PUKRetryCount

			return nil
		} else if isSCardError(err) {
			return restartErr()
		}

		f.pairings.delete(f.cardInfo.instanceUID)
	}

	err := f.pair(kc)

	if err != nil {
		return err
	}

	return f.openSC(kc)
}

func (f *KeycardFlow) unblockPUK(kc *keycardContext) error {
	return errors.New("not yet implemented")
}

func (f *KeycardFlow) authenticate(kc *keycardContext) error {
	if f.cardInfo.pukRetries == 0 {
		return f.pauseAndRestart(SwapCard, PUKRetries)
	} else if f.cardInfo.pinRetries == 0 {
		// succesful PUK unblock leaves the card authenticated
		return f.unblockPUK(kc)
	}

	pinError := ""

	if pin, ok := f.params[PIN]; ok {
		err := kc.verifyPin(pin.(string))

		if err == nil {
			f.cardInfo.pinRetries = maxPINRetries
			return nil
		} else if isSCardError(err) {
			return restartErr()
		} else if leftRetries, ok := getPinRetries(err); ok {
			f.cardInfo.pinRetries = leftRetries
		}

		pinError = PIN
	}

	err := f.pauseAndWait(EnterPIN, pinError)

	if err != nil {
		return err
	}

	return f.authenticate(kc)
}

func (f *KeycardFlow) openSCAndAuthenticate(kc *keycardContext) error {
	err := f.openSC(kc)

	if err != nil {
		return err
	}

	return f.authenticate(kc)
}

func (f *KeycardFlow) getAppInfoFlow(kc *keycardContext) (FlowStatus, error) {
	return FlowStatus{ErrorKey: ErrorOK, AppInfo: toAppInfo(kc.cmdSet.ApplicationInfo)}, nil
}

func (f *KeycardFlow) recoverAccountFlow(kc *keycardContext) (FlowStatus, error) {
	err := f.openSCAndAuthenticate(kc)

	if err != nil {
		return nil, err
	}

	return nil, errors.New("not yet implemented")
}
