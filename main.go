package main

import (
	"net/http"
	"os"

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

	for name, provider := range vpn.Providers {
		if hasItem(enabledProviders, name) {
			hs := &HealthService{
				Name: name,
				Checker: &VPNChecker{
					Provider: provider,
				},
			}
			hs.Start()
		}
	}

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", homeHandler)
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
