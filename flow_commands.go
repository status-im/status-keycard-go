package statuskeycardgo

import "errors"

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
