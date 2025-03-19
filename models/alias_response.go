package models

type RequestAliasResult struct {
	Alias string
}

type AliasAddress struct {
	Address string
}

type AliasEnsResult struct {
	Address    string `json:"address"`
	AutoChoose bool   `json:"autoChoose"`
	Name       string `json:"name"`
}

type AliasEnsAddressResult struct {
	Address string `json:"address"`
}
