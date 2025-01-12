package config

import (
	"net/url"

	"git.grassecon.net/grassrootseconomics/visedriver/env"
)

const (
	createAccountPath          = "/api/v2/account/create"
	trackStatusPath            = "/api/track"
	balancePathPrefix          = "/api/account"
	trackPath                  = "/api/v2/account/status"
	tokenTransferPrefix        = "/api/v2/token/transfer"
	voucherHoldingsPathPrefix  = "/api/v1/holdings"
	voucherTransfersPathPrefix = "/api/v1/transfers/last10"
	voucherDataPathPrefix      = "/api/v1/token"
	AliasPrefix                = "api/v1/alias"
)

var (
	custodialURLBase string
	dataURLBase      string
	BearerToken      string
)

var (
	CreateAccountURL    string
	TrackStatusURL      string
	BalanceURL          string
	TrackURL            string
	TokenTransferURL    string
	VoucherHoldingsURL  string
	VoucherTransfersURL string
	VoucherDataURL      string
	CheckAliasURL       string
)

const (
	defaultHTTPHost string = "127.0.0.1"
	defaultHTTPPort uint = 7123
)

func setBase() error {
	var err error

	custodialURLBase = env.GetEnv("CUSTODIAL_URL_BASE", "http://localhost:5003")
	dataURLBase = env.GetEnv("DATA_URL_BASE", "http://localhost:5006")
	BearerToken = env.GetEnv("BEARER_TOKEN", "")

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
	CheckAliasURL, _ = url.JoinPath(dataURLBase, AliasPrefix)
	return nil
}

func Host() string {
	return env.GetEnv("HOST", defaultHTTPHost)
}

func Port() uint {
	return env.GetEnvUint("PORT", defaultHTTPPort)
}
