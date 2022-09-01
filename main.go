package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"sort"
	"text/template"

	"github.com/ainghazal/torii/vpn"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	listeningPort = ":8081"
	msgHomeStr    = "nothing to see here"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(msgHomeStr))
}

var healthServiceMap = make(map[string]*HealthService)

var enabledProviders = []string{"riseup"}

type httpHandler func(http.ResponseWriter, *http.Request)

type result struct {
	Action string `json:"action"`
	Result bool   `json:"result"`
}

func healthQueryHandlerJSON(phs map[string]*HealthService, provider string) httpHandler {
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
func healthSummaryHandlerText(phs map[string]*HealthService, provider string) httpHandler {
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

// User facing status (html)
func healthQueryHandlerHTML(phs map[string]*HealthService, provider string) httpHandler {
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

func main() {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Out = os.Stdout
	log.Level = logrus.DebugLevel

	if os.Getenv("SKIP_INIT") != "1" {
		err := vpn.InitAllProviders()
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Starting health-check service")
	log.Println("Bootstrapping providers")

	phs := make(map[string]*HealthService)
	for name, provider := range vpn.Providers {
		if hasItem(enabledProviders, name) {
			hs := &HealthService{
				Name: name,
				Checker: &VPNChecker{
					Provider: provider,
				},
			}
			hs.Start()
			phs[name] = hs
		}
	}

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/riseup/status/json", healthQueryHandlerJSON(phs, "riseup")).Queries("addr", "{addr}").Queries("tr", "{tr}")
	r.HandleFunc("/riseup/summary", healthSummaryHandlerText(phs, "riseup"))
	r.HandleFunc("/riseup/status", healthQueryHandlerHTML(phs, "riseup"))
	log.Fatal(http.ListenAndServe(listeningPort, r))
}

func hasItem(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
