package statuskeycardgo

import (
	"github.com/status-im/status-keycard-go/signal"
)

func (mkf *MockedKeycardFlow) handleGetAppInfoFlow() {
	flowStatus := FlowStatus{}

	if mkf.insertedKeycard.NotStatusKeycard {
		flowStatus[ErrorKey] = ErrorNotAKeycard
		flowStatus[InstanceUID] = ""
		flowStatus[KeyUID] = ""
		flowStatus[FreeSlots] = 0
		mkf.state = Paused
		signal.Send(SwapCard, flowStatus)
		return
	}

	flowStatus = FlowStatus{
		PINRetries: mkf.insertedKeycard.PinRetries,
		PUKRetries: mkf.insertedKeycard.PukRetries,
	}

	if mkf.insertedKeycard.InstanceUID == "" || mkf.insertedKeycard.KeyUID == "" {
		flowStatus[ErrorKey] = ErrorNoKeys
		flowStatus[FreeSlots] = 0
		mkf.state = Paused
		signal.Send(SwapCard, flowStatus)
		return
	}

	var (
		enteredPIN   string
		factoryReset bool
	)

	if v, ok := mkf.params[PIN]; ok {
		enteredPIN = v.(string)
	}

	if v, ok := mkf.params[FactoryReset]; ok {
		factoryReset = v.(bool)
	}

	if factoryReset {
		mkf.state = Idle
		*mkf.insertedKeycard = MockedKeycard{}
		signal.Send(FlowResult, FlowStatus{
			ErrorKey: ErrorOK,
			Paired:   false,
			AppInfo: ApplicationInfo{
				Initialized:    false,
				InstanceUID:    []byte(""),
				Version:        0,
				AvailableSlots: 0,
				KeyUID:         []byte(""),
			},
		})
		return
	}

	keycardStoresKeys := mkf.insertedKeycard.InstanceUID != "" && mkf.insertedKeycard.KeyUID != ""
	if len(enteredPIN) == defPINLen && enteredPIN == mkf.insertedKeycard.Pin || !keycardStoresKeys {
		flowStatus[ErrorKey] = ErrorOK
		flowStatus[Paired] = keycardStoresKeys
		flowStatus[AppInfo] = ApplicationInfo{
			Initialized:    keycardStoresKeys,
			InstanceUID:    hexString(mkf.insertedKeycard.InstanceUID),
			Version:        123,
			AvailableSlots: mkf.insertedKeycard.FreePairingSlots,
			KeyUID:         hexString(mkf.insertedKeycard.KeyUID),
		}
		mkf.state = Idle
		signal.Send(FlowResult, flowStatus)
		return
	}

	flowStatus[FreeSlots] = mkf.insertedKeycard.FreePairingSlots
	flowStatus[InstanceUID] = mkf.insertedKeycard.InstanceUID
	flowStatus[KeyUID] = mkf.insertedKeycard.KeyUID
	mkf.state = Paused
	signal.Send(EnterPIN, flowStatus)
}
