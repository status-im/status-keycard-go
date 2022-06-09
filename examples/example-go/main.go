package main

import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	skg "github.com/status-im/status-keycard-go"
	"github.com/status-im/status-keycard-go/signal"
)

var flow *skg.KeycardFlow
var finished chan (struct{})
var correctPairing = "KeycardDefaultPairing"
var correctPIN = "123456"
var correctPUK = "123456123456"
var keyUID = "136cbfc087cf7df6cf3248bce7563d4253b302b2f9e2b5eef8713fa5091409bc"

func signalHandler(j []byte) {
	var sig signal.Envelope
	json.Unmarshal(j, &sig)
	fmt.Printf("Received signal: %+v\n", sig)

	go func() {
		switch sig.Type {
		case skg.InsertCard:
			fmt.Print("Insert card\n")
		case skg.CardInserted:
			fmt.Printf("Card inserted\n")
		case skg.SwapCard:
			fmt.Printf("Swap card. Changing constraint\n")
			flow.Resume(skg.FlowParams{skg.KeyUID: keyUID})
		case skg.EnterPairing:
			fmt.Printf("Entering pass: %+v\n", correctPairing)
			flow.Resume(skg.FlowParams{skg.PairingPass: correctPairing})
		case skg.EnterPIN:
			fmt.Printf("Entering PIN: %+v\n", correctPIN)
			flow.Resume(skg.FlowParams{skg.PIN: correctPIN})
		case skg.EnterNewPIN:
			fmt.Printf("Creating PIN: %+v\n", correctPIN)
			flow.Resume(skg.FlowParams{skg.NewPIN: correctPIN})
		case skg.EnterNewPUK:
			fmt.Printf("Creating PUK: %+v\n", correctPUK)
			flow.Resume(skg.FlowParams{skg.NewPUK: correctPUK})
		case skg.EnterNewPair:
			fmt.Printf("Creating pairing: %+v\n", correctPairing)
			flow.Resume(skg.FlowParams{skg.NewPairing: correctPairing})
		case skg.EnterMnemonic:
			fmt.Printf("Loading mnemonic\n")
			flow.Resume(skg.FlowParams{skg.Mnemonic: "receive fan copper bracket end train again sustain wet siren throw cigar"})
		case skg.FlowResult:
			fmt.Printf("Flow result: %+v\n", sig.Event)
			close(finished)
		}
	}()
}

func testFlow(typ skg.FlowType, params skg.FlowParams) {
	finished = make(chan struct{})
	err := flow.Start(typ, params)

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	<-finished
}

func testRecoverAccount() {
	finished = make(chan struct{})
	err := flow.Start(skg.RecoverAccount, skg.FlowParams{})

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	<-finished
}

func main() {
	dir, err := os.MkdirTemp("", "status-keycard-go")
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		return
	}

	defer os.RemoveAll(dir)

	pairingsFile := filepath.Join(dir, "keycard-pairings.json")

	flow, err = skg.NewFlow(pairingsFile)

	if err != nil {
		fmt.Printf("error: %+v\n", err)
		return
	}

	signal.SetKeycardSignalHandler(signalHandler)

	testFlow(skg.GetAppInfo, skg.FlowParams{skg.FactoryReset: true})
	testFlow(skg.LoadAccount, skg.FlowParams{skg.MnemonicLen: 12})
	testFlow(skg.UnpairThis, skg.FlowParams{skg.PIN: correctPIN})
	testFlow(skg.RecoverAccount, skg.FlowParams{skg.PairingPass: "WrongPass", skg.PIN: "234567"})
	testFlow(skg.Login, skg.FlowParams{skg.KeyUID: "60a78c98d5dd659f714eb7072bfb2c0d8a65f74a8f6aff7bb27cf56ae1feec17"})
	testFlow(skg.GetAppInfo, skg.FlowParams{})
	testFlow(skg.ExportPublic, skg.FlowParams{skg.BIP44Path: "m/44'/60'/0'/0/1"})
	testFlow(skg.Sign, skg.FlowParams{skg.TXHash: "60a78c98d5dd659f714eb7072bfb2c0d8a65f74a8f6aff7bb27cf56ae1feec17", skg.BIP44Path: "m/44'/60'/0'/0/0"})
	testFlow(skg.UnpairThis, skg.FlowParams{skg.PIN: correctPIN})
}
