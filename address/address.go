package address

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"

	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"

	"github.com/btcsuite/btcutil/base58"

	"github.com/filecoin-project/go-filecoin/bls-signatures"
	"github.com/filecoin-project/go-filecoin/crypto"
)

var log = logging.Logger("address")

func init() {
	logging.SetDebugLogging()
}

// Address is the go type that represents an address in the filecoin network.
type Address string

// Network represents which network the address belongs to.
func (a Address) Network() Network {
	return a[0]
}

// Type represents the type of content contained in the address.
func (a Address) Type() Type {
	return a[1]
}

// Data represents the content contained in the address.
func (a Address) Data() []byte {
	return []byte(a[2:])
}

// Empty returns true if the address is empty.
func (a Address) Empty() bool {
	return 0 == len(a)
}

// Bytes returns the byte representation of the address.
func (a Address) Bytes() []byte {
	return []byte(a)
}

// String returns the string representation of an address.
func (a Address) String() string {
	// TODO sanity ensure the address is not empty
	// returning an error from the String() methods is really annoying though

	var prefix string
	switch a.Network() {
	case Mainnet:
		prefix = "fc"
		break
	case Testnet:
		prefix = "tf"
		break
	default:
		panic("BAD")
	}
	// ID's do not have a checksum, bail
	if a.Type() == ID {
		return prefix + base58.Encode(a.suffix())
	}
	return prefix + base58.Encode(append(a.suffix(), checksum(a.suffix())...))
}

func decodeFromString(addr string) (string, Type, []byte, []byte, error) {
	// | fc | t |...ID...|
	// | 2  | 1 |   8    |
	// min length of an address: 11
	if len(addr) < 11 {
		return "", 0, []byte{}, []byte{}, ErrInvalidBytes
	}

	var ntwrk string
	switch addr[:2] {
	case "fc":
		ntwrk = "fc"
		break
	case "tf":
		ntwrk = "tf"
		break
	default:
		return "", 0, []byte{}, []byte{}, ErrInvalidBytes
	}

	// decode the address to its 4 parts
	addrRaw := base58.Decode(addr[2:])
	//    | network | type |    data    |    checksum    |
	//TODO data extration iswrong
	return ntwrk, addrRaw[0], addrRaw[1:][:len(addrRaw[1:])-4], addrRaw[:4], nil
}

// NewFromString tries to parse a given string into a filecoin address.
func NewFromString(addr string) (Address, error) {
	// decode the address and break it up into its components
	ntwrk, typ, data, cksm, err := decodeFromString(addr)
	if err != nil {
		return "", err
	}
	log.Infof("network: %s, type: %x, data: %v, cksm: %x", ntwrk, typ, data, cksm)

	// check the address network
	var in []byte
	switch ntwrk {
	case "fc":
		in = append(in, Mainnet)
	case "tf":
		in = append(in, Testnet)
	default:
		return "", ErrUnknownNetwork
	}

	// check the address type, data length and checksum
	switch typ {
	case SECP256K1:
		if len(data) != LEN_SECP256K1 {
			panic(len(data))
			return "", ErrInvalidBytes
		}
		if !verifyChecksum(data, cksm) {
			return "", ErrInvalidChecksum
		}
		break
	case ID:
		if len(data) != LEN_ID {
			return "", ErrInvalidBytes
		}
		// ID's do not have a checksum
		break
	case Actor:
		if len(data) != LEN_Actor {
			return "", ErrInvalidBytes
		}
		if !verifyChecksum(data, cksm) {
			return "", ErrInvalidBytes
		}
		break
	case BLS:
		if len(data) != LEN_BLS {
			return "", ErrInvalidBytes
		}
		if !verifyChecksum(data, cksm) {
			return "", ErrInvalidBytes
		}
		break
	default:
		return "", ErrUnknownType
	}

	// we didn't error, data is valid
	in = append(in, typ)
	return Address(append(in, data...)), nil
}

// TODO maybe we want to accept a hash instead?
func NewFromSECP256K1(n Network, pk *ecdsa.PublicKey) (Address, error) {
	// TODO should we be hashing here?
	return New(n, SECP256K1, Hash(crypto.ECDSAPubToBytes(pk)))
}

func NewFromBLS(n Network, pk bls.PublicKey) (Address, error) {
	return New(n, BLS, pk[:])
}

func NewFromActorID(n Network, id uint64) (Address, error) {
	data := make([]byte, 8)
	binary.PutUvarint(data, id)
	return New(n, ID, data)
}

func NewFromActor(n Network, randomData []byte) (Address, error) {
	return New(n, Actor, randomData)
}

// New returns an address for network `n`, for data type `t`.
func New(n Network, t Type, data []byte) (Address, error) {
	var out []byte
	if n != Mainnet && n != Testnet {
		return "", ErrUnknownNetwork
	}
	if t != SECP256K1 && t != ID && t != Actor && t != BLS {
		return "", ErrUnknownType
	}
	out = append([]byte{n, t}, data...)
	return Address(out), nil
}

// NetworkToString creates a human readable representation of the network.
func NetworkToString(n Network) string {
	switch n {
	case Mainnet:
		return "fc"
	case Testnet:
		return "tf"
	default:
		panic("invalid network")
	}
}

func checksum(data []byte) []byte {
	// checksum is the last 4 bytes of the sha of address type and data
	cksm := sha256.Sum256(data)
	return cksm[:4]
}

func verifyChecksum(data, cksm []byte) bool {
	maybeCksm := sha256.Sum256(data)
	return (0 == bytes.Compare(maybeCksm[:4], cksm))
}

// Address sufix is everything encoded on an address. All but the first byte.
func (a Address) suffix() []byte {
	return []byte(a[1:])
}
