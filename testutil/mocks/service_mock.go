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
