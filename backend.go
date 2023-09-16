package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL       *url.URL
	isAlive_  bool
	proxy     *httputil.ReverseProxy
	mu        *sync.Mutex
	weight    int
	curWeight int
	load      int64
}

func NewBackend(urlStr string) (*Backend, error) {
	URL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	backend := &Backend{
		URL:       URL,
		isAlive_:  true,
		mu:        &sync.Mutex{},
		weight:    1,
		curWeight: 1,
	}
	revProxy := httputil.NewSingleHostReverseProxy(URL)
	revProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		backend.setAlive(false)
	}
	backend.proxy = revProxy
	return backend, nil
}

func (b *Backend) isAlive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.isAlive_
}

func (b *Backend) setAlive(isAlive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isAlive_ = isAlive
}

func (b *Backend) checkHealth() {
	conn, err := net.DialTimeout("tcp", b.URL.Host, healthCheckTimeout)
	if err != nil {
		log.Println(err)
		b.setAlive(false)
		return
	}
	defer conn.Close()
	b.setAlive(true)
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&b.load, 1)
	defer atomic.AddInt64(&b.load, -1)
	b.proxy.ServeHTTP(w, r)
}
