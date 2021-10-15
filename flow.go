package main

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

func (f *keycardFlow) Start(flowType FlowType, params map[string]string) error {
	if f.state != Idle {
		return errors.New("already running")
	}

	f.flowType = flowType
	f.params = params
	f.state = Running
	go f.runFlow()

	return nil
}

func (f *keycardFlow) Resume(params map[string]string) error {
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
	var result map[string]string

	for repeat {
		switch f.flowType {
		case GetStatus:
			repeat, result = f.switchFlow()
		}
	}

	if f.state != Cancelling {
		signal.SendEvent(FlowResult, result)
	}

	f.state = Idle
}

func (f *keycardFlow) switchFlow() (bool, map[string]string) {
	kc := f.connect()
	defer f.closeKeycard(kc)

	if kc == nil {
		return false, map[string]string{"Error": "Couldn't connect to the card"}
	}

	switch f.flowType {
	case GetStatus:
		return f.getStatusFlow(kc)
	default:
		return false, map[string]string{"Error": "Unknown flow"}
	}
}

func (f *keycardFlow) pause(action string, status map[string]string) {
	signal.SendEvent(action, status)
	f.state = Paused
}

func (f *keycardFlow) pauseAndWait(action string, status map[string]string) {
	f.pause(action, status)
	<-f.wakeUp
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

	f.pause(InsertCard, map[string]string{})
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

func (f *keycardFlow) getStatusFlow(kc *keycardContext) (bool, map[string]string) {

	return false, nil
}
