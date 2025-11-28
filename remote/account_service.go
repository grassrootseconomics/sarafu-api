package remote

import (
	"context"

	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

type AccountService interface {
	CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error)
	CreateAccount(ctx context.Context) (*models.AccountResult, error)
	TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error)
	FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error)
	FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error)
	VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error)
	TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error)
	CheckAliasAddress(ctx context.Context, alias string) (*models.AliasAddress, error)
	RequestAlias(ctx context.Context, hint string, publicKey string) (*models.RequestAliasResult, error)
	UpdateAlias(ctx context.Context, name string, publicKey string) (*models.RequestAliasResult, error)
	SendUpsellSMS(ctx context.Context, inviterPhone, inviteePhone string) (*models.SendSMSResponse, error)
	SendAddressSMS(ctx context.Context, publicKey, originPhone string) error
	SendPINResetSMS(ctx context.Context, admin, phone string) error
	PoolDeposit(ctx context.Context, amount, from, poolAddress, tokenAddress string) (*models.PoolDepositResult, error)
	FetchTopPools(ctx context.Context) ([]dataserviceapi.PoolDetails, error)
	RetrievePoolDetails(ctx context.Context, sym string) (*dataserviceapi.PoolDetails, error)
	GetPoolSwappableFromVouchers(ctx context.Context, poolAddress, publicKey string) ([]dataserviceapi.TokenHoldings, error)
	GetPoolSwappableVouchers(ctx context.Context, poolAddress string) ([]dataserviceapi.TokenHoldings, error)
	GetPoolSwapQuote(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapQuoteResult, error)
	PoolSwap(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapResult, error)
	GetSwapFromTokenMaxLimit(ctx context.Context, poolAddress, fromTokenAddress, toTokenAddress, publicKey string) (*models.MaxLimitResult, error)
	CheckTokenInPool(ctx context.Context, poolAddress, tokenAddress string) (*models.TokenInPoolResult, error)
	GetCreditSendMaxLimit(ctx context.Context, poolAddress, fromTokenAddress, toTokenAddress, publicKey string) (*models.CreditSendLimitsResult, error)
	GetCreditSendReverseQuote(ctx context.Context, poolAddress, fromTokenAddress, toTokenAddress, toTokenAMount string) (*models.CreditSendReverseQouteResult, error)
	MpesaTriggerOnramp(ctx context.Context, address, phoneNumber, asset string, amount int) (*models.MpesaOnrampResponse, error)
}
