package network

type version struct {
	Version   int32
	Height    uint64
	Timestamp int64
	Addr      string
}

type addr struct {
	AddrList []string
}

type inv struct {
	Type      string
	Inventory [][]byte
}

type getdata struct {
	Inventory [][]byte
}

type getblocks struct {
	LastHash           []byte
	BlockLocatorHashes []uint64
}

type getheaders struct {
	BlockLocatorHashes []uint64
	LastHash           []byte
}
