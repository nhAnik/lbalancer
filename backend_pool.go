package main

import (
	"math/rand"
	"sort"
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
	accWeights []int
	mu         *sync.Mutex
}

func newWeightedRoundRobinPool(backends []*Backend) *weightedRoundRobinPool {
	numOfBackends := len(backends)
	wrrp := &weightedRoundRobinPool{
		mu:          &sync.Mutex{},
		backendList: backends,
		accWeights:  make([]int, numOfBackends),
	}
	wrrp.accWeights[0] = backends[0].weight
	for i := 1; i < numOfBackends; i++ {
		wrrp.accWeights[i] = wrrp.accWeights[i-1] + backends[i].weight
	}
	return wrrp
}

func (wp *weightedRoundRobinPool) getNext() *Backend {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return getWeightedNext(wp.backendList, wp.accWeights)
}

type leastConnPool struct {
	backendList
	mu *sync.Mutex
}

func newLeastConnPool(backends []*Backend) *leastConnPool {
	lcp := &leastConnPool{
		mu:          &sync.Mutex{},
		backendList: backends,
	}
	return lcp
}

func (lp *leastConnPool) getNext() *Backend {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	var minLoad int64
	for i := 1; i < len(lp.backendList); i++ {
		if lp.backendList[i].load < minLoad {
			minLoad = lp.backendList[i].load
		}
	}
	var minBackends []*Backend
	for i := 0; i < len(lp.backendList); i++ {
		if lp.backendList[i].load == minLoad {
			minBackends = append(minBackends, lp.backendList[i])
		}
	}
	accWeights := make([]int, len(minBackends))
	accWeights[0] = minBackends[0].weight
	for i := 1; i < len(minBackends); i++ {
		accWeights[i] = accWeights[i-1] + minBackends[i].weight
	}
	return getWeightedNext(minBackends, accWeights)
}

func getWeightedNext(backends []*Backend, accWeights []int) *Backend {
	if len(backends) != len(accWeights) {
		return nil
	}
	randWeight := rand.Int() % accWeights[len(backends)-1]
	idx := sort.Search(len(backends), func(i int) bool {
		return accWeights[i] > randWeight
	})
	return backends[idx]
}
