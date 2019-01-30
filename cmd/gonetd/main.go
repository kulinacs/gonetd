package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/kulinacs/gonetd/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var (
	config            service.Config
	activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gonetd_active_connections",
		Help: "Number of connections currently active",
	})
)

func updateActiveConnections() {
	go func() {
		for {
			var connCount int64
			connCount = 0
			for _, serv := range config.Service {
				connCount += serv.ActiveConnections()
			}
			activeConnections.Set(float64(connCount))
			time.Sleep(2 * time.Second)
		}
	}()
}

func init() {
	configPath := flag.String("config", "./config.toml", "config file")
	flag.Parse()
	_, err := toml.DecodeFile(*configPath, &config)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("failed to load config")
	}
	prometheus.MustRegister(activeConnections)
}

func main() {
	updateActiveConnections()
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
	var wg sync.WaitGroup
	for _, serv := range config.Service {
		wg.Add(1)
		go serv.Handle(&wg)
	}
	wg.Wait()
}
