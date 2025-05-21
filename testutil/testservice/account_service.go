package testservice

import (
	"context"
	"encoding/json"

	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

type TestAccountService struct {
}

func (tas *TestAccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	return &models.AccountResult{
		TrackingId: "075ccc86-f6ef-4d33-97d5-e91cfb37aa0d",
		PublicKey:  "0x623EFAFa8868df4B934dd12a8B26CB3Dd75A7AdD",
	}, nil
}

func (tas *TestAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	balanceResponse := &models.BalanceResult{
		Balance: "0.003 CELO",
		Nonce:   json.Number("0"),
	}
	return balanceResponse, nil
}

func (tas *TestAccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error) {
	return &models.TrackStatusResult{
		Active: true,
	}, nil
}

func (tas *TestAccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	return []dataserviceapi.TokenHoldings{
		dataserviceapi.TokenHoldings{
			ContractAddress: "0x6CC75A06ac72eB4Db2eE22F781F5D100d8ec03ee",
			TokenSymbol:     "SRF",
			TokenDecimals:   "6",
			Balance:         "2745987",
		},
	}, nil
}

func (tas *TestAccountService) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	return []dataserviceapi.Last10TxResponse{}, nil
}

func (m TestAccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	return &models.VoucherDataResult{}, nil
}

func (tas *TestAccountService) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	return &models.TokenTransferResponse{
		TrackingId: "e034d147-747d-42ea-928d-b5a7cb3426af",
	}, nil
}

func (m TestAccountService) PoolDeposit(ctx context.Context, amount, from, poolAddress, tokenAddress string) (*models.PoolDepositResult, error) {
	return &models.PoolDepositResult{}, nil
}

func (m *TestAccountService) CheckAliasAddress(ctx context.Context, alias string) (*models.AliasAddress, error) {
	return &models.AliasAddress{}, nil
}

func (m *TestAccountService) RequestAlias(ctx context.Context, publicKey string, hint string) (*models.RequestAliasResult, error) {
	return &models.RequestAliasResult{}, nil
}

func (m *TestAccountService) SendUpsellSMS(ctx context.Context, inviterPhone, inviteePhone string) (*models.SendSMSResponse, error) {
	return &models.SendSMSResponse{}, nil
}

func (m *TestAccountService) SendAddressSMS(ctx context.Context, publicKey, originPhone string) error {
	return nil
}

func (m *TestAccountService) SendPINResetSMS(ctx context.Context, admin, phone string) error {
	return nil
}

func (m TestAccountService) FetchTopPools(ctx context.Context) ([]dataserviceapi.PoolDetails, error) {
	return []dataserviceapi.PoolDetails{}, nil
}

func (m TestAccountService) GetPoolSwappableFromVouchers(ctx context.Context, poolAddress, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	return []dataserviceapi.TokenHoldings{}, nil
}

func (m TestAccountService) GetPoolSwappableVouchers(ctx context.Context, poolAddress string) ([]dataserviceapi.TokenDetails, error) {
	return []dataserviceapi.TokenDetails{}, nil
}

func (m TestAccountService) GetPoolSwapQuote(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapQuoteResult, error) {
	return &models.PoolSwapQuoteResult{}, nil
}

func (m TestAccountService) PoolSwap(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapResult, error) {
	return &models.PoolSwapResult{}, nil
}

func (m TestAccountService) GetSwapFromTokenMaxLimit(ctx context.Context, poolAddress, fromTokenAddress, toTokenAddress, publicKey string) (*models.MaxLimitResult, error) {
	return &models.MaxLimitResult{}, nil
}

func (m TestAccountService) CheckTokenInPool(ctx context.Context, poolAddress, tokenAddress string) (*models.TokenInPoolResult, error) {
	return &models.TokenInPoolResult{}, nil
}
