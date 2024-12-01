package proc

import "sync"

type Store struct {
	dat map[string]*Client
	mut sync.Mutex
}

func (s *Store) PutClient(conn string, client *Client) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.dat[conn] = client
}

func (s *Store) GetClientByConn(conn string) (*Client, bool) {
	s.mut.Lock()
	defer s.mut.Unlock()

	c, ok := s.dat[conn]

	return c, ok
}
