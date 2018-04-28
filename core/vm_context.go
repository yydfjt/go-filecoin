package core

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/filecoin-project/go-filecoin/abi"
	"github.com/filecoin-project/go-filecoin/types"
)

// VMContext is the only thing exposed to an actor while executing.
// All methods on the VMContext are ABI methods exposed to actors.
type VMContext struct {
	from    *types.Actor
	to      *types.Actor
	message *types.Message
	state   types.StateTree
}

// NewVMContext returns an initialized context.
func NewVMContext(from, to *types.Actor, msg *types.Message, st types.StateTree) *VMContext {
	return &VMContext{
		from:    from,
		to:      to,
		message: msg,
		state:   st,
	}
}

// Message retrieves the message associated with this context.
func (ctx *VMContext) Message() *types.Message {
	return ctx.message
}

// ReadStorage reads the storage from the associated to actor.
func (ctx *VMContext) ReadStorage() []byte {
	return ctx.to.ReadStorage()
}

// WriteStorage writes to the storage of the associated to actor.
func (ctx *VMContext) WriteStorage(memory []byte) error {
	ctx.to.WriteStorage(memory)
	return ctx.state.SetActor(context.Background(), ctx.message.To, ctx.to)
}

// Send sends a message to another actor.
// This method assumes to be called from inside the `to` actor.
func (ctx *VMContext) Send(to types.Address, method string, value *types.TokenAmount, params []interface{}) ([]byte, uint8,
	error) {
	deps := vmContextSendDeps{
		EncodeValues:     abi.EncodeValues,
		GetOrCreateActor: ctx.state.GetOrCreateActor,
		Send:             Send,
		SetActor:         ctx.state.SetActor,
		ToValues:         abi.ToValues,
	}

	return ctx.send(deps, to, method, value, params) // <== Problem 1
	// The problem here is re-entrancy. Downstream callees may have modified
	// the from or to actor's state but we aren't reloading them here to
	// get the changes. We have a stale copy of the actors. Which is
	// kind of a bummer if downstream callers transferred value out of them
	// for example. We should reload here. Some proposals forthcoming for
	// how to deal with the bigger picture.
}

type vmContextSendDeps struct {
	EncodeValues     func([]*abi.Value) ([]byte, error)
	GetOrCreateActor func(context.Context, types.Address, func() (*types.Actor, error)) (*types.Actor, error)
	Send             func(context.Context, *types.Actor, *types.Actor, *types.Message, types.StateTree) ([]byte, uint8, error)
	SetActor         func(context.Context, types.Address, *types.Actor) error
	ToValues         func([]interface{}) ([]*abi.Value, error)
}

// send sends a message to another actor. It exists alongside send so that we can inject its dependencies during test.
func (ctx *VMContext) send(deps vmContextSendDeps, to types.Address, method string, value *types.TokenAmount, params []interface{}) ([]byte, uint8,
	error) {
	// the message sender is the `to` actor, so this is what we set as `from` in the new message
	from := ctx.Message().To
	fromActor := ctx.to // <== Problem 2
	// The issue is that we're propagating the to actor from the calling context.
	// If there's a	bug or even a different implementation that causes this instance of
	// the actor's state to be different than what's in the state tree for him, we have
	// a problem: this instance is going to overwrite whatever is in the state tree
	// when we call deps.Send below (the second thing it does is save both actors). 
	// It doesn't seem safe to both save/load state to/from the state tree and propagate 
	// objects within these vm calls: if we do both then we'll always have room for error. 
	// We should do exclusively one or the other.
	//
	// A meta-problem is that we don't have a contract for who saves what, or at least
	// we don't have one written down. (A proposal is forthcoming.)

	vals, err := deps.ToValues(params)
	if err != nil {
		return nil, 1, faultErrorWrap(err, "failed to convert inputs to abi values")
	}

	paramData, err := deps.EncodeValues(vals)
	if err != nil {
		return nil, 1, revertErrorWrap(err, "encoding params failed")
	}

	msg := types.NewMessage(from, to, 0, value, method, paramData)
	if msg.From == msg.To {
		// TODO: handle this
		return nil, 1, newFaultErrorf("unhandled: sending to self (%s)", msg.From)
	}

	toActor, err := deps.GetOrCreateActor(context.TODO(), msg.To, func() (*types.Actor, error) {
		return NewAccountActor(nil)
	})

	if err != nil {
		return nil, 1, faultErrorWrapf(err, "failed to get or create To actor %s", msg.To)
	}
	// TODO(fritz) de-dup some of the logic between here and core.Send
	out, ret, err := deps.Send(context.Background(), fromActor, toActor, msg, ctx.state)
	if err != nil {
		return nil, ret, err
	}

	return out, ret, nil
}

// AddressForNewActor creates computes the address for a new actor in the same
// way that ethereum does.  Note that this will not work if we allow the
// creation of multiple contracts in a given invocation (nonce will remain the
// same, resulting in the same address back)
func (ctx *VMContext) AddressForNewActor() (types.Address, error) {
	return computeActorAddress(ctx.message.From, ctx.from.Nonce)
}

func computeActorAddress(creator types.Address, nonce uint64) (types.Address, error) {
	buf := new(bytes.Buffer)

	if _, err := buf.Write(creator.Bytes()); err != nil {
		return types.Address{}, err
	}

	if err := binary.Write(buf, binary.BigEndian, nonce); err != nil {
		return types.Address{}, err
	}

	hash, err := types.AddressHash(buf.Bytes())
	if err != nil {
		return types.Address{}, err
	}

	return types.NewMainnetAddress(hash), nil
}
