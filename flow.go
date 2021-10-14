package main

import "errors"

func NewFlow(storageDir string) (*keycardFlow, error) {
	flow := &keycardFlow{
		wakeUp:  make(chan (struct{})),
		storage: storageDir,
	}

	return flow, nil
}

func (f *keycardFlow) Start(flowType FlowType, params map[string]string) error {
	if f.state != IDLE {
		return errors.New("already running")
	}

	f.flowType = flowType
	f.params = params
	f.state = RUNNING
	go f.runFlow()

	return nil
}

func (f *keycardFlow) Resume(params map[string]string) error {
	if f.state != PAUSED {
		return errors.New("only paused flows can be resumed")
	}

	for k, v := range params {
		f.params[k] = v
	}

	f.state = RESUMING
	f.wakeUp <- struct{}{}

	return nil
}

func (f *keycardFlow) Cancel() error {
	prevState := f.state

	if prevState != IDLE {
		return errors.New("cannot cancel idle flow")
	}

	f.state = CANCELLING
	if prevState == PAUSED {
		f.wakeUp <- struct{}{}
	}

	return nil
}

func (f *keycardFlow) runFlow() {

}
