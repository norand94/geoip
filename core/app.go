package core

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/norand94/geoip/core/api"
	"github.com/norand94/geoip/core/config"
	"log"
	"net/http"
)

type app struct {
	Conf     *config.Config
	RConn    redis.Conn
	apiChans api.Chans
}

func NewApp(conf *config.Config) *app {
	app := new(app)
	app.Conf = conf
	return app
}

// Run - Инициализирует модули и запускает приложение
func (a *app) Run() {
	conn, err := redis.Dial("tcp", a.Conf.DB.Address)
	if err != nil {
		log.Fatalln("Не удалось подключиться к Redis")
		panic(err)
	}
	defer conn.Close()
	a.RConn = conn

	//инициализация api-сервиса
	api := api.GetInstance(a.Conf, conn)
	a.apiChans = api.Start()

	r := gin.Default()

	r.GET("/myip", a.myIpHandler)
	r.GET("/ip/:ip", a.ipHandler)
	r.GET("/provStat", a.provStatHandler)

	http.Handle("/", r)

	log.Println("geoip started")
	log.Fatalln(http.ListenAndServe(a.Conf.HttpPort, nil))
}
