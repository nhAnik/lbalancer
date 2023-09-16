package main

import (
	"math/rand"
	"sync"
)

type backendPool interface {
	getNext() *Backend
	getAllBackends() []*Backend
}

type backendList []*Backend

func (bl backendList) getAllBackends() []*Backend {
	return bl
}

type roundRobinPool struct {
	backendList
	curIdx int
	mu     *sync.Mutex
}

func newRoundRobinPool(backends []*Backend) *roundRobinPool {
	return &roundRobinPool{
		mu:          &sync.Mutex{},
		backendList: backends,
		curIdx:      rand.Int() % len(backends),
	}
}

func (rp *roundRobinPool) getNext() *Backend {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	curBackend := rp.backendList[rp.curIdx]
	rp.curIdx = (rp.curIdx + 1) % len(rp.backendList)
	return curBackend
}

type weightedRoundRobinPool struct {
	backendList
	mu *sync.Mutex
}

func newWeightedRoundRobinPool(backends []*Backend) *weightedRoundRobinPool {
	return &weightedRoundRobinPool{
		mu:          &sync.Mutex{},
		backendList: backends,
	}
}

func (wp *weightedRoundRobinPool) getNext() (next *Backend) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	next, wp.backendList = getWeightedNext(wp.backendList)
	return
}

type leastConnPool struct {
	backendList
	mu *sync.Mutex
}

func newLeastConnPool(backends []*Backend) *leastConnPool {
	return &leastConnPool{
		mu:          &sync.Mutex{},
		backendList: backends,
	}
}

func (lp *leastConnPool) getNext() (next *Backend) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	var minLoad int64
	for i := 1; i < len(lp.backendList); i++ {
		if lp.backendList[i].load < minLoad {
			minLoad = lp.backendList[i].load
		}
	}
	var minBackends []*Backend
	var minIndexes []int
	for i := 0; i < len(lp.backendList); i++ {
		if lp.backendList[i].load == minLoad {
			minBackends = append(minBackends, lp.backendList[i])
			minIndexes = append(minIndexes, i)
		}
	}
	if len(minBackends) == 1 {
		return minBackends[0]
	}

	next, minBackends = getWeightedNext(minBackends)
	for i, idx := range minIndexes {
		lp.backendList[idx] = minBackends[i]
	}
	return next
}

func getWeightedNext(backends []*Backend) (*Backend, []*Backend) {
	numOfBackends := len(backends)
	selIdx := 0
	for i := 1; i < numOfBackends; i++ {
		if backends[i].curWeight > backends[selIdx].curWeight {
			selIdx = i
		}
	}
	weightInc := 0
	for i := 0; i < numOfBackends; i++ {
		if i == selIdx {
			continue
		}
		weightInc += backends[i].weight
		backends[i].curWeight += backends[i].weight
	}
	backends[selIdx].curWeight -= weightInc
	return backends[selIdx], backends
}
