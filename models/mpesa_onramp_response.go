package models

type MpesaOnrampResponse struct {
	Message         string `json:"message"`
	Status          string `json:"status"`
	TransactionCode string `json:"transactionCode"`
}
