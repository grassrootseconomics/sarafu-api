package models

type MpesaOnrampResponse struct {
	Message         string `json:"message"`
	Status          string `json:"status"`
	TransactionCode string `json:"transactionCode"`
}

type MpesaOnrampRatesResponse struct {
	Buy  float64 `json:"buy"`
	Sell float64 `json:"sell"`
}
