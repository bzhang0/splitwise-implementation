package splitwise

import (
	"bzhang0/splitwise-implementation/pair"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/shopspring/decimal"
)

func GetBalances(distribution map[string]map[string]decimal.Decimal) map[string]decimal.Decimal {
	balances := make(map[string]decimal.Decimal)
	for creditor, dist := range distribution {
		bal := decimal.NewFromInt(0)
		for _, amount := range dist {
			bal = bal.Add(amount)
		}
		balances[creditor] = bal
	}
	return balances
}

func SimplifyDebts(balances map[string]decimal.Decimal) map[string]map[string]decimal.Decimal {
	simplifiedDistribution := make(map[string]map[string]decimal.Decimal)

	// create two priority queues. maxheap for pos and minheap for neg
	creditorQueue := queue.NewPriorityQueue(0, false)
	debtorQueue := queue.NewPriorityQueue(0, false)

	for user, balance := range balances {
		if balance.GreaterThan(decimal.NewFromInt(0)) {
			creditorQueue.Put(pair.StringDecimalPairMax{
				First:  user,
				Second: balance,
			})
		} else if balance.LessThan(decimal.NewFromInt(0)) {
			debtorQueue.Put(pair.StringDecimalPairMin{
				First:  user,
				Second: balance,
			})
		}
	}

	for !creditorQueue.Empty() {
		creditor := creditorQueue.Peek().(pair.StringDecimalPairMax)
		debtor := debtorQueue.Peek().(pair.StringDecimalPairMin)

		creditorQueue.Get(1)
		debtorQueue.Get(1)

		toTransfer := decimal.Min(creditor.Second, debtor.Second.Abs())

		if debtor.Second.Add(toTransfer).LessThan(decimal.NewFromInt(0)) {
			debtorQueue.Put(pair.StringDecimalPairMin{
				First:  debtor.First,
				Second: debtor.Second.Add(toTransfer),
			})
		} else if creditor.Second.Sub(toTransfer).GreaterThan(decimal.NewFromInt(0)) {
			creditorQueue.Put(pair.StringDecimalPairMax{
				First:  creditor.First,
				Second: creditor.Second.Sub(toTransfer),
			})
		}

		// regardless, log this transaction
		if _, ok := simplifiedDistribution[debtor.First]; !ok {
			simplifiedDistribution[debtor.First] = make(map[string]decimal.Decimal)
		}
		if _, ok := simplifiedDistribution[creditor.First]; !ok {
			simplifiedDistribution[creditor.First] = make(map[string]decimal.Decimal)
		}
		simplifiedDistribution[debtor.First][creditor.First] = toTransfer.Neg()
		simplifiedDistribution[creditor.First][debtor.First] = toTransfer
	}

	return simplifiedDistribution
}

func SimplifyDebtsFromDistribution(distribution map[string]map[string]decimal.Decimal) map[string]map[string]decimal.Decimal {
	balances := GetBalances(distribution)
	return SimplifyDebts(balances)
}
