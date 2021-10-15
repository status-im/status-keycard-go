package statuskeycardgo

import (
	"errors"

	"github.com/status-im/status-keycard-go/signal"
)

func NewFlow(storageDir string) (*keycardFlow, error) {
	flow := &keycardFlow{
		wakeUp:  make(chan (struct{})),
		storage: storageDir,
	}

	return flow, nil
}

func (f *keycardFlow) Start(flowType FlowType, params map[string]interface{}) error {
	if f.state != Idle {
		return errors.New("already running")
	}

	f.flowType = flowType
	f.params = params
	f.state = Running
	go f.runFlow()

	return nil
}

func (f *keycardFlow) Resume(params map[string]interface{}) error {
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

func (f *keycardFlow) Cancel() error {
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

func (f *keycardFlow) runFlow() {
	repeat := true
	var result map[string]interface{}

	for repeat {
		switch f.flowType {
		case GetAppInfo:
			repeat, result = f.switchFlow()
		}
	}

	if f.state != Cancelling {
		signal.SendEvent(FlowResult, result)
	}

	f.state = Idle
}

func (f *keycardFlow) switchFlow() (bool, map[string]interface{}) {
	kc := f.connect()
	defer f.closeKeycard(kc)

	if kc == nil {
		return false, map[string]interface{}{ErrorKey: ErrorConnection}
	}

	switch f.flowType {
	case GetAppInfo:
		return f.getAppInfoFlow(kc)
	default:
		return false, map[string]interface{}{ErrorKey: ErrorUnknownFlow}
	}
}

func (f *keycardFlow) pause(action string, status map[string]interface{}) {
	signal.SendEvent(action, status)
	f.state = Paused
}

func (f *keycardFlow) pauseAndWait(action string, status map[string]interface{}) bool {
	f.pause(action, status)
	<-f.wakeUp

	if f.state == Resuming {
		f.state = Running
		return true
	} else {
		return false
	}
}

func (f *keycardFlow) closeKeycard(kc *keycardContext) {
	if kc != nil {
		kc.stop()
	}
}

func (f *keycardFlow) connect() *keycardContext {
	kc, err := startKeycardContext()

	if err != nil {
		return nil
	}

	f.pause(InsertCard, map[string]interface{}{})
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

		return kc
	}
}

func (f *keycardFlow) selectKeycard(kc *keycardContext) (bool, error) {
	appInfo, err := kc.selectApplet()

	if err != nil {
		return false, err
	}

	if requiredInstanceUID, ok := f.params[InstanceUID]; ok {
		if instanceUID := tox(appInfo.InstanceUID); instanceUID != requiredInstanceUID {
			if f.pauseAndWait(SwapCard, map[string]interface{}{ErrorKey: InstanceUID, InstanceUID: instanceUID}) {
				return true, nil
			} else {
				return false, errors.New(ErrorCancel)
			}
		}
	}

	if requiredKeyUID, ok := f.params[KeyUID]; ok {
		if keyUID := tox(appInfo.KeyUID); keyUID != requiredKeyUID {
			if f.pauseAndWait(SwapCard, map[string]interface{}{ErrorKey: KeyUID, KeyUID: keyUID}) {
				return true, nil
			} else {
				return false, errors.New(ErrorCancel)
			}
		}
	}

	return false, nil
}

func (f *keycardFlow) getAppInfoFlow(kc *keycardContext) (bool, map[string]interface{}) {
	restart, err := f.selectKeycard(kc)

	if err != nil {
		return false, map[string]interface{}{ErrorKey: err.Error()}
	} else if restart {
		return true, nil
	}

	return false, map[string]interface{}{ErrorKey: ErrorOK, AppInfo: kc.cmdSet.ApplicationInfo}
}
