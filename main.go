package main

import (
	"bzhang0/splitwise-implementation/splitwise"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

func main() {
	sw := splitwise.NewSplitwise()

	parseCSV(sw, "")
	g, _ := sw.GetGroup(0)

	g.PrintBalances()

	simplifiedDistribution := g.SimplifyDebts()

	fmt.Println("\nSimplifying debts...")
	totalTransfers := g.TotalTransfers()
	simplifiedTransfers := splitwise.TotalTransfers(simplifiedDistribution)

	var plural string
	savedTransfers := totalTransfers - simplifiedTransfers
	if savedTransfers > 1 || savedTransfers == 0 {
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

		if total.GreaterThanOrEqual(decimal.NewFromInt(0)) {
			continue
		}

		fmt.Printf("\n%s owes %s in total\n", person, total.Neg())
		for creditor, amount := range simplifiedDistribution[person] {
			fmt.Printf("- owes %s to %s\n", amount.Neg(), creditor)
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
		g.AddUser(user)
	}

	for _, record := range records[1:] {
		// stop parsing when we reach the total balance
		if record[1] == "Total balance" {
			break
		}

		total, _ := decimal.NewFromString(record[3])

		var sb strings.Builder
		for i, amount := range record[5:] {
			amount, err := decimal.NewFromString(amount)
			if err != nil {
				return -1, err
			}

			sb.WriteString(users[i])
			sb.WriteString("=")
			sb.WriteString(amount.String())
			sb.WriteString(",")
		}

		s := sb.String()
		s = s[:len(s)-1] // remove last comma

		g.AddTransaction(total, s)
	}

	return id, nil
}
