package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mercuryoio/tonlib-go"
)

func main() {
	configPath := flag.String("config", "/Users/user/go/src/github.com/yakud/ton-blocks-stream-receiver/cmd/tonlib-test/liteserverconfig.json", "config json")
	flag.Parse()

	log.Println("load config from:", *configPath)

	config, err := tonlib.ParseConfigFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	cln, err := tonlib.NewClient(config, tonlib.Config{
		//Timeout: 1000,
	})
	if err != nil {
		log.Fatalf("Init client error: %v. ", err)
	}
	defer cln.Destroy()

	//accState, err := cln.WalletState("0:06E9EFA9B3CA6C47059E31ED56A36BB8B8C844101BD9ACFF304C0970546A65F6")
	//accState, err := cln.GetAccountState("0:06E9EFA9B3CA6C47059E31ED56A36BB8B8C844101BD9ACFF304C0970546A65F6")
	//addr, err := cln.UnpackAccountAddress("EQDfYZhDfNJ0EePoT5ibfI9oG9bWIU6g872oX5h9rL5PHY9a")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("ADDDDDR", addr)
	//accState, err := cln.GetAccountState(strings.ToLower("0:81CC51838E292BC0D48205DA1F956CDD22B8143D9CD313D75B68EAF936D50278"))
	//err = cln.Sync(313350, 623381, 623381)
	//if err != nil {
	//	log.Fatal(err)
	//}

	accState, err := cln.GetAccountState("kQCIUuaay7U5Sum5NPtF0e3iAAAvcR_IkOdtMP8MPNiysXJS")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("SUPER BALANCE:", accState.Balance)
	fmt.Println("SUPER BALANCE:", accState)

	fmt.Println("EEEEEEEEE")
}
