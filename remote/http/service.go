package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/sarafu-api/config"
	"git.grassecon.net/grassrootseconomics/sarafu-api/dev"
	"git.grassecon.net/grassrootseconomics/sarafu-api/models"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

var (
	aliasRegex = regexp.MustCompile("^\\+?[a-zA-Z0-9\\-_]+$")
	logg       = logging.NewVanilla().WithDomain("sarafu-api.devapi")
)

type HTTPAccountService struct {
	SS     storage.StorageService
	UseApi bool
}

// Parameters:
//   - trackingId: A unique identifier for the account.This should be obtained from a previous call to
//     CreateAccount or a similar function that returns an AccountResponse. The `trackingId` field in the
//     AccountResponse struct can be used here to check the account status during a transaction.
//
// Returns:
//   - string: The status of the transaction as a string. If there is an error during the request or processing, this will be an empty string.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil
func (as *HTTPAccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*models.TrackStatusResult, error) {
	var r models.TrackStatusResult

	ep, err := url.JoinPath(config.TrackURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (as *HTTPAccountService) ToFqdn(alias string) string {
	return alias + ".sarafu.eth"
}

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
func (as *HTTPAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	var balanceResult models.BalanceResult

	ep, err := url.JoinPath(config.BalanceURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &balanceResult)
	return &balanceResult, err
}

// CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *HTTPAccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	var r models.AccountResult
	// Create a new request
	req, err := http.NewRequest("POST", config.CreateAccountURL, nil)
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FetchVouchers retrieves the token holdings for a given public key from the data indexer API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *HTTPAccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	var r struct {
		Holdings []dataserviceapi.TokenHoldings `json:"holdings"`
	}

	ep, err := url.JoinPath(config.VoucherHoldingsURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return r.Holdings, nil
}

// FetchTransactions retrieves the last 10 transactions for a given public key from the data indexer API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *HTTPAccountService) FetchTransactions(ctx context.Context, publicKey string) ([]dataserviceapi.Last10TxResponse, error) {
	var r struct {
		Transfers []dataserviceapi.Last10TxResponse `json:"transfers"`
	}

	ep, err := url.JoinPath(config.VoucherTransfersURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return r.Transfers, nil
}

// VoucherData retrieves voucher metadata from the data indexer API endpoint.
// Parameters:
//   - address: The voucher address.
func (as *HTTPAccountService) VoucherData(ctx context.Context, address string) (*models.VoucherDataResult, error) {
	var r struct {
		TokenDetails models.VoucherDataResult `json:"tokenDetails"`
	}

	ep, err := url.JoinPath(config.VoucherDataURL, address)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	return &r.TokenDetails, err
}

// TokenTransfer creates a new token transfer in the custodial system.
// Returns:
//   - *models.TokenTransferResponse: A pointer to an TokenTransferResponse struct containing the trackingId.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *HTTPAccountService) TokenTransfer(ctx context.Context, amount, from, to, tokenAddress string) (*models.TokenTransferResponse, error) {
	var r models.TokenTransferResponse

	// Create request payload
	payload := map[string]string{
		"amount":       amount,
		"from":         from,
		"to":           to,
		"tokenAddress": tokenAddress,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create a new request
	req, err := http.NewRequest("POST", config.TokenTransferURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// CheckAliasAddress retrieves the address of an alias from the API endpoint.
// Parameters:
//   - alias: The alias of the user.
func (as *HTTPAccountService) CheckAliasAddress(ctx context.Context, alias string) (*models.AliasAddress, error) {
	if as.SS == nil {
		return nil, fmt.Errorf("The storage service cannot be nil")
	}
	logg.InfoCtxf(ctx, "resolving alias before formatting", "alias", alias)
	svc := dev.NewDevAccountService(ctx, as.SS)
	if as.UseApi {
		logg.InfoCtxf(ctx, "resolving alias to address", "alias", alias)
		return resolveAliasAddress(ctx, alias)
	} else {
		return svc.CheckAliasAddress(ctx, alias)
	}
}

func resolveAliasAddress(ctx context.Context, alias string) (*models.AliasAddress, error) {
	var aliasEnsResult models.AliasEnsAddressResult

	fullURL, err := url.JoinPath(config.AliasResolverURL, alias)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &aliasEnsResult)
	if err != nil {
		return nil, err
	}

	return &models.AliasAddress{Address: aliasEnsResult.Address}, nil
}

func (as *HTTPAccountService) FetchTopPools(ctx context.Context) ([]dataserviceapi.PoolDetails, error) {
	svc := dev.NewDevAccountService(ctx, as.SS)
	if as.UseApi {
		return fetchCustodialTopPools(ctx)
	} else {
		return svc.FetchTopPools(ctx)
	}
}

func fetchCustodialTopPools(ctx context.Context) ([]dataserviceapi.PoolDetails, error) {
	var r struct {
		TopPools []dataserviceapi.PoolDetails `json:"topPools"`
	}

	req, err := http.NewRequest("GET", config.TopPoolsURL, nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	return r.TopPools, nil
}

func (as *HTTPAccountService) RetrievePoolDetails(ctx context.Context, sym string) (*dataserviceapi.PoolDetails, error) {
	if as.UseApi {
		return retrievePoolDetails(ctx, sym)
	} else {
		return nil, nil
	}
}

func retrievePoolDetails(ctx context.Context, sym string) (*dataserviceapi.PoolDetails, error) {
	var r struct {
		PoolDetails dataserviceapi.PoolDetails `json:"poolDetails"`
	}

	ep, err := url.JoinPath(config.RetrievePoolDetailsURL, sym)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r.PoolDetails, nil
}

func (as *HTTPAccountService) PoolDeposit(ctx context.Context, amount, from, poolAddress, tokenAddress string) (*models.PoolDepositResult, error) {
	var r models.PoolDepositResult

	//pool deposit payload
	payload := map[string]string{
		"amount":       amount,
		"from":         from,
		"poolAddress":  poolAddress,
		"tokenAddress": tokenAddress,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", config.TokenTransferURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (as *HTTPAccountService) GetPoolSwapQuote(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapQuoteResult, error) {
	var r models.PoolSwapQuoteResult

	//pool swap quote payload
	payload := map[string]string{
		"amount":           amount,
		"from":             from,
		"fromTokenAddress": fromTokenAddress,
		"poolAddress":      poolAddress,
		"toTokenAddress":   toTokenAddress,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", config.PoolSwapQuoteURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (as *HTTPAccountService) GetPoolSwappableFromVouchers(ctx context.Context, poolAddress, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	if as.UseApi {
		return as.getPoolSwappableFromVouchers(ctx, poolAddress, publicKey)
	} else {
		svc := dev.NewDevAccountService(ctx, as.SS)
		return svc.GetPoolSwappableFromVouchers(ctx, poolAddress, publicKey)
	}

}

func (as *HTTPAccountService) getPoolSwappableFromVouchers(ctx context.Context, poolAddress, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	var r struct {
		PoolSwappableVouchers []dataserviceapi.TokenHoldings `json:"filtered"`
	}
	ep, err := url.JoinPath(config.PoolSwappableVouchersURL, poolAddress, "from", publicKey)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)

	return r.PoolSwappableVouchers, nil
}

func (as *HTTPAccountService) GetPoolSwappableVouchers(ctx context.Context, poolAddress string) ([]dataserviceapi.TokenHoldings, error) {
	svc := dev.NewDevAccountService(ctx, as.SS)
	if as.UseApi {
		return as.getPoolSwappableVouchers(ctx, poolAddress)
	} else {
		return svc.GetPoolSwappableVouchers(ctx, poolAddress)
	}
}

func (as HTTPAccountService) getPoolSwappableVouchers(ctx context.Context, poolAddress string) ([]dataserviceapi.TokenHoldings, error) {
	var r struct {
		PoolSwappableVouchers []dataserviceapi.TokenHoldings `json:"filtered"`
	}

	basePath, err := url.JoinPath(config.PoolSwappableVouchersURL, poolAddress, "to")
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}

	query := parsedURL.Query()
	if config.IncludeStablesParam != "" {
		query.Set("stables", config.IncludeStablesParam)
	}
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	_, err = doRequest(ctx, req, &r)
	return r.PoolSwappableVouchers, nil
}

func (as *HTTPAccountService) PoolSwap(ctx context.Context, amount, from, fromTokenAddress, poolAddress, toTokenAddress string) (*models.PoolSwapResult, error) {
	var r models.PoolSwapResult

	//swap payload
	payload := map[string]string{
		"amount":           amount,
		"from":             from,
		"fromTokenAddress": fromTokenAddress,
		"poolAddress":      poolAddress,
		"toTokenAddress":   toTokenAddress,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", config.PoolSwapURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (as *HTTPAccountService) GetSwapFromTokenMaxLimit(ctx context.Context, poolAddress, fromTokenAddress, toTokenAddress, publicKey string) (*models.MaxLimitResult, error) {
	if as.UseApi {
		return as.getSwapFromTokenMaxLimit(ctx, poolAddress, fromTokenAddress, toTokenAddress, publicKey)
	} else {
		svc := dev.NewDevAccountService(ctx, as.SS)
		return svc.GetSwapFromTokenMaxLimit(ctx, poolAddress, fromTokenAddress, toTokenAddress, publicKey)
	}
}

func (as *HTTPAccountService) getSwapFromTokenMaxLimit(ctx context.Context, poolAddress, fromTokenAddress, toTokeAddress, publicKey string) (*models.MaxLimitResult, error) {
	var r models.MaxLimitResult

	ep, err := url.JoinPath(config.PoolSwappableVouchersURL, poolAddress, "limit", fromTokenAddress, toTokeAddress, publicKey)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (as *HTTPAccountService) CheckTokenInPool(ctx context.Context, poolAddress, tokenAddress string) (*models.TokenInPoolResult, error) {
	if as.UseApi {
		return as.checkTokenInPool(ctx, poolAddress, tokenAddress)
	} else {
		svc := dev.NewDevAccountService(ctx, as.SS)
		return svc.CheckTokenInPool(ctx, poolAddress, tokenAddress)
	}
}

func (as *HTTPAccountService) checkTokenInPool(ctx context.Context, poolAddress, tokenAddress string) (*models.TokenInPoolResult, error) {
	var r models.TokenInPoolResult

	ep, err := url.JoinPath(config.PoolSwappableVouchersURL, poolAddress, "check", tokenAddress)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// TODO: Use actual custodial api to request available alias
func (as *HTTPAccountService) RequestAlias(ctx context.Context, publicKey string, hint string) (*models.RequestAliasResult, error) {
	if as.SS == nil {
		return nil, fmt.Errorf("The storage service cannot be nil")
	}
	if as.UseApi {
		if !strings.Contains(hint, ".") {
			hint = as.ToFqdn(hint)
		}
		enr, err := requestEnsAlias(ctx, publicKey, hint)
		if err != nil {
			return nil, err
		}
		return &models.RequestAliasResult{Alias: enr.Name}, nil
	} else {
		svc := dev.NewDevAccountService(ctx, as.SS)
		return svc.RequestAlias(ctx, publicKey, hint)
	}
}

func requestEnsAlias(ctx context.Context, publicKey string, hint string) (*models.AliasEnsResult, error) {
	var r models.AliasEnsResult

	endpoint := config.AliasRegistrationURL

	logg.InfoCtxf(ctx, "requesting alias", "endpoint", endpoint, "hint", hint)
	//Payload with the address and hint to derive an ENS name
	payload := map[string]string{
		"address": publicKey,
		"hint":    hint,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	// Log the request body
	logg.InfoCtxf(ctx, "request body", "payload", string(payloadBytes))
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}
	logg.InfoCtxf(ctx, "alias successfully assigned", "alias", r.Name)
	return &r, nil
}

// SendSMS calls the API to send out an SMS.
// Parameters:
//   - inviterPhone: The user initiating the SMS.
//   - inviteePhone: The number being invited to Sarafu.
func (as *HTTPAccountService) SendUpsellSMS(ctx context.Context, inviterPhone, inviteePhone string) (*models.SendSMSResponse, error) {
	var r models.SendSMSResponse

	// Create request payload
	payload := map[string]string{
		"inviterPhone": inviterPhone,
		"inviteePhone": inviteePhone,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create a new request
	req, err := http.NewRequest("POST", config.SendSMSURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	_, err = doRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (as *HTTPAccountService) SendAddressSMS(ctx context.Context, publicKey, originPhone string) error {
	ep, err := url.JoinPath(config.ExternalSMSURL, "address")
	if err != nil {
		return err
	}
	logg.InfoCtxf(ctx, "sending an address sms", "endpoint", ep, "address", publicKey, "origin-phone", originPhone)
	payload := map[string]string{
		"address":     publicKey,
		"originPhone": originPhone,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", ep, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	_, err = doRequest(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (as *HTTPAccountService) SendPINResetSMS(ctx context.Context, admin, phone string) error {
	ep, err := url.JoinPath(config.ExternalSMSURL, "pinreset")
	if err != nil {
		return err
	}
	logg.InfoCtxf(ctx, "sending pin reset sms", "endpoint", ep, "admin", admin, "phone", phone)
	payload := map[string]string{
		"admin": admin,
		"phone": phone,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", ep, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	_, err = doRequest(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}

// TODO: remove eth-custodial api dependency
func doRequest(ctx context.Context, req *http.Request, rcpt any) (*api.OKResponse, error) {
	var okResponse api.OKResponse
	var errResponse api.ErrResponse

	req.Header.Set("Authorization", "Bearer "+config.BearerToken)
	req.Header.Set("Content-Type", "application/json")

	// Log request
	logRequestDetails(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to make %s request to endpoint: %s with reason: %s", req.Method, req.URL, err.Error())
		errResponse.Description = err.Error()
		return nil, err
	}
	defer resp.Body.Close()

	// Read and log response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Received response for %s: Status Code: %d | Content-Type: %s | Body: %s",
		req.URL, resp.StatusCode, resp.Header.Get("Content-Type"), string(body))

	if resp.StatusCode >= http.StatusBadRequest {
		if err := json.Unmarshal(body, &errResponse); err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.Description)
	}

	if err := json.Unmarshal(body, &okResponse); err != nil {
		return nil, err
	}

	if len(okResponse.Result) == 0 {
		return nil, errors.New("Empty api result")
	}

	v, err := json.Marshal(okResponse.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(v, &rcpt)
	return &okResponse, err
}

func logRequestDetails(req *http.Request) {
	var bodyBytes []byte
	contentType := req.Header.Get("Content-Type")

	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
	} else {
		bodyBytes = []byte("-")
	}

	log.Printf("Outgoing Request -> URL: %s | Method: %s | Content-Type: %s | Body: %s",
		req.URL.String(), req.Method, contentType, string(bodyBytes))
}
