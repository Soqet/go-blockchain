package network

import (
	"bchain/internal/blockchain"
	database "bchain/internal/db"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	commandLen = 12
	protocol   = "tcp"
	port       = "13334"
)

var (
	dns = []string{
		"localhost:13335",
	}
)

func StartServer(db *database.DB, bc *blockchain.Blockchain) {
	listener, err := net.Listen(protocol, ":"+port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	v, err := db.GetVersion()
	if err != nil {
		panic(err)
	}
	if v != blockchain.BlockchainVersion {
		err = db.ClearKnownNodes()
		if err != nil {
			panic(err)
		}
		db.UpdateVersion(blockchain.BlockchainVersion)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("waiting for conn")
	for {
		conn, err := listener.Accept()
		fmt.Println("conn accepted")
		if err != nil {
			continue
		}
		fmt.Println("conn handled")
		go handleConn(conn, bc, db)
	}
}

func handleConn(conn net.Conn, bc *blockchain.Blockchain, db *database.DB) {
	defer conn.Close()
	request, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println(err)
		// return
	}
	comm := commandFromBytes(request)
	fmt.Printf("parsed command %s_%d\n", comm, len(comm))
	if handler, ok := handlers[comm]; ok {
		handler(conn, request, bc, db)
	} else {
		fmt.Println("Uknown command")
	}
}

func commandToBytes(command string) []byte {
	var res [commandLen]byte
	if len([]byte(command)) > commandLen {
		panic("INCORRECT COMMAND LENGTH")
	}
	for i, b := range []byte(command) {
		res[i] = b
	}
	return res[:]
}

func commandFromBytes(data []byte) string {
	command := [commandLen]byte{}
	i := 0
	for _, b := range data[:12] {
		if b != 0 {
			command[i] = b
			i++
		}
	}
	fmt.Println(i)
	fmt.Println(len(string(command[:i])))
	return string(command[:i])
}

func SendVersion(addr string, bc *blockchain.Blockchain) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		return
	}
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
	encoder := gob.NewEncoder(response)
	response.Write(commandToBytes("version"))
	err = encoder.Encode(&resp)
	if err != nil {
		return
	}
	io.Copy(conn, response)
	// todo: recieve version
}
