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
	SendUpsellSMS(ctx context.Context, inviterPhone, inviteePhone string) (*models.SendSMSResponse, error)
	SendAddressSMS(ctx context.Context, publicKey, originPhone string) error
	SendPINResetSMS(ctx context.Context, admin, phone string) error
}
