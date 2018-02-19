package main

import (
	"encoding/json"
	"github.com/norand94/geoip/core/config"
	"io/ioutil"
	"github.com/norand94/geoip/core"
)

func main() {
	bts, err := ioutil.ReadFile("geoip_config.json")
	if err != nil {
		panic(err)
	}
	conf := new(config.Config)
	json.Unmarshal(bts, conf)

	app := core.NewApp(conf)
	app.Run()
}
