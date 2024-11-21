package proc

import "sync"

type InMemStorage struct {
	dat map[string]*Client
	mut sync.RWMutex
}

func (s *InMemStorage) Get(conn string) (*Client, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	cli, ok := s.dat[conn]

	return cli, ok
}

func (s *InMemStorage) Set(conn string, cli *Client) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.dat[conn] = cli
}
