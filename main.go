package statuskeycardgo

var kctx *keycardContext

func HelloWorld() string {
	return "Hello World"
}

////export Start
//func Start() *C.char {
//	var err error
//	kctx, err = startKeycardContext()
//	if err != nil {
//		return retValue("err", err.Error())
//	}
//	return retValue("ok", true)
//}

////export Select
//func Select() *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	info, err := kctx.selectApplet()
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "applicationInfo", ApplicationInfo{
//		Installed:              info.Installed,
//		Initialized:            info.Initialized,
//		InstanceUID:            info.InstanceUID,
//		SecureChannelPublicKey: info.SecureChannelPublicKey,
//		Version:                bytesToInt(info.Version),
//		AvailableSlots:         bytesToInt(info.AvailableSlots),
//		KeyUID:                 info.KeyUID,
//		Capabilities:           Capability(info.Capabilities),
//	})
//}

////export Stop
//func Stop() *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	if err := kctx.stop(); err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export Pair
//func Pair(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params pairParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	pairingInfo, err := kctx.pair(params.PairingPassword)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "pairingInfo", PairingInfo{
//		Key:   pairingInfo.Key,
//		Index: pairingInfo.Index,
//	})
//}

////export OpenSecureChannel
//func OpenSecureChannel(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params openSecureChannelParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.openSecureChannel(params.Index, params.Key)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export VerifyPin
//func VerifyPin(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params verifyPinParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.verifyPin(params.Pin)

//	if wrongPing, ok := err.(*keycard.WrongPINError); ok {
//		return retValue("error", err.Error(), "remainingAttempts", wrongPing.RemainingAttempts)
//	}

//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export GenerateKey
//func GenerateKey() *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	keyUID, err := kctx.generateKey()
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "keyUID", hexString(keyUID))
//}

////export DeriveKey
//func DeriveKey(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params deriveKeyParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.deriveKey(params.Path)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export SignWithPath
//func SignWithPath(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params signWithPathParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	sig, err := kctx.signWithPath(params.Data, params.Path)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "signature", Signature{
//		PublicKey: hexString(sig.PubKey()),
//		R:         hexString(sig.R()),
//		S:         hexString(sig.S()),
//		V:         sig.V(),
//	})
//}

////export ExportKey
//func ExportKey(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params exportKeyParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	privKey, pubKey, address, err := kctx.exportKey(params.Derive, params.MakeCurrent, params.OnlyPublic, params.Path)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "privateKey", hexString(privKey), "publicKey", hexString(pubKey), "address", address)
//}

////export LoadSeed
//func LoadSeed(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params loadSeedParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	pubKey, err := kctx.loadSeed(params.Seed)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "publicKey", hexString(pubKey))
//}

////export Init
//func Init(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params initSeedParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.init(params.Pin, params.Puk, params.PairingPassword)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export Unpair
//func Unpair(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params unpairParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.unpair(params.Index)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export GetStatusApplication
//func GetStatusApplication() *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	status, err := kctx.getStatusApplication()
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true, "status", ApplicationStatus{
//		PinRetryCount:  status.PinRetryCount,
//		PUKRetryCount:  status.PUKRetryCount,
//		KeyInitialized: status.KeyInitialized,
//		Path:           status.Path,
//	})
//}

////export ChangePin
//func ChangePin(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params changeSecretsParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.changePin(params.Pin)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export ChangePuk
//func ChangePuk(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params changeSecretsParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.changePuk(params.Puk)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export ChangePairingPassword
//func ChangePairingPassword(jsonParams *C.char) *C.char {
//	if kctx == nil {
//		l("select: not started")
//		return retValue("error", "not started")
//	}

//	var params changeSecretsParams
//	if err := json.Unmarshal([]byte(C.GoString(jsonParams)), &params); err != nil {
//		return retValue("error", err.Error())
//	}

//	err := kctx.changePairingPassword(params.PairingPassword)
//	if err != nil {
//		return retValue("error", err.Error())
//	}

//	return retValue("ok", true)
//}

////export KeycardSetSignalEventCallback
//func KeycardSetSignalEventCallback(cb unsafe.Pointer) {
//	signal.KeycardSetSignalEventCallback(cb)
//}

//func bytesToInt(s []byte) int {
//	if len(s) > 4 {
//		return 0
//	}

//	var b [4]byte
//	copy(b[4-len(s):], s)
//	return int(binary.BigEndian.Uint32(b[:]))
//}
