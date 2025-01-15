package dev

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/db"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	"git.grassecon.net/grassrootseconomics/sarafu-api/event"
	"git.grassecon.net/grassrootseconomics/common/phone"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

var (
	logg = logging.NewVanilla().WithDomain("sarafu-api.devapi")
	aliasRegex = regexp.MustCompile("^\\+?[a-zA-Z0-9\\-_]+$")
)

const (
	pubKeyLen int = 20
	hashLen int = 32
	defaultDecimals = 6
	zeroAddress string = "0x0000000000000000000000000000000000000000"
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

func (t *Tx) ToTransferEvent() event.EventTokenTransfer {
	return event.EventTokenTransfer{
		To: t.To,
		Value: t.Value,
		VoucherAddress: t.Voucher,
		TxHash: t.Hsh,
		From: t.From,
	}
}

func (t *Tx) ToMintEvent() event.EventTokenMint {
	return event.EventTokenMint{
		To: t.To,
		Value: t.Value,
		VoucherAddress: t.Voucher,
		TxHash: t.Hsh,
	}
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

func (a *Account) ToRegistrationEvent() event.EventCustodialRegistration {
	return event.EventCustodialRegistration{
		Account: a.Address,
	}
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
	emitterFunc event.EmitterFunc
	pfx []byte
//	accountsSession map[string]string
}

func NewDevAccountService(ctx context.Context, ss storage.StorageService) *DevAccountService {
	svc := &DevAccountService{
		accounts: make(map[string]Account),
		accountsTrack: make(map[string]string),
		accountsAlias: make(map[string]string),
		vouchers: make(map[string]Voucher),
		vouchersAddress: make(map[string]string),
		txs: make(map[string]Tx),
		txsTrack: make(map[string]string),
		autoVoucherValue: make(map[string]int),
		defaultAccount: zeroAddress,
		pfx: []byte("__"),
	}
	if ss != nil {
		var err error
		svc.db, err = ss.GetUserdataDb(ctx)
		if err != nil {
			panic(err)
		}
		svc.db.SetSession("")
		svc.db.SetPrefix(db.DATATYPE_USERDATA)
		err = svc.loadAll(ctx)
		if err != nil {
			logg.DebugCtxf(ctx, "loadall error", "err", err)
		}
	}
	acc := Account{
		Address: zeroAddress,
	}
	svc.accounts[acc.Address] = acc
	return svc
}

func (das *DevAccountService) WithEmitter(fn event.EmitterFunc) *DevAccountService {
	das.emitterFunc = fn
	return das
}

func (das *DevAccountService) WithPrefix(pfx []byte) *DevAccountService {
	das.pfx = pfx
	return das
}

func (das *DevAccountService) prefixKeyFor(k string, v string) []byte {
	return append(das.pfx, []byte(k + "_" + v)...)
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
	logg.TraceCtxf(ctx, "add account", "address", acc.Address)
	return nil
}

func (das *DevAccountService) loadTx(ctx context.Context, hsh string, v []byte) error {
	var mytx Tx

	err := json.Unmarshal(v, &mytx)
	if err != nil {
		return fmt.Errorf("malformed tx: %v", hsh)
	}
	das.txs[hsh] = mytx
	das.txsTrack[mytx.Track] = hsh
	logg.TraceCtxf(ctx, "add tx", "hash", hsh)
	return nil
}

func (das *DevAccountService) loadItem(ctx context.Context, k []byte, v []byte) error {
	var err error
	s := string(k)
	ss := strings.SplitN(s[2:], "_", 2)
	if len(ss) != 2 {
		return fmt.Errorf("malformed key: %s", s)
	}
	if ss[0] == "account" {
		err = das.loadAccount(ctx, ss[1], v)
	} else if ss[0] == "tx" {
		err = das.loadTx(ctx, ss[1], v)
	} else {
		logg.ErrorCtxf(ctx, "unknown double underscore key", "key", ss[0])
	}
	return err
}

// TODO: Add connect tx and account
// TODO: update balance
func (das *DevAccountService) loadAll(ctx context.Context) error {
	dumper, err := das.db.Dump(ctx, []byte{})
	if err != nil {
		return err
	}
	for true {
		k, v := dumper.Next(ctx)
		if k == nil {
			break
		}
		if !bytes.HasPrefix(k, das.pfx) {
			continue
		}

		err = das.loadItem(ctx, k, v)
		if err != nil {
			return err
		}
	}
	return das.indexAll(ctx)
}

func (das *DevAccountService) indexAll(ctx context.Context) error {
	for k, v := range(das.txs) {
		acc := das.accounts[v.From]
		acc.Txs = append(acc.Txs, k)
		logg.TraceCtxf(ctx, "add tx to sender index", "from", v.From, "tx", k)
		if v.From == v.To {
			continue
		}
		acc = das.accounts[v.To]
		acc.Txs = append(acc.Txs, k)
		logg.TraceCtxf(ctx, "add tx to recipient index", "from", v.To, "tx", k)
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
// TODO: set max balance for 0x00 address
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
	if das.db == nil {
		return nil
	}
	k := das.prefixKeyFor("account", acc.Address)
	v, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	das.db.SetSession("")
	das.db.SetPrefix(db.DATATYPE_USERDATA)
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
	err = das.balanceAuto(ctx, pubKey)
	if err != nil {
		return nil, err
	}

	if das.defaultAccount == zeroAddress {
		das.defaultAccount = pubKey
	}

	if das.emitterFunc != nil {
		msg := event.Msg{
			Typ: event.EventRegistrationTag,
			Item: acc,
		}
		err = das.emitterFunc(ctx, msg)
		if err != nil {
			logg.ErrorCtxf(ctx, "emitter returned error", "err", err, "msg", msg)
		}
	}
	logg.TraceCtxf(ctx, "account created", "account", acc)

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

func (das *DevAccountService) saveTokenTransfer(ctx context.Context, mytx Tx) error {
	k := das.prefixKeyFor("tx", mytx.Hsh)
	v, err := json.Marshal(mytx)
	if err != nil {
		return err
	}
	das.db.SetSession("")
	das.db.SetPrefix(db.DATATYPE_USERDATA)
	return das.db.Put(ctx, []byte(k), v)
}

// TODO: set default voucher on first received
// TODO: update balance
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
	accTo, ok := das.accounts[to]
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
	mytx := Tx{
		Hsh: hsh,
		To: accTo.Address,
		From: accFrom.Address,
		Voucher: voucher.Symbol,
		Value: value,
		Track: uid.String(),
		When: time.Now(),
	}
	err = das.saveTokenTransfer(ctx, mytx)
	if err != nil {
		return nil, err
	}
	das.txs[hsh] = mytx
	if das.emitterFunc != nil {
		msg := event.Msg{
			Typ: event.EventTokenTransferTag,
			Item: mytx,
		}
		err = das.emitterFunc(ctx, msg)
		if err != nil {
			logg.ErrorCtxf(ctx, "emitter returned error", "err", err, "msg", msg)
		}
	}
	logg.TraceCtxf(ctx, "token transfer created", "tx", mytx)
	return &models.TokenTransferResponse{
		TrackingId: uid.String(),
	}, nil
}

func (das *DevAccountService) CheckAliasAddress(ctx context.Context, alias string) (*models.AliasAddress, error) {
	addr, ok := das.accountsAlias[alias]
	if !ok {
		return nil, fmt.Errorf("alias %s not found", alias)
	}
	acc, ok := das.accounts[addr]
	if !ok {
		return nil, fmt.Errorf("alias %s found but does not resolve", alias)
	}
	return &models.AliasAddress{
		Address: acc.Address,
	}, nil
}

func (das *DevAccountService) applyPhoneAlias(ctx context.Context, publicKey string, phoneNumber string) (bool, error) {
	if phoneNumber[0] == '+' {
		if !phone.IsValidPhoneNumber(phoneNumber) {
			return false, fmt.Errorf("Invalid phoneNumber number: %v", phoneNumber)
		}
		logg.DebugCtxf(ctx, "matched phoneNumber alias", "phoneNumber", phoneNumber, "address", publicKey)
		return true, nil
	}
	return false, nil
}

func (das *DevAccountService) RequestAlias(ctx context.Context, publicKey string, hint string) (*models.RequestAliasResult, error) {
	var alias string
	if !aliasRegex.MatchString(hint) {
		return nil, fmt.Errorf("alias hint does not match: %s", publicKey)
	}
	acc, ok := das.accounts[publicKey]
	if !ok {
		return nil, fmt.Errorf("address %s not found", publicKey)
	}
	alias = hint
	isPhone, err := das.applyPhoneAlias(ctx, publicKey, alias)
	if err != nil {
		return nil, fmt.Errorf("phone parser error: %v", err)
	}
	if !isPhone {
		for true {
			addr, ok := das.accountsAlias[alias]
			if !ok {
				break
			}
			if addr == publicKey {
				break
			}
			alias += "x"
		}
		acc.Alias = alias
		das.accountsAlias[alias] = publicKey
	}
	logg.DebugCtxf(ctx, "set alias", "addr", publicKey, "alias", alias)
	return &models.RequestAliasResult{
		Alias: alias,
	}, nil
}
