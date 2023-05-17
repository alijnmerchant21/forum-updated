package moderators

import (
	"bytes"

	"github.com/interchainio/forum/model"
)

type Set struct {
	l []model.User
	m map[string]struct{}
}

func NewSet() *Set {
	return &Set{
		l: []model.User{},
		m: map[string]struct{}{},
	}
}

func (s *Set) Add(u model.User) bool {
	if _, ok := s.m[string(u.PubKey)]; ok {
		return false
	}
	s.l = append(s.l, u)
	s.m[string(u.PubKey)] = struct{}{}
	return true
}

func (s *Set) Remove(u model.User) bool {
	if _, ok := s.m[string(u.PubKey)]; !ok {
		return false
	}
	delete(s.m, string(u.PubKey))
	for i := 0; i < len(s.l); i++ {
		if bytes.Equal(u.PubKey, s.l[i].PubKey) {
			s.l = append(s.l[:i], s.l[i+1:]...)
			return true
		}
	}
	panic("list was expected to contain pubkey but did not!")
}

func (s *Set) List() []model.User {
	return s.l
}
