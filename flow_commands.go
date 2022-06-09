package statuskeycardgo

func (f *KeycardFlow) factoryReset(kc *keycardContext) error {
	err := kc.factoryReset()

	if err == nil {
		delete(f.params, FactoryReset)
		return restartErr()
	} else {
		return err
	}
}

func (f *KeycardFlow) selectKeycard(kc *keycardContext) error {
	appInfo, err := kc.selectApplet()

	f.cardInfo.instanceUID = btox(appInfo.InstanceUID)
	f.cardInfo.keyUID = btox(appInfo.KeyUID)
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

	pairingPass, ok := f.params[PairingPass]

	if !ok {
		pairingPass = DefPairing
	}

	pairing, err := kc.pair(pairingPass.(string))

	if err == nil {
		return f.pairings.store(f.cardInfo.instanceUID, toPairInfo(pairing))
	} else if isSCardError(err) {
		return restartErr()
	}

	delete(f.params, PairingPass)

	err = f.pauseAndWait(EnterPairing, ErrorPairing)

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
		newPairing = DefPairing
	}

	err := kc.init(newPIN.(string), newPUK.(string), newPairing.(string))

	if err != nil {
		return err
	}

	f.params[PIN] = newPIN
	f.params[PairingPass] = newPairing
	delete(f.params, NewPIN)
	delete(f.params, NewPUK)
	delete(f.params, NewPairing)

	return restartErr()
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

func (f *KeycardFlow) exportBIP44Key(kc *keycardContext) (*KeyPair, error) {
	if path, ok := f.params[BIP44Path]; ok {
		return f.exportKey(kc, path.(string), true)
	}

	err := f.pauseAndWait(EnterPath, ErrorExporting)

	if err != nil {
		return nil, err
	}

	return f.exportBIP44Key(kc)
}

func (f *KeycardFlow) loadKeys(kc *keycardContext) error {
	if mnemonic, ok := f.params[Mnemonic]; ok {
		keyUID, err := kc.loadMnemonic(mnemonic.(string), "")

		if isSCardError(err) {
			return restartErr()
		} else if err != nil {
			return err
		}

		f.cardInfo.keyUID = btox(keyUID)
		return nil
	}

	mnemonicLength, ok := f.params[MnemonicLen]
	if !ok {
		mnemonicLength = defMnemoLen
	}

	indexes, err := kc.generateMnemonic(mnemonicLength.(int) / 3)

	if isSCardError(err) {
		return restartErr()
	} else if err != nil {
		return err
	}

	err = f.pauseAndWaitWithStatus(EnterMnemonic, ErrorLoading, FlowParams{MnemonicIdxs: indexes})

	if err != nil {
		return err
	}

	return f.loadKeys(kc)
}

func (f *KeycardFlow) changePIN(kc *keycardContext) error {
	if newPIN, ok := f.params[NewPIN]; ok {
		err := kc.changePin(newPIN.(string))

		if isSCardError(err) {
			return restartErr()
		} else if err != nil {
			return err
		}

		return nil
	}

	err := f.pauseAndWait(EnterNewPIN, ErrorChanging)

	if err != nil {
		return err
	}

	return f.changePIN(kc)
}

func (f *KeycardFlow) changePUK(kc *keycardContext) error {
	if newPUK, ok := f.params[NewPUK]; ok {
		err := kc.changePuk(newPUK.(string))

		if isSCardError(err) {
			return restartErr()
		} else if err != nil {
			return err
		}

		return nil
	}

	err := f.pauseAndWait(EnterNewPUK, ErrorChanging)

	if err != nil {
		return err
	}

	return f.changePUK(kc)
}

func (f *KeycardFlow) changePairing(kc *keycardContext) error {
	if newPairing, ok := f.params[NewPairing]; ok {
		err := kc.changePairingPassword(newPairing.(string))

		if isSCardError(err) {
			return restartErr()
		} else if err != nil {
			return err
		}

		return nil
	}

	err := f.pauseAndWait(EnterNewPair, ErrorChanging)

	if err != nil {
		return err
	}

	return f.changePairing(kc)
}

func (f *KeycardFlow) sign(kc *keycardContext) (*Signature, error) {
	var err error

	path, pathOK := f.params[BIP44Path]

	if !pathOK {
		err = f.pauseAndWait(EnterPath, ErrorSigning)
		if err != nil {
			return nil, err
		}

		return f.sign(kc)
	}

	hash, hashOK := f.params[TXHash]

	var rawHash []byte

	if hashOK {
		rawHash, err = xtob(hash.(string))
		if err != nil {
			hashOK = false
		}
	}

	if !hashOK {
		err := f.pauseAndWait(EnterTXHash, ErrorSigning)
		if err != nil {
			return nil, err
		}

		return f.sign(kc)
	}

	signature, err := kc.signWithPath(rawHash, path.(string))

	if isSCardError(err) {
		return nil, restartErr()
	} else if err != nil {
		return nil, err
	}

	return toSignature(signature), nil
}
