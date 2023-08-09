package splitwise

import (
	"github.com/shopspring/decimal"
)

type User struct {
	id   int // phone number
	name string

	overallBalance decimal.Decimal
	groupBalances  map[*Group]decimal.Decimal
}

func NewUser(id int, name string) *User {
	return &User{
		id:             id,
		name:           name,
		overallBalance: decimal.NewFromInt(0),
		groupBalances:  make(map[*Group]decimal.Decimal),
	}
}

func (u *User) GetId() int {
	return u.id
}

func (u *User) GetName() string {
	return u.name
}

// hash function
func (u *User) Hash() int {
	return u.id
}
