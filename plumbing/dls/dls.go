package dls

import (
	cbor "gx/ipfs/QmRoARq3nkUb13HSKZGepCZSWe5GrVPwx7xURJGZ7KWv9V/go-ipld-cbor"
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	"gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore/query"

	"github.com/filecoin-project/go-filecoin/protocol/storage/deal"
	"github.com/filecoin-project/go-filecoin/repo"
)

// Querier is plumbing implementation querying deals
type Querier struct {
	dealsDs repo.Datastore
}

// New returns a new Querier.
func New(dealsDatastore repo.Datastore) *Querier {
	return &Querier{dealsDs: dealsDatastore}
}

// Query returns a channel of deals matching the given query.
func (querier *Querier) Query(qry query.Query) (<-chan *deal.Deal, <-chan error) {
	out := make(chan *deal.Deal)
	errorOrDoneC := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errorOrDoneC)

		results, err := querier.dealsDs.Query(qry)
		if err != nil {
			errorOrDoneC <- errors.Wrap(err, "failed to query deals from datastore")
			return
		}
		for entry := range results.Next() {
			var storageDeal deal.Deal
			if err := cbor.DecodeInto(entry.Value, &storageDeal); err != nil {
				errorOrDoneC <- errors.Wrap(err, "failed to unmarshal deals from datastore")
				return
			}
			out <- &storageDeal
		}
		errorOrDoneC <- nil
	}()

	return out, errorOrDoneC
}
