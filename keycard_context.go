package statuskeycardgo

import (
	"fmt"

	"github.com/ebfe/scard"
	"github.com/ethereum/go-ethereum/crypto"
	keycard "github.com/status-im/keycard-go"
	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/io"
	"github.com/status-im/keycard-go/types"
	"github.com/status-im/status-keycard-go/signal"
)

type keycardContext struct {
	cardCtx   *scard.Context
	card      *scard.Card
	readers   []string
	c         types.Channel
	cmdSet    *keycard.CommandSet
	connected chan (struct{})
	runErr    error
}

func startKeycardContext() (*keycardContext, error) {
	kctx = &keycardContext{
		connected: make(chan (struct{})),
	}
	err := kctx.start()
	if err != nil {
		return nil, err
	}

	go kctx.run()

	return kctx, nil
}

func (kc *keycardContext) start() error {
	cardCtx, err := scard.EstablishContext()
	if err != nil {
		err = newKeycardError("no pcsc service")
		l(err.Error())
		close(kc.connected)
		return err
	}

	l("listing readers")
	readers, err := cardCtx.ListReaders()
	if err != nil {
		err = newKeycardError("cannot get readers")
		l(err.Error())
		close(kc.connected)
		return err
	}

	kc.readers = readers

	if len(readers) == 0 {
		l("no smartcard reader found")
		err = newKeycardError("no smartcard reader found")
		l(err.Error())
		close(kc.connected)
		return err
	}

	kc.cardCtx = cardCtx
	return nil
}

func (kc *keycardContext) stop() error {
	if kc.runErr != nil {
		return kc.runErr
	}

	if err := kc.cardCtx.Release(); err != nil {
		err = newKeycardError(fmt.Sprintf("error releasing card context %v", err))
		l(err.Error())
		return err
	}

	return nil
}

func (kc *keycardContext) run() {
	l("waiting for card")
	index, err := kc.waitForCard(kc.cardCtx, kc.readers)
	if err != nil {
		l(err.Error())
		kc.runErr = err
		close(kc.connected)
		return
	}

	signal.SendKeycardConnected("")
	l("card found at index %d", index)
	reader := kc.readers[index]

	l("using reader %s", reader)

	card, err := kc.cardCtx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		// error connecting to card
		kc.runErr = err
		close(kc.connected)
		return
	}

	status, err := card.Status()
	if err != nil {
		kc.runErr = err
		close(kc.connected)
		return
	}

	switch status.ActiveProtocol {
	case scard.ProtocolT0:
		l("card protocol T0")
	case scard.ProtocolT1:
		l("card protocol T1")
	default:
		l("card protocol T unknown")
	}

	kc.card = card
	kc.c = io.NewNormalChannel(card)
	kc.cmdSet = keycard.NewCommandSet(kc.c)
	close(kc.connected)
}

func (kc *keycardContext) waitForCard(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))

	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}

			rs[i].CurrentState = rs[i].EventState
		}

		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			return -1, err
		}
	}
}

func (kc *keycardContext) selectApplet() (*types.ApplicationInfo, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	err := kc.cmdSet.Select()
	if err != nil {
		if e, ok := err.(*apdu.ErrBadResponse); ok && e.Sw == globalplatform.SwFileNotFound {
			err = nil
		} else {
			l("select failed %+v", err)
			return nil, err
		}
	}

	return kc.cmdSet.ApplicationInfo, nil
}

func (kc *keycardContext) pair(pairingPassword string) (*types.PairingInfo, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	err := kc.cmdSet.Pair(pairingPassword)
	if err != nil {
		l("pair failed %+v", err)
		return nil, err
	}

	return kc.cmdSet.PairingInfo, nil
}

func (kc *keycardContext) openSecureChannel(index int, key []byte) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	kc.cmdSet.SetPairingInfo(key, index)
	err := kc.cmdSet.OpenSecureChannel()
	if err != nil {
		l("openSecureChannel failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) verifyPin(pin string) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.VerifyPIN(pin)
	if err != nil {
		l("verifyPin failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) generateKey() ([]byte, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	appStatus, err := kc.cmdSet.GetStatusApplication()
	if err != nil {
		l("getStatus failed %+v", err)
		return nil, err
	}

	if appStatus.KeyInitialized {
		l("generateKey failed - already generated - %+v", err)
		return nil, newKeycardError("key already generated")
	}

	keyUID, err := kc.cmdSet.GenerateKey()
	if err != nil {
		l("generateKey failed %+v", err)
		return nil, err
	}

	return keyUID, nil
}

func (kc *keycardContext) removeKey() error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.RemoveKey()
	if err != nil {
		l("removeKey failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) deriveKey(path string) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.DeriveKey(path)
	if err != nil {
		l("deriveKey failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) signWithPath(data []byte, path string) (*types.Signature, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	sig, err := kc.cmdSet.SignWithPath(data, path)
	if err != nil {
		l("signWithPath failed %+v", err)
		return nil, err
	}

	return sig, nil
}

func (kc *keycardContext) exportKey(derive bool, makeCurrent bool, onlyPublic bool, path string) (*KeyPair, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	address := ""
	privKey, pubKey, err := kc.cmdSet.ExportKey(derive, makeCurrent, onlyPublic, path)
	if err != nil {
		l("exportKey failed %+v", err)
		return nil, err
	}

	if pubKey != nil {
		ecdsaPubKey, err := crypto.UnmarshalPubkey(pubKey)
		if err != nil {
			return nil, err
		}

		address = crypto.PubkeyToAddress(*ecdsaPubKey).Hex()
	}

	return &KeyPair{Address: address, PublicKey: pubKey, PrivateKey: privKey}, nil
}

func (kc *keycardContext) loadSeed(seed []byte) ([]byte, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	pubKey, err := kc.cmdSet.LoadSeed(seed)
	if err != nil {
		l("loadSeed failed %+v", err)
		return nil, err
	}

	return pubKey, nil
}

func (kc *keycardContext) init(pin, puk, pairingPassword string) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	secrets := keycard.NewSecrets(pin, puk, pairingPassword)
	err := kc.cmdSet.Init(secrets)
	if err != nil {
		l("init failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) unpair(index uint8) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.Unpair(index)
	if err != nil {
		l("unpair failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) unpairCurrent() error {
	return kc.unpair(uint8(kc.cmdSet.PairingInfo.Index))
}

func (kc *keycardContext) getStatusApplication() (*types.ApplicationStatus, error) {
	<-kc.connected
	if kc.runErr != nil {
		return nil, kc.runErr
	}

	status, err := kc.cmdSet.GetStatusApplication()
	if err != nil {
		l("getStatusApplication failed %+v", err)
		return nil, err
	}

	return status, nil
}

func (kc *keycardContext) changePin(pin string) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.ChangePIN(pin)
	if err != nil {
		l("chaingePin failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) changePuk(puk string) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.ChangePUK(puk)
	if err != nil {
		l("chaingePuk failed %+v", err)
		return err
	}

	return nil
}

func (kc *keycardContext) changePairingPassword(pairingPassword string) error {
	<-kc.connected
	if kc.runErr != nil {
		return kc.runErr
	}

	err := kc.cmdSet.ChangePairingSecret(pairingPassword)
	if err != nil {
		l("chaingePairingPassword failed %+v", err)
		return err
	}

	return nil
}
