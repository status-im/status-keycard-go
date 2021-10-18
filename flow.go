package statuskeycardgo

import (
	"errors"

	"github.com/status-im/status-keycard-go/signal"
)

type KeycardFlow struct {
	flowType FlowType
	state    runState
	wakeUp   chan (struct{})
	storage  string
	params   map[string]interface{}
}

func NewFlow(storageDir string) (*KeycardFlow, error) {
	flow := &KeycardFlow{
		wakeUp:  make(chan (struct{})),
		storage: storageDir,
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

func (f *KeycardFlow) pause(action string, status FlowStatus) {
	signal.SendEvent(action, status)
	f.state = Paused
}

func (f *KeycardFlow) pauseAndWait(action string, status FlowStatus) error {
	f.pause(action, status)
	<-f.wakeUp

	if f.state == Resuming {
		f.state = Running
		return nil
	} else {
		return errors.New("cancel")
	}
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

	f.pause(InsertCard, FlowStatus{})
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

func restartOrCancel(cancel error) error {
	if cancel == nil {
		return restartErr()
	} else {
		return cancel
	}
}

func (f *KeycardFlow) selectKeycard(kc *keycardContext) error {
	appInfo, err := kc.selectApplet()

	if err != nil {
		return err
	}

	if !appInfo.Installed {
		return restartOrCancel(f.pauseAndWait(SwapCard, FlowStatus{ErrorKey: ErrorNotAKeycard}))
	}

	if requiredInstanceUID, ok := f.params[InstanceUID]; ok {
		if instanceUID := tox(appInfo.InstanceUID); instanceUID != requiredInstanceUID {
			return restartOrCancel(f.pauseAndWait(SwapCard, FlowStatus{ErrorKey: InstanceUID, InstanceUID: instanceUID}))
		}
	}

	if requiredKeyUID, ok := f.params[KeyUID]; ok {
		if keyUID := tox(appInfo.KeyUID); keyUID != requiredKeyUID {
			return restartOrCancel(f.pauseAndWait(SwapCard, FlowStatus{ErrorKey: KeyUID, KeyUID: keyUID}))
		}
	}

	return nil
}

func (f *KeycardFlow) openSCAndAuthenticate(kc *keycardContext) error {
	return errors.New("not yet implemented")
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
