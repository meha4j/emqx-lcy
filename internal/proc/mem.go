package proc

import "sync"

type Store struct {
	dat map[string]*Client
	mut sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		dat: make(map[string]*Client, 5000),
	}
}

func (s *Store) PutClient(conn string, client *Client) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.dat[conn] = client
}

func (s *Store) RemoveClient(conn string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.dat[conn] = nil
}

func (s *Store) GetClientByConn(conn string) (*Client, bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	c, ok := s.dat[conn]

	return c, ok
}
