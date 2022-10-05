package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/gob"
	"os"

	"github.com/akamensky/base58"
)

const checksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type SerializedWallet struct {
	PrivKey []byte
	PubKey  []byte
}

type Wallets struct {
	Wallets map[string]Wallet
}

type SerializedWallets struct {
	Wallets map[string]SerializedWallet
}

func NewWallet() (*Wallet, error) {
	private, public, err := newKeyPair()
	if err != nil {
		return nil, err
	}
	return &Wallet{private, public}, nil
}

func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return *private, nil, err
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey, nil
}

func (w Wallet) Address(version uint32) (string, error) {
	pubKeyHash := sha256.Sum256(w.PublicKey)
	return GetAddress(pubKeyHash[:], version)
}

func GetAddress(pubKeyHash []byte, version uint32) (string, error) {
	// pubKeyHash := sha256.Sum256(publicKey)
	binVers := make([]byte, 4)
	binary.LittleEndian.PutUint32(binVers, version)
	vData := append(binVers, pubKeyHash[:]...)
	data := append(vData, checksum(vData)...)
	addr := base58.Encode(data)
	return addr, nil
}

func ValidateAddress(address string) (bool, error) {
	pubKeyHash, err := base58.Decode(address)
	if err != nil {
		return false, err
	}
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLen:]
	version := pubKeyHash[:4]
	pubKeyHash = pubKeyHash[4 : len(pubKeyHash)-checksumLen]
	targetChecksum := checksum(append(version, pubKeyHash...))
	return bytes.Equal(actualChecksum, targetChecksum), nil
}

func (ws Wallets) GetWallet(address string) Wallet {
	return ws.Wallets[address]
}

func ExtractPubKeyHash(address string) ([]byte, error) {
	pubKeyHash, err := base58.Decode(address)
	if err != nil {
		return nil, err
	}
	return pubKeyHash[4 : len(pubKeyHash)-checksumLen], nil
}

func checksum(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:checksumLen]
}

func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := NewWallet()
	if err != nil {
		return "", nil
	}
	address, err := wallet.Address(BlockchainVersion)
	if err != nil {
		return "", err
	}
	ws.Wallets[address] = *wallet

	return address, nil
}

func NewWallets(filePath string) (*Wallets, error) {
	ws := new(Wallets)
	err := ws.LoadFromFile(filePath)
	if err != nil && err.Error() == "EOF" {
		ws.Wallets = map[string]Wallet{}
		err = nil
	}
	return ws, err
}

func (w Wallet) Serialize() (*SerializedWallet, error) {
	sw := new(SerializedWallet)
	sw.PubKey = w.PublicKey
	encoded, err := x509.MarshalECPrivateKey(&w.PrivateKey)
	if err != nil {
		return nil, err
	}
	sw.PrivKey = encoded
	return sw, nil
}

func (ws *Wallets) Serialize() (*SerializedWallets, error) {
	sws := new(SerializedWallets)
	sws.Wallets = make(map[string]SerializedWallet, len(ws.Wallets))
	for a, w := range ws.Wallets {
		sw, err := w.Serialize()
		if err != nil {
			return nil, err
		}
		sws.Wallets[a] = *sw
	}
	return sws, nil
}

func (sw SerializedWallet) Deserialize() (*Wallet, error) {
	w := new(Wallet)
	w.PublicKey = sw.PubKey
	privKey, err := x509.ParseECPrivateKey(sw.PrivKey)
	if err != nil {
		return nil, err
	}
	w.PrivateKey = *privKey
	return w, nil
}

func (sws *SerializedWallets) Deserialize() (*Wallets, error) {
	ws := new(Wallets)
	ws.Wallets = make(map[string]Wallet, len(sws.Wallets))
	for a, sw := range sws.Wallets {
		w, err := sw.Deserialize()
		if err != nil {
			return nil, err
		}
		ws.Wallets[a] = *w
	}
	return ws, nil
}

func (ws Wallets) SaveToFile(filepath string) error {
	encoded := new(bytes.Buffer)
	encoder := gob.NewEncoder(encoded)
	sws, err := ws.Serialize()
	if err != nil {
		return err
	}
	err = encoder.Encode(sws)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(encoded.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (ws *Wallets) LoadFromFile(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	decoder := gob.NewDecoder(buf)
	sws := new(SerializedWallets)
	err = decoder.Decode(sws)
	if err != nil {
		return err
	}
	decodedWallets, err := sws.Deserialize()
	if err != nil {
		return err
	}
	ws.Wallets = decodedWallets.Wallets
	return nil
}
