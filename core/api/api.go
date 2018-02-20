package api

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/norand94/geoip/core/config"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const SourceApi = "api"
const SourceCache = "cache"

var instalnce *service = nil

//Сервис, отправляющий запросы к ip
type service struct {
	Conf        *config.Config
	RConn       redis.Conn
	chans       Chans
	provs       []*provider
	currProv    *provider
	currProvNum uint
}

func GetInstance(conf *config.Config, conn redis.Conn) *service {
	if instalnce == nil {
		instalnce = newService(conf, conn)
	}
	return instalnce
}

func newService(conf *config.Config, conn redis.Conn) *service {
	api := new(service)
	api.Conf = conf
	api.RConn = conn

	api.chans.ProvReq = make(ProvReq)
	api.chans.ReqCh = make(chan Request, 10)
	api.chans.QuitCh = make(chan struct{}, 1)

	for _, v := range conf.Providers {
		api.provs = append(api.provs, &provider{
			limitReq: v.LimitReqCount,
			apiUrl:   v.ApiUrl,
			name:     v.Name,
		})
	}

	api.currProv = api.provs[0]

	return api
}

// Провайдер geoip
type provider struct {
	limitReq       int
	name           string
	currReqCounter int    //Счетчик запросов к нему за данный период
	apiUrl         string //url, по которому можно получить информацию о ip
}

//Chains - каналы, по которым приложение общается с api-сервисом
type Chans struct {
	ProvReq ProvReq       //Запрос на получение информации о провайдерах
	ReqCh   chan Request  //Запрос на получение информации о ip
	QuitCh  chan struct{} // Запрос на остановку сервиса
}

type Request struct {
	Ip         string
	ResponseCh chan Response
}

type Response struct {
	City   interface{} `json:"city"`
	ReqNum int         `json:"reqNum"`
	Source string      `json:"source"`
	Error  error       `json:"error"`
}

//Канал запроса на получение информации о провайдерах
type ProvReq chan chan ProvStats

type ProvStats struct {
	CurrProvName string
	Stats        []ProvStat
}

type ProvStat struct {
	Name           string
	CurrReqCounter int
}

// Start - запускает сервис
func (api *service) Start() Chans {
	ticker := time.Tick(time.Duration(api.Conf.PeriodSec) * time.Second)
	go func(reqCh <-chan Request, quitCh <-chan struct{}) {
		for {
			select {

			// Событие <-ticker - обнуляет статистику по провайдерам
			case <-ticker:
				for i := range api.provs {
					api.provs[i].currReqCounter = 0
				}

			// out := <-api.chans.ProvReq - Отдает статистику по провайдерам
			case out := <-api.chans.ProvReq:
				func() {
					defer close(out)
					st := ProvStats{CurrProvName: api.currProv.name}
					for _, v := range api.provs {
						st.Stats = append(st.Stats, ProvStat{
							Name:           v.name,
							CurrReqCounter: v.currReqCounter,
						})
					}
					out <- st
				}()

			// req := <-reqCh - Отдает информацию об ip
			case req := <-reqCh:
				api.currProv.currReqCounter++
				if api.currProv.currReqCounter >= api.currProv.limitReq {
					api.nextProvider()
				}
				go api.processReq(&req, api.currProv)

			case <-quitCh:
				return
			}
		}
	}(api.chans.ReqCh, api.chans.QuitCh)

	return api.chans
}

func (api *service) nextProvider() {
	if api.Conf.ResetPrevProvider {
		api.currProv.currReqCounter = 0
	}

	api.currProvNum++
	if api.currProvNum >= uint(len(api.provs)) {
		api.currProvNum = 0
	}
	api.currProv = api.provs[api.currProvNum]
}

// processReq - обрабатывает запрос от приложения
// Отправляет запрос на внешние geoip сервисы, разбирает их ответ,
// кладет его в кеш и отправляет назад в response информацию об ip
func (api *service) processReq(req *Request, prov *provider) {
	defer close(req.ResponseCh)
	url := strings.Replace(prov.apiUrl, "{ip}", req.Ip, 1)
	resp, err := http.Get(url)
	if err != nil {
		req.ResponseCh <- Response{Error: err, ReqNum: prov.currReqCounter, Source: prov.name}
		return
	}
	defer resp.Body.Close()

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		req.ResponseCh <- Response{Error: err, ReqNum: prov.currReqCounter, Source: prov.name}
		return
	}
	key := "ip:" + req.Ip
	api.RConn.Do("HSET", key, "resp", bts)

	r := struct {
		City interface{} `json:"city"`
	}{}
	err = json.Unmarshal(bts, &r)
	if err != nil {
		req.ResponseCh <- Response{Error: err, ReqNum: prov.currReqCounter, Source: prov.name}
		return
	}

	_, rerr := api.RConn.Do("HSET", key, "city", r.City)
	_, rerr = api.RConn.Do("EXPIRE", key, api.Conf.ExpiredIpInfoSec)
	if rerr != nil {
		log.Println("Redis error: ", err)
	}

	req.ResponseCh <- Response{City: r.City, ReqNum: prov.currReqCounter, Source: prov.name}
}
