package api

import (
	"encoding/json"
	"fmt"
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

type service struct {
	Conf        *config.Config
	RConn       redis.Conn
	chans       Chans
	provs       []*provider
	currProv    *provider
	currProvNum uint
}

func New(conf *config.Config, conn redis.Conn) *service {
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

type provider struct {
	limitReq       int
	name           string
	currReqCounter int
	period         time.Duration
	apiUrl         string
}

type Chans struct {
	ProvReq ProvReq
	ReqCh   chan Request
	QuitCh  chan struct{}
}

type Request struct {
	Ip         string
	ResponseCh chan Response
}

type Response struct {
	City   string `json:"city"`
	ReqNum int    `json:"reqNum"`
	Source string `json:"source"`
	Error  error  `json:"error"`
}

type ProvReq chan chan ProvStats

type ProvStats struct {
	CurrProvName string
	Stats        []ProvStat
}

type ProvStat struct {
	Name           string
	CurrReqCounter int
}

type httpResp struct {
	City string `json:"city"`
}

func (api *service) Start() Chans {
	ticker := time.After(time.Duration(api.Conf.PeriodMin) * time.Minute)
	go func(reqCh <-chan Request, quitCh <-chan struct{}) {
		for {
			select {
			case <-ticker:
				for i := range api.provs {
					api.provs[i].currReqCounter = 0
				}

			case out := <-api.chans.ProvReq:
				st := ProvStats{CurrProvName: api.currProv.name}
				for _, v := range api.provs {
					st.Stats = append(st.Stats, ProvStat{
						Name:           v.name,
						CurrReqCounter: v.currReqCounter,
					})
				}
				out <- st

			case req := <-reqCh:
				api.currProv.currReqCounter++
				if api.currProv.currReqCounter >= api.currProv.limitReq {
					api.nextProvider()
				}
				go api.processReq(req, api.currProv)

			case <-quitCh:
				return
			}
		}
	}(api.chans.ReqCh, api.chans.QuitCh)

	return api.chans
}

func (api *service) nextProvider() {
	api.currProv.currReqCounter = 0
	api.currProv = api.provs[(api.currProvNum+1)%uint(len(api.provs))]
}

func (api *service) processReq(req Request, prov *provider) {
	url := strings.Replace(prov.apiUrl, "{ip}", req.Ip, 1)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		req.ResponseCh <- Response{Error: err, ReqNum: prov.currReqCounter, Source: SourceApi}
		return
	}
	defer resp.Body.Close()

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		req.ResponseCh <- Response{Error: err, ReqNum: prov.currReqCounter, Source: SourceApi}
		return
	}
	key := "ip:" + req.Ip
	api.RConn.Do("HSET", key, "resp", bts)

	r := httpResp{}
	err = json.Unmarshal(bts, &r)
	if err != nil {
		req.ResponseCh <- Response{Error: err, ReqNum: prov.currReqCounter, Source: SourceApi}
		return
	}

	_, rerr := api.RConn.Do("HSET", key, "city", r.City)
	_, rerr = api.RConn.Do("EXPIRE", key, api.Conf.ExpiredIpInfoSec)
	if rerr != nil {
		log.Println("Redis error: ", err)
	}

	req.ResponseCh <- Response{City: r.City, ReqNum: prov.currReqCounter, Source: SourceApi}
}
