package models

type PoolDepositResult struct {
	TrackingId string `json:"trackingId"`
}

type PoolSwapQuoteResult struct {
	IncludesFeesDeduction bool   `json:"includesFeesDeduction"`
	OutValue              string `json:"outValue"`
}

type PoolSwapResult struct {
	TrackingId string `json:"trackingId"`
}

type MaxLimitResult struct {
	Max string `json:"max"`
}

type TokenInPoolResult struct {
	CanSwapFrom bool `json:"canSwapFrom"`
}

type CreditSendLimitsResult struct {
	MaxRAT string `json:"maxRAT"`
	MaxSAT string `json:"maxSAT"`
}

type CreditSendReverseQouteResult struct {
	InputAmount  string `json:"inputAmount"`
	OutputAmount string `json:"outputAmount"`
}
