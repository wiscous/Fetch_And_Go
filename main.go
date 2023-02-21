package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
)

type Transaction struct {
	Payer     string
	Points    int
	Timestamp time.Time
}

// Boilerplate to sort transactions via the sort package.
type ByTimestamp []*Transaction

func (self ByTimestamp) Len() int { return len(self) }
func (self ByTimestamp) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self ByTimestamp) Less(i, j int) bool {
	return self[i].Timestamp.Before(self[j].Timestamp)
}

func deductPoints(pointsBalance []*Transaction, points int) []*Transaction {
	// Index of the head of the final queue of non-zero transactions remaining after deduction.
	head := 0
	remainingPoints := points
	for remainingPoints > 0 && head < len(pointsBalance) {
		if remainingPoints <= pointsBalance[head].Points {
			pointsBalance[head].Points -= remainingPoints
			remainingPoints = 0
		} else {
			remainingPoints -= pointsBalance[head].Points
			pointsBalance[head].Points = 0
		}

		// Once we have deducted all the points from a transaction we move on to the next
		// transaction
		if pointsBalance[head].Points == 0 {
			head += 1
		}
	}

	// If remaining points is zero and we have no transaction to deduct points from, then the transaction
	// log received as input is invalid.
	if remainingPoints != 0 {
		fmt.Fprintln(os.Stderr, "Invalid operation. Evaluation of transactions and points yields a negative balance.")
		os.Exit(1)
	}

	// we exclude point balances which have been zeroed out.
	return pointsBalance[head:]
}

// Processed the provided transaction log and returns the final payer balances.
func processTransactions(transactions []*Transaction, pointsToSpend int) map[string]int {
	// We sort the transactions by timestamp in ascending order.
	sort.Sort(ByTimestamp(transactions))

	// Since points chosen for deduction depends on the timestamp of the transaction they originated
	// from we track all points as a list of transactions for the payer as well as the user.
	payerBalances := map[string][]*Transaction{}
	userBalance := []*Transaction{}

	for _, transaction := range transactions {
		payer := transaction.Payer

		if transaction.Points > 0 {
			payerBalances[payer] = append(payerBalances[payer], transaction)
			// The user balance consists of all the positive transactions in the log.
			// The points in the balance will be adjusted as we perform deductions.
			userBalance = append(userBalance, transaction)
		} else {
			// For negative transactions we deduct points from the payer balances in order of transaction.
			payerBalances[payer] = deductPoints(payerBalances[payer], -transaction.Points)
		}
	}

	// Deduct pointsToSpend from the transaction balances once the transaction log has been verified and adjusted
	// to handle point deduction or negative transactions
	deductPoints(userBalance, pointsToSpend)

	finalPayerBalances := map[string]int{}
	for payer, transactionLog := range payerBalances {
		payerBalance := 0
		for _, processedTransaction := range transactionLog {
			payerBalance += processedTransaction.Points
		}

		finalPayerBalances[payer] = payerBalance
	}
	return finalPayerBalances
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: ./Fetch_And_Go <Points to spend> <Path to transaction file>\nAny additional trailing arguments will be ignored.")
		os.Exit(1)
	}

	pointsToSpend, err := strconv.Atoi(os.Args[1])

	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not convert", os.Args[1], "to an integer due to error:", err.Error())
		os.Exit(1)
	}

	transactionFile, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open file", os.Args[2], "due to error: ", err.Error())
		os.Exit(1)
	}

	csvReader := csv.NewReader(transactionFile)

	passedFirstLine := false
	transactions := []*Transaction{}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, "Encountered error while processing CSV:", err.Error())
			os.Exit(1)
		}

		// First line of the CSV file is column headers, not a transaction record so we skip it.
		if !passedFirstLine {
			passedFirstLine = true
			continue
		}

		// CSV records should be of length 3 corresponding to payer, points and timestamp.
		if len(record) != 3 {
			// We skip over records with invalid formats
			fmt.Fprintln(os.Stderr, "Skipping invalid CSV record, only records of length 3 are accepted.")
			continue
		}

		payer := record[0]

		points, err := strconv.Atoi(record[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not convert", record[1], "to an integer due to error:", err)
		}

		timestamp, err := time.Parse(time.RFC3339, record[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not parse timestamp:", err)
		}


		transaction := &Transaction{payer, points, timestamp}
		transactions = append(transactions, transaction)
	}

	finalPayerBalances := processTransactions(transactions, pointsToSpend)

	jsonBytes, err := json.MarshalIndent(finalPayerBalances, "", "\t")

	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonBytes))
}
