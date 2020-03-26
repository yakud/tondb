package blocks_fetcher

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"gitlab.flora.loc/mills/tondb/internal/ton"
	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	HeaderLen     = 4
	blockIdError  = "block_id_error"
	blockNotFound = "block_not_found"

	FormatBoc    BlockFormat = 1
	FormatPretty BlockFormat = 2
)

type BlockFormat uint32

type Client struct {
	addr string
	conn net.Conn
	m    *sync.Mutex
}

func (c *Client) FetchBlockTlb(blockId ton.BlockId, format BlockFormat) ([]byte, error) {
	blockIdStr := fmt.Sprintf(
		"(%d,%s,%d):%d",
		blockId.WorkchainId,
		utils.DecToHex(blockId.Shard),
		blockId.SeqNo,
		format,
	)

	c.m.Lock()
	defer c.m.Unlock()
	if _, err := fmt.Fprintf(c.conn, blockIdStr); err != nil {
		conn, err := c.connect()
		if err != nil {
			return nil, fmt.Errorf("connection error: %w", err)
		}
		c.conn = conn
		return c.FetchBlockTlb(blockId, format)
	}

	blockTlb, err := c.readFromConn(c.conn)
	if err != nil {
		return nil, fmt.Errorf("read from conn error: %w", err)
	}

	return blockTlb, nil
}

func (c *Client) connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) readFromConn(conn net.Conn) ([]byte, error) {
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
					return nil, err
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
				return nil, err
			}

			if bodyLen < int(size) {
				continue
			}
			break
		}

		// read body
		if bodyLen != int(size) {
			return nil, fmt.Errorf("body is not %d bytes. readed len: %d", size, bodyLen)
		}

		return dataBuf[0:size], nil
	}
}

func NewClient(addr string) (*Client, error) {
	c := &Client{
		addr: addr,
		m:    &sync.Mutex{},
	}
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c, nil
}
