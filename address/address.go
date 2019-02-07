package address

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"

	"github.com/btcsuite/btcutil/base58"

	"github.com/filecoin-project/go-filecoin/bls-signatures"
	"github.com/filecoin-project/go-filecoin/crypto"
)

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

// String returns the string representation of an address, or `<invalid address>` on a best effort basis if `a` is invalid.
func (a Address) String() string {
	// I am aware this is going to get ripped to shit in CR.... :')
	// This is '5' so the checksum call below doesn't panic
	if len(a) < 5 {
		return "<invalid address>"
	}

	var prefix string
	switch a.Network() {
	case Mainnet:
		prefix = "fc"
		break
	case Testnet:
		prefix = "tf"
		break
	default:
		return "<invalid address>"
	}

	// ID's do not have a checksum, bail
	if a.Type() == ID {
		return prefix + base58.Encode(a.suffix())
	}

	return prefix + base58.Encode(append(a.suffix(), checksum(a.suffix())...))
}

// decodeFromString is a helper method that attempts to extract the components from an address `addr`.
func decodeFromString(addr string) (string, Type, []byte, []byte, error) {
	//     | network  |  type  |     data     | checksum |
	//     | 1 byte   | 1 byte |  8-48 bytes  |  4 bytes |
	// base58Encoded [                                    ]
	// Not included for type: ID            [             ]
	// min length of an address: 11
	if len(addr) < 11 {
		return "", 0, []byte{}, []byte{}, ErrInvalidBytes
	}
	cksmShft := 4

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

	addrRaw := base58.Decode(addr[2:])
	// type is the first byte
	typ := addrRaw[0]
	// ID's do not have a checksum so adjust the shift bit
	if typ == ID {
		cksmShft = 0
	}
	// data is everything after the type up to the last 4 bytes
	data := addrRaw[1 : len(addrRaw)-cksmShft]
	// checksum is the last 4 bytes
	cksm := addrRaw[len(addrRaw)-cksmShft : len(addrRaw)]

	return ntwrk, typ, data, cksm, nil
}

// NewFromString tries to parse a given string into a filecoin address.
func NewFromString(addr string) (Address, error) {
	// decode the address into is components
	ntwrk, typ, data, cksm, err := decodeFromString(addr)
	if err != nil {
		return "", err
	}

	// does the address belong to a known network
	var in []byte
	switch ntwrk {
	case "fc":
		in = append(in, Mainnet)
	case "tf":
		in = append(in, Testnet)
	default:
		return "", ErrUnknownNetwork
	}

	// the checksum digest is produced from the address type and data, so grab that here.
	cksmIngest := append([]byte{typ}, data...)
	switch typ {
	case SECP256K1:
		if len(data) != LEN_SECP256K1 {
			return "", ErrInvalidBytes
		}
		if !verifyChecksum(cksmIngest, cksm) {
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
		if !verifyChecksum(cksmIngest, cksm) {
			return "", ErrInvalidChecksum
		}
		break
	case BLS:
		if len(data) != LEN_BLS {
			return "", ErrInvalidBytes
		}
		if !verifyChecksum(cksmIngest, cksm) {
			return "", ErrInvalidChecksum
		}
		break
	default:
		return "", ErrUnknownType
	}

	// checksum and data length validates, we have a winner!
	in = append(in, typ)
	return Address(append(in, data...)), nil
}

// NewFromSECP256K1 returns an address for the actor represented by secp256k1 public key `pk`.
// TODO Accept an interface, make type assertion on the key, combine this with the BLS method
// call it NewFromPublicKey()
func NewFromSECP256K1(n Network, pk *ecdsa.PublicKey) (Address, error) {
	// TODO should we be hashing here?
	return New(n, SECP256K1, Hash(crypto.ECDSAPubToBytes(pk)))
}

// NewFromBLS returns an address for the actor represented by BLS public key `pk`.
// TODO Accept an interface, make type assertion on the key, combine this with the SECP256K1 method
// call it NewFromPublicKey()
func NewFromBLS(n Network, pk bls.PublicKey) (Address, error) {
	return New(n, BLS, pk[:])
}

// NewFromActorID returns an address for the actor represented by `id`.
func NewFromActorID(n Network, id uint64) (Address, error) {
	data := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(data, id)
	return New(n, ID, data)
}

// NewFromActor returns an address created for the actor represented by `data`.
func NewFromActor(n Network, data []byte) (Address, error) {
	return New(n, Actor, Hash(data))
}

// New returns an address for network `n`, of type `t`, containing value `data`.
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

// checksum returns the last 4 bytes of sha256 data
func checksum(data []byte) []byte {
	cksm := sha256.Sum256(data)
	return cksm[len(cksm)-4:]
}

// verify checksum returns true if checksum(data) matches cksm
func verifyChecksum(data, cksm []byte) bool {
	maybeCksm := sha256.Sum256(data)
	return (0 == bytes.Compare(maybeCksm[len(maybeCksm)-4:], cksm))
}

// Address sufix is everything encoded on an address. All but the first byte.
func (a Address) suffix() []byte {
	return []byte(a[1:])
}
