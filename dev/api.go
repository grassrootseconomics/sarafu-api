package dev

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gofrs/uuid"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
)

const (
	pubKeyLen int = 20
)

type account struct {
	balance int
	nonce int
}

type DevAccountService struct {
	accounts map[string]account
	accountsTrack map[string]string
//	accountsSession map[string]string
}

func NewDevAccountService() *DevAccountService {
	return &DevAccountService{
		accounts: make(map[string]account),
		accountsTrack: make(map[string]string),
		//accountsSession: make(map[string]string),
	}
}

func (das *DevAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	acc, ok := das.accounts[publicKey]
	if !ok {
		return nil, fmt.Errorf("account not found (publickey): %v", publicKey)
	}
	return &models.BalanceResult {
		Balance: strconv.Itoa(acc.balance),
		Nonce: json.Number(strconv.Itoa(acc.nonce)),
	}, nil
}


func (das *DevAccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	var b [pubKeyLen]byte
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	c, err := rand.Read(b[:])
	if err != nil {
		return nil, err
	}
	if c != pubKeyLen {
		return nil, fmt.Errorf("short read: %d", c)
	}
	pubKey := fmt.Sprintf("0x%x", b)
	das.accounts[pubKey] = account{}
	das.accountsTrack[uid.String()] = pubKey
	return &models.AccountResult{
		PublicKey: pubKey,
		TrackingId: uid.String(),
	}, nil
}
