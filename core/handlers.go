package core

import (
	"github.com/ethereum/go-ethereum-1.8.0/log"

	"github.com/norand94/geoip/core/api"
	"strings"
	"github.com/gin-gonic/gin"

)

func (a *app) myIpHandler(c *gin.Context) {
	ip := strings.Split(c.Request.RemoteAddr, ":")[0]
	resp := a.processIp(ip)

	if resp.Error != nil {
		log.Error(resp.Error.Error())
		c.JSON(500, gin.H{
			"error" : "internal server error",
		})
	}

	c.JSON(200, resp)
}

func (a *app) ipHandler(c *gin.Context) {
	ip := c.Param("ip")
	resp := a.processIp(ip)

	if resp.Error != nil {
		log.Error(resp.Error.Error())
		c.JSON(500, gin.H{
			"error" : "internal server error",
		})
	}

	c.JSON(200, resp)
}

func (a *app) processIp(ip string) api.Response {

	cityBts, err := a.RConn.Do("HGET", "ip:"+ip, "sity")
	if err != nil {
		log.Error(err.Error())
		return a.getResponseFromApi(ip)
	}

	if cityBts == nil {
		return a.getResponseFromApi(ip)

	}
	return api.Response{City: string(cityBts.([]byte)), Source: api.SourceCache}
}

func (a *app) getResponseFromApi(ip string) api.Response  {
	respChan := make(chan api.Response, 1)
	a.apiChans.ReqCh <- api.Request{
		Ip:         ip,
		ResponseCh: respChan,
	}
	return <-respChan
}