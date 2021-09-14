package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type hexString []byte

// MarshalJSON serializes hexString to hex
func (s hexString) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(fmt.Sprintf("%x", string(s)))
	return bytes, err
}

// UnmarshalJSON deserializes hexString to hex
func (s *hexString) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	str, err := hex.DecodeString(x)
	if err != nil {
		return err
	}

	*s = hexString([]byte(str))
	return nil
}

type openSecureChannelParams struct {
	Index int       `json:"index"`
	Key   hexString `json:"key"`
}

type verifyPinParams struct {
	Pin string `json:"pin"`
}

type deriveKeyParams struct {
	Path string `json:"path"`
}

type signWithPathParams struct {
	Data hexString `json:"data"`
	Path string    `json:"path"`
}

type exportKeyParams struct {
	Derive      bool   `json:"derive"`
	MakeCurrent bool   `json:"makeCurrent"`
	OnlyPublic  bool   `json:"onlyPublic"`
	Path        string `json:"path"`
}

type loadSeedParams struct {
	Seed hexString `json:"seed"`
}

type pairParams struct {
	PairingPassword string `json:"pairingPassword"`
}

type initSeedParams struct {
	Pin             string `json:"pin"`
	Puk             string `json:"puk"`
	PairingPassword string `json:"pairingPassword"`
}

type changeSecretsParams struct {
	Pin             string `json:"pin"`
	Puk             string `json:"puk"`
	PairingPassword string `json:"pairingPassword"`
}

type unpairParams struct {
	Index uint8 `json:"index"`
}

type Signature struct {
	PublicKey hexString `json:"publicKey"`
	R         hexString `json:"r"`
	S         hexString `json:"s"`
	V         byte      `json:"v"`
}

type PairingInfo struct {
	Key   hexString `json:"key"`
	Index int       `json:"index"`
}

type ApplicationStatus struct {
	PinRetryCount  int    `json:"pinRetryCount"`
	PUKRetryCount  int    `json:"pukRetryCount"`
	KeyInitialized bool   `json:"keyInitialized"`
	Path           string `json:"path"`
}
