package config

import (
	"net/url"

	"git.grassecon.net/grassrootseconomics/visedriver/env"
)

const (
	createAccountPath            = "/api/v2/account/create"
	trackStatusPath              = "/api/track"
	balancePathPrefix            = "/api/account"
	trackPath                    = "/api/v2/account/status"
	tokenTransferPrefix          = "/api/v2/token/transfer"
	voucherHoldingsPathPrefix    = "/api/v1/holdings"
	voucherTransfersPathPrefix   = "/api/v1/transfers/last10"
	voucherDataPathPrefix        = "/api/v1/token"
	SendSMSPrefix                = "api/v1/external/upsell"
	poolDepositPrefix            = "/api/v2/pool/deposit"
	poolSwapQoutePrefix          = "/api/v2/pool/quote"
	poolSwapPrefix               = "/api/v2/pool/swap"
	topPoolsPrefix               = "/api/v1/pool/top"
	retrievePoolDetailsPrefix    = "/api/v1/pool/reverse"
	poolSwappableVouchersPrefix  = "/api/v1/pool"
	AliasRegistrationPrefix      = "/api/v1/internal/register"
	AliasResolverPrefix          = "/api/v1/resolve"
	ExternalSMSPrefix            = "/api/v1/external"
	AliasUpdatePrefix            = "/api/v1/internal/update"
	CreditSendPrefix             = "/api/v1/credit-send"
	CreditSendReverseQuotePrefix = "/api/v1/pool/reverse-quote"
)

var (
	custodialURLBase    string
	dataURLBase         string
	BearerToken         string
	aliasEnsURLBase     string
	externalSMSBase     string
	IncludeStablesParam string
)

var (
	CreateAccountURL          string
	TrackStatusURL            string
	BalanceURL                string
	TrackURL                  string
	TokenTransferURL          string
	VoucherHoldingsURL        string
	VoucherTransfersURL       string
	VoucherDataURL            string
	PoolDepositURL            string
	PoolSwapQuoteURL          string
	PoolSwapURL               string
	TopPoolsURL               string
	RetrievePoolDetailsURL    string
	PoolSwappableVouchersURL  string
	SendSMSURL                string
	AliasRegistrationURL      string
	AliasResolverURL          string
	ExternalSMSURL            string
	AliasUpdateURL            string
	CreditSendURL             string
	CreditSendReverseQuoteURL string
)

func setBase() error {
	var err error

	custodialURLBase = env.GetEnv("CUSTODIAL_URL_BASE", "http://localhost:5003")
	dataURLBase = env.GetEnv("DATA_URL_BASE", "http://localhost:5006")
	aliasEnsURLBase = env.GetEnv("ALIAS_ENS_BASE", "http://localhost:5015")
	externalSMSBase = env.GetEnv("EXTERNAL_SMS_BASE", "http://localhost:5035")
	BearerToken = env.GetEnv("BEARER_TOKEN", "")
	IncludeStablesParam = env.GetEnv("INCLUDE_STABLES_PARAM", "false")

	_, err = url.Parse(custodialURLBase)
	if err != nil {
		return err
	}
	_, err = url.Parse(dataURLBase)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig() error {
	err := setBase()
	if err != nil {
		return err
	}
	CreateAccountURL, _ = url.JoinPath(custodialURLBase, createAccountPath)
	TrackStatusURL, _ = url.JoinPath(custodialURLBase, trackStatusPath)
	BalanceURL, _ = url.JoinPath(custodialURLBase, balancePathPrefix)
	TrackURL, _ = url.JoinPath(custodialURLBase, trackPath)
	TokenTransferURL, _ = url.JoinPath(custodialURLBase, tokenTransferPrefix)
	VoucherHoldingsURL, _ = url.JoinPath(dataURLBase, voucherHoldingsPathPrefix)
	VoucherTransfersURL, _ = url.JoinPath(dataURLBase, voucherTransfersPathPrefix)
	VoucherDataURL, _ = url.JoinPath(dataURLBase, voucherDataPathPrefix)
	SendSMSURL, _ = url.JoinPath(dataURLBase, SendSMSPrefix)
	PoolDepositURL, _ = url.JoinPath(custodialURLBase, poolDepositPrefix)
	PoolSwapQuoteURL, _ = url.JoinPath(custodialURLBase, poolSwapQoutePrefix)
	PoolSwapURL, _ = url.JoinPath(custodialURLBase, poolSwapPrefix)
	TopPoolsURL, _ = url.JoinPath(dataURLBase, topPoolsPrefix)
	RetrievePoolDetailsURL, _ = url.JoinPath(dataURLBase, retrievePoolDetailsPrefix)
	PoolSwappableVouchersURL, _ = url.JoinPath(dataURLBase, poolSwappableVouchersPrefix)
	AliasRegistrationURL, _ = url.JoinPath(aliasEnsURLBase, AliasRegistrationPrefix)
	AliasResolverURL, _ = url.JoinPath(aliasEnsURLBase, AliasResolverPrefix)
	ExternalSMSURL, _ = url.JoinPath(externalSMSBase, ExternalSMSPrefix)
	AliasUpdateURL, _ = url.JoinPath(aliasEnsURLBase, AliasUpdatePrefix)
	CreditSendURL, _ = url.JoinPath(dataURLBase, CreditSendPrefix)
	CreditSendReverseQuoteURL, _ = url.JoinPath(dataURLBase, CreditSendReverseQuotePrefix)

	return nil
}
