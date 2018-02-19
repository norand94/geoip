package config

type Config struct {
	HttpPort         string `json:"httpPort"`
	ExpiredIpInfoSec string `json:"expiredIpInfoSec"`
	PeriodMin        int64  `json:"periodMin"`

	Providers []Provider `json:"providers"`
	DB        Database   `json:"db"`
}

type Database struct {
	DbNum   string `json:"dbNum"`
	Address string `json:"address"`
}

type Provider struct {
	Name          string `json:"name"`
	ApiUrl        string `json:"apiUrl"`
	LimitReqCount int    `json:"limitReqCount"`
}
