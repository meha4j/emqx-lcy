package proc

import "sync"

type InMemStorage struct {
	data map[string]*Client
	mutx sync.RWMutex
}

func (s *InMemStorage) Get(conn string) (*Client, bool) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	cli, ok := s.data[conn]

	return cli, ok
}

func (s *InMemStorage) Set(conn string, cli *Client) {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.data[conn] = cli
}
