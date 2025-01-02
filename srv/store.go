package srv

import "sync"

type ExclusiveRule struct {

}

type Store struct {
	

	dat map[string]*Client
	mux sync.RWMutex
}

func (s *Store) PutClient(conn string, client *Client) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dat[conn] = client
}

func (s *Store) RemClient(conn string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dat[conn] = nil
}

func (s *Store) GetClient(con string) (*Client, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	c, ok := s.dat[con]

	return c, ok
}
