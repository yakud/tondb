package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton"

	"gitlab.flora.loc/mills/tondb/internal/blocks_fetcher"
)

func main() {
	//msg := "(-1,8000000000000000,958228)"

	// connect to this socket
	//conn, err := net.Dial("tcp", "127.0.0.1:13699")
	//if err != nil {
	//	log.Fatal(err)
	//}
	// iptables -I INPUT 1 -p tcp --dport 13699 -j ACCEPT

	blocksFetcher, err := blocks_fetcher.NewClient("127.0.0.1:13699")
	//blocksFetcher, err := blocks_fetcher.NewClient("10.236.0.3:13699")
	if err != nil {
		log.Fatal(err)
	}

	var count uint64 = 0
	go func() {
		for {
			<-time.After(time.Second)
			fmt.Println(atomic.LoadUint64(&count), "blocks/sec")
			atomic.StoreUint64(&count, 0)
		}
	}()

	for {
		_, err := blocksFetcher.FetchBlockTlb(ton.BlockId{
			WorkchainId: -1,
			Shard:       9223372036854775808,
			SeqNo:       958483,
		})
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(string(b))

		atomic.AddUint64(&count, 1)
		<-time.After(time.Second)
		//fmt.Println(string(block))
	}

	/*
		// send to socket
		fmt.Fprintf(conn, msg+"\n")
		// listen for reply
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print("Message from server: " + message)

	*/
}
