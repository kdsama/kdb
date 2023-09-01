package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	cd "github.com/kdsama/kdb/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var ()

func main() {

	logger := setupLogger()
	clh := cd.NewClientHandler(logger)

	http.HandleFunc("/add-server", clh.AddServer)
	// to get data from leader
	http.HandleFunc("/get", clh.Get)
	// To get data from random spawn server
	http.HandleFunc("/get-random", clh.GetRandom)
	// add key valye
	http.HandleFunc("/set", clh.Set)
	// add key valye
	http.HandleFunc("/leader", clh.Leader)
	// automated get from key
	http.HandleFunc("/automate-get", clh.AutomateGet)
	// automated set (key,value)
	http.HandleFunc("/automate-set", clh.AutomateSet)
	// exposed for metrics monitoring by prometheus
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func setupLogger() *zap.SugaredLogger {

	l, _ := zap.NewDevelopment()
	return l.Sugar()
}
