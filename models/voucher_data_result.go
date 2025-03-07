package models

type VoucherDataResult struct {
	TokenName      string `json:"tokenName"`
	TokenSymbol    string `json:"tokenSymbol"`
	TokenDecimals  int    `json:"tokenDecimals"`
	SinkAddress    string `json:"sinkAddress"`
	TokenCommodity string `json:"tokenCommodity"`
	TokenLocation  string `json:"tokenLocation"`
}

type SwappableVoucher struct {
	ContractAddress string `json:"contractAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimals   string `json:"tokenDecimals"`
	Balance         string `json:"balance"`
}
