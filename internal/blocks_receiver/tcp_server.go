package blocks_receiver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Handler func([]byte) error

type TcpReceiver struct {
}

func (t *TcpReceiver) Run(ctx context.Context, wg *sync.WaitGroup, handler Handler) error {
	defer wg.Done()

	var ServerAddr = "0.0.0.0:7315"

	// Listen for incoming connections.
	l, err := net.Listen("tcp", ServerAddr)
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on " + ServerAddr)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		//headerBuf := make([]byte, 4)
		//dataBuf := make([]byte, 1024*1024*50)
		bodyReader := &io.LimitedReader{R: conn}
		for {
			headerBuf := make([]byte, 4)
			dataBuf := make([]byte, 1024*1024*50)

			reqLen, err := conn.Read(headerBuf)
			if err != nil {
				if err == io.EOF {
					//fmt.Println("WAITING EOF. READED:", reqLen)
					<-time.After(time.Millisecond * 100)
					continue
				} else {
					conn.Close()
					l.Close()
					log.Println("Error reading:", err.Error())
					return err
				}
			}
			if reqLen != 4 {
				// TODO: fix
				log.Fatal("header is not 4 bytes")
			}
			size := binary.LittleEndian.Uint32(headerBuf[:4])
			if size == 0 {
				log.Println("Is empty message size. Continue..")
				break
			}
			//fmt.Println("read header: ", size, string(headerBuf), headerBuf)

			var bodyLen = 0
			for {
				needRead := int(size) - bodyLen
				bodyReader.N = int64(needRead)
				bodyLenPacket, err := bodyReader.Read(dataBuf[bodyLen:size])
				bodyLen += bodyLenPacket
				if err == io.EOF {
					if bodyLen < int(size) {
						continue
					}
					continue
				}
				if err != nil {
					conn.Close()
					l.Close()
					log.Println("Error reading body:", err.Error())
					return err
				}

				if bodyLen < int(size) {
					continue
				}
				break
			}

			// read body
			if bodyLen != int(size) {
				log.Println("body is not ", size, "bytes. readed len:", bodyLen)
			}

			if err := handler(dataBuf[:size]); err != nil {
				log.Fatal("handler fatal: ", err)
			}

		}

		// Send a response back to person contacting us.
		//conn.Write([]byte("Hello my C++!"))
		// Close the connection when you're done with it.
		conn.Close()
	}
}

//func (t *TcpReceiver) onpacket(ctx context.Context, wg *sync.WaitGroup, handler Handler) error {
//
///}

func NewTcpReceiver() *TcpReceiver {
	return &TcpReceiver{}
}
