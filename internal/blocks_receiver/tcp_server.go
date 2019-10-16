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

const HeaderLen = 4

type Handler func([]byte) error

type TcpReceiver struct {
	ServerAddr string
}

func (t *TcpReceiver) Run(ctx context.Context, wg *sync.WaitGroup, handler Handler) error {
	defer wg.Done()

	// Listen for incoming connections.
	l, err := net.Listen("tcp", t.ServerAddr)
	if err != nil {
		log.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()
	log.Println("Listening on " + t.ServerAddr)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go func(conn net.Conn, handler Handler) {
			if err := t.worker(conn, handler); err != nil {
				log.Fatal("Worker receiver error: ", err)
			}
		}(conn, handler)
	}
}

func (t *TcpReceiver) worker(conn net.Conn, handler Handler) error {
	//headerBuf := make([]byte, HeaderLen)
	//dataBuf := make([]byte, 1024*1024*50)

	bodyReader := &io.LimitedReader{R: conn}
	for {
		headerBuf := make([]byte, HeaderLen)

		var size uint32
		var headerLen = 0
		for {
			needRead := HeaderLen - headerLen
			bodyReader.N = int64(needRead)
			reqLen, err := bodyReader.Read(headerBuf[headerLen:HeaderLen])
			headerLen += reqLen
			if err != nil {
				if err == io.EOF {
					//fmt.Println("WAITING EOF. READED:", reqLen)
					<-time.After(time.Millisecond * 100)
					continue
				} else {
					conn.Close()
					log.Println("Error reading:", err.Error())
					return err
				}
			}

			if headerLen < HeaderLen {
				continue
			}

			if headerLen != HeaderLen {
				// TODO: fix
				log.Fatal("header is not 4 bytes:", headerLen)
			}
			size = binary.LittleEndian.Uint32(headerBuf[0:4])
			if size == 0 {
				log.Println("Is empty message size. Continue..")
				break
			}
			//fmt.Println("read header: ", size, string(headerBuf), headerBuf)
			break
		}

		dataBuf := make([]byte, size)
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

		if err := handler(dataBuf[0:size]); err != nil {
			log.Fatal("handler fatal: ", err)
		}
		headerBuf = nil
		dataBuf = nil
	}

	conn.Close()

	return nil
}

func NewTcpReceiver(serverAddr string) *TcpReceiver {
	return &TcpReceiver{
		ServerAddr: serverAddr,
	}
}
