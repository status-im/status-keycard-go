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

	err := f.pauseAndWait(EnterPairing, ErrorPairing)

	if err != nil {
		return err
	}

	return f.pair(kc)
}

func (f *KeycardFlow) initCard(kc *keycardContext) error {
	newPIN, pinOK := f.params[NewPIN]

	if !pinOK {
		err := f.pauseAndWait(EnterNewPIN, ErrorRequireInit)
		if err != nil {
			return err
		}

		return f.initCard(kc)
	}

	newPUK, pukOK := f.params[NewPUK]
	if !pukOK {
		err := f.pauseAndWait(EnterNewPUK, ErrorRequireInit)
		if err != nil {
			return err
		}

		return f.initCard(kc)
	}

	newPairing, pairingOK := f.params[NewPairing]
	if !pairingOK {
		err := f.pauseAndWait(EnterNewPair, ErrorRequireInit)
		if err != nil {
			return err
		}

		return f.initCard(kc)
	}

	err := kc.init(newPIN.(string), newPUK.(string), newPairing.(string))

	if err == nil {
		f.params[PIN] = newPIN
		f.params[PairingPass] = newPairing
		delete(f.params, NewPIN)
		delete(f.params, NewPUK)
		delete(f.params, NewPairing)
	}

	return err
}

func (f *KeycardFlow) openSC(kc *keycardContext, giveup bool) error {
	var pairing *PairingInfo

	if !kc.cmdSet.ApplicationInfo.Initialized && !giveup {
		return f.initCard(kc)
	} else {
		pairing = f.pairings.get(f.cardInfo.instanceUID)
	}

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

	if giveup {
		return giveupErr()
	}

	err := f.pair(kc)

	if err != nil {
		return err
	}

	return f.openSC(kc, giveup)
}

func (f *KeycardFlow) unblockPIN(kc *keycardContext) error {
	pukError := ""
	var err error

	newPIN, pinOK := f.params[NewPIN]
	puk, pukOK := f.params[PUK]

	if pinOK && pukOK {
		err = kc.unblockPIN(puk.(string), newPIN.(string))

		if err == nil {
			f.cardInfo.pinRetries = maxPINRetries
			f.cardInfo.pukRetries = maxPUKRetries
			f.params[PIN] = newPIN
			delete(f.params, NewPIN)
			delete(f.params, PUK)
			return nil
		} else if isSCardError(err) {
			return restartErr()
		} else if leftRetries, ok := getRetries(err); ok {
			f.cardInfo.pukRetries = leftRetries
			delete(f.params, PUK)
			pukOK = false
		}

		pukError = PUK
	}

	if !pukOK {
		err = f.pauseAndWait(EnterPUK, pukError)
	} else if !pinOK {
		err = f.pauseAndWait(EnterNewPIN, ErrorUnblocking)
	}

	if err != nil {
		return err
	}

	return f.unblockPIN(kc)
}

func (f *KeycardFlow) authenticate(kc *keycardContext) error {
	if f.cardInfo.pukRetries == 0 {
		return f.pauseAndRestart(SwapCard, PUKRetries)
	} else if f.cardInfo.pinRetries == 0 {
		// succesful unblock leaves the card authenticated
		return f.unblockPIN(kc)
	}

	pinError := ""

	if pin, ok := f.params[PIN]; ok {
		err := kc.verifyPin(pin.(string))

		if err == nil {
			f.cardInfo.pinRetries = maxPINRetries
			return nil
		} else if isSCardError(err) {
			return restartErr()
		} else if leftRetries, ok := getRetries(err); ok {
			f.cardInfo.pinRetries = leftRetries
			delete(f.params, PIN)
		}

		pinError = PIN
	}

	err := f.pauseAndWait(EnterPIN, pinError)

	if err != nil {
		return err
	}

	return f.authenticate(kc)
}

func (f *KeycardFlow) openSCAndAuthenticate(kc *keycardContext, giveup bool) error {
	err := f.openSC(kc, giveup)

	if err != nil {
		return err
	}

	return f.authenticate(kc)
}

func (f *KeycardFlow) unpairCurrent(kc *keycardContext) error {
	err := kc.unpairCurrent()

	if isSCardError(err) {
		return restartErr()
	}

	return err
}

func (f *KeycardFlow) unpair(kc *keycardContext, idx int) error {
	err := kc.unpair(uint8(idx))

	if isSCardError(err) {
		return restartErr()
	}

	return err
}

func (f *KeycardFlow) removeKey(kc *keycardContext) error {
	err := kc.removeKey()

	if isSCardError(err) {
		return restartErr()
	}

	return err
}

func (f *KeycardFlow) exportKey(kc *keycardContext, path string, onlyPublic bool) (*KeyPair, error) {
	keyPair, err := kc.exportKey(true, false, onlyPublic, path)

	if isSCardError(err) {
		return nil, restartErr()
	}

	return keyPair, err
}
