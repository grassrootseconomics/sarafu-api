package models


type Pool struct {
	PoolName            string `json:"poolName"`
	PoolSymbol          string `json:"poolSymbol"`
	PoolContractAddress string `json:"poolContractAddress"`
	LimiterAddress      string `json:"limiterAddress"`
	VoucherRegistry     string `json:"voucherRegistry"`
}
