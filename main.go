package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	maxAttempts                = 5
	healthCheckTimeout         = 1 * time.Second
	defaultHealthCheckInterval = 10 * time.Second
	defaultLbType              = "round-robin"
)

type loadBalancer struct {
	backendPool         backendPool
	numOfBackends       int
	port                int
	healthCheckInterval time.Duration
	healthCheckEnabled  bool
}

func (lb *loadBalancer) checkHealth() {
	log.Println("Checking health of backend servers...")
	for _, b := range lb.backendPool.getAllBackends() {
		go b.checkHealth()
	}
}

func (lb *loadBalancer) getNextBackend() *Backend {
	tries, maxTries := 0, 2*lb.numOfBackends
	curBackend := lb.backendPool.getNext()

	for !curBackend.isAlive() {
		curBackend = lb.backendPool.getNext()
		tries++
		if tries == maxTries {
			return nil
		}
	}
	return curBackend
}

func createLb(configPath string) (*loadBalancer, error) {
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	type backendConf struct {
		URL    string `yaml:"url"`
		Weight int    `yaml:"weight"`
	}
	var lbConf struct {
		Port                int           `yaml:"port"`
		Type                string        `yaml:"type"`
		HealthCheckInterval int           `yaml:"health-check-interval"`
		Backends            []backendConf `yaml:"backends"`
	}
	if err := yaml.Unmarshal(bytes, &lbConf); err != nil {
		return nil, errors.New("invalid config file")
	}

	if len(lbConf.Backends) == 0 {
		return nil, errors.New("no backend specified")
	}

	var backends []*Backend
	isWeighted := false
	for _, bconf := range lbConf.Backends {
		weight := 1
		if bconf.Weight > 0 {
			isWeighted = true
			weight = bconf.Weight
		} else if bconf.Weight < 0 {
			return nil, errors.New("invalid negative weight")
		}
		backend, err := newBackend(bconf.URL, weight)
		if err != nil {
			return nil, err
		}
		backends = append(backends, backend)
	}
	if lbConf.Type == "" {
		lbConf.Type = defaultLbType
	}

	lb := &loadBalancer{
		port:               lbConf.Port,
		numOfBackends:      len(backends),
		healthCheckEnabled: true,
	}
	switch interval := lbConf.HealthCheckInterval; {
	case interval > 0:
		lb.healthCheckInterval = time.Duration(interval) * time.Second
	case interval < 0:
		lb.healthCheckEnabled = false
	default:
		lb.healthCheckInterval = defaultHealthCheckInterval
	}

	switch lbConf.Type {
	case "round-robin":
		if isWeighted {
			lb.backendPool = newWeightedRoundRobinPool(backends)
		} else {
			lb.backendPool = newRoundRobinPool(backends)
		}
	case "least-conn":
		lb.backendPool = newLeastConnPool(backends)
	case "random":
		lb.backendPool = newRandomPool(backends)
	default:
		return nil, errors.New("invalid load balancer type")
	}

	return lb, nil
}

func sendBadGateway(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("Bad gateway\n"))
}

var configFlag = flag.String("config", "config.yaml", "location to config file in yaml format")

func main() {
	flag.Parse()
	configPath := *configFlag
	if _, err := os.Stat(configPath); err != nil {
		panic(err)
	}

	lb, err := createLb(configPath)
	if err != nil {
		panic(err)
	}

	var t *time.Ticker
	if lb.healthCheckEnabled {
		t = time.NewTicker(lb.healthCheckInterval)
		go func() {
			lb.checkHealth()
			for range t.C {
				lb.checkHealth()
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := 0
		for i < maxAttempts {
			backend := lb.getNextBackend()
			if backend == nil {
				sendBadGateway(w)
				return
			}
			backend.ServeHTTP(w, r)
			if backend.isAlive() {
				log.Println("Forwarded to ", backend.URL.String())
				break
			}
			i++
		}
		if i == maxAttempts {
			sendBadGateway(w)
		}
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", lb.port),
		Handler: mux,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\nClosing load balancer")
		if t != nil {
			t.Stop()
		}
		server.Close()
	}()

	log.Println("load balancer starting at port", lb.port)
	log.Fatal(server.ListenAndServe())
}
