package mem

import (
	"sync"

	"github.com/meha4j/extd/internal/proc"
)

type Storage struct {
	data map[string]*proc.Client
	mutx sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]*proc.Client),
	}
}

func (s *Storage) GetClient(conn string) (*proc.Client, bool) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	cli, ok := s.data[conn]

	return cli, ok
}

func (s *Storage) SetClient(conn string, cli *proc.Client) {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.data[conn] = cli
}
