package main

type FlowType int
type RunState int

const (
	GET_STATUS FlowType = iota
	RECOVER_ACCOUNT
	LOAD_ACCOUNT
	LOGIN
	SIGN
	CHANGE_CREDENTIALS
	UNPAIR
	UNPAIR_OTHERS
	DELETE_ACCOUNT_AND_UNPAIR
)

const (
	IDLE RunState = iota
	RUNNING
	PAUSED
	RESUMING
	CANCELLING
)

type keycardFlow struct {
	flowType FlowType
	state    RunState
	wakeUp   chan (struct{})
	storage  string
	params   map[string]string
}
