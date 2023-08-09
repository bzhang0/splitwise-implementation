package pair

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/shopspring/decimal"
)

type StringDecimalPairMax struct {
	First  string
	Second decimal.Decimal
}

func (p StringDecimalPairMax) Compare(other queue.Item) int {
	otherPair, ok := other.(StringDecimalPairMax)
	if !ok {
		panic("Attempted to compare with a different type")
	}

	if compare := p.Second.Cmp(otherPair.Second); compare != 0 {
		return -compare
	}

	if otherPair.First > p.First {
		return -1
	} else if otherPair.First < p.First {
		return 1
	}

	return 0
}
