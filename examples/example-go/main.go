package main

import "C"

import (
	"encoding/json"
	"fmt"
	"time"

	skg "github.com/status-im/status-keycard-go"
	"github.com/status-im/status-keycard-go/signal"
)

var flow *skg.KeycardFlow
var finished chan (struct{})

func signalHandler(j []byte) {
	var sig signal.Envelope
	json.Unmarshal(j, &sig)

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
		case skg.FlowResult:
			fmt.Printf("Flow result: %+v\n", sig.Event)
			close(finished)
		}
	}()
}

func main() {
	res := skg.HelloWorld()
	fmt.Printf("result: %+v\n", res)
	var err error

	flow, err = skg.NewFlow("")

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	signal.SetKeycardSignalHandler(signalHandler)

	finished = make(chan struct{})
	err = flow.Start(skg.GetAppInfo, skg.FlowParams{})

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	<-finished
}
