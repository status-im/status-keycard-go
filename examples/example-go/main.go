package main

import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	skg "github.com/status-im/status-keycard-go"
	"github.com/status-im/status-keycard-go/signal"
)

var flow *skg.KeycardFlow
var finished chan (struct{})
var ppIdx = 0
var pairingPasses = [2]string{"WrongOne", "KeycardTest"}
var correctPIN = "123456"

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
			fmt.Printf("Swap card. You have 5 seconds\n")
			time.Sleep(5 * time.Second)
			flow.Resume(skg.FlowParams{})
		case skg.EnterPairing:
			fmt.Printf("Entering pass: %+v\n", pairingPasses[ppIdx])
			flow.Resume(skg.FlowParams{skg.PairingPass: pairingPasses[ppIdx]})
			ppIdx = (ppIdx + 1) % 2
		case skg.EnterPIN:
			fmt.Printf("Entering PIN: %+v\n", correctPIN)
			flow.Resume(skg.FlowParams{skg.PIN: correctPIN})
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

	testFlow(skg.GetAppInfo, skg.FlowParams{})
	testFlow(skg.RecoverAccount, skg.FlowParams{skg.PIN: "234567"})
}
