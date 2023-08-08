package Splitwise;

import java.math.BigDecimal;
import java.util.*;

import javafx.util.Pair;

public class Splitwise {
    private boolean DEBUG;

    private int personIDCounter;
    private int expenseIDCounter;
    private int groupIDCounter;

    public Set<String> people;     // for now, we will assume that each person is unique
    public Map<String, BigDecimal> balances;

    // debtorDistributions: <X, <Y, -10>> means that X owes 10 to Y
    public Map<String, Map<String, BigDecimal>> debtorDistributions;
    // note that debtorDistributions can only decrease in value

    public List<String> expenses;

    public Splitwise() {
        clear(false);
    }

    public Splitwise(boolean debug) {
        clear(debug);
    }

    public void clear(boolean debug) {
        this.DEBUG = debug;
        people = new HashSet<>();
        balances = new HashMap<>();
        debtorDistributions = new HashMap<>();
        expenses = new ArrayList<>();
    }

    public void newPerson(String name) {
        if (people.contains(name)) {
            throw new IllegalArgumentException("Splitwise.Person already exists");
        }

        people.add(name);
        balances.put(name, BigDecimal.ZERO);
        debtorDistributions.put(name, new HashMap<>());
    }

    public void newExpense(String creditor, BigDecimal amount, String debtorBreakdown) {
        // i dont want to write breakdown checking right now lol

        if (!people.contains(creditor)) {
            System.out.println("Invalid expense");
            return;
        }

        String[] inputs = debtorBreakdown.split(",");
        for (String input : inputs) {
            String[] tokens = input.split("=");
            if (tokens.length != 2) {
                System.out.println("Invalid expense");
                return;
            }

            String debtor = tokens[0];
            BigDecimal share = new BigDecimal(tokens[1]);

            if (!people.contains(debtor)) {
                System.out.println("Invalid expense");
                return;
            }

            transfer(creditor, share, debtor);
        }

        logExpense(creditor, amount, debtorBreakdown);
    }

    private void logExpense(String creditor, BigDecimal amount, String debtorBreakdown) {
        expenses.add(creditor + " paid " + amount + " for " + debtorBreakdown);
    }

    public int rawTotalTransfers() {
        // we count the number of negative values in debtorDistributions
        return rawTotalTransfersHelper(debtorDistributions);
    }

    private int rawTotalTransfersHelper(Map<String, Map<String, BigDecimal>> distributions) {
        // count number of BigDecimals in distributions (for now, assume that all BigDecimals are negative)
        return distributions.values().stream().mapToInt(Map::size).sum();
    }

    private void transfer(String creditor, BigDecimal amount, String debtor) {
        assert people.contains(creditor) && people.contains(debtor) : "Both people must exist";

        // OVERALL BALANCE UPDATES
        balances.put(creditor, balances.get(creditor).add(amount));
        balances.put(debtor, balances.get(debtor).subtract(amount));

        // DISTRIBUTION UPDATES
        // the debtor has debts to the creditor
        Map<String, BigDecimal> debtorDistribution = debtorDistributions.get(debtor);
        debtorDistribution.put(creditor, debtorDistribution.getOrDefault(creditor, BigDecimal.ZERO).subtract(amount));
    }

    // simplify debts algorithm.
    // given a set of people and their balances, find the minimum number of transfers to settle all debts
    // we follow this algorithm:
    //   sort the people by balance. positive means they are a creditor, negative means they are a debtor.
    //   then take the person who currently owes the most, then have them give as much as they can to the person who is owed the most.
    //   repeat until everyone is settled.
    //   it is guaranteed that this algorithm will terminate since the sum of all balances is 0.
    // return this as a new map of distributions
    public Map<String, Map<String, BigDecimal>> simplifyDebts() {
        Map<String, Map<String, BigDecimal>> simplifiedDebtorDistribution = new HashMap<>();

        // we create two priority queues of pos and neg balances

        // creditor queue sorts by max
        PriorityQueue<Pair<String, BigDecimal>> creditorQueue = new PriorityQueue<>((a, b) -> b.getValue().compareTo(a.getValue()));
        PriorityQueue<Pair<String, BigDecimal>> debtorQueue = new PriorityQueue<>((a, b) -> a.getValue().compareTo(b.getValue()));

        // fill the queues
        for (String person : people) {
            BigDecimal balance = balances.get(person);
            if (balance.compareTo(BigDecimal.ZERO) > 0) {
                creditorQueue.add(new Pair<>(person, balance));
            } else if (balance.compareTo(BigDecimal.ZERO) < 0) {
                debtorQueue.add(new Pair<>(person, balance));
            }
        }

        System.out.println();
        System.out.println(creditorQueue);
        System.out.println();
        System.out.println(debtorQueue);
        System.out.println();

        // sum of all values in creditorQueue
        BigDecimal creditorTotal = creditorQueue.stream().map(Pair::getValue).reduce(BigDecimal.ZERO, BigDecimal::add);
        // sum of all values in debtorQueue
        BigDecimal debtorTotal = debtorQueue.stream().map(Pair::getValue).reduce(BigDecimal.ZERO, BigDecimal::add);

//        System.out.println("creditorTotal: " + creditorTotal);
//        System.out.println("debtorTotal: " + debtorTotal);

        // we should have the same total
        assert creditorTotal.compareTo(debtorTotal.multiply(BigDecimal.valueOf(-1))) == 0 : "Total of creditorQueue and debtorQueue should be the same";

        // execute the algorithm!
        while (!creditorQueue.isEmpty()) {
            Pair<String, BigDecimal> creditor = creditorQueue.remove();
            Pair<String, BigDecimal> debtor = debtorQueue.remove();

            BigDecimal toTransfer = debtor.getValue().multiply(BigDecimal.valueOf(-1)).min(creditor.getValue());

            if (debtor.getValue().add(toTransfer).compareTo(BigDecimal.ZERO) < 0) {
                // debtor still owes more
                debtorQueue.add(new Pair<>(debtor.getKey(), debtor.getValue().add(toTransfer)));
            } else if (creditor.getValue().subtract(toTransfer).compareTo(BigDecimal.ZERO) > 0) {
                // debtor paid all they owe, but did not fill all that the creditor is owed
                creditorQueue.add(new Pair<>(creditor.getKey(), creditor.getValue().subtract(toTransfer)));
            }
            // note the else case is they perfectly satisfy.

            // regardless, log this expense
            simplifiedDebtorDistribution.putIfAbsent(debtor.getKey(), new HashMap<>());
            Map<String, BigDecimal> debtorDistribution = simplifiedDebtorDistribution.get(debtor.getKey());
            // debtor should have never paid creditor before
            assert !debtorDistribution.containsKey(creditor.getKey()) : "Debtor should not have paid creditor before";
            debtorDistribution.put(creditor.getKey(), toTransfer.multiply(BigDecimal.valueOf(-1)));
        }
        assert debtorQueue.isEmpty() : "There should be an equal number of pos and neg balances";

        System.out.println("simplified transfers: " + rawTotalTransfersHelper(simplifiedDebtorDistribution));
        return simplifiedDebtorDistribution;
    }
}