package network

import (
	"bchain/internal/blockchain"
	database "bchain/internal/db"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type handlerFunc func(net.Conn, []byte, *blockchain.Blockchain, *database.DB)

var handlers = map[string]handlerFunc{
	"version": handleVersion,
}

func handleVersion(conn net.Conn, request []byte, bc *blockchain.Blockchain, db *database.DB) {
	payload := request[12:]
	req := new(version)
	decoder := gob.NewDecoder(bytes.NewReader(payload))
	decoder.Decode(req)
	fmt.Printf("got %s %s %s %s\n",
		req.Addr, strconv.FormatUint(req.Height, 10),
		strconv.FormatInt(req.Timestamp, 10), strconv.FormatInt(int64(req.Version), 10),
	)
	db.AddKnownNode(req.Addr, req.Version)
	h, err := bc.GetBestHeight()
	if err != nil {
		return
	}
	resp := version{
		Version:   blockchain.BlockchainVersion,
		Height:    h,
		Timestamp: time.Now().Unix(),
		Addr:      conn.RemoteAddr().String(),
	}
	response := new(bytes.Buffer)
	response.Write(commandToBytes("version"))
	encoder := gob.NewEncoder(response)
	err = encoder.Encode(&resp)
	if err != nil {
		return
	}
	io.Copy(conn, response)
}
