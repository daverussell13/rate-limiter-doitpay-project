package util

import (
	"hash/fnv"
	"sync"
)

type StripedMutex struct {
	stripes []sync.Mutex
	size    uint32
}

func NewStripedMutex(n int) *StripedMutex {
	return &StripedMutex{
		stripes: make([]sync.Mutex, n),
		size:    uint32(n),
	}
}

func (s *StripedMutex) idx(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % s.size
}

func (s *StripedMutex) Lock(key string) func() {
	i := s.idx(key)
	s.stripes[i].Lock()
	return func() { s.stripes[i].Unlock() }
}
