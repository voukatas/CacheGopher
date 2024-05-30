package client

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"
)

type CacheNode struct {
	ID        string
	IsPrimary bool
	Hash      uint32
	*ConnPool
	Unhealthy  bool
	RetryAt    time.Time
	HealthLock sync.Mutex // Lock for controling the health status and the timestamp of it
}

func NewCacheNode(id string, isPrimary bool, pool *ConnPool) *CacheNode {
	hash := sha1.New()
	hash.Write([]byte(pool.address))

	return &CacheNode{
		ID:        id,
		IsPrimary: isPrimary,
		Hash:      binary.BigEndian.Uint32(hash.Sum(nil)[:4]),
		ConnPool:  pool,
	}
}

func (node *CacheNode) SetUnhealthy(delay time.Duration) {
	node.HealthLock.Lock()
	defer node.HealthLock.Unlock()

	node.Unhealthy = true
	node.RetryAt = time.Now().Add(delay)
}

type HashRing interface {
	AddNode(*CacheNode)
	RemoveNode(*CacheNode)
	GetNode(string) (*CacheNode, error)
}

type SimpleHashRing struct {
	nodes []*CacheNode
	lock  sync.RWMutex
}

func NewHashRing() HashRing {

	return &SimpleHashRing{
		nodes: make([]*CacheNode, 0),
	}
}

// maybe i need to implement also virtual nodes to better distribute the keys
func (s *SimpleHashRing) AddNode(node *CacheNode) {
	// s.lock.Lock()
	// defer s.lock.Unlock()
	//fmt.Println("node hash: ", node.Hash)
	//getLogger().Debug("node hash: " + fmt.Sprintf("%v", node.Hash))

	s.nodes = append(s.nodes, node)
	sort.Slice(s.nodes, func(i, j int) bool {
		return s.nodes[i].Hash < s.nodes[j].Hash
	})
}

// Reserved for future use
// If this method is used, verify that the GetNode has the locks code
func (s *SimpleHashRing) RemoveNode(node *CacheNode) {
	// s.lock.Lock()
	// defer s.lock.Unlock()

	for i, n := range s.nodes {
		if node.ID == n.ID {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			break
		}
	}

}

// In case a discovery functionality is added, the mutexes should be used
// Take the risky road and avoid using the mutexes for now since we only read, for now...
func (s *SimpleHashRing) GetNode(key string) (*CacheNode, error) {
	// s.lock.RLock()
	// defer s.lock.RUnlock()
	if len(s.nodes) == 0 {
		return nil, fmt.Errorf("ring is empty")
	}

	hash := sha1.New()
	hash.Write([]byte(key))
	// common practice to use 32-bit, also performs faster in calculations
	keyHash := binary.BigEndian.Uint32(hash.Sum(nil)[:4])
	// fmt.Println("keyHash  ", keyHash)
	//
	// for _, node := range s.nodes {
	//
	// 	fmt.Println("node id: ", node.ID)
	// }

	// O(n) time complexity
	// for _, node := range s.nodes {
	// 	if node.Hash >= keyHash {
	// 		return node, nil
	// 	}
	// }

	// O(logn)
	idx := sort.Search(len(s.nodes), func(i int) bool {
		return s.nodes[i].Hash >= keyHash
	})

	if idx < len(s.nodes) {
		//fmt.Println("node selected to send the request: ", s.nodes[idx])
		return s.nodes[idx], nil
	}

	// if we reached here it means a full circle is done

	return s.nodes[0], nil
}
