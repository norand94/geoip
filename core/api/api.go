package api

import (
	"github.com/garyburd/redigo/redis"
	"github.com/norand94/geoip/core/config"
)

type service struct {
	Conf  *config.Config
	RConn redis.Conn
	chans Chans
}

func New(conf *config.Config, conn redis.Conn) *service {
	api := new(service)
	api.Conf = conf
	api.RConn = conn

	api.chans.ReqCh = make(chan Request, 10)
	api.chans.QuitCh = make(chan struct{}, 1)
	return api
}

type Chans struct {
	ReqCh  chan Request
	QuitCh chan struct{}
}

type Request struct {
	Ip         string
	ResponseCh chan Response
}

type Response struct {
	Sity   string
	ReqNum int
	Error  error
}


func (api *service) Start() Chans {
	go func(reqCh <-chan Request, quitCh <-chan struct{}) {
		for {
			select {
			case req := <-reqCh:
				sityBts, err := api.RConn.Do("HGET", "ip:" + req.Ip, "sity")
				if err != nil {
					req.ResponseCh <- Response{Error: err}
				}
				req.ResponseCh <- Response{Sity: string(sityBts.([]byte))}

			case <-quitCh:
				return
			}
		}
	}(api.chans.ReqCh, api.chans.QuitCh)

	return api.chans
}
