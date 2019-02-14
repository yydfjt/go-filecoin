package mpool

import (
	"context"

	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	"gx/ipfs/QmVRxA4J3UPQpw74dLrQ6NJkfysCA1H4GU28gVpXQt9zMU/go-libp2p-pubsub"

	"github.com/filecoin-project/go-filecoin/core"
	"github.com/filecoin-project/go-filecoin/plumbing/msg"
	"github.com/filecoin-project/go-filecoin/types"
)

// Pool interfaces to the core message pool
type Pool struct {
	pool   *core.MessagePool
	pubsub *pubsub.PubSub
}

// New builds a new Pool.
func New(pool *core.MessagePool, pubsub *pubsub.PubSub) *Pool {
	return &Pool{pool: pool, pubsub: pubsub}
}

// Pending lists un-mined messages in the pool
func (ls *Pool) Pending(ctx context.Context, messageCount uint) ([]*types.SignedMessage, error) {
	pending := ls.pool.Pending()
	if len(pending) < int(messageCount) {
		subscription, err := ls.pubsub.Subscribe(msg.Topic)
		if err != nil {
			return nil, err
		}

		for len(pending) < int(messageCount) {
			_, err = subscription.Next(ctx)
			if err != nil {
				return nil, err
			}
			pending = ls.pool.Pending()
		}
	}

	return pending, nil
}

// Remove removes a message from the pool
func (ls *Pool) Remove(id cid.Cid) {
	ls.pool.Remove(id)
}
