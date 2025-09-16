package graph

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiterData struct {
	limit    int
	interval time.Duration
}

// Globals contains all of the nodes so that
// they can all be run at the start of the program.
type Globals struct {
	currID          int
	resultID        int
	mutex           *sync.Mutex
	nodeMap         map[int]Node
	rateLimiterData map[string]RateLimiterData
	rateLimiters    map[string]*rate.Limiter
}

func NewGlobals() *Globals {
	return &Globals{
		currID:          0,
		mutex:           &sync.Mutex{},
		nodeMap:         make(map[int]Node),
		rateLimiterData: make(map[string]RateLimiterData),
		rateLimiters:    make(map[string]*rate.Limiter),
	}
}

// GenID generates a new ID for a new node.
func (n *Globals) GenID() int {
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

// SetResultNodeID sets the node id of the result node
// (the node that yields the final value in the program).
func (n *Globals) SetResultNodeID(id int) {
	n.resultID = id
}

// ResultNode returns the node in the globals with the result node id.
func (n *Globals) ResultNode() Node {
	return n.nodeMap[n.resultID]
}

// Run runs all of the nodes in the map.
func (n *Globals) Run() {
	for _, node := range n.nodeMap {
		go RunNode(node)
	}
}

// SetRateLimit sets the rate limit for a URL.
func (n *Globals) SetRateLimit(url string, maxTimes int, duration time.Duration) {
	n.rateLimiterData[url] = RateLimiterData{
		maxTimes,
		duration,
	}
	n.rateLimiters[url] = rate.NewLimiter(rate.Every(duration/time.Duration(maxTimes)), 1)
}

// RateLimit blocks until the URL can be called without breaking the rate limit.
func (n *Globals) RateLimit(url string) {
	if rateLimiter, ok := n.rateLimiters[url]; ok {
		rateLimiter.Wait(context.Background())
	}
}
