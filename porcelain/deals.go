package porcelain

import (
	"context"
	"fmt"

	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	"gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore/query"

	"github.com/filecoin-project/go-filecoin/protocol/storage/deal"
)

type dlsPlumbing interface {
	ChainLs(ctx context.Context) <-chan interface{}
	DealQuery(qry query.Query) (<-chan *deal.Deal, <-chan error)
}

// DealsLs returns a channel of all deals and a channel for errors or done
func DealsLs(api dlsPlumbing) (<-chan *deal.Deal, <-chan error) {
	return api.DealQuery(query.Query{Prefix: "/" + deal.ClientDatastorePrefix})
}

// DealByCid returns a single deal matching a given cid or an error
func DealByCid(api dlsPlumbing, dealCid cid.Cid) (*deal.Deal, error) {
	dealsC, errorOrDoneC := api.DealQuery(query.Query{Prefix: "/" + deal.ClientDatastorePrefix})
	select {
	case storageDeal := <-dealsC:
		if storageDeal.Response.ProposalCid == dealCid {
			return storageDeal, nil
		}
	case errOrNil := <-errorOrDoneC:
		return nil, errOrNil
	}
	return nil, fmt.Errorf("Could not find deal with CID: %s", dealCid.String())
}
