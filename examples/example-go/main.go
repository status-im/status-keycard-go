package main

import "C"

import (
	"fmt"

	skg "github.com/status-im/status-keycard-go"
)

func main() {
	res := skg.HelloWorld()
	fmt.Printf("result: %+v\n", res)

	flow, err := skg.NewFlow("")

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	err = flow.Start(skg.GetAppInfo, skg.FlowParams{})

	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}

	flow.Wait()
}
