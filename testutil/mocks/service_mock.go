package mocks

import (
	"context"

	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/stretchr/testify/mock"
)

// MockAccountService implements AccountServiceInterface for testing
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	args := m.Called()
	return args.Get(0).(*models.AccountResult), args.Error(1)
}

func (m *MockAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*models.BalanceResult), args.Error(1)
}

func (m *MockAccountService) TrackAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResult, error) {
	args := m.Called(trackingId)
	return args.Get(0).(*models.TrackStatusResult), args.Error(1)
}

func (m *MockAccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	args := m.Called(publicKey)
	return args.Get(0).([]dataserviceapi.TokenHoldings), args.Error(1)
}

func (m *MockAccountService) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).([]dataserviceapi.Last10TxResponse), args.Error(1)
}

func (m *MockAccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	args := m.Called(address)
	return args.Get(0).(*models.VoucherDataResult), args.Error(1)
}

func (m *MockAccountService) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.TokenTransferResponse), args.Error(1)
}

func (m *MockAccountService) CheckAliasAddress(ctx context.Context, alias string) (*models.AliasAddress, error) {
	args := m.Called(alias)
	return args.Get(0).(*models.AliasAddress), args.Error(1)
}

func (m *MockAccountService) RequestAlias(ctx context.Context, publicKey string, hint string) (*models.RequestAliasResult, error) {
	args := m.Called(publicKey, hint)
	return args.Get(0).(*models.RequestAliasResult), args.Error(1)
}

func (m *MockAccountService) SendUpsellSMS(ctx context.Context, inviterPhone, inviteePhone string) (*models.SendSMSResponse, error) {
	args := m.Called(inviterPhone, inviteePhone)
	return args.Get(0).(*models.SendSMSResponse), args.Error(1)
}

func (m *MockAccountService) SendPINResetSMS(ctx context.Context, admin, phone string) error {
	return nil
}

func (m *MockAccountService) SendAddressSMS(ctx context.Context, publicKey, originPhone string) error {
	return nil
}

func (m MockAccountService) PoolDeposit(ctx context.Context, amount, from, poolAddress, tokenAddress string) (*models.PoolDepositResult, error) {
	args := m.Called(amount, from, poolAddress, tokenAddress)
	return args.Get(0).(*models.PoolDepositResult), args.Error(1)
}

func (m MockAccountService) FetchTopPools(ctx context.Context) ([]dataserviceapi.PoolDetails, error) {
	args := m.Called()
	return args.Get(0).([]dataserviceapi.PoolDetails), args.Error(1)
}

func (m MockAccountService) RetrievePoolDetails(ctx context.Context, sym string) (*dataserviceapi.PoolDetails, error) {
	args := m.Called()
	return args.Get(0).(*dataserviceapi.PoolDetails), args.Error(1)
}

func (m MockAccountService) GetPoolSwappableFromVouchers(ctx context.Context, poolAddress, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	args := m.Called(poolAddress, publicKey)
	return args.Get(0).([]dataserviceapi.TokenHoldings), args.Error(1)
}

func (m MockAccountService) GetPoolSwappableVouchers(ctx context.Context, poolAddress string) ([]dataserviceapi.TokenDetails, error) {
	args := m.Called(poolAddress)
	return args.Get(0).([]dataserviceapi.TokenDetails), args.Error(1)
}

func (m MockAccountService) GetPoolSwapQuote(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapQuoteResult, error) {
	args := m.Called(amount, from, fromTokenAddress, poolAddress, toTokenAddress)
	return args.Get(0).(*models.PoolSwapQuoteResult), args.Error(1)
}

func (m MockAccountService) PoolSwap(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapResult, error) {
	args := m.Called(amount, from, fromTokenAddress, poolAddress, toTokenAddress)
	return args.Get(0).(*models.PoolSwapResult), args.Error(1)
}

func (m MockAccountService) GetSwapFromTokenMaxLimit(ctx context.Context, poolAddress, fromTokenAddress, toTokenAddress, publicKey string) (*models.MaxLimitResult, error) {
	args := m.Called(poolAddress, fromTokenAddress, toTokenAddress, publicKey)
	return args.Get(0).(*models.MaxLimitResult), args.Error(1)
}

func (m MockAccountService) CheckTokenInPool(ctx context.Context, poolAddress, tokenAddress string) (*models.TokenInPoolResult, error) {
	args := m.Called(poolAddress, tokenAddress)
	return args.Get(0).(*models.TokenInPoolResult), args.Error(1)
}
