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

func main() {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Out = os.Stdout
	log.Level = logrus.DebugLevel

	log.Println("Starting health-check service")
	log.Println("Bootstrapping providers")
	err := vpn.InitAllProviders()
	if err != nil {
		log.Fatal(err)
	}

	hs := &HealthService{}
	hs.Start()

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", homeHandler)
	log.Fatal(http.ListenAndServe(listeningPort, r))
}
