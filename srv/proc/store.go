package proc

import "sync"

type Store struct {
	dat map[string]*Client
	mux sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		dat: make(map[string]*Client, 5000),
	}
}

func (s *Store) PutClient(conn string, client *Client) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dat[conn] = client
}

func (s *Store) RemoveClient(conn string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dat[conn] = nil
}

func (s *Store) GetClientByConn(conn string) (*Client, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	c, ok := s.dat[conn]

	return c, ok
}
