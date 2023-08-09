package main

import (
	"bzhang0/splitwise-implementation/splitwise"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

func main() {
	sw := splitwise.NewSplitwise()

	parseCSV(sw, "input/crooze.csv")
	g, _ := sw.GetGroup(0)

	g.PrintBalances()

	simplifiedDistribution := g.SimplifyDebts()

	fmt.Println("\nSimplifying debts...")
	totalTransfers := g.TotalTransfers()
	simplifiedTransfers := splitwise.TotalTransfers(simplifiedDistribution)

	var plural string
	savedTransfers := totalTransfers - simplifiedTransfers
	if savedTransfers > 1 {
		plural = "s"
	}

	fmt.Printf("Simplify debts saved %d balance transfer%s (%d -> %d)\n", savedTransfers, plural, totalTransfers, simplifiedTransfers)

	// Sort the people
	var sortedUsers []string
	for person := range simplifiedDistribution {
		sortedUsers = append(sortedUsers, person)
	}
	sort.Strings(sortedUsers)

	for _, person := range sortedUsers {
		total := decimal.NewFromInt(0)
		for _, value := range simplifiedDistribution[person] {
			total = total.Add(value)
		}

		fmt.Printf("\n%s owes %s in total\n", person, total.Mul(decimal.NewFromInt(-1)))
		for creditor, amount := range simplifiedDistribution[person] {
			fmt.Printf("- owes %s to %s\n", amount.Mul(decimal.NewFromInt(-1)), creditor)
		}
	}
}

func parseCSV(sw *splitwise.Splitwise, filename string) (int, error) {
	fmt.Println("\nParsing CSV...")
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		panic(err)
	}

	id := sw.CreateGroup(filename)
	g, _ := sw.GetGroup(id)

	users := records[0][5:]
	for _, user := range users {
		sw.CreateUser(user)
		g.AddMember(user)
	}

	for _, record := range records[1 : len(records)-1] {
		var creditor string
		if err != nil {
			return -1, err
		}

		// find the creditor
		tokens := record[5:]
		for i, token := range tokens {
			val, err := decimal.NewFromString(token)
			if err != nil {
				return -1, err
			}
			if val.GreaterThan(decimal.NewFromInt(0)) {
				creditor = users[i]
				// TODO: add individualized balance if multiple people paid
				break
			}
		}
		if creditor == "" {
			return -1, errors.New("no creditor found")
		}

		// format data for AddTransaction
		var sb strings.Builder
		for i, token := range tokens {
			bal, err := decimal.NewFromString(token)
			if err != nil {
				return -1, err
			}

			if bal.LessThan(decimal.NewFromInt(0)) {
				sb.WriteString(users[i])
				sb.WriteString("=")
				sb.WriteString(bal.Mul(decimal.NewFromInt(-1)).String())
				sb.WriteString(",")
			}
		}

		s := sb.String()
		s = s[:len(s)-1]
		g.AddTransaction(creditor, s)
	}

	return id, nil
}
