package core

import (
	"github.com/gin-gonic/gin"
	"github.com/norand94/geoip/core/api"
	"log"
	"strings"
)

//myIpHandler - получает ip пользователя, что обратился к нему и возвращает город
func (a *app) myIpHandler(c *gin.Context) {
	ip := strings.Split(c.Request.RemoteAddr, ":")[0]
	resp := a.processIp(ip)

	if resp.Error != nil {
		log.Println(resp.Error.Error())
		c.JSON(500, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, resp)
}

//ipHandler - возвращает город по полученному ip
func (a *app) ipHandler(c *gin.Context) {
	ip := c.Param("ip")
	resp := a.processIp(ip)

	if resp.Error != nil {
		log.Println(resp.Error.Error())
		c.JSON(500, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(200, resp)
}

//provStatHandler - возвращает статистику по провайдерам
func (a *app) provStatHandler(c *gin.Context) {
	provCh := make(chan api.ProvStats)
	a.apiChans.ProvReq <- provCh
	c.JSON(200, <-provCh)
}

//processIp - обрабатывает запрос ip, возвращает информацию о нем
func (a *app) processIp(ip string) api.Response {
	//Получение информации из кеша
	cityBts, err := a.RConn.Do("HGET", "ip:"+ip, "city")
	if err != nil {
		log.Println(err.Error())
		return a.getResponseFromApi(ip)
	}

	if cityBts == nil {
		return a.getResponseFromApi(ip)

	}
	return api.Response{City: string(cityBts.([]byte)), Source: api.SourceCache}
}

func (a *app) getResponseFromApi(ip string) api.Response {
	respChan := make(chan api.Response, 1)
	a.apiChans.ReqCh <- api.Request{
		Ip:         ip,
		ResponseCh: respChan,
	}
	return <-respChan
}
