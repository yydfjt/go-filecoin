package address

import (
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"

	"github.com/filecoin-project/go-filecoin/bls-signatures"
)

// Network represents which network an address belongs to.
type Network = byte

const (
	// Mainnet is the main network.
	Mainnet Network = iota
	// Testnet is the test network.
	Testnet
)

// Type represents the type of data address data holds
type Type = byte

const (
	// SECP256K1 means the address is the hash of a secp256k1 public key
	SECP256K1 Type = iota
	// ID means the address is an actor ID
	ID
	// Actor means the address is an acotr address, which is a fixed address
	Actor
	// BLS means the address is a full BLS public key
	BLS
)

const (
	LEN_SECP256K1 = SecpHashLength
	LEN_Actor     = SecpHashLength
	LEN_ID        = 8
	LEN_BLS       = bls.PrivateKeyBytes
)

var (
	// ErrUnknownNetwork is returned when encountering an unknown network in an address.
	ErrUnknownNetwork = errors.New("unknown network")
	// ErrUnknownType is returned when encountering an unknown address type.
	ErrUnknownType = errors.New("unknown type")
	// ErrInvalidBytes is returned when encountering an invalid byte format.
	ErrInvalidBytes = errors.New("invalid bytes")
	// ErrInvalidChecksum is returned when encountering an invalid checksum.
	ErrInvalidChecksum = errors.New("invalid checksum")
	// ErrSeralizeEmpty is returned when seralize is called on an empty address.
	ErrSeralizeEmpty = errors.New("cannot seralize an empy address")
	// ErrDeseralizeEmpty is returned when deseralize is called on an empty data.
	ErrDeseralizeEmpty = errors.New("cannot seralize an empy address")
)

// TODO all this below stuff needs to go
var (
	// TODO Should probably stop using this pattern
	// TestAddress is an account with some initial funds in it
	TestAddress Address
	// TODO Should probably stop using this pattern
	// TestAddress2 is an account with some initial funds in it
	TestAddress2 Address

	// NetworkAddress is the filecoin network
	NetworkAddress Address
	// StorageMarketAddress is the hard-coded address of the filecoin storage market
	StorageMarketAddress Address
	// PaymentBrokerAddress is the hard-coded address of the filecoin storage market
	PaymentBrokerAddress Address
)

// TODO Should probably stop using this pattern
/*
func init() {
	t := Hash([]byte("satoshi"))
	TestAddress = NewMainnet(t)

	t = Hash([]byte("nakamoto"))
	TestAddress2 = NewMainnet(t)

	n := Hash([]byte("filecoin"))
	NetworkAddress = NewMainnet(n)

	s := Hash([]byte("storage"))
	StorageMarketAddress = NewMainnet(s)

	p := Hash([]byte("payments"))
	PaymentBrokerAddress = NewMainnet(p)
}
*/
