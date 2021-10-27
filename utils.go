package statuskeycardgo

import "C"

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"github.com/ebfe/scard"
	keycard "github.com/status-im/keycard-go"
	ktypes "github.com/status-im/keycard-go/types"
)

func retValue(pairs ...interface{}) *C.char {
	obj := make(map[string]interface{})
	for i := 0; i < len(pairs)/2; i++ {
		key := pairs[i*2]
		value := pairs[(i*2)+1]
		obj[key.(string)] = value
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return C.CString(err.Error())
	}

	return C.CString(string(b))
}

func isSCardError(err error) bool {
	_, ok := err.(scard.Error)
	return ok
}

func getRetries(err error) (int, bool) {
	if wrongPIN, ok := err.(*keycard.WrongPINError); ok {
		return wrongPIN.RemainingAttempts, ok
	} else if wrongPUK, ok := err.(*keycard.WrongPUKError); ok {
		return wrongPUK.RemainingAttempts, ok
	} else {
		return 0, false
	}
}

func btox(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func xtob(str string) ([]byte, error) {
	return hex.DecodeString(str)
}

func bytesToInt(s []byte) int {
	if len(s) > 4 {
		return 0
	}

	var b [4]byte
	copy(b[4-len(s):], s)
	return int(binary.BigEndian.Uint32(b[:]))
}

func toAppInfo(r *ktypes.ApplicationInfo) ApplicationInfo {
	return ApplicationInfo{
		Initialized:    r.Initialized,
		InstanceUID:    r.InstanceUID,
		Version:        bytesToInt(r.Version),
		AvailableSlots: bytesToInt(r.AvailableSlots),
		KeyUID:         r.KeyUID,
	}
}

func toPairInfo(r *ktypes.PairingInfo) *PairingInfo {
	return &PairingInfo{
		Key:   r.Key,
		Index: r.Index,
	}
}

func toSignature(r *ktypes.Signature) *Signature {
	return &Signature{
		R: r.R(),
		S: r.S(),
		V: r.V(),
	}
}
