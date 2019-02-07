package porcelain

import (
	"context"

	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	"gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore/query"

	"github.com/filecoin-project/go-filecoin/protocol/storage/deal"
)

type dlsPlumbing interface {
	ChainLs(ctx context.Context) <-chan interface{}
	Find(qry query.Query) (<-chan *deal.Deal, <-chan error)
}

// DealsLs returns a channel of all deals and a channel for errors or done
func DealsLs(api dlsPlumbing) (<-chan *deal.Deal, <-chan error) {
	return api.Find(query.Query{Prefix: "/" + deal.ClientDatastorePrefix})
}

// Deal returns a single deal matching a given cid
func Deal(api dlsPlumbing, cid cid.Cid) (*deal.Deal, error) {
	dealsC, errorOrDoneC := api.Find(query.Query{Prefix: "/" + deal.ClientDatastorePrefix + "/" + cid.String()})
	select {
	case storageDeal := <-dealsC:
		return storageDeal, nil
	case errOrNil := <-errorOrDoneC:
		return nil, errOrNil
	}
}
