package address

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-filecoin/bls-signatures"
	"github.com/filecoin-project/go-filecoin/crypto"
)

var hashes = make([][]byte, 5)

func init() {
	for i := range hashes {
		hashes[i] = Hash([]byte(fmt.Sprintf("foo-%d", i)))
	}
	logging.SetDebugLogging()
}

func TestEmptyAddress(t *testing.T) {
	assert := assert.New(t)
	var emptyAddr Address
	assert.True(emptyAddr.Empty())
	stuffAddr := Address("stuff")
	assert.False(stuffAddr.Empty())
}

func TestNewAddress(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("New SECP256K1 Address", func(t *testing.T) {
		sk, err := crypto.GenerateKey()
		require.NoError(err)

		pk, ok := sk.Public().(*ecdsa.PublicKey)
		require.True(ok)

		secp256k1Addr, err := NewFromSECP256K1(Testnet, pk)
		assert.NoError(err)
		fmt.Println(secp256k1Addr)
	})

	t.Run("New ID Address", func(t *testing.T) {
		idAddress, err := NewFromActorID(Testnet, uint64(1))
		assert.NoError(err)
		fmt.Println(idAddress)
	})

	t.Run("New Actor Address", func(t *testing.T) {
		actorAddress, err := NewFromActor(Testnet, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		assert.NoError(err)
		fmt.Println(actorAddress)
	})

	t.Run("New BLS Address", func(t *testing.T) {
		blsAddress, err := NewFromBLS(Testnet, bls.PrivateKeyPublicKey((bls.PrivateKeyGenerate())))
		assert.NoError(err)
		fmt.Println(blsAddress)
	})

}

func TestAddressDecodeEncode(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("Encode Decode SECP256K1 Address", func(t *testing.T) {
		sk, err := crypto.GenerateKey()
		require.NoError(err)

		pk, ok := sk.Public().(*ecdsa.PublicKey)
		require.True(ok)

		secp256k1Addr, err := NewFromSECP256K1(Testnet, pk)
		require.NoError(err)

		addrString := secp256k1Addr.String()

		addrFromString, err := NewFromString(addrString)
		assert.NoError(err)
		assert.Equal(secp256k1Addr.Bytes(), addrFromString.Bytes())
	})

	t.Run("Encode Decode ID Address", func(t *testing.T) {
		idAddress, err := NewFromActorID(Testnet, uint64(1))
		require.NoError(err)

		addrString := idAddress.String()

		addrFromString, err := NewFromString(addrString)
		assert.NoError(err)
		assert.Equal(idAddress.Bytes(), addrFromString.Bytes())
	})

	t.Run("Encode Decode Actor Address", func(t *testing.T) {
		actorAddress, err := NewFromActor(Testnet, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		assert.NoError(err)

		addrString := actorAddress.String()

		addrFromString, err := NewFromString(addrString)
		assert.NoError(err)
		assert.Equal(actorAddress.Bytes(), addrFromString.Bytes())
	})

	t.Run("Encode Decode BLS Address", func(t *testing.T) {
		blsAddress, err := NewFromBLS(Testnet, bls.PrivateKeyPublicKey((bls.PrivateKeyGenerate())))
		assert.NoError(err)
		fmt.Println(blsAddress)

		addrString := blsAddress.String()

		addrFromString, err := NewFromString(addrString)
		assert.NoError(err)
		assert.Equal(blsAddress.Bytes(), addrFromString.Bytes())
	})
}
