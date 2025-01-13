package event

import (
	"context"
	"fmt"
)

const (
	// TODO: integrate with sarafu-vise-events
	EventTokenTransferTag = "TOKEN_TRANSFER"
	EventTokenMintTag = "TOKEN_MINT"
	EventRegistrationTag = "CUSTODIAL_REGISTRATION"
)

type Msg struct {
	Typ string
	Item any
}

type EmitterFunc func(context.Context, Msg) error

// fields used for handling custodial registration event.
type EventCustodialRegistration struct {
	Account string
}

// fields used for handling token transfer event.
type EventTokenTransfer struct {
	To string
	Value int
	VoucherAddress string
	TxHash string
	From string
}

type EventTokenMint struct {
	To string
	Value int
	TxHash string
	VoucherAddress string
}

type EventsHandlerFunc func(context.Context, any) error

type EventsHandler struct {
	handlers map[string]EventsHandlerFunc
}

func NewEventsHandler() *EventsHandler {
	return &EventsHandler{
		handlers: make(map[string]EventsHandlerFunc),
	}
}

func (eh *EventsHandler) WithHandler(tag string, fn EventsHandlerFunc) *EventsHandler {
	eh.handlers[tag] = fn
	return eh
}

func (eh *EventsHandler) Handle(ctx context.Context, tag string, o any) error {
	fn, ok := eh.handlers[tag]
	if !ok {
		return fmt.Errorf("Handler not registered for tag: %s", tag)
	}
	return fn(ctx, o)
}
