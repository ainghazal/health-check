package health

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/netip"
	"sort"
	"text/template"

	"github.com/gorilla/mux"
)

type httpHandler func(http.ResponseWriter, *http.Request)

type result struct {
	Action string `json:"action"`
	Result bool   `json:"result"`
}

func HealthQueryHandlerJSON(phs map[string]*HealthService, provider string) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		hs, ok := phs[provider]
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}
		addrStr := getParam("addr", r)
		transport := getParam("tr", r)
		if addrStr == "" {
			http.Error(w, "missing param", http.StatusBadRequest)
			return
		}
		addr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addrStr))
		isHealthy, err := hs.Healthy(addr, transport)
		if err != nil {
			log.Println("ERR", err.Error())
			http.Error(w, "dunno rick", http.StatusInternalServerError)
			return
		}
		res := []*result{&result{
			Action: "healthy",
			Result: isHealthy,
		},
		}
		json.NewEncoder(w).Encode(res)
	}
}

var healthViewTemplate = `{{ range . }}{{ .Addr.String }}/{{ .Transport }}: {{ .Healthy }}
{{ end }}`

// PoC for a general view of status (text/plain)
func HealthSummaryHandlerText(phs map[string]*HealthService, provider string) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		hs, ok := phs[provider]
		if hs.Current == nil {
			http.Error(w, "try again later", http.StatusServiceUnavailable)
			return
		}
		mm := hs.Current.Measurements()
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}
		t := template.New("view")
		t, err := t.Parse(healthViewTemplate)
		if err != nil {
			log.Println(err)
			return
		}
		buf := &bytes.Buffer{}
		sort.Slice(mm, func(i, j int) bool { return mm[i].String() > mm[j].String() })
		err = t.Execute(buf, mm)
		if err != nil {
			log.Println(err)
			return
		}
		w.Write(buf.Bytes())
	}
}

// User facing status (html) (TODO)
func HealthQueryHandlerHTML(phs map[string]*HealthService, provider string) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		hs, ok := phs[provider]
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}
		// ...
		addrStr := getQueryParam("addr", r)
		transport := getQueryParam("tr", r)
		if addrStr == "" {
			http.Error(w, "missing param", http.StatusBadRequest)
			return
		}
		addr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addrStr))
		isHealthy, err := hs.Healthy(addr, transport)
		if err != nil {
			log.Println("ERR", err.Error())
			http.Error(w, "dunno rick", http.StatusInternalServerError)
			return
		}
		res := []*result{&result{
			Action: "healthy",
			Result: isHealthy,
		},
		}
		json.NewEncoder(w).Encode(res)
	}
}

func getParam(param string, r *http.Request) string {
	return mux.Vars(r)[param]
}

func getQueryParam(param string, r *http.Request) string {
	return r.URL.Query().Get(param)
}
