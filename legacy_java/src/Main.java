import java.io.BufferedReader;
import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Arrays;
import java.util.Map;
import java.util.TreeSet;

public class Main {
    public static void main(String[] args) throws IOException {
        Splitwise sw = new Splitwise();

        parseCSV(sw, "");
//        System.out.println(sw.debtorDistributions);
//        System.out.println(sw.balances);

        System.out.println("original raw total transfers: " + sw.rawTotalTransfers());

        System.out.println();
        System.out.println("simplifying transactions...");
        Map<String, Map<String, BigDecimal>> simplifiedDistribution = sw.simplifyDebts();

        // sort the people
        TreeSet<String> sortedPeople = new TreeSet<>(simplifiedDistribution.keySet());

        for (String person : sortedPeople) {
            // get total that this person owes:
            BigDecimal total = simplifiedDistribution.get(person).values().stream().reduce(BigDecimal.ZERO, BigDecimal::add);
            System.out.println(person + " owes " + total.multiply(BigDecimal.valueOf(-1)) + " in total");
            for (String debtor : simplifiedDistribution.get(person).keySet()) {
                System.out.println("\t" + debtor + " " + simplifiedDistribution.get(person).get(debtor));
            }
        }
    }

    public static void parseCSV(Splitwise sw, String filename) throws IOException {
        BufferedReader br = new BufferedReader(new FileReader(filename));

        /* example format:
         * Date,Description,Category,Cost,Currency,A,B,C,D
         *
         * aaa,bbb,ccc,59.71,USD,0.00,-29.85,29.85,0.00
         * aaa,bbb,ccc,407.00,USD,-135.66,-135.67,-135.67,407.00
         * aaa,bbb,ccc,69.50,USD,-23.81,45.69,-21.88,0.00
         * aaa,bbb,ccc,6.99,USD,-2.33,-2.33,4.66,0.00
         * aaa,bbb,ccc,7.65,USD,-2.55,-2.55,5.10,0.00
         * aaa,bbb,ccc,17.73,USD,-5.91,-5.91,11.82,0.00
         *
         * aaa,Total balance, , ,USD,-170.26,-130.62,-106.12,407.00
         */

        // split by , and only consider from 5 on
        String[] people = br.readLine().split(",");
        people = Arrays.copyOfRange(people, 5, people.length);

        // ignore the first five entries. add a new person for each
        for (String person : people) {
            sw.newPerson(person);
        }
        // ignore next line
        br.readLine();

        // keep going until we hit just a newline
        while (true) {
            String line = br.readLine();
            if (line == null || line.equals("")) {
                break;
            }

            String[] tokens = line.split(",");

            String creditor = "";
            BigDecimal amount = new BigDecimal(tokens[3]);

            tokens = Arrays.copyOfRange(tokens, 5, tokens.length);

            // we need to search through first to find who has a positive amount.
            // this person is the creditor.
            for (int i = 0; i < tokens.length; i++) {
                if (new BigDecimal(tokens[i]).compareTo(BigDecimal.ZERO) > 0) {
                    creditor = people[i];
                    amount = new BigDecimal(tokens[i]);
                    break;
                }
            }
            assert creditor.length() > 0 : "No creditor found";

            // now we need to go through and find all the debtors
            StringBuilder sb = new StringBuilder();
            for (int i = 0; i < tokens.length; i++) {
                BigDecimal bal = new BigDecimal(tokens[i]);
                if (bal.compareTo(BigDecimal.ZERO) < 0) {
                    sb.append(people[i]);
                    sb.append("=");
                    sb.append(bal.multiply(BigDecimal.valueOf(-1)));
                    sb.append(",");
                }
            }

            // remove the last comma
            sb.deleteCharAt(sb.length() - 1);

            sw.newTransaction(creditor, amount, sb.toString());
        }
    }
}
