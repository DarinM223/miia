package graph

import (
	"github.com/beefsack/go-rate"
	"sync"
	"time"
)

// Globals contains all of the nodes so that
// they can all be run at the start of the program.
type Globals struct {
	currID       int
	mutex        *sync.Mutex
	nodeMap      map[int]Node
	rateLimiters map[string]*rate.RateLimiter
}

func NewGlobals() *Globals {
	return &Globals{
		currID:       0,
		mutex:        &sync.Mutex{},
		nodeMap:      make(map[int]Node),
		rateLimiters: make(map[string]*rate.RateLimiter),
	}
}

// GenerateID generates a new ID for a new node.
func (n *Globals) GenerateID() int {
	n.mutex.Lock()
	id := n.currID
	n.currID++
	n.mutex.Unlock()
	return id
}

// RegisterNode registers a node for an id.
func (n *Globals) RegisterNode(id int, node Node) {
	n.mutex.Lock()
	n.nodeMap[id] = node
	n.mutex.Unlock()
}

// Run runs all of the nodes in the map.
func (n *Globals) Run() {
	for _, node := range n.nodeMap {
		go RunNode(node)
	}
}

// SetRateLimit sets the rate limit for a URL.
func (n *Globals) SetRateLimit(url string, maxTimes int, duration time.Duration) {
	n.rateLimiters[url] = rate.New(maxTimes, duration)
}

// RateLimit blocks until the URL can be called without breaking the rate limit.
func (n *Globals) RateLimit(url string) {
	if rateLimiter, ok := n.rateLimiters[url]; ok {
		rateLimiter.Wait()
	}
}
