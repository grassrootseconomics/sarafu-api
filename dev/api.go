package dev

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
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

type Tx struct {
	Track string `json: "track"`
	Hsh string `json:"hash"`
	To string `json:"to"`
	From string `json: "from"`
	Voucher string `json: "voucher"`
	Value int `json: "value"`
	When time.Time `json: "when"`
}

type Account struct {
	Track string `json: "track"`
	Address string `json: "address"`
	Nonce int `json: "nonce"`
	DefaultVoucher string `json: "defaultVoucher"`
	Balances map[string]int `json: "balances"` 
	Alias string
	Txs []string `json: "txs"`
}

type Voucher struct {
	Name string `json: "name"`
	Address string `json: "address"`
	Symbol string `json: "symbol"`
	Decimals int `json: "decimals"`
	Sink string `json: "sink"`
	Commodity string `json: "commodity"`
	Location string `json: "location"`
}

type DevAccountService struct {
	dir string
	db db.Db
	accounts map[string]Account
	accountsTrack map[string]string
	accountsAlias map[string]string
	vouchers map[string]Voucher
	vouchersAddress map[string]string
	txs map[string]Tx
	txsTrack map[string]string
	toAutoCreate bool
	autoVouchers []string
	autoVoucherValue map[string]int
	defaultAccount string
//	accountsSession map[string]string
}

func NewDevAccountService(ctx context.Context, d string) *DevAccountService {
	svc := &DevAccountService{
		dir: d,
		db: fsdb.NewFsDb(),
		accounts: make(map[string]Account),
		accountsTrack: make(map[string]string),
		accountsAlias: make(map[string]string),
		vouchers: make(map[string]Voucher),
		vouchersAddress: make(map[string]string),
		txs: make(map[string]Tx),
		txsTrack: make(map[string]string),
		autoVoucherValue: make(map[string]int),
	}
	err := svc.db.Connect(ctx, d)
	if err != nil {
		panic(err)
	}
	svc.db.SetPrefix(db.DATATYPE_USERDATA)
	err = svc.loadAll(ctx)
	if err != nil {
		panic(err)
	}
	return svc
}

func (das *DevAccountService) loadAccount(ctx context.Context, pubKey string, v []byte) error {
	var acc Account

	err := json.Unmarshal(v, &acc)
	if err != nil {
		return fmt.Errorf("malformed account: %v", pubKey)
	}
	das.accounts[pubKey] = acc
	das.accountsTrack[acc.Track] = pubKey
	if acc.Alias != "" {
		das.accountsAlias[acc.Alias] = pubKey
	}
	return nil
}

func (das *DevAccountService) loadItem(ctx context.Context, k []byte, v []byte) error {
	var err error
	s := string(k)
	ss := strings.SplitN(s, "_", 2)
	if len(ss) != 2 {
		return fmt.Errorf("malformed key: %s", s)
	}
	if ss[0] == "account" {
		err = das.loadAccount(ctx, ss[1], v)
	}
	return err
}

func (das *DevAccountService) loadAll(ctx context.Context) error {
	d, err := os.ReadDir(das.dir)
	if err != nil {
		return err
	}
	for _, v := range(d) {
		// TODO: move decoding to vise
		fp := v.Name()
		k := []byte(fp[1:])
		v, err := das.db.Get(ctx, k)
		if err != nil {
			return err
		}
		err = das.loadItem(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return nil
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

// TODO: add persistence for vouchers
func (das *DevAccountService) AddVoucher(ctx context.Context, symbol string) error {
	if symbol == "" {
		return fmt.Errorf("cannot add empty sym voucher")
	}
	v, ok := das.vouchers[symbol]
	if ok {
		return fmt.Errorf("already have voucher with symbol %s", v.Symbol)
	}
	h := sha1.New()
	h.Write([]byte(symbol))
	z := h.Sum(nil)
	address := fmt.Sprintf("0x%x", z)
	das.vouchers[symbol] = Voucher{
		Name: symbol,
		Symbol: symbol,
		Address: address,
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
	if acc.DefaultVoucher == "" {
		return nil, fmt.Errorf("no default voucher set for: %v", publicKey)
	}
	bal, ok := acc.Balances[acc.DefaultVoucher]
	if !ok {
		return nil, fmt.Errorf("balance not found for default token %s pubkey %v", acc.DefaultVoucher, publicKey)
	}
	return &models.BalanceResult {
		Balance: strconv.Itoa(bal),
		Nonce: json.Number(strconv.Itoa(acc.Nonce)),
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
		_, err := das.TokenTransfer(ctx, strconv.Itoa(value), das.defaultAccount, pubKey, voucher.Address)
		if err != nil {
			return err
		}
	}
	return nil
}

func (das *DevAccountService) saveAccount(ctx context.Context, acc Account) error {
	k := "account_" + acc.Address
	v, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	return das.db.Put(ctx, []byte(k), v)
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
	acc := Account{
		Track: uid.String(),
		Address: pubKey,
	}

	err = das.saveAccount(ctx, acc)
	if err != nil {
		return nil, err
	}

	das.accounts[pubKey] = acc
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
	for k, v := range(acc.Balances) {
		voucher, ok := das.vouchers[k]
		if !ok {
			return nil, fmt.Errorf("voucher has balance but object not found: %v", k)
		}
		holdings = append(holdings, dataserviceapi.TokenHoldings{
			ContractAddress: voucher.Address,
			TokenSymbol: voucher.Symbol,
			TokenDecimals: strconv.Itoa(voucher.Decimals),
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
	for i, v := range(acc.Txs) {
		mytx := das.txs[v]
		if i == 10 {
			break	
		}
		voucher, ok := das.vouchers[mytx.Voucher]
		if !ok {
			return nil, fmt.Errorf("voucher %s in tx list but not found in voucher list", mytx.Voucher)
		}
		lasttx = append(lasttx, dataserviceapi.Last10TxResponse{
			Sender: mytx.From,
			Recipient: mytx.To,
			TransferValue: strconv.Itoa(mytx.Value),
			ContractAddress: voucher.Address,
			TxHash: mytx.Hsh,
			DateBlock: mytx.When,
			TokenSymbol: voucher.Symbol,
			TokenDecimals: strconv.Itoa(voucher.Decimals),
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
		TokenName: voucher.Name,
		TokenSymbol: voucher.Symbol,
		TokenDecimals: voucher.Decimals,
		SinkAddress: voucher.Sink,
		TokenCommodity: voucher.Commodity,
		TokenLocation: voucher.Location,

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
	das.txs[hsh] = Tx{
		Hsh: hsh,
		To: accTo.Address,
		From: accFrom.Address,
		Voucher: voucher.Symbol,
		Value: value,
		Track: uid.String(),
		When: time.Now(),
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
		Address: acc.Address,
	}, nil
}
