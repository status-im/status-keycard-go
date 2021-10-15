package main

import "C"

import (
	"fmt"

	skg "github.com/status-im/status-keycard-go"
)

func main() {
	res := skg.HelloWorld()
	fmt.Printf("result: %+v\n", res)
}
