package main

import (
	"fmt"
	"log"

	"gitlab.flora.loc/mills/tondb/internal/liteclient"
)

func main() {
	config := &liteclient.CmdClientConfig{
		//ExecPath:    "lite-client",
		ExecPath: "/data/ton/src/build/lite-client/lite-client",
		//WorkingDir:  "/data/ton/validator-work/",
		WorkingDir:  "/data/ton/liteclient-work",
		PubCertPath: "/data/ton/liteclient-work/liteserver.pub",
		Host:        "144.76.140.152",
		Port:        8271,
	}
	client := liteclient.NewCmdClient(config)

	linesOut, err := client.Exec2("runmethod -1:3333333333333333333333333333333333333333333333333333333333333333 active_election_id")
	if err != nil {
		log.Fatal(err)
	}

	for _, l := range linesOut {
		fmt.Println(l)
	}
}
