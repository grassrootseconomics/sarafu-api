package dev

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

var (
	logg = logging.NewVanilla().WithDomain("sarafu-api.devapi")
)

const (
	pubKeyLen int = 20
	hashLen int = 32
	defaultDecimals = 6
	zeroAccount string = "0x0000000000000000000000000000000000000000"
)

type tx struct {
	hsh string
	to string
	from string
	voucher string
	value int
	when time.Time
	track string
}

type account struct {
	track string
	address string
	nonce int
	defaultVoucher string
	balances map[string]int
	txs []string
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

type DevAccountService struct {
	accounts map[string]account
	accountsTrack map[string]string
	accountsAlias map[string]string
	vouchers map[string]voucher
	vouchersAddress map[string]string
	txs map[string]tx
	txsTrack map[string]string
	toAutoCreate bool
	autoVouchers []string
	autoVoucherValue map[string]int
	defaultAccount string
//	accountsSession map[string]string
}

func NewDevAccountService() *DevAccountService {
	return &DevAccountService{
		accounts: make(map[string]account),
		accountsTrack: make(map[string]string),
		accountsAlias: make(map[string]string),
		vouchers: make(map[string]voucher),
		vouchersAddress: make(map[string]string),
		txs: make(map[string]tx),
		txsTrack: make(map[string]string),
		autoVoucherValue: make(map[string]int),
	}
}

func (das *DevAccountService) WithAutoVoucher(ctx context.Context, symbol string, value int) *DevAccountService {
	err := das.AddVoucher(ctx, symbol)
	if err != nil {
		logg.ErrorCtxf(ctx, "cannot add autovoucher %s: %v", symbol, err)
		return das
	}
	das.autoVouchers = append(das.autoVouchers, symbol)
	das.autoVoucherValue[symbol] = value
	return das
}

func (das *DevAccountService) AddVoucher(ctx context.Context, symbol string) error {
	if symbol == "" {
		return fmt.Errorf("cannot add empty sym voucher")
	}
	v, ok := das.vouchers[symbol]
	if ok {
		return fmt.Errorf("already have voucher with symbol %s", v.symbol)
	}
	h := sha1.New()
	h.Write([]byte(symbol))
	z := h.Sum(nil)
	address := fmt.Sprintf("0x%x", z)
	das.vouchers[symbol] = voucher{
		name: symbol,
		symbol: symbol,
		address: address,
	}
	das.vouchersAddress[address] = symbol
	logg.InfoCtxf(ctx, "added dev voucher", "symbol", symbol, "address", address)
	return nil
}

// AccountService implementation below

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

func (das *DevAccountService) balanceAuto(ctx context.Context, pubKey string) error {
	for _, v := range(das.autoVouchers) {
		voucher, ok := das.vouchers[v]
		if !ok {
			return fmt.Errorf("balance auto voucher set but not resolved: %s", v)
		}
		value, ok := das.autoVoucherValue[v]
		if !ok {
			value = 0
		}
		_, err := das.TokenTransfer(ctx, strconv.Itoa(value), das.defaultAccount, pubKey, voucher.address)
		if err != nil {
			return err
		}
	}
	return nil
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
		address: pubKey,
	}
	das.accountsTrack[uid.String()] = pubKey
	das.balanceAuto(ctx, pubKey)

	if das.defaultAccount == zeroAccount {
		das.defaultAccount = pubKey
	}
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
		voucher, ok := das.vouchers[k]
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
		mytx := das.txs[v]
		if i == 10 {
			break	
		}
		voucher, ok := das.vouchers[mytx.voucher]
		if !ok {
			return nil, fmt.Errorf("voucher %s in tx list but not found in voucher list", mytx.voucher)
		}
		lasttx = append(lasttx, dataserviceapi.Last10TxResponse{
			Sender: mytx.from,
			Recipient: mytx.to,
			TransferValue: strconv.Itoa(mytx.value),
			ContractAddress: voucher.address,
			TxHash: mytx.hsh,
			DateBlock: mytx.when,
			TokenSymbol: voucher.symbol,
			TokenDecimals: strconv.Itoa(voucher.decimals),
		})
	}
	return lasttx, nil
}

func (das *DevAccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	sym, ok := das.vouchersAddress[address]
	if !ok {
		return nil, fmt.Errorf("voucher address %v not found", address)
	}
	voucher, ok := das.vouchers[sym]
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

func (das *DevAccountService) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	var b [hashLen]byte
	value, err := strconv.Atoi(amount)
	if err != nil {
		return nil, err
	}
	accFrom, ok := das.accounts[from]
	if !ok {
		return nil, fmt.Errorf("sender account %v not found", from)	
	}
	accTo, ok := das.accounts[from]
	if !ok {
		if !das.toAutoCreate {
			return nil, fmt.Errorf("recipient account %v not found, and not creating", from)	
		}
	}

	sym, ok := das.vouchersAddress[tokenAddress]
	if !ok {
		return nil, fmt.Errorf("voucher address %v not found", tokenAddress)
	}
	voucher, ok := das.vouchers[sym]
	if !ok {
		return nil, fmt.Errorf("voucher address %v found but does not resolve", tokenAddress)
	}

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	c, err := rand.Read(b[:])
	if err != nil {
		return nil, err
	}
	if c != hashLen {
		return nil, fmt.Errorf("tx hash short read: %d", c)
	}
	hsh := fmt.Sprintf("0x%x", b)
	das.txs[hsh] = tx{
		hsh: hsh,
		to: accTo.address,
		from: accFrom.address,
		voucher: voucher.symbol,
		value: value,
		track: uid.String(),
		when: time.Now(),
	}
	return &models.TokenTransferResponse{
		TrackingId: uid.String(),
	}, nil
}

func (das *DevAccountService) CheckAliasAddress(ctx context.Context, alias string) (*dataserviceapi.AliasAddress, error) {
	addr, ok := das.accountsAlias[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s not found", alias)
	}
	acc, ok := das.accounts[addr]
	if !ok {
		return nil, fmt.Errorf("alias %s found but does not resolve", alias)
	}
	return &dataserviceapi.AliasAddress{
		Address: acc.address,
	}, nil
}
