package core

import (
	"github.com/garyburd/redigo/redis"
	"github.com/norand94/geoip/core/config"
	"net/http"
	"github.com/norand94/geoip/core/api"
	"log"
)

type app struct {
	Conf  *config.Config
	RConn redis.Conn
	apiChans api.Chans
}

func NewApp(conf *config.Config) *app {
	app := new(app)
	app.Conf = conf
	return app
}

func (a *app) Run() {
	conn, err := redis.Dial("tcp", a.Conf.DB.Address)
	if err != nil {
		log.Fatalln("Не удалось подключиться к Redis")
		panic(err)
	}
	defer conn.Close()
	a.RConn = conn

	api := api.New(a.Conf, conn)
	a.apiChans = api.Start()

	http.HandleFunc("/my", a.myHandler)
	log.Println("geoip started")
	log.Fatalln(http.ListenAndServe(a.Conf.HttpPort, nil))
}
