package splitwise

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

type Group struct {
	s *Splitwise

	id   int
	name string

	distribution map[string]map[string]decimal.Decimal
	transactions []string
}

func NewGroup(s *Splitwise, id int, name string) *Group {
	return &Group{
		s:            s,
		id:           id,
		name:         name,
		distribution: make(map[string]map[string]decimal.Decimal),
		transactions: make([]string, 0),
	}
}

func (g *Group) AddUser(user string) (bool, error) {
	if _, ok := g.s.users[user]; !ok {
		return false, errors.New("user " + user + " does not exist")
	}
	if _, ok := g.distribution[user]; ok {
		return false, nil
	}

	g.distribution[user] = make(map[string]decimal.Decimal)
	return true, nil
}

func (g *Group) AddTransaction(total decimal.Decimal, breakdown string) error {
	pairs := strings.Split(breakdown, ",")

	transactionBalances := make(map[string]decimal.Decimal)
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return errors.New("invalid input")
		}
		k := strings.TrimSpace(kv[0])
		if _, ok := g.distribution[k]; !ok {
			return errors.New("user " + k + " does not exist")
		}
		v, err := decimal.NewFromString(kv[1])
		if err != nil {
			return err
		}

		transactionBalances[k] = v
	}
	transactionDistribution := SimplifyDebts(transactionBalances)
	for user, dist := range transactionDistribution {
		for otherUser, amount := range dist {
			if _, ok := g.distribution[user][otherUser]; !ok {
				g.distribution[user][otherUser] = decimal.NewFromInt(0)
			}
			g.distribution[user][otherUser] = g.distribution[user][otherUser].Add(amount)
		}
	}

	g.transactions = append(g.transactions, fmt.Sprintf("%s: %s", total.String(), breakdown))
	return nil
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
	return SimplifyDebtsFromDistribution(g.distribution)
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
	balances := GetBalances(g.distribution)

	var sortedUsers []string
	for person := range balances {
		sortedUsers = append(sortedUsers, person)
	}
	sort.Strings(sortedUsers)

	fmt.Println("\nBalances:")
	for _, user := range sortedUsers {
		fmt.Printf("- %s: %s\n", user, balances[user].String())
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
