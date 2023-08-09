package splitwise

import (
	"bzhang0/splitwise-implementation/pair"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/shopspring/decimal"
)

type Group struct {
	s *Splitwise

	id   int
	name string

	localBalance map[string]decimal.Decimal
	distribution map[string]map[string]decimal.Decimal
	// transactions        map[int]*Transaction		TODO: ADD
}

func NewGroup(s *Splitwise, id int, name string) *Group {
	return &Group{
		s:            s,
		id:           id,
		name:         name,
		localBalance: make(map[string]decimal.Decimal),
		distribution: make(map[string]map[string]decimal.Decimal),
	}
}

func (g *Group) AddMember(user string) (bool, error) {
	if _, ok := g.s.users[user]; !ok {
		return false, errors.New("user " + user + " does not exist")
	}
	if _, ok := g.distribution[user]; ok {
		return false, nil
	}

	g.localBalance[user] = decimal.NewFromInt(0)
	g.distribution[user] = make(map[string]decimal.Decimal)
	return true, nil
}

func (g *Group) AddTransaction(creditor string, debtorBreakdown string) error {
	if _, ok := g.distribution[creditor]; !ok {
		// format string like "creditor %s not in group", creditor.GetName())
		return errors.New("creditor " + creditor + " not in group")
	}

	inputs := strings.Split(debtorBreakdown, ",")
	for _, input := range inputs {
		tokens := strings.Split(input, "=")
		if len(tokens) != 2 {
			return errors.New("invalid input")
		}

		debtor := strings.TrimSpace(tokens[0])
		share, err := decimal.NewFromString(strings.TrimSpace(tokens[1]))
		if err != nil {
			return err
		}
		if _, ok := g.distribution[debtor]; !ok {
			return errors.New("debtor " + debtor + " not in group")
		}

		g.transfer(creditor, share, debtor)
	}

	// TODO: log transaction
	return nil
}

func (g *Group) transfer(creditor string, share decimal.Decimal, debtor string) {
	// invariant: both creditor and debtor are in the group

	// if the amount is zero, do nothing
	if share.Equal(decimal.NewFromInt(0)) {
		return
	}

	// fmt.Println("here")

	// update master user balance updates
	g.s.users[creditor] = g.s.users[creditor].Add(share)
	g.s.users[debtor] = g.s.users[debtor].Sub(share)

	// update local balances
	g.localBalance[creditor] = g.localBalance[creditor].Add(share)
	g.localBalance[debtor] = g.localBalance[debtor].Sub(share)

	// update local distribution
	if _, ok := g.distribution[creditor][debtor]; !ok {
		g.distribution[creditor][debtor] = decimal.NewFromInt(0)
	}
	g.distribution[creditor][debtor] = g.distribution[creditor][debtor].Add(share)

	if _, ok := g.distribution[debtor][creditor]; !ok {
		g.distribution[debtor][creditor] = decimal.NewFromInt(0)
	}
	g.distribution[debtor][creditor] = g.distribution[debtor][creditor].Sub(share)
}

// given a set of people and their balances, find the minimum number of transactions to settle all debts
// we follow this algorithm:
//
//	sort the people by balance. positive means they are a creditor, negative means they are a debtor.
//	then take the person who currently owes the most, then have them give as much as they can to the person who is owed the most.
//	repeat until everyone is settled.
//	it is guaranteed that this algorithm will terminate since the sum of all balances is 0.
//
// return this as a new map of distributions
//
// returns a map of users to how much they need to owe to other users. note this is only people who need to pay (debtors)
func (g *Group) SimplifyDebts() map[string]map[string]decimal.Decimal {
	simplifiedDebtorDistribution := make(map[string]map[string]decimal.Decimal)

	// create two priority queues. maxheap for pos and minheap for neg
	creditorQueue := queue.NewPriorityQueue(0, false)
	debtorQueue := queue.NewPriorityQueue(0, false)

	creditorTotal := decimal.NewFromInt(0)
	debtorTotal := decimal.NewFromInt(0)

	// fill the queues
	for user, balance := range g.localBalance {
		if balance.GreaterThan(decimal.NewFromInt(0)) {
			creditorQueue.Put(pair.StringDecimalPairMax{
				First:  user,
				Second: balance,
			})
			creditorTotal = creditorTotal.Add(balance)
		} else if balance.LessThan(decimal.NewFromInt(0)) {
			debtorQueue.Put(pair.StringDecimalPairMin{
				First:  user,
				Second: balance,
			})
			debtorTotal = debtorTotal.Add(balance)
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
		} else {
			if !creditor.Second.Equal(debtor.Second.Abs()) {
				panic("bad!!")
			}
		}
		// note the else case is they perfectly satisfy

		// regardless, log this transaction
		if _, ok := simplifiedDebtorDistribution[debtor.First]; !ok {
			simplifiedDebtorDistribution[debtor.First] = make(map[string]decimal.Decimal)
		}
		// debtor should ahve never paid creditor before
		if _, ok := simplifiedDebtorDistribution[debtor.First][creditor.First]; ok {
			panic("debtor should have never paid creditor before")
		}
		simplifiedDebtorDistribution[debtor.First][creditor.First] = toTransfer.Mul(decimal.NewFromInt(-1))
	}

	if !debtorQueue.Empty() {
		panic("debtor queue should be empty")
	}

	return simplifiedDebtorDistribution
}

func (g *Group) TotalTransfers() int {
	return TotalTransfers(g.distribution)
}

func TotalTransfers(distribution map[string]map[string]decimal.Decimal) int {
	// count all negative values in the distribution
	total := 0
	for _, debtorDistribution := range distribution {
		for _, transfer := range debtorDistribution {
			if transfer.LessThan(decimal.NewFromInt(0)) {
				total++
			}
		}
	}
	return total
}

func (g *Group) PrintBalances() {
	// Sort the people
	var sortedUsers []string
	for person := range g.localBalance {
		sortedUsers = append(sortedUsers, person)
	}
	sort.Strings(sortedUsers)

	fmt.Println("\nBalances:")
	for _, user := range sortedUsers {
		fmt.Printf("- %s: %s\n", user, g.localBalance[user].String())
	}
}

func (g *Group) GetId() int {
	return g.id
}

func (g *Group) GetName() string {
	return g.name
}

// hash function
func (g *Group) Hash() int {
	return g.id
}
