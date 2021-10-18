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

func (f *keycardFlow) Start(flowType FlowType, params FlowParams) error {
	if f.state != Idle {
		return errors.New("already running")
	}

	f.flowType = flowType
	f.params = params
	f.state = Running
	f.active = make(chan (struct{}))
	go f.runFlow()

	return nil
}

func (f *keycardFlow) Resume(params FlowParams) error {
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

func (f *keycardFlow) Wait() {
	<-f.active
}

func (f *keycardFlow) runFlow() {
	repeat := true
	var result FlowStatus

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
	close(f.active)
}

func (f *keycardFlow) switchFlow() (bool, FlowStatus) {
	kc := f.connect()
	defer f.closeKeycard(kc)

	if kc == nil {
		return false, FlowStatus{ErrorKey: ErrorConnection}
	}

	switch f.flowType {
	case GetAppInfo:
		return f.getAppInfoFlow(kc)
	default:
		return false, FlowStatus{ErrorKey: ErrorUnknownFlow}
	}
}

func (f *keycardFlow) pause(action string, status FlowStatus) {
	signal.SendEvent(action, status)
	f.state = Paused
}

func (f *keycardFlow) pauseAndWait(action string, status FlowStatus) bool {
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
			if f.pauseAndWait(SwapCard, FlowStatus{ErrorKey: InstanceUID, InstanceUID: instanceUID}) {
				return true, nil
			} else {
				return false, errors.New(ErrorCancel)
			}
		}
	}

	if requiredKeyUID, ok := f.params[KeyUID]; ok {
		if keyUID := tox(appInfo.KeyUID); keyUID != requiredKeyUID {
			if f.pauseAndWait(SwapCard, FlowStatus{ErrorKey: KeyUID, KeyUID: keyUID}) {
				return true, nil
			} else {
				return false, errors.New(ErrorCancel)
			}
		}
	}

	return false, nil
}

func (f *keycardFlow) getAppInfoFlow(kc *keycardContext) (bool, FlowStatus) {
	restart, err := f.selectKeycard(kc)

	if err != nil {
		return false, FlowStatus{ErrorKey: err.Error()}
	} else if restart {
		return true, nil
	}

	return false, FlowStatus{ErrorKey: ErrorOK, AppInfo: kc.cmdSet.ApplicationInfo}
}
