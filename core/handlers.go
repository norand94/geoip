package core

import (
	"fmt"
	"net/http"
	"strings"
	"github.com/norand94/geoip/core/api"
)

func (a *app) myHandler(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	fmt.Println(ip)
	respChan := make(chan api.Response, 1)
	a.apiChans.ReqCh <- api.Request{
		Ip : ip,
		ResponseCh: respChan,
	}
	fmt.Fprintf(w, "%+v", <-respChan)
}
