package Splitwise;

import java.math.BigDecimal;

public class Person implements Comparable<Person> {

    private int personID;
    private String name;

    private BigDecimal overallBalance;

    public Person(int personID, String name) {
        this.personID = personID;
        this.name = name;
        this.overallBalance = BigDecimal.ZERO;
    }

    @Override
    public int compareTo(Person other) {
        return Integer.compare(this.personID, other.personID);
    }
}
