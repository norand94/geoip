package config

import "time"

type Config struct {
	HttpPort string `json:"httpPort"`
	Providers []Provider `json:"providers"`
	DB        Database   `json:"db"`
}

type Database struct {
	DbNum   string `json:"dbNum"`
	Address string `json:"address"`
}

type Provider struct {
	Name          string        `json:"name"`
	Url           string        `json:"url"`
	Period        time.Duration `json:"period"`
	LimitReqCount int           `json:"limitReqCount"`
}
