package main

import (
	"fmt"
	"log"

	"github.com/mercuryoio/tonlib-go"
)

func main() {
	config, err := tonlib.ParseConfigFile("/Users/user/go/src/github.com/yakud/ton-blocks-stream-receiver/cmd/tonlib-test/liteserverconfig.json")
	if err != nil {
		log.Fatal(err)
	}

	cln, err := tonlib.NewClient(config, tonlib.Config{})
	if err != nil {
		log.Fatalf("Init client error: %v. ", err)
	}
	defer cln.Destroy()

	accState, err := cln.GetAccountState("EQAG6e-ps8psRwWeMe1Wo2u4uMhEEBvZrP8wTAlwVGpl9hzH")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AAAAAAAAAAAAA: %+v\n", *accState)

	fmt.Println("EEEEEEEEE")
}
