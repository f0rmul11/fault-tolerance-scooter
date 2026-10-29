package dao

type Tariff struct {
	Plan         string `json:"plan"`
	RateMinor    int64  `json:"rate_minor"`
	AdvanceMinor int64  `json:"advance_minor"`
}
