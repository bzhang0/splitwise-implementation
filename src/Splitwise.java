import java.util.*;

import javafx.util.Pair;

public class Splitwise {

    static Set<String> people = new HashSet<>();
    static Map<String, Integer> balances = new HashMap<>();
    static Map<String, Map<String, Integer>> distribution = new HashMap<>();
    static List<String> transactions = new ArrayList<>();

    public static void main(String[] args) {

        newPerson("A");
        newPerson("B");
        newPerson("C");

        newTransaction("A", 10, "B=10");
        newTransaction("B", 10, "C=10");

        System.out.println(distribution);
        // count in one line the number of entries that have negative int value in each entry of distribution
        int totalTransactions = distribution.values().stream()
                .mapToInt(innerMap -> (int) innerMap.values().stream().filter(value -> value < 0).count())
                .sum();
        System.out.println(totalTransactions);

        // in theory, no one should pay
        Map<String, Map<String, Integer>> simplifiedDistribution = simplifyDebts();
        System.out.println(simplifiedDistribution);
        int simplifiedTotalTransactions = simplifiedDistribution.values().stream()
                .mapToInt(innerMap -> (int) innerMap.values().stream().filter(value -> value < 0).count())
                .sum();
        System.out.println(simplifiedTotalTransactions);
    }

    public static void newPerson(String name) {
        assert !people.contains(name) : "Person already exists";

        people.add(name);
        balances.put(name, 0);
    }

    public static void newTransaction(String creditor, int amount, String debtorBreakdown) {
        // i dont want to write breakdown checking right now lol

        if (!people.contains(creditor)) {
            System.out.println("Invalid transaction");
            return;
        }

        String[] inputs = debtorBreakdown.split(",");
        for (String input : inputs) {
            String[] tokens = input.split("=");
            if (tokens.length != 2) {
                System.out.println("Invalid transaction");
                return;
            }

            String debtor = tokens[0];
            int share = Integer.parseInt(tokens[1]);

            if (!people.contains(debtor)) {
                System.out.println("Invalid transaction");
                return;
            }

            processTransaction(creditor, share, debtor);
        }

        // if successful, we log it
        logTransaction(creditor, amount, debtorBreakdown);
    }

    public static void logTransaction(String creditor, int amount, String debtorBreakdown) {
        transactions.add(creditor + " paid " + amount + " for " + debtorBreakdown);
    }

    public static void processTransaction(String creditor, int amount, String debtor) {
        assert people.contains(creditor) && people.contains(debtor) : "Both people must exist";

        // OVERALL BALANCE UPDATES
        // the debtor person increases in debt
        balances.put(debtor, balances.getOrDefault(debtor, 0) - amount);

        // the creditor decreases in debt
        balances.put(creditor, balances.getOrDefault(creditor, 0) + amount);

        // DISTRIBUTION UPDATES
        // the debtor has debts to the creditor
        distribution.putIfAbsent(debtor, new HashMap<>());
        Map<String, Integer> debtorDistribution = distribution.get(debtor);
        debtorDistribution.put(creditor, debtorDistribution.getOrDefault(creditor, 0) - amount);

        // the creditor is owed by the debtor
        distribution.putIfAbsent(creditor, new HashMap<>());
        Map<String, Integer> creditorDistribution = distribution.get(creditor);
        creditorDistribution.put(debtor, creditorDistribution.getOrDefault(debtor, 0) + amount);
    }

    // simplify debts algorithm.
    // given a set of people and their balances, find the minimum number of transactions to settle all debts
    // we follow this algorithm:
    //   sort the people by balance. positive means they are a creditor, negative means they are a debtor.
    //   then take the person who currently owes the most, then have them give as much as they can to the person who is owed the most.
    //   repeat until everyone is settled.
    //   it is guaranteed that this algorithm will terminate since the sum of all balances is 0.
    // return this as a new map of distributions
    public static Map<String, Map<String, Integer>> simplifyDebts() {
        Map<String, Map<String, Integer>> simplifiedDistribution = new HashMap<>();

        // we create two priority queues of pos and neg balances

        // creditors
        PriorityQueue<Pair<String, Integer>> creditorQueue = new PriorityQueue<>((a, b) -> b.getValue() - a.getValue());
        // debtors
        PriorityQueue<Pair<String, Integer>> debtorQueue = new PriorityQueue<>(Comparator.comparingInt(Pair::getValue));

        // fill the queues
        for (String person : people) {
            int balance = balances.get(person);
            if (balance > 0) {
                creditorQueue.add(new Pair<>(person, balance));
            } else if (balance < 0) {
                debtorQueue.add(new Pair<>(person, balance));
            }
        }

        System.out.println(creditorQueue);
        System.out.println(debtorQueue);

        // execute the algorithm!
        while (!creditorQueue.isEmpty()) {
            Pair<String, Integer> creditor = creditorQueue.remove();
            Pair<String, Integer> debtor = debtorQueue.remove();

            int satisfied = Math.min(-1 * debtor.getValue(), creditor.getValue());

            // update the debtor's balance
            if (debtor.getValue() + satisfied < 0) {
                // debtor filled all of creditor balance, but they still owe more
                debtorQueue.add(new Pair<>(debtor.getKey(), debtor.getValue() + satisfied));
            } else if (creditor.getValue() - satisfied > 0) {
                // debtor could not fill all of creditor balance. they have paid all they need
                creditorQueue.add(new Pair<>(debtor.getKey(), debtor.getValue() - satisfied));
            }
            // note the else case is they perfectly satisfy.

            // regardless, log this transaction
            Map<String, Integer> debtorDistribution = simplifiedDistribution.getOrDefault(debtor.getKey(), new HashMap<>());
            // debtor should never had paid creditor before
            assert !debtorDistribution.containsKey(creditor.getKey()) : "Debtor should not have paid creditor before";
            debtorDistribution.put(creditor.getKey(), -1 * satisfied);

            Map<String, Integer> creditorDistribution = simplifiedDistribution.getOrDefault(creditor.getKey(), new HashMap<>());
            // creditor should never had been paid by debtor before
            assert !creditorDistribution.containsKey(debtor.getKey()) : "Creditor should not have been paid by debtor before";
            creditorDistribution.put(debtor.getKey(), satisfied);
        }

        assert debtorQueue.isEmpty() : "There should be an equal number of pos and neg balances";

        return simplifiedDistribution;
    }
}