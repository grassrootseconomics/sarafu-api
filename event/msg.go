package event

import (
	"context"
)

const (
	// TODO: integrate with sarafu-vise-events
	EventTokenTransferTag = "TOKEN_TRANSFER"
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
