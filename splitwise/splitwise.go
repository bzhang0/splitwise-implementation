package splitwise

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type Splitwise struct {
	groupIDCounter int

	users  map[string]decimal.Decimal
	groups map[int]*Group
}

func NewSplitwise() *Splitwise {
	return &Splitwise{
		groupIDCounter: 0,

		users:  make(map[string]decimal.Decimal),
		groups: make(map[int]*Group),
	}
}

func (s *Splitwise) CreateUser(name string) {
	if _, ok := s.users[name]; ok {
		return
	}
	s.users[name] = decimal.NewFromInt(0)
}

func (s *Splitwise) CreateGroup(name string) int {
	groupID := s.getNextGroupID()
	s.groups[groupID] = NewGroup(s, groupID, name)
	return groupID
}

func (s *Splitwise) GetGroup(groupID int) (*Group, error) {
	if _, ok := s.groups[groupID]; !ok {
		// weird error handling lmao
		return nil, errors.New(fmt.Errorf("group %d does not exist", groupID).Error())
	}

	return s.groups[groupID], nil
}

func (s *Splitwise) getNextGroupID() int {
	temp := s.groupIDCounter
	s.groupIDCounter++
	return temp
}
