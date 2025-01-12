package dev

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

const (
	pubKeyLen int = 20
)

type tx struct {
	hsh string
	to string
	from string
	voucher string
	value int
	when time.Time
}

type account struct {
	track string
	nonce int
	defaultVoucher string
	balances map[string]int
	txs []tx
}

type voucher struct {
	name string
	address string
	symbol string
	decimals int
	sink string
	commodity string
	location string
}

var (
	vouchers = make(map[string]voucher)
	vouchersAddress = make(map[string]string)
)

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
	if acc.defaultVoucher == "" {
		return nil, fmt.Errorf("no default voucher set for: %v", publicKey)
	}
	bal, ok := acc.balances[acc.defaultVoucher]
	if !ok {
		return nil, fmt.Errorf("balance not found for default token %s pubkey %v", acc.defaultVoucher, publicKey)
	}
	return &models.BalanceResult {
		Balance: strconv.Itoa(bal),
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
	das.accounts[pubKey] = account{
		track: uid.String(),
	}
	das.accountsTrack[uid.String()] = pubKey
	return &models.AccountResult{
		PublicKey: pubKey,
		TrackingId: uid.String(),
	}, nil
}

func (das *DevAccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error) {
	var ok bool
	_, ok = das.accounts[publicKey]
	if !ok {
		return nil, fmt.Errorf("account not found (publickey): %v", publicKey)
	}
	return &models.TrackStatusResult{
		Active: true,
	}, nil
}

func (das *DevAccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	var holdings []dataserviceapi.TokenHoldings
	acc, ok := das.accounts[publicKey]
	if !ok {
		return nil, fmt.Errorf("account not found (publickey): %v", publicKey)
	}
	for k, v := range(acc.balances) {
		voucher, ok := vouchers[k]
		if !ok {
			return nil, fmt.Errorf("voucher has balance but object not found: %v", k)
		}
		holdings = append(holdings, dataserviceapi.TokenHoldings{
			ContractAddress: voucher.address,
			TokenSymbol: voucher.symbol,
			TokenDecimals: strconv.Itoa(voucher.decimals),
			Balance: strconv.Itoa(v),
		})
	}
	return holdings, nil
}

func (das *DevAccountService) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	var lasttx []dataserviceapi.Last10TxResponse
	acc, ok := das.accounts[publicKey]
	if !ok {
		return nil, fmt.Errorf("account not found (publickey): %v", publicKey)
	}
	for i, v := range(acc.txs) {
		if i == 10 {
			break	
		}
		voucher, ok := vouchers[v.voucher]
		if !ok {
			return nil, fmt.Errorf("voucher %s in tx list but not found in voucher list", v.voucher)
		}
		lasttx = append(lasttx, dataserviceapi.Last10TxResponse{
			Sender: v.from,
			Recipient: v.to,
			TransferValue: strconv.Itoa(v.value),
			ContractAddress: voucher.address,
			TxHash: v.hsh,
			DateBlock: v.when,
			TokenSymbol: voucher.symbol,
			TokenDecimals: strconv.Itoa(voucher.decimals),
		})
	}
	return lasttx, nil
}

func (das *DevAccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	sym, ok := vouchersAddress[address]
	if !ok {
		return nil, fmt.Errorf("voucher address %v not found", address)
	}
	voucher, ok := vouchers[sym]
	if !ok {
		return nil, fmt.Errorf("voucher address %v found but does not resolve", address)
	}
	return &models.VoucherDataResult{
		TokenName: voucher.name,
		TokenSymbol: voucher.symbol,
		TokenDecimals: voucher.decimals,
		SinkAddress: voucher.sink,
		TokenCommodity: voucher.commodity,
		TokenLocation: voucher.location,

	}, nil
}
