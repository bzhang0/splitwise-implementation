package pair

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/shopspring/decimal"
)

type StringDecimalPairMin struct {
	First  string
	Second decimal.Decimal
}

func (p StringDecimalPairMin) Compare(other queue.Item) int {
	otherPair, ok := other.(StringDecimalPairMin)
	if !ok {
		panic("Attempted to compare with a different type")
	}

	if compare := p.Second.Cmp(otherPair.Second); compare != 0 {
		return compare
	}

	if p.First > otherPair.First {
		return -1
	} else if p.First < otherPair.First {
		return 1
	}

	return 0
}
