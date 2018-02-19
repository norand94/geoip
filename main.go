package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum-1.8.0/log"
	"github.com/norand94/geoip/core"
	"github.com/norand94/geoip/core/config"
	"io/ioutil"
)

func main() {
	bts, err := ioutil.ReadFile("geoip_config.json")
	if err != nil {
		panic(err)
	}
	conf := new(config.Config)
	json.Unmarshal(bts, conf)
	fmt.Printf("Config: \n %+v \n", conf)

	if len(conf.Providers) == 0 {
		log.Error("Не указаны провайдеры!")
		return
	}

	app := core.NewApp(conf)
	app.Run()
}
