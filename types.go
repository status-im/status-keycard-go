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
